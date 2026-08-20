package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	deliveryHttp "github.com/havilz/caelus-cloud/backend/internal/delivery/http"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
)

// TestHealthCheckEndpoints memverifikasi bahwa endpoint /health dan /api/v1/health mengembalikan status 200 OK dan struktur JSON yang valid.
func TestHealthCheckEndpoints(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "caelus-cloud-api-test",
			Env:         "test",
			CorsOrigins: []string{"*"},
		},
	}

	router := deliveryHttp.NewRouter(deliveryHttp.RouterConfig{Config: cfg})

	t.Run("GET /health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp response.APIResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success true, got false")
		}
	})

	t.Run("GET /api/v1/health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp response.APIResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected success true, got false")
		}
	})
}
