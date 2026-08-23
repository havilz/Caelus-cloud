package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ServerMetric merepresentasikan entitas rekaman metrik telemetri performa server time-series.
type ServerMetric struct {
	ID                 int64     `json:"id"`
	ServerID           uuid.UUID `json:"server_id"`
	CPUUsagePct        float64   `json:"cpu_usage_pct"`
	MemoryUsedMB       int64     `json:"memory_used_mb"`
	MemoryTotalMB      int64     `json:"memory_total_mb"`
	MemoryUsagePct     float64   `json:"memory_usage_pct"`
	DiskUsedGB         float64   `json:"disk_used_gb"`
	DiskTotalGB        float64   `json:"disk_total_gb"`
	DiskUsagePct       float64   `json:"disk_usage_pct"`
	NetworkInKB        int64     `json:"network_in_kb"`
	NetworkOutKB       int64     `json:"network_out_kb"`
	NetworkInRateKBps  float64   `json:"network_in_rate_kbps"`
	NetworkOutRateKBps float64   `json:"network_out_rate_kbps"`
	UptimeSeconds      int64     `json:"uptime_seconds"`
	ContainersCount    int       `json:"containers_count"`
	DockerAvailable    bool      `json:"docker_available"`
	ContainersJSON     string    `json:"containers_json,omitempty"`
	RecordedAt         time.Time `json:"recorded_at"`
}

// HostMetricsPayload merepresentasikan payload data metrik host dari agent.
type HostMetricsPayload struct {
	CPUUsagePct        float64 `json:"cpu_usage_pct"`
	CPUCores           int     `json:"cpu_cores"`
	LoadAvg1m          float64 `json:"load_avg_1m"`
	LoadAvg5m          float64 `json:"load_avg_5m"`
	LoadAvg15m         float64 `json:"load_avg_15m"`
	MemoryTotalMB      uint64  `json:"memory_total_mb"`
	MemoryUsedMB       uint64  `json:"memory_used_mb"`
	MemoryFreeMB       uint64  `json:"memory_free_mb"`
	MemoryAvailableMB  uint64  `json:"memory_available_mb"`
	MemoryUsagePct     float64 `json:"memory_usage_pct"`
	DiskTotalGB        float64 `json:"disk_total_gb"`
	DiskUsedGB         float64 `json:"disk_used_gb"`
	DiskFreeGB         float64 `json:"disk_free_gb"`
	DiskUsagePct       float64 `json:"disk_usage_pct"`
	NetworkInKB        uint64  `json:"network_in_kb"`
	NetworkOutKB       uint64  `json:"network_out_kb"`
	NetworkInRateKBps  float64 `json:"network_in_rate_kbps"`
	NetworkOutRateKBps float64 `json:"network_out_rate_kbps"`
	UptimeSeconds      uint64  `json:"uptime_seconds"`
	OS                 string  `json:"os"`
	Platform           string  `json:"platform"`
	Hostname           string  `json:"hostname"`
}

// ContainerMetricPayload merepresentasikan metrik container yang dilaporkan oleh agent.
type ContainerMetricPayload struct {
	ID            string   `json:"id"`
	Names         []string `json:"names"`
	Image         string   `json:"image"`
	State         string   `json:"state"`
	Status        string   `json:"status"`
	Created       int64    `json:"created"`
	CPUUsagePct   float64  `json:"cpu_usage_pct"`
	MemoryUsageMB float64  `json:"memory_usage_mb"`
	MemoryLimitMB float64  `json:"memory_limit_mb"`
}

// TelemetryReportPayload merepresentasikan DTO payload lengkap yang dikirim oleh daemon caelus-agent.
type TelemetryReportPayload struct {
	ServerID        uuid.UUID                `json:"server_id"`
	Timestamp       time.Time                `json:"timestamp"`
	Host            HostMetricsPayload       `json:"host"`
	Containers      []ContainerMetricPayload `json:"containers"`
	DockerAvailable bool                     `json:"docker_available"`
}

// MetricRepository mendefinisikan kontrak persistensi data time-series metrik server.
type MetricRepository interface {
	Create(ctx context.Context, metric *ServerMetric) error
	GetLatestByServerID(ctx context.Context, serverID uuid.UUID) (*ServerMetric, error)
	GetHistoryByServerID(ctx context.Context, serverID uuid.UUID, from, to time.Time, limit int) ([]ServerMetric, error)
}

// TelemetryBroadcaster mendefinisikan kontrak pengiriman data telemetri dan status server secara realtime ke client.
type TelemetryBroadcaster interface {
	BroadcastToServer(serverID uuid.UUID, event string, data any)
	BroadcastToOrg(orgID uuid.UUID, event string, data any)
}

