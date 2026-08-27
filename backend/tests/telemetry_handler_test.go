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
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

// setupTelemetryHTTPTest menginisialisasi router dan test server untuk modul telemetri dan alert.
func setupTelemetryHTTPTest() (http.Handler, *mockMetricRepo, *mockAlertRepo, *mockServerRepo, jwt.Manager, uuid.UUID, uuid.UUID) {
	jwtCfg := &config.JWTConfig{
		Secret:            "test_secret_key_at_least_32_characters_long_12345",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	jwtManager := jwt.NewJWTManager(jwtCfg, "caelus-test")
	orgID := uuid.New()
	serverID := uuid.New()

	mockServerRepo := newMockServerRepo()
	_ = mockServerRepo.Create(context.Background(), &domain.Server{
		ID:             serverID,
		OrganizationID: orgID,
		Name:           "api-gateway",
		Status:         domain.ServerStatusRunning,
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
		Handlers: deliveryHttp.Handlers{
			TelemetryHandler: v1.NewTelemetryHandler(monitoringUc),
			AlertHandler:     v1.NewAlertHandler(monitoringUc),
			WSHandler:        ws.NewHandler(wsHub, jwtManager),
		},
	}

	router := deliveryHttp.NewRouter(routerConfig)
	return router, mockMetricRepo, mockAlertRepo, mockServerRepo, jwtManager, orgID, serverID
}

// TestTelemetryHTTP_IngestAndQuery memverifikasi endpoint ingestion telemetri dan pembacaan metrik live & history.
func TestTelemetryHTTP_IngestAndQuery(t *testing.T) {
	router, _, _, _, jwtManager, orgID, serverID := setupTelemetryHTTPTest()
	user := &domain.User{ID: uuid.New(), Email: "admin@caelus.cloud"}
	tokens, err := jwtManager.GenerateTokenPair(user, &orgID)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	// 1. Ingest report via POST /api/v1/telemetry/report
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
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Query Live Metrics via GET /api/v1/servers/{id}/metrics/live
	liveReq := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+serverID.String()+"/metrics/live", nil)
	liveReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	liveW := httptest.NewRecorder()
	router.ServeHTTP(liveW, liveReq)

	if liveW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for live metrics, got %d: %s", liveW.Code, liveW.Body.String())
	}

	// 3. Query Metric History via GET /api/v1/servers/{id}/metrics/history
	histReq := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+serverID.String()+"/metrics/history?duration=1h", nil)
	histReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	histW := httptest.NewRecorder()
	router.ServeHTTP(histW, histReq)

	if histW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for metric history, got %d: %s", histW.Code, histW.Body.String())
	}
}

// TestAlertHTTP_ListAndAcknowledge memverifikasi endpoint daftar alert dan aksi konfirmasi.
func TestAlertHTTP_ListAndAcknowledge(t *testing.T) {
	router, _, mockAlerts, _, jwtManager, orgID, _ := setupTelemetryHTTPTest()
	user := &domain.User{ID: uuid.New(), Email: "admin@caelus.cloud"}
	tokens, err := jwtManager.GenerateTokenPair(user, &orgID)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	// 1. List alerts
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/alerts?status=active", nil)
	listReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for list alerts, got %d: %s", listW.Code, listW.Body.String())
	}

	// 2. Acknowledge alert
	alertID := mockAlerts.alerts[0].ID
	ackReq := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/"+alertID.String()+"/acknowledge", nil)
	ackReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	ackW := httptest.NewRecorder()
	router.ServeHTTP(ackW, ackReq)

	if ackW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for acknowledge alert, got %d: %s", ackW.Code, ackW.Body.String())
	}
}
