package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"chiro/pkg/auth"
	"chiro/pkg/svc"
)

// handleDueBills devuelve facturas activas con next_date dentro de 7 días (port getDueBills).
func (a *App) handleDueBills(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	limit, _ := svc.AdvanceBillDate(svc.Today(), "weekly")
	rows, err := a.Store.Pool().Query(r.Context(),
		`SELECT * FROM bill WHERE user_id=$1 AND deleted=0 AND active=1 AND next_date <= $2 ORDER BY next_date ASC`,
		uid, limit)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	defer rows.Close()
	out, err := rowsToMaps(rows)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handlePayBill crea la transacción correspondiente y avanza la próxima fecha (port payBill).
func (a *App) handlePayBill(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	id := chi.URLParam(r, "id")
	ts := time.Now().UnixMilli()

	err := a.Store.ExecAll(r.Context(), func(tx pgx.Tx) error {
		var name string
		var amount float64
		var categoryID, accountID, notes, typ *string
		var nextDate time.Time
		var frequency string
		if err := tx.QueryRow(r.Context(),
			`SELECT name, amount, category_id, account_id, notes, type, frequency, next_date
			 FROM bill WHERE user_id=$1 AND bill_id=$2 AND deleted=0`, uid, id).
			Scan(&name, &amount, &categoryID, &accountID, &notes, &typ, &frequency, &nextDate); err != nil {
			return err
		}
		t := "expense"
		if typ != nil {
			t = *typ
		}
		expID := svc.GenID("exp")
		if _, err := tx.Exec(r.Context(),
			`INSERT INTO expense (user_id, expense_id, description, amount, date, category_id, account_id, notes, type, updated_at, deleted)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,0)`,
			uid, expID, name, amount, svc.Today(), categoryID, accountID, notes, t, ts); err != nil {
			return err
		}
		next, err := svc.AdvanceBillDate(nextDate.Format("2006-01-02"), frequency)
		if err != nil {
			return err
		}
		_, err = tx.Exec(r.Context(),
			`UPDATE bill SET next_date=$3, updated_at=$4 WHERE user_id=$1 AND bill_id=$2`, uid, id, next, ts)
		return err
	})
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleSkipBill solo avanza la fecha sin crear transacción (port skipBill).
func (a *App) handleSkipBill(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	id := chi.URLParam(r, "id")
	var req struct {
		NextDate string `json:"next_date"`
		Freq     string `json:"frequency"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if req.NextDate == "" || req.Freq == "" {
		writeErr(w, http.StatusBadRequest, "next_date y frequency requeridos")
		return
	}
	next, err := svc.AdvanceBillDate(req.NextDate, req.Freq)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "fecha inválida")
		return
	}
	_, err = a.Store.Pool().Exec(r.Context(),
		`UPDATE bill SET next_date=$3, updated_at=$4 WHERE user_id=$1 AND bill_id=$2 AND deleted=0`,
		uid, id, next, time.Now().UnixMilli())
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
