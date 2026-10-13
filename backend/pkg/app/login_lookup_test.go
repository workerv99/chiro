package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestLoginLookupFailure(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"unknown user", pgx.ErrNoRows, http.StatusUnauthorized},
		{"wrapped unknown user", fmt.Errorf("scan: %w", pgx.ErrNoRows), http.StatusUnauthorized},
		{"pool exhausted", errors.New("FATAL: (EMAXCONNSESSION) max clients reached"), http.StatusServiceUnavailable},
		{"request cancelled", context.Canceled, http.StatusServiceUnavailable},
		{"deadline exceeded", context.DeadlineExceeded, http.StatusServiceUnavailable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, msg := loginLookupFailure(c.err)
			if status != c.wantStatus {
				t.Fatalf("status = %d, want %d", status, c.wantStatus)
			}
			if msg == "" {
				t.Fatal("empty message")
			}
		})
	}
}

// A database failure must never be reported as wrong credentials.
func TestLoginLookupFailureDoesNotBlameCredentials(t *testing.T) {
	_, noRows := loginLookupFailure(pgx.ErrNoRows)
	_, dbDown := loginLookupFailure(errors.New("connection refused"))
	if noRows == dbDown {
		t.Fatalf("db failure and unknown user share the message %q", noRows)
	}
}
