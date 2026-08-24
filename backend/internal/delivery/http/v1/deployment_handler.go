package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/orchestration"
)

type DeploymentHandler struct {
	depUsecase *orchestration.UseCase
}

func NewDeploymentHandler(uc *orchestration.UseCase) *DeploymentHandler {
	return &DeploymentHandler{depUsecase: uc}
}

// CreateDeployment menangani HTTP POST /api/v1/deployments untuk memicu deployment container baru.
func (h *DeploymentHandler) CreateDeployment(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	var req domain.DeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	dep, err := h.depUsecase.CreateDeployment(r.Context(), orgID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to create deployment", err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "Container deployment initiated", dep)
}

// ListDeployments menangani HTTP GET /api/v1/deployments untuk mengambil riwayat deployment container.
func (h *DeploymentHandler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	var serverIDPtr *uuid.UUID
	serverIDStr := r.URL.Query().Get("server_id")
	if serverIDStr != "" {
		if sid, err := uuid.Parse(serverIDStr); err == nil {
			serverIDPtr = &sid
		}
	}

	deployments, err := h.depUsecase.ListDeployments(r.Context(), orgID, serverIDPtr)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list deployments", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Deployments retrieved", deployments)
}

// GetDeployment menangani HTTP GET /api/v1/deployments/{id}.
func (h *DeploymentHandler) GetDeployment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	dep, err := h.depUsecase.GetDeployment(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Deployment not found", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Deployment details retrieved", dep)
}

// GetLogs menangani HTTP GET /api/v1/deployments/{id}/logs untuk mengambil snapshot baris log deployment.
func (h *DeploymentHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	limit := 200
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := h.depUsecase.GetLogs(r.Context(), id, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch deployment logs", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Deployment logs retrieved", logs)
}

// StreamLogsSSE menangani HTTP GET /api/v1/deployments/{id}/logs/stream via Server-Sent Events (SSE).
func (h *DeploymentHandler) StreamLogsSSE(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Poll new logs periodically for active SSE client
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastLogID int64 = 0

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			logs, err := h.depUsecase.GetLogs(r.Context(), id, 100)
			if err != nil {
				continue
			}

			for _, log := range logs {
				if log.ID > lastLogID {
					lastLogID = log.ID
					data, _ := json.Marshal(log)
					fmt.Fprintf(w, "data: %s\n\n", data)
					flusher.Flush()
				}
			}

			// Check if deployment has reached terminal status
			dep, err := h.depUsecase.GetDeployment(r.Context(), id)
			if err == nil && (dep.Status == domain.DeploymentStatusRunning || dep.Status == domain.DeploymentStatusFailed || dep.Status == domain.DeploymentStatusStopped) {
				// Send completed event after final flush
				fmt.Fprintf(w, "event: complete\ndata: {\"status\": \"%s\"}\n\n", dep.Status)
				flusher.Flush()
				return
			}
		}
	}
}

// StopDeployment menghentikan container deployment yang sedang berjalan.
func (h *DeploymentHandler) StopDeployment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	if err := h.depUsecase.StopDeployment(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to stop deployment", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Deployment stopped successfully", nil)
}

// RedeployDeployment memicu ulang eksekusi pipeline deployment yang sudah ada.
func (h *DeploymentHandler) RedeployDeployment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	dep, err := h.depUsecase.RedeployDeployment(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to redeploy container", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Container redeployed successfully", dep)
}

// RollbackDeployment memicu pembuatan instance deployment baru berdasarkan konfigurasi sebelumnya.
func (h *DeploymentHandler) RollbackDeployment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	newDep, err := h.depUsecase.RollbackDeployment(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to rollback deployment", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Rollback deployment triggered successfully", newDep)
}

// DeleteDeployment menghapus deployment dan kontainer fisik dari host.
func (h *DeploymentHandler) DeleteDeployment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	if err := h.depUsecase.DeleteDeployment(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete deployment", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Deployment and container deleted successfully", nil)
}
