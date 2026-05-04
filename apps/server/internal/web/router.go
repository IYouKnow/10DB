package web

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/pedro/10db-launch/apps/server/internal/platform/auth"
	"github.com/pedro/10db-launch/apps/server/internal/project"
	types "github.com/pedro/10db-launch/apps/server/internal/types"
	"github.com/pedro/10db-launch/apps/server/internal/user"
)

type Handler struct {
	auth     *auth.Service
	users    *user.Service
	projects *project.Service
	origins  []string
}

func New(authService *auth.Service, userService *user.Service, projectService *project.Service, allowedOrigins []string) *Handler {
	return &Handler{auth: authService, users: userService, projects: projectService, origins: allowedOrigins}
}

func (h *Handler) Router(staticDir string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/login", h.login)
		api.Post("/auth/register", h.register)
		api.With(h.requireAuth).Post("/auth/logout", h.logout)
		api.With(h.requireAuth).Get("/auth/me", h.me)

		api.Group(func(secure chi.Router) {
			secure.Use(h.requireAuth)
			secure.Get("/projects", h.listProjects)
			secure.Post("/projects", h.createProject)
			secure.Get("/projects/{projectID}", h.getProject)
			secure.Post("/projects/{projectID}/databases/postgres", h.provisionPostgres)
			secure.Patch("/projects/{projectID}/databases/{databaseID}", h.updateProjectDatabase)
			secure.Get("/projects/{projectID}/databases/{databaseID}/schema", h.getSchema)
			secure.Put("/projects/{projectID}/databases/{databaseID}/schema", h.putSchema)
			secure.Post("/projects/{projectID}/databases/{databaseID}/schema/validate", h.validateSchema)
			secure.Post("/projects/{projectID}/databases/{databaseID}/schema/sql-preview", h.sqlPreview)
			secure.Post("/projects/{projectID}/databases/{databaseID}/schema/apply", h.applySchema)
			secure.Post("/projects/{projectID}/databases/{databaseID}/schema/tables/{tableID}/apply", h.applySchemaTable)
			secure.Delete("/projects/{projectID}/databases/{databaseID}/schema/tables/{tableID}", h.deleteSchemaTable)
			secure.Get("/projects/{projectID}/databases/{databaseID}/schema/revisions", h.schemaRevisions)
			secure.Get("/projects/{projectID}/databases/{databaseID}/tables", h.listTables)
			secure.Delete("/projects/{projectID}/databases/{databaseID}", h.removeProvisionedPostgres)
			secure.Delete("/projects/{projectID}", h.deleteProject)
			secure.Post("/projects/{projectID}/reset", h.resetProject)
			secure.Get("/projects/{projectID}/connection", h.projectConnection)
			secure.Get("/projects/{projectID}/schema", h.getSchema)
			secure.Put("/projects/{projectID}/schema", h.putSchema)
			secure.Post("/projects/{projectID}/schema/validate", h.validateSchema)
			secure.Post("/projects/{projectID}/schema/sql-preview", h.sqlPreview)
			secure.Post("/projects/{projectID}/schema/apply", h.applySchema)
			secure.Get("/projects/{projectID}/schema/revisions", h.schemaRevisions)
			secure.Get("/projects/{projectID}/tables", h.listTables)
			secure.Get("/projects/{projectID}/tables/{tableName}/columns", h.listColumns)
			secure.Get("/projects/{projectID}/tables/{tableName}/rows", h.listRows)
		})
	})

	if _, err := os.Stat(path.Join(staticDir, "index.html")); err == nil {
		fileServer := http.FileServer(http.FS(spaFS{os.DirFS(staticDir)}))
		r.Handle("/*", fileServer)
	} else {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>10DB Launch Dev</title>
    <style>
      body { font-family: Segoe UI, Arial, sans-serif; margin: 0; background: #eef5ff; color: #08111f; }
      .wrap { max-width: 760px; margin: 8vh auto; padding: 24px; }
      .card { background: white; border-radius: 24px; padding: 24px; box-shadow: 0 18px 48px rgba(8,17,31,.12); }
      code { background: #0f1b2d; color: white; padding: 2px 6px; border-radius: 8px; }
    </style>
  </head>
  <body>
    <div class="wrap">
      <div class="card">
        <h1>10DB Launch backend is running</h1>
        <p>No built frontend assets were found in <code>apps/server/static</code>.</p>
        <p>For local development, open <code>http://localhost:5173</code>.</p>
        <p>For a production-style local run, build the frontend first with <code>.\scripts\build.ps1</code>.</p>
      </div>
    </div>
  </body>
</html>`)
		})
	}
	return r
}

func (h *Handler) requireAuth(next http.Handler) http.Handler {
	return auth.Require(h.auth, auth.EnforceSameOrigin(h.origins, next))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	user, err := h.users.Authenticate(r.Context(), body.Email, body.Password)
	if err != nil {
		Error(w, http.StatusUnauthorized, "invalid_credentials", err.Error(), nil)
		return
	}
	cookie, err := h.auth.CreateCookie(h.auth.NewSession(user.ID, user.Email, user.Name))
	if err != nil {
		Error(w, http.StatusInternalServerError, "session_error", err.Error(), nil)
		return
	}
	http.SetCookie(w, cookie)
	JSON(w, http.StatusOK, map[string]any{"user": map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	}})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	user, err := h.users.Register(r.Context(), body.Name, body.Email, body.Password)
	if err != nil {
		Error(w, http.StatusBadRequest, "register_failed", err.Error(), nil)
		return
	}
	cookie, err := h.auth.CreateCookie(h.auth.NewSession(user.ID, user.Email, user.Name))
	if err != nil {
		Error(w, http.StatusInternalServerError, "session_error", err.Error(), nil)
		return
	}
	http.SetCookie(w, cookie)
	JSON(w, http.StatusCreated, map[string]any{"user": map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	}})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, h.auth.ClearCookie())
	JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	session, err := h.auth.Verify(r)
	if err != nil {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	userRecord, err := h.users.GetByID(r.Context(), session.UserID)
	if err != nil {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{
		"id":    userRecord.ID,
		"email": userRecord.Email,
		"name":  userRecord.Name,
		"role":  userRecord.Role,
	})
}

func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	projects, err := h.projects.List(r.Context(), session.UserID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "list_projects_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"projects": projects})
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	var body project.CreateProjectInput
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	createdProject, err := h.projects.Create(r.Context(), session.UserID, body)
	if err != nil {
		status := http.StatusBadRequest
		var limitErr *project.LimitError
		if errors.As(err, &limitErr) {
			status = http.StatusForbidden
		}
		Error(w, status, "create_project_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusCreated, createdProject)
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	project, err := h.projects.Get(r.Context(), session.UserID, chi.URLParam(r, "projectID"))
	if err != nil {
		Error(w, http.StatusNotFound, "project_not_found", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, project)
}

func (h *Handler) provisionPostgres(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	var body project.ProvisionPostgresInput
	if r.ContentLength > 0 {
		if err := Decode(r, &body); err != nil {
			Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
			return
		}
	}
	updatedProject, err := h.projects.ProvisionPostgres(r.Context(), session.UserID, chi.URLParam(r, "projectID"), body)
	if err != nil {
		status := http.StatusBadRequest
		var limitErr *project.LimitError
		if errors.As(err, &limitErr) {
			status = http.StatusForbidden
		}
		Error(w, status, "provision_postgres_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, updatedProject)
}

func (h *Handler) updateProjectDatabase(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	var body project.UpdateProjectDatabaseInput
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	project, err := h.projects.UpdateProjectDatabase(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"), body)
	if err != nil {
		Error(w, http.StatusBadRequest, "update_database_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, project)
}

func (h *Handler) removeProvisionedPostgres(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	project, err := h.projects.RemoveProvisionedPostgres(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "remove_postgres_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, project)
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	if err := h.projects.Delete(r.Context(), session.UserID, chi.URLParam(r, "projectID")); err != nil {
		Error(w, http.StatusBadRequest, "delete_project_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) resetProject(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	if err := h.projects.Reset(r.Context(), session.UserID, chi.URLParam(r, "projectID")); err != nil {
		Error(w, http.StatusBadRequest, "reset_project_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) projectConnection(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	connection, err := h.projects.Connection(r.Context(), session.UserID, chi.URLParam(r, "projectID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "connection_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, connection)
}

func (h *Handler) getSchema(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	databaseID, err := h.resolveDatabaseID(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"))
	if err != nil {
		JSON(w, http.StatusOK, types.SchemaBlueprint{Version: 1, ProjectID: chi.URLParam(r, "projectID"), DatabaseID: "", Tables: []types.TableBlueprint{}})
		return
	}
	revision, err := h.projects.LatestSchema(r.Context(), session.UserID, chi.URLParam(r, "projectID"), databaseID)
	if err != nil {
		JSON(w, http.StatusOK, types.SchemaBlueprint{Version: 1, ProjectID: chi.URLParam(r, "projectID"), DatabaseID: databaseID, Tables: []types.TableBlueprint{}})
		return
	}
	JSON(w, http.StatusOK, revision)
}

func (h *Handler) putSchema(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	var blueprint types.SchemaBlueprint
	if err := Decode(r, &blueprint); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	databaseID, err := h.resolveDatabaseID(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "database_not_found", err.Error(), nil)
		return
	}
	revision, validationErrors, err := h.projects.SaveSchema(r.Context(), session.UserID, chi.URLParam(r, "projectID"), databaseID, blueprint)
	if err != nil {
		Error(w, http.StatusBadRequest, "save_schema_failed", err.Error(), nil)
		return
	}
	if len(validationErrors) > 0 {
		Error(w, http.StatusBadRequest, "invalid_schema", "schema validation failed", validationErrors)
		return
	}
	JSON(w, http.StatusOK, revision)
}

func (h *Handler) validateSchema(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	var blueprint types.SchemaBlueprint
	if err := Decode(r, &blueprint); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	if _, err := h.projects.Get(r.Context(), session.UserID, chi.URLParam(r, "projectID")); err != nil {
		Error(w, http.StatusNotFound, "project_not_found", err.Error(), nil)
		return
	}
	databaseID, err := h.resolveDatabaseID(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "database_not_found", err.Error(), nil)
		return
	}
	normalized, errs := h.projects.ValidateBlueprint(r.Context(), chi.URLParam(r, "projectID"), databaseID, blueprint)
	JSON(w, http.StatusOK, map[string]any{"blueprint": normalized, "errors": errs, "valid": len(errs) == 0})
}

func (h *Handler) sqlPreview(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	var blueprint *types.SchemaBlueprint
	if r.ContentLength > 0 {
		var body types.SchemaBlueprint
		if err := Decode(r, &body); err == nil {
			blueprint = &body
		}
	}
	databaseID, err := h.resolveDatabaseID(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "database_not_found", err.Error(), nil)
		return
	}
	sql, errs, err := h.projects.PreviewSQL(r.Context(), session.UserID, chi.URLParam(r, "projectID"), databaseID, blueprint)
	if err != nil {
		Error(w, http.StatusBadRequest, "sql_preview_failed", err.Error(), nil)
		return
	}
	if len(errs) > 0 {
		Error(w, http.StatusBadRequest, "invalid_schema", "schema validation failed", errs)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"sql": sql})
}

func (h *Handler) applySchema(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	databaseID, err := h.resolveDatabaseID(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "database_not_found", err.Error(), nil)
		return
	}
	run, err := h.projects.Apply(r.Context(), session.UserID, chi.URLParam(r, "projectID"), databaseID)
	if err != nil {
		Error(w, http.StatusBadRequest, "apply_failed", err.Error(), run)
		return
	}
	JSON(w, http.StatusOK, run)
}

func (h *Handler) applySchemaTable(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	if err := h.projects.ApplyTable(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"), chi.URLParam(r, "tableID")); err != nil {
		Error(w, http.StatusBadRequest, "apply_table_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) deleteSchemaTable(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	if err := h.projects.DeleteTable(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"), chi.URLParam(r, "tableID")); err != nil {
		Error(w, http.StatusBadRequest, "delete_table_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) schemaRevisions(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	databaseID, err := h.resolveDatabaseID(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "database_not_found", err.Error(), nil)
		return
	}
	revisions, err := h.projects.Revisions(r.Context(), session.UserID, chi.URLParam(r, "projectID"), databaseID)
	if err != nil {
		Error(w, http.StatusBadRequest, "list_revisions_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"revisions": revisions})
}

func (h *Handler) listTables(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	databaseID, err := h.resolveDatabaseID(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "databaseID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "database_not_found", err.Error(), nil)
		return
	}
	tables, err := h.projects.ListDatabaseTables(r.Context(), session.UserID, chi.URLParam(r, "projectID"), databaseID)
	if err != nil {
		Error(w, http.StatusBadRequest, "list_tables_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"tables": tables})
}

func (h *Handler) resolveDatabaseID(ctx context.Context, userID, projectID, databaseID string) (string, error) {
	if databaseID != "" {
		return databaseID, nil
	}
	project, err := h.projects.Get(ctx, userID, projectID)
	if err != nil {
		return "", err
	}
	if len(project.Databases) == 0 {
		return "", fmt.Errorf("project does not have a provisioned database yet")
	}
	return project.Databases[0].ID, nil
}

func (h *Handler) listColumns(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	columns, err := h.projects.ListColumns(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "tableName"))
	if err != nil {
		Error(w, http.StatusBadRequest, "list_columns_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"columns": columns})
}

func (h *Handler) listRows(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	limit := 50
	offset := 0
	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	if value := r.URL.Query().Get("offset"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	rows, err := h.projects.ListRows(r.Context(), session.UserID, chi.URLParam(r, "projectID"), chi.URLParam(r, "tableName"), limit, offset)
	if err != nil {
		Error(w, http.StatusBadRequest, "list_rows_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, rows)
}

type spaFS struct {
	fs fs.FS
}

func (s spaFS) Open(name string) (fs.File, error) {
	clean := path.Clean(strings.TrimPrefix(name, "/"))
	if clean == "." || clean == "" {
		clean = "index.html"
	}
	file, err := s.fs.Open(clean)
	if err == nil {
		return file, nil
	}
	return s.fs.Open("index.html")
}

func WithReadyCheck(next http.Handler, checker func(context.Context) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := checker(r.Context()); err != nil {
			Error(w, http.StatusServiceUnavailable, "not_ready", err.Error(), nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}
