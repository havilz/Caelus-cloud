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

type IaCRepository struct {
	pool *pgxpool.Pool
}

func NewIaCRepository(pool *pgxpool.Pool) *IaCRepository {
	return &IaCRepository{pool: pool}
}

func (r *IaCRepository) CreateConfig(ctx context.Context, config *domain.IaCConfiguration) error {
	query := `
		INSERT INTO iac_configurations (id, organization_id, name, description, raw_yaml, status, current_version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at;
	`
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}
	now := time.Now().UTC()
	config.CreatedAt = now
	config.UpdatedAt = now

	return r.pool.QueryRow(
		ctx,
		query,
		config.ID,
		config.OrganizationID,
		config.Name,
		config.Description,
		config.RawYAML,
		config.Status,
		config.CurrentVersion,
		config.CreatedAt,
		config.UpdatedAt,
	).Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)
}

func (r *IaCRepository) GetConfigByID(ctx context.Context, id uuid.UUID) (*domain.IaCConfiguration, error) {
	query := `
		SELECT id, organization_id, name, description, raw_yaml, status, current_version, created_at, updated_at
		FROM iac_configurations
		WHERE id = $1;
	`
	var c domain.IaCConfiguration
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.OrganizationID,
		&c.Name,
		&c.Description,
		&c.RawYAML,
		&c.Status,
		&c.CurrentVersion,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("iac configuration not found")
		}
		return nil, err
	}
	return &c, nil
}

func (r *IaCRepository) ListConfigsByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.IaCConfiguration, error) {
	query := `
		SELECT id, organization_id, name, description, raw_yaml, status, current_version, created_at, updated_at
		FROM iac_configurations
		WHERE organization_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make([]domain.IaCConfiguration, 0)
	for rows.Next() {
		var c domain.IaCConfiguration
		if err := rows.Scan(
			&c.ID,
			&c.OrganizationID,
			&c.Name,
			&c.Description,
			&c.RawYAML,
			&c.Status,
			&c.CurrentVersion,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, nil
}

func (r *IaCRepository) UpdateConfig(ctx context.Context, config *domain.IaCConfiguration) error {
	query := `
		UPDATE iac_configurations
		SET name = $1, description = $2, raw_yaml = $3, status = $4, current_version = $5, updated_at = $6
		WHERE id = $7;
	`
	config.UpdatedAt = time.Now().UTC()
	_, err := r.pool.Exec(
		ctx,
		query,
		config.Name,
		config.Description,
		config.RawYAML,
		config.Status,
		config.CurrentVersion,
		config.UpdatedAt,
		config.ID,
	)
	return err
}

func (r *IaCRepository) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM iac_configurations WHERE id = $1;`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *IaCRepository) CreateState(ctx context.Context, state *domain.IaCState) error {
	query := `
		INSERT INTO iac_states (id, configuration_id, version, state_data, hash, applied_at, applied_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at;
	`
	if state.ID == uuid.Nil {
		state.ID = uuid.New()
	}
	if state.CreatedAt.IsZero() {
		state.CreatedAt = time.Now().UTC()
	}

	stateJSON, err := json.Marshal(state.StateData)
	if err != nil {
		return err
	}

	return r.pool.QueryRow(
		ctx,
		query,
		state.ID,
		state.ConfigurationID,
		state.Version,
		stateJSON,
		state.Hash,
		state.AppliedAt,
		state.AppliedBy,
		state.CreatedAt,
	).Scan(&state.ID, &state.CreatedAt)
}

func (r *IaCRepository) GetLatestStateByConfigID(ctx context.Context, configID uuid.UUID) (*domain.IaCState, error) {
	query := `
		SELECT id, configuration_id, version, state_data, hash, applied_at, applied_by, created_at
		FROM iac_states
		WHERE configuration_id = $1
		ORDER BY version DESC
		LIMIT 1;
	`
	var s domain.IaCState
	var stateBytes []byte
	err := r.pool.QueryRow(ctx, query, configID).Scan(
		&s.ID,
		&s.ConfigurationID,
		&s.Version,
		&stateBytes,
		&s.Hash,
		&s.AppliedAt,
		&s.AppliedBy,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	_ = json.Unmarshal(stateBytes, &s.StateData)
	return &s, nil
}

func (r *IaCRepository) GetStateByVersion(ctx context.Context, configID uuid.UUID, version int) (*domain.IaCState, error) {
	query := `
		SELECT id, configuration_id, version, state_data, hash, applied_at, applied_by, created_at
		FROM iac_states
		WHERE configuration_id = $1 AND version = $2;
	`
	var s domain.IaCState
	var stateBytes []byte
	err := r.pool.QueryRow(ctx, query, configID, version).Scan(
		&s.ID,
		&s.ConfigurationID,
		&s.Version,
		&stateBytes,
		&s.Hash,
		&s.AppliedAt,
		&s.AppliedBy,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("state version %d not found", version)
		}
		return nil, err
	}
	_ = json.Unmarshal(stateBytes, &s.StateData)
	return &s, nil
}

func (r *IaCRepository) ListStatesByConfigID(ctx context.Context, configID uuid.UUID) ([]domain.IaCState, error) {
	query := `
		SELECT id, configuration_id, version, state_data, hash, applied_at, applied_by, created_at
		FROM iac_states
		WHERE configuration_id = $1
		ORDER BY version DESC;
	`
	rows, err := r.pool.Query(ctx, query, configID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := make([]domain.IaCState, 0)
	for rows.Next() {
		var s domain.IaCState
		var stateBytes []byte
		if err := rows.Scan(
			&s.ID,
			&s.ConfigurationID,
			&s.Version,
			&stateBytes,
			&s.Hash,
			&s.AppliedAt,
			&s.AppliedBy,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(stateBytes, &s.StateData)
		states = append(states, s)
	}
	return states, nil
}

func (r *IaCRepository) CreatePlan(ctx context.Context, plan *domain.IaCPlan) error {
	query := `
		INSERT INTO iac_plans (id, configuration_id, target_version, changes, summary, status, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at;
	`
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = time.Now().UTC()
	}

	changesJSON, _ := json.Marshal(plan.Changes)
	summaryJSON, _ := json.Marshal(plan.Summary)

	return r.pool.QueryRow(
		ctx,
		query,
		plan.ID,
		plan.ConfigurationID,
		plan.TargetVersion,
		changesJSON,
		summaryJSON,
		plan.Status,
		plan.ErrorMessage,
		plan.CreatedAt,
	).Scan(&plan.ID, &plan.CreatedAt)
}

func (r *IaCRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*domain.IaCPlan, error) {
	query := `
		SELECT id, configuration_id, target_version, changes, summary, status, error_message, created_at, executed_at
		FROM iac_plans
		WHERE id = $1;
	`
	var p domain.IaCPlan
	var changesBytes, summaryBytes []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.ConfigurationID,
		&p.TargetVersion,
		&changesBytes,
		&summaryBytes,
		&p.Status,
		&p.ErrorMessage,
		&p.CreatedAt,
		&p.ExecutedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("iac plan not found")
		}
		return nil, err
	}
	_ = json.Unmarshal(changesBytes, &p.Changes)
	_ = json.Unmarshal(summaryBytes, &p.Summary)
	return &p, nil
}

func (r *IaCRepository) UpdatePlanStatus(ctx context.Context, id uuid.UUID, status domain.IaCStatus, errorMsg string) error {
	query := `
		UPDATE iac_plans
		SET status = $1, error_message = $2, executed_at = $3
		WHERE id = $4;
	`
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, query, status, errorMsg, now, id)
	return err
}

func (r *IaCRepository) GetLatestPlanByConfigID(ctx context.Context, configID uuid.UUID) (*domain.IaCPlan, error) {
	query := `
		SELECT id, configuration_id, target_version, changes, summary, status, error_message, created_at, executed_at
		FROM iac_plans
		WHERE configuration_id = $1
		ORDER BY created_at DESC
		LIMIT 1;
	`
	var p domain.IaCPlan
	var changesBytes, summaryBytes []byte
	err := r.pool.QueryRow(ctx, query, configID).Scan(
		&p.ID,
		&p.ConfigurationID,
		&p.TargetVersion,
		&changesBytes,
		&summaryBytes,
		&p.Status,
		&p.ErrorMessage,
		&p.CreatedAt,
		&p.ExecutedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	_ = json.Unmarshal(changesBytes, &p.Changes)
	_ = json.Unmarshal(summaryBytes, &p.Summary)
	return &p, nil
}
