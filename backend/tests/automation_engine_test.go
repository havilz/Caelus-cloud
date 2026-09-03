package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/automation"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/notification"
	"github.com/havilz/caelus-cloud/backend/internal/notification/email"
	"github.com/havilz/caelus-cloud/backend/internal/notification/webhook"
	"github.com/havilz/caelus-cloud/backend/internal/queue/mock"
)

type mockAutomationRepo struct {
	rules []domain.AutomationRule
	logs  []domain.RuleExecutionLog
}

func (m *mockAutomationRepo) CreateRule(ctx context.Context, rule *domain.AutomationRule) error {
	m.rules = append(m.rules, *rule)
	return nil
}
func (m *mockAutomationRepo) GetRuleByID(ctx context.Context, orgID, id uuid.UUID) (*domain.AutomationRule, error) {
	for _, r := range m.rules {
		if r.ID == id && r.OrganizationID == orgID {
			return &r, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (m *mockAutomationRepo) ListRules(ctx context.Context, orgID uuid.UUID, page, limit int) ([]domain.AutomationRule, int, error) {
	var result []domain.AutomationRule
	for _, r := range m.rules {
		if r.OrganizationID == orgID {
			result = append(result, r)
		}
	}
	return result, len(result), nil
}
func (m *mockAutomationRepo) ListActiveRulesByTriggerType(ctx context.Context, triggerType domain.RuleTriggerType) ([]domain.AutomationRule, error) {
	var result []domain.AutomationRule
	for _, r := range m.rules {
		if r.IsActive && r.TriggerType == triggerType {
			result = append(result, r)
		}
	}
	return result, nil
}
func (m *mockAutomationRepo) UpdateRule(ctx context.Context, rule *domain.AutomationRule) error {
	for i, r := range m.rules {
		if r.ID == rule.ID {
			m.rules[i] = *rule
			return nil
		}
	}
	return domain.ErrNotFound
}
func (m *mockAutomationRepo) UpdateLastTriggered(ctx context.Context, ruleID uuid.UUID, triggeredAt time.Time) error {
	for i, r := range m.rules {
		if r.ID == ruleID {
			m.rules[i].LastTriggeredAt = &triggeredAt
			return nil
		}
	}
	return nil
}
func (m *mockAutomationRepo) DeleteRule(ctx context.Context, orgID, id uuid.UUID) error {
	for i, r := range m.rules {
		if r.ID == id && r.OrganizationID == orgID {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}
func (m *mockAutomationRepo) CreateExecutionLog(ctx context.Context, log *domain.RuleExecutionLog) error {
	m.logs = append(m.logs, *log)
	return nil
}
func (m *mockAutomationRepo) ListExecutionLogs(ctx context.Context, orgID uuid.UUID, ruleID *uuid.UUID, status *domain.ExecutionStatus, page, limit int) ([]domain.RuleExecutionLog, int, error) {
	return m.logs, len(m.logs), nil
}

func TestRuleEngine_ConditionEvaluationAndActionExecution(t *testing.T) {
	ctx := context.Background()

	var webhookCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := notification.NewUnifiedDispatcher(webhook.NewClient(""), email.NewClient(email.Config{}))
	repo := &mockAutomationRepo{}
	q := mock.NewMockQueueEngine()
	engine := automation.NewEngine(repo, q, notifier, nil, nil)

	orgID := uuid.New()
	rule := domain.AutomationRule{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "Auto-Alert High Memory",
		IsActive:       true,
		TriggerType:    domain.TriggerTypeMetricThreshold,
		Conditions: []domain.RuleCondition{
			{Field: "memory_usage_percent", Operator: domain.OpGreaterThanEqual, Value: 90.0},
			{Field: "status", Operator: domain.OpEqual, Value: "running"},
		},
		Actions: []domain.RuleAction{
			{Type: domain.ActionTypeSendWebhook, Target: server.URL},
			{Type: domain.ActionTypeSendEmail, Target: "devops@caelus.cloud"},
		},
		CooldownSeconds: 60,
	}
	_ = repo.CreateRule(ctx, &rule)

	unmatchedEvent := domain.SystemEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Type:           "metric.memory_report",
		Data: map[string]any{
			"memory_usage_percent": 80.0,
			"status":               "running",
		},
	}
	_ = engine.EvaluateEvent(ctx, unmatchedEvent)
	time.Sleep(50 * time.Millisecond)

	if webhookCalled {
		t.Errorf("expected webhook NOT to be called when conditions are not met")
	}

	matchedEvent := domain.SystemEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Type:           "metric.memory_report",
		Data: map[string]any{
			"memory_usage_percent": 95.0,
			"status":               "running",
		},
	}
	_ = engine.EvaluateEvent(ctx, matchedEvent)
	time.Sleep(100 * time.Millisecond)

	if !webhookCalled {
		t.Errorf("expected webhook to be called when conditions matched")
	}

	if len(repo.logs) == 0 {
		t.Errorf("expected audit execution log to be recorded")
	}
}

func TestCentralEventDispatcher_FanOut(t *testing.T) {
	dispatcher := automation.NewCentralEventDispatcher()
	defer dispatcher.Stop()

	received := make(chan string, 2)
	dispatcher.Subscribe(func(ctx context.Context, event domain.SystemEvent) error {
		received <- event.Type
		return nil
	})

	dispatcher.Publish(context.Background(), domain.SystemEvent{
		Type: "server.crashed",
	})

	select {
	case evtType := <-received:
		if evtType != "server.crashed" {
			t.Errorf("expected 'server.crashed', got '%s'", evtType)
		}
	case <-time.After(500 * time.Millisecond):
		t.Errorf("timeout waiting for event subscriber")
	}
}
