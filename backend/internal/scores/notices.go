package scores

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// NoticeKind tells a team whether a change published the race or corrected it.
type NoticeKind string

const (
	KindPublished NoticeKind = "published"
	KindCorrected NoticeKind = "corrected"
)

// Status is pending until the team confirms it received the change.
type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
)

// Acknowledgement is who confirmed a change for a team, and when. Both are nil
// while the change is pending.
type Acknowledgement struct {
	Status Status
	At     *time.Time
	By     *string
}

// ChangedPoints is the before and after of one driver in a change. Old is nil
// when the change published the driver's first score in the race.
type ChangedPoints struct {
	DriverID  int64
	FirstName string
	LastName  string
	OldPoints *float64
	NewPoints *float64
}

// Notice is a score change as the team it touched sees it.
type Notice struct {
	ChangeLogID     int64
	Race            Race
	Kind            NoticeKind
	ChangedAt       time.Time
	Acknowledgement Acknowledgement
	Items           []ChangedPoints
}

// NoticeBoard is a team's notices, newest first, with how many await confirmation.
type NoticeBoard struct {
	Notices      []Notice
	PendingCount int
}

// TeamAcknowledgement is where one affected team stands on a change.
type TeamAcknowledgement struct {
	TeamID          int64
	TeamCode        string
	TeamName        string
	Acknowledgement Acknowledgement
}

// ScoreChange is a change as the FIA follows it: every team it touched and
// whether each one confirmed.
type ScoreChange struct {
	ChangeLogID int64
	Race        Race
	Kind        NoticeKind
	ChangedAt   time.Time
	ChangedBy   string
	Teams       []TeamAcknowledgement
}

const acknowledgementGuard = "RC005"

// TeamNotices lists the score changes that touched the team.
func (s *Service) TeamNotices(ctx context.Context, teamID int64) (NoticeBoard, error) {
	notices, err := s.teamNotices(ctx, teamID, pgtype.Int8{})
	if err != nil {
		return NoticeBoard{}, err
	}
	return NoticeBoard{Notices: notices, PendingCount: countPending(notices)}, nil
}

// Acknowledge records that the caller's team received the change. The database
// derives the team from the user and refuses a change that did not touch it
// (000023); confirming again keeps the first confirmation.
func (s *Service) Acknowledge(ctx context.Context, identity domain.Identity, changeLogID int64) (Notice, error) {
	if identity.TeamID == nil {
		return Notice{}, domain.ErrForbidden
	}
	err := s.queries.AcknowledgeScoreChange(ctx, dbgen.AcknowledgeScoreChangeParams{
		ChangeLogID:    changeLogID,
		AcknowledgedBy: identity.UserID,
	})
	if err != nil {
		return Notice{}, translateAcknowledgeError(err, changeLogID)
	}
	return s.teamNotice(ctx, *identity.TeamID, changeLogID)
}

// translateAcknowledgeError answers RC005 as not found: the change is unknown or
// belongs to other teams, and telling the two apart would reveal another team's
// change.
func translateAcknowledgeError(err error, changeLogID int64) error {
	if code, isGuard := db.GuardViolation(err); isGuard && code == acknowledgementGuard {
		return domain.ErrNotFound
	}
	return fmt.Errorf("acknowledge score change %d: %w", changeLogID, err)
}

func (s *Service) teamNotice(ctx context.Context, teamID, changeLogID int64) (Notice, error) {
	notices, err := s.teamNotices(ctx, teamID, db.Int8(changeLogID))
	if err != nil {
		return Notice{}, err
	}
	if len(notices) == 0 {
		return Notice{}, domain.ErrNotFound
	}
	return notices[0], nil
}

func (s *Service) teamNotices(ctx context.Context, teamID int64, changeLogID pgtype.Int8) ([]Notice, error) {
	rows, err := s.queries.ListTeamScoreNotices(ctx, dbgen.ListTeamScoreNoticesParams{TeamID: teamID, ChangeLogID: changeLogID})
	if err != nil {
		return nil, fmt.Errorf("list score notices of team %d: %w", teamID, err)
	}
	itemsByChange, err := s.teamItems(ctx, teamID, changeLogID)
	if err != nil {
		return nil, err
	}

	notices := make([]Notice, 0, len(rows))
	for _, row := range rows {
		notices = append(notices, Notice{
			ChangeLogID:     row.ChangeLogID,
			Race:            raceFromEvent(row.CalendarEvent),
			Kind:            kindOf(row.IsPublication),
			ChangedAt:       row.ChangedAt.Time,
			Acknowledgement: acknowledgementOf(row.AcknowledgedAt, row.AcknowledgedBy),
			Items:           itemsByChange[row.ChangeLogID],
		})
	}
	return notices, nil
}

// teamItems answers only the team's own lines of each change: another team's
// points are none of its business.
func (s *Service) teamItems(ctx context.Context, teamID int64, changeLogID pgtype.Int8) (map[int64][]ChangedPoints, error) {
	rows, err := s.queries.ListTeamScoreChangeItems(ctx, dbgen.ListTeamScoreChangeItemsParams{TeamID: teamID, ChangeLogID: changeLogID})
	if err != nil {
		return nil, fmt.Errorf("list score change items of team %d: %w", teamID, err)
	}
	itemsByChange := make(map[int64][]ChangedPoints)
	for _, row := range rows {
		itemsByChange[row.ChangeLogID] = append(itemsByChange[row.ChangeLogID], changedPointsFromRow(row))
	}
	return itemsByChange, nil
}

func changedPointsFromRow(row dbgen.ListTeamScoreChangeItemsRow) ChangedPoints {
	return ChangedPoints{
		DriverID:  row.DriverID,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		OldPoints: row.OldPoints,
		NewPoints: row.NewPoints,
	}
}

// Changes lists every score change, newest first, with the confirmation of each
// team it touched.
func (s *Service) Changes(ctx context.Context) ([]ScoreChange, error) {
	rows, err := s.queries.ListScoreChangeAcknowledgements(ctx)
	if err != nil {
		return nil, fmt.Errorf("list score change acknowledgements: %w", err)
	}

	changes := make([]ScoreChange, 0)
	for _, row := range rows {
		if len(changes) == 0 || changes[len(changes)-1].ChangeLogID != row.ChangeLogID {
			changes = append(changes, ScoreChange{
				ChangeLogID: row.ChangeLogID,
				Race:        raceFromEvent(row.CalendarEvent),
				Kind:        kindOf(row.IsPublication),
				ChangedAt:   row.ChangedAt.Time,
				ChangedBy:   row.ChangedBy,
			})
		}
		current := &changes[len(changes)-1]
		current.Teams = append(current.Teams, TeamAcknowledgement{
			TeamID:          row.TeamID,
			TeamCode:        row.TeamCode,
			TeamName:        row.TeamName,
			Acknowledgement: acknowledgementOf(row.AcknowledgedAt, row.AcknowledgedBy),
		})
	}
	return changes, nil
}

func kindOf(isPublication bool) NoticeKind {
	if isPublication {
		return KindPublished
	}
	return KindCorrected
}

func acknowledgementOf(at pgtype.Timestamptz, by pgtype.Text) Acknowledgement {
	if !at.Valid {
		return Acknowledgement{Status: StatusPending}
	}
	return Acknowledgement{Status: StatusConfirmed, At: db.TimePointer(at), By: db.TextPointer(by)}
}

func countPending(notices []Notice) int {
	pending := 0
	for _, notice := range notices {
		if notice.Acknowledgement.Status == StatusPending {
			pending++
		}
	}
	return pending
}
