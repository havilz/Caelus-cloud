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
	v1 "github.com/havilz/caelus-cloud/backend/internal/delivery/http/v1"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/ws"
	"github.com/havilz/caelus-cloud/backend/internal/observability/loki"
	"github.com/havilz/caelus-cloud/backend/internal/observability/prometheus"
	provFactory "github.com/havilz/caelus-cloud/backend/internal/provider"
	"github.com/havilz/caelus-cloud/backend/internal/repository/postgres"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/auth"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/monitoring"
	provUsecase "github.com/havilz/caelus-cloud/backend/internal/usecase/provider"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/server"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

// main menginisialisasi konfigurasi sistem, logging terstruktur, koneksi database, usecase, router HTTP, dan menjalankan server API.
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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := postgres.NewClient(ctx, &cfg.Database)
	if err != nil {
		logger.Error("Gagal menghubungkan ke database PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	userRepo := postgres.NewUserRepository(client.Pool)
	orgRepo := postgres.NewOrganizationRepository(client.Pool)
	serverRepo := postgres.NewServerRepository(client.Pool)
	providerRepo := postgres.NewProviderRepository(client.Pool)
	credRepo := postgres.NewCredentialRepository(client.Pool)
	auditRepo := postgres.NewAuditRepository(client.Pool)
	metricRepo := postgres.NewMetricRepository(client.Pool)
	alertRepo := postgres.NewAlertRepository(client.Pool)

	jwtManager := jwt.NewJWTManager(&cfg.JWT, cfg.App.Name)
	factory := provFactory.NewDriverFactory()
	wsHub := ws.NewHub()

	authUc := auth.NewAuthUsecase(userRepo, orgRepo, jwtManager)
	credUc := provUsecase.NewCredentialUsecase(credRepo, providerRepo, []byte(cfg.JWT.EncryptionKey))
	serverUc := server.NewServerUsecase(serverRepo, providerRepo, credRepo, factory)

	alertEvaluator := monitoring.NewAlertEvaluator(alertRepo, wsHub)
	promAdapter := prometheus.NewClient(os.Getenv("PROMETHEUS_URL"))
	lokiAdapter := loki.NewClient(os.Getenv("LOKI_URL"))
	monitoringUc := monitoring.NewMonitoringUsecase(metricRepo, alertRepo, serverRepo, alertEvaluator, wsHub, promAdapter, lokiAdapter)

	routerConfig := deliveryHttp.RouterConfig{
		Config:     cfg,
		JWTManager: jwtManager,
		AuditRepo:  auditRepo,
		Logger:     logger.Get(),
		Handlers: deliveryHttp.Handlers{
			AuthHandler:      v1.NewAuthHandler(authUc),
			ServerHandler:    v1.NewServerHandler(serverUc),
			ProviderHandler:  v1.NewProviderHandler(credUc),
			TelemetryHandler: v1.NewTelemetryHandler(monitoringUc),
			AlertHandler:     v1.NewAlertHandler(monitoringUc),
			WSHandler:        ws.NewHandler(wsHub, jwtManager),
		},
	}

	router := deliveryHttp.NewRouter(routerConfig)

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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server dipaksa berhenti karena error pada shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server HTTP Caelus Cloud berhasil dimatikan secara aman")
}
