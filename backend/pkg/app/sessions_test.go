package app

import (
	"strings"
	"testing"
	"time"
)

func TestTruncateUserAgent(t *testing.T) {
	if got := truncateUserAgent("  Mozilla/5.0  "); got != "Mozilla/5.0" {
		t.Fatalf("got %q", got)
	}
	long := strings.Repeat("a", maxUserAgentLen+50)
	if got := truncateUserAgent(long); len(got) != maxUserAgentLen {
		t.Fatalf("len = %d", len(got))
	}
	// Multibyte: must not cut a rune in half.
	multi := strings.Repeat("ñ", maxUserAgentLen)
	got := truncateUserAgent(multi)
	if len([]rune(got)) != maxUserAgentLen {
		t.Fatalf("runes = %d", len([]rune(got)))
	}
	if strings.ContainsRune(got, '�') {
		t.Fatal("invalid utf-8 produced")
	}
}

func TestShouldTouchSession(t *testing.T) {
	now := time.Now()
	if shouldTouchSession(now.Add(-10*time.Second), now) {
		t.Fatal("recent use must not write")
	}
	if !shouldTouchSession(now.Add(-2*time.Minute), now) {
		t.Fatal("stale use must write")
	}
}

func TestSessionIsCurrent(t *testing.T) {
	if sessionIsCurrent("ses_a", "") {
		t.Fatal("legacy token has no current session")
	}
	if !sessionIsCurrent("ses_a", "ses_a") || sessionIsCurrent("ses_a", "ses_b") {
		t.Fatal("current flag mismatch")
	}
}
