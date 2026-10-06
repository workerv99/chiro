package app

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"chiro/pkg/auth"
	"chiro/pkg/svc"
)

const (
	maxUserAgentLen = 255
	// touchInterval evita una escritura por request al actualizar last_used_at.
	touchInterval = time.Minute
)

// truncateUserAgent recorta espacios y limita a maxUserAgentLen runas.
func truncateUserAgent(ua string) string {
	ua = strings.TrimSpace(ua)
	r := []rune(ua)
	if len(r) > maxUserAgentLen {
		return string(r[:maxUserAgentLen])
	}
	return ua
}

// shouldTouchSession indica si last_used_at está lo bastante viejo para actualizarlo.
func shouldTouchSession(lastUsed, now time.Time) bool {
	return now.Sub(lastUsed) >= touchInterval
}

// sessionIsCurrent marca la sesión del token actual (los tokens legacy no tienen sid).
func sessionIsCurrent(sessionID, currentSID string) bool {
	return currentSID != "" && sessionID == currentSID
}

// createSession registra una sesión de dispositivo para el usuario y devuelve su id.
func (a *App) createSession(r *http.Request, userID string) (string, error) {
	sid := svc.GenID("ses")
	_, err := a.Store.Pool().Exec(r.Context(),
		`INSERT INTO sessions (session_id, user_id, user_agent, ip_address) VALUES ($1,$2,$3,$4)`,
		sid, userID, truncateUserAgent(r.UserAgent()), clientIP(r, a.trustProxy))
	return sid, err
}

// handleListSessions lista las sesiones activas del usuario.
func (a *App) handleListSessions(w http.ResponseWriter, r *http.Request) {
	uid := auth.ContextUser(r.Context())
	currentSID := auth.ContextSID(r.Context())
	rows, err := a.Store.Pool().Query(r.Context(),
		`SELECT session_id, user_agent, ip_address, created_at, last_used_at
		   FROM sessions WHERE user_id=$1 AND revoked_at IS NULL
		  ORDER BY last_used_at DESC`, uid)
	if err != nil {
		writeServerError(w, r, "error al listar sesiones", err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, ua, ip string
		var created, last time.Time
		if err := rows.Scan(&id, &ua, &ip, &created, &last); err != nil {
			writeServerError(w, r, "error al listar sesiones", err)
			return
		}
		out = append(out, map[string]any{
			"session_id":   id,
			"user_agent":   ua,
			"ip_address":   ip,
			"created_at":   created.UTC().Format(time.RFC3339),
			"last_used_at": last.UTC().Format(time.RFC3339),
			"current":      sessionIsCurrent(id, currentSID),
		})
	}
	if err := rows.Err(); err != nil {
		writeServerError(w, r, "error al listar sesiones", err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleRevokeSession revoca una sesión propia; 404 si no existe o es ajena.
func (a *App) handleRevokeSession(w http.ResponseWriter, r *http.Request) {
	res, err := a.Store.Pool().Exec(r.Context(),
		`UPDATE sessions SET revoked_at=now()
		  WHERE session_id=$1 AND user_id=$2 AND revoked_at IS NULL`,
		chi.URLParam(r, "id"), auth.ContextUser(r.Context()))
	if err != nil {
		writeServerError(w, r, "error al revocar la sesión", err)
		return
	}
	if res.RowsAffected() == 0 {
		writeErr(w, http.StatusNotFound, "sesión no encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleLogout revoca la sesión del token actual (no-op para tokens legacy).
func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if sid := auth.ContextSID(r.Context()); sid != "" {
		if _, err := a.Store.Pool().Exec(r.Context(),
			`UPDATE sessions SET revoked_at=now() WHERE session_id=$1 AND user_id=$2 AND revoked_at IS NULL`,
			sid, auth.ContextUser(r.Context())); err != nil {
			writeServerError(w, r, "error al cerrar sesión", err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// refreshSession valida la sesión de un token a refrescar. Con sid, la sesión
// debe existir, ser del usuario y no estar revocada; se actualiza last_used_at.
// Sin sid (token legacy) se crea una sesión nueva. Devuelve el sid a usar.
func (a *App) refreshSession(r *http.Request, userID, sid string) (string, error) {
	if sid == "" {
		return a.createSession(r, userID)
	}
	var revoked bool
	err := a.Store.Pool().QueryRow(r.Context(),
		`UPDATE sessions SET last_used_at=now()
		  WHERE session_id=$1 AND user_id=$2
		RETURNING revoked_at IS NOT NULL`, sid, userID).Scan(&revoked)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errSessionInvalid
	}
	if err != nil {
		return "", err
	}
	if revoked {
		return "", errSessionInvalid
	}
	return sid, nil
}

var errSessionInvalid = errors.New("sesión revocada o inexistente")
