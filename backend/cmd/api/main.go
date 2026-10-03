package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/racecontrol/backend/internal/auth"
	"github.com/racecontrol/backend/internal/config"
	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/sessions"
	"github.com/racecontrol/backend/internal/teams"
	"github.com/racecontrol/backend/internal/transport"
	"github.com/racecontrol/backend/internal/users"
)

const (
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("api: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("api: %v", err)
	}
	defer pool.Close()

	router, err := buildRouter(cfg, pool)
	if err != nil {
		log.Fatalf("api: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go shutdownOnSignal(ctx, server, log.Printf)

	log.Printf("api listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("api: %v", err)
	}
}

func buildRouter(cfg config.Config, pool *pgxpool.Pool) (http.Handler, error) {
	hasher, err := auth.NewPasswordHasher(cfg.Argon2)
	if err != nil {
		return nil, err
	}

	queries := dbgen.New(pool)
	sessionService := sessions.NewService(queries, sessions.Config{
		IdleTTL:     cfg.Sessions.Idle,
		AbsoluteTTL: cfg.Sessions.Absolute,
	})

	return transport.NewRouter(transport.Dependencies{
		Pool:              pool,
		Authenticator:     auth.NewAuthenticator(queries, hasher, sessionService),
		Sessions:          sessionService,
		Users:             users.NewService(queries, db.NewTransactor(pool), hasher),
		Teams:             teams.NewService(queries),
		Cookie:            transport.CookieConfig{Secure: cfg.IsProduction(), MaxAge: cfg.Sessions.Absolute},
		AllowedOrigins:    cfg.CORSAllowedOrigins,
		TrustProxyHeaders: cfg.TrustProxyHeaders,
	}), nil
}

func shutdownOnSignal(ctx context.Context, server *http.Server, logf func(string, ...any)) {
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logf("api: graceful shutdown failed: %v", err)
	}
}
