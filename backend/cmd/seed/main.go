package main

import (
	"context"
	"log"

	"github.com/racecontrol/backend/internal/auth"
	"github.com/racecontrol/backend/internal/config"
	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("seed: %v", err)
	}

	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
	defer pool.Close()

	runner := seed.NewRunner(pool)
	if err := runner.Reference(ctx); err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Println("reference data seeded")

	if cfg.IsProduction() || cfg.SeedDemoPassword == "" {
		return
	}

	hasher, err := auth.NewPasswordHasher(cfg.Argon2)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
	if err := runner.DemoAccounts(ctx, hasher, cfg.SeedDemoPassword); err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Println("demo accounts seeded")
}
