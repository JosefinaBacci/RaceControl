package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/db/dbgen"
)

// Transactor runs a unit of work inside a single database transaction.
type Transactor struct {
	pool *pgxpool.Pool
}

// NewTransactor binds the transactor to a connection pool.
func NewTransactor(pool *pgxpool.Pool) *Transactor {
	return &Transactor{pool: pool}
}

// InTx commits when work succeeds and rolls back when it returns an error.
func (t *Transactor) InTx(ctx context.Context, work func(queries dbgen.Querier) error) error {
	err := pgx.BeginFunc(ctx, t.pool, func(tx pgx.Tx) error {
		return work(dbgen.New(tx))
	})
	if err != nil {
		return fmt.Errorf("transaction: %w", err)
	}
	return nil
}
