package scores

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// Transactor runs the score rewrite and its audit trail in a single transaction.
type Transactor interface {
	InTx(ctx context.Context, work func(queries dbgen.Querier) error) error
}

// Service reads the results, the standings and the score notices, and rewrites a
// race's scores. None of the reads is stored: every position, total and notice
// status is derived from the score entries, the change log and the
// acknowledgements when it is read.
type Service struct {
	queries    dbgen.Querier
	transactor Transactor
}

func NewService(queries dbgen.Querier, transactor Transactor) *Service {
	return &Service{queries: queries, transactor: transactor}
}

// GridEntry is a driver a race can score, with what is already loaded (nil when
// the driver has no entry yet).
type GridEntry struct {
	DriverID    int64
	FirstName   string
	LastName    string
	CountryCode string
	TeamID      int64
	TeamName    string
	Number      *int32
	Points      *float64
}

type Grid struct {
	EventID      int64
	CategoryCode string
	CategoryName string
	RaceEndsAt   time.Time
	Entries      []GridEntry
}

// Entry is one row the admin submits. A nil Points is an empty value and is
// rejected, so a blank field is never read as zero.
type Entry struct {
	DriverID int64
	Points   *float64
}

// Grid is the race's score sheet: its championship and every driver that can be
// scored in it.
func (s *Service) Grid(ctx context.Context, eventID int64) (Grid, error) {
	return loadGrid(ctx, s.queries, eventID)
}

func loadGrid(ctx context.Context, queries dbgen.Querier, eventID int64) (Grid, error) {
	championship, err := raceChampionship(ctx, queries, eventID)
	if err != nil {
		return Grid{}, err
	}
	rows, err := queries.ListRaceEligibleDrivers(ctx, eventID)
	if err != nil {
		return Grid{}, fmt.Errorf("list drivers of race %d: %w", eventID, err)
	}
	grid := Grid{
		EventID:      eventID,
		CategoryCode: championship.CategoryCode,
		CategoryName: championship.CategoryName,
		RaceEndsAt:   championship.EndsAt.Time,
	}
	grid.Entries = make([]GridEntry, 0, len(rows))
	for _, row := range rows {
		grid.Entries = append(grid.Entries, GridEntry{
			DriverID:    row.ID,
			FirstName:   row.FirstName,
			LastName:    row.LastName,
			CountryCode: row.CountryCode,
			TeamID:      row.TeamID,
			TeamName:    row.TeamName,
			Number:      db.Int32Pointer(row.Number),
			Points:      row.Points,
		})
	}
	return grid, nil
}

// Replace sets the scores of a race to the submitted grid: it upserts the listed
// drivers, removes the ones left out and records the whole edit as one change,
// attributed to the actor, in a single transaction.
func (s *Service) Replace(ctx context.Context, eventID, actorUserID int64, entries []Entry) (Grid, error) {
	edit := scoreEdit{eventID: eventID, actorUserID: actorUserID, entries: entries}
	err := s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		return replaceScores(ctx, queries, edit)
	})
	if err != nil {
		return Grid{}, translateWriteError(err)
	}
	return s.Grid(ctx, eventID)
}

type scoreEdit struct {
	eventID     int64
	actorUserID int64
	entries     []Entry
}

// replaceScores diffs against the grid read under the race lock, so the audit
// records the points the edit actually replaced. A race takes results only once
// it has finished: results freeze its date (RC001), so loading them earlier would
// lock the calendar on a guess.
func replaceScores(ctx context.Context, queries dbgen.Querier, edit scoreEdit) error {
	if err := queries.LockRaceScores(ctx, edit.eventID); err != nil {
		return fmt.Errorf("lock scores of race %d: %w", edit.eventID, err)
	}
	grid, err := loadGrid(ctx, queries, edit.eventID)
	if err != nil {
		return err
	}
	if grid.RaceEndsAt.After(time.Now()) {
		return domain.NewConflictError("eventId", "la carrera todavía no terminó: no tiene resultados para cargar")
	}
	changes, err := diff(grid, edit.entries)
	if err != nil {
		return err
	}
	if len(changes) == 0 {
		return nil
	}
	return recordChanges(ctx, queries, edit, changes)
}

func recordChanges(ctx context.Context, queries dbgen.Querier, edit scoreEdit, changes []change) error {
	if err := queries.SetScoreActor(ctx, strconv.FormatInt(edit.actorUserID, 10)); err != nil {
		return fmt.Errorf("declare score actor: %w", err)
	}
	logID, err := queries.CreateScoreChangeLog(ctx, dbgen.CreateScoreChangeLogParams{
		EventID:     edit.eventID,
		ActorUserID: edit.actorUserID,
	})
	if err != nil {
		return fmt.Errorf("record change of race %d: %w", edit.eventID, err)
	}
	for _, change := range changes {
		if err := applyChange(ctx, queries, edit.eventID, logID, change); err != nil {
			return err
		}
	}
	return nil
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

func raceChampionship(ctx context.Context, queries dbgen.Querier, eventID int64) (dbgen.GetRaceChampionshipRow, error) {
	row, err := queries.GetRaceChampionship(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return dbgen.GetRaceChampionshipRow{}, domain.ErrNotFound
	}
	if err != nil {
		return dbgen.GetRaceChampionshipRow{}, fmt.Errorf("load championship of event %d: %w", eventID, err)
	}
	return row, nil
}

// change is one driver's move in an edit: from old to new points, either side nil
// for a row that appears or disappears.
type change struct {
	driverID int64
	old      *float64
	new      *float64
}

// diff turns the submitted grid into the minimal set of changes, touching each
// driver once so the audit lists every driver a single time.
func diff(grid Grid, entries []Entry) ([]change, error) {
	current := make(map[int64]*float64, len(grid.Entries))
	for _, entry := range grid.Entries {
		current[entry.DriverID] = entry.Points
	}

	desired := make(map[int64]float64, len(entries))
	order := make([]int64, 0, len(entries))
	for _, entry := range entries {
		if _, eligible := current[entry.DriverID]; !eligible {
			return nil, domain.NewValidationError("entries",
				fmt.Sprintf("el piloto %d no corre en la categoría de la carrera", entry.DriverID))
		}
		if _, repeated := desired[entry.DriverID]; repeated {
			return nil, domain.NewValidationError("entries", "no se puede repetir un piloto en la grilla")
		}
		if err := validatePoints(entry.Points); err != nil {
			return nil, err
		}
		desired[entry.DriverID] = *entry.Points
		order = append(order, entry.DriverID)
	}

	changes := make([]change, 0, len(entries))
	for _, driverID := range order {
		points := desired[driverID]
		old := current[driverID]
		switch {
		case old == nil:
			changes = append(changes, change{driverID: driverID, new: pointerTo(points)})
		case *old != points:
			changes = append(changes, change{driverID: driverID, old: old, new: pointerTo(points)})
		}
	}
	for _, entry := range grid.Entries {
		if _, submitted := desired[entry.DriverID]; submitted || entry.Points == nil {
			continue
		}
		changes = append(changes, change{driverID: entry.DriverID, old: entry.Points})
	}
	return changes, nil
}

// maxPoints is the largest value score_entries.points (numeric(6, 2)) holds, and
// the column keeps two decimals, so a third one would be rounded away unnoticed.
const (
	maxPoints        = 9999.99
	pointsScale      = 100
	decimalTolerance = 1e-9
)

func validatePoints(points *float64) error {
	switch {
	case points == nil:
		return domain.NewValidationError("entries", "cada piloto necesita sus puntos")
	case *points < 0:
		return domain.NewValidationError("entries", "los puntos no pueden ser negativos")
	case *points > maxPoints:
		return domain.NewValidationError("entries", fmt.Sprintf("los puntos no pueden superar %.2f", maxPoints))
	case !hasAtMostTwoDecimals(*points):
		return domain.NewValidationError("entries", "los puntos admiten hasta dos decimales")
	}
	return nil
}

func hasAtMostTwoDecimals(points float64) bool {
	scaled := points * pointsScale
	return math.Abs(scaled-math.Round(scaled)) < decimalTolerance
}

func applyChange(ctx context.Context, queries dbgen.Querier, eventID, logID int64, change change) error {
	if change.new == nil {
		if err := queries.DeleteScoreEntry(ctx, dbgen.DeleteScoreEntryParams{
			EventID:  eventID,
			DriverID: change.driverID,
		}); err != nil {
			return fmt.Errorf("remove score of driver %d: %w", change.driverID, err)
		}
	} else if err := queries.UpsertScoreEntry(ctx, dbgen.UpsertScoreEntryParams{
		EventID:  eventID,
		DriverID: change.driverID,
		Points:   *change.new,
	}); err != nil {
		return fmt.Errorf("set score of driver %d: %w", change.driverID, err)
	}

	if err := queries.CreateScoreChangeLogItem(ctx, dbgen.CreateScoreChangeLogItemParams{
		ChangeLogID: logID,
		DriverID:    change.driverID,
		OldPoints:   change.old,
		NewPoints:   change.new,
	}); err != nil {
		return fmt.Errorf("record score item of driver %d: %w", change.driverID, err)
	}
	return nil
}

func pointerTo(value float64) *float64 {
	return &value
}

// translateWriteError turns what the schema refuses into the domain error the
// transport maps: a repeated driver, a driver of another category (RC002). The
// missing-actor guard (RC006) is a bug here, since the service always declares the
// actor, so it surfaces as an internal error.
func translateWriteError(err error) error {
	if constraint, isViolation := db.UniqueViolation(err); isViolation {
		if constraint == "score_entries_driver_unique_per_event" {
			return domain.NewConflictError("entries", "ya hay un puntaje para ese piloto en la carrera")
		}
		return domain.NewConflictError("", "el puntaje ya existe")
	}
	if code, isGuard := db.GuardViolation(err); isGuard {
		switch code {
		case "RC002":
			return domain.NewValidationError("entries", "el piloto no corre en la categoría de la carrera")
		default:
			return fmt.Errorf("schema guard %s: %w", code, err)
		}
	}
	return err
}
