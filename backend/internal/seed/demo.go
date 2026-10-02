package seed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/auth"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

type demoAccount struct {
	username string
	role     domain.Role
	teamID   int64
}

var demoAccounts = []demoAccount{
	{username: "fia.admin", role: domain.RoleFIAAdmin},
	{username: "ferrari.admin", role: domain.RoleTeamAdmin, teamID: 1},
	{username: "redbull.admin", role: domain.RoleTeamAdmin, teamID: 2},
	{username: "mercedes.admin", role: domain.RoleTeamAdmin, teamID: 3},
}

const demoEmailDomain = "racecontrol.test"

func (r *Runner) DemoAccounts(ctx context.Context, hasher *auth.PasswordHasher, password string) error {
	queries := dbgen.New(r.pool)
	for _, account := range demoAccounts {
		if err := createDemoAccount(ctx, queries, hasher, account, password); err != nil {
			return err
		}
	}
	return nil
}

func createDemoAccount(ctx context.Context, queries *dbgen.Queries, hasher *auth.PasswordHasher, account demoAccount, password string) error {
	email := account.username + "@" + demoEmailDomain
	if err := auth.ValidatePasswordPolicy(password, account.username, email); err != nil {
		return fmt.Errorf("SEED_DEMO_PASSWORD rejected for %s: %w", account.username, err)
	}

	hash, err := hasher.Hash(password)
	if err != nil {
		return err
	}

	err = queries.CreateUserIfAbsent(ctx, dbgen.CreateUserIfAbsentParams{
		Username:     account.username,
		Email:        pgtype.Text{String: email, Valid: true},
		PasswordHash: hash,
		Role:         string(account.role),
		TeamID:       pgtype.Int8{Int64: account.teamID, Valid: account.teamID != 0},
	})
	if err != nil {
		return fmt.Errorf("create demo account %s: %w", account.username, err)
	}
	return nil
}
