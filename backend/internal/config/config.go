package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
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
	Argon2MemoryKiB    uint32
	Argon2Iterations   uint32
	Argon2Parallelism  uint8
	SessionIdleTTL     time.Duration
	SessionAbsoluteTTL time.Duration
	SeedDemoPassword   string
}

// Load reads the configuration from the environment, failing fast on missing or
// malformed values so a misconfigured process never starts.
func Load() (Config, error) {
	environment, err := environmentFromEnv()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment:      environment,
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		HTTPAddr:         os.Getenv("HTTP_ADDR"),
		SeedDemoPassword: os.Getenv("SEED_DEMO_PASSWORD"),
	}

	if cfg.Argon2MemoryKiB, err = uintFromEnv("ARGON2_MEMORY_KIB", 65536); err != nil {
		return Config{}, err
	}
	if cfg.Argon2Iterations, err = uintFromEnv("ARGON2_ITERATIONS", 3); err != nil {
		return Config{}, err
	}
	argon2Parallelism, err := uintFromEnv("ARGON2_PARALLELISM", 2)
	if err != nil {
		return Config{}, err
	}
	if argon2Parallelism > 255 {
		return Config{}, fmt.Errorf("ARGON2_PARALLELISM: %d exceeds the Argon2id limit of 255", argon2Parallelism)
	}
	cfg.Argon2Parallelism = uint8(argon2Parallelism)

	if cfg.SessionIdleTTL, err = durationFromEnv("SESSION_IDLE_TTL", 12*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.SessionAbsoluteTTL, err = durationFromEnv("SESSION_ABSOLUTE_TTL", 168*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.SessionIdleTTL >= cfg.SessionAbsoluteTTL {
		return Config{}, fmt.Errorf(
			"SESSION_IDLE_TTL (%s) must be shorter than SESSION_ABSOLUTE_TTL (%s)",
			cfg.SessionIdleTTL, cfg.SessionAbsoluteTTL)
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}
	if err := cfg.validateSeedSafety(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// validateSeedSafety refuses to start when demo credentials are configured in
// production, so a leaked .env cannot create a predictable admin account.
func (c Config) validateSeedSafety() error {
	if c.Environment != EnvironmentProduction {
		return nil
	}
	if c.SeedDemoPassword != "" {
		return fmt.Errorf("SEED_DEMO_PASSWORD must not be set when APP_ENV=production")
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
