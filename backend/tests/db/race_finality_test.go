package dbtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/tests/testdb"
)

// TestRacedRaceIsFinal covers RC001 from every side: once a race has results, its
// date, its round link, and the round's championship and status are frozen, since
// any of them would move the results to another date, standings table or track.
// The round number and the Grand Prix name stay editable.
func TestRacedRaceIsFinal(t *testing.T) {
	pool := testdb.Pool(t)

	frozen := []struct {
		what  string
		query string
		arg   int64
	}{
		{"moving the date", `UPDATE events SET starts_at = starts_at + interval '1 day' WHERE id = $1`, seededRacedEventID},
		{"taking the race off the calendar", `UPDATE events SET deactivated_at = now() WHERE id = $1`, seededRacedEventID},
		{"moving the race to another round", `UPDATE races SET round_id = 2 WHERE event_id = $1`, seededRacedEventID},
		{"changing the round championship", `UPDATE rounds SET category_id = 2 WHERE id = $1`, seededRacedRoundID},
		{"cancelling the round", `UPDATE rounds SET deactivated_at = now() WHERE id = $1`, seededRacedRoundID},
	}

	for _, c := range frozen {
		code, isGuard := db.GuardViolation(execRolledBack(t, pool, c.query, c.arg))
		if !isGuard || code != "RC001" {
			t.Errorf("%s = (%q, %t), want (\"RC001\", true)", c.what, code, isGuard)
		}
	}

	for _, query := range []string{
		`UPDATE rounds SET number = number + 1000 WHERE id = $1`,
		`UPDATE rounds SET gp_name = gp_name || ' (edited)' WHERE id = $1`,
	} {
		if err := execRolledBack(t, pool, query, seededRacedRoundID); err != nil {
			t.Errorf("editing the label of a raced round should be allowed: %v", err)
		}
	}
}

// execRolledBack runs a statement in a transaction that is always rolled back, so
// the seeded data a guard test leans on is never changed.
func execRolledBack(t *testing.T, pool *pgxpool.Pool, query string, args ...any) error {
	t.Helper()

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, query, args...)
	return err
}
