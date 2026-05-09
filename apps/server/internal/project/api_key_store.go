package project

import (
	"context"
	"database/sql"
	"errors"
	"time"

	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

func (s *Store) CreateDatabaseAPIKey(ctx context.Context, key types.DatabaseAPIKey) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO database_api_keys (id, database_id, name, key_hash, key_prefix, permission, revoked_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		key.ID,
		key.DatabaseID,
		key.Name,
		key.KeyHash,
		key.KeyPrefix,
		key.Permission,
		timePointerString(key.RevokedAt),
		key.CreatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) ListDatabaseAPIKeysByOwner(ctx context.Context, ownerUserID, databaseID string) ([]types.DatabaseAPIKey, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT k.id, k.database_id, k.name, k.key_hash, k.key_prefix, k.permission, k.revoked_at, k.created_at
		FROM database_api_keys k
		INNER JOIN project_databases d ON d.id = k.database_id
		INNER JOIN projects p ON p.id = d.project_id
		WHERE k.database_id = ? AND p.owner_user_id = ?
		ORDER BY k.created_at DESC
	`, databaseID, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]types.DatabaseAPIKey, 0)
	for rows.Next() {
		key, err := scanDatabaseAPIKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (s *Store) RevokeDatabaseAPIKeyByOwner(ctx context.Context, ownerUserID, databaseID, keyID string, revokedAt time.Time) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE database_api_keys
		SET revoked_at = ?
		WHERE id = ? AND database_id = ? AND revoked_at IS NULL
		  AND EXISTS (
			SELECT 1
			FROM project_databases d
			INNER JOIN projects p ON p.id = d.project_id
			WHERE d.id = database_api_keys.database_id AND p.owner_user_id = ?
		  )
	`, revokedAt.Format(time.RFC3339Nano), keyID, databaseID, ownerUserID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) GetActiveDatabaseAPIKeyByHash(ctx context.Context, keyHash string) (types.DatabaseAPIKey, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, database_id, name, key_hash, key_prefix, permission, revoked_at, created_at
		FROM database_api_keys
		WHERE key_hash = ? AND revoked_at IS NULL
		LIMIT 1
	`, keyHash)
	return scanDatabaseAPIKey(row)
}

func (s *Store) GetProjectDatabaseByID(ctx context.Context, databaseID string) (types.ProjectDatabase, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, server_id, engine, name, status, pg_database_name, pg_role_name, pg_password_encrypted,
		       pg_host, pg_port, pg_ssl_mode, position_x, position_y, created_at, updated_at
		FROM project_databases
		WHERE id = ?
	`, databaseID)
	return scanProjectDatabase(row)
}

func scanDatabaseAPIKey(scanner interface{ Scan(dest ...any) error }) (types.DatabaseAPIKey, error) {
	var key types.DatabaseAPIKey
	var revokedAt sql.NullString
	var createdAt string
	if err := scanner.Scan(
		&key.ID,
		&key.DatabaseID,
		&key.Name,
		&key.KeyHash,
		&key.KeyPrefix,
		&key.Permission,
		&revokedAt,
		&createdAt,
	); err != nil {
		return types.DatabaseAPIKey{}, err
	}
	if revokedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, revokedAt.String)
		if err == nil {
			key.RevokedAt = &parsed
		}
	}
	key.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return key, nil
}

func (s *Store) RequireOwnedDatabase(ctx context.Context, ownerUserID, databaseID string) error {
	_, err := s.GetProjectDatabaseByOwner(ctx, ownerUserID, databaseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return err
	}
	return nil
}
