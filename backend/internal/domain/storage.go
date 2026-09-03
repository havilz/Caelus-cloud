package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type StorageProviderType string

const (
	StorageProviderMinIO        StorageProviderType = "minio"
	StorageProviderS3           StorageProviderType = "s3"
	StorageProviderAWS          StorageProviderType = "aws"
	StorageProviderR2           StorageProviderType = "r2"
	StorageProviderDigitalOcean StorageProviderType = "digitalocean"
	StorageProviderGCP          StorageProviderType = "gcp"
	StorageProviderMock         StorageProviderType = "mock"
)

type Bucket struct {
	ID             uuid.UUID           `json:"id"`
	OrganizationID uuid.UUID           `json:"organization_id"`
	Name           string              `json:"name"`
	ProviderType   StorageProviderType `json:"provider_type"`
	Region         string              `json:"region"`
	IsPublic       bool                `json:"is_public"`
	Versioning     bool                `json:"versioning"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

type ObjectItem struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ETag         string            `json:"etag"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	StorageClass string            `json:"storage_class,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type ObjectContent struct {
	Body          io.ReadCloser
	ContentLength int64
	ContentType   string
	ETag          string
	LastModified  time.Time
}

type UploadObjectInput struct {
	BucketName   string
	Key          string
	Body         io.Reader
	Size         int64
	ContentType  string
	Metadata     map[string]string
	StorageClass string
}

type SignedURLOperation string

const (
	SignedURLOpDownload SignedURLOperation = "download"
	SignedURLOpUpload   SignedURLOperation = "upload"
)

type ObjectStorageAdapter interface {
	CreateBucket(ctx context.Context, bucketName, region string) error

	ListBuckets(ctx context.Context) ([]Bucket, error)

	DeleteBucket(ctx context.Context, bucketName string) error

	BucketExists(ctx context.Context, bucketName string) (bool, error)

	ListObjects(ctx context.Context, bucketName, prefix, delimiter string, maxKeys int32) ([]ObjectItem, []string, error)

	UploadObject(ctx context.Context, input UploadObjectInput) (*ObjectItem, error)

	DownloadObject(ctx context.Context, bucketName, key string) (*ObjectContent, error)

	DeleteObject(ctx context.Context, bucketName, key string) error

	DeleteObjects(ctx context.Context, bucketName string, keys []string) error

	GetObjectMetadata(ctx context.Context, bucketName, key string) (*ObjectItem, error)

	GenerateSignedURL(ctx context.Context, bucketName, key string, operation SignedURLOperation, expiry time.Duration) (string, error)
}

type StorageFactory interface {
	GetAdapter(providerType StorageProviderType) (ObjectStorageAdapter, error)

	RegisterAdapter(providerType StorageProviderType, adapter ObjectStorageAdapter)
}

type BucketRepository interface {
	Create(ctx context.Context, bucket *Bucket) error

	GetByID(ctx context.Context, id uuid.UUID) (*Bucket, error)

	GetByName(ctx context.Context, name string) (*Bucket, error)

	ListByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Bucket, int, error)

	Delete(ctx context.Context, id uuid.UUID) error

	CountByOrgID(ctx context.Context, orgID uuid.UUID) (int, error)
}
