package project

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/pedro/10db-launch/apps/server/internal/platform/crypto"
	"github.com/pedro/10db-launch/apps/server/internal/platform/postgres"
	types "github.com/pedro/10db-launch/apps/server/internal/types"
	"github.com/pedro/10db-launch/apps/server/internal/user"
)

type stubUsers struct {
	users map[string]types.User
}

func (s *stubUsers) GetByID(_ context.Context, id string) (types.User, error) {
	user, ok := s.users[id]
	if !ok {
		return types.User{}, sql.ErrNoRows
	}
	return user, nil
}

type stubPostgres struct{}

func (s *stubPostgres) CreateProjectDatabase(context.Context, string, string, string) error {
	return nil
}

func (s *stubPostgres) CreateProjectDatabaseWithConfig(context.Context, postgres.AdminConfig, string, string, string) error {
	return nil
}

func (s *stubPostgres) DropProjectDatabase(context.Context, string, string) error {
	return nil
}

func (s *stubPostgres) DropProjectDatabaseWithConfig(context.Context, postgres.AdminConfig, string, string) error {
	return nil
}

func (s *stubPostgres) ResetProjectSchema(context.Context, types.Project, string) error {
	return nil
}

func (s *stubPostgres) ApplySQL(context.Context, types.Project, string, string) error {
	return nil
}

func (s *stubPostgres) DropTable(context.Context, types.Project, string, string) error {
	return nil
}

func (s *stubPostgres) ListTables(context.Context, types.Project, string) ([]types.TableInfo, error) {
	return nil, nil
}

func (s *stubPostgres) ListColumns(context.Context, types.Project, string, string) ([]types.ColumnInfo, error) {
	return nil, nil
}

func (s *stubPostgres) ListRows(context.Context, types.Project, string, string, int, int) (types.TableRows, error) {
	return types.TableRows{}, nil
}

func TestNormalUserCannotCreateSecondProject(t *testing.T) {
	service := newTestService(t, map[string]types.User{
		"user-1": {ID: "user-1", Role: types.UserRoleUser},
	})

	ctx := context.Background()
	if _, err := service.Create(ctx, "user-1", CreateProjectInput{Name: "Project One", Slug: "project-one"}); err != nil {
		t.Fatalf("Create() first project error = %v", err)
	}

	_, err := service.Create(ctx, "user-1", CreateProjectInput{Name: "Project Two", Slug: "project-two"})
	if err == nil {
		t.Fatalf("Create() second project error = nil, want limit error")
	}

	limitErr, ok := err.(*LimitError)
	if !ok {
		t.Fatalf("Create() second project error = %T, want *LimitError", err)
	}
	if got := limitErr.Error(); got != "Project limit reached. Normal users can create 1 project." {
		t.Fatalf("Create() second project error = %q", got)
	}
}

func TestNormalUserCannotCreateFourthDatabase(t *testing.T) {
	service := newTestService(t, map[string]types.User{
		"user-1": {ID: "user-1", Role: types.UserRoleUser},
	})

	ctx := context.Background()
	project, err := service.Create(ctx, "user-1", CreateProjectInput{Name: "Project One", Slug: "project-one"})
	if err != nil {
		t.Fatalf("Create() project error = %v", err)
	}

	for i := 0; i < MaxDatabasesPerProject; i++ {
		if _, err := service.ProvisionPostgres(ctx, "user-1", project.ID, ProvisionPostgresInput{}); err != nil {
			t.Fatalf("ProvisionPostgres() database %d error = %v", i+1, err)
		}
	}

	_, err = service.ProvisionPostgres(ctx, "user-1", project.ID, ProvisionPostgresInput{})
	if err == nil {
		t.Fatalf("ProvisionPostgres() fourth database error = nil, want limit error")
	}

	limitErr, ok := err.(*LimitError)
	if !ok {
		t.Fatalf("ProvisionPostgres() fourth database error = %T, want *LimitError", err)
	}
	if got := limitErr.Error(); got != "Database limit reached. Each project can have up to 3 databases." {
		t.Fatalf("ProvisionPostgres() fourth database error = %q", got)
	}
}

func TestAdminCanExceedBothLimits(t *testing.T) {
	service := newTestService(t, map[string]types.User{
		"admin-1": {ID: "admin-1", Role: types.UserRoleAdmin},
	})

	ctx := context.Background()
	firstProject, err := service.Create(ctx, "admin-1", CreateProjectInput{Name: "Project One", Slug: "project-one"})
	if err != nil {
		t.Fatalf("Create() first project error = %v", err)
	}
	if _, err := service.Create(ctx, "admin-1", CreateProjectInput{Name: "Project Two", Slug: "project-two"}); err != nil {
		t.Fatalf("Create() second admin project error = %v", err)
	}

	for i := 0; i < MaxDatabasesPerProject+1; i++ {
		if _, err := service.ProvisionPostgres(ctx, "admin-1", firstProject.ID, ProvisionPostgresInput{}); err != nil {
			t.Fatalf("ProvisionPostgres() admin database %d error = %v", i+1, err)
		}
	}
}

func TestProvisionPostgresCreatesMainCredentialAndReturnsDatabaseURL(t *testing.T) {
	service := newTestService(t, map[string]types.User{
		"user-1": {ID: "user-1", Role: types.UserRoleUser},
	})

	ctx := context.Background()
	project, err := service.Create(ctx, "user-1", CreateProjectInput{Name: "Project One", Slug: "project-one"})
	if err != nil {
		t.Fatalf("Create() project error = %v", err)
	}

	updatedProject, err := service.ProvisionPostgres(ctx, "user-1", project.ID, ProvisionPostgresInput{Name: "Main DB"})
	if err != nil {
		t.Fatalf("ProvisionPostgres() error = %v", err)
	}
	if len(updatedProject.Databases) != 1 {
		t.Fatalf("len(updatedProject.Databases) = %d, want 1", len(updatedProject.Databases))
	}

	database := updatedProject.Databases[0]
	credential, err := service.store.GetActiveMainDatabaseCredential(ctx, database.ID)
	if err != nil {
		t.Fatalf("GetActiveMainDatabaseCredential() error = %v", err)
	}
	if credential.Username != database.PGRoleName {
		t.Fatalf("credential.Username = %q, want %q", credential.Username, database.PGRoleName)
	}

	view, err := service.DatabaseCredentials(ctx, "user-1", database.ID)
	if err != nil {
		t.Fatalf("DatabaseCredentials() error = %v", err)
	}
	if view.Username != database.PGRoleName {
		t.Fatalf("view.Username = %q, want %q", view.Username, database.PGRoleName)
	}
	if view.Database != database.PGDatabaseName {
		t.Fatalf("view.Database = %q, want %q", view.Database, database.PGDatabaseName)
	}
	wantPrefix := "postgres://"
	if len(view.DatabaseURL) <= len(wantPrefix) || view.DatabaseURL[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("view.DatabaseURL = %q, want postgres:// prefix", view.DatabaseURL)
	}
}

func TestLegacyDatabaseWithoutServerLinkReturnsManagedServerErrorOnDelete(t *testing.T) {
	service := newTestService(t, map[string]types.User{
		"user-1": {ID: "user-1", Role: types.UserRoleUser},
	})

	ctx := context.Background()
	project, err := service.Create(ctx, "user-1", CreateProjectInput{Name: "Project One", Slug: "project-one"})
	if err != nil {
		t.Fatalf("Create() project error = %v", err)
	}

	database := types.ProjectDatabase{
		ID:                  "db-legacy",
		ProjectID:           project.ID,
		ServerID:            nil,
		Engine:              "postgresql",
		Name:                "Legacy DB",
		Status:              string(types.ProjectStatusReady),
		PGDatabaseName:      "legacy_db",
		PGRoleName:          "legacy_role",
		PGPasswordEncrypted: "encrypted",
		PGHost:              "localhost",
		PGPort:              5432,
		PGSSLMode:           "disable",
		PositionX:           80,
		PositionY:           80,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
	if err := service.store.CreateProjectDatabase(ctx, database); err != nil {
		t.Fatalf("CreateProjectDatabase() error = %v", err)
	}

	_, err = service.RemoveProvisionedPostgres(ctx, "user-1", project.ID, database.ID)
	if err == nil || err.Error() != "This database is not linked to a managed PostgreSQL server." {
		t.Fatalf("RemoveProvisionedPostgres() error = %v, want managed server link error", err)
	}
}

func newTestService(t *testing.T, users map[string]types.User) *Service {
	t.Helper()

	dbConn := newTestDB(t)
	store := NewStore(dbConn)
	if err := store.EnsureSchema(context.Background()); err != nil {
		t.Fatalf("EnsureSchema() error = %v", err)
	}
	now := time.Now().UTC()
	if err := store.CreateDatabaseServer(context.Background(), types.DatabaseServer{
		ID:              "server-1",
		Name:            "Primary",
		Engine:          "postgres",
		Host:            "localhost",
		Port:            5432,
		AdminUsername:   "postgres",
		AdminPassword:   "secret",
		SSLMode:         "disable",
		DefaultDatabase: "postgres",
		IsDefault:       true,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("CreateDatabaseServer() error = %v", err)
	}
	userStore := user.NewStore(dbConn)
	if err := userStore.EnsureSchema(context.Background()); err != nil {
		t.Fatalf("user EnsureSchema() error = %v", err)
	}

	return New(
		store,
		&stubUsers{users: users},
		&stubPostgres{},
		crypto.New("test-master-key"),
	)
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbConn, err := sql.Open("sqlite", "file::memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = dbConn.Close() })

	migrationPath := filepath.Join("..", "..", "migrations", "001_init.sql")
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", migrationPath, err)
	}
	if _, err := dbConn.Exec(string(content)); err != nil {
		t.Fatalf("Exec(migration) error = %v", err)
	}

	return dbConn
}
