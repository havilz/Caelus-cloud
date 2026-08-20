package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/google/uuid"
)

var (
	ErrMissingServerID   = errors.New("SERVER_ID is required and must not be empty")
	ErrInvalidServerID   = errors.New("SERVER_ID must be a valid UUID")
	ErrMissingAgentSecret = errors.New("AGENT_SECRET is required and must not be empty")
)

// Config merepresentasikan parameter konfigurasi operasional caelus-agent.
type Config struct {
	ServerID               uuid.UUID
	AgentSecret            string
	APIEndpoint            string
	CollectionIntervalSec int
	DockerSocketPath       string
	TLSSkipVerify          bool
	LogLevel               string
}

// LoadConfig membaca variabel lingkungan dan memvalidasi konfigurasi agent.
func LoadConfig() (*Config, error) {
	serverIDStr := os.Getenv("SERVER_ID")
	if serverIDStr == "" {
		return nil, ErrMissingServerID
	}

	serverUUID, err := uuid.Parse(serverIDStr)
	if err != nil {
		return nil, ErrInvalidServerID
	}

	agentSecret := os.Getenv("AGENT_SECRET")
	if agentSecret == "" {
		return nil, ErrMissingAgentSecret
	}

	apiEndpoint := getEnvOrDefault("API_ENDPOINT", "http://localhost:8080")
	intervalSec := getEnvAsIntOrDefault("COLLECTION_INTERVAL_SEC", 15)
	if intervalSec < 1 {
		intervalSec = 15
	}

	dockerSocketPath := getEnvOrDefault("DOCKER_SOCKET_PATH", "/var/run/docker.sock")
	tlsSkipVerify := getEnvAsBoolOrDefault("TLS_SKIP_VERIFY", false)
	logLevel := getEnvOrDefault("LOG_LEVEL", "info")

	return &Config{
		ServerID:               serverUUID,
		AgentSecret:            agentSecret,
		APIEndpoint:            apiEndpoint,
		CollectionIntervalSec: intervalSec,
		DockerSocketPath:       dockerSocketPath,
		TLSSkipVerify:          tlsSkipVerify,
		LogLevel:               logLevel,
	}, nil
}

// getEnvOrDefault mengambil nilai environment variable atau mengembalikan fallback default.
func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// getEnvAsIntOrDefault mengonversi nilai environment variable menjadi integer atau mengembalikan fallback default.
func getEnvAsIntOrDefault(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvAsBoolOrDefault mengonversi nilai environment variable menjadi boolean atau mengembalikan fallback default.
func getEnvAsBoolOrDefault(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return fallback
}
