package dbtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/tests/testdb"
)

// TestScoreForAnotherCategoryIsRejected covers RC002: the team a result credits is
// derived from the assignment history, and that team must race in the category of
// the race. Seeded event 1 is a Formula 1 race, so a driver assigned to a Formula
// 2 team cannot be scored there. The write declares an actor first so the audit
// guard steps aside and the category rule is the one that fires.
func TestScoreForAnotherCategoryIsRejected(t *testing.T) {
	pool := testdb.Pool(t)
	driverID := insertTestDriver(t, pool)
	assignDriver(t, pool, driverID, formulaTwoTeamID, seasonStart)

	err := testdb.RunAsScoreActor(context.Background(), pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(context.Background(),
			`INSERT INTO score_entries (event_id, driver_id, points) VALUES ($1, $2, 10)`,
			seededRacedEventID, driverID)
		return err
	})

	code, isGuard := db.GuardViolation(err)
	if !isGuard || code != "RC002" {
		t.Fatalf("scoring a driver of another category = (%q, %t), want (\"RC002\", true)", code, isGuard)
	}
}

// TestScoreChangeWithoutActorIsRejected covers RC005: the audit guard refuses any
// write to score_entries that did not declare who made it.
func TestScoreChangeWithoutActorIsRejected(t *testing.T) {
	pool := testdb.Pool(t)
	driverID := insertTestDriver(t, pool)
	assignDriver(t, pool, driverID, formulaOneTeamID, seasonStart)

	_, err := pool.Exec(context.Background(),
		`INSERT INTO score_entries (event_id, driver_id, points) VALUES ($1, $2, 10)`,
		seededRacedEventID, driverID)

	assertGuardCode(t, err, "RC005")
}
