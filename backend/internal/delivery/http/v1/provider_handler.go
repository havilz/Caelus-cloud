package v1

import (
	"net/http"

	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/provider"
)

type ProviderHandler struct {
	credUsecase provider.CredentialUsecase
}

func NewProviderHandler(uc provider.CredentialUsecase) *ProviderHandler {
	return &ProviderHandler{credUsecase: uc}
}

func (h *ProviderHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.credUsecase.ListSupportedProviders(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar provider", err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Daftar provider yang didukung berhasil diambil", providers)
}
