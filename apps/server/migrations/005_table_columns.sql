CREATE TABLE IF NOT EXISTS table_columns (
  id TEXT PRIMARY KEY,
  table_id TEXT NOT NULL,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  nullable INTEGER NOT NULL DEFAULT 1,
  primary_key INTEGER NOT NULL DEFAULT 0,
  default_value TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_table_columns_table_id
  ON table_columns(table_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_table_columns_table_name
  ON table_columns(table_id, name);
