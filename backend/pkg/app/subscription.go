package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"chiro/pkg/auth"
	"chiro/pkg/config"
)

// checkLimits verifica que el usuario no haya excedido los límites de su plan.
func (a *App) checkLimits(plan string, resource string) error {
	limits := config.FreeLimits
	if plan == "pro" {
		limits = config.ProLimits
	}

	uid := "" // Se obtiene del contexto en el handler
	ctx := context.Background()

	switch resource {
	case "expense":
		if limits.MaxExpensesPerMonth < 0 {
			return nil // ilimitado
		}
		var count int
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		err := a.Store.Pool().QueryRow(ctx,
			`SELECT COUNT(*) FROM expense WHERE user_id=$1 AND deleted=0 AND date >= $2`,
			uid, start.Format("2006-01-02")).Scan(&count)
		if err != nil {
			return err
		}
		if count >= limits.MaxExpensesPerMonth {
			return fmt.Errorf("límite de %d gastos/mes alcanzado. Actualiza a Pro para gastos ilimitados", limits.MaxExpensesPerMonth)
		}
	case "account":
		if limits.MaxAccounts < 0 {
			return nil
		}
		var count int
		err := a.Store.Pool().QueryRow(ctx,
			`SELECT COUNT(*) FROM account WHERE user_id=$1 AND deleted=0`, uid).Scan(&count)
		if err != nil {
			return err
		}
		if count >= limits.MaxAccounts {
			return fmt.Errorf("límite de %d cuentas alcanzado. Actualiza a Pro", limits.MaxAccounts)
		}
	case "loan":
		if limits.MaxLoans < 0 {
			return nil
		}
		var count int
		err := a.Store.Pool().QueryRow(ctx,
			`SELECT COUNT(*) FROM loan WHERE user_id=$1 AND deleted=0`, uid).Scan(&count)
		if err != nil {
			return err
		}
		if count >= limits.MaxLoans {
			return fmt.Errorf("límite de %d préstamos alcanzado. Actualiza a Pro", limits.MaxLoans)
		}
	}
	return nil
}

// checkPlanLimits es un middleware que verifica límites antes de crear recursos.
func (a *App) checkPlanLimits(resource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid := auth.ContextUser(r.Context())
			if uid == "" {
				writeErr(w, http.StatusUnauthorized, "no autenticado")
				return
			}

			// Obtener plan del usuario
			var plan string
			err := a.Store.Pool().QueryRow(r.Context(),
				`SELECT COALESCE(plan, 'free') FROM subscription WHERE user_id=$1`, uid).Scan(&plan)
			if err != nil {
				plan = "free" // default
			}

			limits := config.FreeLimits
			if plan == "pro" {
				limits = config.ProLimits
			}

			ctx := context.WithValue(r.Context(), "plan", plan)
			ctx = context.WithValue(ctx, "limits", limits)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
