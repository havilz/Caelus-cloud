package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type VolumeStatus string
type VolumeType string
type FileSystemType string

const (
	VolumeStatusAvailable VolumeStatus = "available"
	VolumeStatusInUse     VolumeStatus = "in-use"
	VolumeStatusMounting  VolumeStatus = "mounting"

	VolumeTypeNVMe         VolumeType = "nvme"
	VolumeTypeSSD          VolumeType = "ssd"
	VolumeTypeDockerVolume VolumeType = "docker-volume"

	FileSystemExt4  FileSystemType = "ext4"
	FileSystemXFS   FileSystemType = "xfs"
	FileSystemBtrfs FileSystemType = "btrfs"
)

type Volume struct {
	ID                    uuid.UUID      `json:"id"`
	OrganizationID        uuid.UUID      `json:"organization_id"`
	ServerID              *uuid.UUID     `json:"server_id,omitempty"`
	Name                  string         `json:"name"`
	SizeGB                int            `json:"size_gb"`
	Type                  VolumeType     `json:"type"`
	FSType                FileSystemType `json:"fs_type"`
	MountPath             string         `json:"mount_path"`
	Status                VolumeStatus   `json:"status"`
	AttachedContainerName string         `json:"attached_container_name,omitempty"`
	IOPS                  int            `json:"iops"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

type StoragePoolStats struct {
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	FreeBytes    uint64  `json:"free_bytes"`
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	FreeGB       float64 `json:"free_gb"`
	UsagePercent float64 `json:"usage_percent"`
	StoragePath  string  `json:"storage_path"`
}

type CreateVolumeRequest struct {
	ServerID  *uuid.UUID     `json:"server_id,omitempty"`
	Name      string         `json:"name" validate:"required,min=3,max=50"`
	SizeGB    int            `json:"size_gb" validate:"required,min=1"`
	Type      VolumeType     `json:"type" validate:"required,oneof=nvme ssd docker-volume"`
	FSType    FileSystemType `json:"fs_type" validate:"required,oneof=ext4 xfs btrfs"`
	MountPath string         `json:"mount_path" validate:"required"`
}

type ResizeVolumeRequest struct {
	NewSizeGB int `json:"new_size_gb" validate:"required,min=1"`
}

type VolumeRepository interface {
	CreateVolume(ctx context.Context, vol *Volume) error
	GetVolumeByID(ctx context.Context, id uuid.UUID) (*Volume, error)
	ListVolumesByOrg(ctx context.Context, orgID uuid.UUID) ([]Volume, error)
	ListVolumesByServer(ctx context.Context, serverID uuid.UUID) ([]Volume, error)
	UpdateVolumeStatus(ctx context.Context, id uuid.UUID, status VolumeStatus, attachedContainer string) error
	UpdateVolumeSize(ctx context.Context, id uuid.UUID, newSizeGB int) error
	DeleteVolume(ctx context.Context, id uuid.UUID) error
}

type AgentAction struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Target  string `json:"target"`
	Payload string `json:"payload,omitempty"`
}
