package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func loadEnvFiles() {
	dir, err := os.Getwd()
	if err != nil {
		_ = godotenv.Load()
		return
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if info, err := os.Stat(envPath); err == nil && !info.IsDir() {
			_ = godotenv.Load(envPath)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	_ = godotenv.Load()
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Storage  StorageConfig
}

type AppConfig struct {
	Env         string
	Name        string
	Host        string
	Port        string
	Debug       bool
	LogLevel    string
	CorsOrigins    []string
	TrustedProxies []string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret            string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
	EncryptionKey     string
}

type StorageConfig struct {
	Driver    string
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
}

func Load() (*Config, error) {
	loadEnvFiles()

	cfg := &Config{
		App: AppConfig{
			Env:         getEnv("APP_ENV", "development"),
			Name:        getEnv("APP_NAME", "caelus-cloud-api"),
			Host:        getEnv("APP_HOST", "0.0.0.0"),
			Port:        getEnv("APP_PORT", "8080"),
			Debug:       getEnvAsBool("APP_DEBUG", true),
			LogLevel:    getEnv("APP_LOG_LEVEL", "debug"),
			CorsOrigins:    getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			TrustedProxies: getEnvAsSlice("TRUSTED_PROXIES", []string{"127.0.0.1", "::1"}),
		},
		Database: DatabaseConfig{
			Host:            os.Getenv("DB_HOST"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            os.Getenv("DB_USER"),
			Password:        os.Getenv("DB_PASSWORD"),
			DBName:          os.Getenv("DB_NAME"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:            os.Getenv("JWT_SECRET"),
			AccessExpiration:  getEnvAsDuration("JWT_ACCESS_EXPIRATION", 15*time.Minute),
			RefreshExpiration: getEnvAsDuration("JWT_REFRESH_EXPIRATION", 7*24*time.Hour),
			EncryptionKey:     os.Getenv("ENCRYPTION_KEY"),
		},
		Storage: StorageConfig{
			Driver:    getEnv("STORAGE_DRIVER", "s3"),
			Endpoint:  os.Getenv("STORAGE_ENDPOINT"),
			AccessKey: os.Getenv("STORAGE_ACCESS_KEY"),
			SecretKey: os.Getenv("STORAGE_SECRET_KEY"),
			Bucket:    os.Getenv("STORAGE_BUCKET"),
			Region:    getEnv("STORAGE_REGION", "us-east-1"),
			UseSSL:    getEnvAsBool("STORAGE_USE_SSL", false),
		},
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return errors.New("variabel lingkungan wajib belum diisi: JWT_SECRET")
	}
	if len(c.JWT.Secret) < 32 {
		return errors.New("keamanan JWT_SECRET tidak memenuhi syarat: panjang kunci minimal 32 karakter")
	}
	if c.JWT.EncryptionKey == "" {
		return errors.New("variabel lingkungan wajib belum diisi: ENCRYPTION_KEY")
	}
	if len(c.JWT.EncryptionKey) < 32 {
		return errors.New("keamanan ENCRYPTION_KEY tidak memenuhi syarat: panjang kunci minimal 32 karakter")
	}
	if c.App.Env == "production" {
		if c.Database.Host == "" {
			return errors.New("DB_HOST wajib diisi pada lingkungan production")
		}
		if c.Database.User == "" {
			return errors.New("DB_USER wajib diisi pada lingkungan production")
		}
		if c.Database.Password == "" {
			return errors.New("DB_PASSWORD wajib diisi pada lingkungan production")
		}
		if c.Database.DBName == "" {
			return errors.New("DB_NAME wajib diisi pada lingkungan production")
		}
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	parts := strings.Split(valStr, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
