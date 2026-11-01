package dbtest

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/racecontrol/backend/tests/testdb"
)

// Penalty types of seeds/reference.sql: a reprimand takes no value, a time
// penalty takes 1 to 300 seconds.
const (
	reprimandPenaltyID   = 2
	timePenaltyPenaltyID = 3
)

func insertSanctionOnTeam(t *testing.T, tx pgx.Tx, teamID int64) int64 {
	t.Helper()

	var sanctionID int64
	if err := tx.QueryRow(context.Background(),
		`INSERT INTO sanctions (team_id) VALUES ($1) RETURNING id`, teamID).Scan(&sanctionID); err != nil {
		t.Fatalf("insert sanction: %v", err)
	}
	return sanctionID
}

func insertVersion(tx pgx.Tx, sanctionID int64, action string, penaltyTypeID int64, value any, note any, actorID int64) (int64, error) {
	var versionID int64
	err := inSavepointRow(tx, &versionID,
		`INSERT INTO sanction_versions (sanction_id, action, penalty_type_id, penalty_value, reason, note, actor_user_id)
		 VALUES ($1, $2, $3, $4, 'Motivo de prueba', $5, $6) RETURNING id`,
		sanctionID, action, penaltyTypeID, value, note, actorID)
	return versionID, err
}

func assertSQLState(t *testing.T, err error, want string) {
	t.Helper()

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != want {
		t.Fatalf("error = %v, want SQLSTATE %s", err, want)
	}
}

func TestSanctionFallsOnExactlyOneSubject(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		// Without a subject there is no accountable team either, and the trigger
		// that derives it refuses first.
		assertGuardCode(t, inSavepoint(tx, `INSERT INTO sanctions (driver_id, team_id) VALUES (NULL, NULL)`), "RC011")
		assertSQLState(t, inSavepoint(tx, `INSERT INTO sanctions (driver_id, team_id) VALUES ($1, $2)`,
			seededFerrariDriverID, redBullTeamID), pgerrcode.CheckViolation)
	})
}

func TestDriverSanctionTakesTheDriversTeam(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		var accountable int64
		err := tx.QueryRow(context.Background(),
			`INSERT INTO sanctions (driver_id, accountable_team_id) VALUES ($1, $2) RETURNING accountable_team_id`,
			seededFerrariDriverID, redBullTeamID).Scan(&accountable)
		if err != nil || accountable != formulaOneTeamID {
			t.Fatalf("accountable team = %d (%v), want the driver's team %d whatever was sent", accountable, err, formulaOneTeamID)
		}
	})
}

func TestSanctionOfADriverWithoutTeamIsRefused(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		var driverID int64
		if err := tx.QueryRow(context.Background(),
			`INSERT INTO drivers (first_name, last_name, license_number, country_code)
			 VALUES ('Sin', 'Equipo', $1, 'ar') RETURNING id`, testdb.UniqueCode(t)).Scan(&driverID); err != nil {
			t.Fatalf("insert driver: %v", err)
		}
		assertGuardCode(t, inSavepoint(tx, `INSERT INTO sanctions (driver_id) VALUES ($1)`, driverID), "RC011")
	})
}

func TestPenaltyValueMustFitItsType(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		actor := insertUser(t, tx, "fia_admin", 0)
		for _, invalid := range []struct {
			penaltyTypeID int64
			value         any
		}{
			{reprimandPenaltyID, 5},
			{timePenaltyPenaltyID, nil},
			{timePenaltyPenaltyID, 0},
			{timePenaltyPenaltyID, 301},
		} {
			sanctionID := insertSanctionOnTeam(t, tx, formulaOneTeamID)
			_, err := insertVersion(tx, sanctionID, "issued", invalid.penaltyTypeID, invalid.value, nil, actor)
			assertGuardCode(t, err, "RC010")
		}
	})
}

func TestSanctionHistoryStartsWithTheIssueAndEndsWithTheRevocation(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		actor := insertUser(t, tx, "fia_admin", 0)
		sanctionID := insertSanctionOnTeam(t, tx, formulaOneTeamID)

		_, err := insertVersion(tx, sanctionID, "amended", reprimandPenaltyID, nil, "Antes del alta", actor)
		assertGuardCode(t, err, "RC012")
		if _, err := insertVersion(tx, sanctionID, "issued", reprimandPenaltyID, nil, nil, actor); err != nil {
			t.Fatalf("issue: %v", err)
		}
		_, err = insertVersion(tx, sanctionID, "issued", reprimandPenaltyID, nil, nil, actor)
		assertGuardCode(t, err, "RC012")
		_, err = insertVersion(tx, sanctionID, "revoked", reprimandPenaltyID, nil, nil, actor)
		assertSQLState(t, err, pgerrcode.CheckViolation)
		if _, err := insertVersion(tx, sanctionID, "revoked", reprimandPenaltyID, nil, "Apelación", actor); err != nil {
			t.Fatalf("revoke: %v", err)
		}
		_, err = insertVersion(tx, sanctionID, "amended", reprimandPenaltyID, nil, "Después de revocar", actor)
		assertGuardCode(t, err, "RC012")
	})
}

func TestSanctionRecordIsAppendOnly(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		actor := insertUser(t, tx, "fia_admin", 0)
		ferrariAdmin := insertUser(t, tx, "team_admin", formulaOneTeamID)
		sanctionID := insertSanctionOnTeam(t, tx, formulaOneTeamID)
		versionID, err := insertVersion(tx, sanctionID, "issued", reprimandPenaltyID, nil, nil, actor)
		if err != nil {
			t.Fatalf("issue: %v", err)
		}
		if err := inSavepoint(tx, `INSERT INTO sanction_acknowledgements (sanction_version_id, acknowledged_by) VALUES ($1, $2)`,
			versionID, ferrariAdmin); err != nil {
			t.Fatalf("acknowledge: %v", err)
		}

		for _, statement := range []string{
			`UPDATE sanctions SET team_id = 2 WHERE id = $1`,
			`DELETE FROM sanctions WHERE id = $1`,
			`UPDATE sanction_versions SET reason = 'otro' WHERE sanction_id = $1`,
			`DELETE FROM sanction_versions WHERE sanction_id = $1`,
			`DELETE FROM sanction_recipients WHERE sanction_version_id IN (SELECT id FROM sanction_versions WHERE sanction_id = $1)`,
			`UPDATE sanction_acknowledgements SET acknowledged_at = now() WHERE sanction_version_id IN (SELECT id FROM sanction_versions WHERE sanction_id = $1)`,
		} {
			if err := inSavepoint(tx, statement, sanctionID); err == nil {
				t.Fatalf("%s should be refused", statement)
			}
		}
	})
}

func TestOnlyARecipientTeamConfirmsTheVersionThatStands(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		actor := insertUser(t, tx, "fia_admin", 0)
		formulaTwoAdmin := insertUser(t, tx, "team_admin", formulaTwoTeamID)
		redBullAdmin := insertUser(t, tx, "team_admin", redBullTeamID)
		sanctionID := insertSanctionOnTeam(t, tx, formulaOneTeamID)
		issued, err := insertVersion(tx, sanctionID, "issued", reprimandPenaltyID, nil, nil, actor)
		if err != nil {
			t.Fatalf("issue: %v", err)
		}
		acknowledgeVersion := func(versionID, userID int64) error {
			return inSavepoint(tx, `INSERT INTO sanction_acknowledgements (sanction_version_id, acknowledged_by) VALUES ($1, $2)`,
				versionID, userID)
		}

		assertGuardCode(t, acknowledgeVersion(issued, actor), "RC013")
		assertGuardCode(t, acknowledgeVersion(issued, formulaTwoAdmin), "RC013")
		if _, err := insertVersion(tx, sanctionID, "amended", timePenaltyPenaltyID, 5, "Corrección", actor); err != nil {
			t.Fatalf("amend: %v", err)
		}
		assertGuardCode(t, acknowledgeVersion(issued, redBullAdmin), "RC014")
	})
}
