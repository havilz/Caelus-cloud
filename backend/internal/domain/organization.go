package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OrganizationRole string

const (
	RoleOwner  OrganizationRole = "owner"
	RoleAdmin  OrganizationRole = "admin"
	RoleMember OrganizationRole = "member"
	RoleViewer OrganizationRole = "viewer"
)

type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Tier      string    `json:"tier"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrganizationMember struct {
	ID             uuid.UUID        `json:"id"`
	OrganizationID uuid.UUID        `json:"organization_id"`
	UserID         uuid.UUID        `json:"user_id"`
	Role           OrganizationRole `json:"role"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	User           *User            `json:"user,omitempty"`
}

type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id uuid.UUID) error

	AddMember(ctx context.Context, member *OrganizationMember) error
	GetMember(ctx context.Context, orgID, userID uuid.UUID) (*OrganizationMember, error)
	ListMembers(ctx context.Context, orgID uuid.UUID) ([]OrganizationMember, error)
	UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role OrganizationRole) error
	RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error
}
