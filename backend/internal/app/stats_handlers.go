package app

import (
	"net/http"
	"strconv"
	"time"

	"chiro/internal/auth"
	"chiro/internal/svc"
)

// handleStats devuelve el bundle para la pantalla de estadísticas:
// resumen anual por mes, desglose por categoría, cuentas con balance,
// capital prestado y total pendiente.
func (a *App) handleStats(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	q := r.URL.Query()
	year, _ := strconv.Atoi(q.Get("year"))
	month, _ := strconv.Atoi(q.Get("month"))
	if year == 0 {
		year = time.Now().UTC().Year()
	}
	if month == 0 {
		month = int(time.Now().UTC().Month())
	}

	yearRows, err := svc.YearSummary(r.Context(), a.Store, uid, year)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	breakdown, err := svc.CategoryBreakdown(r.Context(), a.Store, uid, year, month)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	balances, err := svc.AccountBalances(r.Context(), a.Store, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	lentByMonth, err := svc.LentByMonth(r.Context(), a.Store, uid, year)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	lentInYear, err := svc.LentInYear(r.Context(), a.Store, uid, year)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	outstanding, err := svc.OutstandingTotal(r.Context(), a.Store, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"year":          year,
		"month":         month,
		"months":        yearRows,
		"breakdown":     breakdown,
		"accounts":      balances,
		"lent_by_month": lentByMonth,
		"lent_in_year":  lentInYear,
		"outstanding":   outstanding,
	})
}

// handleBudgetProgress devuelve presupuestos con progreso (?year=&month=).
func (a *App) handleBudgetProgress(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	q := r.URL.Query()
	year, _ := strconv.Atoi(q.Get("year"))
	month, _ := strconv.Atoi(q.Get("month"))
	if year == 0 {
		year = time.Now().UTC().Year()
	}
	if month == 0 {
		month = int(time.Now().UTC().Month())
	}
	rows, err := svc.BudgetProgress(r.Context(), a.Store, uid, year, month)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
