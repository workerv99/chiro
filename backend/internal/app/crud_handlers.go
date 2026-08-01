package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"chiro/internal/auth"
	"chiro/internal/store"
	"chiro/internal/svc"
)

// resource define una tabla sincronizable con sus endpoints CRUD.
type resource struct {
	table   string // nombre de la tabla en Postgres (singular)
	plural  string // prefijo de ruta HTTP (plural)
	orderBy string
	// afterList permite enriquecer la lista (ej. person_name en préstamos).
	afterList func(rows []map[string]any) []map[string]any
}

func (a *App) mountResource(r chi.Router, res resource) {
	table := res.table

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		uid := auth.ContextUser(req.Context())
		rows, err := a.Store.List(req.Context(), table, uid, res.orderBy)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		if res.afterList != nil {
			rows = res.afterList(rows)
		}
		writeJSON(w, http.StatusOK, rows)
	})

	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var row map[string]any
		if !readJSON(w, req, &row) {
			return
		}
		a.upsertRow(w, req, table, row)
	})

	r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
		var row map[string]any
		if !readJSON(w, req, &row) {
			return
		}
		pk := primaryKey(table)
		row[pk] = chi.URLParam(req, "id")
		a.upsertRow(w, req, table, row)
	})

	r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
		uid := auth.ContextUser(req.Context())
		if err := a.Store.SoftDelete(req.Context(), table, chi.URLParam(req, "id"), uid); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
}

func (a *App) upsertRow(w http.ResponseWriter, r *http.Request, table string, row map[string]any) {
	uid := auth.ContextUser(r.Context())
	// Ids deterministas para cuotas; el resto los genera el servidor.
	if id := row[primaryKey(table)]; id == nil || id == "" {
		row[primaryKey(table)] = svc.GenID(idPrefix(table))
	}
	if err := a.Store.MergeRows(r.Context(), table, []map[string]any{row}, uid); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	row["user_id"] = uid
	writeJSON(w, http.StatusOK, store.Normalize(row, uid))
}

func primaryKey(table string) string {
	for _, t := range store.Tablas {
		if t.Table == table {
			return t.PK[0]
		}
	}
	return table + "_id"
}

func idPrefix(table string) string {
	prefixes := map[string]string{
		"account":     "acc",
		"category":    "cat",
		"expense":     "exp",
		"person":      "per",
		"loan":        "loan",
		"payment":     "pay",
		"installment": "ins",
		"budget":      "bgt",
		"piggy_bank":  "pig",
		"bill":        "bil",
		"tag":         "tag",
	}
	if p, ok := prefixes[table]; ok {
		return p
	}
	return table
}
