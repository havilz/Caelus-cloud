package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Provider struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Credential struct {
	ID                 uuid.UUID      `json:"id"`
	OrganizationID     uuid.UUID      `json:"organization_id"`
	ProviderID         uuid.UUID      `json:"provider_id"`
	Name               string         `json:"name"`
	EncryptedAPIKey    *string        `json:"-"`
	EncryptedAPISecret *string        `json:"-"`
	EncryptedSSHKey    *string        `json:"-"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	Provider           *Provider      `json:"provider,omitempty"`
}

type ProviderRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Provider, error)
	GetBySlug(ctx context.Context, slug string) (*Provider, error)
	List(ctx context.Context) ([]Provider, error)
}

type CredentialRepository interface {
	Create(ctx context.Context, cred *Credential) error
	GetByID(ctx context.Context, id uuid.UUID) (*Credential, error)
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]Credential, error)
	Update(ctx context.Context, cred *Credential) error
	Delete(ctx context.Context, id uuid.UUID) error
}
