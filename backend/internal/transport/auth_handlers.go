package transport

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/racecontrol/backend/internal/auth"
	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/sessions"
)

const (
	clientPlatformHeader = "X-Client-Platform"
	nativeClientPlatform = "native"
)

type Authenticator interface {
	Login(ctx context.Context, credentials auth.Credentials, client sessions.Client) (auth.LoginResult, error)
}

type SessionRevoker interface {
	Revoke(ctx context.Context, token string) error
}

type CookieConfig struct {
	Secure bool
	MaxAge time.Duration
}

type authHandlers struct {
	authenticator Authenticator
	sessions      SessionRevoker
	cookie        CookieConfig
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TeamID   *int64 `json:"teamId"`
}

type loginResponse struct {
	Token string       `json:"token,omitempty"`
	User  userResponse `json:"user"`
}

func (h *authHandlers) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}

	result, err := h.authenticator.Login(r.Context(),
		auth.Credentials{Username: request.Username, Password: request.Password},
		clientFromRequest(r))
	if err != nil {
		writeError(w, r, err)
		return
	}

	response := loginResponse{User: toUserResponse(result.Identity)}
	if isNativeClient(r) {
		response.Token = result.Token
	} else {
		http.SetCookie(w, h.sessionCookie(result.Token, int(h.cookie.MaxAge.Seconds())))
	}
	writeJSON(w, http.StatusOK, response)
}

// logout answers 204 even without a valid session: the browser cannot delete an
// httpOnly cookie by itself, so an expired one must still be cleared here.
func (h *authHandlers) logout(w http.ResponseWriter, r *http.Request) {
	if token := auth.TokenFromRequest(r); token != "" {
		if err := h.sessions.Revoke(r.Context(), token); err != nil {
			writeError(w, r, err)
			return
		}
	}
	http.SetCookie(w, h.sessionCookie("", -1))
	w.WriteHeader(http.StatusNoContent)
}

func (h *authHandlers) me(w http.ResponseWriter, r *http.Request) {
	identity, found := auth.IdentityFromContext(r.Context())
	if !found {
		writeError(w, r, domain.ErrUnauthenticated)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(identity))
}

func (h *authHandlers) sessionCookie(value string, maxAgeSeconds int) *http.Cookie {
	return &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAgeSeconds,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}
}

func toUserResponse(identity domain.Identity) userResponse {
	return userResponse{
		ID:       identity.UserID,
		Username: identity.Username,
		Role:     string(identity.Role),
		TeamID:   identity.TeamID,
	}
}

func isNativeClient(r *http.Request) bool {
	return r.Header.Get(clientPlatformHeader) == nativeClientPlatform
}

func clientFromRequest(r *http.Request) sessions.Client {
	return sessions.Client{IP: remoteIP(r.RemoteAddr), UserAgent: r.UserAgent()}
}

func remoteIP(remoteAddr string) *netip.Addr {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return nil
	}
	unmapped := addr.Unmap()
	return &unmapped
}
