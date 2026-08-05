package app

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"chiro/pkg/auth"
	"chiro/pkg/model"
	"chiro/pkg/svc"
)

// ── Personas ──────────────────────────────────────────────────────────────────

// handleListPersons lista personas con lo prestado y lo pendiente (port getPersons).
func (a *App) handleListPersons(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	rows, err := a.Store.Pool().Query(r.Context(),
		`SELECT p.person_id, p.name, p.notes,
		        COALESCE(loan_agg.total_loaned, 0) AS total_loaned,
		        COALESCE(inst_agg.total_pending, 0) AS total_pending
		 FROM person p
		 LEFT JOIN (
		     SELECT l.person_id, SUM(l.amount + l.amount * COALESCE(l.interest_rate,0)/100.0) AS total_loaned
		     FROM loan l WHERE l.user_id=$1 AND l.deleted=0 GROUP BY l.person_id
		 ) loan_agg ON loan_agg.person_id = p.person_id
		 LEFT JOIN (
		     SELECT l.person_id, SUM(GREATEST(0, i.amount - i.paid_amount)) AS total_pending
		     FROM installment i JOIN loan l ON l.user_id=i.user_id AND l.loan_id=i.loan_id
		     WHERE l.user_id=$1 AND l.deleted=0 AND i.deleted=0 GROUP BY l.person_id
		 ) inst_agg ON inst_agg.person_id = p.person_id
		 WHERE p.user_id=$1 AND p.deleted=0
		 ORDER BY p.name ASC`, uid)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	defer rows.Close()
	out := make([]model.PersonWithTotal, 0)
	for rows.Next() {
		var p model.PersonWithTotal
		if err := rows.Scan(&p.PersonID, &p.Name, &p.Notes, &p.TotalLoaned, &p.TotalPending); err != nil {
			writeServerError(w, r, "error interno del servidor", err)
			return
		}
		out = append(out, p)
	}
	writeJSON(w, http.StatusOK, out)
}

// handleDeletePerson borra en cascada pagos, préstamos y la persona (port deletePerson).
func (a *App) handleDeletePerson(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	id := chi.URLParam(r, "id")
	ts := time.Now().UnixMilli()
	err := a.Store.ExecAll(r.Context(), func(tx pgx.Tx) error {
		if _, err := tx.Exec(r.Context(),
			`UPDATE installment SET deleted=1, updated_at=$3 WHERE user_id=$1 AND loan_id IN (SELECT loan_id FROM loan WHERE user_id=$1 AND person_id=$2)`,
			uid, id, ts); err != nil {
			return err
		}
		if _, err := tx.Exec(r.Context(),
			`UPDATE payment SET deleted=1, updated_at=$3 WHERE user_id=$1 AND loan_id IN (SELECT loan_id FROM loan WHERE user_id=$1 AND person_id=$2)`,
			uid, id, ts); err != nil {
			return err
		}
		if _, err := tx.Exec(r.Context(),
			`UPDATE loan SET deleted=1, updated_at=$3 WHERE user_id=$1 AND person_id=$2`, uid, id, ts); err != nil {
			return err
		}
		_, err := tx.Exec(r.Context(),
			`UPDATE person SET deleted=1, updated_at=$3 WHERE user_id=$1 AND person_id=$2`, uid, id, ts)
		return err
	})
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handlePersonLoans lista los préstamos de una persona.
func (a *App) handlePersonLoans(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	pid := chi.URLParam(r, "id")
	rows, err := a.Store.Pool().Query(r.Context(),
		`SELECT l.*, CAST(COALESCE((SELECT SUM(i.paid_amount) FROM installment i WHERE i.user_id=l.user_id AND i.loan_id=l.loan_id AND i.deleted=0),0) AS float8) AS total_paid
		 FROM loan l WHERE l.user_id=$1 AND l.person_id=$2 AND l.deleted=0 ORDER BY l.date DESC`, uid, pid)
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

// ── Préstamos ─────────────────────────────────────────────────────────────────

type loanInput struct {
	PersonID          string  `json:"person_id"`
	Description       string  `json:"description"`
	Amount            float64 `json:"amount"`
	Date              string  `json:"date"`
	InterestRate      float64 `json:"interest_rate"`
	InterestType      string  `json:"interest_type"`
	Months            int     `json:"months"`
	Frequency         string  `json:"frequency"`
	FirstDueDate      *string `json:"first_due_date"`
	CustomInstallment float64 `json:"custom_installment"`
}

// handleListLoans lista préstamos con persona y total pagado (?year=&month=).
func (a *App) handleListLoans(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))

	sql := `SELECT l.loan_id, l.person_id, COALESCE(p.name,'') AS person_name, l.description, l.amount, to_char(l.date,'YYYY-MM-DD'),
	               l.is_paid, l.interest_rate, l.interest_type, l.months, l.frequency, to_char(l.first_due_date,'YYYY-MM-DD'),
	               COALESCE(inst_agg.total_paid, 0) AS total_paid,
	               CAST(l.amount + l.amount * COALESCE(l.interest_rate,0)/100.0 AS float8) AS total_interest,
	               CAST(l.amount + l.amount * COALESCE(l.interest_rate,0)/100.0 AS float8) AS total_amount
	        FROM loan l
	        LEFT JOIN person p ON p.user_id = l.user_id AND p.person_id = l.person_id
	        LEFT JOIN (
	            SELECT i.loan_id, SUM(i.paid_amount) AS total_paid
	            FROM installment i WHERE i.user_id=$1 AND i.deleted=0 GROUP BY i.loan_id
	        ) inst_agg ON inst_agg.loan_id = l.loan_id
	        WHERE l.user_id=$1 AND l.deleted=0 AND COALESCE(p.deleted,0)=0`
	args := []any{uid}
	if year > 0 {
		sql += ` AND EXTRACT(YEAR FROM l.date)::int = $` + strconv.Itoa(len(args)+1)
		args = append(args, year)
		if month > 0 {
			sql += ` AND EXTRACT(MONTH FROM l.date)::int = $` + strconv.Itoa(len(args)+1)
			args = append(args, month)
		}
	}
	sql += ` ORDER BY l.date DESC`

	rows, err := a.Store.Pool().Query(r.Context(), sql, args...)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	defer rows.Close()
	out := make([]model.LoanWithPerson, 0)
	for rows.Next() {
		var l model.LoanWithPerson
		if err := rows.Scan(&l.LoanID, &l.PersonID, &l.PersonName, &l.Description, &l.Amount, &l.Date,
			&l.IsPaid, &l.InterestRate, &l.InterestType, &l.Months, &l.Frequency, &l.FirstDueDate,
			&l.TotalPaid, &l.TotalInterest, &l.TotalAmount); err != nil {
			writeServerError(w, r, "error interno del servidor", err)
			return
		}
		out = append(out, l)
	}
	writeJSON(w, http.StatusOK, out)
}

// handleCreateLoan crea el préstamo y genera su cronograma (port createLoan).
func (a *App) handleCreateLoan(w http.ResponseWriter, r *http.Request) {
	var in loanInput
	if !readJSON(w, r, &in) {
		return
	}
	if in.Description == "" || in.Amount <= 0 || in.PersonID == "" {
		writeErr(w, http.StatusBadRequest, "description, amount y person_id requeridos")
		return
	}
	// Validaciones de campos.
	if err := svc.ValidateAmountPositive(in.Amount); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := svc.ValidateDate(in.Date); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := svc.ValidateInterestType(in.InterestType); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := svc.ValidateFrequency(in.Frequency); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if in.FirstDueDate != nil {
		if err := svc.ValidateDate(*in.FirstDueDate); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	uid := auth.ContextUser(r.Context())
	// Validar que la persona existe y pertenece al usuario.
	var exists int
	if err := a.Store.Pool().QueryRow(r.Context(),
		`SELECT COUNT(*) FROM person WHERE user_id=$1 AND person_id=$2 AND deleted=0`, uid, in.PersonID).Scan(&exists); err != nil || exists == 0 {
		writeErr(w, http.StatusBadRequest, "persona no encontrada")
		return
	}
	freq := orDefault(in.Frequency, "monthly")
	n := max(1, in.Months)
	firstDue := ""
	if in.FirstDueDate != nil {
		firstDue = *in.FirstDueDate
	}
	if firstDue == "" {
		firstDue, _ = svc.AdvanceByFreq(in.Date, freq, 1)
	}
	loanID := svc.GenID("loan")
	ts := time.Now().UnixMilli()
	total := svc.LoanTotal(in.Amount, in.InterestRate, n, in.InterestType)

	err := a.Store.ExecAll(r.Context(), func(tx pgx.Tx) error {
		if _, err := tx.Exec(r.Context(),
			`INSERT INTO loan (user_id, loan_id, person_id, description, amount, date, is_paid,
			   interest_rate, interest_type, months, frequency, first_due_date, updated_at, deleted)
			 VALUES ($1,$2,$3,$4,$5,$6,0,$7,$8,$9,$10,$11,$12,0)`,
			uid, loanID, in.PersonID, in.Description, in.Amount, in.Date,
			in.InterestRate, orDefault(in.InterestType, "simple"), n, freq, firstDue, ts); err != nil {
			return err
		}
		var amounts []float64
		if in.CustomInstallment > 0 {
			amounts = svc.SplitAmountsCustom(total, in.CustomInstallment)
		} else {
			amounts = svc.SplitAmounts(total, n)
		}
		for i := range amounts {
			num := i + 1
			due, err := svc.AdvanceByFreq(firstDue, freq, i)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(r.Context(),
				`INSERT INTO installment (user_id, installment_id, loan_id, number, due_date, amount, paid_date, paid_amount, updated_at, deleted)
				 VALUES ($1,$2,$3,$4,$5,$6,NULL,0,$7,0)
				 ON CONFLICT (user_id, installment_id) DO NOTHING`,
				uid, svc.InstallmentID(loanID, num), loanID, num, due, amounts[i], ts); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"loan_id": loanID})
}

// handleUpdateLoan actualiza y regenera el cronograma (port updateLoan).
func (a *App) handleUpdateLoan(w http.ResponseWriter, r *http.Request) {
	var in loanInput
	if !readJSON(w, r, &in) {
		return
	}
	if err := svc.ValidateDate(in.Date); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := svc.ValidateInterestType(in.InterestType); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := svc.ValidateFrequency(in.Frequency); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	uid := auth.ContextUser(r.Context())
	loanID := chi.URLParam(r, "id")
	freq := orDefault(in.Frequency, "monthly")
	n := max(1, in.Months)
	firstDue := ""
	if in.FirstDueDate != nil {
		firstDue = *in.FirstDueDate
	}
	if firstDue == "" {
		firstDue, _ = svc.AdvanceByFreq(in.Date, freq, 1)
	}
	total := svc.LoanTotal(in.Amount, in.InterestRate, n, in.InterestType)

	err := a.Store.ExecAll(r.Context(), func(tx pgx.Tx) error {
		res, err := tx.Exec(r.Context(),
			`UPDATE loan SET description=$3, amount=$4, date=$5, interest_rate=$6, interest_type=$7,
			   months=$8, frequency=$9, first_due_date=$10, updated_at=$11
			 WHERE user_id=$1 AND loan_id=$2 AND deleted=0`,
			uid, loanID, in.Description, in.Amount, in.Date, in.InterestRate,
			orDefault(in.InterestType, "simple"), n, freq, firstDue, time.Now().UnixMilli())
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return errNotFound
		}
		return nil
	})
	if err != nil {
		if err == errNotFound {
			writeErr(w, http.StatusNotFound, "préstamo no encontrado")
			return
		}
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	if err := svc.RegenerateSchedule(r.Context(), a.Store, uid, loanID, total, n, firstDue, freq, in.CustomInstallment); err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleDeleteLoan borra pagos y préstamo (port deleteLoan).
func (a *App) handleDeleteLoan(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	loanID := chi.URLParam(r, "id")
	ts := time.Now().UnixMilli()
	err := a.Store.ExecAll(r.Context(), func(tx pgx.Tx) error {
		if _, err := tx.Exec(r.Context(),
			`UPDATE installment SET deleted=1, updated_at=$3 WHERE user_id=$1 AND loan_id=$2`, uid, loanID, ts); err != nil {
			return err
		}
		if _, err := tx.Exec(r.Context(),
			`UPDATE payment SET deleted=1, updated_at=$3 WHERE user_id=$1 AND loan_id=$2`, uid, loanID, ts); err != nil {
			return err
		}
		_, err := tx.Exec(r.Context(),
			`UPDATE loan SET deleted=1, updated_at=$3 WHERE user_id=$1 AND loan_id=$2`, uid, loanID, ts)
		return err
	})
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── Cuotas ────────────────────────────────────────────────────────────────────

// handleLoanInstallments devuelve las cuotas de un préstamo con montos efectivos.
func (a *App) handleLoanInstallments(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	loanID := chi.URLParam(r, "id")
	rows, err := a.Store.Pool().Query(r.Context(),
		`SELECT installment_id, loan_id, number, to_char(due_date,'YYYY-MM-DD'),
		        amount, to_char(paid_date,'YYYY-MM-DD'), paid_amount
		 FROM installment WHERE user_id=$1 AND loan_id=$2 AND deleted=0 ORDER BY number ASC`, uid, loanID)
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	defer rows.Close()

	var insts []svc.InstallmentIn
	for rows.Next() {
		var i svc.InstallmentIn
		var paidDate *string
		if err := rows.Scan(&i.InstallmentID, &i.LoanID, &i.Number, &i.DueDate, &i.Amount, &paidDate, &i.PaidAmount); err != nil {
			writeServerError(w, r, "error interno del servidor", err)
			return
		}
		i.PaidDate = paidDate
		insts = append(insts, i)
	}
	if err := rows.Err(); err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}

	effective := svc.ComputeEffectiveAmounts(insts, svc.Today())
	out := make([]model.Installment, len(effective))
	for i, inst := range effective {
		isOverdue := inst.DueDate < svc.Today()
		fullyPaid := inst.PaidAmount >= inst.Amount
		out[i] = model.Installment{
			InstallmentID: inst.InstallmentID,
			LoanID:        inst.LoanID,
			Number:        inst.Number,
			DueDate:       inst.DueDate,
			Amount:        inst.Amount,
			PaidDate:      inst.PaidDate,
			PaidAmount:    inst.PaidAmount,
			Effective:     svc.Round2(inst.Amount - inst.PaidAmount),
			IsOverdue:     isOverdue,
			IsPartial:     inst.PaidAmount > 0 && !fullyPaid,
			IsPaid:        fullyPaid,
			Remaining:     max(0, svc.Round2(inst.Amount-inst.PaidAmount)),
		}
	}
	writeJSON(w, http.StatusOK, out)
}

type payReq struct {
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
}

func (a *App) handlePayInstallment(w http.ResponseWriter, r *http.Request) {
	a.installmentAction(w, r, "pay")
}
func (a *App) handleCascadeInstallment(w http.ResponseWriter, r *http.Request) {
	a.installmentAction(w, r, "cascade")
}
func (a *App) handleUnpayInstallment(w http.ResponseWriter, r *http.Request) {
	a.installmentAction(w, r, "unpay")
}

func (a *App) installmentAction(w http.ResponseWriter, r *http.Request, kind string) {
	uid := auth.ContextUser(r.Context())
	id := chi.URLParam(r, "id")
	date := svc.Today()
	var amount float64
	if kind != "unpay" {
		var req payReq
		if !readJSON(w, r, &req) {
			return
		}
		amount = req.Amount
		if req.Date != "" {
			if err := svc.ValidateDate(req.Date); err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			date = req.Date
		}
	}
	var err error
	switch kind {
	case "pay":
		err = svc.PayInstallment(r.Context(), a.Store, uid, id, amount, date)
	case "cascade":
		err = svc.PayInstallmentCascade(r.Context(), a.Store, uid, id, amount, date)
	case "unpay":
		err = svc.UnpayInstallment(r.Context(), a.Store, uid, id)
	}
	if err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleMigrateLoans genera cuotas para préstamos sin ellas (migración idempotente).
func (a *App) handleMigrateLoans(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	if err := svc.MigrateLoansToInstallments(r.Context(), a.Store, uid); err != nil {
		writeServerError(w, r, "error interno del servidor", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

var errNotFound = &notFoundErr{}

type notFoundErr struct{}

func (*notFoundErr) Error() string { return "no encontrado" }
