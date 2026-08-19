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

// NewMockDriver menginisialisasi driver provider Mock untuk simulasi komputasi cloud secara lokal.
// Mengembalikan implementasi interface domain.ProviderDriver.
func NewMockDriver() domain.ProviderDriver {
	return &MockDriver{
		servers: make(map[string]*domain.ProviderServer),
	}
}

// CreateServer menyimulasikan provisioning server VPS baru pada infrastruktur cloud tiruan.
// Parameter ctx merupakan konteks eksekusi operasi.
// Parameter cred merupakan kredensial provider yang digunakan.
// Parameter req memuat spesifikasi dan konfigurasi server yang akan dibuat.
// Mengembalikan pointer *domain.ProviderServer jika provisioning berhasil atau error jika parameter tidak valid.
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

// GetServer mengambil status dan detail informasi server dari provider tiruan berdasarkan identifier eksternal.
// Parameter ctx merupakan konteks eksekusi operasi.
// Parameter cred merupakan kredensial provider.
// Parameter externalID merupakan identifier eksternal unik server pada provider.
// Mengembalikan pointer *domain.ProviderServer atau domain.ErrNotFound jika server tidak ditemukan.
func (d *MockDriver) GetServer(ctx context.Context, cred *domain.Credential, externalID string) (*domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	server, exists := d.servers[externalID]
	if !exists {
		return nil, domain.ErrNotFound
	}

	return server, nil
}

// ListServers mengambil seluruh daftar instance server yang terdaftar pada provider tiruan.
// Parameter ctx merupakan konteks eksekusi operasi.
// Parameter cred merupakan kredensial provider.
// Mengembalikan slice []domain.ProviderServer dan error jika terjadi kegagalan.
func (d *MockDriver) ListServers(ctx context.Context, cred *domain.Credential) ([]domain.ProviderServer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]domain.ProviderServer, 0, len(d.servers))
	for _, s := range d.servers {
		list = append(list, *s)
	}

	return list, nil
}

// RebootServer menyimulasikan restart sistem operasi server pada provider tiruan.
// Parameter ctx merupakan konteks eksekusi operasi.
// Parameter cred merupakan kredensial provider.
// Parameter externalID merupakan identifier eksternal server.
// Mengembalikan domain.ErrNotFound jika server tidak ditemukan.
func (d *MockDriver) RebootServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	return d.updateServerStatus(externalID, domain.ServerStatusRunning)
}

// ShutdownServer menyimulasikan pematian daya (power off / stop) pada instance server tiruan.
// Parameter ctx merupakan konteks eksekusi operasi.
// Parameter cred merupakan kredensial provider.
// Parameter externalID merupakan identifier eksternal server.
// Mengembalikan domain.ErrNotFound jika server tidak ditemukan.
func (d *MockDriver) ShutdownServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	return d.updateServerStatus(externalID, domain.ServerStatusStopped)
}

// StartServer menyimulasikan penyalaan daya (power on / start) pada instance server tiruan yang sedang berhenti.
// Parameter ctx merupakan konteks eksekusi operasi.
// Parameter cred merupakan kredensial provider.
// Parameter externalID merupakan identifier eksternal server.
// Mengembalikan domain.ErrNotFound jika server tidak ditemukan.
func (d *MockDriver) StartServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	return d.updateServerStatus(externalID, domain.ServerStatusRunning)
}

// ResizeServer menyimulasikan pengubahan kapasitas spesifikasi vCPU, RAM, dan Disk instance server.
// Parameter ctx merupakan konteks eksekusi operasi.
// Parameter cred merupakan kredensial provider.
// Parameter req memuat identifier server dan spesifikasi target baru.
// Mengembalikan domain.ErrNotFound jika server tidak ditemukan.
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

// DeleteServer menyimulasikan terminasi dan penghapusan permanen instance server dari provider.
// Parameter ctx merupakan konteks eksekusi operasi.
// Parameter cred merupakan kredensial provider.
// Parameter externalID merupakan identifier eksternal server.
// Mengembalikan domain.ErrNotFound jika server tidak ditemukan.
func (d *MockDriver) DeleteServer(ctx context.Context, cred *domain.Credential, externalID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.servers[externalID]; !exists {
		return domain.ErrNotFound
	}

	delete(d.servers, externalID)
	return nil
}

// updateServerStatus memperbarui status operasional server secara thread-safe.
// Parameter externalID merupakan identifier unik server.
// Parameter targetStatus merupakan status baru yang akan diterapkan.
// Mengembalikan domain.ErrNotFound jika server tidak ditemukan.
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
