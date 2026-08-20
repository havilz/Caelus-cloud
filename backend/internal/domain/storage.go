package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

// StorageProviderType mendefinisikan tipe penyedia layanan Object Storage yang didukung.
type StorageProviderType string

const (
	StorageProviderMinIO StorageProviderType = "minio"
	StorageProviderS3    StorageProviderType = "s3"
	StorageProviderR2    StorageProviderType = "r2"
	StorageProviderMock  StorageProviderType = "mock"
)

// Bucket merepresentasikan entitas bucket penyimpanan objek dalam multi-tenant Caelus Cloud.
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

// ObjectItem merepresentasikan metadata dari sebuah objek/file di dalam bucket.
type ObjectItem struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ETag         string            `json:"etag"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	StorageClass string            `json:"storage_class,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ObjectContent membungkus pembacaan stream data objek beserta metadata responnya.
type ObjectContent struct {
	Body          io.ReadCloser
	ContentLength int64
	ContentType   string
	ETag          string
	LastModified  time.Time
}

// UploadObjectInput mendefinisikan parameter yang dibutuhkan untuk mengunggah objek ke bucket.
type UploadObjectInput struct {
	BucketName   string
	Key          string
	Body         io.Reader
	Size         int64
	ContentType  string
	Metadata     map[string]string
	StorageClass string
}

// SignedURLOperation menentukan jenis operasi presigned URL yang diizinkan (download/upload).
type SignedURLOperation string

const (
	SignedURLOpDownload SignedURLOperation = "download" // HTTP GET
	SignedURLOpUpload   SignedURLOperation = "upload"   // HTTP PUT
)

// ObjectStorageAdapter mendefinisikan kontrak interface operasi multi-provider Object Storage.
type ObjectStorageAdapter interface {
	// CreateBucket membuat bucket baru pada penyedia storage dengan nama dan region yang ditentukan.
	CreateBucket(ctx context.Context, bucketName, region string) error

	// ListBuckets mengambil daftar seluruh bucket yang tersedia di bawah akun/kredensial penyedia.
	ListBuckets(ctx context.Context) ([]Bucket, error)

	// DeleteBucket menghapus bucket berdasarkan nama (bucket harus dalam keadaan kosong).
	DeleteBucket(ctx context.Context, bucketName string) error

	// BucketExists memeriksa apakah bucket dengan nama tertentu sudah ada pada penyedia storage.
	BucketExists(ctx context.Context, bucketName string) (bool, error)

	// ListObjects mengambil daftar objek di dalam bucket dengan dukungan prefix, delimiter (folder), dan limit jumlah.
	ListObjects(ctx context.Context, bucketName, prefix, delimiter string, maxKeys int32) ([]ObjectItem, []string, error)

	// UploadObject mengunggah stream objek ke dalam bucket.
	UploadObject(ctx context.Context, input UploadObjectInput) (*ObjectItem, error)

	// DownloadObject mengambil stream konten objek beserta metadatanya dari bucket.
	DownloadObject(ctx context.Context, bucketName, key string) (*ObjectContent, error)

	// DeleteObject menghapus sebuah objek berdasarkan kunci spesifik dari bucket.
	DeleteObject(ctx context.Context, bucketName, key string) error

	// DeleteObjects menghapus beberapa objek sekaligus dalam satu operasi batch.
	DeleteObjects(ctx context.Context, bucketName string, keys []string) error

	// GetObjectMetadata mengambil metadata dari sebuah objek tanpa mengunduh konten body-nya.
	GetObjectMetadata(ctx context.Context, bucketName, key string) (*ObjectItem, error)

	// GenerateSignedURL membuat URL bertanda tangan (Presigned URL) dengan masa kedaluwarsa tertentu.
	GenerateSignedURL(ctx context.Context, bucketName, key string, operation SignedURLOperation, expiry time.Duration) (string, error)
}

// StorageFactory mendefinisikan kontrak registri dan penyedia adapter storage multi-provider.
type StorageFactory interface {
	// GetAdapter mengambil implementasi ObjectStorageAdapter berdasarkan tipe penyedia storage.
	GetAdapter(providerType StorageProviderType) (ObjectStorageAdapter, error)

	// RegisterAdapter mendaftarkan implementasi ObjectStorageAdapter baru ke dalam factory.
	RegisterAdapter(providerType StorageProviderType, adapter ObjectStorageAdapter)
}

// BucketRepository mendefinisikan kontrak persistensi metadata bucket ke basis data.
type BucketRepository interface {
	// Create menyimpan rekaman metadata bucket baru ke basis data.
	Create(ctx context.Context, bucket *Bucket) error

	// GetByID mengambil data bucket berdasarkan ID unik.
	GetByID(ctx context.Context, id uuid.UUID) (*Bucket, error)

	// GetByName mengambil data bucket berdasarkan nama uniknya.
	GetByName(ctx context.Context, name string) (*Bucket, error)

	// ListByOrgID mengambil daftar seluruh bucket milik organisasi tertentu dengan paginasi.
	ListByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Bucket, int, error)

	// Delete menghapus rekaman metadata bucket dari basis data.
	Delete(ctx context.Context, id uuid.UUID) error

	// CountByOrgID menghitung total jumlah bucket milik organisasi.
	CountByOrgID(ctx context.Context, orgID uuid.UUID) (int, error)
}
