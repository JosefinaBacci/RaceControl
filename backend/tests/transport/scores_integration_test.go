package transporttest

import (
	"fmt"
	"net/http"
	"testing"
	"time"

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
