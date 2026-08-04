package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"chiro/pkg/config"
	"chiro/pkg/store"
)

// Handler construye el router HTTP completo de la API.
func (a *App) Handler(cfg config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(CORS(cfg.CORSOrigins))
	r.Use(securityHeaders)
	r.Use(rateLimit(cfg))

	r.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// ── Público ────────────────────────────────────────────────────────────────
	r.Post("/api/auth/register", a.handleRegister)
	r.Post("/api/auth/login", a.handleLogin)
	r.Post("/api/auth/refresh", a.handleRefresh)

	// ── Protegido ──────────────────────────────────────────────────────────────
	r.Group(func(pr chi.Router) {
		pr.Use(a.Auth.Middleware)
		pr.Use(a.requireActive)
		pr.Get("/api/auth/me", a.handleMe)

		// ── Administración (solo rol admin) ────────────────────────────────────
		pr.Route("/api/admin", func(ar chi.Router) {
			ar.Use(a.requireAdmin)
			ar.Get("/users", a.handleAdminListUsers)
			ar.Get("/users/{id}", a.handleAdminGetUser)
			ar.Put("/users/{id}", a.handleAdminUpdateUser)
		})

		// Recursos genéricos (CRUD + soft-delete).
		for _, res := range []resource{
			{table: "account", plural: "accounts", orderBy: "name ASC"},
			{table: "category", plural: "categories", orderBy: "name ASC"},
			{table: "payment", plural: "payments", orderBy: "date DESC"},
			{table: "installment", plural: "installments", orderBy: "number ASC"},
			{table: "budget", plural: "budgets", orderBy: "year DESC, month DESC"},
			{table: "piggy_bank", plural: "piggy", orderBy: "name ASC"},
			{table: "bill", plural: "bills", orderBy: "next_date ASC"},
			{table: "tag", plural: "tags", orderBy: "name ASC"},
		} {
			pr.Route("/api/"+res.plural, func(rc chi.Router) {
				a.mountResource(rc, res)
			})
		}

		// Transacciones.
		pr.Get("/api/expenses", a.handleListExpenses)
		pr.Get("/api/expenses/{id}", a.handleGetExpense)
		pr.Post("/api/expenses", a.handleCreateExpense)
		pr.Put("/api/expenses/{id}", a.handleUpdateExpense)
		pr.Delete("/api/expenses/{id}", a.handleDeleteExpense)
		pr.Get("/api/summary", a.handleMonthSummary)
		pr.Post("/api/expenses/transfer", a.handleCreateTransfer)

		// Personas y préstamos.
		pr.Get("/api/persons", a.handleListPersons)
		pr.Delete("/api/persons/{id}", a.handleDeletePerson)
		pr.Get("/api/persons/{id}/loans", a.handlePersonLoans)

		pr.Get("/api/loans", a.handleListLoans)
		pr.Post("/api/loans", a.handleCreateLoan)
		pr.Put("/api/loans/{id}", a.handleUpdateLoan)
		pr.Delete("/api/loans/{id}", a.handleDeleteLoan)
		pr.Get("/api/loans/{id}/installments", a.handleLoanInstallments)
		pr.Post("/api/loans/migrate", a.handleMigrateLoans)

		pr.Post("/api/installments/{id}/pay", a.handlePayInstallment)
		pr.Post("/api/installments/{id}/cascade", a.handleCascadeInstallment)
		pr.Post("/api/installments/{id}/unpay", a.handleUnpayInstallment)

		// Facturas recurrentes.
		pr.Get("/api/bills/due", a.handleDueBills)
		pr.Post("/api/bills/{id}/pay", a.handlePayBill)
		pr.Post("/api/bills/{id}/skip", a.handleSkipBill)

		// Estadísticas y sync.
		pr.Get("/api/stats", a.handleStats)
		pr.Get("/api/budgets/progress", a.handleBudgetProgress)
		pr.Post("/api/sync", a.handleSync)

		// Suscripciones.
		pr.Get("/api/subscription", a.handleGetSubscription)
		pr.Post("/api/subscription/activate", a.handleActivatePro)
		pr.Post("/api/subscription/cancel", a.handleCancelSubscription)

		// Admin.
		pr.Get("/api/admin/stats", a.handleAdminStats)
		pr.Get("/api/admin/users", a.handleAdminListUsers)
	})

	return r
}

// Tables expone las tablas sincronizables (para tools y tests).
func Tables() []store.Tabla { return store.Tablas }
