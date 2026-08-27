package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type DomainHandler struct {
	domainUsecase domain.DomainUsecase
}

func NewDomainHandler(uc domain.DomainUsecase) *DomainHandler {
	return &DomainHandler{domainUsecase: uc}
}

// CreateDomain handles POST /api/v1/domains
func (h *DomainHandler) CreateDomain(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	var req domain.CreateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	d, err := h.domainUsecase.CreateDomain(r.Context(), orgID, &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to register custom domain", err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "Custom domain registered successfully", d)
}

// ListDomains handles GET /api/v1/domains
func (h *DomainHandler) ListDomains(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	domains, err := h.domainUsecase.ListDomains(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve custom domains", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Custom domains retrieved successfully", domains)
}

// GetDomain handles GET /api/v1/domains/{id}
func (h *DomainHandler) GetDomain(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid domain ID", err.Error())
		return
	}

	d, err := h.domainUsecase.GetDomain(r.Context(), orgID, id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Domain not found", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Domain retrieved successfully", d)
}

// DeleteDomain handles DELETE /api/v1/domains/{id}
func (h *DomainHandler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid domain ID", err.Error())
		return
	}

	if err := h.domainUsecase.DeleteDomain(r.Context(), orgID, id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete custom domain", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Custom domain deleted successfully", nil)
}

// VerifyDomain handles POST /api/v1/domains/{id}/verify
func (h *DomainHandler) VerifyDomain(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid domain ID", err.Error())
		return
	}

	verification, err := h.domainUsecase.VerifyDomain(r.Context(), orgID, id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Domain verification failed", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Domain verification completed", verification)
}
