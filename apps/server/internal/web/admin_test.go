package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/pedro/10db-launch/apps/server/internal/admin"
	"github.com/pedro/10db-launch/apps/server/internal/platform/auth"
	"github.com/pedro/10db-launch/apps/server/internal/platform/crypto"
	"github.com/pedro/10db-launch/apps/server/internal/platform/postgres"
	"github.com/pedro/10db-launch/apps/server/internal/project"
	types "github.com/pedro/10db-launch/apps/server/internal/types"
	"github.com/pedro/10db-launch/apps/server/internal/user"
)

func TestNonAdminCannotAccessAdminOverview(t *testing.T) {
	handler, authService := newTestHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/api/admin/overview", nil)
	addSessionCookie(t, request, authService, "user-1")

	recorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	var body ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error != "Admin access required." {
		t.Fatalf("error message = %q, want %q", body.Error, "Admin access required.")
	}
}

func TestAdminCanAccessAdminOverview(t *testing.T) {
	handler, authService := newTestHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/api/admin/overview", nil)
	addSessionCookie(t, request, authService, "admin-1")

	recorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var body types.AdminOverview
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if body.TotalUsers != 3 {
		t.Fatalf("TotalUsers = %d, want 3", body.TotalUsers)
	}
}

type testPostgres struct{}

func (s *testPostgres) CreateProjectDatabaseWithConfig(context.Context, postgres.AdminConfig, string, string, string) error {
	return nil
}

func (s *testPostgres) DropProjectDatabaseWithConfig(context.Context, postgres.AdminConfig, string, string) error {
	return nil
}

func (s *testPostgres) ResetProjectSchema(context.Context, types.Project, string) error {
	return nil
}

func (s *testPostgres) ApplySQL(context.Context, types.Project, string, string) error {
	return nil
}

func (s *testPostgres) DropTable(context.Context, types.Project, string, string) error {
	return nil
}

func (s *testPostgres) ListTables(context.Context, types.Project, string) ([]types.TableInfo, error) {
	return nil, nil
}

func (s *testPostgres) ListColumns(context.Context, types.Project, string, string) ([]types.ColumnInfo, error) {
	return nil, nil
}

func (s *testPostgres) ListRows(context.Context, types.Project, string, string, int, int) (types.TableRows, error) {
	return types.TableRows{}, nil
}

func TestDatabaseCredentialsAreLimitedToDatabaseOwner(t *testing.T) {
	handler, authService := newTestHandler(t)

	createdProject, err := handler.projects.Create(context.Background(), "user-1", project.CreateProjectInput{
		Name: "Project One",
		Slug: "project-one",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updatedProject, err := handler.projects.ProvisionPostgres(context.Background(), "user-1", createdProject.ID, project.ProvisionPostgresInput{Name: "Main DB"})
	if err != nil {
		t.Fatalf("ProvisionPostgres() error = %v", err)
	}
	databaseID := updatedProject.Databases[0].ID

	ownerRequest := httptest.NewRequest(http.MethodGet, "/api/v1/databases/"+databaseID+"/credentials", nil)
	addSessionCookie(t, ownerRequest, authService, "user-1")

	ownerRecorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(ownerRecorder, ownerRequest)

	if ownerRecorder.Code != http.StatusOK {
		t.Fatalf("owner status = %d, want %d", ownerRecorder.Code, http.StatusOK)
	}
	var ownerBody types.DatabaseCredentialsView
	if err := json.NewDecoder(ownerRecorder.Body).Decode(&ownerBody); err != nil {
		t.Fatalf("decode owner body: %v", err)
	}
	if ownerBody.Username == "" || ownerBody.Password == "" || ownerBody.DatabaseURL == "" {
		t.Fatalf("owner credentials missing fields: %+v", ownerBody)
	}

	otherRequest := httptest.NewRequest(http.MethodGet, "/api/v1/databases/"+databaseID+"/credentials", nil)
	addSessionCookie(t, otherRequest, authService, "user-2")

	otherRecorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(otherRecorder, otherRequest)

	if otherRecorder.Code != http.StatusNotFound {
		t.Fatalf("non-owner status = %d, want %d", otherRecorder.Code, http.StatusNotFound)
	}
}

func newTestHandler(t *testing.T) (*Handler, *auth.Service) {
	t.Helper()

	dbConn := newAdminTestDB(t)
	projectStore := project.NewStore(dbConn)
	if err := projectStore.EnsureSchema(context.Background()); err != nil {
		t.Fatalf("EnsureSchema(project) error = %v", err)
	}
	userStore := user.NewStore(dbConn)
	if err := userStore.EnsureSchema(context.Background()); err != nil {
		t.Fatalf("EnsureSchema(user) error = %v", err)
	}

	now := time.Now().UTC()
	if err := userStore.Create(context.Background(), types.User{
		ID:           "admin-1",
		Email:        "admin@example.com",
		Name:         "Admin",
		Role:         types.UserRoleAdmin,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatalf("Create(admin) error = %v", err)
	}
	if err := userStore.Create(context.Background(), types.User{
		ID:           "user-1",
		Email:        "user@example.com",
		Name:         "User",
		Role:         types.UserRoleUser,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatalf("Create(user) error = %v", err)
	}
	if err := userStore.Create(context.Background(), types.User{
		ID:           "user-2",
		Email:        "user2@example.com",
		Name:         "User Two",
		Role:         types.UserRoleUser,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatalf("Create(user-2) error = %v", err)
	}
	if err := projectStore.CreateDatabaseServer(context.Background(), types.DatabaseServer{
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

	userService := user.New(userStore)
	authService := auth.New("test-master-key", "http://localhost:8080", time.Hour)
	adminService := admin.New(userService, projectStore, admin.PostgresTester{})
	projectService := project.New(projectStore, userService, &testPostgres{}, crypto.New("test-master-key"))

	return New(authService, adminService, userService, projectService, []string{"http://localhost:5173"}), authService
}

func addSessionCookie(t *testing.T, request *http.Request, authService *auth.Service, userID string) {
	t.Helper()

	cookie, err := authService.CreateCookie(authService.NewSession(userID, userID+"@example.com", userID))
	if err != nil {
		t.Fatalf("CreateCookie() error = %v", err)
	}
	request.AddCookie(cookie)
}

func newAdminTestDB(t *testing.T) *sql.DB {
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
