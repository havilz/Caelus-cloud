package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RuleTriggerType mendefinisikan tipe kejadian pemicu suatu aturan otomasi.
type RuleTriggerType string

const (
	TriggerTypeMetricThreshold    RuleTriggerType = "metric_threshold"
	TriggerTypeServerStatusChanged RuleTriggerType = "server_status_changed"
	TriggerTypeBackupEvent        RuleTriggerType = "backup_event"
	TriggerTypeScheduledCron      RuleTriggerType = "scheduled_cron"
)

// ConditionOperator mendefinisikan operator perbandingan logika pada evaluasi kondisi.
type ConditionOperator string

const (
	OpGreaterThan      ConditionOperator = ">"
	OpGreaterThanEqual ConditionOperator = ">="
	OpLessThan         ConditionOperator = "<"
	OpLessThanEqual    ConditionOperator = "<="
	OpEqual            ConditionOperator = "=="
	OpNotEqual         ConditionOperator = "!="
	OpIn               ConditionOperator = "in"
	OpContains         ConditionOperator = "contains"
)

// RuleCondition merepresentasikan satu klausa kondisi yang harus dipenuhi agar aksi dieksekusi.
type RuleCondition struct {
	Field           string            `json:"field"`                      // misal: "cpu_usage_percent", "memory_usage_percent", "status"
	Operator        ConditionOperator `json:"operator"`                   // misal: ">=", "==", "<"
	Value           any               `json:"value"`                      // nilai ambang batas pembanding
	DurationMinutes int               `json:"duration_minutes,omitempty"` // durasi berturut-turut kondisi harus terpenuhi
}

// ActionType mendefinisikan jenis aksi yang dijalankan saat kondisi aturan terpenuhi.
type ActionType string

const (
	ActionTypeSendEmail      ActionType = "send_email"
	ActionTypeSendWebhook    ActionType = "send_webhook"
	ActionTypeRebootServer   ActionType = "reboot_server"
	ActionTypeShutdownServer ActionType = "shutdown_server"
	ActionTypeTriggerBackup  ActionType = "trigger_backup"
	ActionTypeScaleServer    ActionType = "scale_server"
)

// RuleAction merepresentasikan satu aksi yang akan dijalankan oleh rule engine.
type RuleAction struct {
	Type   ActionType      `json:"type"`
	Target string          `json:"target,omitempty"` // email address, webhook URL, atau server ID
	Config json.RawMessage `json:"config,omitempty"` // parameter tambahan spesifik aksi
}

// AutomationRule merepresentasikan entitas aturan otomasi Event-Condition-Action (ECA).
type AutomationRule struct {
	ID              uuid.UUID       `json:"id"`
	OrganizationID  uuid.UUID       `json:"organization_id"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	IsActive        bool            `json:"is_active"`
	TriggerType     RuleTriggerType `json:"trigger_type"`
	TriggerConfig   json.RawMessage `json:"trigger_config"`
	Conditions      []RuleCondition `json:"conditions"`
	Actions         []RuleAction    `json:"actions"`
	CooldownSeconds int             `json:"cooldown_seconds"`
	LastTriggeredAt *time.Time      `json:"last_triggered_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// CreateRuleInput merepresentasikan parameter masukan untuk pembuatan aturan otomasi baru.
type CreateRuleInput struct {
	OrganizationID  uuid.UUID       `json:"-"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	IsActive        *bool           `json:"is_active,omitempty"`
	TriggerType     RuleTriggerType `json:"trigger_type"`
	TriggerConfig   json.RawMessage `json:"trigger_config,omitempty"`
	Conditions      []RuleCondition `json:"conditions"`
	Actions         []RuleAction    `json:"actions"`
	CooldownSeconds int             `json:"cooldown_seconds,omitempty"`
}

// UpdateRuleInput merepresentasikan parameter masukan untuk pembaruan aturan otomasi.
type UpdateRuleInput struct {
	Name            *string          `json:"name,omitempty"`
	Description     *string          `json:"description,omitempty"`
	IsActive        *bool            `json:"is_active,omitempty"`
	TriggerType     *RuleTriggerType `json:"trigger_type,omitempty"`
	TriggerConfig   json.RawMessage  `json:"trigger_config,omitempty"`
	Conditions      []RuleCondition  `json:"conditions,omitempty"`
	Actions         []RuleAction     `json:"actions,omitempty"`
	CooldownSeconds *int             `json:"cooldown_seconds,omitempty"`
}

// ExecutionStatus mendefinisikan status hasil eksekusi aturan otomasi.
type ExecutionStatus string

const (
	ExecutionStatusSuccess   ExecutionStatus = "success"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusPartial   ExecutionStatus = "partially_failed"
	ExecutionStatusSkipped   ExecutionStatus = "skipped"
)

// ActionResultItem mencatat rincian hasil eksekusi per aksi.
type ActionResultItem struct {
	ActionType ActionType `json:"action_type"`
	Target     string     `json:"target,omitempty"`
	Status     string     `json:"status"` // "success" / "failed"
	Response   string     `json:"response,omitempty"`
	Error      string     `json:"error,omitempty"`
}

// RuleExecutionLog merepresentasikan catatan audit jejak eksekusi aturan otomasi.
type RuleExecutionLog struct {
	ID                   uuid.UUID          `json:"id"`
	RuleID               uuid.UUID          `json:"rule_id"`
	OrganizationID       uuid.UUID          `json:"organization_id"`
	RuleName             string             `json:"rule_name,omitempty"`
	TriggerEvent         string             `json:"trigger_event"`
	Status               ExecutionStatus    `json:"status"`
	EvaluatedConditions  json.RawMessage    `json:"evaluated_conditions"`
	ExecutedActions      []ActionResultItem `json:"executed_actions"`
	ErrorMessage         string             `json:"error_message,omitempty"`
	ExecutionDurationMs  int                `json:"execution_duration_ms"`
	ExecutedAt           time.Time          `json:"executed_at"`
}

// SystemEvent merepresentasikan payload kejadian sistem yang didistribusikan ke Rule Engine.
type SystemEvent struct {
	ID             uuid.UUID      `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	Type           string         `json:"type"` // "metric.cpu_high", "server.down", "backup.failed"
	SourceResource string         `json:"source_resource"` // "server:833cf7c3-..."
	Data           map[string]any `json:"data"`
	OccurredAt     time.Time      `json:"occurred_at"`
}

// AutomationRepository mendefinisikan antarmuka repositori data aturan dan log eksekusi otomasi.
type AutomationRepository interface {
	// CreateRule membuat aturan otomasi baru di database.
	CreateRule(ctx context.Context, rule *AutomationRule) error

	// GetRuleByID mengambil satu aturan berdasarkan ID dan ID organisasi.
	GetRuleByID(ctx context.Context, orgID, id uuid.UUID) (*AutomationRule, error)

	// ListRules mengambil seluruh aturan otomasi milik organisasi.
	ListRules(ctx context.Context, orgID uuid.UUID, page, limit int) ([]AutomationRule, int, error)

	// ListActiveRulesByTriggerType mengambil aturan aktif berdasarkan tipe trigger untuk evaluasi cepat.
	ListActiveRulesByTriggerType(ctx context.Context, triggerType RuleTriggerType) ([]AutomationRule, error)

	// UpdateRule memperbarui data aturan otomasi.
	UpdateRule(ctx context.Context, rule *AutomationRule) error

	// UpdateLastTriggered memperbarui stempel waktu terakhir aturan dieksekusi.
	UpdateLastTriggered(ctx context.Context, ruleID uuid.UUID, triggeredAt time.Time) error

	// DeleteRule menghapus aturan otomasi.
	DeleteRule(ctx context.Context, orgID, id uuid.UUID) error

	// CreateExecutionLog mencatat log riwayat eksekusi ke database.
	CreateExecutionLog(ctx context.Context, log *RuleExecutionLog) error

	// ListExecutionLogs mengambil daftar riwayat log eksekusi berpaginasi.
	ListExecutionLogs(ctx context.Context, orgID uuid.UUID, ruleID *uuid.UUID, status *ExecutionStatus, page, limit int) ([]RuleExecutionLog, int, error)
}
