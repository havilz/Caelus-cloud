package custom

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type CustomDriver struct {
	mu      sync.RWMutex
	servers map[string]*domain.ProviderServer
}

func NewCustomDriver() domain.ProviderDriver {
	return &CustomDriver{
		servers: make(map[string]*domain.ProviderServer),
	}
}

func (d *CustomDriver) CreateServer(ctx context.Context, cred *domain.Credential, req domain.CreateServerRequest) (*domain.ProviderServer, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	externalID := fmt.Sprintf("byos-%s", uuid.New().String()[:8])
	server := &domain.ProviderServer{
		ExternalID: externalID,
		Name:       req.Name,
		Status:     domain.ServerStatusPending,
		PublicIP:   "0.0.0.0",
		PrivateIP:  "127.0.0.1",
		Region:     req.Region,
		CPUCores:   req.CPUCores,
		MemoryMB:   req.MemoryMB,
		DiskGB:     req.DiskGB,
		CreatedAt:  time.Now().UTC(),
	}

	d.servers[externalID] = server
	return server, nil
}

func (d *CustomDriver) GetServer(ctx context.Context, cred *domain.Credential, externalID string) (*domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	server, exists := d.servers[externalID]
	if !exists {
		return &domain.ProviderServer{
			ExternalID: externalID,
			Name:       "BYOS Server",
			Status:     domain.ServerStatusRunning,
		}, nil
	}
	return server, nil
}

func (d *CustomDriver) ListServers(ctx context.Context, cred *domain.Credential) ([]domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]domain.ProviderServer, 0, len(d.servers))
	for _, s := range d.servers {
		list = append(list, *s)
	}
	return list, nil
}

func (d *CustomDriver) StartServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

func (d *CustomDriver) ShutdownServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusStopped
	}
	return nil
}

func (d *CustomDriver) RebootServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

func (d *CustomDriver) ResizeServer(ctx context.Context, cred *domain.Credential, req domain.ResizeServerRequest) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.servers[req.ExternalID]; ok {
		s.CPUCores = req.CPUCores
		s.MemoryMB = req.MemoryMB
		s.DiskGB = req.DiskGB
	}
	return nil
}

func (d *CustomDriver) DeleteServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.servers, externalID)
	return nil
}
