package server

import (
	"encoding/json"
	"net/http"
)

// Rutas de los probes. Se responden en el borde (api/index.go y cmd/server),
// nunca dentro del router: un probe no puede depender del servicio que chequea.
const (
	HealthPath = "/api/health"
	ReadyPath  = "/api/ready"
)

// Ready indica si el handler quedó inicializado, SIN disparar la
// inicialización. Un probe de salud no debe provocar un cold start.
func Ready() bool {
	mu.Lock()
	defer mu.Unlock()
	return handler != nil
}

// HealthHandler responde el probe de liveness: el proceso está vivo y contesta
// HTTP. Devuelve siempre 200 a propósito, para que un incidente de base de
// datos siga visible como "vivo pero no listo" en el monitoreo, en vez de
// desaparecer del panel a la vez que se rompe la app.
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeHealth(w, http.StatusOK, "ok")
	})
}

// ReadyHandler responde el probe de readiness: 503 mientras el handler no esté
// inicializado, para que un balanceador deje de mandar tráfico hacia una
// instancia que solo puede devolver errores.
func ReadyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if !Ready() {
			writeHealth(w, http.StatusServiceUnavailable, "degraded")
			return
		}
		writeHealth(w, http.StatusOK, "ok")
	})
}

// WithProbes responde /api/health y /api/ready antes de delegar en h. Lo usa el
// servidor standalone, donde el router no está expuesto al borde.
func WithProbes(h http.Handler) http.Handler {
	health, ready := HealthHandler(), ReadyHandler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case HealthPath:
			health.ServeHTTP(w, r)
			return
		case ReadyPath:
			ready.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func writeHealth(w http.ResponseWriter, status int, state string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": state})
}
