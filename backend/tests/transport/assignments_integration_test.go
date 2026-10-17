package transporttest

import (
	"fmt"
	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/tests/testdb"
	"net/http"
	"testing"
	"time"
)

func TestAssignmentHistoryIsScopedAndKeepsDepartures(t *testing.T) {
	h := newHarness(t)
	c := testdb.NewChampionship(t, h.pool)
	team := c.AddTeam("Origen")
	other := c.AddTeam("Destino")
	driver := c.AddDriver(team, "Historial")
	c.Transfer(driver, other, testdb.SeasonStart.Add(24*time.Hour))
	third := c.AddTeam("Ajena")
	c.AddDriver(third, "Privado")
	admin, _ := h.signedIn(domain.RoleTeamAdmin, team)
	fia, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/assignments", nil, nil), http.StatusUnauthorized)
	expectStatus(t, h.do(admin, http.MethodGet, fmt.Sprintf("/assignments?teamId=%d", third), nil, nil), http.StatusNotFound)
	own := h.do(admin, http.MethodGet, "/assignments", nil, nil)
	expectStatus(t, own, http.StatusOK)
	rows := listOf(t, own.body, "assignments")
	if len(rows) != 2 {
		t.Fatalf("own history = %v, want entry and departure", rows)
	}
	for _, row := range rows {
		if row["driverId"] != float64(driver) {
			t.Fatal("foreign history leaked")
		}
	}
	filtered := h.do(fia, http.MethodGet, fmt.Sprintf("/assignments?teamId=%d", team), nil, nil)
	expectStatus(t, filtered, http.StatusOK)
	if len(listOf(t, filtered.body, "assignments")) != 2 {
		t.Fatal("FIA filter must include departures")
	}
	expectStatus(t, h.do(fia, http.MethodGet, "/assignments?teamId=invalid", nil, nil), http.StatusBadRequest)
}
