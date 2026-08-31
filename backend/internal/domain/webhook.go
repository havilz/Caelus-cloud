package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Webhook struct {
	ID              uuid.UUID  `json:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id"`
	Name            string     `json:"name"`
	URL             string     `json:"url"`
	Secret          *string    `json:"secret,omitempty"`
	Events          []string   `json:"events"`
	IsActive        bool       `json:"is_active"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	LastStatus      *int       `json:"last_status,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateWebhookRequest struct {
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Secret *string  `json:"secret,omitempty"`
	Events []string `json:"events"`
}

type UpdateWebhookRequest struct {
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Secret   *string  `json:"secret,omitempty"`
	Events   []string `json:"events"`
	IsActive bool     `json:"is_active"`
}

type WebhookRepository interface {
	Create(ctx context.Context, webhook *Webhook) error
	GetByID(ctx context.Context, id uuid.UUID) (*Webhook, error)
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]Webhook, error)
	ListByEvent(ctx context.Context, orgID uuid.UUID, event string) ([]Webhook, error)
	Update(ctx context.Context, webhook *Webhook) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status int, triggeredAt time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
}
