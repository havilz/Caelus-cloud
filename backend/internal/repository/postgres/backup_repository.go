package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type backupRepository struct {
	pool *pgxpool.Pool
}

// NewBackupRepository membuat instance baru PostgreSQL BackupRepository.
func NewBackupRepository(pool *pgxpool.Pool) domain.BackupRepository {
	return &backupRepository{pool: pool}
}

// CreatePolicy menyimpan kebijakan backup baru ke basis data.
func (r *backupRepository) CreatePolicy(ctx context.Context, policy *domain.BackupPolicy) error {
	query := `
		INSERT INTO backup_policies (
			id, organization_id, server_id, bucket_id, name, cron_expression,
			retention_days, include_disks, is_active, next_run_at, last_run_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at
	`

	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		policy.ID,
		policy.OrganizationID,
		policy.ServerID,
		policy.BucketID,
		policy.Name,
		policy.CronExpression,
		policy.RetentionDays,
		policy.IncludeDisks,
		policy.IsActive,
		policy.NextRunAt,
		policy.LastRunAt,
		policy.CreatedAt,
		policy.UpdatedAt,
	).Scan(&policy.CreatedAt, &policy.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create backup policy: %w", err)
	}

	return nil
}

// GetPolicyByID mengambil detail kebijakan backup berdasarkan ID beserta nama server & bucket-nya.
func (r *backupRepository) GetPolicyByID(ctx context.Context, id uuid.UUID) (*domain.BackupPolicy, error) {
	query := `
		SELECT 
			p.id, p.organization_id, p.server_id, p.bucket_id, p.name, p.cron_expression,
			p.retention_days, p.include_disks, p.is_active, p.next_run_at, p.last_run_at,
			p.created_at, p.updated_at,
			s.name AS server_name,
			b.name AS bucket_name
		FROM backup_policies p
		LEFT JOIN servers s ON s.id = p.server_id
		LEFT JOIN buckets b ON b.id = p.bucket_id
		WHERE p.id = $1
	`

	var p domain.BackupPolicy
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.OrganizationID,
		&p.ServerID,
		&p.BucketID,
		&p.Name,
		&p.CronExpression,
		&p.RetentionDays,
		&p.IncludeDisks,
		&p.IsActive,
		&p.NextRunAt,
		&p.LastRunAt,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.ServerName,
		&p.BucketName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query backup policy by ID: %w", err)
	}

	return &p, nil
}

// ListPoliciesByOrgID mengambil seluruh kebijakan backup milik organisasi.
func (r *backupRepository) ListPoliciesByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.BackupPolicy, error) {
	query := `
		SELECT 
			p.id, p.organization_id, p.server_id, p.bucket_id, p.name, p.cron_expression,
			p.retention_days, p.include_disks, p.is_active, p.next_run_at, p.last_run_at,
			p.created_at, p.updated_at,
			s.name AS server_name,
			b.name AS bucket_name
		FROM backup_policies p
		LEFT JOIN servers s ON s.id = p.server_id
		LEFT JOIN buckets b ON b.id = p.bucket_id
		WHERE p.organization_id = $1
		ORDER BY p.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup policies: %w", err)
	}
	defer rows.Close()

	policies := make([]domain.BackupPolicy, 0)
	for rows.Next() {
		var p domain.BackupPolicy
		if err := rows.Scan(
			&p.ID,
			&p.OrganizationID,
			&p.ServerID,
			&p.BucketID,
			&p.Name,
			&p.CronExpression,
			&p.RetentionDays,
			&p.IncludeDisks,
			&p.IsActive,
			&p.NextRunAt,
			&p.LastRunAt,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.ServerName,
			&p.BucketName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan backup policy row: %w", err)
		}
		policies = append(policies, p)
	}

	return policies, nil
}

// ListPoliciesByServerID mengambil kebijakan backup untuk server tertentu.
func (r *backupRepository) ListPoliciesByServerID(ctx context.Context, serverID uuid.UUID) ([]domain.BackupPolicy, error) {
	query := `
		SELECT 
			p.id, p.organization_id, p.server_id, p.bucket_id, p.name, p.cron_expression,
			p.retention_days, p.include_disks, p.is_active, p.next_run_at, p.last_run_at,
			p.created_at, p.updated_at,
			s.name AS server_name,
			b.name AS bucket_name
		FROM backup_policies p
		LEFT JOIN servers s ON s.id = p.server_id
		LEFT JOIN buckets b ON b.id = p.bucket_id
		WHERE p.server_id = $1
		ORDER BY p.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to list server backup policies: %w", err)
	}
	defer rows.Close()

	policies := make([]domain.BackupPolicy, 0)
	for rows.Next() {
		var p domain.BackupPolicy
		if err := rows.Scan(
			&p.ID,
			&p.OrganizationID,
			&p.ServerID,
			&p.BucketID,
			&p.Name,
			&p.CronExpression,
			&p.RetentionDays,
			&p.IncludeDisks,
			&p.IsActive,
			&p.NextRunAt,
			&p.LastRunAt,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.ServerName,
			&p.BucketName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan server backup policy: %w", err)
		}
		policies = append(policies, p)
	}

	return policies, nil
}

// ListDuePolicies mengambil daftar kebijakan aktif yang jadwal eksekusinya telah jatuh tempo.
func (r *backupRepository) ListDuePolicies(ctx context.Context, now time.Time) ([]domain.BackupPolicy, error) {
	query := `
		SELECT 
			p.id, p.organization_id, p.server_id, p.bucket_id, p.name, p.cron_expression,
			p.retention_days, p.include_disks, p.is_active, p.next_run_at, p.last_run_at,
			p.created_at, p.updated_at,
			s.name AS server_name,
			b.name AS bucket_name
		FROM backup_policies p
		LEFT JOIN servers s ON s.id = p.server_id
		LEFT JOIN buckets b ON b.id = p.bucket_id
		WHERE p.is_active = true AND (p.next_run_at IS NULL OR p.next_run_at <= $1)
		LIMIT 50
	`

	rows, err := r.pool.Query(ctx, query, now)
	if err != nil {
		return nil, fmt.Errorf("failed to query due backup policies: %w", err)
	}
	defer rows.Close()

	policies := make([]domain.BackupPolicy, 0)
	for rows.Next() {
		var p domain.BackupPolicy
		if err := rows.Scan(
			&p.ID,
			&p.OrganizationID,
			&p.ServerID,
			&p.BucketID,
			&p.Name,
			&p.CronExpression,
			&p.RetentionDays,
			&p.IncludeDisks,
			&p.IsActive,
			&p.NextRunAt,
			&p.LastRunAt,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.ServerName,
			&p.BucketName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan due backup policy: %w", err)
		}
		policies = append(policies, p)
	}

	return policies, nil
}

// UpdatePolicy memperbarui konfigurasi kebijakan backup.
func (r *backupRepository) UpdatePolicy(ctx context.Context, policy *domain.BackupPolicy) error {
	query := `
		UPDATE backup_policies
		SET 
			name = $2,
			cron_expression = $3,
			retention_days = $4,
			include_disks = $5,
			is_active = $6,
			bucket_id = $7,
			next_run_at = $8,
			last_run_at = $9,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		policy.ID,
		policy.Name,
		policy.CronExpression,
		policy.RetentionDays,
		policy.IncludeDisks,
		policy.IsActive,
		policy.BucketID,
		policy.NextRunAt,
		policy.LastRunAt,
	).Scan(&policy.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update backup policy: %w", err)
	}

	return nil
}

// DeletePolicy menghapus kebijakan backup.
func (r *backupRepository) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM backup_policies WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete backup policy: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CreateRecord membuat entri rekaman backup baru.
func (r *backupRepository) CreateRecord(ctx context.Context, record *domain.BackupRecord) error {
	query := `
		INSERT INTO backup_records (
			id, organization_id, policy_id, server_id, bucket_id, backup_name,
			storage_key, size_bytes, status, error_message, checksum_sha256,
			started_at, completed_at, expires_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at
	`

	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		record.ID,
		record.OrganizationID,
		record.PolicyID,
		record.ServerID,
		record.BucketID,
		record.BackupName,
		record.StorageKey,
		record.SizeBytes,
		record.Status,
		record.ErrorMessage,
		record.ChecksumSHA256,
		record.StartedAt,
		record.CompletedAt,
		record.ExpiresAt,
		record.CreatedAt,
		record.UpdatedAt,
	).Scan(&record.CreatedAt, &record.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert backup record: %w", err)
	}

	return nil
}

// GetRecordByID mengambil rekaman backup berdasarkan ID.
func (r *backupRepository) GetRecordByID(ctx context.Context, id uuid.UUID) (*domain.BackupRecord, error) {
	query := `
		SELECT 
			r.id, r.organization_id, r.policy_id, r.server_id, r.bucket_id, r.backup_name,
			r.storage_key, r.size_bytes, r.status, r.error_message, r.checksum_sha256,
			r.started_at, r.completed_at, r.expires_at, r.created_at, r.updated_at,
			s.name AS server_name,
			b.name AS bucket_name
		FROM backup_records r
		LEFT JOIN servers s ON s.id = r.server_id
		LEFT JOIN buckets b ON b.id = r.bucket_id
		WHERE r.id = $1
	`

	var rec domain.BackupRecord
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rec.ID,
		&rec.OrganizationID,
		&rec.PolicyID,
		&rec.ServerID,
		&rec.BucketID,
		&rec.BackupName,
		&rec.StorageKey,
		&rec.SizeBytes,
		&rec.Status,
		&rec.ErrorMessage,
		&rec.ChecksumSHA256,
		&rec.StartedAt,
		&rec.CompletedAt,
		&rec.ExpiresAt,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&rec.ServerName,
		&rec.BucketName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query backup record by ID: %w", err)
	}

	return &rec, nil
}

// UpdateRecordStatus memperbarui status proses, ukuran, dan error rekaman backup.
func (r *backupRepository) UpdateRecordStatus(ctx context.Context, id uuid.UUID, status domain.BackupStatus, sizeBytes int64, checksum, errMsg *string, completedAt *time.Time) error {
	query := `
		UPDATE backup_records
		SET 
			status = $2,
			size_bytes = $3,
			checksum_sha256 = $4,
			error_message = $5,
			completed_at = $6,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, status, sizeBytes, checksum, errMsg, completedAt)
	if err != nil {
		return fmt.Errorf("failed to update backup record status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ListRecordsByOrgID mengambil daftar riwayat backup milik organisasi dengan paginasi.
func (r *backupRepository) ListRecordsByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.BackupRecord, int, error) {
	countQuery := `SELECT COUNT(*) FROM backup_records WHERE organization_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count backup records: %w", err)
	}

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT 
			r.id, r.organization_id, r.policy_id, r.server_id, r.bucket_id, r.backup_name,
			r.storage_key, r.size_bytes, r.status, r.error_message, r.checksum_sha256,
			r.started_at, r.completed_at, r.expires_at, r.created_at, r.updated_at,
			s.name AS server_name,
			b.name AS bucket_name
		FROM backup_records r
		LEFT JOIN servers s ON s.id = r.server_id
		LEFT JOIN buckets b ON b.id = r.bucket_id
		WHERE r.organization_id = $1
		ORDER BY r.started_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list backup records: %w", err)
	}
	defer rows.Close()

	records := make([]domain.BackupRecord, 0)
	for rows.Next() {
		var rec domain.BackupRecord
		if err := rows.Scan(
			&rec.ID,
			&rec.OrganizationID,
			&rec.PolicyID,
			&rec.ServerID,
			&rec.BucketID,
			&rec.BackupName,
			&rec.StorageKey,
			&rec.SizeBytes,
			&rec.Status,
			&rec.ErrorMessage,
			&rec.ChecksumSHA256,
			&rec.StartedAt,
			&rec.CompletedAt,
			&rec.ExpiresAt,
			&rec.CreatedAt,
			&rec.UpdatedAt,
			&rec.ServerName,
			&rec.BucketName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan backup record row: %w", err)
		}
		records = append(records, rec)
	}

	return records, total, nil
}

// ListRecordsByServerID mengambil riwayat backup untuk server tertentu.
func (r *backupRepository) ListRecordsByServerID(ctx context.Context, serverID uuid.UUID, limit, offset int) ([]domain.BackupRecord, int, error) {
	countQuery := `SELECT COUNT(*) FROM backup_records WHERE server_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, serverID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count server backup records: %w", err)
	}

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT 
			r.id, r.organization_id, r.policy_id, r.server_id, r.bucket_id, r.backup_name,
			r.storage_key, r.size_bytes, r.status, r.error_message, r.checksum_sha256,
			r.started_at, r.completed_at, r.expires_at, r.created_at, r.updated_at,
			s.name AS server_name,
			b.name AS bucket_name
		FROM backup_records r
		LEFT JOIN servers s ON s.id = r.server_id
		LEFT JOIN buckets b ON b.id = r.bucket_id
		WHERE r.server_id = $1
		ORDER BY r.started_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, serverID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list server backup records: %w", err)
	}
	defer rows.Close()

	records := make([]domain.BackupRecord, 0)
	for rows.Next() {
		var rec domain.BackupRecord
		if err := rows.Scan(
			&rec.ID,
			&rec.OrganizationID,
			&rec.PolicyID,
			&rec.ServerID,
			&rec.BucketID,
			&rec.BackupName,
			&rec.StorageKey,
			&rec.SizeBytes,
			&rec.Status,
			&rec.ErrorMessage,
			&rec.ChecksumSHA256,
			&rec.StartedAt,
			&rec.CompletedAt,
			&rec.ExpiresAt,
			&rec.CreatedAt,
			&rec.UpdatedAt,
			&rec.ServerName,
			&rec.BucketName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan server backup record: %w", err)
		}
		records = append(records, rec)
	}

	return records, total, nil
}

// ListExpiredRecords mengambil daftar arsip backup yang masa retensinya telah kedaluwarsa.
func (r *backupRepository) ListExpiredRecords(ctx context.Context, now time.Time, limit int) ([]domain.BackupRecord, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT 
			r.id, r.organization_id, r.policy_id, r.server_id, r.bucket_id, r.backup_name,
			r.storage_key, r.size_bytes, r.status, r.error_message, r.checksum_sha256,
			r.started_at, r.completed_at, r.expires_at, r.created_at, r.updated_at,
			s.name AS server_name,
			b.name AS bucket_name
		FROM backup_records r
		LEFT JOIN servers s ON s.id = r.server_id
		LEFT JOIN buckets b ON b.id = r.bucket_id
		WHERE r.expires_at IS NOT NULL AND r.expires_at <= $1 AND r.status = 'completed'
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, now, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired backup records: %w", err)
	}
	defer rows.Close()

	records := make([]domain.BackupRecord, 0)
	for rows.Next() {
		var rec domain.BackupRecord
		if err := rows.Scan(
			&rec.ID,
			&rec.OrganizationID,
			&rec.PolicyID,
			&rec.ServerID,
			&rec.BucketID,
			&rec.BackupName,
			&rec.StorageKey,
			&rec.SizeBytes,
			&rec.Status,
			&rec.ErrorMessage,
			&rec.ChecksumSHA256,
			&rec.StartedAt,
			&rec.CompletedAt,
			&rec.ExpiresAt,
			&rec.CreatedAt,
			&rec.UpdatedAt,
			&rec.ServerName,
			&rec.BucketName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan expired backup record: %w", err)
		}
		records = append(records, rec)
	}

	return records, nil
}

// DeleteRecord menghapus rekaman metadata backup dari basis data.
func (r *backupRepository) DeleteRecord(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM backup_records WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete backup record: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
