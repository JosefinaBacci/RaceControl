package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/racecontrol/backend/internal/config"
	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/transport"
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

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           transport.NewRouter(pool),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go shutdownOnSignal(ctx, server, log.Printf)

	log.Printf("api listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("api: %v", err)
	}
}

func shutdownOnSignal(ctx context.Context, server *http.Server, logf func(string, ...any)) {
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logf("api: graceful shutdown failed: %v", err)
	}
}
