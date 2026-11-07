package app

import (
	"net/http"
	"strconv"
	"time"

	"chiro/pkg/auth"
	"chiro/pkg/store"
	"chiro/pkg/svc"
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
// Requiere plan Pro activo y vigente (402 en caso contrario). El user_id de las
// filas enviadas siempre se sobrescribe con el del usuario autenticado.
//
// Soporta paginación via ?limit=N. El pull es inclusivo (updated_at >= since)
// y se acota a `limit` filas por tabla; `next_since` se calcula con
// svc.NextSince (mínimo entre las tablas truncadas) y `has_more` indica que el
// cliente debe repetir el pull con ese cursor. Puede haber filas repetidas:
// el merge LWW del cliente es idempotente.
func (a *App) handleSync(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	ok, err := a.hasActivePro(r.Context(), uid)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	if !ok {
		writeErr(w, http.StatusPaymentRequired, "la sincronización en la nube requiere una suscripción Pro activa")
		return
	}
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
	var pages []svc.TablePage
	for _, t := range store.Tablas {
		rows, err := a.Store.ListChangedLimit(r.Context(), t.Table, uid, since, limit)
		if err != nil {
			writeServerError(w, r, "error interno del servidor", err)
			return
		}
		if len(rows) == 0 {
			continue
		}
		pulled[t.Table] = rows
		last, _ := rows[len(rows)-1]["updated_at"].(int64)
		pages = append(pages, svc.TablePage{Count: len(rows), Limit: limit, LastUpdatedAt: last})
	}

	nextSince, hasMore := svc.NextSince(since, pages)
	resp := map[string]any{
		"pulled":     pulled,
		"serverNow":  time.Now().UnixMilli(),
		"next_since": nextSince,
		"has_more":   hasMore,
	}
	writeJSON(w, http.StatusOK, resp)
}
