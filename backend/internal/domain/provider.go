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

type CreateServerRequest struct {
	Name      string            `json:"name"`
	Region    string            `json:"region"`
	OSType    string            `json:"os_type"`
	PlanID    string            `json:"plan_id"`
	CPUCores  int               `json:"cpu_cores"`
	MemoryMB  int               `json:"memory_mb"`
	DiskGB    int               `json:"disk_gb"`
	SSHKey    string            `json:"ssh_key,omitempty"`
	UserData  string            `json:"user_data,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type ResizeServerRequest struct {
	ExternalID string `json:"external_id"`
	PlanID     string `json:"plan_id"`
	CPUCores   int    `json:"cpu_cores"`
	MemoryMB   int    `json:"memory_mb"`
	DiskGB     int    `json:"disk_gb"`
}

type ProviderServer struct {
	ExternalID string       `json:"external_id"`
	Name       string       `json:"name"`
	Status     ServerStatus `json:"status"`
	PublicIP   string       `json:"public_ip"`
	PrivateIP  string       `json:"private_ip,omitempty"`
	Region     string       `json:"region"`
	CPUCores   int          `json:"cpu_cores"`
	MemoryMB   int          `json:"memory_mb"`
	DiskGB     int          `json:"disk_gb"`
	CreatedAt  time.Time    `json:"created_at"`
}

type ProviderDriver interface {
	CreateServer(ctx context.Context, cred *Credential, req CreateServerRequest) (*ProviderServer, error)
	GetServer(ctx context.Context, cred *Credential, externalID string) (*ProviderServer, error)
	ListServers(ctx context.Context, cred *Credential) ([]ProviderServer, error)
	RebootServer(ctx context.Context, cred *Credential, externalID string) error
	ShutdownServer(ctx context.Context, cred *Credential, externalID string) error
	StartServer(ctx context.Context, cred *Credential, externalID string) error
	ResizeServer(ctx context.Context, cred *Credential, req ResizeServerRequest) error
	DeleteServer(ctx context.Context, cred *Credential, externalID string) error
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
