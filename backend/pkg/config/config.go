// Package config lee la configuración desde variables de entorno.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL     string
	JWTSecret       string
	Port            string
	CORSOrigins     []string
	WebDist         string
	DBMaxConns      int32
	DBMinConns      int32
	JWT_TTL_Hours   int    // TTL del access token (horas, default 24)
	RequireSecret   bool   // si true, JWT_SECRET debe ser seguro
	TrustedProxies  bool   // si true, se confía en X-Forwarded-For para IP cliente
	RateLimitPerMin int     // requests/min por IP para endpoints sensibles
}

// PlanLimits define los límites por plan.
type PlanLimits struct {
	MaxExpensesPerMonth int
	MaxAccounts         int
	MaxLoans            int
	MaxCategories       int
}

var FreeLimits = PlanLimits{
	MaxExpensesPerMonth: 50,
	MaxAccounts:         3,
	MaxLoans:            10,
	MaxCategories:       10,
}

var ProLimits = PlanLimits{
	MaxExpensesPerMonth: -1, // ilimitado
	MaxAccounts:         -1,
	MaxLoans:            -1,
	MaxCategories:       -1,
}

const devJWTSecret = "dev-only-secret-change-me"

// Load lee la configuración del entorno. requireSecret=true falla si JWT_SECRET
// es el default o está vacío: en producción un secreto débil permite forjar
// tokens. Las herramientas de migración / admin-create / import-sqlite pasan
// requireSecret=false porque no firman JWTs.
//
// DATABASE_URL acepta la URL completa de Postgres, incluyendo:
//   - Supabase directo:  postgresql://postgres.[ref]:[pw]@db.[ref].supabase.co:5432/postgres?sslmode=require
//   - Supabase pooler:   postgresql://postgres.[ref]:[pw]@aws-0-[region].pooler.supabase.com:6543/postgres?sslmode=require&pgbouncer=true
// pgxpool parsea la URL y respeta sslmode/connect_timeout/etc.
func Load(requireSecret bool) (Config, error) {
	secret := env("JWT_SECRET", "")
	if requireSecret && (secret == "" || secret == devJWTSecret) {
		return Config{}, errors.New("JWT_SECRET requerido y no puede ser el default 'dev-only-secret-change-me'")
	}
	url := env("DATABASE_URL", "")
	if url == "" {
		return Config{}, errors.New("DATABASE_URL requerido")
	}
	if requireSecret && strings.Contains(url, "sslmode=disable") {
		return Config{}, errors.New("DATABASE_URL no puede usar sslmode=disable en producción")
	}

	c := Config{
		DatabaseURL:     url,
		JWTSecret:       secret,
		Port:            env("PORT", "8080"),
		CORSOrigins:     split(env("CORS_ORIGINS", "http://localhost:5173,http://localhost:4173")),
		WebDist:         env("WEB_DIST", "../web/build"),
		DBMaxConns:      int32(envInt("DB_MAX_CONNS", 10)),
		DBMinConns:      int32(envInt("DB_MIN_CONNS", 1)),
		JWT_TTL_Hours:   envInt("JWT_TTL_HOURS", 24),
		RequireSecret:   requireSecret,
		TrustedProxies:  envBool("TRUSTED_PROXIES", true),
		RateLimitPerMin: envInt("RATE_LIMIT_PER_MIN", 20),
	}
	return c, nil
}

// String devuelve una vista segura para logs (oculta secretos).
func (c Config) String() string {
	return fmt.Sprintf("port=%s db_max=%d db_min=%d trusted_proxies=%v rate_limit=%d origins=%d",
		c.Port, c.DBMaxConns, c.DBMinConns, c.TrustedProxies, c.RateLimitPerMin, len(c.CORSOrigins))
}

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	var n int
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func envBool(k string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v == "1" || strings.EqualFold(v, "true")
}

func split(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
