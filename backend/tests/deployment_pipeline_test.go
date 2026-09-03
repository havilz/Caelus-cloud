package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/orchestration"
)

type MockDeploymentRepo struct {
	deployments map[uuid.UUID]*domain.Deployment
	logs        map[uuid.UUID][]domain.DeploymentLog
}

func NewMockDeploymentRepo() *MockDeploymentRepo {
	return &MockDeploymentRepo{
		deployments: make(map[uuid.UUID]*domain.Deployment),
		logs:        make(map[uuid.UUID][]domain.DeploymentLog),
	}
}

func (m *MockDeploymentRepo) CreateDeployment(ctx context.Context, dep *domain.Deployment) error {
	m.deployments[dep.ID] = dep
	return nil
}

func (m *MockDeploymentRepo) GetDeploymentByID(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	d, ok := m.deployments[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return d, nil
}

func (m *MockDeploymentRepo) ListDeploymentsByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.Deployment, error) {
	var list []domain.Deployment
	for _, d := range m.deployments {
		if d.OrganizationID == orgID {
			list = append(list, *d)
		}
	}
	return list, nil
}

func (m *MockDeploymentRepo) ListDeploymentsByServer(ctx context.Context, serverID uuid.UUID) ([]domain.Deployment, error) {
	var list []domain.Deployment
	for _, d := range m.deployments {
		if d.ServerID != nil && *d.ServerID == serverID {
			list = append(list, *d)
		}
	}
	return list, nil
}

func (m *MockDeploymentRepo) UpdateDeploymentStatus(ctx context.Context, id uuid.UUID, status domain.DeploymentStatus, errorMsg string, finishedAt *time.Time) error {
	d, ok := m.deployments[id]
	if ok {
		d.Status = status
		d.ErrorMessage = errorMsg
		d.FinishedAt = finishedAt
	}
	return nil
}

func (m *MockDeploymentRepo) DeleteDeployment(ctx context.Context, id uuid.UUID) error {
	delete(m.deployments, id)
	delete(m.logs, id)
	return nil
}

func (m *MockDeploymentRepo) AppendLog(ctx context.Context, log *domain.DeploymentLog) error {
	m.logs[log.DeploymentID] = append(m.logs[log.DeploymentID], *log)
	return nil
}

func (m *MockDeploymentRepo) GetLogsByDeployment(ctx context.Context, deploymentID uuid.UUID, limit int) ([]domain.DeploymentLog, error) {
	return m.logs[deploymentID], nil
}

func TestDeployment_PipelineExecution(t *testing.T) {
	ctx := context.Background()
	repo := NewMockDeploymentRepo()
	uc := orchestration.NewUseCase(repo, nil)

	orgID := uuid.New()
	req := domain.DeploymentRequest{
		AppName:       "backend-api",
		ImageTag:      "mock:latest",
		ContainerName: "caelus-nginx",
		PortBindings: []domain.PortBinding{
			{HostPort: 80, ContainerPort: 80, Protocol: "tcp"},
		},
		RestartPolicy: "always",
	}

	dep, err := uc.CreateDeployment(ctx, orgID, req)
	if err != nil {
		t.Fatalf("failed creating deployment: %v", err)
	}

	if dep == nil || dep.ID == uuid.Nil {
		t.Fatal("deployment ID is empty")
	}

	var updated *domain.Deployment
	for i := 0; i < 40; i++ {
		time.Sleep(250 * time.Millisecond)
		updated, _ = uc.GetDeployment(ctx, dep.ID)
		if updated != nil && (updated.Status == domain.DeploymentStatusRunning || updated.Status == domain.DeploymentStatusFailed) {
			break
		}
	}

	if updated == nil {
		t.Fatal("failed retrieving deployment")
	}

	if updated.Status != domain.DeploymentStatusRunning {
		t.Errorf("expected deployment to be Running, got %s (error: %s)", updated.Status, updated.ErrorMessage)
	}

	logs, err := uc.GetLogs(ctx, dep.ID, 100)
	if err != nil {
		t.Fatalf("failed retrieving deployment logs: %v", err)
	}

	if len(logs) == 0 {
		t.Error("expected deployment logs to be recorded, got 0 logs")
	}

	if err := uc.StopDeployment(ctx, dep.ID); err != nil {
		t.Fatalf("failed stopping deployment: %v", err)
	}
	stopped, _ := uc.GetDeployment(ctx, dep.ID)
	if stopped.Status != domain.DeploymentStatusStopped {
		t.Errorf("expected deployment status Stopped, got %s", stopped.Status)
	}

	rolledBackDep, err := uc.RollbackDeployment(ctx, dep.ID)
	if err != nil {
		t.Fatalf("failed triggering rollback deployment: %v", err)
	}
	if rolledBackDep == nil || rolledBackDep.ID == dep.ID {
		t.Errorf("expected new deployment instance from rollback, got %v", rolledBackDep)
	}
}
