// Package server construye el handler HTTP una única vez (pool + migraciones).
package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"chiro/pkg/app"
	"chiro/pkg/auth"
	"chiro/pkg/config"
	"chiro/pkg/migrate"
	"chiro/pkg/store"
)

var (
	once    sync.Once
	handler http.Handler
	initErr error
)

// Handler devuelve el router HTTP (inicializa pool + migraciones una sola vez).
func Handler() (http.Handler, error) {
	once.Do(func() {
		cfg, err := config.Load(true)
		if err != nil {
			initErr = err
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		pool, err := buildPool(ctx, cfg)
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
		authMgr := auth.NewManager(cfg.JWTSecret, cfg.JWT_TTL_Hours)
		handler = withSPA(app.New(st, authMgr).Handler(cfg), cfg)
	})
	return handler, initErr
}

// buildPool crea un pgxpool con tuning según la URL (directa o pooler Supabase).
// Detecta ?pgbouncer=true para activar el modo compatible con transaction
// pooling (prepared statements por conexión en vez de por statement).
func buildPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if cfg.DBMaxConns > 0 {
		pcfg.MaxConns = cfg.DBMaxConns
	}
	if cfg.DBMinConns > 0 {
		pcfg.MinConns = cfg.DBMinConns
	}
	// Pooler Supabase / PgBouncer en transaction mode: deshabilita prepared
	// statements nombrados (no son compatibles con el proxy de conexión).
	if containsQueryParam(cfg.DatabaseURL, "pgbouncer=true") {
		pcfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}
	return pgxpool.NewWithConfig(ctx, pcfg)
}

func containsQueryParam(url, want string) bool {
	i := strings.Index(url, "?")
	if i < 0 {
		return false
	}
	for _, kv := range strings.Split(url[i+1:], "&") {
		if strings.EqualFold(kv, want) {
			return true
		}
	}
	return false
}
