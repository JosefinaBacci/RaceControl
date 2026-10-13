package db

import (
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// Int64Pointer converts a nullable bigint into nil or a pointer to its value.
func Int64Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

// Timestamptz wraps a time as a non-null timestamptz parameter.
func Timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

// UniqueViolation reports whether err is a unique-constraint violation and on which constraint.
func UniqueViolation(err error) (constraint string, isViolation bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return pgErr.ConstraintName, true
	}
	return "", false
}

// guardCodePrefix is the SQLSTATE class the schema guards raise with, so a service
// can tell a deliberate rule apart from an accidental database error.
const guardCodePrefix = "RC"

// GuardViolation reports whether err is a rule a schema guard raised on purpose
// and returns the SQLSTATE that identifies which one.
func GuardViolation(err error) (code string, isGuard bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && strings.HasPrefix(pgErr.Code, guardCodePrefix) {
		return pgErr.Code, true
	}
	return "", false
}
