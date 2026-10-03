package db

import (
	"errors"
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
