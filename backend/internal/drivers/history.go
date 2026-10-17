package drivers

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/domain"
	"time"
)

type Assignment struct {
	ID               int64
	DriverID         int64
	FirstName        string
	LastName         string
	TeamID           *int64
	TeamName         string
	PreviousTeamName string
	IsStarter        bool
	EffectiveFrom    time.Time
	EffectiveUntil   *time.Time
}

func (s *Service) History(ctx context.Context, reader domain.Identity, teamID *int64) ([]Assignment, error) {
	scope, err := assignmentScope(reader, teamID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListAssignmentHistory(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("list assignment history: %w", err)
	}
	result := make([]Assignment, 0, len(rows))
	for _, row := range rows {
		var until *time.Time
		if row.EffectiveUntil.Valid {
			value := row.EffectiveUntil.Time
			until = &value
		}
		result = append(result, Assignment{ID: row.ID, DriverID: row.DriverID, FirstName: row.FirstName,
			LastName: row.LastName, TeamID: db.Int64Pointer(row.TeamID), TeamName: row.TeamName,
			PreviousTeamName: row.PreviousTeamName, IsStarter: row.IsStarter,
			EffectiveFrom: row.EffectiveFrom.Time, EffectiveUntil: until})
	}
	return result, nil
}

func assignmentScope(reader domain.Identity, requested *int64) (pgtype.Int8, error) {
	if requested != nil && *requested <= 0 {
		return pgtype.Int8{}, domain.NewValidationError("teamId", "la escudería no es válida")
	}
	if reader.Role == domain.RoleFIAAdmin {
		if requested == nil {
			return pgtype.Int8{}, nil
		}
		return db.Int8(*requested), nil
	}
	if reader.Role != domain.RoleTeamAdmin || reader.TeamID == nil {
		return pgtype.Int8{}, domain.ErrForbidden
	}
	if requested != nil && *requested != *reader.TeamID {
		return pgtype.Int8{}, domain.ErrNotFound
	}
	return db.Int8(*reader.TeamID), nil
}
