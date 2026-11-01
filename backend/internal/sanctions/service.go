// Package sanctions keeps the FIA's record of sanctions (US15), publishes it
// (US25) and follows which teams confirmed each version (US26).
//
// A sanction is never rewritten: issuing, amending and revoking each append a
// version, the latest one stands, and every version is sent to the teams of the
// championship to confirm.
package sanctions

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// Transactor runs a sanction and its version in a single transaction.
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

// Status filters the list by whether the sanction still stands.
type Status string

const (
	StatusAll     Status = "all"
	StatusActive  Status = "active"
	StatusRevoked Status = "revoked"
)

// isRevokedByStatus is the list filter each status asks for; nil matches both.
var isRevokedByStatus = map[Status]pgtype.Bool{
	StatusAll:     {},
	StatusActive:  {Bool: false, Valid: true},
	StatusRevoked: {Bool: true, Valid: true},
}

// Filter narrows the public list. TeamID matches the accountable team, so it
// also finds the sanctions of the team's drivers.
type Filter struct {
	DriverID *int64
	TeamID   *int64
	Category string
	Status   Status
	Range    domain.DateRange
}

// PenaltyTypes lists the catalogue the FIA picks a penalty from.
func (s *Service) PenaltyTypes(ctx context.Context) ([]PenaltyType, error) {
	rows, err := s.queries.ListPenaltyTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list penalty types: %w", err)
	}
	types := make([]PenaltyType, 0, len(rows))
	for _, row := range rows {
		types = append(types, PenaltyType{
			ID:       row.ID,
			Code:     row.Code,
			Name:     row.Name,
			Unit:     db.TextPointer(row.Unit),
			MaxValue: db.Int32Pointer(row.MaxValue),
		})
	}
	return types, nil
}

// List answers the sanctions as they stand, newest issue first.
func (s *Service) List(ctx context.Context, filter Filter) ([]Sanction, error) {
	params, err := listParams(filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListCurrentSanctions(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list sanctions: %w", err)
	}
	sanctions := make([]Sanction, 0, len(rows))
	for _, row := range rows {
		sanctions = append(sanctions, sanctionFromRow(row))
	}
	return sanctions, nil
}

func listParams(filter Filter) (dbgen.ListCurrentSanctionsParams, error) {
	isRevoked, err := isRevokedOf(filter.Status)
	if err != nil {
		return dbgen.ListCurrentSanctionsParams{}, err
	}
	return dbgen.ListCurrentSanctionsParams{
		DriverID:     optionalInt8(filter.DriverID),
		TeamID:       optionalInt8(filter.TeamID),
		CategoryCode: optionalCategory(filter.Category),
		IsRevoked:    isRevoked,
		FromAt:       db.OptionalTimestamptz(filter.Range.From),
		UntilAt:      db.OptionalTimestamptz(filter.Range.Until),
	}, nil
}

func isRevokedOf(status Status) (pgtype.Bool, error) {
	if status == "" {
		status = StatusAll
	}
	isRevoked, known := isRevokedByStatus[status]
	if !known {
		return pgtype.Bool{}, domain.NewValidationError("status", "el estado debe ser active, revoked o all")
	}
	return isRevoked, nil
}

func optionalCategory(raw string) pgtype.Text {
	category := strings.ToLower(strings.TrimSpace(raw))
	return pgtype.Text{String: category, Valid: category != ""}
}

// Get answers a sanction with its whole history. A revoked sanction is still
// answered: the revocation is part of the official record.
func (s *Service) Get(ctx context.Context, id int64) (Detail, error) {
	rows, err := s.queries.ListSanctionVersions(ctx, id)
	if err != nil {
		return Detail{}, fmt.Errorf("list versions of sanction %d: %w", id, err)
	}
	if len(rows) == 0 {
		return Detail{}, domain.ErrNotFound
	}
	versions := make([]Version, 0, len(rows))
	for _, row := range rows {
		versions = append(versions, versionFromRow(row))
	}
	return Detail{Sanction: sanctionFromRow(rows[len(rows)-1]), Versions: versions}, nil
}

func optionalInt8(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return db.Int8(*value)
}

func optionalInt4(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return db.Int4(*value)
}

func optionalNote(note string) pgtype.Text {
	return pgtype.Text{String: note, Valid: note != ""}
}
