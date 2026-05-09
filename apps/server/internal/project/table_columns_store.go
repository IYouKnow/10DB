package project

import (
	"context"
	"database/sql"
	"time"

	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

type ownedDraftTable struct {
	ProjectID  string
	DatabaseID string
	Revision   types.SchemaRevision
	Table      types.TableBlueprint
}

func (s *Store) ListDraftTableColumns(ctx context.Context, tableID string) ([]types.DraftTableColumn, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, table_id, name, type, nullable, primary_key, default_value, created_at, updated_at
		FROM table_columns
		WHERE table_id = ?
		ORDER BY created_at ASC
	`, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]types.DraftTableColumn, 0)
	for rows.Next() {
		column, err := scanDraftTableColumn(rows)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

func (s *Store) GetDraftTableColumn(ctx context.Context, tableID, columnID string) (types.DraftTableColumn, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, table_id, name, type, nullable, primary_key, default_value, created_at, updated_at
		FROM table_columns
		WHERE table_id = ? AND id = ?
	`, tableID, columnID)
	return scanDraftTableColumn(row)
}

func (s *Store) CreateDraftTableColumn(ctx context.Context, column types.DraftTableColumn) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO table_columns (id, table_id, name, type, nullable, primary_key, default_value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		column.ID,
		column.TableID,
		column.Name,
		column.Type,
		boolToInt(column.Nullable),
		boolToInt(column.PrimaryKey),
		column.DefaultValue,
		column.CreatedAt.Format(time.RFC3339Nano),
		column.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) UpdateDraftTableColumn(ctx context.Context, column types.DraftTableColumn) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE table_columns
		SET name = ?, type = ?, nullable = ?, primary_key = ?, default_value = ?, updated_at = ?
		WHERE table_id = ? AND id = ?
	`,
		column.Name,
		column.Type,
		boolToInt(column.Nullable),
		boolToInt(column.PrimaryKey),
		column.DefaultValue,
		column.UpdatedAt.Format(time.RFC3339Nano),
		column.TableID,
		column.ID,
	)
	return err
}

func (s *Store) DeleteDraftTableColumn(ctx context.Context, tableID, columnID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM table_columns WHERE table_id = ? AND id = ?`, tableID, columnID)
	return err
}

func (s *Store) ReplaceDraftTableColumns(ctx context.Context, tableID string, columns []types.DraftTableColumn) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM table_columns WHERE table_id = ?`, tableID); err != nil {
		return err
	}
	for _, column := range columns {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO table_columns (id, table_id, name, type, nullable, primary_key, default_value, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			column.ID,
			column.TableID,
			column.Name,
			column.Type,
			boolToInt(column.Nullable),
			boolToInt(column.PrimaryKey),
			column.DefaultValue,
			column.CreatedAt.Format(time.RFC3339Nano),
			column.UpdatedAt.Format(time.RFC3339Nano),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) FindOwnedDraftTable(ctx context.Context, ownerUserID, tableID string) (ownedDraftTable, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.id, s.project_id, s.database_id, s.version_number, s.blueprint_json, s.blueprint_hash, s.generated_sql, s.created_at
		FROM schema_revisions s
		INNER JOIN projects p ON p.id = s.project_id
		WHERE p.owner_user_id = ?
		ORDER BY s.created_at DESC
	`, ownerUserID)
	if err != nil {
		return ownedDraftTable{}, err
	}
	defer rows.Close()

	for rows.Next() {
		revision, err := scanRevision(rows)
		if err != nil {
			return ownedDraftTable{}, err
		}
		for _, table := range revision.Blueprint.Tables {
			if table.ID == tableID {
				return ownedDraftTable{
					ProjectID:  revision.ProjectID,
					DatabaseID: revision.DatabaseID,
					Revision:   revision,
					Table:      table,
				}, nil
			}
		}
	}
	if err := rows.Err(); err != nil {
		return ownedDraftTable{}, err
	}
	return ownedDraftTable{}, sql.ErrNoRows
}

func (s *Store) BackfillDraftTableColumns(ctx context.Context, tableID string, columns []types.DraftTableColumn) error {
	existing, err := s.ListDraftTableColumns(ctx, tableID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}
	return s.ReplaceDraftTableColumns(ctx, tableID, columns)
}

func scanDraftTableColumn(scanner interface{ Scan(dest ...any) error }) (types.DraftTableColumn, error) {
	var column types.DraftTableColumn
	var createdAt, updatedAt string
	var nullable, primaryKey int
	if err := scanner.Scan(
		&column.ID,
		&column.TableID,
		&column.Name,
		&column.Type,
		&nullable,
		&primaryKey,
		&column.DefaultValue,
		&createdAt,
		&updatedAt,
	); err != nil {
		return types.DraftTableColumn{}, err
	}
	column.Nullable = nullable == 1
	column.PrimaryKey = primaryKey == 1
	column.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	column.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return column, nil
}
