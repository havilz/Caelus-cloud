package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	UserID         uuid.UUID  `json:"user_id"`
	Name           string     `json:"name"`
	KeyPrefix      string     `json:"key_prefix"`
	KeyHash        string     `json:"-"`
	Scopes         []string   `json:"scopes"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Token mentah hanya diisi saat pembuatan pertama kali (tidak disimpan di database)
	RawToken string `json:"raw_token,omitempty"`
}

type CreateAPIKeyRequest struct {
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	ExpiresIn *int     `json:"expires_in_days,omitempty"`
}

type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*APIKey, error)
	GetByHash(ctx context.Context, keyHash string) (*APIKey, error)
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]APIKey, error)
	UpdateLastUsed(ctx context.Context, id uuid.UUID, lastUsed time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
}
