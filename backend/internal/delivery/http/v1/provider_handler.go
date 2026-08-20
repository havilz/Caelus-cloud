package v1

import (
	"net/http"

	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/provider"
)

type ProviderHandler struct {
	credUsecase provider.CredentialUsecase
}

// NewProviderHandler menginisialisasi HTTP Handler untuk operasi data provider cloud dan kredensial.
// Parameter uc merupakan implementasi provider.CredentialUsecase.
// Mengembalikan pointer *ProviderHandler.
func NewProviderHandler(uc provider.CredentialUsecase) *ProviderHandler {
	return &ProviderHandler{credUsecase: uc}
}

// ListProviders menangani HTTP GET /api/v1/providers untuk mengambil seluruh daftar provider cloud yang didukung sistem beserta ID uniknya.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request.
func (h *ProviderHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.credUsecase.ListSupportedProviders(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar provider", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Daftar provider yang didukung berhasil diambil", providers)
}
