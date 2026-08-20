package tests

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/storage"
	"github.com/havilz/caelus-cloud/backend/internal/storage/minio"
	"github.com/havilz/caelus-cloud/backend/internal/storage/mock"
	"github.com/havilz/caelus-cloud/backend/internal/storage/r2"
	"github.com/havilz/caelus-cloud/backend/internal/storage/s3"
)

func TestStorageFactory_RegisterAndGetAdapter(t *testing.T) {
	factory := storage.NewStorageFactory()

	// Default Mock Adapter harus langsung tersedia
	mockAdapter, err := factory.GetAdapter(domain.StorageProviderMock)
	if err != nil {
		t.Fatalf("expected mock adapter to be registered, got error: %v", err)
	}
	if mockAdapter == nil {
		t.Fatal("expected non-nil mock adapter")
	}

	// Adapter yang belum didaftarkan harus menghasilkan error NotFound
	_, err = factory.GetAdapter(domain.StorageProviderS3)
	if err == nil {
		t.Fatal("expected error for unregistered s3 adapter, got nil")
	}

	// Daftarkan custom adapter
	customMock := mock.NewMockStorageAdapter()
	factory.RegisterAdapter(domain.StorageProviderS3, customMock)

	retrieved, err := factory.GetAdapter(domain.StorageProviderS3)
	if err != nil {
		t.Fatalf("failed to retrieve registered adapter: %v", err)
	}
	if retrieved != customMock {
		t.Fatal("retrieved adapter does not match registered instance")
	}
}

func TestStorageAdapter_BucketLifecycle(t *testing.T) {
	ctx := context.Background()
	adapter := mock.NewMockStorageAdapter()

	bucketName := "caelus-production-assets"

	// 1. Bucket belum ada
	exists, err := adapter.BucketExists(ctx, bucketName)
	if err != nil {
		t.Fatalf("failed to check bucket exists: %v", err)
	}
	if exists {
		t.Fatal("expected bucket to not exist initially")
	}

	// 2. Buat Bucket
	if err := adapter.CreateBucket(ctx, bucketName, "ap-southeast-1"); err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	// 3. Bucket harus ada
	exists, err = adapter.BucketExists(ctx, bucketName)
	if err != nil {
		t.Fatalf("failed to check bucket exists after creation: %v", err)
	}
	if !exists {
		t.Fatal("expected bucket to exist after creation")
	}

	// 4. Mencegah duplikasi nama bucket
	if err := adapter.CreateBucket(ctx, bucketName, "ap-southeast-1"); err == nil {
		t.Fatal("expected conflict error when creating duplicate bucket, got nil")
	}

	// 5. List Buckets
	buckets, err := adapter.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("failed to list buckets: %v", err)
	}
	if len(buckets) != 1 || buckets[0].Name != bucketName {
		t.Fatalf("expected 1 bucket with name %s, got: %+v", bucketName, buckets)
	}

	// 6. Hapus Bucket kosong
	if err := adapter.DeleteBucket(ctx, bucketName); err != nil {
		t.Fatalf("failed to delete empty bucket: %v", err)
	}

	// 7. Bucket harus sudah hilang
	exists, _ = adapter.BucketExists(ctx, bucketName)
	if exists {
		t.Fatal("expected bucket to be deleted")
	}
}

func TestStorageAdapter_ObjectOperations(t *testing.T) {
	ctx := context.Background()
	adapter := mock.NewMockStorageAdapter()
	bucketName := "my-app-bucket"

	_ = adapter.CreateBucket(ctx, bucketName, "us-east-1")

	testContent := "Hello, Caelus Cloud Object Storage! Multi-cloud infrastructure."
	testKey := "documents/reports/q3_summary.txt"

	// 1. Upload Object
	uploadRes, err := adapter.UploadObject(ctx, domain.UploadObjectInput{
		BucketName:  bucketName,
		Key:         testKey,
		Body:        strings.NewReader(testContent),
		Size:        int64(len(testContent)),
		ContentType: "text/plain",
		Metadata: map[string]string{
			"author": "caelus-agent",
		},
	})
	if err != nil {
		t.Fatalf("failed to upload object: %v", err)
	}
	if uploadRes.Key != testKey || uploadRes.Size != int64(len(testContent)) {
		t.Fatalf("unexpected upload result: %+v", uploadRes)
	}
	if uploadRes.ETag == "" {
		t.Fatal("expected valid ETag for uploaded object")
	}

	// 2. Mencegah penghapusan bucket yang masih berisi objek
	if err := adapter.DeleteBucket(ctx, bucketName); err == nil {
		t.Fatal("expected error when deleting non-empty bucket, got nil")
	}

	// 3. Get Object Metadata (HeadObject)
	meta, err := adapter.GetObjectMetadata(ctx, bucketName, testKey)
	if err != nil {
		t.Fatalf("failed to get object metadata: %v", err)
	}
	if meta.Size != int64(len(testContent)) || meta.ContentType != "text/plain" {
		t.Fatalf("unexpected metadata result: %+v", meta)
	}

	// 4. Download Object
	content, err := adapter.DownloadObject(ctx, bucketName, testKey)
	if err != nil {
		t.Fatalf("failed to download object: %v", err)
	}
	defer content.Body.Close()

	downloadedBytes, err := io.ReadAll(content.Body)
	if err != nil {
		t.Fatalf("failed to read downloaded body: %v", err)
	}
	if string(downloadedBytes) != testContent {
		t.Fatalf("downloaded content mismatch. Expected %q, got %q", testContent, string(downloadedBytes))
	}

	// 5. Upload second object in different folder
	_, _ = adapter.UploadObject(ctx, domain.UploadObjectInput{
		BucketName: bucketName,
		Key:        "images/banner.png",
		Body:       bytes.NewReader([]byte("fake-png-binary")),
		Size:       15,
	})

	// 6. List Objects dengan Delimiter "/" (Virtual Folders)
	objects, folders, err := adapter.ListObjects(ctx, bucketName, "", "/", 100)
	if err != nil {
		t.Fatalf("failed to list objects: %v", err)
	}
	if len(folders) != 2 {
		t.Fatalf("expected 2 root folders ('documents/', 'images/'), got: %+v", folders)
	}
	if len(objects) != 0 {
		t.Fatalf("expected 0 root objects, got: %+v", objects)
	}

	// 7. List Objects di dalam subfolder "documents/"
	subObjects, _, err := adapter.ListObjects(ctx, bucketName, "documents/", "", 100)
	if err != nil {
		t.Fatalf("failed to list subfolder objects: %v", err)
	}
	if len(subObjects) != 1 || subObjects[0].Key != testKey {
		t.Fatalf("expected 1 sub object, got: %+v", subObjects)
	}

	// 8. Generate Signed URLs (Download & Upload)
	downloadURL, err := adapter.GenerateSignedURL(ctx, bucketName, testKey, domain.SignedURLOpDownload, 10*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate presigned download URL: %v", err)
	}
	if !strings.Contains(downloadURL, testKey) || !strings.Contains(downloadURL, "op=download") {
		t.Fatalf("invalid presigned download URL: %s", downloadURL)
	}

	uploadURL, err := adapter.GenerateSignedURL(ctx, bucketName, "new_upload.pdf", domain.SignedURLOpUpload, 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate presigned upload URL: %v", err)
	}
	if !strings.Contains(uploadURL, "op=upload") {
		t.Fatalf("invalid presigned upload URL: %s", uploadURL)
	}

	// 9. Batch Delete Objects
	if err := adapter.DeleteObjects(ctx, bucketName, []string{testKey, "images/banner.png"}); err != nil {
		t.Fatalf("failed to batch delete objects: %v", err)
	}

	// 10. Verifikasi objek sudah kosong
	remainingObjs, _, _ := adapter.ListObjects(ctx, bucketName, "", "", 100)
	if len(remainingObjs) != 0 {
		t.Fatalf("expected 0 remaining objects after batch delete, got %d", len(remainingObjs))
	}

	// 11. Hapus bucket setelah kosong
	if err := adapter.DeleteBucket(ctx, bucketName); err != nil {
		t.Fatalf("failed to delete bucket after cleanup: %v", err)
	}
}

func TestStorageAdapter_ProviderInitialization(t *testing.T) {
	// MinIO Adapter initialization
	minioAdapter, err := minio.NewAdapter(minio.Config{
		Endpoint:        "http://localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
	})
	if err != nil {
		t.Fatalf("failed to initialize MinIO adapter: %v", err)
	}
	if minioAdapter == nil {
		t.Fatal("expected non-nil MinIO adapter")
	}

	// S3 Adapter initialization
	s3Adapter, err := s3.NewAdapter(s3.Config{
		Region:          "ap-southeast-1",
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		ProviderType:    domain.StorageProviderS3,
	})
	if err != nil {
		t.Fatalf("failed to initialize S3 adapter: %v", err)
	}
	if s3Adapter == nil {
		t.Fatal("expected non-nil S3 adapter")
	}

	// Cloudflare R2 Adapter initialization validation
	_, err = r2.NewAdapter(r2.Config{
		AccountID: "", // Account ID kosong harus gagal
	})
	if err == nil {
		t.Fatal("expected validation error for empty Cloudflare account ID, got nil")
	}

	r2Adapter, err := r2.NewAdapter(r2.Config{
		AccountID:       "abc123def456789",
		AccessKeyID:     "r2_access_key",
		SecretAccessKey: "r2_secret_key",
	})
	if err != nil {
		t.Fatalf("failed to initialize Cloudflare R2 adapter: %v", err)
	}
	if r2Adapter == nil {
		t.Fatal("expected non-nil Cloudflare R2 adapter")
	}
}
