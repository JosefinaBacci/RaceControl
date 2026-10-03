package testdb

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/db"
)

var uniqueSequence atomic.Int64

func Pool(t *testing.T) *pgxpool.Pool {
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

func UniqueUsername(t *testing.T) string {
	t.Helper()

	return fmt.Sprintf("t%d.%d", time.Now().UnixNano()%1_000_000_000_000, uniqueSequence.Add(1))
}

// UniqueCode returns a value that satisfies the short code formats of the
// reference tables, such as teams_code_format.
func UniqueCode(t *testing.T) string {
	t.Helper()

	return fmt.Sprintf("t%d_%d", time.Now().UnixNano()%1_000_000_000, uniqueSequence.Add(1))
}
