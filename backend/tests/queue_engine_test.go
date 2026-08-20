package tests

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/queue"
	"github.com/havilz/caelus-cloud/backend/internal/queue/mock"
)

// TestMockQueueEngine_Lifecycle menguji fungsionalitas dasar enqueue, pendaftaran handler, dan pemrosesan task.
func TestMockQueueEngine_Lifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	engine := mock.NewMockQueueEngine()

	var processedCount int32
	engine.RegisterHandler(queue.TaskTypeSendEmailNotification, func(ctx context.Context, payload *queue.TaskPayload) error {
		atomic.AddInt32(&processedCount, 1)
		return nil
	})

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("failed to start mock queue engine: %v", err)
	}
	defer engine.Stop()

	task := &queue.TaskPayload{
		ID:             uuid.New(),
		Type:           queue.TaskTypeSendEmailNotification,
		OrganizationID: uuid.New(),
		Data:           json.RawMessage(`{"to": "admin@caelus.cloud"}`),
	}

	if err := engine.Enqueue(ctx, task); err != nil {
		t.Fatalf("failed to enqueue task: %v", err)
	}

	// Tunggu pemrosesan
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&processedCount) != 1 {
		t.Errorf("expected 1 processed task, got %d", processedCount)
	}

	allTasks := engine.GetAllTasks()
	if len(allTasks) != 1 {
		t.Errorf("expected 1 task in history, got %d", len(allTasks))
	}
}

// TestDistributedScheduler_RegisterAndTrigger menguji registrasi dan eksekusi periodik scheduler terdistribusi.
func TestDistributedScheduler_RegisterAndTrigger(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	engine := mock.NewMockQueueEngine()
	scheduler := queue.NewDistributedScheduler(engine)

	var payloadCount int32
	scheduler.RegisterJob("test_telemetry_cleanup", 10*time.Millisecond, queue.TaskTypeCleanupTelemetry, func() (*queue.TaskPayload, error) {
		atomic.AddInt32(&payloadCount, 1)
		return &queue.TaskPayload{
			ID:   uuid.New(),
			Type: queue.TaskTypeCleanupTelemetry,
		}, nil
	})

	scheduler.StartWithInterval(ctx, 10*time.Millisecond)
	defer scheduler.Stop()

	// Tunggu setidaknya 1 siklus
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&payloadCount) < 1 {
		t.Errorf("expected at least 1 job trigger, got %d", payloadCount)
	}
}
