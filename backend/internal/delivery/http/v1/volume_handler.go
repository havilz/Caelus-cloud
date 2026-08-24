package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	volumeUsecase "github.com/havilz/caelus-cloud/backend/internal/usecase/volume"
)

type VolumeHandler struct {
	volUsecase *volumeUsecase.UseCase
}

func NewVolumeHandler(uc *volumeUsecase.UseCase) *VolumeHandler {
	return &VolumeHandler{volUsecase: uc}
}

// GetStoragePoolStats menangani GET /api/v1/volumes/stats untuk mengambil kapasitas disk fisik host
func (h *VolumeHandler) GetStoragePoolStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.volUsecase.GetStoragePoolStats(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil statistik storage pool", err.Error())
		return
	}
	response.Success(w, http.StatusOK, "Statistik storage pool berhasil diambil", stats)
}

// CreateVolume menangani POST /api/v1/volumes
func (h *VolumeHandler) CreateVolume(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	var req domain.CreateVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	vol, err := h.volUsecase.CreateVolume(r.Context(), orgID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal membuat volume", err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "Volume berhasil dibuat", vol)
}

// ListVolumes menangani GET /api/v1/volumes
func (h *VolumeHandler) ListVolumes(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	list, err := h.volUsecase.ListVolumes(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar volume", err.Error())
		return
	}

	if list == nil {
		list = []domain.Volume{}
	}

	response.Success(w, http.StatusOK, "Daftar volume berhasil diambil", list)
}

// DeleteVolume menangani DELETE /api/v1/volumes/{id}
func (h *VolumeHandler) DeleteVolume(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid volume ID", err.Error())
		return
	}

	if err := h.volUsecase.DeleteVolume(r.Context(), orgID, id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus volume", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Volume berhasil dihapus", nil)
}
