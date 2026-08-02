package app

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"chiro/internal/auth"
	"chiro/internal/store"
	"chiro/internal/svc"
)

// handleListExpenses lista transacciones con filtros ?year=&month=&q=
func (a *App) handleListExpenses(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	q := r.URL.Query()
	year, _ := strconv.Atoi(q.Get("year"))
	month, _ := strconv.Atoi(q.Get("month"))
	search := strings.TrimSpace(q.Get("q"))

	// Excluye el lado "income" de las transferencias (port de TRANSFER_FILTER).
	sql := `SELECT * FROM expense WHERE user_id=$1 AND deleted=0
	        AND (transfer_pair_id IS NULL OR type != 'income')`
	args := []any{uid}
	if year > 0 {
		sql += ` AND EXTRACT(YEAR FROM date)::int = $` + strconv.Itoa(len(args)+1)
		args = append(args, year)
		if month > 0 {
			sql += ` AND EXTRACT(MONTH FROM date)::int = $` + strconv.Itoa(len(args)+1)
			args = append(args, month)
		}
	}
	if search != "" {
		sql += ` AND (description ILIKE $` + strconv.Itoa(len(args)+1)
		args = append(args, "%"+search+"%")
		sql += ` OR notes ILIKE $` + strconv.Itoa(len(args)+1) + `)`
		args = append(args, "%"+search+"%")
	}
	sql += ` ORDER BY date DESC`

	rows, err := a.Store.Pool().Query(r.Context(), sql, args...)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	out, err := rowsToMaps(rows)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleCreateExpense crea una transacción y sus tags (port createExpense).
func (a *App) handleCreateExpense(w http.ResponseWriter, r *http.Request) {
	var row map[string]any
	if !readJSON(w, r, &row) {
		return
	}
	if row["expense_id"] == nil || row["expense_id"] == "" {
		row["expense_id"] = svc.GenID("exp")
	}
	a.saveExpenseWithTags(w, r, row)
}

// handleUpdateExpense actualiza una transacción y reemplaza sus tags.
func (a *App) handleUpdateExpense(w http.ResponseWriter, r *http.Request) {
	var row map[string]any
	if !readJSON(w, r, &row) {
		return
	}
	row["expense_id"] = chi.URLParam(r, "id")
	a.saveExpenseWithTags(w, r, row)
}

// saveExpenseWithTags guarda la transacción y sincroniza expense_tag.
func (a *App) saveExpenseWithTags(w http.ResponseWriter, r *http.Request, row map[string]any) {
	uid := auth.ContextUser(r.Context())
	ts := time.Now().UnixMilli()

	var tags []any
	if t, ok := row["tags"].([]any); ok {
		tags = t
	}
	delete(row, "tags")

	expID := row["expense_id"].(string)
	err := a.Store.ExecAll(r.Context(), func(tx pgx.Tx) error {
		if err := a.Store.MergeRowsTx(r.Context(), tx, "expense", []map[string]any{row}, uid); err != nil {
			return err
		}
		// Reemplaza los tags: marca borrados y reinserta la selección actual.
		if _, err := tx.Exec(r.Context(),
			`UPDATE expense_tag SET deleted=1, updated_at=$3 WHERE user_id=$1 AND expense_id=$2 AND deleted=0`,
			uid, expID, ts); err != nil {
			return err
		}
		for _, tag := range tags {
			tagID, ok := tag.(string)
			if !ok || tagID == "" {
				continue
			}
			if _, err := tx.Exec(r.Context(),
				`INSERT INTO expense_tag (user_id, expense_id, tag_id, updated_at, deleted)
				 VALUES ($1,$2,$3,$4,0)
				 ON CONFLICT (user_id, expense_id, tag_id) DO UPDATE SET deleted=0, updated_at=EXCLUDED.updated_at`,
				uid, expID, tagID, ts); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	out := store.Normalize(row, uid)
	out["tags"] = tags
	writeJSON(w, http.StatusOK, out)
}

type transferReq struct {
	Description          string  `json:"description"`
	Amount               float64 `json:"amount"`
	Date                 string  `json:"date"`
	AccountID            string  `json:"account_id"`
	DestinationAccountID string  `json:"destination_account_id"`
	Notes                *string `json:"notes"`
}

// handleCreateTransfer crea el par (retiro/gasto + ingreso) de una transferencia.
// Port de utils/repo/expenses.ts createTransfer (doble entrada, misma updated_at).
func (a *App) handleCreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req transferReq
	if !readJSON(w, r, &req) {
		return
	}
	if req.Description == "" || req.Amount <= 0 || req.AccountID == "" || req.DestinationAccountID == "" {
		writeErr(w, http.StatusBadRequest, "description, amount, account_id y destination_account_id requeridos")
		return
	}
	if req.AccountID == req.DestinationAccountID {
		writeErr(w, http.StatusBadRequest, "las cuentas deben ser distintas")
		return
	}
	uid := auth.ContextUser(r.Context())
	ts := time.Now().UnixMilli()
	pairID := svc.GenID("trp")

	err := a.Store.ExecAll(r.Context(), func(tx pgx.Tx) error {
		for _, side := range []struct {
			id       string
			from, to string
			typ      string
		}{
			{svc.GenID("exp"), req.AccountID, req.DestinationAccountID, "expense"},
			{svc.GenID("exp"), req.DestinationAccountID, req.AccountID, "income"},
		} {
			_, err := tx.Exec(r.Context(),
				`INSERT INTO expense
				  (user_id, expense_id, description, amount, date, account_id,
				   destination_account_id, transfer_pair_id, notes, type, updated_at, deleted)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,0)`,
				uid, side.id, req.Description, req.Amount, req.Date,
				side.from, side.to, pairID, req.Notes, side.typ, ts)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"transfer_pair_id": pairID})
}

// handleDeleteExpense borra la transacción y, si es transferencia, ambos lados
// y sus tags (port de utils/repo/expenses.ts deleteExpense).
func (a *App) handleDeleteExpense(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	id := chi.URLParam(r, "id")
	ts := time.Now().UnixMilli()

	var pair *string
	row := a.Store.Pool().QueryRow(r.Context(),
		`SELECT transfer_pair_id FROM expense WHERE user_id=$1 AND expense_id=$2`, uid, id)
	var p any
	if err := row.Scan(&p); err != nil {
		writeErr(w, http.StatusNotFound, "no encontrado")
		return
	}
	if p != nil {
		s := p.(string)
		pair = &s
	}

	err := a.Store.ExecAll(r.Context(), func(tx pgx.Tx) error {
		if pair != nil {
			if _, err := tx.Exec(r.Context(),
				`UPDATE expense_tag SET deleted=1, updated_at=$3
				 WHERE user_id=$1 AND expense_id IN (SELECT expense_id FROM expense WHERE user_id=$1 AND transfer_pair_id=$2)`,
				uid, *pair, ts); err != nil {
				return err
			}
			_, err := tx.Exec(r.Context(),
				`UPDATE expense SET deleted=1, updated_at=$3 WHERE user_id=$1 AND transfer_pair_id=$2`,
				uid, *pair, ts)
			return err
		}
		if _, err := tx.Exec(r.Context(),
			`UPDATE expense_tag SET deleted=1, updated_at=$3 WHERE user_id=$1 AND expense_id=$2`,
			uid, id, ts); err != nil {
			return err
		}
		_, err := tx.Exec(r.Context(),
			`UPDATE expense SET deleted=1, updated_at=$3 WHERE user_id=$1 AND expense_id=$2`,
			uid, id, ts)
		return err
	})
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleGetExpense devuelve una transacción con sus tags (port getExpense).
func (a *App) handleGetExpense(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	id := chi.URLParam(r, "id")

	rows, err := a.Store.Pool().Query(r.Context(),
		`SELECT * FROM expense WHERE user_id=$1 AND expense_id=$2 AND deleted=0`, uid, id)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	out, err := rowsToMaps(rows)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	if len(out) == 0 {
		writeErr(w, http.StatusNotFound, "no encontrado")
		return
	}
	row := out[0]

	tagRows, err := a.Store.Pool().Query(r.Context(),
		`SELECT t.tag_id FROM tag t JOIN expense_tag et ON et.user_id=t.user_id AND et.tag_id=t.tag_id
		 WHERE et.user_id=$1 AND et.expense_id=$2 AND et.deleted=0 AND t.deleted=0 ORDER BY t.name ASC`, uid, id)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	tags := make([]any, 0)
	for tagRows.Next() {
		var tagID string
		if err := tagRows.Scan(&tagID); err != nil {
			writeServerError(w, r, "error interno del servidor", err)
			return
		}
		tags = append(tags, tagID)
	}
	tagRows.Close()

	row["tags"] = tags
	writeJSON(w, http.StatusOK, row)
}

// handleMonthSummary devuelve el resumen del mes (?year=&month=).
func (a *App) handleMonthSummary(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if year == 0 {
		now := time.Now().UTC()
		year, month = now.Year(), int(now.Month())
	}
	if month == 0 {
		month = int(time.Now().UTC().Month())
	}
	sum, err := svc.MonthSummary(r.Context(), a.Store, uid, year, month)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, sum)
}
