// Package server construye el handler HTTP una única vez (pool + migraciones).
package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"chiro/internal/app"
	"chiro/internal/auth"
	"chiro/internal/config"
	"chiro/internal/migrate"
	"chiro/internal/store"
)

var (
	once    sync.Once
	handler http.Handler
	initErr error
)

// Handler devuelve el router HTTP (inicializa pool + migraciones una sola vez).
func Handler() (http.Handler, error) {
	once.Do(func() {
		cfg := config.Load()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			initErr = err
			return
		}
		if err := pool.Ping(ctx); err != nil {
			initErr = err
			return
		}
		if err := migrate.Apply(ctx, pool); err != nil {
			initErr = err
			return
		}

		st := store.New(pool)
		authMgr := auth.NewManager(cfg.JWTSecret)
		handler = withSPA(app.New(st, authMgr).Handler(cfg.CORSOrigins), cfg)
	})
	return handler, initErr
}
