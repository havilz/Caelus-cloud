package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// IaCAction merepresentasikan jenis tindakan rekonsiliasi yang diusulkan oleh Plan Engine.
type IaCAction string

const (
	ActionCreate IaCAction = "create"
	ActionUpdate IaCAction = "update"
	ActionDelete IaCAction = "delete"
	ActionNoOp   IaCAction = "noop"
)

// IaCStatus merepresentasikan status siklus hidup konfigurasi atau rencana IaC.
type IaCStatus string

const (
	IaCStatusDraft      IaCStatus = "draft"
	IaCStatusPlanned    IaCStatus = "planned"
	IaCStatusApplying   IaCStatus = "applying"
	IaCStatusApplied    IaCStatus = "applied"
	IaCStatusFailed     IaCStatus = "failed"
	IaCStatusDrifted    IaCStatus = "drifted"
	IaCStatusRolledBack IaCStatus = "rolled_back"
)

// ResourceType mendefinisikan jenis resource yang didukung dalam declarative YAML.
type ResourceType string

const (
	ResourceTypeServer    ResourceType = "server"
	ResourceTypeStorage   ResourceType = "storage"
	ResourceTypeContainer ResourceType = "container"
	ResourceTypeRule      ResourceType = "rule"
)

// ServerSpec mendefinisikan spesifikasi deklaratif server VPS/Cloud.
type ServerSpec struct {
	Name     string            `json:"name" yaml:"name"`
	Provider string            `json:"provider" yaml:"provider"` // aws, digitalocean, hetzner, contabo, mock
	Region   string            `json:"region" yaml:"region"`
	Size     string            `json:"size" yaml:"size"` // e.g. t3.micro, s-1vcpu-1gb
	Image    string            `json:"image" yaml:"image"`
	Tags     map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
	SSHKeys  []string          `json:"ssh_keys,omitempty" yaml:"ssh_keys,omitempty"`
}

// StorageSpec mendefinisikan spesifikasi deklaratif object storage bucket.
type StorageSpec struct {
	Name       string `json:"name" yaml:"name"`
	Type       string `json:"type" yaml:"type"` // local, s3, r2
	Region     string `json:"region,omitempty" yaml:"region,omitempty"`
	Versioning bool   `json:"versioning,omitempty" yaml:"versioning,omitempty"`
	Access     string `json:"access,omitempty" yaml:"access,omitempty"` // private, public-read
}

// ContainerSpec mendefinisikan spesifikasi deklaratif container Docker.
type ContainerSpec struct {
	Name          string            `json:"name" yaml:"name"`
	Server        string            `json:"server,omitempty" yaml:"server,omitempty"` // Server Name or Server UUID
	Image         string            `json:"image" yaml:"image"`
	Ports         []string          `json:"ports,omitempty" yaml:"ports,omitempty"` // e.g. ["80:80", "443:443"]
	Environment   map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Volumes       []string          `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	RestartPolicy string            `json:"restart_policy,omitempty" yaml:"restart_policy,omitempty"`
}

// RuleSpec mendefinisikan spesifikasi deklaratif automation rule.
type RuleSpec struct {
	Name      string                 `json:"name" yaml:"name"`
	Trigger   string                 `json:"trigger" yaml:"trigger"`
	Condition map[string]interface{} `json:"condition,omitempty" yaml:"condition,omitempty"`
	Action    map[string]interface{} `json:"action" yaml:"action"`
}

// DeclarativeManifest merepresentasikan dokumen YAML utuh yang diparsing.
type DeclarativeManifest struct {
	Version    string          `json:"version" yaml:"version"`
	Servers    []ServerSpec    `json:"servers,omitempty" yaml:"servers,omitempty"`
	Storages   []StorageSpec   `json:"storages,omitempty" yaml:"storages,omitempty"`
	Containers []ContainerSpec `json:"containers,omitempty" yaml:"containers,omitempty"`
	Rules      []RuleSpec      `json:"rules,omitempty" yaml:"rules,omitempty"`
}

// IaCChange merepresentasikan satu unit diff perubahan pada resource tertentu.
type IaCChange struct {
	ResourceType  ResourceType           `json:"resource_type"`
	ResourceName  string                 `json:"resource_name"`
	Action        IaCAction              `json:"action"` // create, update, delete, noop
	Before        map[string]interface{} `json:"before,omitempty"`
	After         map[string]interface{} `json:"after,omitempty"`
	ChangedFields []string               `json:"changed_fields,omitempty"`
	Reason        string                 `json:"reason,omitempty"`
}

// IaCSummary merangkum jumlah aksi dalam sebuah rencana.
type IaCSummary struct {
	Create int `json:"create"`
	Update int `json:"update"`
	Delete int `json:"delete"`
	NoOp   int `json:"noop"`
	Total  int `json:"total"`
}

// IaCConfiguration merepresentasikan file konfigurasi deklaratif pengguna.
type IaCConfiguration struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	RawYAML        string    `json:"raw_yaml"`
	Status         IaCStatus `json:"status"`
	CurrentVersion int       `json:"current_version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// IaCState merepresentasikan snapshot actual state infrastruktur yang tercatat.
type IaCState struct {
	ID              uuid.UUID              `json:"id"`
	ConfigurationID uuid.UUID              `json:"configuration_id"`
	Version         int                    `json:"version"`
	StateData       map[string]interface{} `json:"state_data"`
	Hash            string                 `json:"hash"`
	AppliedAt       time.Time              `json:"applied_at"`
	AppliedBy       *uuid.UUID             `json:"applied_by,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// IaCPlan merepresentasikan hasil komputasi Plan Engine sebelum diaplikasikan.
type IaCPlan struct {
	ID              uuid.UUID   `json:"id"`
	ConfigurationID uuid.UUID   `json:"configuration_id"`
	TargetVersion   int         `json:"target_version"`
	Changes         []IaCChange `json:"changes"`
	Summary         IaCSummary  `json:"summary"`
	Status          IaCStatus   `json:"status"`
	ErrorMessage    string      `json:"error_message,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	ExecutedAt      *time.Time  `json:"executed_at,omitempty"`
}

// IaCValidationError mendefinisikan error validasi sintaks / skema dengan baris dan kolom.
type IaCValidationError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// IaCValidationResponse mengembalikan hasil validasi sintaks konfigurasi deklaratif.
type IaCValidationResponse struct {
	Valid    bool                 `json:"valid"`
	Errors   []IaCValidationError `json:"errors,omitempty"`
	Manifest *DeclarativeManifest `json:"manifest,omitempty"`
}

// IaCRepository mendefinisikan interface persistensi database untuk modul IaC.
type IaCRepository interface {
	CreateConfig(ctx context.Context, config *IaCConfiguration) error
	GetConfigByID(ctx context.Context, id uuid.UUID) (*IaCConfiguration, error)
	ListConfigsByOrg(ctx context.Context, orgID uuid.UUID) ([]IaCConfiguration, error)
	UpdateConfig(ctx context.Context, config *IaCConfiguration) error
	DeleteConfig(ctx context.Context, id uuid.UUID) error

	CreateState(ctx context.Context, state *IaCState) error
	GetLatestStateByConfigID(ctx context.Context, configID uuid.UUID) (*IaCState, error)
	GetStateByVersion(ctx context.Context, configID uuid.UUID, version int) (*IaCState, error)
	ListStatesByConfigID(ctx context.Context, configID uuid.UUID) ([]IaCState, error)

	CreatePlan(ctx context.Context, plan *IaCPlan) error
	GetPlanByID(ctx context.Context, id uuid.UUID) (*IaCPlan, error)
	UpdatePlanStatus(ctx context.Context, id uuid.UUID, status IaCStatus, errorMsg string) error
	GetLatestPlanByConfigID(ctx context.Context, configID uuid.UUID) (*IaCPlan, error)
}
