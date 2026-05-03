CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  pg_database_name TEXT NOT NULL,
  pg_role_name TEXT NOT NULL,
  pg_password_encrypted TEXT NOT NULL,
  pg_host TEXT NOT NULL,
  pg_port INTEGER NOT NULL,
  pg_ssl_mode TEXT NOT NULL,
  last_applied_revision_id TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS schema_revisions (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  version_number INTEGER NOT NULL,
  blueprint_json TEXT NOT NULL,
  blueprint_hash TEXT NOT NULL,
  generated_sql TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_schema_revisions_project_version
  ON schema_revisions(project_id, version_number);

CREATE TABLE IF NOT EXISTS apply_runs (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  schema_revision_id TEXT NOT NULL,
  status TEXT NOT NULL,
  sql_executed TEXT NOT NULL DEFAULT '',
  error_message TEXT NOT NULL DEFAULT '',
  started_at TEXT NOT NULL,
  finished_at TEXT,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(schema_revision_id) REFERENCES schema_revisions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS project_events (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  type TEXT NOT NULL,
  message TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
