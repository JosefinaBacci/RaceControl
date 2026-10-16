package scores

import (
	"context"
	"fmt"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// DriverStanding is a driver's accumulated score in a championship. The team is
// the one credited in their latest race.
type DriverStanding struct {
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
	Races       int32
}

// TeamStanding sums the points of every result credited to the team, whoever
// drove for it at the time.
type TeamStanding struct {
	Position int32
	TeamID   int64
	TeamCode string
	TeamName string
	Points   float64
}

type Standings struct {
	SeasonYear   int32
	CategoryCode string
	Drivers      []DriverStanding
	Teams        []TeamStanding
}

// Standings is a championship table, so it needs the category: points of
// different categories never add up.
func (s *Service) Standings(ctx context.Context, filter Filter) (Standings, error) {
	category := normalizedCategory(filter.Category)
	if category == "" {
		return Standings{}, domain.NewValidationError("category", "la categoría es obligatoria")
	}
	year, err := s.seasonYear(ctx, filter.Season)
	if err != nil {
		return Standings{}, err
	}

	drivers, err := s.driverStandings(ctx, dbgen.DriverStandingsParams{SeasonYear: year, CategoryCode: category})
	if err != nil {
		return Standings{}, err
	}
	teams, err := s.teamStandings(ctx, dbgen.TeamStandingsParams{SeasonYear: year, CategoryCode: category})
	if err != nil {
		return Standings{}, err
	}
	return Standings{SeasonYear: year, CategoryCode: category, Drivers: drivers, Teams: teams}, nil
}

func (s *Service) driverStandings(ctx context.Context, params dbgen.DriverStandingsParams) ([]DriverStanding, error) {
	rows, err := s.queries.DriverStandings(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("driver standings of %s %d: %w", params.CategoryCode, params.SeasonYear, err)
	}
	standings := make([]DriverStanding, 0, len(rows))
	for _, row := range rows {
		standings = append(standings, DriverStanding{
			Position:    row.Position,
			DriverID:    row.DriverID,
			FirstName:   row.FirstName,
			LastName:    row.LastName,
			CountryCode: row.CountryCode,
			Number:      db.Int32Pointer(row.DriverNumber),
			TeamID:      row.TeamID,
			TeamCode:    row.TeamCode,
			TeamName:    row.TeamName,
			Points:      row.Points,
			Races:       row.Races,
		})
	}
	return standings, nil
}

func (s *Service) teamStandings(ctx context.Context, params dbgen.TeamStandingsParams) ([]TeamStanding, error) {
	rows, err := s.queries.TeamStandings(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("team standings of %s %d: %w", params.CategoryCode, params.SeasonYear, err)
	}
	standings := make([]TeamStanding, 0, len(rows))
	for _, row := range rows {
		standings = append(standings, TeamStanding{
			Position: row.Position,
			TeamID:   row.TeamID,
			TeamCode: row.TeamCode,
			TeamName: row.TeamName,
			Points:   row.Points,
		})
	}
	return standings, nil
}
