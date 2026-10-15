package transporttest

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/tests/testdb"
)

func (h *harness) driverDirectory(query string) []map[string]any {
	h.t.Helper()

	got := h.do(h.browser(), http.MethodGet, "/drivers"+query, nil, nil)
	expectStatus(h.t, got, http.StatusOK)
	raw, _ := got.body["drivers"].([]any)
	drivers := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if driver, ok := item.(map[string]any); ok {
			drivers = append(drivers, driver)
		}
	}
	return drivers
}

func TestDriverDirectoryIsPublic(t *testing.T) {
	h := newHarness(t)

	got := h.do(h.browser(), http.MethodGet, "/drivers", nil, nil)
	expectStatus(t, got, http.StatusOK)

	detail := h.do(h.browser(), http.MethodGet, "/drivers/1", nil, nil)
	expectStatus(t, detail, http.StatusOK)
	if detail.body["teamName"] != "Scuderia Ferrari" || detail.body["categoryCode"] != "f1" {
		t.Fatalf("seeded driver 1 = %v, want Charles Leclerc at Ferrari in f1", detail.body)
	}
	if detail.body["firstName"] != "Charles" || detail.body["lastName"] != "Leclerc" {
		t.Fatalf("detail person facts = %v", detail.body)
	}
}

func TestDirectoryListsEveryTeamCurrentLineup(t *testing.T) {
	h := newHarness(t)

	lineup := h.driverDirectory("?category=f1")
	if len(lineup) == 0 {
		t.Fatal("the seeded Formula 1 lineup should not be empty")
	}
	for _, driver := range lineup {
		if driver["teamName"] == nil || driver["categoryCode"] != "f1" {
			t.Fatalf("a lineup row must carry its team and championship: %v", driver)
		}
	}
}

func TestDirectoryReflectsALineupChange(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	name := testdb.UniqueCode(t)
	payload := h.newDriverPayload()
	payload["lastName"] = name
	created := h.createTeamDriver(teamAdmin, payload)
	expectStatus(t, created, http.StatusCreated)

	query := "?q=" + url.QueryEscape(name)
	found := h.driverDirectory(query)
	if len(found) != 1 || found[0]["lastName"] != name {
		t.Fatalf("the new driver should be in the directory, got %v", found)
	}
	if found[0]["teamName"] != "Scuderia Ferrari" {
		t.Fatalf("the directory must credit the team, got %v", found[0]["teamName"])
	}

	expectStatus(t, h.do(teamAdmin, http.MethodPost, driverPath(created, "/release"), nil, nil), http.StatusNoContent)
	if len(h.driverDirectory(query)) != 0 {
		t.Fatal("a released driver must leave the directory")
	}
}

func TestDirectorySearchAndFilter(t *testing.T) {
	h := newHarness(t)

	if len(h.driverDirectory("?q=Leclerc")) == 0 {
		t.Fatal("searching a surname should find the driver")
	}
	for _, driver := range h.driverDirectory("?q=%25") {
		if driver["lastName"] != nil && driver["lastName"] == "%" {
			continue
		}
		t.Fatal("a literal % must not act as a wildcard")
	}
	for _, driver := range h.driverDirectory("?category=academy") {
		if driver["categoryCode"] != "academy" {
			t.Fatalf("the category filter leaked %v", driver["categoryCode"])
		}
	}
}

func TestDirectoryDetailOfAnUnknownIdIsNotFound(t *testing.T) {
	h := newHarness(t)

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/drivers/999999999", nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/drivers/not-a-number", nil, nil), http.StatusNotFound)
}

func TestDirectoryIsReadOnly(t *testing.T) {
	h := newHarness(t)

	expectStatus(t, h.do(h.browser(), http.MethodPost, "/drivers", map[string]any{"firstName": "X"}, nil), http.StatusMethodNotAllowed)
	expectStatus(t, h.do(h.browser(), http.MethodPatch, "/drivers/1", map[string]any{"lastName": "X"}, nil), http.StatusMethodNotAllowed)
}
