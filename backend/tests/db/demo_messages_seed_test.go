package dbtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/tests/testdb"
)

const (
	demoFerrariConversationID = 1
	mclarenTeamID             = 4
)

type conversationFootprint struct {
	messages int
	reads    int
}

// TestDemoMessagesSeedSkipsAForeignConversation covers a database where
// conversation 1 was opened by someone else before the demo accounts existed:
// the seed must not graft the demo thread or a read cursor onto it.
func TestDemoMessagesSeedSkipsAForeignConversation(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		ensureForeignConversationSlot(t, tx)
		ensureDemoAccount(t, tx, "fia.admin", "fia_admin", 0)
		ensureDemoAccount(t, tx, "ferrari.admin", "team_admin", formulaOneTeamID)
		before := conversationFootprintOf(t, tx, demoFerrariConversationID)

		if _, err := tx.Exec(context.Background(), seedInserts(t, "demo_messages.sql", "-- 4. setval")); err != nil {
			t.Fatalf("apply the demo messages seed: %v", err)
		}

		if after := conversationFootprintOf(t, tx, demoFerrariConversationID); after != before {
			t.Fatalf("conversation %d = %+v after seeding, want it untouched: %+v", demoFerrariConversationID, after, before)
		}
	})
}

// ensureForeignConversationSlot makes conversation 1 belong to McLaren, unless an
// earlier test run already left a conversation there.
func ensureForeignConversationSlot(t *testing.T, tx pgx.Tx) {
	t.Helper()

	author := insertUser(t, tx, "team_admin", mclarenTeamID)
	if _, err := tx.Exec(context.Background(),
		`INSERT INTO conversations (id, team_id, subject, created_by)
		 OVERRIDING SYSTEM VALUE VALUES ($1, $2, 'Ajena', $3)
		 ON CONFLICT (id) DO NOTHING`,
		demoFerrariConversationID, mclarenTeamID, author); err != nil {
		t.Fatalf("insert foreign conversation: %v", err)
	}

	var username string
	if err := tx.QueryRow(context.Background(),
		`SELECT users.username FROM conversations JOIN users ON users.id = conversations.created_by WHERE conversations.id = $1`,
		demoFerrariConversationID).Scan(&username); err != nil {
		t.Fatalf("read conversation %d: %v", demoFerrariConversationID, err)
	}
	if username == "ferrari.admin" {
		t.Fatalf("conversation %d was opened by ferrari.admin, so it is not foreign", demoFerrariConversationID)
	}
}

func conversationFootprintOf(t *testing.T, tx pgx.Tx, conversationID int64) conversationFootprint {
	t.Helper()

	var footprint conversationFootprint
	err := tx.QueryRow(context.Background(), `
		SELECT (SELECT count(*) FROM messages WHERE conversation_id = $1),
		       (SELECT count(*) FROM conversation_reads WHERE conversation_id = $1)`,
		conversationID).Scan(&footprint.messages, &footprint.reads)
	if err != nil {
		t.Fatalf("footprint of conversation %d: %v", conversationID, err)
	}
	return footprint
}
