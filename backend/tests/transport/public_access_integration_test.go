package transporttest

import (
	"io"
	"net/http"
	"testing"

	"github.com/racecontrol/backend/internal/domain"
)

// publicReads are the reads an anonymous visitor has. A signed-in user must see
// the very same, so none of them may depend on the session.
var publicReads = []string{
	"/teams",
	"/circuits",
	"/seasons",
	"/events",
	"/events/1",
	"/drivers",
	"/drivers/1",
	"/results",
	"/results/1",
	"/standings?category=f1",
}

// publicDownloads stamp the moment they were generated, so only their status is
// compared.
var publicDownloads = []string{"/calendar.ics", "/calendar.pdf"}

type viewer struct {
	name    string
	client  *http.Client
	headers map[string]string
}

func (h *harness) publicViewers() []viewer {
	h.t.Helper()

	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)
	return []viewer{
		{name: "fia admin", client: fiaAdmin},
		{name: "team admin", client: teamAdmin},
		{name: "stale token", client: h.browser(), headers: map[string]string{"Authorization": "Bearer not-a-session"}},
	}
}

func (h *harness) fetch(who viewer, path string) (int, string) {
	h.t.Helper()

	request, err := http.NewRequest(http.MethodGet, h.server.URL+path, nil)
	if err != nil {
		h.t.Fatalf("new request: %v", err)
	}
	for name, value := range who.headers {
		request.Header.Set(name, value)
	}
	result, err := who.client.Do(request)
	if err != nil {
		h.t.Fatalf("GET %s as %s: %v", path, who.name, err)
	}
	defer result.Body.Close()
	raw, err := io.ReadAll(result.Body)
	if err != nil {
		h.t.Fatalf("read %s: %v", path, err)
	}
	return result.StatusCode, string(raw)
}

func TestSignedInUsersSeeEverythingAnAnonymousVisitorSees(t *testing.T) {
	h := newHarness(t)
	anonymous := viewer{name: "anonymous", client: h.browser()}
	viewers := h.publicViewers()

	for _, path := range publicReads {
		wantStatus, wantBody := h.fetch(anonymous, path)
		if wantStatus != http.StatusOK {
			t.Fatalf("anonymous GET %s = %d, want 200", path, wantStatus)
		}
		for _, who := range viewers {
			status, body := h.fetch(who, path)
			if status != wantStatus || body != wantBody {
				t.Errorf("GET %s as %s = %d, differs from what an anonymous visitor gets", path, who.name, status)
			}
		}
	}
	for _, path := range publicDownloads {
		for _, who := range append(viewers, anonymous) {
			if status, _ := h.fetch(who, path); status != http.StatusOK {
				t.Errorf("GET %s as %s = %d, want 200", path, who.name, status)
			}
		}
	}
}

// privateReads need a session; an anonymous visitor never gets past the guard.
var privateReads = []string{
	"/users",
	"/team/drivers",
	"/team/score-notices",
	"/score-changes",
	"/events/1/scores",
	"/events/1/scores/history",
	"/conversations",
	"/auth/me",
}

func TestAnonymousVisitorsSeeNothingPrivate(t *testing.T) {
	h := newHarness(t)
	anonymous := viewer{name: "anonymous", client: h.browser()}

	for _, path := range privateReads {
		if status, _ := h.fetch(anonymous, path); status != http.StatusUnauthorized {
			t.Errorf("anonymous GET %s = %d, want 401", path, status)
		}
	}
}
