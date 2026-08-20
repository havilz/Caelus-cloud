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
	"github.com/havilz/caelus-cloud/backend/internal/usecase/server"
)

const (
	errInvalidRequestParams = "Parameter tidak valid"
	errInvalidRequestBody   = "Permintaan tidak valid"
)

type ServerHandler struct {
	serverUsecase server.ServerUsecase
}

// NewServerHandler menginisialisasi HTTP Handler untuk operasi manajemen dan siklus hidup server VPS.
// Parameter uc merupakan implementasi server.ServerUsecase.
// Mengembalikan pointer *ServerHandler.
func NewServerHandler(uc server.ServerUsecase) *ServerHandler {
	return &ServerHandler{serverUsecase: uc}
}

// ListServers menangani HTTP GET /api/v1/servers untuk mengambil daftar server organisasi dengan paginasi.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request.
func (h *ServerHandler) ListServers(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, errInvalidRequestBody, "Organization ID tidak ditemukan pada sesi request")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	servers, total, err := h.serverUsecase.ListServers(r.Context(), orgID, page, limit)
	if err != nil {
		handleServerError(w, err)
		return
	}

	response.Paginated(w, http.StatusOK, "Daftar server berhasil diambil", servers, page, limit, total)
}

// CreateServer menangani HTTP POST /api/v1/servers untuk provisioning server VPS baru.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request yang memuat payload JSON server.CreateServerInput.
func (h *ServerHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusBadRequest, errInvalidRequestBody, "Organization ID tidak ditemukan pada sesi request")
		return
	}

	var input server.CreateServerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, errInvalidPayloadFormat, err.Error())
		return
	}
	input.OrganizationID = orgID

	created, err := h.serverUsecase.CreateServer(r.Context(), input)
	if err != nil {
		handleServerError(w, err)
		return
	}

	middleware.SetAuditMetadata(r.Context(), "server", created.ID.String())
	response.Success(w, http.StatusCreated, "Server berhasil dibuat dan dalam proses running", created)
}

// GetServer menangani HTTP GET /api/v1/servers/{id} untuk mengambil data detail server.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request.
func (h *ServerHandler) GetServer(w http.ResponseWriter, r *http.Request) {
	orgID, serverID, err := extractOrgAndServerID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, errInvalidRequestParams, err.Error())
		return
	}

	srv, err := h.serverUsecase.GetServer(r.Context(), orgID, serverID)
	if err != nil {
		handleServerError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Detail server berhasil diambil", srv)
}

// RebootServer menangani HTTP POST /api/v1/servers/{id}/reboot untuk melakukan reboot server VPS.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request.
func (h *ServerHandler) RebootServer(w http.ResponseWriter, r *http.Request) {
	orgID, serverID, err := extractOrgAndServerID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, errInvalidRequestParams, err.Error())
		return
	}

	if err := h.serverUsecase.RebootServer(r.Context(), orgID, serverID); err != nil {
		handleServerError(w, err)
		return
	}

	middleware.SetAuditMetadata(r.Context(), "server", serverID.String())
	response.Success(w, http.StatusOK, "Perintah reboot server berhasil dikirim", nil)
}

// ShutdownServer menangani HTTP POST /api/v1/servers/{id}/shutdown untuk mematikan instance server VPS.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request.
func (h *ServerHandler) ShutdownServer(w http.ResponseWriter, r *http.Request) {
	orgID, serverID, err := extractOrgAndServerID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, errInvalidRequestParams, err.Error())
		return
	}

	if err := h.serverUsecase.ShutdownServer(r.Context(), orgID, serverID); err != nil {
		handleServerError(w, err)
		return
	}

	middleware.SetAuditMetadata(r.Context(), "server", serverID.String())
	response.Success(w, http.StatusOK, "Perintah shutdown server berhasil dikirim", nil)
}

// StartServer menangani HTTP POST /api/v1/servers/{id}/start untuk menyalakan instance server VPS yang berhenti.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request.
func (h *ServerHandler) StartServer(w http.ResponseWriter, r *http.Request) {
	orgID, serverID, err := extractOrgAndServerID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, errInvalidRequestParams, err.Error())
		return
	}

	if err := h.serverUsecase.StartServer(r.Context(), orgID, serverID); err != nil {
		handleServerError(w, err)
		return
	}

	middleware.SetAuditMetadata(r.Context(), "server", serverID.String())
	response.Success(w, http.StatusOK, "Perintah start server berhasil dikirim", nil)
}

// ResizeServer menangani HTTP PATCH /api/v1/servers/{id}/resize untuk mengubah spesifikasi vCPU/RAM/Disk server.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request yang memuat payload JSON server.ResizeServerInput.
func (h *ServerHandler) ResizeServer(w http.ResponseWriter, r *http.Request) {
	orgID, serverID, err := extractOrgAndServerID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, errInvalidRequestParams, err.Error())
		return
	}

	var input server.ResizeServerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, errInvalidPayloadFormat, err.Error())
		return
	}

	if err := h.serverUsecase.ResizeServer(r.Context(), orgID, serverID, input); err != nil {
		handleServerError(w, err)
		return
	}

	middleware.SetAuditMetadata(r.Context(), "server", serverID.String())
	response.Success(w, http.StatusOK, "Perubahan spesifikasi server berhasil diterapkan", nil)
}

// DeleteServer menangani HTTP DELETE /api/v1/servers/{id} untuk menterminasi dan menghapus server dari sistem.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request.
func (h *ServerHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	orgID, serverID, err := extractOrgAndServerID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, errInvalidRequestParams, err.Error())
		return
	}

	if err := h.serverUsecase.DeleteServer(r.Context(), orgID, serverID); err != nil {
		handleServerError(w, err)
		return
	}

	middleware.SetAuditMetadata(r.Context(), "server", serverID.String())
	response.Success(w, http.StatusOK, "Server berhasil dihapus", nil)
}

// extractOrgAndServerID mengambil OrganizationID dari context dan ServerID dari parameter URL.
// Parameter r merupakan pointer HTTP request.
// Mengembalikan UUID organisasi, UUID server, dan error jika format ID tidak valid.
func extractOrgAndServerID(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, uuid.Nil, domain.ErrUnauthorized
	}

	idParam := chi.URLParam(r, "id")
	serverID, err := uuid.Parse(idParam)
	if err != nil {
		return uuid.Nil, uuid.Nil, domain.ErrBadRequest
	}

	return orgID, serverID, nil
}

// handleServerError memetakan domain error server ke format respon HTTP standar.
// Parameter w merupakan HTTP response writer.
// Parameter err merupakan error domain yang terjadi.
func handleServerError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrNotFound:
		response.Error(w, http.StatusNotFound, "Resource tidak ditemukan", err.Error())
	case domain.ErrProviderNotSupported:
		response.Error(w, http.StatusBadRequest, "Provider cloud belum didukung", err.Error())
	case domain.ErrForbidden:
		response.Error(w, http.StatusForbidden, "Akses dilarang", err.Error())
	case domain.ErrBadRequest:
		response.Error(w, http.StatusBadRequest, "Permintaan data tidak valid", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "Terjadi kesalahan internal pada server", err.Error())
	}
}
