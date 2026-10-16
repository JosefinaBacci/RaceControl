package drivers

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

const (
	licenseConstraint = "drivers_license_number_key"
	numberConstraint  = "season_drivers_unique_per_season_category"
)

// Transactor runs the multi-row writes (a person and their first assignment, a
// role change that inserts a new period) in a single transaction.
type Transactor interface {
	InTx(ctx context.Context, work func(queries dbgen.Querier) error) error
}

type Service struct {
	queries    dbgen.Querier
	transactor Transactor
}

func NewService(queries dbgen.Querier, transactor Transactor) *Service {
	return &Service{queries: queries, transactor: transactor}
}

// Lineup lists the drivers a team holds right now, starters first.
func (s *Service) Lineup(ctx context.Context, teamID int64) ([]Driver, error) {
	rows, err := s.queries.ListTeamDrivers(ctx, db.Int8(teamID))
	if err != nil {
		return nil, fmt.Errorf("list lineup of team %d: %w", teamID, err)
	}
	drivers := make([]Driver, 0, len(rows))
	for _, row := range rows {
		drivers = append(drivers, driverFromRow(row.ID, row.FirstName, row.LastName, row.LicenseNumber, row.CountryCode, row.IsStarter, row.Number))
	}
	return drivers, nil
}

// Get answers a driver only when it belongs to the team: a stranger and an unknown
// id are both not found, so the team never learns about another team's driver.
func (s *Service) Get(ctx context.Context, teamID, id int64) (Driver, error) {
	row, err := s.queries.GetTeamDriver(ctx, dbgen.GetTeamDriverParams{TeamID: db.Int8(teamID), DriverID: id})
	if errors.Is(err, pgx.ErrNoRows) {
		return Driver{}, domain.ErrNotFound
	}
	if err != nil {
		return Driver{}, fmt.Errorf("get driver %d of team %d: %w", id, teamID, err)
	}
	return driverFromRow(row.ID, row.FirstName, row.LastName, row.LicenseNumber, row.CountryCode, row.IsStarter, row.Number), nil
}

// Create files a new driver and their first assignment to the team, plus the
// season number when one is given.
func (s *Service) Create(ctx context.Context, teamID int64, input NewDriver) (Driver, error) {
	state := newState(input)
	if err := validateState(state); err != nil {
		return Driver{}, err
	}

	var id int64
	err := s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		created, err := queries.CreateDriver(ctx, dbgen.CreateDriverParams{
			FirstName:     state.firstName,
			LastName:      state.lastName,
			LicenseNumber: state.licenseNumber,
			CountryCode:   state.countryCode,
		})
		if err != nil {
			return translateWriteError(err)
		}
		id = created
		if err := queries.CreateAssignment(ctx, dbgen.CreateAssignmentParams{
			DriverID:  id,
			TeamID:    db.Int8(teamID),
			IsStarter: state.isStarter,
		}); err != nil {
			return translateWriteError(err)
		}
		if state.number == nil {
			return nil
		}
		return storeNumber(ctx, queries, id, teamID, *state.number)
	})
	if err != nil {
		return Driver{}, err
	}
	return s.Get(ctx, teamID, id)
}

// Update edits the person facts and, when the role changed, closes the current
// membership period with a new one: the role rides on the assignment, never on an
// edit (000011).
func (s *Service) Update(ctx context.Context, teamID, id int64, changes DriverChanges) (Driver, error) {
	current, err := s.Get(ctx, teamID, id)
	if err != nil {
		return Driver{}, err
	}
	next := driverState{
		firstName:     current.FirstName,
		lastName:      current.LastName,
		licenseNumber: current.LicenseNumber,
		countryCode:   current.CountryCode,
		isStarter:     current.IsStarter,
		number:        current.Number,
	}.withChanges(changes)
	if err := validateState(next); err != nil {
		return Driver{}, err
	}

	err = s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		if err := queries.UpdateDriver(ctx, dbgen.UpdateDriverParams{
			ID:            id,
			FirstName:     next.firstName,
			LastName:      next.lastName,
			LicenseNumber: next.licenseNumber,
			CountryCode:   next.countryCode,
		}); err != nil {
			return translateWriteError(err)
		}
		if changes.IsStarter != nil && *changes.IsStarter != current.IsStarter {
			if err := queries.CreateAssignment(ctx, dbgen.CreateAssignmentParams{
				DriverID:  id,
				TeamID:    db.Int8(teamID),
				IsStarter: *changes.IsStarter,
			}); err != nil {
				return translateWriteError(err)
			}
		}
		if changes.Number != nil {
			return storeNumber(ctx, queries, id, teamID, *changes.Number)
		}
		return nil
	})
	if err != nil {
		return Driver{}, err
	}
	return s.Get(ctx, teamID, id)
}

// Release ends the membership: a new period without a team leaves the driver out
// of the lineup from now on, and the history stays append-only (000022).
func (s *Service) Release(ctx context.Context, teamID, id int64) error {
	if _, err := s.Get(ctx, teamID, id); err != nil {
		return err
	}
	return s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		if err := queries.ReleaseAssignment(ctx, id); err != nil {
			return translateWriteError(err)
		}
		return nil
	})
}

// storeNumber files the number for the latest season under the team's category.
// The category is the team's, never the caller's choice, and a trigger checks the
// driver really raced in it during the season (000019).
func storeNumber(ctx context.Context, queries dbgen.Querier, driverID, teamID int64, number int32) error {
	seasonID, err := latestSeasonID(ctx, queries)
	if err != nil {
		return err
	}
	categoryID, err := teamCategoryID(ctx, queries, teamID)
	if err != nil {
		return err
	}
	if err := queries.UpsertSeasonDriver(ctx, dbgen.UpsertSeasonDriverParams{
		DriverID:   driverID,
		SeasonID:   seasonID,
		CategoryID: categoryID,
		Number:     number,
	}); err != nil {
		return translateWriteError(err)
	}
	return nil
}

func latestSeasonID(ctx context.Context, queries dbgen.Querier) (int64, error) {
	season, err := queries.LatestSeason(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.NewValidationError("number", "no hay ninguna temporada cargada")
	}
	if err != nil {
		return 0, fmt.Errorf("load latest season: %w", err)
	}
	return season.ID, nil
}

func teamCategoryID(ctx context.Context, queries dbgen.Querier, teamID int64) (int64, error) {
	categoryID, err := queries.TeamCategory(ctx, teamID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("load category of team %d: %w", teamID, err)
	}
	return categoryID, nil
}

// translateWriteError turns what the schema refuses into the domain error the
// transport maps. Anything else is a genuine failure and surfaces as a 500.
func translateWriteError(err error) error {
	if constraint, isViolation := db.UniqueViolation(err); isViolation {
		switch constraint {
		case licenseConstraint:
			return domain.NewConflictError("licenseNumber", "ya hay un piloto con esa licencia")
		case numberConstraint:
			return domain.NewConflictError("number", "ya hay un piloto con ese número en la categoría")
		default:
			return domain.NewConflictError("", "el piloto ya existe")
		}
	}
	if code, isGuard := db.GuardViolation(err); isGuard {
		switch code {
		case "RC003":
			return domain.NewConflictError("number", "el piloto no tiene una asignación en esa categoría durante la temporada")
		case "RC004":
			return domain.NewConflictError("", "el historial de asignaciones solo puede extenderse hacia adelante")
		default:
			return fmt.Errorf("schema guard %s: %w", code, err)
		}
	}
	return err
}
