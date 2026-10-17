package calendar

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

const (
	periodUpcoming = "upcoming"
	periodPast     = "past"

	roundNumberConstraint = "rounds_number_unique_per_championship"
)

// Transactor runs the multi-row writes (event plus its child, the cancellation of
// a race and its round) in a single transaction.
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

// Filter narrows the calendar: a category code, a kind and a period are all
// optional, and an empty one does not filter.
type Filter struct {
	Category string
	Kind     string
	Period   string
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Event, error) {
	params, err := listParams(filter)
	if err != nil {
		return nil, err
	}

	rows, err := s.queries.ListCalendarEvents(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list calendar events: %w", err)
	}

	events := make([]Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, eventFromRow(row))
	}
	return events, nil
}

func listParams(filter Filter) (dbgen.ListCalendarEventsParams, error) {
	var params dbgen.ListCalendarEventsParams
	if category := strings.TrimSpace(filter.Category); category != "" {
		params.CategoryCode = db.OptionalText(category)
	}
	if filter.Kind != "" {
		kind, err := parseKind(filter.Kind)
		if err != nil {
			return dbgen.ListCalendarEventsParams{}, err
		}
		params.Kind = db.OptionalText(string(kind))
	}
	switch filter.Period {
	case "":
	case periodUpcoming, periodPast:
		params.Period = db.OptionalText(filter.Period)
	default:
		return dbgen.ListCalendarEventsParams{},
			domain.NewValidationError("period", "el período debe ser upcoming o past")
	}
	return params, nil
}

func (s *Service) Get(ctx context.Context, id int64) (Event, error) {
	row, err := s.queries.GetCalendarEvent(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Event{}, domain.ErrNotFound
	}
	if err != nil {
		return Event{}, fmt.Errorf("get calendar event %d: %w", id, err)
	}
	return eventFromRow(row), nil
}

func (s *Service) Circuits(ctx context.Context) ([]Circuit, error) {
	rows, err := s.queries.ListCircuits(ctx)
	if err != nil {
		return nil, fmt.Errorf("list circuits: %w", err)
	}

	circuits := make([]Circuit, 0, len(rows))
	for _, row := range rows {
		circuits = append(circuits, Circuit{
			ID:          row.ID,
			Name:        row.Name,
			Locality:    row.Locality,
			CountryCode: row.CountryCode,
		})
	}
	return circuits, nil
}

func (s *Service) Seasons(ctx context.Context) ([]Season, error) {
	rows, err := s.queries.ListSeasons(ctx)
	if err != nil {
		return nil, fmt.Errorf("list seasons: %w", err)
	}

	seasons := make([]Season, 0, len(rows))
	for _, row := range rows {
		seasons = append(seasons, Season{ID: row.ID, Year: row.Year})
	}
	return seasons, nil
}

// NewEvent is a create request. The attributes only some kinds use are optional
// here and required by validation for the kind that needs them.
type NewEvent struct {
	Kind         string
	StartsAt     time.Time
	EndsAt       time.Time
	SeasonID     int64
	CategoryCode string
	CircuitID    *int64
	Number       *int32
	GPName       *string
	ControlType  *string
}

func (s *Service) Create(ctx context.Context, input NewEvent) (Event, error) {
	if !input.StartsAt.IsZero() && !input.StartsAt.After(time.Now()) {
		return Event{}, domain.NewValidationError("startsAt", "el inicio debe ser posterior a la fecha y hora actuales")
	}
	kind, err := parseKind(input.Kind)
	if err != nil {
		return Event{}, err
	}
	state, err := s.newState(ctx, input)
	if err != nil {
		return Event{}, err
	}
	if err := validateState(ctx, s.queries, kind, state); err != nil {
		return Event{}, err
	}

	var id int64
	err = s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		eventID, err := insertEvent(ctx, queries, kind, state)
		if err != nil {
			return err
		}
		id = eventID
		return insertChild(ctx, queries, kind, eventID, state)
	})
	if err != nil {
		return Event{}, translateWriteError(err)
	}
	return s.Get(ctx, id)
}

func (s *Service) newState(ctx context.Context, input NewEvent) (eventState, error) {
	categoryID, err := s.categoryID(ctx, input.CategoryCode)
	if err != nil {
		return eventState{}, err
	}
	return eventState{
		startsAt:    input.StartsAt,
		endsAt:      input.EndsAt,
		seasonID:    input.SeasonID,
		categoryID:  categoryID,
		circuitID:   input.CircuitID,
		number:      input.Number,
		gpName:      trimmed(input.GPName),
		controlType: trimmed(input.ControlType),
	}, nil
}

// EventChanges carries the fields an update may touch. A nil field is left as it
// was, and no field can be cleared.
type EventChanges struct {
	StartsAt     *time.Time
	EndsAt       *time.Time
	SeasonID     *int64
	CategoryCode *string
	CircuitID    *int64
	Number       *int32
	GPName       *string
	ControlType  *string
}

func (s *Service) Update(ctx context.Context, id int64, changes EventChanges) (Event, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return Event{}, err
	}
	if changes.StartsAt != nil && !changes.StartsAt.Equal(current.StartsAt) && !changes.StartsAt.After(time.Now()) {
		return Event{}, domain.NewValidationError("startsAt", "el nuevo inicio debe ser posterior a la fecha y hora actuales")
	}
	next, err := s.applyChanges(ctx, current, changes)
	if err != nil {
		return Event{}, err
	}
	if err := validateState(ctx, s.queries, current.Kind, next); err != nil {
		return Event{}, err
	}

	err = s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		if err := queries.UpdateEventTimes(ctx, dbgen.UpdateEventTimesParams{
			ID:       id,
			StartsAt: db.Timestamptz(next.startsAt),
			EndsAt:   db.Timestamptz(next.endsAt),
		}); err != nil {
			return fmt.Errorf("update event %d: %w", id, err)
		}
		return updateChild(ctx, queries, current, next)
	})
	if err != nil {
		return Event{}, translateWriteError(err)
	}
	return s.Get(ctx, id)
}

func (s *Service) applyChanges(ctx context.Context, current Event, changes EventChanges) (eventState, error) {
	next := eventState{
		startsAt:    current.StartsAt,
		endsAt:      current.EndsAt,
		seasonID:    current.SeasonID,
		categoryID:  current.CategoryID,
		circuitID:   current.CircuitID,
		number:      current.RoundNumber,
		gpName:      textOf(current.GPName),
		controlType: textOf(current.ControlType),
	}
	if changes.StartsAt != nil {
		next.startsAt = *changes.StartsAt
	}
	if changes.EndsAt != nil {
		next.endsAt = *changes.EndsAt
	}
	if changes.SeasonID != nil {
		next.seasonID = *changes.SeasonID
	}
	if changes.CategoryCode != nil {
		categoryID, err := s.categoryID(ctx, *changes.CategoryCode)
		if err != nil {
			return eventState{}, err
		}
		next.categoryID = categoryID
	}
	if changes.CircuitID != nil {
		next.circuitID = changes.CircuitID
	}
	if changes.Number != nil {
		next.number = changes.Number
	}
	if changes.GPName != nil {
		next.gpName = strings.TrimSpace(*changes.GPName)
	}
	if changes.ControlType != nil {
		next.controlType = strings.TrimSpace(*changes.ControlType)
	}
	return next, nil
}

// Cancel deactivates the event and, when it is a race, the round it runs. Only an
// event that has not started can be cancelled: once it starts it is part of the
// season's history. The schema guards also refuse a race with results (RC001),
// which the time check already covers but the database enforces on its own.
func (s *Service) Cancel(ctx context.Context, id int64) (Event, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return Event{}, err
	}
	if !current.IsActive {
		return Event{}, domain.ErrNotFound
	}
	if current.HasStarted(time.Now()) {
		return Event{}, domain.NewConflictError("", "el evento ya empezó: no se puede cancelar")
	}

	err = s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		if err := queries.DeactivateEvent(ctx, id); err != nil {
			return fmt.Errorf("deactivate event %d: %w", id, err)
		}
		if current.Kind != KindRace || current.RoundID == nil {
			return nil
		}
		if err := queries.DeactivateRound(ctx, *current.RoundID); err != nil {
			return fmt.Errorf("cancel round of event %d: %w", id, err)
		}
		return nil
	})
	if err != nil {
		return Event{}, translateWriteError(err)
	}
	return s.Get(ctx, id)
}

func (s *Service) categoryID(ctx context.Context, code string) (int64, error) {
	id, err := s.queries.CategoryIDByCode(ctx, strings.ToLower(strings.TrimSpace(code)))
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.NewValidationError("categoryCode", "la categoría no existe")
	}
	if err != nil {
		return 0, fmt.Errorf("resolve category %q: %w", code, err)
	}
	return id, nil
}

func insertEvent(ctx context.Context, queries dbgen.Querier, kind Kind, state eventState) (int64, error) {
	id, err := queries.CreateEvent(ctx, dbgen.CreateEventParams{
		Kind:     string(kind),
		StartsAt: db.Timestamptz(state.startsAt),
		EndsAt:   db.Timestamptz(state.endsAt),
	})
	if err != nil {
		return 0, fmt.Errorf("create %s event: %w", kind, err)
	}
	return id, nil
}

func insertChild(ctx context.Context, queries dbgen.Querier, kind Kind, eventID int64, state eventState) error {
	switch kind {
	case KindRace:
		roundID, err := queries.CreateRound(ctx, dbgen.CreateRoundParams{
			SeasonID:   state.seasonID,
			CategoryID: state.categoryID,
			Number:     numberValue(state.number),
			CircuitID:  int64Value(state.circuitID),
			GpName:     state.gpName,
		})
		if err != nil {
			return fmt.Errorf("create round of event %d: %w", eventID, err)
		}
		if err := queries.CreateRace(ctx, dbgen.CreateRaceParams{EventID: eventID, RoundID: roundID}); err != nil {
			return fmt.Errorf("create race of event %d: %w", eventID, err)
		}
		return nil

	case KindTyreTest:
		err := queries.CreateTyreTest(ctx, dbgen.CreateTyreTestParams{
			EventID:    eventID,
			SeasonID:   state.seasonID,
			CategoryID: state.categoryID,
			CircuitID:  int64Value(state.circuitID),
		})
		if err != nil {
			return fmt.Errorf("create tyre test of event %d: %w", eventID, err)
		}
		return nil

	case KindTechnicalControl:
		err := queries.CreateTechnicalControl(ctx, dbgen.CreateTechnicalControlParams{
			EventID:     eventID,
			ControlType: state.controlType,
			SeasonID:    state.seasonID,
			CategoryID:  state.categoryID,
		})
		if err != nil {
			return fmt.Errorf("create technical control of event %d: %w", eventID, err)
		}
		return nil

	default:
		return fmt.Errorf("create child of event %d: unknown kind %q", eventID, kind)
	}
}

func updateChild(ctx context.Context, queries dbgen.Querier, current Event, next eventState) error {
	switch current.Kind {
	case KindRace:
		if current.RoundID == nil {
			return fmt.Errorf("race event %d has no round", current.ID)
		}
		err := queries.UpdateRound(ctx, dbgen.UpdateRoundParams{
			ID:         *current.RoundID,
			SeasonID:   next.seasonID,
			CategoryID: next.categoryID,
			Number:     numberValue(next.number),
			CircuitID:  int64Value(next.circuitID),
			GpName:     next.gpName,
		})
		if err != nil {
			return fmt.Errorf("update round of event %d: %w", current.ID, err)
		}
		return nil

	case KindTyreTest:
		err := queries.UpdateTyreTest(ctx, dbgen.UpdateTyreTestParams{
			EventID:    current.ID,
			SeasonID:   next.seasonID,
			CategoryID: next.categoryID,
			CircuitID:  int64Value(next.circuitID),
		})
		if err != nil {
			return fmt.Errorf("update tyre test of event %d: %w", current.ID, err)
		}
		return nil

	case KindTechnicalControl:
		err := queries.UpdateTechnicalControl(ctx, dbgen.UpdateTechnicalControlParams{
			EventID:     current.ID,
			SeasonID:    next.seasonID,
			CategoryID:  next.categoryID,
			ControlType: next.controlType,
		})
		if err != nil {
			return fmt.Errorf("update technical control of event %d: %w", current.ID, err)
		}
		return nil

	default:
		return fmt.Errorf("update child of event %d: unknown kind %q", current.ID, current.Kind)
	}
}

// translateWriteError turns what the schema refuses into the domain error the
// transport maps: the round number is taken, or a finished race is frozen by a
// guard (RC001). Anything else is a genuine failure and is left to surface as 500.
func translateWriteError(err error) error {
	if constraint, isViolation := db.UniqueViolation(err); isViolation {
		if constraint == roundNumberConstraint {
			return domain.NewConflictError("number", "ya hay una ronda activa con ese número en el campeonato")
		}
		return domain.NewConflictError("", "el evento ya existe")
	}
	if code, isGuard := db.GuardViolation(err); isGuard {
		switch code {
		case "RC001":
			return domain.NewConflictError("", "la carrera ya tiene resultados: no se puede modificar ni cancelar")
		default:
			return fmt.Errorf("schema guard %s: %w", code, err)
		}
	}
	return err
}

func trimmed(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func numberValue(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
