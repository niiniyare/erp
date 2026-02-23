package authz

import (
	"context"
	"os"
	"testing"

	casbin "github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/niiniyare/erp/internal/shared/logger"
)

// ---------------------------------------------------------------------------
// noopLogger satisfies logger.Logger for tests without any output.
// ---------------------------------------------------------------------------

type noopLogger struct{}

func (noopLogger) Debug(msg string, fields ...logger.Fields)                             {}
func (noopLogger) Info(msg string, fields ...logger.Fields)                              {}
func (noopLogger) Warn(msg string, fields ...logger.Fields)                              {}
func (noopLogger) Error(msg string, fields ...logger.Fields)                             {}
func (noopLogger) Fatal(msg string, fields ...logger.Fields)                             {}
func (noopLogger) DebugContext(_ context.Context, _ string, _ ...logger.Fields)          {}
func (noopLogger) InfoContext(_ context.Context, _ string, _ ...logger.Fields)           {}
func (noopLogger) WarnContext(_ context.Context, _ string, _ ...logger.Fields)           {}
func (noopLogger) ErrorContext(_ context.Context, _ string, _ ...logger.Fields)          {}
func (noopLogger) WithFields(_ logger.Fields) logger.Logger                              { return noopLogger{} }
func (noopLogger) WithContext(_ context.Context) logger.Logger                           { return noopLogger{} }
func (noopLogger) SetLevel(_ logger.LogLevel)                                            {}
func (noopLogger) Close() error                                                          { return nil }

// ---------------------------------------------------------------------------
// DB helpers — all DB-backed tests require DATABASE_URL.
// ---------------------------------------------------------------------------

// testPool opens a connection pool.
// The test is skipped if DATABASE_URL is not set.
// Callers are responsible for closing the pool (typically in TearDownSuite).
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set — skipping DB test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	return pool
}

// newTestService creates a fully initialised Service backed by the test DB.
func newTestService(t *testing.T, pool *pgxpool.Pool) Service {
	t.Helper()
	svc, err := New(Config{Pool: pool, Logger: noopLogger{}})
	require.NoError(t, err)
	return svc
}

// seedTestTenant inserts a minimal tenant row so the role_assignments FK is
// satisfied. Uses ON CONFLICT DO NOTHING so it is safe to call multiple times.
func seedTestTenant(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO tenants (id, name, slug, email, status, timezone, currency_code, metadata, settings)
		VALUES ($1, 'Test Tenant', 'test-tenant', 'test@authz.test', 'ACTIVE', 'UTC', 'USD', '{}', '{}')
		ON CONFLICT (id) DO NOTHING`,
		testTenantID,
	)
	require.NoError(t, err)
}

// cleanTables truncates the two authz tables before each test so each case
// starts with an empty slate. The tenant row is preserved (not truncated)
// because role_assignments FK references tenants, not the other way around.
func cleanTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`TRUNCATE casbin_rule, role_assignments CASCADE`)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// In-memory service — no database required.
// ---------------------------------------------------------------------------

// newMemService creates a *service backed by a pure in-memory Casbin enforcer.
//
// The pool is pointed at a guaranteed-closed port so that revokeExpiredRoles
// returns a non-fatal connection error rather than panicking on a nil pool.
// Enforce() treats revokeExpiredRoles failures as warnings and continues, so
// the in-memory enforcer is evaluated normally.
//
// With EnableAutoSave(false), AddPolicy/RemovePolicy only update the in-memory
// model (no adapter writes), making the entire Enforce → Policy lifecycle
// testable without a database.
//
// To add roles in unit tests, call svc.enforcer.AddRoleForUserInDomain()
// directly, because AssignRole writes to the DB.
func newMemService(t *testing.T) *service {
	t.Helper()
	m, err := casbinmodel.NewModelFromString(casbinModel)
	require.NoError(t, err)

	// NewEnforcer with only the model creates an in-memory enforcer (nil adapter).
	// LoadPolicy is a no-op when adapter is nil.
	e, err := casbin.NewEnforcer(m)
	require.NoError(t, err)
	e.EnableAutoSave(false) // no adapter — writes stay in-memory only

	// Non-nil pool pointing to a closed port. pool.Query() returns an immediate
	// "connection refused" error (not a panic), which Enforce() swallows.
	fakePool, _ := pgxpool.New(context.Background(),
		"postgres://localhost:5999/authz_unit_test_fake?connect_timeout=1")

	return &service{
		enforcer: e,
		pool:     fakePool,
		log:      noopLogger{},
	}
}

// memRole adds a Casbin g-rule directly to the in-memory enforcer.
// Use this instead of AssignRole in unit tests (which needs a real DB).
func memRole(t *testing.T, svc *service, subject, role, domain string) {
	t.Helper()
	_, err := svc.enforcer.AddRoleForUserInDomain(subject, role, domain)
	require.NoError(t, err)
}

// constants reused across integration tests
const (
	testTenantID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	testDomain   = testTenantID // TenantDomain just returns the ID as-is
	testSubject  = "tenant:usr_test_001"
	testRole     = "role:test-role"
)
