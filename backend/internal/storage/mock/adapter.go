package mock

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type mockObject struct {
	Key          string
	Data         []byte
	Size         int64
	ETag         string
	ContentType  string
	LastModified time.Time
	StorageClass string
	Metadata     map[string]string
}

type mockBucket struct {
	Name      string
	Region    string
	CreatedAt time.Time
	Objects   map[string]*mockObject
	mu        sync.RWMutex
}

type MockStorageAdapter struct {
	buckets map[string]*mockBucket
	mu      sync.RWMutex
}

func NewMockStorageAdapter() *MockStorageAdapter {
	return &MockStorageAdapter{
		buckets: make(map[string]*mockBucket),
	}
}

func (m *MockStorageAdapter) CreateBucket(_ context.Context, bucketName, region string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buckets[bucketName]; exists {
		return fmt.Errorf("%w: bucket %s already exists", domain.ErrConflict, bucketName)
	}

	if region == "" {
		region = "us-east-1"
	}

	m.buckets[bucketName] = &mockBucket{
		Name:      bucketName,
		Region:    region,
		CreatedAt: time.Now().UTC(),
		Objects:   make(map[string]*mockObject),
	}

	return nil
}

func (m *MockStorageAdapter) ListBuckets(_ context.Context) ([]domain.Bucket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]domain.Bucket, 0, len(m.buckets))
	for _, b := range m.buckets {
		result = append(result, domain.Bucket{
			ID:           uuid.New(),
			Name:         b.Name,
			ProviderType: domain.StorageProviderMock,
			Region:       b.Region,
			CreatedAt:    b.CreatedAt,
			UpdatedAt:    b.CreatedAt,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (m *MockStorageAdapter) DeleteBucket(_ context.Context, bucketName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, exists := m.buckets[bucketName]
	if !exists {
		return fmt.Errorf("%w: bucket %s not found", domain.ErrNotFound, bucketName)
	}

	b.mu.RLock()
	objCount := len(b.Objects)
	b.mu.RUnlock()

	if objCount > 0 {
		return fmt.Errorf("%w: bucket %s is not empty", domain.ErrConflict, bucketName)
	}

	delete(m.buckets, bucketName)
	return nil
}

func (m *MockStorageAdapter) BucketExists(_ context.Context, bucketName string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.buckets[bucketName]
	return exists, nil
}

func (m *MockStorageAdapter) ListObjects(_ context.Context, bucketName, prefix, delimiter string, maxKeys int32) ([]domain.ObjectItem, []string, error) {
	m.mu.RLock()
	b, exists := m.buckets[bucketName]
	m.mu.RUnlock()

	if !exists {
		return nil, nil, fmt.Errorf("%w: bucket %s not found", domain.ErrNotFound, bucketName)
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	if maxKeys <= 0 || maxKeys > 1000 {
		maxKeys = 1000
	}

	objects := make([]domain.ObjectItem, 0)
	folderSet := make(map[string]bool)

	for key, obj := range b.Objects {
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			continue
		}

		if delimiter != "" {
			relativeKey := strings.TrimPrefix(key, prefix)
			if idx := strings.Index(relativeKey, delimiter); idx >= 0 {
				folder := prefix + relativeKey[:idx+len(delimiter)]
				folderSet[folder] = true
				continue
			}
		}

		objects = append(objects, domain.ObjectItem{
			Key:          obj.Key,
			Size:         obj.Size,
			ETag:         obj.ETag,
			ContentType:  obj.ContentType,
			LastModified: obj.LastModified,
			StorageClass: obj.StorageClass,
			Metadata:     obj.Metadata,
		})

		if int32(len(objects)) >= maxKeys {
			break
		}
	}

	folders := make([]string, 0, len(folderSet))
	for f := range folderSet {
		folders = append(folders, f)
	}
	sort.Strings(folders)
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Key < objects[j].Key
	})

	return objects, folders, nil
}

func (m *MockStorageAdapter) UploadObject(_ context.Context, input domain.UploadObjectInput) (*domain.ObjectItem, error) {
	m.mu.RLock()
	b, exists := m.buckets[input.BucketName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: bucket %s not found", domain.ErrNotFound, input.BucketName)
	}

	data, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	hash := md5.Sum(data)
	etag := hex.EncodeToString(hash[:])

	contentType := input.ContentType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(input.Key))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	obj := &mockObject{
		Key:          input.Key,
		Data:         data,
		Size:         int64(len(data)),
		ETag:         etag,
		ContentType:  contentType,
		LastModified: time.Now().UTC(),
		StorageClass: input.StorageClass,
		Metadata:     input.Metadata,
	}

	b.mu.Lock()
	b.Objects[input.Key] = obj
	b.mu.Unlock()

	return &domain.ObjectItem{
		Key:          obj.Key,
		Size:         obj.Size,
		ETag:         obj.ETag,
		ContentType:  obj.ContentType,
		LastModified: obj.LastModified,
		StorageClass: obj.StorageClass,
		Metadata:     obj.Metadata,
	}, nil
}

func (m *MockStorageAdapter) DownloadObject(_ context.Context, bucketName, key string) (*domain.ObjectContent, error) {
	m.mu.RLock()
	b, exists := m.buckets[bucketName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: bucket %s not found", domain.ErrNotFound, bucketName)
	}

	b.mu.RLock()
	obj, exists := b.Objects[key]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: object %s not found in bucket %s", domain.ErrNotFound, key, bucketName)
	}

	return &domain.ObjectContent{
		Body:          io.NopCloser(bytes.NewReader(obj.Data)),
		ContentLength: obj.Size,
		ContentType:   obj.ContentType,
		ETag:          obj.ETag,
		LastModified:  obj.LastModified,
	}, nil
}

func (m *MockStorageAdapter) DeleteObject(_ context.Context, bucketName, key string) error {
	m.mu.RLock()
	b, exists := m.buckets[bucketName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: bucket %s not found", domain.ErrNotFound, bucketName)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.Objects, key)
	return nil
}

func (m *MockStorageAdapter) DeleteObjects(_ context.Context, bucketName string, keys []string) error {
	m.mu.RLock()
	b, exists := m.buckets[bucketName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: bucket %s not found", domain.ErrNotFound, bucketName)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, k := range keys {
		delete(b.Objects, k)
	}
	return nil
}

func (m *MockStorageAdapter) GetObjectMetadata(_ context.Context, bucketName, key string) (*domain.ObjectItem, error) {
	m.mu.RLock()
	b, exists := m.buckets[bucketName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: bucket %s not found", domain.ErrNotFound, bucketName)
	}

	b.mu.RLock()
	obj, exists := b.Objects[key]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: object %s not found in bucket %s", domain.ErrNotFound, key, bucketName)
	}

	return &domain.ObjectItem{
		Key:          obj.Key,
		Size:         obj.Size,
		ETag:         obj.ETag,
		ContentType:  obj.ContentType,
		LastModified: obj.LastModified,
		StorageClass: obj.StorageClass,
		Metadata:     obj.Metadata,
	}, nil
}

func (m *MockStorageAdapter) GenerateSignedURL(_ context.Context, bucketName, key string, operation domain.SignedURLOperation, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}
	token := uuid.New().String()
	return fmt.Sprintf("https://storage.caelus.local/%s/%s?op=%s&expires=%d&token=%s", bucketName, key, operation, time.Now().Add(expiry).Unix(), token), nil
}
