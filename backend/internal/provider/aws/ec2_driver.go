package aws

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/encryptor"
)

type EC2Driver struct {
	mu            sync.RWMutex
	encryptionKey []byte
	httpClient    *http.Client
	servers       map[string]*domain.ProviderServer
}

func NewEC2Driver(encryptionKey []byte) domain.ProviderDriver {
	return &EC2Driver{
		encryptionKey: encryptionKey,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		servers:       make(map[string]*domain.ProviderServer),
	}
}

func (d *EC2Driver) CreateServer(ctx context.Context, cred *domain.Credential, req domain.CreateServerRequest) (*domain.ProviderServer, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var accessKey, secretKey string
	if cred != nil && len(d.encryptionKey) == 32 {
		if cred.EncryptedAPIKey != nil {
			if decrypted, err := encryptor.Decrypt(*cred.EncryptedAPIKey, d.encryptionKey); err == nil {
				accessKey = decrypted
			}
		}
		if cred.EncryptedAPISecret != nil {
			if decrypted, err := encryptor.Decrypt(*cred.EncryptedAPISecret, d.encryptionKey); err == nil {
				secretKey = decrypted
			}
		}
	}

	region := "us-east-1"
	if req.Region != "" {
		region = req.Region
	} else if cred != nil && cred.Metadata != nil {
		if r, ok := cred.Metadata["region"].(string); ok && r != "" {
			region = r
		}
	}

	externalID := fmt.Sprintf("i-%s", uuid.New().String()[:17])
	publicIP := fmt.Sprintf("54.%d.%d.%d", rand.Intn(200)+1, rand.Intn(250)+1, rand.Intn(250)+1)
	privateIP := fmt.Sprintf("172.31.%d.%d", rand.Intn(250)+1, rand.Intn(250)+1)

	_ = accessKey
	_ = secretKey

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

func (d *EC2Driver) GetServer(ctx context.Context, cred *domain.Credential, externalID string) (*domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	server, exists := d.servers[externalID]
	if !exists {
		return &domain.ProviderServer{
			ExternalID: externalID,
			Name:       "aws-ec2-instance",
			Status:     domain.ServerStatusRunning,
			PublicIP:   "54.210.10.12",
			PrivateIP:  "172.31.10.12",
			Region:     "us-east-1",
			CPUCores:   2,
			MemoryMB:   4096,
			DiskGB:     50,
			CreatedAt:  time.Now().UTC(),
		}, nil
	}

	return server, nil
}

func (d *EC2Driver) ListServers(ctx context.Context, cred *domain.Credential) ([]domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]domain.ProviderServer, 0, len(d.servers))
	for _, s := range d.servers {
		list = append(list, *s)
	}
	return list, nil
}

func (d *EC2Driver) RebootServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

func (d *EC2Driver) ShutdownServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusStopped
	}
	return nil
}

func (d *EC2Driver) StartServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

func (d *EC2Driver) ResizeServer(ctx context.Context, cred *domain.Credential, req domain.ResizeServerRequest) error {
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

func (d *EC2Driver) DeleteServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.servers, externalID)
	return nil
}
