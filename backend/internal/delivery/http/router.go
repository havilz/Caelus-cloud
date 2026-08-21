package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	customMiddleware "github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	v1 "github.com/havilz/caelus-cloud/backend/internal/delivery/http/v1"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/ws"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

type Handlers struct {
	AuthHandler       *v1.AuthHandler
	ServerHandler     *v1.ServerHandler
	ProviderHandler   *v1.ProviderHandler
	TelemetryHandler  *v1.TelemetryHandler
	AlertHandler      *v1.AlertHandler
	StorageHandler    *v1.StorageHandler
	BackupHandler     *v1.BackupHandler
	AutomationHandler *v1.AutomationHandler
	SecurityHandler   *v1.SecurityHandler
	WSHandler         *ws.Handler
}

type RouterConfig struct {
	Config     *config.Config
	JWTManager jwt.Manager
	AuditRepo  domain.AuditLogRepository
	Logger     *slog.Logger
	Handlers   Handlers
}

// NewRouter menginisialisasi router Chi dengan middleware global, endpoint kesehatan, rute publik auth, telemetri, dan rute terproteksi server/provider/alert.
func NewRouter(rc RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	if rc.Config != nil {
		r.Use(customMiddleware.CORS(rc.Config.App.CorsOrigins))
	}
	r.Use(customMiddleware.RequestLogger())

	registerHealthRoutes(r, rc.Config)
	registerAPIRoutes(r, rc)

	return r
}

// registerHealthRoutes mendaftarkan endpoint pemantauan kesehatan aplikasi (/health dan /api/v1/health).
func registerHealthRoutes(r *chi.Mux, cfg *config.Config) {
	serviceName := "caelus-cloud-api"
	envName := "development"
	if cfg != nil {
		serviceName = cfg.App.Name
		envName = cfg.App.Env
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, "Caelus Cloud API is healthy", map[string]string{
			"status":  "ok",
			"service": serviceName,
			"env":     envName,
		})
	})

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, "API v1 is operational", map[string]string{
			"status":  "ok",
			"version": "v1",
		})
	})

	r.Get("/install.sh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		script := `#!/usr/bin/env bash
set -e

echo "=== Caelus Cloud Agent Auto-Installer ==="

SERVER_ID=""
AGENT_SECRET=""
API_ENDPOINT="http://localhost:8080"

for arg in "$@"; do
  case $arg in
    --server-id=*) SERVER_ID="${arg#*=}" ;;
    --secret=*) AGENT_SECRET="${arg#*=}" ;;
    --api=*) API_ENDPOINT="${arg#*=}" ;;
  esac
done

if [ -z "$SERVER_ID" ] || [ -z "$AGENT_SECRET" ]; then
  echo "Error: --server-id and --secret are required."
  exit 1
fi

echo ">> Configured for Server ID: $SERVER_ID"
echo ">> API Endpoint: $API_ENDPOINT"

sudo mkdir -p /etc/caelus /var/log/caelus /opt/caelus

sudo tee /etc/caelus/agent.env > /dev/null <<EOF
SERVER_ID=$SERVER_ID
AGENT_SECRET=$AGENT_SECRET
API_ENDPOINT=$API_ENDPOINT
COLLECTION_INTERVAL_SEC=5
LOG_LEVEL=info
EOF

sudo tee /etc/systemd/system/caelus-agent.service > /dev/null <<EOF
[Unit]
Description=Caelus Cloud Telemetry & Management Agent
After=network.target

[Service]
Type=simple
EnvironmentFile=/etc/caelus/agent.env
ExecStart=/opt/caelus/caelus-agent
Restart=always
RestartSec=5s
StandardOutput=append:/var/log/caelus/agent.log
StandardError=append:/var/log/caelus/agent.log

[Install]
WantedBy=multi-user.target
EOF

echo ">> Caelus Agent service installed at /etc/systemd/system/caelus-agent.service"
echo ">> You can now start the agent daemon."
`
		_, _ = w.Write([]byte(script))
	})
}

// registerAPIRoutes mendaftarkan seluruh rute API v1 publik dan terproteksi ke router Chi.
func registerAPIRoutes(r *chi.Mux, rc RouterConfig) {
	r.Route("/api/v1", func(apiRouter chi.Router) {
		if rc.Handlers.AuthHandler != nil {
			apiRouter.Route("/auth", func(authRouter chi.Router) {
				authRouter.Post("/register", rc.Handlers.AuthHandler.Register)
				authRouter.Post("/login", rc.Handlers.AuthHandler.Login)
				authRouter.Post("/refresh", rc.Handlers.AuthHandler.RefreshToken)
			})
		}

		if rc.Handlers.ProviderHandler != nil {
			apiRouter.Get("/providers", rc.Handlers.ProviderHandler.ListProviders)
		}

		if rc.Handlers.TelemetryHandler != nil {
			apiRouter.Post("/telemetry/report", rc.Handlers.TelemetryHandler.IngestReport)
			apiRouter.Get("/telemetry/stream/{server_id}", rc.Handlers.WSHandler.HandleSSE)
		}

		if rc.Handlers.WSHandler != nil {
			apiRouter.Get("/ws", rc.Handlers.WSHandler.HandleWebSocket)
		}

		if rc.JWTManager != nil {
			apiRouter.Group(func(protectedRouter chi.Router) {
				protectedRouter.Use(customMiddleware.Authenticate(rc.JWTManager))
				if rc.AuditRepo != nil {
					protectedRouter.Use(customMiddleware.AuditLogInterceptor(rc.AuditRepo, rc.Logger))
				}

				if rc.Handlers.ServerHandler != nil || rc.Handlers.TelemetryHandler != nil {
					registerServerRoutes(protectedRouter, rc.Handlers.ServerHandler, rc.Handlers.TelemetryHandler)
				}

				if rc.Handlers.AlertHandler != nil {
					registerAlertRoutes(protectedRouter, rc.Handlers.AlertHandler)
				}

				if rc.Handlers.StorageHandler != nil {
					registerStorageRoutes(protectedRouter, rc.Handlers.StorageHandler)
				}

				if rc.Handlers.BackupHandler != nil {
					registerBackupRoutes(protectedRouter, rc.Handlers.BackupHandler)
				}

				if rc.Handlers.AutomationHandler != nil {
					registerAutomationRoutes(protectedRouter, rc.Handlers.AutomationHandler)
				}

				if rc.Handlers.SecurityHandler != nil {
					registerSecurityRoutes(protectedRouter, rc.Handlers.SecurityHandler)
				}
			})
		}
	})
}

// registerSecurityRoutes mendaftarkan rute endpoint manajemen keamanan Sentinel.
func registerSecurityRoutes(r chi.Router, secH *v1.SecurityHandler) {
	r.Route("/security", func(secRouter chi.Router) {
		secRouter.Get("/overview", secH.GetPostureOverview)
		secRouter.Post("/scans", secH.TriggerScan)
		secRouter.Get("/scans", secH.ListScans)
		secRouter.Get("/scans/{id}", secH.GetScan)

		secRouter.Get("/findings", secH.ListFindings)
		secRouter.Get("/findings/{id}", secH.GetFinding)
		secRouter.Patch("/findings/{id}/status", secH.UpdateFindingStatus)

		secRouter.Get("/incidents", secH.ListIncidents)
		secRouter.Post("/incidents", secH.CreateIncident)
		secRouter.Patch("/incidents/{id}/status", secH.UpdateIncidentStatus)
	})
}

// registerServerRoutes mendaftarkan seluruh rute endpoint manajemen server VPS dan telemetri metrik.
func registerServerRoutes(r chi.Router, h *v1.ServerHandler, th *v1.TelemetryHandler) {
	r.Route("/servers", func(serverRouter chi.Router) {
		if h != nil {
			serverRouter.Get("/", h.ListServers)
			serverRouter.Post("/", h.CreateServer)
			serverRouter.Get("/{id}", h.GetServer)
			serverRouter.Patch("/{id}/resize", h.ResizeServer)
			serverRouter.Delete("/{id}", h.DeleteServer)
			serverRouter.Post("/{id}/reboot", h.RebootServer)
			serverRouter.Post("/{id}/shutdown", h.ShutdownServer)
			serverRouter.Post("/{id}/start", h.StartServer)
		}

		if th != nil {
			serverRouter.Get("/{id}/metrics/live", th.GetLiveMetrics)
			serverRouter.Get("/{id}/metrics/history", th.GetMetricHistory)
		}
	})
}

// registerAlertRoutes mendaftarkan rute endpoint manajemen insiden alert dan aturan threshold.
func registerAlertRoutes(r chi.Router, ah *v1.AlertHandler) {
	r.Route("/alerts", func(alertRouter chi.Router) {
		alertRouter.Get("/", ah.ListAlerts)
		alertRouter.Post("/{id}/acknowledge", ah.AcknowledgeAlert)
		alertRouter.Post("/{id}/resolve", ah.ResolveAlert)
		alertRouter.Get("/rules", ah.ListRules)
		alertRouter.Post("/rules", ah.CreateRule)
		alertRouter.Delete("/rules/{id}", ah.DeleteRule)
	})
}

// registerStorageRoutes mendaftarkan rute endpoint pengelolaan bucket dan file object storage.
func registerStorageRoutes(r chi.Router, sh *v1.StorageHandler) {
	r.Route("/storage", func(storageRouter chi.Router) {
		storageRouter.Get("/buckets", sh.ListBuckets)
		storageRouter.Post("/buckets", sh.CreateBucket)
		storageRouter.Get("/buckets/{name}", sh.GetBucket)
		storageRouter.Delete("/buckets/{name}", sh.DeleteBucket)

		storageRouter.Get("/buckets/{name}/objects", sh.ListObjects)
		storageRouter.Post("/buckets/{name}/objects", sh.UploadObject)
		storageRouter.Get("/buckets/{name}/objects/download", sh.DownloadObject)
		storageRouter.Delete("/buckets/{name}/objects", sh.DeleteObject)
		storageRouter.Post("/buckets/{name}/objects/signed-url", sh.GenerateSignedURL)
	})
}

// registerBackupRoutes mendaftarkan rute endpoint kebijakan dan riwayat backup server.
func registerBackupRoutes(r chi.Router, bh *v1.BackupHandler) {
	r.Route("/backups", func(backupRouter chi.Router) {
		backupRouter.Get("/policies", bh.ListPolicies)
		backupRouter.Post("/policies", bh.CreatePolicy)
		backupRouter.Delete("/policies/{id}", bh.DeletePolicy)

		backupRouter.Post("/trigger/{server_id}", bh.TriggerBackup)
		backupRouter.Get("/records", bh.ListRecords)
		backupRouter.Delete("/records/{id}", bh.DeleteRecord)
	})
}

// registerAutomationRoutes mendaftarkan rute endpoint aturan otomasi dan riwayat log eksekusi.
func registerAutomationRoutes(r chi.Router, autoH *v1.AutomationHandler) {
	r.Route("/automation", func(autoRouter chi.Router) {
		autoRouter.Get("/rules", autoH.ListRules)
		autoRouter.Post("/rules", autoH.CreateRule)
		autoRouter.Get("/rules/{id}", autoH.GetRule)
		autoRouter.Put("/rules/{id}", autoH.UpdateRule)
		autoRouter.Delete("/rules/{id}", autoH.DeleteRule)
		autoRouter.Post("/rules/{id}/test", autoH.TestRule)

		autoRouter.Get("/logs", autoH.ListLogs)
	})
}
