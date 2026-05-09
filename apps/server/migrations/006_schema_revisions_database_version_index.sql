DROP INDEX IF EXISTS idx_schema_revisions_project_version;

CREATE UNIQUE INDEX IF NOT EXISTS idx_schema_revisions_project_database_version
  ON schema_revisions(project_id, database_id, version_number);
