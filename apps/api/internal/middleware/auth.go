package middleware

import (
	"net/http"
	"strings"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

func Authenticate(tokens auth.TokenManager, users repository.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httpx.Error(w, http.StatusUnauthorized, "unauthorized", "Missing Authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				httpx.Error(w, http.StatusUnauthorized, "unauthorized", "Malformed Authorization header")
				return
			}

			claims, err := tokens.Parse(parts[1])
			if err != nil || claims.Type != "access" {
				httpx.Error(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired token")
				return
			}
			if users != nil {
				user, err := users.FindByID(r.Context(), claims.UserID)
				if err != nil || normalizeSessionVersion(user.SessionVersion) != normalizeSessionVersion(claims.SessionVersion) {
					httpx.Error(w, http.StatusUnauthorized, "unauthorized", "Session is no longer valid")
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

func normalizeSessionVersion(version int) int {
	if version <= 0 {
		return 1
	}
	return version
}

func RequireRole(role domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok || claims.Role != role {
				httpx.Error(w, http.StatusForbidden, "forbidden", "Insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAtLeastRole(role domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok || !domain.RoleAtLeast(claims.Role, role) {
				httpx.Error(w, http.StatusForbidden, "forbidden", "Insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
