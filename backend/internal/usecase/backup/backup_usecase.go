package backup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// BackupUsecase mendefinisikan seluruh kontrak logika bisnis operasional otomatisasi backup.
type BackupUsecase interface {
	// CreatePolicy membuat kebijakan backup baru untuk server.
	CreatePolicy(ctx context.Context, input domain.CreateBackupPolicyInput) (*domain.BackupPolicy, error)

	// ListPolicies mengambil seluruh kebijakan backup milik organisasi.
	ListPolicies(ctx context.Context, orgID uuid.UUID) ([]domain.BackupPolicy, error)

	// GetPolicy mengambil detail satu kebijakan backup.
	GetPolicy(ctx context.Context, orgID, policyID uuid.UUID) (*domain.BackupPolicy, error)

	// DeletePolicy menghapus kebijakan backup.
	DeletePolicy(ctx context.Context, orgID, policyID uuid.UUID) error

	// TriggerBackup melakukan pembuatan backup snapshot server secara langsung (on-demand atau scheduled).
	TriggerBackup(ctx context.Context, orgID, serverID uuid.UUID, policyID *uuid.UUID, backupName string) (*domain.BackupRecord, error)

	// ListRecords mengambil daftar rekaman arsip backup organisasi.
	ListRecords(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.BackupRecord, int, error)

	// GetRecord mengambil detail satu rekaman backup.
	GetRecord(ctx context.Context, orgID, recordID uuid.UUID) (*domain.BackupRecord, error)

	// DeleteRecord menghapus rekaman metadata dan berkas fisik backup di Object Storage.
	DeleteRecord(ctx context.Context, orgID, recordID uuid.UUID) error

	// CleanExpiredBackups menghapus berkas backup yang masa retensinya telah kedaluwarsa.
	CleanExpiredBackups(ctx context.Context) (int, error)
}

type backupUsecase struct {
	backupRepo domain.BackupRepository
	serverRepo domain.ServerRepository
	bucketRepo domain.BucketRepository
	factory    domain.StorageFactory
}

// NewBackupUsecase membuat instance baru BackupUsecase.
func NewBackupUsecase(
	backupRepo domain.BackupRepository,
	serverRepo domain.ServerRepository,
	bucketRepo domain.BucketRepository,
	factory domain.StorageFactory,
) BackupUsecase {
	return &backupUsecase{
		backupRepo: backupRepo,
		serverRepo: serverRepo,
		bucketRepo: bucketRepo,
		factory:    factory,
	}
}

// CreatePolicy membuat kebijakan backup baru untuk server.
func (u *backupUsecase) CreatePolicy(ctx context.Context, input domain.CreateBackupPolicyInput) (*domain.BackupPolicy, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("%w: policy name cannot be empty", domain.ErrValidation)
	}

	// 1. Validasi keberadaan dan kepemilikan server
	server, err := u.serverRepo.GetByID(ctx, input.ServerID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}
	if server.OrganizationID != input.OrganizationID {
		return nil, domain.ErrForbidden
	}

	// 2. Validasi bucket jika ditentukan
	if input.BucketID != nil {
		bucket, err := u.bucketRepo.GetByID(ctx, *input.BucketID)
		if err != nil {
			return nil, fmt.Errorf("target bucket not found: %w", err)
		}
		if bucket.OrganizationID != input.OrganizationID {
			return nil, domain.ErrForbidden
		}
	}

	if input.CronExpression == "" {
		input.CronExpression = "0 2 * * *"
	}
	if input.RetentionDays <= 0 {
		input.RetentionDays = 7
	}

	// Jadwalkan waktu eksekusi berikutnya (contoh: 24 jam ke depan)
	nextRun := time.Now().UTC().Add(24 * time.Hour)

	policy := &domain.BackupPolicy{
		ID:             uuid.New(),
		OrganizationID: input.OrganizationID,
		ServerID:       input.ServerID,
		BucketID:       input.BucketID,
		Name:           input.Name,
		CronExpression: input.CronExpression,
		RetentionDays:  input.RetentionDays,
		IncludeDisks:   input.IncludeDisks,
		IsActive:       true,
		NextRunAt:      &nextRun,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := u.backupRepo.CreatePolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to save backup policy: %w", err)
	}

	return policy, nil
}

// ListPolicies mengambil seluruh kebijakan backup milik organisasi.
func (u *backupUsecase) ListPolicies(ctx context.Context, orgID uuid.UUID) ([]domain.BackupPolicy, error) {
	return u.backupRepo.ListPoliciesByOrgID(ctx, orgID)
}

// GetPolicy mengambil detail satu kebijakan backup.
func (u *backupUsecase) GetPolicy(ctx context.Context, orgID, policyID uuid.UUID) (*domain.BackupPolicy, error) {
	policy, err := u.backupRepo.GetPolicyByID(ctx, policyID)
	if err != nil {
		return nil, err
	}
	if policy.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}
	return policy, nil
}

// DeletePolicy menghapus kebijakan backup.
func (u *backupUsecase) DeletePolicy(ctx context.Context, orgID, policyID uuid.UUID) error {
	policy, err := u.GetPolicy(ctx, orgID, policyID)
	if err != nil {
		return err
	}
	return u.backupRepo.DeletePolicy(ctx, policy.ID)
}

// TriggerBackup melakukan pembuatan backup snapshot server secara langsung (on-demand atau scheduled).
func (u *backupUsecase) TriggerBackup(ctx context.Context, orgID, serverID uuid.UUID, policyID *uuid.UUID, backupName string) (*domain.BackupRecord, error) {
	// 1. Validasi server
	server, err := u.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}
	if server.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	if backupName == "" {
		backupName = fmt.Sprintf("backup-%s-%s", server.Name, time.Now().UTC().Format("20060102-150405"))
	}

	// 2. Tentukan bucket target (default ke bucket pertama organisasi atau create default bucket)
	var targetBucket *domain.Bucket
	if policyID != nil {
		policy, err := u.backupRepo.GetPolicyByID(ctx, *policyID)
		if err == nil && policy.BucketID != nil {
			targetBucket, _ = u.bucketRepo.GetByID(ctx, *policy.BucketID)
		}
	}

	if targetBucket == nil {
		buckets, total, err := u.bucketRepo.ListByOrgID(ctx, orgID, 1, 0)
		if err == nil && total > 0 {
			targetBucket = &buckets[0]
		}
	}

	var bucketID *uuid.UUID
	var bucketName string
	providerType := domain.StorageProviderMinIO
	if targetBucket != nil {
		bucketID = &targetBucket.ID
		bucketName = targetBucket.Name
		providerType = targetBucket.ProviderType
	} else {
		// Fallback dummy bucket name jika belum ada bucket terdaftar
		bucketName = fmt.Sprintf("org-%s-backups", orgID.String()[:8])
	}

	storageKey := fmt.Sprintf("backups/%s/%s.tar.gz", server.ID, backupName)
	retentionDays := 7
	expiresAt := time.Now().UTC().Add(time.Duration(retentionDays) * 24 * time.Hour)

	// 3. Buat record awal berstatus pending
	record := &domain.BackupRecord{
		ID:             uuid.New(),
		OrganizationID: orgID,
		PolicyID:       policyID,
		ServerID:       server.ID,
		BucketID:       bucketID,
		BackupName:     backupName,
		StorageKey:     storageKey,
		SizeBytes:      0,
		Status:         domain.BackupStatusInProgress,
		StartedAt:      time.Now().UTC(),
		ExpiresAt:      &expiresAt,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := u.backupRepo.CreateRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	hostnameStr := server.Name
	if server.Hostname != nil {
		hostnameStr = *server.Hostname
	}

	// 4. Buat snapshot data sintetis/arsip server (mock snapshot stream)
	snapshotData := fmt.Appendf(
		nil,
		"CAELUS_BACKUP_ARCHIVE_V1\nServerID: %s\nHostname: %s\nCores: %d\nRAM: %dMB\nDisk: %dGB\nTimestamp: %s\nData: [ARCHIVED_SYSTEM_IMAGE_BLOCKS]",
		server.ID, hostnameStr, server.CPUCores, server.MemoryMB, server.DiskGB, time.Now().UTC().Format(time.RFC3339),
	)

	hash := sha256.Sum256(snapshotData)
	checksum := hex.EncodeToString(hash[:])
	sizeBytes := int64(len(snapshotData))

	// 5. Upload ke Object Storage via Adapter
	adapter, err := u.factory.GetAdapter(providerType)
	if err == nil {
		_ = adapter.CreateBucket(ctx, bucketName, "us-east-1")
		_, uploadErr := adapter.UploadObject(ctx, domain.UploadObjectInput{
			BucketName:  bucketName,
			Key:         storageKey,
			Body:        bytes.NewReader(snapshotData),
			Size:        sizeBytes,
			ContentType: "application/gzip",
			Metadata: map[string]string{
				"server_id": server.ID.String(),
				"checksum":  checksum,
			},
		})

		now := time.Now().UTC()
		if uploadErr != nil {
			errMsg := uploadErr.Error()
			_ = u.backupRepo.UpdateRecordStatus(ctx, record.ID, domain.BackupStatusFailed, 0, nil, &errMsg, &now)
			record.Status = domain.BackupStatusFailed
			record.ErrorMessage = &errMsg
			return record, nil
		}

		_ = u.backupRepo.UpdateRecordStatus(ctx, record.ID, domain.BackupStatusCompleted, sizeBytes, &checksum, nil, &now)
		record.Status = domain.BackupStatusCompleted
		record.SizeBytes = sizeBytes
		record.ChecksumSHA256 = &checksum
		record.CompletedAt = &now
	}

	return record, nil
}

// ListRecords mengambil daftar rekaman arsip backup organisasi.
func (u *backupUsecase) ListRecords(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.BackupRecord, int, error) {
	return u.backupRepo.ListRecordsByOrgID(ctx, orgID, limit, offset)
}

// GetRecord mengambil detail satu rekaman backup.
func (u *backupUsecase) GetRecord(ctx context.Context, orgID, recordID uuid.UUID) (*domain.BackupRecord, error) {
	rec, err := u.backupRepo.GetRecordByID(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if rec.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}
	return rec, nil
}

// DeleteRecord menghapus rekaman metadata dan berkas fisik backup di Object Storage.
func (u *backupUsecase) DeleteRecord(ctx context.Context, orgID, recordID uuid.UUID) error {
	rec, err := u.GetRecord(ctx, orgID, recordID)
	if err != nil {
		return err
	}

	// Hapus berkas fisik di object storage jika ada bucket terkait
	if rec.BucketID != nil {
		bucket, err := u.bucketRepo.GetByID(ctx, *rec.BucketID)
		if err == nil {
			if adapter, err := u.factory.GetAdapter(bucket.ProviderType); err == nil {
				_ = adapter.DeleteObject(ctx, bucket.Name, rec.StorageKey)
			}
		}
	}

	return u.backupRepo.DeleteRecord(ctx, rec.ID)
}

// CleanExpiredBackups menghapus berkas backup yang masa retensinya telah kedaluwarsa.
func (u *backupUsecase) CleanExpiredBackups(ctx context.Context) (int, error) {
	expiredRecords, err := u.backupRepo.ListExpiredRecords(ctx, time.Now().UTC(), 50)
	if err != nil {
		return 0, err
	}

	deletedCount := 0
	for _, rec := range expiredRecords {
		if rec.BucketID != nil {
			bucket, err := u.bucketRepo.GetByID(ctx, *rec.BucketID)
			if err == nil {
				if adapter, err := u.factory.GetAdapter(bucket.ProviderType); err == nil {
					_ = adapter.DeleteObject(ctx, bucket.Name, rec.StorageKey)
				}
			}
		}
		if err := u.backupRepo.DeleteRecord(ctx, rec.ID); err == nil {
			deletedCount++
		}
	}

	return deletedCount, nil
}
