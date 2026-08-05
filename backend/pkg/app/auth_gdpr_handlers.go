package app

import (
	"net/http"
	"strings"
	"time"

	"chiro/pkg/auth"
	"chiro/pkg/email"
	"chiro/pkg/svc"
)

// handleSendVerification envía un email de verificación al usuario.
func (a *App) handleSendVerification(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())

	var emailAddr, name string
	err := a.Store.Pool().QueryRow(r.Context(),
		`SELECT email, name FROM users WHERE user_id=$1`, uid).Scan(&emailAddr, &name)
	if err != nil {
		writeErr(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	// Generar token de verificación.
	token := svc.GenID("verify")
	expires := time.Now().Add(24 * time.Hour)

	_, err = a.Store.Pool().Exec(r.Context(),
		`INSERT INTO verification_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET token=$2, expires_at=$3`,
		uid, token, expires)
	if err != nil {
		writeServerError(w, r, "error al generar token", err)
		return
	}

	// Enviar email.
	verifyURL := "https://chiro.app/verify?token=" + token
	if err := email.SendVerifyEmail(emailAddr, name, verifyURL); err != nil {
		// No fallar si el email no está configurado.
		writeJSON(w, http.StatusOK, map[string]string{"status": "pending", "message": "email no configurado"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

// handleVerifyEmail confirma la verificación del email.
func (a *App) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeErr(w, http.StatusBadRequest, "token requerido")
		return
	}

	var userID string
	var expires time.Time
	err := a.Store.Pool().QueryRow(r.Context(),
		`SELECT user_id, expires_at FROM verification_tokens WHERE token=$1`, token).Scan(&userID, &expires)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "token inválido")
		return
	}

	if time.Now().After(expires) {
		writeErr(w, http.StatusBadRequest, "token expirado")
		return
	}

	// Marcar email como verificado.
	_, _ = a.Store.Pool().Exec(r.Context(),
		`UPDATE users SET email_verified=true WHERE user_id=$1`, userID)
	_, _ = a.Store.Pool().Exec(r.Context(),
		`DELETE FROM verification_tokens WHERE token=$1`, token)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleForgotPassword envía un email con link de restablecimiento.
func (a *App) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	if !readJSON(w, r, &in) {
		return
	}

	emailAddr := strings.ToLower(strings.TrimSpace(in.Email))
	if emailAddr == "" {
		writeErr(w, http.StatusBadRequest, "email requerido")
		return
	}

	var userID, name string
	err := a.Store.Pool().QueryRow(r.Context(),
		`SELECT user_id, name FROM users WHERE email=$1`, emailAddr).Scan(&userID, &name)
	if err != nil {
		// No revelar si el email existe.
		writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
		return
	}

	// Generar token de reset.
	token := svc.GenID("reset")
	expires := time.Now().Add(1 * time.Hour)

	_, _ = a.Store.Pool().Exec(r.Context(),
		`INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET token=$2, expires_at=$3`,
		userID, token, expires)

	resetURL := "https://chiro.app/reset-password?token=" + token
	_ = email.SendPasswordReset(emailAddr, name, resetURL)

	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

// handleResetPassword establece una nueva contraseña con el token.
func (a *App) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &in) {
		return
	}

	if len(in.Password) < 6 {
		writeErr(w, http.StatusBadRequest, "contraseña de al menos 6 caracteres")
		return
	}

	var userID string
	var expires time.Time
	err := a.Store.Pool().QueryRow(r.Context(),
		`SELECT user_id, expires_at FROM password_reset_tokens WHERE token=$1`, in.Token).Scan(&userID, &expires)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "token inválido")
		return
	}

	if time.Now().After(expires) {
		writeErr(w, http.StatusBadRequest, "token expirado")
		return
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		writeServerError(w, r, "error", err)
		return
	}

	_, _ = a.Store.Pool().Exec(r.Context(),
		`UPDATE users SET password_hash=$2 WHERE user_id=$1`, userID, hash)
	_, _ = a.Store.Pool().Exec(r.Context(),
		`DELETE FROM password_reset_tokens WHERE token=$1`, in.Token)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
