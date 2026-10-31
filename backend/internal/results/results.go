// Package results holds the rules every official result shares, whatever it
// grades (a team in a technical control, a compound in a tyre test): when it may
// be dated, what its notes must say, and how a correction supersedes it.
//
// A result is an official record, so it is never rewritten. A correction is a new
// row that supersedes the current one; the chain of rows is the audit trail.
package results

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/domain"
)

const MaxNotesLength = 2000

// EventWindow is the part of a calendar event a result is dated against.
type EventWindow struct {
	StartsAt time.Time
	EndsAt   time.Time
	IsActive bool
}

// ValidateConductedOn accepts the day a result was obtained only if the event
// still stands, the day falls inside the event's days and it is not in the future:
// nothing can be graded before it happened.
func ValidateConductedOn(conductedOn time.Time, window EventWindow, now time.Time) error {
	const field = "conductedOn"
	if !window.IsActive {
		return domain.NewValidationError(field, "el evento está cancelado")
	}
	if conductedOn.After(domain.DayOf(now)) {
		return domain.NewValidationError(field, "la fecha todavía no ocurrió")
	}
	if conductedOn.Before(domain.DayOf(window.StartsAt)) || conductedOn.After(domain.DayOf(window.EndsAt)) {
		return domain.NewValidationError(field, "la fecha está fuera de los días del evento")
	}
	return nil
}

// NormalizeNotes trims the notes and accepts them only within the length limit,
// and only if a failed result says why: a bare "not passed" is no record.
func NormalizeNotes(passed bool, notes string) (string, error) {
	trimmed := strings.TrimSpace(notes)
	if !passed && trimmed == "" {
		return "", domain.NewValidationError("notes", "un resultado no aprobado necesita observaciones")
	}
	if utf8.RuneCountInString(trimmed) > MaxNotesLength {
		return "", domain.NewValidationError("notes", "las observaciones no pueden superar los 2000 caracteres")
	}
	return trimmed, nil
}

// NormalizeCorrectionNote trims the reason for a correction, which is mandatory.
func NormalizeCorrectionNote(note string) (string, error) {
	trimmed := strings.TrimSpace(note)
	if trimmed == "" {
		return "", domain.NewValidationError("correctionNote", "la corrección necesita un motivo")
	}
	if utf8.RuneCountInString(trimmed) > MaxNotesLength {
		return "", domain.NewValidationError("correctionNote", "el motivo no puede superar los 2000 caracteres")
	}
	return trimmed, nil
}

// Constraints names the two unique constraints that keep a results table to one
// current row per graded subject: one original row, and each row superseded once.
type Constraints struct {
	OneOriginal    string
	SupersedesOnce string
}

// TranslateWriteError turns a lost race on the results chain into the conflict
// the transport maps; anything else is returned untouched.
func TranslateWriteError(err error, constraints Constraints) error {
	constraint, isViolation := db.UniqueViolation(err)
	if !isViolation {
		return err
	}
	switch constraint {
	case constraints.OneOriginal:
		return domain.NewConflictError("", "ya hay un resultado registrado: corregilo en lugar de cargar otro")
	case constraints.SupersedesOnce:
		return domain.NewConflictError("", "ese resultado ya fue corregido: corregí el vigente")
	default:
		return err
	}
}
