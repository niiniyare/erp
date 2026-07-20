package migration

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Runner executes framework migrations against a PostgreSQL database.
// It wraps golang-migrate and uses the registered module sources as the
// migration source, eliminating manual SQL file management.
type Runner struct {
	dbURL string
}

// NewRunner creates a Runner that will apply migrations to the database
// identified by dbURL (standard PostgreSQL DSN or connection string).
func NewRunner(dbURL string) *Runner {
	return &Runner{dbURL: dbURL}
}

// Up applies all pending migrations in dependency order.
// Returns nil if no migrations are pending (ErrNoChange is suppressed).
func (r *Runner) Up() error {
	m, err := r.newMigrate()
	if err != nil {
		return err
	}
	defer r.close(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	v, _, _ := m.Version()
	slog.Info("migrations applied", "version", v)
	return nil
}

// Steps applies n migrations (positive = up, negative = down).
func (r *Runner) Steps(n int) error {
	m, err := r.newMigrate()
	if err != nil {
		return err
	}
	defer r.close(m)

	if err := m.Steps(n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate steps(%d): %w", n, err)
	}
	return nil
}

// Version returns the current applied migration version and dirty state.
func (r *Runner) Version() (uint, bool, error) {
	m, err := r.newMigrate()
	if err != nil {
		return 0, false, err
	}
	defer r.close(m)
	return m.Version()
}

// Force sets the migration version without running any SQL. Use to recover
// from a dirty state after manual intervention.
func (r *Runner) Force(version int) error {
	m, err := r.newMigrate()
	if err != nil {
		return err
	}
	defer r.close(m)
	return m.Force(version)
}

func (r *Runner) newMigrate() (*migrate.Migrate, error) {
	vfs := BuildFS()
	d, err := iofs.New(vfs, ".")
	if err != nil {
		return nil, fmt.Errorf("migration: build iofs source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, r.dbURL)
	if err != nil {
		return nil, fmt.Errorf("migration: init runner: %w", err)
	}
	return m, nil
}

func (r *Runner) close(m *migrate.Migrate) {
	src, db := m.Close()
	if src != nil {
		slog.Debug("migration source close error", "err", src)
	}
	if db != nil {
		slog.Debug("migration db close error", "err", db)
	}
}
