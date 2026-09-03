package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/iac/applier"
	"github.com/havilz/caelus-cloud/backend/internal/iac/parser"
	"github.com/havilz/caelus-cloud/backend/internal/iac/planner"
)

type MockIaCRepo struct {
	configs map[uuid.UUID]*domain.IaCConfiguration
	states  map[uuid.UUID][]*domain.IaCState
	plans   map[uuid.UUID]*domain.IaCPlan
}

func NewMockIaCRepo() *MockIaCRepo {
	return &MockIaCRepo{
		configs: make(map[uuid.UUID]*domain.IaCConfiguration),
		states:  make(map[uuid.UUID][]*domain.IaCState),
		plans:   make(map[uuid.UUID]*domain.IaCPlan),
	}
}

func (m *MockIaCRepo) CreateConfig(ctx context.Context, config *domain.IaCConfiguration) error {
	m.configs[config.ID] = config
	return nil
}

func (m *MockIaCRepo) GetConfigByID(ctx context.Context, id uuid.UUID) (*domain.IaCConfiguration, error) {
	c, ok := m.configs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return c, nil
}

func (m *MockIaCRepo) ListConfigsByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.IaCConfiguration, error) {
	var list []domain.IaCConfiguration
	for _, c := range m.configs {
		if c.OrganizationID == orgID {
			list = append(list, *c)
		}
	}
	return list, nil
}

func (m *MockIaCRepo) UpdateConfig(ctx context.Context, config *domain.IaCConfiguration) error {
	m.configs[config.ID] = config
	return nil
}

func (m *MockIaCRepo) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	delete(m.configs, id)
	return nil
}

func (m *MockIaCRepo) CreateState(ctx context.Context, state *domain.IaCState) error {
	m.states[state.ConfigurationID] = append(m.states[state.ConfigurationID], state)
	return nil
}

func (m *MockIaCRepo) GetLatestStateByConfigID(ctx context.Context, configID uuid.UUID) (*domain.IaCState, error) {
	list := m.states[configID]
	if len(list) == 0 {
		return nil, nil
	}
	return list[len(list)-1], nil
}

func (m *MockIaCRepo) GetStateByVersion(ctx context.Context, configID uuid.UUID, version int) (*domain.IaCState, error) {
	list := m.states[configID]
	for _, s := range list {
		if s.Version == version {
			return s, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockIaCRepo) ListStatesByConfigID(ctx context.Context, configID uuid.UUID) ([]domain.IaCState, error) {
	list := m.states[configID]
	var res []domain.IaCState
	for _, s := range list {
		res = append(res, *s)
	}
	return res, nil
}

func (m *MockIaCRepo) CreatePlan(ctx context.Context, plan *domain.IaCPlan) error {
	m.plans[plan.ID] = plan
	return nil
}

func (m *MockIaCRepo) GetPlanByID(ctx context.Context, id uuid.UUID) (*domain.IaCPlan, error) {
	p, ok := m.plans[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (m *MockIaCRepo) UpdatePlanStatus(ctx context.Context, id uuid.UUID, status domain.IaCStatus, errorMsg string) error {
	p, ok := m.plans[id]
	if ok {
		p.Status = status
		p.ErrorMessage = errorMsg
	}
	return nil
}

func (m *MockIaCRepo) GetLatestPlanByConfigID(ctx context.Context, configID uuid.UUID) (*domain.IaCPlan, error) {
	for _, p := range m.plans {
		if p.ConfigurationID == configID {
			return p, nil
		}
	}
	return nil, nil
}

func TestIaC_YAMLParser(t *testing.T) {
	p := parser.NewParser()

	t.Run("Valid Manifest", func(t *testing.T) {
		yamlDoc := `
version: v1
servers:
  - name: web-api-prod
    provider: aws
    region: us-east-1
    size: t3.micro
    image: ubuntu-22.04
storages:
  - name: prod-backups
    type: s3
    region: us-east-1
containers:
  - name: redis-cache
    image: redis:7-alpine
    ports:
      - "6379:6379"
`
		manifest, errs := p.Parse(yamlDoc)
		if len(errs) > 0 {
			t.Fatalf("unexpected validation errors: %v", errs)
		}
		if manifest == nil {
			t.Fatal("manifest should not be nil")
		}
		if len(manifest.Servers) != 1 || manifest.Servers[0].Name != "web-api-prod" {
			t.Errorf("expected 1 server 'web-api-prod', got %v", manifest.Servers)
		}
		if len(manifest.Storages) != 1 || manifest.Storages[0].Name != "prod-backups" {
			t.Errorf("expected 1 storage 'prod-backups', got %v", manifest.Storages)
		}
		if len(manifest.Containers) != 1 || manifest.Containers[0].Name != "redis-cache" {
			t.Errorf("expected 1 container 'redis-cache', got %v", manifest.Containers)
		}
	})

	t.Run("Invalid Syntax", func(t *testing.T) {
		yamlDoc := `
servers:
  - name: test
    provider: [invalid: yaml: string
`
		_, errs := p.Parse(yamlDoc)
		if len(errs) == 0 {
			t.Fatal("expected syntax error, got none")
		}
	})

	t.Run("Duplicate Resource Name", func(t *testing.T) {
		yamlDoc := `
servers:
  - name: web-server
    provider: aws
  - name: web-server
    provider: hetzner
`
		_, errs := p.Parse(yamlDoc)
		if len(errs) == 0 {
			t.Fatal("expected duplicate name error, got none")
		}
	})
}

func TestIaC_PlanEngine(t *testing.T) {
	eng := planner.NewEngine()
	configID := uuid.New()

	desired := &domain.DeclarativeManifest{
		Version: "v1",
		Servers: []domain.ServerSpec{
			{Name: "app-server-1", Provider: "aws", Region: "us-east-1", Size: "t3.medium"},
			{Name: "app-server-2", Provider: "digitalocean", Region: "nyc1", Size: "s-1vcpu-1gb"},
		},
		Storages: []domain.StorageSpec{
			{Name: "app-assets", Type: "s3", Region: "us-east-1"},
		},
	}

	t.Run("Initial Plan (All Create)", func(t *testing.T) {
		plan, err := eng.GeneratePlan(configID, 1, desired, nil)
		if err != nil {
			t.Fatalf("failed generating plan: %v", err)
		}

		if plan.Summary.Create != 3 {
			t.Errorf("expected 3 creates, got %d", plan.Summary.Create)
		}
		if plan.Summary.Delete != 0 || plan.Summary.Update != 0 {
			t.Errorf("expected 0 delete/update, got %d/%d", plan.Summary.Delete, plan.Summary.Update)
		}
	})

	t.Run("Drift / Update & Delete Plan", func(t *testing.T) {

		currentState := &domain.IaCState{
			ID:              uuid.New(),
			ConfigurationID: configID,
			Version:         1,
			StateData: map[string]interface{}{
				"version": "v1",
				"servers": []domain.ServerSpec{
					{Name: "app-server-1", Provider: "aws", Region: "us-east-1", Size: "t3.micro"},
					{Name: "old-server", Provider: "aws", Region: "us-east-1", Size: "t3.nano"},
				},
				"storages": []domain.StorageSpec{
					{Name: "app-assets", Type: "s3", Region: "us-east-1"},
				},
			},
		}

		plan, err := eng.GeneratePlan(configID, 2, desired, currentState)
		if err != nil {
			t.Fatalf("failed generating plan: %v", err)
		}

		if plan.Summary.Create != 1 {
			t.Errorf("expected 1 create (app-server-2), got %d", plan.Summary.Create)
		}
		if plan.Summary.Update != 1 {
			t.Errorf("expected 1 update (app-server-1), got %d", plan.Summary.Update)
		}
		if plan.Summary.Delete != 1 {
			t.Errorf("expected 1 delete (old-server), got %d", plan.Summary.Delete)
		}
		if plan.Summary.NoOp != 1 {
			t.Errorf("expected 1 noop (app-assets), got %d", plan.Summary.NoOp)
		}
	})
}

func TestIaC_ApplyAndRollback(t *testing.T) {
	ctx := context.Background()
	repo := NewMockIaCRepo()
	app := applier.NewApplier(applier.Dependencies{IaCRepo: repo})

	config := &domain.IaCConfiguration{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Name:           "infra-stack",
		Status:         domain.IaCStatusDraft,
		CurrentVersion: 1,
	}
	_ = repo.CreateConfig(ctx, config)

	manifest := &domain.DeclarativeManifest{
		Version: "v1",
		Servers: []domain.ServerSpec{
			{Name: "test-node", Provider: "aws", Region: "us-east-1", Size: "t3.micro"},
		},
	}

	plan := &domain.IaCPlan{
		ID:              uuid.New(),
		ConfigurationID: config.ID,
		TargetVersion:   1,
		Status:          domain.IaCStatusPlanned,
		Changes: []domain.IaCChange{
			{
				ResourceType: domain.ResourceTypeServer,
				ResourceName: "test-node",
				Action:       domain.ActionCreate,
			},
		},
	}
	_ = repo.CreatePlan(ctx, plan)

	state, err := app.Apply(ctx, config, plan, manifest, nil)
	if err != nil {
		t.Fatalf("failed applying plan: %v", err)
	}
	if state == nil || state.Version != 1 {
		t.Fatalf("expected state version 1, got %v", state)
	}
	if config.Status != domain.IaCStatusApplied {
		t.Errorf("expected config status applied, got %s", config.Status)
	}

	restored, err := app.RollbackState(ctx, config, 1, nil)
	if err != nil {
		t.Fatalf("failed rolling back state: %v", err)
	}
	if restored.Version != 2 {
		t.Errorf("expected restored state version to be incremented to 2, got %d", restored.Version)
	}
	if config.Status != domain.IaCStatusRolledBack {
		t.Errorf("expected config status rolled_back, got %s", config.Status)
	}
}
