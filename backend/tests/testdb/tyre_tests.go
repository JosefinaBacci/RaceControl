package testdb

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// AddTyreTest creates a tyre test of the championship in the seeded season. Its
// results are append-only and outlive the test, like the championship itself.
func (c *Championship) AddTyreTest(startsAt, endsAt time.Time) int64 {
	c.t.Helper()

	var eventID int64
	err := pgx.BeginFunc(context.Background(), c.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(context.Background(),
			`INSERT INTO events (kind, starts_at, ends_at) VALUES ('tyre_test', $1, $2) RETURNING id`,
			startsAt, endsAt).Scan(&eventID); err != nil {
			return err
		}
		_, err := tx.Exec(context.Background(),
			`INSERT INTO tyre_tests (event_id, season_id, category_id, circuit_id) VALUES ($1, $2, $3, $4)`,
			eventID, seededSeasonID, c.CategoryID, seededCircuitID)
		return err
	})
	if err != nil {
		c.t.Fatalf("create tyre test of %s: %v", c.CategoryCode, err)
	}
	return eventID
}

// TyreResult stores the original result of a compound in a tyre test straight in
// the table, for the tests that probe the schema rather than the API.
func (c *Championship) TyreResult(eventID int64, compoundCode string, conductedOn time.Time, actorID int64) int64 {
	c.t.Helper()

	var resultID int64
	c.queryRow(&resultID,
		`INSERT INTO tyre_test_results (event_id, compound_id, conducted_on, passed, recorded_by)
		 SELECT $1, id, $3, true, $4 FROM tyre_compounds WHERE code = $2
		 RETURNING id`,
		eventID, compoundCode, conductedOn, actorID)
	return resultID
}
