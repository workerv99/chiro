// Package config lee la configuración desde variables de entorno.
package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	CORSOrigins []string
	WebDist     string
}

func Load() Config {
	c := Config{
		DatabaseURL: env("DATABASE_URL", "postgres://chiro:chiro@localhost:5432/chiro?sslmode=disable"),
		JWTSecret:   env("JWT_SECRET", "dev-only-secret-change-me"),
		Port:        env("PORT", "8080"),
		CORSOrigins: split(env("CORS_ORIGINS", "http://localhost:5173,http://localhost:4173")),
		WebDist:     env("WEB_DIST", "../web/build"),
	}
	return c
}

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
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
