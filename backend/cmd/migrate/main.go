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
	direction := flag.String("direction", "up", "Arah migrasi: up")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Gagal memuat konfigurasi: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.App.LogLevel, cfg.App.Debug)

	if cfg.Database.Host == "" {
		logger.Error("Konfigurasi DB_HOST belum diatur pada environment")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := postgres.NewClient(ctx, &cfg.Database)
	if err != nil {
		logger.Error("Gagal tersambung ke basis data", "error", err)
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
			logger.Error("Eksekusi migrasi up gagal", "error", err)
			os.Exit(1)
		}
		logger.Info("Seluruh migrasi basis data berhasil diaplikasikan")
	}
}
