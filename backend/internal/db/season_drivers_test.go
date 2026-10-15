package db_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/testdb"
)

// Seeded reference data these tests lean on: season 1 is 2026, teams 1 and 5
// belong to Formula 1 and Formula 2 respectively.
const (
	demoSeasonID         = 1
	formulaOneCategoryID = 1
	formulaOneTeamID     = 1
	formulaTwoCategoryID = 2
	formulaTwoTeamID     = 5
)

var (
	seasonStart = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	midSeason   = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
)

var seasonNumberSequence atomic.Int32

// TestSeasonDriverAcceptsADriverWhoJoinsMidSeason locks in the fix for the rookie
// case: a driver whose first assignment starts after January 1st still gets a
// number, because the check spans the whole season instead of its first day.
func TestSeasonDriverAcceptsADriverWhoJoinsMidSeason(t *testing.T) {
	pool := testdb.Pool(t)
	driverID := insertTestDriver(t, pool)
	assignDriver(t, pool, driverID, formulaOneTeamID, midSeason)

	if err := insertSeasonDriver(pool, driverID, formulaOneCategoryID); err != nil {
		t.Fatalf("mid-season rookie should get a number: %v", err)
	}
}

// TestSeasonDriverRejectsACategoryTheDriverNeverRacedIn asserts that an admin
// cannot file a number under a championship the driver has no assignment in.
func TestSeasonDriverRejectsACategoryTheDriverNeverRacedIn(t *testing.T) {
	pool := testdb.Pool(t)
	driverID := insertTestDriver(t, pool)
	assignDriver(t, pool, driverID, formulaOneTeamID, midSeason)

	err := insertSeasonDriver(pool, driverID, formulaTwoCategoryID)
	code, isGuard := db.GuardViolation(err)
	if !isGuard || code != "RC003" {
		t.Fatalf("category without assignment = (%q, %t), want (\"RC003\", true)", code, isGuard)
	}
}

// TestSeasonDriverKeepsOneNumberPerCategory covers the promotion case: a driver
// who changes category mid-season holds a number in each championship at once.
func TestSeasonDriverKeepsOneNumberPerCategory(t *testing.T) {
	pool := testdb.Pool(t)
	driverID := insertTestDriver(t, pool)
	assignDriver(t, pool, driverID, formulaOneTeamID, seasonStart)
	assignDriver(t, pool, driverID, formulaTwoTeamID, midSeason)

	if err := insertSeasonDriver(pool, driverID, formulaOneCategoryID); err != nil {
		t.Fatalf("number in the first category: %v", err)
	}
	if err := insertSeasonDriver(pool, driverID, formulaTwoCategoryID); err != nil {
		t.Fatalf("number in the second category: %v", err)
	}
}

// TestSeasonDriverRejectsTwoNumbersInTheSameCategory asserts the unique key moved
// from (driver, season) to (driver, season, category).
func TestSeasonDriverRejectsTwoNumbersInTheSameCategory(t *testing.T) {
	pool := testdb.Pool(t)
	driverID := insertTestDriver(t, pool)
	assignDriver(t, pool, driverID, formulaOneTeamID, seasonStart)

	if err := insertSeasonDriver(pool, driverID, formulaOneCategoryID); err != nil {
		t.Fatalf("first number: %v", err)
	}
	if _, isViolation := db.UniqueViolation(insertSeasonDriver(pool, driverID, formulaOneCategoryID)); !isViolation {
		t.Fatalf("second number in the same category should violate a unique key")
	}
}

func insertTestDriver(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()

	var driverID int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO drivers (first_name, last_name, license_number, country_code)
		 VALUES ('Test', 'Driver', $1, 'ar')
		 RETURNING id`,
		testdb.UniqueCode(t),
	).Scan(&driverID)
	if err != nil {
		t.Fatalf("insert driver: %v", err)
	}

	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(),
			`DELETE FROM season_drivers WHERE driver_id = $1`, driverID); err != nil {
			t.Errorf("clean up season_drivers: %v", err)
		}
	})
	return driverID
}

func assignDriver(t *testing.T, pool *pgxpool.Pool, driverID, teamID int64, effectiveFrom time.Time) {
	t.Helper()

	_, err := pool.Exec(context.Background(),
		`INSERT INTO driver_team_assignments (driver_id, team_id, effective_from, is_starter)
		 VALUES ($1, $2, $3, true)`,
		driverID, teamID, effectiveFrom)
	if err != nil {
		t.Fatalf("assign driver %d to team %d: %v", driverID, teamID, err)
	}
}

func insertSeasonDriver(pool *pgxpool.Pool, driverID, categoryID int64) error {
	_, err := pool.Exec(context.Background(),
		`INSERT INTO season_drivers (driver_id, season_id, category_id, number)
		 VALUES ($1, $2, $3, $4)`,
		driverID, demoSeasonID, categoryID, uniqueSeasonNumber())
	return err
}

// uniqueSeasonNumber stays above the seeded numbers and below the schema's 999
// ceiling, so the (season, category, number) unique key never collides.
func uniqueSeasonNumber() int32 {
	return 500 + seasonNumberSequence.Add(1)%400
}
