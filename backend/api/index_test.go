package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubServerHandler replaces the lazy init with a stub.
func stubServerHandler(t *testing.T, h http.Handler, err error) {
	t.Helper()
	prev := serverHandler
	serverHandler = func() (http.Handler, error) { return h, err }
	t.Cleanup(func() { serverHandler = prev })
}

func withOrigin(origin string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	return r
}

const prodOrigin = "https://chiro-4rs.pages.dev"

// dbLeakError is the kind of error the production DB returns on pool
// exhaustion. It must never reach the client. Values are deliberately fake
// (RFC 5737 TEST-NET) so no real infrastructure identifiers get committed.
func dbLeakError() error {
	return errors.New("failed to connect to `user=postgres.aaaaaaaaaaaaaaaaaaaaaa database=postgres`: " +
		"203.0.113.42:5432 (aws-0-eu-central-1.pooler.supabase.com): server error: " +
		"FATAL: (EMAXCONNSESSION) max clients reached in session mode - pool_size: 15")
}

// TestInitFailureStillSendsCORS is the regression test for the original symptom:
// a 500 without Access-Control-Allow-Origin made the browser report the outage
// as "blocked by CORS policy".
func TestInitFailureStillSendsCORS(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	stubServerHandler(t, nil, dbLeakError())

	rec := httptest.NewRecorder()
	Handler(rec, withOrigin(prodOrigin))

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != prodOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q (el error debe traer CORS)", got, prodOrigin)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("falta Access-Control-Allow-Methods en la respuesta de error")
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want %q", got, "Origin")
	}
}

// TestInitFailureDoesNotLeakInternals: the DB error must stay in the log.
func TestInitFailureDoesNotLeakInternals(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	stubServerHandler(t, nil, dbLeakError())

	rec := httptest.NewRecorder()
	Handler(rec, withOrigin(prodOrigin))

	body := rec.Body.String()
	leaks := []string{
		"aaaaaaaaaaaaaaaaaaaaaa", // project ref / DB user
		"pooler.supabase.com",    // DB host
		"EMAXCONNSESSION",        // internal SQLSTATE
		"203.0.113.42",           // DB IP
		"pool_size",              // pool internals
	}
	for _, leak := range leaks {
		if strings.Contains(body, leak) {
			t.Errorf("el cuerpo de la respuesta filtra %q:\n%s", leak, body)
		}
	}
}

// TestInitFailureStatus: 503 because the cause is transient, not 500.
func TestInitFailureStatus(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	stubServerHandler(t, nil, dbLeakError())

	rec := httptest.NewRecorder()
	Handler(rec, withOrigin(prodOrigin))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d (503, la causa es transitoria)", rec.Code, http.StatusServiceUnavailable)
	}
}

// TestPreflightSucceedsWithoutInit: the browser's OPTIONS probe must pass even
// when the backend cannot connect, otherwise every request looks like a CORS bug.
func TestPreflightSucceedsWithoutInit(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	stubServerHandler(t, nil, dbLeakError())

	r := httptest.NewRequest(http.MethodOptions, "/api/auth/me", nil)
	r.Header.Set("Origin", prodOrigin)
	r.Header.Set("Access-Control-Request-Method", "GET")

	rec := httptest.NewRecorder()
	Handler(rec, r)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != prodOrigin {
		t.Errorf("preflight Access-Control-Allow-Origin = %q, want %q", got, prodOrigin)
	}
}

// TestSuccessPathDelegates: CORS must not break the normal flow.
func TestSuccessPathDelegates(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	sentinel := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	stubServerHandler(t, sentinel, nil)

	rec := httptest.NewRecorder()
	Handler(rec, withOrigin(prodOrigin))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Errorf("no delegó al handler real, body = %q", rec.Body.String())
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != prodOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, prodOrigin)
	}
}

// TestUnlistedOriginGetsNoACAO: a disallowed origin must not receive the header.
func TestUnlistedOriginGetsNoACAO(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	stubServerHandler(t, nil, dbLeakError())

	rec := httptest.NewRecorder()
	Handler(rec, withOrigin("https://evil.example"))

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want vacío para origen no permitido", got)
	}
}
