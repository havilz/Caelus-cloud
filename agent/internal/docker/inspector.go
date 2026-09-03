package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/havilz/caelus-cloud/agent/internal/transport"
)

type Inspector interface {
	IsAvailable(ctx context.Context) bool
	InspectContainers(ctx context.Context) ([]transport.ContainerMetrics, error)
	InspectNetworks(ctx context.Context) ([]transport.DiscoveredNetwork, error)
	InspectVolumes(ctx context.Context) ([]transport.DiscoveredVolume, error)
	CreateVolume(ctx context.Context, name string) error
	RemoveVolume(ctx context.Context, name string) error
	DeployContainer(ctx context.Context, payload transport.ContainerDeployPayload) error
	RemoveContainer(ctx context.Context, name string) error
	StartContainer(ctx context.Context, name string) error
	StopContainer(ctx context.Context, name string) error
	RestartContainer(ctx context.Context, name string) error
}

type UnixSocketInspector struct {
	socketPath string
	client     *http.Client
}

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
			Timeout:   10 * time.Second,
		},
	}
}

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
			metric.Logs = i.fetchContainerLogs(ctx, c.ID)
		}

		result = append(result, metric)
	}

	return result, nil
}

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

func (i *UnixSocketInspector) fetchContainerLogs(ctx context.Context, containerID string) []string {
	url := fmt.Sprintf("http://localhost/containers/%s/logs?stdout=1&stderr=1&tail=15&timestamps=0", containerID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16384))
	if err != nil || len(raw) == 0 {
		return nil
	}

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 8 && (line[0] <= 2) && line[1] == 0 && line[2] == 0 && line[3] == 0 {
			line = line[8:]
		}
		str := strings.TrimSpace(string(line))
		if str != "" {
			lines = append(lines, str)
		}
	}
	return lines
}

func (i *UnixSocketInspector) CreateVolume(ctx context.Context, name string) error {
	if !i.IsAvailable(ctx) {
		return fmt.Errorf("docker daemon is not available")
	}

	body := fmt.Sprintf(`{"Name":"%s","Driver":"local"}`, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/volumes/create", strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build create volume request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute create volume request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("docker API returned status %d when creating volume %s", resp.StatusCode, name)
	}
	return nil
}

func (i *UnixSocketInspector) RemoveVolume(ctx context.Context, name string) error {
	if !i.IsAvailable(ctx) {
		return fmt.Errorf("docker daemon is not available")
	}

	url := fmt.Sprintf("http://localhost/volumes/%s?force=1", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build remove volume request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute remove volume request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("docker API returned status %d when removing volume %s", resp.StatusCode, name)
	}
	return nil
}

func (i *UnixSocketInspector) DeployContainer(ctx context.Context, payload transport.ContainerDeployPayload) error {
	if !i.IsAvailable(ctx) {
		return fmt.Errorf("docker daemon is not available")
	}

	containerName := strings.TrimSpace(payload.Name)
	if containerName == "" {
		return fmt.Errorf("container name cannot be empty")
	}
	imageName := strings.TrimSpace(payload.Image)
	if imageName == "" {
		return fmt.Errorf("image name cannot be empty")
	}

	pullTransport := &http.Transport{
		DialContext: func(pCtx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(pCtx, "unix", i.socketPath)
		},
		DisableCompression: true,
	}
	pullClient := &http.Client{
		Transport: pullTransport,
		Timeout:   5 * time.Minute,
	}
	pullURL := fmt.Sprintf("http://localhost/images/create?fromImage=%s", url.QueryEscape(imageName))
	pullReq, err := http.NewRequestWithContext(ctx, http.MethodPost, pullURL, nil)
	if err != nil {
		return fmt.Errorf("failed building image pull request: %w", err)
	}
	pullResp, pErr := pullClient.Do(pullReq)
	if pErr != nil {
		return fmt.Errorf("image pull request failed for %s: %w", imageName, pErr)
	}
	_, _ = io.Copy(io.Discard, pullResp.Body)
	_ = pullResp.Body.Close()
	if pullResp.StatusCode != http.StatusOK {
		return fmt.Errorf("docker image pull returned status %d for image %s", pullResp.StatusCode, imageName)
	}

	_ = i.RemoveContainer(ctx, containerName)

	exposedPorts := make(map[string]struct{})
	portBindings := make(map[string][]map[string]string)

	for _, p := range payload.Ports {
		parts := strings.Split(p, ":")
		if len(parts) == 2 {
			hostPort := strings.TrimSpace(parts[0])
			containerPort := strings.TrimSpace(parts[1])
			if !strings.Contains(containerPort, "/") {
				containerPort += "/tcp"
			}
			exposedPorts[containerPort] = struct{}{}
			portBindings[containerPort] = []map[string]string{
				{"HostPort": hostPort},
			}
		}
	}

	var envList []string
	for k, v := range payload.Environment {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}

	restartPolicyName := strings.TrimSpace(payload.RestartPolicy)
	if restartPolicyName == "" {
		restartPolicyName = "unless-stopped"
	}

	createPayload := map[string]interface{}{
		"Image":        imageName,
		"ExposedPorts": exposedPorts,
		"Env":          envList,
		"HostConfig": map[string]interface{}{
			"PortBindings": portBindings,
			"Binds":        payload.Volumes,
			"RestartPolicy": map[string]string{
				"Name": restartPolicyName,
			},
		},
	}

	bodyBytes, err := json.Marshal(createPayload)
	if err != nil {
		return fmt.Errorf("failed marshaling container create payload: %w", err)
	}

	createURL := fmt.Sprintf("http://localhost/containers/create?name=%s", url.QueryEscape(containerName))
	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, createURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed building container create request: %w", err)
	}
	createReq.Header.Set("Content-Type", "application/json")

	createResp, err := i.client.Do(createReq)
	if err != nil {
		return fmt.Errorf("failed executing container create request: %w", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusOK && createResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(io.LimitReader(createResp.Body, 1024))
		return fmt.Errorf("docker API returned status %d when creating container %s: %s", createResp.StatusCode, containerName, string(respBody))
	}

	var createResult struct {
		ID string `json:"Id"`
	}
	_ = json.NewDecoder(createResp.Body).Decode(&createResult)

	startID := containerName
	if createResult.ID != "" {
		startID = createResult.ID
	}

	startURL := fmt.Sprintf("http://localhost/containers/%s/start", url.PathEscape(startID))
	startReq, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, nil)
	if err != nil {
		return fmt.Errorf("failed building container start request: %w", err)
	}

	startResp, err := i.client.Do(startReq)
	if err != nil {
		return fmt.Errorf("failed executing container start request: %w", err)
	}
	defer startResp.Body.Close()

	if startResp.StatusCode != http.StatusOK && startResp.StatusCode != http.StatusNoContent && startResp.StatusCode != http.StatusNotModified {
		respBody, _ := io.ReadAll(io.LimitReader(startResp.Body, 1024))
		return fmt.Errorf("docker API returned status %d when starting container %s: %s", startResp.StatusCode, containerName, string(respBody))
	}

	return nil
}

func (i *UnixSocketInspector) RemoveContainer(ctx context.Context, name string) error {
	if !i.IsAvailable(ctx) {
		return fmt.Errorf("docker daemon is not available")
	}

	url := fmt.Sprintf("http://localhost/containers/%s?force=1", url.PathEscape(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build remove container request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute remove container request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("docker API returned status %d when removing container %s", resp.StatusCode, name)
	}
	return nil
}

func (i *UnixSocketInspector) StartContainer(ctx context.Context, name string) error {
	if !i.IsAvailable(ctx) {
		return fmt.Errorf("docker daemon is not available")
	}

	url := fmt.Sprintf("http://localhost/containers/%s/start", url.PathEscape(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build start container request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute start container request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotModified {
		return fmt.Errorf("docker API returned status %d when starting container %s", resp.StatusCode, name)
	}
	return nil
}

func (i *UnixSocketInspector) StopContainer(ctx context.Context, name string) error {
	if !i.IsAvailable(ctx) {
		return fmt.Errorf("docker daemon is not available")
	}

	url := fmt.Sprintf("http://localhost/containers/%s/stop?t=10", url.PathEscape(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build stop container request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute stop container request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotModified && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("docker API returned status %d when stopping container %s", resp.StatusCode, name)
	}
	return nil
}

func (i *UnixSocketInspector) RestartContainer(ctx context.Context, name string) error {
	if !i.IsAvailable(ctx) {
		return fmt.Errorf("docker daemon is not available")
	}

	url := fmt.Sprintf("http://localhost/containers/%s/restart?t=10", url.PathEscape(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build restart container request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute restart container request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("docker API returned status %d when restarting container %s", resp.StatusCode, name)
	}
	return nil
}
