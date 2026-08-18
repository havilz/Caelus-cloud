package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID             uuid.UUID      `json:"id"`
	OrganizationID *uuid.UUID     `json:"organization_id,omitempty"`
	UserID         *uuid.UUID     `json:"user_id,omitempty"`
	Action         string         `json:"action"`
	ResourceType   string         `json:"resource_type"`
	ResourceID     *string        `json:"resource_id,omitempty"`
	IPAddress      *string        `json:"ip_address,omitempty"`
	UserAgent      *string        `json:"user_agent,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	ListByOrg(ctx context.Context, orgID uuid.UUID, page, limit int) ([]AuditLog, int64, error)
}
