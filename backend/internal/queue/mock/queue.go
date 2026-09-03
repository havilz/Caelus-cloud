package mock

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/queue"
)

type MockQueueEngine struct {
	tasks      []*queue.TaskPayload
	deadLetter []*queue.TaskPayload
	handlers   map[queue.TaskType]queue.TaskHandler
	mu         sync.Mutex

	taskChan chan *queue.TaskPayload
	stopChan chan struct{}
	wg       sync.WaitGroup
	isClosed bool
}

func NewMockQueueEngine() *MockQueueEngine {
	return &MockQueueEngine{
		tasks:      make([]*queue.TaskPayload, 0),
		deadLetter: make([]*queue.TaskPayload, 0),
		handlers:   make(map[queue.TaskType]queue.TaskHandler),
		taskChan:   make(chan *queue.TaskPayload, 100),
		stopChan:   make(chan struct{}),
	}
}

func (m *MockQueueEngine) Enqueue(ctx context.Context, task *queue.TaskPayload) error {
	if task == nil {
		return errors.New("task payload cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isClosed {
		return errors.New("queue engine is closed")
	}

	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now().UTC()
	}

	m.tasks = append(m.tasks, task)

	select {
	case m.taskChan <- task:
	default:
	}

	return nil
}

func (m *MockQueueEngine) EnqueueDelayed(ctx context.Context, task *queue.TaskPayload, delay time.Duration) error {
	if task == nil {
		return errors.New("task payload cannot be nil")
	}

	go func() {
		select {
		case <-time.After(delay):
			_ = m.Enqueue(context.Background(), task)
		case <-m.stopChan:
			return
		}
	}()

	return nil
}

func (m *MockQueueEngine) RegisterHandler(taskType queue.TaskType, handler queue.TaskHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[taskType] = handler
}

func (m *MockQueueEngine) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.isClosed {
		m.mu.Unlock()
		return errors.New("queue engine is closed")
	}
	m.mu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.stopChan:
				return
			case task := <-m.taskChan:
				m.mu.Lock()
				handler, exists := m.handlers[task.Type]
				m.mu.Unlock()

				if exists && handler != nil {
					_ = handler(ctx, task)
				}
			}
		}
	}()

	return nil
}

func (m *MockQueueEngine) Stop() error {
	m.mu.Lock()
	if m.isClosed {
		m.mu.Unlock()
		return nil
	}
	m.isClosed = true
	close(m.stopChan)
	m.mu.Unlock()

	m.wg.Wait()
	return nil
}

func (m *MockQueueEngine) GetAllTasks() []*queue.TaskPayload {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]*queue.TaskPayload, len(m.tasks))
	copy(copied, m.tasks)
	return copied
}
