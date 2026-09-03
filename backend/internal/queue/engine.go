package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TaskType string

const (
	TaskTypeExecuteRuleAction       TaskType = "automation.execute_action"
	TaskTypeSendEmailNotification   TaskType = "notification.send_email"
	TaskTypeSendWebhookNotification TaskType = "notification.send_webhook"
	TaskTypeTriggerBackup           TaskType = "backup.trigger"
	TaskTypeCleanupTelemetry        TaskType = "telemetry.cleanup"
)

type TaskPayload struct {
	ID             uuid.UUID       `json:"id"`
	Type           TaskType        `json:"type"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	Data           json.RawMessage `json:"data"`
	MaxRetries     int             `json:"max_retries"`
	RetryCount     int             `json:"retry_count"`
	LastError      string          `json:"last_error,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	ExecuteAt      time.Time       `json:"execute_at"`
}

type TaskHandler func(ctx context.Context, payload *TaskPayload) error

type QueueEngine interface {
	Enqueue(ctx context.Context, task *TaskPayload) error

	EnqueueDelayed(ctx context.Context, task *TaskPayload, delay time.Duration) error

	RegisterHandler(taskType TaskType, handler TaskHandler)

	Start(ctx context.Context) error

	Stop() error
}
