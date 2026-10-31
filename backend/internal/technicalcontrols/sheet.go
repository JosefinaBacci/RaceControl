package technicalcontrols

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// Control is the calendar event a sheet grades.
type Control struct {
	EventID      int64
	ControlType  string
	StartsAt     time.Time
	EndsAt       time.Time
	IsActive     bool
	SeasonYear   int32
	CategoryCode string
	CategoryName string
}

// GradedTeam is a team of the control's championship with every result it was
// given there, oldest first. No revisions means it is still pending.
type GradedTeam struct {
	Team
	IsActive  bool
	Revisions []Result
}

// Sheet is what the FIA works on: the control and each team's chain of results.
type Sheet struct {
	Control Control
	Teams   []GradedTeam
}

func (s *Service) Sheet(ctx context.Context, eventID int64) (Sheet, error) {
	control, err := s.queries.GetTechnicalControl(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Sheet{}, domain.ErrNotFound
	}
	if err != nil {
		return Sheet{}, fmt.Errorf("get technical control %d: %w", eventID, err)
	}
	teams, err := s.queries.ListControlledTeams(ctx, eventID)
	if err != nil {
		return Sheet{}, fmt.Errorf("list teams of control %d: %w", eventID, err)
	}
	revisions, err := s.queries.ListTechnicalControlRevisions(ctx, eventID)
	if err != nil {
		return Sheet{}, fmt.Errorf("list results of control %d: %w", eventID, err)
	}
	return Sheet{Control: controlFromRow(control), Teams: gradedTeams(teams, revisions)}, nil
}

func gradedTeams(teams []dbgen.ListControlledTeamsRow, revisions []dbgen.TechnicalControlRevision) []GradedTeam {
	revisionsByTeam := make(map[int64][]Result, len(teams))
	for _, revision := range revisions {
		revisionsByTeam[revision.TeamID] = append(revisionsByTeam[revision.TeamID], resultFromRow(revision))
	}
	graded := make([]GradedTeam, 0, len(teams))
	for _, team := range teams {
		graded = append(graded, GradedTeam{
			Team:      Team{ID: team.ID, Name: team.Name},
			IsActive:  team.IsActive,
			Revisions: revisionsByTeam[team.ID],
		})
	}
	return graded
}

func controlFromRow(row dbgen.GetTechnicalControlRow) Control {
	return Control{
		EventID:      row.EventID,
		ControlType:  row.ControlType,
		StartsAt:     row.StartsAt.Time,
		EndsAt:       row.EndsAt.Time,
		IsActive:     row.IsActive,
		SeasonYear:   row.SeasonYear,
		CategoryCode: row.CategoryCode,
		CategoryName: row.CategoryName,
	}
}
