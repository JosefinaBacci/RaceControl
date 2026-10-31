package dbtest

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/tests/testdb"
)

type controlChain struct {
	pool       *pgxpool.Pool
	eventID    int64
	categoryID int64
	teamID     int64
	otherTeam  int64
	authorID   int64
	originalID int64
}

func newControlChain(t *testing.T) controlChain {
	t.Helper()
	pool := testdb.Pool(t)
	championship := testdb.NewChampionship(t, pool)
	start := time.Date(2026, 3, 5, 9, 0, 0, 0, time.UTC)
	chain := controlChain{
		pool:       pool,
		eventID:    championship.AddTechnicalControl("Pesaje", start, start.Add(time.Hour)),
		categoryID: championship.CategoryID,
		teamID:     championship.AddTeam("Equipo A"),
		otherTeam:  championship.AddTeam("Equipo B"),
	}
	ctx := context.Background()
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash, role) VALUES ($1, 'not-a-real-hash', 'fia_admin') RETURNING id`,
		testdb.UniqueUsername(t)).Scan(&chain.authorID); err != nil {
		t.Fatalf("insert author: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO technical_control_results (event_id, category_id, team_id, conducted_on, passed, recorded_by)
		 VALUES ($1, $2, $3, '2026-03-05', true, $4) RETURNING id`,
		chain.eventID, chain.categoryID, chain.teamID, chain.authorID).Scan(&chain.originalID); err != nil {
		t.Fatalf("insert original result: %v", err)
	}
	return chain
}

const insertResult = `INSERT INTO technical_control_results
    (event_id, category_id, team_id, conducted_on, passed, notes, supersedes_id, correction_note, recorded_by)
    VALUES ($1, $2, $3, '2026-03-05', false, 'peso', $4, $5, $6)`

func (c controlChain) correctionArgs(teamID int64, supersedesID, reason any) []any {
	return []any{c.eventID, c.categoryID, teamID, supersedesID, reason, c.authorID}
}

// TestControlResultChainStaysLinear asserts the schema keeps one result in force
// per team: a correction names its reason and the same team, and nothing is
// graded twice from scratch or superseded twice.
func TestControlResultChainStaysLinear(t *testing.T) {
	chain := newControlChain(t)
	const reason = "cargado por error"

	refused := map[string][]any{
		"a correction without a reason": chain.correctionArgs(chain.teamID, chain.originalID, nil),
		"a reason without a correction": chain.correctionArgs(chain.otherTeam, nil, reason),
		"a correction of another team":  chain.correctionArgs(chain.otherTeam, chain.originalID, reason),
		"a second original":             chain.correctionArgs(chain.teamID, nil, nil),
	}
	for what, args := range refused {
		if err := execRolledBack(t, chain.pool, insertResult, args...); err == nil {
			t.Errorf("%s was accepted", what)
		}
	}

	ctx := context.Background()
	tx, err := chain.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	correction := chain.correctionArgs(chain.teamID, chain.originalID, reason)
	if _, err := tx.Exec(ctx, insertResult, correction...); err != nil {
		t.Fatalf("first correction: %v", err)
	}
	if _, err := tx.Exec(ctx, insertResult, correction...); err == nil {
		t.Error("a result was superseded twice")
	}
}

// TestControlWithResultsIsFinal covers RC020 on both tables a control spans and
// lets a write that changes nothing through.
func TestControlWithResultsIsFinal(t *testing.T) {
	chain := newControlChain(t)

	for _, query := range []string{
		`UPDATE events SET ends_at = ends_at + interval '1 day' WHERE id = $1`,
		`UPDATE events SET deactivated_at = now() WHERE id = $1`,
		`UPDATE technical_controls SET control_type = 'Otro' WHERE event_id = $1`,
	} {
		code, isGuard := db.GuardViolation(execRolledBack(t, chain.pool, query, chain.eventID))
		if !isGuard || code != "RC020" {
			t.Errorf("%q = (%q, %t), want (\"RC020\", true)", query, code, isGuard)
		}
	}
	for _, query := range []string{
		`UPDATE events SET starts_at = starts_at, ends_at = ends_at WHERE id = $1`,
		`UPDATE technical_controls SET control_type = control_type WHERE event_id = $1`,
	} {
		if err := execRolledBack(t, chain.pool, query, chain.eventID); err != nil {
			t.Errorf("%q = %v, want a no-op write to pass", query, err)
		}
	}
}
