package tests

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/storage"
	"github.com/havilz/caelus-cloud/backend/internal/storage/mock"
	storageUcPkg "github.com/havilz/caelus-cloud/backend/internal/usecase/storage"
)

// mockBucketRepo menyediakan implementasi in-memory domain.BucketRepository untuk pengujian usecase.
type mockBucketRepo struct {
	buckets map[uuid.UUID]*domain.Bucket
	mu      sync.RWMutex
}

func newMockBucketRepo() *mockBucketRepo {
	return &mockBucketRepo{
		buckets: make(map[uuid.UUID]*domain.Bucket),
	}
}

func (m *mockBucketRepo) Create(_ context.Context, bucket *domain.Bucket) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, b := range m.buckets {
		if b.Name == bucket.Name {
			return domain.ErrConflict
		}
	}

	if bucket.ID == uuid.Nil {
		bucket.ID = uuid.New()
	}
	m.buckets[bucket.ID] = bucket
	return nil
}

func (m *mockBucketRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Bucket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	b, exists := m.buckets[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return b, nil
}

func (m *mockBucketRepo) GetByName(_ context.Context, name string) (*domain.Bucket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, b := range m.buckets {
		if b.Name == name {
			return b, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockBucketRepo) ListByOrgID(_ context.Context, orgID uuid.UUID, limit, offset int) ([]domain.Bucket, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]domain.Bucket, 0)
	for _, b := range m.buckets {
		if b.OrganizationID == orgID {
			result = append(result, *b)
		}
	}

	total := len(result)
	if offset >= total {
		return []domain.Bucket{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return result[offset:end], total, nil
}

func (m *mockBucketRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[id]; !exists {
		return domain.ErrNotFound
	}
	delete(m.buckets, id)
	return nil
}

func (m *mockBucketRepo) CountByOrgID(_ context.Context, orgID uuid.UUID) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, b := range m.buckets {
		if b.OrganizationID == orgID {
			count++
		}
	}
	return count, nil
}

func TestStorageUsecase_BucketLifecycleAndObjects(t *testing.T) {
	ctx := context.Background()
	bucketRepo := newMockBucketRepo()
	factory := storage.NewStorageFactory()
	mockAdapter := mock.NewMockStorageAdapter()
	factory.RegisterAdapter(domain.StorageProviderMinIO, mockAdapter)

	uc := storageUcPkg.NewStorageUsecase(bucketRepo, factory, nil, nil)
	orgID := uuid.New()
	bucketName := "project-assets"

	// 1. Create Bucket
	bucket, err := uc.CreateBucket(ctx, orgID, bucketName, domain.StorageProviderMinIO, "us-east-1", false, false)
	if err != nil {
		t.Fatalf("failed to create bucket via usecase: %v", err)
	}
	if bucket.Name != bucketName || bucket.OrganizationID != orgID {
		t.Fatalf("unexpected bucket returned: %+v", bucket)
	}

	// 2. List Buckets
	buckets, total, err := uc.ListBuckets(ctx, orgID, 10, 0)
	if err != nil {
		t.Fatalf("failed to list buckets: %v", err)
	}
	if total != 1 || len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got total=%d, len=%d", total, len(buckets))
	}

	// 3. Upload Object
	content := "Welcome to Caelus Cloud multi-cloud storage!"
	uploadRes, err := uc.UploadObject(ctx, orgID, bucketName, "docs/intro.txt", strings.NewReader(content), int64(len(content)), "text/plain", nil)
	if err != nil {
		t.Fatalf("failed to upload object: %v", err)
	}
	if uploadRes.Key != "docs/intro.txt" {
		t.Fatalf("unexpected uploaded key: %s", uploadRes.Key)
	}

	// 4. Download Object
	downloadRes, err := uc.DownloadObject(ctx, orgID, bucketName, "docs/intro.txt")
	if err != nil {
		t.Fatalf("failed to download object: %v", err)
	}
	defer downloadRes.Body.Close()

	bodyBytes, _ := io.ReadAll(downloadRes.Body)
	if string(bodyBytes) != content {
		t.Fatalf("content mismatch. Expected %s, got %s", content, string(bodyBytes))
	}

	// 5. Generate Signed URL
	signedURL, err := uc.GenerateSignedURL(ctx, orgID, bucketName, "docs/intro.txt", domain.SignedURLOpDownload, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate signed URL: %v", err)
	}
	if !strings.Contains(signedURL, "docs/intro.txt") {
		t.Fatalf("invalid signed URL: %s", signedURL)
	}

	// 6. Delete Object
	if err := uc.DeleteObject(ctx, orgID, bucketName, "docs/intro.txt"); err != nil {
		t.Fatalf("failed to delete object: %v", err)
	}

	// 7. Delete Bucket
	if err := uc.DeleteBucket(ctx, orgID, bucketName); err != nil {
		t.Fatalf("failed to delete bucket: %v", err)
	}

	// 8. Verifikasi bucket sudah terhapus
	_, err = uc.GetBucket(ctx, orgID, bucketName)
	if err == nil {
		t.Fatal("expected bucket to be not found after deletion")
	}
}
