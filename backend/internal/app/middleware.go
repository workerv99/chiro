package app

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"chiro/internal/auth"
	"chiro/internal/config"
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
				if allowAll && !allowed {
					// origen * con auth es superficie: rechazar
					allowed = false
				}
				if allowed {
					if allowAll {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					} else {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Vary", "Origin")
					}
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
		// HSTS: solo si el request es HTTPS (Vercel/proxies setean X-Forwarded-Proto).
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
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

// requireAdmin exige rol admin leyendo de la DB (no del JWT) para que un
// demote se aplique inmediatamente, igual que requireActive.
func (a *App) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var role string
		err := a.Store.Pool().QueryRow(r.Context(),
			`SELECT role FROM users WHERE user_id=$1`, auth.ContextUser(r.Context())).Scan(&role)
		if err != nil || role != "admin" {
			writeErr(w, http.StatusForbidden, "se requiere rol admin")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── Rate limiting por IP ──────────────────────────────────────────────────────

// ipLimiter limita requests por IP con un bucket golang.org/x/time/rate.
// Implementación minimalista: token bucket por IP con limpieza periódica.
type ipLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // requests por minuto
	interval time.Duration // ventana de llenado
	last     time.Time
}

type bucket struct {
	tokens  int
	updated time.Time
}

func newIPLimiter(perMin int) *ipLimiter {
	il := &ipLimiter{
		buckets:  map[string]*bucket{},
		rate:     perMin,
		interval: time.Minute,
	}
	go il.gc()
	return il
}

func (l *ipLimiter) gc() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		l.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for k, b := range l.buckets {
			if b.updated.Before(cutoff) {
				delete(l.buckets, k)
			}
		}
		l.mu.Unlock()
	}
}

// allow devuelve true si la IP puede hacer un request; en caso contrario
// false. Repone 1 token cada interval/rate.
func (l *ipLimiter) allow(ip string) bool {
	if l.rate <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.buckets[ip]
	if !ok {
		l.buckets[ip] = &bucket{tokens: l.rate - 1, updated: now}
		return true
	}
	elapsed := now.Sub(b.updated)
	refill := int(elapsed / l.interval) * l.rate
	if refill > 0 {
		b.tokens = min2(l.rate, b.tokens+refill)
		b.updated = now
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if h := r.Header.Get("X-Forwarded-For"); h != "" {
			parts := strings.Split(h, ",")
			return strings.TrimSpace(parts[0])
		}
		if h := r.Header.Get("X-Real-IP"); h != "" {
			return strings.TrimSpace(h)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// rateLimit aplica el limiter solo a /api/auth/* (login, register).
// Devuelve 429 cuando se excede.
func rateLimit(cfg config.Config) func(http.Handler) http.Handler {
	if cfg.RateLimitPerMin <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	lim := newIPLimiter(cfg.RateLimitPerMin)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/auth/") {
				ip := clientIP(r, cfg.TrustedProxies)
				if !lim.allow(ip) {
					w.Header().Set("Retry-After", "60")
					writeErr(w, http.StatusTooManyRequests, "demasiadas solicitudes, intenta más tarde")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
