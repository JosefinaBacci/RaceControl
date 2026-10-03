package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/racecontrol/backend/internal/domain"
)

type fakeResolver map[string]domain.Identity

func (f fakeResolver) Resolve(_ context.Context, token string) (domain.Identity, error) {
	identity, found := f[token]
	if !found {
		return domain.Identity{}, domain.ErrUnauthenticated
	}
	return identity, nil
}

func statusWriter(w http.ResponseWriter, _ *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		w.WriteHeader(http.StatusUnauthorized)
	case errors.Is(err, domain.ErrForbidden):
		w.WriteHeader(http.StatusForbidden)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func TestRequireSessionAndRole(t *testing.T) {
	teamID := int64(1)
	resolver := fakeResolver{
		"fia-token":  {UserID: 1, Role: domain.RoleFIAAdmin},
		"team-token": {UserID: 2, Role: domain.RoleTeamAdmin, TeamID: &teamID},
	}
	guard := NewMiddleware(resolver, statusWriter)
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	fiaOnly := guard.RequireSession(guard.RequireRole(domain.RoleFIAAdmin)(ok))

	cases := []struct {
		name    string
		prepare func(*http.Request)
		want    int
	}{
		{"no credentials", func(*http.Request) {}, http.StatusUnauthorized},
		{"unknown token", func(r *http.Request) { r.Header.Set("Authorization", "Bearer nope") }, http.StatusUnauthorized},
		{"team admin on fia route", func(r *http.Request) { r.Header.Set("Authorization", "Bearer team-token") }, http.StatusForbidden},
		{"fia admin via bearer", func(r *http.Request) { r.Header.Set("Authorization", "Bearer fia-token") }, http.StatusOK},
		{"fia admin via cookie", func(r *http.Request) { r.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "fia-token"}) }, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			tc.prepare(request)
			recorder := httptest.NewRecorder()

			fiaOnly.ServeHTTP(recorder, request)

			if recorder.Code != tc.want {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.want)
			}
		})
	}
}

func TestRequireRoleWithoutSessionIsUnauthenticated(t *testing.T) {
	guard := NewMiddleware(fakeResolver{}, statusWriter)
	handler := guard.RequireRole(domain.RoleFIAAdmin)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}
