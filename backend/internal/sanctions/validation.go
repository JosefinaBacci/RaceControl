package sanctions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// The limits match the CHECKs of sanction_versions (000035).
const (
	minReasonLength = 3
	maxReasonLength = 2000
	maxNoteLength   = 1000
)

// Terms is what a version imposes and why. The FIA submits them on issue and on
// every amendment.
type Terms struct {
	PenaltyType  string
	PenaltyValue *int32
	Reason       string
	EventID      *int64
}

// resolvedTerms are the terms checked against the catalogue, ready to store.
type resolvedTerms struct {
	penaltyTypeID int64
	penaltyValue  *int32
	reason        string
	eventID       *int64
}

func (t resolvedTerms) equals(other resolvedTerms) bool {
	return t.penaltyTypeID == other.penaltyTypeID &&
		equalPointers(t.penaltyValue, other.penaltyValue) &&
		t.reason == other.reason &&
		equalPointers(t.eventID, other.eventID)
}

func equalPointers[T comparable](left, right *T) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

// resolveTerms validates everything the terms say on their own; whether the event
// belongs to the championship needs the sanction and is checked later.
func resolveTerms(ctx context.Context, queries dbgen.Querier, terms Terms) (resolvedTerms, error) {
	reason, err := normalizeReason(terms.Reason)
	if err != nil {
		return resolvedTerms{}, err
	}
	penaltyType, err := penaltyTypeByCode(ctx, queries, terms.PenaltyType)
	if err != nil {
		return resolvedTerms{}, err
	}
	if err := validatePenaltyValue(penaltyType, terms.PenaltyValue); err != nil {
		return resolvedTerms{}, err
	}
	return resolvedTerms{
		penaltyTypeID: penaltyType.ID,
		penaltyValue:  terms.PenaltyValue,
		reason:        reason,
		eventID:       terms.EventID,
	}, nil
}

func normalizeReason(raw string) (string, error) {
	reason := strings.TrimSpace(raw)
	length := utf8.RuneCountInString(reason)
	if length < minReasonLength {
		return "", domain.NewValidationError("reason", "el motivo es obligatorio (al menos 3 caracteres)")
	}
	if length > maxReasonLength {
		return "", domain.NewValidationError("reason", fmt.Sprintf("el motivo admite hasta %d caracteres", maxReasonLength))
	}
	return reason, nil
}

// normalizeNote checks the explanation an amendment or a revocation must carry.
func normalizeNote(raw string) (string, error) {
	note := strings.TrimSpace(raw)
	if note == "" {
		return "", domain.NewValidationError("note", "explicá el motivo del cambio")
	}
	if utf8.RuneCountInString(note) > maxNoteLength {
		return "", domain.NewValidationError("note", fmt.Sprintf("la nota admite hasta %d caracteres", maxNoteLength))
	}
	return note, nil
}

func penaltyTypeByCode(ctx context.Context, queries dbgen.Querier, code string) (dbgen.GetPenaltyTypeByCodeRow, error) {
	row, err := queries.GetPenaltyTypeByCode(ctx, strings.TrimSpace(code))
	if errors.Is(err, pgx.ErrNoRows) {
		return dbgen.GetPenaltyTypeByCodeRow{}, domain.NewValidationError("penaltyType", "la penalización no existe")
	}
	if err != nil {
		return dbgen.GetPenaltyTypeByCodeRow{}, fmt.Errorf("load penalty type %q: %w", code, err)
	}
	return row, nil
}

// validatePenaltyValue mirrors the schema guard (000035, RC010) so the FIA gets a
// message that names the field.
func validatePenaltyValue(penaltyType dbgen.GetPenaltyTypeByCodeRow, value *int32) error {
	const field = "penaltyValue"
	if !penaltyType.Unit.Valid {
		if value != nil {
			return domain.NewValidationError(field, "esta penalización no lleva valor")
		}
		return nil
	}
	maxValue := penaltyType.MaxValue.Int32
	if value == nil || *value < 1 || *value > maxValue {
		return domain.NewValidationError(field, fmt.Sprintf("el valor debe ser un entero entre 1 y %d", maxValue))
	}
	return nil
}

// validateEvent accepts only an event that still stands in the accountable
// team's championship.
func validateEvent(ctx context.Context, queries dbgen.Querier, eventID *int64, teamID int64) error {
	if eventID == nil {
		return nil
	}
	inChampionship, err := queries.ActiveEventInTeamCategory(ctx, dbgen.ActiveEventInTeamCategoryParams{EventID: *eventID, TeamID: teamID})
	if err != nil {
		return fmt.Errorf("check event %d of team %d: %w", *eventID, teamID, err)
	}
	if !inChampionship {
		return domain.NewValidationError("eventId", "el evento no existe, está cancelado o es de otro campeonato")
	}
	return nil
}

// Target is who a new sanction falls on: exactly one of the two.
type Target struct {
	DriverID *int64
	TeamID   *int64
}

func validateTarget(ctx context.Context, queries dbgen.Querier, target Target) error {
	if (target.DriverID == nil) == (target.TeamID == nil) {
		return domain.NewValidationError("target", "la sanción va a un piloto o a una escudería, exactamente uno")
	}
	if target.DriverID != nil {
		return requireActive(ctx, queries.ActiveDriverExists, *target.DriverID, "driverId", "el piloto no existe o está dado de baja")
	}
	return requireActive(ctx, queries.ActiveTeamExists, *target.TeamID, "teamId", "la escudería no existe o está dada de baja")
}

func requireActive(ctx context.Context, exists func(context.Context, int64) (bool, error), id int64, field, message string) error {
	found, err := exists(ctx, id)
	if err != nil {
		return fmt.Errorf("check %s %d: %w", field, id, err)
	}
	if !found {
		return domain.NewValidationError(field, message)
	}
	return nil
}

// Guard codes of the sanction schema (000035, 000036).
const (
	guardPenaltyValue = "RC010"
	guardDriverNoTeam = "RC011"
	guardVersionOrder = "RC012"
	guardNotRecipient = "RC013"
	guardVersionStale = "RC014"
)

// translateWriteError turns what the schema refuses into the domain error the
// transport maps. The service checks each rule first, so a guard firing here
// means another request got in between.
func translateWriteError(err error) error {
	code, isGuard := db.GuardViolation(err)
	if !isGuard {
		return err
	}
	switch code {
	case guardPenaltyValue:
		return domain.NewValidationError("penaltyValue", "el valor no corresponde a la penalización")
	case guardDriverNoTeam:
		return domain.NewValidationError("driverId", "el piloto no tiene escudería hoy: no hay a quién notificar")
	case guardVersionOrder:
		return domain.NewConflictError("id", "la sanción ya fue revocada")
	default:
		return fmt.Errorf("schema guard %s: %w", code, err)
	}
}
