// Package testutil provides shared helpers for integration tests.
//
// All helpers are safe to call from t.Parallel tests and register t.Cleanup
// automatically so no state leaks between tests.
package testutil

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NewPostgres starts an isolated PostgreSQL 17 container, applies all up
// migrations from store/migrations/, and returns a ready pool.
//
// Skipped automatically when -short is set (fast CI gate).
// Container and pool are cleaned up via t.Cleanup.
func NewPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping Postgres integration test (-short)")
	}

	ctx := context.Background()

	pg, err := tcpg.Run(ctx,
		"postgres:17-alpine",
		tcpg.WithDatabase("llmstatus_test"),
		tcpg.WithUsername("llmstatus"),
		tcpg.WithPassword("llmstatus"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("testutil.NewPostgres: start container: %v", err)
	}
	t.Cleanup(func() {
		if terr := pg.Terminate(context.Background()); terr != nil {
			t.Logf("testutil.NewPostgres: terminate container: %v", terr)
		}
	})

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("testutil.NewPostgres: connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("testutil.NewPostgres: new pool: %v", err)
	}
	t.Cleanup(pool.Close)

	applyMigrations(t, ctx, pool)

	return pool
}

// applyMigrations executes all up-migration SQL files (*.sql, excluding
// *.down.sql) from store/migrations/ in lexicographic order.
func applyMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Clean(
		filepath.Join(filepath.Dir(thisFile), "../../store/migrations"),
	)

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("applyMigrations: readdir %s: %v", migrationsDir, err)
	}

	var files []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasSuffix(name, ".sql") && !strings.HasSuffix(name, ".down.sql") {
			files = append(files, filepath.Join(migrationsDir, name))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		sql, rerr := os.ReadFile(f)
		if rerr != nil {
			t.Fatalf("applyMigrations: readfile %s: %v", f, rerr)
		}
		if _, eerr := pool.Exec(ctx, string(sql)); eerr != nil {
			t.Fatalf("applyMigrations: exec %s: %v", f, eerr)
		}
	}
}
