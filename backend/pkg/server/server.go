// Package server construye el handler HTTP una única vez (pool + migraciones).
package server

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"chiro/pkg/app"
	"chiro/pkg/auth"
	"chiro/pkg/config"
	"chiro/pkg/migrate"
	"chiro/pkg/store"
)

// Tunning de la inicialización.
const (
	// initAttempts son los intentos dentro de una misma petición. Saturación de
	// pool o un reinicio de la DB son fallos transitorios: un reintento evita que
	// un cold start se convierta en un 500 permanente.
	initAttempts = 3
	// initTimeout es el presupuesto TOTAL de la inicialización, no por intento.
	// Debe quedar por debajo del maxDuration de la función en la plataforma: si
	// lo excede, Vercel mata la invocación y el reintento nunca ocurre.
	initTimeout = 20 * time.Second
	// initBackoff es la espera base entre intentos (doble en cada vuelta).
	initBackoff = 500 * time.Millisecond
	// initCooldown evita que cada request entrante dispare su propio intento
	// mientras el anterior sigue fallando.
	initCooldown = 5 * time.Second
)

var (
	mu         sync.Mutex
	handler    http.Handler
	initErr    error
	lastFailed time.Time
)

// Handler devuelve el router HTTP, inicializando pool + migraciones la primera
// vez. A diferencia de un sync.Once, un fallo de init NO es terminal: en
// serverless una instancia puede vivir minutos, así que cachear el error la
// dejaba muerta de forma permanente ante un fallo transitorio. Solo el éxito
// se cachea.
func Handler() (http.Handler, error) {
	mu.Lock()
	defer mu.Unlock()

	if handler != nil {
		return handler, nil
	}
	// Dentro del cooldown se devuelve el error cacheado sin volver a golpear la
	// DB en cada request concurrente.
	if initErr != nil && time.Since(lastFailed) < initCooldown {
		return nil, initErr
	}

	h, err := initWithRetry()
	if err != nil {
		initErr, lastFailed = err, time.Now()
		log.Printf("chiro: init falló; nuevo intento tras %s: %v", initCooldown, err)
		return nil, err
	}

	initErr = nil
	handler = h
	return handler, nil
}

// initWithRetry reintenta connectInit con backoff exponencial, siempre dentro
// del presupuesto initTimeout. Solo reintenta fallos transitorios: un error de
// configuración se devuelve de inmediato.
func initWithRetry() (http.Handler, error) {
	cfg, err := config.Load(true)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(initTimeout)
	var errTransient error
	for attempt := 1; attempt <= initAttempts; attempt++ {
		// Cada intento usa el presupuesto restante, nunca el total: así un
		// intento colgado no puede agotar el tiempo de los reintentos.
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}

		h, err := connectInit(cfg, remaining)
		if err == nil {
			return h, nil
		}
		// Un JWT_SECRET débil o una URL inválida no se arreglan reintentando.
		if !isTransient(err) {
			return nil, err
		}

		errTransient = err
		if attempt < initAttempts {
			log.Printf("chiro: init intento %d/%d falló: %v", attempt, initAttempts, err)
			time.Sleep(time.Duration(1<<(attempt-1)) * initBackoff)
		}
	}
	// errTransient solo es nil si el presupuesto se agotó antes del primer
	// intento; devolver (nil, nil) haría que Handler() aceptara un handler nil.
	if errTransient == nil {
		errTransient = errors.New("init: presupuesto agotado antes de intentar")
	}
	return nil, errTransient
}

// connectInit permite simular fallos en los tests.
var connectInit = connect

// connect ejecuta un intento completo de inicialización con el timeout dado.
func connect(cfg config.Config, timeout time.Duration) (http.Handler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pool, err := buildPool(ctx, cfg)
	if err != nil {
		return nil, err
	}
	// El pool debe cerrarse en cualquier fallo posterior: sin esto los reintentos
	// acumularían conexiones y agravarían la saturación que intentamos evitar.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if err := migrate.Apply(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	st := store.New(pool)
	authMgr := auth.NewManager(cfg.JWTSecret, cfg.JWT_TTL_Hours)
	return withSPA(app.New(st, authMgr).Handler(cfg), cfg), nil
}

// isTransient distingue fallos que valen un reintento (saturación de pool, DB
// reiniciándose, timeout de red) de los permanentes.
func isTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return true
	}
	var opErr *net.OpError
	return errors.As(err, &opErr)
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
