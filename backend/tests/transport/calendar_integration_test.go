package transporttest

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/racecontrol/backend/internal/domain"
)

const (
	seededSeasonID     = 1
	seededCircuitID    = 1
	seededRacedEventID = 1
	seededRoundNumber  = 18
	formulaOneCategory = "f1"
)

// nextRoundNumber answers a number no active round of the championship holds, so
// the tests leave the calendar the demo seed describes intact. Deactivated rounds
// are counted too, which is safe and never reuses a number.
func (h *harness) nextRoundNumber(categoryCode string) int32 {
	h.t.Helper()

	var number int32
	err := h.pool.QueryRow(context.Background(), `
		SELECT COALESCE(MAX(r.number), 0) + 1
		FROM rounds r JOIN categories c ON c.id = r.category_id
		WHERE r.season_id = $1 AND c.code = $2`,
		seededSeasonID, categoryCode).Scan(&number)
	if err != nil {
		h.t.Fatalf("next round number: %v", err)
	}
	return number
}

func (h *harness) roundIsActive(categoryCode string, number int32) bool {
	h.t.Helper()

	var isActive bool
	err := h.pool.QueryRow(context.Background(), `
		SELECT r.is_active
		FROM rounds r JOIN categories c ON c.id = r.category_id
		WHERE r.season_id = $1 AND c.code = $2 AND r.number = $3`,
		seededSeasonID, categoryCode, number).Scan(&isActive)
	if err != nil {
		h.t.Fatalf("round %d of %s: %v", number, categoryCode, err)
	}
	return isActive
}

func (h *harness) createRace(fiaAdmin *http.Client, startsAt time.Time, number int32, gpName string) response {
	h.t.Helper()

	return h.do(fiaAdmin, http.MethodPost, "/events", map[string]any{
		"kind":         "race",
		"startsAt":     startsAt.Format(time.RFC3339),
		"endsAt":       startsAt.Add(2 * time.Hour).Format(time.RFC3339),
		"seasonId":     seededSeasonID,
		"categoryCode": formulaOneCategory,
		"circuitId":    seededCircuitID,
		"number":       number,
		"gpName":       gpName,
	}, nil)
}

func eventID(t *testing.T, created response) int64 {
	t.Helper()

	id, ok := created.body["id"].(float64)
	if !ok {
		t.Fatalf("the event has no id: %s", created.raw)
	}
	return int64(id)
}

func (h *harness) listEvents(path string) []map[string]any {
	h.t.Helper()

	got := h.do(h.browser(), http.MethodGet, path, nil, nil)
	expectStatus(h.t, got, http.StatusOK)

	raw, _ := got.body["events"].([]any)
	events := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if event, ok := item.(map[string]any); ok {
			events = append(events, event)
		}
	}
	return events
}

func listHasID(events []map[string]any, id int64) bool {
	for _, event := range events {
		if event["id"] == float64(id) {
			return true
		}
	}
	return false
}

func TestCalendarIsReadableWithoutASession(t *testing.T) {
	h := newHarness(t)

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/events", nil, nil), http.StatusOK)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/circuits", nil, nil), http.StatusOK)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/seasons", nil, nil), http.StatusOK)
}

func TestCalendarWritesRequireFIAAdmin(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)
	payload := map[string]any{"kind": "technical_control"}

	expectStatus(t, h.do(h.browser(), http.MethodPost, "/events", payload, nil), http.StatusUnauthorized)
	expectStatus(t, h.do(teamAdmin, http.MethodPost, "/events", payload, nil), http.StatusForbidden)

	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	got := h.do(fiaAdmin, http.MethodPost, "/events", payload, nil)
	if got.status == http.StatusUnauthorized || got.status == http.StatusForbidden {
		t.Fatalf("the FIA admin must reach the handler, got %d: %s", got.status, got.raw)
	}
}

func TestFIAAdminCreatesARaceWithItsRound(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	number := h.nextRoundNumber(formulaOneCategory)
	created := h.createRace(fiaAdmin, time.Date(2026, 5, 3, 15, 0, 0, 0, time.UTC), number, "Gran Premio de Prueba")
	expectStatus(t, created, http.StatusCreated)

	wantName := fmt.Sprintf("Gran Premio de Prueba · Ronda %d", number)
	if created.body["name"] != wantName {
		t.Fatalf("name = %v, want %q", created.body["name"], wantName)
	}
	if created.body["categoryCode"] != formulaOneCategory || created.body["seasonYear"] != float64(2026) {
		t.Fatalf("championship = %v / %v, want f1 / 2026", created.body["categoryCode"], created.body["seasonYear"])
	}
	if created.body["circuitName"] != "Albert Park" || created.body["countryCode"] != "au" {
		t.Fatalf("circuit = %v / %v, want Albert Park / au", created.body["circuitName"], created.body["countryCode"])
	}
	if created.body["isActive"] != true {
		t.Fatalf("isActive = %v, want true", created.body["isActive"])
	}
}

func TestFIAAdminCreatesATyreTestAndATechnicalControl(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	tyreTest := h.do(fiaAdmin, http.MethodPost, "/events", map[string]any{
		"kind":         "tyre_test",
		"startsAt":     "2026-06-10T09:00:00Z",
		"endsAt":       "2026-06-10T17:00:00Z",
		"seasonId":     seededSeasonID,
		"categoryCode": formulaOneCategory,
		"circuitId":    seededCircuitID,
	}, nil)
	expectStatus(t, tyreTest, http.StatusCreated)
	if tyreTest.body["name"] != "Prueba de neumáticos · 10/06/2026" {
		t.Fatalf("tyre test name = %v", tyreTest.body["name"])
	}
	if tyreTest.body["roundNumber"] != nil {
		t.Fatalf("a tyre test has no round, got %v", tyreTest.body["roundNumber"])
	}

	control := h.do(fiaAdmin, http.MethodPost, "/events", map[string]any{
		"kind":         "technical_control",
		"startsAt":     "2026-06-11T09:00:00Z",
		"endsAt":       "2026-06-11T12:00:00Z",
		"seasonId":     seededSeasonID,
		"categoryCode": formulaOneCategory,
		"controlType":  "Verificación de pesos",
	}, nil)
	expectStatus(t, control, http.StatusCreated)
	if control.body["name"] != "Verificación de pesos" {
		t.Fatalf("control name = %v", control.body["name"])
	}
	if control.body["circuitId"] != nil {
		t.Fatalf("a technical control has no circuit, got %v", control.body["circuitId"])
	}
}

func TestCreateEventRejectsBadInput(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	valid := map[string]any{
		"kind":         "race",
		"startsAt":     "2026-05-03T15:00:00Z",
		"endsAt":       "2026-05-03T17:00:00Z",
		"seasonId":     seededSeasonID,
		"categoryCode": formulaOneCategory,
		"circuitId":    seededCircuitID,
		"number":       float64(1),
		"gpName":       "Gran Premio de Prueba",
	}

	cases := map[string]func(map[string]any){
		"unknown kind":     func(p map[string]any) { p["kind"] = "practice" },
		"end before start": func(p map[string]any) { p["endsAt"] = "2026-05-03T14:00:00Z" },
		"date outside season": func(p map[string]any) {
			p["startsAt"] = "2027-05-03T15:00:00Z"
			p["endsAt"] = "2027-05-03T17:00:00Z"
		},
		"unknown category":     func(p map[string]any) { p["categoryCode"] = "f0" },
		"unknown season":       func(p map[string]any) { p["seasonId"] = float64(999999) },
		"unknown circuit":      func(p map[string]any) { p["circuitId"] = float64(999999) },
		"race without number":  func(p map[string]any) { delete(p, "number") },
		"race without gp name": func(p map[string]any) { delete(p, "gpName") },
		"tyre test without circuit": func(p map[string]any) {
			p["kind"] = "tyre_test"
			delete(p, "circuitId")
		},
		"control without type": func(p map[string]any) {
			p["kind"] = "technical_control"
			delete(p, "circuitId")
		},
	}

	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			payload := make(map[string]any, len(valid))
			for key, value := range valid {
				payload[key] = value
			}
			corrupt(payload)

			got := h.do(fiaAdmin, http.MethodPost, "/events", payload, nil)
			expectStatus(t, got, http.StatusBadRequest)
			if got.body["code"] != "validation" {
				t.Fatalf("code = %v, want validation; body: %s", got.body["code"], got.raw)
			}
		})
	}
}

func TestCreateRejectsARoundNumberInUse(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	got := h.createRace(fiaAdmin, time.Date(2026, 5, 3, 15, 0, 0, 0, time.UTC), seededRoundNumber, "Gran Premio Repetido")
	expectStatus(t, got, http.StatusConflict)
	if got.body["code"] != "conflict" {
		t.Fatalf("code = %v, want conflict", got.body["code"])
	}
	if got.body["field"] != "number" {
		t.Fatalf("field = %v, want number", got.body["field"])
	}
}

func TestCancellingARaceTakesItsRoundOffTheCalendar(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	number := h.nextRoundNumber(formulaOneCategory)
	created := h.createRace(fiaAdmin, time.Date(2026, 5, 3, 15, 0, 0, 0, time.UTC), number, "Gran Premio Cancelado")
	expectStatus(t, created, http.StatusCreated)
	id := eventID(t, created)

	cancelled := h.do(fiaAdmin, http.MethodDelete, fmt.Sprintf("/events/%d", id), nil, nil)
	expectStatus(t, cancelled, http.StatusOK)
	if cancelled.body["isActive"] != false {
		t.Fatalf("isActive = %v, want false", cancelled.body["isActive"])
	}
	if h.roundIsActive(formulaOneCategory, number) {
		t.Fatalf("cancelling the race must also cancel its round %d", number)
	}

	if listHasID(h.listEvents("/events?category=f1"), eventID(t, created)) {
		t.Fatal("a cancelled event must leave the calendar")
	}
}

func TestCancellingARacedRaceIsRefused(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	got := h.do(fiaAdmin, http.MethodDelete, fmt.Sprintf("/events/%d", seededRacedEventID), nil, nil)
	expectStatus(t, got, http.StatusConflict)
	if got.body["code"] != "conflict" {
		t.Fatalf("code = %v, want conflict", got.body["code"])
	}
}

func TestEditingARacedRaceIsRefused(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	got := h.do(fiaAdmin, http.MethodPatch, fmt.Sprintf("/events/%d", seededRacedEventID),
		map[string]any{"startsAt": "2026-10-04T16:00:00Z", "endsAt": "2026-10-04T18:00:00Z"}, nil)
	expectStatus(t, got, http.StatusConflict)
}

func TestUpdateEditsTheRoundOfARace(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	number := h.nextRoundNumber(formulaOneCategory)
	created := h.createRace(fiaAdmin, time.Date(2026, 5, 3, 15, 0, 0, 0, time.UTC), number, "Gran Premio Original")
	expectStatus(t, created, http.StatusCreated)
	id := eventID(t, created)

	updated := h.do(fiaAdmin, http.MethodPatch, fmt.Sprintf("/events/%d", id),
		map[string]any{"gpName": "Gran Premio Renombrado"}, nil)
	expectStatus(t, updated, http.StatusOK)

	wantName := fmt.Sprintf("Gran Premio Renombrado · Ronda %d", number)
	if updated.body["name"] != wantName {
		t.Fatalf("name = %v, want %q", updated.body["name"], wantName)
	}

	fetched := h.do(h.browser(), http.MethodGet, fmt.Sprintf("/events/%d", id), nil, nil)
	expectStatus(t, fetched, http.StatusOK)
	if fetched.body["name"] != wantName {
		t.Fatalf("the edit must persist, got %v", fetched.body["name"])
	}
}

func TestEventListFiltersByCategoryAndKind(t *testing.T) {
	h := newHarness(t)

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/events?period=whenever", nil, nil), http.StatusBadRequest)

	for _, event := range h.listEvents("/events?kind=tyre_test") {
		if event["kind"] != "tyre_test" {
			t.Fatalf("kind filter leaked %v", event["kind"])
		}
	}

	academy := h.listEvents("/events?category=academy")
	if len(academy) == 0 {
		t.Fatal("the academy round of the demo seed should be listed")
	}
	for _, event := range academy {
		if event["categoryCode"] != "academy" {
			t.Fatalf("category filter leaked %v", event["categoryCode"])
		}
	}
}

func TestEventListFiltersByPeriod(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	now := time.Now().UTC()
	number := h.nextRoundNumber(formulaOneCategory)

	past := h.createRace(fiaAdmin, now.Add(-48*time.Hour), number, "Gran Premio Pasado")
	expectStatus(t, past, http.StatusCreated)
	future := h.createRace(fiaAdmin, now.Add(48*time.Hour), number+1, "Gran Premio Futuro")
	expectStatus(t, future, http.StatusCreated)

	pastList := h.listEvents("/events?period=past")
	if !listHasID(pastList, eventID(t, past)) {
		t.Fatal("a past event must be in the past list")
	}
	if listHasID(pastList, eventID(t, future)) {
		t.Fatal("a future event must not be in the past list")
	}

	futureList := h.listEvents("/events?period=upcoming")
	if !listHasID(futureList, eventID(t, future)) {
		t.Fatal("a future event must be in the upcoming list")
	}
	if listHasID(futureList, eventID(t, past)) {
		t.Fatal("a past event must not be in the upcoming list")
	}
}

func TestEventDetailOfAnUnknownIdIsNotFound(t *testing.T) {
	h := newHarness(t)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/events/999999999", nil, nil), http.StatusNotFound)
}

func TestCircuitsAndSeasonsAreCatalogued(t *testing.T) {
	h := newHarness(t)

	circuits := h.do(h.browser(), http.MethodGet, "/circuits", nil, nil)
	expectStatus(t, circuits, http.StatusOK)
	list, _ := circuits.body["circuits"].([]any)
	if len(list) == 0 {
		t.Fatalf("circuits = %s", circuits.raw)
	}
	first, _ := list[0].(map[string]any)
	for _, field := range []string{"id", "name", "locality", "countryCode"} {
		if first[field] == nil {
			t.Fatalf("circuit %v has no %s", first, field)
		}
	}

	seasons := h.do(h.browser(), http.MethodGet, "/seasons", nil, nil)
	expectStatus(t, seasons, http.StatusOK)
	years, _ := seasons.body["seasons"].([]any)
	if len(years) == 0 {
		t.Fatalf("seasons = %s", seasons.raw)
	}
	firstSeason, _ := years[0].(map[string]any)
	if firstSeason["year"] != float64(2026) {
		t.Fatalf("season = %v, want 2026", firstSeason["year"])
	}
}
