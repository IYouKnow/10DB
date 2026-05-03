package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS project_databases (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			engine TEXT NOT NULL,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			pg_database_name TEXT NOT NULL,
			pg_role_name TEXT NOT NULL,
			pg_password_encrypted TEXT NOT NULL,
			pg_host TEXT NOT NULL,
			pg_port INTEGER NOT NULL,
			pg_ssl_mode TEXT NOT NULL,
			position_x REAL NOT NULL DEFAULT 80,
			position_y REAL NOT NULL DEFAULT 80,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_project_databases_project_id ON project_databases(project_id)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	if err := ensureColumn(ctx, s.db, "schema_revisions", "database_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_schema_revisions_database_id ON schema_revisions(database_id, version_number DESC)`); err != nil {
		return err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, status, pg_database_name, pg_role_name, pg_password_encrypted, pg_host, pg_port, pg_ssl_mode, created_at, updated_at
		FROM projects
		WHERE pg_database_name <> ''
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			projectID string
			status    string
			dbName    string
			roleName  string
			password  string
			host      string
			port      int
			sslMode   string
			createdAt string
			updatedAt string
		)
		if err := rows.Scan(&projectID, &status, &dbName, &roleName, &password, &host, &port, &sslMode, &createdAt, &updatedAt); err != nil {
			return err
		}

		var count int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM project_databases WHERE project_id = ? AND pg_database_name = ?`, projectID, dbName).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO project_databases (
				id, project_id, engine, name, status, pg_database_name, pg_role_name, pg_password_encrypted,
				pg_host, pg_port, pg_ssl_mode, position_x, position_y, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			uuid.NewString(),
			projectID,
			"postgresql",
			"PostgreSQL Database",
			status,
			dbName,
			roleName,
			password,
			host,
			port,
			sslMode,
			80.0,
			80.0,
			createdAt,
			updatedAt,
		); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (s *Store) ListProjects(ctx context.Context, ownerUserID string) ([]types.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, owner_user_id, name, slug, description, status, pg_database_name, pg_role_name, pg_password_encrypted,
		       pg_host, pg_port, pg_ssl_mode, last_applied_revision_id, created_at, updated_at
		FROM projects
		WHERE owner_user_id = ?
		ORDER BY created_at DESC
	`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]types.Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for index := range projects {
		databases, err := s.ListProjectDatabases(ctx, projects[index].ID)
		if err != nil {
			return nil, err
		}
		projects[index].Databases = databases
	}

	return projects, nil
}

func (s *Store) GetProject(ctx context.Context, ownerUserID, id string) (types.Project, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, owner_user_id, name, slug, description, status, pg_database_name, pg_role_name, pg_password_encrypted,
		       pg_host, pg_port, pg_ssl_mode, last_applied_revision_id, created_at, updated_at
		FROM projects WHERE id = ? AND owner_user_id = ?
	`, id, ownerUserID)
	project, err := scanProject(row)
	if err != nil {
		return types.Project{}, err
	}
	project.Databases, err = s.ListProjectDatabases(ctx, project.ID)
	if err != nil {
		return types.Project{}, err
	}
	return project, nil
}

func (s *Store) CreateProject(ctx context.Context, project types.Project) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO projects (
			id, owner_user_id, name, slug, description, status, pg_database_name, pg_role_name, pg_password_encrypted,
			pg_host, pg_port, pg_ssl_mode, last_applied_revision_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		project.ID,
		project.OwnerUserID,
		project.Name,
		project.Slug,
		project.Description,
		project.Status,
		project.PGDatabaseName,
		project.PGRoleName,
		project.PGPasswordEncrypted,
		project.PGHost,
		project.PGPort,
		project.PGSSLMode,
		project.LastAppliedRevisionID,
		project.CreatedAt.Format(time.RFC3339Nano),
		project.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) UpdateProject(ctx context.Context, project types.Project) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE projects
		SET owner_user_id = ?, name = ?, slug = ?, description = ?, status = ?, pg_database_name = ?, pg_role_name = ?,
		    pg_password_encrypted = ?, pg_host = ?, pg_port = ?, pg_ssl_mode = ?, last_applied_revision_id = ?, updated_at = ?
		WHERE id = ?
	`,
		project.OwnerUserID,
		project.Name,
		project.Slug,
		project.Description,
		project.Status,
		project.PGDatabaseName,
		project.PGRoleName,
		project.PGPasswordEncrypted,
		project.PGHost,
		project.PGPort,
		project.PGSSLMode,
		project.LastAppliedRevisionID,
		project.UpdatedAt.Format(time.RFC3339Nano),
		project.ID,
	)
	return err
}

func (s *Store) DeleteProject(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	return err
}

func (s *Store) ListProjectDatabases(ctx context.Context, projectID string) ([]types.ProjectDatabase, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, engine, name, status, pg_database_name, pg_role_name, pg_password_encrypted,
		       pg_host, pg_port, pg_ssl_mode, position_x, position_y, created_at, updated_at
		FROM project_databases
		WHERE project_id = ?
		ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	databases := make([]types.ProjectDatabase, 0)
	for rows.Next() {
		database, err := scanProjectDatabase(rows)
		if err != nil {
			return nil, err
		}
		databases = append(databases, database)
	}
	return databases, rows.Err()
}

func (s *Store) GetProjectDatabase(ctx context.Context, projectID, databaseID string) (types.ProjectDatabase, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, engine, name, status, pg_database_name, pg_role_name, pg_password_encrypted,
		       pg_host, pg_port, pg_ssl_mode, position_x, position_y, created_at, updated_at
		FROM project_databases
		WHERE id = ? AND project_id = ?
	`, databaseID, projectID)
	return scanProjectDatabase(row)
}

func (s *Store) CreateProjectDatabase(ctx context.Context, database types.ProjectDatabase) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO project_databases (
			id, project_id, engine, name, status, pg_database_name, pg_role_name, pg_password_encrypted,
			pg_host, pg_port, pg_ssl_mode, position_x, position_y, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		database.ID,
		database.ProjectID,
		database.Engine,
		database.Name,
		database.Status,
		database.PGDatabaseName,
		database.PGRoleName,
		database.PGPasswordEncrypted,
		database.PGHost,
		database.PGPort,
		database.PGSSLMode,
		database.PositionX,
		database.PositionY,
		database.CreatedAt.Format(time.RFC3339Nano),
		database.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) UpdateProjectDatabase(ctx context.Context, database types.ProjectDatabase) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE project_databases
		SET engine = ?, name = ?, status = ?, pg_database_name = ?, pg_role_name = ?, pg_password_encrypted = ?,
		    pg_host = ?, pg_port = ?, pg_ssl_mode = ?, position_x = ?, position_y = ?, updated_at = ?
		WHERE id = ? AND project_id = ?
	`,
		database.Engine,
		database.Name,
		database.Status,
		database.PGDatabaseName,
		database.PGRoleName,
		database.PGPasswordEncrypted,
		database.PGHost,
		database.PGPort,
		database.PGSSLMode,
		database.PositionX,
		database.PositionY,
		database.UpdatedAt.Format(time.RFC3339Nano),
		database.ID,
		database.ProjectID,
	)
	return err
}

func (s *Store) DeleteProjectDatabase(ctx context.Context, projectID, databaseID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM project_databases WHERE id = ? AND project_id = ?`, databaseID, projectID)
	return err
}

func (s *Store) SaveSchemaRevision(ctx context.Context, projectID string, blueprint types.SchemaBlueprint, blueprintHash, generatedSQL string) (types.SchemaRevision, error) {
	var currentVersion int
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_number), 0) FROM schema_revisions WHERE project_id = ? AND database_id = ?`, projectID, blueprint.DatabaseID).Scan(&currentVersion)
	revision := types.SchemaRevision{
		ID:            uuid.NewString(),
		ProjectID:     projectID,
		DatabaseID:    blueprint.DatabaseID,
		VersionNumber: currentVersion + 1,
		Blueprint:     blueprint,
		BlueprintHash: blueprintHash,
		GeneratedSQL:  generatedSQL,
		CreatedAt:     time.Now().UTC(),
	}
	raw, err := json.Marshal(revision.Blueprint)
	if err != nil {
		return types.SchemaRevision{}, err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO schema_revisions (id, project_id, database_id, version_number, blueprint_json, blueprint_hash, generated_sql, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, revision.ID, revision.ProjectID, revision.DatabaseID, revision.VersionNumber, string(raw), revision.BlueprintHash, revision.GeneratedSQL, revision.CreatedAt.Format(time.RFC3339Nano))
	return revision, err
}

func (s *Store) GetLatestSchemaRevision(ctx context.Context, projectID, databaseID string) (types.SchemaRevision, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, database_id, version_number, blueprint_json, blueprint_hash, generated_sql, created_at
		FROM schema_revisions
		WHERE project_id = ? AND database_id = ?
		ORDER BY version_number DESC
		LIMIT 1
	`, projectID, databaseID)
	return scanRevision(row)
}

func (s *Store) ListSchemaRevisions(ctx context.Context, projectID, databaseID string) ([]types.SchemaRevision, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, database_id, version_number, blueprint_json, blueprint_hash, generated_sql, created_at
		FROM schema_revisions
		WHERE project_id = ? AND database_id = ?
		ORDER BY version_number DESC
	`, projectID, databaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	revisions := make([]types.SchemaRevision, 0)
	for rows.Next() {
		revision, err := scanRevision(rows)
		if err != nil {
			return nil, err
		}
		revisions = append(revisions, revision)
	}
	return revisions, rows.Err()
}

func (s *Store) CreateApplyRun(ctx context.Context, run types.ApplyRun) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO apply_runs (id, project_id, schema_revision_id, status, sql_executed, error_message, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, run.ID, run.ProjectID, run.SchemaRevisionID, run.Status, run.SQLExecuted, run.ErrorMessage, run.StartedAt.Format(time.RFC3339Nano), timePointerString(run.FinishedAt))
	return err
}

func (s *Store) UpdateApplyRun(ctx context.Context, run types.ApplyRun) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE apply_runs
		SET status = ?, sql_executed = ?, error_message = ?, finished_at = ?
		WHERE id = ?
	`, run.Status, run.SQLExecuted, run.ErrorMessage, timePointerString(run.FinishedAt), run.ID)
	return err
}

func scanProject(scanner interface{ Scan(dest ...any) error }) (types.Project, error) {
	var project types.Project
	var createdAt, updatedAt string
	var lastApplied sql.NullString
	err := scanner.Scan(
		&project.ID,
		&project.OwnerUserID,
		&project.Name,
		&project.Slug,
		&project.Description,
		&project.Status,
		&project.PGDatabaseName,
		&project.PGRoleName,
		&project.PGPasswordEncrypted,
		&project.PGHost,
		&project.PGPort,
		&project.PGSSLMode,
		&lastApplied,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return types.Project{}, err
	}
	project.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	project.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	if lastApplied.Valid {
		project.LastAppliedRevisionID = &lastApplied.String
	}
	return project, nil
}

func scanProjectDatabase(scanner interface{ Scan(dest ...any) error }) (types.ProjectDatabase, error) {
	var database types.ProjectDatabase
	var createdAt, updatedAt string
	err := scanner.Scan(
		&database.ID,
		&database.ProjectID,
		&database.Engine,
		&database.Name,
		&database.Status,
		&database.PGDatabaseName,
		&database.PGRoleName,
		&database.PGPasswordEncrypted,
		&database.PGHost,
		&database.PGPort,
		&database.PGSSLMode,
		&database.PositionX,
		&database.PositionY,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return types.ProjectDatabase{}, err
	}
	database.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	database.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return database, nil
}

func scanRevision(scanner interface{ Scan(dest ...any) error }) (types.SchemaRevision, error) {
	var revision types.SchemaRevision
	var raw string
	var createdAt string
	if err := scanner.Scan(&revision.ID, &revision.ProjectID, &revision.DatabaseID, &revision.VersionNumber, &raw, &revision.BlueprintHash, &revision.GeneratedSQL, &createdAt); err != nil {
		return types.SchemaRevision{}, err
	}
	if err := json.Unmarshal([]byte(raw), &revision.Blueprint); err != nil {
		return types.SchemaRevision{}, fmt.Errorf("decode blueprint: %w", err)
	}
	revision.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return revision, nil
}

func ensureColumn(ctx context.Context, db *sql.DB, tableName, columnName, columnDef string) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &primaryKey); err != nil {
			return err
		}
		if name == columnName {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnDef))
	return err
}

func timePointerString(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339Nano)
}
