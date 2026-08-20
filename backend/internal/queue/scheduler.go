package queue

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

// ScheduledJob merepresentasikan definisi pekerjaan berulang dengan interval waktu.
type ScheduledJob struct {
	ID          string
	Name        string
	Interval    time.Duration
	TaskType    TaskType
	PayloadFunc func() (*TaskPayload, error)
	LastRun     time.Time
}

// DistributedScheduler mengelola eksekusi tugas berkala dan memasukkannya ke TaskQueue.
type DistributedScheduler struct {
	queue    QueueEngine
	jobs     []*ScheduledJob
	jobsMu   sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup
	isClosed bool
	mu       sync.Mutex
}

// NewDistributedScheduler membuat instance baru DistributedScheduler.
// Parameter q merupakan engine antrean tempat pekerjaan dijadwalkan.
// Mengembalikan pointer *DistributedScheduler.
func NewDistributedScheduler(q QueueEngine) *DistributedScheduler {
	return &DistributedScheduler{
		queue:    q,
		jobs:     make([]*ScheduledJob, 0),
		stopChan: make(chan struct{}),
	}
}

// RegisterJob mendaftarkan tugas berulang dengan interval eksekusi tertentu.
// Parameter name merupakan nama tugas penjadwalan.
// Parameter interval merupakan durasi jeda waktu antar eksekusi.
// Parameter taskType merupakan tipe tugas antrean.
// Parameter payloadFunc merupakan fungsi pembuat payload saat waktu eksekusi tiba.
func (s *DistributedScheduler) RegisterJob(name string, interval time.Duration, taskType TaskType, payloadFunc func() (*TaskPayload, error)) {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	job := &ScheduledJob{
		ID:          uuid.New().String(),
		Name:        name,
		Interval:    interval,
		TaskType:    taskType,
		PayloadFunc: payloadFunc,
		LastRun:     time.Now().UTC(),
	}
	s.jobs = append(s.jobs, job)
	logger.Info("Pekerjaan terjadwal berhasil didaftarkan", "job_name", name, "interval", interval)
}

// Start menjalankan loop scheduler untuk mengevaluasi seluruh pekerjaan berkala dengan interval default 1 detik.
// Parameter ctx merupakan context siklus hidup scheduler.
func (s *DistributedScheduler) Start(ctx context.Context) {
	s.StartWithInterval(ctx, 1*time.Second)
}

// StartWithInterval menjalankan loop scheduler dengan interval pengecekan kustom.
// Parameter ctx merupakan context siklus hidup scheduler.
// Parameter evalInterval merupakan durasi siklus pengecekan jadwal pekerjaan.
func (s *DistributedScheduler) StartWithInterval(ctx context.Context, evalInterval time.Duration) {
	s.mu.Lock()
	if s.isClosed {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	if evalInterval <= 0 {
		evalInterval = 1 * time.Second
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(evalInterval)
		defer ticker.Stop()

		// Jalankan evaluasi awal saat pertama kali start
		s.evaluateJobs(ctx, time.Now().UTC())

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopChan:
				return
			case now := <-ticker.C:
				s.evaluateJobs(ctx, now)
			}
		}
	}()

	logger.Info("Distributed Task Scheduler berhasil dijalankan", "eval_interval", evalInterval)
}

// Stop menghentikan loop penjadwalan secara aman.
func (s *DistributedScheduler) Stop() {
	s.mu.Lock()
	if s.isClosed {
		s.mu.Unlock()
		return
	}
	s.isClosed = true
	close(s.stopChan)
	s.mu.Unlock()

	s.wg.Wait()
	logger.Info("Distributed Task Scheduler berhasil dihentikan")
}

// evaluateJobs memeriksa seluruh job yang terdaftar dan memasukkannya ke antrean jika interval telah terpenuhi.
func (s *DistributedScheduler) evaluateJobs(ctx context.Context, now time.Time) {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	for _, job := range s.jobs {
		if now.Sub(job.LastRun) >= job.Interval {
			job.LastRun = now

			if job.PayloadFunc == nil {
				continue
			}

			payload, err := job.PayloadFunc()
			if err != nil {
				logger.Error("Gagal membuat payload untuk pekerjaan terjadwal", "job_name", job.Name, "error", err)
				continue
			}

			if payload == nil {
				continue
			}

			if err := s.queue.Enqueue(ctx, payload); err != nil {
				logger.Error("Gagal memasukkan tugas terjadwal ke antrean", "job_name", job.Name, "error", err)
			} else {
				logger.Debug("Tugas terjadwal berhasil dimasukkan ke antrean", "job_name", job.Name, "task_id", payload.ID)
			}
		}
	}
}
