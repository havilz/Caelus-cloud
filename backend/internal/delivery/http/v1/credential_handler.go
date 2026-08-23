package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	provFactory "github.com/havilz/caelus-cloud/backend/internal/provider"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/provider"
)

type CredentialHandler struct {
	credUsecase   provider.CredentialUsecase
	driverFactory provFactory.Factory
}

// NewCredentialHandler menginisialisasi HTTP Handler untuk operasi manajemen kredensial cloud provider.
func NewCredentialHandler(uc provider.CredentialUsecase, factory provFactory.Factory) *CredentialHandler {
	return &CredentialHandler{
		credUsecase:   uc,
		driverFactory: factory,
	}
}

type createCredentialRequest struct {
	ProviderID uuid.UUID      `json:"provider_id"`
	Name       string         `json:"name"`
	APIKey     string         `json:"api_key,omitempty"`
	APISecret  string         `json:"api_secret,omitempty"`
	SSHKey     string         `json:"ssh_key,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type updateCredentialRequest struct {
	Name      string         `json:"name"`
	APIKey    string         `json:"api_key,omitempty"`
	APISecret string         `json:"api_secret,omitempty"`
	SSHKey    string         `json:"ssh_key,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// ListCredentials menangani HTTP GET /api/v1/credentials untuk mengambil seluruh kredensial milik organisasi.
func (h *CredentialHandler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan pada sesi")
		return
	}

	creds, err := h.credUsecase.ListCredentials(r.Context(), orgID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Daftar kredensial provider berhasil diambil", creds)
}

// CreateCredential menangani HTTP POST /api/v1/credentials untuk menambahkan kredensial provider terenkripsi baru.
func (h *CredentialHandler) CreateCredential(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan pada sesi")
		return
	}

	var req createCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format payload tidak valid", err.Error())
		return
	}

	cred, err := h.credUsecase.CreateCredential(r.Context(), provider.CreateCredentialInput{
		OrganizationID: orgID,
		ProviderID:     req.ProviderID,
		Name:           req.Name,
		APIKey:         req.APIKey,
		APISecret:      req.APISecret,
		SSHKey:         req.SSHKey,
		Metadata:       req.Metadata,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, "Kredensial provider berhasil disimpan dan dienkripsi", cred)
}

// GetCredential menangani HTTP GET /api/v1/credentials/{id} untuk mengambil metadata kredensial.
func (h *CredentialHandler) GetCredential(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan pada sesi")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID kredensial tidak valid", err.Error())
		return
	}

	cred, err := h.credUsecase.GetCredential(r.Context(), orgID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Detail kredensial provider berhasil diambil", cred)
}

// UpdateCredential menangani HTTP PUT /api/v1/credentials/{id} untuk memperbarui metadata atau token kredensial.
func (h *CredentialHandler) UpdateCredential(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan pada sesi")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID kredensial tidak valid", err.Error())
		return
	}

	var req updateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format payload tidak valid", err.Error())
		return
	}

	cred, err := h.credUsecase.UpdateCredential(r.Context(), provider.UpdateCredentialInput{
		ID:             id,
		OrganizationID: orgID,
		Name:           req.Name,
		APIKey:         req.APIKey,
		APISecret:      req.APISecret,
		SSHKey:         req.SSHKey,
		Metadata:       req.Metadata,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Kredensial provider berhasil diperbarui", cred)
}

// DeleteCredential menangani HTTP DELETE /api/v1/credentials/{id} untuk menghapus kredensial provider.
func (h *CredentialHandler) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan pada sesi")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID kredensial tidak valid", err.Error())
		return
	}

	if err := h.credUsecase.DeleteCredential(r.Context(), orgID, id); err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Kredensial provider berhasil dihapus", nil)
}

// TestCredential menangani HTTP POST /api/v1/credentials/{id}/test untuk memvalidasi koneksi kredensial ke provider cloud.
func (h *CredentialHandler) TestCredential(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Organization ID tidak ditemukan pada sesi")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID kredensial tidak valid", err.Error())
		return
	}

	cred, err := h.credUsecase.GetCredential(r.Context(), orgID, id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	if cred.Provider == nil {
		response.Error(w, http.StatusBadRequest, "Provider tidak ditemukan untuk kredensial ini", nil)
		return
	}

	driver, err := h.driverFactory.GetDriver(cred.Provider.Slug)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Driver provider tidak didukung", err.Error())
		return
	}

	// Uji koneksi dengan mengambil daftar server atau instance info dari driver
	servers, err := driver.ListServers(r.Context(), cred)
	if err != nil {
		response.Error(w, http.StatusBadGateway, "Gagal terhubung ke provider cloud", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Koneksi ke cloud provider berhasil diverifikasi", map[string]any{
		"provider":     cred.Provider.Slug,
		"status":       "connected",
		"server_count": len(servers),
	})
}

func (h *CredentialHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.Error(w, http.StatusNotFound, "Kredensial tidak ditemukan", err.Error())
	case errors.Is(err, domain.ErrBadRequest):
		response.Error(w, http.StatusBadRequest, "Parameter permintaan tidak valid", err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		response.Error(w, http.StatusUnauthorized, "Tidak memiliki izin akses", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "Terjadi kegagalan server", err.Error())
	}
}
