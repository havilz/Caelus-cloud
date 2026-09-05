package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/repository/postgres"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

func main() {
	direction := flag.String("direction", "up", "Migration direction: up")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.App.LogLevel, cfg.App.Debug)

	if cfg.Database.Host == "" {
		logger.Error("DB_HOST configuration not set in environment")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := postgres.NewClient(ctx, &cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	migrationsDir := filepath.Join("migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = filepath.Join("..", "..", "migrations")
	}

	migrator := postgres.NewMigrator(client.Pool)

	if *direction == "up" {
		if err := migrator.Up(ctx, migrationsDir); err != nil {
			logger.Error("Migration up execution failed", "error", err)
			os.Exit(1)
		}
		logger.Info("All database migrations applied successfully")
	}
}
