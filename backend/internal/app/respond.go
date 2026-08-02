package app

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// writeServerError loguea el error real (para debugging) y devuelve un
// mensaje genérico al cliente para no filtrar información interna.
func writeServerError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	log.Printf("chiro error [%s %s]: %v", r.Method, r.URL.Path, err)
	writeErr(w, http.StatusInternalServerError, msg)
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeErr(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return false
	}
	return true
}

// rowsToMaps convierte filas a []map[string]any con fechas en YYYY-MM-DD.
// Devuelve un slice vacío (no nil) cuando no hay filas, para JSON "[]".
func rowsToMaps(rows pgx.Rows) ([]map[string]any, error) {
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		fds := rows.FieldDescriptions()
		m := make(map[string]any, len(fds))
		for i, fd := range fds {
			if t, ok := vals[i].(time.Time); ok {
				m[fd.Name] = t.Format("2006-01-02")
				continue
			}
			m[fd.Name] = vals[i]
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func parseQueryBool(r *http.Request, key string) bool {
	return r.URL.Query().Get(key) == "true" || r.URL.Query().Get(key) == "1"
}
