package dbtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/tests/testdb"
)

const demoSanctionsSetval = "-- 4. setval"

var demoSanctionIDs = []int64{1, 2, 3, 4}

// TestDemoSanctionSeedIsIdempotentAndLeavesForeignSanctionsAlone runs the seed
// twice next to whatever sanctions the test database already holds under the
// demo ids: the second run must change nothing, and a sanction the seed does not
// own keeps its versions.
func TestDemoSanctionSeedIsIdempotentAndLeavesForeignSanctionsAlone(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		ensureDemoAccount(t, tx, "fia.admin", "fia_admin", 0)
		ensureDemoAccount(t, tx, "ferrari.admin", "team_admin", formulaOneTeamID)
		ensureDemoAccount(t, tx, "redbull.admin", "team_admin", redBullTeamID)
		ensureDemoAccount(t, tx, "mercedes.admin", "team_admin", 3)
		foreign := foreignSanctionVersions(t, tx)
		inserts := seedInserts(t, "demo_sanctions.sql", demoSanctionsSetval)

		applySeed(t, tx, inserts)
		afterFirst := sanctionRowCounts(t, tx)
		applySeed(t, tx, inserts)

		if afterSecond := sanctionRowCounts(t, tx); afterSecond != afterFirst {
			t.Fatalf("second run changed the rows: %v, then %v", afterFirst, afterSecond)
		}
		for sanctionID, versions := range foreign {
			if got := versionCount(t, tx, sanctionID); got != versions {
				t.Fatalf("foreign sanction %d has %d versions after seeding, want %d", sanctionID, got, versions)
			}
		}
	})
}

func applySeed(t *testing.T, tx pgx.Tx, inserts string) {
	t.Helper()

	if _, err := tx.Exec(context.Background(), inserts); err != nil {
		t.Fatalf("apply the demo sanctions seed: %v", err)
	}
}

// foreignSanctionVersions answers the version count of every demo id already
// taken before the seed runs: none of them was written by this seed in this
// database, since the test database has no demo accounts.
func foreignSanctionVersions(t *testing.T, tx pgx.Tx) map[int64]int {
	t.Helper()

	foreign := map[int64]int{}
	for _, sanctionID := range demoSanctionIDs {
		var exists bool
		if err := tx.QueryRow(context.Background(), `SELECT EXISTS (SELECT 1 FROM sanctions WHERE id = $1)`, sanctionID).Scan(&exists); err != nil {
			t.Fatalf("look up sanction %d: %v", sanctionID, err)
		}
		if exists {
			foreign[sanctionID] = versionCount(t, tx, sanctionID)
		}
	}
	return foreign
}

func versionCount(t *testing.T, tx pgx.Tx, sanctionID int64) int {
	t.Helper()

	var count int
	if err := tx.QueryRow(context.Background(), `SELECT count(*) FROM sanction_versions WHERE sanction_id = $1`, sanctionID).Scan(&count); err != nil {
		t.Fatalf("count versions of sanction %d: %v", sanctionID, err)
	}
	return count
}

func sanctionRowCounts(t *testing.T, tx pgx.Tx) [3]int {
	t.Helper()

	var counts [3]int
	err := tx.QueryRow(context.Background(), `
		SELECT (SELECT count(*) FROM sanctions),
		       (SELECT count(*) FROM sanction_versions),
		       (SELECT count(*) FROM sanction_acknowledgements)`).Scan(&counts[0], &counts[1], &counts[2])
	if err != nil {
		t.Fatalf("count sanction rows: %v", err)
	}
	return counts
}
