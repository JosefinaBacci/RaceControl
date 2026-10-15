package dbtest

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/tests/testdb"
)

// TestDeletingAnEventChildIsAllowed locks in the 000020 decision: the mirror
// guard that stopped an event from losing its child is gone. The application
// never removes a child on its own (a child is pinned by deactivating its event),
// so the delete now goes through instead of aborting the commit on a deferred
// constraint trigger.
func TestDeletingAnEventChildIsAllowed(t *testing.T) {
	pool := testdb.Pool(t)
	eventID := insertTyreTestEvent(t, pool)

	if _, err := pool.Exec(context.Background(),
		`DELETE FROM tyre_tests WHERE event_id = $1`, eventID); err != nil {
		t.Fatalf("deleting the child of event %d should be allowed: %v", eventID, err)
	}
}

// insertTyreTestEvent builds the smallest valid event with a child. Both rows go
// in one transaction because assert_event_has_child runs deferred at commit: an
// event committed without its child would fail the other half of the invariant.
func insertTyreTestEvent(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()

	ctx := context.Background()
	var eventID int64

	err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`INSERT INTO events (kind, starts_at, ends_at)
			 VALUES ('tyre_test', $1, $2) RETURNING id`,
			seasonStart, seasonStart.Add(2*time.Hour)).Scan(&eventID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx,
			`INSERT INTO tyre_tests (event_id, season_id, category_id, circuit_id)
			 VALUES ($1, $2, $3, $4)`,
			eventID, demoSeasonID, formulaOneCategoryID, seededCircuitID)
		return err
	})
	if err != nil {
		t.Fatalf("insert tyre test event: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM tyre_tests WHERE event_id = $1`, eventID)
		if _, err := pool.Exec(ctx, `DELETE FROM events WHERE id = $1`, eventID); err != nil {
			t.Errorf("clean up event %d: %v", eventID, err)
		}
	})

	return eventID
}
