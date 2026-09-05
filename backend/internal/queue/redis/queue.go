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

type Config struct {
	Client         *redis.Client
	QueueName      string
	Concurrency    int
	PollTimeout    time.Duration
	BaseRetryDelay time.Duration
}

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

func (e *RedisQueueEngine) RegisterHandler(taskType queue.TaskType, handler queue.TaskHandler) {
	e.handlersMu.Lock()
	defer e.handlersMu.Unlock()
	e.handlers[taskType] = handler
}

func (e *RedisQueueEngine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.isClosed {
		e.mu.Unlock()
		return errors.New("cannot start stopped queue engine")
	}
	e.mu.Unlock()

	logger.Info("Starting Redis Queue Engine",
		"queue", e.queueName,
		"concurrency", e.concurrency,
	)

	e.wg.Add(1)
	go e.delayedTaskDispatcher(ctx)

	for i := 0; i < e.concurrency; i++ {
		e.wg.Add(1)
		go e.workerConsumer(ctx, i+1)
	}

	return nil
}

func (e *RedisQueueEngine) Stop() error {
	e.mu.Lock()
	if e.isClosed {
		e.mu.Unlock()
		return nil
	}
	e.isClosed = true
	close(e.stopChan)
	e.mu.Unlock()

	logger.Info("Stopping Redis Queue Engine gracefully...")
	e.wg.Wait()
	logger.Info("Redis Queue Engine stopped successfully")
	return nil
}

func (e *RedisQueueEngine) workerConsumer(ctx context.Context, workerID int) {
	defer e.wg.Done()
	logger.Debug("Worker consumer loop active", "worker_id", workerID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		default:

			res, err := e.client.BRPop(ctx, e.pollTimeout, e.queueName).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {

					continue
				}

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
				logger.Error("Failed to deserialize queue payload", "raw", taskRaw, "error", err)
				continue
			}

			e.processTask(ctx, &task)
		}
	}
}

func (e *RedisQueueEngine) processTask(ctx context.Context, task *queue.TaskPayload) {
	e.handlersMu.RLock()
	handler, exists := e.handlers[task.Type]
	e.handlersMu.RUnlock()

	if !exists {
		logger.Warn("No handler registered for task type", "task_type", task.Type, "task_id", task.ID)
		e.sendToDeadLetterQueue(ctx, task, fmt.Sprintf("no handler registered for task type %s", task.Type))
		return
	}

	execErr := handler(ctx, task)
	if execErr == nil {

		return
	}

	task.RetryCount++
	task.LastError = execErr.Error()

	if task.RetryCount > task.MaxRetries {
		logger.Error("Task exceeded maximum retry attempts, routed to Dead Letter Queue",
			"task_id", task.ID,
			"task_type", task.Type,
			"retry_count", task.RetryCount,
			"error", execErr,
		)
		e.sendToDeadLetterQueue(ctx, task, execErr.Error())
		return
	}

	multiplier := math.Pow(2, float64(task.RetryCount-1))
	delay := time.Duration(float64(e.baseRetryDelay) * multiplier)

	logger.Warn("Task processing failed, scheduling retry",
		"task_id", task.ID,
		"task_type", task.Type,
		"retry_count", task.RetryCount,
		"delay", delay,
		"error", execErr,
	)

	_ = e.EnqueueDelayed(ctx, task, delay)
}

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

			items, err := e.client.ZRangeByScore(ctx, e.delayedKey, &redis.ZRangeBy{
				Min:   "-inf",
				Max:   fmt.Sprintf("%f", nowScore),
				Count: 50,
			}).Result()

			if err != nil || len(items) == 0 {
				continue
			}

			for _, item := range items {

				removed, err := e.client.ZRem(ctx, e.delayedKey, item).Result()
				if err != nil || removed == 0 {
					continue
				}

				_ = e.client.LPush(ctx, e.queueName, item)
			}
		}
	}
}

func (e *RedisQueueEngine) sendToDeadLetterQueue(ctx context.Context, task *queue.TaskPayload, reason string) {
	task.LastError = reason
	data, err := json.Marshal(task)
	if err != nil {
		return
	}
	_ = e.client.LPush(ctx, e.deadLetterKey, data)
}
