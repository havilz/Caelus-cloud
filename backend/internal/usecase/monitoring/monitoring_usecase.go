package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// MonitoringUsecase mendefinisikan interface logika bisnis ingestion telemetri, metrik history, dan pengelolaan alert.
type MonitoringUsecase interface {
	IngestTelemetry(ctx context.Context, payload *domain.TelemetryReportPayload) error
	GetServerLiveMetrics(ctx context.Context, serverID uuid.UUID) (*domain.ServerMetric, error)
	GetServerMetricHistory(ctx context.Context, serverID uuid.UUID, duration time.Duration) ([]domain.ServerMetric, error)
	ListAlerts(ctx context.Context, orgID uuid.UUID, status *domain.AlertStatus, page, limit int) ([]domain.Alert, int64, error)
	AcknowledgeAlert(ctx context.Context, alertID, userID uuid.UUID) error
	ResolveAlert(ctx context.Context, alertID, userID uuid.UUID) error
	CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error
	ListAlertRules(ctx context.Context, orgID uuid.UUID) ([]domain.AlertRule, error)
	DeleteAlertRule(ctx context.Context, ruleID uuid.UUID) error
}

type monitoringUsecase struct {
	metricRepo  domain.MetricRepository
	alertRepo   domain.AlertRepository
	serverRepo  domain.ServerRepository
	evaluator   *AlertEvaluator
	broadcaster domain.TelemetryBroadcaster
	promAdapter domain.MetricsQueryAdapter
	lokiAdapter domain.LogQueryAdapter
}

// NewMonitoringUsecase membuat instance baru implementasi MonitoringUsecase.
func NewMonitoringUsecase(
	metricRepo domain.MetricRepository,
	alertRepo domain.AlertRepository,
	serverRepo domain.ServerRepository,
	evaluator *AlertEvaluator,
	broadcaster domain.TelemetryBroadcaster,
	promAdapter domain.MetricsQueryAdapter,
	lokiAdapter domain.LogQueryAdapter,
) MonitoringUsecase {
	return &monitoringUsecase{
		metricRepo:  metricRepo,
		alertRepo:   alertRepo,
		serverRepo:  serverRepo,
		evaluator:   evaluator,
		broadcaster: broadcaster,
		promAdapter: promAdapter,
		lokiAdapter: lokiAdapter,
	}
}

// IngestTelemetry memproses laporan telemetri yang dikirimkan oleh caelus-agent.
func (u *monitoringUsecase) IngestTelemetry(ctx context.Context, payload *domain.TelemetryReportPayload) error {
	server, err := u.serverRepo.GetByID(ctx, payload.ServerID)
	if err != nil {
		return fmt.Errorf("server not found for telemetry ingestion: %w", err)
	}

	containersJSONBytes, _ := json.Marshal(payload.Containers)

	metric := &domain.ServerMetric{
		ServerID:           payload.ServerID,
		CPUUsagePct:        payload.Host.CPUUsagePct,
		MemoryUsedMB:       int64(payload.Host.MemoryUsedMB),
		MemoryTotalMB:      int64(payload.Host.MemoryTotalMB),
		MemoryUsagePct:     payload.Host.MemoryUsagePct,
		DiskUsedGB:         payload.Host.DiskUsedGB,
		DiskTotalGB:        payload.Host.DiskTotalGB,
		DiskUsagePct:       payload.Host.DiskUsagePct,
		NetworkInKB:        int64(payload.Host.NetworkInKB),
		NetworkOutKB:       int64(payload.Host.NetworkOutKB),
		NetworkInRateKBps:  payload.Host.NetworkInRateKBps,
		NetworkOutRateKBps: payload.Host.NetworkOutRateKBps,
		UptimeSeconds:      int64(payload.Host.UptimeSeconds),
		ContainersCount:    len(payload.Containers),
		DockerAvailable:    payload.DockerAvailable,
		ContainersJSON:     string(containersJSONBytes),
		RecordedAt:         payload.Timestamp,
	}

	if err := u.metricRepo.Create(ctx, metric); err != nil {
		return fmt.Errorf("failed to persist server metric: %w", err)
	}

	needsUpdate := false
	if server.Status != domain.ServerStatusRunning {
		server.Status = domain.ServerStatusRunning
		needsUpdate = true
	}
	if payload.Host.CPUCores > 0 && server.CPUCores != payload.Host.CPUCores {
		server.CPUCores = payload.Host.CPUCores
		needsUpdate = true
	}
	if payload.Host.MemoryTotalMB > 0 && server.MemoryMB != int(payload.Host.MemoryTotalMB) {
		server.MemoryMB = int(payload.Host.MemoryTotalMB)
		needsUpdate = true
	}
	if payload.Host.DiskTotalGB > 0 && server.DiskGB != int(payload.Host.DiskTotalGB) {
		server.DiskGB = int(payload.Host.DiskTotalGB)
		needsUpdate = true
	}
	if payload.Host.Platform != "" && server.OSType != payload.Host.Platform {
		server.OSType = payload.Host.Platform
		needsUpdate = true
	}
	if payload.Host.Hostname != "" && (server.Hostname == nil || *server.Hostname != payload.Host.Hostname) {
		h := payload.Host.Hostname
		server.Hostname = &h
		needsUpdate = true
	}
	if needsUpdate {
		_ = u.serverRepo.Update(ctx, server)
	}

	if u.evaluator != nil {
		_ = u.evaluator.EvaluateMetrics(ctx, server, metric)
	}

	if u.broadcaster != nil {
		u.broadcaster.BroadcastToServer(server.ID, "metrics.updated", metric)
		u.broadcaster.BroadcastToOrg(server.OrganizationID, "server.telemetry", map[string]any{
			"server_id":   server.ID,
			"cpu_pct":     metric.CPUUsagePct,
			"mem_pct":     metric.MemoryUsagePct,
			"disk_pct":    metric.DiskUsagePct,
			"recorded_at": metric.RecordedAt,
		})
	}

	return nil
}

// GetServerLiveMetrics mengambil snapshot metrik terbaru dari database.
func (u *monitoringUsecase) GetServerLiveMetrics(ctx context.Context, serverID uuid.UUID) (*domain.ServerMetric, error) {
	return u.metricRepo.GetLatestByServerID(ctx, serverID)
}

// GetServerMetricHistory mengambil deret waktu metrik server dalam rentang waktu yang ditentukan.
func (u *monitoringUsecase) GetServerMetricHistory(ctx context.Context, serverID uuid.UUID, duration time.Duration) ([]domain.ServerMetric, error) {
	if duration <= 0 {
		duration = 1 * time.Hour
	}
	now := time.Now().UTC()
	from := now.Add(-duration)

	return u.metricRepo.GetHistoryByServerID(ctx, serverID, from, now, 100)
}

// ListAlerts mengambil daftar peringatan aktif atau riwayat berdasarkan filter status.
func (u *monitoringUsecase) ListAlerts(ctx context.Context, orgID uuid.UUID, status *domain.AlertStatus, page, limit int) ([]domain.Alert, int64, error) {
	return u.alertRepo.ListAlertsByOrg(ctx, orgID, status, page, limit)
}

// AcknowledgeAlert menandai alert telah ditinjau oleh pengguna.
func (u *monitoringUsecase) AcknowledgeAlert(ctx context.Context, alertID, userID uuid.UUID) error {
	now := time.Now().UTC()
	return u.alertRepo.UpdateAlertStatus(ctx, alertID, domain.AlertStatusAcknowledged, &userID, &now)
}

// ResolveAlert menandai insiden alert telah terselesaikan.
func (u *monitoringUsecase) ResolveAlert(ctx context.Context, alertID, userID uuid.UUID) error {
	now := time.Now().UTC()
	return u.alertRepo.UpdateAlertStatus(ctx, alertID, domain.AlertStatusResolved, &userID, &now)
}

// CreateAlertRule membuat aturan ambang batas alert baru.
func (u *monitoringUsecase) CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	if rule.Name == "" {
		return fmt.Errorf("%w: alert rule name is required", domain.ErrValidation)
	}
	if rule.MetricName == "" {
		return fmt.Errorf("%w: metric name is required", domain.ErrValidation)
	}
	if rule.Threshold <= 0 {
		return fmt.Errorf("%w: threshold must be greater than 0", domain.ErrValidation)
	}
	return u.alertRepo.CreateRule(ctx, rule)
}

// ListAlertRules mengambil seluruh aturan evaluasi alert milik organisasi.
func (u *monitoringUsecase) ListAlertRules(ctx context.Context, orgID uuid.UUID) ([]domain.AlertRule, error) {
	return u.alertRepo.ListRulesByOrg(ctx, orgID)
}

// DeleteAlertRule menghapus aturan alert berdasarkan ID.
func (u *monitoringUsecase) DeleteAlertRule(ctx context.Context, ruleID uuid.UUID) error {
	return u.alertRepo.DeleteRule(ctx, ruleID)
}
