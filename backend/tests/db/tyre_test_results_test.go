package dbtest

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/tests/testdb"
)

const recordedEventFreezeCode = "RC020"

var tyreTestDay = time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

type tyreTestFixture struct {
	pool     *pgxpool.Pool
	eventID  int64
	resultID int64
	actorID  int64
}

func newTyreTestFixture(t *testing.T) tyreTestFixture {
	t.Helper()

	pool := testdb.Pool(t)
	championship := testdb.NewChampionship(t, pool)
	eventID := championship.AddTyreTest(tyreTestDay.Add(9*time.Hour), tyreTestDay.Add(17*time.Hour))
	actorID := insertFIAAdmin(t, pool)
	resultID := championship.TyreResult(eventID, "c3", tyreTestDay, actorID)
	return tyreTestFixture{pool: pool, eventID: eventID, resultID: resultID, actorID: actorID}
}

func insertFIAAdmin(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()

	var userID int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash, role) VALUES ($1, 'not-a-real-hash', 'fia_admin') RETURNING id`,
		testdb.UniqueUsername(t)).Scan(&userID)
	if err != nil {
		t.Fatalf("insert fia admin: %v", err)
	}
	return userID
}

// correct inserts a correction of supersededID for the given compound, the way
// the service does, inside a rolled-back transaction.
func (f tyreTestFixture) correct(t *testing.T, supersededID int64, compoundCode string, note any) error {
	t.Helper()
	return execRolledBack(t, f.pool,
		`INSERT INTO tyre_test_results (event_id, compound_id, conducted_on, passed, supersedes_id, correction_note, recorded_by)
		 SELECT $1, id, $3, true, $4, $5, $6 FROM tyre_compounds WHERE code = $2`,
		f.eventID, compoundCode, tyreTestDay, supersededID, note, f.actorID)
}

func TestTyreTestResultsAreAppendOnly(t *testing.T) {
	fixture := newTyreTestFixture(t)

	for _, query := range []string{
		`UPDATE tyre_test_results SET passed = false WHERE id = $1`,
		`DELETE FROM tyre_test_results WHERE id = $1`,
	} {
		if err := execRolledBack(t, fixture.pool, query, fixture.resultID); err == nil {
			t.Errorf("%q should be rejected: a tyre test result is an official record", query)
		}
	}
}

func TestTyreTestKeepsOneOriginalPerCompound(t *testing.T) {
	fixture := newTyreTestFixture(t)

	err := execRolledBack(t, fixture.pool,
		`INSERT INTO tyre_test_results (event_id, compound_id, conducted_on, passed, recorded_by)
		 SELECT $1, id, $2, true, $3 FROM tyre_compounds WHERE code = 'c3'`,
		fixture.eventID, tyreTestDay, fixture.actorID)
	if constraint, isViolation := db.UniqueViolation(err); !isViolation || constraint != "tyre_test_results_one_original" {
		t.Fatalf("second original of c3 = %v, want the one_original violation", err)
	}
}

func TestTyreTestCorrectionChainIsLinear(t *testing.T) {
	fixture := newTyreTestFixture(t)

	if err := fixture.correct(t, fixture.resultID, "c3", "motivo"); err != nil {
		t.Fatalf("a correction of the same compound should be stored: %v", err)
	}
	if err := fixture.correct(t, fixture.resultID, "c4", "motivo"); err == nil {
		t.Error("a correction cannot move the result to another compound")
	}
	if err := fixture.correct(t, fixture.resultID, "c3", nil); err == nil {
		t.Error("a correction without its note should be rejected")
	}
}

func TestTyreTestResultIsSupersededOnce(t *testing.T) {
	fixture := newTyreTestFixture(t)
	ctx := context.Background()

	tx, err := fixture.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	insertCorrection := `INSERT INTO tyre_test_results (event_id, compound_id, conducted_on, passed, supersedes_id, correction_note, recorded_by)
		SELECT $1, id, $2, true, $3, 'motivo', $4 FROM tyre_compounds WHERE code = 'c3'`
	if _, err := tx.Exec(ctx, insertCorrection, fixture.eventID, tyreTestDay, fixture.resultID, fixture.actorID); err != nil {
		t.Fatalf("first correction: %v", err)
	}
	_, err = tx.Exec(ctx, insertCorrection, fixture.eventID, tyreTestDay, fixture.resultID, fixture.actorID)
	if constraint, isViolation := db.UniqueViolation(err); !isViolation || constraint != "tyre_test_results_supersedes_once" {
		t.Fatalf("second correction of the same row = %v, want the supersedes_once violation", err)
	}
}

func TestTestedTyreTestIsFinal(t *testing.T) {
	fixture := newTyreTestFixture(t)

	frozen := map[string]string{
		"moving the date":        `UPDATE events SET starts_at = starts_at - interval '1 hour' WHERE id = $1`,
		"cancelling the test":    `UPDATE events SET deactivated_at = now() WHERE id = $1`,
		"moving it to a circuit": `UPDATE tyre_tests SET circuit_id = circuit_id + 1 WHERE event_id = $1`,
	}
	for what, query := range frozen {
		code, isGuard := db.GuardViolation(execRolledBack(t, fixture.pool, query, fixture.eventID))
		if !isGuard || code != recordedEventFreezeCode {
			t.Errorf("%s = (%q, %t), want (%q, true)", what, code, isGuard, recordedEventFreezeCode)
		}
	}

	resent := `UPDATE events SET starts_at = starts_at, deactivated_at = deactivated_at WHERE id = $1`
	if err := execRolledBack(t, fixture.pool, resent, fixture.eventID); err != nil {
		t.Errorf("re-sending the same values should pass: %v", err)
	}
}

func TestUntestedTyreTestStaysEditable(t *testing.T) {
	pool := testdb.Pool(t)
	eventID := testdb.NewChampionship(t, pool).AddTyreTest(tyreTestDay.Add(9*time.Hour), tyreTestDay.Add(17*time.Hour))

	if err := execRolledBack(t, pool, `UPDATE events SET deactivated_at = now() WHERE id = $1`, eventID); err != nil {
		t.Fatalf("a tyre test without results should still be cancellable: %v", err)
	}
}
