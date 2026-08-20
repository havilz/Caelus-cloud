package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/queue"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
	"github.com/redis/go-redis/v9"
)

const (
	defaultQueueName      = "caelus:tasks:active"
	defaultWorkerPoolSize = 5
	pollInterval          = 1 * time.Second
)

// Config mendefinisikan opsi konfigurasi koneksi dan operasi antrean Redis.
type Config struct {
	Client         *redis.Client
	QueueName      string
	Concurrency    int
	PollTimeout    time.Duration
	BaseRetryDelay time.Duration
}

// RedisQueueEngine mengimplementasikan interface QueueEngine menggunakan Redis.
type RedisQueueEngine struct {
	client         *redis.Client
	queueName      string
	delayedKey     string
	deadLetterKey  string
	concurrency    int
	pollTimeout    time.Duration
	baseRetryDelay time.Duration

	handlers   map[queue.TaskType]queue.TaskHandler
	handlersMu sync.RWMutex

	stopChan chan struct{}
	wg       sync.WaitGroup
	isClosed bool
	mu       sync.Mutex
}

// NewRedisQueueEngine membuat dan menginisialisasi instance RedisQueueEngine baru.
// Parameter cfg memuat konfigurasi Redis client dan parameter antrean.
// Mengembalikan pointer *RedisQueueEngine.
func NewRedisQueueEngine(cfg Config) *RedisQueueEngine {
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = defaultWorkerPoolSize
	}

	queueName := cfg.QueueName
	if queueName == "" {
		queueName = defaultQueueName
	}

	pollTimeout := cfg.PollTimeout
	if pollTimeout <= 0 {
		pollTimeout = 2 * time.Second
	}

	baseRetryDelay := cfg.BaseRetryDelay
	if baseRetryDelay <= 0 {
		baseRetryDelay = 5 * time.Second
	}

	return &RedisQueueEngine{
		client:         cfg.Client,
		queueName:      queueName,
		delayedKey:     fmt.Sprintf("%s:delayed", queueName),
		deadLetterKey:  fmt.Sprintf("%s:dead_letter", queueName),
		concurrency:    concurrency,
		pollTimeout:    pollTimeout,
		baseRetryDelay: baseRetryDelay,
		handlers:       make(map[queue.TaskType]queue.TaskHandler),
		stopChan:       make(chan struct{}),
	}
}

// Enqueue memasukkan tugas baru ke dalam antrean aktif Redis.
// Parameter ctx merupakan context pemanggilan.
// Parameter task merupakan payload pekerjaan yang akan dimasukkan.
// Mengembalikan error jika serialisasi JSON atau operasi LPUSH Redis gagal.
func (e *RedisQueueEngine) Enqueue(ctx context.Context, task *queue.TaskPayload) error {
	if task == nil {
		return errors.New("task payload cannot be nil")
	}

	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now().UTC()
	}
	if task.MaxRetries <= 0 {
		task.MaxRetries = 3
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	if err := e.client.LPush(ctx, e.queueName, data).Err(); err != nil {
		return fmt.Errorf("failed to push task to redis queue: %w", err)
	}

	return nil
}

// EnqueueDelayed memasukkan tugas ke antrean tertunda (Sorted Set) Redis dengan skor waktu eksekusi.
// Parameter ctx merupakan context pemanggilan.
// Parameter task merupakan payload pekerjaan yang akan dijadwalkan.
// Parameter delay merupakan durasi waktu tunda eksekusi.
// Mengembalikan error jika operasi ZADD Redis gagal.
func (e *RedisQueueEngine) EnqueueDelayed(ctx context.Context, task *queue.TaskPayload, delay time.Duration) error {
	if task == nil {
		return errors.New("task payload cannot be nil")
	}

	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now().UTC()
	}
	task.ExecuteAt = time.Now().UTC().Add(delay)

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal delayed task payload: %w", err)
	}

	score := float64(task.ExecuteAt.Unix())
	if err := e.client.ZAdd(ctx, e.delayedKey, redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("failed to schedule delayed task in redis: %w", err)
	}

	return nil
}

// RegisterHandler mendaftarkan fungsi pemroses untuk tipe tugas tertentu.
// Parameter taskType merupakan nama jenis tugas.
// Parameter handler merupakan fungsi pemroses tugas.
func (e *RedisQueueEngine) RegisterHandler(taskType queue.TaskType, handler queue.TaskHandler) {
	e.handlersMu.Lock()
	defer e.handlersMu.Unlock()
	e.handlers[taskType] = handler
}

// Start menjalankan consumer pool dan scheduler loop untuk memproses antrean.
// Parameter ctx merupakan context siklus hidup worker.
// Mengembalikan error jika engine telah ditutup atau bermasalah.
func (e *RedisQueueEngine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.isClosed {
		e.mu.Unlock()
		return errors.New("cannot start stopped queue engine")
	}
	e.mu.Unlock()

	logger.Info("Memulai Redis Queue Engine",
		"queue", e.queueName,
		"concurrency", e.concurrency,
	)

	// 1. Jalankan goroutine pemindah tugas tertunda (Delayed Tasks Dispatcher)
	e.wg.Add(1)
	go e.delayedTaskDispatcher(ctx)

	// 2. Jalankan worker pool consumer
	for i := 0; i < e.concurrency; i++ {
		e.wg.Add(1)
		go e.workerConsumer(ctx, i+1)
	}

	return nil
}

// Stop menghentikan seluruh worker consumer dan menutup antrean secara anggun.
// Mengembalikan error jika terjadi kegagalan saat penutupan.
func (e *RedisQueueEngine) Stop() error {
	e.mu.Lock()
	if e.isClosed {
		e.mu.Unlock()
		return nil
	}
	e.isClosed = true
	close(e.stopChan)
	e.mu.Unlock()

	logger.Info("Menghentikan Redis Queue Engine secara anggun...")
	e.wg.Wait()
	logger.Info("Redis Queue Engine berhasil dihentikan")
	return nil
}

// workerConsumer merupakan loop consumer yang membaca antrean via BRPOP dan memproses tugas.
func (e *RedisQueueEngine) workerConsumer(ctx context.Context, workerID int) {
	defer e.wg.Done()
	logger.Debug("Worker consumer loop aktif", "worker_id", workerID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		default:
			// Ambil item dari antrean dengan timeout blocking (BRPOP)
			res, err := e.client.BRPop(ctx, e.pollTimeout, e.queueName).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					// Queue kosong pada jendela polling timeout, lanjutkan loop
					continue
				}
				// Jika ada error jaringan sementara, tunggu sejenak
				select {
				case <-ctx.Done():
					return
				case <-e.stopChan:
					return
				case <-time.After(500 * time.Millisecond):
					continue
				}
			}

			if len(res) < 2 {
				continue
			}

			taskRaw := res[1]
			var task queue.TaskPayload
			if err := json.Unmarshal([]byte(taskRaw), &task); err != nil {
				logger.Error("Gagal mendeserialisasi payload antrean", "raw", taskRaw, "error", err)
				continue
			}

			e.processTask(ctx, &task)
		}
	}
}

// processTask mengeksekusi handler tugas, menangani retry bertahap (exponential backoff), atau mengirim ke Dead Letter Queue.
func (e *RedisQueueEngine) processTask(ctx context.Context, task *queue.TaskPayload) {
	e.handlersMu.RLock()
	handler, exists := e.handlers[task.Type]
	e.handlersMu.RUnlock()

	if !exists {
		logger.Warn("Tidak ditemukan handler untuk tipe tugas", "task_type", task.Type, "task_id", task.ID)
		e.sendToDeadLetterQueue(ctx, task, fmt.Sprintf("no handler registered for task type %s", task.Type))
		return
	}

	execErr := handler(ctx, task)
	if execErr == nil {
		// Tugas berhasil diselesaikan
		return
	}

	task.RetryCount++
	task.LastError = execErr.Error()

	if task.RetryCount > task.MaxRetries {
		logger.Error("Tugas melebihi batas percobaan maksimal, dialihkan ke Dead Letter Queue",
			"task_id", task.ID,
			"task_type", task.Type,
			"retry_count", task.RetryCount,
			"error", execErr,
		)
		e.sendToDeadLetterQueue(ctx, task, execErr.Error())
		return
	}

	// Hitung jeda waktu exponential backoff: baseDelay * 2^(retry-1)
	multiplier := math.Pow(2, float64(task.RetryCount-1))
	delay := time.Duration(float64(e.baseRetryDelay) * multiplier)

	logger.Warn("Tugas gagal diproses, menjadwalkan percobaan ulang",
		"task_id", task.ID,
		"task_type", task.Type,
		"retry_count", task.RetryCount,
		"delay", delay,
		"error", execErr,
	)

	_ = e.EnqueueDelayed(ctx, task, delay)
}

// delayedTaskDispatcher memeriksa tugas tertunda pada sorted set dan memindahkannya ke antrean aktif saat waktunya tiba.
func (e *RedisQueueEngine) delayedTaskDispatcher(ctx context.Context) {
	defer e.wg.Done()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		case <-ticker.C:
			nowScore := float64(time.Now().UTC().Unix())

			// Ambil semua item yang skornya <= waktu sekarang
			items, err := e.client.ZRangeByScore(ctx, e.delayedKey, &redis.ZRangeBy{
				Min:   "-inf",
				Max:   fmt.Sprintf("%f", nowScore),
				Count: 50,
			}).Result()

			if err != nil || len(items) == 0 {
				continue
			}

			for _, item := range items {
				// Hapus dari sorted set
				removed, err := e.client.ZRem(ctx, e.delayedKey, item).Result()
				if err != nil || removed == 0 {
					continue
				}

				// Masukkan ke antrean aktif
				_ = e.client.LPush(ctx, e.queueName, item)
			}
		}
	}
}

// sendToDeadLetterQueue menyimpan tugas yang gagal secara permanen ke daftar Dead Letter Queue untuk analisis audit.
func (e *RedisQueueEngine) sendToDeadLetterQueue(ctx context.Context, task *queue.TaskPayload, reason string) {
	task.LastError = reason
	data, err := json.Marshal(task)
	if err != nil {
		return
	}
	_ = e.client.LPush(ctx, e.deadLetterKey, data)
}
