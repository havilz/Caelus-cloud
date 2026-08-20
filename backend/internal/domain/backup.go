package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// BackupStatus mendefinisikan status siklus hidup proses pembuatan backup.
type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusInProgress BackupStatus = "in_progress"
	BackupStatusCompleted  BackupStatus = "completed"
	BackupStatusFailed     BackupStatus = "failed"
)

// BackupPolicy merepresentasikan konfigurasi kebijakan jadwal dan retensi backup server.
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

	// Relational views
	ServerName *string `json:"server_name,omitempty"`
	BucketName *string `json:"bucket_name,omitempty"`
}

// BackupRecord merepresentasikan rekaman arsip berkas backup yang tersimpan di Object Storage.
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

	// Relational views
	ServerName *string `json:"server_name,omitempty"`
	BucketName *string `json:"bucket_name,omitempty"`
}

// CreateBackupPolicyInput parameter pembuatan kebijakan backup.
type CreateBackupPolicyInput struct {
	OrganizationID uuid.UUID
	ServerID       uuid.UUID
	BucketID       *uuid.UUID
	Name           string
	CronExpression string
	RetentionDays  int
	IncludeDisks   bool
}

// BackupRepository mendefinisikan kontrak persistensi basis data untuk kebijakan dan arsip backup.
type BackupRepository interface {
	// CreatePolicy menyimpan kebijakan backup baru.
	CreatePolicy(ctx context.Context, policy *BackupPolicy) error

	// GetPolicyByID mengambil detail kebijakan backup berdasarkan ID.
	GetPolicyByID(ctx context.Context, id uuid.UUID) (*BackupPolicy, error)

	// ListPoliciesByOrgID mengambil seluruh kebijakan backup milik organisasi.
	ListPoliciesByOrgID(ctx context.Context, orgID uuid.UUID) ([]BackupPolicy, error)

	// ListPoliciesByServerID mengambil kebijakan backup untuk server tertentu.
	ListPoliciesByServerID(ctx context.Context, serverID uuid.UUID) ([]BackupPolicy, error)

	// ListDuePolicies mengambil daftar kebijakan aktif yang jadwal eksekusinya telah jatuh tempo.
	ListDuePolicies(ctx context.Context, now time.Time) ([]BackupPolicy, error)

	// UpdatePolicy memperbarui konfigurasi kebijakan backup.
	UpdatePolicy(ctx context.Context, policy *BackupPolicy) error

	// DeletePolicy menghapus kebijakan backup.
	DeletePolicy(ctx context.Context, id uuid.UUID) error

	// CreateRecord membuat entri rekaman backup baru.
	CreateRecord(ctx context.Context, record *BackupRecord) error

	// GetRecordByID mengambil rekaman backup berdasarkan ID.
	GetRecordByID(ctx context.Context, id uuid.UUID) (*BackupRecord, error)

	// UpdateRecordStatus memperbarui status proses, ukuran, dan error rekaman backup.
	UpdateRecordStatus(ctx context.Context, id uuid.UUID, status BackupStatus, sizeBytes int64, checksum, errMsg *string, completedAt *time.Time) error

	// ListRecordsByOrgID mengambil daftar riwayat backup milik organisasi dengan paginasi.
	ListRecordsByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]BackupRecord, int, error)

	// ListRecordsByServerID mengambil riwayat backup untuk server tertentu.
	ListRecordsByServerID(ctx context.Context, serverID uuid.UUID, limit, offset int) ([]BackupRecord, int, error)

	// ListExpiredRecords mengambil daftar arsip backup yang masa retensinya telah kedaluwarsa.
	ListExpiredRecords(ctx context.Context, now time.Time, limit int) ([]BackupRecord, error)

	// DeleteRecord menghapus rekaman metadata backup dari basis data.
	DeleteRecord(ctx context.Context, id uuid.UUID) error
}
