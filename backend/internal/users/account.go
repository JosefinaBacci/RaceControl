package users

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

type Account struct {
	ID            int64
	Username      string
	Email         *string
	Role          domain.Role
	TeamID        *int64
	TeamName      *string
	IsActive      bool
	DeactivatedAt *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type NewAccount struct {
	Username string
	Email    string
	Password string
	Role     string
	TeamID   *int64
}

// A nil field keeps its current value and an empty Email clears it.
type AccountChanges struct {
	Email    *string
	Role     *string
	TeamID   *int64
	Password *string
}

type Filter struct {
	Search string
	Role   string
	Status string
}

func accountFromRow(row dbgen.GetAccountRow) Account {
	return Account{
		ID:            row.ID,
		Username:      row.Username,
		Email:         textPointer(row.Email),
		Role:          domain.Role(row.Role),
		TeamID:        db.Int64Pointer(row.TeamID),
		TeamName:      textPointer(row.TeamName),
		IsActive:      row.IsActive,
		DeactivatedAt: timePointer(row.DeactivatedAt),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func optionalText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func optionalInt8(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}
