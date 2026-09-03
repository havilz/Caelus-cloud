package tests

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/agent/internal/config"
)

func TestLoadConfig_MissingServerID(t *testing.T) {
	os.Clearenv()

	_, err := config.LoadConfig()
	if err != config.ErrMissingServerID {
		t.Fatalf("expected ErrMissingServerID, got: %v", err)
	}
}

func TestLoadConfig_InvalidServerID(t *testing.T) {
	os.Clearenv()
	t.Setenv("SERVER_ID", "invalid-uuid-string")

	_, err := config.LoadConfig()
	if err != config.ErrInvalidServerID {
		t.Fatalf("expected ErrInvalidServerID, got: %v", err)
	}
}

func TestLoadConfig_MissingAgentSecret(t *testing.T) {
	os.Clearenv()
	validUUID := uuid.New().String()
	t.Setenv("SERVER_ID", validUUID)

	_, err := config.LoadConfig()
	if err != config.ErrMissingAgentSecret {
		t.Fatalf("expected ErrMissingAgentSecret, got: %v", err)
	}
}

func TestLoadConfig_SuccessWithDefaults(t *testing.T) {
	os.Clearenv()
	validUUID := uuid.New()
	t.Setenv("SERVER_ID", validUUID.String())
	t.Setenv("AGENT_SECRET", "super-secret-agent-key")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.ServerID != validUUID {
		t.Errorf("expected serverID %v, got %v", validUUID, cfg.ServerID)
	}
	if cfg.AgentSecret != "super-secret-agent-key" {
		t.Errorf("expected secret 'super-secret-agent-key', got '%s'", cfg.AgentSecret)
	}
	if cfg.APIEndpoint != "http://localhost:8080" {
		t.Errorf("expected default api endpoint 'http://localhost:8080', got '%s'", cfg.APIEndpoint)
	}
	if cfg.CollectionIntervalSec != 15 {
		t.Errorf("expected default interval 15, got %d", cfg.CollectionIntervalSec)
	}
	if cfg.DockerSocketPath != "/var/run/docker.sock" {
		t.Errorf("expected default socket '/var/run/docker.sock', got '%s'", cfg.DockerSocketPath)
	}
	if cfg.TLSSkipVerify != false {
		t.Errorf("expected TLSSkipVerify false, got true")
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	os.Clearenv()
	validUUID := uuid.New()
	t.Setenv("SERVER_ID", validUUID.String())
	t.Setenv("AGENT_SECRET", "custom-secret")
	t.Setenv("API_ENDPOINT", "https://api.caelus.cloud")
	t.Setenv("COLLECTION_INTERVAL_SEC", "30")
	t.Setenv("DOCKER_SOCKET_PATH", "/custom/docker.sock")
	t.Setenv("TLS_SKIP_VERIFY", "true")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.APIEndpoint != "https://api.caelus.cloud" {
		t.Errorf("expected 'https://api.caelus.cloud', got '%s'", cfg.APIEndpoint)
	}
	if cfg.CollectionIntervalSec != 30 {
		t.Errorf("expected 30, got %d", cfg.CollectionIntervalSec)
	}
	if cfg.DockerSocketPath != "/custom/docker.sock" {
		t.Errorf("expected '/custom/docker.sock', got '%s'", cfg.DockerSocketPath)
	}
	if !cfg.TLSSkipVerify {
		t.Errorf("expected TLSSkipVerify true, got false")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected logLevel 'debug', got '%s'", cfg.LogLevel)
	}
}
