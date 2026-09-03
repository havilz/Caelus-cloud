package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	deliveryHttp "github.com/havilz/caelus-cloud/backend/internal/delivery/http"
	v1 "github.com/havilz/caelus-cloud/backend/internal/delivery/http/v1"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/ws"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/monitoring"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/hasher"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

func setupTelemetryHTTPTest() (http.Handler, *mockMetricRepo, *mockAlertRepo, *mockServerRepo, jwt.Manager, uuid.UUID, uuid.UUID) {
	jwtCfg := &config.JWTConfig{
		Secret:            "test_secret_key_at_least_32_characters_long_12345",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	jwtManager := jwt.NewJWTManager(jwtCfg, "caelus-test")
	orgID := uuid.New()
	serverID := uuid.New()

	secret := "test-agent-secret-12345"
	secretHash, _ := hasher.Hash(secret, nil)
	mockServerRepo := newMockServerRepo()
	_ = mockServerRepo.Create(context.Background(), &domain.Server{
		ID:              serverID,
		OrganizationID:  orgID,
		Name:            "api-gateway",
		Status:          domain.ServerStatusRunning,
		AgentSecretHash: &secretHash,
	})

	mockMetricRepo := &mockMetricRepo{
		metrics: []domain.ServerMetric{
			{
				ServerID:    serverID,
				CPUUsagePct: 35.0,
				RecordedAt:  time.Now().UTC(),
			},
		},
	}
	mockAlertRepo := &mockAlertRepo{
		alerts: []domain.Alert{
			{
				ID:             uuid.New(),
				OrganizationID: orgID,
				ServerID:       serverID,
				Title:          "High Load",
				Status:         domain.AlertStatusActive,
			},
		},
	}

	wsHub := ws.NewHub()
	evaluator := monitoring.NewAlertEvaluator(mockAlertRepo, wsHub)
	monitoringUc := monitoring.NewMonitoringUsecase(mockMetricRepo, mockAlertRepo, mockServerRepo, evaluator, wsHub, nil, nil, nil)

	routerConfig := deliveryHttp.RouterConfig{
		JWTManager: jwtManager,
		ServerRepo: mockServerRepo,
		Handlers: deliveryHttp.Handlers{
			TelemetryHandler: v1.NewTelemetryHandler(monitoringUc),
			AlertHandler:     v1.NewAlertHandler(monitoringUc),
			WSHandler:        ws.NewHandler(wsHub, jwtManager),
		},
	}

	router := deliveryHttp.NewRouter(routerConfig)
	return router, mockMetricRepo, mockAlertRepo, mockServerRepo, jwtManager, orgID, serverID
}

func TestTelemetryHTTP_IngestAndQuery(t *testing.T) {
	router, _, _, _, jwtManager, orgID, serverID := setupTelemetryHTTPTest()
	user := &domain.User{ID: uuid.New(), Email: "admin@caelus.cloud"}
	tokens, err := jwtManager.GenerateTokenPair(user, &orgID)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	payload := domain.TelemetryReportPayload{
		ServerID:  serverID,
		Timestamp: time.Now().UTC(),
		Host: domain.HostMetricsPayload{
			CPUUsagePct: 62.0,
			CPUCores:    4,
		},
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/report", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Server-ID", serverID.String())
	req.Header.Set("Authorization", "Bearer test-agent-secret-12345")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	liveReq := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+serverID.String()+"/metrics/live", nil)
	liveReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	liveW := httptest.NewRecorder()
	router.ServeHTTP(liveW, liveReq)

	if liveW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for live metrics, got %d: %s", liveW.Code, liveW.Body.String())
	}

	histReq := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+serverID.String()+"/metrics/history?duration=1h", nil)
	histReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	histW := httptest.NewRecorder()
	router.ServeHTTP(histW, histReq)

	if histW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for metric history, got %d: %s", histW.Code, histW.Body.String())
	}
}

func TestAlertHTTP_ListAndAcknowledge(t *testing.T) {
	router, _, mockAlerts, _, jwtManager, orgID, _ := setupTelemetryHTTPTest()
	user := &domain.User{ID: uuid.New(), Email: "admin@caelus.cloud"}
	tokens, err := jwtManager.GenerateTokenPair(user, &orgID)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/alerts?status=active", nil)
	listReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for list alerts, got %d: %s", listW.Code, listW.Body.String())
	}

	alertID := mockAlerts.alerts[0].ID
	ackReq := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/"+alertID.String()+"/acknowledge", nil)
	ackReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	ackW := httptest.NewRecorder()
	router.ServeHTTP(ackW, ackReq)

	if ackW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for acknowledge alert, got %d: %s", ackW.Code, ackW.Body.String())
	}
}
