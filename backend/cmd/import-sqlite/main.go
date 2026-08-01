// Importa un archivo SQLite (gastos.db o un backup) a Postgres bajo un usuario.
//
// Uso:
//
//	# Importar bajo un usuario existente:
//	DATABASE_URL=postgres://... go run ./cmd/import-sqlite -db backup.db -user-id usr_xxx
//
//	# O crear el usuario a la vez:
//	DATABASE_URL=postgres://... go run ./cmd/import-sqlite \
//	  -db backup.db -email user@dominio -password secreta -name "Nombre"
//
// Copia las 12 tablas sincronizables respetando updated_at/deleted originales
// (el merge LWW del store se encarga de no pisar filas más nuevas). Valida que
// las fechas sean YYYY-MM-DD reales: si una fila trae una fecha inválida,
// aborta indicando la fila para corregirla en el .db y volver a ejecutar.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"chiro/internal/auth"
	"chiro/internal/config"
	"chiro/internal/store"
	"chiro/internal/svc"
)

func main() {
	dbPath := flag.String("db", "", "ruta al archivo SQLite (.db)")
	userID := flag.String("user-id", "", "user_id destino (si no se da -email)")
	email := flag.String("email", "", "email del usuario destino (lo crea si no existe)")
	password := flag.String("password", "", "contraseña si se crea el usuario")
	name := flag.String("name", "", "nombre si se crea el usuario")
	flag.Parse()

	if *dbPath == "" {
		log.Fatal("chiro: falta -db")
	}
	if *userID == "" && *email == "" {
		log.Fatal("chiro: falta -user-id o -email")
	}

	ctx := context.Background()
	cfg := config.Load()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("chiro: pool: %v", err)
	}
	defer pool.Close()

	// ── Resolver o crear el usuario destino ────────────────────────────────────
	target := *userID
	if *email != "" {
		em := strings.ToLower(strings.TrimSpace(*email))
		err := pool.QueryRow(ctx, `SELECT user_id FROM users WHERE email=$1`, em).Scan(&target)
		if err == pgx.ErrNoRows {
			if len(*password) < 6 {
				log.Fatal("chiro: -password (≥6 caracteres) es obligatorio al crear usuario")
			}
			hash, err := auth.HashPassword(*password)
			if err != nil {
				log.Fatalf("chiro: hash: %v", err)
			}
			nm := *name
			if nm == "" {
				nm = em
			}
			target = svc.GenID("usr")
			_, err = pool.Exec(ctx,
				`INSERT INTO users (user_id, email, name, password_hash, role, status, created_at)
				 VALUES ($1,$2,$3,$4,'user','active',$5)`,
				target, em, nm, hash, time.Now().UTC().UnixMilli())
			if err != nil {
				log.Fatalf("chiro: crear usuario: %v", err)
			}
			log.Printf("chiro: usuario creado: %s (%s)", em, target)
		} else if err != nil {
			log.Fatalf("chiro: consulta usuario: %v", err)
		}
	}
	if target == "" {
		log.Fatal("chiro: no se pudo resolver el usuario destino")
	}
	// Verificar que el usuario existe (las filas tienen FK a users).
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)`, target).Scan(&exists); err != nil || !exists {
		log.Fatalf("chiro: el usuario %s no existe", target)
	}

	// ── Leer SQLite y fusionar ─────────────────────────────────────────────────
	src, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("chiro: abrir sqlite: %v", err)
	}
	defer src.Close()

	st := store.New(pool)
	total := 0
	for _, t := range store.Tablas {
		rows, err := src.Query(`SELECT * FROM ` + t.Table)
		if err != nil {
			log.Fatalf("chiro: leer %s: %v", t.Table, err)
		}
		cols, err := rows.Columns()
		if err != nil {
			rows.Close()
			log.Fatalf("chiro: columnas %s: %v", t.Table, err)
		}
		recs := make([]map[string]any, 0)
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		for rows.Next() {
			if err := rows.Scan(ptrs...); err != nil {
				rows.Close()
				log.Fatalf("chiro: scan %s: %v", t.Table, err)
			}
			rec := make(map[string]any, len(cols))
			for i, c := range cols {
				switch v := vals[i].(type) {
				case nil:
					rec[c] = nil
				case []byte:
					rec[c] = string(v)
				default:
					rec[c] = v
				}
			}
			if col, val := badDate(rec); col != "" {
				pk := ""
				for _, k := range t.PK {
					pk += fmt.Sprintf(" %s=%v", k, rec[k])
				}
				log.Fatalf("chiro: %s%s: fecha inválida en %q: %q — corrígela en el .db y vuelve a ejecutar", t.Table, pk, col, val)
			}
			recs = append(recs, rec)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			log.Fatalf("chiro: iterar %s: %v", t.Table, err)
		}

		if len(recs) > 0 {
			if err := st.MergeRows(ctx, t.Table, recs, target); err != nil {
				log.Fatalf("chiro: importar %s: %v", t.Table, err)
			}
		}
		fmt.Printf("  %-14s %d\n", t.Table, len(recs))
		total += len(recs)
	}
	log.Printf("chiro: importadas %d filas en %d tablas para %s", total, len(store.Tablas), target)
	_ = os.Stdout.Sync()
}

var dateShape = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// badDate devuelve (columna, valor) del primer valor con forma de fecha que no
// sea una fecha válida YYYY-MM-DD. Devuelve ("", "") si no hay ninguno.
func badDate(rec map[string]any) (string, string) {
	for k, v := range rec {
		s, ok := v.(string)
		if !ok || !dateShape.MatchString(s) {
			continue
		}
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return k, s
		}
	}
	return "", ""
}
