package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
	handler, authService, _ := newTestHandler(t)
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
	handler, authService, _ := newTestHandler(t)
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

type testPostgres struct {
	lastListLimit int
	lastTableName string
}

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

func (s *testPostgres) ListDataRows(_ context.Context, _ types.Project, _ string, tableName string, limit int) (types.TableRows, error) {
	s.lastTableName = tableName
	s.lastListLimit = limit
	return types.TableRows{
		Columns: []string{"id", "name"},
		Rows: []map[string]any{
			{"id": "row-1", "name": "Alice"},
		},
		Limit: limit,
	}, nil
}

func (s *testPostgres) GetDataRow(_ context.Context, _ types.Project, _ string, tableName, id string) (map[string]any, error) {
	s.lastTableName = tableName
	return map[string]any{"id": id, "name": "Alice"}, nil
}

func (s *testPostgres) InsertDataRow(_ context.Context, _ types.Project, _ string, tableName string, values map[string]any) (map[string]any, error) {
	s.lastTableName = tableName
	result := map[string]any{"id": "new-row"}
	for key, value := range values {
		result[key] = value
	}
	return result, nil
}

func (s *testPostgres) UpdateDataRow(_ context.Context, _ types.Project, _ string, tableName, id string, values map[string]any) (map[string]any, error) {
	s.lastTableName = tableName
	result := map[string]any{"id": id}
	for key, value := range values {
		result[key] = value
	}
	return result, nil
}

func (s *testPostgres) DeleteDataRow(_ context.Context, _ types.Project, _ string, tableName, id string) error {
	s.lastTableName = tableName
	return nil
}

func TestOwnerCanCreateAPIKeyAndListDoesNotReturnSecret(t *testing.T) {
	handler, authService, _ := newTestHandler(t)

	projectRecord, err := handler.projects.Create(context.Background(), "user-1", project.CreateProjectInput{
		Name: "Project One",
		Slug: "project-one",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updatedProject, err := handler.projects.ProvisionPostgres(context.Background(), "user-1", projectRecord.ID, project.ProvisionPostgresInput{Name: "Main DB"})
	if err != nil {
		t.Fatalf("ProvisionPostgres() error = %v", err)
	}
	databaseID := updatedProject.Databases[0].ID

	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/databases/"+databaseID+"/api-keys", strings.NewReader(`{"name":"My app key"}`))
	createRequest.Header.Set("Content-Type", "application/json")
	createRequest.Header.Set("Origin", "http://localhost:5173")
	addSessionCookie(t, createRequest, authService, "user-1")

	createRecorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createRecorder.Code, http.StatusCreated)
	}
	var created types.DatabaseAPIKeySecret
	if err := json.NewDecoder(createRecorder.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if !strings.HasPrefix(created.Key, "tdb_live_") {
		t.Fatalf("created.Key = %q, want tdb_live_ prefix", created.Key)
	}
	if created.KeyPrefix == "" {
		t.Fatalf("created.KeyPrefix is empty")
	}
	if created.Permission != "read_write" {
		t.Fatalf("created.Permission = %q, want %q", created.Permission, "read_write")
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/databases/"+databaseID+"/api-keys", nil)
	addSessionCookie(t, listRequest, authService, "user-1")

	listRecorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(listRecorder, listRequest)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRecorder.Code, http.StatusOK)
	}
	var listed struct {
		APIKeys []map[string]any `json:"apiKeys"`
	}
	if err := json.NewDecoder(listRecorder.Body).Decode(&listed); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listed.APIKeys) != 1 {
		t.Fatalf("len(listed.APIKeys) = %d, want 1", len(listed.APIKeys))
	}
	if _, ok := listed.APIKeys[0]["key"]; ok {
		t.Fatalf("list response unexpectedly included full key")
	}
	if _, ok := listed.APIKeys[0]["keyHash"]; ok {
		t.Fatalf("list response unexpectedly included key hash")
	}
	if got, ok := listed.APIKeys[0]["permission"].(string); !ok || got != "read_write" {
		t.Fatalf("listed permission = %v, want read_write", listed.APIKeys[0]["permission"])
	}
}

func TestNonOwnerCannotCreateAPIKey(t *testing.T) {
	handler, authService, _ := newTestHandler(t)

	projectRecord, err := handler.projects.Create(context.Background(), "user-1", project.CreateProjectInput{
		Name: "Project One",
		Slug: "project-one",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updatedProject, err := handler.projects.ProvisionPostgres(context.Background(), "user-1", projectRecord.ID, project.ProvisionPostgresInput{Name: "Main DB"})
	if err != nil {
		t.Fatalf("ProvisionPostgres() error = %v", err)
	}
	databaseID := updatedProject.Databases[0].ID

	request := httptest.NewRequest(http.MethodPost, "/api/v1/databases/"+databaseID+"/api-keys", strings.NewReader(`{"name":"Other key"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://localhost:5173")
	addSessionCookie(t, request, authService, "user-2")

	recorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestInvalidKeyReturnsUnauthorizedForDataRoutes(t *testing.T) {
	handler, _, _ := newTestHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/data/users", nil)
	request.Header.Set("Authorization", "Bearer tdb_live_invalid")

	recorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestRevokedKeyCannotAccessDataRoutes(t *testing.T) {
	handler, _, _ := newTestHandler(t)

	projectRecord, err := handler.projects.Create(context.Background(), "user-1", project.CreateProjectInput{
		Name: "Project One",
		Slug: "project-one",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updatedProject, err := handler.projects.ProvisionPostgres(context.Background(), "user-1", projectRecord.ID, project.ProvisionPostgresInput{Name: "Main DB"})
	if err != nil {
		t.Fatalf("ProvisionPostgres() error = %v", err)
	}
	databaseID := updatedProject.Databases[0].ID

	secret, err := handler.projects.CreateAPIKey(context.Background(), "user-1", databaseID, "My app key", "read_write")
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}
	if err := handler.projects.RevokeAPIKey(context.Background(), "user-1", databaseID, secret.ID); err != nil {
		t.Fatalf("RevokeAPIKey() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/data/users", nil)
	request.Header.Set("Authorization", "Bearer "+secret.Key)

	recorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestDataRouteValidatesTableName(t *testing.T) {
	handler, _, _ := newTestHandler(t)

	projectRecord, err := handler.projects.Create(context.Background(), "user-1", project.CreateProjectInput{
		Name: "Project One",
		Slug: "project-one",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updatedProject, err := handler.projects.ProvisionPostgres(context.Background(), "user-1", projectRecord.ID, project.ProvisionPostgresInput{Name: "Main DB"})
	if err != nil {
		t.Fatalf("ProvisionPostgres() error = %v", err)
	}
	secret, err := handler.projects.CreateAPIKey(context.Background(), "user-1", updatedProject.Databases[0].ID, "My app key", "read_write")
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/data/bad-table!", nil)
	request.Header.Set("Authorization", "Bearer "+secret.Key)

	recorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestDataRouteCapsLimitAt500(t *testing.T) {
	handler, _, pg := newTestHandler(t)

	projectRecord, err := handler.projects.Create(context.Background(), "user-1", project.CreateProjectInput{
		Name: "Project One",
		Slug: "project-one",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updatedProject, err := handler.projects.ProvisionPostgres(context.Background(), "user-1", projectRecord.ID, project.ProvisionPostgresInput{Name: "Main DB"})
	if err != nil {
		t.Fatalf("ProvisionPostgres() error = %v", err)
	}
	secret, err := handler.projects.CreateAPIKey(context.Background(), "user-1", updatedProject.Databases[0].ID, "My app key", "read_write")
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/data/users?limit=999", nil)
	request.Header.Set("Authorization", "Bearer "+secret.Key)

	recorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if pg.lastListLimit != 500 {
		t.Fatalf("lastListLimit = %d, want 500", pg.lastListLimit)
	}
}

func TestReadKeyCanGetButCannotWrite(t *testing.T) {
	handler, _, _ := newTestHandler(t)
	key := createAPIKeyForTest(t, handler, "read")

	getRequest := httptest.NewRequest(http.MethodGet, "/data/users", nil)
	getRequest.Header.Set("Authorization", "Bearer "+key.Key)
	getRecorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(getRecorder, getRequest)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", getRecorder.Code, http.StatusOK)
	}

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodPost, path: "/data/users", body: `{"name":"Alice"}`},
		{method: http.MethodPatch, path: "/data/users/row-1", body: `{"name":"Bob"}`},
		{method: http.MethodDelete, path: "/data/users/row-1", body: ``},
	} {
		request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		request.Header.Set("Authorization", "Bearer "+key.Key)
		if tc.body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.Router(t.TempDir()).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("%s status = %d, want %d", tc.method, recorder.Code, http.StatusForbidden)
		}
		var body ErrorResponse
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("decode %s error: %v", tc.method, err)
		}
		if body.Error != "API key does not have permission for this action." {
			t.Fatalf("%s error = %q", tc.method, body.Error)
		}
	}
}

func TestWriteKeyCannotGetButCanWrite(t *testing.T) {
	handler, _, _ := newTestHandler(t)
	key := createAPIKeyForTest(t, handler, "write")

	getRequest := httptest.NewRequest(http.MethodGet, "/data/users", nil)
	getRequest.Header.Set("Authorization", "Bearer "+key.Key)
	getRecorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(getRecorder, getRequest)
	if getRecorder.Code != http.StatusForbidden {
		t.Fatalf("GET status = %d, want %d", getRecorder.Code, http.StatusForbidden)
	}

	for _, tc := range []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{method: http.MethodPost, path: "/data/users", body: `{"name":"Alice"}`, want: http.StatusCreated},
		{method: http.MethodPatch, path: "/data/users/row-1", body: `{"name":"Bob"}`, want: http.StatusOK},
		{method: http.MethodDelete, path: "/data/users/row-1", body: ``, want: http.StatusOK},
	} {
		request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		request.Header.Set("Authorization", "Bearer "+key.Key)
		if tc.body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.Router(t.TempDir()).ServeHTTP(recorder, request)
		if recorder.Code != tc.want {
			t.Fatalf("%s status = %d, want %d", tc.method, recorder.Code, tc.want)
		}
	}
}

func TestReadWriteKeyCanUseAllAllowedMethods(t *testing.T) {
	handler, _, _ := newTestHandler(t)
	key := createAPIKeyForTest(t, handler, "read_write")

	for _, tc := range []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{method: http.MethodGet, path: "/data/users", body: ``, want: http.StatusOK},
		{method: http.MethodPost, path: "/data/users", body: `{"name":"Alice"}`, want: http.StatusCreated},
		{method: http.MethodPatch, path: "/data/users/row-1", body: `{"name":"Bob"}`, want: http.StatusOK},
		{method: http.MethodDelete, path: "/data/users/row-1", body: ``, want: http.StatusOK},
	} {
		request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		request.Header.Set("Authorization", "Bearer "+key.Key)
		if tc.body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.Router(t.TempDir()).ServeHTTP(recorder, request)
		if recorder.Code != tc.want {
			t.Fatalf("%s status = %d, want %d", tc.method, recorder.Code, tc.want)
		}
	}
}

func TestOwnerCanManageDraftTableColumns(t *testing.T) {
	handler, authService, _ := newTestHandler(t)
	tableID := createDraftTableForTest(t, handler, "user-1", "project-one-columns")

	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/tables/"+tableID+"/columns", strings.NewReader(`{"name":"email","type":"text","nullable":false,"primaryKey":false,"defaultValue":""}`))
	createRequest.Header.Set("Content-Type", "application/json")
	createRequest.Header.Set("Origin", "http://localhost:5173")
	addSessionCookie(t, createRequest, authService, "user-1")
	createRecorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createRecorder.Code, http.StatusCreated)
	}
	var created types.DraftTableColumn
	if err := json.NewDecoder(createRecorder.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.Name != "email" {
		t.Fatalf("created.Name = %q, want email", created.Name)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/tables/"+tableID+"/columns", nil)
	addSessionCookie(t, listRequest, authService, "user-1")
	listRecorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(listRecorder, listRequest)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRecorder.Code, http.StatusOK)
	}

	updateRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/tables/"+tableID+"/columns/"+created.ID, strings.NewReader(`{"name":"user_email","type":"text","nullable":true,"primaryKey":false,"defaultValue":"''"}`))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateRequest.Header.Set("Origin", "http://localhost:5173")
	addSessionCookie(t, updateRequest, authService, "user-1")
	updateRecorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(updateRecorder, updateRequest)
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d", updateRecorder.Code, http.StatusOK)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/tables/"+tableID+"/columns/"+created.ID, nil)
	deleteRequest.Header.Set("Origin", "http://localhost:5173")
	addSessionCookie(t, deleteRequest, authService, "user-1")
	deleteRecorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(deleteRecorder, deleteRequest)
	if deleteRecorder.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want %d", deleteRecorder.Code, http.StatusOK)
	}
}

func TestNonOwnerCannotManageDraftTableColumns(t *testing.T) {
	handler, authService, _ := newTestHandler(t)
	tableID := createDraftTableForTest(t, handler, "user-1", "project-one-owner-only")

	request := httptest.NewRequest(http.MethodGet, "/api/v1/tables/"+tableID+"/columns", nil)
	addSessionCookie(t, request, authService, "user-2")
	recorder := httptest.NewRecorder()
	handler.Router(t.TempDir()).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDraftTableColumnsRejectDuplicateNames(t *testing.T) {
	handler, authService, _ := newTestHandler(t)
	tableID := createDraftTableForTest(t, handler, "user-1", "project-one-duplicate-cols")

	for _, body := range []string{
		`{"name":"email","type":"text","nullable":false,"primaryKey":false,"defaultValue":""}`,
		`{"name":"email","type":"text","nullable":true,"primaryKey":false,"defaultValue":""}`,
	} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/tables/"+tableID+"/columns", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", "http://localhost:5173")
		addSessionCookie(t, request, authService, "user-1")
		recorder := httptest.NewRecorder()
		handler.Router(t.TempDir()).ServeHTTP(recorder, request)
		if strings.Contains(body, `"nullable":true`) {
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("duplicate status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}
		} else if recorder.Code != http.StatusCreated {
			t.Fatalf("first create status = %d, want %d", recorder.Code, http.StatusCreated)
		}
	}
}

func createAPIKeyForTest(t *testing.T, handler *Handler, permission string) types.DatabaseAPIKeySecret {
	t.Helper()

	projectRecord, err := handler.projects.Create(context.Background(), "user-1", project.CreateProjectInput{
		Name: "Project One",
		Slug: "project-one-" + strings.ReplaceAll(permission, "_", "-"),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updatedProject, err := handler.projects.ProvisionPostgres(context.Background(), "user-1", projectRecord.ID, project.ProvisionPostgresInput{Name: "Main DB"})
	if err != nil {
		t.Fatalf("ProvisionPostgres() error = %v", err)
	}
	secret, err := handler.projects.CreateAPIKey(context.Background(), "user-1", updatedProject.Databases[0].ID, "My app key", permission)
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}
	return secret
}

func createDraftTableForTest(t *testing.T, handler *Handler, ownerUserID, slug string) string {
	t.Helper()

	projectRecord, err := handler.projects.Create(context.Background(), ownerUserID, project.CreateProjectInput{
		Name: "Project One",
		Slug: slug,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	updatedProject, err := handler.projects.ProvisionPostgres(context.Background(), ownerUserID, projectRecord.ID, project.ProvisionPostgresInput{Name: "Main DB"})
	if err != nil {
		t.Fatalf("ProvisionPostgres() error = %v", err)
	}
	blueprint := types.SchemaBlueprint{
		Version:    1,
		ProjectID:  projectRecord.ID,
		DatabaseID: updatedProject.Databases[0].ID,
		Tables: []types.TableBlueprint{
			{
				ID:       "tbl_users_" + slug,
				Name:     "users",
				Position: types.Position{X: 120, Y: 120},
				Columns: []types.ColumnBlueprint{
					{
						ID:         "col_id_" + slug,
						Name:       "id",
						Type:       "uuid",
						PrimaryKey: true,
						Nullable:   false,
						Default:    &types.DefaultValue{Kind: "expression", Value: "gen_random_uuid()"},
						Config:     types.ColumnConfig{},
					},
				},
			},
		},
	}
	if _, _, err := handler.projects.SaveSchema(context.Background(), ownerUserID, projectRecord.ID, updatedProject.Databases[0].ID, blueprint); err != nil {
		t.Fatalf("SaveSchema() error = %v", err)
	}
	return blueprint.Tables[0].ID
}

func newTestHandler(t *testing.T) (*Handler, *auth.Service, *testPostgres) {
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
	postgresService := &testPostgres{}
	projectService := project.New(projectStore, userService, postgresService, crypto.New("test-master-key"))

	return New(authService, adminService, userService, projectService, []string{"http://localhost:5173"}), authService, postgresService
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
