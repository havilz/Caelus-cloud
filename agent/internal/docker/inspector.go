package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/havilz/caelus-cloud/agent/internal/transport"
)

// Inspector mendefinisikan interface inspeksi status dan metrik Docker daemon.
type Inspector interface {
	IsAvailable(ctx context.Context) bool
	InspectContainers(ctx context.Context) ([]transport.ContainerMetrics, error)
}

// UnixSocketInspector mengimplementasikan Inspector melalui komunikasi Unix domain socket ke Docker daemon.
type UnixSocketInspector struct {
	socketPath string
	client     *http.Client
}

// NewInspector membuat instance baru UnixSocketInspector menggunakan path socket yang ditentukan.
func NewInspector(socketPath string) *UnixSocketInspector {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
		DisableCompression: true,
	}

	return &UnixSocketInspector{
		socketPath: socketPath,
		client: &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
		},
	}
}

// IsAvailable memeriksa apakah socket Docker ada dan merespons ping status.
func (i *UnixSocketInspector) IsAvailable(ctx context.Context) bool {
	if _, err := os.Stat(i.socketPath); os.IsNotExist(err) {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/_ping", nil)
	if err != nil {
		return false
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// InspectContainers mengambil daftar seluruh container Docker dan metrik utilisasinya.
func (i *UnixSocketInspector) InspectContainers(ctx context.Context) ([]transport.ContainerMetrics, error) {
	if !i.IsAvailable(ctx) {
		return []transport.ContainerMetrics{}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/containers/json?all=1", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker containers request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker containers request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker API returned status code %d", resp.StatusCode)
	}

	var rawContainers []struct {
		ID      string   `json:"Id"`
		Names   []string `json:"Names"`
		Image   string   `json:"Image"`
		State   string   `json:"State"`
		Status  string   `json:"Status"`
		Created int64    `json:"Created"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawContainers); err != nil {
		return nil, fmt.Errorf("failed to decode docker containers json: %w", err)
	}

	result := make([]transport.ContainerMetrics, 0, len(rawContainers))
	for _, c := range rawContainers {
		metric := transport.ContainerMetrics{
			ID:      c.ID,
			Names:   c.Names,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: c.Created,
		}

		if c.State == "running" {
			cpuPct, memMB, limitMB := i.fetchContainerStats(ctx, c.ID)
			metric.CPUUsagePct = math.Round(cpuPct*100) / 100
			metric.MemoryUsageMB = math.Round(memMB*100) / 100
			metric.MemoryLimitMB = math.Round(limitMB*100) / 100
		}

		result = append(result, metric)
	}

	return result, nil
}

// fetchContainerStats mengambil ringkasan metrik CPU dan RAM container yang sedang berjalan.
func (i *UnixSocketInspector) fetchContainerStats(ctx context.Context, containerID string) (float64, float64, float64) {
	url := fmt.Sprintf("http://localhost/containers/%s/stats?stream=false", containerID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, 0, 0
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return 0, 0, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, 0
	}

	var stats struct {
		CPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
			OnlineCPUs     uint32 `json:"online_cpus"`
		} `json:"cpu_stats"`
		PreCPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
		MemoryStats struct {
			Usage uint64 `json:"usage"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return 0, 0, 0
	}

	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)

	var cpuPercent float64
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		onlineCPUs := float64(stats.CPUStats.OnlineCPUs)
		if onlineCPUs == 0 {
			onlineCPUs = 1.0
		}
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}

	memMB := float64(stats.MemoryStats.Usage) / (1024.0 * 1024.0)
	limitMB := float64(stats.MemoryStats.Limit) / (1024.0 * 1024.0)

	return cpuPercent, memMB, limitMB
}
