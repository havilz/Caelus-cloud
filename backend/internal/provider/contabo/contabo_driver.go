package contabo

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

type ContaboDriver struct {
	mu            sync.RWMutex
	encryptionKey []byte
	httpClient    *http.Client
	servers       map[string]*domain.ProviderServer
}

func NewContaboDriver(encryptionKey []byte) domain.ProviderDriver {
	return &ContaboDriver{
		encryptionKey: encryptionKey,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		servers:       make(map[string]*domain.ProviderServer),
	}
}

func (d *ContaboDriver) CreateServer(ctx context.Context, cred *domain.Credential, req domain.CreateServerRequest) (*domain.ProviderServer, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var clientID, clientSecret string
	if cred != nil && len(d.encryptionKey) == 32 {
		if cred.EncryptedAPIKey != nil {
			if decrypted, err := encryptor.Decrypt(*cred.EncryptedAPIKey, d.encryptionKey); err == nil {
				clientID = decrypted
			}
		}
		if cred.EncryptedAPISecret != nil {
			if decrypted, err := encryptor.Decrypt(*cred.EncryptedAPISecret, d.encryptionKey); err == nil {
				clientSecret = decrypted
			}
		}
	}
	_ = clientID
	_ = clientSecret

	region := "EU"
	if req.Region != "" {
		region = req.Region
	} else if cred != nil && cred.Metadata != nil {
		if r, ok := cred.Metadata["region"].(string); ok && r != "" {
			region = r
		}
	}

	externalID := fmt.Sprintf("ctb-%d", rand.Intn(9000000)+1000000)
	publicIP := fmt.Sprintf("194.163.%d.%d", rand.Intn(250)+1, rand.Intn(250)+1)
	privateIP := fmt.Sprintf("10.0.3.%d", rand.Intn(250)+1)

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

func (d *ContaboDriver) GetServer(ctx context.Context, cred *domain.Credential, externalID string) (*domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	server, exists := d.servers[externalID]
	if !exists {
		return &domain.ProviderServer{
			ExternalID: externalID,
			Name:       "contabo-vps",
			Status:     domain.ServerStatusRunning,
			PublicIP:   "194.163.150.88",
			PrivateIP:  "10.0.3.12",
			Region:     "EU",
			CPUCores:   4,
			MemoryMB:   8192,
			DiskGB:     200,
			CreatedAt:  time.Now().UTC(),
		}, nil
	}
	return server, nil
}

func (d *ContaboDriver) ListServers(ctx context.Context, cred *domain.Credential) ([]domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]domain.ProviderServer, 0, len(d.servers))
	for _, s := range d.servers {
		list = append(list, *s)
	}
	return list, nil
}

func (d *ContaboDriver) RebootServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

func (d *ContaboDriver) ShutdownServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusStopped
	}
	return nil
}

func (d *ContaboDriver) StartServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

func (d *ContaboDriver) ResizeServer(ctx context.Context, cred *domain.Credential, req domain.ResizeServerRequest) error {
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

func (d *ContaboDriver) DeleteServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.servers, externalID)
	return nil
}
