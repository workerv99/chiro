package app

import (
	"net/http"

	"chiro/pkg/auth"
	"github.com/go-chi/chi/v5"
)

type adminStats struct {
	TotalUsers    int `json:"total_users"`
	TotalExpenses int `json:"total_expenses"`
	TotalLoans    int `json:"total_loans"`
	ProUsers      int `json:"pro_users"`
	FreeUsers     int `json:"free_users"`
}

type userInfo struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Plan     string `json:"plan"`
}

// handleAdminStats devuelve métricas generales (solo admin).
func (a *App) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	var role string
	a.Store.Pool().QueryRow(r.Context(), `SELECT role FROM users WHERE user_id=$1`, uid).Scan(&role)
	if role != "admin" {
		writeErr(w, http.StatusForbidden, "se requiere rol admin")
		return
	}

	ctx := r.Context()
	var stats adminStats
	a.Store.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	a.Store.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM expense WHERE deleted=0`).Scan(&stats.TotalExpenses)
	a.Store.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM loan WHERE deleted=0`).Scan(&stats.TotalLoans)
	a.Store.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM subscription WHERE plan='pro'`).Scan(&stats.ProUsers)
	stats.FreeUsers = stats.TotalUsers - stats.ProUsers

	writeJSON(w, http.StatusOK, stats)
}

// handleAdminListUsers lista todos los usuarios (solo admin).
func (a *App) handleAdminListUsers(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	var role string
	a.Store.Pool().QueryRow(r.Context(), `SELECT role FROM users WHERE user_id=$1`, uid).Scan(&role)
	if role != "admin" {
		writeErr(w, http.StatusForbidden, "se requiere rol admin")
		return
	}

	rows, err := a.Store.Pool().Query(r.Context(),
		`SELECT u.user_id, u.email, u.name, u.role, u.status,
		 COALESCE(s.plan, 'free') as plan
		 FROM users u
		 LEFT JOIN subscription s ON s.user_id = u.user_id
		 ORDER BY u.created_at DESC`)
	if err != nil {
		writeServerError(w, r, "error", err)
		return
	}
	defer rows.Close()

	var users []userInfo
	for rows.Next() {
		var u userInfo
		rows.Scan(&u.UserID, &u.Email, &u.Name, &u.Role, &u.Status, &u.Plan)
		users = append(users, u)
	}

	writeJSON(w, http.StatusOK, users)
}

// handleAdminGetUser obtiene un usuario específico (solo admin).
func (a *App) handleAdminGetUser(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	var role string
	a.Store.Pool().QueryRow(r.Context(), `SELECT role FROM users WHERE user_id=$1`, uid).Scan(&role)
	if role != "admin" {
		writeErr(w, http.StatusForbidden, "se requiere rol admin")
		return
	}

	userID := chi.URLParam(r, "id")
	var u userInfo
	err := a.Store.Pool().QueryRow(r.Context(),
		`SELECT u.user_id, u.email, u.name, u.role, u.status,
		 COALESCE(s.plan, 'free') as plan
		 FROM users u
		 LEFT JOIN subscription s ON s.user_id = u.user_id
		 WHERE u.user_id=$1`, userID).Scan(&u.UserID, &u.Email, &u.Name, &u.Role, &u.Status, &u.Plan)
	if err != nil {
		writeErr(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	writeJSON(w, http.StatusOK, u)
}

// handleAdminUpdateUser actualiza un usuario (solo admin).
func (a *App) handleAdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	var role string
	a.Store.Pool().QueryRow(r.Context(), `SELECT role FROM users WHERE user_id=$1`, uid).Scan(&role)
	if role != "admin" {
		writeErr(w, http.StatusForbidden, "se requiere rol admin")
		return
	}

	var in struct {
		Status string `json:"status"`
		Role   string `json:"role"`
	}
	if !readJSON(w, r, &in) {
		return
	}

	userID := chi.URLParam(r, "id")
	_, err := a.Store.Pool().Exec(r.Context(),
		`UPDATE users SET status=$2, role=$3 WHERE user_id=$1`,
		userID, in.Status, in.Role)
	if err != nil {
		writeServerError(w, r, "error", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
