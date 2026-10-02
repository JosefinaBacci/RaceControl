package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/racecontrol/backend/internal/domain"
)

const (
	SessionCookieName = "rc_session"
	bearerPrefix      = "Bearer "
)

type SessionResolver interface {
	Resolve(ctx context.Context, token string) (domain.Identity, error)
}

type ErrorWriter func(w http.ResponseWriter, r *http.Request, err error)

type Middleware struct {
	resolver   SessionResolver
	writeError ErrorWriter
}

func NewMiddleware(resolver SessionResolver, writeError ErrorWriter) *Middleware {
	return &Middleware{resolver: resolver, writeError: writeError}
}

func (m *Middleware) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, err := m.resolver.Resolve(r.Context(), TokenFromRequest(r))
		if err != nil {
			m.writeError(w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithIdentity(r.Context(), identity)))
	})
}

func (m *Middleware) RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, found := IdentityFromContext(r.Context())
			if !found {
				m.writeError(w, r, domain.ErrUnauthenticated)
				return
			}
			if !identity.HasRole(roles...) {
				m.writeError(w, r, domain.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func TokenFromRequest(r *http.Request) string {
	if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, bearerPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
	}
	return ""
}

type identityContextKey struct{}

func WithIdentity(ctx context.Context, identity domain.Identity) context.Context {
	return context.WithValue(ctx, identityContextKey{}, identity)
}

func IdentityFromContext(ctx context.Context) (domain.Identity, bool) {
	identity, found := ctx.Value(identityContextKey{}).(domain.Identity)
	return identity, found
}
