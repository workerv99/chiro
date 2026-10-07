// Entrada de la API en Vercel (serverless Go).
package api

import (
	"log"
	"net/http"

	"chiro/pkg/app"
	"chiro/pkg/config"
	"chiro/pkg/server"
)

// serverHandler es un punto de inyección para los tests.
var serverHandler = server.Handler

func Handler(w http.ResponseWriter, r *http.Request) {
	// CORS va PRIMERO, antes de tocar server.Handler(): si la inicialización
	// falla, la respuesta de error también necesita las cabeceras. Sin esto el
	// navegador ve un 500 sin Access-Control-Allow-Origin y lo reporta como
	// "blocked by CORS policy", enviando a diagnosticar la configuración de CORS
	// cuando el problema real es la base de datos.
	app.SetCORS(w, r, config.CORSOriginsFromEnv())

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h, err := serverHandler()
	if err != nil {
		// El error real (hostnames de DB, usuario, pool_size) se queda en el log:
		// filtrarlo al cliente expone topología interna. 503 y no 500 porque la
		// causa es transitoria y el cliente tiene sentido reintentar.
		log.Printf("chiro: handler no disponible [%s %s]: %v", r.Method, r.URL.Path, err)
		http.Error(w, "servicio no disponible, reintenta en unos segundos", http.StatusServiceUnavailable)
		return
	}
	h.ServeHTTP(w, r)
}
