package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/pedro/10db-launch/apps/server/internal/platform/auth"
	"github.com/pedro/10db-launch/apps/server/internal/project"
)

func (h *Handler) listDatabaseAPIKeys(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	keys, err := h.projects.ListAPIKeys(r.Context(), session.UserID, chi.URLParam(r, "databaseID"))
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		Error(w, status, "list_api_keys_failed", err.Error(), nil)
		return
	}

	JSON(w, http.StatusOK, map[string]any{"apiKeys": keys})
}

func (h *Handler) createDatabaseAPIKey(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	var body struct {
		Name       string `json:"name"`
		Permission string `json:"permission"`
	}
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}

	secret, err := h.projects.CreateAPIKey(r.Context(), session.UserID, chi.URLParam(r, "databaseID"), body.Name, body.Permission)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		Error(w, status, "create_api_key_failed", err.Error(), nil)
		return
	}

	JSON(w, http.StatusCreated, secret)
}

func (h *Handler) revokeDatabaseAPIKey(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	err := h.projects.RevokeAPIKey(r.Context(), session.UserID, chi.URLParam(r, "databaseID"), chi.URLParam(r, "keyID"))
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		Error(w, status, "revoke_api_key_failed", err.Error(), nil)
		return
	}

	JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) listDataRows(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	limit := 100
	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			limit = parsed
		}
	}

	rows, err := h.projects.ListDataByAPIKey(r.Context(), token, r.Method, chi.URLParam(r, "table"), limit)
	if err != nil {
		h.writeDataAPIError(w, err)
		return
	}
	JSON(w, http.StatusOK, rows)
}

func (h *Handler) getDataRow(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	row, err := h.projects.GetDataByAPIKey(r.Context(), token, r.Method, chi.URLParam(r, "table"), chi.URLParam(r, "id"))
	if err != nil {
		h.writeDataAPIError(w, err)
		return
	}
	JSON(w, http.StatusOK, row)
}

func (h *Handler) insertDataRow(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	body, err := decodeJSONObject(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	row, err := h.projects.InsertDataByAPIKey(r.Context(), token, r.Method, chi.URLParam(r, "table"), body)
	if err != nil {
		h.writeDataAPIError(w, err)
		return
	}
	JSON(w, http.StatusCreated, row)
}

func (h *Handler) updateDataRow(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	body, err := decodeJSONObject(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	row, err := h.projects.UpdateDataByAPIKey(r.Context(), token, r.Method, chi.URLParam(r, "table"), chi.URLParam(r, "id"), body)
	if err != nil {
		h.writeDataAPIError(w, err)
		return
	}
	JSON(w, http.StatusOK, row)
}

func (h *Handler) deleteDataRow(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	err := h.projects.DeleteDataByAPIKey(r.Context(), token, r.Method, chi.URLParam(r, "table"), chi.URLParam(r, "id"))
	if err != nil {
		h.writeDataAPIError(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *Handler) writeDataAPIError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, project.ErrInvalidAPIKey):
		Error(w, http.StatusUnauthorized, "invalid_api_key", "invalid api key", nil)
	case errors.Is(err, project.ErrAPIKeyPermission):
		Error(w, http.StatusForbidden, "forbidden", "API key does not have permission for this action.", nil)
	case errors.Is(err, pgx.ErrNoRows), errors.Is(err, sql.ErrNoRows):
		Error(w, http.StatusNotFound, "not_found", "resource not found", nil)
	default:
		status := http.StatusBadRequest
		Error(w, status, "data_api_failed", err.Error(), nil)
	}
}

func bearerToken(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

func decodeJSONObject(r *http.Request) (map[string]any, error) {
	defer r.Body.Close()
	if r.ContentLength == 0 {
		return map[string]any{}, nil
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body == nil {
		return map[string]any{}, nil
	}
	return body, nil
}
