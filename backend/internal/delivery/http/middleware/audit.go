package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type auditResourceKeyType string

const auditResourceKey auditResourceKeyType = "audit_resource"

type AuditResourceMetadata struct {
	ResourceType string
	ResourceID   string
	Payload      map[string]any
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// AuditLogInterceptor mengembalikan middleware HTTP yang secara otomatis mencatat riwayat aktivitas mutasi data oleh pengguna terautentikasi ke tabel audit_logs.
// Parameter auditRepo merupakan implementasi domain.AuditLogRepository untuk persistensi log.
// Parameter logger merupakan pointer *slog.Logger untuk pencatatan log kegagalan internal.
// Mengembalikan fungsi middleware func(http.Handler) http.Handler.
func AuditLogInterceptor(auditRepo domain.AuditLogRepository, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isMutatingMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			rec := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			meta := &AuditResourceMetadata{
				Payload: make(map[string]any),
			}
			ctx := context.WithValue(r.Context(), auditResourceKey, meta)

			next.ServeHTTP(rec, r.WithContext(ctx))

			recordAuditEntry(r.Context(), auditRepo, logger, r, rec.statusCode, meta)
		})
	}
}

// SetAuditMetadata menyematkan informasi resource yang sedang dimodifikasi ke dalam konteks request untuk dicatat oleh AuditLogInterceptor.
// Parameter ctx merupakan konteks request HTTP.
// Parameter resourceType merupakan tipe resource yang diubah (misal: "server", "user", "organization").
// Parameter resourceID merupakan identifier resource target.
func SetAuditMetadata(ctx context.Context, resourceType, resourceID string) {
	if meta, ok := ctx.Value(auditResourceKey).(*AuditResourceMetadata); ok {
		meta.ResourceType = resourceType
		meta.ResourceID = resourceID
	}
}

// recordAuditEntry menyusun entitas AuditLog dan menyimpannya ke database repository.
// Parameter ctx merupakan konteks request HTTP.
// Parameter auditRepo merupakan repository audit logs.
// Parameter logger merupakan structured logger.
// Parameter r merupakan pointer request HTTP.
// Parameter statusCode merupakan HTTP status code hasil respons.
// Parameter meta merupakan metadata resource yang terlampir pada konteks.
func recordAuditEntry(ctx context.Context, auditRepo domain.AuditLogRepository, logger *slog.Logger, r *http.Request, statusCode int, meta *AuditResourceMetadata) {
	userID, hasUser := GetUserIDFromContext(ctx)
	orgID, hasOrg := GetOrganizationIDFromContext(ctx)

	if !hasUser {
		return
	}

	clientIP := extractClientIP(r)
	userAgent := r.UserAgent()
	action := r.Method + " " + r.URL.Path

	var userUUID *uuid.UUID
	if hasUser {
		userUUID = &userID
	}

	var orgUUID *uuid.UUID
	if hasOrg {
		orgUUID = &orgID
	}

	resourceType := meta.ResourceType
	if resourceType == "" {
		resourceType = "http_request"
	}

	var resourceIDPtr *string
	if meta.ResourceID != "" {
		resourceIDPtr = &meta.ResourceID
	}

	payload := meta.Payload
	if payload == nil {
		payload = make(map[string]any)
	}
	payload["status_code"] = statusCode
	payload["query"] = r.URL.RawQuery

	auditEntry := &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgUUID,
		UserID:         userUUID,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceIDPtr,
		IPAddress:      &clientIP,
		UserAgent:      &userAgent,
		Payload:        payload,
		CreatedAt:      time.Now(),
	}

	if err := auditRepo.Create(context.Background(), auditEntry); err != nil && logger != nil {
		logger.Error("gagal menyimpan audit log", "error", err, "action", action)
	}
}

// isMutatingMethod memeriksa apakah metode HTTP melakukan perubahan state data (POST, PUT, PATCH, DELETE).
// Parameter method merupakan nama metode HTTP.
// Mengembalikan true jika metode termasuk mutasi data.
func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// extractClientIP mengambil alamat IP klien asli dari header proxy atau remote address.
// Parameter r merupakan pointer request HTTP.
// Mengembalikan string alamat IP bersih tanpa nomor port.
func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
