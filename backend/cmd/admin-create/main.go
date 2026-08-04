// Crea o promueve el primer usuario admin.
//
// Uso:
//
//	ADMIN_EMAIL=admin@dominio ADMIN_PASSWORD=secreta ADMIN_NAME="Admin" \
//	  DATABASE_URL=postgres://... go run ./cmd/admin-create
//
// Si el email ya existe, lo promueve a admin (y actualiza password/name/status).
package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"chiro/pkg/auth"
	"chiro/pkg/config"
	"chiro/pkg/svc"
)

func main() {
	cfg, err := config.Load(false)
	if err != nil {
		log.Fatalf("chiro: %v", err)
	}
	email := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || len(password) < 6 {
		log.Fatal("chiro: ADMIN_EMAIL y ADMIN_PASSWORD (≥6 caracteres) son obligatorios")
	}
	name := os.Getenv("ADMIN_NAME")
	if name == "" {
		name = "Admin"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("chiro: pool: %v", err)
	}
	defer pool.Close()

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("chiro: hash: %v", err)
	}
	now := time.Now().UTC().UnixMilli()

	var uid string
	err = pool.QueryRow(ctx, `SELECT user_id FROM users WHERE email=$1`, email).Scan(&uid)
	switch {
	case err == pgx.ErrNoRows:
		uid = svc.GenID("usr")
		_, err = pool.Exec(ctx,
			`INSERT INTO users (user_id, email, name, password_hash, role, status, created_at)
			 VALUES ($1,$2,$3,$4,'admin','active',$5)`,
			uid, email, name, hash, now)
		if err != nil {
			log.Fatalf("chiro: insert: %v", err)
		}
		log.Printf("chiro: admin creado: %s (%s)", email, uid)
	case err == nil:
		_, err = pool.Exec(ctx,
			`UPDATE users SET role='admin', status='active', name=$1, password_hash=$2 WHERE user_id=$3`,
			name, hash, uid)
		if err != nil {
			log.Fatalf("chiro: update: %v", err)
		}
		log.Printf("chiro: promovido a admin: %s (%s)", email, uid)
	default:
		log.Fatalf("chiro: consulta: %v", err)
	}
}
