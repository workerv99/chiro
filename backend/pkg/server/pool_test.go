package server

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"chiro/pkg/config"
)

func poolCfg(t *testing.T, url string) *pgxpool.Config {
	t.Helper()
	pcfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("ParseConfig(%q): %v", url, err)
	}
	return pcfg
}

// TestMaxConnIdleTimeApplied: the lever that lets an idle serverless instance
// release its session instead of holding one of the few available.
func TestMaxConnIdleTimeApplied(t *testing.T) {
	pcfg := poolCfg(t, "postgres://u:p@localhost:5432/db")
	applyPoolTuning(pcfg, config.Config{DatabaseURL: pcfg.ConnString(), DBMaxIdleSecs: 60})

	if pcfg.MaxConnIdleTime != 60*time.Second {
		t.Errorf("MaxConnIdleTime = %v, want 60s", pcfg.MaxConnIdleTime)
	}
}

// TestMaxConnIdleTimeDefaultUntouched: 0 must leave pgx's own default in place,
// so deploying does not silently change behaviour for existing setups.
func TestMaxConnIdleTimeDefaultUntouched(t *testing.T) {
	pcfg := poolCfg(t, "postgres://u:p@localhost:5432/db")
	before := pcfg.MaxConnIdleTime

	applyPoolTuning(pcfg, config.Config{DatabaseURL: pcfg.ConnString()})

	if pcfg.MaxConnIdleTime != before {
		t.Errorf("MaxConnIdleTime = %v, want el default de pgx %v", pcfg.MaxConnIdleTime, before)
	}
}

func TestConnsTuning(t *testing.T) {
	t.Run("aplica max y min", func(t *testing.T) {
		pcfg := poolCfg(t, "postgres://u:p@localhost:5432/db")
		applyPoolTuning(pcfg, config.Config{
			DatabaseURL: pcfg.ConnString(),
			DBMaxConns:  1,
			DBMinConns:  0,
		})
		if pcfg.MaxConns != 1 {
			t.Errorf("MaxConns = %d, want 1", pcfg.MaxConns)
		}
		// DB_MIN_CONNS=0 debe dejar el default de pgx (0), no fijar 0 explícito
		// ni dejar el 1 previo.
		if pcfg.MinConns != 0 {
			t.Errorf("MinConns = %d, want 0", pcfg.MinConns)
		}
	})

	t.Run("respeta defaults de pgx si no se configura", func(t *testing.T) {
		pcfg := poolCfg(t, "postgres://u:p@localhost:5432/db")
		maxBefore, minBefore := pcfg.MaxConns, pcfg.MinConns

		applyPoolTuning(pcfg, config.Config{DatabaseURL: pcfg.ConnString()})

		if pcfg.MaxConns != maxBefore || pcfg.MinConns != minBefore {
			t.Errorf("MaxConns/MinConns = %d/%d, want %d/%d",
				pcfg.MaxConns, pcfg.MinConns, maxBefore, minBefore)
		}
	})
}

// TestTransactionModeDisablesPreparedStatements guards the PgBouncer
// compatibility flag that makes transaction pooling usable at all.
func TestTransactionModeDisablesPreparedStatements(t *testing.T) {
	t.Run("transaction pooler usa simple protocol", func(t *testing.T) {
		url := "postgres://u:p@pooler:6543/db?pgbouncer=true"
		pcfg := poolCfg(t, url)
		applyPoolTuning(pcfg, config.Config{DatabaseURL: url})

		if pcfg.ConnConfig.DefaultQueryExecMode != pgx.QueryExecModeSimpleProtocol {
			t.Errorf("exec mode = %v, want simple protocol", pcfg.ConnConfig.DefaultQueryExecMode)
		}
	})

	t.Run("session pooler conserva prepared statements", func(t *testing.T) {
		url := "postgres://u:p@pooler:5432/db"
		pcfg := poolCfg(t, url)
		before := pcfg.ConnConfig.DefaultQueryExecMode

		applyPoolTuning(pcfg, config.Config{DatabaseURL: url})

		if pcfg.ConnConfig.DefaultQueryExecMode != before {
			t.Errorf("exec mode = %v, want sin cambios %v", pcfg.ConnConfig.DefaultQueryExecMode, before)
		}
	})
}

// TestContainsQueryParamIsCaseInsensitive: pgbouncer=TRUE must work too, since
// this flag decides whether prepared statements are safe.
func TestContainsQueryParamIsCaseInsensitive(t *testing.T) {
	cases := map[string]bool{
		"postgres://u:p@h:6543/db?pgbouncer=true":                 true,
		"postgres://u:p@h:6543/db?pgbouncer=TRUE":                 true,
		"postgres://u:p@h:6543/db?sslmode=require&pgbouncer=true": true,
		"postgres://u:p@h:5432/db":                                false,
		"postgres://u:p@h:5432/db?pgbouncer=false":                false,
	}
	for url, want := range cases {
		if got := containsQueryParam(url, "pgbouncer=true"); got != want {
			t.Errorf("containsQueryParam(%q) = %v, want %v", url, got, want)
		}
	}
}

// TestSessionModeIdleTimeoutDoesNotBreakTuning: the session-mode mitigation
// (low idle timeout, one connection) must compose with the rest of the tuning.
func TestSessionModeIdleTimeoutDoesNotBreakTuning(t *testing.T) {
	url := "postgres://u:p@aws-0-us-east-1.pooler.supabase.com:5432/postgres?sslmode=require"
	pcfg := poolCfg(t, url)
	applyPoolTuning(pcfg, config.Config{
		DatabaseURL:   url,
		DBMaxConns:    1,
		DBMaxIdleSecs: 60,
	})

	if pcfg.MaxConns != 1 {
		t.Errorf("MaxConns = %d, want 1", pcfg.MaxConns)
	}
	if pcfg.MaxConnIdleTime != 60*time.Second {
		t.Errorf("MaxConnIdleTime = %v, want 60s", pcfg.MaxConnIdleTime)
	}
	if pcfg.ConnConfig.DefaultQueryExecMode != pgx.QueryExecModeCacheStatement {
		t.Errorf("session mode debe conservar prepared statements, exec mode = %v",
			pcfg.ConnConfig.DefaultQueryExecMode)
	}
}
