package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pedro/10db-launch/apps/server/internal/admin"
)

func (h *Handler) adminOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.admin.Overview(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "admin_overview_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, overview)
}

func (h *Handler) adminListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.admin.ListServers(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "list_servers_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"servers": servers})
}

func (h *Handler) adminCreateServer(w http.ResponseWriter, r *http.Request) {
	var body admin.CreateServerInput
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	server, err := h.admin.CreateServer(r.Context(), body)
	if err != nil {
		Error(w, http.StatusBadRequest, "create_server_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusCreated, server)
}

func (h *Handler) adminUpdateServer(w http.ResponseWriter, r *http.Request) {
	var body admin.UpdateServerInput
	if err := Decode(r, &body); err != nil {
		Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	server, err := h.admin.UpdateServer(r.Context(), chi.URLParam(r, "serverID"), body)
	if err != nil {
		status := http.StatusBadRequest
		if admin.IsNotFound(err) {
			status = http.StatusNotFound
		}
		Error(w, status, "update_server_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, server)
}

func (h *Handler) adminTestServer(w http.ResponseWriter, r *http.Request) {
	result, err := h.admin.TestServer(r.Context(), chi.URLParam(r, "serverID"))
	if err != nil {
		status := http.StatusBadRequest
		if admin.IsNotFound(err) {
			status = http.StatusNotFound
		}
		Error(w, status, "test_server_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, result)
}

func (h *Handler) adminSetDefaultServer(w http.ResponseWriter, r *http.Request) {
	server, err := h.admin.SetDefaultServer(r.Context(), chi.URLParam(r, "serverID"))
	if err != nil {
		status := http.StatusBadRequest
		if admin.IsNotFound(err) {
			status = http.StatusNotFound
		}
		Error(w, status, "set_default_server_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, server)
}

func (h *Handler) adminDeleteServer(w http.ResponseWriter, r *http.Request) {
	err := h.admin.DeleteServer(r.Context(), chi.URLParam(r, "serverID"))
	if err != nil {
		status := http.StatusBadRequest
		if admin.IsNotFound(err) {
			status = http.StatusNotFound
		}
		Error(w, status, "delete_server_failed", err.Error(), nil)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"ok": true})
}
