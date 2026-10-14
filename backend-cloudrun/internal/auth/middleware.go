package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey struct{}

// FromContext returns verified claims, or nil for anonymous (guest) requests.
func FromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(ctxKey{}).(*Claims)
	return c
}

// Middleware parses an optional Bearer token. A present-but-invalid token is rejected with 401;
// requests without a token continue as anonymous. Privileged routes are guarded by RequireRole.
// A nil verifier disables verification entirely (local/demo mode).
func Middleware(v *Verifier, next http.Handler) http.Handler {
	if v == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		token, ok := strings.CutPrefix(h, "Bearer ")
		if !ok {
			http.Error(w, `{"error":"invalid authorization header"}`, http.StatusUnauthorized)
			return
		}
		claims, err := v.Verify(r.Context(), token)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, claims)))
	})
}

// RequireRole allows only the listed roles. When verification is disabled (v == nil) it is a no-op.
func RequireRole(v *Verifier, next http.HandlerFunc, roles ...string) http.HandlerFunc {
	if v == nil {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		c := FromContext(r.Context())
		if c == nil {
			http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
			return
		}
		for _, role := range roles {
			if c.Role == role {
				next(w, r)
				return
			}
		}
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
	}
}
