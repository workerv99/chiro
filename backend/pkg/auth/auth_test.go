package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueParseSidRoundTrip(t *testing.T) {
	m := NewManager("secret", 24)
	tok, err := m.Issue("usr_1", "a@b.c", "admin", "ses_abc")
	if err != nil {
		t.Fatal(err)
	}
	id, role, sid, err := m.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	if id != "usr_1" || role != "admin" || sid != "ses_abc" {
		t.Fatalf("got %q %q %q", id, role, sid)
	}
	c, err := m.ParseClaims(tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.SessionID != "ses_abc" {
		t.Fatalf("ParseClaims sid = %q", c.SessionID)
	}
}

func signRaw(t *testing.T, m *Manager, c Claims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestParseLegacyTokenWithoutSid(t *testing.T) {
	m := NewManager("secret", 24)
	tok := signRaw(t, m, Claims{Email: "a@b.c", RegisteredClaims: jwt.RegisteredClaims{
		Subject: "usr_1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}})
	id, role, sid, err := m.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	if id != "usr_1" || role != "user" || sid != "" {
		t.Fatalf("got %q %q %q", id, role, sid)
	}
}

func TestExpiredToken(t *testing.T) {
	m := NewManager("secret", 24)
	tok := signRaw(t, m, Claims{Email: "a@b.c", SessionID: "ses_x", RegisteredClaims: jwt.RegisteredClaims{
		Subject: "usr_1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}})
	if _, _, _, err := m.Parse(tok); err == nil {
		t.Fatal("Parse must reject expired token")
	}
	c, err := m.ParseClaims(tok)
	if err != nil {
		t.Fatalf("ParseClaims must accept expired token: %v", err)
	}
	if c.SessionID != "ses_x" || c.Subject != "usr_1" {
		t.Fatalf("claims = %+v", c)
	}
}

func TestParseRejectsBadSignature(t *testing.T) {
	tok, _ := NewManager("one", 24).Issue("u", "e", "user", "s")
	if _, _, _, err := NewManager("two", 24).Parse(tok); err == nil {
		t.Fatal("expected signature error")
	}
}
