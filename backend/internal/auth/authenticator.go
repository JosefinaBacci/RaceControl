package auth

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/sessions"
)

const (
	MaxConsecutiveFailures = 5
	LockoutWindow          = 15 * time.Minute
)

type UserStore interface {
	GetUserByUsername(ctx context.Context, username string) (dbgen.User, error)
	UpdateUserPasswordHash(ctx context.Context, arg dbgen.UpdateUserPasswordHashParams) error
	CountFailuresSinceLastSuccess(ctx context.Context, arg dbgen.CountFailuresSinceLastSuccessParams) (int64, error)
	InsertLoginAttempt(ctx context.Context, arg dbgen.InsertLoginAttemptParams) error
}

type SessionIssuer interface {
	Issue(ctx context.Context, userID int64, client sessions.Client) (string, error)
}

type Credentials struct {
	Username string
	Password string
}

type LoginResult struct {
	Token    string
	Identity domain.Identity
}

type Authenticator struct {
	users    UserStore
	hasher   *PasswordHasher
	sessions SessionIssuer
	clock    func() time.Time
}

func NewAuthenticator(users UserStore, hasher *PasswordHasher, sessions SessionIssuer) *Authenticator {
	return &Authenticator{users: users, hasher: hasher, sessions: sessions, clock: time.Now}
}

func (a *Authenticator) Login(ctx context.Context, credentials Credentials, client sessions.Client) (LoginResult, error) {
	username := domain.NormalizeUsername(credentials.Username)
	if err := validateCredentialsFormat(username, credentials.Password); err != nil {
		return LoginResult{}, err
	}

	if err := a.ensureNotLocked(ctx, username); err != nil {
		return LoginResult{}, err
	}

	user, err := a.verifyCredentials(ctx, username, credentials.Password)
	if errors.Is(err, domain.ErrInvalidCredentials) {
		if recordErr := a.recordAttempt(ctx, username, false, client.IP); recordErr != nil {
			return LoginResult{}, recordErr
		}
		return LoginResult{}, err
	}
	if err != nil {
		return LoginResult{}, err
	}

	if err := a.recordAttempt(ctx, username, true, client.IP); err != nil {
		return LoginResult{}, err
	}

	token, err := a.sessions.Issue(ctx, user.ID, client)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Token: token, Identity: identityOf(user)}, nil
}

func validateCredentialsFormat(username, password string) error {
	if !domain.IsValidUsername(username) {
		return domain.NewValidationError("username",
			"el usuario debe tener entre 3 y 32 caracteres: letras, números, punto, guion o guion bajo")
	}
	if password == "" {
		return domain.NewValidationError("password", "la contraseña es obligatoria")
	}
	if utf8.RuneCountInString(password) > MaxPasswordLength {
		return domain.NewValidationError("password", "la contraseña es demasiado larga")
	}
	return nil
}

func (a *Authenticator) ensureNotLocked(ctx context.Context, username string) error {
	failures, err := a.users.CountFailuresSinceLastSuccess(ctx, dbgen.CountFailuresSinceLastSuccessParams{
		Username: username,
		Since:    pgtype.Timestamptz{Time: a.clock().Add(-LockoutWindow), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("count failed logins: %w", err)
	}
	if failures >= MaxConsecutiveFailures {
		return domain.ErrAccountLocked
	}
	return nil
}

func (a *Authenticator) verifyCredentials(ctx context.Context, username, password string) (dbgen.User, error) {
	user, err := a.users.GetUserByUsername(ctx, username)
	if errors.Is(err, pgx.ErrNoRows) {
		a.hasher.BurnDecoy(password)
		return dbgen.User{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return dbgen.User{}, fmt.Errorf("load user: %w", err)
	}

	check, err := a.hasher.Verify(password, user.PasswordHash)
	if err != nil {
		return dbgen.User{}, fmt.Errorf("verify password of user %d: %w", user.ID, err)
	}
	if !check.Matches || !user.IsActive {
		return dbgen.User{}, domain.ErrInvalidCredentials
	}

	if check.NeedsRehash {
		if err := a.rehash(ctx, user.ID, password); err != nil {
			return dbgen.User{}, err
		}
	}
	return user, nil
}

func (a *Authenticator) rehash(ctx context.Context, userID int64, password string) error {
	hash, err := a.hasher.Hash(password)
	if err != nil {
		return err
	}
	if err := a.users.UpdateUserPasswordHash(ctx, dbgen.UpdateUserPasswordHashParams{ID: userID, PasswordHash: hash}); err != nil {
		return fmt.Errorf("store upgraded hash of user %d: %w", userID, err)
	}
	return nil
}

func (a *Authenticator) recordAttempt(ctx context.Context, username string, succeeded bool, ip *netip.Addr) error {
	err := a.users.InsertLoginAttempt(ctx, dbgen.InsertLoginAttemptParams{Username: username, Succeeded: succeeded, Ip: ip})
	if err != nil {
		return fmt.Errorf("record login attempt: %w", err)
	}
	return nil
}

func identityOf(user dbgen.User) domain.Identity {
	identity := domain.Identity{UserID: user.ID, Username: user.Username, Role: domain.Role(user.Role)}
	if user.TeamID.Valid {
		teamID := user.TeamID.Int64
		identity.TeamID = &teamID
	}
	return identity
}
