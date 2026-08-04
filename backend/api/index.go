package api

import (
	"net/http"

	"chiro/internal/server"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	h, err := server.Handler()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.ServeHTTP(w, r)
}
