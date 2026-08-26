package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/havilz/caelus-cloud/agent/internal/transport"
)

// Inspector mendefinisikan interface inspeksi status, metrik, network, dan volume Docker daemon.
type Inspector interface {
	IsAvailable(ctx context.Context) bool
	InspectContainers(ctx context.Context) ([]transport.ContainerMetrics, error)
	InspectNetworks(ctx context.Context) ([]transport.DiscoveredNetwork, error)
	InspectVolumes(ctx context.Context) ([]transport.DiscoveredVolume, error)
}

// UnixSocketInspector mengimplementasikan Inspector melalui komunikasi Unix domain socket ke Docker daemon.
type UnixSocketInspector struct {
	socketPath string
	client     *http.Client
}

// NewInspector membuat instance baru UnixSocketInspector menggunakan path socket yang ditentukan.
func NewInspector(socketPath string) *UnixSocketInspector {
	tr := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
		DisableCompression: true,
	}

	return &UnixSocketInspector{
		socketPath: socketPath,
		client: &http.Client{
			Transport: tr,
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

// InspectContainers mengambil daftar seluruh container Docker beserta konfigurasi port, volume, dan network-nya.
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
		Ports   []struct {
			IP          string `json:"IP"`
			PrivatePort int    `json:"PrivatePort"`
			PublicPort  int    `json:"PublicPort"`
			Type        string `json:"Type"`
		} `json:"Ports"`
		Mounts []struct {
			Name        string `json:"Name"`
			Source      string `json:"Source"`
			Destination string `json:"Destination"`
			Mode        string `json:"Mode"`
			Type        string `json:"Type"`
		} `json:"Mounts"`
		NetworkSettings struct {
			Networks map[string]struct {
				NetworkID string `json:"NetworkID"`
				IPAddress string `json:"IPAddress"`
				Gateway   string `json:"Gateway"`
			} `json:"Networks"`
		} `json:"NetworkSettings"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawContainers); err != nil {
		return nil, fmt.Errorf("failed to decode docker containers json: %w", err)
	}

	result := make([]transport.ContainerMetrics, 0, len(rawContainers))
	for _, c := range rawContainers {
		var ports []transport.PortBindingInfo
		for _, p := range c.Ports {
			ports = append(ports, transport.PortBindingInfo{
				HostPort:      p.PublicPort,
				ContainerPort: p.PrivatePort,
				Protocol:      p.Type,
				HostIP:        p.IP,
			})
		}

		var mounts []transport.VolumeMountInfo
		for _, m := range c.Mounts {
			mounts = append(mounts, transport.VolumeMountInfo{
				Name:        m.Name,
				Source:      m.Source,
				Destination: m.Destination,
				Mode:        m.Mode,
				Type:        m.Type,
			})
		}

		var networks []string
		var ipAddress string
		for netName, netInfo := range c.NetworkSettings.Networks {
			networks = append(networks, netName)
			if ipAddress == "" && netInfo.IPAddress != "" {
				ipAddress = netInfo.IPAddress
			}
		}

		cleanNames := make([]string, 0, len(c.Names))
		for _, n := range c.Names {
			cleanNames = append(cleanNames, strings.TrimPrefix(n, "/"))
		}

		metric := transport.ContainerMetrics{
			ID:           c.ID,
			Names:        cleanNames,
			Image:        c.Image,
			State:        c.State,
			Status:       c.Status,
			Created:      c.Created,
			PortBindings: ports,
			VolumeMounts: mounts,
			Networks:     networks,
			IPAddress:    ipAddress,
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

// InspectNetworks mengambil seluruh konfigurasi Docker network / VPC bridge pada host.
func (i *UnixSocketInspector) InspectNetworks(ctx context.Context) ([]transport.DiscoveredNetwork, error) {
	if !i.IsAvailable(ctx) {
		return []transport.DiscoveredNetwork{}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/networks", nil)
	if err != nil {
		return nil, err
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("networks API returned status %d", resp.StatusCode)
	}

	var rawNetworks []struct {
		ID       string `json:"Id"`
		Name     string `json:"Name"`
		Driver   string `json:"Driver"`
		Scope    string `json:"Scope"`
		Internal bool   `json:"Internal"`
		IPAM     struct {
			Config []struct {
				Subnet  string `json:"Subnet"`
				Gateway string `json:"Gateway"`
			} `json:"Config"`
		} `json:"IPAM"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawNetworks); err != nil {
		return nil, err
	}

	result := make([]transport.DiscoveredNetwork, 0, len(rawNetworks))
	for _, n := range rawNetworks {
		subnet := ""
		gateway := ""
		if len(n.IPAM.Config) > 0 {
			subnet = n.IPAM.Config[0].Subnet
			gateway = n.IPAM.Config[0].Gateway
		}

		result = append(result, transport.DiscoveredNetwork{
			ID:         n.ID,
			Name:       n.Name,
			Driver:     n.Driver,
			Scope:      n.Scope,
			SubnetCIDR: subnet,
			Gateway:    gateway,
			Internal:   n.Internal,
		})
	}

	return result, nil
}

// InspectVolumes mengambil seluruh persistent block volumes yang ada pada host.
func (i *UnixSocketInspector) InspectVolumes(ctx context.Context) ([]transport.DiscoveredVolume, error) {
	if !i.IsAvailable(ctx) {
		return []transport.DiscoveredVolume{}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/volumes", nil)
	if err != nil {
		return nil, err
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("volumes API returned status %d", resp.StatusCode)
	}

	var rawData struct {
		Volumes []struct {
			Name       string `json:"Name"`
			Driver     string `json:"Driver"`
			Mountpoint string `json:"Mountpoint"`
			CreatedAt  string `json:"CreatedAt"`
			UsageData  struct {
				Size int64 `json:"Size"`
			} `json:"UsageData"`
		} `json:"Volumes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, err
	}

	result := make([]transport.DiscoveredVolume, 0, len(rawData.Volumes))
	for _, v := range rawData.Volumes {
		sizeGB := float64(v.UsageData.Size) / (1024.0 * 1024.0 * 1024.0)
		result = append(result, transport.DiscoveredVolume{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			SizeGB:     math.Round(sizeGB*100) / 100,
			CreatedAt:  v.CreatedAt,
		})
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
