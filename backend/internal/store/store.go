// Package store es la capa de datos. Soporta CRUD genérico por tabla con
// soft-delete y UPSERT de fusión (último-en-escribir-gana vía updated_at),
// port de utils/dbSchema.js (SYNC_TABLES + buildMergeStatement).
package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Tabla describe una tabla sincronizable con su(s) columna(s) de clave primaria.
type Tabla struct {
	Table string
	PK    []string
}

// Tablas es el port de SYNC_TABLES del proyecto original.
var Tablas = []Tabla{
	{Table: "account", PK: []string{"account_id"}},
	{Table: "category", PK: []string{"category_id"}},
	{Table: "expense", PK: []string{"expense_id"}},
	{Table: "person", PK: []string{"person_id"}},
	{Table: "loan", PK: []string{"loan_id"}},
	{Table: "payment", PK: []string{"payment_id"}},
	{Table: "installment", PK: []string{"installment_id"}},
	{Table: "budget", PK: []string{"budget_id"}},
	{Table: "piggy_bank", PK: []string{"piggy_bank_id"}},
	{Table: "bill", PK: []string{"bill_id"}},
	{Table: "tag", PK: []string{"tag_id"}},
	{Table: "expense_tag", PK: []string{"expense_id", "tag_id"}},
}

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// ── Utilidades ────────────────────────────────────────────────────────────────

var allowedTables = func() map[string]struct{} {
	m := make(map[string]struct{}, len(Tablas))
	for _, t := range Tablas {
		m[t.Table] = struct{}{}
	}
	return m
}()

func checkTable(t string) (Tabla, error) {
	if _, ok := allowedTables[t]; !ok {
		return Tabla{}, errors.New("tabla no permitida: " + t)
	}
	var pk []string
	for _, tb := range Tablas {
		if tb.Table == t {
			pk = tb.PK
		}
	}
	return Tabla{Table: t, PK: pk}, nil
}

// Normalize devuelve la fila con valores por defecto para el servidor:
// user_id, deleted=0 y updated_at si faltan.
func Normalize(row map[string]any, userID string) map[string]any {
	out := make(map[string]any, len(row)+2)
	for k, v := range row {
		if v == nil {
			continue
		}
		out[k] = v
	}
	if _, ok := out["user_id"]; !ok {
		out["user_id"] = userID
	}
	if _, ok := out["deleted"]; !ok {
		out["deleted"] = 0
	}
	if _, ok := out["updated_at"]; !ok {
		out["updated_at"] = time.Now().UnixMilli()
	}
	return out
}

// ── Upsert de fusión (LWW) ────────────────────────────────────────────────────

// MergeRows inserta filas de un cliente; ante conflicto por PK pisa la fila
// local SOLO si la remota es más nueva (excluded.updated_at > t.updated_at).
func (s *Store) MergeRows(ctx context.Context, table string, rows []map[string]any, userID string) error {
	return s.mergeRows(ctx, s.pool, table, rows, userID)
}

// MergeRowsTx es MergeRows dentro de una transacción ya abierta.
func (s *Store) MergeRowsTx(ctx context.Context, tx pgx.Tx, table string, rows []map[string]any, userID string) error {
	return s.mergeRows(ctx, tx, table, rows, userID)
}

type batchSender interface {
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

func (s *Store) mergeRows(ctx context.Context, sender batchSender, table string, rows []map[string]any, userID string) error {
	t, err := checkTable(table)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	pkSet := map[string]bool{}
	for _, pk := range t.PK {
		pkSet[pk] = true
	}

	// Unión de columnas presentes en las filas NORMALIZADAS (robusto ante
	// campos faltantes y garantiza user_id/updated_at/deleted en el INSERT).
	norm := make([]map[string]any, 0, len(rows))
	colSet := map[string]bool{}
	for _, r := range rows {
		nr := Normalize(r, userID)
		norm = append(norm, nr)
		for k := range nr {
			if k == "user_id" {
				continue
			}
			colSet[k] = true
		}
	}
	cols := make([]string, 0, len(colSet))
	for k := range colSet {
		cols = append(cols, k)
	}
	sortStrings(cols)

	var nonPK []string
	for _, c := range cols {
		if !pkSet[c] {
			nonPK = append(nonPK, c)
		}
	}

	quoted := func(list []string) string {
		out := make([]string, len(list))
		for i, c := range list {
			out[i] = `"` + c + `"`
		}
		return strings.Join(out, ", ")
	}

	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
	}
	setClause := make([]string, len(nonPK))
	for i, c := range nonPK {
		setClause[i] = fmt.Sprintf(`"%s" = EXCLUDED."%s"`, c, c)
	}
	conflict := make([]string, len(t.PK)+1)
	conflict[0] = `"user_id"`
	for i, pk := range t.PK {
		conflict[i+1] = `"` + pk + `"`
	}
	pkRef := `"user_id"`
	for _, pk := range t.PK {
		pkRef += fmt.Sprintf(` AND "%s" = EXCLUDED."%s"`, pk, pk)
	}

	sql := fmt.Sprintf(
		`INSERT INTO "%s" ("user_id", %s) VALUES ($1, %s)
		 ON CONFLICT (%s) DO UPDATE SET %s WHERE EXCLUDED."updated_at" > "%s"."updated_at"`,
		t.Table, quoted(cols), strings.Join(placeholders, ", "),
		strings.Join(conflict, ", "), strings.Join(setClause, ", "), t.Table,
	)

	batch := &pgx.Batch{}
	for _, nr := range norm {
		args := make([]any, 0, len(cols)+1)
		args = append(args, userID)
		for _, c := range cols {
			v, ok := nr[c]
			if !ok {
				v = nil
			}
			args = append(args, v)
		}
		batch.Queue(sql, args...)
	}

	return sender.SendBatch(ctx, batch).Close()
}

// ── Lecturas ──────────────────────────────────────────────────────────────────

// List devuelve las filas activas del usuario.
func (s *Store) List(ctx context.Context, table string, userID string, orderBy string) ([]map[string]any, error) {
	t, err := checkTable(table)
	if err != nil {
		return nil, err
	}
	if orderBy == "" {
		orderBy = t.PK[0]
	}
	sql := fmt.Sprintf(`SELECT * FROM "%s" WHERE "user_id" = $1 AND "deleted" = 0 ORDER BY %s`, t.Table, orderBy)
	return queryRows(ctx, s.pool, sql, userID)
}

// ListDeleted incluye las filas borradas (para sync de tombstones).
func (s *Store) ListDeleted(ctx context.Context, table string, userID string) ([]map[string]any, error) {
	t, err := checkTable(table)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`SELECT * FROM "%s" WHERE "user_id" = $1 AND "deleted" = 1 ORDER BY "updated_at" DESC`, t.Table)
	return queryRows(ctx, s.pool, sql, userID)
}

// ListChanged devuelve filas (vivas o borradas) con updated_at mayor a since.
func (s *Store) ListChanged(ctx context.Context, table string, userID string, since int64) ([]map[string]any, error) {
	t, err := checkTable(table)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`SELECT * FROM "%s" WHERE "user_id" = $1 AND "updated_at" > $2 ORDER BY "updated_at" ASC`, t.Table)
	return queryRows(ctx, s.pool, sql, userID, since)
}

// ListChangedLimit es como ListChanged pero con un limite por tabla (paginación).
func (s *Store) ListChangedLimit(ctx context.Context, table string, userID string, since int64, limit int) ([]map[string]any, error) {
	t, err := checkTable(table)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`SELECT * FROM "%s" WHERE "user_id" = $1 AND "updated_at" > $2 ORDER BY "updated_at" ASC LIMIT $3`, t.Table)
	return queryRows(ctx, s.pool, sql, userID, since, limit)
}

// Get devuelve una fila activa por su clave primaria.
func (s *Store) Get(ctx context.Context, table string, id string, userID string) (map[string]any, error) {
	t, err := checkTable(table)
	if err != nil {
		return nil, err
	}
	pk := t.PK[0]
	sql := fmt.Sprintf(`SELECT * FROM "%s" WHERE "user_id" = $1 AND "%s" = $2 AND "deleted" = 0`, t.Table, pk)
	rows, err := queryRows(ctx, s.pool, sql, userID, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// SoftDelete marca deleted=1 en la fila (y solo si pertenece al usuario).
func (s *Store) SoftDelete(ctx context.Context, table string, id string, userID string) error {
	t, err := checkTable(table)
	if err != nil {
		return err
	}
	pk := t.PK[0]
	sql := fmt.Sprintf(`UPDATE "%s" SET "deleted" = 1, "updated_at" = $3 WHERE "user_id" = $1 AND "%s" = $2 AND "deleted" = 0`, t.Table, pk)
	_, err = s.pool.Exec(ctx, sql, userID, id, time.Now().UnixMilli())
	return err
}

// ── Queries SQL directas ──────────────────────────────────────────────────────

// ExecAll ejecuta varias sentencias dentro de una transacción.
func (s *Store) ExecAll(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// queryRows mapea filas a []map[string]any convirtiendo fechas y numéricos.
func queryRows(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) ([]map[string]any, error) {
	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		fds := rows.FieldDescriptions()
		m := make(map[string]any, len(fds))
		for i, fd := range fds {
			m[fd.Name] = convertVal(fd.DataTypeOID, vals[i])
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func convertVal(oid uint32, v any) any {
	switch oid {
	case pgtype.DateOID:
		if t, ok := v.(time.Time); ok {
			return t.Format("2006-01-02")
		}
	}
	return v
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
