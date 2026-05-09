CREATE TABLE IF NOT EXISTS database_api_keys (
  id TEXT PRIMARY KEY,
  database_id TEXT NOT NULL,
  name TEXT NOT NULL,
  key_hash TEXT NOT NULL,
  key_prefix TEXT NOT NULL,
  revoked_at TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(database_id) REFERENCES project_databases(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_database_api_keys_database_id
  ON database_api_keys(database_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_database_api_keys_key_hash
  ON database_api_keys(key_hash);
