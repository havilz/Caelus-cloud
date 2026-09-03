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

type StorageUsecase interface {
	CreateBucket(ctx context.Context, orgID uuid.UUID, name string, providerType domain.StorageProviderType, region string, isPublic, versioning bool) (*domain.Bucket, error)

	ListBuckets(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.Bucket, int, error)

	SyncBuckets(ctx context.Context, orgID uuid.UUID) ([]domain.Bucket, error)

	GetBucket(ctx context.Context, orgID uuid.UUID, name string) (*domain.Bucket, error)

	DeleteBucket(ctx context.Context, orgID uuid.UUID, name string) error

	ListObjects(ctx context.Context, orgID uuid.UUID, bucketName, prefix, delimiter string, maxKeys int32) ([]domain.ObjectItem, []string, error)

	UploadObject(ctx context.Context, orgID uuid.UUID, bucketName, key string, body io.Reader, size int64, contentType string, metadata map[string]string) (*domain.ObjectItem, error)

	DownloadObject(ctx context.Context, orgID uuid.UUID, bucketName, key string) (*domain.ObjectContent, error)

	DeleteObject(ctx context.Context, orgID uuid.UUID, bucketName, key string) error

	DeleteObjects(ctx context.Context, orgID uuid.UUID, bucketName string, keys []string) error

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

func NewStorageUsecase(bucketRepo domain.BucketRepository, factory domain.StorageFactory, credRepo domain.CredentialRepository, encryptionKey []byte) StorageUsecase {
	return &storageUsecase{
		bucketRepo:    bucketRepo,
		factory:       factory,
		credRepo:      credRepo,
		encryptionKey: encryptionKey,
	}
}

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

	adapter, err := u.getAdapter(ctx, orgID, providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage adapter for provider %s: %w", providerType, err)
	}

	if err := adapter.CreateBucket(ctx, name, region); err != nil {
		return nil, err
	}

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

		_ = adapter.DeleteBucket(ctx, name)
		return nil, fmt.Errorf("failed to persist bucket metadata: %w", err)
	}

	return bucket, nil
}

func (u *storageUsecase) ListBuckets(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.Bucket, int, error) {
	return u.bucketRepo.ListByOrgID(ctx, orgID, limit, offset)
}

func (u *storageUsecase) SyncBuckets(ctx context.Context, orgID uuid.UUID) ([]domain.Bucket, error) {
	existingDBBuckets, _, err := u.bucketRepo.ListByOrgID(ctx, orgID, 1000, 0)
	if err != nil {
		return nil, err
	}

	dbBucketMap := make(map[string]domain.Bucket)
	for _, b := range existingDBBuckets {
		dbBucketMap[fmt.Sprintf("%s:%s", b.ProviderType, b.Name)] = b
	}

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

		for _, b := range existingDBBuckets {
			if b.ProviderType == prov {
				if !remoteNames[b.Name] {
					_ = u.bucketRepo.Delete(ctx, b.ID)
				}
			}
		}
	}

	synced, _, err := u.bucketRepo.ListByOrgID(ctx, orgID, 1000, 0)
	return synced, err
}

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

		if delErr := adapter.DeleteBucket(ctx, name); delErr != nil {

			if strings.Contains(delErr.Error(), "BucketNotEmpty") || strings.Contains(delErr.Error(), "409") {
				return delErr
			}

		}
	}

	return u.bucketRepo.Delete(ctx, bucket.ID)
}

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
