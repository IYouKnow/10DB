package web

import (
	"context"
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
)

type Handler struct {
	auth     *auth.Service
	projects *project.Service
	origins  []string
}

func New(authService *auth.Service, projectService *project.Service, allowedOrigins []string) *Handler {
	return &Handler{auth: authService, projects: projectService, origins: allowedOrigins}
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
		api.With(h.requireAuth).Post("/auth/logout", h.logout)
		api.With(h.requireAuth).Get("/auth/me", h.me)

		api.Group(func(secure chi.Router) {
			secure.Use(h.requireAuth)
			secure.Get("/projects", h.listProjects)
			secure.Post("/projects", h.createProject)
			secure.Get("/projects/{projectID}", h.getProject)
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
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	if !h.auth.Login(body.Username, body.Password) {
		Error(w, http.StatusUnauthorized, "invalid_credentials", "invalid credentials", nil)
		return
	}
	cookie, err := h.auth.CreateCookie()
	if err != nil {
		Error(w, http.StatusInternalServerError, "session_error", err.Error(), nil)
		return
	}
	http.SetCookie(w, cookie)
	JSON(w, http.StatusOK, map[string]any{"ok": true})
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
	JSON(w, http.StatusOK, map[string]any{"username": session.Username})
}

func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projects.List(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "list_projects_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"projects": projects})
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var body project.CreateProjectInput
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	project, err := h.projects.Create(r.Context(), body)
	if err != nil {
		Error(w, http.StatusBadRequest, "create_project_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusCreated, project)
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	project, err := h.projects.Get(r.Context(), chi.URLParam(r, "projectID"))
	if err != nil {
		Error(w, http.StatusNotFound, "project_not_found", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, project)
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	if err := h.projects.Delete(r.Context(), chi.URLParam(r, "projectID")); err != nil {
		Error(w, http.StatusBadRequest, "delete_project_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) resetProject(w http.ResponseWriter, r *http.Request) {
	if err := h.projects.Reset(r.Context(), chi.URLParam(r, "projectID")); err != nil {
		Error(w, http.StatusBadRequest, "reset_project_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) projectConnection(w http.ResponseWriter, r *http.Request) {
	connection, err := h.projects.Connection(r.Context(), chi.URLParam(r, "projectID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "connection_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, connection)
}

func (h *Handler) getSchema(w http.ResponseWriter, r *http.Request) {
	revision, err := h.projects.LatestSchema(r.Context(), chi.URLParam(r, "projectID"))
	if err != nil {
		JSON(w, http.StatusOK, types.SchemaBlueprint{Version: 1, ProjectID: chi.URLParam(r, "projectID"), Tables: []types.TableBlueprint{}})
		return
	}
	JSON(w, http.StatusOK, revision)
}

func (h *Handler) putSchema(w http.ResponseWriter, r *http.Request) {
	var blueprint types.SchemaBlueprint
	if err := Decode(r, &blueprint); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	revision, validationErrors, err := h.projects.SaveSchema(r.Context(), chi.URLParam(r, "projectID"), blueprint)
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
	var blueprint types.SchemaBlueprint
	if err := Decode(r, &blueprint); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	normalized, errs := h.projects.ValidateBlueprint(r.Context(), chi.URLParam(r, "projectID"), blueprint)
	JSON(w, http.StatusOK, map[string]any{"blueprint": normalized, "errors": errs, "valid": len(errs) == 0})
}

func (h *Handler) sqlPreview(w http.ResponseWriter, r *http.Request) {
	var blueprint *types.SchemaBlueprint
	if r.ContentLength > 0 {
		var body types.SchemaBlueprint
		if err := Decode(r, &body); err == nil {
			blueprint = &body
		}
	}
	sql, errs, err := h.projects.PreviewSQL(r.Context(), chi.URLParam(r, "projectID"), blueprint)
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
	run, err := h.projects.Apply(r.Context(), chi.URLParam(r, "projectID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "apply_failed", err.Error(), run)
		return
	}
	JSON(w, http.StatusOK, run)
}

func (h *Handler) schemaRevisions(w http.ResponseWriter, r *http.Request) {
	revisions, err := h.projects.Revisions(r.Context(), chi.URLParam(r, "projectID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "list_revisions_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"revisions": revisions})
}

func (h *Handler) listTables(w http.ResponseWriter, r *http.Request) {
	tables, err := h.projects.ListTables(r.Context(), chi.URLParam(r, "projectID"))
	if err != nil {
		Error(w, http.StatusBadRequest, "list_tables_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"tables": tables})
}

func (h *Handler) listColumns(w http.ResponseWriter, r *http.Request) {
	columns, err := h.projects.ListColumns(r.Context(), chi.URLParam(r, "projectID"), chi.URLParam(r, "tableName"))
	if err != nil {
		Error(w, http.StatusBadRequest, "list_columns_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"columns": columns})
}

func (h *Handler) listRows(w http.ResponseWriter, r *http.Request) {
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
	rows, err := h.projects.ListRows(r.Context(), chi.URLParam(r, "projectID"), chi.URLParam(r, "tableName"), limit, offset)
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
