package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/automation"
	deliveryHttp "github.com/havilz/caelus-cloud/backend/internal/delivery/http"
	v1 "github.com/havilz/caelus-cloud/backend/internal/delivery/http/v1"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/notification"
	"github.com/havilz/caelus-cloud/backend/internal/notification/email"
	"github.com/havilz/caelus-cloud/backend/internal/notification/webhook"
	"github.com/havilz/caelus-cloud/backend/internal/queue/mock"
	automationUc "github.com/havilz/caelus-cloud/backend/internal/usecase/automation"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

func TestAutomationHTTP_Endpoints(t *testing.T) {
	repo := &mockAutomationRepo{}
	q := mock.NewMockQueueEngine()
	notifier := notification.NewUnifiedDispatcher(webhook.NewClient(""), email.NewClient(email.Config{}))
	dispatcher := automation.NewCentralEventDispatcher()
	defer dispatcher.Stop()
	engine := automation.NewEngine(repo, q, notifier, nil, nil)

	uc := automationUc.NewAutomationUsecase(repo, engine, dispatcher)
	handler := v1.NewAutomationHandler(uc)

	jwtSecret := "test_secret_for_automation_http_testing_123"
	jwtMgr := jwt.NewJWTManager(&config.JWTConfig{
		Secret:            jwtSecret,
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 24 * time.Hour,
	}, "caelus-test")

	orgID := uuid.New()
	user := &domain.User{ID: uuid.New(), Email: "devops@caelus.cloud", FullName: "DevOps User", IsActive: true}
	tokenPair, err := jwtMgr.GenerateTokenPair(user, &orgID)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}
	token := tokenPair.AccessToken

	routerConfig := deliveryHttp.RouterConfig{
		JWTManager: jwtMgr,
		Handlers: deliveryHttp.Handlers{
			AutomationHandler: handler,
		},
	}
	router := deliveryHttp.NewRouter(routerConfig)

	createBody := map[string]any{
		"name":         "Auto-Scale on CPU",
		"trigger_type": "metric_threshold",
		"conditions": []map[string]any{
			{"field": "cpu_usage_percent", "operator": ">=", "value": 90},
		},
		"actions": []map[string]any{
			{"type": "send_email", "target": "ops@caelus.cloud"},
		},
		"cooldown_seconds": 300,
	}
	bodyBytes, _ := json.Marshal(createBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/automation/rules", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var createdResp struct {
		Data domain.AutomationRule `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createdResp)
	ruleID := createdResp.Data.ID

	req, _ = http.NewRequest(http.MethodGet, "/api/v1/automation/rules", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", w.Code)
	}

	testBody := map[string]any{
		"mock_data": map[string]any{
			"cpu_usage_percent": 95,
		},
	}
	testBytes, _ := json.Marshal(testBody)
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/automation/rules/"+ruleID.String()+"/test", bytes.NewBuffer(testBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", w.Code)
	}

	req, _ = http.NewRequest(http.MethodGet, "/api/v1/automation/logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", w.Code)
	}
}
