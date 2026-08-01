// Package auth maneja registro, login, JWT y middleware de autenticación.
package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type ctxKey string

const userKey ctxKey = "chiro_user"

// Claims es el payload del JWT.
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// Manager firma y valida JWTs.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret), ttl: 30 * 24 * time.Hour}
}

// HashPassword genera el hash bcrypt de una contraseña.
func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// Issue firma un token para un usuario.
func (m *Manager) Issue(userID, email string) (string, error) {
	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse valida un token y devuelve el user_id.
func (m *Manager) Parse(tokenStr string) (string, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inesperado")
		}
		return m.secret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return "", errors.New("token inválido")
	}
	return claims.Subject, nil
}

// ContextUser devuelve el user_id guardado por Middleware.
func ContextUser(ctx context.Context) string {
	if v, ok := ctx.Value(userKey).(string); ok {
		return v
	}
	return ""
}

// Middleware exige un Bearer token válido.
func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			http.Error(w, `{"error":"no autorizado"}`, http.StatusUnauthorized)
			return
		}
		uid, err := m.Parse(strings.TrimPrefix(authz, "Bearer "))
		if err != nil || uid == "" {
			http.Error(w, `{"error":"sesión inválida"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
