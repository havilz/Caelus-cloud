package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AutomationRepository struct {
	pool *pgxpool.Pool
}

func NewAutomationRepository(pool *pgxpool.Pool) *AutomationRepository {
	return &AutomationRepository{pool: pool}
}

func (r *AutomationRepository) CreateRule(ctx context.Context, rule *domain.AutomationRule) error {
	query := `
		INSERT INTO automation_rules (
			id, organization_id, name, description, is_active,
			trigger_type, trigger_config, conditions, actions,
			cooldown_seconds, last_triggered_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13
		);
	`

	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	now := time.Now().UTC()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	triggerConfigJSON := rule.TriggerConfig
	if len(triggerConfigJSON) == 0 {
		triggerConfigJSON = []byte("{}")
	}

	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal rule conditions: %w", err)
	}

	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal rule actions: %w", err)
	}

	_, err = r.pool.Exec(ctx, query,
		rule.ID,
		rule.OrganizationID,
		rule.Name,
		rule.Description,
		rule.IsActive,
		rule.TriggerType,
		triggerConfigJSON,
		conditionsJSON,
		actionsJSON,
		rule.CooldownSeconds,
		rule.LastTriggeredAt,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert automation rule: %w", err)
	}

	return nil
}

func (r *AutomationRepository) GetRuleByID(ctx context.Context, orgID, id uuid.UUID) (*domain.AutomationRule, error) {
	query := `
		SELECT
			id, organization_id, name, description, is_active,
			trigger_type, trigger_config, conditions, actions,
			cooldown_seconds, last_triggered_at, created_at, updated_at
		FROM automation_rules
		WHERE id = $1 AND organization_id = $2;
	`

	var rule domain.AutomationRule
	var conditionsJSON, actionsJSON []byte

	err := r.pool.QueryRow(ctx, query, id, orgID).Scan(
		&rule.ID,
		&rule.OrganizationID,
		&rule.Name,
		&rule.Description,
		&rule.IsActive,
		&rule.TriggerType,
		&rule.TriggerConfig,
		&conditionsJSON,
		&actionsJSON,
		&rule.CooldownSeconds,
		&rule.LastTriggeredAt,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query automation rule: %w", err)
	}

	if len(conditionsJSON) > 0 {
		_ = json.Unmarshal(conditionsJSON, &rule.Conditions)
	}
	if len(actionsJSON) > 0 {
		_ = json.Unmarshal(actionsJSON, &rule.Actions)
	}

	return &rule, nil
}

func (r *AutomationRepository) ListRules(ctx context.Context, orgID uuid.UUID, page, limit int) ([]domain.AutomationRule, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	countQuery := `SELECT COUNT(*) FROM automation_rules WHERE organization_id = $1;`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count automation rules: %w", err)
	}

	query := `
		SELECT
			id, organization_id, name, description, is_active,
			trigger_type, trigger_config, conditions, actions,
			cooldown_seconds, last_triggered_at, created_at, updated_at
		FROM automation_rules
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;
	`

	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list automation rules: %w", err)
	}
	defer rows.Close()

	rules := make([]domain.AutomationRule, 0)
	for rows.Next() {
		var rule domain.AutomationRule
		var conditionsJSON, actionsJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.OrganizationID,
			&rule.Name,
			&rule.Description,
			&rule.IsActive,
			&rule.TriggerType,
			&rule.TriggerConfig,
			&conditionsJSON,
			&actionsJSON,
			&rule.CooldownSeconds,
			&rule.LastTriggeredAt,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan automation rule row: %w", err)
		}

		if len(conditionsJSON) > 0 {
			_ = json.Unmarshal(conditionsJSON, &rule.Conditions)
		}
		if len(actionsJSON) > 0 {
			_ = json.Unmarshal(actionsJSON, &rule.Actions)
		}

		rules = append(rules, rule)
	}

	return rules, total, nil
}

func (r *AutomationRepository) ListActiveRulesByTriggerType(ctx context.Context, triggerType domain.RuleTriggerType) ([]domain.AutomationRule, error) {
	query := `
		SELECT
			id, organization_id, name, description, is_active,
			trigger_type, trigger_config, conditions, actions,
			cooldown_seconds, last_triggered_at, created_at, updated_at
		FROM automation_rules
		WHERE is_active = true AND trigger_type = $1;
	`

	rows, err := r.pool.Query(ctx, query, triggerType)
	if err != nil {
		return nil, fmt.Errorf("failed to list active automation rules by trigger: %w", err)
	}
	defer rows.Close()

	rules := make([]domain.AutomationRule, 0)
	for rows.Next() {
		var rule domain.AutomationRule
		var conditionsJSON, actionsJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.OrganizationID,
			&rule.Name,
			&rule.Description,
			&rule.IsActive,
			&rule.TriggerType,
			&rule.TriggerConfig,
			&conditionsJSON,
			&actionsJSON,
			&rule.CooldownSeconds,
			&rule.LastTriggeredAt,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active automation rule: %w", err)
		}

		if len(conditionsJSON) > 0 {
			_ = json.Unmarshal(conditionsJSON, &rule.Conditions)
		}
		if len(actionsJSON) > 0 {
			_ = json.Unmarshal(actionsJSON, &rule.Actions)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func (r *AutomationRepository) UpdateRule(ctx context.Context, rule *domain.AutomationRule) error {
	query := `
		UPDATE automation_rules
		SET
			name = $3,
			description = $4,
			is_active = $5,
			trigger_type = $6,
			trigger_config = $7,
			conditions = $8,
			actions = $9,
			cooldown_seconds = $10,
			updated_at = $11
		WHERE id = $1 AND organization_id = $2;
	`

	rule.UpdatedAt = time.Now().UTC()

	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal rule conditions: %w", err)
	}

	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal rule actions: %w", err)
	}

	res, err := r.pool.Exec(ctx, query,
		rule.ID,
		rule.OrganizationID,
		rule.Name,
		rule.Description,
		rule.IsActive,
		rule.TriggerType,
		rule.TriggerConfig,
		conditionsJSON,
		actionsJSON,
		rule.CooldownSeconds,
		rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update automation rule: %w", err)
	}

	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *AutomationRepository) UpdateLastTriggered(ctx context.Context, ruleID uuid.UUID, triggeredAt time.Time) error {
	query := `UPDATE automation_rules SET last_triggered_at = $2 WHERE id = $1;`
	_, err := r.pool.Exec(ctx, query, ruleID, triggeredAt)
	return err
}

func (r *AutomationRepository) DeleteRule(ctx context.Context, orgID, id uuid.UUID) error {
	query := `DELETE FROM automation_rules WHERE id = $1 AND organization_id = $2;`
	res, err := r.pool.Exec(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete automation rule: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AutomationRepository) CreateExecutionLog(ctx context.Context, log *domain.RuleExecutionLog) error {
	query := `
		INSERT INTO automation_execution_logs (
			id, rule_id, organization_id, trigger_event, status,
			evaluated_conditions, executed_actions, error_message,
			execution_duration_ms, executed_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10
		);
	`

	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.ExecutedAt.IsZero() {
		log.ExecutedAt = time.Now().UTC()
	}

	evaluatedJSON := log.EvaluatedConditions
	if len(evaluatedJSON) == 0 {
		evaluatedJSON = []byte("{}")
	}

	actionsJSON, err := json.Marshal(log.ExecutedActions)
	if err != nil {
		actionsJSON = []byte("[]")
	}

	_, err = r.pool.Exec(ctx, query,
		log.ID,
		log.RuleID,
		log.OrganizationID,
		log.TriggerEvent,
		log.Status,
		evaluatedJSON,
		actionsJSON,
		log.ErrorMessage,
		log.ExecutionDurationMs,
		log.ExecutedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert automation execution log: %w", err)
	}

	return nil
}

func (r *AutomationRepository) ListExecutionLogs(
	ctx context.Context,
	orgID uuid.UUID,
	ruleID *uuid.UUID,
	status *domain.ExecutionStatus,
	page, limit int,
) ([]domain.RuleExecutionLog, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	whereClause := `WHERE l.organization_id = $1`
	args := []any{orgID}
	argIdx := 2

	if ruleID != nil && *ruleID != uuid.Nil {
		whereClause += fmt.Sprintf(` AND l.rule_id = $%d`, argIdx)
		args = append(args, *ruleID)
		argIdx++
	}

	if status != nil && *status != "" {
		whereClause += fmt.Sprintf(` AND l.status = $%d`, argIdx)
		args = append(args, *status)
		argIdx++
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM automation_execution_logs l %s;`, whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count execution logs: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT
			l.id, l.rule_id, l.organization_id, COALESCE(r.name, 'Deleted Rule'),
			l.trigger_event, l.status, l.evaluated_conditions, l.executed_actions,
			COALESCE(l.error_message, ''), l.execution_duration_ms, l.executed_at
		FROM automation_execution_logs l
		LEFT JOIN automation_rules r ON l.rule_id = r.id
		%s
		ORDER BY l.executed_at DESC
		LIMIT $%d OFFSET $%d;
	`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list execution logs: %w", err)
	}
	defer rows.Close()

	logs := make([]domain.RuleExecutionLog, 0)
	for rows.Next() {
		var l domain.RuleExecutionLog
		var actionsJSON []byte

		err := rows.Scan(
			&l.ID,
			&l.RuleID,
			&l.OrganizationID,
			&l.RuleName,
			&l.TriggerEvent,
			&l.Status,
			&l.EvaluatedConditions,
			&actionsJSON,
			&l.ErrorMessage,
			&l.ExecutionDurationMs,
			&l.ExecutedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan execution log row: %w", err)
		}

		if len(actionsJSON) > 0 {
			_ = json.Unmarshal(actionsJSON, &l.ExecutedActions)
		}

		logs = append(logs, l)
	}

	return logs, total, nil
}
