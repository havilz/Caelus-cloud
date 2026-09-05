package queue

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

type ScheduledJob struct {
	ID          string
	Name        string
	Interval    time.Duration
	TaskType    TaskType
	PayloadFunc func() (*TaskPayload, error)
	LastRun     time.Time
}

type DistributedScheduler struct {
	queue    QueueEngine
	jobs     []*ScheduledJob
	jobsMu   sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup
	isClosed bool
	mu       sync.Mutex
}

func NewDistributedScheduler(q QueueEngine) *DistributedScheduler {
	return &DistributedScheduler{
		queue:    q,
		jobs:     make([]*ScheduledJob, 0),
		stopChan: make(chan struct{}),
	}
}

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
	logger.Info("Scheduled task registered successfully", "job_name", name, "interval", interval)
}

func (s *DistributedScheduler) Start(ctx context.Context) {
	s.StartWithInterval(ctx, 1*time.Second)
}

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

	logger.Info("Distributed Task Scheduler started successfully", "eval_interval", evalInterval)
}

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
	logger.Info("Distributed Task Scheduler stopped successfully")
}

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
				logger.Error("Failed to generate payload for scheduled task", "job_name", job.Name, "error", err)
				continue
			}

			if payload == nil {
				continue
			}

			if err := s.queue.Enqueue(ctx, payload); err != nil {
				logger.Error("Failed to enqueue scheduled task", "job_name", job.Name, "error", err)
			} else {
				logger.Debug("Scheduled task enqueued successfully", "job_name", job.Name, "task_id", payload.ID)
			}
		}
	}
}
