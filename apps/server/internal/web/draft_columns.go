package web

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pedro/10db-launch/apps/server/internal/platform/auth"
	"github.com/pedro/10db-launch/apps/server/internal/project"
)

func (h *Handler) listDraftTableColumns(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	columns, err := h.projects.ListDraftColumns(r.Context(), session.UserID, chi.URLParam(r, "tableID"))
	if err != nil {
		h.writeDraftColumnError(w, "list_columns_failed", err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"columns": columns})
}

func (h *Handler) createDraftTableColumn(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	var body project.DraftColumnInput
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	column, err := h.projects.CreateDraftColumn(r.Context(), session.UserID, chi.URLParam(r, "tableID"), body)
	if err != nil {
		h.writeDraftColumnError(w, "create_column_failed", err)
		return
	}
	JSON(w, http.StatusCreated, column)
}

func (h *Handler) updateDraftTableColumn(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	var body project.DraftColumnInput
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	column, err := h.projects.UpdateDraftColumn(r.Context(), session.UserID, chi.URLParam(r, "tableID"), chi.URLParam(r, "columnID"), body)
	if err != nil {
		h.writeDraftColumnError(w, "update_column_failed", err)
		return
	}
	JSON(w, http.StatusOK, column)
}

func (h *Handler) deleteDraftTableColumn(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}
	if err := h.projects.DeleteDraftColumn(r.Context(), session.UserID, chi.URLParam(r, "tableID"), chi.URLParam(r, "columnID")); err != nil {
		h.writeDraftColumnError(w, "delete_column_failed", err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) writeDraftColumnError(w http.ResponseWriter, code string, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, sql.ErrNoRows) {
		status = http.StatusNotFound
	}
	Error(w, status, code, err.Error(), nil)
}
