package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"awo.so/internal/shared"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ------------------------------------------------------------------------------------------------
// Test infrastructure
// ------------------------------------------------------------------------------------------------

// testDBMS holds a store and helpers for a single test run.
type testDBMS struct {
	store    Store
	pool     *pgxpool.Pool
	tenantID uuid.UUID // a seeded ACTIVE tenant
}

// setupTestDB connects to TEST_DATABASE_URL and bootstraps the minimal schema
// needed for tenant context tests. The schema is created in a fresh schema
// named after the test and torn down in t.Cleanup.
//
// Skip the test if TEST_DATABASE_URL is not set — CI sets it, local devs can opt in.
func setupTestDB(t *testing.T) *testDBMS {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://admin:admin@localhost:5432/ledger?sslmode=disable"
		t.Skip("TEST_DATABASE_URL not set, skipping database tests")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err, "connect to test database")

	schemaName := fmt.Sprintf("test_%s_%d", sanitiseName(t.Name()), time.Now().UnixNano())

	// Bootstrap schema + functions + tenants table in an isolated schema.
	_, err = pool.Exec(ctx, bootstrapSQL(schemaName))
	require.NoError(t, err, "bootstrap test schema")

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName),
		)
		pool.Close()
	})

	// Set search_path so the store finds the right schema.
	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	cfg.ConnConfig.RuntimeParams["search_path"] = schemaName + ",pg_catalog"

	scopedPool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(scopedPool.Close)

	store := NewStore(scopedPool)

	// Seed one ACTIVE tenant and one SUSPENDED tenant.
	activeTenantID := uuid.New()
	_, err = scopedPool.Exec(ctx,
		`INSERT INTO tenants (id, slug, name, email, "Status")
		 VALUES ($1, $2, $3, $4, 'ACTIVE')`,
		activeTenantID,
		fmt.Sprintf("active-tenant-%s", activeTenantID),
		fmt.Sprintf("Active Tenant %s", activeTenantID),
		"active@example.com",
	)
	require.NoError(t, err, "seed active tenant")

	return &testDBMS{
		store:    store,
		pool:     scopedPool,
		tenantID: activeTenantID,
	}
}

// ctxWithTenant returns a context carrying the given tenant ID via the
// shared package (mirrors what the auth middleware does in production).
func ctxWithTenant(tenantID uuid.UUID) context.Context {
	return shared.WithTenantID(context.Background(), tenantID)
}

// ------------------------------------------------------------------------------------------------
// Unit tests — no database required
// ------------------------------------------------------------------------------------------------

// TestGetTenantIDFromContext_Unit tests the pure context extraction logic
// without touching the database. We call SetTenantContextFromCtx with a
// context that has/doesn't have a tenant ID and assert the right error path.
//
// These tests work without a database because SetTenantContextFromCtx's first
// action is to read from the context — it fails before hitting the pool.
func TestGetTenantIDFromContext_Unit(t *testing.T) {
	t.Run("returns error when tenant ID not in context", func(t *testing.T) {
		// Use a real store backed by a nil pool — the test must fail at the
		// context-read step before any pool call is made.
		store := NewStore(nil)
		err := store.SetTenantContextFromCtx(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tenant ID not found in context")
	})

	t.Run("WithTenantFromCtx propagates missing tenant error", func(t *testing.T) {
		store := NewStore(nil)
		err := store.WithTenantFromCtx(context.Background(), func(_ context.Context, _ Store) error {
			t.Fatal("fn must not be called when tenant ID is missing")
			return nil
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tenant ID not found in context")
	})

	t.Run("BeginTxWithTenantFromCtx propagates missing tenant error", func(t *testing.T) {
		store := NewStore(nil)
		tx, s, err := store.BeginTxWithTenantFromCtx(context.Background())
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.Nil(t, s)
		assert.Contains(t, err.Error(), "tenant ID not found in context")
	})
}

// ------------------------------------------------------------------------------------------------
// Integration tests — tenant context (set / clear / validate)
// ------------------------------------------------------------------------------------------------

func TestSetTenantContextFromCtx(t *testing.T) {
	tdb := setupTestDB(t)
	ctx := ctxWithTenant(tdb.tenantID)

	t.Run("sets context for active tenant", func(t *testing.T) {
		err := tdb.store.SetTenantContextFromCtx(ctx)
		require.NoError(t, err)

		// Verify the session variable is now set.
		var gotID uuid.UUID
		err = tdb.pool.QueryRow(ctx, "SELECT current_tenant_id()").Scan(&gotID)
		require.NoError(t, err)
		assert.Equal(t, tdb.tenantID, gotID)
	})

	t.Run("rejects non-existent tenant", func(t *testing.T) {
		ctx := ctxWithTenant(uuid.New()) // random UUID — does not exist
		err := tdb.store.SetTenantContextFromCtx(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tenant not found")
	})

	t.Run("rejects soft-deleted tenant", func(t *testing.T) {
		deletedID := seedTenant(t, tdb.pool, "ACTIVE")
		softDelete(t, tdb.pool, deletedID)

		ctx := ctxWithTenant(deletedID)
		err := tdb.store.SetTenantContextFromCtx(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tenant not found or has been deleted")
	})

	t.Run("rejects suspended tenant", func(t *testing.T) {
		suspendedID := seedTenant(t, tdb.pool, "SUSPENDED")

		ctx := ctxWithTenant(suspendedID)
		err := tdb.store.SetTenantContextFromCtx(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not ACTIVE")
	})
}

// Deprecated variant — same contract, different call path.
func TestSetTenantContext_Deprecated(t *testing.T) {
	tdb := setupTestDB(t)
	ctx := context.Background()

	t.Run("sets context for active tenant", func(t *testing.T) {
		err := tdb.store.SetTenantContext(ctx, tdb.tenantID)
		require.NoError(t, err)
	})

	t.Run("rejects non-existent tenant", func(t *testing.T) {
		err := tdb.store.SetTenantContext(ctx, uuid.New())
		require.Error(t, err)
	})
}

func TestValidateTenantContext(t *testing.T) {
	tdb := setupTestDB(t)

	t.Run("returns tenant UUID when context is set and tenant is active", func(t *testing.T) {
		// set_tenant_context is transaction-local so we need a transaction.
		err := tdb.store.WithTenantFromCtx(ctxWithTenant(tdb.tenantID), func(ctx context.Context, s Store) error {
			gotID, err := s.ValidateTenantContext(ctx)
			require.NoError(t, err)
			assert.Equal(t, tdb.tenantID, gotID)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("errors when no context is set", func(t *testing.T) {
		// Call on the raw pool with no tenant set — validate_tenant_context
		// should raise because current_tenant_id() returns NULL.
		_, err := tdb.store.ValidateTenantContext(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no tenant context is set")
	})

	t.Run("errors when tenant becomes suspended mid-transaction", func(t *testing.T) {
		volatileID := seedTenant(t, tdb.pool, "ACTIVE")
		ctx := ctxWithTenant(volatileID)

		err := tdb.store.WithTenantFromCtx(ctx, func(ctx context.Context, s Store) error {
			// Suspend the tenant from outside this transaction.
			_, execErr := tdb.pool.Exec(context.Background(),
				`UPDATE tenants SET "Status" = 'SUSPENDED' WHERE id = $1`, volatileID,
			)
			require.NoError(t, execErr)

			_, valErr := s.ValidateTenantContext(ctx)
			return valErr
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not ACTIVE")
	})
}

// ------------------------------------------------------------------------------------------------
// Integration tests — WithTenantFromCtx
// ------------------------------------------------------------------------------------------------

func TestWithTenantFromCtx(t *testing.T) {
	tdb := setupTestDB(t)

	t.Run("fn executes with tenant context and commits", func(t *testing.T) {
		var capturedID uuid.UUID
		err := tdb.store.WithTenantFromCtx(ctxWithTenant(tdb.tenantID), func(ctx context.Context, s Store) error {
			return tdb.pool.QueryRow(ctx, "SELECT current_tenant_id()").Scan(&capturedID)
		})
		require.NoError(t, err)
		assert.Equal(t, tdb.tenantID, capturedID)
	})

	t.Run("rolls back when fn returns an error", func(t *testing.T) {
		sentinel := errors.New("deliberate failure")
		var fnCalled bool

		err := tdb.store.WithTenantFromCtx(ctxWithTenant(tdb.tenantID), func(_ context.Context, _ Store) error {
			fnCalled = true
			return sentinel
		})

		assert.True(t, fnCalled)
		assert.ErrorIs(t, err, sentinel)
	})

	t.Run("tenant context is cleared after transaction ends", func(t *testing.T) {
		_ = tdb.store.WithTenantFromCtx(ctxWithTenant(tdb.tenantID), func(_ context.Context, _ Store) error {
			return nil
		})

		// After the transaction, the session variable should be gone
		// (transaction-local scope — cleared on COMMIT).
		var gotID *uuid.UUID
		err := tdb.pool.QueryRow(context.Background(), "SELECT current_tenant_id()").Scan(&gotID)
		require.NoError(t, err)
		assert.Nil(t, gotID, "tenant context must be cleared after transaction")
	})

	t.Run("rejects non-existent tenant before fn is called", func(t *testing.T) {
		ctx := ctxWithTenant(uuid.New())
		err := tdb.store.WithTenantFromCtx(ctx, func(_ context.Context, _ Store) error {
			t.Fatal("fn must not be called for invalid tenant")
			return nil
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tenant not found")
	})

	// Deprecated variant.
	t.Run("WithTenant deprecated variant behaves identically", func(t *testing.T) {
		var capturedID uuid.UUID
		err := tdb.store.WithTenant(context.Background(), tdb.tenantID, func(ctx context.Context, _ Store) error {
			return tdb.pool.QueryRow(ctx, "SELECT current_tenant_id()").Scan(&capturedID)
		})
		require.NoError(t, err)
		assert.Equal(t, tdb.tenantID, capturedID)
	})
}

// ------------------------------------------------------------------------------------------------
// Integration tests — BeginTxWithTenantFromCtx
// ------------------------------------------------------------------------------------------------

func TestBeginTxWithTenantFromCtx(t *testing.T) {
	tdb := setupTestDB(t)

	t.Run("returns tx and store with tenant context set", func(t *testing.T) {
		ctx := ctxWithTenant(tdb.tenantID)
		tx, s, err := tdb.store.BeginTxWithTenantFromCtx(ctx)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.NotNil(t, s)

		var gotID uuid.UUID
		scanErr := tdb.pool.QueryRow(ctx, "SELECT current_tenant_id()").Scan(&gotID)
		require.NoError(t, scanErr)
		assert.Equal(t, tdb.tenantID, gotID)

		require.NoError(t, tx.Commit(ctx))
	})

	t.Run("explicit rollback clears transaction", func(t *testing.T) {
		ctx := ctxWithTenant(tdb.tenantID)
		tx, _, err := tdb.store.BeginTxWithTenantFromCtx(ctx)
		require.NoError(t, err)

		require.NoError(t, tx.Rollback(ctx))

		// After rollback, tenant context must be cleared.
		var gotID *uuid.UUID
		err = tdb.pool.QueryRow(context.Background(), "SELECT current_tenant_id()").Scan(&gotID)
		require.NoError(t, err)
		assert.Nil(t, gotID)
	})

	t.Run("returned store satisfies TxStore interface", func(t *testing.T) {
		ctx := ctxWithTenant(tdb.tenantID)
		tx, s, err := tdb.store.BeginTxWithTenantFromCtx(ctx)
		require.NoError(t, err)

		txStore, ok := s.(TxStore)
		assert.True(t, ok, "returned store must implement TxStore")
		assert.NotNil(t, txStore.GetTx())

		_ = tx.Rollback(ctx)
	})

	t.Run("rejects non-existent tenant and rolls back", func(t *testing.T) {
		ctx := ctxWithTenant(uuid.New())
		tx, s, err := tdb.store.BeginTxWithTenantFromCtx(ctx)
		require.Error(t, err)
		assert.Nil(t, tx)
		assert.Nil(t, s)
		assert.Contains(t, err.Error(), "tenant not found")
	})

	// Deprecated variant.
	t.Run("BeginTxWithTenant deprecated variant works", func(t *testing.T) {
		tx, s, err := tdb.store.BeginTxWithTenant(context.Background(), tdb.tenantID)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.NotNil(t, s)
		_ = tx.Rollback(context.Background())
	})
}

// ------------------------------------------------------------------------------------------------
// Integration tests — WithTx / WithTxOptions
// ------------------------------------------------------------------------------------------------

func TestWithTx(t *testing.T) {
	tdb := setupTestDB(t)

	t.Run("commits on success", func(t *testing.T) {
		var called bool
		err := tdb.store.WithTx(context.Background(), func(_ context.Context, _ Store) error {
			called = true
			return nil
		})
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("rolls back and returns error when fn fails", func(t *testing.T) {
		sentinel := errors.New("fn error")
		err := tdb.store.WithTx(context.Background(), func(_ context.Context, _ Store) error {
			return sentinel
		})
		assert.ErrorIs(t, err, sentinel)
	})

	t.Run("nested WithTenantFromCtx inside WithTx works", func(t *testing.T) {
		ctx := ctxWithTenant(tdb.tenantID)
		err := tdb.store.WithTx(context.Background(), func(_ context.Context, s Store) error {
			return s.WithTenantFromCtx(ctx, func(ctx context.Context, ts Store) error {
				var gotID uuid.UUID
				return tdb.pool.QueryRow(ctx, "SELECT current_tenant_id()").Scan(&gotID)
			})
		})
		require.NoError(t, err)
	})
}

func TestWithTxOptions(t *testing.T) {
	tdb := setupTestDB(t)

	t.Run("read-only transaction rejects writes", func(t *testing.T) {
		opts := pgx.TxOptions{AccessMode: pgx.ReadOnly}
		err := tdb.store.WithTxOptions(context.Background(), opts, func(ctx context.Context, _ Store) error {
			_, execErr := tdb.pool.Exec(ctx,
				`INSERT INTO tenants (id, slug, name, email, "Status")
				 VALUES ($1, $2, $3, $4, 'ACTIVE')`,
				uuid.New(), "readonly-test", "ReadOnly Tenant", "ro@example.com",
			)
			return execErr
		})
		require.Error(t, err, "write inside read-only tx must fail")
	})

	t.Run("serializable isolation commits cleanly on non-conflicting work", func(t *testing.T) {
		opts := pgx.TxOptions{IsoLevel: pgx.Serializable}
		err := tdb.store.WithTxOptions(context.Background(), opts, func(_ context.Context, _ Store) error {
			return nil
		})
		require.NoError(t, err)
	})
}

// ------------------------------------------------------------------------------------------------
// Integration tests — TxStore explicit control
// ------------------------------------------------------------------------------------------------

func TestTxStore_CommitAndRollback(t *testing.T) {
	tdb := setupTestDB(t)

	t.Run("Commit makes changes visible", func(t *testing.T) {
		ctx := ctxWithTenant(tdb.tenantID)
		newID := uuid.New()

		tx, s, err := tdb.store.BeginTxWithTenantFromCtx(ctx)
		require.NoError(t, err)

		txStore, ok := s.(TxStore)
		require.True(t, ok)

		// Write inside the transaction.
		_, err = tdb.pool.Exec(ctx,
			`INSERT INTO tenants (id, slug, name, email, "Status")
			 VALUES ($1, $2, $3, $4, 'ACTIVE')`,
			newID,
			fmt.Sprintf("commit-test-%s", newID),
			fmt.Sprintf("Commit Test %s", newID),
			"commit@example.com",
		)
		require.NoError(t, err)

		require.NoError(t, txStore.Commit(ctx))
		_ = tx // tx already committed

		// Should be visible after commit.
		var count int
		err = tdb.pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM tenants WHERE id = $1", newID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Rollback discards changes", func(t *testing.T) {
		ctx := ctxWithTenant(tdb.tenantID)
		newID := uuid.New()

		tx, s, err := tdb.store.BeginTxWithTenantFromCtx(ctx)
		require.NoError(t, err)

		txStore, ok := s.(TxStore)
		require.True(t, ok)

		_, err = tdb.pool.Exec(ctx,
			`INSERT INTO tenants (id, slug, name, email, "Status")
			 VALUES ($1, $2, $3, $4, 'ACTIVE')`,
			newID,
			fmt.Sprintf("rollback-test-%s", newID),
			fmt.Sprintf("Rollback Test %s", newID),
			"rollback@example.com",
		)
		require.NoError(t, err)

		require.NoError(t, txStore.Rollback(ctx))
		_ = tx

		// Should NOT be visible after rollback.
		var count int
		err = tdb.pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM tenants WHERE id = $1", newID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("double Rollback is a no-op", func(t *testing.T) {
		ctx := ctxWithTenant(tdb.tenantID)
		_, s, err := tdb.store.BeginTxWithTenantFromCtx(ctx)
		require.NoError(t, err)

		txStore := s.(TxStore)
		require.NoError(t, txStore.Rollback(ctx))
		require.NoError(t, txStore.Rollback(ctx), "second rollback must be a no-op, not an error")
	})
}

// ------------------------------------------------------------------------------------------------
// Integration tests — HealthCheck
// ------------------------------------------------------------------------------------------------

func TestHealthCheck(t *testing.T) {
	tdb := setupTestDB(t)

	t.Run("returns nil on healthy database", func(t *testing.T) {
		err := tdb.store.HealthCheck(context.Background())
		require.NoError(t, err)
	})
}

// ------------------------------------------------------------------------------------------------
// Helpers
// ------------------------------------------------------------------------------------------------

// seedTenant inserts a tenant with the given status and returns its UUID.
func seedTenant(t *testing.T, pool *pgxpool.Pool, status string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO tenants (id, slug, name, email, "Status")
		 VALUES ($1, $2, $3, $4, $5)`,
		id,
		fmt.Sprintf("tenant-%s", id),
		fmt.Sprintf("Tenant %s", id),
		fmt.Sprintf("%s@example.com", id),
		status,
	)
	require.NoError(t, err, "seedTenant")
	return id
}

// softDelete marks a tenant as soft-deleted.
func softDelete(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"UPDATE tenants SET deleted_at = NOW() WHERE id = $1", id,
	)
	require.NoError(t, err, "softDelete")
}

// sanitiseName returns a postgres-safe identifier fragment from a test name.
func sanitiseName(name string) string {
	out := make([]byte, 0, len(name))
	for i := 0; i < len(name) && i < 30; i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			out = append(out, c)
		} else if c >= 'A' && c <= 'Z' {
			out = append(out, c+32)
		} else {
			out = append(out, '_')
		}
	}
	return string(out)
}

// bootstrapSQL returns the SQL to set up an isolated test schema containing
// the minimal tenants table and tenant context functions.
// This mirrors the real migrations — any drift here is a test bug.
func bootstrapSQL(schema string) string {
	return fmt.Sprintf(`
CREATE SCHEMA IF NOT EXISTS %[1]s;
SET search_path = %[1]s, pg_catalog;

-- Minimal tenants table (mirrors migration 000053).
CREATE TABLE IF NOT EXISTS %[1]s.tenants (
  id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  slug       VARCHAR(50) NOT NULL UNIQUE
             CHECK (slug ~* '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'),
  name       VARCHAR(255) NOT NULL UNIQUE,
  email      VARCHAR(255) NOT NULL,
  "Status"   VARCHAR(20)  NOT NULL DEFAULT 'PENDING'
             CHECK ("Status" IN ('ACTIVE', 'SUSPENDED', 'PENDING', 'ARCHIVED')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

-- current_tenant_id() — mirrors migration 000055.
CREATE OR REPLACE FUNCTION %[1]s.current_tenant_id()
  RETURNS UUID LANGUAGE plpgsql STABLE SECURITY INVOKER
AS $$
BEGIN
  RETURN COALESCE(NULLIF(current_setting('app.current_tenant_id', TRUE), ''), NULL)::UUID;
EXCEPTION
  WHEN invalid_text_representation THEN
    RAISE WARNING 'current_tenant_id(): invalid UUID in session variable — returning NULL';
    RETURN NULL;
END;
$$;

-- set_tenant_context(UUID) — mirrors migration 000055.
CREATE OR REPLACE FUNCTION %[1]s.set_tenant_context(p_tenant_id UUID)
  RETURNS VOID LANGUAGE plpgsql SECURITY DEFINER
  SET search_path = pg_catalog, %[1]s
AS $$
DECLARE
  v_status TEXT;
BEGIN
  SELECT "Status" INTO v_status
    FROM %[1]s.tenants
   WHERE id = p_tenant_id AND deleted_at IS NULL;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'set_tenant_context: tenant not found or has been deleted — id: %%', p_tenant_id
      USING ERRCODE = 'no_data_found';
  END IF;

  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION 'set_tenant_context: tenant is not ACTIVE — id: %%, current status: %%', p_tenant_id, v_status
      USING ERRCODE = 'check_violation';
  END IF;

  PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, TRUE);
END;
$$;

-- clear_tenant_context() — mirrors migration 000055.
CREATE OR REPLACE FUNCTION %[1]s.clear_tenant_context()
  RETURNS VOID LANGUAGE plpgsql SECURITY INVOKER
AS $$
BEGIN
  PERFORM set_config('app.current_tenant_id', '', TRUE);
END;
$$;

-- validate_tenant_context() — mirrors migration 000055.
CREATE OR REPLACE FUNCTION %[1]s.validate_tenant_context()
  RETURNS UUID LANGUAGE plpgsql STABLE SECURITY DEFINER
  SET search_path = pg_catalog, %[1]s
AS $$
DECLARE
  v_tid    UUID;
  v_status TEXT;
BEGIN
  v_tid := %[1]s.current_tenant_id();

  IF v_tid IS NULL THEN
    RAISE EXCEPTION 'validate_tenant_context: no tenant context is set — call set_tenant_context() before proceeding'
      USING ERRCODE = 'no_data_found';
  END IF;

  SELECT "Status" INTO v_status
    FROM %[1]s.tenants
   WHERE id = v_tid AND deleted_at IS NULL;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'validate_tenant_context: tenant no longer exists or has been soft-deleted — id: %%', v_tid
      USING ERRCODE = 'no_data_found';
  END IF;

  IF v_status <> 'ACTIVE' THEN
    RAISE EXCEPTION 'validate_tenant_context: tenant context is stale — id: %%, current status: %%', v_tid, v_status
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN v_tid;
END;
$$;
`, schema)
}
