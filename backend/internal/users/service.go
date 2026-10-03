package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/auth"
	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

const (
	statusActive   = "active"
	statusInactive = "inactive"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) (auth.PasswordCheck, error)
}

type Transactor interface {
	InTx(ctx context.Context, work func(queries dbgen.Querier) error) error
}

type Service struct {
	queries    dbgen.Querier
	transactor Transactor
	hasher     PasswordHasher
}

func NewService(queries dbgen.Querier, transactor Transactor, hasher PasswordHasher) *Service {
	return &Service{queries: queries, transactor: transactor, hasher: hasher}
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Account, error) {
	params, err := listParams(filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListAccounts(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	accounts := make([]Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, accountFromRow(dbgen.GetAccountRow(row)))
	}
	return accounts, nil
}

func (s *Service) Get(ctx context.Context, id int64) (Account, error) {
	row, err := s.queries.GetAccount(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, domain.ErrNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("get account %d: %w", id, err)
	}
	return accountFromRow(row), nil
}

func (s *Service) Create(ctx context.Context, input NewAccount) (Account, error) {
	state, err := newAccountState(input)
	if err != nil {
		return Account{}, err
	}
	if err := validateState(ctx, s.queries, state); err != nil {
		return Account{}, err
	}
	if err := validatePassword(input.Password, state); err != nil {
		return Account{}, err
	}
	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return Account{}, err
	}

	user, err := s.queries.CreateUser(ctx, dbgen.CreateUserParams{
		Username:     state.username,
		Email:        optionalText(state.email),
		PasswordHash: hash,
		Role:         string(state.role),
		TeamID:       optionalInt8(state.teamID),
	})
	if err != nil {
		return Account{}, conflictOrWrap(err, "create account")
	}
	return s.Get(ctx, user.ID)
}

func (s *Service) Update(ctx context.Context, actor domain.Identity, id int64, changes AccountChanges) (Account, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return Account{}, err
	}
	next, err := changes.applyTo(current)
	if err != nil {
		return Account{}, err
	}
	if actor.UserID == id && next.role != current.Role {
		return Account{}, domain.ErrSelfLockout
	}
	if err := validateState(ctx, s.queries, next); err != nil {
		return Account{}, err
	}
	newHash, err := s.hashChangedPassword(changes.Password, next)
	if err != nil {
		return Account{}, err
	}

	err = s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		return applyUpdate(ctx, queries, id, next, newHash, revokesSessions(current, next, newHash))
	})
	if err != nil {
		return Account{}, err
	}
	return s.Get(ctx, id)
}

func (s *Service) Deactivate(ctx context.Context, actor domain.Identity, id int64) (Account, error) {
	if actor.UserID == id {
		return Account{}, domain.ErrSelfLockout
	}
	if _, err := s.Get(ctx, id); err != nil {
		return Account{}, err
	}
	err := s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		if err := queries.DeactivateUser(ctx, id); err != nil {
			return fmt.Errorf("deactivate account %d: %w", id, err)
		}
		return revokeSessions(ctx, queries, dbgen.RevokeUserSessionsParams{UserID: id})
	})
	if err != nil {
		return Account{}, err
	}
	return s.Get(ctx, id)
}

func (s *Service) Reactivate(ctx context.Context, id int64) (Account, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return Account{}, err
	}
	if err := s.queries.ReactivateUser(ctx, id); err != nil {
		return Account{}, fmt.Errorf("reactivate account %d: %w", id, err)
	}
	return s.Get(ctx, id)
}

// ChangeOwnPassword keeps the session that made the request, so the user is not
// signed out of the device they are using, and ends every other one.
func (s *Service) ChangeOwnPassword(ctx context.Context, actor domain.Identity, current, next string) error {
	user, err := s.queries.GetUserByUsername(ctx, actor.Username)
	if err != nil {
		return fmt.Errorf("load account %d: %w", actor.UserID, err)
	}
	if err := s.verifyCurrentPassword(current, user.PasswordHash); err != nil {
		return err
	}
	if next == current {
		return domain.NewValidationError("password", "la nueva contraseña tiene que ser distinta de la actual")
	}
	newHash, err := s.hashChangedPassword(&next, accountState{username: user.Username, email: user.Email.String})
	if err != nil {
		return err
	}

	return s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		if err := storePasswordHash(ctx, queries, user.ID, newHash); err != nil {
			return err
		}
		return revokeSessions(ctx, queries, dbgen.RevokeUserSessionsParams{
			UserID:        user.ID,
			KeptSessionID: pgtype.Int8{Int64: actor.SessionID, Valid: true},
		})
	})
}

func (s *Service) verifyCurrentPassword(current, storedHash string) error {
	check, err := s.hasher.Verify(current, storedHash)
	if err != nil {
		return fmt.Errorf("verify current password: %w", err)
	}
	if !check.Matches {
		return domain.NewValidationError("currentPassword", "la contraseña actual no es correcta")
	}
	return nil
}

func (s *Service) hashChangedPassword(password *string, state accountState) (string, error) {
	if password == nil {
		return "", nil
	}
	if err := validatePassword(*password, state); err != nil {
		return "", err
	}
	return s.hasher.Hash(*password)
}

func newAccountState(input NewAccount) (accountState, error) {
	role, err := parseRole(input.Role)
	if err != nil {
		return accountState{}, err
	}
	return accountState{
		username: domain.NormalizeUsername(input.Username),
		email:    normalizeEmail(input.Email),
		role:     role,
		teamID:   input.TeamID,
	}, nil
}

func (c AccountChanges) applyTo(current Account) (accountState, error) {
	next := accountState{username: current.Username, role: current.Role, teamID: current.TeamID}
	if current.Email != nil {
		next.email = *current.Email
	}
	if c.Email != nil {
		next.email = normalizeEmail(*c.Email)
	}
	if c.Role != nil {
		role, err := parseRole(*c.Role)
		if err != nil {
			return accountState{}, err
		}
		next.role = role
		next.teamID = nil
	}
	if c.TeamID != nil {
		next.teamID = c.TeamID
	} else if next.role == domain.RoleTeamAdmin && current.Role == domain.RoleTeamAdmin {
		next.teamID = current.TeamID
	}
	return next, nil
}

// A changed role, team or password means existing sessions carry stale privileges
// or a compromised credential, so the user has to sign in again.
func revokesSessions(current Account, next accountState, newHash string) bool {
	return next.role != current.Role || !sameTeam(current.TeamID, next.teamID) || newHash != ""
}

func sameTeam(first, second *int64) bool {
	if first == nil || second == nil {
		return first == second
	}
	return *first == *second
}

func applyUpdate(ctx context.Context, queries dbgen.Querier, id int64, next accountState, newHash string, mustRevoke bool) error {
	err := queries.UpdateAccount(ctx, dbgen.UpdateAccountParams{
		ID:     id,
		Email:  optionalText(next.email),
		Role:   string(next.role),
		TeamID: optionalInt8(next.teamID),
	})
	if err != nil {
		return conflictOrWrap(err, "update account")
	}
	if newHash != "" {
		if err := storePasswordHash(ctx, queries, id, newHash); err != nil {
			return err
		}
	}
	if !mustRevoke {
		return nil
	}
	return revokeSessions(ctx, queries, dbgen.RevokeUserSessionsParams{UserID: id})
}

func storePasswordHash(ctx context.Context, queries dbgen.Querier, id int64, hash string) error {
	if err := queries.UpdateUserPasswordHash(ctx, dbgen.UpdateUserPasswordHashParams{ID: id, PasswordHash: hash}); err != nil {
		return fmt.Errorf("update password of account %d: %w", id, err)
	}
	return nil
}

func revokeSessions(ctx context.Context, queries dbgen.Querier, params dbgen.RevokeUserSessionsParams) error {
	if err := queries.RevokeUserSessions(ctx, params); err != nil {
		return fmt.Errorf("revoke sessions of account %d: %w", params.UserID, err)
	}
	return nil
}

func conflictOrWrap(err error, action string) error {
	constraint, isViolation := db.UniqueViolation(err)
	if !isViolation {
		return fmt.Errorf("%s: %w", action, err)
	}
	if constraint == "users_email_key" {
		return domain.NewConflictError("email", "ya hay un usuario con ese email")
	}
	return domain.NewConflictError("username", "ya hay un usuario con ese nombre")
}

func listParams(filter Filter) (dbgen.ListAccountsParams, error) {
	params := dbgen.ListAccountsParams{Search: optionalText(escapeLike(strings.ToLower(strings.TrimSpace(filter.Search))))}
	if filter.Role != "" {
		role, err := parseRole(filter.Role)
		if err != nil {
			return dbgen.ListAccountsParams{}, err
		}
		params.Role = optionalText(string(role))
	}
	switch filter.Status {
	case "":
	case statusActive, statusInactive:
		params.IsActive = pgtype.Bool{Bool: filter.Status == statusActive, Valid: true}
	default:
		return dbgen.ListAccountsParams{}, domain.NewValidationError("status", "el estado debe ser active o inactive")
	}
	return params, nil
}

// Usernames may contain "_", which LIKE would otherwise read as a wildcard.
func escapeLike(search string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(search)
}
