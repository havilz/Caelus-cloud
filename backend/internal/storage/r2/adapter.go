package r2

import (
	"fmt"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/storage/s3"
)

type Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

func NewAdapter(cfg Config) (domain.ObjectStorageAdapter, error) {
	if cfg.AccountID == "" {
		return nil, fmt.Errorf("%w: cloudflare account ID is required for R2", domain.ErrValidation)
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)
	region := cfg.Region
	if region == "" {
		region = "auto"
	}

	adapter, err := s3.NewAdapter(s3.Config{
		Endpoint:        endpoint,
		Region:          region,
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		UsePathStyle:    false,
		ProviderType:    domain.StorageProviderR2,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudflare R2 storage adapter: %w", err)
	}

	return adapter, nil
}
