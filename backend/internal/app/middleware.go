package app

import (
	"net/http"

	"chiro/internal/auth"
)

// CORS permite los orígenes configurados (o todos en desarrollo).
func CORS(origins []string) func(http.Handler) http.Handler {
	allowAll := false
	set := map[string]bool{}
	for _, o := range origins {
		if o == "*" {
			allowAll = true
		}
		set[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				allowed := allowAll || set[origin]
				if allowAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// securityHeaders agrega encabezados mínimos de seguridad.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// requireActive rechaza cuentas deshabilitadas. Verifica el estado en la DB
// en cada petición: un disable aplica aunque el JWT siga vigente.
func (a *App) requireActive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var status string
		err := a.Store.Pool().QueryRow(r.Context(),
			`SELECT status FROM users WHERE user_id=$1`, auth.ContextUser(r.Context())).Scan(&status)
		if err != nil || status != "active" {
			writeErr(w, http.StatusForbidden, "cuenta deshabilitada")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireAdmin exige rol admin (leído del JWT por el middleware de auth).
func (a *App) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.ContextRole(r.Context()) != "admin" {
			writeErr(w, http.StatusForbidden, "se requiere rol admin")
			return
		}
		next.ServeHTTP(w, r)
	})
}
