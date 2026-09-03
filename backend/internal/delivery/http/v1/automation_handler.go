package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	automationUc "github.com/havilz/caelus-cloud/backend/internal/usecase/automation"
)

type AutomationHandler struct {
	usecase automationUc.AutomationUsecase
}

func NewAutomationHandler(uc automationUc.AutomationUsecase) *AutomationHandler {
	return &AutomationHandler{usecase: uc}
}

func (h *AutomationHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	var input domain.CreateRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	input.OrganizationID = orgID

	rule, err := h.usecase.CreateRule(r.Context(), input)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusCreated, "Automation rule created successfully", rule)
}

func (h *AutomationHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	rules, total, err := h.usecase.ListRules(r.Context(), orgID, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve automation rules", nil)
		return
	}

	response.Paginated(w, http.StatusOK, "Automation rules retrieved successfully", rules, page, limit, int64(total))
}

func (h *AutomationHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	ruleIDStr := chi.URLParam(r, "id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid rule ID format", nil)
		return
	}

	rule, err := h.usecase.GetRule(r.Context(), orgID, ruleID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Automation rule not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "Automation rule retrieved successfully", rule)
}

func (h *AutomationHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	ruleIDStr := chi.URLParam(r, "id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid rule ID format", nil)
		return
	}

	var input domain.UpdateRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload format", nil)
		return
	}

	rule, err := h.usecase.UpdateRule(r.Context(), orgID, ruleID, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Automation rule not found", nil)
			return
		}
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "Automation rule updated successfully", rule)
}

func (h *AutomationHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	ruleIDStr := chi.URLParam(r, "id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid rule ID format", nil)
		return
	}

	if err := h.usecase.DeleteRule(r.Context(), orgID, ruleID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Automation rule not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "Automation rule deleted successfully", nil)
}

func (h *AutomationHandler) TestRule(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	ruleIDStr := chi.URLParam(r, "id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid rule ID format", nil)
		return
	}

	var reqBody struct {
		MockData map[string]any `json:"mock_data"`
	}
	_ = json.NewDecoder(r.Body).Decode(&reqBody)

	log, err := h.usecase.TestRule(r.Context(), orgID, ruleID, reqBody.MockData)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Automation rule not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "Automation rule test executed successfully", log)
}

func (h *AutomationHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var ruleIDFilter *uuid.UUID
	if ruleIDStr := r.URL.Query().Get("rule_id"); ruleIDStr != "" {
		if parsed, err := uuid.Parse(ruleIDStr); err == nil {
			ruleIDFilter = &parsed
		}
	}

	var statusFilter *domain.ExecutionStatus
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		st := domain.ExecutionStatus(statusStr)
		statusFilter = &st
	}

	logs, total, err := h.usecase.ListLogs(r.Context(), orgID, ruleIDFilter, statusFilter, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve execution logs", nil)
		return
	}

	response.Paginated(w, http.StatusOK, "Execution logs retrieved successfully", logs, page, limit, int64(total))
}
