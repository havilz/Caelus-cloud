package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type IaCAction string

const (
	ActionCreate IaCAction = "create"
	ActionUpdate IaCAction = "update"
	ActionDelete IaCAction = "delete"
	ActionNoOp   IaCAction = "noop"
)

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

type ResourceType string

const (
	ResourceTypeServer    ResourceType = "server"
	ResourceTypeStorage   ResourceType = "storage"
	ResourceTypeContainer ResourceType = "container"
	ResourceTypeRule      ResourceType = "rule"
)

type ServerSpec struct {
	Name     string            `json:"name" yaml:"name"`
	Provider string            `json:"provider" yaml:"provider"`
	Region   string            `json:"region" yaml:"region"`
	Size     string            `json:"size" yaml:"size"`
	Image    string            `json:"image" yaml:"image"`
	Tags     map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
	SSHKeys  []string          `json:"ssh_keys,omitempty" yaml:"ssh_keys,omitempty"`
}

type StorageSpec struct {
	Name       string `json:"name" yaml:"name"`
	Type       string `json:"type" yaml:"type"`
	Region     string `json:"region,omitempty" yaml:"region,omitempty"`
	Versioning bool   `json:"versioning,omitempty" yaml:"versioning,omitempty"`
	Access     string `json:"access,omitempty" yaml:"access,omitempty"`
}

type ContainerSpec struct {
	Name          string            `json:"name" yaml:"name"`
	Server        string            `json:"server,omitempty" yaml:"server,omitempty"`
	Image         string            `json:"image" yaml:"image"`
	Ports         []string          `json:"ports,omitempty" yaml:"ports,omitempty"`
	Environment   map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Volumes       []string          `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	RestartPolicy string            `json:"restart_policy,omitempty" yaml:"restart_policy,omitempty"`
}

type RuleSpec struct {
	Name      string                 `json:"name" yaml:"name"`
	Trigger   string                 `json:"trigger" yaml:"trigger"`
	Condition map[string]interface{} `json:"condition,omitempty" yaml:"condition,omitempty"`
	Action    map[string]interface{} `json:"action" yaml:"action"`
}

type DeclarativeManifest struct {
	Version    string          `json:"version" yaml:"version"`
	Servers    []ServerSpec    `json:"servers,omitempty" yaml:"servers,omitempty"`
	Storages   []StorageSpec   `json:"storages,omitempty" yaml:"storages,omitempty"`
	Containers []ContainerSpec `json:"containers,omitempty" yaml:"containers,omitempty"`
	Rules      []RuleSpec      `json:"rules,omitempty" yaml:"rules,omitempty"`
}

type IaCChange struct {
	ResourceType  ResourceType           `json:"resource_type"`
	ResourceName  string                 `json:"resource_name"`
	Action        IaCAction              `json:"action"`
	Before        map[string]interface{} `json:"before,omitempty"`
	After         map[string]interface{} `json:"after,omitempty"`
	ChangedFields []string               `json:"changed_fields,omitempty"`
	Reason        string                 `json:"reason,omitempty"`
}

type IaCSummary struct {
	Create int `json:"create"`
	Update int `json:"update"`
	Delete int `json:"delete"`
	NoOp   int `json:"noop"`
	Total  int `json:"total"`
}

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

type IaCValidationError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type IaCValidationResponse struct {
	Valid    bool                 `json:"valid"`
	Errors   []IaCValidationError `json:"errors,omitempty"`
	Manifest *DeclarativeManifest `json:"manifest,omitempty"`
}

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
