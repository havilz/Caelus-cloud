package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliveryHttp "github.com/havilz/caelus-cloud/backend/internal/delivery/http"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

// main menginisialisasi konfigurasi sistem, logging terstruktur, router HTTP, dan menjalankan HTTP server dengan mekanisme graceful shutdown.
func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Gagal memuat konfigurasi: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.App.LogLevel, cfg.App.Debug)

	if err := cfg.Validate(); err != nil {
		logger.Warn("Peringatan konfigurasi environment", "error", err)
	}
	logger.Info("Menginisialisasi Caelus Cloud API",
		"app_name", cfg.App.Name,
		"env", cfg.App.Env,
		"port", cfg.App.Port,
	)

	router := deliveryHttp.NewRouter(cfg)

	serverAddr := fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Server HTTP Caelus Cloud berjalan", "addr", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Gagal menjalankan server HTTP", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

	logger.Info("Menerima sinyal terminasi, mematikan server secara graceful...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server dipaksa berhenti karena error pada shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server HTTP Caelus Cloud berhasil dimatikan secara aman")
}
