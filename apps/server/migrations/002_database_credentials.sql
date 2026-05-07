CREATE TABLE IF NOT EXISTS database_credentials (
  id TEXT PRIMARY KEY,
  database_id TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT 'Main credentials',
  username TEXT NOT NULL,
  password TEXT NOT NULL,
  type TEXT NOT NULL DEFAULT 'main',
  revoked_at TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(database_id) REFERENCES project_databases(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_database_credentials_database_id
  ON database_credentials(database_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_database_credentials_one_active_main
  ON database_credentials(database_id)
  WHERE type = 'main' AND revoked_at IS NULL;
