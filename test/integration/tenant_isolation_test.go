package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TestTenantIsolation verifies that Row Level Security (RLS) properly isolates tenant data
func TestTenantIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// Setup database connection
	dbURL := getTestDatabaseURL()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	// Create test dependencies
	store := db.NewStore(pool)
	tracer := tracing.NewNoOpService()
	cacheService := cache.NewMockService(nil)

	service := tenant.NewService(tenant.Dependencies{
		Store:  store,
		Cache:  cacheService,
		Tracer: tracer,
	})

	t.Run("tenant_context_isolation", func(t *testing.T) {
		// Create two test tenants
		tenant1, err := service.CreateTenant(ctx, tenant.CreateTenantRequest{
			Name:         "Isolation Test Alpha",
			Email:        "alpha@isolation-test.com",
			Slug:         "isolation-alpha-" + uuid.New().String()[:8],
			CurrencyCode: "USD",
		})
		require.NoError(t, err)

		tenant2, err := service.CreateTenant(ctx, tenant.CreateTenantRequest{
			Name:         "Isolation Test Beta",
			Email:        "beta@isolation-test.com",
			Slug:         "isolation-beta-" + uuid.New().String()[:8],
			CurrencyCode: "EUR",
		})
		require.NoError(t, err)

		t.Logf("Created test tenants: Alpha(%s) and Beta(%s)", tenant1.ID, tenant2.ID)

		// Test 1: Set tenant context and verify isolation
		err = service.SetTenant(ctx, tenant1.ID)
		require.NoError(t, err)

		// Verify current tenant
		currentTenant, err := service.GetCurrentTenant(ctx)
		require.NoError(t, err)
		assert.Equal(t, tenant1.ID, currentTenant.ID)

		// Test 2: Create tenant configuration with RLS
		// First, create configuration for tenant1
		_, err = store.CreateTenantConfiguration(ctx, db.CreateTenantConfigurationParams{
			TenantID:        tenant1.ID,
			DefaultCurrency: "USD",
		})
		require.NoError(t, err)

		// Switch to tenant2 context
		err = service.SetTenant(ctx, tenant2.ID)
		require.NoError(t, err)

		// Try to get tenant1's configuration (should fail or return empty)
		_, err = store.GetTenantConfiguration(ctx)
		assert.Error(t, err, "Should not be able to access tenant1's config from tenant2 context")

		// Create configuration for tenant2
		_, err = store.CreateTenantConfiguration(ctx, db.CreateTenantConfigurationParams{
			TenantID:        tenant2.ID,
			DefaultCurrency: "EUR",
		})
		require.NoError(t, err)

		// Verify we can access tenant2's configuration
		config, err := store.GetTenantConfiguration(ctx)
		require.NoError(t, err)
		assert.Equal(t, "EUR", config.DefaultCurrency)

		// Test 3: Reset tenant context
		err = service.ResetTenant(ctx)
		require.NoError(t, err)

		// Without tenant context, operations should fail
		_, err = store.GetTenantConfiguration(ctx)
		assert.Error(t, err, "Should not be able to access config without tenant context")
	})

	t.Run("cross_tenant_data_protection", func(t *testing.T) {
		// Create test tenants
		tenantA, err := service.CreateTenant(ctx, tenant.CreateTenantRequest{
			Name:         "Protected Alpha",
			Email:        "protected-alpha@test.com",
			Slug:         "protected-alpha-" + uuid.New().String()[:8],
			CurrencyCode: "USD",
		})
		require.NoError(t, err)

		tenantB, err := service.CreateTenant(ctx, tenant.CreateTenantRequest{
			Name:         "Protected Beta",
			Email:        "protected-beta@test.com",
			Slug:         "protected-beta-" + uuid.New().String()[:8],
			CurrencyCode: "GBP",
		})
		require.NoError(t, err)

		// Set context to tenantA
		err = service.SetTenant(ctx, tenantA.ID)
		require.NoError(t, err)

		// Create usage stats for tenantA
		_, err = store.InitializeUsageStats(ctx, tenantA.ID)
		require.NoError(t, err)

		// Switch to tenantB
		err = service.SetTenant(ctx, tenantB.ID)
		require.NoError(t, err)

		// Try to access tenantA's usage stats (should fail)
		periodStart := time.Now().Truncate(24 * time.Hour)
		_, err = store.GetTenantUsageStats(ctx, periodStart)
		assert.Error(t, err, "Should not access tenantA's usage stats from tenantB context")

		// Create and verify tenantB's own usage stats
		_, err = store.InitializeUsageStats(ctx, tenantB.ID)
		require.NoError(t, err)

		stats, err := store.GetLatestTenantUsageStats(ctx)
		require.NoError(t, err)
		assert.NotNil(t, stats)
	})

	t.Run("admin_operations_bypass_rls", func(t *testing.T) {
		// Admin operations like ListTenants should see all tenants
		// This tests that certain operations can bypass RLS when needed

		// Clear any tenant context
		err = service.ResetTenant(ctx)
		require.NoError(t, err)

		// List all tenants (admin operation)
		allTenants, err := service.ListTenants(ctx, 0, 100)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allTenants), 4, "Should see all created test tenants")

		// Verify we can see tenants created in previous tests
		var foundAlpha, foundBeta bool
		for _, t := range allTenants {
			if t.Name == "Isolation Test Alpha" {
				foundAlpha = true
			}
			if t.Name == "Isolation Test Beta" {
				foundBeta = true
			}
		}
		assert.True(t, foundAlpha, "Should find Alpha tenant in admin list")
		assert.True(t, foundBeta, "Should find Beta tenant in admin list")
	})

	t.Run("tenant_context_validation", func(t *testing.T) {
		// Test invalid tenant context
		invalidTenantID := uuid.New()

		err := service.SetTenant(ctx, invalidTenantID)
		assert.Error(t, err, "Should fail to set context for non-existent tenant")

		// Test suspended tenant context
		suspendedTenant, err := service.CreateTenant(ctx, tenant.CreateTenantRequest{
			Name:         "Suspended Test",
			Email:        "suspended@test.com",
			Slug:         "suspended-" + uuid.New().String()[:8],
			CurrencyCode: "USD",
			Status:       tenant.StatusSuspended,
		})
		require.NoError(t, err)

		err = service.SetTenant(ctx, suspendedTenant.ID)
		assert.Error(t, err, "Should fail to set context for suspended tenant")
	})
}

// getTestDatabaseURL returns the test database connection string
func getTestDatabaseURL() string {
	// Use environment variable if set, otherwise use default
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://awo_user:awo_secure_2024@localhost:5432/awoerp_test?sslmode=disable"
}
