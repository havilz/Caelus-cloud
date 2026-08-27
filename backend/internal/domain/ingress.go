package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DomainStatus string
type SSLStatus string
type IngressTargetType string

const (
	DomainStatusPendingDNS DomainStatus = "pending_dns"
	DomainStatusVerifying  DomainStatus = "verifying"
	DomainStatusActive     DomainStatus = "active"
	DomainStatusError      DomainStatus = "error"

	SSLStatusNone    SSLStatus = "none"
	SSLStatusPending SSLStatus = "pending"
	SSLStatusActive  SSLStatus = "active"
	SSLStatusExpired SSLStatus = "expired"
	SSLStatusError   SSLStatus = "error"

	IngressTargetContainer IngressTargetType = "container"
	IngressTargetPort      IngressTargetType = "port"
	IngressTargetService   IngressTargetType = "service"
	IngressTargetStorage   IngressTargetType = "storage"
)

type CustomDomain struct {
	ID                   uuid.UUID         `json:"id"`
	OrganizationID       uuid.UUID         `json:"organization_id"`
	ServerID             *uuid.UUID        `json:"server_id,omitempty"`
	ServerName           string            `json:"server_name,omitempty"`
	ServerPublicIP       string            `json:"server_public_ip,omitempty"`
	DomainName           string            `json:"domain_name"`
	TargetType           IngressTargetType `json:"target_type"`
	TargetID             string            `json:"target_id"`
	TargetPort           int               `json:"target_port"`
	Status               DomainStatus      `json:"status"`
	VerificationToken    string            `json:"verification_token"`
	SSLStatus            SSLStatus         `json:"ssl_status"`
	AutoSSL              bool              `json:"auto_ssl"`
	CloudflareDNSManaged bool              `json:"cloudflare_dns_managed"`
	CloudflareRecordID   string            `json:"cloudflare_record_id,omitempty"`
	ErrorMessage         string            `json:"error_message,omitempty"`
	LastCheckedAt        *time.Time        `json:"last_checked_at,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

type CreateDomainRequest struct {
	ServerID             *uuid.UUID        `json:"server_id,omitempty"`
	DomainName           string            `json:"domain_name" validate:"required,hostname"`
	TargetType           IngressTargetType `json:"target_type" validate:"required,oneof=container port service storage"`
	TargetID             string            `json:"target_id" validate:"required"`
	TargetPort           int               `json:"target_port" validate:"required,min=1,max=65535"`
	AutoSSL              bool              `json:"auto_ssl"`
	CloudflareDNSManaged bool              `json:"cloudflare_dns_managed"`
}

type VerifyDomainResponse struct {
	DomainID      uuid.UUID    `json:"domain_id"`
	DomainName    string       `json:"domain_name"`
	Status        DomainStatus `json:"status"`
	Verified      bool         `json:"verified"`
	ExpectedIP    string       `json:"expected_ip"`
	ResolvedIPs   []string     `json:"resolved_ips"`
	ExpectedTXT   string       `json:"expected_txt"`
	ResolvedTXT   []string     `json:"resolved_txt"`
	SSLStatus     SSLStatus    `json:"ssl_status"`
	Message       string       `json:"message"`
}

type DomainRepository interface {
	Create(ctx context.Context, domain *CustomDomain) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*CustomDomain, error)
	GetByDomainName(ctx context.Context, domainName string) (*CustomDomain, error)
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]CustomDomain, error)
	Update(ctx context.Context, domain *CustomDomain) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status DomainStatus, errMsg string) error
	UpdateSSL(ctx context.Context, id uuid.UUID, sslStatus SSLStatus) error
}

type DomainUsecase interface {
	CreateDomain(ctx context.Context, orgID uuid.UUID, req *CreateDomainRequest) (*CustomDomain, error)
	GetDomain(ctx context.Context, orgID, id uuid.UUID) (*CustomDomain, error)
	ListDomains(ctx context.Context, orgID uuid.UUID) ([]CustomDomain, error)
	DeleteDomain(ctx context.Context, orgID, id uuid.UUID) error
	VerifyDomain(ctx context.Context, orgID, id uuid.UUID) (*VerifyDomainResponse, error)
}
