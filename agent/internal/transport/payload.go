package transport

import (
	"time"

	"github.com/google/uuid"
)

// HostMetrics merepresentasikan ringkasan metrik utilisasi perangkat keras host sistem.
type HostMetrics struct {
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

// ContainerMetrics merepresentasikan data status dan utilisasi resource sebuah container Docker.
type ContainerMetrics struct {
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

// AgentReportPayload adalah payload agregat yang dikirim oleh agent ke control plane Caelus API.
type AgentReportPayload struct {
	ServerID        uuid.UUID          `json:"server_id"`
	Timestamp       time.Time          `json:"timestamp"`
	Host            HostMetrics        `json:"host"`
	Containers      []ContainerMetrics `json:"containers"`
	DockerAvailable bool               `json:"docker_available"`
}
