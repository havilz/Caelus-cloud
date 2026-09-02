package orchestration

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/orchestration/pipeline"
)

type UseCase struct {
	repo     domain.DeploymentRepository
	pipeline *pipeline.DockerPipeline
}

func NewUseCase(repo domain.DeploymentRepository, broadcaster pipeline.LogBroadcaster) *UseCase {
	pipe := pipeline.NewDockerPipeline(repo, broadcaster)
	return &UseCase{
		repo:     repo,
		pipeline: pipe,
	}
}

func (u *UseCase) CreateDeployment(ctx context.Context, orgID uuid.UUID, req domain.DeploymentRequest) (*domain.Deployment, error) {
	if req.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}
	if err := validateDockerName(req.AppName, "app_name"); err != nil {
		return nil, err
	}
	if req.ImageTag == "" {
		return nil, fmt.Errorf("image_tag is required")
	}
	if err := validateImageTag(req.ImageTag); err != nil {
		return nil, err
	}

	containerName := req.ContainerName
	if containerName == "" {
		containerName = fmt.Sprintf("caelus-%s", req.AppName)
	}
	if err := validateDockerName(containerName, "container_name"); err != nil {
		return nil, err
	}

	if req.NetworkName != "" {
		if err := validateNetworkName(req.NetworkName); err != nil {
			return nil, err
		}
	}

	restartPolicy := req.RestartPolicy
	if restartPolicy == "" {
		restartPolicy = "unless-stopped"
	}
	if err := validateRestartPolicy(restartPolicy); err != nil {
		return nil, err
	}

	// Validasi dan sanitasi seluruh bind-mount path host (C-2: Container Escape Mitigation)
	for _, vb := range req.VolumeBindings {
		if err := validateHostPath(vb.HostPath); err != nil {
			return nil, fmt.Errorf("volume binding tidak aman: %w", err)
		}
	}

	dep := &domain.Deployment{
		ID:                   uuid.New(),
		OrganizationID:       orgID,
		ServerID:             req.ServerID,
		AppName:              req.AppName,
		ImageTag:             req.ImageTag,
		ContainerName:        containerName,
		EnvironmentVariables: req.EnvironmentVariables,
		PortBindings:         req.PortBindings,
		VolumeBindings:       req.VolumeBindings,
		RestartPolicy:        restartPolicy,
		NetworkName:          req.NetworkName,
		Command:              req.Command,
		Status:               domain.DeploymentStatusQueued,
	}

	if err := u.repo.CreateDeployment(ctx, dep); err != nil {
		return nil, fmt.Errorf("failed creating deployment record: %w", err)
	}

	// Trigger asynchronous Docker deployment pipeline
	if err := u.pipeline.Execute(ctx, dep); err != nil {
		return nil, fmt.Errorf("failed triggering deployment pipeline: %w", err)
	}

	return dep, nil
}

// validateHostPath memvalidasi path host bind-mount untuk mencegah teknik container escape (C-2).
// Menolak path root, direktori sistem sensitif, dan socket Docker daemon.
func validateHostPath(hostPath string) error {
	if hostPath == "" {
		return fmt.Errorf("host_path tidak boleh kosong")
	}

	// Bersihkan path untuk mencegah traversal (../../etc)
	clean := filepath.Clean(hostPath)

	// Daftar path yang secara absolut dilarang
	blockedPaths := []string{
		"/",
		"/etc",
		"/root",
		"/bin",
		"/sbin",
		"/usr",
		"/lib",
		"/lib64",
		"/boot",
		"/sys",
		"/proc",
		"/dev",
		"/run",
		"/var/run",
		"/var/run/docker.sock",
		"/var/lib/docker",
		"/home/docker-data",
	}

	for _, blocked := range blockedPaths {
		if clean == blocked {
			return fmt.Errorf("host_path '%s' adalah direktori sistem yang dilarang", hostPath)
		}
	}

	// Blokir path yang merupakan sub-direktori dari path sensitif
	sensitivePrefixes := []string{
		"/etc/",
		"/root/",
		"/bin/",
		"/sbin/",
		"/usr/",
		"/lib/",
		"/lib64/",
		"/boot/",
		"/sys/",
		"/proc/",
		"/dev/",
		"/var/run/",
		"/var/lib/docker/",
		"/home/docker-data/",
	}

	for _, prefix := range sensitivePrefixes {
		if strings.HasPrefix(clean+"/", prefix) {
			return fmt.Errorf("host_path '%s' berada dalam direktori sistem yang dilarang", hostPath)
		}
	}

	return nil
}

func (u *UseCase) GetDeployment(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	return u.repo.GetDeploymentByID(ctx, id)
}

func (u *UseCase) ListDeployments(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) ([]domain.Deployment, error) {
	if serverID != nil && *serverID != uuid.Nil {
		return u.repo.ListDeploymentsByServer(ctx, *serverID)
	}
	return u.repo.ListDeploymentsByOrg(ctx, orgID)
}

func (u *UseCase) GetLogs(ctx context.Context, deploymentID uuid.UUID, limit int) ([]domain.DeploymentLog, error) {
	return u.repo.GetLogsByDeployment(ctx, deploymentID, limit)
}

func (u *UseCase) StopDeployment(ctx context.Context, id uuid.UUID) error {
	dep, err := u.repo.GetDeploymentByID(ctx, id)
	if err != nil {
		return err
	}

	// Hentikan container fisik di host jika ada
	_ = exec.CommandContext(ctx, "docker", "stop", dep.ContainerName).Run()

	now := time.Now().UTC()
	return u.repo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusStopped, "Manually stopped by operator", &now)
}

func (u *UseCase) RedeployDeployment(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	dep, err := u.repo.GetDeploymentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	dep.Status = domain.DeploymentStatusQueued
	_ = u.repo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusQueued, "", nil)

	if err := u.pipeline.Execute(ctx, dep); err != nil {
		return nil, fmt.Errorf("failed triggering redeployment: %w", err)
	}

	return dep, nil
}

func (u *UseCase) DeleteDeployment(ctx context.Context, id uuid.UUID) error {
	dep, err := u.repo.GetDeploymentByID(ctx, id)
	if err == nil && dep != nil {
		// Hapus container fisik dari Docker daemon
		_ = exec.CommandContext(ctx, "docker", "rm", "-f", dep.ContainerName).Run()
	}

	return u.repo.DeleteDeployment(ctx, id)
}

func (u *UseCase) RollbackDeployment(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	oldDep, err := u.repo.GetDeploymentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create a new rollback deployment
	req := domain.DeploymentRequest{
		ServerID:             oldDep.ServerID,
		AppName:              oldDep.AppName,
		ImageTag:             oldDep.ImageTag,
		ContainerName:        oldDep.ContainerName,
		EnvironmentVariables: oldDep.EnvironmentVariables,
		PortBindings:         oldDep.PortBindings,
		VolumeBindings:       oldDep.VolumeBindings,
		RestartPolicy:        oldDep.RestartPolicy,
	}

	newDep, err := u.CreateDeployment(ctx, oldDep.OrganizationID, req)
	if err != nil {
		return nil, fmt.Errorf("failed triggering rollback deployment: %w", err)
	}

	now := time.Now().UTC()
	_ = u.repo.UpdateDeploymentStatus(ctx, oldDep.ID, domain.DeploymentStatusRolledBack, fmt.Sprintf("Rolled back to %s", newDep.ID), &now)

	return newDep, nil
}

var (
	validDockerNameRegex  = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`)
	validImageTagRegex    = regexp.MustCompile(`^[a-zA-Z0-9_.:/\-@]{1,255}$`)
	validNetworkNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,128}$`)
)

// validateDockerName memvalidasi nama container atau aplikasi berbasis aturan penamaan standar Docker (M-1).
func validateDockerName(name, fieldName string) error {
	if !validDockerNameRegex.MatchString(name) {
		return fmt.Errorf("%s '%s' mengandung karakter tidak diizinkan atau melebihi batas panjang", fieldName, name)
	}
	return nil
}

// validateImageTag memvalidasi format tag image Docker untuk mencegah flag injection (M-1).
func validateImageTag(tag string) error {
	if !validImageTagRegex.MatchString(tag) {
		return fmt.Errorf("image_tag '%s' tidak valid atau mengandung karakter berbahaya", tag)
	}
	return nil
}

// validateNetworkName memvalidasi nama network Docker (M-1).
func validateNetworkName(netName string) error {
	if !validNetworkNameRegex.MatchString(netName) {
		return fmt.Errorf("network_name '%s' tidak valid", netName)
	}
	return nil
}

// validateRestartPolicy memvalidasi nilai restart policy yang diperbolehkan oleh Docker CLI (M-1).
func validateRestartPolicy(policy string) error {
	switch policy {
	case "no", "always", "unless-stopped", "on-failure":
		return nil
	default:
		return fmt.Errorf("restart_policy '%s' tidak valid (diizinkan: no, always, unless-stopped, on-failure)", policy)
	}
}
