package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/automation"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/notification"
	"github.com/havilz/caelus-cloud/backend/internal/notification/email"
	"github.com/havilz/caelus-cloud/backend/internal/notification/webhook"
	"github.com/havilz/caelus-cloud/backend/internal/queue"
	redisQueue "github.com/havilz/caelus-cloud/backend/internal/queue/redis"
	"github.com/havilz/caelus-cloud/backend/internal/repository/postgres"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Gagal memuat konfigurasi worker: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.App.LogLevel, cfg.App.Debug)
	logger.Info("Menginisialisasi Caelus Worker Daemon (caelus-worker)...",
		"env", cfg.App.Env,
		"redis_host", cfg.Redis.Host,
		"redis_port", cfg.Redis.Port,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbCtx, dbCancel := context.WithTimeout(ctx, 15*time.Second)
	defer dbCancel()

	client, err := postgres.NewClient(dbCtx, &cfg.Database)
	if err != nil {
		logger.Error("Gagal menghubungkan ke database PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	automationRepo := postgres.NewAutomationRepository(client.Pool)

	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	if cfg.Redis.Host == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		logger.Warn("Koneksi Redis tidak tersedia, worker berjalan dalam mode terbatas", "error", err)
	} else {
		logger.Info("Koneksi Redis berhasil tersambung", "addr", redisAddr)
	}
	defer rdb.Close()

	webhookClient := webhook.NewClient(os.Getenv("WEBHOOK_SIGNING_SECRET"))
	emailClient := email.NewClient(email.Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     587,
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("SMTP_FROM"),
	})
	notifier := notification.NewUnifiedDispatcher(webhookClient, emailClient)

	queueEngine := redisQueue.NewRedisQueueEngine(redisQueue.Config{
		Client:      rdb,
		Concurrency: 5,
		PollTimeout: 2 * time.Second,
	})

	ruleEngine := automation.NewEngine(automationRepo, queueEngine, notifier, nil, nil)

	registerWorkerHandlers(queueEngine, ruleEngine, notifier)

	if err := queueEngine.Start(ctx); err != nil {
		logger.Error("Gagal menjalankan worker queue engine", "error", err)
		os.Exit(1)
	}

	logger.Info("Caelus Worker siap memproses antrean pekerjaan terdistribusi")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Menerima sinyal terminasi, mematikan worker secara anggun...")

	cancel()
	_ = queueEngine.Stop()
	logger.Info("Caelus Worker berhasil dimatikan secara aman")
}

func registerWorkerHandlers(q queue.QueueEngine, engine automation.RuleEngine, notifier notification.Dispatcher) {

	q.RegisterHandler(queue.TaskTypeSendWebhookNotification, func(ctx context.Context, task *queue.TaskPayload) error {
		var payload struct {
			URL  string                 `json:"url"`
			Data webhook.WebhookPayload `json:"data"`
		}
		if err := json.Unmarshal(task.Data, &payload); err != nil {
			return fmt.Errorf("invalid webhook task data: %w", err)
		}
		return notifier.SendWebhook(ctx, payload.URL, payload.Data)
	})

	q.RegisterHandler(queue.TaskTypeSendEmailNotification, func(ctx context.Context, task *queue.TaskPayload) error {
		var payload email.EmailMessage
		if err := json.Unmarshal(task.Data, &payload); err != nil {
			return fmt.Errorf("invalid email task data: %w", err)
		}
		return notifier.SendEmail(ctx, payload)
	})

	q.RegisterHandler(queue.TaskTypeExecuteRuleAction, func(ctx context.Context, task *queue.TaskPayload) error {
		var payload struct {
			Rule      domain.AutomationRule `json:"rule"`
			Action    domain.RuleAction     `json:"action"`
			EventData map[string]any        `json:"event_data"`
		}
		if err := json.Unmarshal(task.Data, &payload); err != nil {
			return fmt.Errorf("invalid rule action task data: %w", err)
		}

		res := engine.ExecuteRuleAction(ctx, &payload.Rule, payload.Action, payload.EventData)
		if res.Status == "failed" {
			return fmt.Errorf("action %s execution failed: %s", payload.Action.Type, res.Error)
		}
		return nil
	})
}
