package tyretests

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/results"
)

var writeConstraints = results.Constraints{
	OneOriginal:    "tyre_test_results_one_original",
	SupersedesOnce: "tyre_test_results_supersedes_once",
}

type Service struct {
	queries dbgen.Querier
}

func NewService(queries dbgen.Querier) *Service {
	return &Service{queries: queries}
}

// Types lists the compounds of the catalogue: the types a tyre test can have.
func (s *Service) Types(ctx context.Context) ([]results.Type, error) {
	rows, err := s.queries.ListTyreCompounds(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tyre compounds: %w", err)
	}
	types := make([]results.Type, 0, len(rows))
	for _, row := range rows {
		types = append(types, results.Type{Code: row.Code, Name: row.Name})
	}
	return types, nil
}

// History lists the current result of every compound tested, newest first. The
// type is a compound of a fixed catalogue, so an unknown one is a client error
// rather than an empty history.
func (s *Service) History(ctx context.Context, filter domain.HistoryFilter) ([]Result, error) {
	compoundCode := pgtype.Text{}
	if filter.Type != "" {
		if _, err := s.compoundID(ctx, "type", filter.Type); err != nil {
			return nil, err
		}
		compoundCode = db.OptionalText(filter.Type)
	}
	rows, err := s.queries.ListCurrentTyreTestResults(ctx, dbgen.ListCurrentTyreTestResultsParams{
		FromDay:      db.OptionalDate(filter.Range.From),
		UntilDay:     db.OptionalDate(filter.Range.Until),
		CompoundCode: compoundCode,
	})
	if err != nil {
		return nil, fmt.Errorf("list tyre test results: %w", err)
	}
	return toResults(rows), nil
}

// Chain lists every result of a tyre test, corrections included, for the FIA.
func (s *Service) Chain(ctx context.Context, eventID int64) ([]Result, error) {
	if _, err := s.window(ctx, eventID); err != nil {
		return nil, err
	}
	rows, err := s.queries.ListTyreTestResultChain(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("list results of tyre test %d: %w", eventID, err)
	}
	return toResults(rows), nil
}

func (s *Service) Record(ctx context.Context, eventID, actorUserID int64, input NewResult) (Result, error) {
	window, err := s.window(ctx, eventID)
	if err != nil {
		return Result{}, err
	}
	compoundID, err := s.compoundID(ctx, "compoundCode", strings.TrimSpace(input.CompoundCode))
	if err != nil {
		return Result{}, err
	}
	outcome, err := results.ParseInput(input.Input, window, time.Now())
	if err != nil {
		return Result{}, err
	}
	return s.insert(ctx, dbgen.InsertTyreTestResultParams{
		EventID:     eventID,
		CompoundID:  compoundID,
		ConductedOn: db.Date(outcome.ConductedOn),
		Passed:      outcome.Passed,
		Notes:       outcome.Notes,
		RecordedBy:  actorUserID,
	})
}

// Correct supersedes a result with a new one for the same test and compound.
// Only the current result can be corrected; the schema refuses a second
// correction of the same row, which also settles two concurrent corrections.
func (s *Service) Correct(ctx context.Context, resultID, actorUserID int64, input Correction) (Result, error) {
	subject, err := s.queries.GetTyreTestResultSubject(ctx, resultID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, domain.ErrNotFound
	}
	if err != nil {
		return Result{}, fmt.Errorf("load tyre test result %d: %w", resultID, err)
	}
	window, err := s.window(ctx, subject.EventID)
	if err != nil {
		return Result{}, err
	}
	outcome, err := results.ParseInput(input.Input, window, time.Now())
	if err != nil {
		return Result{}, err
	}
	correctionNote, err := results.NormalizeCorrectionNote(input.CorrectionNote)
	if err != nil {
		return Result{}, err
	}
	return s.insert(ctx, dbgen.InsertTyreTestResultParams{
		EventID:        subject.EventID,
		CompoundID:     subject.CompoundID,
		ConductedOn:    db.Date(outcome.ConductedOn),
		Passed:         outcome.Passed,
		Notes:          outcome.Notes,
		SupersedesID:   db.Int8(resultID),
		CorrectionNote: db.OptionalText(correctionNote),
		RecordedBy:     actorUserID,
	})
}

func (s *Service) insert(ctx context.Context, params dbgen.InsertTyreTestResultParams) (Result, error) {
	id, err := s.queries.InsertTyreTestResult(ctx, params)
	if err != nil {
		return Result{}, results.TranslateWriteError(err, writeConstraints)
	}
	row, err := s.queries.GetTyreTestResultEntry(ctx, id)
	if err != nil {
		return Result{}, fmt.Errorf("load tyre test result %d: %w", id, err)
	}
	return toResult(row), nil
}

// window loads the timing of a tyre test. Any other event, or none, is not found:
// only a tyre test has tyre results.
func (s *Service) window(ctx context.Context, eventID int64) (results.EventWindow, error) {
	row, err := s.queries.GetTyreTestWindow(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return results.EventWindow{}, domain.ErrNotFound
	}
	if err != nil {
		return results.EventWindow{}, fmt.Errorf("load tyre test %d: %w", eventID, err)
	}
	return results.EventWindow{StartsAt: row.StartsAt.Time, EndsAt: row.EndsAt.Time, IsActive: row.IsActive}, nil
}

func (s *Service) compoundID(ctx context.Context, field, code string) (int64, error) {
	if code == "" {
		return 0, domain.NewValidationError(field, "el compuesto es obligatorio")
	}
	row, err := s.queries.GetTyreCompoundByCode(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.NewValidationError(field, "el compuesto no existe")
	}
	if err != nil {
		return 0, fmt.Errorf("load tyre compound %q: %w", code, err)
	}
	return row.ID, nil
}
