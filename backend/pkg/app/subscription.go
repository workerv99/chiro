package app

import (
	"context"
	"time"

	"chiro/pkg/config"
	"chiro/pkg/svc"
)

// planLimits obtiene los límites del plan del usuario (Free o Pro).
func (a *App) planLimits(ctx context.Context, uid string) config.PlanLimits {
	var plan string
	if err := a.Store.Pool().QueryRow(ctx,
		`SELECT COALESCE(plan, 'free') FROM subscription WHERE user_id=$1`, uid).Scan(&plan); err != nil {
		plan = "free"
	}
	if plan == "pro" {
		return config.ProLimits
	}
	return config.FreeLimits
}

// checkLimits verifica que el usuario no haya excedido el límite de su plan
// para el recurso indicado ("expense", "account" o "loan").
func (a *App) checkLimits(ctx context.Context, uid, resource string) error {
	limits := a.planLimits(ctx, uid)

	var limit int
	switch resource {
	case "expense":
		limit = limits.MaxExpensesPerMonth
	case "account":
		limit = limits.MaxAccounts
	case "loan":
		limit = limits.MaxLoans
	default:
		return nil
	}
	if limit < 0 {
		return nil
	}

	var count int
	var err error
	switch resource {
	case "expense":
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		err = a.Store.Pool().QueryRow(ctx,
			`SELECT COUNT(*) FROM expense WHERE user_id=$1 AND deleted=0 AND date >= $2`,
			uid, start.Format("2006-01-02")).Scan(&count)
	case "account":
		err = a.Store.Pool().QueryRow(ctx,
			`SELECT COUNT(*) FROM account WHERE user_id=$1 AND deleted=0`, uid).Scan(&count)
	case "loan":
		err = a.Store.Pool().QueryRow(ctx,
			`SELECT COUNT(*) FROM loan WHERE user_id=$1 AND deleted=0`, uid).Scan(&count)
	}
	if err != nil {
		return err
	}
	return svc.CheckLimit(resource, limit, count)
}
