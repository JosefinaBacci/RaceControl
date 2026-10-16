package dbtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/tests/testdb"
)

// Seeded driver 1 races for Ferrari (team 1) on the seeded race, so a change that
// lists only them touches Ferrari and no one else. Team 2 is Red Bull.
const (
	seededFerrariDriverID = 1
	redBullTeamID         = 2
)

// TestAcknowledgementTakesTheTeamOfTheUser covers the derivation of 000023: the
// team stored is the acknowledger's, whatever team_id the writer sent.
func TestAcknowledgementTakesTheTeamOfTheUser(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		changeLogID := insertFerrariChange(t, tx)
		ferrariAdmin := insertUser(t, tx, "team_admin", formulaOneTeamID)

		var storedTeamID int64
		err := tx.QueryRow(context.Background(),
			`INSERT INTO score_acknowledgements (change_log_id, team_id, acknowledged_by)
			 VALUES ($1, $2, $3) RETURNING team_id`,
			changeLogID, redBullTeamID, ferrariAdmin).Scan(&storedTeamID)
		if err != nil {
			t.Fatalf("a Ferrari administrator should acknowledge a Ferrari change: %v", err)
		}
		if storedTeamID != formulaOneTeamID {
			t.Fatalf("team_id = %d, want the acknowledger's team %d", storedTeamID, formulaOneTeamID)
		}
	})
}

// TestFIAAdminCannotAcknowledge: a FIA administrator has no team, so no change can
// ever be theirs to confirm.
func TestFIAAdminCannotAcknowledge(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		changeLogID := insertFerrariChange(t, tx)
		fiaAdmin := insertUser(t, tx, "fia_admin", 0)

		assertGuardCode(t, acknowledge(tx, changeLogID, fiaAdmin), "RC005")
	})
}

// TestAcknowledgementOfAnUntouchedChangeIsRejected: a team confirms only the
// changes whose items include it.
func TestAcknowledgementOfAnUntouchedChangeIsRejected(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		changeLogID := insertFerrariChange(t, tx)
		redBullAdmin := insertUser(t, tx, "team_admin", redBullTeamID)

		assertGuardCode(t, acknowledge(tx, changeLogID, redBullAdmin), "RC005")
	})
}

func acknowledge(tx pgx.Tx, changeLogID, userID int64) error {
	_, err := tx.Exec(context.Background(),
		`INSERT INTO score_acknowledgements (change_log_id, acknowledged_by) VALUES ($1, $2)`,
		changeLogID, userID)
	return err
}

// insertFerrariChange records a change on the seeded race that lists only a
// Ferrari driver.
func insertFerrariChange(t *testing.T, tx pgx.Tx) int64 {
	t.Helper()

	ctx := context.Background()
	actor := insertUser(t, tx, "fia_admin", 0)
	var changeLogID int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO score_change_log (event_id, actor_user_id) VALUES ($1, $2) RETURNING id`,
		seededRacedEventID, actor).Scan(&changeLogID); err != nil {
		t.Fatalf("insert score change: %v", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO score_change_log_items (change_log_id, driver_id, old_points, new_points)
		 VALUES ($1, $2, NULL, 25)`,
		changeLogID, seededFerrariDriverID); err != nil {
		t.Fatalf("insert score change item: %v", err)
	}
	return changeLogID
}

func insertUser(t *testing.T, tx pgx.Tx, role string, teamID int64) int64 {
	t.Helper()

	var userID int64
	err := tx.QueryRow(context.Background(),
		`INSERT INTO users (username, password_hash, role, team_id)
		 VALUES ($1, 'not-a-real-hash', $2, NULLIF($3, 0)) RETURNING id`,
		testdb.UniqueUsername(t), role, teamID).Scan(&userID)
	if err != nil {
		t.Fatalf("insert %s: %v", role, err)
	}
	return userID
}

// inRolledBackTx runs the test body in a transaction that is always rolled back,
// since score changes and acknowledgements are append-only and could never be
// cleaned up otherwise.
func inRolledBackTx(t *testing.T, pool *pgxpool.Pool, body func(tx pgx.Tx)) {
	t.Helper()

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	body(tx)
}
