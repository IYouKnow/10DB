package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppAddr        string
	AppBaseURL     string
	AllowedOrigins []string
	MasterKey      string
	SessionTTL     time.Duration
	ControlDBPath  string
}

func Load() (Config, error) {
	cfg := Config{
		AppAddr:        getEnv("APP_ADDR", ":8080"),
		AppBaseURL:     getEnv("APP_BASE_URL", "http://localhost:8080"),
		AllowedOrigins: parseOrigins(getEnv("APP_ALLOWED_ORIGINS", "http://localhost:8080,http://localhost:5173,http://127.0.0.1:5173")),
		MasterKey:      os.Getenv("APP_MASTER_KEY"),
		ControlDBPath:  resolveControlDBPath(getEnv("CONTROL_DB_PATH", "apps/server/data/10db-launch.sqlite")),
	}

	ttlHours, err := strconv.Atoi(getEnv("APP_SESSION_TTL_HOURS", "24"))
	if err != nil {
		return Config{}, fmt.Errorf("parse APP_SESSION_TTL_HOURS: %w", err)
	}
	cfg.SessionTTL = time.Duration(ttlHours) * time.Hour

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var missing []string
	for key, value := range map[string]string{
		"APP_MASTER_KEY": c.MasterKey,
	} {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	if _, err := url.Parse(c.AppBaseURL); err != nil {
		return fmt.Errorf("invalid APP_BASE_URL: %w", err)
	}
	for _, origin := range c.AllowedOrigins {
		if _, err := url.Parse(origin); err != nil {
			return fmt.Errorf("invalid APP_ALLOWED_ORIGINS entry %q: %w", origin, err)
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func parseOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(strings.TrimRight(part, "/"))
		if trimmed == "" {
			continue
		}
		origins = append(origins, trimmed)
	}
	return origins
}

func resolveControlDBPath(raw string) string {
	if strings.TrimSpace(raw) == "" {
		raw = "apps/server/data/10db-launch.sqlite"
	}
	if filepath.IsAbs(raw) {
		return raw
	}

	baseDir := os.Getenv("APP_DOTENV_DIR")
	if strings.TrimSpace(baseDir) == "" {
		baseDir, _ = os.Getwd()
	}
	if strings.TrimSpace(baseDir) == "" {
		return raw
	}
	return filepath.Clean(filepath.Join(baseDir, raw))
}
