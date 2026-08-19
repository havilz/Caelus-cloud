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

type mockAuditRepo struct {
	logs []domain.AuditLog
}

func newMockAuditRepo() *mockAuditRepo {
	return &mockAuditRepo{
		logs: make([]domain.AuditLog, 0),
	}
}

func (m *mockAuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	m.logs = append(m.logs, *log)
	return nil
}

func (m *mockAuditRepo) ListByOrg(ctx context.Context, orgID uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error) {
	var result []domain.AuditLog
	for _, l := range m.logs {
		if l.OrganizationID != nil && *l.OrganizationID == orgID {
			result = append(result, l)
		}
	}
	return result, int64(len(result)), nil
}

// TestAuditRepository_CreateAndList memverifikasi penyimpanan entitas audit log ke repository dan pengambilan data log berdasarkan organisasi.
func TestAuditRepository_CreateAndList(t *testing.T) {
	repo := newMockAuditRepo()
	ctx := context.Background()

	orgID := uuid.New()
	userID := uuid.New()
	resID := "srv-12345"

	entry := &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: &orgID,
		UserID:         &userID,
		Action:         "POST /api/v1/servers",
		ResourceType:   "server",
		ResourceID:     &resID,
		Payload: map[string]any{
			"server_name": "production-node-1",
		},
		CreatedAt: time.Now(),
	}

	if err := repo.Create(ctx, entry); err != nil {
		t.Fatalf("gagal membuat audit log: %v", err)
	}

	logs, total, err := repo.ListByOrg(ctx, orgID, 1, 10)
	if err != nil {
		t.Fatalf("gagal mengambil audit logs: %v", err)
	}

	if total != 1 {
		t.Errorf("total log harus 1, didapat: %d", total)
	}

	if len(logs) == 0 || logs[0].Action != "POST /api/v1/servers" {
		t.Errorf("data audit log tidak sesuai")
	}
}

// TestAuditInterceptor_MutatingMethod_Captured memverifikasi bahwa HTTP request mutasi data (POST) otomatis dicatat ke dalam audit log lengkap dengan identitas pengguna.
func TestAuditInterceptor_MutatingMethod_Captured(t *testing.T) {
	jwtCfg := &config.JWTConfig{
		Secret:            "test_secret_key_at_least_32_characters_long_12345",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	jwtManager := jwt.NewJWTManager(jwtCfg, "caelus-test")
	auditRepo := newMockAuditRepo()

	orgID := uuid.New()
	user := &domain.User{
		ID:        uuid.New(),
		Email:     "operator@example.com",
		FullName:  "Operator User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tokens, _ := jwtManager.GenerateTokenPair(user, &orgID)

	authMw := middleware.Authenticate(jwtManager)
	auditMw := middleware.AuditLogInterceptor(auditRepo, nil)

	handler := authMw(auditMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middleware.SetAuditMetadata(r.Context(), "server", "srv-test-999")
		w.WriteHeader(http.StatusCreated)
	})))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	req.Header.Set("User-Agent", "CaelusTestClient/1.0")
	req.Header.Set("X-Forwarded-For", "203.0.113.195")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status code harus %d, didapat %d", http.StatusCreated, rec.Code)
	}

	if len(auditRepo.logs) != 1 {
		t.Fatalf("harus ada 1 log audit yang tercatat, didapat: %d", len(auditRepo.logs))
	}

	logged := auditRepo.logs[0]
	if logged.UserID == nil || *logged.UserID != user.ID {
		t.Errorf("UserID tidak sesuai")
	}
	if logged.OrganizationID == nil || *logged.OrganizationID != orgID {
		t.Errorf("OrganizationID tidak sesuai")
	}
	if logged.ResourceType != "server" {
		t.Errorf("ResourceType tidak sesuai: %s", logged.ResourceType)
	}
	if logged.ResourceID == nil || *logged.ResourceID != "srv-test-999" {
		t.Errorf("ResourceID tidak sesuai")
	}
	if logged.IPAddress == nil || *logged.IPAddress != "203.0.113.195" {
		t.Errorf("IPAddress tidak sesuai: %v", logged.IPAddress)
	}
}

// TestAuditInterceptor_NonMutatingMethod_Skipped memverifikasi bahwa HTTP request pembacaan data (GET) diabaikan dan tidak dicatat ke tabel audit log.
func TestAuditInterceptor_NonMutatingMethod_Skipped(t *testing.T) {
	jwtCfg := &config.JWTConfig{
		Secret:            "test_secret_key_at_least_32_characters_long_12345",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	jwtManager := jwt.NewJWTManager(jwtCfg, "caelus-test")
	auditRepo := newMockAuditRepo()

	user := &domain.User{
		ID:        uuid.New(),
		Email:     "operator@example.com",
		FullName:  "Operator User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	tokens, _ := jwtManager.GenerateTokenPair(user, nil)

	authMw := middleware.Authenticate(jwtManager)
	auditMw := middleware.AuditLogInterceptor(auditRepo, nil)

	handler := authMw(auditMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if len(auditRepo.logs) != 0 {
		t.Errorf("request GET tidak boleh dicatat ke audit logs")
	}
}

// TestAuditInterceptor_Unauthenticated_Skipped memverifikasi bahwa request yang tidak memiliki konteks pengguna terautentikasi tidak dicatat ke audit logs.
func TestAuditInterceptor_Unauthenticated_Skipped(t *testing.T) {
	auditRepo := newMockAuditRepo()
	auditMw := middleware.AuditLogInterceptor(auditRepo, nil)

	handler := auditMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/public/endpoint", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if len(auditRepo.logs) != 0 {
		t.Errorf("request tanpa autentikasi tidak boleh dicatat ke audit logs")
	}
}
