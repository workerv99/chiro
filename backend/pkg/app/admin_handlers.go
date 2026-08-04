package app

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"chiro/pkg/auth"
	"chiro/pkg/model"
)

// ── Consultas ─────────────────────────────────────────────────────────────────

const adminUserSelect = `
SELECT u.user_id, u.email, u.name, u.role, u.status, u.created_at,
       (SELECT COUNT(*) FROM expense e WHERE e.user_id = u.user_id AND e.deleted = 0)  AS expenses,
       (SELECT COUNT(*) FROM loan l   WHERE l.user_id = u.user_id AND l.deleted = 0)  AS loans
FROM users u`

func scanAdminUser(row interface{ Scan(...any) error }) (model.AdminUser, error) {
	var u model.AdminUser
	err := row.Scan(&u.UserID, &u.Email, &u.Name, &u.Role, &u.Status, &u.CreatedAt, &u.Expenses, &u.Loans)
	return u, err
}

// GET /api/admin/users?limit=100&offset=0
func (a *App) handleAdminListUsers(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	rows, err := a.Store.Pool().Query(r.Context(), adminUserSelect+` ORDER BY u.created_at ASC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "error al listar usuarios")
		return
	}
	defer rows.Close()

	out := make([]model.AdminUser, 0)
	for rows.Next() {
		u, err := scanAdminUser(rows)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "error al leer usuarios")
			return
		}
		out = append(out, u)
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /api/admin/users/{id}
func (a *App) handleAdminGetUser(w http.ResponseWriter, r *http.Request) {
	u, err := scanAdminUser(a.Store.Pool().QueryRow(r.Context(),
		adminUserSelect+` WHERE u.user_id = $1`, chi.URLParam(r, "id")))
	if err != nil {
		writeErr(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// ── Actualización ─────────────────────────────────────────────────────────────

type adminUpdateReq struct {
	Role     *string `json:"role"`
	Status   *string `json:"status"`
	Name     *string `json:"name"`
	Password *string `json:"password"`
}

// PUT /api/admin/users/{id}
// Body: { role?: "user"|"admin", status?: "active"|"disabled", name?: string, password?: string }
func (a *App) handleAdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req adminUpdateReq
	if !readJSON(w, r, &req) {
		return
	}

	if req.Role != nil && *req.Role != "user" && *req.Role != "admin" {
		writeErr(w, http.StatusBadRequest, "rol inválido (user|admin)")
		return
	}
	if req.Status != nil && *req.Status != "active" && *req.Status != "disabled" {
		writeErr(w, http.StatusBadRequest, "estado inválido (active|disabled)")
		return
	}
	if req.Password != nil && len(*req.Password) < 6 {
		writeErr(w, http.StatusBadRequest, "contraseña de al menos 6 caracteres")
		return
	}

	// Evitar que un admin se quite el acceso a sí mismo (lockout).
	if id == auth.ContextUser(r.Context()) {
		if (req.Role != nil && *req.Role == "user") || (req.Status != nil && *req.Status == "disabled") {
			writeErr(w, http.StatusBadRequest, "no puedes quitarte el rol admin ni desactivar tu cuenta")
			return
		}
	}

	cols := []string{}
	args := []any{}
	add := func(c string, v any) {
		cols = append(cols, c)
		args = append(args, v)
	}
	if req.Role != nil {
		add("role", *req.Role)
	}
	if req.Status != nil {
		add("status", *req.Status)
	}
	if req.Name != nil {
		add("name", *req.Name)
	}
	if req.Password != nil {
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "error al hashear")
			return
		}
		add("password_hash", hash)
	}
	if len(cols) == 0 {
		writeErr(w, http.StatusBadRequest, "nada que actualizar")
		return
	}

	sets := make([]string, len(cols))
	for i, c := range cols {
		sets[i] = c + " = $" + strconv.Itoa(i+1)
	}
	tag, err := a.Store.Pool().Exec(r.Context(),
		`UPDATE users SET `+strings.Join(sets, ", ")+` WHERE user_id = $`+strconv.Itoa(len(cols)+1),
		append(args, id)...)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "error al actualizar")
		return
	}
	if tag.RowsAffected() == 0 {
		writeErr(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	u, err := scanAdminUser(a.Store.Pool().QueryRow(r.Context(),
		adminUserSelect+` WHERE u.user_id = $1`, id))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "error al leer usuario")
		return
	}
	writeJSON(w, http.StatusOK, u)
}
