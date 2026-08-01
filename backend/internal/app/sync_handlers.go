package app

import (
	"net/http"
	"time"

	"chiro/internal/auth"
	"chiro/internal/store"
)

// syncRequest es el payload de sincronización: un mapa tabla → filas locales
// (incluye tombstones deleted=1) y un marca de agua (updated_at) para el pull.
type syncRequest struct {
	Tables map[string][]map[string]any `json:"tables"`
	Since  int64                       `json:"since"`
}

// handleSync aplica el merge LWW por registro y devuelve lo que cambió.
// Port de utils/sync.ts: "último que guarda gana" por fila, sin pérdida salvo
// edición concurrente del MISMO registro.
func (a *App) handleSync(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	var req syncRequest
	if !readJSON(w, r, &req) {
		return
	}

	// 1) Merge de lo que envía el cliente (su lado gana si es más nuevo).
	for table, rows := range req.Tables {
		if len(rows) == 0 {
			continue
		}
		if err := a.Store.MergeRows(r.Context(), table, rows, uid); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// 2) Pull de lo que cambió desde el watermark del cliente.
	since := req.Since
	if since <= 0 {
		since = 1
	}
	pulled := map[string][]map[string]any{}
	for _, t := range store.Tablas {
		rows, err := a.Store.ListChanged(r.Context(), t.Table, uid, since)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		if len(rows) > 0 {
			pulled[t.Table] = rows
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"pulled":    pulled,
		"serverNow": time.Now().UnixMilli(),
	})
}
