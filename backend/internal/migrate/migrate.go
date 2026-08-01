// Package migrate aplica las migraciones SQL embebidas.
package migrate

import (
	"context"
	"embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var files embed.FS

// Apply ejecuta todas las migraciones dentro de una transacción.
func Apply(ctx context.Context, pool *pgxpool.Pool) error {
	names, err := files.ReadDir("migrations")
	if err != nil {
		return err
	}
	var sqls []string
	for _, f := range names {
		if f.IsDir() || len(f.Name()) < 5 || f.Name()[len(f.Name())-4:] != ".sql" {
			continue
		}
		b, err := files.ReadFile("migrations/" + f.Name())
		if err != nil {
			return err
		}
		sqls = append(sqls, string(b))
	}
	sort.Strings(sqls)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, sql := range sqls {
		if _, err := tx.Exec(ctx, sql); err != nil {
			return fmt.Errorf("migracion %q: %w", short(sql), err)
		}
	}
	return tx.Commit(ctx)
}

func short(sql string) string {
	if len(sql) > 60 {
		return sql[:60] + "..."
	}
	return sql
}
