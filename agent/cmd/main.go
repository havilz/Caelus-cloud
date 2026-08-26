package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/havilz/caelus-cloud/agent/internal/collector"
	"github.com/havilz/caelus-cloud/agent/internal/config"
	"github.com/havilz/caelus-cloud/agent/internal/docker"
	"github.com/havilz/caelus-cloud/agent/internal/transport"
	"github.com/havilz/caelus-cloud/agent/pkg/logger"
)

// main adalah titik masuk utama daemon caelus-agent.
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load agent configuration", "error", err)
		os.Exit(1)
	}

	appLogger := logger.NewLogger(cfg.LogLevel)
	slog.SetDefault(appLogger)

	slog.Info("starting caelus-agent daemon",
		"server_id", cfg.ServerID,
		"api_endpoint", cfg.APIEndpoint,
		"interval_sec", cfg.CollectionIntervalSec,
		"docker_socket", cfg.DockerSocketPath,
	)

	sysCollector := collector.NewCollector()
	dockerInspector := docker.NewInspector(cfg.DockerSocketPath)
	transportClient := transport.NewHTTPClient(cfg.APIEndpoint, cfg.ServerID, cfg.AgentSecret, cfg.TLSSkipVerify)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	runCollectionCycle(ctx, sysCollector, dockerInspector, transportClient, cfg)

	ticker := time.NewTicker(time.Duration(cfg.CollectionIntervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down caelus-agent gracefully")
			return
		case <-ticker.C:
			runCollectionCycle(ctx, sysCollector, dockerInspector, transportClient, cfg)
		}
	}
}

// runCollectionCycle menjalankan satu siklus pengumpulan data telemetri host, container, network, dan volume.
func runCollectionCycle(
	ctx context.Context,
	sysCollector collector.Collector,
	dockerInspector docker.Inspector,
	transportClient transport.Client,
	cfg *config.Config,
) {
	start := time.Now()

	hostMetrics, err := sysCollector.Collect(ctx)
	if err != nil {
		slog.Error("failed to collect host metrics", "error", err)
		return
	}

	dockerAvail := dockerInspector.IsAvailable(ctx)
	var containers []transport.ContainerMetrics
	var networks []transport.DiscoveredNetwork
	var volumes []transport.DiscoveredVolume

	if dockerAvail {
		containers, err = dockerInspector.InspectContainers(ctx)
		if err != nil {
			slog.Warn("failed to inspect docker containers", "error", err)
		}

		networks, err = dockerInspector.InspectNetworks(ctx)
		if err != nil {
			slog.Warn("failed to inspect docker networks", "error", err)
		}

		volumes, err = dockerInspector.InspectVolumes(ctx)
		if err != nil {
			slog.Warn("failed to inspect docker volumes", "error", err)
		}
	}

	payload := &transport.AgentReportPayload{
		ServerID:        cfg.ServerID,
		Timestamp:       time.Now().UTC(),
		Host:            *hostMetrics,
		Containers:      containers,
		Networks:        networks,
		Volumes:         volumes,
		DockerAvailable: dockerAvail,
	}

	sendCtx, sendCancel := context.WithTimeout(ctx, 15*time.Second)
	defer sendCancel()

	if err := transportClient.SendReport(sendCtx, payload); err != nil {
		slog.Error("failed to send telemetry report to api", "error", err)
		return
	}

	slog.Info("telemetry report dispatched successfully",
		"duration_ms", time.Since(start).Milliseconds(),
		"cpu_pct", hostMetrics.CPUUsagePct,
		"mem_pct", hostMetrics.MemoryUsagePct,
		"disk_pct", hostMetrics.DiskUsagePct,
		"containers_count", len(containers),
		"networks_count", len(networks),
		"volumes_count", len(volumes),
		"docker_active", dockerAvail,
	)
}
