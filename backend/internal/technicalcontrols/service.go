// Package technicalcontrols records and lists whether each team's car passed a
// technical control (US21, US22). The control itself is a calendar event; this
// package owns its official results.
package technicalcontrols

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/results"
)

var chainConstraints = results.Constraints{
	OneOriginal:    "technical_control_results_one_original",
	SupersedesOnce: "technical_control_results_supersedes_once",
}

type Transactor interface {
	InTx(ctx context.Context, work func(queries dbgen.Querier) error) error
}

type Service struct {
	queries    dbgen.Querier
	transactor Transactor
	now        func() time.Time
}

func NewService(queries dbgen.Querier, transactor Transactor) *Service {
	return &Service{queries: queries, transactor: transactor, now: time.Now}
}

type Team struct {
	ID   int64
	Name string
}

// Result grades one team in one control.
type Result struct {
	results.Entry
	Team Team
}

// Filter narrows the public history. A zero Category or TeamID does not filter.
type Filter struct {
	History  domain.HistoryFilter
	Category string
	TeamID   int64
}

// NewResult is the first result of a team in a control.
type NewResult struct {
	EventID    int64
	TeamID     int64
	Input      results.Input
	RecordedBy int64
}

// Correction replaces the result in force with a new one and says why.
type Correction struct {
	ResultID   int64
	Input      results.Input
	Note       string
	RecordedBy int64
}

func (s *Service) History(ctx context.Context, filter Filter) ([]Result, error) {
	rows, err := s.queries.ListTechnicalControlHistory(ctx, historyParams(filter))
	if err != nil {
		return nil, fmt.Errorf("list technical control history: %w", err)
	}
	history := make([]Result, 0, len(rows))
	for _, row := range rows {
		history = append(history, resultFromRow(dbgen.TechnicalControlRevision(row)))
	}
	return history, nil
}

// Types lists the control types that have a result in force, which are the only
// ones a history filter can match.
func (s *Service) Types(ctx context.Context) ([]results.Type, error) {
	controlTypes, err := s.queries.ListTechnicalControlTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list technical control types: %w", err)
	}
	types := make([]results.Type, 0, len(controlTypes))
	for _, controlType := range controlTypes {
		types = append(types, typeOf(controlType))
	}
	return types, nil
}

func (s *Service) Record(ctx context.Context, input NewResult) (Result, error) {
	var id int64
	err := s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		control, err := lockControl(ctx, queries, input.EventID)
		if err != nil {
			return err
		}
		if err := checkGradedTeam(ctx, queries, input.TeamID, control.CategoryID); err != nil {
			return err
		}
		outcome, err := results.ParseInput(input.Input, windowOf(control), s.now())
		if err != nil {
			return err
		}
		id, err = queries.InsertTechnicalControlResult(ctx, insertParams(control, input.TeamID, outcome, input.RecordedBy))
		return err
	})
	if err != nil {
		return Result{}, translateRecordError(err, "record technical control result")
	}
	return s.revision(ctx, id)
}

func (s *Service) Correct(ctx context.Context, correction Correction) (Result, error) {
	var id int64
	err := s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		superseded, err := supersededResult(ctx, queries, correction.ResultID)
		if err != nil {
			return err
		}
		control, err := lockControl(ctx, queries, superseded.EventID)
		if err != nil {
			return err
		}
		outcome, err := results.ParseInput(correction.Input, windowOf(control), s.now())
		if err != nil {
			return err
		}
		note, err := results.NormalizeCorrectionNote(correction.Note)
		if err != nil {
			return err
		}
		params := insertParams(control, superseded.TeamID, outcome, correction.RecordedBy)
		params.SupersedesID = db.Int8(superseded.ID)
		params.CorrectionNote = db.OptionalText(note)
		id, err = queries.InsertTechnicalControlResult(ctx, params)
		return err
	})
	if err != nil {
		return Result{}, translateRecordError(err, "correct technical control result")
	}
	return s.revision(ctx, id)
}

func (s *Service) revision(ctx context.Context, id int64) (Result, error) {
	row, err := s.queries.GetTechnicalControlRevision(ctx, id)
	if err != nil {
		return Result{}, fmt.Errorf("read recorded result %d: %w", id, err)
	}
	return resultFromRow(row), nil
}

func supersededResult(ctx context.Context, queries dbgen.Querier, id int64) (dbgen.GetTechnicalControlResultSubjectRow, error) {
	result, err := queries.GetTechnicalControlResultSubject(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, domain.ErrNotFound
	}
	if err != nil {
		return result, fmt.Errorf("get result %d: %w", id, err)
	}
	return result, nil
}

func lockControl(ctx context.Context, queries dbgen.Querier, eventID int64) (dbgen.LockTechnicalControlRow, error) {
	control, err := queries.LockTechnicalControl(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return control, domain.ErrNotFound
	}
	if err != nil {
		return control, fmt.Errorf("lock technical control %d: %w", eventID, err)
	}
	return control, nil
}

// checkGradedTeam accepts only an active team of the control's championship. The
// schema refuses another championship too, but this answers on the field.
func checkGradedTeam(ctx context.Context, queries dbgen.Querier, teamID, categoryID int64) error {
	const field = "teamId"
	if teamID <= 0 {
		return domain.NewValidationError(field, "elegí la escudería")
	}
	team, err := queries.GetGradedTeam(ctx, teamID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.NewValidationError(field, "la escudería no existe")
	}
	if err != nil {
		return fmt.Errorf("get team %d: %w", teamID, err)
	}
	if team.CategoryID != categoryID {
		return domain.NewValidationError(field, "la escudería no corre en el campeonato del control")
	}
	if !team.IsActive {
		return domain.NewValidationError(field, "la escudería está dada de baja")
	}
	return nil
}

func translateRecordError(err error, action string) error {
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) || errors.Is(err, domain.ErrNotFound) {
		return err
	}
	translated := results.TranslateWriteError(err, chainConstraints)
	if errors.Is(translated, domain.ErrConflict) {
		return translated
	}
	return fmt.Errorf("%s: %w", action, err)
}

func windowOf(control dbgen.LockTechnicalControlRow) results.EventWindow {
	return results.EventWindow{StartsAt: control.StartsAt.Time, EndsAt: control.EndsAt.Time, IsActive: control.IsActive}
}

func insertParams(control dbgen.LockTechnicalControlRow, teamID int64, outcome results.Outcome, recordedBy int64) dbgen.InsertTechnicalControlResultParams {
	return dbgen.InsertTechnicalControlResultParams{
		EventID:     control.EventID,
		CategoryID:  control.CategoryID,
		TeamID:      teamID,
		ConductedOn: db.Date(outcome.ConductedOn),
		Passed:      outcome.Passed,
		Notes:       outcome.Notes,
		RecordedBy:  recordedBy,
	}
}

func historyParams(filter Filter) dbgen.ListTechnicalControlHistoryParams {
	params := dbgen.ListTechnicalControlHistoryParams{
		FromDay:  db.OptionalDate(filter.History.Range.From),
		UntilDay: db.OptionalDate(filter.History.Range.Until),
	}
	if filter.History.Type != "" {
		params.ControlType = db.OptionalText(filter.History.Type)
	}
	if filter.Category != "" {
		params.CategoryCode = db.OptionalText(filter.Category)
	}
	if filter.TeamID != 0 {
		params.TeamID = db.Int8(filter.TeamID)
	}
	return params
}

// A control type is free text typed in the calendar, so it is its own code.
func typeOf(controlType string) results.Type {
	return results.Type{Code: controlType, Name: controlType}
}

func resultFromRow(row dbgen.TechnicalControlRevision) Result {
	return Result{
		Entry: results.Entry{
			ID:             row.ID,
			EventID:        row.EventID,
			ConductedOn:    row.ConductedOn.Time,
			Type:           typeOf(row.ControlType),
			Passed:         row.Passed,
			Notes:          row.Notes,
			SupersedesID:   db.Int64Pointer(row.SupersedesID),
			CorrectionNote: db.TextPointer(row.CorrectionNote),
			SeasonYear:     row.SeasonYear,
			CategoryCode:   row.CategoryCode,
			RecordedBy:     row.RecordedBy,
			RecordedAt:     row.RecordedAt.Time,
		},
		Team: Team{ID: row.TeamID, Name: row.TeamName},
	}
}
