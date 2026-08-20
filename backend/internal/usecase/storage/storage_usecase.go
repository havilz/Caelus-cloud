package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// StorageUsecase mendefinisikan seluruh kontrak logika bisnis operasional Object Storage.
type StorageUsecase interface {
	// CreateBucket membuat bucket baru pada penyedia storage dan mencatat metadatanya ke database.
	CreateBucket(ctx context.Context, orgID uuid.UUID, name string, providerType domain.StorageProviderType, region string, isPublic, versioning bool) (*domain.Bucket, error)

	// ListBuckets mengambil daftar seluruh bucket milik organisasi.
	ListBuckets(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.Bucket, int, error)

	// GetBucket mengambil detail satu bucket berdasarkan nama dan kepemilikan organisasi.
	GetBucket(ctx context.Context, orgID uuid.UUID, name string) (*domain.Bucket, error)

	// DeleteBucket menghapus bucket dari database dan dari penyedia storage.
	DeleteBucket(ctx context.Context, orgID uuid.UUID, name string) error

	// ListObjects mengambil daftar objek di dalam bucket dengan opsi navigasi folder.
	ListObjects(ctx context.Context, orgID uuid.UUID, bucketName, prefix, delimiter string, maxKeys int32) ([]domain.ObjectItem, []string, error)

	// UploadObject mengunggah stream data objek ke dalam bucket.
	UploadObject(ctx context.Context, orgID uuid.UUID, bucketName, key string, body io.Reader, size int64, contentType string, metadata map[string]string) (*domain.ObjectItem, error)

	// DownloadObject mengunduh stream data objek dari bucket.
	DownloadObject(ctx context.Context, orgID uuid.UUID, bucketName, key string) (*domain.ObjectContent, error)

	// DeleteObject menghapus satu file/objek dari bucket.
	DeleteObject(ctx context.Context, orgID uuid.UUID, bucketName, key string) error

	// DeleteObjects menghapus beberapa file sekaligus dari bucket.
	DeleteObjects(ctx context.Context, orgID uuid.UUID, bucketName string, keys []string) error

	// GenerateSignedURL membuat URL bertanda tangan untuk unduh atau unggah file.
	GenerateSignedURL(ctx context.Context, orgID uuid.UUID, bucketName, key string, operation domain.SignedURLOperation, expiry time.Duration) (string, error)
}

type storageUsecase struct {
	bucketRepo domain.BucketRepository
	factory    domain.StorageFactory
}

// NewStorageUsecase membuat instance baru StorageUsecase.
func NewStorageUsecase(bucketRepo domain.BucketRepository, factory domain.StorageFactory) StorageUsecase {
	return &storageUsecase{
		bucketRepo: bucketRepo,
		factory:    factory,
	}
}

// CreateBucket membuat bucket baru pada penyedia storage dan mencatat metadatanya ke database.
func (u *storageUsecase) CreateBucket(ctx context.Context, orgID uuid.UUID, name string, providerType domain.StorageProviderType, region string, isPublic, versioning bool) (*domain.Bucket, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: bucket name cannot be empty", domain.ErrValidation)
	}

	if providerType == "" {
		providerType = domain.StorageProviderMinIO
	}
	if region == "" {
		region = "us-east-1"
	}

	// 1. Ambil adapter storage sesuai provider
	adapter, err := u.factory.GetAdapter(providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage adapter for provider %s: %w", providerType, err)
	}

	// 2. Buat bucket pada penyedia storage fisik/cloud
	if err := adapter.CreateBucket(ctx, name, region); err != nil {
		return nil, err
	}

	// 3. Simpan metadata ke PostgreSQL DB
	bucket := &domain.Bucket{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		ProviderType:   providerType,
		Region:         region,
		IsPublic:       isPublic,
		Versioning:     versioning,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := u.bucketRepo.Create(ctx, bucket); err != nil {
		// Rollback bucket di storage jika gagal simpan di database
		_ = adapter.DeleteBucket(ctx, name)
		return nil, fmt.Errorf("failed to persist bucket metadata: %w", err)
	}

	return bucket, nil
}

// ListBuckets mengambil daftar seluruh bucket milik organisasi.
func (u *storageUsecase) ListBuckets(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.Bucket, int, error) {
	return u.bucketRepo.ListByOrgID(ctx, orgID, limit, offset)
}

// GetBucket mengambil detail satu bucket berdasarkan nama dan kepemilikan organisasi.
func (u *storageUsecase) GetBucket(ctx context.Context, orgID uuid.UUID, name string) (*domain.Bucket, error) {
	bucket, err := u.bucketRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if bucket.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	return bucket, nil
}

// DeleteBucket menghapus bucket dari database dan dari penyedia storage.
func (u *storageUsecase) DeleteBucket(ctx context.Context, orgID uuid.UUID, name string) error {
	bucket, err := u.GetBucket(ctx, orgID, name)
	if err != nil {
		return err
	}

	adapter, err := u.factory.GetAdapter(bucket.ProviderType)
	if err != nil {
		return fmt.Errorf("failed to get storage adapter: %w", err)
	}

	// 1. Hapus dari penyedia storage (akan error jika bucket masih berisi objek)
	if err := adapter.DeleteBucket(ctx, name); err != nil {
		return err
	}

	// 2. Hapus dari basis data PostgreSQL
	return u.bucketRepo.Delete(ctx, bucket.ID)
}

// ListObjects mengambil daftar objek di dalam bucket dengan opsi navigasi folder.
func (u *storageUsecase) ListObjects(ctx context.Context, orgID uuid.UUID, bucketName, prefix, delimiter string, maxKeys int32) ([]domain.ObjectItem, []string, error) {
	bucket, err := u.GetBucket(ctx, orgID, bucketName)
	if err != nil {
		return nil, nil, err
	}

	adapter, err := u.factory.GetAdapter(bucket.ProviderType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get storage adapter: %w", err)
	}

	return adapter.ListObjects(ctx, bucketName, prefix, delimiter, maxKeys)
}

// UploadObject mengunggah stream data objek ke dalam bucket.
func (u *storageUsecase) UploadObject(ctx context.Context, orgID uuid.UUID, bucketName, key string, body io.Reader, size int64, contentType string, metadata map[string]string) (*domain.ObjectItem, error) {
	bucket, err := u.GetBucket(ctx, orgID, bucketName)
	if err != nil {
		return nil, err
	}

	adapter, err := u.factory.GetAdapter(bucket.ProviderType)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage adapter: %w", err)
	}

	input := domain.UploadObjectInput{
		BucketName:  bucketName,
		Key:         key,
		Body:        body,
		Size:        size,
		ContentType: contentType,
		Metadata:    metadata,
	}

	return adapter.UploadObject(ctx, input)
}

// DownloadObject mengunduh stream data objek dari bucket.
func (u *storageUsecase) DownloadObject(ctx context.Context, orgID uuid.UUID, bucketName, key string) (*domain.ObjectContent, error) {
	bucket, err := u.GetBucket(ctx, orgID, bucketName)
	if err != nil {
		return nil, err
	}

	adapter, err := u.factory.GetAdapter(bucket.ProviderType)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage adapter: %w", err)
	}

	return adapter.DownloadObject(ctx, bucketName, key)
}

// DeleteObject menghapus satu file/objek dari bucket.
func (u *storageUsecase) DeleteObject(ctx context.Context, orgID uuid.UUID, bucketName, key string) error {
	bucket, err := u.GetBucket(ctx, orgID, bucketName)
	if err != nil {
		return err
	}

	adapter, err := u.factory.GetAdapter(bucket.ProviderType)
	if err != nil {
		return fmt.Errorf("failed to get storage adapter: %w", err)
	}

	return adapter.DeleteObject(ctx, bucketName, key)
}

// DeleteObjects menghapus beberapa file sekaligus dari bucket.
func (u *storageUsecase) DeleteObjects(ctx context.Context, orgID uuid.UUID, bucketName string, keys []string) error {
	bucket, err := u.GetBucket(ctx, orgID, bucketName)
	if err != nil {
		return err
	}

	adapter, err := u.factory.GetAdapter(bucket.ProviderType)
	if err != nil {
		return fmt.Errorf("failed to get storage adapter: %w", err)
	}

	return adapter.DeleteObjects(ctx, bucketName, keys)
}

// GenerateSignedURL membuat URL bertanda tangan untuk unduh atau unggah file.
func (u *storageUsecase) GenerateSignedURL(ctx context.Context, orgID uuid.UUID, bucketName, key string, operation domain.SignedURLOperation, expiry time.Duration) (string, error) {
	bucket, err := u.GetBucket(ctx, orgID, bucketName)
	if err != nil {
		return "", err
	}

	adapter, err := u.factory.GetAdapter(bucket.ProviderType)
	if err != nil {
		return "", fmt.Errorf("failed to get storage adapter: %w", err)
	}

	return adapter.GenerateSignedURL(ctx, bucketName, key, operation, expiry)
}
