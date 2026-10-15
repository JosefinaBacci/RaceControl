package drivers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// PublicDriver is a directory entry: the person facts, the team they currently
// belong to and the number they hold in its championship.
type PublicDriver struct {
	ID           int64
	FirstName    string
	LastName     string
	CountryCode  string
	IsStarter    bool
	Number       *int32
	TeamID       int64
	TeamName     string
	CategoryCode string
	CategoryName string
}

// PublicFilter narrows the directory: a category code and a name search, both
// optional.
type PublicFilter struct {
	Category string
	Search   string
}

// Directory lists the current lineup of every team, grouped by championship and
// team so the reader can render it as a directory.
func (s *Service) Directory(ctx context.Context, filter PublicFilter) ([]PublicDriver, error) {
	rows, err := s.queries.ListActiveDrivers(ctx, directoryParams(filter))
	if err != nil {
		return nil, fmt.Errorf("list drivers: %w", err)
	}
	drivers := make([]PublicDriver, 0, len(rows))
	for _, row := range rows {
		drivers = append(drivers, publicDriverFromRow(dbgen.ListActiveDriversRow(row)))
	}
	return drivers, nil
}

// PublicDriver answers one driver of the current lineup; a released driver and an
// unknown id are both not found.
func (s *Service) PublicDriver(ctx context.Context, id int64) (PublicDriver, error) {
	row, err := s.queries.GetActiveDriver(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicDriver{}, domain.ErrNotFound
	}
	if err != nil {
		return PublicDriver{}, fmt.Errorf("get driver %d: %w", id, err)
	}
	return publicDriverFromRow(dbgen.ListActiveDriversRow(row)), nil
}

func directoryParams(filter PublicFilter) dbgen.ListActiveDriversParams {
	var params dbgen.ListActiveDriversParams
	if category := strings.ToLower(strings.TrimSpace(filter.Category)); category != "" {
		params.CategoryCode = pgtype.Text{String: category, Valid: true}
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		params.Search = pgtype.Text{String: escapeLike(search), Valid: true}
	}
	return params
}

func publicDriverFromRow(row dbgen.ListActiveDriversRow) PublicDriver {
	return PublicDriver{
		ID:           row.ID,
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		CountryCode:  row.CountryCode,
		IsStarter:    row.IsStarter,
		Number:       db.Int32Pointer(row.Number),
		TeamID:       row.TeamID,
		TeamName:     row.TeamName,
		CategoryCode: row.CategoryCode,
		CategoryName: row.CategoryName,
	}
}

// A name may contain "_", which LIKE would read as a wildcard.
func escapeLike(search string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(search)
}
