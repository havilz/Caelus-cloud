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
	"github.com/havilz/caelus-cloud/backend/internal/usecase/backup"
)

type BackupHandler struct {
	usecase backup.BackupUsecase
}

func NewBackupHandler(usecase backup.BackupUsecase) *BackupHandler {
	return &BackupHandler{usecase: usecase}
}

type CreatePolicyRequest struct {
	ServerID       uuid.UUID  `json:"server_id"`
	BucketID       *uuid.UUID `json:"bucket_id,omitempty"`
	Name           string     `json:"name"`
	CronExpression string     `json:"cron_expression"`
	RetentionDays  int        `json:"retention_days"`
	IncludeDisks   bool       `json:"include_disks"`
}

func (h *BackupHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	var req CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	policy, err := h.usecase.CreatePolicy(r.Context(), domain.CreateBackupPolicyInput{
		OrganizationID: orgID,
		ServerID:       req.ServerID,
		BucketID:       req.BucketID,
		Name:           req.Name,
		CronExpression: req.CronExpression,
		RetentionDays:  req.RetentionDays,
		IncludeDisks:   req.IncludeDisks,
	})
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusCreated, "backup policy created successfully", policy)
}

func (h *BackupHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	policies, err := h.usecase.ListPolicies(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "backup policies retrieved successfully", policies)
}

func (h *BackupHandler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	policyID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid policy ID format", nil)
		return
	}

	if err := h.usecase.DeletePolicy(r.Context(), orgID, policyID); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "backup policy deleted successfully", nil)
}

type TriggerBackupRequest struct {
	BackupName string     `json:"backup_name,omitempty"`
	PolicyID   *uuid.UUID `json:"policy_id,omitempty"`
}

func (h *BackupHandler) TriggerBackup(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	serverIDStr := chi.URLParam(r, "server_id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid server ID format", nil)
		return
	}

	var req TriggerBackupRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	record, err := h.usecase.TriggerBackup(r.Context(), orgID, serverID, req.PolicyID, req.BackupName)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusCreated, "backup triggered successfully", record)
}

func (h *BackupHandler) ListRecords(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	records, total, err := h.usecase.ListRecords(r.Context(), orgID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.Paginated(w, http.StatusOK, "backup records retrieved successfully", records, page, limit, int64(total))
}

func (h *BackupHandler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	recordID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid record ID format", nil)
		return
	}

	if err := h.usecase.DeleteRecord(r.Context(), orgID, recordID); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "backup record deleted successfully", nil)
}
