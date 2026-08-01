// Package model define las formas tipadas de las respuestas calculadas.
package model

// MonthSummary es el resumen de un mes (excluye transferencias y borrados).
type MonthSummary struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
}

// MonthRow es el resumen por mes dentro de un año.
type MonthRow struct {
	Month   int     `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
}

// CategoryBreakdown es el total gastado/ingresado por categoría.
type CategoryBreakdown struct {
	CategoryID    string  `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	CategoryColor string  `json:"category_color"`
	Total         float64 `json:"total"`
	Type          string  `json:"type"`
}

// AccountWithBalance es una cuenta con su balance calculado.
type AccountWithBalance struct {
	AccountID   string  `json:"account_id"`
	Name        string  `json:"name"`
	Currency    string  `json:"currency"`
	AccountType string  `json:"account_type"`
	Balance     float64 `json:"balance"`
}

// PersonWithTotal es una persona con lo prestado y lo pendiente de cobrar.
type PersonWithTotal struct {
	PersonID     string  `json:"person_id"`
	Name         string  `json:"name"`
	Notes        *string `json:"notes"`
	TotalLoaned  float64 `json:"total_loaned"`
	TotalPending float64 `json:"total_pending"`
}

// LoanWithPerson es un préstamo con el nombre de la persona.
type LoanWithPerson struct {
	LoanID        string  `json:"loan_id"`
	PersonID      string  `json:"person_id"`
	PersonName    string  `json:"person_name"`
	Description   string  `json:"description"`
	Amount        float64 `json:"amount"`
	Date          string  `json:"date"`
	IsPaid        int     `json:"is_paid"`
	InterestRate  float64 `json:"interest_rate"`
	InterestType  string  `json:"interest_type"`
	Months        int     `json:"months"`
	Frequency     string  `json:"frequency"`
	FirstDueDate  *string `json:"first_due_date"`
	TotalPaid     float64 `json:"total_paid"`
	TotalInterest float64 `json:"total_interest"`
	TotalAmount   float64 `json:"total_amount"`
}

// Installment es una cuota con su monto efectivo (arrastre de deuda vencida).
type Installment struct {
	InstallmentID string  `json:"installment_id"`
	LoanID        string  `json:"loan_id"`
	Number        int     `json:"number"`
	DueDate       string  `json:"due_date"`
	Amount        float64 `json:"amount"`
	PaidDate      *string `json:"paid_date"`
	PaidAmount    float64 `json:"paid_amount"`
	Effective     float64 `json:"effective"`
	IsOverdue     bool    `json:"is_overdue"`
	IsPartial     bool    `json:"is_partial"`
	IsPaid        bool    `json:"is_paid"`
	Remaining     float64 `json:"remaining"`
}

// BudgetWithProgress es un presupuesto con el gasto de la categoría.
type BudgetWithProgress struct {
	BudgetID      string  `json:"budget_id"`
	CategoryID    string  `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	CategoryColor string  `json:"category_color"`
	Amount        float64 `json:"amount"`
	Month         int     `json:"month"`
	Year          int     `json:"year"`
	Spent         float64 `json:"spent"`
	Percentage    float64 `json:"percentage"`
}

// User es el usuario autenticado.
type User struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// AdminUser es un usuario visto por el panel de administración.
type AdminUser struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
	Expenses  int    `json:"expenses"`
	Loans     int    `json:"loans"`
}

// AuthResponse es la respuesta de register/login.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
