package resultstest

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/results"
)

func day(year int, month time.Month, dayOfMonth int) time.Time {
	return time.Date(year, month, dayOfMonth, 0, 0, 0, 0, time.UTC)
}

var window = results.EventWindow{
	StartsAt: time.Date(2026, 3, 5, 9, 0, 0, 0, time.UTC),
	EndsAt:   time.Date(2026, 3, 6, 18, 0, 0, 0, time.UTC),
	IsActive: true,
}

var afterTheEvent = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func assertValidationField(t *testing.T, err error, field string) {
	t.Helper()
	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) || validationErr.Field != field {
		t.Fatalf("err = %v, want a validation error on %q", err, field)
	}
}

func TestValidateConductedOnAcceptsTheEventDays(t *testing.T) {
	for _, conductedOn := range []time.Time{day(2026, 3, 5), day(2026, 3, 6)} {
		if err := results.ValidateConductedOn(conductedOn, window, afterTheEvent); err != nil {
			t.Errorf("ValidateConductedOn(%v) = %v, want it accepted", conductedOn, err)
		}
	}
}

func TestValidateConductedOnRejects(t *testing.T) {
	cancelled := window
	cancelled.IsActive = false
	cases := []struct {
		name        string
		conductedOn time.Time
		window      results.EventWindow
		now         time.Time
	}{
		{name: "before the event", conductedOn: day(2026, 3, 4), window: window, now: afterTheEvent},
		{name: "after the event", conductedOn: day(2026, 3, 7), window: window, now: afterTheEvent},
		{name: "in the future", conductedOn: day(2026, 3, 6), window: window, now: time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC)},
		{name: "cancelled event", conductedOn: day(2026, 3, 5), window: cancelled, now: afterTheEvent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertValidationField(t, results.ValidateConductedOn(tc.conductedOn, tc.window, tc.now), "conductedOn")
		})
	}
}

func TestNormalizeNotes(t *testing.T) {
	if notes, err := results.NormalizeNotes(true, "  "); err != nil || notes != "" {
		t.Errorf("passed without notes = (%q, %v), want accepted and empty", notes, err)
	}
	if notes, err := results.NormalizeNotes(false, "  alerón fuera de medida "); err != nil || notes != "alerón fuera de medida" {
		t.Errorf("failed with notes = (%q, %v), want them trimmed", notes, err)
	}
	_, err := results.NormalizeNotes(false, " ")
	assertValidationField(t, err, "notes")
	_, err = results.NormalizeNotes(true, strings.Repeat("ñ", results.MaxNotesLength+1))
	assertValidationField(t, err, "notes")
}

func TestNormalizeCorrectionNote(t *testing.T) {
	if note, err := results.NormalizeCorrectionNote(" equipo mal elegido "); err != nil || note != "equipo mal elegido" {
		t.Errorf("NormalizeCorrectionNote = (%q, %v), want it trimmed", note, err)
	}
	_, err := results.NormalizeCorrectionNote("")
	assertValidationField(t, err, "correctionNote")
	_, err = results.NormalizeCorrectionNote(strings.Repeat("a", results.MaxNotesLength+1))
	assertValidationField(t, err, "correctionNote")
}

func TestTranslateWriteError(t *testing.T) {
	constraints := results.Constraints{OneOriginal: "one_original", SupersedesOnce: "supersedes_once"}
	for _, constraint := range []string{"one_original", "supersedes_once"} {
		err := results.TranslateWriteError(&pgconn.PgError{Code: "23505", ConstraintName: constraint}, constraints)
		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("violation of %s = %v, want a conflict", constraint, err)
		}
	}
	other := &pgconn.PgError{Code: "23505", ConstraintName: "something_else"}
	if err := results.TranslateWriteError(other, constraints); err != other {
		t.Errorf("unrelated violation = %v, want it untouched", err)
	}
}
