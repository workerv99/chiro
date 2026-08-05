package app

import (
	"encoding/json"
	"net/http"
	"time"

	"chiro/pkg/auth"
)

// handleDeleteAccount elimina la cuenta del usuario y todos sus datos (GDPR).
func (a *App) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	ts := time.Now().UnixMilli()

	// Soft-delete de todos los datos del usuario.
	tables := []string{
		"expense_tag", "expense", "installment", "payment", "loan",
		"budget", "piggy_bank", "bill", "tag", "category", "account", "person",
	}
	for _, table := range tables {
		_, _ = a.Store.Pool().Exec(r.Context(),
			`UPDATE `+table+` SET deleted=$2, updated_at=$3 WHERE user_id=$1`, uid, 1, ts)
	}

	// Eliminar suscripción y consentimiento.
	_, _ = a.Store.Pool().Exec(r.Context(), `DELETE FROM subscription WHERE user_id=$1`, uid)
	_, _ = a.Store.Pool().Exec(r.Context(), `DELETE FROM user_consent WHERE user_id=$1`, uid)

	// Eliminar usuario.
	_, err := a.Store.Pool().Exec(r.Context(), `DELETE FROM users WHERE user_id=$1`, uid)
	if err != nil {
		writeServerError(w, r, "error al eliminar cuenta", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleExportData exporta todos los datos del usuario en JSON (GDPR).
func (a *App) handleExportData(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	ctx := r.Context()

	export := map[string]any{}

	// Usuario.
	var email, name string
	a.Store.Pool().QueryRow(ctx, `SELECT email, name FROM users WHERE user_id=$1`, uid).Scan(&email, &name)
	export["user"] = map[string]string{"email": email, "name": name}

	// Cuentas.
	rows, _ := a.Store.Pool().Query(ctx, `SELECT account_id, name, currency, account_type FROM account WHERE user_id=$1 AND deleted=0`, uid)
	if rows != nil {
		var items []map[string]any
		for rows.Next() {
			var id, n, c, t string
			rows.Scan(&id, &n, &c, &t)
			items = append(items, map[string]any{"account_id": id, "name": n, "currency": c, "type": t})
		}
		rows.Close()
		export["accounts"] = items
	}

	// Categorías.
	rows, _ = a.Store.Pool().Query(ctx, `SELECT category_id, name, color, type FROM category WHERE user_id=$1 AND deleted=0`, uid)
	if rows != nil {
		var items []map[string]any
		for rows.Next() {
			var id, n, c, t string
			rows.Scan(&id, &n, &c, &t)
			items = append(items, map[string]any{"category_id": id, "name": n, "color": c, "type": t})
		}
		rows.Close()
		export["categories"] = items
	}

	// Gastos.
	rows, _ = a.Store.Pool().Query(ctx, `SELECT expense_id, description, amount, date, type FROM expense WHERE user_id=$1 AND deleted=0`, uid)
	if rows != nil {
		var items []map[string]any
		for rows.Next() {
			var id, desc, t string
			var amt float64
			var d time.Time
			rows.Scan(&id, &desc, &amt, &d, &t)
			items = append(items, map[string]any{"expense_id": id, "description": desc, "amount": amt, "date": d.Format("2006-01-02"), "type": t})
		}
		rows.Close()
		export["expenses"] = items
	}

	// Préstamos.
	rows, _ = a.Store.Pool().Query(ctx, `SELECT loan_id, description, amount, date FROM loan WHERE user_id=$1 AND deleted=0`, uid)
	if rows != nil {
		var items []map[string]any
		for rows.Next() {
			var id, desc string
			var amt float64
			var d time.Time
			rows.Scan(&id, &desc, &amt, &d)
			items = append(items, map[string]any{"loan_id": id, "description": desc, "amount": amt, "date": d.Format("2006-01-02")})
		}
		rows.Close()
		export["loans"] = items
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=chiro-export.json")
	json.NewEncoder(w).Encode(export)
}
