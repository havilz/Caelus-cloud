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
	"github.com/havilz/caelus-cloud/backend/internal/usecase/monitoring"
)

type AlertHandler struct {
	usecase monitoring.MonitoringUsecase
}

func NewAlertHandler(usecase monitoring.MonitoringUsecase) *AlertHandler {
	return &AlertHandler{usecase: usecase}
}

func (h *AlertHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized or missing organization context", nil)
		return
	}

	statusParam := r.URL.Query().Get("status")
	var statusFilter *domain.AlertStatus
	if statusParam != "" {
		st := domain.AlertStatus(statusParam)
		statusFilter = &st
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	alerts, total, err := h.usecase.ListAlerts(r.Context(), orgID, statusFilter, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to query alerts", nil)
		return
	}

	response.Paginated(w, http.StatusOK, "alerts retrieved successfully", alerts, page, limit, total)
}

func (h *AlertHandler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized access", nil)
		return
	}

	alertIDStr := chi.URLParam(r, "id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid alert ID format", nil)
		return
	}

	if err := h.usecase.AcknowledgeAlert(r.Context(), alertID, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to acknowledge alert", nil)
		return
	}

	response.Success(w, http.StatusOK, "alert acknowledged successfully", map[string]any{"id": alertID})
}

func (h *AlertHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized access", nil)
		return
	}

	alertIDStr := chi.URLParam(r, "id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid alert ID format", nil)
		return
	}

	if err := h.usecase.ResolveAlert(r.Context(), alertID, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve alert", nil)
		return
	}

	response.Success(w, http.StatusOK, "alert resolved successfully", map[string]any{"id": alertID})
}

func (h *AlertHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized or missing organization context", nil)
		return
	}

	rules, err := h.usecase.ListAlertRules(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to query alert rules", nil)
		return
	}

	response.Success(w, http.StatusOK, "alert rules retrieved successfully", rules)
}

func (h *AlertHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized or missing organization context", nil)
		return
	}

	var req struct {
		ServerID        *uuid.UUID           `json:"server_id,omitempty"`
		Name            string               `json:"name"`
		MetricName      string               `json:"metric_name"`
		Operator        string               `json:"operator"`
		Threshold       float64              `json:"threshold"`
		DurationSeconds int                  `json:"duration_seconds"`
		Severity        domain.AlertSeverity `json:"severity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body format", nil)
		return
	}

	operator := req.Operator
	if operator == "" {
		operator = ">"
	}
	severity := req.Severity
	if severity == "" {
		severity = domain.AlertSeverityWarning
	}
	duration := req.DurationSeconds
	if duration <= 0 {
		duration = 60
	}

	rule := &domain.AlertRule{
		OrganizationID:  orgID,
		ServerID:        req.ServerID,
		Name:            req.Name,
		MetricName:      req.MetricName,
		Operator:        operator,
		Threshold:       req.Threshold,
		DurationSeconds: duration,
		Severity:        severity,
		IsEnabled:       true,
	}

	if err := h.usecase.CreateAlertRule(r.Context(), rule); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusCreated, "alert rule created successfully", rule)
}

func (h *AlertHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	ruleIDStr := chi.URLParam(r, "id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid rule ID format", nil)
		return
	}

	if err := h.usecase.DeleteAlertRule(r.Context(), ruleID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete alert rule", nil)
		return
	}

	response.Success(w, http.StatusOK, "alert rule deleted successfully", map[string]any{"id": ruleID})
}
