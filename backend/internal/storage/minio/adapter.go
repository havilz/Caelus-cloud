package minio

import (
	"fmt"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/storage/s3"
)

// Config merepresentasikan konfigurasi koneksi ke instance MinIO Object Storage.
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

// NewAdapter menginisialisasi instance ObjectStorageAdapter khusus MinIO (S3-compatible dengan path-style addressing).
func NewAdapter(cfg Config) (domain.ObjectStorageAdapter, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:9000"
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	adapter, err := s3.NewAdapter(s3.Config{
		Endpoint:        cfg.Endpoint,
		Region:          cfg.Region,
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		UsePathStyle:    true, // MinIO membutuhkan path-style URL (http://endpoint/bucket/key)
		ProviderType:    domain.StorageProviderMinIO,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio storage adapter: %w", err)
	}

	return adapter, nil
}
