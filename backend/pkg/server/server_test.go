package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"chiro/pkg/config"
)

// resetForTest limpia el estado global de inicialización entre tests.
func resetForTest() {
	mu.Lock()
	defer mu.Unlock()
	handler = nil
	initErr = nil
	lastFailed = time.Time{}
}

// stubConnect reemplaza connectInit por un stub y lo restaura al terminar.
func stubConnect(t *testing.T, fn func(config.Config, time.Duration) (http.Handler, error)) {
	t.Helper()
	prev := connectInit
	connectInit = fn
	t.Cleanup(func() { connectInit = prev })
}

// validEnv satisface config.Load(true): DATABASE_URL sin sslmode=disable y un
// JWT_SECRET que no sea el default de desarrollo.
func validEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db?sslmode=require")
	t.Setenv("JWT_SECRET", "test-secret-not-the-dev-default")
	t.Setenv("REQUIRE_SECRET", "")
}

// maxConnsErr reproduce la forma del error EMAXCONNSESSION visto en producción:
// un *pgconn.PgError envuelto con %w y luego con errors.Join, que es
// exactamente la cadena que pgconn construye (ConnectError.Unwrap ->
// errors.Join -> perDialConnectError.Unwrap -> PgError). Si errors.As no
// atravesara esa cadena, el retry no se dispararía nunca.
func maxConnsErr() error {
	pgErr := &pgconn.PgError{
		Code:    "53300",
		Message: "max clients reached in session mode - max clients are limited to pool_size: 15",
	}
	return errors.Join(fmt.Errorf(
		"203.0.113.42:5432 (aws-0-eu-central-1.pooler.supabase.com): server error: %w", pgErr))
}

func TestIsTransient(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"EMAXCONNSESSION de Supabase (PgError envuelto)", maxConnsErr(), true},
		{"PgError desnudo", &pgconn.PgError{Code: "53300"}, true},
		{"deadline exceeded", context.DeadlineExceeded, true},
		{"timeout de red", &net.OpError{Op: "dial", Err: errors.New("timeout")}, true},
		{"error de configuracion", errors.New("JWT_SECRET requerido y no puede ser el default"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isTransient(tc.err); got != tc.want {
				t.Errorf("isTransient() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestHandlerRetriesTransientFailure es el test que habría detectado el bug
// original: un fallo transitorio NO debe dejar la instancia muerta.
func TestHandlerRetriesTransientFailure(t *testing.T) {
	resetForTest()
	validEnv(t)

	var calls atomic.Int32
	stubConnect(t, func(config.Config, time.Duration) (http.Handler, error) {
		if calls.Add(1) < initAttempts {
			return nil, maxConnsErr()
		}
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}), nil
	})

	h, err := Handler()
	if err != nil {
		t.Fatalf("Handler() error = %v, want nil (debería reintentar)", err)
	}
	if h == nil {
		t.Fatal("Handler() devolvió nil handler")
	}
	if got := calls.Load(); got != initAttempts {
		t.Errorf("connectInit llamado %d veces, want %d", got, initAttempts)
	}
}

// TestHandlerCachesSuccess: tras un init exitoso no se vuelve a intentar.
func TestHandlerCachesSuccess(t *testing.T) {
	resetForTest()
	validEnv(t)

	var calls atomic.Int32
	stubConnect(t, func(config.Config, time.Duration) (http.Handler, error) {
		calls.Add(1)
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}), nil
	})

	first, err := Handler()
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	for range 5 {
		if _, err := Handler(); err != nil {
			t.Fatalf("Handler() en cache error = %v", err)
		}
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("connectInit llamado %d veces, want 1 (el éxito debe cachearse)", got)
	}
	_ = first
}

// TestHandlerDoesNotRetryPermanentError: un error de configuración no se reintenta.
func TestHandlerDoesNotRetryPermanentError(t *testing.T) {
	resetForTest()
	validEnv(t)

	var calls atomic.Int32
	stubConnect(t, func(config.Config, time.Duration) (http.Handler, error) {
		calls.Add(1)
		return nil, errors.New("JWT_SECRET requerido y no puede ser el default")
	})

	if _, err := Handler(); err == nil {
		t.Fatal("Handler() error = nil, want error permanente")
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("connectInit llamado %d veces, want 1 (no debe reintentar permanentes)", got)
	}
}

// TestHandlerCooldown evita el stampede: tras fallar, los requests siguientes
// devuelven el error cacheado sin volver a golpear la DB.
func TestHandlerCooldown(t *testing.T) {
	resetForTest()
	validEnv(t)

	var calls atomic.Int32
	stubConnect(t, func(config.Config, time.Duration) (http.Handler, error) {
		calls.Add(1)
		return nil, maxConnsErr()
	})

	if _, err := Handler(); err == nil {
		t.Fatal("Handler() error = nil, want error")
	}
	after := calls.Load()

	for range 5 {
		if _, err := Handler(); err == nil {
			t.Fatal("Handler() error = nil, want error cacheado")
		}
	}
	if got := calls.Load(); got != after {
		t.Errorf("connectInit llamado %d veces tras el cooldown, want %d (no debe reintentar)", got, after)
	}
}
