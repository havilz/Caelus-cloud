package mock

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type MockDriver struct {
	mu      sync.RWMutex
	servers map[string]*domain.ProviderServer
}

func NewMockDriver() domain.ProviderDriver {
	return &MockDriver{
		servers: make(map[string]*domain.ProviderServer),
	}
}

func (d *MockDriver) CreateServer(ctx context.Context, cred *domain.Credential, req domain.CreateServerRequest) (*domain.ProviderServer, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	externalID := fmt.Sprintf("mock-srv-%s", uuid.New().String()[:8])
	publicIP := fmt.Sprintf("198.51.100.%d", rand.Intn(250)+2)
	privateIP := fmt.Sprintf("10.0.0.%d", rand.Intn(250)+2)

	server := &domain.ProviderServer{
		ExternalID: externalID,
		Name:       req.Name,
		Status:     domain.ServerStatusRunning,
		PublicIP:   publicIP,
		PrivateIP:  privateIP,
		Region:     req.Region,
		CPUCores:   req.CPUCores,
		MemoryMB:   req.MemoryMB,
		DiskGB:     req.DiskGB,
		CreatedAt:  time.Now(),
	}

	d.servers[externalID] = server
	return server, nil
}

func (d *MockDriver) GetServer(ctx context.Context, cred *domain.Credential, externalID string) (*domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	server, exists := d.servers[externalID]
	if !exists {
		return nil, domain.ErrNotFound
	}

	return server, nil
}

func (d *MockDriver) ListServers(ctx context.Context, cred *domain.Credential) ([]domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]domain.ProviderServer, 0, len(d.servers))
	for _, s := range d.servers {
		list = append(list, *s)
	}

	return list, nil
}

func (d *MockDriver) RebootServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	return d.updateServerStatus(externalID, domain.ServerStatusRunning)
}

func (d *MockDriver) ShutdownServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	return d.updateServerStatus(externalID, domain.ServerStatusStopped)
}

func (d *MockDriver) StartServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	return d.updateServerStatus(externalID, domain.ServerStatusRunning)
}

func (d *MockDriver) ResizeServer(ctx context.Context, cred *domain.Credential, req domain.ResizeServerRequest) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	server, exists := d.servers[req.ExternalID]
	if !exists {
		return domain.ErrNotFound
	}

	if req.CPUCores > 0 {
		server.CPUCores = req.CPUCores
	}
	if req.MemoryMB > 0 {
		server.MemoryMB = req.MemoryMB
	}
	if req.DiskGB > 0 {
		server.DiskGB = req.DiskGB
	}

	return nil
}

func (d *MockDriver) DeleteServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.servers[externalID]; !exists {
		return domain.ErrNotFound
	}

	delete(d.servers, externalID)
	return nil
}

func (d *MockDriver) updateServerStatus(externalID string, targetStatus domain.ServerStatus) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	server, exists := d.servers[externalID]
	if !exists {
		return domain.ErrNotFound
	}

	server.Status = targetStatus
	return nil
}
