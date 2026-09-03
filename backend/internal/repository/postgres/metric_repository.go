package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricRepository struct {
	pool *pgxpool.Pool
}

func NewMetricRepository(pool *pgxpool.Pool) *MetricRepository {
	return &MetricRepository{pool: pool}
}

func (r *MetricRepository) Create(ctx context.Context, metric *domain.ServerMetric) error {
	query := `
		INSERT INTO server_metrics (
			server_id, cpu_usage_pct, memory_used_mb, memory_total_mb, memory_usage_pct,
			disk_used_gb, disk_total_gb, disk_usage_pct, network_in_kb, network_out_kb,
			network_in_rate_kbps, network_out_rate_kbps, uptime_seconds, containers_count,
			docker_available, containers_json, recorded_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		) RETURNING id;
	`

	if metric.RecordedAt.IsZero() {
		metric.RecordedAt = time.Now().UTC()
	}

	containersJSON := metric.ContainersJSON
	if containersJSON == "" {
		containersJSON = "[]"
	}

	return r.pool.QueryRow(
		ctx,
		query,
		metric.ServerID,
		metric.CPUUsagePct,
		metric.MemoryUsedMB,
		metric.MemoryTotalMB,
		metric.MemoryUsagePct,
		metric.DiskUsedGB,
		metric.DiskTotalGB,
		metric.DiskUsagePct,
		metric.NetworkInKB,
		metric.NetworkOutKB,
		metric.NetworkInRateKBps,
		metric.NetworkOutRateKBps,
		metric.UptimeSeconds,
		metric.ContainersCount,
		metric.DockerAvailable,
		containersJSON,
		metric.RecordedAt,
	).Scan(&metric.ID)
}

func (r *MetricRepository) GetLatestByServerID(ctx context.Context, serverID uuid.UUID) (*domain.ServerMetric, error) {
	query := `
		SELECT
			id, server_id, cpu_usage_pct, memory_used_mb, memory_total_mb, memory_usage_pct,
			disk_used_gb, disk_total_gb, disk_usage_pct, network_in_kb, network_out_kb,
			network_in_rate_kbps, network_out_rate_kbps, uptime_seconds, containers_count,
			docker_available, containers_json, recorded_at
		FROM server_metrics
		WHERE server_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1;
	`

	var m domain.ServerMetric
	err := r.pool.QueryRow(ctx, query, serverID).Scan(
		&m.ID,
		&m.ServerID,
		&m.CPUUsagePct,
		&m.MemoryUsedMB,
		&m.MemoryTotalMB,
		&m.MemoryUsagePct,
		&m.DiskUsedGB,
		&m.DiskTotalGB,
		&m.DiskUsagePct,
		&m.NetworkInKB,
		&m.NetworkOutKB,
		&m.NetworkInRateKBps,
		&m.NetworkOutRateKBps,
		&m.UptimeSeconds,
		&m.ContainersCount,
		&m.DockerAvailable,
		&m.ContainersJSON,
		&m.RecordedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: metric for server %s", domain.ErrNotFound, serverID)
		}
		return nil, fmt.Errorf("failed to query latest metric: %w", err)
	}

	return &m, nil
}

func (r *MetricRepository) GetHistoryByServerID(ctx context.Context, serverID uuid.UUID, from, to time.Time, limit int) ([]domain.ServerMetric, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query := `
		SELECT
			id, server_id, cpu_usage_pct, memory_used_mb, memory_total_mb, memory_usage_pct,
			disk_used_gb, disk_total_gb, disk_usage_pct, network_in_kb, network_out_kb,
			network_in_rate_kbps, network_out_rate_kbps, uptime_seconds, containers_count,
			docker_available, containers_json, recorded_at
		FROM server_metrics
		WHERE server_id = $1 AND recorded_at >= $2 AND recorded_at <= $3
		ORDER BY recorded_at ASC
		LIMIT $4;
	`

	rows, err := r.pool.Query(ctx, query, serverID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query metric history: %w", err)
	}
	defer rows.Close()

	metrics := make([]domain.ServerMetric, 0)
	for rows.Next() {
		var m domain.ServerMetric
		err := rows.Scan(
			&m.ID,
			&m.ServerID,
			&m.CPUUsagePct,
			&m.MemoryUsedMB,
			&m.MemoryTotalMB,
			&m.MemoryUsagePct,
			&m.DiskUsedGB,
			&m.DiskTotalGB,
			&m.DiskUsagePct,
			&m.NetworkInKB,
			&m.NetworkOutKB,
			&m.NetworkInRateKBps,
			&m.NetworkOutRateKBps,
			&m.UptimeSeconds,
			&m.ContainersCount,
			&m.DockerAvailable,
			&m.ContainersJSON,
			&m.RecordedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric row: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}
