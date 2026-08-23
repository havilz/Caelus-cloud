package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/iac"
)

type IaCHandler struct {
	iacUsecase *iac.UseCase
}

func NewIaCHandler(uc *iac.UseCase) *IaCHandler {
	return &IaCHandler{iacUsecase: uc}
}

type validateYAMLRequest struct {
	RawYAML string `json:"raw_yaml"`
}

type createConfigRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RawYAML     string `json:"raw_yaml"`
}

type updateConfigRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RawYAML     string `json:"raw_yaml"`
}

type rollbackRequest struct {
	TargetVersion int `json:"target_version"`
}

// ValidateYAML memvalidasi sintaks dan skema YAML deklaratif tanpa menyimpannya ke database.
func (h *IaCHandler) ValidateYAML(w http.ResponseWriter, r *http.Request) {
	var req validateYAMLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	result := h.iacUsecase.ValidateYAML(req.RawYAML)
	response.Success(w, http.StatusOK, "YAML validation executed", result)
}

// ListConfigs mengambil seluruh daftar konfigurasi IaC milik organisasi.
func (h *IaCHandler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	configs, err := h.iacUsecase.ListConfigs(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list IaC configs", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "IaC configs retrieved", configs)
}

// CreateConfig membuat template konfigurasi deklaratif baru.
func (h *IaCHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	var req createConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.Name == "" || req.RawYAML == "" {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Name and raw_yaml are required")
		return
	}

	config, err := h.iacUsecase.CreateConfig(r.Context(), orgID, req.Name, req.Description, req.RawYAML)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to create IaC config", err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "IaC config created successfully", config)
}

// GetConfig mengambil detail konfigurasi spesifik berdasarkan ID.
func (h *IaCHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	config, err := h.iacUsecase.GetConfig(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "IaC config not found", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "IaC config retrieved", config)
}

// UpdateConfig memperbarui konfigurasi deklaratif.
func (h *IaCHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	var req updateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	config, err := h.iacUsecase.UpdateConfig(r.Context(), id, req.Name, req.Description, req.RawYAML)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to update IaC config", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "IaC config updated", config)
}

// DeleteConfig menghapus konfigurasi deklaratif.
func (h *IaCHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	if err := h.iacUsecase.DeleteConfig(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete IaC config", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "IaC config deleted", nil)
}

// GeneratePlan menghasilkan diff perbandingan Desired State vs Actual State.
func (h *IaCHandler) GeneratePlan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	plan, err := h.iacUsecase.GeneratePlan(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to generate plan", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Execution plan generated", plan)
}

// GetLatestPlan mengambil rencana eksekusi terakhir untuk konfigurasi tertentu.
func (h *IaCHandler) GetLatestPlan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	plan, err := h.iacUsecase.GetLatestPlan(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch plan", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Latest plan retrieved", plan)
}

// ApplyPlan mengeksekusi rencana IaC dan menyimpan snapshot state baru.
func (h *IaCHandler) ApplyPlan(w http.ResponseWriter, r *http.Request) {
	planIDStr := chi.URLParam(r, "id")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid plan ID format", err.Error())
		return
	}

	var userIDPtr *uuid.UUID
	if userID, ok := middleware.GetUserIDFromContext(r.Context()); ok {
		userIDPtr = &userID
	}

	state, err := h.iacUsecase.ApplyPlan(r.Context(), planID, userIDPtr)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to apply plan", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Plan applied successfully", state)
}

// RollbackState mengembalikan status konfigurasi ke snapshot versi sebelumnya.
func (h *IaCHandler) RollbackState(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	configID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	var req rollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.TargetVersion <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid request", "target_version must be greater than 0")
		return
	}

	var userIDPtr *uuid.UUID
	if userID, ok := middleware.GetUserIDFromContext(r.Context()); ok {
		userIDPtr = &userID
	}

	restoredState, err := h.iacUsecase.RollbackState(r.Context(), configID, req.TargetVersion, userIDPtr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to rollback state", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "State rolled back successfully", restoredState)
}

// ListStates mengambil riwayat snapshot state versi dari sebuah konfigurasi.
func (h *IaCHandler) ListStates(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	configID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	states, err := h.iacUsecase.ListStates(r.Context(), configID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list states", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "State history retrieved", states)
}
