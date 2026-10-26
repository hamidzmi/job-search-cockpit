package httpadapter

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"job-search-cockpit/internal/usecase/port/inbound"
)

type BackupHandler struct {
	backup inbound.BackupUseCase
}

func NewBackupHandler(backup inbound.BackupUseCase) *BackupHandler {
	return &BackupHandler{backup: backup}
}

func (h *BackupHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/backup/export", h.handleExportBackup)
	mux.HandleFunc("POST /api/backup/export", h.handleExportBackup)
}

func (h *BackupHandler) handleExportBackup(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	timestamp := time.Now().UTC().Format("20060102-150405")

	if format == "sqlite" || format == "db" {
		data, err := h.backup.ExportSQLiteBackup(r.Context())
		if err != nil {
			writeDomainError(w, err)
			return
		}
		filename := fmt.Sprintf("cockpit-backup-%s.db", timestamp)
		w.Header().Set("Content-Type", "application/x-sqlite3")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		return
	}

	data, err := h.backup.ExportJSONBackup(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	filename := fmt.Sprintf("cockpit-backup-%s.json", timestamp)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
