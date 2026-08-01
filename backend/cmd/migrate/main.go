// Aplica las migraciones. Uso: DATABASE_URL=... go run ./cmd/migrate
package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"chiro/internal/config"
	"chiro/internal/migrate"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("chiro: pool: %v", err)
	}
	defer pool.Close()

	if err := migrate.Apply(ctx, pool); err != nil {
		log.Fatalf("chiro: migraciones: %v", err)
	}
	log.Println("chiro: migraciones aplicadas")
}
