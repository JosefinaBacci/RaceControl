package auth

import (
	"strings"
	"unicode/utf8"

	"github.com/racecontrol/backend/internal/domain"
)

const (
	MinPasswordLength            = 12
	MaxPasswordLength            = 128
	minIdentifierLengthToCompare = 3
)

func ValidatePasswordPolicy(password, username, email string) error {
	length := utf8.RuneCountInString(password)
	if length < MinPasswordLength {
		return domain.NewValidationError("password", "la contraseña debe tener al menos 12 caracteres")
	}
	if length > MaxPasswordLength {
		return domain.NewValidationError("password", "la contraseña no puede superar los 128 caracteres")
	}
	if resemblesAny(password, identifiersOf(username, email)) {
		return domain.NewValidationError("password", "la contraseña no puede parecerse al usuario o al email")
	}
	return nil
}

func identifiersOf(username, email string) []string {
	identifiers := []string{username, email}
	if localPart, _, found := strings.Cut(email, "@"); found {
		identifiers = append(identifiers, localPart)
	}
	return identifiers
}

func resemblesAny(password string, identifiers []string) bool {
	normalizedPassword := strings.ToLower(password)
	for _, identifier := range identifiers {
		normalizedIdentifier := strings.ToLower(strings.TrimSpace(identifier))
		if len(normalizedIdentifier) < minIdentifierLengthToCompare {
			continue
		}
		if strings.Contains(normalizedPassword, normalizedIdentifier) {
			return true
		}
	}
	return false
}
