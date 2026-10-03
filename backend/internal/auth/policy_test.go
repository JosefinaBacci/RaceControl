package auth

import (
	"errors"
	"strings"
	"testing"

	"github.com/racecontrol/backend/internal/domain"
)

func TestValidatePasswordPolicy(t *testing.T) {
	cases := []struct {
		name     string
		password string
		valid    bool
	}{
		{"long passphrase", "tres tristes tigres", true},
		{"exactly twelve", "abcdefghijkl", true},
		{"too short", "abcdefghijk", false},
		{"too long", strings.Repeat("a", MaxPasswordLength+1), false},
		{"contains username", "xx-Juan.Perez-2026", false},
		{"contains email local part", "juanp-secreto-2026", false},
		{"multibyte counts runes", "ñandú-ñandú-ñ", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePasswordPolicy(tc.password, "juan.perez", "juanp@fia.com")
			var validationErr *domain.ValidationError
			if tc.valid && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.valid && !errors.As(err, &validationErr) {
				t.Fatalf("error = %v, want a validation error", err)
			}
		})
	}
}
