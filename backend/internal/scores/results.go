package scores

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/calendar"
	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// DriverResult is one line of a race: drivers on equal points share a position.
type DriverResult struct {
	Position    int32
	DriverID    int64
	FirstName   string
	LastName    string
	CountryCode string
	Number      *int32
	TeamID      int64
	TeamCode    string
	TeamName    string
	Points      float64
}

// RaceResults is a race with the results the FIA loaded for it.
type RaceResults struct {
	Race    Race
	Results []DriverResult
}

// Results lists the races of the championship that already have results, in
// calendar order.
func (s *Service) Results(ctx context.Context, filter Filter) ([]RaceResults, error) {
	year, err := s.seasonYear(ctx, filter.Season)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListRaceResults(ctx, dbgen.ListRaceResultsParams{
		SeasonYear:   db.Int4(year),
		CategoryCode: optionalCategory(filter.Category),
	})
	if err != nil {
		return nil, fmt.Errorf("list race results of season %d: %w", year, err)
	}
	return groupByRace(rows), nil
}

// RaceResults answers one race, with no results when none were loaded yet. An
// unknown id, another kind of event and a cancelled race are all not found.
func (s *Service) RaceResults(ctx context.Context, eventID int64) (RaceResults, error) {
	event, err := s.queries.GetCalendarEvent(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return RaceResults{}, domain.ErrNotFound
	}
	if err != nil {
		return RaceResults{}, fmt.Errorf("get race %d: %w", eventID, err)
	}
	if calendar.Kind(event.Kind) != calendar.KindRace || !event.IsActive {
		return RaceResults{}, domain.ErrNotFound
	}

	rows, err := s.queries.ListRaceResults(ctx, dbgen.ListRaceResultsParams{EventID: db.Int8(eventID)})
	if err != nil {
		return RaceResults{}, fmt.Errorf("list results of race %d: %w", eventID, err)
	}
	results := make([]DriverResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, driverResultFromRow(row))
	}
	return RaceResults{Race: raceFromEvent(event), Results: results}, nil
}

// groupByRace folds the rows, which arrive ordered by race, into one entry per race.
func groupByRace(rows []dbgen.ListRaceResultsRow) []RaceResults {
	races := make([]RaceResults, 0)
	for _, row := range rows {
		if len(races) == 0 || races[len(races)-1].Race.ID != row.CalendarEvent.ID {
			races = append(races, RaceResults{Race: raceFromEvent(row.CalendarEvent)})
		}
		current := &races[len(races)-1]
		current.Results = append(current.Results, driverResultFromRow(row))
	}
	return races
}

func driverResultFromRow(row dbgen.ListRaceResultsRow) DriverResult {
	return DriverResult{
		Position:    row.Position,
		DriverID:    row.DriverID,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		CountryCode: row.DriverCountryCode,
		Number:      db.Int32Pointer(row.DriverNumber),
		TeamID:      row.TeamID,
		TeamCode:    row.TeamCode,
		TeamName:    row.TeamName,
		Points:      row.Points,
	}
}
