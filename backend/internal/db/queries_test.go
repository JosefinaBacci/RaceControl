package db_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/db"
)

// usernameSequence makes concurrent test binaries produce distinct usernames.
var usernameSequence atomic.Int64

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set; skipping database integration test")
	}

	pool, err := db.Connect(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

// TestGetUserByUsernameNotFound asserts that a missing username yields pgx.ErrNoRows
// rather than a zero value, which is what the auth service turns into a 401.
func TestGetUserByUsernameNotFound(t *testing.T) {
	queries := db.NewQueries(testPool(t))

	_, err := queries.GetUserByUsername(context.Background(), "nobody-here")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}

// TestGetUserByUsernameIsCaseInsensitive locks in the login behaviour that lets a
// user type their username in any casing while the schema stores only lowercase.
func TestGetUserByUsernameIsCaseInsensitive(t *testing.T) {
	pool := testPool(t)
	queries := db.NewQueries(pool)

	username := uniqueUsername(t)
	insertTestUser(t, pool, username)

	for _, input := range []string{username, strings.ToUpper(username)} {
		user, err := queries.GetUserByUsername(context.Background(), input)
		if err != nil {
			t.Fatalf("GetUserByUsername(%q): %v", input, err)
		}
		if user.Username != username {
			t.Errorf("GetUserByUsername(%q) = %q, want %q", input, user.Username, username)
		}
	}
}

// TestInsertAfterSeededRowsDoesNotCollide proves the seed migration left the
// identity sequences ahead of the rows it inserted explicitly.
func TestInsertAfterSeededRowsDoesNotCollide(t *testing.T) {
	pool := testPool(t)
	queries := db.NewQueries(pool)

	insertTestUser(t, pool, uniqueUsername(t))

	user, err := queries.GetUserByUsername(context.Background(), seededUsername(t))
	if err != nil {
		t.Fatalf("read seeded user: %v", err)
	}
	if user.ID <= 0 {
		t.Fatalf("seeded user has id %d, want a positive identity value", user.ID)
	}
}

// seededUsername returns a username known to exist from the seed migration.
func seededUsername(t *testing.T) string {
	t.Helper()

	const knownSeededUsername = "seeded_probe"
	pool := testPool(t)
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO users (username, password_hash, role) VALUES ($1, 'hash', 'fia_admin')
		 ON CONFLICT (username) DO NOTHING`,
		knownSeededUsername,
	); err != nil {
		t.Fatalf("insert probe user: %v", err)
	}
	return knownSeededUsername
}

func insertTestUser(t *testing.T, pool *pgxpool.Pool, username string) {
	t.Helper()

	if _, err := pool.Exec(context.Background(),
		`INSERT INTO users (username, password_hash, role) VALUES ($1, 'test-hash', 'fia_admin')`,
		username,
	); err != nil {
		t.Fatalf("insert test user %q: %v", username, err)
	}
}

// uniqueUsername builds a lowercase username honouring the users_username_format
// check constraint, unique across concurrent test binaries.
func uniqueUsername(t *testing.T) string {
	t.Helper()

	return fmt.Sprintf("test_%d_%d", time.Now().UnixNano(), usernameSequence.Add(1))
}
