package calendar

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// eventState is an event reduced to the values a write needs, whatever its kind.
// It is what a create carries and what an update turns the current row into, so
// both paths validate the very same shape.
type eventState struct {
	startsAt    time.Time
	endsAt      time.Time
	seasonID    int64
	categoryID  int64
	circuitID   *int64
	number      *int32
	gpName      string
	controlType string
}

func parseKind(raw string) (Kind, error) {
	kind := Kind(raw)
	if !kind.IsValid() {
		return "", domain.NewValidationError("kind", "el tipo debe ser race, tyre_test o technical_control")
	}
	return kind, nil
}

func validateState(ctx context.Context, queries dbgen.Querier, kind Kind, state eventState) error {
	if err := validateTiming(state); err != nil {
		return err
	}
	if err := validateWithinSeason(ctx, queries, state); err != nil {
		return err
	}
	switch kind {
	case KindRace:
		return validateRace(ctx, queries, state)
	case KindTyreTest:
		return validateTyreTest(ctx, queries, state)
	case KindTechnicalControl:
		return validateControl(state)
	default:
		return domain.NewValidationError("kind", "el tipo de evento no es válido")
	}
}

func validateTiming(state eventState) error {
	if state.startsAt.IsZero() || state.endsAt.IsZero() {
		return domain.NewValidationError("startsAt", "la fecha de inicio y de fin son obligatorias")
	}
	if !state.endsAt.After(state.startsAt) {
		return domain.NewValidationError("endsAt", "el fin tiene que ser posterior al inicio")
	}
	return nil
}

// validateWithinSeason keeps the event's date inside its championship year: a
// round points at a season, so an event outside that year would lie about it.
func validateWithinSeason(ctx context.Context, queries dbgen.Querier, state eventState) error {
	year, err := queries.SeasonYear(ctx, state.seasonID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.NewValidationError("seasonId", "la temporada no existe")
	}
	if err != nil {
		return fmt.Errorf("load season %d: %w", state.seasonID, err)
	}

	start := time.Date(int(year), time.January, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	if state.startsAt.Before(start) || !state.startsAt.Before(end) {
		return domain.NewValidationError("startsAt",
			fmt.Sprintf("la fecha tiene que caer dentro de la temporada %d", year))
	}
	return nil
}

func validateRace(ctx context.Context, queries dbgen.Querier, state eventState) error {
	if state.number == nil || *state.number <= 0 {
		return domain.NewValidationError("number", "la ronda necesita un número mayor que cero")
	}
	if state.gpName == "" {
		return domain.NewValidationError("gpName", "la carrera necesita el nombre del Gran Premio")
	}
	return validateCircuit(ctx, queries, state.circuitID)
}

func validateTyreTest(ctx context.Context, queries dbgen.Querier, state eventState) error {
	return validateCircuit(ctx, queries, state.circuitID)
}

func validateControl(state eventState) error {
	if state.controlType == "" {
		return domain.NewValidationError("controlType", "el control necesita un tipo")
	}
	return nil
}

func validateCircuit(ctx context.Context, queries dbgen.Querier, circuitID *int64) error {
	if circuitID == nil {
		return domain.NewValidationError("circuitId", "el evento necesita un circuito")
	}
	exists, err := queries.CircuitExists(ctx, *circuitID)
	if err != nil {
		return fmt.Errorf("check circuit %d: %w", *circuitID, err)
	}
	if !exists {
		return domain.NewValidationError("circuitId", "el circuito no existe")
	}
	return nil
}
