package custom

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// CustomDriver mengimplementasikan domain.ProviderDriver untuk server mandiri (BYOS / Home Server / Existing VPS).
type CustomDriver struct {
	mu      sync.RWMutex
	servers map[string]*domain.ProviderServer
}

// NewCustomDriver membuat instance baru CustomDriver.
// Mengembalikan implementasi interface domain.ProviderDriver.
func NewCustomDriver() domain.ProviderDriver {
	return &CustomDriver{
		servers: make(map[string]*domain.ProviderServer),
	}
}

// CreateServer mendaftarkan server kustom/eksternal ke dalam sistem Caelus dengan status pending menunggu heartbeat agent.
// Parameter ctx merupakan konteks eksekusi operasi.
// Parameter cred merupakan kredensial provider opsional.
// Parameter req memuat spesifikasi dan data server yang didaftarkan.
// Mengembalikan pointer *domain.ProviderServer.
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

// GetServer mengambil detail server kustom dari penyimpanan internal.
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

// ListServers mengambil seluruh daftar server kustom.
func (d *CustomDriver) ListServers(ctx context.Context, cred *domain.Credential) ([]domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]domain.ProviderServer, 0, len(d.servers))
	for _, s := range d.servers {
		list = append(list, *s)
	}
	return list, nil
}

// StartServer menyimulasikan perintah start pada server kustom.
func (d *CustomDriver) StartServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

// ShutdownServer menyimulasikan perintah shutdown pada server kustom.
func (d *CustomDriver) ShutdownServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusStopped
	}
	return nil
}

// RebootServer menyimulasikan perintah reboot pada server kustom.
func (d *CustomDriver) RebootServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.servers[externalID]; ok {
		s.Status = domain.ServerStatusRunning
	}
	return nil
}

// ResizeServer mengubah alokasi spesifikasi komputasi server kustom.
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

// DeleteServer menghapus entri server kustom dari registri driver.
func (d *CustomDriver) DeleteServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.servers, externalID)
	return nil
}
