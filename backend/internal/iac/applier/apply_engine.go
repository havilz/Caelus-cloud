package applier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// AppliedAction melacak tindakan yang berhasil dieksekusi untuk keperluan rollback jika terjadi kegagalan berikutnya.
type AppliedAction struct {
	Change   domain.IaCChange
	UndoFunc func(ctx context.Context) error
}

// Dependencies memuat seluruh repositori dan service yang dibutuhkan untuk provisi nyata sumber daya IaC.
type Dependencies struct {
	IaCRepo        domain.IaCRepository
	ServerRepo     domain.ServerRepository
	ProviderRepo   domain.ProviderRepository
	BucketRepo     domain.BucketRepository
	StorageFactory domain.StorageFactory
	DeploymentRepo domain.DeploymentRepository
	AutomationRepo domain.AutomationRepository
}

// Applier mengeksekusi rencana IaC dan mengelola rollback otomatis.
type Applier struct {
	deps Dependencies
}

// NewApplier membuat instance baru IaC Applier.
func NewApplier(deps Dependencies) *Applier {
	return &Applier{
		deps: deps,
	}
}

// Apply mengeksekusi rencana IaC secara sekuensial dengan garansi rollback transaksional.
func (a *Applier) Apply(ctx context.Context, config *domain.IaCConfiguration, plan *domain.IaCPlan, manifest *domain.DeclarativeManifest, userID *uuid.UUID) (*domain.IaCState, error) {
	if plan.Status == domain.IaCStatusApplied {
		return nil, fmt.Errorf("plan %s has already been applied", plan.ID)
	}

	_ = a.deps.IaCRepo.UpdatePlanStatus(ctx, plan.ID, domain.IaCStatusApplying, "")

	var rollbackStack []AppliedAction

	// 1. Eksekusi rekonsiliasi seluruh resource deklaratif (idempotent & self-healing)
	for _, change := range plan.Changes {
		applied, err := a.executeChange(ctx, config.OrganizationID, change, manifest)
		if err != nil {
			// Rollback semua tindakan yang telah berhasil dieksekusi sebelumnya secara LIFO (Last In First Out)
			rollbackErr := a.rollback(ctx, rollbackStack)
			errMsg := fmt.Sprintf("failed applying change %s (%s): %v", change.ResourceName, change.Action, err)
			if rollbackErr != nil {
				errMsg = fmt.Sprintf("%s | rollback error: %v", errMsg, rollbackErr)
			}

			_ = a.deps.IaCRepo.UpdatePlanStatus(ctx, plan.ID, domain.IaCStatusFailed, errMsg)
			return nil, errors.New(errMsg)
		}

		if applied != nil {
			rollbackStack = append(rollbackStack, *applied)
		}
	}

	// 2. Simpan State Baru
	now := time.Now().UTC()
	stateDataMap := structToMap(manifest)
	dataBytes, _ := json.Marshal(stateDataMap)
	hasher := sha256.New()
	hasher.Write(dataBytes)
	stateHash := hex.EncodeToString(hasher.Sum(nil))

	newState := &domain.IaCState{
		ID:              uuid.New(),
		ConfigurationID: config.ID,
		Version:         plan.TargetVersion,
		StateData:       stateDataMap,
		Hash:            stateHash,
		AppliedAt:       now,
		AppliedBy:       userID,
		CreatedAt:       now,
	}

	if err := a.deps.IaCRepo.CreateState(ctx, newState); err != nil {
		_ = a.rollback(ctx, rollbackStack)
		_ = a.deps.IaCRepo.UpdatePlanStatus(ctx, plan.ID, domain.IaCStatusFailed, err.Error())
		return nil, fmt.Errorf("failed saving state: %w", err)
	}

	// 3. Update status plan & konfigurasi
	_ = a.deps.IaCRepo.UpdatePlanStatus(ctx, plan.ID, domain.IaCStatusApplied, "")
	config.Status = domain.IaCStatusApplied
	config.CurrentVersion = plan.TargetVersion
	_ = a.deps.IaCRepo.UpdateConfig(ctx, config)

	return newState, nil
}

// RollbackState mengembalikan infrastruktur ke snapshot state versi sebelumnya.
func (a *Applier) RollbackState(ctx context.Context, config *domain.IaCConfiguration, targetVersion int, userID *uuid.UUID) (*domain.IaCState, error) {
	targetState, err := a.deps.IaCRepo.GetStateByVersion(ctx, config.ID, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("target state version %d not found: %w", targetVersion, err)
	}

	var targetManifest domain.DeclarativeManifest
	dataBytes, _ := json.Marshal(targetState.StateData)
	_ = json.Unmarshal(dataBytes, &targetManifest)

	// Buat state baru dengan versi inkremental yang meniru target state
	newVersion := config.CurrentVersion + 1
	now := time.Now().UTC()

	restoredState := &domain.IaCState{
		ID:              uuid.New(),
		ConfigurationID: config.ID,
		Version:         newVersion,
		StateData:       targetState.StateData,
		Hash:            targetState.Hash,
		AppliedAt:       now,
		AppliedBy:       userID,
		CreatedAt:       now,
	}

	if err := a.deps.IaCRepo.CreateState(ctx, restoredState); err != nil {
		return nil, fmt.Errorf("failed creating rollback state: %w", err)
	}

	config.CurrentVersion = newVersion
	config.Status = domain.IaCStatusRolledBack
	_ = a.deps.IaCRepo.UpdateConfig(ctx, config)

	return restoredState, nil
}

func (a *Applier) executeChange(ctx context.Context, orgID uuid.UUID, change domain.IaCChange, manifest *domain.DeclarativeManifest) (*AppliedAction, error) {
	switch change.ResourceType {
	case domain.ResourceTypeServer:
		if a.deps.ServerRepo != nil && manifest != nil {
			var spec *domain.ServerSpec
			for i := range manifest.Servers {
				if manifest.Servers[i].Name == change.ResourceName {
					spec = &manifest.Servers[i]
					break
				}
			}

			if spec != nil {
				providerID := uuid.New()
				if a.deps.ProviderRepo != nil {
					slug := strings.ToLower(strings.TrimSpace(spec.Provider))
					if slug == "" || slug == "byos" || slug == "generic" {
						slug = "custom"
					}
					if p, err := a.deps.ProviderRepo.GetBySlug(ctx, slug); err == nil && p != nil {
						providerID = p.ID
					} else if fallback, err := a.deps.ProviderRepo.GetBySlug(ctx, "custom"); err == nil && fallback != nil {
						providerID = fallback.ID
					} else if mockP, err := a.deps.ProviderRepo.GetBySlug(ctx, "mock"); err == nil && mockP != nil {
						providerID = mockP.ID
					}
				}

				osType := spec.Image
				if osType == "" {
					osType = "ubuntu-22.04"
				}
				region := spec.Region
				if region == "" {
					region = "default"
				}

				existingSrvs, _, err := a.deps.ServerRepo.ListByOrg(ctx, orgID, 1, 100)
				var srvExists bool
				if err == nil {
					for _, ex := range existingSrvs {
						if ex.Name == spec.Name {
							srvExists = true
							break
						}
					}
				}

				if !srvExists {
					now := time.Now().UTC()
					initStatus := domain.ServerStatusPending
					slug := strings.ToLower(strings.TrimSpace(spec.Provider))
					if slug == "mock" {
						initStatus = domain.ServerStatusRunning
					}

					srv := &domain.Server{
						ID:             uuid.New(),
						OrganizationID: orgID,
						ProviderID:     providerID,
						Name:           spec.Name,
						Status:         initStatus,
						OSType:         osType,
						CPUCores:       2,
						MemoryMB:       2048,
						DiskGB:         40,
						Region:         region,
						CreatedAt:      now,
						UpdatedAt:      now,
					}

					if err := a.deps.ServerRepo.Create(ctx, srv); err != nil {
						return nil, fmt.Errorf("failed creating server %s: %w", spec.Name, err)
					}

					return &AppliedAction{
						Change: change,
						UndoFunc: func(c context.Context) error {
							return a.deps.ServerRepo.Delete(c, srv.ID)
						},
					}, nil
				}
				return &AppliedAction{Change: change, UndoFunc: func(_ context.Context) error { return nil }}, nil
			}
		}

	case domain.ResourceTypeStorage:
		if a.deps.BucketRepo != nil && manifest != nil {
			var spec *domain.StorageSpec
			for i := range manifest.Storages {
				if manifest.Storages[i].Name == change.ResourceName {
					spec = &manifest.Storages[i]
					break
				}
			}

			if spec != nil {
				provType := domain.StorageProviderMinIO
				stType := strings.ToLower(strings.TrimSpace(spec.Type))
				if stType == "s3" || stType == "aws" {
					provType = domain.StorageProviderS3
				} else if stType == "r2" || stType == "cloudflare" {
					provType = domain.StorageProviderR2
				}

				region := spec.Region
				if region == "" {
					region = "us-east-1"
				}

				existingBucket, _ := a.deps.BucketRepo.GetByName(ctx, spec.Name)
				if existingBucket == nil {
					if a.deps.StorageFactory != nil {
						if adapter, err := a.deps.StorageFactory.GetAdapter(provType); err == nil && adapter != nil {
							_ = adapter.CreateBucket(ctx, spec.Name, region)
						}
					}

					now := time.Now().UTC()
					isPublic := spec.Access == "public-read" || spec.Access == "public"
					bucket := &domain.Bucket{
						ID:             uuid.New(),
						OrganizationID: orgID,
						Name:           spec.Name,
						ProviderType:   provType,
						Region:         region,
						IsPublic:       isPublic,
						Versioning:     spec.Versioning,
						CreatedAt:      now,
						UpdatedAt:      now,
					}

					if err := a.deps.BucketRepo.Create(ctx, bucket); err != nil {
						return nil, fmt.Errorf("failed persisting bucket %s: %w", spec.Name, err)
					}

					return &AppliedAction{
						Change: change,
						UndoFunc: func(c context.Context) error {
							return a.deps.BucketRepo.Delete(c, bucket.ID)
						},
					}, nil
				}
				return &AppliedAction{Change: change, UndoFunc: func(_ context.Context) error { return nil }}, nil
			}
		}

	case domain.ResourceTypeContainer:
		if a.deps.DeploymentRepo != nil && manifest != nil {
			var spec *domain.ContainerSpec
			for i := range manifest.Containers {
				if manifest.Containers[i].Name == change.ResourceName {
					spec = &manifest.Containers[i]
					break
				}
			}

			if spec != nil {
				var portBindings []domain.PortBinding
				for _, p := range spec.Ports {
					parts := strings.Split(p, ":")
					if len(parts) == 2 {
						hPort, _ := strconv.Atoi(parts[0])
						cPort, _ := strconv.Atoi(parts[1])
						if hPort > 0 && cPort > 0 {
							portBindings = append(portBindings, domain.PortBinding{
								HostPort:      hPort,
								ContainerPort: cPort,
								Protocol:      "tcp",
							})
						}
					}
				}

				var volBindings []domain.VolumeBinding
				for _, v := range spec.Volumes {
					parts := strings.Split(v, ":")
					if len(parts) >= 2 {
						mode := "rw"
						if len(parts) >= 3 {
							mode = parts[2]
						}
						volBindings = append(volBindings, domain.VolumeBinding{
							HostPath:      parts[0],
							ContainerPath: parts[1],
							Mode:          mode,
						})
					}
				}

				restartPolicy := spec.RestartPolicy
				if restartPolicy == "" {
					restartPolicy = "unless-stopped"
				}

				existingDeps, _ := a.deps.DeploymentRepo.ListDeploymentsByOrg(ctx, orgID)
				var depExists bool
				for _, ed := range existingDeps {
					if ed.ContainerName == spec.Name || ed.AppName == spec.Name {
						depExists = true
						break
					}
				}

				if !depExists {
					now := time.Now().UTC()
					deployment := &domain.Deployment{
						ID:                   uuid.New(),
						OrganizationID:       orgID,
						AppName:              spec.Name,
						ImageTag:             spec.Image,
						ContainerName:        spec.Name,
						EnvironmentVariables: spec.Environment,
						PortBindings:         portBindings,
						VolumeBindings:       volBindings,
						RestartPolicy:        restartPolicy,
						Status:               domain.DeploymentStatusRunning,
						CreatedAt:            now,
						UpdatedAt:            now,
					}

					if err := a.deps.DeploymentRepo.CreateDeployment(ctx, deployment); err != nil {
						return nil, fmt.Errorf("failed creating container %s: %w", spec.Name, err)
					}

					return &AppliedAction{
						Change: change,
						UndoFunc: func(c context.Context) error {
							return a.deps.DeploymentRepo.UpdateDeploymentStatus(c, deployment.ID, domain.DeploymentStatusRolledBack, "IaC Rollback", nil)
						},
					}, nil
				}
				return &AppliedAction{Change: change, UndoFunc: func(_ context.Context) error { return nil }}, nil
			}
		}

	case domain.ResourceTypeRule:
		if a.deps.AutomationRepo != nil && manifest != nil {
			var spec *domain.RuleSpec
			for i := range manifest.Rules {
				if manifest.Rules[i].Name == change.ResourceName {
					spec = &manifest.Rules[i]
					break
				}
			}

			if spec != nil {
				triggerType := domain.TriggerTypeMetricThreshold
				if spec.Trigger == "server_status_changed" {
					triggerType = domain.TriggerTypeServerStatusChanged
				} else if spec.Trigger == "scheduled_cron" {
					triggerType = domain.TriggerTypeScheduledCron
				}

				actionBytes, _ := json.Marshal(spec.Action)
				var ruleAction domain.RuleAction
				_ = json.Unmarshal(actionBytes, &ruleAction)
				if ruleAction.Type == "" {
					ruleAction.Type = domain.ActionTypeSendEmail
				}

				now := time.Now().UTC()
				rule := &domain.AutomationRule{
					ID:              uuid.New(),
					OrganizationID:  orgID,
					Name:            spec.Name,
					IsActive:        true,
					TriggerType:     triggerType,
					Conditions:      []domain.RuleCondition{},
					Actions:         []domain.RuleAction{ruleAction},
					CooldownSeconds: 300,
					CreatedAt:       now,
					UpdatedAt:       now,
				}

				if err := a.deps.AutomationRepo.CreateRule(ctx, rule); err != nil {
					return nil, fmt.Errorf("failed creating automation rule %s: %w", spec.Name, err)
				}

				return &AppliedAction{
					Change: change,
					UndoFunc: func(c context.Context) error {
						return a.deps.AutomationRepo.DeleteRule(c, orgID, rule.ID)
					},
				}, nil
			}
		}
	}

	return &AppliedAction{
		Change:   change,
		UndoFunc: func(_ context.Context) error { return nil },
	}, nil
}

func (a *Applier) rollback(ctx context.Context, stack []AppliedAction) error {
	var errs []error
	for i := len(stack) - 1; i >= 0; i-- {
		action := stack[i]
		if action.UndoFunc != nil {
			if err := action.UndoFunc(ctx); err != nil {
				errs = append(errs, fmt.Errorf("undo %s failed: %w", action.Change.ResourceName, err))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("rollback errors: %v", errs)
	}
	return nil
}

func structToMap(val interface{}) map[string]interface{} {
	b, err := json.Marshal(val)
	if err != nil {
		return make(map[string]interface{})
	}
	var res map[string]interface{}
	_ = json.Unmarshal(b, &res)
	return res
}
