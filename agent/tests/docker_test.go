package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/havilz/caelus-cloud/agent/internal/docker"
)

func TestDockerInspector_UnavailableSocket(t *testing.T) {
	inspector := docker.NewInspector("/tmp/nonexistent_caelus_test.sock")
	ctx := context.Background()

	if inspector.IsAvailable(ctx) {
		t.Error("expected IsAvailable to be false for non-existent socket")
	}

	containers, err := inspector.InspectContainers(ctx)
	if err != nil {
		t.Fatalf("expected no error when docker unavailable, got: %v", err)
	}

	if len(containers) != 0 {
		t.Errorf("expected 0 containers, got %d", len(containers))
	}
}

func TestDockerInspector_MockUnixSocket(t *testing.T) {
	socketDir := t.TempDir()
	socketPath := filepath.Join(socketDir, "docker_mock.sock")

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create unix socket listener: %v", err)
	}
	defer listener.Close()
	defer os.Remove(socketPath)

	mux := http.NewServeMux()
	mux.HandleFunc("/_ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	mux.HandleFunc("/containers/json", func(w http.ResponseWriter, _ *http.Request) {
		containers := []map[string]any{
			{
				"Id":      "c1a2b3c4d5e6",
				"Names":   []string{"/caelus-web"},
				"Image":   "nginx:alpine",
				"State":   "running",
				"Status":  "Up 3 hours",
				"Created": int64(1700000000),
			},
			{
				"Id":      "c9f8e7d6c5b4",
				"Names":   []string{"/caelus-db"},
				"Image":   "postgres:16",
				"State":   "exited",
				"Status":  "Exited (0) 10 minutes ago",
				"Created": int64(1700000000),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(containers)
	})

	mux.HandleFunc("/containers/c1a2b3c4d5e6/stats", func(w http.ResponseWriter, _ *http.Request) {
		stats := map[string]any{
			"cpu_stats": map[string]any{
				"cpu_usage": map[string]any{
					"total_usage": uint64(500000000),
				},
				"system_cpu_usage": uint64(1000000000),
				"online_cpus":      uint32(2),
			},
			"precpu_stats": map[string]any{
				"cpu_usage": map[string]any{
					"total_usage": uint64(400000000),
				},
				"system_cpu_usage": uint64(900000000),
			},
			"memory_stats": map[string]any{
				"usage": uint64(128 * 1024 * 1024),
				"limit": uint64(1024 * 1024 * 1024),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	})

	server := &http.Server{Handler: mux}
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Close()

	inspector := docker.NewInspector(socketPath)
	ctx := context.Background()

	if !inspector.IsAvailable(ctx) {
		t.Fatal("expected IsAvailable to be true for mock socket")
	}

	containers, err := inspector.InspectContainers(ctx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(containers))
	}

	runningContainer := containers[0]
	if runningContainer.ID != "c1a2b3c4d5e6" {
		t.Errorf("expected ID 'c1a2b3c4d5e6', got '%s'", runningContainer.ID)
	}
	if runningContainer.State != "running" {
		t.Errorf("expected state 'running', got '%s'", runningContainer.State)
	}
	if runningContainer.MemoryUsageMB <= 0 {
		t.Errorf("expected MemoryUsageMB > 0, got %f", runningContainer.MemoryUsageMB)
	}

	exitedContainer := containers[1]
	if exitedContainer.State != "exited" {
		t.Errorf("expected state 'exited', got '%s'", exitedContainer.State)
	}
	if exitedContainer.CPUUsagePct != 0.0 {
		t.Errorf("expected CPUUsagePct 0 for exited container, got %f", exitedContainer.CPUUsagePct)
	}
}

func TestDockerInspector_InvalidJSONResponse(t *testing.T) {
	socketDir := t.TempDir()
	socketPath := filepath.Join(socketDir, "docker_invalid.sock")

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create unix socket listener: %v", err)
	}
	defer listener.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/_ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/containers/json", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "{invalid-json")
	})

	server := &http.Server{Handler: mux}
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Close()

	inspector := docker.NewInspector(socketPath)
	ctx := context.Background()

	_, err = inspector.InspectContainers(ctx)
	if err == nil {
		t.Error("expected error for invalid JSON response, got nil")
	}
}
