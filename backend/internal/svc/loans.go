package svc

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"chiro/internal/store"
)

// CreateInstallments genera las n cuotas de un préstamo con ids deterministas
// (ins_<loan>_<num>) para que la sincronización deduplique en vez de repetir.
func CreateInstallments(ctx context.Context, st *store.Store, userID string, loanID string, total float64, n int, firstDue string, freq string) error {
	amounts := SplitAmounts(total, n)
	ts := time.Now().UnixMilli()
	return st.ExecAll(ctx, func(tx pgx.Tx) error {
		for i := range amounts {
			num := i + 1
			due, err := AdvanceByFreq(firstDue, freq, i)
			if err != nil {
				return err
			}
			if err := insertInstallment(ctx, tx, userID, InstallmentID(loanID, num), loanID, num, due, amounts[i], nil, 0, ts, 0); err != nil {
				return err
			}
		}
		return nil
	})
}

// RegenerateSchedule recalcula vencimientos y montos conservando lo pagado en
// cada número de cuota; las sobrantes quedan borradas (soft).
func RegenerateSchedule(ctx context.Context, st *store.Store, userID string, loanID string, total float64, n int, firstDue string, freq string) error {
	count := max(1, n)
	amounts := SplitAmounts(total, count)
	ts := time.Now().UnixMilli()

	return st.ExecAll(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`UPDATE installment SET deleted=1, updated_at=$3 WHERE user_id=$1 AND loan_id=$2 AND number > $4 AND deleted=0`,
			userID, loanID, ts, count); err != nil {
			return err
		}
		for i := 0; i < count; i++ {
			num := i + 1
			due, err := AdvanceByFreq(firstDue, freq, i)
			if err != nil {
				return err
			}
			if err := upsertInstallment(ctx, tx, userID, InstallmentID(loanID, num), loanID, num, due, amounts[i], ts); err != nil {
				return err
			}
		}
		return recomputeLoanPaid(ctx, tx, userID, loanID, ts)
	})
}

// ── Pagos de cuotas ───────────────────────────────────────────────────────────

// PayInstallment registra un pago (total o parcial). Si el monto es <= 0 lo
// trata como un "unpay" (port de utils/repo/installments.ts).
func PayInstallment(ctx context.Context, st *store.Store, userID string, installmentID string, paidAmount float64, paidDate string) error {
	if paidAmount <= 0 {
		return UnpayInstallment(ctx, st, userID, installmentID)
	}
	return st.ExecAll(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`UPDATE installment SET paid_amount=$3, paid_date=$4, updated_at=$5
			 WHERE user_id=$1 AND installment_id=$2 AND deleted=0`,
			userID, installmentID, Round2(paidAmount), paidDate, time.Now().UnixMilli()); err != nil {
			return err
		}
		loanID, err := loanIDOf(ctx, tx, userID, installmentID)
		if err != nil {
			return err
		}
		return recomputeLoanPaid(ctx, tx, userID, loanID, time.Now().UnixMilli())
	})
}

// PayInstallmentCascade aplica un abono empezando en una cuota y volcando el
// excedente a las siguientes, sin pisar lo ya pagado.
func PayInstallmentCascade(ctx context.Context, st *store.Store, userID string, installmentID string, amount float64, paidDate string) error {
	return st.ExecAll(ctx, func(tx pgx.Tx) error {
		var loanID string
		var num int
		if err := tx.QueryRow(ctx,
			`SELECT loan_id, number FROM installment WHERE user_id=$1 AND installment_id=$2 AND deleted=0`,
			userID, installmentID).Scan(&loanID, &num); err != nil {
			return err
		}

		rows, err := tx.Query(ctx,
			`SELECT installment_id, amount, paid_amount FROM installment
			 WHERE user_id=$1 AND loan_id=$2 AND number >= $3 AND deleted=0 ORDER BY number ASC`,
			userID, loanID, num)
		if err != nil {
			return err
		}
		type ins struct {
			id   string
			amt  float64
			paid float64
		}
		var list []ins
		for rows.Next() {
			var i ins
			if err := rows.Scan(&i.id, &i.amt, &i.paid); err != nil {
				rows.Close()
				return err
			}
			list = append(list, i)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

		remaining := amount
		ts := time.Now().UnixMilli()
		for _, i := range list {
			if remaining <= 0 {
				break
			}
			space := Round2(i.amt - i.paid)
			if space <= 0 {
				continue
			}
			give := min(remaining, space)
			newPaid := Round2(i.paid + give)
			if _, err := tx.Exec(ctx,
				`UPDATE installment SET paid_amount=$3, paid_date=$4, updated_at=$5 WHERE user_id=$1 AND installment_id=$2`,
				userID, i.id, newPaid, paidDate, ts); err != nil {
				return err
			}
			remaining = Round2(remaining - give)
		}
		return recomputeLoanPaid(ctx, tx, userID, loanID, ts)
	})
}

// UnpayInstallment revierte el pago de una cuota.
func UnpayInstallment(ctx context.Context, st *store.Store, userID string, installmentID string) error {
	return st.ExecAll(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`UPDATE installment SET paid_amount=0, paid_date=NULL, updated_at=$3 WHERE user_id=$1 AND installment_id=$2`,
			userID, installmentID, time.Now().UnixMilli()); err != nil {
			return err
		}
		loanID, err := loanIDOf(ctx, tx, userID, installmentID)
		if err != nil {
			return err
		}
		return recomputeLoanPaid(ctx, tx, userID, loanID, time.Now().UnixMilli())
	})
}

// ── Internos ──────────────────────────────────────────────────────────────────

func insertInstallment(ctx context.Context, tx pgx.Tx, userID, id, loanID string, num int, due string, amount float64, paidDate *string, paidAmount float64, ts int64, deleted int) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO installment (user_id, installment_id, loan_id, number, due_date, amount, paid_date, paid_amount, updated_at, deleted)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 ON CONFLICT (user_id, installment_id) DO NOTHING`,
		userID, id, loanID, num, due, amount, paidDate, paidAmount, ts, deleted)
	return err
}

func upsertInstallment(ctx context.Context, tx pgx.Tx, userID, id, loanID string, num int, due string, amount float64, ts int64) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO installment (user_id, installment_id, loan_id, number, due_date, amount, paid_date, paid_amount, updated_at, deleted)
		 VALUES ($1,$2,$3,$4,$5,$6,NULL,0,$7,0)
		 ON CONFLICT (user_id, installment_id) DO UPDATE SET
		   due_date=EXCLUDED.due_date, amount=EXCLUDED.amount, deleted=0, updated_at=EXCLUDED.updated_at`,
		userID, id, loanID, num, due, amount, ts)
	return err
}

func loanIDOf(ctx context.Context, tx pgx.Tx, userID, installmentID string) (string, error) {
	var loanID string
	err := tx.QueryRow(ctx,
		`SELECT loan_id FROM installment WHERE user_id=$1 AND installment_id=$2`,
		userID, installmentID).Scan(&loanID)
	return loanID, err
}

// recomputeLoanPaid: pagado cuando toda cuota tiene paid_amount >= amount.
func recomputeLoanPaid(ctx context.Context, tx pgx.Tx, userID, loanID string, ts int64) error {
	var total, paid int
	err := tx.QueryRow(ctx,
		`SELECT COUNT(*)::int AS total, COALESCE(SUM(CASE WHEN paid_amount >= amount THEN 1 ELSE 0 END),0)::int AS paid
		 FROM installment WHERE user_id=$1 AND loan_id=$2 AND deleted=0`, userID, loanID).Scan(&total, &paid)
	if err != nil {
		return err
	}
	isPaid := 0
	if total > 0 && paid == total {
		isPaid = 1
	}
	_, err = tx.Exec(ctx,
		`UPDATE loan SET is_paid=$3, updated_at=$4 WHERE user_id=$1 AND loan_id=$2`,
		userID, loanID, isPaid, ts)
	return err
}

// MigrateLoansToInstallments genera cuotas para préstamos sin ellas,
// repartiendo lo ya pagado (tabla payment) entre las primeras cuotas.
func MigrateLoansToInstallments(ctx context.Context, st *store.Store, userID string) error {
	type loanRow struct {
		LoanID       string
		Amount       float64
		InterestRate float64
		Months       int
		Date         string
		Frequency    string
		FirstDue     *string
	}
	var loans []loanRow
	rows, err := st.Pool().Query(ctx,
		`SELECT loan_id, amount, COALESCE(interest_rate,0), COALESCE(months,1),
		        to_char(date,'YYYY-MM-DD'), COALESCE(frequency,'monthly'), first_due_date
		 FROM loan WHERE user_id=$1 AND deleted=0`, userID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var l loanRow
		var firstDue any
		if err := rows.Scan(&l.LoanID, &l.Amount, &l.InterestRate, &l.Months, &l.Date, &l.Frequency, &firstDue); err != nil {
			rows.Close()
			return err
		}
		if fd, ok := firstDue.(time.Time); ok {
			s := fd.Format("2006-01-02")
			l.FirstDue = &s
		}
		loans = append(loans, l)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, loan := range loans {
		var count int
		if err := st.Pool().QueryRow(ctx,
			`SELECT COUNT(*)::int FROM installment WHERE user_id=$1 AND loan_id=$2 AND deleted=0`,
			userID, loan.LoanID).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		n := max(1, loan.Months)
		freq := loan.Frequency
		if freq == "" {
			freq = "monthly"
		}
		firstDue := ""
		if loan.FirstDue != nil {
			firstDue = *loan.FirstDue
		}
		if firstDue == "" {
			if fd, err := AdvanceByFreq(loan.Date, freq, 1); err == nil {
				firstDue = fd
			}
		}

		total := LoanTotal(loan.Amount, loan.InterestRate, n, "simple")
		amounts := SplitAmounts(total, n)

		pays, err := paymentTotals(ctx, st, userID, loan.LoanID)
		if err != nil {
			return err
		}
		remaining := pays.total
		payIdx := 0
		ts := time.Now().UnixMilli()

		if err := st.ExecAll(ctx, func(tx pgx.Tx) error {
			for i := 0; i < n; i++ {
				var pa float64
				var pd *string
				if remaining > 0 {
					pa = min(amounts[i], remaining)
					remaining = Round2(remaining - pa)
					for payIdx < len(pays.dates)-1 && remaining > 0 {
						payIdx++
					}
					last := pays.dates[min(payIdx, len(pays.dates)-1)]
					pd = &last
				}
				due, err := AdvanceByFreq(firstDue, freq, i)
				if err != nil {
					return err
				}
				if err := insertInstallment(ctx, tx, userID, InstallmentID(loan.LoanID, i+1), loan.LoanID, i+1, due, amounts[i], pd, pa, ts, 0); err != nil {
					return err
				}
			}
			return recomputeLoanPaid(ctx, tx, userID, loan.LoanID, ts)
		}); err != nil {
			return err
		}
	}
	return nil
}

type paymentSums struct {
	total float64
	dates []string
}

func paymentTotals(ctx context.Context, st *store.Store, userID, loanID string) (paymentSums, error) {
	rows, err := st.Pool().Query(ctx,
		`SELECT to_char(date,'YYYY-MM-DD'), amount FROM payment
		 WHERE user_id=$1 AND loan_id=$2 AND deleted=0 ORDER BY date ASC`, userID, loanID)
	if err != nil {
		return paymentSums{}, err
	}
	defer rows.Close()
	var ps paymentSums
	for rows.Next() {
		var d string
		var a float64
		if err := rows.Scan(&d, &a); err != nil {
			return ps, err
		}
		ps.total = Round2(ps.total + a)
		ps.dates = append(ps.dates, d)
	}
	return ps, rows.Err()
}
