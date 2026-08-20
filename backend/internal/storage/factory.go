package storage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/storage/mock"
)

type storageFactory struct {
	mu       sync.RWMutex
	adapters map[domain.StorageProviderType]domain.ObjectStorageAdapter
}

// NewStorageFactory menginisialisasi factory penyimpanan multi-provider dengan adapter bawaan MockStorageAdapter.
func NewStorageFactory() domain.StorageFactory {
	f := &storageFactory{
		adapters: make(map[domain.StorageProviderType]domain.ObjectStorageAdapter),
	}
	f.RegisterAdapter(domain.StorageProviderMock, mock.NewMockStorageAdapter())
	return f
}

// GetAdapter mengambil implementasi ObjectStorageAdapter berdasarkan tipe penyedia storage.
func (f *storageFactory) GetAdapter(providerType domain.StorageProviderType) (domain.ObjectStorageAdapter, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	normalized := domain.StorageProviderType(strings.ToLower(strings.TrimSpace(string(providerType))))
	adapter, exists := f.adapters[normalized]
	if !exists {
		return nil, fmt.Errorf("%w: storage adapter for provider %s not registered", domain.ErrNotFound, providerType)
	}

	return adapter, nil
}

// RegisterAdapter mendaftarkan implementasi ObjectStorageAdapter baru ke dalam factory.
func (f *storageFactory) RegisterAdapter(providerType domain.StorageProviderType, adapter domain.ObjectStorageAdapter) {
	f.mu.Lock()
	defer f.mu.Unlock()

	normalized := domain.StorageProviderType(strings.ToLower(strings.TrimSpace(string(providerType))))
	f.adapters[normalized] = adapter
}
