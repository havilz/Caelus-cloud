package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type AlertRepository struct {
	pool *pgxpool.Pool
}

// NewAlertRepository menginisialisasi repository Alert dan AlertRule berbasis database PostgreSQL.
func NewAlertRepository(pool *pgxpool.Pool) *AlertRepository {
	return &AlertRepository{pool: pool}
}

// CreateAlert menyimpan data insiden alert baru ke dalam database.
func (r *AlertRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	query := `
		INSERT INTO alerts (
			id, organization_id, server_id, rule_id, alert_type, severity,
			title, message, status, current_value, threshold_value, triggered_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at;
	`

	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	if alert.TriggeredAt.IsZero() {
		alert.TriggeredAt = time.Now().UTC()
	}
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = time.Now().UTC()
	}
	if alert.Status == "" {
		alert.Status = domain.AlertStatusActive
	}

	return r.pool.QueryRow(
		ctx,
		query,
		alert.ID,
		alert.OrganizationID,
		alert.ServerID,
		alert.RuleID,
		alert.AlertType,
		alert.Severity,
		alert.Title,
		alert.Message,
		alert.Status,
		alert.CurrentValue,
		alert.ThresholdValue,
		alert.TriggeredAt,
		alert.CreatedAt,
	).Scan(&alert.ID, &alert.CreatedAt)
}

// GetAlertByID mengambil detail satu entitas alert berdasarkan ID.
func (r *AlertRepository) GetAlertByID(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	query := `
		SELECT
			id, organization_id, server_id, rule_id, alert_type, severity,
			title, message, status, current_value, threshold_value, acknowledged_at,
			acknowledged_by, resolved_at, resolved_by, triggered_at, created_at
		FROM alerts
		WHERE id = $1;
	`

	var a domain.Alert
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID,
		&a.OrganizationID,
		&a.ServerID,
		&a.RuleID,
		&a.AlertType,
		&a.Severity,
		&a.Title,
		&a.Message,
		&a.Status,
		&a.CurrentValue,
		&a.ThresholdValue,
		&a.AcknowledgedAt,
		&a.AcknowledgedBy,
		&a.ResolvedAt,
		&a.ResolvedBy,
		&a.TriggeredAt,
		&a.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: alert %s", domain.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to query alert by id: %w", err)
	}

	return &a, nil
}

// ListAlertsByOrg mengambil daftar alert milik organisasi dengan paginasi dan filter status opsional.
func (r *AlertRepository) ListAlertsByOrg(ctx context.Context, orgID uuid.UUID, status *domain.AlertStatus, page, limit int) ([]domain.Alert, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var countQuery string
	var listQuery string
	var err error
	var total int64
	var rows pgx.Rows

	if status != nil {
		countQuery = `SELECT COUNT(*) FROM alerts WHERE organization_id = $1 AND status = $2;`
		err = r.pool.QueryRow(ctx, countQuery, orgID, *status).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
		}

		listQuery = `
			SELECT
				id, organization_id, server_id, rule_id, alert_type, severity,
				title, message, status, current_value, threshold_value, acknowledged_at,
				acknowledged_by, resolved_at, resolved_by, triggered_at, created_at
			FROM alerts
			WHERE organization_id = $1 AND status = $2
			ORDER BY triggered_at DESC
			LIMIT $3 OFFSET $4;
		`
		rows, err = r.pool.Query(ctx, listQuery, orgID, *status, limit, offset)
	} else {
		countQuery = `SELECT COUNT(*) FROM alerts WHERE organization_id = $1;`
		err = r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
		}

		listQuery = `
			SELECT
				id, organization_id, server_id, rule_id, alert_type, severity,
				title, message, status, current_value, threshold_value, acknowledged_at,
				acknowledged_by, resolved_at, resolved_by, triggered_at, created_at
			FROM alerts
			WHERE organization_id = $1
			ORDER BY triggered_at DESC
			LIMIT $2 OFFSET $3;
		`
		rows, err = r.pool.Query(ctx, listQuery, orgID, limit, offset)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list alerts: %w", err)
	}
	defer rows.Close()

	alerts := make([]domain.Alert, 0)
	for rows.Next() {
		var a domain.Alert
		err := rows.Scan(
			&a.ID,
			&a.OrganizationID,
			&a.ServerID,
			&a.RuleID,
			&a.AlertType,
			&a.Severity,
			&a.Title,
			&a.Message,
			&a.Status,
			&a.CurrentValue,
			&a.ThresholdValue,
			&a.AcknowledgedAt,
			&a.AcknowledgedBy,
			&a.ResolvedAt,
			&a.ResolvedBy,
			&a.TriggeredAt,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan alert row: %w", err)
		}
		alerts = append(alerts, a)
	}

	return alerts, total, nil
}

// ListActiveAlertsByServer mengambil seluruh alert berstatus aktif untuk server tertentu.
func (r *AlertRepository) ListActiveAlertsByServer(ctx context.Context, serverID uuid.UUID) ([]domain.Alert, error) {
	query := `
		SELECT
			id, organization_id, server_id, rule_id, alert_type, severity,
			title, message, status, current_value, threshold_value, acknowledged_at,
			acknowledged_by, resolved_at, resolved_by, triggered_at, created_at
		FROM alerts
		WHERE server_id = $1 AND status = 'active'
		ORDER BY triggered_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active alerts: %w", err)
	}
	defer rows.Close()

	alerts := make([]domain.Alert, 0)
	for rows.Next() {
		var a domain.Alert
		err := rows.Scan(
			&a.ID,
			&a.OrganizationID,
			&a.ServerID,
			&a.RuleID,
			&a.AlertType,
			&a.Severity,
			&a.Title,
			&a.Message,
			&a.Status,
			&a.CurrentValue,
			&a.ThresholdValue,
			&a.AcknowledgedAt,
			&a.AcknowledgedBy,
			&a.ResolvedAt,
			&a.ResolvedBy,
			&a.TriggeredAt,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert row: %w", err)
		}
		alerts = append(alerts, a)
	}

	return alerts, nil
}

// UpdateAlertStatus memperbarui status siklus hidup alert (misal Acknowledged atau Resolved).
func (r *AlertRepository) UpdateAlertStatus(ctx context.Context, id uuid.UUID, status domain.AlertStatus, userID *uuid.UUID, timestamp *time.Time) error {
	var query string
	if status == domain.AlertStatusAcknowledged {
		query = `
			UPDATE alerts
			SET status = $2, acknowledged_at = $3, acknowledged_by = $4
			WHERE id = $1;
		`
	} else if status == domain.AlertStatusResolved {
		query = `
			UPDATE alerts
			SET status = $2, resolved_at = $3, resolved_by = $4
			WHERE id = $1;
		`
	} else {
		query = `
			UPDATE alerts
			SET status = $2
			WHERE id = $1;
		`
		_, err := r.pool.Exec(ctx, query, id, status)
		return err
	}

	res, err := r.pool.Exec(ctx, query, id, status, timestamp, userID)
	if err != nil {
		return fmt.Errorf("failed to update alert status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%w: alert %s", domain.ErrNotFound, id)
	}
	return nil
}

// CreateRule menyimpan aturan evaluasi threshold alert baru.
func (r *AlertRepository) CreateRule(ctx context.Context, rule *domain.AlertRule) error {
	query := `
		INSERT INTO alert_rules (
			id, organization_id, server_id, name, metric_name, operator,
			threshold, duration_seconds, severity, is_enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at;
	`

	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now().UTC()
	}
	if rule.UpdatedAt.IsZero() {
		rule.UpdatedAt = time.Now().UTC()
	}

	return r.pool.QueryRow(
		ctx,
		query,
		rule.ID,
		rule.OrganizationID,
		rule.ServerID,
		rule.Name,
		rule.MetricName,
		rule.Operator,
		rule.Threshold,
		rule.DurationSeconds,
		rule.Severity,
		rule.IsEnabled,
		rule.CreatedAt,
		rule.UpdatedAt,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
}

// GetRuleByID mengambil satu aturan evaluasi alert berdasarkan ID.
func (r *AlertRepository) GetRuleByID(ctx context.Context, id uuid.UUID) (*domain.AlertRule, error) {
	query := `
		SELECT
			id, organization_id, server_id, name, metric_name, operator,
			threshold, duration_seconds, severity, is_enabled, created_at, updated_at
		FROM alert_rules
		WHERE id = $1;
	`

	var rule domain.AlertRule
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rule.ID,
		&rule.OrganizationID,
		&rule.ServerID,
		&rule.Name,
		&rule.MetricName,
		&rule.Operator,
		&rule.Threshold,
		&rule.DurationSeconds,
		&rule.Severity,
		&rule.IsEnabled,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: alert rule %s", domain.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to query alert rule by id: %w", err)
	}

	return &rule, nil
}

// ListRulesByOrg mengambil seluruh aturan alert milik sebuah organisasi.
func (r *AlertRepository) ListRulesByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.AlertRule, error) {
	query := `
		SELECT
			id, organization_id, server_id, name, metric_name, operator,
			threshold, duration_seconds, severity, is_enabled, created_at, updated_at
		FROM alert_rules
		WHERE organization_id = $1
		ORDER BY created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}
	defer rows.Close()

	rules := make([]domain.AlertRule, 0)
	for rows.Next() {
		var rule domain.AlertRule
		err := rows.Scan(
			&rule.ID,
			&rule.OrganizationID,
			&rule.ServerID,
			&rule.Name,
			&rule.MetricName,
			&rule.Operator,
			&rule.Threshold,
			&rule.DurationSeconds,
			&rule.Severity,
			&rule.IsEnabled,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert rule row: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// ListRulesForServer mengambil aturan alert yang berlaku untuk server tertentu (aturan spesifik server atau aturan global organisasi).
func (r *AlertRepository) ListRulesForServer(ctx context.Context, orgID, serverID uuid.UUID) ([]domain.AlertRule, error) {
	query := `
		SELECT
			id, organization_id, server_id, name, metric_name, operator,
			threshold, duration_seconds, severity, is_enabled, created_at, updated_at
		FROM alert_rules
		WHERE organization_id = $1 AND is_enabled = true AND (server_id = $2 OR server_id IS NULL)
		ORDER BY created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, orgID, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules for server: %w", err)
	}
	defer rows.Close()

	rules := make([]domain.AlertRule, 0)
	for rows.Next() {
		var rule domain.AlertRule
		err := rows.Scan(
			&rule.ID,
			&rule.OrganizationID,
			&rule.ServerID,
			&rule.Name,
			&rule.MetricName,
			&rule.Operator,
			&rule.Threshold,
			&rule.DurationSeconds,
			&rule.Severity,
			&rule.IsEnabled,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert rule row: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// DeleteRule menghapus satu aturan alert berdasarkan ID.
func (r *AlertRepository) DeleteRule(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM alert_rules WHERE id = $1;`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("%w: alert rule %s", domain.ErrNotFound, id)
	}
	return nil
}
