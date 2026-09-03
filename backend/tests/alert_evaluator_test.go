package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/monitoring"
)

type mockAlertRepo struct {
	alerts       []domain.Alert
	rules        []domain.AlertRule
	activeAlerts []domain.Alert
}

func (m *mockAlertRepo) CreateAlert(_ context.Context, alert *domain.Alert) error {
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	m.alerts = append(m.alerts, *alert)
	m.activeAlerts = append(m.activeAlerts, *alert)
	return nil
}

func (m *mockAlertRepo) GetAlertByID(_ context.Context, id uuid.UUID) (*domain.Alert, error) {
	for _, a := range m.alerts {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockAlertRepo) ListAlertsByOrg(_ context.Context, _ uuid.UUID, _ *domain.AlertStatus, _, _ int) ([]domain.Alert, int64, error) {
	return m.alerts, int64(len(m.alerts)), nil
}

func (m *mockAlertRepo) ListActiveAlertsByServer(_ context.Context, serverID uuid.UUID) ([]domain.Alert, error) {
	var result []domain.Alert
	for _, a := range m.activeAlerts {
		if a.ServerID == serverID && a.Status == domain.AlertStatusActive {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAlertRepo) UpdateAlertStatus(_ context.Context, id uuid.UUID, status domain.AlertStatus, _ *uuid.UUID, _ *time.Time) error {
	for i, a := range m.alerts {
		if a.ID == id {
			m.alerts[i].Status = status
			return nil
		}
	}
	return domain.ErrNotFound
}

func (m *mockAlertRepo) CreateRule(_ context.Context, rule *domain.AlertRule) error {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	m.rules = append(m.rules, *rule)
	return nil
}

func (m *mockAlertRepo) GetRuleByID(_ context.Context, id uuid.UUID) (*domain.AlertRule, error) {
	for _, r := range m.rules {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockAlertRepo) ListRulesByOrg(_ context.Context, _ uuid.UUID) ([]domain.AlertRule, error) {
	return m.rules, nil
}

func (m *mockAlertRepo) ListRulesForServer(_ context.Context, _, _ uuid.UUID) ([]domain.AlertRule, error) {
	return m.rules, nil
}

func (m *mockAlertRepo) DeleteRule(_ context.Context, id uuid.UUID) error {
	for i, r := range m.rules {
		if r.ID == id {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}

func TestAlertEvaluator_TriggerOnThresholdExceeded(t *testing.T) {
	orgID := uuid.New()
	serverID := uuid.New()
	ruleID := uuid.New()

	mockRepo := &mockAlertRepo{
		rules: []domain.AlertRule{
			{
				ID:             ruleID,
				OrganizationID: orgID,
				Name:           "High CPU Usage",
				MetricName:     "cpu_usage",
				Operator:       ">",
				Threshold:      80.0,
				Severity:       domain.AlertSeverityCritical,
				IsEnabled:      true,
			},
		},
	}

	evaluator := monitoring.NewAlertEvaluator(mockRepo, nil)
	ctx := context.Background()

	server := &domain.Server{
		ID:             serverID,
		OrganizationID: orgID,
		Name:           "web-prod-1",
	}

	m1 := &domain.ServerMetric{
		ServerID:    serverID,
		CPUUsagePct: 75.0,
	}
	if err := evaluator.EvaluateMetrics(ctx, server, m1); err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	if len(mockRepo.alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(mockRepo.alerts))
	}

	m2 := &domain.ServerMetric{
		ServerID:    serverID,
		CPUUsagePct: 92.0,
	}
	if err := evaluator.EvaluateMetrics(ctx, server, m2); err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	if len(mockRepo.alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(mockRepo.alerts))
	}

	createdAlert := mockRepo.alerts[0]
	if createdAlert.Severity != domain.AlertSeverityCritical {
		t.Errorf("expected severity critical, got %s", createdAlert.Severity)
	}
	if createdAlert.Status != domain.AlertStatusActive {
		t.Errorf("expected status active, got %s", createdAlert.Status)
	}

	m3 := &domain.ServerMetric{
		ServerID:    serverID,
		CPUUsagePct: 95.0,
	}
	if err := evaluator.EvaluateMetrics(ctx, server, m3); err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}
	if len(mockRepo.alerts) != 1 {
		t.Errorf("expected still 1 alert (no duplicate spam), got %d", len(mockRepo.alerts))
	}
}

func TestAlertEvaluator_MemoryAndDiskRules(t *testing.T) {
	orgID := uuid.New()
	serverID := uuid.New()

	mockRepo := &mockAlertRepo{
		rules: []domain.AlertRule{
			{
				ID:             uuid.New(),
				OrganizationID: orgID,
				Name:           "RAM Full",
				MetricName:     "memory_usage",
				Operator:       ">=",
				Threshold:      90.0,
				Severity:       domain.AlertSeverityWarning,
				IsEnabled:      true,
			},
			{
				ID:             uuid.New(),
				OrganizationID: orgID,
				Name:           "Disk Full",
				MetricName:     "disk_usage",
				Operator:       ">",
				Threshold:      85.0,
				Severity:       domain.AlertSeverityCritical,
				IsEnabled:      true,
			},
		},
	}

	evaluator := monitoring.NewAlertEvaluator(mockRepo, nil)
	ctx := context.Background()

	server := &domain.Server{
		ID:             serverID,
		OrganizationID: orgID,
		Name:           "db-server-1",
	}

	metric := &domain.ServerMetric{
		ServerID:       serverID,
		MemoryUsagePct: 91.5,
		DiskUsagePct:   89.0,
	}

	if err := evaluator.EvaluateMetrics(ctx, server, metric); err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}

	if len(mockRepo.alerts) != 2 {
		t.Fatalf("expected 2 alerts triggered, got %d", len(mockRepo.alerts))
	}
}
