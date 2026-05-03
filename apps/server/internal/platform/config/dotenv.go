package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadDotEnvIfPresent walks upward from the current working directory and
// loads the first .env file it finds without overriding existing env vars.
func LoadDotEnvIfPresent() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	for {
		candidate := filepath.Join(wd, ".env")
		if _, err := os.Stat(candidate); err == nil {
			return loadDotEnvFile(candidate)
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			return nil
		}
		wd = parent
	}
}

func loadDotEnvFile(path string) error {
	if _, exists := os.LookupEnv("APP_DOTENV_DIR"); !exists {
		if err := os.Setenv("APP_DOTENV_DIR", filepath.Dir(path)); err != nil {
			return fmt.Errorf("set APP_DOTENV_DIR: %w", err)
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %s: %w", key, err)
		}
	}

	return scanner.Err()
}
