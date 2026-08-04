package app

import (
	"context"
	"net/http"
	"strings"
	"time"

	"chiro/pkg/auth"
	"chiro/pkg/model"
	"chiro/pkg/svc"
)

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if !readJSON(w, r, &req) {
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || len(req.Password) < 6 {
		writeErr(w, http.StatusBadRequest, "Email requerido y contraseña de al menos 6 caracteres")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "error al hashear")
		return
	}
	uid := svc.GenID("usr")

	// Insertar el usuario.
	_, err = a.Store.Pool().Exec(r.Context(),
		`INSERT INTO users (user_id, email, name, password_hash, created_at) VALUES ($1,$2,$3,$4,$5)`,
		uid, req.Email, req.Name, hash, time.Now().UTC().UnixMilli())
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "23505") {
			writeErr(w, http.StatusConflict, "ese email ya está registrado")
			return
		}
		writeErr(w, http.StatusInternalServerError, "error al crear el usuario")
		return
	}

	// Semilla por defecto (port de SEED del proyecto original): cuentas y categorías.
	seedDefaults(r.Context(), a, uid)

	token, err := a.Auth.Issue(uid, req.Email, "user")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "error al firmar token")
		return
	}
	writeJSON(w, http.StatusOK, model.AuthResponse{
		Token: token,
		User:  model.User{UserID: uid, Email: req.Email, Name: req.Name, Role: "user", Status: "active"},
	})
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if !readJSON(w, r, &req) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	var uid, name, hash, role, status string
	err := a.Store.Pool().QueryRow(r.Context(),
		`SELECT user_id, name, password_hash, role, status FROM users WHERE email=$1`, email).
		Scan(&uid, &name, &hash, &role, &status)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "email o contraseña incorrectos")
		return
	}
	if !auth.CheckPassword(hash, req.Password) {
		writeErr(w, http.StatusUnauthorized, "email o contraseña incorrectos")
		return
	}
	if status == "disabled" {
		writeErr(w, http.StatusForbidden, "cuenta deshabilitada")
		return
	}
	if role == "" {
		role = "user"
	}
	token, err := a.Auth.Issue(uid, email, role)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "error al firmar token")
		return
	}
	writeJSON(w, http.StatusOK, model.AuthResponse{
		Token: token,
		User:  model.User{UserID: uid, Email: email, Name: name, Role: role, Status: status},
	})
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	var email, name, role, status string
	err := a.Store.Pool().QueryRow(r.Context(),
		`SELECT email, name, role, status FROM users WHERE user_id=$1`, uid).Scan(&email, &name, &role, &status)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "sesión inválida")
		return
	}
	writeJSON(w, http.StatusOK, model.User{UserID: uid, Email: email, Name: name, Role: role, Status: status})
}

// handleRefresh renueva un token válido (o expirado <7 días) con la firma
// actual. No requiere autenticación previa — el token viejo se envía como Bearer.
func (a *App) handleRefresh(w http.ResponseWriter, r *http.Request) {
	authz := r.Header.Get("Authorization")
	if len(authz) < 8 {
		writeErr(w, http.StatusUnauthorized, "token requerido")
		return
	}
	tokenStr := authz[7:]
	claims, err := a.Auth.ParseClaims(tokenStr)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "token inválido")
		return
	}
	if claims.ExpiresAt != nil {
		if time.Until(claims.ExpiresAt.Time) < -7*24*time.Hour {
			writeErr(w, http.StatusUnauthorized, "token expirado demasiado, inicia sesión de nuevo")
			return
		}
	}
	newToken, err := a.Auth.Issue(claims.Subject, claims.Email, claims.Role)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "error al firmar token")
		return
	}
	var name, role, status string
	err = a.Store.Pool().QueryRow(r.Context(),
		`SELECT name, role, status FROM users WHERE user_id=$1`, claims.Subject).Scan(&name, &role, &status)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "sesión inválida")
		return
	}
	if status == "disabled" {
		writeErr(w, http.StatusForbidden, "cuenta deshabilitada")
		return
	}
	if role == "" {
		role = "user"
	}
	writeJSON(w, http.StatusOK, model.AuthResponse{
		Token: newToken,
		User:  model.User{UserID: claims.Subject, Email: claims.Email, Name: name, Role: role, Status: status},
	})
}

// seedDefaults crea cuentas y categorías por defecto (port de utils/dbSchema SEED).
func seedDefaults(ctx context.Context, a *App, uid string) {
	accounts := []map[string]any{
		{"account_id": "acc_cash", "name": "Efectivo", "currency": "USD", "account_type": "asset"},
		{"account_id": "acc_bank", "name": "Banco", "currency": "USD", "account_type": "asset"},
	}
	categories := []map[string]any{
		{"category_id": "cat_food", "name": "Alimentación", "color": "#FF6B6B", "icon": "restaurant", "type": "expense"},
		{"category_id": "cat_transport", "name": "Transporte", "color": "#4ECDC4", "icon": "car", "type": "expense"},
		{"category_id": "cat_housing", "name": "Vivienda", "color": "#45B7D1", "icon": "home", "type": "expense"},
		{"category_id": "cat_health", "name": "Salud", "color": "#96CEB4", "icon": "medical", "type": "expense"},
		{"category_id": "cat_entertain", "name": "Entretenimiento", "color": "#F7B731", "icon": "game-controller", "type": "expense"},
		{"category_id": "cat_clothes", "name": "Ropa", "color": "#DDA0DD", "icon": "shirt", "type": "expense"},
		{"category_id": "cat_education", "name": "Educación", "color": "#54A0FF", "icon": "school", "type": "expense"},
		{"category_id": "cat_services", "name": "Servicios", "color": "#F7DC6F", "icon": "flash", "type": "expense"},
		{"category_id": "cat_other_exp", "name": "Otros", "color": "#BDC3C7", "icon": "ellipsis-horizontal", "type": "expense"},
		{"category_id": "cat_salary", "name": "Salario", "color": "#27AE60", "icon": "cash", "type": "income"},
		{"category_id": "cat_freelance", "name": "Freelance", "color": "#2ECC71", "icon": "laptop", "type": "income"},
		{"category_id": "cat_invest", "name": "Inversión", "color": "#16A085", "icon": "trending-up", "type": "income"},
		{"category_id": "cat_gift", "name": "Regalo", "color": "#8E44AD", "icon": "gift", "type": "income"},
		{"category_id": "cat_other_inc", "name": "Otros ingresos", "color": "#1ABC9C", "icon": "add-circle", "type": "income"},
	}
	_ = a.Store.MergeRows(ctx, "account", accounts, uid)
	_ = a.Store.MergeRows(ctx, "category", categories, uid)
}
