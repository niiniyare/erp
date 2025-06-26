package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

// TestDB provides database utilities for testing
type TestDB struct {
	Pool *pgxpool.Pool
	DB   *sql.DB
}

// NewTestDB creates a new test database connection
func NewTestDB(t *testing.T) *TestDB {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://admin:admin@localhost:5432/ledger_test?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	return &TestDB{
		Pool: pool,
		DB:   db,
	}
}

// Close closes the database connections
func (tdb *TestDB) Close() {
	if tdb.Pool != nil {
		tdb.Pool.Close()
	}
	if tdb.DB != nil {
		tdb.DB.Close()
	}
}

// Truncate truncates specified tables for test cleanup
func (tdb *TestDB) Truncate(tables ...string) error {
	if len(tables) == 0 {
		return nil
	}

	query := "TRUNCATE TABLE " + tables[0]
	for i := 1; i < len(tables); i++ {
		query += ", " + tables[i]
	}
	query += " RESTART IDENTITY CASCADE"

	_, err := tdb.DB.Exec(query)
	return err
}

// RunInTransaction runs a function within a database transaction
func (tdb *TestDB) RunInTransaction(fn func(*sql.Tx) error) error {
	tx, err := tdb.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}