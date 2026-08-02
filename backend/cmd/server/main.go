// Servidor HTTP standalone. Uso: DATABASE_URL=... go run ./cmd/server
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chiro/internal/config"
	"chiro/internal/server"
)

func main() {
	cfg, err := config.Load(true)
	if err != nil {
		log.Fatalf("chiro: %v", err)
	}
	h, err := server.Handler()
	if err != nil {
		log.Fatalf("chiro: error de arranque: %v", err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("chiro api escuchando en :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("chiro: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
