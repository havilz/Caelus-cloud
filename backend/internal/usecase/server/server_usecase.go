package server

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	provFactory "github.com/havilz/caelus-cloud/backend/internal/provider"
)

type CreateServerInput struct {
	OrganizationID uuid.UUID         `json:"organization_id"`
	ProviderID     uuid.UUID         `json:"provider_id"`
	CredentialID   *uuid.UUID        `json:"credential_id,omitempty"`
	Name           string            `json:"name"`
	Region         string            `json:"region"`
	OSType         string            `json:"os_type"`
	PlanID         string            `json:"plan_id"`
	CPUCores       int               `json:"cpu_cores"`
	MemoryMB       int               `json:"memory_mb"`
	DiskGB         int               `json:"disk_gb"`
	SSHKey         string            `json:"ssh_key,omitempty"`
	UserData       string            `json:"user_data,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
}

type ResizeServerInput struct {
	CPUCores int    `json:"cpu_cores"`
	MemoryMB int    `json:"memory_mb"`
	DiskGB   int    `json:"disk_gb"`
	PlanID   string `json:"plan_id,omitempty"`
}

type ServerUsecase interface {
	CreateServer(ctx context.Context, input CreateServerInput) (*domain.Server, error)
	GetServer(ctx context.Context, orgID, serverID uuid.UUID) (*domain.Server, error)
	ListServers(ctx context.Context, orgID uuid.UUID, page, limit int) ([]domain.Server, int64, error)
	RebootServer(ctx context.Context, orgID, serverID uuid.UUID) error
	ShutdownServer(ctx context.Context, orgID, serverID uuid.UUID) error
	StartServer(ctx context.Context, orgID, serverID uuid.UUID) error
	ResizeServer(ctx context.Context, orgID, serverID uuid.UUID, input ResizeServerInput) error
	DeleteServer(ctx context.Context, orgID, serverID uuid.UUID) error
}

type serverUsecase struct {
	serverRepo    domain.ServerRepository
	providerRepo  domain.ProviderRepository
	credRepo      domain.CredentialRepository
	driverFactory provFactory.Factory
}

// NewServerUsecase menginisialisasi use case manajemen server VPS dengan integrasi repositori dan driver provider.
// Parameter serverRepo merupakan implementasi domain.ServerRepository.
// Parameter providerRepo merupakan implementasi domain.ProviderRepository.
// Parameter credRepo merupakan implementasi domain.CredentialRepository.
// Parameter factory merupakan implementasi provFactory.Factory.
// Mengembalikan instance interface ServerUsecase.
func NewServerUsecase(
	serverRepo domain.ServerRepository,
	providerRepo domain.ProviderRepository,
	credRepo domain.CredentialRepository,
	factory provFactory.Factory,
) ServerUsecase {
	return &serverUsecase{
		serverRepo:    serverRepo,
		providerRepo:  providerRepo,
		credRepo:      credRepo,
		driverFactory: factory,
	}
}

// CreateServer memvalidasi input, melakukan provisioning instance pada provider cloud via driver, dan mencatat data server ke database.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter input memuat konfigurasi server yang akan dibuat.
// Mengembalikan pointer *domain.Server yang dibuat atau error jika proses gagal.
func (u *serverUsecase) CreateServer(ctx context.Context, input CreateServerInput) (*domain.Server, error) {
	if err := validateCreateServerInput(&input); err != nil {
		return nil, err
	}

	provider, driver, cred, err := u.resolveDriverAndCredential(ctx, input.ProviderID, input.CredentialID)
	if err != nil {
		return nil, err
	}

	driverReq := domain.CreateServerRequest{
		Name:     input.Name,
		Region:   input.Region,
		OSType:   input.OSType,
		PlanID:   input.PlanID,
		CPUCores: input.CPUCores,
		MemoryMB: input.MemoryMB,
		DiskGB:   input.DiskGB,
		SSHKey:   input.SSHKey,
		UserData: input.UserData,
		Tags:     input.Tags,
	}

	provServer, err := driver.CreateServer(ctx, cred, driverReq)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	newServer := &domain.Server{
		ID:               uuid.New(),
		OrganizationID:   input.OrganizationID,
		CredentialID:     input.CredentialID,
		ProviderID:       input.ProviderID,
		ExternalServerID: &provServer.ExternalID,
		Name:             input.Name,
		Hostname:         &provServer.Name,
		IPAddress:        &provServer.PublicIP,
		Status:           provServer.Status,
		OSType:           input.OSType,
		CPUCores:         input.CPUCores,
		MemoryMB:         input.MemoryMB,
		DiskGB:           input.DiskGB,
		Region:           input.Region,
		CreatedAt:        now,
		UpdatedAt:        now,
		Provider:         provider,
	}

	if err := u.serverRepo.Create(ctx, newServer); err != nil {
		return nil, err
	}

	return newServer, nil
}

// GetServer mengambil data detail server dan memastikan kepemilikan organisasi yang sah.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter orgID merupakan UUID organisasi pemilik server.
// Parameter serverID merupakan UUID server yang diminta.
// Mengembalikan pointer *domain.Server atau error jika server tidak ditemukan atau akses dilarang.
func (u *serverUsecase) GetServer(ctx context.Context, orgID, serverID uuid.UUID) (*domain.Server, error) {
	server, err := u.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return nil, err
	}

	if server.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	return server, nil
}

// ListServers mengambil daftar server milik suatu organisasi dengan paginasi.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter orgID merupakan UUID organisasi pemilik server.
// Parameter page merupakan nomor halaman data (1-based).
// Parameter limit merupakan batas data per halaman.
// Mengembalikan slice []domain.Server, total data int64, dan error jika query gagal.
func (u *serverUsecase) ListServers(ctx context.Context, orgID uuid.UUID, page, limit int) ([]domain.Server, int64, error) {
	return u.serverRepo.ListByOrg(ctx, orgID, page, limit)
}

// RebootServer menyalakan ulang server pada provider cloud dan memperbarui status lokal menjadi running.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter orgID merupakan UUID organisasi pemilik server.
// Parameter serverID merupakan UUID server.
// Mengembalikan error jika operasi gagal atau server tidak ditemukan.
func (u *serverUsecase) RebootServer(ctx context.Context, orgID, serverID uuid.UUID) error {
	return u.executeServerAction(ctx, orgID, serverID, domain.ServerStatusRunning, func(d domain.ProviderDriver, c *domain.Credential, extID string) error {
		return d.RebootServer(ctx, c, extID)
	})
}

// ShutdownServer mematikan daya server pada provider cloud dan memperbarui status lokal menjadi stopped.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter orgID merupakan UUID organisasi pemilik server.
// Parameter serverID merupakan UUID server.
// Mengembalikan error jika operasi gagal atau server tidak ditemukan.
func (u *serverUsecase) ShutdownServer(ctx context.Context, orgID, serverID uuid.UUID) error {
	return u.executeServerAction(ctx, orgID, serverID, domain.ServerStatusStopped, func(d domain.ProviderDriver, c *domain.Credential, extID string) error {
		return d.ShutdownServer(ctx, c, extID)
	})
}

// StartServer menyalakan kembali server pada provider cloud dan memperbarui status lokal menjadi running.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter orgID merupakan UUID organisasi pemilik server.
// Parameter serverID merupakan UUID server.
// Mengembalikan error jika operasi gagal atau server tidak ditemukan.
func (u *serverUsecase) StartServer(ctx context.Context, orgID, serverID uuid.UUID) error {
	return u.executeServerAction(ctx, orgID, serverID, domain.ServerStatusRunning, func(d domain.ProviderDriver, c *domain.Credential, extID string) error {
		return d.StartServer(ctx, c, extID)
	})
}

// ResizeServer mengubah kapasitas spesifikasi vCPU, RAM, dan Disk server pada provider dan memperbarui data lokal.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter orgID merupakan UUID organisasi pemilik server.
// Parameter serverID merupakan UUID server.
// Parameter input memuat spesifikasi baru.
// Mengembalikan error jika operasi gagal atau server tidak ditemukan.
func (u *serverUsecase) ResizeServer(ctx context.Context, orgID, serverID uuid.UUID, input ResizeServerInput) error {
	server, err := u.GetServer(ctx, orgID, serverID)
	if err != nil {
		return err
	}

	_, driver, cred, err := u.resolveDriverAndCredential(ctx, server.ProviderID, server.CredentialID)
	if err != nil {
		return err
	}

	if server.ExternalServerID != nil {
		req := domain.ResizeServerRequest{
			ExternalID: *server.ExternalServerID,
			PlanID:     input.PlanID,
			CPUCores:   input.CPUCores,
			MemoryMB:   input.MemoryMB,
			DiskGB:     input.DiskGB,
		}
		if err := driver.ResizeServer(ctx, cred, req); err != nil {
			return err
		}
	}

	if input.CPUCores > 0 {
		server.CPUCores = input.CPUCores
	}
	if input.MemoryMB > 0 {
		server.MemoryMB = input.MemoryMB
	}
	if input.DiskGB > 0 {
		server.DiskGB = input.DiskGB
	}
	server.UpdatedAt = time.Now()

	return u.serverRepo.Update(ctx, server)
}

// DeleteServer menghapus instance server dari provider cloud dan menghapus rekaman server dari database.
// Parameter ctx merupakan konteks eksekusi use case.
// Parameter orgID merupakan UUID organisasi pemilik server.
// Parameter serverID merupakan UUID server yang akan dihapus.
// Mengembalikan error jika operasi gagal atau server tidak ditemukan.
func (u *serverUsecase) DeleteServer(ctx context.Context, orgID, serverID uuid.UUID) error {
	server, err := u.GetServer(ctx, orgID, serverID)
	if err != nil {
		return err
	}

	_, driver, cred, err := u.resolveDriverAndCredential(ctx, server.ProviderID, server.CredentialID)
	if err == nil && server.ExternalServerID != nil {
		_ = driver.DeleteServer(ctx, cred, *server.ExternalServerID)
	}

	return u.serverRepo.Delete(ctx, serverID)
}

// validateCreateServerInput memverifikasi kelengkapan dan validitas parameter pembuatan server baru.
// Parameter input merupakan pointer data input pembuatan server.
// Mengembalikan error domain.ErrBadRequest jika ada data wajib yang kosong.
func validateCreateServerInput(input *CreateServerInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Region = strings.TrimSpace(input.Region)
	input.OSType = strings.TrimSpace(input.OSType)

	if input.Name == "" || input.OrganizationID == uuid.Nil || input.ProviderID == uuid.Nil {
		return domain.ErrBadRequest
	}
	if input.Region == "" {
		input.Region = "custom"
	}
	if input.OSType == "" {
		input.OSType = "auto-detect"
	}
	if input.CPUCores <= 0 {
		input.CPUCores = 1
	}
	if input.MemoryMB <= 0 {
		input.MemoryMB = 1024
	}
	if input.DiskGB <= 0 {
		input.DiskGB = 25
	}
	return nil
}

// resolveDriverAndCredential memuat entitas Provider, menginstansiasi driver provider dari factory, dan mengambil kredensial terkait jika tersedia.
// Parameter ctx merupakan konteks eksekusi.
// Parameter providerID merupakan UUID provider.
// Parameter credID merupakan pointer UUID kredensial (opsional).
// Mengembalikan pointer *domain.Provider, instance domain.ProviderDriver, pointer *domain.Credential, dan error.
func (u *serverUsecase) resolveDriverAndCredential(ctx context.Context, providerID uuid.UUID, credID *uuid.UUID) (*domain.Provider, domain.ProviderDriver, *domain.Credential, error) {
	provider, err := u.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return nil, nil, nil, err
	}

	driver, err := u.driverFactory.GetDriver(provider.Slug)
	if err != nil {
		return nil, nil, nil, err
	}

	var cred *domain.Credential
	if credID != nil && *credID != uuid.Nil {
		cred, _ = u.credRepo.GetByID(ctx, *credID)
	}

	return provider, driver, cred, nil
}

// executeServerAction menjalankan aksi kontrol pada driver provider dan memperbarui status server pada database.
// Parameter ctx merupakan konteks eksekusi.
// Parameter orgID merupakan UUID organisasi.
// Parameter serverID merupakan UUID server.
// Parameter targetStatus merupakan status baru yang akan ditetapkan ke database.
// Parameter actionFn merupakan fungsi eksekusi kontrol driver.
// Mengembalikan error jika operasi gagal.
func (u *serverUsecase) executeServerAction(
	ctx context.Context,
	orgID, serverID uuid.UUID,
	targetStatus domain.ServerStatus,
	actionFn func(domain.ProviderDriver, *domain.Credential, string) error,
) error {
	server, err := u.GetServer(ctx, orgID, serverID)
	if err != nil {
		return err
	}

	_, driver, cred, err := u.resolveDriverAndCredential(ctx, server.ProviderID, server.CredentialID)
	if err != nil {
		return err
	}

	if server.ExternalServerID != nil {
		if err := actionFn(driver, cred, *server.ExternalServerID); err != nil {
			return err
		}
	}

	return u.serverRepo.UpdateStatus(ctx, serverID, targetStatus)
}
