package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/term"

	"github.com/racecontrol/backend/internal/auth"
	"github.com/racecontrol/backend/internal/config"
	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

const passwordEnvVar = "ADMIN_PASSWORD"

func main() {
	username := flag.String("username", "", "username of the new FIA administrator")
	email := flag.String("email", "", "optional email of the new FIA administrator")
	flag.Parse()

	if err := run(domain.NormalizeUsername(*username), strings.ToLower(strings.TrimSpace(*email))); err != nil {
		log.Fatalf("create-admin: %v", err)
	}
}

func run(username, email string) error {
	if err := domain.ValidateUsername(username); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	password, err := readPassword()
	if err != nil {
		return err
	}
	if err := auth.ValidatePasswordPolicy(password, username, email); err != nil {
		return err
	}

	hasher, err := auth.NewPasswordHasher(cfg.Argon2)
	if err != nil {
		return err
	}
	hash, err := hasher.Hash(password)
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	user, err := dbgen.New(pool).CreateUser(ctx, dbgen.CreateUserParams{
		Username:     username,
		Email:        pgtype.Text{String: email, Valid: email != ""},
		PasswordHash: hash,
		Role:         string(domain.RoleFIAAdmin),
	})
	if constraint, isViolation := db.UniqueViolation(err); isViolation {
		return fmt.Errorf("username or email already in use: %s", constraint)
	}
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	log.Printf("FIA administrator %q created with id %d", user.Username, user.ID)
	return nil
}

func readPassword() (string, error) {
	if password := os.Getenv(passwordEnvVar); password != "" {
		return password, nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return "", fmt.Errorf("read password from stdin: %w", err)
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	first, err := promptHidden("Password: ")
	if err != nil {
		return "", err
	}
	second, err := promptHidden("Repeat password: ")
	if err != nil {
		return "", err
	}
	if first != second {
		return "", errors.New("passwords do not match")
	}
	return first, nil
}

func promptHidden(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(raw), nil
}
