package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/racecontrol/backend/internal/config"
	"github.com/racecontrol/backend/internal/db/migrate"
)

func main() {
	steps := flag.Int("down", 0, "roll back this many migrations instead of applying them; a negative value rolls back everything")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}

	if err := run(cfg.DatabaseURL, *steps); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}

func run(databaseURL string, downSteps int) error {
	migrator, err := migrate.New(databaseURL)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := migrator.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "migrate: closing: %v\n", closeErr)
		}
	}()

	if downSteps != 0 {
		version, err := migrator.Down(downSteps)
		if err != nil {
			return err
		}
		fmt.Printf("rolled back, database is at version %d\n", version)
		return nil
	}

	version, err := migrator.Up()
	if err != nil {
		return err
	}
	fmt.Printf("migrations applied, database is at version %d\n", version)
	return nil
}
