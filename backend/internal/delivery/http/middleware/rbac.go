package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

const memberKey contextKey = "auth_org_member"

func RequireOrganizationRole(orgRepo domain.OrganizationRepository, allowedRoles ...domain.OrganizationRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "Autentikasi diperlukan", "Identitas pengguna tidak ditemukan pada sesi request")
				return
			}

			orgID, err := resolveOrganizationID(r)
			if err != nil {
				response.Error(w, http.StatusBadRequest, "Permintaan tidak valid", "Identifier organisasi target tidak ditemukan")
				return
			}

			member, err := orgRepo.GetMember(r.Context(), orgID, userID)
			if err != nil {
				response.Error(w, http.StatusForbidden, "Akses ditolak", "Pengguna bukan anggota dari organisasi ini")
				return
			}

			if !isRoleAuthorized(member.Role, allowedRoles) {
				response.Error(w, http.StatusForbidden, "Akses ditolak", "Hak akses peran Anda tidak mencukupi untuk melakukan tindakan ini")
				return
			}

			ctx := context.WithValue(r.Context(), memberKey, member)
			ctx = context.WithValue(ctx, orgIDKey, orgID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetMemberFromContext(ctx context.Context) (*domain.OrganizationMember, bool) {
	member, ok := ctx.Value(memberKey).(*domain.OrganizationMember)
	return member, ok
}

func resolveOrganizationID(r *http.Request) (uuid.UUID, error) {
	if param := chi.URLParam(r, "org_id"); param != "" {
		return uuid.Parse(param)
	}

	if param := chi.URLParam(r, "organization_id"); param != "" {
		return uuid.Parse(param)
	}

	if header := r.Header.Get("X-Organization-ID"); header != "" {
		return uuid.Parse(header)
	}

	if orgID, ok := GetOrgIDFromContext(r.Context()); ok && orgID != uuid.Nil {
		return orgID, nil
	}

	if orgID, ok := GetOrganizationIDFromContext(r.Context()); ok && orgID != uuid.Nil {
		return orgID, nil
	}

	return uuid.Nil, domain.ErrNotFound
}

func RequireAdmin(orgRepo domain.OrganizationRepository) func(http.Handler) http.Handler {
	return RequireOrganizationRole(orgRepo, domain.RoleAdmin)
}

func RequireOwner(orgRepo domain.OrganizationRepository) func(http.Handler) http.Handler {
	return RequireOrganizationRole(orgRepo, domain.RoleOwner)
}

func isRoleAuthorized(userRole domain.OrganizationRole, allowedRoles []domain.OrganizationRole) bool {
	if len(allowedRoles) == 0 {
		return true
	}

	userRank := getRoleRank(userRole)
	for _, allowed := range allowedRoles {
		allowedRank := getRoleRank(allowed)
		if userRank >= allowedRank {
			return true
		}
	}

	return false
}

func getRoleRank(role domain.OrganizationRole) int {
	switch role {
	case domain.RoleOwner:
		return 4
	case domain.RoleAdmin:
		return 3
	case domain.RoleMember:
		return 2
	case domain.RoleViewer:
		return 1
	default:
		return 0
	}
}
