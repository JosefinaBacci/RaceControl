package users

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/racecontrol/backend/internal/auth"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

const maxEmailLength = 254

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type accountState struct {
	username string
	email    string
	role     domain.Role
	teamID   *int64
}

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func validateUsername(username string) error {
	if !domain.IsValidUsername(username) {
		return domain.NewValidationError("username",
			"el usuario debe tener entre 3 y 32 caracteres: minúsculas, números, punto, guion o guion bajo")
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return nil
	}
	if len(email) > maxEmailLength || !emailPattern.MatchString(email) {
		return domain.NewValidationError("email", "el email no es válido")
	}
	return nil
}

func parseRole(raw string) (domain.Role, error) {
	role := domain.Role(raw)
	if !role.IsValid() {
		return "", domain.NewValidationError("role", "el rol debe ser administrador FIA o administrador de escudería")
	}
	return role, nil
}

func validateState(ctx context.Context, queries dbgen.Querier, state accountState) error {
	if err := validateUsername(state.username); err != nil {
		return err
	}
	if err := validateEmail(state.email); err != nil {
		return err
	}
	return validateTeamAssignment(ctx, queries, state.role, state.teamID)
}

func validateTeamAssignment(ctx context.Context, queries dbgen.Querier, role domain.Role, teamID *int64) error {
	if role == domain.RoleFIAAdmin {
		if teamID != nil {
			return domain.NewValidationError("teamId", "un administrador FIA no pertenece a una escudería")
		}
		return nil
	}
	if teamID == nil {
		return domain.NewValidationError("teamId", "un administrador de escudería necesita una escudería")
	}
	exists, err := queries.TeamExists(ctx, *teamID)
	if err != nil {
		return fmt.Errorf("check team %d: %w", *teamID, err)
	}
	if !exists {
		return domain.NewValidationError("teamId", "la escudería no existe")
	}
	return nil
}

func validatePassword(password string, state accountState) error {
	return auth.ValidatePasswordPolicy(password, state.username, state.email)
}
