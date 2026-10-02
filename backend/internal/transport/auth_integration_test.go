package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/auth"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/sessions"
	"github.com/racecontrol/backend/internal/testdb"
	"github.com/racecontrol/backend/internal/transport"
)

const (
	testPassword  = "una frase larga de prueba"
	wrongPassword = "otra frase que no es"
)

type harness struct {
	t       *testing.T
	pool    *pgxpool.Pool
	queries *dbgen.Queries
	hasher  *auth.PasswordHasher
	server  *httptest.Server
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	pool := testdb.Pool(t)
	hasher, err := auth.NewPasswordHasher(auth.Argon2Params{MemoryKiB: 1024, Iterations: 1, Parallelism: 1})
	if err != nil {
		t.Fatalf("NewPasswordHasher: %v", err)
	}

	queries := dbgen.New(pool)
	sessionService := sessions.NewService(queries, sessions.Config{IdleTTL: time.Hour, AbsoluteTTL: 24 * time.Hour})
	router := transport.NewRouter(transport.Dependencies{
		Pool:           pool,
		Authenticator:  auth.NewAuthenticator(queries, hasher, sessionService),
		Sessions:       sessionService,
		Cookie:         transport.CookieConfig{MaxAge: 24 * time.Hour},
		AllowedOrigins: []string{"http://localhost:8081"},
	})

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &harness{t: t, pool: pool, queries: queries, hasher: hasher, server: server}
}

func (h *harness) createUser(role domain.Role, teamID int64) string {
	h.t.Helper()

	hash, err := h.hasher.Hash(testPassword)
	if err != nil {
		h.t.Fatalf("Hash: %v", err)
	}
	user, err := h.queries.CreateUser(context.Background(), dbgen.CreateUserParams{
		Username:     testdb.UniqueUsername(h.t),
		PasswordHash: hash,
		Role:         string(role),
		TeamID:       pgtype.Int8{Int64: teamID, Valid: teamID != 0},
	})
	if err != nil {
		h.t.Fatalf("CreateUser: %v", err)
	}
	return user.Username
}

func (h *harness) deactivate(username string) {
	h.t.Helper()

	_, err := h.pool.Exec(context.Background(),
		`UPDATE users SET is_active = false, deactivated_at = now() WHERE username = $1`, username)
	if err != nil {
		h.t.Fatalf("deactivate %s: %v", username, err)
	}
}

func (h *harness) browser() *http.Client {
	h.t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		h.t.Fatalf("cookiejar: %v", err)
	}
	return &http.Client{Jar: jar}
}

type response struct {
	status int
	body   map[string]any
	raw    string
}

func (h *harness) do(client *http.Client, method, path string, payload any, headers map[string]string) response {
	h.t.Helper()

	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			h.t.Fatalf("marshal: %v", err)
		}
		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, h.server.URL+path, body)
	if err != nil {
		h.t.Fatalf("new request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	for name, value := range headers {
		request.Header.Set(name, value)
	}

	result, err := client.Do(request)
	if err != nil {
		h.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer result.Body.Close()

	raw, err := io.ReadAll(result.Body)
	if err != nil {
		h.t.Fatalf("read body: %v", err)
	}
	decoded := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &decoded); err != nil {
			h.t.Fatalf("decode body %q: %v", raw, err)
		}
	}

	return response{status: result.StatusCode, body: decoded, raw: string(raw)}
}

func (h *harness) login(client *http.Client, username, password string) response {
	h.t.Helper()
	return h.do(client, http.MethodPost, "/auth/login", map[string]string{"username": username, "password": password}, nil)
}

func expectStatus(t *testing.T, got response, want int) {
	t.Helper()
	if got.status != want {
		t.Fatalf("status = %d, want %d; body: %s", got.status, want, got.raw)
	}
}

func TestLoginWithCookieIdentifiesRole(t *testing.T) {
	h := newHarness(t)
	username := h.createUser(domain.RoleTeamAdmin, 1)
	browser := h.browser()

	login := h.login(browser, username, testPassword)
	expectStatus(t, login, http.StatusOK)
	if _, leaked := login.body["token"]; leaked {
		t.Fatal("web login must not expose the session token in the body")
	}

	me := h.do(browser, http.MethodGet, "/auth/me", nil, nil)
	expectStatus(t, me, http.StatusOK)
	if me.body["role"] != string(domain.RoleTeamAdmin) || me.body["teamId"] != float64(1) {
		t.Fatalf("me = %v, want team_admin of team 1", me.body)
	}
}

func TestLoginIsCaseInsensitiveOnUsername(t *testing.T) {
	h := newHarness(t)
	username := h.createUser(domain.RoleFIAAdmin, 0)

	expectStatus(t, h.login(h.browser(), "  "+strings.ToUpper(username)+" ", testPassword), http.StatusOK)
}

func TestNativeLoginReturnsBearerToken(t *testing.T) {
	h := newHarness(t)
	username := h.createUser(domain.RoleFIAAdmin, 0)
	client := &http.Client{}

	login := h.do(client, http.MethodPost, "/auth/login",
		map[string]string{"username": username, "password": testPassword},
		map[string]string{"X-Client-Platform": "native"})
	expectStatus(t, login, http.StatusOK)

	token, _ := login.body["token"].(string)
	if token == "" {
		t.Fatal("native login must return the session token")
	}

	me := h.do(client, http.MethodGet, "/auth/me", nil, map[string]string{"Authorization": "Bearer " + token})
	expectStatus(t, me, http.StatusOK)
	if me.body["role"] != string(domain.RoleFIAAdmin) {
		t.Fatalf("role = %v, want fia_admin", me.body["role"])
	}
}

func TestFailedLoginDoesNotRevealWhetherUserExists(t *testing.T) {
	h := newHarness(t)
	username := h.createUser(domain.RoleFIAAdmin, 0)

	wrongPasswordResponse := h.login(h.browser(), username, wrongPassword)
	unknownUserResponse := h.login(h.browser(), testdb.UniqueUsername(t), wrongPassword)

	expectStatus(t, wrongPasswordResponse, http.StatusUnauthorized)
	expectStatus(t, unknownUserResponse, http.StatusUnauthorized)
	if wrongPasswordResponse.raw != unknownUserResponse.raw {
		t.Fatalf("responses differ:\n wrong password: %s\n unknown user:   %s",
			wrongPasswordResponse.raw, unknownUserResponse.raw)
	}
}

func TestInvalidInputIsRejectedBeforeAuthenticating(t *testing.T) {
	h := newHarness(t)

	cases := map[string]map[string]string{
		"forbidden characters": {"username": "fia admin!", "password": testPassword},
		"empty password":       {"username": "fia.admin", "password": ""},
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			got := h.do(h.browser(), http.MethodPost, "/auth/login", payload, nil)
			expectStatus(t, got, http.StatusBadRequest)
			if got.body["code"] != "validation" {
				t.Fatalf("code = %v, want validation", got.body["code"])
			}
		})
	}
}

func TestDeactivatedUserCannotLogIn(t *testing.T) {
	h := newHarness(t)
	username := h.createUser(domain.RoleFIAAdmin, 0)
	h.deactivate(username)

	expectStatus(t, h.login(h.browser(), username, testPassword), http.StatusUnauthorized)
}

func TestDeactivationEndsExistingSessionImmediately(t *testing.T) {
	h := newHarness(t)
	username := h.createUser(domain.RoleFIAAdmin, 0)
	browser := h.browser()
	expectStatus(t, h.login(browser, username, testPassword), http.StatusOK)

	h.deactivate(username)

	expectStatus(t, h.do(browser, http.MethodGet, "/auth/me", nil, nil), http.StatusUnauthorized)
}

func TestAccountLocksAfterRepeatedFailures(t *testing.T) {
	h := newHarness(t)
	username := h.createUser(domain.RoleFIAAdmin, 0)

	for attempt := 1; attempt <= auth.MaxConsecutiveFailures; attempt++ {
		expectStatus(t, h.login(h.browser(), username, wrongPassword), http.StatusUnauthorized)
	}

	locked := h.login(h.browser(), username, testPassword)
	expectStatus(t, locked, http.StatusTooManyRequests)
	if locked.body["code"] != "account_locked" {
		t.Fatalf("code = %v, want account_locked", locked.body["code"])
	}
}

func TestSuccessfulLoginResetsFailureCount(t *testing.T) {
	h := newHarness(t)
	username := h.createUser(domain.RoleFIAAdmin, 0)

	for attempt := 1; attempt < auth.MaxConsecutiveFailures; attempt++ {
		expectStatus(t, h.login(h.browser(), username, wrongPassword), http.StatusUnauthorized)
	}
	expectStatus(t, h.login(h.browser(), username, testPassword), http.StatusOK)
	expectStatus(t, h.login(h.browser(), username, wrongPassword), http.StatusUnauthorized)

	expectStatus(t, h.login(h.browser(), username, testPassword), http.StatusOK)
}

func TestLogoutRevokesSession(t *testing.T) {
	h := newHarness(t)
	username := h.createUser(domain.RoleFIAAdmin, 0)
	browser := h.browser()
	expectStatus(t, h.login(browser, username, testPassword), http.StatusOK)

	expectStatus(t, h.do(browser, http.MethodPost, "/auth/logout", nil, nil), http.StatusNoContent)
	expectStatus(t, h.do(browser, http.MethodGet, "/auth/me", nil, nil), http.StatusUnauthorized)
}

func TestMeWithoutSessionIsUnauthorized(t *testing.T) {
	h := newHarness(t)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/auth/me", nil, nil), http.StatusUnauthorized)
}
