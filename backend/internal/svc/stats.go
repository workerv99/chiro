package svc

import (
	"context"
	"fmt"
	"time"

	"chiro/internal/model"
	"chiro/internal/store"
)

// ── Resúmenes de gastos (excluyen transferencias y borrados) ─────────────────

// MonthSummary port de utils/repo/expenses.ts getMonthSummary (rango de fechas).
func MonthSummary(ctx context.Context, st *store.Store, userID string, year, month int) (model.MonthSummary, error) {
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	endY, endM := year, month+1
	if month == 12 {
		endY, endM = year+1, 1
	}
	end := fmt.Sprintf("%04d-%02d-01", endY, endM)

	rows, err := st.Pool().Query(ctx,
		`SELECT type, CAST(COALESCE(SUM(amount),0) AS float8) AS total
		 FROM expense
		 WHERE user_id=$1 AND deleted=0 AND date >= $2 AND date < $3 AND transfer_pair_id IS NULL
		 GROUP BY type`,
		userID, start, end)
	if err != nil {
		return model.MonthSummary{}, err
	}
	defer rows.Close()

	var income, expense float64
	for rows.Next() {
		var t string
		var total float64
		if err := rows.Scan(&t, &total); err != nil {
			return model.MonthSummary{}, err
		}
		if t == "income" {
			income = total
		} else {
			expense = total
		}
	}
	return model.MonthSummary{Income: income, Expense: expense, Balance: Round2(income - expense)}, rows.Err()
}

// YearSummary port de getYearSummary.
func YearSummary(ctx context.Context, st *store.Store, userID string, year int) ([]model.MonthRow, error) {
	rows, err := st.Pool().Query(ctx,
		`SELECT EXTRACT(MONTH FROM date)::int AS month, type, CAST(COALESCE(SUM(amount),0) AS float8) AS total
		 FROM expense
		 WHERE user_id=$1 AND deleted=0 AND EXTRACT(YEAR FROM date)::int = $2 AND transfer_pair_id IS NULL
		 GROUP BY month, type
		 ORDER BY month`,
		userID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byMonth := map[int]model.MonthRow{}
	for rows.Next() {
		var m int
		var t string
		var total float64
		if err := rows.Scan(&m, &t, &total); err != nil {
			return nil, err
		}
		r := byMonth[m]
		r.Month = m
		if t == "income" {
			r.Income = total
		} else {
			r.Expense = total
		}
		r.Balance = Round2(r.Income - r.Expense)
		byMonth[m] = r
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]model.MonthRow, 12)
	for i := 1; i <= 12; i++ {
		out[i-1] = byMonth[i]
		out[i-1].Month = i
		out[i-1].Balance = Round2(out[i-1].Income - out[i-1].Expense)
	}
	return out, nil
}

// CategoryBreakdown port de getCategoryBreakdown.
func CategoryBreakdown(ctx context.Context, st *store.Store, userID string, year, month int) ([]model.CategoryBreakdown, error) {
	sql := `SELECT c.category_id, c.name AS category_name, c.color AS category_color,
	           e.type, CAST(COALESCE(SUM(e.amount),0) AS float8) AS total
	        FROM expense e
	        JOIN category c ON c.user_id = e.user_id AND c.category_id = e.category_id
	        WHERE e.user_id=$1 AND e.deleted=0 AND e.transfer_pair_id IS NULL
	          AND e.category_id IS NOT NULL AND EXTRACT(YEAR FROM e.date)::int = $2`
	args := []any{userID, year}
	if month > 0 {
		sql += ` AND EXTRACT(MONTH FROM e.date)::int = $3`
		args = append(args, month)
	}
	sql += ` GROUP BY c.category_id, c.name, c.color, e.type ORDER BY total DESC`

	rows, err := st.Pool().Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.CategoryBreakdown, 0)
	for rows.Next() {
		var b model.CategoryBreakdown
		if err := rows.Scan(&b.CategoryID, &b.CategoryName, &b.CategoryColor, &b.Type, &b.Total); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// AccountBalances devuelve cuentas con balance calculado (suma de expense por cuenta).
func AccountBalances(ctx context.Context, st *store.Store, userID string) ([]model.AccountWithBalance, error) {
	rows, err := st.Pool().Query(ctx,
		`SELECT a.account_id, a.name, a.currency, a.account_type,
		        CAST(COALESCE((
		          SELECT SUM(CASE WHEN e.type='income' THEN e.amount ELSE -e.amount END)
		          FROM expense e
		          WHERE e.user_id=a.user_id AND e.account_id=a.account_id AND e.deleted=0
		        ), 0) AS float8) AS balance
		 FROM account a
		 WHERE a.user_id=$1 AND a.deleted=0
		 ORDER BY a.name ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.AccountWithBalance, 0)
	for rows.Next() {
		var a model.AccountWithBalance
		if err := rows.Scan(&a.AccountID, &a.Name, &a.Currency, &a.AccountType, &a.Balance); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// LentByMonth devuelve capital prestado por mes (índice 0 = enero).
func LentByMonth(ctx context.Context, st *store.Store, userID string, year int) ([]float64, error) {
	rows, err := st.Pool().Query(ctx,
		`SELECT EXTRACT(MONTH FROM date)::int AS month, CAST(COALESCE(SUM(amount),0) AS float8) AS total
		 FROM loan WHERE user_id=$1 AND deleted=0 AND EXTRACT(YEAR FROM date)::int = $2
		 GROUP BY month`, userID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]float64, 12)
	for rows.Next() {
		var m int
		var total float64
		if err := rows.Scan(&m, &total); err != nil {
			return nil, err
		}
		out[m-1] = total
	}
	return out, rows.Err()
}

// LentInYear es el capital prestado en un año.
func LentInYear(ctx context.Context, st *store.Store, userID string, year int) (float64, error) {
	row := st.Pool().QueryRow(ctx,
		`SELECT CAST(COALESCE(SUM(amount),0) AS float8) FROM loan
		 WHERE user_id=$1 AND deleted=0 AND EXTRACT(YEAR FROM date)::int = $2`, userID, year)
	var total float64
	err := row.Scan(&total)
	return total, err
}

// OutstandingTotal es el total global pendiente de cobrar (saldo de cuotas impagas).
func OutstandingTotal(ctx context.Context, st *store.Store, userID string) (float64, error) {
	row := st.Pool().QueryRow(ctx,
		`SELECT CAST(COALESCE(SUM(GREATEST(0, i.amount - i.paid_amount)),0) AS float8)
		 FROM installment i JOIN loan l ON l.user_id = i.user_id AND l.loan_id = i.loan_id
		 WHERE i.user_id=$1 AND i.deleted=0 AND l.deleted=0`, userID)
	var total float64
	err := row.Scan(&total)
	return total, err
}

// BudgetProgress devuelve presupuestos con su progreso de gasto.
func BudgetProgress(ctx context.Context, st *store.Store, userID string, year, month int) ([]model.BudgetWithProgress, error) {
	rows, err := st.Pool().Query(ctx,
		`SELECT b.budget_id, b.category_id, c.name AS category_name, c.color AS category_color,
		        b.amount, b.month, b.year,
		        CAST(COALESCE((
		          SELECT SUM(e.amount) FROM expense e
		          WHERE e.user_id=b.user_id AND e.category_id=b.category_id
		            AND e.deleted=0 AND e.transfer_pair_id IS NULL
		            AND EXTRACT(YEAR FROM e.date)::int = b.year
		            AND EXTRACT(MONTH FROM e.date)::int = b.month
		        ),0) AS float8) AS spent
		 FROM budget b
		 JOIN category c ON c.user_id = b.user_id AND c.category_id = b.category_id
		 WHERE b.user_id=$1 AND b.deleted=0 AND b.year=$2 AND b.month=$3
		 ORDER BY b.amount DESC`, userID, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.BudgetWithProgress, 0)
	for rows.Next() {
		var b model.BudgetWithProgress
		if err := rows.Scan(&b.BudgetID, &b.CategoryID, &b.CategoryName, &b.CategoryColor,
			&b.Amount, &b.Month, &b.Year, &b.Spent); err != nil {
			return nil, err
		}
		if b.Amount > 0 {
			b.Percentage = Round2(b.Spent / b.Amount * 100)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// PersonTotals calcula lo prestado y lo pendiente por persona.
func PersonTotals(ctx context.Context, st *store.Store, userID string, personID string) (loaned, pending float64, err error) {
	row := st.Pool().QueryRow(ctx,
		`SELECT CAST(COALESCE(SUM(l.amount + l.amount * COALESCE(l.interest_rate,0)/100.0),0) AS float8)
		 FROM loan l WHERE l.user_id=$1 AND l.person_id=$2 AND l.deleted=0`, userID, personID)
	if err = row.Scan(&loaned); err != nil {
		return 0, 0, err
	}
	row = st.Pool().QueryRow(ctx,
		`SELECT CAST(COALESCE(SUM(GREATEST(0, i.amount - i.paid_amount)),0) AS float8)
		 FROM installment i JOIN loan l ON l.user_id = i.user_id AND l.loan_id = i.loan_id
		 WHERE l.user_id=$1 AND l.person_id=$2 AND l.deleted=0 AND i.deleted=0`, userID, personID)
	err = row.Scan(&pending)
	return loaned, pending, err
}

// Helper para reusar queries con fecha actual.
func Today() string { return time.Now().UTC().Format("2006-01-02") }
