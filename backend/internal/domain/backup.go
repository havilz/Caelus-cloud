package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusInProgress BackupStatus = "in_progress"
	BackupStatusCompleted  BackupStatus = "completed"
	BackupStatusFailed     BackupStatus = "failed"
)

type BackupPolicy struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	ServerID       uuid.UUID  `json:"server_id"`
	BucketID       *uuid.UUID `json:"bucket_id,omitempty"`
	Name           string     `json:"name"`
	CronExpression string     `json:"cron_expression"`
	RetentionDays  int        `json:"retention_days"`
	IncludeDisks   bool       `json:"include_disks"`
	IsActive       bool       `json:"is_active"`
	NextRunAt      *time.Time `json:"next_run_at,omitempty"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	ServerName *string `json:"server_name,omitempty"`
	BucketName *string `json:"bucket_name,omitempty"`
}

type BackupRecord struct {
	ID             uuid.UUID    `json:"id"`
	OrganizationID uuid.UUID    `json:"organization_id"`
	PolicyID       *uuid.UUID   `json:"policy_id,omitempty"`
	ServerID       uuid.UUID    `json:"server_id"`
	BucketID       *uuid.UUID   `json:"bucket_id,omitempty"`
	BackupName     string       `json:"backup_name"`
	StorageKey     string       `json:"storage_key"`
	SizeBytes      int64        `json:"size_bytes"`
	Status         BackupStatus `json:"status"`
	ErrorMessage   *string      `json:"error_message,omitempty"`
	ChecksumSHA256 *string      `json:"checksum_sha256,omitempty"`
	StartedAt      time.Time    `json:"started_at"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`
	ExpiresAt      *time.Time   `json:"expires_at,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`

	ServerName *string `json:"server_name,omitempty"`
	BucketName *string `json:"bucket_name,omitempty"`
}

type CreateBackupPolicyInput struct {
	OrganizationID uuid.UUID
	ServerID       uuid.UUID
	BucketID       *uuid.UUID
	Name           string
	CronExpression string
	RetentionDays  int
	IncludeDisks   bool
}

type BackupRepository interface {
	CreatePolicy(ctx context.Context, policy *BackupPolicy) error

	GetPolicyByID(ctx context.Context, id uuid.UUID) (*BackupPolicy, error)

	ListPoliciesByOrgID(ctx context.Context, orgID uuid.UUID) ([]BackupPolicy, error)

	ListPoliciesByServerID(ctx context.Context, serverID uuid.UUID) ([]BackupPolicy, error)

	ListDuePolicies(ctx context.Context, now time.Time) ([]BackupPolicy, error)

	UpdatePolicy(ctx context.Context, policy *BackupPolicy) error

	DeletePolicy(ctx context.Context, id uuid.UUID) error

	CreateRecord(ctx context.Context, record *BackupRecord) error

	GetRecordByID(ctx context.Context, id uuid.UUID) (*BackupRecord, error)

	UpdateRecordStatus(ctx context.Context, id uuid.UUID, status BackupStatus, sizeBytes int64, checksum, errMsg *string, completedAt *time.Time) error

	ListRecordsByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]BackupRecord, int, error)

	ListRecordsByServerID(ctx context.Context, serverID uuid.UUID, limit, offset int) ([]BackupRecord, int, error)

	ListExpiredRecords(ctx context.Context, now time.Time, limit int) ([]BackupRecord, error)

	DeleteRecord(ctx context.Context, id uuid.UUID) error
}
