package db_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/testdb"
)

const highestSeededTeamID = 8

// TestGetUserByUsernameNotFound asserts that a missing username yields pgx.ErrNoRows
// rather than a zero value, which is what the authenticator turns into a 401.
func TestGetUserByUsernameNotFound(t *testing.T) {
	queries := dbgen.New(testdb.Pool(t))

	_, err := queries.GetUserByUsername(context.Background(), "nobody-here")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}

// TestGetUserByUsernameIsCaseInsensitive locks in the login behaviour that lets a
// user type their username in any casing while the schema stores only lowercase.
func TestGetUserByUsernameIsCaseInsensitive(t *testing.T) {
	queries := dbgen.New(testdb.Pool(t))
	created, err := queries.CreateUser(context.Background(), dbgen.CreateUserParams{
		Username:     testdb.UniqueUsername(t),
		PasswordHash: "test-hash",
		Role:         "fia_admin",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	for _, input := range []string{created.Username, strings.ToUpper(created.Username)} {
		user, err := queries.GetUserByUsername(context.Background(), input)
		if err != nil {
			t.Fatalf("GetUserByUsername(%q): %v", input, err)
		}
		if user.ID != created.ID {
			t.Errorf("GetUserByUsername(%q) returned user %d, want %d", input, user.ID, created.ID)
		}
	}
}

// TestInsertAfterSeedDoesNotCollide proves the seed left the identity sequence
// ahead of the rows it inserted with fixed ids.
func TestInsertAfterSeedDoesNotCollide(t *testing.T) {
	var teamID int64
	err := testdb.Pool(t).QueryRow(context.Background(),
		`INSERT INTO teams (category_id, code, name) VALUES (1, $1, 'Test team') RETURNING id`,
		testdb.UniqueCode(t),
	).Scan(&teamID)
	if err != nil {
		t.Fatalf("insert team: %v", err)
	}
	if teamID <= highestSeededTeamID {
		t.Fatalf("new team got id %d, want an id above the seeded %d", teamID, highestSeededTeamID)
	}
}
