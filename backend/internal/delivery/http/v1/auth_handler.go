package v1

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/auth"
)

const (
	errInvalidPayloadFormat = "Format payload tidak valid"
)

type AuthHandler struct {
	authUsecase auth.Usecase
	auditRepo   domain.AuditLogRepository
}

func NewAuthHandler(uc auth.Usecase, auditRepo domain.AuditLogRepository) *AuthHandler {
	return &AuthHandler{
		authUsecase: uc,
		auditRepo:   auditRepo,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input auth.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.recordAuthAudit(r, "auth.register_failed", input.Email, nil, nil, http.StatusBadRequest, err.Error())
		response.Error(w, http.StatusBadRequest, errInvalidPayloadFormat, err.Error())
		return
	}

	output, err := h.authUsecase.Register(r.Context(), input)
	if err != nil {
		statusCode := getAuthErrorStatusCode(err)
		h.recordAuthAudit(r, "auth.register_failed", input.Email, nil, nil, statusCode, err.Error())
		handleAuthError(w, err)
		return
	}

	var userIDPtr *uuid.UUID
	var orgIDPtr *uuid.UUID
	if output != nil {
		userIDPtr = &output.User.ID
		if output.Organization != nil {
			orgIDPtr = &output.Organization.ID
		}
	}

	h.recordAuthAudit(r, "auth.register_success", input.Email, userIDPtr, orgIDPtr, http.StatusCreated, "registered successfully")
	response.Success(w, http.StatusCreated, "Registrasi akun berhasil", output)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input auth.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.recordAuthAudit(r, "auth.login_failed", input.Email, nil, nil, http.StatusBadRequest, err.Error())
		response.Error(w, http.StatusBadRequest, errInvalidPayloadFormat, err.Error())
		return
	}

	output, err := h.authUsecase.Login(r.Context(), input)
	if err != nil {
		statusCode := getAuthErrorStatusCode(err)
		h.recordAuthAudit(r, "auth.login_failed", input.Email, nil, nil, statusCode, err.Error())
		handleAuthError(w, err)
		return
	}

	var userIDPtr *uuid.UUID
	if output != nil && output.User != nil {
		userIDPtr = &output.User.ID
	}

	h.recordAuthAudit(r, "auth.login_success", input.Email, userIDPtr, nil, http.StatusOK, "login successful")
	response.Success(w, http.StatusOK, "Login berhasil", output)
}

func (h *AuthHandler) recordAuthAudit(r *http.Request, action, email string, userID, orgID *uuid.UUID, statusCode int, message string) {
	if h.auditRepo == nil {
		return
	}

	clientIP := extractIP(r)
	userAgent := r.UserAgent()

	payload := map[string]any{
		"email":       email,
		"status_code": statusCode,
		"message":     message,
	}

	auditEntry := &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Action:         action,
		ResourceType:   "auth",
		IPAddress:      &clientIP,
		UserAgent:      &userAgent,
		Payload:        payload,
		CreatedAt:      time.Now(),
	}

	_ = h.auditRepo.Create(context.Background(), auditEntry)
}

func extractIP(r *http.Request) string {
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

func getAuthErrorStatusCode(err error) int {
	switch err {
	case domain.ErrEmailAlreadyInUse:
		return http.StatusConflict
	case domain.ErrEmailInvalid, domain.ErrPasswordTooShort, domain.ErrBadRequest:
		return http.StatusBadRequest
	case domain.ErrInvalidCredentials:
		return http.StatusUnauthorized
	case domain.ErrUserInactive:
		return http.StatusForbidden
	case domain.ErrInvalidToken, domain.ErrUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, errInvalidPayloadFormat, "Field refresh_token wajib diisi")
		return
	}

	tokenPair, err := h.authUsecase.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Token berhasil diperbarui", tokenPair)
}

func handleAuthError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrEmailAlreadyInUse:
		response.Error(w, http.StatusConflict, "Email sudah terdaftar", err.Error())
	case domain.ErrEmailInvalid, domain.ErrPasswordTooShort, domain.ErrBadRequest:
		response.Error(w, http.StatusBadRequest, "Data registrasi tidak valid", err.Error())
	case domain.ErrInvalidCredentials:
		response.Error(w, http.StatusUnauthorized, "Email atau kata sandi tidak sesuai", err.Error())
	case domain.ErrUserInactive:
		response.Error(w, http.StatusForbidden, "Akun pengguna dinonaktifkan", err.Error())
	case domain.ErrInvalidToken, domain.ErrUnauthorized:
		response.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau telah berakhir", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "Terjadi kesalahan internal server", err.Error())
	}
}
