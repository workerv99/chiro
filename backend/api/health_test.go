package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"chiro/pkg/server"
)

// TestHealthAnswersWhenInitFails is the whole point: the probe must report on
// the very initialization it sits in front of.
func TestHealthAnswersWhenInitFails(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	stubServerHandler(t, nil, dbLeakError())

	for _, path := range []string{server.HealthPath, server.ReadyPath} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Handler(rec, httptest.NewRequest(http.MethodGet, path, nil))

			if rec.Code == http.StatusServiceUnavailable && path == server.HealthPath {
				t.Errorf("liveness devolvió 503; el proceso SÍ está vivo, debe reportarse como tal")
			}
			if rec.Code == 0 {
				t.Fatal("no escribió status")
			}
			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body no es JSON: %q", rec.Body.String())
			}
			if body["status"] == "" {
				t.Errorf("body sin campo status: %q", rec.Body.String())
			}
		})
	}
}

// TestLivenessIsAlwaysOK: liveness must never fail, or monitoring loses the
// instance at the exact moment it starts failing.
func TestLivenessIsAlwaysOK(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	stubServerHandler(t, nil, dbLeakError())

	rec := httptest.NewRecorder()
	Handler(rec, httptest.NewRequest(http.MethodGet, server.HealthPath, nil))

	if rec.Code != http.StatusOK {
		t.Errorf("liveness status = %d, want 200 aunque la DB esté caída", rec.Code)
	}
}

// TestReadinessReportsDegraded: readiness must fail while init has not succeeded,
// so a load balancer stops sending traffic.
func TestReadinessReportsDegraded(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	stubServerHandler(t, nil, dbLeakError())

	rec := httptest.NewRecorder()
	Handler(rec, httptest.NewRequest(http.MethodGet, server.ReadyPath, nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("readiness status = %d, want 503 cuando la init falló", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body no es JSON: %q", rec.Body.String())
	}
	if body["status"] != "degraded" {
		t.Errorf("status = %q, want %q", body["status"], "degraded")
	}
}

// TestProbesCarryCORS: monitoring calling from a browser-like client also needs headers.
func TestProbesCarryCORS(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	stubServerHandler(t, nil, dbLeakError())

	r := httptest.NewRequest(http.MethodGet, server.ReadyPath, nil)
	r.Header.Set("Origin", prodOrigin)

	rec := httptest.NewRecorder()
	Handler(rec, r)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != prodOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, prodOrigin)
	}
}

// TestProbesDoNotLeakInternals: the probe body must stay clean too.
func TestProbesDoNotLeakInternals(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	stubServerHandler(t, nil, dbLeakError())

	for _, path := range []string{server.HealthPath, server.ReadyPath} {
		rec := httptest.NewRecorder()
		Handler(rec, httptest.NewRequest(http.MethodGet, path, nil))

		for _, leak := range []string{"supabase", "EMAXCONNSESSION", "pool_size", "203.0.113.42"} {
			if strings.Contains(rec.Body.String(), leak) {
				t.Errorf("%s filtró %q: %q", path, leak, rec.Body.String())
			}
		}
	}
}

// TestHealthDoesNotTriggerInit: a probe must not cause a cold start. If it did,
// every health check would open a DB connection.
func TestHealthDoesNotTriggerInit(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	var calls int
	prev := serverHandler
	serverHandler = func() (http.Handler, error) {
		calls++
		return nil, dbLeakError()
	}
	t.Cleanup(func() { serverHandler = prev })

	for range 3 {
		rec := httptest.NewRecorder()
		Handler(rec, httptest.NewRequest(http.MethodGet, server.HealthPath, nil))
	}

	if calls != 0 {
		t.Errorf("serverHandler llamado %d veces, want 0 (el probe no debe inicializar)", calls)
	}
}

// TestNormalPathStillDelegates: removing the route from the router must not
// break regular API traffic.
func TestNormalPathStillDelegates(t *testing.T) {
	t.Setenv("CORS_ORIGINS", prodOrigin)
	sentinel := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})
	stubServerHandler(t, sentinel, nil)

	rec := httptest.NewRecorder()
	Handler(rec, httptest.NewRequest(http.MethodGet, "/api/accounts", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != `[]` {
		t.Errorf("body = %q, want %q", rec.Body.String(), `[]`)
	}
}
