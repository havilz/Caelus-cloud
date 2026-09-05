package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

func setupMiddlewareTest() (jwt.Manager, *mockOrgRepo, *domain.User, uuid.UUID) {
	jwtCfg := &config.JWTConfig{
		Secret:            "test_secret_key_at_least_32_characters_long_12345",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	jwtManager := jwt.NewJWTManager(jwtCfg, "caelus-test")
	orgRepo := newMockOrgRepo()

	orgID := uuid.New()
	user := &domain.User{
		ID:        uuid.New(),
		Email:     "admin@example.com",
		FullName:  "Admin User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return jwtManager, orgRepo, user, orgID
}

func TestAuthenticateMiddleware_MissingHeader(t *testing.T) {
	jwtManager, _, _, _ := setupMiddlewareTest()
	authMiddleware := middleware.Authenticate(jwtManager)

	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status harus %d, didapat %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthenticateMiddleware_InvalidFormat(t *testing.T) {
	jwtManager, _, _, _ := setupMiddlewareTest()
	authMiddleware := middleware.Authenticate(jwtManager)

	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic invalid_token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status harus %d, didapat %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthenticateMiddleware_InvalidToken(t *testing.T) {
	jwtManager, _, _, _ := setupMiddlewareTest()
	authMiddleware := middleware.Authenticate(jwtManager)

	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status harus %d, didapat %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthenticateMiddleware_Success(t *testing.T) {
	jwtManager, _, user, orgID := setupMiddlewareTest()
	authMiddleware := middleware.Authenticate(jwtManager)

	tokens, err := jwtManager.GenerateTokenPair(user, &orgID)
	if err != nil {
		t.Fatalf("gagal menghasilkan token: %v", err)
	}

	var capturedUserID uuid.UUID
	var capturedOrgID uuid.UUID

	handler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, _ := middleware.GetUserIDFromContext(r.Context())
		oid, _ := middleware.GetOrganizationIDFromContext(r.Context())
		capturedUserID = uid
		capturedOrgID = oid
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status harus %d, didapat %d", http.StatusOK, rec.Code)
	}

	if capturedUserID != user.ID {
		t.Errorf("UserID pada konteks tidak sesuai: didapat %v, diharapkan %v", capturedUserID, user.ID)
	}
	if capturedOrgID != orgID {
		t.Errorf("OrganizationID pada konteks tidak sesuai: didapat %v, diharapkan %v", capturedOrgID, orgID)
	}
}

func TestRequireOrganizationRole_NotMember(t *testing.T) {
	jwtManager, orgRepo, user, orgID := setupMiddlewareTest()
	authMiddleware := middleware.Authenticate(jwtManager)
	rbacMiddleware := middleware.RequireOrganizationRole(orgRepo, domain.RoleAdmin)

	tokens, _ := jwtManager.GenerateTokenPair(user, &orgID)

	handler := authMiddleware(rbacMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/org-admin", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("X-Organization-ID", orgID.String())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status harus %d, didapat %d", http.StatusForbidden, rec.Code)
	}
}

func TestRequireOrganizationRole_InsufficientRole(t *testing.T) {
	jwtManager, orgRepo, user, orgID := setupMiddlewareTest()
	authMiddleware := middleware.Authenticate(jwtManager)
	rbacMiddleware := middleware.RequireOrganizationRole(orgRepo, domain.RoleAdmin)

	_ = orgRepo.AddMember(context.Background(), &domain.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         user.ID,
		Role:           domain.RoleViewer,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	tokens, _ := jwtManager.GenerateTokenPair(user, &orgID)

	handler := authMiddleware(rbacMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/org-admin", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("X-Organization-ID", orgID.String())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status harus %d, didapat %d", http.StatusForbidden, rec.Code)
	}
}

func TestRequireOrganizationRole_AuthorizedRole(t *testing.T) {
	jwtManager, orgRepo, user, orgID := setupMiddlewareTest()
	authMiddleware := middleware.Authenticate(jwtManager)
	rbacMiddleware := middleware.RequireOrganizationRole(orgRepo, domain.RoleAdmin)

	_ = orgRepo.AddMember(context.Background(), &domain.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         user.ID,
		Role:           domain.RoleOwner,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	tokens, _ := jwtManager.GenerateTokenPair(user, &orgID)

	handler := authMiddleware(rbacMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/org-admin", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("X-Organization-ID", orgID.String())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status harus %d, didapat %d", http.StatusOK, rec.Code)
	}
}

func TestRequireAdmin_RejectsMember(t *testing.T) {
	jwtManager, orgRepo, user, orgID := setupMiddlewareTest()
	authMiddleware := middleware.Authenticate(jwtManager)
	adminMiddleware := middleware.RequireAdmin(orgRepo)

	_ = orgRepo.AddMember(context.Background(), &domain.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         user.ID,
		Role:           domain.RoleMember,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	tokens, _ := jwtManager.GenerateTokenPair(user, &orgID)

	handler := authMiddleware(adminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/credentials", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status harus %d untuk role member, didapat %d", http.StatusForbidden, rec.Code)
	}
}

func TestRequireOwner_RejectsAdmin(t *testing.T) {
	jwtManager, orgRepo, user, orgID := setupMiddlewareTest()
	authMiddleware := middleware.Authenticate(jwtManager)
	ownerMiddleware := middleware.RequireOwner(orgRepo)

	_ = orgRepo.AddMember(context.Background(), &domain.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         user.ID,
		Role:           domain.RoleAdmin,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	tokens, _ := jwtManager.GenerateTokenPair(user, &orgID)

	handler := authMiddleware(ownerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodPut, "/settings/members/123/role", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status harus %d untuk role admin pada endpoint khusus owner, didapat %d", http.StatusForbidden, rec.Code)
	}
}

func TestRequireOwner_AllowsOwner(t *testing.T) {
	jwtManager, orgRepo, user, orgID := setupMiddlewareTest()
	authMiddleware := middleware.Authenticate(jwtManager)
	ownerMiddleware := middleware.RequireOwner(orgRepo)

	_ = orgRepo.AddMember(context.Background(), &domain.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         user.ID,
		Role:           domain.RoleOwner,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	tokens, _ := jwtManager.GenerateTokenPair(user, &orgID)

	handler := authMiddleware(ownerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodPut, "/settings/members/123/role", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status harus %d untuk role owner, didapat %d", http.StatusOK, rec.Code)
	}
}

