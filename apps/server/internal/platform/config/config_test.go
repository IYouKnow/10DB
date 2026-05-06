package config

import "testing"

func TestLoadDoesNotRequirePGAdminEnvVars(t *testing.T) {
	t.Setenv("APP_MASTER_KEY", "test-master-key")
	t.Setenv("PG_ADMIN_HOST", "")
	t.Setenv("PG_ADMIN_PORT", "")
	t.Setenv("PG_ADMIN_DB", "")
	t.Setenv("PG_ADMIN_USER", "")
	t.Setenv("PG_ADMIN_PASSWORD", "")
	t.Setenv("PG_SSL_MODE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MasterKey != "test-master-key" {
		t.Fatalf("MasterKey = %q, want %q", cfg.MasterKey, "test-master-key")
	}
}
