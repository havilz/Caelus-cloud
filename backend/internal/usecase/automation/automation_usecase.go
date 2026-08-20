package automation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/automation"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// AutomationUsecase mendefinisikan seluruh kontrak logika bisnis operasional aturan otomasi.
type AutomationUsecase interface {
	// CreateRule membuat aturan otomasi baru untuk organisasi.
	CreateRule(ctx context.Context, input domain.CreateRuleInput) (*domain.AutomationRule, error)

	// GetRule mengambil detail satu aturan otomasi.
	GetRule(ctx context.Context, orgID, ruleID uuid.UUID) (*domain.AutomationRule, error)

	// ListRules mengambil seluruh aturan otomasi milik organisasi.
	ListRules(ctx context.Context, orgID uuid.UUID, page, limit int) ([]domain.AutomationRule, int, error)

	// UpdateRule memperbarui aturan otomasi yang ada.
	UpdateRule(ctx context.Context, orgID, ruleID uuid.UUID, input domain.UpdateRuleInput) (*domain.AutomationRule, error)

	// DeleteRule menghapus aturan otomasi.
	DeleteRule(ctx context.Context, orgID, ruleID uuid.UUID) error

	// TestRule melakukan simulasi uji coba eksekusi aturan otomasi secara manual.
	TestRule(ctx context.Context, orgID, ruleID uuid.UUID, mockData map[string]any) (*domain.RuleExecutionLog, error)

	// ListLogs mengambil riwayat log audit eksekusi aturan.
	ListLogs(ctx context.Context, orgID uuid.UUID, ruleID *uuid.UUID, status *domain.ExecutionStatus, page, limit int) ([]domain.RuleExecutionLog, int, error)
}

// usecase mengimplementasikan AutomationUsecase.
type usecase struct {
	repo       domain.AutomationRepository
	engine     automation.RuleEngine
	dispatcher automation.EventDispatcher
}

// NewAutomationUsecase membuat instance baru usecase otomasi.
// Parameter repo merupakan repositori PostgreSQL otomasi.
// Parameter engine merupakan mesin evaluasi aturan.
// Parameter dispatcher merupakan penyebar kejadian sistem.
// Mengembalikan pointer instance usecase.
func NewAutomationUsecase(
	repo domain.AutomationRepository,
	engine automation.RuleEngine,
	dispatcher automation.EventDispatcher,
) AutomationUsecase {
	return &usecase{
		repo:       repo,
		engine:     engine,
		dispatcher: dispatcher,
	}
}

// CreateRule memvalidasi dan membuat aturan otomasi baru.
func (u *usecase) CreateRule(ctx context.Context, input domain.CreateRuleInput) (*domain.AutomationRule, error) {
	if input.OrganizationID == uuid.Nil {
		return nil, errors.New("organization_id is required")
	}
	if input.Name == "" {
		return nil, errors.New("rule name is required")
	}
	if input.TriggerType == "" {
		return nil, errors.New("trigger_type is required")
	}
	if len(input.Actions) == 0 {
		return nil, errors.New("at least one action is required")
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	cooldown := input.CooldownSeconds
	if cooldown <= 0 {
		cooldown = 300 // default 5 menit
	}

	rule := &domain.AutomationRule{
		ID:              uuid.New(),
		OrganizationID:  input.OrganizationID,
		Name:            input.Name,
		Description:     input.Description,
		IsActive:        isActive,
		TriggerType:     input.TriggerType,
		TriggerConfig:   input.TriggerConfig,
		Conditions:      input.Conditions,
		Actions:         input.Actions,
		CooldownSeconds: cooldown,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if err := u.repo.CreateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create automation rule: %w", err)
	}

	return rule, nil
}

// GetRule mengambil detail satu aturan otomasi.
func (u *usecase) GetRule(ctx context.Context, orgID, ruleID uuid.UUID) (*domain.AutomationRule, error) {
	return u.repo.GetRuleByID(ctx, orgID, ruleID)
}

// ListRules mengambil daftar aturan otomasi terpaginasi.
func (u *usecase) ListRules(ctx context.Context, orgID uuid.UUID, page, limit int) ([]domain.AutomationRule, int, error) {
	return u.repo.ListRules(ctx, orgID, page, limit)
}

// UpdateRule memperbarui data aturan otomasi yang ada.
func (u *usecase) UpdateRule(ctx context.Context, orgID, ruleID uuid.UUID, input domain.UpdateRuleInput) (*domain.AutomationRule, error) {
	rule, err := u.repo.GetRuleByID(ctx, orgID, ruleID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil && *input.Name != "" {
		rule.Name = *input.Name
	}
	if input.Description != nil {
		rule.Description = *input.Description
	}
	if input.IsActive != nil {
		rule.IsActive = *input.IsActive
	}
	if input.TriggerType != nil && *input.TriggerType != "" {
		rule.TriggerType = *input.TriggerType
	}
	if len(input.TriggerConfig) > 0 {
		rule.TriggerConfig = input.TriggerConfig
	}
	if input.Conditions != nil {
		rule.Conditions = input.Conditions
	}
	if input.Actions != nil {
		rule.Actions = input.Actions
	}
	if input.CooldownSeconds != nil && *input.CooldownSeconds >= 0 {
		rule.CooldownSeconds = *input.CooldownSeconds
	}

	if err := u.repo.UpdateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update automation rule: %w", err)
	}

	return rule, nil
}

// DeleteRule menghapus aturan otomasi.
func (u *usecase) DeleteRule(ctx context.Context, orgID, ruleID uuid.UUID) error {
	return u.repo.DeleteRule(ctx, orgID, ruleID)
}

// TestRule melakukan simulasi uji coba eksekusi aturan otomasi secara manual.
func (u *usecase) TestRule(ctx context.Context, orgID, ruleID uuid.UUID, mockData map[string]any) (*domain.RuleExecutionLog, error) {
	rule, err := u.repo.GetRuleByID(ctx, orgID, ruleID)
	if err != nil {
		return nil, err
	}

	if mockData == nil {
		mockData = map[string]any{
			"cpu_usage_percent":    95.5,
			"memory_usage_percent": 88.0,
			"status":               "offline",
			"server_id":            uuid.New().String(),
		}
	}

	event := domain.SystemEvent{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Type:           string(rule.TriggerType),
		SourceResource: "manual_test_run",
		Data:           mockData,
		OccurredAt:     time.Now().UTC(),
	}

	startTime := time.Now()
	actionResults := make([]domain.ActionResultItem, 0, len(rule.Actions))
	for _, action := range rule.Actions {
		res := u.engine.ExecuteRuleAction(ctx, rule, action, mockData)
		actionResults = append(actionResults, res)
	}

	log := &domain.RuleExecutionLog{
		ID:                  uuid.New(),
		RuleID:              rule.ID,
		OrganizationID:      rule.OrganizationID,
		RuleName:            rule.Name,
		TriggerEvent:        event.Type,
		Status:              domain.ExecutionStatusSuccess,
		EvaluatedConditions: []byte(`{"manual_test": true}`),
		ExecutedActions:     actionResults,
		ExecutionDurationMs: int(time.Since(startTime).Milliseconds()),
		ExecutedAt:          time.Now().UTC(),
	}

	_ = u.repo.CreateExecutionLog(ctx, log)
	return log, nil
}

// ListLogs mengambil daftar catatan log riwayat eksekusi otomasi.
func (u *usecase) ListLogs(
	ctx context.Context,
	orgID uuid.UUID,
	ruleID *uuid.UUID,
	status *domain.ExecutionStatus,
	page, limit int,
) ([]domain.RuleExecutionLog, int, error) {
	return u.repo.ListExecutionLogs(ctx, orgID, ruleID, status, page, limit)
}
