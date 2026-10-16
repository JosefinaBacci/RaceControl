package testdb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Seeded reference rows a championship hangs from: season 1 is 2026 and circuit
// 1 is any track of the catalogue.
const (
	seededSeasonID  = 1
	seededCircuitID = 1
	raceDuration    = 2 * time.Hour
)

// SeasonStart is the first instant of the seeded 2026 season.
var SeasonStart = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// Championship is a category of its own in the seeded season, so its results,
// standings and score notices are only what the test puts there. Score changes
// are append-only and outlive the test, which is why each test builds a new one
// instead of sharing the seeded Formula 1.
type Championship struct {
	t            *testing.T
	pool         *pgxpool.Pool
	CategoryID   int64
	CategoryCode string
}

// NewChampionship creates the category.
func NewChampionship(t *testing.T, pool *pgxpool.Pool) *Championship {
	t.Helper()

	championship := &Championship{t: t, pool: pool, CategoryCode: UniqueCode(t)}
	championship.queryRow(&championship.CategoryID,
		`INSERT INTO categories (code, name) VALUES ($1, 'Test championship') RETURNING id`,
		championship.CategoryCode)
	return championship
}

// AddTeam creates a team racing in the championship.
func (c *Championship) AddTeam(name string) int64 {
	c.t.Helper()

	var teamID int64
	c.queryRow(&teamID,
		`INSERT INTO teams (category_id, code, name) VALUES ($1, $2, $3) RETURNING id`,
		c.CategoryID, fmt.Sprintf("team_%d", uniqueSequence.Add(1)), name)
	return teamID
}

// AddDriver creates a driver assigned to the team for the whole 2026 season.
func (c *Championship) AddDriver(teamID int64, lastName string) int64 {
	c.t.Helper()
	return c.AddDriverSince(teamID, lastName, SeasonStart)
}

// AddDriverSince creates a driver assigned to the team from the given date.
func (c *Championship) AddDriverSince(teamID int64, lastName string, from time.Time) int64 {
	c.t.Helper()

	var driverID int64
	c.queryRow(&driverID,
		`INSERT INTO drivers (first_name, last_name, license_number, country_code)
		 VALUES ('Test', $1, $2, 'ar') RETURNING id`,
		lastName, UniqueCode(c.t))
	c.exec(`INSERT INTO driver_team_assignments (driver_id, team_id, effective_from, is_starter)
	        VALUES ($1, $2, $3, true)`, driverID, teamID, from)
	return driverID
}

// Transfer moves the driver to another team from the given date. The history
// only grows forward (000022), so the date must follow the driver's last race.
func (c *Championship) Transfer(driverID, teamID int64, from time.Time) {
	c.t.Helper()

	c.exec(`INSERT INTO driver_team_assignments (driver_id, team_id, effective_from, is_starter)
	        VALUES ($1, $2, $3, true)`, driverID, teamID, from)
}

// AddRace creates a race of the championship in the seeded 2026 season.
func (c *Championship) AddRace(startsAt time.Time) int64 {
	c.t.Helper()
	return c.AddRaceIn(seededSeasonID, startsAt)
}

// AddRaceIn creates a race of the championship on its own round of the season.
func (c *Championship) AddRaceIn(seasonID int64, startsAt time.Time) int64 {
	c.t.Helper()

	var roundID int64
	c.queryRow(&roundID,
		`INSERT INTO rounds (season_id, category_id, number, circuit_id, gp_name)
		 SELECT $1, $2, COALESCE(max(number), 0) + 1, $3, 'Gran Premio de Prueba'
		 FROM rounds WHERE season_id = $1 AND category_id = $2
		 RETURNING id`,
		seasonID, c.CategoryID, seededCircuitID)

	var eventID int64
	err := pgx.BeginFunc(context.Background(), c.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(context.Background(),
			`INSERT INTO events (kind, starts_at, ends_at) VALUES ('race', $1, $2) RETURNING id`,
			startsAt, startsAt.Add(raceDuration)).Scan(&eventID); err != nil {
			return err
		}
		_, err := tx.Exec(context.Background(),
			`INSERT INTO races (event_id, round_id) VALUES ($1, $2)`, eventID, roundID)
		return err
	})
	if err != nil {
		c.t.Fatalf("create race of %s: %v", c.CategoryCode, err)
	}
	return eventID
}

// Score stores a driver's result in a race, the way the FIA loads it.
func (c *Championship) Score(eventID, driverID int64, points float64) {
	c.t.Helper()

	c.exec(`INSERT INTO score_entries (event_id, driver_id, points) VALUES ($1, $2, $3)`,
		eventID, driverID, points)
}

// PointsChange is one driver's line in a score change. A nil Old publishes the
// driver's first score in the race.
type PointsChange struct {
	DriverID int64
	Old      *float64
	New      float64
}

// ScoreChange describes a publication or a correction of a race's scores.
type ScoreChange struct {
	EventID   int64
	ActorID   int64
	ChangedAt time.Time
	Lines     []PointsChange
}

// RecordChange writes the change log row and its lines, as score loading does.
func (c *Championship) RecordChange(change ScoreChange) int64 {
	c.t.Helper()

	var changeLogID int64
	c.queryRow(&changeLogID,
		`INSERT INTO score_change_log (event_id, actor_user_id, changed_at) VALUES ($1, $2, $3) RETURNING id`,
		change.EventID, change.ActorID, change.ChangedAt)
	for _, line := range change.Lines {
		c.exec(`INSERT INTO score_change_log_items (change_log_id, driver_id, old_points, new_points)
		        VALUES ($1, $2, $3, $4)`, changeLogID, line.DriverID, line.Old, line.New)
	}
	return changeLogID
}

// Season answers the id of the season of that year, creating it when missing.
// Only a year before the seeded 2026 is safe: a later one would become the
// latest season the API defaults to.
func Season(t *testing.T, pool *pgxpool.Pool, year int32) int64 {
	t.Helper()

	var seasonID int64
	err := pool.QueryRow(context.Background(),
		`WITH created AS (
		     INSERT INTO seasons (year) VALUES ($1) ON CONFLICT (year) DO NOTHING RETURNING id
		 )
		 SELECT id FROM created
		 UNION ALL
		 SELECT id FROM seasons WHERE year = $1`,
		year).Scan(&seasonID)
	if err != nil {
		t.Fatalf("season %d: %v", year, err)
	}
	return seasonID
}

// UserID resolves a username to its id.
func UserID(t *testing.T, pool *pgxpool.Pool, username string) int64 {
	t.Helper()

	var userID int64
	if err := pool.QueryRow(context.Background(),
		`SELECT id FROM users WHERE username = $1`, username).Scan(&userID); err != nil {
		t.Fatalf("user %s: %v", username, err)
	}
	return userID
}

func (c *Championship) queryRow(destination *int64, query string, args ...any) {
	c.t.Helper()

	if err := c.pool.QueryRow(context.Background(), query, args...).Scan(destination); err != nil {
		c.t.Fatalf("championship %s: %v", c.CategoryCode, err)
	}
}

func (c *Championship) exec(query string, args ...any) {
	c.t.Helper()

	if _, err := c.pool.Exec(context.Background(), query, args...); err != nil {
		c.t.Fatalf("championship %s: %v", c.CategoryCode, err)
	}
}
