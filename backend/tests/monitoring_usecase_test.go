package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/ws"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/monitoring"
)

type mockMetricRepo struct {
	metrics []domain.ServerMetric
}

func (m *mockMetricRepo) Create(_ context.Context, metric *domain.ServerMetric) error {
	m.metrics = append(m.metrics, *metric)
	return nil
}

func (m *mockMetricRepo) GetLatestByServerID(_ context.Context, serverID uuid.UUID) (*domain.ServerMetric, error) {
	for i := len(m.metrics) - 1; i >= 0; i-- {
		if m.metrics[i].ServerID == serverID {
			return &m.metrics[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockMetricRepo) GetHistoryByServerID(_ context.Context, serverID uuid.UUID, _, _ time.Time, _ int) ([]domain.ServerMetric, error) {
	var result []domain.ServerMetric
	for _, metric := range m.metrics {
		if metric.ServerID == serverID {
			result = append(result, metric)
		}
	}
	return result, nil
}

func TestMonitoringUsecase_IngestTelemetry(t *testing.T) {
	serverID := uuid.New()
	orgID := uuid.New()

	mockServerRepo := newMockServerRepo()
	_ = mockServerRepo.Create(context.Background(), &domain.Server{
		ID:             serverID,
		OrganizationID: orgID,
		Name:           "node-alpha",
		Status:         domain.ServerStatusPending,
	})

	mockMetricRepo := &mockMetricRepo{}
	mockAlertRepo := &mockAlertRepo{}
	wsHub := ws.NewHub()
	evaluator := monitoring.NewAlertEvaluator(mockAlertRepo, wsHub)

	uc := monitoring.NewMonitoringUsecase(
		mockMetricRepo,
		mockAlertRepo,
		mockServerRepo,
		evaluator,
		wsHub,
		nil,
		nil,
		nil,
	)

	payload := &domain.TelemetryReportPayload{
		ServerID:  serverID,
		Timestamp: time.Now().UTC(),
		Host: domain.HostMetricsPayload{
			CPUUsagePct:       45.5,
			CPUCores:          2,
			MemoryTotalMB:     2048,
			MemoryUsedMB:      1024,
			MemoryUsagePct:    50.0,
			DiskTotalGB:       40.0,
			DiskUsedGB:        10.0,
			DiskUsagePct:      25.0,
			UptimeSeconds:     3600,
			NetworkInKB:       5000,
			NetworkOutKB:      2500,
			NetworkInRateKBps: 12.5,
		},
		Containers: []domain.ContainerMetricPayload{
			{
				ID:          "cont-1",
				Names:       []string{"/web"},
				Image:       "nginx:latest",
				State:       "running",
				CPUUsagePct: 5.0,
			},
		},
		DockerAvailable: true,
	}

	ctx := context.Background()
	if err := uc.IngestTelemetry(ctx, payload); err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	if len(mockMetricRepo.metrics) != 1 {
		t.Fatalf("expected 1 metric recorded, got %d", len(mockMetricRepo.metrics))
	}

	recorded := mockMetricRepo.metrics[0]
	if recorded.CPUUsagePct != 45.5 {
		t.Errorf("expected CPU 45.5, got %f", recorded.CPUUsagePct)
	}
	if recorded.ContainersCount != 1 {
		t.Errorf("expected 1 container, got %d", recorded.ContainersCount)
	}

	server, _ := mockServerRepo.GetByID(ctx, serverID)
	if server.Status != domain.ServerStatusRunning {
		t.Errorf("expected server status to be running, got %s", server.Status)
	}
}

func TestMonitoringUsecase_GetMetricsHistory(t *testing.T) {
	serverID := uuid.New()
	mockMetricRepo := &mockMetricRepo{
		metrics: []domain.ServerMetric{
			{ServerID: serverID, CPUUsagePct: 10.0, RecordedAt: time.Now().Add(-30 * time.Minute)},
			{ServerID: serverID, CPUUsagePct: 20.0, RecordedAt: time.Now().Add(-15 * time.Minute)},
			{ServerID: serverID, CPUUsagePct: 30.0, RecordedAt: time.Now()},
		},
	}

	uc := monitoring.NewMonitoringUsecase(
		mockMetricRepo,
		&mockAlertRepo{},
		newMockServerRepo(),
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	history, err := uc.GetServerMetricHistory(ctx, serverID, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to query history: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("expected 3 history points, got %d", len(history))
	}
}

func TestMonitoringUsecase_AlertLifecycle(t *testing.T) {
	orgID := uuid.New()
	serverID := uuid.New()
	alertID := uuid.New()
	userID := uuid.New()

	mockAlertRepo := &mockAlertRepo{
		alerts: []domain.Alert{
			{
				ID:             alertID,
				OrganizationID: orgID,
				ServerID:       serverID,
				Title:          "Memory High",
				Status:         domain.AlertStatusActive,
			},
		},
	}

	uc := monitoring.NewMonitoringUsecase(
		&mockMetricRepo{},
		mockAlertRepo,
		newMockServerRepo(),
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()

	if err := uc.AcknowledgeAlert(ctx, alertID, userID); err != nil {
		t.Fatalf("acknowledge failed: %v", err)
	}
	a, _ := mockAlertRepo.GetAlertByID(ctx, alertID)
	if a.Status != domain.AlertStatusAcknowledged {
		t.Errorf("expected status acknowledged, got %s", a.Status)
	}

	if err := uc.ResolveAlert(ctx, alertID, userID); err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	a2, _ := mockAlertRepo.GetAlertByID(ctx, alertID)
	if a2.Status != domain.AlertStatusResolved {
		t.Errorf("expected status resolved, got %s", a2.Status)
	}
}
