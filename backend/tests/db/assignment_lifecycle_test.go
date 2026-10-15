package dbtest

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/tests/testdb"
)

// TestAssignmentCanBeReleased covers the "baja": a row without a team ends the
// membership period, so the team derived after that date is no one.
func TestAssignmentCanBeReleased(t *testing.T) {
	pool := testdb.Pool(t)
	driverID := insertTestDriver(t, pool)
	assignDriver(t, pool, driverID, formulaOneTeamID, seasonStart)

	release := seasonStart.AddDate(0, 6, 0)
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO driver_team_assignments (driver_id, team_id, effective_from, is_starter)
		 VALUES ($1, NULL, $2, false)`,
		driverID, release); err != nil {
		t.Fatalf("release the driver: %v", err)
	}

	if team := assignedTeam(t, pool, driverID, release.Add(time.Hour)); team != nil {
		t.Fatalf("assigned team after the release = %v, want none", *team)
	}
}

// TestAssignmentMustExtendHistory covers the first half of RC004: a new row has to
// start after the driver's last assignment, so the derived team of a past date
// never changes.
func TestAssignmentMustExtendHistory(t *testing.T) {
	pool := testdb.Pool(t)
	driverID := insertTestDriver(t, pool)
	assignDriver(t, pool, driverID, formulaOneTeamID, midSeason)

	// earlier than the first assignment, so it would rewrite the derived team of
	// dates between the two rows
	_, err := pool.Exec(context.Background(),
		`INSERT INTO driver_team_assignments (driver_id, team_id, effective_from, is_starter)
		 VALUES ($1, $2, $3, true)`,
		driverID, formulaTwoTeamID, seasonStart)
	assertGuardCode(t, err, "RC004")
}

// TestAssignmentCannotRewriteARacedRace covers the second half of RC004: a driver
// who already scored cannot be reassigned for a date at or before that race, since
// the score entry already stored the team they raced for.
func TestAssignmentCannotRewriteARacedRace(t *testing.T) {
	pool := testdb.Pool(t)
	driverID := insertTestDriver(t, pool)
	assignDriver(t, pool, driverID, formulaOneTeamID, seasonStart)
	insertScoreEntry(t, pool, driverID, seededRacedEventID)

	// Seeded event 1 is raced on 2026-10-04, between the first assignment and the
	// next one the test tries to insert.
	_, err := pool.Exec(context.Background(),
		`INSERT INTO driver_team_assignments (driver_id, team_id, effective_from, is_starter)
		 VALUES ($1, $2, $3, true)`,
		driverID, formulaTwoTeamID, midSeason)
	assertGuardCode(t, err, "RC004")
}

// TestAssignmentAfterTheLastRaceIsAccepted is the mirror of the guard: extending
// the history past everything the driver raced is allowed.
func TestAssignmentAfterTheLastRaceIsAccepted(t *testing.T) {
	pool := testdb.Pool(t)
	driverID := insertTestDriver(t, pool)
	assignDriver(t, pool, driverID, formulaOneTeamID, seasonStart)
	insertScoreEntry(t, pool, driverID, seededRacedEventID)

	later := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO driver_team_assignments (driver_id, team_id, effective_from, is_starter)
		 VALUES ($1, $2, $3, true)`,
		driverID, formulaTwoTeamID, later); err != nil {
		t.Fatalf("an assignment after the last race should be accepted: %v", err)
	}
}

func assignedTeam(t *testing.T, pool *pgxpool.Pool, driverID int64, at time.Time) *int64 {
	t.Helper()

	var teamID *int64
	if err := pool.QueryRow(context.Background(),
		`SELECT assigned_team($1, $2)`, driverID, at).Scan(&teamID); err != nil {
		t.Fatalf("assigned_team(%d): %v", driverID, err)
	}
	return teamID
}

func insertScoreEntry(t *testing.T, pool *pgxpool.Pool, driverID, eventID int64) {
	t.Helper()

	_, err := pool.Exec(context.Background(),
		`INSERT INTO score_entries (event_id, driver_id, points) VALUES ($1, $2, 10)`,
		eventID, driverID)
	if err != nil {
		t.Fatalf("insert score entry: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(),
			`DELETE FROM score_entries WHERE driver_id = $1`, driverID); err != nil {
			t.Errorf("clean up score entries: %v", err)
		}
	})
}

func assertGuardCode(t *testing.T, err error, want string) {
	t.Helper()

	code, isGuard := db.GuardViolation(err)
	if !isGuard || code != want {
		t.Fatalf("guard = (%q, %t), want (%q, true)", code, isGuard, want)
	}
}
