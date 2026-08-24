package iac

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/iac/applier"
	"github.com/havilz/caelus-cloud/backend/internal/iac/parser"
	"github.com/havilz/caelus-cloud/backend/internal/iac/planner"
)

type UseCase struct {
	repo    domain.IaCRepository
	parser  *parser.Parser
	planner *planner.Engine
	applier *applier.Applier
}

func NewUseCase(repo domain.IaCRepository) *UseCase {
	return NewUseCaseWithDeps(applier.Dependencies{
		IaCRepo: repo,
	})
}

func NewUseCaseWithDeps(deps applier.Dependencies) *UseCase {
	p := parser.NewParser()
	pl := planner.NewEngine()
	ap := applier.NewApplier(deps)
	return &UseCase{
		repo:    deps.IaCRepo,
		parser:  p,
		planner: pl,
		applier: ap,
	}
}

func (u *UseCase) ValidateYAML(rawYAML string) domain.IaCValidationResponse {
	manifest, errs := u.parser.Parse(rawYAML)
	if len(errs) > 0 {
		return domain.IaCValidationResponse{
			Valid:  false,
			Errors: errs,
		}
	}
	return domain.IaCValidationResponse{
		Valid:    true,
		Manifest: manifest,
	}
}

func (u *UseCase) CreateConfig(ctx context.Context, orgID uuid.UUID, name, description, rawYAML string) (*domain.IaCConfiguration, error) {
	// Parse first to ensure basic validity
	_, errs := u.parser.Parse(rawYAML)
	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid YAML syntax: %s", errs[0].Message)
	}

	config := &domain.IaCConfiguration{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
		RawYAML:        rawYAML,
		Status:         domain.IaCStatusDraft,
		CurrentVersion: 1,
	}

	if err := u.repo.CreateConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("failed creating iac configuration: %w", err)
	}

	return config, nil
}

func (u *UseCase) GetConfig(ctx context.Context, id uuid.UUID) (*domain.IaCConfiguration, error) {
	return u.repo.GetConfigByID(ctx, id)
}

func (u *UseCase) ListConfigs(ctx context.Context, orgID uuid.UUID) ([]domain.IaCConfiguration, error) {
	return u.repo.ListConfigsByOrg(ctx, orgID)
}

func (u *UseCase) UpdateConfig(ctx context.Context, id uuid.UUID, name, description, rawYAML string) (*domain.IaCConfiguration, error) {
	config, err := u.repo.GetConfigByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rawYAML != "" {
		_, errs := u.parser.Parse(rawYAML)
		if len(errs) > 0 {
			return nil, fmt.Errorf("invalid YAML: %s", errs[0].Message)
		}
		config.RawYAML = rawYAML
	}
	if name != "" {
		config.Name = name
	}
	config.Description = description

	if err := u.repo.UpdateConfig(ctx, config); err != nil {
		return nil, fmt.Errorf("failed updating iac configuration: %w", err)
	}

	return config, nil
}

func (u *UseCase) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	return u.repo.DeleteConfig(ctx, id)
}

func (u *UseCase) GeneratePlan(ctx context.Context, configID uuid.UUID) (*domain.IaCPlan, error) {
	config, err := u.repo.GetConfigByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed getting config: %w", err)
	}

	manifest, errs := u.parser.Parse(config.RawYAML)
	if len(errs) > 0 {
		return nil, fmt.Errorf("cannot plan: YAML is invalid (%s)", errs[0].Message)
	}

	latestState, err := u.repo.GetLatestStateByConfigID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed checking latest state: %w", err)
	}

	targetVersion := 1
	if latestState != nil {
		targetVersion = latestState.Version + 1
	}

	plan, err := u.planner.GeneratePlan(configID, targetVersion, manifest, latestState)
	if err != nil {
		return nil, fmt.Errorf("failed generating plan: %w", err)
	}

	if err := u.repo.CreatePlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed saving plan: %w", err)
	}

	config.Status = domain.IaCStatusPlanned
	_ = u.repo.UpdateConfig(ctx, config)

	return plan, nil
}

func (u *UseCase) ApplyPlan(ctx context.Context, planID uuid.UUID, userID *uuid.UUID) (*domain.IaCState, error) {
	plan, err := u.repo.GetPlanByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	config, err := u.repo.GetConfigByID(ctx, plan.ConfigurationID)
	if err != nil {
		return nil, fmt.Errorf("config not found: %w", err)
	}

	manifest, errs := u.parser.Parse(config.RawYAML)
	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid YAML syntax during apply: %s", errs[0].Message)
	}

	state, err := u.applier.Apply(ctx, config, plan, manifest, userID)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (u *UseCase) RollbackState(ctx context.Context, configID uuid.UUID, targetVersion int, userID *uuid.UUID) (*domain.IaCState, error) {
	config, err := u.repo.GetConfigByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("config not found: %w", err)
	}

	return u.applier.RollbackState(ctx, config, targetVersion, userID)
}

func (u *UseCase) ListStates(ctx context.Context, configID uuid.UUID) ([]domain.IaCState, error) {
	return u.repo.ListStatesByConfigID(ctx, configID)
}

func (u *UseCase) GetLatestPlan(ctx context.Context, configID uuid.UUID) (*domain.IaCPlan, error) {
	return u.repo.GetLatestPlanByConfigID(ctx, configID)
}
