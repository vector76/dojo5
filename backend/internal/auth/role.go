package auth

import "net/http"

// RequireRole returns middleware that checks the user's role from JWT context
// against the list of allowed roles. Returns 401 if no claims are present,
// or 403 if the user's role is not in the allowed list.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := UserFromContext(r.Context())
			if claims == nil {
				writeError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !allowed[claims.Role] {
				writeError(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
