package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/monitoring"
)

type TelemetryHandler struct {
	usecase monitoring.MonitoringUsecase
}

func NewTelemetryHandler(usecase monitoring.MonitoringUsecase) *TelemetryHandler {
	return &TelemetryHandler{usecase: usecase}
}

func (h *TelemetryHandler) IngestReport(w http.ResponseWriter, r *http.Request) {
	var payload domain.TelemetryReportPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid telemetry payload format", nil)
		return
	}

	if payload.ServerID == uuid.Nil {
		serverIDHeader := r.Header.Get("X-Server-ID")
		if parsedID, err := uuid.Parse(serverIDHeader); err == nil {
			payload.ServerID = parsedID
		}
	}

	if payload.ServerID == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "missing server ID in payload or header", nil)
		return
	}

	if err := h.usecase.IngestTelemetry(r.Context(), &payload); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	actions := h.usecase.GetPendingActions(payload.ServerID)

	response.Success(w, http.StatusOK, "telemetry report ingested successfully", map[string]any{
		"server_id":   payload.ServerID,
		"recorded_at": payload.Timestamp,
		"actions":     actions,
	})
}

func (h *TelemetryHandler) GetLiveMetrics(w http.ResponseWriter, r *http.Request) {
	serverIDStr := chi.URLParam(r, "id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid server ID format", nil)
		return
	}

	metric, err := h.usecase.GetServerLiveMetrics(r.Context(), serverID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "live metric not found for server", nil)
		return
	}

	response.Success(w, http.StatusOK, "live metric retrieved successfully", metric)
}

func (h *TelemetryHandler) GetMetricHistory(w http.ResponseWriter, r *http.Request) {
	serverIDStr := chi.URLParam(r, "id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid server ID format", nil)
		return
	}

	durationParam := r.URL.Query().Get("duration")
	var duration time.Duration
	switch durationParam {
	case "1h":
		duration = 1 * time.Hour
	case "6h":
		duration = 6 * time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	default:
		duration = 1 * time.Hour
	}

	history, err := h.usecase.GetServerMetricHistory(r.Context(), serverID, duration)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to query metric history", nil)
		return
	}

	response.Success(w, http.StatusOK, "metric history retrieved successfully", history)
}
