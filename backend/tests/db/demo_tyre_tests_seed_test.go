package dbtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/tests/testdb"
)

// TestDemoTyreTestsSeedIsRepeatable applies the seed twice, as every `make seed`
// does. Results are append-only and the test database's ids may already belong to
// other tests' results, so the seed must skip what is taken instead of failing,
// and its correction may only land on the seed's own c4 original.
func TestDemoTyreTestsSeedIsRepeatable(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		ensureDemoAccount(t, tx, "fia.admin", "fia_admin", 0)
		inserts := seedInserts(t, "demo_tyre_tests.sql", "-- setval")

		for range 2 {
			if _, err := tx.Exec(context.Background(), inserts); err != nil {
				t.Fatalf("apply the demo tyre tests seed: %v", err)
			}
		}

		var graftedCorrections int
		if err := tx.QueryRow(context.Background(),
			`SELECT count(*)
			 FROM tyre_test_results correction
			 JOIN tyre_test_results original ON original.id = correction.supersedes_id
			 JOIN tyre_compounds ON tyre_compounds.id = original.compound_id
			 JOIN users ON users.id = original.recorded_by
			 WHERE correction.id = 3
			   AND (tyre_compounds.code <> 'c4' OR users.username <> 'fia.admin' OR original.event_id <> 4)`,
		).Scan(&graftedCorrections); err != nil {
			t.Fatalf("read the demo correction: %v", err)
		}
		if graftedCorrections != 0 {
			t.Fatal("the demo correction was grafted onto a result the seed does not own")
		}
	})
}
