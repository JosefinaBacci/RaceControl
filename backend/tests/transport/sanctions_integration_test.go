package transporttest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/tests/testdb"
)

// sanctionFixture is a championship of its own with two teams, a driver of team
// A and a race, plus a FIA administrator. Every active team of the championship
// receives each sanction, so a fresh one keeps other tests' sanctions out.
type sanctionFixture struct {
	h            *harness
	championship *testdb.Championship
	teamA        int64
	teamB        int64
	alonso       int64
	race         int64
	fiaAdmin     *http.Client
	fiaName      string
}

func (h *harness) newSanctionFixture() sanctionFixture {
	h.t.Helper()

	championship := testdb.NewChampionship(h.t, h.pool)
	teamA := championship.AddTeam("Equipo A")
	fiaAdmin, fiaName := h.signedIn(domain.RoleFIAAdmin, 0)
	return sanctionFixture{
		h:            h,
		championship: championship,
		teamA:        teamA,
		teamB:        championship.AddTeam("Equipo B"),
		alonso:       championship.AddDriver(teamA, "Alonso"),
		race:         championship.AddRace(firstRaceStart),
		fiaAdmin:     fiaAdmin,
		fiaName:      fiaName,
	}
}

func (f sanctionFixture) driverSanction() map[string]any {
	return map[string]any{
		"driverId":     f.alonso,
		"penaltyType":  "time_penalty",
		"penaltyValue": 5,
		"reason":       "Exceder los límites de pista en la vuelta 12",
		"eventId":      f.race,
	}
}

func (f sanctionFixture) issue(payload map[string]any) response {
	f.h.t.Helper()
	return f.h.do(f.fiaAdmin, http.MethodPost, "/sanctions", payload, nil)
}

func (f sanctionFixture) issueOnAlonso() response {
	f.h.t.Helper()

	created := f.issue(f.driverSanction())
	expectStatus(f.h.t, created, http.StatusCreated)
	return created
}

func (f sanctionFixture) amendment(note string) map[string]any {
	payload := f.driverSanction()
	delete(payload, "driverId")
	payload["penaltyValue"] = 10
	payload["note"] = note
	return payload
}

func (f sanctionFixture) listQuery(extra string) string {
	return "/sanctions?category=" + url.QueryEscape(f.championship.CategoryCode) + extra
}

func sanctionPath(created response, suffix string) string {
	return fmt.Sprintf("/sanctions/%v%s", created.body["id"], suffix)
}

func currentVersionID(t *testing.T, sanction map[string]any) float64 {
	t.Helper()
	return objectOf(t, sanction, "current")["id"].(float64)
}

func sanctionAcknowledgementPath(versionID float64) string {
	return fmt.Sprintf("/team/sanction-notices/%v/acknowledgement", versionID)
}

func (h *harness) sanctionNotices(client *http.Client) response {
	h.t.Helper()

	got := h.do(client, http.MethodGet, "/team/sanction-notices", nil, nil)
	expectStatus(h.t, got, http.StatusOK)
	return got
}

func sanctionNoticeFor(t *testing.T, board response, sanctionID any) map[string]any {
	t.Helper()

	for _, notice := range listOf(t, board.body, "notices") {
		if objectOf(t, notice, "sanction")["id"] == sanctionID {
			return notice
		}
	}
	t.Fatalf("no notice for sanction %v in %v", sanctionID, board.body)
	return nil
}

func (f sanctionFixture) followUpOf(sanctionID any) map[string]any {
	f.h.t.Helper()

	got := f.h.do(f.fiaAdmin, http.MethodGet, "/sanction-acknowledgements", nil, nil)
	expectStatus(f.h.t, got, http.StatusOK)
	for _, followUp := range listOf(f.h.t, got.body, "followUps") {
		if objectOf(f.h.t, followUp, "sanction")["id"] == sanctionID {
			return followUp
		}
	}
	f.h.t.Fatalf("no follow-up for sanction %v", sanctionID)
	return nil
}

func listIDs(t *testing.T, got response) []any {
	t.Helper()

	ids := []any{}
	for _, sanction := range listOf(t, got.body, "sanctions") {
		ids = append(ids, sanction["id"])
	}
	return ids
}

func expectIDs(t *testing.T, got response, want ...any) {
	t.Helper()

	expectStatus(t, got, http.StatusOK)
	ids := listIDs(t, got)
	if fmt.Sprintf("%v", ids) != fmt.Sprintf("%v", want) {
		t.Fatalf("sanctions = %v, want %v", ids, want)
	}
}

func TestSanctionWritesRequireAFIAAdmin(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	created := fixture.issueOnAlonso()
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, fixture.teamA)
	writes := []struct{ method, path string }{
		{http.MethodPost, "/sanctions"},
		{http.MethodPut, sanctionPath(created, "")},
		{http.MethodPost, sanctionPath(created, "/revocation")},
		{http.MethodGet, "/sanction-acknowledgements"},
	}

	for _, write := range writes {
		expectStatus(t, h.do(h.browser(), write.method, write.path, fixture.driverSanction(), nil), http.StatusUnauthorized)
		expectStatus(t, h.do(teamAdmin, write.method, write.path, fixture.driverSanction(), nil), http.StatusForbidden)
	}
	expectStatus(t, h.do(h.browser(), http.MethodDelete, sanctionPath(created, ""), nil, nil), http.StatusMethodNotAllowed)
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodGet, "/team/sanction-notices", nil, nil), http.StatusForbidden)
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPost, sanctionAcknowledgementPath(currentVersionID(t, created.body)), nil, nil), http.StatusForbidden)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/team/sanction-notices", nil, nil), http.StatusUnauthorized)
}

func TestFIAIssuesASanctionOnADriverWithDateAndAuthor(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	before := time.Now()
	created := fixture.issueOnAlonso()

	subject := objectOf(t, created.body, "subject")
	if subject["kind"] != "driver" || subject["driverId"] != float64(fixture.alonso) || subject["driverLastName"] != "Alonso" {
		t.Fatalf("subject = %v, want driver Alonso", subject)
	}
	if objectOf(t, created.body, "team")["id"] != float64(fixture.teamA) || created.body["status"] != "active" {
		t.Fatalf("sanction = %v, want an active sanction accountable to team A", created.body)
	}
	expectInstantBetween(t, created.body["issuedAt"], before, time.Now())
	current := objectOf(t, created.body, "current")
	penalty := objectOf(t, current, "penalty")
	if penalty["code"] != "time_penalty" || penalty["value"] != float64(5) || penalty["unit"] != "seconds" {
		t.Fatalf("penalty = %v, want 5 seconds", penalty)
	}
	if objectOf(t, current, "event")["id"] != float64(fixture.race) {
		t.Fatalf("event = %v, want the race", current["event"])
	}

	author := objectOf(t, fixture.followUpOf(created.body["id"]), "sanction")
	versions := listOf(t, fixture.followUpOf(created.body["id"]), "versions")
	if len(versions) != 1 || versions[0]["changedBy"] != fixture.fiaName || author["id"] != created.body["id"] {
		t.Fatalf("follow-up versions = %v, want one issue by %s", versions, fixture.fiaName)
	}
}

func expectInstantBetween(t *testing.T, raw any, from, to time.Time) {
	t.Helper()

	text, _ := raw.(string)
	at, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		t.Fatalf("%q is not a timestamp: %v", text, err)
	}
	if at.Before(from.Add(-clockSkew)) || at.After(to.Add(clockSkew)) {
		t.Fatalf("instant = %s, want between %s and %s", at, from, to)
	}
}

func TestFIAIssuesASanctionOnATeam(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()

	created := fixture.issue(map[string]any{
		"teamId": fixture.teamB, "penaltyType": "fine", "penaltyValue": 25000, "reason": "Liberación insegura desde el box",
	})
	expectStatus(t, created, http.StatusCreated)
	subject := objectOf(t, created.body, "subject")
	if subject["kind"] != "team" || subject["driverId"] != nil || objectOf(t, created.body, "team")["id"] != float64(fixture.teamB) {
		t.Fatalf("sanction = %v, want one on team B", created.body)
	}
	if objectOf(t, created.body, "current")["event"] != nil {
		t.Fatal("a sanction without an event must answer a null event")
	}
}

func TestSanctionIsRejectedWithoutReasonTargetOrValidPenalty(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	otherRace := testdb.NewChampionship(t, h.pool).AddRace(firstRaceStart)
	cases := []struct {
		name   string
		change func(payload map[string]any)
		field  string
	}{
		{"no target", func(p map[string]any) { delete(p, "driverId") }, "target"},
		{"two targets", func(p map[string]any) { p["teamId"] = fixture.teamA }, "target"},
		{"unknown driver", func(p map[string]any) { p["driverId"] = 999999999 }, "driverId"},
		{"unknown team", func(p map[string]any) { delete(p, "driverId"); p["teamId"] = 999999999 }, "teamId"},
		{"no reason", func(p map[string]any) { delete(p, "reason") }, "reason"},
		{"blank reason", func(p map[string]any) { p["reason"] = "    " }, "reason"},
		{"reason too long", func(p map[string]any) { p["reason"] = strings.Repeat("a", 2001) }, "reason"},
		{"no penalty", func(p map[string]any) { delete(p, "penaltyType") }, "penaltyType"},
		{"unknown penalty", func(p map[string]any) { p["penaltyType"] = "points_deduction" }, "penaltyType"},
		{"missing value", func(p map[string]any) { delete(p, "penaltyValue") }, "penaltyValue"},
		{"zero value", func(p map[string]any) { p["penaltyValue"] = 0 }, "penaltyValue"},
		{"negative value", func(p map[string]any) { p["penaltyValue"] = -5 }, "penaltyValue"},
		{"value over the ceiling", func(p map[string]any) { p["penaltyValue"] = 301 }, "penaltyValue"},
		{"value on a penalty without unit", func(p map[string]any) { p["penaltyType"] = "reprimand" }, "penaltyValue"},
		{"fractional value", func(p map[string]any) { p["penaltyValue"] = 2.5 }, "body"},
		{"value as text", func(p map[string]any) { p["penaltyValue"] = "5" }, "body"},
		{"overflowing value", func(p map[string]any) { p["penaltyValue"] = 1e12 }, "body"},
		{"unknown field", func(p map[string]any) { p["issuedAt"] = "2020-01-01T00:00:00Z" }, "body"},
		{"event of another championship", func(p map[string]any) { p["eventId"] = otherRace }, "eventId"},
		{"unknown event", func(p map[string]any) { p["eventId"] = 999999999 }, "eventId"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := fixture.driverSanction()
			tc.change(payload)
			got := fixture.issue(payload)
			expectStatus(t, got, http.StatusBadRequest)
			if got.body["field"] != tc.field {
				t.Fatalf("field = %v, want %s; body %s", got.body["field"], tc.field, got.raw)
			}
		})
	}
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPost, "/sanctions", nil, nil), http.StatusBadRequest)
	expectIDs(t, h.do(h.browser(), http.MethodGet, fixture.listQuery(""), nil, nil))
}

func TestDriverWithoutATeamCannotBeSanctioned(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	payload := fixture.driverSanction()
	payload["driverId"] = fixture.championship.AddDriverSince(fixture.teamA, "Futuro", time.Now().Add(24*time.Hour))

	got := fixture.issue(payload)
	expectStatus(t, got, http.StatusBadRequest)
	if got.body["field"] != "driverId" {
		t.Fatalf("field = %v, want driverId", got.body["field"])
	}
}

func TestDriverSanctionStaysWithTheTeamOfTheDay(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	created := fixture.issueOnAlonso()
	fixture.championship.Transfer(fixture.alonso, fixture.teamB, time.Now().Add(time.Hour))

	detail := h.do(h.browser(), http.MethodGet, sanctionPath(created, ""), nil, nil)
	expectStatus(t, detail, http.StatusOK)
	if objectOf(t, detail.body, "team")["id"] != float64(fixture.teamA) {
		t.Fatalf("team = %v, want team A, the driver's team when sanctioned", detail.body["team"])
	}
}

func TestPublicSeesSanctionsWithoutSigningInAndFiltersThem(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	onDriver := fixture.issueOnAlonso()
	onTeam := fixture.issue(map[string]any{"teamId": fixture.teamB, "penaltyType": "reprimand", "reason": "Conducta antideportiva"})
	expectStatus(t, onTeam, http.StatusCreated)
	revoked := fixture.issueOnAlonso()
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPost, sanctionPath(revoked, "/revocation"), map[string]any{"note": "Apelación aceptada"}, nil), http.StatusOK)
	public := h.browser()
	today := time.Now().UTC().Format(time.DateOnly)
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format(time.DateOnly)

	all := h.do(public, http.MethodGet, fixture.listQuery(""), nil, nil)
	expectIDs(t, all, revoked.body["id"], onTeam.body["id"], onDriver.body["id"])
	if strings.Contains(all.raw, fixture.fiaName) {
		t.Fatal("the public list must not name the FIA administrator")
	}
	expectIDs(t, h.do(public, http.MethodGet, fixture.listQuery(fmt.Sprintf("&driverId=%d&status=active", fixture.alonso)), nil, nil), onDriver.body["id"])
	expectIDs(t, h.do(public, http.MethodGet, fixture.listQuery(fmt.Sprintf("&teamId=%d", fixture.teamB)), nil, nil), onTeam.body["id"])
	expectIDs(t, h.do(public, http.MethodGet, fixture.listQuery(fmt.Sprintf("&teamId=%d&status=all", fixture.teamA)), nil, nil), revoked.body["id"], onDriver.body["id"])
	expectIDs(t, h.do(public, http.MethodGet, fixture.listQuery("&status=revoked"), nil, nil), revoked.body["id"])
	expectIDs(t, h.do(public, http.MethodGet, fixture.listQuery("&from="+today+"&to="+today+"&status=active"), nil, nil), onTeam.body["id"], onDriver.body["id"])
	expectIDs(t, h.do(public, http.MethodGet, fixture.listQuery("&to="+yesterday), nil, nil))

	for _, query := range []string{"&status=vigentes", "&driverId=abc", "&teamId=-1", "&from=2026-13-01", "&from=" + today + "&to=" + yesterday} {
		expectStatus(t, h.do(public, http.MethodGet, fixture.listQuery(query), nil, nil), http.StatusBadRequest)
	}
}

func TestPublicDetailShowsWhatTheFIARecorded(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	created := fixture.issueOnAlonso()

	detail := h.do(h.browser(), http.MethodGet, sanctionPath(created, ""), nil, nil)
	expectStatus(t, detail, http.StatusOK)
	current := objectOf(t, detail.body, "current")
	if current["reason"] != fixture.driverSanction()["reason"] || objectOf(t, current, "penalty")["value"] != float64(5) {
		t.Fatalf("detail = %v, want the recorded reason and penalty", detail.body)
	}
	if versions := listOf(t, detail.body, "versions"); len(versions) != 1 || versions[0]["action"] != "issued" {
		t.Fatalf("versions = %v, want only the issue", versions)
	}
	if strings.Contains(detail.raw, fixture.fiaName) {
		t.Fatal("the public detail must not name the FIA administrator")
	}
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/sanctions/999999999", nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(h.browser(), http.MethodGet, "/sanctions/abc", nil, nil), http.StatusNotFound)
}

func TestEveryTeamOfTheChampionshipIsNotifiedAndConfirms(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	teamA, adminA := h.signedIn(domain.RoleTeamAdmin, fixture.teamA)
	teamB, _ := h.signedIn(domain.RoleTeamAdmin, fixture.teamB)
	created := fixture.issueOnAlonso()

	noticeA := sanctionNoticeFor(t, h.sanctionNotices(teamA), created.body["id"])
	noticeB := sanctionNoticeFor(t, h.sanctionNotices(teamB), created.body["id"])
	if noticeA["status"] != "pending" || noticeA["isOwnTeam"] != true || noticeB["status"] != "pending" || noticeB["isOwnTeam"] != false {
		t.Fatalf("notices = %v / %v, want both pending and only A's own", noticeA, noticeB)
	}
	if h.sanctionNotices(teamA).body["pendingCount"] != float64(1) {
		t.Fatal("team A should have one pending sanction")
	}

	before := time.Now()
	confirmed := h.do(teamA, http.MethodPost, sanctionAcknowledgementPath(noticeA["id"].(float64)), nil, nil)
	expectStatus(t, confirmed, http.StatusOK)
	firstAt := expectAcknowledgedBetween(t, confirmed.body, adminA, before, time.Now())
	again := h.do(teamA, http.MethodPost, sanctionAcknowledgementPath(noticeA["id"].(float64)), nil, nil)
	expectStatus(t, again, http.StatusOK)
	if expectAcknowledgedBetween(t, again.body, adminA, before, time.Now()) != firstAt {
		t.Fatal("confirming again must keep the first confirmation")
	}
	if h.sanctionNotices(teamA).body["pendingCount"] != float64(0) {
		t.Fatal("team A has nothing pending once it confirmed")
	}

	version := listOf(t, fixture.followUpOf(created.body["id"]), "versions")[0]
	if version["confirmedCount"] != float64(1) || len(listOf(t, version, "teams")) != 2 {
		t.Fatalf("follow-up = %v, want 1 of 2 teams confirmed", version)
	}
}

func TestOnlyTeamsOfTheChampionshipActiveAtTheVersionAreNotified(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	retired := fixture.championship.AddTeam("Equipo retirado")
	if _, err := h.pool.Exec(context.Background(), `UPDATE teams SET deactivated_at = now() WHERE id = $1`, retired); err != nil {
		t.Fatalf("deactivate team: %v", err)
	}
	created := fixture.issueOnAlonso()
	latecomer, _ := h.signedIn(domain.RoleTeamAdmin, fixture.championship.AddTeam("Equipo nuevo"))
	foreigner, _ := h.signedIn(domain.RoleTeamAdmin, testdb.NewChampionship(t, h.pool).AddTeam("Otro campeonato"))
	versionID := currentVersionID(t, created.body)

	for _, team := range listOf(t, listOf(t, fixture.followUpOf(created.body["id"]), "versions")[0], "teams") {
		if objectOf(t, team, "team")["id"] == float64(retired) {
			t.Fatal("a deactivated team must not be asked to confirm")
		}
	}
	for _, outsider := range []*http.Client{latecomer, foreigner} {
		if len(listOf(t, h.sanctionNotices(outsider).body, "notices")) != 0 {
			t.Fatal("a team not notified of the version must not see it in its notices")
		}
		expectStatus(t, h.do(outsider, http.MethodPost, sanctionAcknowledgementPath(versionID), nil, nil), http.StatusNotFound)
	}
	teamA, _ := h.signedIn(domain.RoleTeamAdmin, fixture.teamA)
	expectStatus(t, h.do(teamA, http.MethodPost, "/team/sanction-notices/999999999/acknowledgement", nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(teamA, http.MethodPost, "/team/sanction-notices/abc/acknowledgement", nil, nil), http.StatusNotFound)
}

func TestAmendmentAsksEveryTeamToConfirmAgain(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	teamA, _ := h.signedIn(domain.RoleTeamAdmin, fixture.teamA)
	created := fixture.issueOnAlonso()
	issuedVersion := currentVersionID(t, created.body)
	expectStatus(t, h.do(teamA, http.MethodPost, sanctionAcknowledgementPath(issuedVersion), nil, nil), http.StatusOK)

	amended := h.do(fixture.fiaAdmin, http.MethodPut, sanctionPath(created, ""), fixture.amendment("Revisión de la telemetría"), nil)
	expectStatus(t, amended, http.StatusOK)
	current := objectOf(t, amended.body, "current")
	if current["action"] != "amended" || current["note"] != "Revisión de la telemetría" || objectOf(t, current, "penalty")["value"] != float64(10) {
		t.Fatalf("current = %v, want the amendment", current)
	}
	if len(listOf(t, amended.body, "versions")) != 2 || amended.body["issuedAt"] != created.body["issuedAt"] {
		t.Fatalf("detail = %v, want two versions and the original issue date", amended.body)
	}

	notice := sanctionNoticeFor(t, h.sanctionNotices(teamA), created.body["id"])
	if notice["status"] != "pending" || notice["id"] != currentVersionID(t, amended.body) {
		t.Fatalf("notice = %v, want the amendment pending", notice)
	}
	expectStatus(t, h.do(teamA, http.MethodPost, sanctionAcknowledgementPath(issuedVersion), nil, nil), http.StatusConflict)

	unchanged := h.do(fixture.fiaAdmin, http.MethodPut, sanctionPath(created, ""), fixture.amendment("Sin cambios reales"), nil)
	expectStatus(t, unchanged, http.StatusOK)
	if len(listOf(t, unchanged.body, "versions")) != 2 {
		t.Fatal("submitting the standing terms must not record a version")
	}
	followUp := listOf(t, fixture.followUpOf(created.body["id"]), "versions")
	if len(followUp) != 2 || followUp[0]["confirmedCount"] != float64(0) || followUp[1]["confirmedCount"] != float64(1) {
		t.Fatalf("follow-up = %v, want the amendment unconfirmed and the issue confirmed once", followUp)
	}
}

func TestAmendmentIsValidated(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	created := fixture.issueOnAlonso()
	withoutNote := fixture.amendment("  ")
	retargeted := fixture.amendment("Otro destinatario")
	retargeted["teamId"] = fixture.teamB

	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPut, sanctionPath(created, ""), withoutNote, nil), http.StatusBadRequest)
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPut, sanctionPath(created, ""), retargeted, nil), http.StatusBadRequest)
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPut, "/sanctions/999999999", fixture.amendment("Nota"), nil), http.StatusNotFound)
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPost, "/sanctions/999999999/revocation", map[string]any{"note": "Nota"}, nil), http.StatusNotFound)
}

func TestRevocationStaysOnTheRecordAndIsNotified(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	teamB, _ := h.signedIn(domain.RoleTeamAdmin, fixture.teamB)
	created := fixture.issueOnAlonso()

	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPost, sanctionPath(created, "/revocation"), map[string]any{"note": " "}, nil), http.StatusBadRequest)
	revoked := h.do(fixture.fiaAdmin, http.MethodPost, sanctionPath(created, "/revocation"), map[string]any{"note": "Apelación aceptada"}, nil)
	expectStatus(t, revoked, http.StatusOK)

	detail := h.do(h.browser(), http.MethodGet, sanctionPath(created, ""), nil, nil)
	expectStatus(t, detail, http.StatusOK)
	current := objectOf(t, detail.body, "current")
	if detail.body["status"] != "revoked" || current["action"] != "revoked" || current["note"] != "Apelación aceptada" {
		t.Fatalf("detail = %v, want the revocation with its note", detail.body)
	}
	if objectOf(t, current, "penalty")["code"] != "time_penalty" || current["reason"] != fixture.driverSanction()["reason"] {
		t.Fatal("the revocation must keep the terms it annuls")
	}
	notice := sanctionNoticeFor(t, h.sanctionNotices(teamB), created.body["id"])
	if notice["status"] != "pending" || objectOf(t, notice, "sanction")["status"] != "revoked" {
		t.Fatalf("notice = %v, want the revocation pending", notice)
	}
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPost, sanctionPath(created, "/revocation"), map[string]any{"note": "Otra vez"}, nil), http.StatusConflict)
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPut, sanctionPath(created, ""), fixture.amendment("Tarde"), nil), http.StatusConflict)
}
