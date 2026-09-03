package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeploymentStatus string

const (
	DeploymentStatusQueued     DeploymentStatus = "queued"
	DeploymentStatusPulling    DeploymentStatus = "pulling"
	DeploymentStatusBuilding   DeploymentStatus = "building"
	DeploymentStatusDeploying  DeploymentStatus = "deploying"
	DeploymentStatusRunning    DeploymentStatus = "running"
	DeploymentStatusFailed     DeploymentStatus = "failed"
	DeploymentStatusStopped    DeploymentStatus = "stopped"
	DeploymentStatusRolledBack DeploymentStatus = "rolled_back"
)

type PortBinding struct {
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"`
}

type VolumeBinding struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	Mode          string `json:"mode"`
}

type Deployment struct {
	ID                   uuid.UUID         `json:"id"`
	OrganizationID       uuid.UUID         `json:"organization_id"`
	ServerID             *uuid.UUID        `json:"server_id,omitempty"`
	AppName              string            `json:"app_name"`
	ImageTag             string            `json:"image_tag"`
	ContainerName        string            `json:"container_name"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
	PortBindings         []PortBinding     `json:"port_bindings,omitempty"`
	VolumeBindings       []VolumeBinding   `json:"volume_bindings,omitempty"`
	RestartPolicy        string            `json:"restart_policy"`
	NetworkName          string            `json:"network_name,omitempty"`
	Command              string            `json:"command,omitempty"`
	Status               DeploymentStatus  `json:"status"`
	ErrorMessage         string            `json:"error_message,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	FinishedAt           *time.Time        `json:"finished_at,omitempty"`
}

type DeploymentLog struct {
	ID           int64     `json:"id"`
	DeploymentID uuid.UUID `json:"deployment_id"`
	Timestamp    time.Time `json:"timestamp"`
	Stream       string    `json:"stream"`
	Message      string    `json:"message"`
}

type DeploymentRequest struct {
	ServerID             *uuid.UUID        `json:"server_id,omitempty"`
	AppName              string            `json:"app_name"`
	ImageTag             string            `json:"image_tag"`
	ContainerName        string            `json:"container_name,omitempty"`
	Command              string            `json:"command,omitempty"`
	NetworkName          string            `json:"network_name,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
	PortBindings         []PortBinding     `json:"port_bindings,omitempty"`
	VolumeBindings       []VolumeBinding   `json:"volume_bindings,omitempty"`
	RestartPolicy        string            `json:"restart_policy,omitempty"`
}

type DeploymentRepository interface {
	CreateDeployment(ctx context.Context, dep *Deployment) error
	GetDeploymentByID(ctx context.Context, id uuid.UUID) (*Deployment, error)
	ListDeploymentsByOrg(ctx context.Context, orgID uuid.UUID) ([]Deployment, error)
	ListDeploymentsByServer(ctx context.Context, serverID uuid.UUID) ([]Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, id uuid.UUID, status DeploymentStatus, errorMsg string, finishedAt *time.Time) error
	DeleteDeployment(ctx context.Context, id uuid.UUID) error
	AppendLog(ctx context.Context, log *DeploymentLog) error
	GetLogsByDeployment(ctx context.Context, deploymentID uuid.UUID, limit int) ([]DeploymentLog, error)
}
