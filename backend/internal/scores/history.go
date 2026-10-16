package scores

import (
	"context"
	"fmt"
	"time"

	"github.com/racecontrol/backend/internal/db/dbgen"
)

// Revision is one change of a race's scores as the FIA audits it: when, who and
// the before and after of every driver it touched.
type Revision struct {
	ChangeLogID int64
	Kind        NoticeKind
	ChangedAt   time.Time
	ChangedBy   string
	Items       []ChangedPoints
}

// History lists the changes of a race's scores, newest first. The oldest one
// published the race and every later one corrected it.
func (s *Service) History(ctx context.Context, eventID int64) ([]Revision, error) {
	if _, err := raceChampionship(ctx, s.queries, eventID); err != nil {
		return nil, err
	}
	rows, err := s.queries.ListRaceScoreHistory(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("list score history of race %d: %w", eventID, err)
	}
	revisions := groupRevisions(rows)
	if len(revisions) > 0 {
		revisions[len(revisions)-1].Kind = KindPublished
	}
	return revisions, nil
}

func groupRevisions(rows []dbgen.ListRaceScoreHistoryRow) []Revision {
	revisions := make([]Revision, 0)
	for _, row := range rows {
		if len(revisions) == 0 || revisions[len(revisions)-1].ChangeLogID != row.ChangeLogID {
			revisions = append(revisions, Revision{
				ChangeLogID: row.ChangeLogID,
				Kind:        KindCorrected,
				ChangedAt:   row.ChangedAt.Time,
				ChangedBy:   row.ChangedBy,
			})
		}
		current := &revisions[len(revisions)-1]
		current.Items = append(current.Items, ChangedPoints{
			DriverID:  row.DriverID,
			FirstName: row.FirstName,
			LastName:  row.LastName,
			OldPoints: row.OldPoints,
			NewPoints: row.NewPoints,
		})
	}
	return revisions
}
