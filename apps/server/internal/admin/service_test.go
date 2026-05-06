package admin

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/pedro/10db-launch/apps/server/internal/platform/postgres"
	"github.com/pedro/10db-launch/apps/server/internal/project"
	types "github.com/pedro/10db-launch/apps/server/internal/types"
	"github.com/pedro/10db-launch/apps/server/internal/user"
)

type stubUserCounter struct {
	count int
}

func (s stubUserCounter) Count(context.Context) (int, error) {
	return s.count, nil
}

type stubTester struct {
	err error
}

func (s stubTester) TestAdminConnection(context.Context, postgres.AdminConfig) error {
	return s.err
}

func TestFirstCreatedServerBecomesDefault(t *testing.T) {
	service, store := newTestService(t, stubTester{})

	server, err := service.CreateServer(context.Background(), CreateServerInput{
		Name:            "Primary",
		Host:            "127.0.0.1",
		Port:            5432,
		AdminUsername:   "postgres",
		AdminPassword:   "secret",
		SSLMode:         "disable",
		DefaultDatabase: "postgres",
	})
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if !server.IsDefault {
		t.Fatalf("CreateServer() IsDefault = false, want true")
	}

	servers, err := store.ListDatabaseServers(context.Background())
	if err != nil {
		t.Fatalf("ListDatabaseServers() error = %v", err)
	}
	if len(servers) != 1 || !servers[0].IsDefault {
		t.Fatalf("ListDatabaseServers() default = %#v, want single default server", servers)
	}
}

func TestOnlyOneServerCanBeDefault(t *testing.T) {
	service, store := newTestService(t, stubTester{})
	ctx := context.Background()

	first, err := service.CreateServer(ctx, CreateServerInput{
		Name:            "Primary",
		Host:            "127.0.0.1",
		Port:            5432,
		AdminUsername:   "postgres",
		AdminPassword:   "secret",
		SSLMode:         "disable",
		DefaultDatabase: "postgres",
	})
	if err != nil {
		t.Fatalf("CreateServer(first) error = %v", err)
	}

	second, err := service.CreateServer(ctx, CreateServerInput{
		Name:            "Secondary",
		Host:            "127.0.0.2",
		Port:            5432,
		AdminUsername:   "postgres",
		AdminPassword:   "secret",
		SSLMode:         "disable",
		DefaultDatabase: "postgres",
		IsDefault:       true,
	})
	if err != nil {
		t.Fatalf("CreateServer(second) error = %v", err)
	}
	if !second.IsDefault {
		t.Fatalf("CreateServer(second) IsDefault = false, want true")
	}

	servers, err := store.ListDatabaseServers(ctx)
	if err != nil {
		t.Fatalf("ListDatabaseServers() error = %v", err)
	}
	defaultCount := 0
	firstDefault := false
	secondDefault := false
	for _, server := range servers {
		if server.IsDefault {
			defaultCount++
		}
		if server.ID == first.ID {
			firstDefault = server.IsDefault
		}
		if server.ID == second.ID {
			secondDefault = server.IsDefault
		}
	}
	if defaultCount != 1 || firstDefault || !secondDefault {
		t.Fatalf("defaults after second create = first:%v second:%v count:%d, want only second default", firstDefault, secondDefault, defaultCount)
	}
}

func TestServerTestEndpointHandlesConnectionFailureCleanly(t *testing.T) {
	service, _ := newTestService(t, stubTester{err: errors.New("dial error")})
	ctx := context.Background()

	server, err := service.CreateServer(ctx, CreateServerInput{
		Name:            "Primary",
		Host:            "127.0.0.1",
		Port:            5432,
		AdminUsername:   "postgres",
		AdminPassword:   "secret",
		SSLMode:         "disable",
		DefaultDatabase: "postgres",
	})
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}

	result, err := service.TestServer(ctx, server.ID)
	if err != nil {
		t.Fatalf("TestServer() error = %v", err)
	}
	if result.Success {
		t.Fatalf("TestServer() Success = true, want false")
	}
	if result.Message == "" {
		t.Fatalf("TestServer() Message = empty, want clear failure message")
	}
}

func TestCannotDeleteDefaultServer(t *testing.T) {
	service, _ := newTestService(t, stubTester{})
	ctx := context.Background()

	server, err := service.CreateServer(ctx, CreateServerInput{
		Name:            "Primary",
		Host:            "127.0.0.1",
		Port:            5432,
		AdminUsername:   "postgres",
		AdminPassword:   "secret",
		SSLMode:         "disable",
		DefaultDatabase: "postgres",
	})
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}

	err = service.DeleteServer(ctx, server.ID)
	if err == nil || err.Error() != "Cannot delete the default server." {
		t.Fatalf("DeleteServer() error = %v, want default server error", err)
	}
}

func TestCannotDeleteServerWithDatabases(t *testing.T) {
	service, store := newTestService(t, stubTester{})
	ctx := context.Background()

	first, err := service.CreateServer(ctx, CreateServerInput{
		Name:            "Primary",
		Host:            "127.0.0.1",
		Port:            5432,
		AdminUsername:   "postgres",
		AdminPassword:   "secret",
		SSLMode:         "disable",
		DefaultDatabase: "postgres",
	})
	if err != nil {
		t.Fatalf("CreateServer(first) error = %v", err)
	}

	second, err := service.CreateServer(ctx, CreateServerInput{
		Name:            "Secondary",
		Host:            "127.0.0.2",
		Port:            5432,
		AdminUsername:   "postgres",
		AdminPassword:   "secret",
		SSLMode:         "disable",
		DefaultDatabase: "postgres",
	})
	if err != nil {
		t.Fatalf("CreateServer(second) error = %v", err)
	}

	projectRecord := types.Project{
		ID:          "project-1",
		OwnerUserID: "user-1",
		Name:        "Project",
		Slug:        "project",
		Description: "",
		Status:      types.ProjectStatusReady,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := store.CreateProject(ctx, projectRecord); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if err := store.CreateProjectDatabase(ctx, types.ProjectDatabase{
		ID:                  "db-1",
		ProjectID:           projectRecord.ID,
		ServerID:            &second.ID,
		Engine:              "postgresql",
		Name:                "App DB",
		Status:              string(types.ProjectStatusReady),
		PGDatabaseName:      "app_db",
		PGRoleName:          "app_role",
		PGPasswordEncrypted: "encrypted",
		PGHost:              second.Host,
		PGPort:              second.Port,
		PGSSLMode:           second.SSLMode,
		PositionX:           80,
		PositionY:           80,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}); err != nil {
		t.Fatalf("CreateProjectDatabase() error = %v", err)
	}

	err = service.DeleteServer(ctx, second.ID)
	if err == nil || err.Error() != "Cannot delete server while databases are using it." {
		t.Fatalf("DeleteServer() error = %v, want server in use error", err)
	}

	if first.ID == second.ID {
		t.Fatalf("server IDs unexpectedly match")
	}
}

func newTestService(t *testing.T, tester postgresTester) (*Service, *project.Store) {
	t.Helper()

	dbConn := newTestDB(t)
	projectStore := project.NewStore(dbConn)
	if err := projectStore.EnsureSchema(context.Background()); err != nil {
		t.Fatalf("EnsureSchema(project) error = %v", err)
	}
	userStore := user.NewStore(dbConn)
	if err := userStore.EnsureSchema(context.Background()); err != nil {
		t.Fatalf("EnsureSchema(user) error = %v", err)
	}

	return New(stubUserCounter{count: 1}, projectStore, tester), projectStore
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
