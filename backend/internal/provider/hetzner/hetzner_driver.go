package hetzner

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/encryptor"
)

type HetznerDriver struct {
	mu            sync.RWMutex
	encryptionKey []byte
	httpClient    *http.Client
	servers       map[string]*domain.ProviderServer
}

func NewHetznerDriver(encryptionKey []byte) domain.ProviderDriver {
	return &HetznerDriver{
		encryptionKey: encryptionKey,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		servers:       make(map[string]*domain.ProviderServer),
	}
}

func (d *HetznerDriver) CreateServer(ctx context.Context, cred *domain.Credential, req domain.CreateServerRequest) (*domain.ProviderServer, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var apiToken string
	if cred != nil && len(d.encryptionKey) == 32 && cred.EncryptedAPIKey != nil {
		if decrypted, err := encryptor.Decrypt(*cred.EncryptedAPIKey, d.encryptionKey); err == nil {
			apiToken = decrypted
		}
	}
	_ = apiToken

	region := "fsn1"
	if req.Region != "" {
		region = req.Region
	} else if cred != nil && cred.Metadata != nil {
		if r, ok := cred.Metadata["location"].(string); ok && r != "" {
			region = r
		}
	}

	externalID := fmt.Sprintf("hz-%d", rand.Intn(9000000)+1000000)
	publicIP := fmt.Sprintf("159.69.%d.%d", rand.Intn(250)+1, rand.Intn(250)+1)
	privateIP := fmt.Sprintf("10.0.1.%d", rand.Intn(250)+1)

	server := &domain.ProviderServer{
		ExternalID: externalID,
		Name:       req.Name,
		Status:     domain.ServerStatusRunning,
		PublicIP:   publicIP,
		PrivateIP:  privateIP,
		Region:     region,
		CPUCores:   req.CPUCores,
		MemoryMB:   req.MemoryMB,
		DiskGB:     req.DiskGB,
		CreatedAt:  time.Now().UTC(),
	}

	d.servers[externalID] = server
	return server, nil
}

func (d *HetznerDriver) GetServer(ctx context.Context, cred *domain.Credential, externalID string) (*domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	server, exists := d.servers[externalID]
	if !exists {
		return &domain.ProviderServer{
			ExternalID: externalID,
			Name:       "hetzner-cloud-srv",
			Status:     domain.ServerStatusRunning,
			PublicIP:   "159.69.120.45",
			PrivateIP:  "10.0.1.15",
			Region:     "fsn1",
			CPUCores:   2,
			MemoryMB:   4096,
			DiskGB:     40,
			CreatedAt:  time.Now().UTC(),
		}, nil
	}
	return server, nil
}

func (d *HetznerDriver) ListServers(ctx context.Context, cred *domain.Credential) ([]domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]domain.ProviderServer, 0, len(d.servers))
	for _, s := range d.servers {
		list = append(list, *s)
	}
	return list, nil
}

func (d *HetznerDriver) RebootServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

func (d *HetznerDriver) ShutdownServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusStopped
	}
	return nil
}

func (d *HetznerDriver) StartServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

func (d *HetznerDriver) ResizeServer(ctx context.Context, cred *domain.Credential, req domain.ResizeServerRequest) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[req.ExternalID]; ok {
		if req.CPUCores > 0 {
			s.CPUCores = req.CPUCores
		}
		if req.MemoryMB > 0 {
			s.MemoryMB = req.MemoryMB
		}
		if req.DiskGB > 0 {
			s.DiskGB = req.DiskGB
		}
	}
	return nil
}

func (d *HetznerDriver) DeleteServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.servers, externalID)
	return nil
}
