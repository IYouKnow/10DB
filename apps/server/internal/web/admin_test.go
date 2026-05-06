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
	if body.TotalUsers != 2 {
		t.Fatalf("TotalUsers = %d, want 2", body.TotalUsers)
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

	userService := user.New(userStore)
	authService := auth.New("test-master-key", "http://localhost:8080", time.Hour)
	adminService := admin.New(userService, projectStore, admin.PostgresTester{})
	projectService := project.New(projectStore, userService, &postgres.Service{}, crypto.New("test-master-key"))

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
