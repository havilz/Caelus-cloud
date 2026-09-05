package orchestration

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/orchestration/pipeline"
	"github.com/havilz/caelus-cloud/backend/pkg/security"
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

	for i, vb := range req.VolumeBindings {
		canonicalPath, err := security.ValidateHostPath(vb.HostPath)
		if err != nil {
			return nil, fmt.Errorf("volume binding tidak aman: %w", err)
		}
		req.VolumeBindings[i].HostPath = canonicalPath
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

	if err := u.pipeline.Execute(ctx, dep); err != nil {
		return nil, fmt.Errorf("failed triggering deployment pipeline: %w", err)
	}

	return dep, nil
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

		_ = exec.CommandContext(ctx, "docker", "rm", "-f", dep.ContainerName).Run()
	}

	return u.repo.DeleteDeployment(ctx, id)
}

func (u *UseCase) RollbackDeployment(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	oldDep, err := u.repo.GetDeploymentByID(ctx, id)
	if err != nil {
		return nil, err
	}

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

func validateDockerName(name, fieldName string) error {
	if !validDockerNameRegex.MatchString(name) {
		return fmt.Errorf("%s '%s' mengandung karakter tidak diizinkan atau melebihi batas panjang", fieldName, name)
	}
	return nil
}

func validateImageTag(tag string) error {
	if !validImageTagRegex.MatchString(tag) {
		return fmt.Errorf("image_tag '%s' tidak valid atau mengandung karakter berbahaya", tag)
	}
	return nil
}

func validateNetworkName(netName string) error {
	if !validNetworkNameRegex.MatchString(netName) {
		return fmt.Errorf("network_name '%s' tidak valid", netName)
	}
	return nil
}

func validateRestartPolicy(policy string) error {
	switch policy {
	case "no", "always", "unless-stopped", "on-failure":
		return nil
	default:
		return fmt.Errorf("restart_policy '%s' tidak valid (diizinkan: no, always, unless-stopped, on-failure)", policy)
	}
}
