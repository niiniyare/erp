package pgstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EnsureRLSFunction verifies that the set_tenant_context(uuid) stored procedure
// exists in the connected PostgreSQL database and is callable.
//
// This is called during application startup (after the pool is established but
// before any request handling begins). A missing function means RLS isolation
// is broken — the process must not start.
//
// Returns a non-nil error if:
//   - the pg_proc lookup fails
//   - set_tenant_context does not exist
//   - a test invocation with uuid.Nil fails
func EnsureRLSFunction(ctx context.Context, pool *pgxpool.Pool) error {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM   pg_proc p
			JOIN   pg_namespace n ON n.oid = p.pronamespace
			WHERE  p.proname = 'set_tenant_context'
			  AND  n.nspname = 'public'
		)
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("pgstore: check set_tenant_context existence: %w", err)
	}
	if !exists {
		return errors.New(
			"pgstore: set_tenant_context() not found in public schema — " +
				"apply database migrations before starting the server")
	}

	// Smoke-test the function with uuid.Nil so we know it's callable.
	// set_tenant_context should accept any UUID (validation is internal).
	if _, err := pool.Exec(ctx, "SELECT set_tenant_context($1)", "00000000-0000-0000-0000-000000000000"); err != nil {
		return fmt.Errorf("pgstore: set_tenant_context smoke test failed: %w", err)
	}

	return nil
}
