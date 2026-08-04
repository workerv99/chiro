package app

import (
	"net/http"
	"time"

	"chiro/pkg/auth"
	"chiro/pkg/config"
)

type subscriptionResponse struct {
	Plan      string  `json:"plan"`
	Status    string  `json:"status"`
	Limits    config.PlanLimits `json:"limits"`
	Usage     usageInfo `json:"usage"`
}

type usageInfo struct {
	ExpensesThisMonth int `json:"expenses_this_month"`
	TotalAccounts     int `json:"total_accounts"`
	TotalLoans        int `json:"total_loans"`
}

// handleGetSubscription devuelve el estado actual de la suscripción del usuario.
func (a *App) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())

	var plan, status string
	err := a.Store.Pool().QueryRow(r.Context(),
		`SELECT COALESCE(plan, 'free'), COALESCE(status, 'active') FROM subscription WHERE user_id=$1`,
		uid).Scan(&plan, &status)
	if err != nil {
		plan = "free"
		status = "active"
	}

	limits := config.FreeLimits
	if plan == "pro" {
		limits = config.ProLimits
	}

	// Obtener uso actual
	ctx := r.Context()
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var expenses int
	a.Store.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM expense WHERE user_id=$1 AND deleted=0 AND date >= $2`,
		uid, start.Format("2006-01-02")).Scan(&expenses)

	var accounts int
	a.Store.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM account WHERE user_id=$1 AND deleted=0`, uid).Scan(&accounts)

	var loans int
	a.Store.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM loan WHERE user_id=$1 AND deleted=0`, uid).Scan(&loans)

	writeJSON(w, http.StatusOK, subscriptionResponse{
		Plan:   plan,
		Status: status,
		Limits: limits,
		Usage: usageInfo{
			ExpensesThisMonth: expenses,
			TotalAccounts:     accounts,
			TotalLoans:        loans,
		},
	})
}

// handleActivatePro activa el plan Pro para el usuario (simulado - sin Stripe por ahora).
func (a *App) handleActivatePro(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())

	_, err := a.Store.Pool().Exec(r.Context(),
		`INSERT INTO subscription (user_id, plan, status, started_at, expires_at, updated_at)
		 VALUES ($1, 'pro', 'active', now(), now() + interval '30 days', now())
		 ON CONFLICT (user_id) DO UPDATE SET plan='pro', status='active',
		 expires_at=now() + interval '30 days', updated_at=now()`,
		uid)
	if err != nil {
		writeServerError(w, r, "error al activar plan", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleCancelSubscription cancela la suscripción del usuario.
func (a *App) handleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())

	_, err := a.Store.Pool().Exec(r.Context(),
		`UPDATE subscription SET status='cancelled', updated_at=now() WHERE user_id=$1`,
		uid)
	if err != nil {
		writeServerError(w, r, "error al cancelar", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
