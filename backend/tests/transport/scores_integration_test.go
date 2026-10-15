package transporttest

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/racecontrol/backend/internal/domain"
)

func (h *harness) createScoreRace(fiaAdmin *http.Client, gpName string) int64 {
	h.t.Helper()

	number := h.nextRoundNumber(formulaOneCategory)
	created := h.createRace(fiaAdmin, time.Date(2026, 6, 7, 15, 0, 0, 0, time.UTC), number, gpName)
	expectStatus(h.t, created, http.StatusCreated)
	return eventID(h.t, created)
}

func (h *harness) scoreGrid(client *http.Client, eventID int64) response {
	h.t.Helper()

	got := h.do(client, http.MethodGet, fmt.Sprintf("/events/%d/scores", eventID), nil, nil)
	expectStatus(h.t, got, http.StatusOK)
	return got
}

func scoreEntries(t *testing.T, got response) []map[string]any {
	t.Helper()

	raw, _ := got.body["entries"].([]any)
	entries := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if entry, ok := item.(map[string]any); ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

func pointsOf(entries []map[string]any, driverID float64) any {
	for _, entry := range entries {
		if entry["driverId"] == driverID {
			return entry["points"]
		}
	}
	return nil
}

func (h *harness) loadScores(client *http.Client, eventID int64, entries []map[string]any) response {
	h.t.Helper()
	return h.do(client, http.MethodPut, fmt.Sprintf("/events/%d/scores", eventID), map[string]any{"entries": entries}, nil)
}

func TestScoreManagementRequiresFIAAdmin(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/events/1/scores", nil, nil), http.StatusUnauthorized)
	expectStatus(t, h.do(teamAdmin, http.MethodGet, "/events/1/scores", nil, nil), http.StatusForbidden)
	expectStatus(t, h.loadScores(teamAdmin, 1, nil), http.StatusForbidden)
}

func TestFIAAdminLoadsAndEditsScores(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	id := h.createScoreRace(fiaAdmin, "Gran Premio de Puntajes")

	grid := h.scoreGrid(fiaAdmin, id)
	if grid.body["categoryCode"] != formulaOneCategory {
		t.Fatalf("grid championship = %v, want f1", grid.body["categoryCode"])
	}
	if len(scoreEntries(t, grid)) == 0 {
		t.Fatal("the grid must list the drivers of the championship")
	}

	loaded := h.loadScores(fiaAdmin, id, []map[string]any{
		{"driverId": 1, "points": 25},
		{"driverId": 2, "points": 18},
		{"driverId": 3, "points": 15},
	})
	expectStatus(t, loaded, http.StatusOK)
	if pointsOf(scoreEntries(t, loaded), 1) != float64(25) {
		t.Fatalf("driver 1 points = %v, want 25", pointsOf(scoreEntries(t, loaded), 1))
	}

	fetched := h.scoreGrid(fiaAdmin, id)
	if pointsOf(scoreEntries(t, fetched), 3) != float64(15) {
		t.Fatalf("the loaded scores must persist, got %v", scoreEntries(t, fetched))
	}

	removed := h.loadScores(fiaAdmin, id, []map[string]any{
		{"driverId": 1, "points": 25},
		{"driverId": 2, "points": 18},
	})
	expectStatus(t, removed, http.StatusOK)
	if pointsOf(scoreEntries(t, removed), 3) != nil {
		t.Fatal("a driver left out of the grid must lose their score")
	}
}

func TestScoresRejectBadInput(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	id := h.createScoreRace(fiaAdmin, "Gran Premio Inválido")

	cases := map[string][]map[string]any{
		"negative points": {{"driverId": 1, "points": -5}},
		"empty points":    {{"driverId": 1}},
		"repeated driver": {{"driverId": 1, "points": 10}, {"driverId": 1, "points": 20}},
		"unknown driver":  {{"driverId": 999999999, "points": 10}},
	}
	for name, entries := range cases {
		t.Run(name, func(t *testing.T) {
			got := h.loadScores(fiaAdmin, id, entries)
			expectStatus(t, got, http.StatusBadRequest)
			if got.body["code"] != "validation" {
				t.Fatalf("code = %v, want validation; body: %s", got.body["code"], got.raw)
			}
		})
	}
}

func TestScoresOnANonRaceAreNotFound(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	// Seeded event 4 is a tyre test: only races take results.
	expectStatus(t, h.do(fiaAdmin, http.MethodGet, "/events/4/scores", nil, nil), http.StatusNotFound)
	expectStatus(t, h.loadScores(fiaAdmin, 4, []map[string]any{{"driverId": 1, "points": 10}}), http.StatusNotFound)
}

func TestScoreChangeIsAuditedWithItsAuthor(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, username := h.signedIn(domain.RoleFIAAdmin, 0)
	id := h.createScoreRace(fiaAdmin, "Gran Premio Auditado")

	expectStatus(t, h.loadScores(fiaAdmin, id, []map[string]any{{"driverId": 1, "points": 10}}), http.StatusOK)

	var actorID int64
	if err := h.pool.QueryRow(context.Background(), `SELECT id FROM users WHERE username = $1`, username).Scan(&actorID); err != nil {
		t.Fatalf("load the acting user: %v", err)
	}

	var loggedActor int64
	err := h.pool.QueryRow(context.Background(),
		`SELECT actor_user_id FROM score_change_log WHERE event_id = $1 ORDER BY id DESC LIMIT 1`, id).Scan(&loggedActor)
	if err != nil {
		t.Fatalf("load the score change log: %v", err)
	}
	if loggedActor != actorID {
		t.Fatalf("logged actor = %d, want %d", loggedActor, actorID)
	}

	var items int
	if err := h.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM score_change_log_items i
		 JOIN score_change_log l ON l.id = i.change_log_id
		 WHERE l.event_id = $1`, id).Scan(&items); err != nil {
		t.Fatalf("count the score change items: %v", err)
	}
	if items != 1 {
		t.Fatalf("items = %d, want the one driver of the edit", items)
	}
}

func TestStandingsArePublicAndReflectAChange(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	id := h.createScoreRace(fiaAdmin, "Gran Premio de Clasificación")

	before := h.driverStandingPoints(1, formulaOneCategory)
	expectStatus(t, h.loadScores(fiaAdmin, id, []map[string]any{{"driverId": 1, "points": 7}}), http.StatusOK)

	if gained := h.driverStandingPoints(1, formulaOneCategory) - before; gained != 7 {
		t.Fatalf("standings moved by %v, want 7", gained)
	}

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/standings?category=f1&season=2026", nil, nil), http.StatusOK)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/standings", nil, nil), http.StatusBadRequest)
}

func (h *harness) driverStandingPoints(driverID float64, category string) float64 {
	h.t.Helper()

	got := h.do(h.browser(), http.MethodGet, "/standings?category="+category+"&season=2026", nil, nil)
	expectStatus(h.t, got, http.StatusOK)
	raw, _ := got.body["standings"].([]any)
	for _, item := range raw {
		standing, ok := item.(map[string]any)
		if !ok || standing["driverId"] != driverID {
			continue
		}
		points, _ := standing["points"].(float64)
		return points
	}
	return 0
}
