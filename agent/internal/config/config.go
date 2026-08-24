package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

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
	serverIDStr := getFirstEnv("SERVER_ID", "CAELUS_SERVER_ID")
	if serverIDStr == "" {
		return nil, ErrMissingServerID
	}

	serverUUID, err := uuid.Parse(serverIDStr)
	if err != nil {
		return nil, ErrInvalidServerID
	}

	agentSecret := getFirstEnv("AGENT_SECRET", "CAELUS_AGENT_SECRET")
	if agentSecret == "" {
		return nil, ErrMissingAgentSecret
	}

	apiEndpoint := getFirstEnvWithFallback("http://localhost:8080", "API_ENDPOINT", "CAELUS_API_ENDPOINT")
	
	intervalSec := 15
	if intervalVal := getFirstEnv("COLLECTION_INTERVAL_SEC", "CAELUS_INTERVAL"); intervalVal != "" {
		if sec, err := strconv.Atoi(intervalVal); err == nil && sec > 0 {
			intervalSec = sec
		} else if strings.HasSuffix(intervalVal, "s") {
			if sec, err := strconv.Atoi(strings.TrimSuffix(intervalVal, "s")); err == nil && sec > 0 {
				intervalSec = sec
			}
		}
	}

	dockerSocketPath := getFirstEnvWithFallback("/var/run/docker.sock", "DOCKER_SOCKET_PATH", "CAELUS_DOCKER_SOCKET_PATH")
	tlsSkipVerify := getEnvAsBoolOrDefault("TLS_SKIP_VERIFY", false)
	logLevel := getFirstEnvWithFallback("info", "LOG_LEVEL", "CAELUS_LOG_LEVEL")

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

// getFirstEnv mengembalikan nilai dari key pertama yang tidak kosong.
func getFirstEnv(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return ""
}

// getFirstEnvWithFallback mengembalikan nilai dari key pertama yang tidak kosong atau fallback jika semua kosong.
func getFirstEnvWithFallback(fallback string, keys ...string) string {
	if val := getFirstEnv(keys...); val != "" {
		return val
	}
	return fallback
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
