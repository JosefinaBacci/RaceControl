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

// Demo applies the sample calendar, drivers and scores. It is reserved for
// non-production environments so a real deployment starts with an empty season.
func (r *Runner) Demo(ctx context.Context) error {
	if err := applyFile(ctx, r.pool, "demo.sql"); err != nil {
		return fmt.Errorf("apply demo seed: %w", err)
	}
	return nil
}

// DemoScoreHistory applies the score changes and acknowledgements of the demo
// race. They name the demo accounts, so it runs after DemoAccounts.
func (r *Runner) DemoScoreHistory(ctx context.Context) error {
	if err := applyFile(ctx, r.pool, "demo_scores.sql"); err != nil {
		return fmt.Errorf("apply demo score history seed: %w", err)
	}
	return nil
}

// DemoMessages applies the demo conversations. They name the demo accounts, so it
// runs after DemoAccounts.
func (r *Runner) DemoMessages(ctx context.Context) error {
	if err := applyFile(ctx, r.pool, "demo_messages.sql"); err != nil {
		return fmt.Errorf("apply demo messages seed: %w", err)
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
