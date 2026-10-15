package transporttest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/tests/testdb"
)

// Seeded drivers: 1 to 3 race for Ferrari (team 1) and 6 and 7 for Mercedes
// (team 3), so a Ferrari administrator has no business with the latter.
const seededMercedesDriverID = 6

// uniqueDriverNumber answers a number no other driver holds in the team's
// championship, so a re-run against the same test database never collides.
func (h *harness) uniqueDriverNumber(categoryCode string) int32 {
	h.t.Helper()

	var number int32
	err := h.pool.QueryRow(context.Background(), `
		SELECT COALESCE(MAX(sd.number), 0) + 1
		FROM season_drivers sd
		JOIN categories c ON c.id = sd.category_id
		WHERE sd.season_id = $1 AND c.code = $2`,
		seededSeasonID, categoryCode).Scan(&number)
	if err != nil {
		h.t.Fatalf("unique driver number: %v", err)
	}
	return number
}

func (h *harness) newDriverPayload() map[string]any {
	return map[string]any{
		"firstName":     "Nuevo",
		"lastName":      "Piloto",
		"licenseNumber": testdb.UniqueCode(h.t),
		"countryCode":   "ar",
		"isStarter":     true,
		"number":        h.uniqueDriverNumber("f1"),
	}
}

func (h *harness) createTeamDriver(client *http.Client, payload map[string]any) response {
	h.t.Helper()
	return h.do(client, http.MethodPost, "/team/drivers", payload, nil)
}

func (h *harness) listTeamDrivers(client *http.Client) []map[string]any {
	h.t.Helper()

	got := h.do(client, http.MethodGet, "/team/drivers", nil, nil)
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

func driverPath(driver response, action string) string {
	return fmt.Sprintf("/team/drivers/%.0f%s", driver.body["id"], action)
}

func TestDriverManagementRequiresTeamAdmin(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/team/drivers", nil, nil), http.StatusUnauthorized)
	expectStatus(t, h.do(fiaAdmin, http.MethodGet, "/team/drivers", nil, nil), http.StatusForbidden)
	expectStatus(t, h.createTeamDriver(fiaAdmin, h.newDriverPayload()), http.StatusForbidden)
}

func TestTeamAdminCreatesAndEditsADriver(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	payload := h.newDriverPayload()
	created := h.createTeamDriver(teamAdmin, payload)
	expectStatus(t, created, http.StatusCreated)
	if created.body["firstName"] != "Nuevo" || created.body["isStarter"] != true {
		t.Fatalf("created = %v, want a starter named Nuevo", created.body)
	}
	if created.body["number"] != float64(payload["number"].(int32)) {
		t.Fatalf("number = %v, want %v", created.body["number"], payload["number"])
	}

	if !lineupHas(created, h.listTeamDrivers(teamAdmin)) {
		t.Fatal("the new driver must show up in the team lineup")
	}

	edited := h.do(teamAdmin, http.MethodPatch, driverPath(created, ""), map[string]any{
		"lastName":  "Piloto Editado",
		"isStarter": false,
	}, nil)
	expectStatus(t, edited, http.StatusOK)
	if edited.body["lastName"] != "Piloto Editado" || edited.body["isStarter"] != false {
		t.Fatalf("edited = %v, want the renamed reserve", edited.body)
	}

	fetched := h.do(teamAdmin, http.MethodGet, driverPath(created, ""), nil, nil)
	expectStatus(t, fetched, http.StatusOK)
	if fetched.body["lastName"] != "Piloto Editado" {
		t.Fatalf("the edit must persist, got %v", fetched.body["lastName"])
	}
}

func TestTeamAdminCannotReachAnotherTeamsDriver(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	mercedesDriver := response{body: map[string]any{"id": float64(seededMercedesDriverID)}}
	expectStatus(t, h.do(teamAdmin, http.MethodGet, driverPath(mercedesDriver, ""), nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(teamAdmin, http.MethodPatch, driverPath(mercedesDriver, ""), map[string]any{"lastName": "Robado"}, nil), http.StatusNotFound)
	expectStatus(t, h.do(teamAdmin, http.MethodPost, driverPath(mercedesDriver, "/release"), nil, nil), http.StatusNotFound)
}

func TestCreateDriverRejectsBadInput(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	cases := map[string]struct {
		change    func(map[string]any)
		wantField string
	}{
		"blank first name":    {func(p map[string]any) { p["firstName"] = "   " }, "firstName"},
		"missing last name":   {func(p map[string]any) { delete(p, "lastName") }, "lastName"},
		"short license":       {func(p map[string]any) { p["licenseNumber"] = "ab" }, "licenseNumber"},
		"bad country":         {func(p map[string]any) { p["countryCode"] = "Argentina" }, "countryCode"},
		"number out of range": {func(p map[string]any) { p["number"] = 1000 }, "number"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			payload := h.newDriverPayload()
			tc.change(payload)

			got := h.createTeamDriver(teamAdmin, payload)
			expectStatus(t, got, http.StatusBadRequest)
			if got.body["field"] != tc.wantField {
				t.Fatalf("field = %v, want %s; body: %s", got.body["field"], tc.wantField, got.raw)
			}
		})
	}
}

func TestCreateRejectsATakenLicense(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	payload := h.newDriverPayload()
	payload["licenseNumber"] = "FIA-0001"

	got := h.createTeamDriver(teamAdmin, payload)
	expectStatus(t, got, http.StatusConflict)
	if got.body["field"] != "licenseNumber" {
		t.Fatalf("field = %v, want licenseNumber; body: %s", got.body["field"], got.raw)
	}
}

func TestReleasingADriverTakesThemOffTheLineup(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	created := h.createTeamDriver(teamAdmin, h.newDriverPayload())
	expectStatus(t, created, http.StatusCreated)

	expectStatus(t, h.do(teamAdmin, http.MethodPost, driverPath(created, "/release"), nil, nil), http.StatusNoContent)
	if lineupHas(created, h.listTeamDrivers(teamAdmin)) {
		t.Fatal("a released driver must leave the lineup")
	}
	expectStatus(t, h.do(teamAdmin, http.MethodGet, driverPath(created, ""), nil, nil), http.StatusNotFound)
}

func TestDriverDetailOfAnUnknownIdIsNotFound(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	expectStatus(t, h.do(teamAdmin, http.MethodGet, "/team/drivers/999999999", nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(teamAdmin, http.MethodGet, "/team/drivers/not-a-number", nil, nil), http.StatusNotFound)
}

func TestSeededLineupListsStartersFirst(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	lineup := h.listTeamDrivers(teamAdmin)
	if len(lineup) == 0 {
		t.Fatal("the Ferrari lineup of the demo seed should not be empty")
	}
	if lineup[0]["lastName"] == nil {
		t.Fatalf("a lineup row must carry the person facts: %v", lineup[0])
	}
}

func lineupHas(driver response, lineup []map[string]any) bool {
	for _, item := range lineup {
		if item["id"] == driver.body["id"] {
			return true
		}
	}
	return false
}
