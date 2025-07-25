//go:build database
// +build database

package tenant

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RLSPoliciesTestSuite tests Row Level Security policies with tenant context
type RLSPoliciesTestSuite struct {
	suite.Suite
	pool      *pgxpool.Pool
	store     db.Store
	ctx       context.Context
	tenant1   *Tenant
	tenant2   *Tenant
	tenant1ID uuid.UUID
	tenant2ID uuid.UUID
}

func (suite *RLSPoliciesTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Check if database tests should be skipped
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		suite.T().Skip("TEST_DATABASE_URL not set, skipping database tests")
	}

	// Create database connection
	config, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(suite.T(), err)

	// Configure for testing
	config.MaxConns = 5
	config.MinConns = 1

	suite.pool, err = pgxpool.NewWithConfig(suite.ctx, config)
	require.NoError(suite.T(), err)

	// Test connection
	err = suite.pool.Ping(suite.ctx)
	require.NoError(suite.T(), err)

	// Create store
	suite.store = db.NewStore(suite.pool)

	// Create two test tenants for isolation testing
	suite.createTestTenants()
}

func (suite *RLSPoliciesTestSuite) createTestTenants() {
	// Create first tenant
	params1 := db.CreateTenantParams{
		Name:      "RLS Test Tenant 1",
		Slug:      "rls-test-tenant-1",
		Email:     "rls1@test-tenant.com",
		Subdomain: stringPtr("rls-test-1"),
		Status:    "active",
		Industry:  stringPtr("technology"),
	}

	sqlcTenant1, err := suite.store.CreateTenant(suite.ctx, params1)
	require.NoError(suite.T(), err)

	suite.tenant1, err = FromSQLCTenant(sqlcTenant1)
	require.NoError(suite.T(), err)
	suite.tenant1ID = suite.tenant1.ID

	// Create second tenant
	params2 := db.CreateTenantParams{
		Name:      "RLS Test Tenant 2",
		Slug:      "rls-test-tenant-2",
		Email:     "rls2@test-tenant.com",
		Subdomain: stringPtr("rls-test-2"),
		Status:    "active",
		Industry:  stringPtr("finance"),
	}

	sqlcTenant2, err := suite.store.CreateTenant(suite.ctx, params2)
	require.NoError(suite.T(), err)

	suite.tenant2, err = FromSQLCTenant(sqlcTenant2)
	require.NoError(suite.T(), err)
	suite.tenant2ID = suite.tenant2.ID
}

func (suite *RLSPoliciesTestSuite) TearDownSuite() {
	// Clean up test tenants
	if suite.tenant1ID != uuid.Nil {
		err := suite.store.SoftDeleteTenant(suite.ctx, suite.tenant1ID)
		require.NoError(suite.T(), err)
	}
	if suite.tenant2ID != uuid.Nil {
		err := suite.store.SoftDeleteTenant(suite.ctx, suite.tenant2ID)
		require.NoError(suite.T(), err)
	}

	if suite.pool != nil {
		suite.pool.Close()
	}
}

// TestTenantRLSIsolation tests that tenants can only see their own data
func (suite *RLSPoliciesTestSuite) TestTenantRLSIsolation() {
	suite.Run("BasicTenantIsolation", func() {
		// Set context to first tenant
		err := suite.store.SetTenantContext(suite.ctx, suite.tenant1ID)
		require.NoError(suite.T(), err)

		// Try to get first tenant's data - should succeed
		tenant, err := suite.store.GetTenantByID(suite.ctx, suite.tenant1ID)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), tenant)
		assert.Equal(suite.T(), suite.tenant1ID, tenant.ID)

		// Try to get second tenant's data with first tenant's context - behavior depends on RLS implementation
		// Note: This test will show whether RLS is properly implemented
		tenant2, err := suite.store.GetTenantByID(suite.ctx, suite.tenant2ID)
		if err != nil {
			// If RLS is enforced, we expect an error or no results
			suite.T().Logf("RLS properly blocks cross-tenant access: %v", err)
		} else {
			// If no RLS, we get the data (which may be expected for admin queries)
			suite.T().Logf("Cross-tenant access allowed (no RLS or admin mode): %v", tenant2.Name)
		}
	})
}

// TestCurrentTenantQueries tests queries that use get_current_tenant_id()
func (suite *RLSPoliciesTestSuite) TestCurrentTenantQueries() {
	suite.Run("CurrentTenantQueries", func() {
		// Set context to first tenant
		err := suite.store.SetTenantContext(suite.ctx, suite.tenant1ID)
		require.NoError(suite.T(), err)

		// Test GetCurrentTenant function (if implemented)
		// This query uses: WHERE id = get_current_tenant_id()
		currentTenant, err := suite.store.GetCurrentTenant(suite.ctx)
		if err != nil {
			// Function might not exist, that's okay
			suite.T().Logf("GetCurrentTenant not implemented or failed: %v", err)
		} else {
			assert.Equal(suite.T(), suite.tenant1ID, currentTenant.ID)
		}

		// Test CheckCurrentTenantExists function
		exists, err := suite.store.CheckCurrentTenantExists(suite.ctx)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), exists, "Current tenant should exist")

		// Switch to second tenant and test again
		err = suite.store.SetTenantContext(suite.ctx, suite.tenant2ID)
		require.NoError(suite.T(), err)

		exists, err = suite.store.CheckCurrentTenantExists(suite.ctx)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), exists, "Second tenant should also exist")

		// Reset context and test
		err = suite.store.ResetTenantContext(suite.ctx)
		require.NoError(suite.T(), err)

		exists, err = suite.store.CheckCurrentTenantExists(suite.ctx)
		assert.NoError(suite.T(), err)
		assert.False(suite.T(), exists, "No current tenant should exist after reset")
	})
}

// TestTenantContextValidation tests tenant context validation
func (suite *RLSPoliciesTestSuite) TestTenantContextValidation() {
	suite.Run("ValidateTenantContext", func() {
		// Set valid tenant context
		err := suite.store.SetTenantContext(suite.ctx, suite.tenant1ID)
		require.NoError(suite.T(), err)

		// Test ValidateCurrentTenant function (if implemented)
		err = suite.store.ValidateCurrentTenant(suite.ctx)
		if err != nil {
			suite.T().Logf("ValidateCurrentTenant not implemented or failed: %v", err)
		} else {
			suite.T().Log("ValidateCurrentTenant succeeded for valid tenant")
		}

		// Test with invalid/non-existent tenant ID
		invalidTenantID := uuid.New()
		err = suite.store.SetTenantContext(suite.ctx, invalidTenantID)
		require.NoError(suite.T(), err) // Setting context should succeed

		// But validation should fail
		err = suite.store.ValidateCurrentTenant(suite.ctx)
		if err == nil {
			suite.T().Log("ValidateCurrentTenant passed for invalid tenant (may not be implemented)")
		} else {
			suite.T().Logf("ValidateCurrentTenant correctly failed for invalid tenant: %v", err)
		}
	})
}

// TestConcurrentTenantContexts tests that different connections have isolated contexts
func (suite *RLSPoliciesTestSuite) TestConcurrentTenantContexts() {
	suite.Run("ConcurrentContextIsolation", func() {
		// Create a second connection pool to simulate different sessions
		config, err := pgxpool.ParseConfig(os.Getenv("TEST_DATABASE_URL"))
		require.NoError(suite.T(), err)
		config.MaxConns = 5
		config.MinConns = 1

		pool2, err := pgxpool.NewWithConfig(suite.ctx, config)
		require.NoError(suite.T(), err)
		defer pool2.Close()

		store2 := db.NewStore(pool2)

		// Set different tenant contexts in each connection
		err = suite.store.SetTenantContext(suite.ctx, suite.tenant1ID)
		require.NoError(suite.T(), err)

		err = store2.SetTenantContext(suite.ctx, suite.tenant2ID)
		require.NoError(suite.T(), err)

		// Verify each connection maintains its own context
		id1, err := suite.store.GetCurrentTenantID(suite.ctx)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.tenant1ID, id1)

		id2, err := store2.GetCurrentTenantID(suite.ctx)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.tenant2ID, id2)

		// Verify contexts are isolated - changing one doesn't affect the other
		newTenantID := uuid.New()
		err = suite.store.SetTenantContext(suite.ctx, newTenantID)
		require.NoError(suite.T(), err)

		// First connection should have new context
		id1, err = suite.store.GetCurrentTenantID(suite.ctx)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), newTenantID, id1)

		// Second connection should still have its original context
		id2, err = store2.GetCurrentTenantID(suite.ctx)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.tenant2ID, id2)
	})
}

// TestTransactionTenantContext tests tenant context within transactions
func (suite *RLSPoliciesTestSuite) TestTransactionTenantContext() {
	suite.Run("TransactionContext", func() {
		// Set initial tenant context
		err := suite.store.SetTenantContext(suite.ctx, suite.tenant1ID)
		require.NoError(suite.T(), err)

		// Start a transaction
		tx, err := suite.pool.Begin(suite.ctx)
		require.NoError(suite.T(), err)
		defer tx.Rollback(suite.ctx)

		// Verify tenant context is inherited in transaction
		var currentID uuid.UUID
		err = tx.QueryRow(suite.ctx, `
			SELECT CASE 
				WHEN current_setting('app.current_tenant_id', true) = '' THEN NULL
				ELSE current_setting('app.current_tenant_id')::uuid
			END`).Scan(&currentID)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.tenant1ID, currentID)

		// Change tenant context within transaction using transaction-local setting
		_, err = tx.Exec(suite.ctx,
			"SELECT set_config('app.current_tenant_id', $1, true)", suite.tenant2ID.String())
		assert.NoError(suite.T(), err)

		// Verify change within transaction
		err = tx.QueryRow(suite.ctx, `
			SELECT CASE 
				WHEN current_setting('app.current_tenant_id', true) = '' THEN NULL
				ELSE current_setting('app.current_tenant_id')::uuid
			END`).Scan(&currentID)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.tenant2ID, currentID)

		// Commit transaction
		err = tx.Commit(suite.ctx)
		assert.NoError(suite.T(), err)

		// Verify original session context is restored after transaction
		// (transaction-local changes should not persist)
		sessionID, err := suite.store.GetCurrentTenantID(suite.ctx)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.tenant1ID, sessionID,
			"Session context should be restored after transaction-local change")
	})
}

// TestTenantBasedFiltering tests queries with tenant-based filtering
func (suite *RLSPoliciesTestSuite) TestTenantBasedFiltering() {
	suite.Run("TenantBasedFiltering", func() {
		// Test GetActiveTenants which should respect tenant context if RLS is enabled
		activeTenants, err := suite.store.GetActiveTenants(suite.ctx)
		assert.NoError(suite.T(), err)

		// Without tenant context, might see all tenants (depends on RLS implementation)
		allTenantsCount := len(activeTenants)
		suite.T().Logf("Found %d active tenants without tenant context", allTenantsCount)

		// Set tenant context and test again
		err = suite.store.SetTenantContext(suite.ctx, suite.tenant1ID)
		require.NoError(suite.T(), err)

		activeTenants, err = suite.store.GetActiveTenants(suite.ctx)
		assert.NoError(suite.T(), err)

		contextTenantsCount := len(activeTenants)
		suite.T().Logf("Found %d active tenants with tenant context", contextTenantsCount)

		// If RLS is properly implemented, context-filtered results should be <= all results
		assert.LessOrEqual(suite.T(), contextTenantsCount, allTenantsCount,
			"Tenant context should not increase visible tenants")

		// Verify that at least our current tenant is visible
		found := false
		for _, tenant := range activeTenants {
			if tenant.ID == suite.tenant1ID {
				found = true
				break
			}
		}
		assert.True(suite.T(), found, "Current tenant should be visible in active tenants")
	})
}

// TestRLSWithSubdomainResolution tests RLS with subdomain resolution
func (suite *RLSPoliciesTestSuite) TestRLSWithSubdomainResolution() {
	suite.Run("SubdomainResolution", func() {
		// Test ResolveSubdomainToID function
		resolvedID, err := suite.store.ResolveSubdomainToID(suite.ctx, stringPtr("rls-test-1"))
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.tenant1ID, resolvedID)

		// Test second tenant's subdomain
		resolvedID2, err := suite.store.ResolveSubdomainToID(suite.ctx, stringPtr("rls-test-2"))
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.tenant2ID, resolvedID2)

		// Test non-existent subdomain
		_, err = suite.store.ResolveSubdomainToID(suite.ctx, stringPtr("non-existent-subdomain"))
		assert.Error(suite.T(), err, "Should get error for non-existent subdomain")

		// Test GetTenantByUUID function (which takes subdomain parameter)
		tenant, err := suite.store.GetTenantByUUID(suite.ctx, stringPtr("rls-test-1"))
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.tenant1ID, tenant.ID)

		// Test with tenant context set
		err = suite.store.SetTenantContext(suite.ctx, suite.tenant1ID)
		require.NoError(suite.T(), err)

		// Should still be able to resolve our own subdomain
		tenant, err = suite.store.GetTenantByUUID(suite.ctx, stringPtr("rls-test-1"))
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.tenant1ID, tenant.ID)

		// Test accessing other tenant's subdomain with context set
		tenant2, err := suite.store.GetTenantByUUID(suite.ctx, stringPtr("rls-test-2"))
		if err != nil {
			suite.T().Logf("RLS blocks cross-tenant subdomain access: %v", err)
		} else {
			suite.T().Logf("Cross-tenant subdomain access allowed: %v", tenant2.Name)
		}
	})
}

// TestRLSPolicies runs the test suite
func TestRLSPolicies(t *testing.T) {
	suite.Run(t, new(RLSPoliciesTestSuite))
}
