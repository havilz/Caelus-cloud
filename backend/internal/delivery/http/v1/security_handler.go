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
	"github.com/havilz/caelus-cloud/backend/internal/usecase/security"
)

// SecurityHandler menangani rute HTTP REST API untuk modul Sentinel Security.
type SecurityHandler struct {
	securityUsecase security.SecurityUsecase
}

// NewSecurityHandler membuat instance baru SecurityHandler.
func NewSecurityHandler(uc security.SecurityUsecase) *SecurityHandler {
	return &SecurityHandler{securityUsecase: uc}
}

// TriggerScanRequest payload body untuk memicu scan keamanan baru.
type TriggerScanRequest struct {
	ServerID *uuid.UUID       `json:"server_id,omitempty"`
	ScanType domain.ScanType `json:"scan_type,omitempty"`
}

// TriggerScan menangani HTTP POST /api/v1/security/scans
func (h *SecurityHandler) TriggerScan(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Organization ID tidak ditemukan pada sesi", nil)
		return
	}

	var req TriggerScanRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	scan, err := h.securityUsecase.TriggerScan(r.Context(), orgID, req.ServerID, req.ScanType)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memicu pemindaian keamanan", err.Error())
		return
	}

	response.Success(w, http.StatusAccepted, "Pemindaian keamanan Sentinel berhasil dipicu", scan)
}

// ListScans menangani HTTP GET /api/v1/security/scans
func (h *SecurityHandler) ListScans(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Organization ID tidak ditemukan", nil)
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

	var serverID *uuid.UUID
	if srvStr := r.URL.Query().Get("server_id"); srvStr != "" {
		if id, err := uuid.Parse(srvStr); err == nil {
			serverID = &id
		}
	}

	scans, total, err := h.securityUsecase.ListScans(r.Context(), orgID, serverID, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil riwayat scan", err.Error())
		return
	}

	response.Paginated(w, http.StatusOK, "Riwayat pemindaian berhasil diambil", scans, page, limit, int64(total))
}

// GetScan menangani HTTP GET /api/v1/security/scans/{id}
func (h *SecurityHandler) GetScan(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Organization ID tidak ditemukan", nil)
		return
	}

	scanID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Format Scan ID tidak valid", err.Error())
		return
	}

	scan, err := h.securityUsecase.GetScan(r.Context(), orgID, scanID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Data pemindaian tidak ditemukan", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Detail pemindaian berhasil diambil", scan)
}

// ListFindings menangani HTTP GET /api/v1/security/findings
func (h *SecurityHandler) ListFindings(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Organization ID tidak ditemukan", nil)
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

	var serverID *uuid.UUID
	if srvStr := r.URL.Query().Get("server_id"); srvStr != "" {
		if id, err := uuid.Parse(srvStr); err == nil {
			serverID = &id
		}
	}

	var category *domain.FindingCategory
	if catStr := r.URL.Query().Get("category"); catStr != "" {
		c := domain.FindingCategory(catStr)
		category = &c
	}

	var severity *domain.FindingSeverity
	if sevStr := r.URL.Query().Get("severity"); sevStr != "" {
		s := domain.FindingSeverity(sevStr)
		severity = &s
	}

	var status *domain.FindingStatus
	if stStr := r.URL.Query().Get("status"); stStr != "" {
		st := domain.FindingStatus(stStr)
		status = &st
	}

	findings, total, err := h.securityUsecase.ListFindings(r.Context(), orgID, serverID, category, severity, status, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar temuan", err.Error())
		return
	}

	response.Paginated(w, http.StatusOK, "Daftar temuan keamanan berhasil diambil", findings, page, limit, int64(total))
}

// GetFinding menangani HTTP GET /api/v1/security/findings/{id}
func (h *SecurityHandler) GetFinding(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Organization ID tidak ditemukan", nil)
		return
	}

	findingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Format Finding ID tidak valid", err.Error())
		return
	}

	finding, err := h.securityUsecase.GetFinding(r.Context(), orgID, findingID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Temuan keamanan tidak ditemukan", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Detail temuan berhasil diambil", finding)
}

// UpdateFindingStatusRequest payload untuk update status temuan.
type UpdateFindingStatusRequest struct {
	Status domain.FindingStatus `json:"status"`
}

// UpdateFindingStatus menangani HTTP PATCH /api/v1/security/findings/{id}/status
func (h *SecurityHandler) UpdateFindingStatus(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Organization ID tidak ditemukan", nil)
		return
	}

	findingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Format Finding ID tidak valid", err.Error())
		return
	}

	var req UpdateFindingStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload request tidak valid", err.Error())
		return
	}

	if err := h.securityUsecase.UpdateFindingStatus(r.Context(), orgID, findingID, req.Status); err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui status temuan", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Status temuan berhasil diperbarui", nil)
}

// GetPostureOverview menangani HTTP GET /api/v1/security/overview
func (h *SecurityHandler) GetPostureOverview(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Organization ID tidak ditemukan", nil)
		return
	}

	overview, err := h.securityUsecase.GetPostureOverview(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil ringkasan postur keamanan", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Postur keamanan berhasil dihitung", overview)
}

// CreateIncidentRequest payload untuk membuat insiden baru.
type CreateIncidentRequest struct {
	Title      string                 `json:"title"`
	Summary    string                 `json:"summary"`
	Severity   domain.FindingSeverity `json:"severity"`
	FindingIDs []uuid.UUID            `json:"finding_ids"`
}

// CreateIncident menangani HTTP POST /api/v1/security/incidents
func (h *SecurityHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Organization ID tidak ditemukan", nil)
		return
	}

	var req CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload request tidak valid", err.Error())
		return
	}

	incident, err := h.securityUsecase.CreateIncident(r.Context(), orgID, req.Title, req.Summary, req.Severity, req.FindingIDs)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal membuat insiden keamanan", err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "Insiden keamanan berhasil dibuat", incident)
}

// ListIncidents menangani HTTP GET /api/v1/security/incidents
func (h *SecurityHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Organization ID tidak ditemukan", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	var status *domain.IncidentStatus
	if stStr := r.URL.Query().Get("status"); stStr != "" {
		st := domain.IncidentStatus(stStr)
		status = &st
	}

	incidents, total, err := h.securityUsecase.ListIncidents(r.Context(), orgID, status, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar insiden", err.Error())
		return
	}

	response.Paginated(w, http.StatusOK, "Daftar insiden berhasil diambil", incidents, page, limit, int64(total))
}

// UpdateIncidentStatusRequest payload update status insiden.
type UpdateIncidentStatusRequest struct {
	Status domain.IncidentStatus `json:"status"`
	Notes  string                `json:"notes"`
}

// UpdateIncidentStatus menangani HTTP PATCH /api/v1/security/incidents/{id}/status
func (h *SecurityHandler) UpdateIncidentStatus(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Organization ID tidak ditemukan", nil)
		return
	}

	incidentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Format Incident ID tidak valid", err.Error())
		return
	}

	var req UpdateIncidentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload request tidak valid", err.Error())
		return
	}

	if err := h.securityUsecase.UpdateIncidentStatus(r.Context(), orgID, incidentID, req.Status, req.Notes); err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal update status insiden", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Status insiden berhasil diperbarui", nil)
}
