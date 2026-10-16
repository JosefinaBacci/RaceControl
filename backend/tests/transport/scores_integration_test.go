package transporttest

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/tests/testdb"
)

var (
	firstRaceStart  = time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC)
	secondRaceStart = time.Date(2026, 4, 1, 14, 0, 0, 0, time.UTC)
)

// standingsFixture is a championship of its own with two teams and two races:
//
//	race 1: Alonso (A) 25, Bianchi (A) 18, Castro (B) 18
//	race 2: Castro (B) 25, Alonso (A) 18, Bianchi (A) 0.5
//
// so Alonso and Castro tie on 43, Bianchi has 18.5, team A 61.5 and team B 43.
type standingsFixture struct {
	championship *testdb.Championship
	firstRace    int64
	secondRace   int64
	teamA        int64
	teamB        int64
}

func (h *harness) newStandingsFixture() standingsFixture {
	h.t.Helper()

	championship := testdb.NewChampionship(h.t, h.pool)
	teamA := championship.AddTeam("Equipo A")
	teamB := championship.AddTeam("Equipo B")
	alonso := championship.AddDriver(teamA, "Alonso")
	bianchi := championship.AddDriver(teamA, "Bianchi")
	castro := championship.AddDriver(teamB, "Castro")

	firstRace := championship.AddRace(firstRaceStart)
	secondRace := championship.AddRace(secondRaceStart)
	championship.Score(firstRace, alonso, 25)
	championship.Score(firstRace, bianchi, 18)
	championship.Score(firstRace, castro, 18)
	championship.Score(secondRace, castro, 25)
	championship.Score(secondRace, alonso, 18)
	championship.Score(secondRace, bianchi, 0.5)

	return standingsFixture{championship: championship, firstRace: firstRace, secondRace: secondRace, teamA: teamA, teamB: teamB}
}

func listOf(t *testing.T, body map[string]any, key string) []map[string]any {
	t.Helper()

	raw, ok := body[key].([]any)
	if !ok {
		t.Fatalf("%q is not a list in %v", key, body)
	}
	list := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		list = append(list, item.(map[string]any))
	}
	return list
}

func objectOf(t *testing.T, body map[string]any, key string) map[string]any {
	t.Helper()

	object, ok := body[key].(map[string]any)
	if !ok {
		t.Fatalf("%q is not an object in %v", key, body)
	}
	return object
}

type rankedLine struct {
	position float64
	name     string
	points   float64
}

func expectRanking(t *testing.T, rows []map[string]any, nameKey string, want []rankedLine) {
	t.Helper()

	if len(rows) != len(want) {
		t.Fatalf("got %d rows, want %d: %v", len(rows), len(want), rows)
	}
	for index, line := range want {
		row := rows[index]
		if row["position"] != line.position || row[nameKey] != line.name || row["points"] != line.points {
			t.Errorf("row %d = (%v, %v, %v), want (%v, %s, %v)",
				index, row["position"], row[nameKey], row["points"], line.position, line.name, line.points)
		}
	}
}

func TestResultsAndStandingsAreReadableWithoutASession(t *testing.T) {
	h := newHarness(t)

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/results", nil, nil), http.StatusOK)
	expectStatus(t, h.do(h.browser(), http.MethodGet, fmt.Sprintf("/results/%d", seededRacedEventID), nil, nil), http.StatusOK)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/standings?category=f1", nil, nil), http.StatusOK)
}

func TestStandingsShareTiedPositionsAndSumTheEntries(t *testing.T) {
	h := newHarness(t)
	fixture := h.newStandingsFixture()

	got := h.do(h.browser(), http.MethodGet, "/standings?season=2026&category="+fixture.championship.CategoryCode, nil, nil)
	expectStatus(t, got, http.StatusOK)

	expectRanking(t, listOf(t, got.body, "drivers"), "lastName", []rankedLine{
		{1, "Alonso", 43},
		{1, "Castro", 43},
		{3, "Bianchi", 18.5},
	})
	expectRanking(t, listOf(t, got.body, "teams"), "teamName", []rankedLine{
		{1, "Equipo A", 61.5},
		{2, "Equipo B", 43},
	})
}

func TestStandingsRequireACategoryAndAValidSeason(t *testing.T) {
	h := newHarness(t)

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/standings", nil, nil), http.StatusBadRequest)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/standings?category=f1&season=dos-mil", nil, nil), http.StatusBadRequest)
}

func TestResultsListEachRaceWithItsPositions(t *testing.T) {
	h := newHarness(t)
	fixture := h.newStandingsFixture()

	got := h.do(h.browser(), http.MethodGet, "/results?category="+fixture.championship.CategoryCode, nil, nil)
	expectStatus(t, got, http.StatusOK)

	races := listOf(t, got.body, "races")
	if len(races) != 2 {
		t.Fatalf("got %d races, want the two raced ones: %s", len(races), got.raw)
	}
	if objectOf(t, races[0], "race")["id"] != float64(fixture.firstRace) {
		t.Fatalf("races must come in calendar order, got %s", got.raw)
	}
	expectRanking(t, listOf(t, races[0], "results"), "lastName", []rankedLine{
		{1, "Alonso", 25},
		{2, "Bianchi", 18},
		{2, "Castro", 18},
	})
}

func TestRaceResultsMatchTheLoadedScores(t *testing.T) {
	h := newHarness(t)
	fixture := h.newStandingsFixture()

	got := h.do(h.browser(), http.MethodGet, fmt.Sprintf("/results/%d", fixture.secondRace), nil, nil)
	expectStatus(t, got, http.StatusOK)

	race := objectOf(t, got.body, "race")
	if race["categoryCode"] != fixture.championship.CategoryCode || race["name"] == "" {
		t.Fatalf("race = %v, want a labelled race of the fixture championship", race)
	}
	expectRanking(t, listOf(t, got.body, "results"), "lastName", []rankedLine{
		{1, "Castro", 25},
		{2, "Alonso", 18},
		{3, "Bianchi", 0.5},
	})
}

func TestRaceWithoutResultsHasAnEmptyTable(t *testing.T) {
	h := newHarness(t)
	championship := testdb.NewChampionship(t, h.pool)
	race := championship.AddRace(firstRaceStart)

	got := h.do(h.browser(), http.MethodGet, fmt.Sprintf("/results/%d", race), nil, nil)
	expectStatus(t, got, http.StatusOK)
	if results := listOf(t, got.body, "results"); len(results) != 0 {
		t.Fatalf("results = %v, want none", results)
	}
}

func TestResultsOfSomethingOtherThanARaceAreNotFound(t *testing.T) {
	h := newHarness(t)

	const seededTechnicalControlID = 2
	expectStatus(t, h.do(h.browser(), http.MethodGet, fmt.Sprintf("/results/%d", seededTechnicalControlID), nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/results/999999999", nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/results/not-a-number", nil, nil), http.StatusNotFound)
}

// TestStandingsCreditTheTeamAtRaceTime covers a mid-season transfer: the team
// table keeps each result with the team the driver raced for, while the driver
// row shows the team of their latest race.
func TestStandingsCreditTheTeamAtRaceTime(t *testing.T) {
	h := newHarness(t)
	championship := testdb.NewChampionship(t, h.pool)
	teamA := championship.AddTeam("Equipo A")
	teamB := championship.AddTeam("Equipo B")
	driver := championship.AddDriver(teamA, "Alonso")

	firstRace := championship.AddRace(firstRaceStart)
	secondRace := championship.AddRace(secondRaceStart)
	championship.Score(firstRace, driver, 25)
	championship.Transfer(driver, teamB, firstRaceStart.AddDate(0, 0, 14))
	championship.Score(secondRace, driver, 10)

	got := h.do(h.browser(), http.MethodGet, "/standings?category="+championship.CategoryCode, nil, nil)
	expectStatus(t, got, http.StatusOK)

	expectRanking(t, listOf(t, got.body, "teams"), "teamName", []rankedLine{
		{1, "Equipo A", 25},
		{2, "Equipo B", 10},
	})
	drivers := listOf(t, got.body, "drivers")
	expectRanking(t, drivers, "lastName", []rankedLine{{1, "Alonso", 35}})
	if drivers[0]["teamId"] != float64(teamB) {
		t.Fatalf("driver row team = %v, want the team of the latest race %d", drivers[0]["teamId"], teamB)
	}
}

// TestScoresStayInsideTheirChampionship: another category and another season
// never leak into the results or the standings asked for.
func TestScoresStayInsideTheirChampionship(t *testing.T) {
	h := newHarness(t)
	const previousYear = 2025
	previousSeason := testdb.Season(t, h.pool, previousYear)

	asked := testdb.NewChampionship(t, h.pool)
	askedDriver := asked.AddDriver(asked.AddTeam("Equipo A"), "Alonso")
	askedRace := asked.AddRace(firstRaceStart)
	asked.Score(askedRace, askedDriver, 25)

	other := testdb.NewChampionship(t, h.pool)
	otherDriver := other.AddDriver(other.AddTeam("Equipo B"), "Castro")
	otherRace := other.AddRace(firstRaceStart)
	other.Score(otherRace, otherDriver, 18)

	previousSeasonStart := time.Date(previousYear, 1, 1, 0, 0, 0, 0, time.UTC)
	pastDriver := asked.AddDriverSince(asked.AddTeam("Equipo C"), "Bianchi", previousSeasonStart)
	pastRace := asked.AddRaceIn(previousSeason, previousSeasonStart.AddDate(0, 2, 0))
	asked.Score(pastRace, pastDriver, 12)

	standings := h.do(h.browser(), http.MethodGet, "/standings?season=2026&category="+asked.CategoryCode, nil, nil)
	expectStatus(t, standings, http.StatusOK)
	expectRanking(t, listOf(t, standings.body, "drivers"), "lastName", []rankedLine{{1, "Alonso", 25}})
	expectRanking(t, listOf(t, standings.body, "teams"), "teamName", []rankedLine{{1, "Equipo A", 25}})

	past := h.do(h.browser(), http.MethodGet, fmt.Sprintf("/standings?season=%d&category=%s", previousYear, asked.CategoryCode), nil, nil)
	expectStatus(t, past, http.StatusOK)
	expectRanking(t, listOf(t, past.body, "drivers"), "lastName", []rankedLine{{1, "Bianchi", 12}})

	results := h.do(h.browser(), http.MethodGet, "/results?season=2026&category="+asked.CategoryCode, nil, nil)
	expectStatus(t, results, http.StatusOK)
	races := listOf(t, results.body, "races")
	if len(races) != 1 || objectOf(t, races[0], "race")["id"] != float64(askedRace) {
		t.Fatalf("races = %s, want only race %d of the asked championship and season", results.raw, askedRace)
	}
}

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

func gridPointsOf(entries []map[string]any, driverID float64) any {
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
	if gridPointsOf(scoreEntries(t, loaded), 1) != float64(25) {
		t.Fatalf("driver 1 points = %v, want 25", gridPointsOf(scoreEntries(t, loaded), 1))
	}

	fetched := h.scoreGrid(fiaAdmin, id)
	if gridPointsOf(scoreEntries(t, fetched), 3) != float64(15) {
		t.Fatalf("the loaded scores must persist, got %v", scoreEntries(t, fetched))
	}

	removed := h.loadScores(fiaAdmin, id, []map[string]any{
		{"driverId": 1, "points": 25},
		{"driverId": 2, "points": 18},
	})
	expectStatus(t, removed, http.StatusOK)
	if gridPointsOf(scoreEntries(t, removed), 3) != nil {
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
	raw, _ := got.body["drivers"].([]any)
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
