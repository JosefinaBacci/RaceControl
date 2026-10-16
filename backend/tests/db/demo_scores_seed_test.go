package dbtest

import (
	"context"
	"io/fs"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/seeds"
	"github.com/racecontrol/backend/tests/testdb"
)

const (
	demoPublicationID = 1
	demoCorrectionID  = 2
	seededSecondRace  = 3
)

type changeFootprint struct {
	changeExists bool
	items        int
	acks         int
}

// TestDemoScoreSeedSkipsAForeignChange covers a database where change 1 was
// written by someone else before the demo accounts existed: the seed must leave
// that append-only change alone instead of grafting the demo lines and the
// acknowledgement onto it, and must not add a correction of a race it never
// published.
func TestDemoScoreSeedSkipsAForeignChange(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		ensureForeignPublicationSlot(t, tx)
		ensureDemoAccount(t, tx, "fia.admin", "fia_admin", 0)
		ensureDemoAccount(t, tx, "redbull.admin", "team_admin", redBullTeamID)
		before := map[int64]changeFootprint{
			demoPublicationID: footprintOf(t, tx, demoPublicationID),
			demoCorrectionID:  footprintOf(t, tx, demoCorrectionID),
		}

		if _, err := tx.Exec(context.Background(), demoScoreInserts(t)); err != nil {
			t.Fatalf("apply the demo score seed: %v", err)
		}

		for changeLogID, want := range before {
			if got := footprintOf(t, tx, changeLogID); got != want {
				t.Errorf("change %d = %+v after seeding, want it untouched: %+v", changeLogID, got, want)
			}
		}
	})
}

// demoScoreInserts is the seed without its trailing setval: setval is not
// transactional, so running it here would move the shared test sequences.
func demoScoreInserts(t *testing.T) string {
	t.Helper()

	contents, err := fs.ReadFile(seeds.Files, "demo_scores.sql")
	if err != nil {
		t.Fatalf("read demo score seed: %v", err)
	}
	inserts, _, found := strings.Cut(string(contents), "-- 4. setval")
	if !found {
		t.Fatal("the demo score seed no longer ends with its setval section")
	}
	return inserts
}

// ensureForeignPublicationSlot makes change 1 belong to another race and user,
// unless an earlier test run already left a change there, which is foreign too.
func ensureForeignPublicationSlot(t *testing.T, tx pgx.Tx) {
	t.Helper()

	ctx := context.Background()
	actor := insertUser(t, tx, "fia_admin", 0)
	if _, err := tx.Exec(ctx,
		`INSERT INTO score_change_log (id, event_id, actor_user_id)
		 OVERRIDING SYSTEM VALUE VALUES ($1, $2, $3)
		 ON CONFLICT (id) DO NOTHING`,
		demoPublicationID, seededSecondRace, actor); err != nil {
		t.Fatalf("insert foreign change: %v", err)
	}

	var eventID int64
	if err := tx.QueryRow(ctx, `SELECT event_id FROM score_change_log WHERE id = $1`, demoPublicationID).Scan(&eventID); err != nil {
		t.Fatalf("read change %d: %v", demoPublicationID, err)
	}
	if eventID == seededRacedEventID {
		t.Fatalf("change %d is on the demo race, so it is not foreign", demoPublicationID)
	}
}

func ensureDemoAccount(t *testing.T, tx pgx.Tx, username, role string, teamID int64) {
	t.Helper()

	if _, err := tx.Exec(context.Background(),
		`INSERT INTO users (username, password_hash, role, team_id)
		 VALUES ($1, 'not-a-real-hash', $2, NULLIF($3, 0))
		 ON CONFLICT (username) DO NOTHING`,
		username, role, teamID); err != nil {
		t.Fatalf("insert %s: %v", username, err)
	}
}

func footprintOf(t *testing.T, tx pgx.Tx, changeLogID int64) changeFootprint {
	t.Helper()

	var footprint changeFootprint
	err := tx.QueryRow(context.Background(), `
		SELECT EXISTS (SELECT 1 FROM score_change_log WHERE id = $1),
		       (SELECT count(*) FROM score_change_log_items WHERE change_log_id = $1),
		       (SELECT count(*) FROM score_acknowledgements WHERE change_log_id = $1)`,
		changeLogID).Scan(&footprint.changeExists, &footprint.items, &footprint.acks)
	if err != nil {
		t.Fatalf("footprint of change %d: %v", changeLogID, err)
	}
	return footprint
}
