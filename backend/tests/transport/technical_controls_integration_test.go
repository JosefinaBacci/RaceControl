package transporttest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/tests/testdb"
)

// controlDay is the day of the past control every test grades: early in the
// season, so it has always happened by the time the suite runs.
var controlStart = time.Date(2026, 3, 5, 9, 0, 0, 0, time.UTC)

const controlDay = "2026-03-05"

type controlFixture struct {
	h            *harness
	championship *testdb.Championship
	eventID      int64
	controlType  string
	teamA        int64
	teamB        int64
	fia          *http.Client
}

func newControlFixture(t *testing.T) *controlFixture {
	t.Helper()
	h := newHarness(t)
	championship := testdb.NewChampionship(t, h.pool)
	controlType := "Verificación " + championship.CategoryCode
	fixture := &controlFixture{
		h:            h,
		championship: championship,
		controlType:  controlType,
		eventID:      championship.AddTechnicalControl(controlType, controlStart, controlStart.Add(27*time.Hour)),
		teamA:        championship.AddTeam("Equipo A"),
		teamB:        championship.AddTeam("Equipo B"),
		fia:          h.browser(),
	}
	expectStatus(t, h.login(fixture.fia, h.createUser(domain.RoleFIAAdmin, 0), testPassword), http.StatusOK)
	return fixture
}

func (f *controlFixture) record(eventID int64, payload map[string]any) response {
	f.h.t.Helper()
	return f.h.do(f.fia, http.MethodPost, fmt.Sprintf("/technical-controls/%d/results", eventID), payload, nil)
}

func (f *controlFixture) correct(resultID int64, payload map[string]any) response {
	f.h.t.Helper()
	return f.h.do(f.fia, http.MethodPost, fmt.Sprintf("/technical-controls/results/%d/corrections", resultID), payload, nil)
}

func (f *controlFixture) history(query string) []map[string]any {
	f.h.t.Helper()
	got := f.h.do(&http.Client{}, http.MethodGet, "/technical-controls/results?category="+f.championship.CategoryCode+query, nil, nil)
	expectStatus(f.h.t, got, http.StatusOK)
	return entries(got.body["results"])
}

func passedResult(teamID int64) map[string]any {
	return map[string]any{"teamId": teamID, "conductedOn": controlDay, "passed": true, "notes": ""}
}

func failedResult(teamID int64, notes string) map[string]any {
	return map[string]any{"teamId": teamID, "conductedOn": controlDay, "passed": false, "notes": notes}
}

// correctionOf is a correction body: it never names the team, which the
// corrected result already fixes.
func correctionOf(passed bool, notes, reason string) map[string]any {
	return map[string]any{"conductedOn": controlDay, "passed": passed, "notes": notes, "correctionNote": reason}
}

func entries(value any) []map[string]any {
	list, _ := value.([]any)
	out := make([]map[string]any, 0, len(list))
	for _, item := range list {
		entry, _ := item.(map[string]any)
		out = append(out, entry)
	}
	return out
}

func idOf(t *testing.T, got response) int64 {
	t.Helper()
	id, ok := got.body["id"].(float64)
	if !ok {
		t.Fatalf("response has no id: %s", got.raw)
	}
	return int64(id)
}

func expectField(t *testing.T, got response, field string) {
	t.Helper()
	expectStatus(t, got, http.StatusBadRequest)
	if got.body["field"] != field {
		t.Fatalf("field = %v, want %q; body: %s", got.body["field"], field, got.raw)
	}
}

func TestRecordedControlResultsMatchThePublicHistory(t *testing.T) {
	f := newControlFixture(t)

	recorded := f.record(f.eventID, failedResult(f.teamA, "  Fondo plano fuera de medida "))
	expectStatus(t, recorded, http.StatusCreated)
	if recorded.body["recordedBy"] == nil || recorded.body["notes"] != "Fondo plano fuera de medida" {
		t.Fatalf("recorded = %s, want the FIA view with trimmed notes", recorded.raw)
	}
	expectStatus(t, f.record(f.eventID, passedResult(f.teamB)), http.StatusCreated)

	history := f.history("")
	if len(history) != 2 {
		t.Fatalf("history = %v, want both results", history)
	}
	failed := history[0]
	if failed["conductedOn"] != controlDay || failed["passed"] != false || failed["notes"] != "Fondo plano fuera de medida" {
		t.Errorf("failed result = %v, want what the FIA recorded", failed)
	}
	if failed["type"].(map[string]any)["name"] != f.controlType || failed["team"].(map[string]any)["name"] != "Equipo A" {
		t.Errorf("failed result = %v, want its control type and team", failed)
	}
	for _, entry := range history {
		if _, leaked := entry["recordedBy"]; leaked {
			t.Errorf("public entry %v names the administrator", entry)
		}
	}
}

func TestControlTypesListTheRecordedOnes(t *testing.T) {
	f := newControlFixture(t)
	expectStatus(t, f.record(f.eventID, passedResult(f.teamA)), http.StatusCreated)

	got := f.h.do(&http.Client{}, http.MethodGet, "/technical-controls/types", nil, nil)
	expectStatus(t, got, http.StatusOK)
	if !strings.Contains(got.raw, f.controlType) {
		t.Fatalf("types = %s, want %q", got.raw, f.controlType)
	}
}

func TestOnlyTheFIARecordsAndCorrectsControlResults(t *testing.T) {
	f := newControlFixture(t)
	resultID := idOf(t, f.record(f.eventID, passedResult(f.teamA)))
	team := f.h.browser()
	expectStatus(t, f.h.login(team, f.h.createUser(domain.RoleTeamAdmin, f.teamA), testPassword), http.StatusOK)

	writes := []struct{ method, path string }{
		{http.MethodPost, fmt.Sprintf("/technical-controls/%d/results", f.eventID)},
		{http.MethodGet, fmt.Sprintf("/technical-controls/%d/results", f.eventID)},
		{http.MethodPost, fmt.Sprintf("/technical-controls/results/%d/corrections", resultID)},
	}
	for _, write := range writes {
		expectStatus(t, f.h.do(&http.Client{}, write.method, write.path, passedResult(f.teamB), nil), http.StatusUnauthorized)
		expectStatus(t, f.h.do(team, write.method, write.path, passedResult(f.teamB), nil), http.StatusForbidden)
	}
	if len(f.history("")) != 1 {
		t.Fatal("a refused write changed the history")
	}
}

func TestControlResultsCannotBeEditedOrDeleted(t *testing.T) {
	f := newControlFixture(t)
	resultID := idOf(t, f.record(f.eventID, passedResult(f.teamA)))

	for _, method := range []string{http.MethodPut, http.MethodPatch, http.MethodDelete} {
		got := f.h.do(f.fia, method, fmt.Sprintf("/technical-controls/%d/results", f.eventID), passedResult(f.teamA), nil)
		expectStatus(t, got, http.StatusMethodNotAllowed)
	}
	for _, statement := range []string{
		`UPDATE technical_control_results SET passed = false, notes = 'x' WHERE id = $1`,
		`DELETE FROM technical_control_results WHERE id = $1`,
	} {
		if _, err := f.h.pool.Exec(context.Background(), statement, resultID); err == nil {
			t.Errorf("%q succeeded, want the append-only guard to refuse it", statement)
		}
	}
}

func TestRecordControlResultRejectsInvalidInput(t *testing.T) {
	f := newControlFixture(t)
	otherChampionship := testdb.NewChampionship(t, f.h.pool)
	foreignTeam := otherChampionship.AddTeam("Equipo ajeno")
	retiredTeam := f.championship.AddTeam("Equipo retirado")
	if _, err := f.h.pool.Exec(context.Background(), `UPDATE teams SET deactivated_at = now() WHERE id = $1`, retiredTeam); err != nil {
		t.Fatalf("deactivate team: %v", err)
	}

	cases := []struct {
		name    string
		payload map[string]any
		field   string
	}{
		{"malformed day", map[string]any{"teamId": f.teamA, "conductedOn": "05/03/2026", "passed": true}, "conductedOn"},
		{"impossible day", map[string]any{"teamId": f.teamA, "conductedOn": "2026-02-30", "passed": true}, "conductedOn"},
		{"missing day", map[string]any{"teamId": f.teamA, "passed": true}, "conductedOn"},
		{"day before the control", map[string]any{"teamId": f.teamA, "conductedOn": "2026-03-04", "passed": true}, "conductedOn"},
		{"day after the control", map[string]any{"teamId": f.teamA, "conductedOn": "2026-03-07", "passed": true}, "conductedOn"},
		{"missing outcome", map[string]any{"teamId": f.teamA, "conductedOn": controlDay}, "passed"},
		{"failed without notes", failedResult(f.teamA, "   "), "notes"},
		{"notes too long", failedResult(f.teamA, strings.Repeat("a", 2001)), "notes"},
		{"missing team", map[string]any{"conductedOn": controlDay, "passed": true}, "teamId"},
		{"unknown team", passedResult(999999999), "teamId"},
		{"team of another championship", passedResult(foreignTeam), "teamId"},
		{"deactivated team", passedResult(retiredTeam), "teamId"},
		{"unknown field", map[string]any{"teamId": f.teamA, "conductedOn": controlDay, "passed": true, "recordedBy": 1}, "body"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectField(t, f.record(f.eventID, tc.payload), tc.field)
		})
	}
	if len(f.history("")) != 0 {
		t.Fatal("a rejected result was stored")
	}
}

func TestControlResultsWaitForTheControlToHappen(t *testing.T) {
	f := newControlFixture(t)
	tomorrow := time.Now().UTC().Add(24 * time.Hour)
	upcoming := f.championship.AddTechnicalControl(f.controlType, tomorrow, tomorrow.Add(time.Hour))

	expectField(t, f.record(upcoming, map[string]any{
		"teamId": f.teamA, "conductedOn": tomorrow.Format(time.DateOnly), "passed": true,
	}), "conductedOn")
}

func TestCancelledControlsTakeNoResults(t *testing.T) {
	f := newControlFixture(t)
	if _, err := f.h.pool.Exec(context.Background(), `UPDATE events SET deactivated_at = now() WHERE id = $1`, f.eventID); err != nil {
		t.Fatalf("cancel control: %v", err)
	}
	expectField(t, f.record(f.eventID, passedResult(f.teamA)), "conductedOn")
}

func TestControlResultsNeedAnExistingControl(t *testing.T) {
	f := newControlFixture(t)
	race := f.championship.AddRace(controlStart)

	expectStatus(t, f.record(999999999, passedResult(f.teamA)), http.StatusNotFound)
	expectStatus(t, f.record(race, passedResult(f.teamA)), http.StatusNotFound)
	expectStatus(t, f.h.do(f.fia, http.MethodPost, "/technical-controls/abc/results", passedResult(f.teamA), nil), http.StatusNotFound)
}

func TestASecondResultForTheSameTeamIsAConflict(t *testing.T) {
	f := newControlFixture(t)
	expectStatus(t, f.record(f.eventID, passedResult(f.teamA)), http.StatusCreated)
	expectStatus(t, f.record(f.eventID, failedResult(f.teamA, "otra vez")), http.StatusConflict)
}

func TestConcurrentResultsForTheSameTeamStoreOne(t *testing.T) {
	f := newControlFixture(t)
	statuses := concurrently(10, func() int { return f.record(f.eventID, passedResult(f.teamA)).status })

	if statuses[http.StatusCreated] != 1 || statuses[http.StatusConflict] != 9 {
		t.Fatalf("statuses = %v, want one 201 and nine 409", statuses)
	}
}

func TestACorrectionSupersedesTheResultInForce(t *testing.T) {
	f := newControlFixture(t)
	originalID := idOf(t, f.record(f.eventID, passedResult(f.teamA)))

	corrected := f.correct(originalID, correctionOf(false, "Alerón trasero fuera de medida", "Se cargó como aprobado por error"))
	expectStatus(t, corrected, http.StatusCreated)
	if corrected.body["supersedesId"] != float64(originalID) || corrected.body["isCorrection"] != true {
		t.Fatalf("corrected = %s, want it to supersede %d", corrected.raw, originalID)
	}

	history := f.history("")
	if len(history) != 1 || history[0]["passed"] != false || history[0]["isCorrection"] != true ||
		history[0]["correctionNote"] != "Se cargó como aprobado por error" {
		t.Fatalf("history = %v, want only the correction, marked as such", history)
	}

	sheet := f.h.do(f.fia, http.MethodGet, fmt.Sprintf("/technical-controls/%d/results", f.eventID), nil, nil)
	expectStatus(t, sheet, http.StatusOK)
	for _, team := range entries(sheet.body["teams"]) {
		revisions := entries(team["revisions"])
		if team["name"] == "Equipo A" && len(revisions) != 2 {
			t.Errorf("team A revisions = %v, want the original and its correction", revisions)
		}
		if team["name"] == "Equipo B" && len(revisions) != 0 {
			t.Errorf("team B revisions = %v, want it pending", revisions)
		}
	}
}

func TestCorrectionsAreChecked(t *testing.T) {
	f := newControlFixture(t)
	originalID := idOf(t, f.record(f.eventID, passedResult(f.teamA)))
	movedDay := correctionOf(true, "", "fecha")
	movedDay["conductedOn"] = "2026-04-01"
	withTeam := correctionOf(true, "", "equipo")
	withTeam["teamId"] = f.teamB

	expectField(t, f.correct(originalID, correctionOf(true, "", "  ")), "correctionNote")
	expectField(t, f.correct(originalID, movedDay), "conductedOn")
	expectField(t, f.correct(originalID, withTeam), "body")
	expectStatus(t, f.correct(999999999, correctionOf(true, "", "motivo")), http.StatusNotFound)

	expectStatus(t, f.correct(originalID, correctionOf(false, "peso", "motivo")), http.StatusCreated)
	expectStatus(t, f.correct(originalID, correctionOf(true, "", "de nuevo")), http.StatusConflict)
}

func TestConcurrentCorrectionsOfOneResultStoreOne(t *testing.T) {
	f := newControlFixture(t)
	originalID := idOf(t, f.record(f.eventID, passedResult(f.teamA)))
	statuses := concurrently(10, func() int {
		return f.correct(originalID, correctionOf(false, "peso", "motivo")).status
	})

	if statuses[http.StatusCreated] != 1 || statuses[http.StatusConflict] != 9 {
		t.Fatalf("statuses = %v, want one 201 and nine 409", statuses)
	}
}

func TestAControlWithResultsIsFinal(t *testing.T) {
	f := newControlFixture(t)
	expectStatus(t, f.record(f.eventID, passedResult(f.teamA)), http.StatusCreated)
	path := fmt.Sprintf("/events/%d", f.eventID)
	later := time.Now().UTC().Add(48 * time.Hour)

	expectStatus(t, f.h.do(f.fia, http.MethodPatch, path, map[string]any{"controlType": "Otro tipo"}, nil), http.StatusConflict)
	expectStatus(t, f.h.do(f.fia, http.MethodPatch, path, map[string]any{
		"startsAt": later, "endsAt": later.Add(time.Hour),
	}, nil), http.StatusConflict)
	expectStatus(t, f.h.do(f.fia, http.MethodPatch, path, map[string]any{"controlType": f.controlType}, nil), http.StatusOK)
}

func TestControlHistoryFilters(t *testing.T) {
	f := newControlFixture(t)
	laterStart := controlStart.AddDate(0, 0, 10)
	otherType := "Pesaje " + f.championship.CategoryCode
	laterControl := f.championship.AddTechnicalControl(otherType, laterStart, laterStart.Add(time.Hour))
	expectStatus(t, f.record(f.eventID, passedResult(f.teamA)), http.StatusCreated)
	expectStatus(t, f.record(laterControl, map[string]any{
		"teamId": f.teamB, "conductedOn": laterStart.Format(time.DateOnly), "passed": true,
	}), http.StatusCreated)

	cases := []struct {
		query string
		want  int
	}{
		{"", 2},
		{"&from=" + controlDay + "&to=" + controlDay, 1},
		{"&from=2026-03-06", 1},
		{"&to=2026-03-04", 0},
		{"&type=" + strings.ReplaceAll(otherType, " ", "%20"), 1},
		{"&type=inexistente", 0},
		{fmt.Sprintf("&team=%d", f.teamA), 1},
	}
	for _, tc := range cases {
		if got := f.history(tc.query); len(got) != tc.want {
			t.Errorf("history%s = %d entries, want %d", tc.query, len(got), tc.want)
		}
	}

	for query, field := range map[string]string{
		"?from=2026-13-01":               "from",
		"?to=ayer":                       "to",
		"?from=2026-03-02&to=2026-03-01": "to",
		"?team=abc":                      "team",
	} {
		expectField(t, f.h.do(&http.Client{}, http.MethodGet, "/technical-controls/results"+query, nil, nil), field)
	}
}

// concurrently runs request from many goroutines at once and counts the statuses.
func concurrently(times int, request func() int) map[int]int {
	var (
		mutex    sync.Mutex
		group    sync.WaitGroup
		statuses = map[int]int{}
	)
	start := make(chan struct{})
	for range times {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			status := request()
			mutex.Lock()
			statuses[status]++
			mutex.Unlock()
		}()
	}
	close(start)
	group.Wait()
	return statuses
}
