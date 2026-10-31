package transporttest

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/tests/testdb"
)

const tyreDayLayout = time.DateOnly

var pastTyreTestDay = time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC)

type tyreTestScenario struct {
	h       *harness
	fia     *http.Client
	eventID int64
}

func newTyreTestScenario(t *testing.T) tyreTestScenario {
	t.Helper()

	h := newHarness(t)
	fia := h.browser()
	expectStatus(t, h.login(fia, h.createUser(domain.RoleFIAAdmin, 0), testPassword), http.StatusOK)
	championship := testdb.NewChampionship(t, h.pool)
	eventID := championship.AddTyreTest(pastTyreTestDay.Add(9*time.Hour), pastTyreTestDay.Add(17*time.Hour))
	return tyreTestScenario{h: h, fia: fia, eventID: eventID}
}

func tyreResultBody(day time.Time, compoundCode string, passed bool, notes string) map[string]any {
	return map[string]any{
		"conductedOn":  day.Format(tyreDayLayout),
		"compoundCode": compoundCode,
		"passed":       passed,
		"notes":        notes,
	}
}

func (s tyreTestScenario) record(client *http.Client, eventID int64, body any) response {
	return s.h.do(client, http.MethodPost, fmt.Sprintf("/tyre-tests/%d/results", eventID), body, nil)
}

func (s tyreTestScenario) correct(client *http.Client, resultID int64, body any) response {
	return s.h.do(client, http.MethodPost, fmt.Sprintf("/tyre-tests/results/%d/corrections", resultID), body, nil)
}

func (s tyreTestScenario) history(query string) []map[string]any {
	s.h.t.Helper()

	got := s.h.do(&http.Client{}, http.MethodGet, "/tyre-tests/results"+query, nil, nil)
	expectStatus(s.h.t, got, http.StatusOK)
	rows, _ := got.body["results"].([]any)
	entries := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, row.(map[string]any))
	}
	return entries
}

func historyEntry(entries []map[string]any, id float64) (map[string]any, bool) {
	for _, entry := range entries {
		if entry["id"] == id {
			return entry, true
		}
	}
	return nil, false
}

func TestRecordedTyreTestResultIsPublishedAsLoaded(t *testing.T) {
	s := newTyreTestScenario(t)

	recorded := s.record(s.fia, s.eventID, tyreResultBody(pastTyreTestDay, "c3", true, "  Degradación baja  "))
	expectStatus(t, recorded, http.StatusCreated)
	if recorded.body["recordedBy"] == nil {
		t.Fatalf("the FIA answer should name its author: %v", recorded.body)
	}

	entry, found := historyEntry(s.history(""), recorded.body["id"].(float64))
	if !found {
		t.Fatalf("result %v missing from the public history", recorded.body["id"])
	}
	want := map[string]any{
		"eventId":      float64(s.eventID),
		"conductedOn":  pastTyreTestDay.Format(tyreDayLayout),
		"passed":       true,
		"notes":        "Degradación baja",
		"isCorrection": false,
	}
	for key, value := range want {
		if entry[key] != value {
			t.Errorf("public %s = %v, want %v", key, entry[key], value)
		}
	}
	if entry["type"].(map[string]any)["code"] != "c3" {
		t.Errorf("public type = %v, want c3", entry["type"])
	}
	if _, leaked := entry["recordedBy"]; leaked {
		t.Error("the public history must not name the FIA administrator")
	}
}

func TestOnlyTheFIARecordsTyreTestResults(t *testing.T) {
	s := newTyreTestScenario(t)
	team := s.h.browser()
	expectStatus(t, s.h.login(team, s.h.createUser(domain.RoleTeamAdmin, 1), testPassword), http.StatusOK)
	body := tyreResultBody(pastTyreTestDay, "c2", true, "")
	chainPath := fmt.Sprintf("/tyre-tests/%d/results", s.eventID)

	expectStatus(t, s.record(&http.Client{}, s.eventID, body), http.StatusUnauthorized)
	expectStatus(t, s.record(team, s.eventID, body), http.StatusForbidden)
	expectStatus(t, s.correct(team, 1, body), http.StatusForbidden)
	expectStatus(t, s.h.do(team, http.MethodGet, chainPath, nil, nil), http.StatusForbidden)
	expectStatus(t, s.h.do(&http.Client{}, http.MethodGet, chainPath, nil, nil), http.StatusUnauthorized)
}

func TestTyreTestResultsCannotBeEditedOrDeleted(t *testing.T) {
	s := newTyreTestScenario(t)
	recorded := s.record(s.fia, s.eventID, tyreResultBody(pastTyreTestDay, "c2", true, ""))
	expectStatus(t, recorded, http.StatusCreated)
	resultPath := fmt.Sprintf("/tyre-tests/results/%d", int64(recorded.body["id"].(float64)))
	chainPath := fmt.Sprintf("/tyre-tests/%d/results", s.eventID)

	for _, method := range []string{http.MethodPut, http.MethodPatch, http.MethodDelete} {
		for _, path := range []string{resultPath, chainPath} {
			if status := s.rawStatus(method, path); status < http.StatusBadRequest {
				t.Errorf("%s %s = %d, want it refused", method, path, status)
			}
		}
	}
}

// rawStatus sends a request the router has no route for, whose answer is plain
// text rather than the JSON the harness decodes.
func (s tyreTestScenario) rawStatus(method, path string) int {
	s.h.t.Helper()

	request, err := http.NewRequest(method, s.h.server.URL+path, nil)
	if err != nil {
		s.h.t.Fatalf("new request: %v", err)
	}
	answer, err := s.fia.Do(request)
	if err != nil {
		s.h.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer answer.Body.Close()
	return answer.StatusCode
}

func TestTyreTestResultValidation(t *testing.T) {
	s := newTyreTestScenario(t)
	tomorrow := domain.DayOf(time.Now()).AddDate(0, 0, 1)
	runningEventID := testdb.NewChampionship(t, s.h.pool).AddTyreTest(time.Now().Add(-24*time.Hour), time.Now().Add(72*time.Hour))
	cancelledEventID := testdb.NewChampionship(t, s.h.pool).AddTyreTest(pastTyreTestDay.Add(9*time.Hour), pastTyreTestDay.Add(17*time.Hour))
	if _, err := s.h.pool.Exec(context.Background(), `UPDATE events SET deactivated_at = now() WHERE id = $1`, cancelledEventID); err != nil {
		t.Fatalf("cancel event: %v", err)
	}

	cases := []struct {
		name    string
		eventID int64
		body    map[string]any
		field   string
	}{
		{"future day", runningEventID, tyreResultBody(tomorrow, "c1", true, ""), "conductedOn"},
		{"day before the test", s.eventID, tyreResultBody(pastTyreTestDay.AddDate(0, 0, -1), "c1", true, ""), "conductedOn"},
		{"day after the test", s.eventID, tyreResultBody(pastTyreTestDay.AddDate(0, 0, 1), "c1", true, ""), "conductedOn"},
		{"cancelled test", cancelledEventID, tyreResultBody(pastTyreTestDay, "c1", true, ""), "conductedOn"},
		{"impossible day", s.eventID, map[string]any{"conductedOn": "2026-02-30", "compoundCode": "c1", "passed": true}, "conductedOn"},
		{"missing day", s.eventID, map[string]any{"compoundCode": "c1", "passed": true}, "conductedOn"},
		{"missing verdict", s.eventID, map[string]any{"conductedOn": "2026-04-14", "compoundCode": "c1"}, "passed"},
		{"failure without notes", s.eventID, tyreResultBody(pastTyreTestDay, "c1", false, "   "), "notes"},
		{"unknown compound", s.eventID, tyreResultBody(pastTyreTestDay, "slick", true, ""), "compoundCode"},
		{"missing compound", s.eventID, tyreResultBody(pastTyreTestDay, "", true, ""), "compoundCode"},
		{"unknown field", s.eventID, map[string]any{"conductedOn": "2026-04-14", "compoundCode": "c1", "passed": true, "recordedBy": 1}, "body"},
	}
	for _, c := range cases {
		got := s.record(s.fia, c.eventID, c.body)
		if got.status != http.StatusBadRequest || got.body["field"] != c.field {
			t.Errorf("%s = %d %v, want 400 on %s", c.name, got.status, got.body, c.field)
		}
	}

	tooLong := tyreResultBody(pastTyreTestDay, "c1", true, string(make([]byte, 2001)))
	if got := s.record(s.fia, s.eventID, tooLong); got.status != http.StatusBadRequest {
		t.Errorf("notes over the limit = %d, want 400", got.status)
	}
}

func TestTyreTestResultNeedsATyreTest(t *testing.T) {
	s := newTyreTestScenario(t)
	raceID := testdb.NewChampionship(t, s.h.pool).AddRace(pastTyreTestDay.Add(15 * time.Hour))
	body := tyreResultBody(pastTyreTestDay, "c1", true, "")

	expectStatus(t, s.record(s.fia, raceID, body), http.StatusNotFound)
	expectStatus(t, s.record(s.fia, 999_999_999, body), http.StatusNotFound)
	correction := map[string]any{"conductedOn": "2026-04-14", "passed": true, "correctionNote": "motivo"}
	expectStatus(t, s.correct(s.fia, 999_999_999, correction), http.StatusNotFound)
}

func TestTyreTestResultIsCorrectedNotDuplicated(t *testing.T) {
	s := newTyreTestScenario(t)
	original := s.record(s.fia, s.eventID, tyreResultBody(pastTyreTestDay, "c5", true, ""))
	expectStatus(t, original, http.StatusCreated)
	originalID := int64(original.body["id"].(float64))

	expectStatus(t, s.record(s.fia, s.eventID, tyreResultBody(pastTyreTestDay, "c5", false, "otra")), http.StatusConflict)

	correction := map[string]any{"conductedOn": "2026-04-14", "passed": false, "notes": "Ampollas", "correctionNote": ""}
	if got := s.correct(s.fia, originalID, correction); got.status != http.StatusBadRequest || got.body["field"] != "correctionNote" {
		t.Fatalf("correction without a reason = %d %v, want 400 on correctionNote", got.status, got.body)
	}
	correction["correctionNote"] = "Se cargó como aprobado"
	corrected := s.correct(s.fia, originalID, correction)
	expectStatus(t, corrected, http.StatusCreated)
	if corrected.body["supersedesId"] != float64(originalID) {
		t.Fatalf("supersedesId = %v, want %d", corrected.body["supersedesId"], originalID)
	}
	expectStatus(t, s.correct(s.fia, originalID, correction), http.StatusConflict)

	history := s.history("?type=c5&from=2026-04-14&to=2026-04-14")
	if _, stillListed := historyEntry(history, float64(originalID)); stillListed {
		t.Error("the superseded result must leave the public history")
	}
	entry, found := historyEntry(history, corrected.body["id"].(float64))
	if !found || entry["isCorrection"] != true || entry["passed"] != false || entry["correctionNote"] != "Se cargó como aprobado" {
		t.Fatalf("public correction = %v, want the current result marked as a correction", entry)
	}

	chain := s.h.do(s.fia, http.MethodGet, fmt.Sprintf("/tyre-tests/%d/results", s.eventID), nil, nil)
	expectStatus(t, chain, http.StatusOK)
	if revisions, _ := chain.body["results"].([]any); len(revisions) != 2 {
		t.Fatalf("FIA chain = %v, want the original and its correction", chain.body)
	}
}

func TestTyreTestHistoryFilters(t *testing.T) {
	s := newTyreTestScenario(t)
	secondDay := pastTyreTestDay.AddDate(0, 0, 1)
	twoDayEventID := testdb.NewChampionship(t, s.h.pool).AddTyreTest(pastTyreTestDay.Add(9*time.Hour), secondDay.Add(17*time.Hour))
	first := s.record(s.fia, twoDayEventID, tyreResultBody(pastTyreTestDay, "intermediate", true, ""))
	second := s.record(s.fia, twoDayEventID, tyreResultBody(secondDay, "wet", true, ""))
	expectStatus(t, first, http.StatusCreated)
	expectStatus(t, second, http.StatusCreated)
	firstID, secondID := first.body["id"].(float64), second.body["id"].(float64)

	cases := []struct {
		query      string
		wantFirst  bool
		wantSecond bool
	}{
		{"?from=2026-04-14&to=2026-04-14", true, false},
		{"?from=2026-04-15", false, true},
		{"?to=2026-04-15", true, true},
		{"?type=wet", false, true},
		{"?type=intermediate&from=2026-04-15", false, false},
	}
	for _, c := range cases {
		entries := s.history(c.query)
		_, hasFirst := historyEntry(entries, firstID)
		_, hasSecond := historyEntry(entries, secondID)
		if hasFirst != c.wantFirst || hasSecond != c.wantSecond {
			t.Errorf("%s lists first=%t second=%t, want %t %t", c.query, hasFirst, hasSecond, c.wantFirst, c.wantSecond)
		}
	}

	for _, query := range []string{"?type=zz", "?from=14-04-2026", "?from=2026-04-15&to=2026-04-14"} {
		expectStatus(t, s.h.do(&http.Client{}, http.MethodGet, "/tyre-tests/results"+query, nil, nil), http.StatusBadRequest)
	}

	types := s.h.do(&http.Client{}, http.MethodGet, "/tyre-tests/types", nil, nil)
	expectStatus(t, types, http.StatusOK)
	if listed, _ := types.body["types"].([]any); len(listed) != 8 {
		t.Errorf("types = %v, want the 8 compounds of the catalogue", types.body)
	}
}

// TestConcurrentTyreTestWritesKeepOneCurrentResult fires the same write twice at
// once: the schema, not a read-then-write check, decides who wins.
func TestConcurrentTyreTestWritesKeepOneCurrentResult(t *testing.T) {
	s := newTyreTestScenario(t)

	originals := s.concurrently(func() response {
		return s.record(s.fia, s.eventID, tyreResultBody(pastTyreTestDay, "c6", true, ""))
	})
	expectOneWinner(t, "original", originals)

	var originalID int64
	for _, got := range originals {
		if got.status == http.StatusCreated {
			originalID = int64(got.body["id"].(float64))
		}
	}
	correction := map[string]any{"conductedOn": "2026-04-14", "passed": true, "notes": "", "correctionNote": "Vueltas mal contadas"}
	expectOneWinner(t, "correction", s.concurrently(func() response { return s.correct(s.fia, originalID, correction) }))
}

func (s tyreTestScenario) concurrently(write func() response) []response {
	const writers = 2
	responses := make([]response, writers)
	var group sync.WaitGroup
	for index := range responses {
		group.Add(1)
		go func() {
			defer group.Done()
			responses[index] = write()
		}()
	}
	group.Wait()
	return responses
}

func expectOneWinner(t *testing.T, what string, responses []response) {
	t.Helper()

	statuses := map[int]int{}
	for _, got := range responses {
		statuses[got.status]++
	}
	if statuses[http.StatusCreated] != 1 || statuses[http.StatusConflict] != 1 {
		t.Fatalf("concurrent %s statuses = %v, want one 201 and one 409", what, statuses)
	}
}
