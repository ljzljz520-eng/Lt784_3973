package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabasePath string
	HTTPAddress  string
	PageSize     int
	SeedDemo     bool
}

func Load() Config {
	return Config{DatabasePath: valueOr("SELECTOR_DB", "selector.db"), HTTPAddress: valueOr("SELECTOR_ADDR", ":8080"), PageSize: positiveInt("SELECTOR_PAGE_SIZE", 10), SeedDemo: boolValue("SELECTOR_SEED", true)}
}

func (c Config) Validate() error {
	if c.DatabasePath == "" || c.HTTPAddress == "" {
		return fmt.Errorf("database path and HTTP address are required")
	}
	if c.PageSize < 1 || c.PageSize > 100 {
		return fmt.Errorf("page size must be between 1 and 100")
	}
	return nil
}

func valueOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func positiveInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func boolValue(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
