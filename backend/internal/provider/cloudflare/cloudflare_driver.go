package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/encryptor"
)

type CloudflareDriver struct {
	mu            sync.RWMutex
	encryptionKey []byte
	httpClient    *http.Client
	servers       map[string]*domain.ProviderServer
}

func NewCloudflareDriver(encryptionKey []byte) domain.ProviderDriver {
	return &CloudflareDriver{
		encryptionKey: encryptionKey,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		servers:       make(map[string]*domain.ProviderServer),
	}
}

func (d *CloudflareDriver) CreateServer(ctx context.Context, cred *domain.Credential, req domain.CreateServerRequest) (*domain.ProviderServer, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	region := "auto"
	if req.Region != "" {
		region = req.Region
	} else if cred != nil && cred.Metadata != nil {
		if r, ok := cred.Metadata["region"].(string); ok && r != "" {
			region = r
		}
	}

	publicIP := "edge.cloudflare.com"
	if ip, ok := req.Tags["public_ip"]; ok && ip != "" {
		publicIP = ip
	} else if hostname, ok := req.Tags["tunnel_hostname"]; ok && hostname != "" {
		publicIP = hostname
	}

	privateIP := "127.0.0.1"
	if pip, ok := req.Tags["private_ip"]; ok && pip != "" {
		privateIP = pip
	}

	externalID := fmt.Sprintf("cf-%s", uuid.New().String()[:12])
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

func (d *CloudflareDriver) GetServer(ctx context.Context, cred *domain.Credential, externalID string) (*domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	server, exists := d.servers[externalID]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return server, nil
}

func (d *CloudflareDriver) ListServers(ctx context.Context, cred *domain.Credential) ([]domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var result []domain.ProviderServer
	for _, s := range d.servers {
		result = append(result, *s)
	}
	return result, nil
}

func (d *CloudflareDriver) RebootServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	return nil
}

func (d *CloudflareDriver) ShutdownServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusStopped
	}
	return nil
}

func (d *CloudflareDriver) StartServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

func (d *CloudflareDriver) ResizeServer(ctx context.Context, cred *domain.Credential, req domain.ResizeServerRequest) error {
	return nil
}

func (d *CloudflareDriver) DeleteServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.servers, externalID)
	return nil
}

func (d *CloudflareDriver) ValidateCredential(ctx context.Context, cred *domain.Credential) error {
	if cred == nil {
		return fmt.Errorf("credential cannot be nil")
	}

	var apiKey, apiSecret string
	if len(d.encryptionKey) == 32 {
		if cred.EncryptedAPIKey != nil {
			if decrypted, err := encryptor.Decrypt(*cred.EncryptedAPIKey, d.encryptionKey); err == nil {
				apiKey = decrypted
			}
		}
		if cred.EncryptedAPISecret != nil {
			if decrypted, err := encryptor.Decrypt(*cred.EncryptedAPISecret, d.encryptionKey); err == nil {
				apiSecret = decrypted
			}
		}
	}

	if apiKey == "" && apiSecret == "" {
		return fmt.Errorf("API key/token or secret cannot be empty")
	}

	return nil
}
