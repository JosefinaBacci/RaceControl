package db

import (
	"errors"
	"fmt"
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

// Int8 wraps an int64 as a non-null bigint parameter.
func Int8(value int64) pgtype.Int8 {
	return pgtype.Int8{Int64: value, Valid: true}
}

// Int4 wraps an int32 as a non-null integer parameter.
func Int4(value int32) pgtype.Int4 {
	return pgtype.Int4{Int32: value, Valid: true}
}

// TextPointer converts a nullable text into nil or a pointer to its value.
func TextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

// TimePointer converts a nullable timestamptz into nil or a pointer to its value.
func TimePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

// Int32Pointer converts a nullable int4 into nil or a pointer to its value.
func Int32Pointer(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

// Float64Pointer converts a nullable numeric into nil or a pointer to its value.
func Float64Pointer(value pgtype.Numeric) (*float64, error) {
	if !value.Valid {
		return nil, nil
	}
	converted, err := value.Float64Value()
	if err != nil {
		return nil, fmt.Errorf("convert numeric to float64: %w", err)
	}
	return &converted.Float64, nil
}

// OptionalText wraps a Go string as a present text parameter.
func OptionalText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
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
