package transport_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/testdb"
)

const (
	ferrariTeamID = 1
	unknownTeamID = 999_999_999
	newPassword   = "otra frase larga y nueva"
)

func (h *harness) signedIn(role domain.Role, teamID int64) (*http.Client, string) {
	h.t.Helper()

	username := h.createUser(role, teamID)
	client := h.browser()
	expectStatus(h.t, h.login(client, username, testPassword), http.StatusOK)
	return client, username
}

func (h *harness) createAccount(admin *http.Client, payload map[string]any) response {
	h.t.Helper()
	return h.do(admin, http.MethodPost, "/users", payload, nil)
}

func (h *harness) newTeamAdminPayload() map[string]any {
	return map[string]any{
		"username": testdb.UniqueUsername(h.t),
		"password": testPassword,
		"role":     "team_admin",
		"teamId":   ferrariTeamID,
	}
}

func accountPath(account response, action string) string {
	return fmt.Sprintf("/users/%.0f%s", account.body["id"], action)
}

func TestUserManagementRequiresFIAAdmin(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/users", nil, nil), http.StatusUnauthorized)
	expectStatus(t, h.do(teamAdmin, http.MethodGet, "/users", nil, nil), http.StatusForbidden)
	expectStatus(t, h.createAccount(teamAdmin, h.newTeamAdminPayload()), http.StatusForbidden)
}

func TestFIAAdminCreatesTeamAdminWhoCanLogIn(t *testing.T) {
	h := newHarness(t)
	admin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	payload := h.newTeamAdminPayload()
	payload["email"] = "  Nuevo." + payload["username"].(string) + "@Example.com "

	created := h.createAccount(admin, payload)
	expectStatus(t, created, http.StatusCreated)
	if created.body["teamName"] != "Scuderia Ferrari" || created.body["isActive"] != true {
		t.Fatalf("created = %v, want an active Ferrari team admin", created.body)
	}
	if _, leaked := created.body["password"]; leaked {
		t.Fatal("the account response must never include the password")
	}
	if created.body["email"] != "nuevo."+payload["username"].(string)+"@example.com" {
		t.Fatalf("email = %v, want it trimmed and lowercased", created.body["email"])
	}

	expectStatus(t, h.login(h.browser(), payload["username"].(string), testPassword), http.StatusOK)
}

func TestCreateRejectsInvalidAccounts(t *testing.T) {
	h := newHarness(t)
	admin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	cases := map[string]struct {
		change    func(map[string]any)
		wantField string
	}{
		"short password":          {func(p map[string]any) { p["password"] = "corta" }, "password"},
		"team admin without team": {func(p map[string]any) { delete(p, "teamId") }, "teamId"},
		"fia admin with team":     {func(p map[string]any) { p["role"] = "fia_admin" }, "teamId"},
		"unknown team":            {func(p map[string]any) { p["teamId"] = unknownTeamID }, "teamId"},
		"invented role":           {func(p map[string]any) { p["role"] = "public" }, "role"},
		"malformed email":         {func(p map[string]any) { p["email"] = "no-es-un-email" }, "email"},
		"malformed username":      {func(p map[string]any) { p["username"] = "con espacios" }, "username"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			payload := h.newTeamAdminPayload()
			tc.change(payload)

			got := h.createAccount(admin, payload)
			expectStatus(t, got, http.StatusBadRequest)
			if got.body["field"] != tc.wantField {
				t.Fatalf("field = %v, want %s", got.body["field"], tc.wantField)
			}
		})
	}
}

func TestCreateRejectsTakenUsernameAndEmail(t *testing.T) {
	h := newHarness(t)
	admin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	first := h.newTeamAdminPayload()
	first["email"] = first["username"].(string) + "@example.com"
	expectStatus(t, h.createAccount(admin, first), http.StatusCreated)

	sameUsername := h.newTeamAdminPayload()
	sameUsername["username"] = first["username"]
	sameEmail := h.newTeamAdminPayload()
	sameEmail["email"] = first["email"]

	for field, payload := range map[string]map[string]any{"username": sameUsername, "email": sameEmail} {
		got := h.createAccount(admin, payload)
		expectStatus(t, got, http.StatusConflict)
		if got.body["field"] != field {
			t.Fatalf("field = %v, want %s", got.body["field"], field)
		}
	}
}

func TestListSearchesAndFilters(t *testing.T) {
	h := newHarness(t)
	admin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	kept := h.createAccount(admin, h.newTeamAdminPayload())
	deactivated := h.createAccount(admin, h.newTeamAdminPayload())
	expectStatus(t, h.do(admin, http.MethodPost, accountPath(deactivated, "/deactivate"), nil, nil), http.StatusOK)

	countMatches := func(query string) int {
		t.Helper()
		got := h.do(admin, http.MethodGet, "/users?"+query, nil, nil)
		expectStatus(t, got, http.StatusOK)
		return len(got.body["users"].([]any))
	}

	keptName := kept.body["username"].(string)
	if countMatches("q="+keptName) != 1 {
		t.Fatalf("searching %q should find exactly that account", keptName)
	}
	if countMatches("q="+keptName+"&role=fia_admin") != 0 {
		t.Fatal("the role filter should exclude a team admin")
	}
	if countMatches("q="+deactivated.body["username"].(string)+"&status=active") != 0 {
		t.Fatal("the status filter should exclude a deactivated account")
	}
	if countMatches("q=%25") != 0 {
		t.Fatal("a literal % must not act as a wildcard")
	}
	expectStatus(t, h.do(admin, http.MethodGet, "/users?status=asleep", nil, nil), http.StatusBadRequest)
}

func TestRoleChangeClearsTeamAndRevokesSessions(t *testing.T) {
	h := newHarness(t)
	admin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	teamAdmin, username := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)
	target := h.findAccount(admin, username)

	changed := h.do(admin, http.MethodPatch, accountPath(target, ""), map[string]any{"role": "fia_admin"}, nil)
	expectStatus(t, changed, http.StatusOK)
	if changed.body["role"] != "fia_admin" || changed.body["teamId"] != nil {
		t.Fatalf("changed = %v, want an FIA admin without team", changed.body)
	}

	expectStatus(t, h.do(teamAdmin, http.MethodGet, "/auth/me", nil, nil), http.StatusUnauthorized)
}

func TestEmailChangeKeepsSessions(t *testing.T) {
	h := newHarness(t)
	admin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	teamAdmin, username := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)
	target := h.findAccount(admin, username)

	changed := h.do(admin, http.MethodPatch, accountPath(target, ""), map[string]any{"email": username + "@example.com"}, nil)
	expectStatus(t, changed, http.StatusOK)

	expectStatus(t, h.do(teamAdmin, http.MethodGet, "/auth/me", nil, nil), http.StatusOK)
}

func TestPasswordChangeRevokesSessions(t *testing.T) {
	h := newHarness(t)
	admin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	teamAdmin, username := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)
	target := h.findAccount(admin, username)

	expectStatus(t, h.do(admin, http.MethodPatch, accountPath(target, ""), map[string]any{"password": newPassword}, nil), http.StatusOK)

	expectStatus(t, h.do(teamAdmin, http.MethodGet, "/auth/me", nil, nil), http.StatusUnauthorized)
	expectStatus(t, h.login(h.browser(), username, testPassword), http.StatusUnauthorized)
	expectStatus(t, h.login(h.browser(), username, newPassword), http.StatusOK)
}

func TestDeactivationEndsSessionsAndReactivationRestoresAccess(t *testing.T) {
	h := newHarness(t)
	admin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	teamAdmin, username := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)
	secondDevice := h.browser()
	expectStatus(t, h.login(secondDevice, username, testPassword), http.StatusOK)
	target := h.findAccount(admin, username)

	deactivated := h.do(admin, http.MethodPost, accountPath(target, "/deactivate"), nil, nil)
	expectStatus(t, deactivated, http.StatusOK)
	if deactivated.body["isActive"] != false || deactivated.body["deactivatedAt"] == nil {
		t.Fatalf("deactivated = %v, want an inactive account with its date", deactivated.body)
	}
	for _, device := range []*http.Client{teamAdmin, secondDevice} {
		expectStatus(t, h.do(device, http.MethodGet, "/auth/me", nil, nil), http.StatusUnauthorized)
	}
	expectStatus(t, h.login(h.browser(), username, testPassword), http.StatusUnauthorized)

	expectStatus(t, h.do(admin, http.MethodPost, accountPath(target, "/reactivate"), nil, nil), http.StatusOK)
	expectStatus(t, h.login(h.browser(), username, testPassword), http.StatusOK)
}

func TestAdminCannotLockThemselvesOut(t *testing.T) {
	h := newHarness(t)
	admin, username := h.signedIn(domain.RoleFIAAdmin, 0)
	self := h.findAccount(admin, username)

	for _, attempt := range []struct {
		method, path string
		payload      any
	}{
		{http.MethodPost, accountPath(self, "/deactivate"), nil},
		{http.MethodPatch, accountPath(self, ""), map[string]any{"role": "team_admin", "teamId": ferrariTeamID}},
	} {
		got := h.do(admin, attempt.method, attempt.path, attempt.payload, nil)
		expectStatus(t, got, http.StatusConflict)
		if got.body["code"] != "self_lockout" {
			t.Fatalf("code = %v, want self_lockout", got.body["code"])
		}
	}
}

func TestUnknownAccountIsNotFound(t *testing.T) {
	h := newHarness(t)
	admin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	expectStatus(t, h.do(admin, http.MethodGet, "/users/999999999", nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(admin, http.MethodGet, "/users/not-a-number", nil, nil), http.StatusNotFound)
}

func TestTeamsCatalogIsPublic(t *testing.T) {
	h := newHarness(t)

	got := h.do(h.browser(), http.MethodGet, "/teams", nil, nil)
	expectStatus(t, got, http.StatusOK)
	first := got.body["teams"].([]any)[0].(map[string]any)
	if first["categoryCode"] != "f1" {
		t.Fatalf("first team = %v, want the catalogue ordered by category", first)
	}
}

func (h *harness) findAccount(admin *http.Client, username string) response {
	h.t.Helper()

	got := h.do(admin, http.MethodGet, "/users?q="+username, nil, nil)
	expectStatus(h.t, got, http.StatusOK)
	matches := got.body["users"].([]any)
	if len(matches) != 1 {
		h.t.Fatalf("expected exactly one account named %s, got %d", username, len(matches))
	}
	return response{status: got.status, body: matches[0].(map[string]any)}
}
