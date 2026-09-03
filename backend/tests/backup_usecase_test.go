package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/storage"
	"github.com/havilz/caelus-cloud/backend/internal/storage/mock"
	backupUcPkg "github.com/havilz/caelus-cloud/backend/internal/usecase/backup"
)

type mockBackupRepo struct {
	policies map[uuid.UUID]*domain.BackupPolicy
	records  map[uuid.UUID]*domain.BackupRecord
	mu       sync.RWMutex
}

func newMockBackupRepo() *mockBackupRepo {
	return &mockBackupRepo{
		policies: make(map[uuid.UUID]*domain.BackupPolicy),
		records:  make(map[uuid.UUID]*domain.BackupRecord),
	}
}

func (m *mockBackupRepo) CreatePolicy(_ context.Context, policy *domain.BackupPolicy) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	m.policies[policy.ID] = policy
	return nil
}

func (m *mockBackupRepo) GetPolicyByID(_ context.Context, id uuid.UUID) (*domain.BackupPolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, exists := m.policies[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockBackupRepo) ListPoliciesByOrgID(_ context.Context, orgID uuid.UUID) ([]domain.BackupPolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]domain.BackupPolicy, 0)
	for _, p := range m.policies {
		if p.OrganizationID == orgID {
			res = append(res, *p)
		}
	}
	return res, nil
}

func (m *mockBackupRepo) ListPoliciesByServerID(_ context.Context, serverID uuid.UUID) ([]domain.BackupPolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]domain.BackupPolicy, 0)
	for _, p := range m.policies {
		if p.ServerID == serverID {
			res = append(res, *p)
		}
	}
	return res, nil
}

func (m *mockBackupRepo) ListDuePolicies(_ context.Context, now time.Time) ([]domain.BackupPolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]domain.BackupPolicy, 0)
	for _, p := range m.policies {
		if p.IsActive && (p.NextRunAt == nil || p.NextRunAt.Before(now) || p.NextRunAt.Equal(now)) {
			res = append(res, *p)
		}
	}
	return res, nil
}

func (m *mockBackupRepo) UpdatePolicy(_ context.Context, policy *domain.BackupPolicy) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.policies[policy.ID] = policy
	return nil
}

func (m *mockBackupRepo) DeletePolicy(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.policies, id)
	return nil
}

func (m *mockBackupRepo) CreateRecord(_ context.Context, record *domain.BackupRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	m.records[record.ID] = record
	return nil
}

func (m *mockBackupRepo) GetRecordByID(_ context.Context, id uuid.UUID) (*domain.BackupRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, exists := m.records[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return r, nil
}

func (m *mockBackupRepo) UpdateRecordStatus(_ context.Context, id uuid.UUID, status domain.BackupStatus, sizeBytes int64, checksum, errMsg *string, completedAt *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, exists := m.records[id]
	if !exists {
		return domain.ErrNotFound
	}
	r.Status = status
	r.SizeBytes = sizeBytes
	r.ChecksumSHA256 = checksum
	r.ErrorMessage = errMsg
	r.CompletedAt = completedAt
	return nil
}

func (m *mockBackupRepo) ListRecordsByOrgID(_ context.Context, orgID uuid.UUID, limit, offset int) ([]domain.BackupRecord, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]domain.BackupRecord, 0)
	for _, r := range m.records {
		if r.OrganizationID == orgID {
			res = append(res, *r)
		}
	}
	total := len(res)
	if offset >= total {
		return []domain.BackupRecord{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return res[offset:end], total, nil
}

func (m *mockBackupRepo) ListRecordsByServerID(_ context.Context, serverID uuid.UUID, limit, offset int) ([]domain.BackupRecord, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]domain.BackupRecord, 0)
	for _, r := range m.records {
		if r.ServerID == serverID {
			res = append(res, *r)
		}
	}
	total := len(res)
	if offset >= total {
		return []domain.BackupRecord{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return res[offset:end], total, nil
}

func (m *mockBackupRepo) ListExpiredRecords(_ context.Context, now time.Time, limit int) ([]domain.BackupRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]domain.BackupRecord, 0)
	for _, r := range m.records {
		if r.ExpiresAt != nil && (r.ExpiresAt.Before(now) || r.ExpiresAt.Equal(now)) && r.Status == domain.BackupStatusCompleted {
			res = append(res, *r)
			if len(res) >= limit {
				break
			}
		}
	}
	return res, nil
}

func (m *mockBackupRepo) DeleteRecord(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.records, id)
	return nil
}

func TestBackupUsecase_PolicyAndSnapshotLifecycle(t *testing.T) {
	ctx := context.Background()
	backupRepo := newMockBackupRepo()
	bucketRepo := newMockBucketRepo()
	serverRepo := newMockServerRepo()
	factory := storage.NewStorageFactory()
	mockAdapter := mock.NewMockStorageAdapter()
	factory.RegisterAdapter(domain.StorageProviderMinIO, mockAdapter)

	uc := backupUcPkg.NewBackupUsecase(backupRepo, serverRepo, bucketRepo, factory)

	orgID := uuid.New()
	hostname := "web-node-01.caelus.internal"
	server := &domain.Server{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "web-node-01",
		Hostname:       &hostname,
		CPUCores:       4,
		MemoryMB:       8192,
		DiskGB:         100,
		Status:         domain.ServerStatusRunning,
	}
	_ = serverRepo.Create(ctx, server)

	policy, err := uc.CreatePolicy(ctx, domain.CreateBackupPolicyInput{
		OrganizationID: orgID,
		ServerID:       server.ID,
		Name:           "Daily Database Backup",
		CronExpression: "0 3 * * *",
		RetentionDays:  7,
		IncludeDisks:   true,
	})
	if err != nil {
		t.Fatalf("failed to create backup policy: %v", err)
	}
	if policy.Name != "Daily Database Backup" || policy.ServerID != server.ID {
		t.Fatalf("unexpected policy created: %+v", policy)
	}

	policies, err := uc.ListPolicies(ctx, orgID)
	if err != nil || len(policies) != 1 {
		t.Fatalf("expected 1 policy, got len=%d, err=%v", len(policies), err)
	}

	record, err := uc.TriggerBackup(ctx, orgID, server.ID, &policy.ID, "manual-snapshot-01")
	if err != nil {
		t.Fatalf("failed to trigger backup: %v", err)
	}
	if record.Status != domain.BackupStatusCompleted {
		t.Fatalf("expected backup status completed, got %s", record.Status)
	}
	if record.SizeBytes <= 0 || record.ChecksumSHA256 == nil {
		t.Fatalf("expected valid size and sha256 checksum in backup record: %+v", record)
	}

	records, total, err := uc.ListRecords(ctx, orgID, 10, 0)
	if err != nil || total != 1 || len(records) != 1 {
		t.Fatalf("expected 1 backup record, got total=%d, err=%v", total, err)
	}

	if err := uc.DeleteRecord(ctx, orgID, record.ID); err != nil {
		t.Fatalf("failed to delete backup record: %v", err)
	}

	if err := uc.DeletePolicy(ctx, orgID, policy.ID); err != nil {
		t.Fatalf("failed to delete backup policy: %v", err)
	}
}
