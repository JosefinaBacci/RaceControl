package sessions

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

const (
	tokenByteLength = 32
	touchInterval   = 5 * time.Minute
)

type Store interface {
	CreateSession(ctx context.Context, arg dbgen.CreateSessionParams) (int64, error)
	GetActiveSessionByTokenHash(ctx context.Context, arg dbgen.GetActiveSessionByTokenHashParams) (dbgen.GetActiveSessionByTokenHashRow, error)
	TouchSession(ctx context.Context, arg dbgen.TouchSessionParams) error
	RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error
	RevokeUserSessions(ctx context.Context, userID int64) error
}

type Config struct {
	IdleTTL     time.Duration
	AbsoluteTTL time.Duration
	Clock       func() time.Time
}

type Client struct {
	IP        *netip.Addr
	UserAgent string
}

type Service struct {
	store  Store
	config Config
}

func NewService(store Store, config Config) *Service {
	if config.Clock == nil {
		config.Clock = time.Now
	}
	return &Service{store: store, config: config}
}

func (s *Service) Issue(ctx context.Context, userID int64, client Client) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}

	_, err = s.store.CreateSession(ctx, dbgen.CreateSessionParams{
		UserID:            userID,
		TokenHash:         hashToken(token),
		AbsoluteExpiresAt: db.Timestamptz(s.config.Clock().Add(s.config.AbsoluteTTL)),
		Ip:                client.IP,
		UserAgent:         pgtype.Text{String: client.UserAgent, Valid: client.UserAgent != ""},
	})
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return token, nil
}

func (s *Service) Resolve(ctx context.Context, token string) (domain.Identity, error) {
	if token == "" {
		return domain.Identity{}, domain.ErrUnauthenticated
	}

	now := s.config.Clock()
	row, err := s.store.GetActiveSessionByTokenHash(ctx, dbgen.GetActiveSessionByTokenHashParams{
		TokenHash:  hashToken(token),
		Now:        db.Timestamptz(now),
		IdleCutoff: db.Timestamptz(now.Add(-s.config.IdleTTL)),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Identity{}, domain.ErrUnauthenticated
	}
	if err != nil {
		return domain.Identity{}, fmt.Errorf("resolve session: %w", err)
	}

	if err := s.touchIfStale(ctx, row, now); err != nil {
		return domain.Identity{}, err
	}

	return identityFromRow(row), nil
}

func (s *Service) Revoke(ctx context.Context, token string) error {
	if err := s.store.RevokeSessionByTokenHash(ctx, hashToken(token)); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (s *Service) RevokeAllForUser(ctx context.Context, userID int64) error {
	if err := s.store.RevokeUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("revoke sessions of user %d: %w", userID, err)
	}
	return nil
}

func (s *Service) touchIfStale(ctx context.Context, row dbgen.GetActiveSessionByTokenHashRow, now time.Time) error {
	if now.Sub(row.LastUsedAt.Time) < touchInterval {
		return nil
	}
	if err := s.store.TouchSession(ctx, dbgen.TouchSessionParams{ID: row.SessionID, LastUsedAt: db.Timestamptz(now)}); err != nil {
		return fmt.Errorf("touch session: %w", err)
	}
	return nil
}

func identityFromRow(row dbgen.GetActiveSessionByTokenHashRow) domain.Identity {
	return domain.Identity{
		UserID:    row.UserID,
		Username:  row.Username,
		Role:      domain.Role(row.Role),
		TeamID:    db.Int64Pointer(row.TeamID),
		SessionID: row.SessionID,
	}
}

func newToken() (string, error) {
	raw := make([]byte, tokenByteLength)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
