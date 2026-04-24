// Command migrate applies or rolls back numbered SQL migrations.
//
// Usage:
//
//	migrate [-db DSN] <command> [args]
//
// Commands:
//
//	up           Apply all pending migrations (default)
//	down [N]     Roll back N steps (default 1)
//	version      Print current migration version and dirty flag
//	status       Human-readable migration state
//	force N      Force-set version without running SQL (dirty-state recovery)
//
// The database DSN is read from DATABASE_URL or the -db flag. Use the
// pgx5:// scheme, e.g. pgx5://user:pass@host/dbname?sslmode=disable
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	migrations "github.com/llmstatus/llmstatus/store/migrations"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("migrate: ")

	dbURL := flag.String("db", os.Getenv("DATABASE_URL"), "Postgres DSN (pgx5:// scheme)")
	flag.Usage = usage
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("DATABASE_URL is not set — pass -db or export DATABASE_URL")
	}

	cmd := "up"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	m := newMigrator(*dbURL)
	defer closeMigrator(m)

	switch cmd {
	case "up":
		cmdUp(m)
	case "down":
		runDown(m)
	case "version":
		cmdVersion(m)
	case "status":
		cmdStatus(m)
	case "force":
		runForce(m)
	default:
		log.Fatalf("unknown command %q — run migrate -help", cmd)
	}
}

func newMigrator(dbURL string) *migrate.Migrate {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatalf("source: %v", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, dbURL)
	if err != nil {
		log.Fatalf("init: %v", err)
	}
	return m
}

func closeMigrator(m *migrate.Migrate) {
	srcErr, dbErr := m.Close()
	if srcErr != nil {
		log.Printf("close source: %v", srcErr)
	}
	if dbErr != nil {
		log.Printf("close db: %v", dbErr)
	}
}

func runDown(m *migrate.Migrate) {
	n := 1
	if flag.NArg() > 1 {
		var err error
		n, err = strconv.Atoi(flag.Arg(1))
		if err != nil || n < 1 {
			log.Fatalf("down: N must be a positive integer, got %q", flag.Arg(1))
		}
	}
	cmdDown(m, n)
}

func runForce(m *migrate.Migrate) {
	if flag.NArg() < 2 {
		log.Fatal("force: version number required")
	}
	v, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		log.Fatalf("force: invalid version %q", flag.Arg(1))
	}
	cmdForce(m, v)
}

func cmdUp(m *migrate.Migrate) {
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("already up to date")
			return
		}
		log.Fatalf("up: %v", err)
	}
	v, _, _ := m.Version()
	fmt.Printf("up: applied to version %d\n", v)
}

func cmdDown(m *migrate.Migrate, n int) {
	if err := m.Steps(-n); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("nothing to roll back")
			return
		}
		log.Fatalf("down: %v", err)
	}
	v, _, verr := m.Version()
	if errors.Is(verr, migrate.ErrNilVersion) {
		fmt.Println("down: rolled back to clean state (no migrations applied)")
		return
	}
	fmt.Printf("down: rolled back %d step(s), now at version %d\n", n, v)
}

func cmdVersion(m *migrate.Migrate) {
	v, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("version: none (no migrations applied)")
		return
	}
	if err != nil {
		log.Fatalf("version: %v", err)
	}
	fmt.Printf("version: %d  dirty: %v\n", v, dirty)
}

func cmdStatus(m *migrate.Migrate) {
	v, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("status: no migrations applied")
		return
	}
	if err != nil {
		log.Fatalf("status: %v", err)
	}
	if dirty {
		fmt.Printf("status: DIRTY at version %d — run 'migrate force N' to recover\n", v)
		return
	}
	fmt.Printf("status: ok  version: %d\n", v)
}

func cmdForce(m *migrate.Migrate, v int) {
	if err := m.Force(v); err != nil {
		log.Fatalf("force: %v", err)
	}
	fmt.Printf("force: version set to %d (dirty flag cleared)\n", v)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: migrate [-db DSN] <command> [args]

Commands:
  up           Apply all pending migrations (default)
  down [N]     Roll back N steps (default 1)
  version      Print current version and dirty flag
  status       Human-readable migration state
  force N      Force-set version (dirty-state recovery)

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
DATABASE_URL / -db must use the pgx5:// scheme:
  pgx5://user:pass@localhost:5432/dbname?sslmode=disable
`)
}
