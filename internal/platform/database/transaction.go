package database

import (
	"context"
	"database/sql"
)



// transaction implements Transaction interface
type transaction struct {
    tx *sql.Tx
}

// ExecContext executes a query with context
func (t *transaction) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result,error) {
    return t.tx.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows
func (t *transaction) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    return t.tx.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns a single row
func (t *transaction) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
    return t.tx.QueryRowContext(ctx, query, args...)
}

// Commit commits the transaction
func (t *transaction) Commit() error {
    return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *transaction) Rollback() error {
    return t.tx.Rollback()
}
