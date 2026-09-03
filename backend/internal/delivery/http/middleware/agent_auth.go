package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/hasher"
)

func RequireAgentAuth(serverRepo domain.ServerRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

			srv, err := serverRepo.GetByIDWithSecret(r.Context(), serverID)
			if err != nil {

				response.Error(w, http.StatusUnauthorized, "agent authentication failed", "server not found or unauthorized")
				return
			}

			if srv.AgentSecretHash == nil || *srv.AgentSecretHash == "" {
				cleanID := strings.ReplaceAll(srv.ID.String(), "-", "")
				prefix16 := cleanID
				if len(prefix16) > 16 {
					prefix16 = prefix16[:16]
				}
				expectedDefault := "caelus_agent_sec_" + prefix16
				if agentSecret == expectedDefault {

					if hash, err := hasher.Hash(agentSecret, nil); err == nil {
						prefix := agentSecret
						if len(prefix) > 16 {
							prefix = prefix[:16]
						}
						_ = serverRepo.SetAgentSecret(r.Context(), srv.ID, hash, prefix)
					}
					next.ServeHTTP(w, r)
					return
				}

				response.Error(w, http.StatusUnauthorized, "agent authentication failed", "agent secret not configured for this server")
				return
			}

			match, err := hasher.Compare(agentSecret, *srv.AgentSecretHash)
			if err != nil || !match {
				response.Error(w, http.StatusUnauthorized, "agent authentication failed", "invalid agent secret")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
