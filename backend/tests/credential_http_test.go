package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	deliveryHttp "github.com/havilz/caelus-cloud/backend/internal/delivery/http"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	v1 "github.com/havilz/caelus-cloud/backend/internal/delivery/http/v1"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	provFactory "github.com/havilz/caelus-cloud/backend/internal/provider"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/provider"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

func TestCredentialHTTPEndpoints(t *testing.T) {
	encKey := []byte("12345678901234567890123456789012")
	credRepo := newMockCredRepo()
	provRepo := newMockProviderRepo()
	factory := provFactory.NewDriverFactoryWithKey(encKey)

	provID := uuid.New()
	provRepo.providers[provID] = &domain.Provider{
		ID:       provID,
		Name:     "Amazon Web Services",
		Slug:     "aws",
		IsActive: true,
	}

	credUc := provider.NewCredentialUsecase(credRepo, provRepo, encKey)
	credHandler := v1.NewCredentialHandler(credUc, factory)

	jwtCfg := &config.JWTConfig{
		Secret:            "test_secret_key_at_least_32_characters_long_12345",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	jwtManager := jwt.NewJWTManager(jwtCfg, "caelus-test")

	router := deliveryHttp.NewRouter(deliveryHttp.RouterConfig{
		Config:     &config.Config{App: config.AppConfig{Name: "test-api", Env: "test"}},
		JWTManager: jwtManager,
		Handlers: deliveryHttp.Handlers{
			CredentialHandler: credHandler,
		},
	})

	orgID := uuid.New()
	user := &domain.User{ID: uuid.New(), Email: "admin@caelus.cloud", FullName: "Admin User", IsActive: true}
	tokens, _ := jwtManager.GenerateTokenPair(user, &orgID)
	authHeader := "Bearer " + tokens.AccessToken

	var createdCredID string

	t.Run("POST /api/v1/credentials - Create", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{
			"provider_id": provID.String(),
			"name":        "Production AWS Key",
			"api_key":     "AKIAIOSFODNN7EXAMPLE",
			"api_secret":  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"metadata": map[string]any{
				"region": "us-east-1",
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials", bytes.NewReader(body))
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var res response.APIResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		dataMap, ok := res.Data.(map[string]any)
		if !ok || dataMap["id"] == nil {
			t.Fatalf("expected credential id in response, got %v", res.Data)
		}
		createdCredID = dataMap["id"].(string)
	})

	t.Run("GET /api/v1/credentials - List", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials", nil)
		req.Header.Set("Authorization", authHeader)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GET /api/v1/credentials/{id} - Get", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials/"+createdCredID, nil)
		req.Header.Set("Authorization", authHeader)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /api/v1/credentials/{id}/test - Test Connection", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials/"+createdCredID+"/test", nil)
		req.Header.Set("Authorization", authHeader)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PUT /api/v1/credentials/{id} - Update", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{
			"name": "Updated AWS Credentials",
		})

		req := httptest.NewRequest(http.MethodPut, "/api/v1/credentials/"+createdCredID, bytes.NewReader(body))
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("DELETE /api/v1/credentials/{id} - Delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/credentials/"+createdCredID, nil)
		req.Header.Set("Authorization", authHeader)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})
}
