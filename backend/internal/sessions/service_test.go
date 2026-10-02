package sessions_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/sessions"
	"github.com/racecontrol/backend/internal/testdb"
)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

func newUserID(t *testing.T, queries *dbgen.Queries) int64 {
	t.Helper()

	user, err := queries.CreateUser(context.Background(), dbgen.CreateUserParams{
		Username:     testdb.UniqueUsername(t),
		PasswordHash: "unused-in-session-tests",
		Role:         string(domain.RoleFIAAdmin),
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return user.ID
}

func TestSessionExpiry(t *testing.T) {
	const idleTTL, absoluteTTL = time.Hour, 3 * time.Hour

	cases := []struct {
		name     string
		schedule []time.Duration
		valid    bool
	}{
		{"fresh session", []time.Duration{0}, true},
		{"idle beyond ttl", []time.Duration{idleTTL + time.Minute}, false},
		{"activity keeps it alive", []time.Duration{50 * time.Minute, 100 * time.Minute, 150 * time.Minute}, true},
		{"absolute ttl wins over activity", []time.Duration{50 * time.Minute, 100 * time.Minute, 150 * time.Minute, absoluteTTL + time.Minute}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queries := dbgen.New(testdb.Pool(t))
			clock := &fakeClock{now: time.Now()}
			service := sessions.NewService(queries, sessions.Config{IdleTTL: idleTTL, AbsoluteTTL: absoluteTTL, Clock: clock.Now})
			start := clock.now

			token, err := service.Issue(context.Background(), newUserID(t, queries), sessions.Client{})
			if err != nil {
				t.Fatalf("Issue: %v", err)
			}

			var resolveErr error
			for _, offset := range tc.schedule {
				clock.now = start.Add(offset)
				_, resolveErr = service.Resolve(context.Background(), token)
			}

			if tc.valid && resolveErr != nil {
				t.Fatalf("Resolve: %v", resolveErr)
			}
			if !tc.valid && !errors.Is(resolveErr, domain.ErrUnauthenticated) {
				t.Fatalf("Resolve error = %v, want ErrUnauthenticated", resolveErr)
			}
		})
	}
}

func TestRevokeAllForUserEndsEverySession(t *testing.T) {
	queries := dbgen.New(testdb.Pool(t))
	service := sessions.NewService(queries, sessions.Config{IdleTTL: time.Hour, AbsoluteTTL: 2 * time.Hour})
	userID := newUserID(t, queries)

	var tokens []string
	for range 2 {
		token, err := service.Issue(context.Background(), userID, sessions.Client{})
		if err != nil {
			t.Fatalf("Issue: %v", err)
		}
		tokens = append(tokens, token)
	}

	if err := service.RevokeAllForUser(context.Background(), userID); err != nil {
		t.Fatalf("RevokeAllForUser: %v", err)
	}

	for _, token := range tokens {
		if _, err := service.Resolve(context.Background(), token); !errors.Is(err, domain.ErrUnauthenticated) {
			t.Fatalf("Resolve after revoke = %v, want ErrUnauthenticated", err)
		}
	}
}
