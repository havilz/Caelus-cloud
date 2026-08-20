package provider

import (
	"strings"
	"sync"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/provider/mock"
)

type Factory interface {
	GetDriver(providerSlug string) (domain.ProviderDriver, error)
	RegisterDriver(providerSlug string, driver domain.ProviderDriver)
}

type driverFactory struct {
	mu      sync.RWMutex
	drivers map[string]domain.ProviderDriver
}

// NewDriverFactory menginisialisasi factory registri driver provider cloud dengan driver bawaan MockDriver.
// Mengembalikan implementasi interface Factory.
func NewDriverFactory() Factory {
	f := &driverFactory{
		drivers: make(map[string]domain.ProviderDriver),
	}
	f.RegisterDriver("mock", mock.NewMockDriver())
	return f
}

// GetDriver mengambil instance ProviderDriver berdasarkan slug unik penyedia cloud.
// Parameter providerSlug merupakan identifier slug provider (misal: "mock", "aws", "hetzner").
// Mengembalikan domain.ProviderDriver atau domain.ErrNotFound jika driver untuk provider tersebut belum terdaftar.
func (f *driverFactory) GetDriver(providerSlug string) (domain.ProviderDriver, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	slug := strings.ToLower(strings.TrimSpace(providerSlug))
	driver, exists := f.drivers[slug]
	if !exists {
		return nil, domain.ErrNotFound
	}

	return driver, nil
}

// RegisterDriver mendaftarkan implementasi ProviderDriver baru ke dalam factory secara dinamis.
// Parameter providerSlug merupakan identifier slug unik provider.
// Parameter driver merupakan implementasi domain.ProviderDriver.
func (f *driverFactory) RegisterDriver(providerSlug string, driver domain.ProviderDriver) {
	f.mu.Lock()
	defer f.mu.Unlock()

	slug := strings.ToLower(strings.TrimSpace(providerSlug))
	f.drivers[slug] = driver
}
