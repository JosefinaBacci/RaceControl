package main

import (
	"context"
	"log"

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

	if err := seed.NewRunner(pool).Reference(ctx); err != nil {
		log.Fatalf("seed: %v", err)
	}

	log.Println("reference data seeded")
}
