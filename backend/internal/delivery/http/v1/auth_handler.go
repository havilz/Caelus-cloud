package v1

import (
	"encoding/json"
	"net/http"

	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/auth"
)

const (
	errInvalidPayloadFormat = "Format payload tidak valid"
)

type AuthHandler struct {
	authUsecase auth.Usecase
}

// NewAuthHandler menginisialisasi HTTP Handler untuk operasi autentikasi pengguna.
// Parameter uc merupakan implementasi auth.Usecase.
// Mengembalikan pointer *AuthHandler.
func NewAuthHandler(uc auth.Usecase) *AuthHandler {
	return &AuthHandler{authUsecase: uc}
}

// Register menangani HTTP POST /api/v1/auth/register untuk pendaftaran akun dan inisialisasi organisasi.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request yang memuat payload JSON auth.RegisterInput.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input auth.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, errInvalidPayloadFormat, err.Error())
		return
	}

	output, err := h.authUsecase.Register(r.Context(), input)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, "Registrasi akun berhasil", output)
}

// Login menangani HTTP POST /api/v1/auth/login untuk autentikasi kredensial pengguna dan penerbitan token JWT.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request yang memuat payload JSON auth.LoginInput.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input auth.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, errInvalidPayloadFormat, err.Error())
		return
	}

	output, err := h.authUsecase.Login(r.Context(), input)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Login berhasil", output)
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken menangani HTTP POST /api/v1/auth/refresh untuk perpanjangan masa aktif sesi pengguna.
// Parameter w merupakan HTTP response writer.
// Parameter r merupakan pointer HTTP request yang memuat payload JSON refreshTokenRequest.
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, errInvalidPayloadFormat, "Field refresh_token wajib diisi")
		return
	}

	tokenPair, err := h.authUsecase.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Token berhasil diperbarui", tokenPair)
}

// handleAuthError memetakan domain error autentikasi ke format respon HTTP standar.
// Parameter w merupakan HTTP response writer.
// Parameter err merupakan error domain yang terjadi.
func handleAuthError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrEmailAlreadyInUse:
		response.Error(w, http.StatusConflict, "Email sudah terdaftar", err.Error())
	case domain.ErrEmailInvalid, domain.ErrPasswordTooShort, domain.ErrBadRequest:
		response.Error(w, http.StatusBadRequest, "Data registrasi tidak valid", err.Error())
	case domain.ErrInvalidCredentials:
		response.Error(w, http.StatusUnauthorized, "Email atau kata sandi tidak sesuai", err.Error())
	case domain.ErrUserInactive:
		response.Error(w, http.StatusForbidden, "Akun pengguna dinonaktifkan", err.Error())
	case domain.ErrInvalidToken, domain.ErrUnauthorized:
		response.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau telah berakhir", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "Terjadi kesalahan internal server", err.Error())
	}
}
