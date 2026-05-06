package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

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
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			password_hash TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	if err := ensureColumn(ctx, s.db, "users", "role", "TEXT NOT NULL DEFAULT 'user'"); err != nil {
		return err
	}

	exists, err := s.columnExists(ctx, "projects", "owner_user_id")
	if err != nil {
		return err
	}
	if !exists {
		if _, err := s.db.ExecContext(ctx, `ALTER TABLE projects ADD COLUMN owner_user_id TEXT`); err != nil {
			return err
		}
	}

	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_projects_owner_user_id ON projects(owner_user_id)`); err != nil {
		return err
	}

	return nil
}

func (s *Store) Create(ctx context.Context, user types.User) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (id, email, name, role, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		user.ID,
		user.Email,
		user.Name,
		user.Role,
		user.PasswordHash,
		user.CreatedAt.Format(time.RFC3339Nano),
		user.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetByEmail(ctx context.Context, email string) (types.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, name, role, password_hash, created_at, updated_at
		FROM users
		WHERE email = ?
	`, strings.ToLower(strings.TrimSpace(email)))
	return scanUser(row)
}

func (s *Store) GetByID(ctx context.Context, id string) (types.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, name, role, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id)
	return scanUser(row)
}

func (s *Store) Count(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (s *Store) columnExists(ctx context.Context, tableName, columnName string) (bool, error) {
	rows, err := s.db.QueryContext(ctx, `PRAGMA table_info(`+tableName+`)`)
	if err != nil {
		return false, err
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
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}

	return false, rows.Err()
}

func scanUser(scanner interface{ Scan(dest ...any) error }) (types.User, error) {
	var user types.User
	var createdAt, updatedAt string
	if err := scanner.Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.PasswordHash, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return types.User{}, err
		}
		return types.User{}, err
	}
	user.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return user, nil
}

func ensureColumn(ctx context.Context, db *sql.DB, tableName, columnName, columnDef string) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(`+tableName+`)`)
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

	_, err = db.ExecContext(ctx, `ALTER TABLE `+tableName+` ADD COLUMN `+columnName+` `+columnDef)
	return err
}
