package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/racecontrol/backend/internal/auth"
)

// Environment tells the process which deployment it is running in.
type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
)

// Config holds every runtime setting, read once from the environment.
type Config struct {
	Environment        Environment
	DatabaseURL        string
	HTTPAddr           string
	Argon2             auth.Argon2Params
	Sessions           SessionTTLs
	SeedDemoPassword   string
	CORSAllowedOrigins []string
	TrustProxyHeaders  bool
}

type SessionTTLs struct {
	Idle     time.Duration
	Absolute time.Duration
}

const maxArgon2Parallelism = 255

// Load reads the configuration from the environment, failing fast on missing or
// malformed values so a misconfigured process never starts.
func Load() (Config, error) {
	environment, err := environmentFromEnv()
	if err != nil {
		return Config{}, err
	}
	argon2Params, err := argon2FromEnv()
	if err != nil {
		return Config{}, err
	}
	sessionTTLs, err := sessionTTLsFromEnv()
	if err != nil {
		return Config{}, err
	}
	// Only a reverse proxy that overwrites X-Forwarded-For makes it trustworthy;
	// otherwise any client could forge the IP recorded in login_attempts.
	trustProxyHeaders, err := boolFromEnv("TRUST_PROXY_HEADERS", false)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment:        environment,
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		HTTPAddr:           os.Getenv("HTTP_ADDR"),
		Argon2:             argon2Params,
		Sessions:           sessionTTLs,
		SeedDemoPassword:   os.Getenv("SEED_DEMO_PASSWORD"),
		CORSAllowedOrigins: listFromEnv("CORS_ALLOWED_ORIGINS", defaultDevelopmentOrigins),
		TrustProxyHeaders:  trustProxyHeaders,
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func argon2FromEnv() (auth.Argon2Params, error) {
	memory, err := uintFromEnv("ARGON2_MEMORY_KIB", 65536)
	if err != nil {
		return auth.Argon2Params{}, err
	}
	iterations, err := uintFromEnv("ARGON2_ITERATIONS", 3)
	if err != nil {
		return auth.Argon2Params{}, err
	}
	parallelism, err := uintFromEnv("ARGON2_PARALLELISM", 2)
	if err != nil {
		return auth.Argon2Params{}, err
	}
	if parallelism > maxArgon2Parallelism {
		return auth.Argon2Params{}, fmt.Errorf("ARGON2_PARALLELISM: %d exceeds the Argon2id limit of %d",
			parallelism, maxArgon2Parallelism)
	}
	return auth.Argon2Params{MemoryKiB: memory, Iterations: iterations, Parallelism: uint8(parallelism)}, nil
}

func sessionTTLsFromEnv() (SessionTTLs, error) {
	idle, err := durationFromEnv("SESSION_IDLE_TTL", 12*time.Hour)
	if err != nil {
		return SessionTTLs{}, err
	}
	absolute, err := durationFromEnv("SESSION_ABSOLUTE_TTL", 168*time.Hour)
	if err != nil {
		return SessionTTLs{}, err
	}
	if idle >= absolute {
		return SessionTTLs{}, fmt.Errorf("SESSION_IDLE_TTL (%s) must be shorter than SESSION_ABSOLUTE_TTL (%s)", idle, absolute)
	}
	return SessionTTLs{Idle: idle, Absolute: absolute}, nil
}

func (c Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.HTTPAddr == "" {
		return errors.New("HTTP_ADDR is required")
	}
	// A leaked .env must not be able to create a predictable admin account in production.
	if c.IsProduction() && c.SeedDemoPassword != "" {
		return errors.New("SEED_DEMO_PASSWORD must not be set when APP_ENV=production")
	}
	return nil
}

// IsProduction reports whether the process runs in production.
func (c Config) IsProduction() bool {
	return c.Environment == EnvironmentProduction
}

func environmentFromEnv() (Environment, error) {
	raw := os.Getenv("APP_ENV")
	if raw == "" {
		return EnvironmentDevelopment, nil
	}

	switch Environment(raw) {
	case EnvironmentDevelopment, EnvironmentProduction:
		return Environment(raw), nil
	default:
		return "", fmt.Errorf("APP_ENV: %q is not one of %q or %q",
			raw, EnvironmentDevelopment, EnvironmentProduction)
	}
}

var defaultDevelopmentOrigins = []string{"http://localhost:8081", "http://localhost:19006"}

func listFromEnv(name string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	var values []string
	for _, value := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func uintFromEnv(name string, fallback uint32) (uint32, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be an unsigned integer: %w", name, err)
	}
	return uint32(parsed), nil
}

func durationFromEnv(name string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a Go duration such as 12h: %w", name, err)
	}
	return parsed, nil
}

func boolFromEnv(name string, fallback bool) (bool, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false: %w", name, err)
	}
	return parsed, nil
}
