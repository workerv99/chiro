// Sirve el SPA (build de Vite) con fallback a index.html para rutas del router.
package server

import (
	"net/http"
	"path"
	"strings"

	"chiro/internal/config"
)

func withSPA(api http.Handler, cfg config.Config) http.Handler {
	dist := http.Dir(cfg.WebDist)
	fileServer := http.FileServer(dist)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)
			return
		}

		if r.URL.Path != "/" {
			f, err := dist.Open(path.Clean(strings.TrimPrefix(r.URL.Path, "/")))
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		// SPA fallback: todo lo que no sea un asset existente sirve index.html.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
