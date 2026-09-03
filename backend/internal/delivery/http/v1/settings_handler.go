package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/settings"
)

type SettingsHandler struct {
	settingsUsecase settings.Usecase
}

func NewSettingsHandler(uc settings.Usecase) *SettingsHandler {
	return &SettingsHandler{settingsUsecase: uc}
}

func (h *SettingsHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autentikasi gagal", "User ID tidak ditemukan")
		return
	}

	user, err := h.settingsUsecase.GetProfile(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil profil", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Profil pengguna berhasil diambil", user)
}

func (h *SettingsHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autentikasi gagal", "User ID tidak ditemukan")
		return
	}

	var input settings.UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", err.Error())
		return
	}

	user, err := h.settingsUsecase.UpdateProfile(r.Context(), userID, input)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal memperbarui profil", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Profil berhasil diperbarui", user)
}

func (h *SettingsHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autentikasi gagal", "User ID tidak ditemukan")
		return
	}

	var input settings.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", err.Error())
		return
	}

	if err := h.settingsUsecase.ChangePassword(r.Context(), userID, input); err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal mengganti kata sandi", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Kata sandi berhasil diperbarui", nil)
}

func (h *SettingsHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan pada sesi")
		return
	}

	org, err := h.settingsUsecase.GetOrganization(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data organisasi", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Data organisasi berhasil diambil", org)
}

func (h *SettingsHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan pada sesi")
		return
	}

	var input settings.UpdateOrganizationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", err.Error())
		return
	}

	org, err := h.settingsUsecase.UpdateOrganization(r.Context(), orgID, input)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal memperbarui data organisasi", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Organisasi berhasil diperbarui", org)
}

func (h *SettingsHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	members, err := h.settingsUsecase.ListMembers(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar anggota tim", err.Error())
		return
	}

	invitations, _ := h.settingsUsecase.ListInvitations(r.Context(), orgID)

	response.Success(w, http.StatusOK, "Daftar anggota organisasi berhasil diambil", map[string]any{
		"members":     members,
		"invitations": invitations,
	})
}

func (h *SettingsHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	userID, _ := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	var input settings.InviteMemberInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", err.Error())
		return
	}

	inv, err := h.settingsUsecase.InviteMember(r.Context(), orgID, userID, input)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal mengirim undangan", err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "Undangan anggota berhasil dibuat", inv)
}

func (h *SettingsHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	targetUserIDStr := chi.URLParam(r, "user_id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID pengguna tidak valid", err.Error())
		return
	}

	var body struct {
		Role domain.OrganizationRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", err.Error())
		return
	}

	if err := h.settingsUsecase.UpdateMemberRole(r.Context(), orgID, targetUserID, body.Role); err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal memperbarui peran anggota", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Peran anggota berhasil diperbarui", nil)
}

func (h *SettingsHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	targetUserIDStr := chi.URLParam(r, "user_id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID pengguna tidak valid", err.Error())
		return
	}

	if err := h.settingsUsecase.RemoveMember(r.Context(), orgID, targetUserID); err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal mengeluarkan anggota tim", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Anggota tim berhasil dikeluarkan", nil)
}

func (h *SettingsHandler) DeleteInvitation(w http.ResponseWriter, r *http.Request) {
	invIDStr := chi.URLParam(r, "invitation_id")
	invID, err := uuid.Parse(invIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID undangan tidak valid", err.Error())
		return
	}

	if err := h.settingsUsecase.DeleteInvitation(r.Context(), invID); err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal membatalkan undangan", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Undangan berhasil dibatalkan", nil)
}

func (h *SettingsHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	keys, err := h.settingsUsecase.ListAPIKeys(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar API keys", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Daftar API keys berhasil diambil", keys)
}

func (h *SettingsHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	userID, _ := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	var req domain.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", err.Error())
		return
	}

	apiKey, err := h.settingsUsecase.CreateAPIKey(r.Context(), orgID, userID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal membuat API key", err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "API key berhasil dibuat. Harap salin kunci rahasia ini sekarang karena tidak akan ditampilkan lagi.", apiKey)
}

func (h *SettingsHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	keyIDStr := chi.URLParam(r, "key_id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID API key tidak valid", err.Error())
		return
	}

	if err := h.settingsUsecase.DeleteAPIKey(r.Context(), orgID, keyID); err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal menghapus API key", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "API key berhasil dihapus / direvoke", nil)
}

func (h *SettingsHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	webhooks, err := h.settingsUsecase.ListWebhooks(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar webhook", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Daftar webhook berhasil diambil", webhooks)
}

func (h *SettingsHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	var req domain.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", err.Error())
		return
	}

	webhook, err := h.settingsUsecase.CreateWebhook(r.Context(), orgID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal mendaftarkan webhook", err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "Webhook berhasil didaftarkan", webhook)
}

func (h *SettingsHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	webhookIDStr := chi.URLParam(r, "webhook_id")
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID webhook tidak valid", err.Error())
		return
	}

	var req domain.UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", err.Error())
		return
	}

	webhook, err := h.settingsUsecase.UpdateWebhook(r.Context(), orgID, webhookID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal memperbarui webhook", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Webhook berhasil diperbarui", webhook)
}

func (h *SettingsHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	webhookIDStr := chi.URLParam(r, "webhook_id")
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID webhook tidak valid", err.Error())
		return
	}

	statusCode, err := h.settingsUsecase.TestWebhook(r.Context(), orgID, webhookID)
	if err != nil {
		response.Error(w, http.StatusBadGateway, "Pengiriman test webhook gagal", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Test ping webhook berhasil dikirim", map[string]any{
		"http_status": statusCode,
	})
}

func (h *SettingsHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	webhookIDStr := chi.URLParam(r, "webhook_id")
	webhookID, err := uuid.Parse(webhookIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID webhook tidak valid", err.Error())
		return
	}

	if err := h.settingsUsecase.DeleteWebhook(r.Context(), orgID, webhookID); err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal menghapus webhook", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Webhook berhasil dihapus", nil)
}

func (h *SettingsHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logs, total, err := h.settingsUsecase.ListAuditLogs(r.Context(), orgID, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil audit logs", err.Error())
		return
	}

	response.Paginated(w, http.StatusOK, "Audit logs berhasil diambil", logs, page, limit, total)
}
