package transporttest

import (
	"context"
	"errors"
	"testing"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/sanctions"
)

type expiredNoticeRead struct{ dbgen.Querier }

func (expiredNoticeRead) ListTeamSanctionNotices(context.Context, dbgen.ListTeamSanctionNoticesParams) ([]dbgen.ListTeamSanctionNoticesRow, error) {
	return nil, context.DeadlineExceeded
}

type expireReceiptRead struct{ *db.Transactor }

func (tx expireReceiptRead) InTx(ctx context.Context, work func(dbgen.Querier) error) error {
	return tx.Transactor.InTx(ctx, func(q dbgen.Querier) error { return work(expiredNoticeRead{q}) })
}

func TestFailedReceiptReadLeavesSanctionPending(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	team, username := h.signedIn(domain.RoleTeamAdmin, fixture.teamA)
	created := fixture.issueOnAlonso()
	version := int64(currentVersionID(t, created.body))
	ctx := context.Background()
	actor := domain.Identity{TeamID: &fixture.teamA}
	if err := h.pool.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&actor.UserID); err != nil {
		t.Fatal(err)
	}
	service := sanctions.NewService(expiredNoticeRead{h.queries}, expireReceiptRead{db.NewTransactor(h.pool)})
	if _, err := service.Acknowledge(ctx, actor, version); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected receipt read deadline, got %v", err)
	}
	board := h.sanctionNotices(team)
	if board.body["pendingCount"] != float64(1) {
		t.Fatalf("receipt persisted after failed read: %v", board.body)
	}
}
