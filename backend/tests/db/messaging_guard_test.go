package dbtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/tests/testdb"
)

const participantGuard = "RC007"

// TestTeamAdminCannotOpenAConversationWithAnotherTeam covers the creation guard
// of 000028: a team administrator only opens conversations with their own team.
func TestTeamAdminCannotOpenAConversationWithAnotherTeam(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		ferrariAdmin := insertUser(t, tx, "team_admin", formulaOneTeamID)

		_, err := openConversation(tx, redBullTeamID, ferrariAdmin)
		assertGuardCode(t, err, participantGuard)
	})
}

// TestOnlyParticipantsWriteInAConversation: the team's administrators and every
// FIA administrator write; another team's administrator and an inactive account
// are refused with RC007.
func TestOnlyParticipantsWriteInAConversation(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		ferrariAdmin := insertUser(t, tx, "team_admin", formulaOneTeamID)
		conversationID := mustOpenConversation(t, tx, formulaOneTeamID, ferrariAdmin)

		for _, participant := range []int64{
			insertUser(t, tx, "team_admin", formulaOneTeamID),
			insertUser(t, tx, "fia_admin", 0),
		} {
			if _, err := writeMessage(tx, conversationID, participant); err != nil {
				t.Fatalf("participant %d should write: %v", participant, err)
			}
		}

		redBullAdmin := insertUser(t, tx, "team_admin", redBullTeamID)
		_, err := writeMessage(tx, conversationID, redBullAdmin)
		assertGuardCode(t, err, participantGuard)

		inactiveFIAAdmin := insertUser(t, tx, "fia_admin", 0)
		if _, err := tx.Exec(context.Background(), `UPDATE users SET deactivated_at = now() WHERE id = $1`, inactiveFIAAdmin); err != nil {
			t.Fatalf("deactivate: %v", err)
		}
		_, err = writeMessage(tx, conversationID, inactiveFIAAdmin)
		assertGuardCode(t, err, participantGuard)
	})
}

// TestMessagesAreAppendOnly: what was said is never edited nor erased.
func TestMessagesAreAppendOnly(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		author := insertUser(t, tx, "team_admin", formulaOneTeamID)
		messageID := mustWriteMessage(t, tx, mustOpenConversation(t, tx, formulaOneTeamID, author), author)

		for _, statement := range []string{
			`UPDATE messages SET body = 'otro texto' WHERE id = $1`,
			`DELETE FROM messages WHERE id = $1`,
		} {
			if err := inSavepoint(tx, statement, messageID); err == nil {
				t.Fatalf("%q succeeded, want messages to be append-only", statement)
			}
		}
	})
}

// TestReadCursorOnlyMovesForward: reading an older message again keeps the cursor
// where the newest read left it.
func TestReadCursorOnlyMovesForward(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		ctx := context.Background()
		author := insertUser(t, tx, "team_admin", formulaOneTeamID)
		reader := insertUser(t, tx, "fia_admin", 0)
		conversationID := mustOpenConversation(t, tx, formulaOneTeamID, author)
		first := mustWriteMessage(t, tx, conversationID, author)
		second := mustWriteMessage(t, tx, conversationID, author)

		queries := dbgen.New(tx)
		for _, messageID := range []int64{second, first} {
			err := queries.MarkConversationRead(ctx, dbgen.MarkConversationReadParams{
				ConversationID: conversationID, UserID: reader, LastReadMessageID: messageID,
			})
			if err != nil {
				t.Fatalf("mark read up to %d: %v", messageID, err)
			}
		}

		var cursor int64
		if err := tx.QueryRow(ctx,
			`SELECT last_read_message_id FROM conversation_reads WHERE conversation_id = $1 AND user_id = $2`,
			conversationID, reader).Scan(&cursor); err != nil {
			t.Fatalf("read cursor: %v", err)
		}
		if cursor != second {
			t.Fatalf("cursor = %d, want it to stay on %d", cursor, second)
		}
	})
}

// TestReadCursorPointsIntoItsOwnConversation: a cursor naming a message of
// another conversation is refused by the composite foreign key.
func TestReadCursorPointsIntoItsOwnConversation(t *testing.T) {
	inRolledBackTx(t, testdb.Pool(t), func(tx pgx.Tx) {
		author := insertUser(t, tx, "team_admin", formulaOneTeamID)
		conversationID := mustOpenConversation(t, tx, formulaOneTeamID, author)
		foreignMessage := mustWriteMessage(t, tx, mustOpenConversation(t, tx, formulaOneTeamID, author), author)

		err := inSavepoint(tx,
			`INSERT INTO conversation_reads (conversation_id, user_id, last_read_message_id) VALUES ($1, $2, $3)`,
			conversationID, author, foreignMessage)
		if err == nil {
			t.Fatal("a cursor on another conversation's message was accepted")
		}
	})
}

func openConversation(tx pgx.Tx, teamID, createdBy int64) (int64, error) {
	var conversationID int64
	err := inSavepointRow(tx, &conversationID,
		`INSERT INTO conversations (team_id, subject, created_by) VALUES ($1, 'Consulta', $2) RETURNING id`,
		teamID, createdBy)
	return conversationID, err
}

func mustOpenConversation(t *testing.T, tx pgx.Tx, teamID, createdBy int64) int64 {
	t.Helper()

	conversationID, err := openConversation(tx, teamID, createdBy)
	if err != nil {
		t.Fatalf("open conversation with team %d: %v", teamID, err)
	}
	return conversationID
}

func writeMessage(tx pgx.Tx, conversationID, authorID int64) (int64, error) {
	var messageID int64
	err := inSavepointRow(tx, &messageID,
		`INSERT INTO messages (conversation_id, author_id, body) VALUES ($1, $2, 'Hola') RETURNING id`,
		conversationID, authorID)
	return messageID, err
}

func mustWriteMessage(t *testing.T, tx pgx.Tx, conversationID, authorID int64) int64 {
	t.Helper()

	messageID, err := writeMessage(tx, conversationID, authorID)
	if err != nil {
		t.Fatalf("write in conversation %d: %v", conversationID, err)
	}
	return messageID
}

// inSavepoint runs a statement that may fail without aborting the surrounding
// test transaction.
func inSavepoint(tx pgx.Tx, statement string, args ...any) error {
	return pgx.BeginFunc(context.Background(), tx, func(savepoint pgx.Tx) error {
		_, err := savepoint.Exec(context.Background(), statement, args...)
		return err
	})
}

func inSavepointRow(tx pgx.Tx, destination *int64, statement string, args ...any) error {
	return pgx.BeginFunc(context.Background(), tx, func(savepoint pgx.Tx) error {
		return savepoint.QueryRow(context.Background(), statement, args...).Scan(destination)
	})
}
