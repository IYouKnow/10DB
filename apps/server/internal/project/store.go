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
	type legacyProjectDatabase struct {
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
	}

	type credentialBackfill struct {
		databaseID string
		username   string
		password   string
		createdAt  string
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS project_databases (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			server_id TEXT NULL,
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
		`CREATE TABLE IF NOT EXISTS database_credentials (
			id TEXT PRIMARY KEY,
			database_id TEXT NOT NULL,
			label TEXT NOT NULL DEFAULT 'Main credentials',
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'main',
			revoked_at TEXT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY(database_id) REFERENCES project_databases(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_database_credentials_database_id ON database_credentials(database_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_database_credentials_one_active_main
			ON database_credentials(database_id)
			WHERE type = 'main' AND revoked_at IS NULL`,
		`CREATE TABLE IF NOT EXISTS database_api_keys (
			id TEXT PRIMARY KEY,
			database_id TEXT NOT NULL,
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			key_prefix TEXT NOT NULL,
			permission TEXT NOT NULL DEFAULT 'read_write',
			revoked_at TEXT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY(database_id) REFERENCES project_databases(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_database_api_keys_database_id ON database_api_keys(database_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_database_api_keys_key_hash ON database_api_keys(key_hash)`,
		`CREATE TABLE IF NOT EXISTS table_columns (
			id TEXT PRIMARY KEY,
			table_id TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			nullable INTEGER NOT NULL DEFAULT 1,
			primary_key INTEGER NOT NULL DEFAULT 0,
			default_value TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_table_columns_table_id ON table_columns(table_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_table_columns_table_name ON table_columns(table_id, name)`,
		`CREATE TABLE IF NOT EXISTS database_servers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			engine TEXT NOT NULL DEFAULT 'postgres',
			host TEXT NOT NULL,
			port INTEGER NOT NULL DEFAULT 5432,
			admin_username TEXT NOT NULL,
			admin_password TEXT NOT NULL,
			ssl_mode TEXT NOT NULL DEFAULT 'disable',
			default_database TEXT NOT NULL DEFAULT 'postgres',
			is_default INTEGER NOT NULL DEFAULT 0,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_database_servers_default ON database_servers(is_default)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	if err := ensureColumn(ctx, s.db, "project_databases", "server_id", "TEXT NULL"); err != nil {
		return err
	}

	if err := ensureColumn(ctx, s.db, "schema_revisions", "database_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DROP INDEX IF EXISTS idx_schema_revisions_project_version`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_schema_revisions_project_database_version ON schema_revisions(project_id, database_id, version_number)`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_schema_revisions_database_id ON schema_revisions(database_id, version_number DESC)`); err != nil {
		return err
	}

	if err := ensureColumn(ctx, s.db, "database_api_keys", "permission", "TEXT NOT NULL DEFAULT 'read_write'"); err != nil {
		return err
	}

	legacyDatabases := make([]legacyProjectDatabase, 0)
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
		legacyDatabases = append(legacyDatabases, legacyProjectDatabase{
			projectID: projectID,
			status:    status,
			dbName:    dbName,
			roleName:  roleName,
			password:  password,
			host:      host,
			port:      port,
			sslMode:   sslMode,
			createdAt: createdAt,
			updatedAt: updatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, item := range legacyDatabases {
		var count int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM project_databases WHERE project_id = ? AND pg_database_name = ?`, item.projectID, item.dbName).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO project_databases (
				id, project_id, server_id, engine, name, status, pg_database_name, pg_role_name, pg_password_encrypted,
				pg_host, pg_port, pg_ssl_mode, position_x, position_y, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			uuid.NewString(),
			item.projectID,
			nil,
			"postgresql",
			"PostgreSQL Database",
			item.status,
			item.dbName,
			item.roleName,
			item.password,
			item.host,
			item.port,
			item.sslMode,
			80.0,
			80.0,
			item.createdAt,
			item.updatedAt,
		); err != nil {
			return err
		}
	}

	credentialBackfills := make([]credentialBackfill, 0)
	credentialRows, err := s.db.QueryContext(ctx, `
		SELECT id, pg_role_name, pg_password_encrypted, created_at
		FROM project_databases
	`)
	if err != nil {
		return err
	}
	defer credentialRows.Close()

	for credentialRows.Next() {
		var (
			databaseID string
			username   string
			password   string
			createdAt  string
		)
		if err := credentialRows.Scan(&databaseID, &username, &password, &createdAt); err != nil {
			return err
		}
		credentialBackfills = append(credentialBackfills, credentialBackfill{
			databaseID: databaseID,
			username:   username,
			password:   password,
			createdAt:  createdAt,
		})
	}

	if err := credentialRows.Err(); err != nil {
		return err
	}
	if err := credentialRows.Close(); err != nil {
		return err
	}

	for _, item := range credentialBackfills {
		var count int
		if err := s.db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM database_credentials
			WHERE database_id = ? AND type = 'main' AND revoked_at IS NULL
		`, item.databaseID).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO database_credentials (id, database_id, label, username, password, type, revoked_at, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			uuid.NewString(),
			item.databaseID,
			"Main credentials",
			item.username,
			item.password,
			"main",
			nil,
			item.createdAt,
		); err != nil {
			return err
		}
	}

	return nil
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

func (s *Store) CountProjectsByOwner(ctx context.Context, ownerUserID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects WHERE owner_user_id = ?`, ownerUserID).Scan(&count)
	return count, err
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
		SELECT id, project_id, server_id, engine, name, status, pg_database_name, pg_role_name, pg_password_encrypted,
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

func (s *Store) CountProjectDatabases(ctx context.Context, projectID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM project_databases WHERE project_id = ?`, projectID).Scan(&count)
	return count, err
}

func (s *Store) GetProjectDatabase(ctx context.Context, projectID, databaseID string) (types.ProjectDatabase, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, server_id, engine, name, status, pg_database_name, pg_role_name, pg_password_encrypted,
		       pg_host, pg_port, pg_ssl_mode, position_x, position_y, created_at, updated_at
		FROM project_databases
		WHERE id = ? AND project_id = ?
	`, databaseID, projectID)
	return scanProjectDatabase(row)
}

func (s *Store) GetProjectDatabaseByOwner(ctx context.Context, ownerUserID, databaseID string) (types.ProjectDatabase, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT d.id, d.project_id, d.server_id, d.engine, d.name, d.status, d.pg_database_name, d.pg_role_name, d.pg_password_encrypted,
		       d.pg_host, d.pg_port, d.pg_ssl_mode, d.position_x, d.position_y, d.created_at, d.updated_at
		FROM project_databases d
		INNER JOIN projects p ON p.id = d.project_id
		WHERE d.id = ? AND p.owner_user_id = ?
	`, databaseID, ownerUserID)
	return scanProjectDatabase(row)
}

func (s *Store) CreateProjectDatabase(ctx context.Context, database types.ProjectDatabase) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO project_databases (
			id, project_id, server_id, engine, name, status, pg_database_name, pg_role_name, pg_password_encrypted,
			pg_host, pg_port, pg_ssl_mode, position_x, position_y, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		database.ID,
		database.ProjectID,
		database.ServerID,
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
		SET server_id = ?, engine = ?, name = ?, status = ?, pg_database_name = ?, pg_role_name = ?, pg_password_encrypted = ?,
		    pg_host = ?, pg_port = ?, pg_ssl_mode = ?, position_x = ?, position_y = ?, updated_at = ?
		WHERE id = ? AND project_id = ?
	`,
		database.ServerID,
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

func (s *Store) CreateDatabaseCredential(ctx context.Context, credential types.DatabaseCredential) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO database_credentials (id, database_id, label, username, password, type, revoked_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		credential.ID,
		credential.DatabaseID,
		credential.Label,
		credential.Username,
		credential.Password,
		credential.Type,
		timePointerString(credential.RevokedAt),
		credential.CreatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetActiveMainDatabaseCredential(ctx context.Context, databaseID string) (types.DatabaseCredential, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, database_id, label, username, password, type, revoked_at, created_at
		FROM database_credentials
		WHERE database_id = ? AND type = 'main' AND revoked_at IS NULL
		LIMIT 1
	`, databaseID)
	return scanDatabaseCredential(row)
}

func (s *Store) CountProjects(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects`).Scan(&count)
	return count, err
}

func (s *Store) CountProjectDatabasesTotal(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM project_databases`).Scan(&count)
	return count, err
}

func (s *Store) CountFailedProjectDatabases(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM project_databases WHERE status = ?`, string(types.ProjectStatusApplyFailed)).Scan(&count)
	return count, err
}

func (s *Store) CountDatabaseServers(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM database_servers`).Scan(&count)
	return count, err
}

func (s *Store) CountActiveDatabaseServers(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM database_servers WHERE is_active = 1`).Scan(&count)
	return count, err
}

func (s *Store) ListDatabaseServers(ctx context.Context) ([]types.DatabaseServer, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, engine, host, port, admin_username, admin_password, ssl_mode, default_database,
		       is_default, is_active, created_at, updated_at
		FROM database_servers
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := make([]types.DatabaseServer, 0)
	for rows.Next() {
		server, err := scanDatabaseServer(rows)
		if err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}
	return servers, rows.Err()
}

func (s *Store) GetDatabaseServer(ctx context.Context, id string) (types.DatabaseServer, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, engine, host, port, admin_username, admin_password, ssl_mode, default_database,
		       is_default, is_active, created_at, updated_at
		FROM database_servers
		WHERE id = ?
	`, id)
	return scanDatabaseServer(row)
}

func (s *Store) GetDefaultDatabaseServer(ctx context.Context) (types.DatabaseServer, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, engine, host, port, admin_username, admin_password, ssl_mode, default_database,
		       is_default, is_active, created_at, updated_at
		FROM database_servers
		WHERE is_default = 1
		LIMIT 1
	`)
	return scanDatabaseServer(row)
}

func (s *Store) GetActiveDefaultDatabaseServer(ctx context.Context) (types.DatabaseServer, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, engine, host, port, admin_username, admin_password, ssl_mode, default_database,
		       is_default, is_active, created_at, updated_at
		FROM database_servers
		WHERE is_default = 1 AND is_active = 1
		LIMIT 1
	`)
	return scanDatabaseServer(row)
}

func (s *Store) CreateDatabaseServer(ctx context.Context, server types.DatabaseServer) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO database_servers (
			id, name, engine, host, port, admin_username, admin_password, ssl_mode, default_database,
			is_default, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		server.ID,
		server.Name,
		server.Engine,
		server.Host,
		server.Port,
		server.AdminUsername,
		server.AdminPassword,
		server.SSLMode,
		server.DefaultDatabase,
		boolToInt(server.IsDefault),
		boolToInt(server.IsActive),
		server.CreatedAt.Format(time.RFC3339Nano),
		server.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) UpdateDatabaseServer(ctx context.Context, server types.DatabaseServer) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE database_servers
		SET name = ?, engine = ?, host = ?, port = ?, admin_username = ?, admin_password = ?, ssl_mode = ?,
		    default_database = ?, is_default = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`,
		server.Name,
		server.Engine,
		server.Host,
		server.Port,
		server.AdminUsername,
		server.AdminPassword,
		server.SSLMode,
		server.DefaultDatabase,
		boolToInt(server.IsDefault),
		boolToInt(server.IsActive),
		server.UpdatedAt.Format(time.RFC3339Nano),
		server.ID,
	)
	return err
}

func (s *Store) SetDefaultDatabaseServer(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `UPDATE database_servers SET is_default = 0, updated_at = ?`, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE database_servers SET is_default = 1, updated_at = ? WHERE id = ?`, time.Now().UTC().Format(time.RFC3339Nano), id); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) DeleteDatabaseServer(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM database_servers WHERE id = ?`, id)
	return err
}

func (s *Store) CountDatabasesByServerID(ctx context.Context, serverID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM project_databases WHERE server_id = ?`, serverID).Scan(&count)
	return count, err
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
	var serverID sql.NullString
	err := scanner.Scan(
		&database.ID,
		&database.ProjectID,
		&serverID,
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
	if serverID.Valid {
		database.ServerID = &serverID.String
	}
	database.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	database.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return database, nil
}

func scanDatabaseServer(scanner interface{ Scan(dest ...any) error }) (types.DatabaseServer, error) {
	var server types.DatabaseServer
	var createdAt, updatedAt string
	var isDefault, isActive int
	err := scanner.Scan(
		&server.ID,
		&server.Name,
		&server.Engine,
		&server.Host,
		&server.Port,
		&server.AdminUsername,
		&server.AdminPassword,
		&server.SSLMode,
		&server.DefaultDatabase,
		&isDefault,
		&isActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return types.DatabaseServer{}, err
	}
	server.IsDefault = isDefault == 1
	server.IsActive = isActive == 1
	server.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	server.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return server, nil
}

func scanDatabaseCredential(scanner interface{ Scan(dest ...any) error }) (types.DatabaseCredential, error) {
	var credential types.DatabaseCredential
	var revokedAt sql.NullString
	var createdAt string
	if err := scanner.Scan(
		&credential.ID,
		&credential.DatabaseID,
		&credential.Label,
		&credential.Username,
		&credential.Password,
		&credential.Type,
		&revokedAt,
		&createdAt,
	); err != nil {
		return types.DatabaseCredential{}, err
	}
	if revokedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, revokedAt.String)
		if err == nil {
			credential.RevokedAt = &parsed
		}
	}
	credential.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return credential, nil
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

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
