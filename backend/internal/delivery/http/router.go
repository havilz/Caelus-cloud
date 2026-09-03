package http

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	customMiddleware "github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	v1 "github.com/havilz/caelus-cloud/backend/internal/delivery/http/v1"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/ws"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	AuthHandler       *v1.AuthHandler
	ServerHandler     *v1.ServerHandler
	ProviderHandler   *v1.ProviderHandler
	CredentialHandler *v1.CredentialHandler
	TelemetryHandler  *v1.TelemetryHandler
	AlertHandler      *v1.AlertHandler
	StorageHandler    *v1.StorageHandler
	BackupHandler     *v1.BackupHandler
	AutomationHandler *v1.AutomationHandler
	SecurityHandler   *v1.SecurityHandler
	IaCHandler        *v1.IaCHandler
	DeploymentHandler *v1.DeploymentHandler
	NetworkHandler    *v1.NetworkHandler
	VolumeHandler     *v1.VolumeHandler
	DomainHandler     *v1.DomainHandler
	SettingsHandler   *v1.SettingsHandler
	WSHandler         *ws.Handler
}

type RouterConfig struct {
	Config     *config.Config
	JWTManager jwt.Manager
	AuditRepo  domain.AuditLogRepository
	ServerRepo domain.ServerRepository

	PgxPool  *pgxpool.Pool
	Logger   *slog.Logger
	Handlers Handlers
}

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

	r.Get("/agent-bin", func(w http.ResponseWriter, r *http.Request) {
		candidatePaths := []string{
			filepath.Join("agent", "bin", "caelus-agent"),
			filepath.Join("..", "agent", "bin", "caelus-agent"),
			filepath.Join("..", "..", "agent", "bin", "caelus-agent"),
			filepath.Join("/opt", "caelus", "caelus-agent"),
		}

		var targetBin string
		for _, path := range candidatePaths {
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				targetBin = path
				break
			}
		}

		if targetBin == "" {
			response.Error(w, http.StatusNotFound, "Binary caelus-agent belum dikompilasi pada server API (Jalankan 'make build-agent')", nil)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=caelus-agent")
		http.ServeFile(w, r, targetBin)
	})

	r.Get("/install.sh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		script := `#!/usr/bin/env bash
set -e

SERVER_ID=""
AGENT_SECRET=""
API_ENDPOINT="http://localhost:8080"

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --server-id=*) SERVER_ID="${1#*=}" ;;
        --server-id) SERVER_ID="$2"; shift ;;
        --secret=*) AGENT_SECRET="${1#*=}" ;;
        --secret) AGENT_SECRET="$2"; shift ;;
        --api=*|--endpoint=*) API_ENDPOINT="${1#*=}" ;;
        --api|--endpoint) API_ENDPOINT="$2"; shift ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

if [ -z "$SERVER_ID" ] || [ -z "$AGENT_SECRET" ]; then
    echo "Usage: curl -sSL https://caelus.cloud/install.sh | bash -s -- --server-id <ID> --secret <SECRET> [--api <URL>]"
    exit 1
fi

INSTALL_DIR="/opt/caelus"
mkdir -p "$INSTALL_DIR"

echo "-> Menghentikan service lama jika sedang berjalan..."
systemctl stop caelus-agent 2>/dev/null || true

echo "-> Mengunduh binary agent terbaru..."
curl -sSL "$API_ENDPOINT/agent-bin" -o "$INSTALL_DIR/caelus-agent.tmp"
mv -f "$INSTALL_DIR/caelus-agent.tmp" "$INSTALL_DIR/caelus-agent"
chmod +x "$INSTALL_DIR/caelus-agent"

echo "-> Membuat konfigurasi agent..."
cat <<EOF > "$INSTALL_DIR/agent.env"
SERVER_ID=$SERVER_ID
AGENT_SECRET=$AGENT_SECRET
API_ENDPOINT=$API_ENDPOINT
COLLECTION_INTERVAL_SEC=5
CAELUS_SERVER_ID=$SERVER_ID
CAELUS_AGENT_SECRET=$AGENT_SECRET
CAELUS_API_ENDPOINT=$API_ENDPOINT
CAELUS_INTERVAL=5s
EOF

if [ -n "$ALL_PROXY" ]; then
    echo "ALL_PROXY=$ALL_PROXY" >> "$INSTALL_DIR/agent.env"
    echo "HTTP_PROXY=$ALL_PROXY" >> "$INSTALL_DIR/agent.env"
    echo "HTTPS_PROXY=$ALL_PROXY" >> "$INSTALL_DIR/agent.env"
fi

echo "-> Mendaftarkan service systemd..."
cat <<EOF > /etc/systemd/system/caelus-agent.service
[Unit]
Description=Caelus Cloud Monitoring & Telemetry Agent
After=network.target

[Service]
Type=simple
EnvironmentFile=$INSTALL_DIR/agent.env
ExecStart=$INSTALL_DIR/caelus-agent
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload 2>/dev/null || true
systemctl enable caelus-agent 2>/dev/null || true
systemctl restart caelus-agent 2>/dev/null || true

echo "=== Caelus Cloud Agent Berhasil Diinstal dan Berjalan! ==="
`
		_, _ = w.Write([]byte(script))
	})
}

func registerAPIRoutes(r *chi.Mux, rc RouterConfig) {
	r.Route("/api/v1", func(apiRouter chi.Router) {
		if rc.Handlers.AuthHandler != nil {

			authLimiter := customMiddleware.NewAuthRateLimiter(5, 1*time.Minute)
			apiRouter.With(authLimiter.Limit()).Post("/auth/register", rc.Handlers.AuthHandler.Register)
			apiRouter.With(authLimiter.Limit()).Post("/auth/login", rc.Handlers.AuthHandler.Login)
			apiRouter.Post("/auth/refresh", rc.Handlers.AuthHandler.RefreshToken)
		}

		if rc.Handlers.ProviderHandler != nil {
			apiRouter.Get("/providers", rc.Handlers.ProviderHandler.ListProviders)
		}

		if rc.Handlers.TelemetryHandler != nil {

			if rc.ServerRepo != nil {
				apiRouter.With(customMiddleware.RequireAgentAuth(rc.ServerRepo)).Post("/telemetry/report", rc.Handlers.TelemetryHandler.IngestReport)
			}
		}

		if rc.Handlers.WSHandler != nil {
			apiRouter.Get("/ws", rc.Handlers.WSHandler.HandleWebSocket)
		}

		if rc.JWTManager != nil {
			apiRouter.Group(func(protectedRouter chi.Router) {
				protectedRouter.Use(customMiddleware.Authenticate(rc.JWTManager))

				protectedRouter.Use(customMiddleware.InjectOrgContext())
				if rc.AuditRepo != nil {
					protectedRouter.Use(customMiddleware.AuditLogInterceptor(rc.AuditRepo, rc.Logger))
				}

				if rc.Handlers.ServerHandler != nil || rc.Handlers.TelemetryHandler != nil {
					registerServerRoutes(protectedRouter, rc.Handlers.ServerHandler, rc.Handlers.TelemetryHandler)
				}

				if rc.Handlers.WSHandler != nil {
					protectedRouter.Get("/telemetry/stream/{server_id}", rc.Handlers.WSHandler.HandleSSE)
				}

				if rc.Handlers.CredentialHandler != nil {
					registerCredentialRoutes(protectedRouter, rc.Handlers.CredentialHandler)
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

				if rc.Handlers.IaCHandler != nil {
					registerIaCRoutes(protectedRouter, rc.Handlers.IaCHandler)
				}

				if rc.Handlers.DeploymentHandler != nil {
					registerDeploymentRoutes(protectedRouter, rc.Handlers.DeploymentHandler)
				}

				if rc.Handlers.NetworkHandler != nil {
					registerNetworkRoutes(protectedRouter, rc.Handlers.NetworkHandler)
				}

				if rc.Handlers.VolumeHandler != nil {
					registerVolumeRoutes(protectedRouter, rc.Handlers.VolumeHandler)
				}

				if rc.Handlers.DomainHandler != nil {
					registerDomainRoutes(protectedRouter, rc.Handlers.DomainHandler)
				}

				if rc.Handlers.SettingsHandler != nil {
					registerSettingsRoutes(protectedRouter, rc.Handlers.SettingsHandler)
				}
			})
		}
	})
}

func registerSettingsRoutes(r chi.Router, sh *v1.SettingsHandler) {
	r.Route("/settings", func(sr chi.Router) {
		sr.Get("/profile", sh.GetProfile)
		sr.Put("/profile", sh.UpdateProfile)
		sr.Post("/change-password", sh.ChangePassword)

		sr.Get("/organization", sh.GetOrganization)
		sr.Put("/organization", sh.UpdateOrganization)

		sr.Get("/members", sh.ListMembers)
		sr.Post("/members/invite", sh.InviteMember)
		sr.Put("/members/{user_id}/role", sh.UpdateMemberRole)
		sr.Delete("/members/{user_id}", sh.RemoveMember)
		sr.Delete("/invitations/{invitation_id}", sh.DeleteInvitation)

		sr.Get("/api-keys", sh.ListAPIKeys)
		sr.Post("/api-keys", sh.CreateAPIKey)
		sr.Delete("/api-keys/{key_id}", sh.DeleteAPIKey)

		sr.Get("/webhooks", sh.ListWebhooks)
		sr.Post("/webhooks", sh.CreateWebhook)
		sr.Put("/webhooks/{webhook_id}", sh.UpdateWebhook)
		sr.Post("/webhooks/{webhook_id}/test", sh.TestWebhook)
		sr.Delete("/webhooks/{webhook_id}", sh.DeleteWebhook)

		sr.Get("/audit-logs", sh.ListAuditLogs)
	})
}

func registerDomainRoutes(r chi.Router, dh *v1.DomainHandler) {
	r.Route("/domains", func(domainRouter chi.Router) {
		domainRouter.Get("/", dh.ListDomains)
		domainRouter.Post("/", dh.CreateDomain)
		domainRouter.Get("/{id}", dh.GetDomain)
		domainRouter.Delete("/{id}", dh.DeleteDomain)
		domainRouter.Post("/{id}/verify", dh.VerifyDomain)
	})
}

func registerVolumeRoutes(r chi.Router, volH *v1.VolumeHandler) {
	r.Route("/volumes", func(volRouter chi.Router) {
		volRouter.Get("/stats", volH.GetStoragePoolStats)
		volRouter.Post("/", volH.CreateVolume)
		volRouter.Get("/", volH.ListVolumes)
		volRouter.Delete("/{id}", volH.DeleteVolume)
	})
}

func registerIaCRoutes(r chi.Router, iacH *v1.IaCHandler) {
	r.Route("/iac", func(iacRouter chi.Router) {
		iacRouter.Post("/validate", iacH.ValidateYAML)
		iacRouter.Get("/configs", iacH.ListConfigs)
		iacRouter.Post("/configs", iacH.CreateConfig)
		iacRouter.Get("/configs/{id}", iacH.GetConfig)
		iacRouter.Put("/configs/{id}", iacH.UpdateConfig)
		iacRouter.Delete("/configs/{id}", iacH.DeleteConfig)
		iacRouter.Post("/configs/{id}/plan", iacH.GeneratePlan)
		iacRouter.Get("/configs/{id}/plan", iacH.GetLatestPlan)
		iacRouter.Post("/plans/{id}/apply", iacH.ApplyPlan)
		iacRouter.Post("/configs/{id}/rollback", iacH.RollbackState)
		iacRouter.Get("/configs/{id}/states", iacH.ListStates)
	})
}

func registerDeploymentRoutes(r chi.Router, depH *v1.DeploymentHandler) {
	r.Route("/deployments", func(depRouter chi.Router) {
		depRouter.Post("/", depH.CreateDeployment)
		depRouter.Get("/", depH.ListDeployments)
		depRouter.Get("/{id}", depH.GetDeployment)
		depRouter.Get("/{id}/logs", depH.GetLogs)
		depRouter.Get("/{id}/logs/stream", depH.StreamLogsSSE)
		depRouter.Post("/{id}/stop", depH.StopDeployment)
		depRouter.Post("/{id}/redeploy", depH.RedeployDeployment)
		depRouter.Post("/{id}/rollback", depH.RollbackDeployment)
		depRouter.Delete("/{id}", depH.DeleteDeployment)
	})
}

func registerNetworkRoutes(r chi.Router, netH *v1.NetworkHandler) {
	r.Route("/networks", func(netRouter chi.Router) {
		netRouter.Post("/", netH.CreateNetwork)
		netRouter.Get("/", netH.ListNetworks)
		netRouter.Delete("/{id}", netH.DeleteNetwork)
	})

	r.Route("/firewall-rules", func(fwRouter chi.Router) {
		fwRouter.Post("/", netH.CreateFirewallRule)
		fwRouter.Get("/", netH.ListFirewallRules)
		fwRouter.Delete("/{id}", netH.DeleteFirewallRule)
	})
}

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

func registerStorageRoutes(r chi.Router, sh *v1.StorageHandler) {
	r.Route("/storage", func(storageRouter chi.Router) {
		storageRouter.Get("/buckets", sh.ListBuckets)
		storageRouter.Post("/buckets", sh.CreateBucket)
		storageRouter.Post("/sync", sh.SyncBuckets)
		storageRouter.Get("/buckets/{name}", sh.GetBucket)
		storageRouter.Delete("/buckets/{name}", sh.DeleteBucket)

		storageRouter.Get("/buckets/{name}/objects", sh.ListObjects)
		storageRouter.Post("/buckets/{name}/objects", sh.UploadObject)
		storageRouter.Get("/buckets/{name}/objects/download", sh.DownloadObject)
		storageRouter.Delete("/buckets/{name}/objects", sh.DeleteObject)
		storageRouter.Post("/buckets/{name}/objects/signed-url", sh.GenerateSignedURL)
	})
}

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

func registerCredentialRoutes(r chi.Router, ch *v1.CredentialHandler) {
	r.Route("/credentials", func(credRouter chi.Router) {
		credRouter.Get("/", ch.ListCredentials)
		credRouter.Post("/", ch.CreateCredential)
		credRouter.Get("/{id}", ch.GetCredential)
		credRouter.Put("/{id}", ch.UpdateCredential)
		credRouter.Delete("/{id}", ch.DeleteCredential)
		credRouter.Post("/{id}/test", ch.TestCredential)
	})
}
