package migration

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"awo.so/framework/db"
)

// Direction controls which migrations to apply.
type Direction int

const (
	Up   Direction = iota // apply all pending migrations
	Down                  // revert the last applied migration
)

// Run applies or reverts framework migrations against dsn.
// dsn must be a valid postgres:// connection string.
func Run(_ context.Context, dsn string, dir Direction) error {
	m, err := newMigrate(dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	var migrateErr error
	switch dir {
	case Up:
		migrateErr = m.Up()
	case Down:
		migrateErr = m.Steps(-1)
	}

	if errors.Is(migrateErr, migrate.ErrNoChange) {
		return nil
	}
	return migrateErr
}

// Version returns the current schema version and whether a migration is dirty.
func Version(_ context.Context, dsn string) (version uint, dirty bool, err error) {
	m, err := newMigrate(dsn)
	if err != nil {
		return 0, false, err
	}
	defer m.Close()

	v, d, vErr := m.Version()
	if errors.Is(vErr, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return v, d, vErr
}

func newMigrate(dsn string) (*migrate.Migrate, error) {
	sub, err := fs.Sub(db.Migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("migration: embed sub: %w", err)
	}
	src, err := iofs.New(sub, ".")
	if err != nil {
		return nil, fmt.Errorf("migration: iofs source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return nil, fmt.Errorf("migration: new: %w", err)
	}
	return m, nil
}
