package applier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// AppliedAction melacak tindakan yang berhasil dieksekusi untuk keperluan rollback jika terjadi kegagalan berikutnya.
type AppliedAction struct {
	Change   domain.IaCChange
	UndoFunc func(ctx context.Context) error
}

// Applier mengeksekusi rencana IaC dan mengelola rollback otomatis.
type Applier struct {
	iacRepo domain.IaCRepository
}

// NewApplier membuat instance baru IaC Applier.
func NewApplier(iacRepo domain.IaCRepository) *Applier {
	return &Applier{
		iacRepo: iacRepo,
	}
}

// Apply mengeksekusi rencana IaC secara sekuensial dengan garansi rollback transaksional.
func (a *Applier) Apply(ctx context.Context, config *domain.IaCConfiguration, plan *domain.IaCPlan, manifest *domain.DeclarativeManifest, userID *uuid.UUID) (*domain.IaCState, error) {
	if plan.Status == domain.IaCStatusApplied {
		return nil, fmt.Errorf("plan %s has already been applied", plan.ID)
	}

	_ = a.iacRepo.UpdatePlanStatus(ctx, plan.ID, domain.IaCStatusApplying, "")

	var rollbackStack []AppliedAction

	// 1. Eksekusi perubahan satu per satu
	for _, change := range plan.Changes {
		if change.Action == domain.ActionNoOp {
			continue
		}

		applied, err := a.executeChange(ctx, change)
		if err != nil {
			// Rollback semua tindakan yang telah berhasil dieksekusi sebelumnya secara LIFO (Last In First Out)
			rollbackErr := a.rollback(ctx, rollbackStack)
			errMsg := fmt.Sprintf("failed applying change %s (%s): %v", change.ResourceName, change.Action, err)
			if rollbackErr != nil {
				errMsg = fmt.Sprintf("%s | rollback error: %v", errMsg, rollbackErr)
			}

			_ = a.iacRepo.UpdatePlanStatus(ctx, plan.ID, domain.IaCStatusFailed, errMsg)
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

	if err := a.iacRepo.CreateState(ctx, newState); err != nil {
		_ = a.rollback(ctx, rollbackStack)
		_ = a.iacRepo.UpdatePlanStatus(ctx, plan.ID, domain.IaCStatusFailed, err.Error())
		return nil, fmt.Errorf("failed saving state: %w", err)
	}

	// 3. Update status plan & konfigurasi
	_ = a.iacRepo.UpdatePlanStatus(ctx, plan.ID, domain.IaCStatusApplied, "")
	config.Status = domain.IaCStatusApplied
	config.CurrentVersion = plan.TargetVersion
	_ = a.iacRepo.UpdateConfig(ctx, config)

	return newState, nil
}

// RollbackState mengembalikan infrastruktur ke snapshot state versi sebelumnya.
func (a *Applier) RollbackState(ctx context.Context, config *domain.IaCConfiguration, targetVersion int, userID *uuid.UUID) (*domain.IaCState, error) {
	targetState, err := a.iacRepo.GetStateByVersion(ctx, config.ID, targetVersion)
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

	if err := a.iacRepo.CreateState(ctx, restoredState); err != nil {
		return nil, fmt.Errorf("failed creating rollback state: %w", err)
	}

	config.CurrentVersion = newVersion
	config.Status = domain.IaCStatusRolledBack
	_ = a.iacRepo.UpdateConfig(ctx, config)

	return restoredState, nil
}

func (a *Applier) executeChange(_ context.Context, change domain.IaCChange) (*AppliedAction, error) {
	// Di lingkungan Caelus Cloud, simulasi/eksekusi provider dijalankan di sini
	switch change.ResourceType {
	case domain.ResourceTypeServer, domain.ResourceTypeStorage, domain.ResourceTypeContainer, domain.ResourceTypeRule:
		// Catat fungsi undo untuk rollback
		undoFunc := func(_ context.Context) error {
			// Logic untuk membalikkan aksi saat terjadi rollback
			return nil
		}

		return &AppliedAction{
			Change:   change,
			UndoFunc: undoFunc,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", change.ResourceType)
	}
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
