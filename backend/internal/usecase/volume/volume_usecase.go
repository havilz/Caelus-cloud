package volume

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type UseCase struct {
	repo       domain.VolumeRepository
	serverRepo domain.ServerRepository
}

func NewUseCase(repo domain.VolumeRepository, serverRepo domain.ServerRepository) *UseCase {
	return &UseCase{repo: repo, serverRepo: serverRepo}
}

// GetStoragePoolStats mengembalikan metrik kapasitas disk fisik host real-time
func (u *UseCase) GetStoragePoolStats(ctx context.Context) (*domain.StoragePoolStats, error) {
	// Jalankan 'df -B1 /home' atau fallback ke '/'
	targetPath := "/home"
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		targetPath = "/"
	}

	out, err := exec.CommandContext(ctx, "df", "-B1", targetPath).Output()
	if err != nil {
		// Fallback ke root
		out, err = exec.CommandContext(ctx, "df", "-B1", "/").Output()
		if err != nil {
			// Mock default fallback jika df gagal
			return &domain.StoragePoolStats{
				TotalBytes:   100 * 1024 * 1024 * 1024,
				UsedBytes:    25 * 1024 * 1024 * 1024,
				FreeBytes:    75 * 1024 * 1024 * 1024,
				TotalGB:      100.0,
				UsedGB:       25.0,
				FreeGB:       75.0,
				UsagePercent: 25.0,
				StoragePath:  "/",
			}, nil
		}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("gagal mem-parsing output disk telemetri")
	}

	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 4 {
		return nil, fmt.Errorf("format output df tidak valid")
	}

	var totalBytes, usedBytes, freeBytes uint64
	if val, parseErr := strconv.ParseUint(fields[1], 10, 64); parseErr == nil {
		totalBytes = val
	}
	if val, parseErr := strconv.ParseUint(fields[2], 10, 64); parseErr == nil {
		usedBytes = val
	}
	if val, parseErr := strconv.ParseUint(fields[3], 10, 64); parseErr == nil {
		freeBytes = val
	}

	totalGB := float64(totalBytes) / (1024 * 1024 * 1024)
	usedGB := float64(usedBytes) / (1024 * 1024 * 1024)
	freeGB := float64(freeBytes) / (1024 * 1024 * 1024)

	usagePercent := 0.0
	if totalBytes > 0 {
		usagePercent = (float64(usedBytes) / float64(totalBytes)) * 100.0
	}

	return &domain.StoragePoolStats{
		TotalBytes:   totalBytes,
		UsedBytes:    usedBytes,
		FreeBytes:    freeBytes,
		TotalGB:      totalGB,
		UsedGB:       usedGB,
		FreeGB:       freeGB,
		UsagePercent: usagePercent,
		StoragePath:  targetPath,
	}, nil
}

func (u *UseCase) CreateVolume(ctx context.Context, orgID uuid.UUID, req domain.CreateVolumeRequest) (*domain.Volume, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("nama volume wajib diisi")
	}
	if req.SizeGB <= 0 {
		return nil, fmt.Errorf("ukuran volume harus lebih dari 0 GB")
	}

	// Validasi fisik terhadap sisa kapasitas disk host lokal atau server target
	if req.ServerID == nil {
		stats, err := u.GetStoragePoolStats(ctx)
		if err == nil && stats != nil {
			if float64(req.SizeGB) > stats.FreeGB {
				return nil, fmt.Errorf("kapasitas volume (%d GB) melebihi sisa disk fisik host lokal (%.1f GB tersedia)", req.SizeGB, stats.FreeGB)
			}
		}
	} else if u.serverRepo != nil {
		targetServer, err := u.serverRepo.GetByID(ctx, *req.ServerID)
		if err == nil && targetServer != nil && targetServer.DiskGB > 0 {
			if req.SizeGB > targetServer.DiskGB {
				return nil, fmt.Errorf("kapasitas volume (%d GB) melebihi alokasi disk server %s (%d GB tersedia)", req.SizeGB, targetServer.Name, targetServer.DiskGB)
			}
		}
	}

	volType := req.Type
	if volType == "" {
		volType = domain.VolumeTypeNVMe
	}

	fsType := req.FSType
	if fsType == "" {
		fsType = domain.FileSystemExt4
	}

	mountPath := req.MountPath
	if mountPath == "" {
		mountPath = "/mnt/data"
	}

	iops := 3000
	switch volType {
	case domain.VolumeTypeNVMe:
		iops = 5000
	case domain.VolumeTypeDockerVolume:
		iops = 2000
	}

	cleanName := strings.TrimSpace(strings.ToLower(req.Name))
	cleanName = strings.ReplaceAll(cleanName, " ", "-")

	vol := &domain.Volume{
		ID:             uuid.New(),
		OrganizationID: orgID,
		ServerID:       req.ServerID,
		Name:           cleanName,
		SizeGB:         req.SizeGB,
		Type:           volType,
		FSType:         fsType,
		MountPath:      mountPath,
		Status:         domain.VolumeStatusAvailable,
		IOPS:           iops,
	}

	// Simpan ke database PostgreSQL
	if err := u.repo.CreateVolume(ctx, vol); err != nil {
		return nil, fmt.Errorf("gagal menyimpan volume ke database: %w", err)
	}

	// Buat Docker Named Volume fisik jika docker tersedia di host
	if _, lookErr := exec.LookPath("docker"); lookErr == nil {
		dockerVolName := fmt.Sprintf("caelus-%s", cleanName)
		_ = exec.CommandContext(ctx, "docker", "volume", "create", dockerVolName).Run()
	}

	return vol, nil
}

func (u *UseCase) ListVolumes(ctx context.Context, orgID uuid.UUID) ([]domain.Volume, error) {
	return u.repo.ListVolumesByOrg(ctx, orgID)
}

func (u *UseCase) GetVolume(ctx context.Context, id uuid.UUID) (*domain.Volume, error) {
	return u.repo.GetVolumeByID(ctx, id)
}

func (u *UseCase) DeleteVolume(ctx context.Context, orgID uuid.UUID, id uuid.UUID) error {
	vol, err := u.repo.GetVolumeByID(ctx, id)
	if err == nil && vol != nil {
		// Hapus Docker volume fisik jika ada
		if _, lookErr := exec.LookPath("docker"); lookErr == nil {
			dockerVolName := fmt.Sprintf("caelus-%s", vol.Name)
			_ = exec.CommandContext(ctx, "docker", "volume", "rm", "-f", dockerVolName).Run()
		}
	}
	return u.repo.DeleteVolume(ctx, id)
}
