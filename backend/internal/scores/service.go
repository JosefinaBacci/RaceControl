package scores

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// Service reads the results, the standings and the score notices. None of them is
// stored: every position, total and notice status is derived from the score
// entries, the change log and the acknowledgements when it is read.
type Service struct {
	queries dbgen.Querier
}

func NewService(queries dbgen.Querier) *Service {
	return &Service{queries: queries}
}

// Filter picks a championship. The season is a year and defaults to the latest
// one; the category is a code such as f1.
type Filter struct {
	Season   string
	Category string
}

// seasonYear resolves the requested year, or the latest season when none is
// given. Without any season loaded it answers zero, which matches no result.
func (s *Service) seasonYear(ctx context.Context, raw string) (int32, error) {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		return parseSeason(raw)
	}
	season, err := s.queries.LatestSeason(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("load latest season: %w", err)
	}
	return season.Year, nil
}

func parseSeason(raw string) (int32, error) {
	year, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, domain.NewValidationError("season", "la temporada es un año, por ejemplo 2026")
	}
	return int32(year), nil
}

func normalizedCategory(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func optionalCategory(raw string) pgtype.Text {
	category := normalizedCategory(raw)
	return pgtype.Text{String: category, Valid: category != ""}
}

func presentYear(year int32) pgtype.Int4 {
	return pgtype.Int4{Int32: year, Valid: true}
}

func presentID(id int64) pgtype.Int8 {
	return pgtype.Int8{Int64: id, Valid: true}
}
