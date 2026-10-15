package dbtest

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/racecontrol/backend/internal/db"
)

// TestGuardViolationReadsTheGuardCode asserts that a service can recognise the
// exception a schema guard raises by the SQLSTATE it carries.
func TestGuardViolationReadsTheGuardCode(t *testing.T) {
	err := &pgconn.PgError{Code: "RC001", Message: "event 1 has results: its date is final"}

	code, isGuard := db.GuardViolation(err)
	if !isGuard || code != "RC001" {
		t.Fatalf("GuardViolation(%v) = (%q, %t), want (\"RC001\", true)", err, code, isGuard)
	}
}

// TestGuardViolationIgnoresOtherErrors asserts that a constraint violation and a
// plain error are not mistaken for a guard.
func TestGuardViolationIgnoresOtherErrors(t *testing.T) {
	for _, err := range []error{&pgconn.PgError{Code: "23505"}, errors.New("connection reset")} {
		if code, isGuard := db.GuardViolation(err); isGuard {
			t.Errorf("GuardViolation(%v) = (%q, true), want it not to be a guard", err, code)
		}
	}
}
