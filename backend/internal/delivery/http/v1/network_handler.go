package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/network"
)

type NetworkHandler struct {
	netUsecase *network.UseCase
}

func NewNetworkHandler(uc *network.UseCase) *NetworkHandler {
	return &NetworkHandler{netUsecase: uc}
}

// CreateNetwork menangani HTTP POST /api/v1/networks untuk membuat VPC / Network baru.
func (h *NetworkHandler) CreateNetwork(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	var req domain.CreateNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	net, err := h.netUsecase.CreateNetwork(r.Context(), orgID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to create network", err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "Network created successfully", net)
}

// ListNetworks menangani HTTP GET /api/v1/networks.
func (h *NetworkHandler) ListNetworks(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	list, err := h.netUsecase.ListNetworks(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list networks", err.Error())
		return
	}

	if list == nil {
		list = []domain.Network{}
	}

	response.Success(w, http.StatusOK, "Networks retrieved successfully", list)
}

// DeleteNetwork menangani HTTP DELETE /api/v1/networks/{id}.
func (h *NetworkHandler) DeleteNetwork(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	if err := h.netUsecase.DeleteNetwork(r.Context(), orgID, id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete network", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Network deleted successfully", nil)
}

// CreateFirewallRule menangani HTTP POST /api/v1/firewall-rules.
func (h *NetworkHandler) CreateFirewallRule(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	var req domain.CreateFirewallRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	rule, err := h.netUsecase.CreateFirewallRule(r.Context(), orgID, req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Failed to create firewall rule", err.Error())
		return
	}

	response.Success(w, http.StatusCreated, "Firewall rule created successfully", rule)
}

// ListFirewallRules menangani HTTP GET /api/v1/firewall-rules.
func (h *NetworkHandler) ListFirewallRules(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	list, err := h.netUsecase.ListFirewallRules(r.Context(), orgID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list firewall rules", err.Error())
		return
	}

	if list == nil {
		list = []domain.FirewallRule{}
	}

	response.Success(w, http.StatusOK, "Firewall rules retrieved successfully", list)
}

// DeleteFirewallRule menangani HTTP DELETE /api/v1/firewall-rules/{id}.
func (h *NetworkHandler) DeleteFirewallRule(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, "Invalid request", "Organization ID not found in session")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	if err := h.netUsecase.DeleteFirewallRule(r.Context(), orgID, id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete firewall rule", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Firewall rule deleted successfully", nil)
}
