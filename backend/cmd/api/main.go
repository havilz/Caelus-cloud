package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/automation"
	deliveryHttp "github.com/havilz/caelus-cloud/backend/internal/delivery/http"
	v1 "github.com/havilz/caelus-cloud/backend/internal/delivery/http/v1"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/ws"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/notification"
	"github.com/havilz/caelus-cloud/backend/internal/notification/email"
	"github.com/havilz/caelus-cloud/backend/internal/notification/webhook"
	"github.com/havilz/caelus-cloud/backend/internal/observability/loki"
	"github.com/havilz/caelus-cloud/backend/internal/observability/prometheus"
	provFactory "github.com/havilz/caelus-cloud/backend/internal/provider"
	provSync "github.com/havilz/caelus-cloud/backend/internal/provider/sync"
	"github.com/havilz/caelus-cloud/backend/internal/repository/postgres"
	"github.com/havilz/caelus-cloud/backend/internal/sentinel"
	storageFactory "github.com/havilz/caelus-cloud/backend/internal/storage"
	minioStorage "github.com/havilz/caelus-cloud/backend/internal/storage/minio"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/auth"
	automationUsecase "github.com/havilz/caelus-cloud/backend/internal/usecase/automation"
	backupUsecase "github.com/havilz/caelus-cloud/backend/internal/usecase/backup"
	iacApplier "github.com/havilz/caelus-cloud/backend/internal/iac/applier"
	iacUsecase "github.com/havilz/caelus-cloud/backend/internal/usecase/iac"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/monitoring"
	orchestrationUsecase "github.com/havilz/caelus-cloud/backend/internal/usecase/orchestration"
	provUsecase "github.com/havilz/caelus-cloud/backend/internal/usecase/provider"
	securityUsecase "github.com/havilz/caelus-cloud/backend/internal/usecase/security"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/server"
	storageUsecase "github.com/havilz/caelus-cloud/backend/internal/usecase/storage"
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
	bucketRepo := postgres.NewBucketRepository(client.Pool)
	backupRepo := postgres.NewBackupRepository(client.Pool)
	automationRepo := postgres.NewAutomationRepository(client.Pool)

	jwtManager := jwt.NewJWTManager(&cfg.JWT, cfg.App.Name)
	factory := provFactory.NewDriverFactoryWithKey([]byte(cfg.JWT.EncryptionKey))
	wsHub := ws.NewHub()

	// Inisialisasi Storage Factory & Adapters
	storageFactoryInstance := storageFactory.NewStorageFactory()
	minioEndpoint := os.Getenv("STORAGE_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "http://localhost:9000"
	}
	minioAccessKey := os.Getenv("STORAGE_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "minioadmin"
	}
	minioSecretKey := os.Getenv("STORAGE_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "minioadmin"
	}

	if minioAdapter, err := minioStorage.NewAdapter(minioStorage.Config{
		Endpoint:        minioEndpoint,
		AccessKeyID:     minioAccessKey,
		SecretAccessKey: minioSecretKey,
		Region:          "us-east-1",
	}); err == nil {
		storageFactoryInstance.RegisterAdapter(domain.StorageProviderMinIO, minioAdapter)
	}

	authUc := auth.NewAuthUsecase(userRepo, orgRepo, jwtManager)
	credUc := provUsecase.NewCredentialUsecase(credRepo, providerRepo, []byte(cfg.JWT.EncryptionKey))
	serverUc := server.NewServerUsecase(serverRepo, providerRepo, credRepo, factory)

	alertEvaluator := monitoring.NewAlertEvaluator(alertRepo, wsHub)
	promAdapter := prometheus.NewClient(os.Getenv("PROMETHEUS_URL"))
	lokiAdapter := loki.NewClient(os.Getenv("LOKI_URL"))
	monitoringUc := monitoring.NewMonitoringUsecase(metricRepo, alertRepo, serverRepo, alertEvaluator, wsHub, promAdapter, lokiAdapter)

	storageUc := storageUsecase.NewStorageUsecase(bucketRepo, storageFactoryInstance)
	backupUc := backupUsecase.NewBackupUsecase(backupRepo, serverRepo, bucketRepo, storageFactoryInstance)

	// Inisialisasi Automation & Notification Dispatcher
	webhookClient := webhook.NewClient(os.Getenv("WEBHOOK_SIGNING_SECRET"))
	emailClient := email.NewClient(email.Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     587,
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("SMTP_FROM"),
	})
	unifiedNotifier := notification.NewUnifiedDispatcher(webhookClient, emailClient)

	// Inisialisasi Central Event Dispatcher & Rule Engine
	centralDispatcher := automation.NewCentralEventDispatcher()
	ruleEngine := automation.NewEngine(automationRepo, nil, unifiedNotifier, serverUc, backupUc)
	centralDispatcher.Subscribe(func(ctx context.Context, event domain.SystemEvent) error {
		return ruleEngine.EvaluateEvent(ctx, event)
	})

	automationUc := automationUsecase.NewAutomationUsecase(automationRepo, ruleEngine, centralDispatcher)

	// Background Backup Scheduler & Retention Cleaner Worker
	backupScheduler := backupUsecase.NewScheduler(backupRepo, backupUc, logger.Get())
	backupScheduler.Start(60 * time.Second)
	defer backupScheduler.Stop()

	// Background Heartbeat Liveness Watchdog (Mendeteksi server mati setelah 15s tanpa telemetri)
	watchdog := monitoring.NewHeartbeatWatchdog(
		serverRepo,
		metricRepo,
		wsHub,
		func(ctx context.Context, event domain.SystemEvent) {
			centralDispatcher.Publish(ctx, event)
		},
		15*time.Second,
	)
	watchdog.Start()
	defer watchdog.Stop()

	// Background Multi-Provider Resource Status Sync Engine (Rekonsiliasi status setiap 60 detik)
	syncEngine := provSync.NewSyncEngine(
		serverRepo,
		providerRepo,
		credRepo,
		factory,
		func(ctx context.Context, event domain.SystemEvent) {
			centralDispatcher.Publish(ctx, event)
		},
		60*time.Second,
	)
	syncEngine.Start(context.Background())
	defer syncEngine.Stop()

	securityRepo := postgres.NewSecurityRepository(client.Pool)
	sentinelOrchestrator := sentinel.NewOrchestrator(securityRepo, func(ctx context.Context, event domain.SystemEvent) {
		centralDispatcher.Publish(ctx, event)
	})
	securityUc := securityUsecase.NewSecurityUsecase(securityRepo, serverRepo, metricRepo, sentinelOrchestrator)

	// Inisialisasi Repositori dan Usecase IaC & Orkestrasi Deployment
	iacRepo := postgres.NewIaCRepository(client.Pool)
	deploymentRepo := postgres.NewDeploymentRepository(client.Pool)
	iacUc := iacUsecase.NewUseCaseWithDeps(iacApplier.Dependencies{
		IaCRepo:        iacRepo,
		ServerRepo:     serverRepo,
		ProviderRepo:   providerRepo,
		BucketRepo:     bucketRepo,
		StorageFactory: storageFactoryInstance,
		DeploymentRepo: deploymentRepo,
		AutomationRepo: automationRepo,
	})
	deploymentUc := orchestrationUsecase.NewUseCase(deploymentRepo, wsHub)

	routerConfig := deliveryHttp.RouterConfig{
		Config:     cfg,
		JWTManager: jwtManager,
		AuditRepo:  auditRepo,
		Logger:     logger.Get(),
		Handlers: deliveryHttp.Handlers{
			AuthHandler:       v1.NewAuthHandler(authUc),
			ServerHandler:     v1.NewServerHandler(serverUc),
			ProviderHandler:   v1.NewProviderHandler(credUc),
			CredentialHandler: v1.NewCredentialHandler(credUc, factory),
			TelemetryHandler:  v1.NewTelemetryHandler(monitoringUc),
			AlertHandler:      v1.NewAlertHandler(monitoringUc),
			StorageHandler:    v1.NewStorageHandler(storageUc),
			BackupHandler:     v1.NewBackupHandler(backupUc),
			AutomationHandler: v1.NewAutomationHandler(automationUc),
			SecurityHandler:   v1.NewSecurityHandler(securityUc),
			IaCHandler:        v1.NewIaCHandler(iacUc),
			DeploymentHandler: v1.NewDeploymentHandler(deploymentUc),
			WSHandler:         ws.NewHandler(wsHub, jwtManager),
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
