package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func InjectOrgContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID, hasOrg := GetOrganizationIDFromContext(r.Context())
			if hasOrg && orgID != uuid.Nil {
				ctx := context.WithValue(r.Context(), orgContextKey, orgID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type orgContextKeyType struct{}

var orgContextKey = orgContextKeyType{}

func GetOrgIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	orgID, ok := ctx.Value(orgContextKey).(uuid.UUID)
	if ok && orgID != uuid.Nil {
		return orgID, true
	}
	return GetOrganizationIDFromContext(ctx)
}
