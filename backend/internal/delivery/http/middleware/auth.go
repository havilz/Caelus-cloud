package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/jwt"
)

type contextKey string

const (
	claimsKey contextKey = "auth_claims"
	userIDKey contextKey = "auth_user_id"
	orgIDKey  contextKey = "auth_org_id"
)

// Authenticate mengembalikan middleware HTTP yang memvalidasi Access Token JWT dari header Authorization.
// Parameter jwtManager merupakan instance interface jwt.Manager untuk memvalidasi tanda tangan dan klaim token.
// Mengembalikan fungsi middleware func(http.Handler) http.Handler.
func Authenticate(jwtManager jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := extractBearerToken(r.Header.Get("Authorization"))
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "Autentikasi gagal", err.Error())
				return
			}

			claims, err := jwtManager.ValidateAccessToken(tokenString)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "Token autentikasi tidak valid atau telah kedaluwarsa", err.Error())
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, claimsKey, claims)
			ctx = context.WithValue(ctx, userIDKey, claims.UserID)
			if claims.OrganizationID != nil {
				ctx = context.WithValue(ctx, orgIDKey, *claims.OrganizationID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaimsFromContext mengambil data *jwt.UserClaims dari konteks request HTTP.
// Parameter ctx merupakan konteks request HTTP.
// Mengembalikan pointer *jwt.UserClaims dan boolean bernilai true jika klaim ditemukan.
func GetClaimsFromContext(ctx context.Context) (*jwt.UserClaims, bool) {
	claims, ok := ctx.Value(claimsKey).(*jwt.UserClaims)
	return claims, ok
}

// GetUserIDFromContext mengambil UUID pengguna terautentikasi dari konteks request HTTP.
// Parameter ctx merupakan konteks request HTTP.
// Mengembalikan uuid.UUID dan boolean bernilai true jika ID pengguna ditemukan.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

// GetOrganizationIDFromContext mengambil UUID organisasi aktif dari konteks request HTTP.
// Parameter ctx merupakan konteks request HTTP.
// Mengembalikan uuid.UUID dan boolean bernilai true jika identifier organisasi ditemukan.
func GetOrganizationIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(orgIDKey).(uuid.UUID)
	return id, ok
}

// extractBearerToken memisahkan skema "Bearer" dan mengekstrak string token mentah dari header Authorization.
// Parameter authHeader merupakan nilai string header Authorization.
// Mengembalikan string token mentah atau domain.ErrUnauthorized jika format tidak sah.
func extractBearerToken(authHeader string) (string, error) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return "", domain.ErrUnauthorized
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", domain.ErrUnauthorized
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", domain.ErrUnauthorized
	}

	return token, nil
}
