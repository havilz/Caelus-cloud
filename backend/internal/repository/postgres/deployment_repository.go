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

type DeploymentRepository struct {
	pool *pgxpool.Pool
}

func NewDeploymentRepository(pool *pgxpool.Pool) *DeploymentRepository {
	return &DeploymentRepository{pool: pool}
}

func (r *DeploymentRepository) CreateDeployment(ctx context.Context, dep *domain.Deployment) error {
	query := `
		INSERT INTO deployments (id, organization_id, server_id, app_name, image_tag, container_name, network_name, environment_variables, port_bindings, volume_bindings, restart_policy, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at;
	`
	if dep.ID == uuid.Nil {
		dep.ID = uuid.New()
	}
	now := time.Now().UTC()
	dep.CreatedAt = now
	dep.UpdatedAt = now

	envJSON, _ := json.Marshal(dep.EnvironmentVariables)
	portsJSON, _ := json.Marshal(dep.PortBindings)
	volsJSON, _ := json.Marshal(dep.VolumeBindings)

	return r.pool.QueryRow(
		ctx,
		query,
		dep.ID,
		dep.OrganizationID,
		dep.ServerID,
		dep.AppName,
		dep.ImageTag,
		dep.ContainerName,
		dep.NetworkName,
		envJSON,
		portsJSON,
		volsJSON,
		dep.RestartPolicy,
		dep.Status,
		dep.CreatedAt,
		dep.UpdatedAt,
	).Scan(&dep.ID, &dep.CreatedAt, &dep.UpdatedAt)
}

func (r *DeploymentRepository) GetDeploymentByID(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	query := `
		SELECT id, organization_id, server_id, app_name, image_tag, container_name, COALESCE(network_name, ''),
		       COALESCE(environment_variables, '{}'::jsonb), 
		       COALESCE(port_bindings, '[]'::jsonb), 
		       COALESCE(volume_bindings, '[]'::jsonb), 
		       restart_policy, status, COALESCE(error_message, ''), created_at, updated_at, finished_at
		FROM deployments
		WHERE id = $1;
	`
	var d domain.Deployment
	var envBytes, portsBytes, volsBytes []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&d.ID,
		&d.OrganizationID,
		&d.ServerID,
		&d.AppName,
		&d.ImageTag,
		&d.ContainerName,
		&d.NetworkName,
		&envBytes,
		&portsBytes,
		&volsBytes,
		&d.RestartPolicy,
		&d.Status,
		&d.ErrorMessage,
		&d.CreatedAt,
		&d.UpdatedAt,
		&d.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("deployment not found")
		}
		return nil, err
	}
	_ = json.Unmarshal(envBytes, &d.EnvironmentVariables)
	_ = json.Unmarshal(portsBytes, &d.PortBindings)
	_ = json.Unmarshal(volsBytes, &d.VolumeBindings)
	return &d, nil
}

func (r *DeploymentRepository) ListDeploymentsByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.Deployment, error) {
	query := `
		SELECT id, organization_id, server_id, app_name, image_tag, container_name, COALESCE(network_name, ''),
		       COALESCE(environment_variables, '{}'::jsonb), 
		       COALESCE(port_bindings, '[]'::jsonb), 
		       COALESCE(volume_bindings, '[]'::jsonb), 
		       restart_policy, status, COALESCE(error_message, ''), created_at, updated_at, finished_at
		FROM deployments
		WHERE organization_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deployments := make([]domain.Deployment, 0)
	for rows.Next() {
		var d domain.Deployment
		var envBytes, portsBytes, volsBytes []byte
		if err := rows.Scan(
			&d.ID,
			&d.OrganizationID,
			&d.ServerID,
			&d.AppName,
			&d.ImageTag,
			&d.ContainerName,
			&d.NetworkName,
			&envBytes,
			&portsBytes,
			&volsBytes,
			&d.RestartPolicy,
			&d.Status,
			&d.ErrorMessage,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.FinishedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(envBytes, &d.EnvironmentVariables)
		_ = json.Unmarshal(portsBytes, &d.PortBindings)
		_ = json.Unmarshal(volsBytes, &d.VolumeBindings)
		deployments = append(deployments, d)
	}
	return deployments, nil
}

func (r *DeploymentRepository) ListDeploymentsByServer(ctx context.Context, serverID uuid.UUID) ([]domain.Deployment, error) {
	query := `
		SELECT id, organization_id, server_id, app_name, image_tag, container_name, COALESCE(network_name, ''),
		       COALESCE(environment_variables, '{}'::jsonb), 
		       COALESCE(port_bindings, '[]'::jsonb), 
		       COALESCE(volume_bindings, '[]'::jsonb), 
		       restart_policy, status, COALESCE(error_message, ''), created_at, updated_at, finished_at
		FROM deployments
		WHERE server_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deployments := make([]domain.Deployment, 0)
	for rows.Next() {
		var d domain.Deployment
		var envBytes, portsBytes, volsBytes []byte
		if err := rows.Scan(
			&d.ID,
			&d.OrganizationID,
			&d.ServerID,
			&d.AppName,
			&d.ImageTag,
			&d.ContainerName,
			&d.NetworkName,
			&envBytes,
			&portsBytes,
			&volsBytes,
			&d.RestartPolicy,
			&d.Status,
			&d.ErrorMessage,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.FinishedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(envBytes, &d.EnvironmentVariables)
		_ = json.Unmarshal(portsBytes, &d.PortBindings)
		_ = json.Unmarshal(volsBytes, &d.VolumeBindings)
		deployments = append(deployments, d)
	}
	return deployments, nil
}

func (r *DeploymentRepository) UpdateDeploymentStatus(ctx context.Context, id uuid.UUID, status domain.DeploymentStatus, errorMsg string, finishedAt *time.Time) error {
	query := `
		UPDATE deployments
		SET status = $1, error_message = $2, finished_at = $3, updated_at = $4
		WHERE id = $5;
	`
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, query, status, errorMsg, finishedAt, now, id)
	return err
}

func (r *DeploymentRepository) DeleteDeployment(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM deployments WHERE id = $1;`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *DeploymentRepository) AppendLog(ctx context.Context, log *domain.DeploymentLog) error {
	query := `
		INSERT INTO deployment_logs (deployment_id, timestamp, stream, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now().UTC()
	}
	return r.pool.QueryRow(ctx, query, log.DeploymentID, log.Timestamp, log.Stream, log.Message).Scan(&log.ID)
}

func (r *DeploymentRepository) GetLogsByDeployment(ctx context.Context, deploymentID uuid.UUID, limit int) ([]domain.DeploymentLog, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	query := `
		SELECT id, deployment_id, timestamp, stream, message
		FROM deployment_logs
		WHERE deployment_id = $1
		ORDER BY timestamp ASC
		LIMIT $2;
	`
	rows, err := r.pool.Query(ctx, query, deploymentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]domain.DeploymentLog, 0)
	for rows.Next() {
		var l domain.DeploymentLog
		if err := rows.Scan(
			&l.ID,
			&l.DeploymentID,
			&l.Timestamp,
			&l.Stream,
			&l.Message,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
