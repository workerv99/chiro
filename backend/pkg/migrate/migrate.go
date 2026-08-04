// Package migrate aplica las migraciones SQL embebidas con tracking de versión.
package migrate

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var files embed.FS

// Apply ejecuta solo las migraciones que no se han aplicado aún.
// Crea la tabla schema_migrations (si no existe) y registra cada migración
// aplicada con su nombre y timestamp.
func Apply(ctx context.Context, pool *pgxpool.Pool) error {
	// 1) Asegurar que la tabla de tracking existe.
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("crear schema_migrations: %w", err)
	}

	// 2) Obtener las migraciones ya aplicadas.
	applied := map[string]bool{}
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("leer schema_migrations: %w", err)
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	// 3) Listar archivos de migración embebidos.
	entries, err := files.ReadDir("migrations")
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}

	// 4) Aplicar solo las nuevas, en orden.
	for _, name := range names {
		if applied[name] {
			continue
		}
		b, err := files.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("leer %s: %w", name, err)
		}
		sql := string(b)

		// Ejecutar en una transacción: migración + registro.
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, sql); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("migracion %s: %w", short(sql), err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`, name); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("registrar %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", name, err)
		}
	}
	return nil
}

func short(sql string) string {
	if len(sql) > 60 {
		return sql[:60] + "..."
	}
	return sql
}
