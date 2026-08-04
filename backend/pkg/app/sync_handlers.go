package app

import (
	"net/http"
	"strconv"
	"time"

	"chiro/pkg/auth"
	"chiro/pkg/store"
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
//
// Soporta paginación via ?limit=N. El pull se acota a `limit` filas por tabla
// y devuelve `next_since` con el maximo `updated_at` visto. El cliente puede
// hacer pull incremental hasta que no haya más cambios.
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
	limit := 500
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 5000 {
			limit = n
		}
	}
	pulled := map[string][]map[string]any{}
	maxSince := since
	for _, t := range store.Tablas {
		rows, err := a.Store.ListChangedLimit(r.Context(), t.Table, uid, since, limit)
		if err != nil {
			writeServerError(w, r, "error interno del servidor", err)
			return
		}
		if len(rows) > 0 {
			pulled[t.Table] = rows
			// Actualizar el cursor con el maximo updated_at de esta tabla.
			for _, row := range rows {
				if ts, ok := row["updated_at"].(int64); ok && ts > maxSince {
					maxSince = ts
				}
			}
		}
	}

	resp := map[string]any{
		"pulled":    pulled,
		"serverNow": time.Now().UnixMilli(),
	}
	if maxSince > since {
		resp["next_since"] = maxSince
	}
	writeJSON(w, http.StatusOK, resp)
}
