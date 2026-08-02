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
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// ctxUser es lo que el middleware guarda en el contexto.
type ctxUser struct {
	id   string
	role string
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
func (m *Manager) Issue(userID, email, role string) (string, error) {
	claims := Claims{
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse valida un token y devuelve el user_id y el rol.
func (m *Manager) Parse(tokenStr string) (id, role string, err error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inesperado")
		}
		return m.secret, nil
	})
	if err != nil {
		return "", "", err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return "", "", errors.New("token inválido")
	}
	if claims.Role == "" {
		claims.Role = "user"
	}
	return claims.Subject, claims.Role, nil
}

// ParseClaims extrae claims de un token sin validar expiración (para refresh).
// Solo valida la firma.
func (m *Manager) ParseClaims(tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inesperado")
		}
		return m.secret, nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}

// ContextUser devuelve el user_id guardado por Middleware.
func ContextUser(ctx context.Context) string {
	if v, ok := ctx.Value(userKey).(*ctxUser); ok {
		return v.id
	}
	return ""
}

// ContextRole devuelve el rol guardado por Middleware.
func ContextRole(ctx context.Context) string {
	if v, ok := ctx.Value(userKey).(*ctxUser); ok {
		return v.role
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
		uid, role, err := m.Parse(strings.TrimPrefix(authz, "Bearer "))
		if err != nil || uid == "" {
			http.Error(w, `{"error":"sesión inválida"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, &ctxUser{id: uid, role: role})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
