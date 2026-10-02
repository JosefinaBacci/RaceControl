package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds every runtime setting, read once from the environment.
type Config struct {
	DatabaseURL        string
	HTTPAddr           string
	Argon2MemoryKiB    uint32
	Argon2Iterations   uint32
	Argon2Parallelism  uint8
	SessionIdleTTL     time.Duration
	SessionAbsoluteTTL time.Duration
}

// Load reads the configuration from the environment, failing fast on missing or
// malformed values so a misconfigured process never starts.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		HTTPAddr:    os.Getenv("HTTP_ADDR"),
	}

	var err error
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

	return cfg, nil
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
