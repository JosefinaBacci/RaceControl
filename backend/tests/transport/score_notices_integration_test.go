package transporttest

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/tests/testdb"
)

// clockSkew tolerates the gap between the test clock and the database's now().
const clockSkew = 5 * time.Second

// noticeFixture is a championship whose race is raced by Alonso (team A) and
// Castro (team B), plus the FIA administrator who loads its scores.
type noticeFixture struct {
	championship *testdb.Championship
	race         int64
	teamA        int64
	teamB        int64
	alonso       int64
	castro       int64
	actorID      int64
	actorName    string
}

func (h *harness) newNoticeFixture() noticeFixture {
	h.t.Helper()

	championship := testdb.NewChampionship(h.t, h.pool)
	teamA := championship.AddTeam("Equipo A")
	teamB := championship.AddTeam("Equipo B")
	actorName := h.createUser(domain.RoleFIAAdmin, 0)
	return noticeFixture{
		championship: championship,
		race:         championship.AddRace(firstRaceStart),
		teamA:        teamA,
		teamB:        teamB,
		alonso:       championship.AddDriver(teamA, "Alonso"),
		castro:       championship.AddDriver(teamB, "Castro"),
		actorID:      testdb.UserID(h.t, h.pool, actorName),
		actorName:    actorName,
	}
}

func (f noticeFixture) publish(lines ...testdb.PointsChange) int64 {
	return f.championship.RecordChange(testdb.ScoreChange{
		EventID: f.race, ActorID: f.actorID, ChangedAt: firstRaceStart.Add(3 * time.Hour), Lines: lines,
	})
}

func (f noticeFixture) correct(lines ...testdb.PointsChange) int64 {
	return f.championship.RecordChange(testdb.ScoreChange{
		EventID: f.race, ActorID: f.actorID, ChangedAt: firstRaceStart.Add(48 * time.Hour), Lines: lines,
	})
}

func pointsOf(value float64) *float64 {
	return &value
}

func (h *harness) teamNotices(client *http.Client) response {
	h.t.Helper()

	got := h.do(client, http.MethodGet, "/team/score-notices", nil, nil)
	expectStatus(h.t, got, http.StatusOK)
	return got
}

func acknowledgementPath(changeLogID int64) string {
	return fmt.Sprintf("/team/score-notices/%d/acknowledgement", changeLogID)
}

func noticeWithID(t *testing.T, notices []map[string]any, changeLogID int64) map[string]any {
	t.Helper()

	for _, notice := range notices {
		if notice["id"] == float64(changeLogID) {
			return notice
		}
	}
	t.Fatalf("no notice for change %d in %v", changeLogID, notices)
	return nil
}

func expectAcknowledgedBetween(t *testing.T, notice map[string]any, username string, from, to time.Time) time.Time {
	t.Helper()

	if notice["status"] != "confirmed" || notice["acknowledgedBy"] != username {
		t.Fatalf("notice = %v, want confirmed by %s", notice, username)
	}
	raw, _ := notice["acknowledgedAt"].(string)
	at, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		t.Fatalf("acknowledgedAt %q is not a timestamp: %v", raw, err)
	}
	if at.Before(from.Add(-clockSkew)) || at.After(to.Add(clockSkew)) {
		t.Fatalf("acknowledgedAt = %s, want between %s and %s", at, from, to)
	}
	return at
}

func TestScoreNoticesRequireATeamAdmin(t *testing.T) {
	h := newHarness(t)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/team/score-notices", nil, nil), http.StatusUnauthorized)
	expectStatus(t, h.do(h.browser(), http.MethodPost, acknowledgementPath(1), nil, nil), http.StatusUnauthorized)
	expectStatus(t, h.do(fiaAdmin, http.MethodGet, "/team/score-notices", nil, nil), http.StatusForbidden)
	expectStatus(t, h.do(fiaAdmin, http.MethodPost, acknowledgementPath(1), nil, nil), http.StatusForbidden)
}

func TestScoreChangesRequireAFIAAdmin(t *testing.T) {
	h := newHarness(t)
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, ferrariTeamID)

	expectStatus(t, h.do(h.browser(), http.MethodGet, "/score-changes", nil, nil), http.StatusUnauthorized)
	expectStatus(t, h.do(teamAdmin, http.MethodGet, "/score-changes", nil, nil), http.StatusForbidden)
}

func TestTeamAdminIsNotifiedOfAPublicationAndConfirmsIt(t *testing.T) {
	h := newHarness(t)
	fixture := h.newNoticeFixture()
	teamAdmin, username := h.signedIn(domain.RoleTeamAdmin, fixture.teamA)
	publication := fixture.publish(
		testdb.PointsChange{DriverID: fixture.alonso, New: 25},
		testdb.PointsChange{DriverID: fixture.castro, New: 18},
	)

	board := h.teamNotices(teamAdmin)
	if board.body["pendingCount"] != float64(1) {
		t.Fatalf("pendingCount = %v, want 1", board.body["pendingCount"])
	}
	notice := noticeWithID(t, listOf(t, board.body, "notices"), publication)
	if notice["kind"] != "published" || notice["status"] != "pending" || notice["acknowledgedAt"] != nil {
		t.Fatalf("notice = %v, want a pending publication", notice)
	}
	items := listOf(t, notice, "items")
	if len(items) != 1 || items[0]["driverId"] != float64(fixture.alonso) || items[0]["oldPoints"] != nil || items[0]["newPoints"] != float64(25) {
		t.Fatalf("items = %v, want only the team's own driver going from nothing to 25", items)
	}

	before := time.Now()
	confirmed := h.do(teamAdmin, http.MethodPost, acknowledgementPath(publication), nil, nil)
	expectStatus(t, confirmed, http.StatusOK)
	firstAcknowledgement := expectAcknowledgedBetween(t, confirmed.body, username, before, time.Now())

	board = h.teamNotices(teamAdmin)
	if board.body["pendingCount"] != float64(0) {
		t.Fatalf("pendingCount = %v after confirming, want 0", board.body["pendingCount"])
	}
	expectAcknowledgedBetween(t, noticeWithID(t, listOf(t, board.body, "notices"), publication), username, before, time.Now())

	again := h.do(teamAdmin, http.MethodPost, acknowledgementPath(publication), nil, nil)
	expectStatus(t, again, http.StatusOK)
	if repeated := expectAcknowledgedBetween(t, again.body, username, before, time.Now()); !repeated.Equal(firstAcknowledgement) {
		t.Fatalf("confirming again moved the timestamp from %s to %s", firstAcknowledgement, repeated)
	}
}

func TestACorrectionAsksForANewConfirmation(t *testing.T) {
	h := newHarness(t)
	fixture := h.newNoticeFixture()
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, fixture.teamA)
	publication := fixture.publish(testdb.PointsChange{DriverID: fixture.alonso, New: 25})
	expectStatus(t, h.do(teamAdmin, http.MethodPost, acknowledgementPath(publication), nil, nil), http.StatusOK)

	correction := fixture.correct(testdb.PointsChange{DriverID: fixture.alonso, Old: pointsOf(25), New: 18})

	board := h.teamNotices(teamAdmin)
	notices := listOf(t, board.body, "notices")
	if len(notices) != 2 || notices[0]["id"] != float64(correction) {
		t.Fatalf("notices = %v, want the correction first", notices)
	}
	if notices[0]["kind"] != "corrected" || notices[0]["status"] != "pending" {
		t.Fatalf("correction = %v, want a pending correction", notices[0])
	}
	if notices[1]["kind"] != "published" || notices[1]["status"] != "confirmed" {
		t.Fatalf("publication = %v, want the confirmed publication", notices[1])
	}
	if board.body["pendingCount"] != float64(1) {
		t.Fatalf("pendingCount = %v, want 1", board.body["pendingCount"])
	}
}

func TestTeamAdminCannotSeeOrConfirmAnotherTeamsNotice(t *testing.T) {
	h := newHarness(t)
	fixture := h.newNoticeFixture()
	teamAdmin, _ := h.signedIn(domain.RoleTeamAdmin, fixture.teamA)
	foreign := fixture.publish(testdb.PointsChange{DriverID: fixture.castro, New: 18})

	if notices := listOf(t, h.teamNotices(teamAdmin).body, "notices"); len(notices) != 0 {
		t.Fatalf("notices = %v, want none: the change only touched team B", notices)
	}
	expectStatus(t, h.do(teamAdmin, http.MethodPost, acknowledgementPath(foreign), nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(teamAdmin, http.MethodPost, acknowledgementPath(999_999_999), nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(teamAdmin, http.MethodPost, "/team/score-notices/not-a-number/acknowledgement", nil, nil), http.StatusNotFound)
}

func TestFIAAdminSeesWhichTeamsConfirmed(t *testing.T) {
	h := newHarness(t)
	fixture := h.newNoticeFixture()
	teamAdmin, username := h.signedIn(domain.RoleTeamAdmin, fixture.teamA)
	fiaAdmin, _ := h.signedIn(domain.RoleFIAAdmin, 0)
	publication := fixture.publish(
		testdb.PointsChange{DriverID: fixture.alonso, New: 25},
		testdb.PointsChange{DriverID: fixture.castro, New: 18},
	)
	before := time.Now()
	expectStatus(t, h.do(teamAdmin, http.MethodPost, acknowledgementPath(publication), nil, nil), http.StatusOK)

	got := h.do(fiaAdmin, http.MethodGet, "/score-changes", nil, nil)
	expectStatus(t, got, http.StatusOK)
	change := noticeWithID(t, listOf(t, got.body, "changes"), publication)
	if change["kind"] != "published" || change["changedBy"] != fixture.actorName {
		t.Fatalf("change = %v, want a publication by %s", change, fixture.actorName)
	}

	teams := listOf(t, change, "teams")
	if len(teams) != 2 {
		t.Fatalf("teams = %v, want both affected teams", teams)
	}
	for _, team := range teams {
		switch team["teamId"] {
		case float64(fixture.teamA):
			expectAcknowledgedBetween(t, team, username, before, time.Now())
		case float64(fixture.teamB):
			if team["status"] != "pending" || team["acknowledgedAt"] != nil {
				t.Fatalf("team B = %v, want pending", team)
			}
		default:
			t.Fatalf("unexpected team %v", team)
		}
	}
}
