package authz

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	casbin "github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// ---------------------------------------------------------------------------
// noopLogger satisfies logger.Logger for tests without any output.
// ---------------------------------------------------------------------------

type noopLogger struct{}

func (noopLogger) Debug(msg string, fields ...logger.Fields)                    {}
func (noopLogger) Info(msg string, fields ...logger.Fields)                     {}
func (noopLogger) Warn(msg string, fields ...logger.Fields)                     {}
func (noopLogger) Error(msg string, fields ...logger.Fields)                    {}
func (noopLogger) Fatal(msg string, fields ...logger.Fields)                    {}
func (noopLogger) DebugContext(_ context.Context, _ string, _ ...logger.Fields) {}
func (noopLogger) InfoContext(_ context.Context, _ string, _ ...logger.Fields)  {}
func (noopLogger) WarnContext(_ context.Context, _ string, _ ...logger.Fields)  {}
func (noopLogger) ErrorContext(_ context.Context, _ string, _ ...logger.Fields) {}
func (noopLogger) WithFields(_ logger.Fields) logger.Logger                     { return noopLogger{} }
func (noopLogger) WithContext(_ context.Context) logger.Logger                  { return noopLogger{} }
func (noopLogger) SetLevel(_ logger.LogLevel)                                   {}
func (noopLogger) Close() error                                                 { return nil }

// ---------------------------------------------------------------------------
// noopCache satisfies cache.Service for unit tests (all ops are no-ops).
// ---------------------------------------------------------------------------

type noopCache struct{}

var errCacheMiss = errors.New("cache miss")

func (noopCache) Get(_ context.Context, _ string, _ any) error                          { return errCacheMiss }
func (noopCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error         { return nil }
func (noopCache) Delete(_ context.Context, _ string) error                              { return nil }
func (noopCache) Flush(_ context.Context) error                                         { return nil }
func (noopCache) MGet(_ context.Context, _ []string) ([]cache.Result, error)            { return nil, nil }
func (noopCache) MSet(_ context.Context, _ map[string]any, _ time.Duration) error       { return nil }
func (noopCache) MDelete(_ context.Context, _ []string) error                           { return nil }
func (noopCache) DeletePattern(_ context.Context, _ string) error                       { return nil }
func (noopCache) Keys(_ context.Context, _ string) ([]string, error)                    { return nil, nil }
func (noopCache) Exists(_ context.Context, _ string) (bool, error)                      { return false, nil }
func (noopCache) TTL(_ context.Context, _ string) (time.Duration, error)                { return 0, nil }
func (noopCache) Expire(_ context.Context, _ string, _ time.Duration) error             { return nil }
func (noopCache) GetMemory(_ context.Context, _ string, _ any) error                   { return errCacheMiss }
func (noopCache) SetMemory(_ context.Context, _ string, _ any, _ time.Duration) error  { return nil }
func (noopCache) DeleteMemory(_ context.Context, _ string) error                        { return nil }
func (noopCache) GetGlobalMemory(_ string, _ any) error                                 { return errCacheMiss }
func (noopCache) SetGlobalMemory(_ string, _ any, _ time.Duration) error                { return nil }
func (noopCache) DeleteGlobalMemory(_ string) error                                     { return nil }
func (noopCache) Ping(_ context.Context) error                                          { return nil }
func (noopCache) Stats() cache.CacheStats                                               { return cache.CacheStats{} }
func (noopCache) Reset()                                                                {}
func (noopCache) Close() error                                                          { return nil }

// ---------------------------------------------------------------------------
// noopRepo satisfies Repository for in-memory unit tests.
// ListExpiredActiveRoleNames returns an error to preserve the original
// "non-fatal connection failure" behavior in newMemService.
// ---------------------------------------------------------------------------

type noopRepo struct{}

func (noopRepo) UpsertRoleAssignment(_ context.Context, _ uuid.UUID, _, _, _ string, _, _ *string, _ *time.Time) error {
	return nil
}
func (noopRepo) DeactivateRoleAssignment(_ context.Context, _, _, _ string) error { return nil }
func (noopRepo) ListRoleAssignments(_ context.Context, _, _ string) ([]RoleAssignment, error) {
	return nil, nil
}
func (noopRepo) ListExpiredActiveRoleNames(_ context.Context, _, _ string) ([]string, error) {
	return nil, fmt.Errorf("not connected")
}

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
	svc, err := New(Config{
		Store:  db.NewStore(pool),
		Cache:  noopCache{},
		Logger: noopLogger{},
	})
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
// Uses noopRepo so that revokeExpiredRoles returns a non-fatal error rather
// than requiring a database connection. Enforce() treats revokeExpiredRoles
// failures as warnings and continues, so the in-memory enforcer is evaluated
// normally.
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

	e, err := casbin.NewEnforcer(m)
	require.NoError(t, err)
	e.EnableAutoSave(false)

	return &service{
		enforcer: e,
		repo:     noopRepo{},
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
