package svc

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"chiro/pkg/store"
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
func RegenerateSchedule(ctx context.Context, st *store.Store, userID string, loanID string, total float64, n int, firstDue string, freq string, customInstallment float64) error {
	count := max(1, n)
	var amounts []float64
	if customInstallment > 0 {
		amounts = SplitAmountsCustom(total, customInstallment)
	} else {
		amounts = SplitAmounts(total, count)
	}
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
		var loanID string
		var number int
		if err := tx.QueryRow(ctx,
			`SELECT loan_id, number FROM installment WHERE user_id=$1 AND installment_id=$2 AND deleted=0`,
			userID, installmentID).Scan(&loanID, &number); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE installment SET paid_amount=$3, paid_date=$4, updated_at=$5
			 WHERE user_id=$1 AND installment_id=$2 AND deleted=0`,
			userID, installmentID, Round2(paidAmount), paidDate, time.Now().UnixMilli()); err != nil {
			return err
		}
		if err := insertPaymentHistory(ctx, tx, userID, loanID, installmentID, paidAmount, paidDate,
			GenID("payhist"), []paymentAllocation{{id: installmentID, number: number, amount: Round2(paidAmount)}}); err != nil {
			return err
		}
		return recomputeLoanPaid(ctx, tx, userID, loanID, time.Now().UnixMilli())
	})
}

// CascadeResult resume lo que se aplicó al distribuir un pago en cascada.
type CascadeResult struct {
	InstallmentsAffected int     `json:"installments_affected"`
	TotalApplied         float64 `json:"total_applied"`
	StartingAmount       float64 `json:"starting_amount"`
	Excess               float64 `json:"excess"`
	AppliedCount         int     `json:"applied_count"`
	PaymentID            string  `json:"payment_id"`
}

// PayInstallmentCascade aplica un abono empezando en una cuota y volcando el
// excedente a las siguientes, sin pisar lo ya pagado.
func PayInstallmentCascade(ctx context.Context, st *store.Store, userID string, installmentID string, amount float64, paidDate string) (*CascadeResult, error) {
	var res CascadeResult
	err := st.ExecAll(ctx, func(tx pgx.Tx) error {
		var loanID string
		var num int
		if err := tx.QueryRow(ctx,
			`SELECT loan_id, number FROM installment WHERE user_id=$1 AND installment_id=$2 AND deleted=0`,
			userID, installmentID).Scan(&loanID, &num); err != nil {
			return err
		}

		rows, err := tx.Query(ctx,
			`SELECT installment_id, number, amount, paid_amount FROM installment
			 WHERE user_id=$1 AND loan_id=$2 AND number >= $3 AND deleted=0 ORDER BY number ASC`,
			userID, loanID, num)
		if err != nil {
			return err
		}
		var list []CascadeInstallment
		for rows.Next() {
			var i CascadeInstallment
			if err := rows.Scan(&i.ID, &i.Number, &i.Amount, &i.Paid); err != nil {
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

		allocs, startingRemaining, excess, err := AllocateCascade(list, amount)
		if err != nil {
			return err
		}
		res.StartingAmount = startingRemaining
		res.Excess = excess

		paidByID := make(map[string]float64, len(list))
		for _, i := range list {
			paidByID[i.ID] = i.Paid
		}

		ts := time.Now().UnixMilli()
		allocations := make([]paymentAllocation, 0, len(allocs))
		var totalApplied float64
		for _, a := range allocs {
			newPaid := Round2(paidByID[a.ID] + a.Amount)
			if _, err := tx.Exec(ctx,
				`UPDATE installment SET paid_amount=$3, paid_date=$4, updated_at=$5 WHERE user_id=$1 AND installment_id=$2`,
				userID, a.ID, newPaid, paidDate, ts); err != nil {
				return err
			}
			totalApplied = Round2(totalApplied + a.Amount)
			allocations = append(allocations, paymentAllocation{id: a.ID, number: a.Number, amount: a.Amount})
		}
		res.AppliedCount = len(allocs)
		res.InstallmentsAffected = res.AppliedCount
		res.TotalApplied = totalApplied
		res.PaymentID = GenID("payhist")
		if err := insertPaymentHistory(ctx, tx, userID, loanID, installmentID, totalApplied, paidDate, res.PaymentID, allocations); err != nil {
			return err
		}
		return recomputeLoanPaid(ctx, tx, userID, loanID, ts)
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

type paymentAllocation struct {
	id     string
	number int
	amount float64
}

func insertPaymentHistory(ctx context.Context, tx pgx.Tx, userID, loanID, startingID string, amount float64, date, paymentID string, allocations []paymentAllocation) error {
	ts := time.Now().UnixMilli()
	if _, err := tx.Exec(ctx,
		`INSERT INTO payment_history (user_id, payment_id, loan_id, starting_installment_id, amount, date, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		userID, paymentID, loanID, startingID, Round2(amount), date, ts); err != nil {
		return err
	}
	for _, allocation := range allocations {
		if _, err := tx.Exec(ctx,
			`INSERT INTO payment_history_allocation (user_id, payment_id, installment_id, installment_number, amount)
			 VALUES ($1,$2,$3,$4,$5)`,
			userID, paymentID, allocation.id, allocation.number, Round2(allocation.amount)); err != nil {
			return err
		}
	}
	return nil
}

// UnpayInstallment revierte el pago de una cuota. Un mismo pago puede haber
// asignado montos a varias cuotas (abono en cascada): sólo se revierte la
// asignación de esta cuota, restando su monto del total del pago; el pago
// sólo se marca como borrado cuando ya no le queda ninguna asignación. Nunca
// se toca created_at (no hay updated_at en payment_history).
func UnpayInstallment(ctx context.Context, st *store.Store, userID string, installmentID string) error {
	return st.ExecAll(ctx, func(tx pgx.Tx) error {
		loanID, err := loanIDOf(ctx, tx, userID, installmentID)
		if err != nil {
			return err
		}

		rows, err := tx.Query(ctx,
			`SELECT payment_id, amount FROM payment_history_allocation
			 WHERE user_id=$1 AND installment_id=$2`, userID, installmentID)
		if err != nil {
			return err
		}
		type alloc struct {
			paymentID string
			amount    float64
		}
		var allocs []alloc
		for rows.Next() {
			var a alloc
			if err := rows.Scan(&a.paymentID, &a.amount); err != nil {
				rows.Close()
				return err
			}
			allocs = append(allocs, a)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

		for _, a := range allocs {
			var current float64
			if err := tx.QueryRow(ctx,
				`SELECT amount FROM payment_history WHERE user_id=$1 AND payment_id=$2`,
				userID, a.paymentID).Scan(&current); err != nil {
				return err
			}
			newAmount := max(0, Round2(current-a.amount))
			if _, err := tx.Exec(ctx,
				`UPDATE payment_history SET amount=$3 WHERE user_id=$1 AND payment_id=$2`,
				userID, a.paymentID, newAmount); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx,
				`DELETE FROM payment_history_allocation WHERE user_id=$1 AND payment_id=$2 AND installment_id=$3`,
				userID, a.paymentID, installmentID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx,
				`UPDATE payment_history SET deleted=1
				 WHERE user_id=$1 AND payment_id=$2 AND deleted=0
				   AND NOT EXISTS (
				     SELECT 1 FROM payment_history_allocation
				     WHERE user_id=$1 AND payment_id=$2
				   )`,
				userID, a.paymentID); err != nil {
				return err
			}
		}

		if _, err := tx.Exec(ctx,
			`UPDATE installment SET paid_amount=0, paid_date=NULL, updated_at=$3 WHERE user_id=$1 AND installment_id=$2`,
			userID, installmentID, time.Now().UnixMilli()); err != nil {
			return err
		}
		return recomputeLoanPaid(ctx, tx, userID, loanID, time.Now().UnixMilli())
	})
}

// UnpayLastPayment revierte el pago más reciente del préstamo al que pertenece
// la cuota dada: resta cada asignación de ese pago a su cuota, limpia paid_date
// donde la cuota queda sin pagos y elimina el pago del historial. Si el
// préstamo no tiene historial (datos anteriores), revierte sólo la cuota dada.
func UnpayLastPayment(ctx context.Context, st *store.Store, userID string, installmentID string) error {
	return st.ExecAll(ctx, func(tx pgx.Tx) error {
		loanID, err := loanIDOf(ctx, tx, userID, installmentID)
		if err != nil {
			return err
		}
		var paymentID string
		err = tx.QueryRow(ctx,
			`SELECT payment_id FROM payment_history
			 WHERE user_id=$1 AND loan_id=$2 AND deleted=0
			 ORDER BY date DESC, created_at DESC LIMIT 1`, userID, loanID).Scan(&paymentID)
		if err == pgx.ErrNoRows {
			return unpayInstallmentTx(ctx, tx, userID, loanID, installmentID)
		}
		if err != nil {
			return err
		}

		rows, err := tx.Query(ctx,
			`SELECT installment_id, amount FROM payment_history_allocation
			 WHERE user_id=$1 AND payment_id=$2`, userID, paymentID)
		if err != nil {
			return err
		}
		var allocs []CascadeAllocation
		for rows.Next() {
			var a CascadeAllocation
			if err := rows.Scan(&a.ID, &a.Amount); err != nil {
				rows.Close()
				return err
			}
			allocs = append(allocs, a)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

		ts := time.Now().UnixMilli()
		for _, a := range allocs {
			var paid float64
			if err := tx.QueryRow(ctx,
				`SELECT paid_amount FROM installment WHERE user_id=$1 AND installment_id=$2`,
				userID, a.ID).Scan(&paid); err != nil {
				return err
			}
			newPaid, clearDate := ReversePaid(paid, a.Amount)
			q := `UPDATE installment SET paid_amount=$3, updated_at=$4 WHERE user_id=$1 AND installment_id=$2`
			if clearDate {
				q = `UPDATE installment SET paid_amount=$3, updated_at=$4, paid_date=NULL WHERE user_id=$1 AND installment_id=$2`
			}
			if _, err := tx.Exec(ctx, q, userID, a.ID, newPaid, ts); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx,
			`UPDATE payment_history SET deleted=1 WHERE user_id=$1 AND payment_id=$2`, userID, paymentID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`DELETE FROM payment_history_allocation WHERE user_id=$1 AND payment_id=$2`, userID, paymentID); err != nil {
			return err
		}
		return recomputeLoanPaid(ctx, tx, userID, loanID, ts)
	})
}

// unpayInstallmentTx pone la cuota en cero sin tocar el historial.
func unpayInstallmentTx(ctx context.Context, tx pgx.Tx, userID, loanID, installmentID string) error {
	ts := time.Now().UnixMilli()
	if _, err := tx.Exec(ctx,
		`UPDATE installment SET paid_amount=0, paid_date=NULL, updated_at=$3 WHERE user_id=$1 AND installment_id=$2`,
		userID, installmentID, ts); err != nil {
		return err
	}
	return recomputeLoanPaid(ctx, tx, userID, loanID, ts)
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
