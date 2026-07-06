// Command migrate runs database migrations using golang-migrate.
//
// Usage:
//
//	migrate -dir ./db/migration -db $DATABASE_URL up
//	migrate -dir ./db/migration -db $DATABASE_URL down 1
//	migrate -dir ./db/migration -db $DATABASE_URL version
//	migrate -dir ./db/migration -db $DATABASE_URL force <version>
//
// This command runs as a separate process from the API server. It is designed
// to run in CI (before deployment) and is never invoked automatically at
// server startup — auto-migration is explicitly prohibited.
//
// The process exits with code 0 on success, non-zero on failure.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dir := flag.String("dir", "./db/migration", "migration files directory")
	db := flag.String("db", os.Getenv("DATABASE_URL"), "database connection URL")
	flag.Parse()

	if *db == "" {
		log.Fatal("migrate: -db or DATABASE_URL is required")
	}

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("migrate: command required: up | down [N] | version | force <version>")
	}

	m, err := newMigrate(*db, *dir)
	if err != nil {
		log.Fatalf("migrate: init: %v", err)
	}
	defer func() { src, db := m.Close(); _ = src; _ = db }()

	cmd := args[0]
	switch cmd {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("migrate: up complete")

	case "down":
		n := 1
		if len(args) > 1 {
			v, err := strconv.Atoi(args[1])
			if err != nil {
				log.Fatalf("migrate: invalid N for down: %v", err)
			}
			n = v
		}
		if err := m.Steps(-n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("migrate down: %v", err)
		}
		fmt.Printf("migrate: down %d complete\n", n)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("migrate version: %v", err)
		}
		fmt.Printf("version: %d, dirty: %v\n", version, dirty)

	case "force":
		if len(args) < 2 {
			log.Fatal("migrate force: version required")
		}
		v, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("migrate force: invalid version: %v", err)
		}
		if err := m.Force(v); err != nil {
			log.Fatalf("migrate force: %v", err)
		}
		fmt.Printf("migrate: forced to version %d\n", v)

	default:
		log.Fatalf("migrate: unknown command %q", cmd)
	}
}

func newMigrate(dbURL, dir string) (*migrate.Migrate, error) {
	sourceURL := "file://" + dir
	// golang-migrate postgres driver expects the standard PostgreSQL DSN.
	// Prepend "postgres://" if the URL uses "postgresql://" scheme variant.
	return migrate.New(sourceURL, dbURL)
}
