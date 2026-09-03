package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RuleTriggerType string

const (
	TriggerTypeMetricThreshold     RuleTriggerType = "metric_threshold"
	TriggerTypeServerStatusChanged RuleTriggerType = "server_status_changed"
	TriggerTypeBackupEvent         RuleTriggerType = "backup_event"
	TriggerTypeScheduledCron       RuleTriggerType = "scheduled_cron"
)

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

type RuleCondition struct {
	Field           string            `json:"field"`
	Operator        ConditionOperator `json:"operator"`
	Value           any               `json:"value"`
	DurationMinutes int               `json:"duration_minutes,omitempty"`
}

type ActionType string

const (
	ActionTypeSendEmail      ActionType = "send_email"
	ActionTypeSendWebhook    ActionType = "send_webhook"
	ActionTypeRebootServer   ActionType = "reboot_server"
	ActionTypeShutdownServer ActionType = "shutdown_server"
	ActionTypeTriggerBackup  ActionType = "trigger_backup"
	ActionTypeScaleServer    ActionType = "scale_server"
)

type RuleAction struct {
	Type   ActionType      `json:"type"`
	Target string          `json:"target,omitempty"`
	Config json.RawMessage `json:"config,omitempty"`
}

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

type ExecutionStatus string

const (
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusFailed  ExecutionStatus = "failed"
	ExecutionStatusPartial ExecutionStatus = "partially_failed"
	ExecutionStatusSkipped ExecutionStatus = "skipped"
)

type ActionResultItem struct {
	ActionType ActionType `json:"action_type"`
	Target     string     `json:"target,omitempty"`
	Status     string     `json:"status"`
	Response   string     `json:"response,omitempty"`
	Error      string     `json:"error,omitempty"`
}

type RuleExecutionLog struct {
	ID                  uuid.UUID          `json:"id"`
	RuleID              uuid.UUID          `json:"rule_id"`
	OrganizationID      uuid.UUID          `json:"organization_id"`
	RuleName            string             `json:"rule_name,omitempty"`
	TriggerEvent        string             `json:"trigger_event"`
	Status              ExecutionStatus    `json:"status"`
	EvaluatedConditions json.RawMessage    `json:"evaluated_conditions"`
	ExecutedActions     []ActionResultItem `json:"executed_actions"`
	ErrorMessage        string             `json:"error_message,omitempty"`
	ExecutionDurationMs int                `json:"execution_duration_ms"`
	ExecutedAt          time.Time          `json:"executed_at"`
}

type SystemEvent struct {
	ID             uuid.UUID      `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	Type           string         `json:"type"`
	SourceResource string         `json:"source_resource"`
	Data           map[string]any `json:"data"`
	OccurredAt     time.Time      `json:"occurred_at"`
}

type AutomationRepository interface {
	CreateRule(ctx context.Context, rule *AutomationRule) error

	GetRuleByID(ctx context.Context, orgID, id uuid.UUID) (*AutomationRule, error)

	ListRules(ctx context.Context, orgID uuid.UUID, page, limit int) ([]AutomationRule, int, error)

	ListActiveRulesByTriggerType(ctx context.Context, triggerType RuleTriggerType) ([]AutomationRule, error)

	UpdateRule(ctx context.Context, rule *AutomationRule) error

	UpdateLastTriggered(ctx context.Context, ruleID uuid.UUID, triggeredAt time.Time) error

	DeleteRule(ctx context.Context, orgID, id uuid.UUID) error

	CreateExecutionLog(ctx context.Context, log *RuleExecutionLog) error

	ListExecutionLogs(ctx context.Context, orgID uuid.UUID, ruleID *uuid.UUID, status *ExecutionStatus, page, limit int) ([]RuleExecutionLog, int, error)
}
