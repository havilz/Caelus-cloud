package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/storage/s3"
	"github.com/havilz/caelus-cloud/backend/pkg/encryptor"
)

// StorageUsecase mendefinisikan seluruh kontrak logika bisnis operasional Object Storage.
type StorageUsecase interface {
	// CreateBucket membuat bucket baru pada penyedia storage dan mencatat metadatanya ke database.
	CreateBucket(ctx context.Context, orgID uuid.UUID, name string, providerType domain.StorageProviderType, region string, isPublic, versioning bool) (*domain.Bucket, error)

	// ListBuckets mengambil daftar seluruh bucket milik organisasi.
	ListBuckets(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.Bucket, int, error)

	// SyncBuckets melakukan sinkronisasi dua arah (Two-Way Discovery Sync) antara Caelus dan penyedia storage aktif (Cloudflare R2, AWS, MinIO).
	SyncBuckets(ctx context.Context, orgID uuid.UUID) ([]domain.Bucket, error)

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
	bucketRepo    domain.BucketRepository
	factory       domain.StorageFactory
	credRepo      domain.CredentialRepository
	encryptionKey []byte
}

func cleanInputToken(input string) string {
	input = strings.TrimSpace(input)
	if idx := strings.Index(input, ":"); idx != -1 {
		input = strings.TrimSpace(input[idx+1:])
	}
	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimPrefix(input, "http://")
	if idx := strings.Index(input, "."); idx != -1 {
		input = input[:idx]
	}
	var b strings.Builder
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// NewStorageUsecase membuat instance baru StorageUsecase dengan dukungan multi-cloud credential dinamis.
func NewStorageUsecase(bucketRepo domain.BucketRepository, factory domain.StorageFactory, credRepo domain.CredentialRepository, encryptionKey []byte) StorageUsecase {
	return &storageUsecase{
		bucketRepo:    bucketRepo,
		factory:       factory,
		credRepo:      credRepo,
		encryptionKey: encryptionKey,
	}
}

// getAdapter menyelesaikan adapter storage secara dinamis berdasarkan kredensial cloud aktif jika tersedia.
func (u *storageUsecase) getAdapter(ctx context.Context, orgID uuid.UUID, providerType domain.StorageProviderType) (domain.ObjectStorageAdapter, error) {
	if u.credRepo != nil && providerType != domain.StorageProviderMinIO && providerType != domain.StorageProviderMock {
		creds, err := u.credRepo.ListByOrg(ctx, orgID)
		if err == nil && len(creds) > 0 {
			targetSlug := string(providerType)
			if providerType == domain.StorageProviderR2 {
				targetSlug = "cloudflare"
			} else if providerType == domain.StorageProviderS3 {
				targetSlug = "aws"
			}

			for _, c := range creds {
				if c.Provider != nil && strings.EqualFold(c.Provider.Slug, targetSlug) {
					var accessKey, secretKey string
					if c.EncryptedAPIKey != nil {
						if dec, err := encryptor.Decrypt(*c.EncryptedAPIKey, u.encryptionKey); err == nil {
							accessKey = dec
						}
					}
					if c.EncryptedAPISecret != nil {
						if dec, err := encryptor.Decrypt(*c.EncryptedAPISecret, u.encryptionKey); err == nil {
							secretKey = dec
						}
					}

					endpoint := ""
					region := "us-east-1"
					if c.Metadata != nil {
						if r, ok := c.Metadata["region"].(string); ok && r != "" {
							region = r
						}
					}

					switch providerType {
					case domain.StorageProviderR2:
						region = "auto"
						accountID := ""
						if c.Metadata != nil {
							if acc, ok := c.Metadata["account_id"].(string); ok && acc != "" {
								accountID = cleanInputToken(acc)
							}
						}
						if accountID == "" {
							accountID = cleanInputToken(accessKey)
						}
						if accountID != "" {
							endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
						}
					case domain.StorageProviderDigitalOcean:
						if region == "auto" || region == "" {
							region = "sgp1"
						}
						endpoint = fmt.Sprintf("https://%s.digitaloceanspaces.com", region)
					case domain.StorageProviderGCP:
						endpoint = "https://storage.googleapis.com"
					}

					if accessKey != "" && secretKey != "" {
						adapter, err := s3.NewAdapter(s3.Config{
							Endpoint:        endpoint,
							AccessKeyID:     accessKey,
							SecretAccessKey: secretKey,
							Region:          region,
							ProviderType:    providerType,
							UsePathStyle:    providerType == domain.StorageProviderR2 || providerType == domain.StorageProviderGCP,
						})
						if err == nil {
							return adapter, nil
						}
					}
				}
			}
		}
	}

	return u.factory.GetAdapter(providerType)
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

	// 1. Ambil adapter storage dinamis sesuai provider dan kredensial terdaftar
	adapter, err := u.getAdapter(ctx, orgID, providerType)
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

// SyncBuckets melakukan sinkronisasi dua arah (Two-Way Discovery Sync) antara Caelus dan penyedia storage aktif (Cloudflare R2, AWS, MinIO).
func (u *storageUsecase) SyncBuckets(ctx context.Context, orgID uuid.UUID) ([]domain.Bucket, error) {
	existingDBBuckets, _, err := u.bucketRepo.ListByOrgID(ctx, orgID, 1000, 0)
	if err != nil {
		return nil, err
	}

	dbBucketMap := make(map[string]domain.Bucket)
	for _, b := range existingDBBuckets {
		dbBucketMap[fmt.Sprintf("%s:%s", b.ProviderType, b.Name)] = b
	}

	// 1. Kumpulkan seluruh provider yang aktif untuk organisasi ini
	providersToSync := []domain.StorageProviderType{domain.StorageProviderMinIO}

	if u.credRepo != nil {
		creds, _ := u.credRepo.ListByOrg(ctx, orgID)
		for _, c := range creds {
			if c.Provider != nil {
				slug := strings.ToLower(c.Provider.Slug)
				switch slug {
				case "cloudflare", "r2":
					providersToSync = append(providersToSync, domain.StorageProviderR2)
				case "aws", "s3":
					providersToSync = append(providersToSync, domain.StorageProviderS3)
				case "gcp":
					providersToSync = append(providersToSync, domain.StorageProviderGCP)
				case "digitalocean":
					providersToSync = append(providersToSync, domain.StorageProviderDigitalOcean)
				}
			}
		}
	}

	// 2. Iterasi setiap provider dan lakukan rekonsiliasi
	for _, prov := range providersToSync {
		adapter, err := u.getAdapter(ctx, orgID, prov)
		if err != nil || adapter == nil {
			continue
		}

		remoteBuckets, err := adapter.ListBuckets(ctx)
		if err != nil {
			continue
		}

		remoteNames := make(map[string]bool)
		for _, rb := range remoteBuckets {
			remoteNames[rb.Name] = true
			key := fmt.Sprintf("%s:%s", prov, rb.Name)

			if _, exists := dbBucketMap[key]; !exists {
				// Bucket baru terdeteksi di remote cloud / MinIO -> Daftarkan ke database
				newBucket := &domain.Bucket{
					ID:             uuid.New(),
					OrganizationID: orgID,
					Name:           rb.Name,
					ProviderType:   prov,
					Region:         rb.Region,
					IsPublic:       false,
					Versioning:     false,
					CreatedAt:      rb.CreatedAt,
					UpdatedAt:      time.Now().UTC(),
				}
				_ = u.bucketRepo.Create(ctx, newBucket)
			}
		}

		// Hapus bucket di DB yang sudah dihapus secara fisik di cloud untuk provider ini
		for _, b := range existingDBBuckets {
			if b.ProviderType == prov {
				if !remoteNames[b.Name] {
					_ = u.bucketRepo.Delete(ctx, b.ID)
				}
			}
		}
	}

	// Kembalikan daftar bucket terbaru setelah sinkronisasi
	synced, _, err := u.bucketRepo.ListByOrgID(ctx, orgID, 1000, 0)
	return synced, err
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
	bucket, err := u.bucketRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}

	adapter, err := u.getAdapter(ctx, orgID, bucket.ProviderType)
	if err == nil && adapter != nil {
		// 1. Hapus dari penyedia storage (otomatis membersihkan objek)
		if delErr := adapter.DeleteBucket(ctx, name); delErr != nil {
			// Jika error adalah BucketNotEmpty, hentikan proses dan beri tahu user
			if strings.Contains(delErr.Error(), "BucketNotEmpty") || strings.Contains(delErr.Error(), "409") {
				return delErr
			}
			// Abaikan error lain (misal NoSuchBucket, invalid endpoint, auth expired) dan tetap izinkan pembersihan record database
		}
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

	adapter, err := u.getAdapter(ctx, orgID, bucket.ProviderType)
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

	adapter, err := u.getAdapter(ctx, orgID, bucket.ProviderType)
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

	adapter, err := u.getAdapter(ctx, orgID, bucket.ProviderType)
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

	adapter, err := u.getAdapter(ctx, orgID, bucket.ProviderType)
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

	adapter, err := u.getAdapter(ctx, orgID, bucket.ProviderType)
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

	adapter, err := u.getAdapter(ctx, orgID, bucket.ProviderType)
	if err != nil {
		return "", fmt.Errorf("failed to get storage adapter: %w", err)
	}

	return adapter.GenerateSignedURL(ctx, bucketName, key, operation, expiry)
}
