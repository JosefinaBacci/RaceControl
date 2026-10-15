package dbtest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/tests/testdb"
)

// TestEventWithoutAChildIsRejected locks the half of "exactly one child" that
// stays in the database after 000020: assert_event_has_child runs deferred, so an
// event committed on its own fails instead of hanging from nothing.
func TestEventWithoutAChildIsRejected(t *testing.T) {
	pool := testdb.Pool(t)

	_, err := pool.Exec(context.Background(),
		`INSERT INTO events (kind, starts_at, ends_at) VALUES ('race', $1, $2)`,
		seasonStart, seasonStart.Add(2*time.Hour))

	if err == nil || !strings.Contains(err.Error(), "has no child row") {
		t.Fatalf("event without a child = %v, want the deferred child guard to reject it", err)
	}
}

// TestSecondChildOfTheSameKindIsRejected: the child's primary key is its event, so
// an event cannot hang two races. "At most one child" needs no trigger.
func TestSecondChildOfTheSameKindIsRejected(t *testing.T) {
	pool := testdb.Pool(t)
	eventID, _ := createRaceEvent(t, pool)
	otherRoundID := createRound(t, pool)

	_, err := pool.Exec(context.Background(),
		`INSERT INTO races (event_id, round_id) VALUES ($1, $2)`, eventID, otherRoundID)

	if _, isViolation := db.UniqueViolation(err); !isViolation {
		t.Fatalf("second race for event %d = %v, want a unique violation on the child key", eventID, err)
	}
}

// TestChildOfAnotherKindIsRejected: the composite FK pins a child to an event of
// its own kind, so a race cannot hang from a tyre test.
func TestChildOfAnotherKindIsRejected(t *testing.T) {
	pool := testdb.Pool(t)
	eventID := insertTyreTestEvent(t, pool)
	roundID := createRound(t, pool)

	_, err := pool.Exec(context.Background(),
		`INSERT INTO races (event_id, round_id) VALUES ($1, $2)`, eventID, roundID)

	if err == nil || !strings.Contains(err.Error(), "races_event_fkey") {
		t.Fatalf("race on a tyre test event = %v, want the kind FK to reject it", err)
	}
}

// TestEventCannotChangeKindWhileItHasAChild: the kind is part of the child's FK,
// so an event cannot switch to a type whose child it does not have.
func TestEventCannotChangeKindWhileItHasAChild(t *testing.T) {
	pool := testdb.Pool(t)
	eventID, _ := createRaceEvent(t, pool)

	_, err := pool.Exec(context.Background(),
		`UPDATE events SET kind = 'tyre_test' WHERE id = $1`, eventID)

	if err == nil || !strings.Contains(err.Error(), "races_event_fkey") {
		t.Fatalf("changing the kind of event %d = %v, want the kind FK to reject it", eventID, err)
	}
}

// createRaceEvent builds a race event with its child in one transaction, the way
// the application does it, and returns both ids.
func createRaceEvent(t *testing.T, pool *pgxpool.Pool) (eventID, roundID int64) {
	t.Helper()

	ctx := context.Background()
	roundID = createRound(t, pool)

	err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`INSERT INTO events (kind, starts_at, ends_at)
			 VALUES ('race', $1, $2) RETURNING id`,
			seasonStart, seasonStart.Add(2*time.Hour)).Scan(&eventID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx,
			`INSERT INTO races (event_id, round_id) VALUES ($1, $2)`, eventID, roundID)
		return err
	})
	if err != nil {
		t.Fatalf("create race event: %v", err)
	}

	t.Cleanup(func() {
		if err := testdb.RunAsScoreActor(ctx, pool, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM score_entries WHERE event_id = $1`, eventID)
			return err
		}); err != nil {
			t.Errorf("clean up score entries of event %d: %v", eventID, err)
		}
		_, _ = pool.Exec(ctx, `DELETE FROM races WHERE event_id = $1`, eventID)
		if _, err := pool.Exec(ctx, `DELETE FROM events WHERE id = $1`, eventID); err != nil {
			t.Errorf("clean up race event %d: %v", eventID, err)
		}
	})
	return eventID, roundID
}

// createRound inserts an active round with a free number, so the tournament never
// holds two active rounds with the same one.
func createRound(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()

	var roundID int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO rounds (season_id, category_id, number, circuit_id, gp_name)
		 SELECT $1, $2, COALESCE(max(number), 0) + 1, $3, 'Test round'
		 FROM rounds WHERE season_id = $1 AND category_id = $2
		 RETURNING id`,
		demoSeasonID, formulaOneCategoryID, seededCircuitID).Scan(&roundID)
	if err != nil {
		t.Fatalf("insert round: %v", err)
	}

	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM rounds WHERE id = $1`, roundID); err != nil {
			t.Errorf("clean up round %d: %v", roundID, err)
		}
	})
	return roundID
}
