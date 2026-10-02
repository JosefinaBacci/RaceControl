package domain

import (
	"regexp"
	"strings"
)

type Role string

const (
	RoleFIAAdmin  Role = "fia_admin"
	RoleTeamAdmin Role = "team_admin"
)

func (r Role) IsValid() bool {
	return r == RoleFIAAdmin || r == RoleTeamAdmin
}

type Identity struct {
	UserID    int64
	Username  string
	Role      Role
	TeamID    *int64
	SessionID int64
}

func (i Identity) HasRole(roles ...Role) bool {
	for _, role := range roles {
		if i.Role == role {
			return true
		}
	}
	return false
}

var usernamePattern = regexp.MustCompile(`^[a-z0-9._-]{3,32}$`)

func NormalizeUsername(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func IsValidUsername(normalized string) bool {
	return usernamePattern.MatchString(normalized)
}
