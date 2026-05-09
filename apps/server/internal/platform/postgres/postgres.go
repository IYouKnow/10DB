package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

type AdminConfig struct {
	Host     string
	Port     int
	DBName   string
	User     string
	Password string
	SSLMode  string
}

type Service struct {
	cfg  AdminConfig
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg AdminConfig) (*Service, error) {
	dsn := buildDSN(cfg.Host, cfg.Port, cfg.DBName, cfg.User, cfg.Password, cfg.SSLMode)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Service{cfg: cfg, pool: pool}, nil
}

func (s *Service) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Service) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Service) CreateProjectDatabase(ctx context.Context, databaseName, roleName, password string) error {
	return CreateProjectDatabaseWithConfig(ctx, s.cfg, databaseName, roleName, password)
}

func (s *Service) CreateProjectDatabaseWithConfig(ctx context.Context, cfg AdminConfig, databaseName, roleName, password string) error {
	return CreateProjectDatabaseWithConfig(ctx, cfg, databaseName, roleName, password)
}

func (s *Service) DropProjectDatabase(ctx context.Context, databaseName, roleName string) error {
	return DropProjectDatabaseWithConfig(ctx, s.cfg, databaseName, roleName)
}

func (s *Service) DropProjectDatabaseWithConfig(ctx context.Context, cfg AdminConfig, databaseName, roleName string) error {
	return DropProjectDatabaseWithConfig(ctx, cfg, databaseName, roleName)
}

func (s *Service) ResetProjectSchema(ctx context.Context, project types.Project, password string) error {
	conn, err := pgx.Connect(ctx, buildDSN(project.PGHost, project.PGPort, project.PGDatabaseName, project.PGRoleName, password, project.PGSSLMode))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, `DROP SCHEMA IF EXISTS public CASCADE`); err != nil {
		return err
	}
	if _, err := conn.Exec(ctx, `CREATE SCHEMA public`); err != nil {
		return err
	}
	_, err = conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`)
	return err
}

func (s *Service) ApplySQL(ctx context.Context, project types.Project, password, sql string) error {
	conn, err := pgx.Connect(ctx, buildDSN(project.PGHost, project.PGPort, project.PGDatabaseName, project.PGRoleName, password, project.PGSSLMode))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, sql)
	return err
}

func (s *Service) DropTable(ctx context.Context, project types.Project, password, tableName string) error {
	conn, err := pgx.Connect(ctx, buildDSN(project.PGHost, project.PGPort, project.PGDatabaseName, project.PGRoleName, password, project.PGSSLMode))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", quoteIdent(tableName)))
	return err
}

func (s *Service) PublicTableCount(ctx context.Context, project types.Project, password string) (int, error) {
	conn, err := pgx.Connect(ctx, buildDSN(project.PGHost, project.PGPort, project.PGDatabaseName, project.PGRoleName, password, project.PGSSLMode))
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)
	var count int
	err = conn.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`).Scan(&count)
	return count, err
}

func (s *Service) ListTables(ctx context.Context, project types.Project, password string) ([]types.TableInfo, error) {
	conn, err := pgx.Connect(ctx, buildDSN(project.PGHost, project.PGPort, project.PGDatabaseName, project.PGRoleName, password, project.PGSSLMode))
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)
	rows, err := conn.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tables := make([]types.TableInfo, 0)
	for rows.Next() {
		var item types.TableInfo
		if err := rows.Scan(&item.Name); err != nil {
			return nil, err
		}
		tables = append(tables, item)
	}
	return tables, rows.Err()
}

func (s *Service) ListColumns(ctx context.Context, project types.Project, password, tableName string) ([]types.ColumnInfo, error) {
	conn, err := pgx.Connect(ctx, buildDSN(project.PGHost, project.PGPort, project.PGDatabaseName, project.PGRoleName, password, project.PGSSLMode))
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)
	rows, err := conn.Query(ctx, `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position
	`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := make([]types.ColumnInfo, 0)
	for rows.Next() {
		var item types.ColumnInfo
		var nullable string
		if err := rows.Scan(&item.Name, &item.DataType, &nullable); err != nil {
			return nil, err
		}
		item.Nullable = nullable == "YES"
		columns = append(columns, item)
	}
	return columns, rows.Err()
}

func (s *Service) ListRows(ctx context.Context, project types.Project, password, tableName string, limit, offset int) (types.TableRows, error) {
	conn, err := s.connectProject(ctx, project, password)
	if err != nil {
		return types.TableRows{}, err
	}
	defer conn.Close(ctx)
	query := fmt.Sprintf(`SELECT * FROM %s LIMIT %d OFFSET %d`, quoteIdent(tableName), limit, offset)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return types.TableRows{}, err
	}
	defer rows.Close()
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, 0, len(fieldDescriptions))
	for _, fd := range fieldDescriptions {
		columns = append(columns, string(fd.Name))
	}
	items := make([]map[string]any, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return types.TableRows{}, err
		}
		row := map[string]any{}
		for i, name := range columns {
			row[name] = values[i]
		}
		items = append(items, row)
	}
	return types.TableRows{Columns: columns, Rows: items, Limit: limit, Offset: offset}, rows.Err()
}

func (s *Service) ListDataRows(ctx context.Context, project types.Project, password, tableName string, limit int) (types.TableRows, error) {
	conn, err := s.connectProject(ctx, project, password)
	if err != nil {
		return types.TableRows{}, err
	}
	defer conn.Close(ctx)

	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY id::text LIMIT $1`, quoteIdent(tableName))
	rows, err := conn.Query(ctx, query, limit)
	if err != nil {
		return types.TableRows{}, err
	}
	defer rows.Close()

	return scanRows(rows, limit, 0)
}

func (s *Service) GetDataRow(ctx context.Context, project types.Project, password, tableName, id string) (map[string]any, error) {
	conn, err := s.connectProject(ctx, project, password)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	query := fmt.Sprintf(`SELECT * FROM %s WHERE id::text = $1 LIMIT 1`, quoteIdent(tableName))
	return querySingleRow(ctx, conn, query, id)
}

func (s *Service) InsertDataRow(ctx context.Context, project types.Project, password, tableName string, values map[string]any) (map[string]any, error) {
	conn, err := s.connectProject(ctx, project, password)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	if len(values) == 0 {
		query := fmt.Sprintf(`INSERT INTO %s DEFAULT VALUES RETURNING *`, quoteIdent(tableName))
		return querySingleRow(ctx, conn, query)
	}

	keys := sortedMapKeys(values)
	quotedColumns := make([]string, 0, len(keys))
	placeholders := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys))
	for index, key := range keys {
		quotedColumns = append(quotedColumns, quoteIdent(key))
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
		args = append(args, values[key])
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) RETURNING *`,
		quoteIdent(tableName),
		strings.Join(quotedColumns, ", "),
		strings.Join(placeholders, ", "),
	)
	return querySingleRow(ctx, conn, query, args...)
}

func (s *Service) UpdateDataRow(ctx context.Context, project types.Project, password, tableName, id string, values map[string]any) (map[string]any, error) {
	conn, err := s.connectProject(ctx, project, password)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	if len(values) == 0 {
		query := fmt.Sprintf(`SELECT * FROM %s WHERE id::text = $1 LIMIT 1`, quoteIdent(tableName))
		return querySingleRow(ctx, conn, query, id)
	}

	keys := sortedMapKeys(values)
	setClauses := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys)+1)
	for index, key := range keys {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", quoteIdent(key), index+1))
		args = append(args, values[key])
	}
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE %s SET %s WHERE id::text = $%d RETURNING *`,
		quoteIdent(tableName),
		strings.Join(setClauses, ", "),
		len(args),
	)
	return querySingleRow(ctx, conn, query, args...)
}

func (s *Service) DeleteDataRow(ctx context.Context, project types.Project, password, tableName, id string) error {
	conn, err := s.connectProject(ctx, project, password)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	commandTag, err := conn.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE id::text = $1`, quoteIdent(tableName)), id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func CreateProjectDatabaseWithConfig(ctx context.Context, cfg AdminConfig, databaseName, roleName, password string) error {
	pool, err := pgxpool.New(ctx, buildDSN(cfg.Host, cfg.Port, cfg.DBName, cfg.User, cfg.Password, cfg.SSLMode))
	if err != nil {
		return err
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, fmt.Sprintf(
		"CREATE ROLE %s LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION PASSWORD %s",
		quoteIdent(roleName),
		quoteLiteral(password),
	)); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s OWNER %s", quoteIdent(databaseName), quoteIdent(roleName))); err != nil {
		_, _ = pool.Exec(ctx, fmt.Sprintf("DROP ROLE IF EXISTS %s", quoteIdent(roleName)))
		return err
	}

	// Remove default PUBLIC connectivity and grant the tenant role access only
	// to its own database. This prevents connecting to other managed databases
	// that would otherwise inherit PUBLIC CONNECT privileges.
	if err := lockRoleToDatabase(ctx, pool, cfg.DBName, databaseName, roleName); err != nil {
		_, _ = pool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", quoteIdent(databaseName)))
		_, _ = pool.Exec(ctx, fmt.Sprintf("DROP ROLE IF EXISTS %s", quoteIdent(roleName)))
		return err
	}

	projectConn, err := pgx.Connect(ctx, buildDSN(cfg.Host, cfg.Port, databaseName, cfg.User, cfg.Password, cfg.SSLMode))
	if err != nil {
		return err
	}
	defer projectConn.Close(ctx)
	if _, err := projectConn.Exec(ctx, `REVOKE ALL ON SCHEMA public FROM PUBLIC`); err != nil {
		return err
	}
	if _, err := projectConn.Exec(ctx, fmt.Sprintf("GRANT USAGE, CREATE ON SCHEMA public TO %s", quoteIdent(roleName))); err != nil {
		return err
	}
	_, err = projectConn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`)
	return err
}

func lockRoleToDatabase(ctx context.Context, pool *pgxpool.Pool, adminDatabaseName, databaseName, roleName string) error {
	databases := []string{adminDatabaseName, databaseName}
	if adminDatabaseName != "postgres" {
		databases = append(databases, "postgres")
	}

	for _, name := range databases {
		if _, err := pool.Exec(ctx, fmt.Sprintf("REVOKE CONNECT, TEMPORARY ON DATABASE %s FROM PUBLIC", quoteIdent(name))); err != nil {
			return err
		}
	}

	if _, err := pool.Exec(ctx, fmt.Sprintf("GRANT CONNECT, TEMPORARY ON DATABASE %s TO %s", quoteIdent(databaseName), quoteIdent(roleName))); err != nil {
		return err
	}
	if adminDatabaseName != databaseName {
		if _, err := pool.Exec(ctx, fmt.Sprintf("REVOKE CONNECT, TEMPORARY ON DATABASE %s FROM %s", quoteIdent(adminDatabaseName), quoteIdent(roleName))); err != nil {
			return err
		}
	}
	if adminDatabaseName != "postgres" && databaseName != "postgres" {
		if _, err := pool.Exec(ctx, fmt.Sprintf("REVOKE CONNECT, TEMPORARY ON DATABASE %s FROM %s", quoteIdent("postgres"), quoteIdent(roleName))); err != nil {
			return err
		}
	}
	return nil
}

func DropProjectDatabaseWithConfig(ctx context.Context, cfg AdminConfig, databaseName, roleName string) error {
	pool, err := pgxpool.New(ctx, buildDSN(cfg.Host, cfg.Port, cfg.DBName, cfg.User, cfg.Password, cfg.SSLMode))
	if err != nil {
		return err
	}
	defer pool.Close()

	_, _ = pool.Exec(ctx, `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`, databaseName)
	if _, err := pool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", quoteIdent(databaseName))); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, fmt.Sprintf("DROP ROLE IF EXISTS %s", quoteIdent(roleName))); err != nil {
		return err
	}
	return nil
}

func TestAdminConnection(ctx context.Context, cfg AdminConfig) error {
	conn, err := pgx.Connect(ctx, buildDSN(cfg.Host, cfg.Port, cfg.DBName, cfg.User, cfg.Password, cfg.SSLMode))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	return conn.Ping(ctx)
}

func (s *Service) connectProject(ctx context.Context, project types.Project, password string) (*pgx.Conn, error) {
	return pgx.Connect(ctx, buildDSN(project.PGHost, project.PGPort, project.PGDatabaseName, project.PGRoleName, password, project.PGSSLMode))
}

func buildDSN(host string, port int, dbName, user, password, sslMode string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		url.PathEscape(dbName),
		url.QueryEscape(sslMode),
	)
}

func quoteIdent(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func quoteLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func scanRows(rows pgx.Rows, limit, offset int) (types.TableRows, error) {
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, 0, len(fieldDescriptions))
	for _, fd := range fieldDescriptions {
		columns = append(columns, string(fd.Name))
	}

	items := make([]map[string]any, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return types.TableRows{}, err
		}
		row := map[string]any{}
		for index, name := range columns {
			row[name] = values[index]
		}
		items = append(items, row)
	}
	if err := rows.Err(); err != nil {
		return types.TableRows{}, err
	}
	return types.TableRows{Columns: columns, Rows: items, Limit: limit, Offset: offset}, nil
}

func querySingleRow(ctx context.Context, conn *pgx.Conn, query string, args ...any) (map[string]any, error) {
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := scanRows(rows, 1, 0)
	if err != nil {
		return nil, err
	}
	if len(result.Rows) == 0 {
		return nil, pgx.ErrNoRows
	}
	if len(result.Rows) > 1 {
		return nil, errors.New("expected a single row")
	}
	return result.Rows[0], nil
}

func sortedMapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
