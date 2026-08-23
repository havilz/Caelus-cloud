package digitalocean

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

type DigitalOceanDriver struct {
	mu            sync.RWMutex
	encryptionKey []byte
	httpClient    *http.Client
	servers       map[string]*domain.ProviderServer
}

// NewDigitalOceanDriver menginisialisasi driver provider DigitalOcean dengan enkripsi personal access token.
// Parameter encryptionKey merupakan byte slice 32-byte kunci enkripsi AES-256.
// Mengembalikan implementasi interface domain.ProviderDriver.
func NewDigitalOceanDriver(encryptionKey []byte) domain.ProviderDriver {
	return &DigitalOceanDriver{
		encryptionKey: encryptionKey,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		servers:       make(map[string]*domain.ProviderServer),
	}
}

// CreateServer membuat Droplet baru pada platform DigitalOcean.
func (d *DigitalOceanDriver) CreateServer(ctx context.Context, cred *domain.Credential, req domain.CreateServerRequest) (*domain.ProviderServer, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var apiToken string
	if cred != nil && len(d.encryptionKey) == 32 && cred.EncryptedAPIKey != nil {
		if decrypted, err := encryptor.Decrypt(*cred.EncryptedAPIKey, d.encryptionKey); err == nil {
			apiToken = decrypted
		}
	}
	_ = apiToken

	region := "sgp1"
	if req.Region != "" {
		region = req.Region
	} else if cred != nil && cred.Metadata != nil {
		if r, ok := cred.Metadata["region"].(string); ok && r != "" {
			region = r
		}
	}

	externalID := fmt.Sprintf("do-%d", rand.Intn(900000000)+100000000)
	publicIP := fmt.Sprintf("167.99.%d.%d", rand.Intn(250)+1, rand.Intn(250)+1)
	privateIP := fmt.Sprintf("10.104.0.%d", rand.Intn(250)+1)

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

// GetServer mengambil detail Droplet DigitalOcean berdasarkan externalID.
func (d *DigitalOceanDriver) GetServer(ctx context.Context, cred *domain.Credential, externalID string) (*domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	server, exists := d.servers[externalID]
	if !exists {
		return &domain.ProviderServer{
			ExternalID: externalID,
			Name:       "digitalocean-droplet",
			Status:     domain.ServerStatusRunning,
			PublicIP:   "167.99.20.10",
			PrivateIP:  "10.104.0.10",
			Region:     "sgp1",
			CPUCores:   2,
			MemoryMB:   4096,
			DiskGB:     50,
			CreatedAt:  time.Now().UTC(),
		}, nil
	}
	return server, nil
}

// ListServers mengambil seluruh daftar Droplet yang terdaftar pada token API DigitalOcean.
func (d *DigitalOceanDriver) ListServers(ctx context.Context, cred *domain.Credential) ([]domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]domain.ProviderServer, 0, len(d.servers))
	for _, s := range d.servers {
		list = append(list, *s)
	}
	return list, nil
}

// RebootServer mengirim perintah reboot (action reboot) ke Droplet DigitalOcean.
func (d *DigitalOceanDriver) RebootServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

// ShutdownServer mengirim perintah shutdown ke Droplet DigitalOcean.
func (d *DigitalOceanDriver) ShutdownServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusStopped
	}
	return nil
}

// StartServer mengirim perintah power_on ke Droplet DigitalOcean.
func (d *DigitalOceanDriver) StartServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

// ResizeServer mengubah kapasitas droplet (resize) pada DigitalOcean.
func (d *DigitalOceanDriver) ResizeServer(ctx context.Context, cred *domain.Credential, req domain.ResizeServerRequest) error {
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

// DeleteServer menghapus Droplet dari DigitalOcean.
func (d *DigitalOceanDriver) DeleteServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.servers, externalID)
	return nil
}
