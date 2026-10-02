package auth

import (
	"errors"
	"strings"
	"testing"
)

var fastParams = Argon2Params{MemoryKiB: 1024, Iterations: 1, Parallelism: 1}

func newTestHasher(t *testing.T, params Argon2Params) *PasswordHasher {
	t.Helper()

	hasher, err := NewPasswordHasher(params)
	if err != nil {
		t.Fatalf("NewPasswordHasher: %v", err)
	}
	return hasher
}

func TestHashEncodesParametersInPHCFormat(t *testing.T) {
	hash, err := newTestHasher(t, fastParams).Hash("correct horse battery")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=1024,t=1,p=1$") {
		t.Fatalf("unexpected hash format %q", hash)
	}
}

func TestHashUsesFreshSaltEachTime(t *testing.T) {
	hasher := newTestHasher(t, fastParams)
	first, _ := hasher.Hash("same password here")
	second, _ := hasher.Hash("same password here")
	if first == second {
		t.Fatal("two hashes of the same password are identical; salt is not random")
	}
}

func TestVerify(t *testing.T) {
	hasher := newTestHasher(t, fastParams)
	hash, err := hasher.Hash("correct horse battery")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	cases := []struct {
		name     string
		password string
		matches  bool
	}{
		{"correct password", "correct horse battery", true},
		{"wrong password", "correct horse batterY", false},
		{"empty password", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check, err := hasher.Verify(tc.password, hash)
			if err != nil {
				t.Fatalf("Verify: %v", err)
			}
			if check.Matches != tc.matches {
				t.Fatalf("Matches = %v, want %v", check.Matches, tc.matches)
			}
		})
	}
}

func TestVerifyFlagsRehashWhenConfiguredCostIncreases(t *testing.T) {
	hash, err := newTestHasher(t, fastParams).Hash("correct horse battery")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	stronger := newTestHasher(t, Argon2Params{MemoryKiB: 2048, Iterations: 1, Parallelism: 1})
	check, err := stronger.Verify("correct horse battery", hash)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !check.Matches || !check.NeedsRehash {
		t.Fatalf("check = %+v, want a match that needs rehash", check)
	}
}

func TestVerifyRejectsMalformedHash(t *testing.T) {
	hasher := newTestHasher(t, fastParams)
	for _, malformed := range []string{"", "plain", "$bcrypt$v=19$m=1,t=1,p=1$c2FsdA$a2V5", "$argon2id$v=19$m=x$c2FsdA$a2V5"} {
		if _, err := hasher.Verify("whatever", malformed); !errors.Is(err, ErrMalformedHash) {
			t.Errorf("Verify(%q) error = %v, want ErrMalformedHash", malformed, err)
		}
	}
}
