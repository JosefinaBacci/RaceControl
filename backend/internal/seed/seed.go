package seed

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/seeds"
)

// Runner applies the reference data and, outside production, the demo accounts.
type Runner struct {
	pool *pgxpool.Pool
}

// NewRunner binds the seeder to a connection pool.
func NewRunner(pool *pgxpool.Pool) *Runner {
	return &Runner{pool: pool}
}

// Reference applies the embedded reference data. It is idempotent, so calling it
// repeatedly neither duplicates rows nor rewinds identity sequences.
func (r *Runner) Reference(ctx context.Context) error {
	if err := applyFile(ctx, r.pool, "reference.sql"); err != nil {
		return fmt.Errorf("apply reference seed: %w", err)
	}
	return nil
}

func applyFile(ctx context.Context, pool *pgxpool.Pool, name string) error {
	contents, err := fs.ReadFile(seeds.Files, name)
	if err != nil {
		return fmt.Errorf("read seed %s: %w", name, err)
	}

	if _, err := pool.Exec(ctx, string(contents)); err != nil {
		return fmt.Errorf("execute seed %s: %w", name, err)
	}
	return nil
}
