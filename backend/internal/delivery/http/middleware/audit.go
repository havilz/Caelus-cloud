package middleware

import (
	"context"
	"log/slog"
	"net/http"
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

func SetAuditMetadata(ctx context.Context, resourceType, resourceID string) {
	if meta, ok := ctx.Value(auditResourceKey).(*AuditResourceMetadata); ok {
		meta.ResourceType = resourceType
		meta.ResourceID = resourceID
	}
}

func recordAuditEntry(ctx context.Context, auditRepo domain.AuditLogRepository, logger *slog.Logger, r *http.Request, statusCode int, meta *AuditResourceMetadata) {
	userID, hasUser := GetUserIDFromContext(ctx)
	orgID, hasOrg := GetOrganizationIDFromContext(ctx)

	if !hasUser {
		return
	}

	clientIP := ExtractClientIP(r)
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
		logger.Error("failed to persist audit log", "error", err, "action", action)
	}
}

func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
