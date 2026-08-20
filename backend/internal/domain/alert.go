package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
)

// Alert merepresentasikan entitas insiden peringatan sistem yang terpicu oleh anomali metrik.
type Alert struct {
	ID             uuid.UUID     `json:"id"`
	OrganizationID uuid.UUID     `json:"organization_id"`
	ServerID       uuid.UUID     `json:"server_id"`
	RuleID         *uuid.UUID    `json:"rule_id,omitempty"`
	AlertType      string        `json:"alert_type"`
	Severity       AlertSeverity `json:"severity"`
	Title          string        `json:"title"`
	Message        string        `json:"message"`
	Status         AlertStatus   `json:"status"`
	CurrentValue   *float64      `json:"current_value,omitempty"`
	ThresholdValue *float64      `json:"threshold_value,omitempty"`
	AcknowledgedAt *time.Time    `json:"acknowledged_at,omitempty"`
	AcknowledgedBy *uuid.UUID    `json:"acknowledged_by,omitempty"`
	ResolvedAt     *time.Time    `json:"resolved_at,omitempty"`
	ResolvedBy     *uuid.UUID    `json:"resolved_by,omitempty"`
	TriggeredAt    time.Time     `json:"triggered_at"`
	CreatedAt      time.Time     `json:"created_at"`
	Server         *Server       `json:"server,omitempty"`
}

// AlertRule merepresentasikan aturan ambang batas evaluasi otomatis untuk metrik tertentu.
type AlertRule struct {
	ID              uuid.UUID     `json:"id"`
	OrganizationID  uuid.UUID     `json:"organization_id"`
	ServerID        *uuid.UUID    `json:"server_id,omitempty"`
	Name            string        `json:"name"`
	MetricName      string        `json:"metric_name"`
	Operator        string        `json:"operator"`
	Threshold       float64       `json:"threshold"`
	DurationSeconds int           `json:"duration_seconds"`
	Severity        AlertSeverity `json:"severity"`
	IsEnabled       bool          `json:"is_enabled"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// AlertRepository mendefinisikan interface persistensi alert dan aturan threshold.
type AlertRepository interface {
	CreateAlert(ctx context.Context, alert *Alert) error
	GetAlertByID(ctx context.Context, id uuid.UUID) (*Alert, error)
	ListAlertsByOrg(ctx context.Context, orgID uuid.UUID, status *AlertStatus, page, limit int) ([]Alert, int64, error)
	ListActiveAlertsByServer(ctx context.Context, serverID uuid.UUID) ([]Alert, error)
	UpdateAlertStatus(ctx context.Context, id uuid.UUID, status AlertStatus, userID *uuid.UUID, timestamp *time.Time) error
	CreateRule(ctx context.Context, rule *AlertRule) error
	GetRuleByID(ctx context.Context, id uuid.UUID) (*AlertRule, error)
	ListRulesByOrg(ctx context.Context, orgID uuid.UUID) ([]AlertRule, error)
	ListRulesForServer(ctx context.Context, orgID, serverID uuid.UUID) ([]AlertRule, error)
	DeleteRule(ctx context.Context, id uuid.UUID) error
}
