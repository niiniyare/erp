package pgx

import (
	"context"

	pgxlib "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"awo.so/awo/tx"
)

// pgConn wraps either a pool connection or a transaction and implements tx.Conn.
// This type is internal to the pgx driver; nothing outside contrib/pgx sees it.
type pgConn struct {
	pool *pgxpool.Pool
	txn  pgxlib.Tx // non-nil when inside a transaction
}

func (c *pgConn) InTx() bool { return c.txn != nil }

// Querier returns the underlying pgx querier — either the active transaction
// or the pool (for auto-commit queries).
func (c *pgConn) Querier() pgxlib.Row {
	panic("use QueryRow/Exec directly — see pgConn.pool or pgConn.txn")
}

// connFromContext extracts the *pgConn from ctx, or returns a pool-backed
// connection if none is present.
func connFromContext(ctx context.Context, pool *pgxpool.Pool) *pgConn {
	if c, ok := tx.ConnFromContext(ctx).(*pgConn); ok {
		return c
	}
	return &pgConn{pool: pool}
}

// execer provides a pgx-compatible Exec / QueryRow / Query interface that
// works regardless of whether we are inside a transaction.
type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgxlib.Row
	Query(ctx context.Context, sql string, args ...any) (pgxlib.Rows, error)
	SendBatch(ctx context.Context, b *pgxlib.Batch) pgxlib.BatchResults
}

func (c *pgConn) db() execer {
	if c.txn != nil {
		return c.txn
	}
	return c.pool
}

// setTenantContext calls the stored procedure that sets the RLS session variable.
// Must be called inside the transaction BEFORE any DML on tenant-scoped tables.
func setTenantContext(ctx context.Context, e execer, tenantID string) error {
	_, err := e.Exec(ctx, "SELECT set_tenant_context($1)", tenantID)
	return err
}
