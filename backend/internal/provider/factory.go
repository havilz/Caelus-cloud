package provider

import (
	"strings"
	"sync"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/provider/aws"
	"github.com/havilz/caelus-cloud/backend/internal/provider/cloudflare"
	"github.com/havilz/caelus-cloud/backend/internal/provider/contabo"
	"github.com/havilz/caelus-cloud/backend/internal/provider/custom"
	"github.com/havilz/caelus-cloud/backend/internal/provider/digitalocean"
	"github.com/havilz/caelus-cloud/backend/internal/provider/hetzner"
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

func NewDriverFactory() Factory {
	return NewDriverFactoryWithKey(nil)
}

func NewDriverFactoryWithKey(encryptionKey []byte) Factory {
	f := &driverFactory{
		drivers: make(map[string]domain.ProviderDriver),
	}
	f.RegisterDriver("mock", mock.NewMockDriver())
	f.RegisterDriver("custom", custom.NewCustomDriver())
	f.RegisterDriver("aws", aws.NewEC2Driver(encryptionKey))
	f.RegisterDriver("cloudflare", cloudflare.NewCloudflareDriver(encryptionKey))
	f.RegisterDriver("hetzner", hetzner.NewHetznerDriver(encryptionKey))
	f.RegisterDriver("digitalocean", digitalocean.NewDigitalOceanDriver(encryptionKey))
	f.RegisterDriver("contabo", contabo.NewContaboDriver(encryptionKey))
	return f
}

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

func (f *driverFactory) RegisterDriver(providerSlug string, driver domain.ProviderDriver) {
	f.mu.Lock()
	defer f.mu.Unlock()

	slug := strings.ToLower(strings.TrimSpace(providerSlug))
	f.drivers[slug] = driver
}
