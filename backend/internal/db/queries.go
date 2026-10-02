package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/db/dbgen"
)

// Queries exposes the generated, type-safe data access layer.
type Queries struct {
	inner *dbgen.Queries
}

// NewQueries binds the generated queries to a connection pool.
func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{inner: dbgen.New(pool)}
}

// GetUserByUsername looks up a user by username, case-insensitively.
func (q *Queries) GetUserByUsername(ctx context.Context, username string) (dbgen.User, error) {
	return q.inner.GetUserByUsername(ctx, username)
}

// GetUserByID looks up a user by primary key.
func (q *Queries) GetUserByID(ctx context.Context, id int64) (dbgen.User, error) {
	return q.inner.GetUserByID(ctx, id)
}
