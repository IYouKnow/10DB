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

func (s *Store) ListProjects(ctx context.Context) ([]types.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, slug, description, status, pg_database_name, pg_role_name, pg_password_encrypted,
		       pg_host, pg_port, pg_ssl_mode, last_applied_revision_id, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`)
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
	return projects, rows.Err()
}

func (s *Store) GetProject(ctx context.Context, id string) (types.Project, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, slug, description, status, pg_database_name, pg_role_name, pg_password_encrypted,
		       pg_host, pg_port, pg_ssl_mode, last_applied_revision_id, created_at, updated_at
		FROM projects WHERE id = ?
	`, id)
	return scanProject(row)
}

func (s *Store) CreateProject(ctx context.Context, project types.Project) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO projects (
			id, name, slug, description, status, pg_database_name, pg_role_name, pg_password_encrypted,
			pg_host, pg_port, pg_ssl_mode, last_applied_revision_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		project.ID,
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
		SET name = ?, slug = ?, description = ?, status = ?, pg_database_name = ?, pg_role_name = ?,
		    pg_password_encrypted = ?, pg_host = ?, pg_port = ?, pg_ssl_mode = ?, last_applied_revision_id = ?, updated_at = ?
		WHERE id = ?
	`,
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

func (s *Store) SaveSchemaRevision(ctx context.Context, projectID string, blueprint types.SchemaBlueprint, blueprintHash, generatedSQL string) (types.SchemaRevision, error) {
	var currentVersion int
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_number), 0) FROM schema_revisions WHERE project_id = ?`, projectID).Scan(&currentVersion)
	revision := types.SchemaRevision{
		ID:            uuid.NewString(),
		ProjectID:     projectID,
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
		INSERT INTO schema_revisions (id, project_id, version_number, blueprint_json, blueprint_hash, generated_sql, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, revision.ID, revision.ProjectID, revision.VersionNumber, string(raw), revision.BlueprintHash, revision.GeneratedSQL, revision.CreatedAt.Format(time.RFC3339Nano))
	return revision, err
}

func (s *Store) GetLatestSchemaRevision(ctx context.Context, projectID string) (types.SchemaRevision, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, version_number, blueprint_json, blueprint_hash, generated_sql, created_at
		FROM schema_revisions
		WHERE project_id = ?
		ORDER BY version_number DESC
		LIMIT 1
	`, projectID)
	return scanRevision(row)
}

func (s *Store) ListSchemaRevisions(ctx context.Context, projectID string) ([]types.SchemaRevision, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, version_number, blueprint_json, blueprint_hash, generated_sql, created_at
		FROM schema_revisions
		WHERE project_id = ?
		ORDER BY version_number DESC
	`, projectID)
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

func scanRevision(scanner interface{ Scan(dest ...any) error }) (types.SchemaRevision, error) {
	var revision types.SchemaRevision
	var raw string
	var createdAt string
	if err := scanner.Scan(&revision.ID, &revision.ProjectID, &revision.VersionNumber, &raw, &revision.BlueprintHash, &revision.GeneratedSQL, &createdAt); err != nil {
		return types.SchemaRevision{}, err
	}
	if err := json.Unmarshal([]byte(raw), &revision.Blueprint); err != nil {
		return types.SchemaRevision{}, fmt.Errorf("decode blueprint: %w", err)
	}
	revision.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return revision, nil
}

func timePointerString(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339Nano)
}
