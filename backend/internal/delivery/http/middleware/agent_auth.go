package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/hasher"
)

// RequireAgentAuth mengembalikan middleware yang memvalidasi request telemetri dari caelus-agent.
// Middleware ini mengekstrak server_id dari header X-Server-ID dan secret dari header Authorization,
// lalu memverifikasi hash Argon2id secret yang tersimpan di database.
// Mengembalikan 401 Unauthorized jika validasi gagal.
func RequireAgentAuth(serverRepo domain.ServerRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Ekstrak Bearer token dari Authorization header
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "agent authentication required", "missing Authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || strings.TrimSpace(parts[1]) == "" {
				response.Error(w, http.StatusUnauthorized, "agent authentication required", "invalid Authorization header format")
				return
			}
			agentSecret := strings.TrimSpace(parts[1])

			// Ekstrak server_id dari header X-Server-ID
			serverIDStr := strings.TrimSpace(r.Header.Get("X-Server-ID"))
			if serverIDStr == "" {
				response.Error(w, http.StatusUnauthorized, "agent authentication required", "missing X-Server-ID header")
				return
			}

			serverID, err := uuid.Parse(serverIDStr)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "agent authentication required", "invalid X-Server-ID format")
				return
			}

			// Ambil server beserta hash secret dari database
			srv, err := serverRepo.GetByIDWithSecret(r.Context(), serverID)
			if err != nil {
				// Kembalikan 401 (bukan 404) untuk mencegah enumerasi UUID server
				response.Error(w, http.StatusUnauthorized, "agent authentication failed", "server not found or unauthorized")
				return
			}

			// Server belum diset secret-nya (server lama sebelum migrasi)
			if srv.AgentSecretHash == nil || *srv.AgentSecretHash == "" {
				response.Error(w, http.StatusUnauthorized, "agent authentication failed", "agent secret not configured for this server")
				return
			}

			// Verifikasi secret menggunakan constant-time Argon2id comparison
			match, err := hasher.Compare(agentSecret, *srv.AgentSecretHash)
			if err != nil || !match {
				response.Error(w, http.StatusUnauthorized, "agent authentication failed", "invalid agent secret")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
