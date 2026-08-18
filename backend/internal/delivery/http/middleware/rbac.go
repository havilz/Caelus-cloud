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

// RequireOrganizationRole mengembalikan middleware HTTP yang memvalidasi bahwa pengguna terautentikasi memiliki peran yang diizinkan dalam organisasi target.
// Parameter orgRepo merupakan implementasi domain.OrganizationRepository untuk memverifikasi keanggotaan pengguna.
// Parameter allowedRoles merupakan daftar peran yang berhak mengakses rute endpoint.
// Mengembalikan fungsi middleware func(http.Handler) http.Handler.
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

// GetMemberFromContext mengambil entitas *domain.OrganizationMember dari konteks request HTTP.
// Parameter ctx merupakan konteks request HTTP.
// Mengembalikan pointer *domain.OrganizationMember dan boolean bernilai true jika data anggota ditemukan.
func GetMemberFromContext(ctx context.Context) (*domain.OrganizationMember, bool) {
	member, ok := ctx.Value(memberKey).(*domain.OrganizationMember)
	return member, ok
}

// resolveOrganizationID mengekstrak UUID organisasi dari URL parameter, header HTTP, atau klaim konteks.
// Parameter r merupakan pointer request HTTP yang sedang diproses.
// Mengembalikan uuid.UUID organisasi yang ditemukan atau error jika identifier tidak tersedia atau format cacat.
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

	if orgID, ok := GetOrganizationIDFromContext(r.Context()); ok && orgID != uuid.Nil {
		return orgID, nil
	}

	return uuid.Nil, domain.ErrNotFound
}

// isRoleAuthorized memeriksa apakah peran pengguna memenuhi salah satu peran yang diizinkan atau memiliki hierarki peran yang lebih tinggi.
// Parameter userRole merupakan peran yang dimiliki pengguna saat ini.
// Parameter allowedRoles merupakan daftar peran yang diperbolehkan.
// Mengembalikan true jika peran pengguna diizinkan.
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

// getRoleRank mengonversi tipe peran organisasi menjadi nilai numerik hierarki hak akses.
// Parameter role merupakan peran organisasi.
// Mengembalikan integer bobot peran (Owner: 4, Admin: 3, Member: 2, Viewer: 1, Lainnya: 0).
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
