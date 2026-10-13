package app

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prevOut, prevFlags := log.Writer(), log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})
	return &buf
}

func TestLogRequestError(t *testing.T) {
	buf := captureLog(t)
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)

	logRequestError(r, errors.New("boom"))

	want := "chiro error [POST /api/auth/login]: boom"
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Fatalf("log = %q, want %q", got, want)
	}
}

// writeServerError and logRequestError must share one format.
func TestWriteServerErrorUsesSharedLogFormat(t *testing.T) {
	buf := captureLog(t)
	r := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	w := httptest.NewRecorder()

	writeServerError(w, r, "generic", errors.New("db down"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	want := "chiro error [GET /api/x]: db down"
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Fatalf("log = %q, want %q", got, want)
	}
}
