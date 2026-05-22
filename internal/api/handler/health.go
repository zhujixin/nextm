package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nextm/nextm/internal/pkg/httputil"
)

type HealthHandler struct {
	sqliteDB *sql.DB
	postgres *sql.DB
	startAt  time.Time
}

type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	Timestamp int64  `json:"timestamp"`
	SQLite    string `json:"sqlite"`
	Postgres  string `json:"postgres,omitempty"`
}

func NewHealthHandler(sqliteDB *sql.DB, postgres *sql.DB) *HealthHandler {
	return &HealthHandler{
		sqliteDB: sqliteDB,
		postgres: postgres,
		startAt:  time.Now(),
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	sqliteStatus := "ok"
	if err := h.sqliteDB.Ping(); err != nil {
		sqliteStatus = "error: " + err.Error()
	}

	resp := HealthResponse{
		Status:    "ok",
		Version:   "0.1.0",
		Uptime:    time.Since(h.startAt).Round(time.Second).String(),
		Timestamp: time.Now().UnixMilli(),
		SQLite:    sqliteStatus,
	}

	if h.postgres != nil {
		if err := h.postgres.Ping(); err != nil {
			resp.Postgres = "error: " + err.Error()
		} else {
			resp.Postgres = "ok"
		}
	}

	if sqliteStatus != "ok" {
		resp.Status = "degraded"
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if resp.Status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.sqliteDB.Ping(); err != nil {
		httputil.WriteError(w, httputil.ErrUnavailable)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive"}`))
}
