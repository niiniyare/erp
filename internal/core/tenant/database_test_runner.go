//go:build database
// +build database

package tenant

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/niiniyare/erp/db/sqlc"
)

// DatabaseTestRunner provides utilities for running database tests
type DatabaseTestRunner struct {
	pool  *pgxpool.Pool
	store db.Store
	ctx   context.Context
}

// NewDatabaseTestRunner creates a new database test runner
func NewDatabaseTestRunner() (*DatabaseTestRunner, error) {
	// Check if database tests should be run
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("TEST_DATABASE_URL not set, database tests cannot run")
	}

	ctx := context.Background()

	// Create database connection
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure for testing
	config.MaxConns = 5
	config.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create store
	store := db.NewStore(pool)

	return &DatabaseTestRunner{
		pool:  pool,
		store: store,
		ctx:   ctx,
	}, nil
}

// Close closes the database connection
func (r *DatabaseTestRunner) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

// TestTenantContextOperations demonstrates tenant context operations
func (r *DatabaseTestRunner) TestTenantContextOperations() error {
	fmt.Println("=== Testing Tenant Context Operations ===")

	// Test 1: Create a test tenant
	fmt.Println("1. Creating test tenant...")
	params := db.CreateTenantParams{
		Name:      "Database Test Demo",
		Slug:      "db-test-demo",
		Email:     "demo@database-test.com",
		Subdomain: stringPtr("db-demo"),
		Status:    "active",
		Industry:  stringPtr("technology"),
	}

	sqlcTenant, err := r.store.CreateTenant(r.ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create test tenant: %w", err)
	}

	fmt.Printf("   ✓ Created tenant: %s (ID: %s)\n", sqlcTenant.Name, sqlcTenant.ID)

	// Test 2: Set tenant context
	fmt.Println("2. Setting tenant context...")
	err = r.store.SetTenantContext(r.ctx, sqlcTenant.ID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}
	fmt.Printf("   ✓ Set tenant context to: %s\n", sqlcTenant.ID)

	// Test 3: Get current tenant ID
	fmt.Println("3. Getting current tenant ID...")
	currentID, err := r.store.GetCurrentTenantID(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to get current tenant ID: %w", err)
	}
	fmt.Printf("   ✓ Current tenant ID: %s\n", currentID)

	if currentID != sqlcTenant.ID {
		return fmt.Errorf("tenant ID mismatch: expected %s, got %s", sqlcTenant.ID, currentID)
	}

	// Test 4: Check current tenant exists
	fmt.Println("4. Checking if current tenant exists...")
	exists, err := r.store.CheckCurrentTenantExists(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to check current tenant exists: %w", err)
	}
	fmt.Printf("   ✓ Current tenant exists: %t\n", exists)

	if !exists {
		return fmt.Errorf("current tenant should exist but doesn't")
	}

	// Test 5: Test subdomain resolution
	fmt.Println("5. Testing subdomain resolution...")
	resolvedID, err := r.store.ResolveSubdomainToID(r.ctx, stringPtr("db-demo"))
	if err != nil {
		return fmt.Errorf("failed to resolve subdomain: %w", err)
	}
	fmt.Printf("   ✓ Resolved subdomain 'db-demo' to: %s\n", resolvedID)

	if resolvedID != sqlcTenant.ID {
		return fmt.Errorf("subdomain resolution mismatch: expected %s, got %s", sqlcTenant.ID, resolvedID)
	}

	// Test 6: Reset tenant context
	fmt.Println("6. Resetting tenant context...")
	err = r.store.ResetTenantContext(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to reset tenant context: %w", err)
	}
	fmt.Println("   ✓ Reset tenant context")

	// Test 7: Verify context was reset
	fmt.Println("7. Verifying context was reset...")
	resetID, err := r.store.GetCurrentTenantID(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to get current tenant ID after reset: %w", err)
	}
	fmt.Printf("   ✓ Current tenant ID after reset: %s\n", resetID)

	if resetID != uuid.Nil {
		return fmt.Errorf("tenant context should be reset but got: %s", resetID)
	}

	// Test 8: Check current tenant doesn't exist after reset
	fmt.Println("8. Checking current tenant exists after reset...")
	existsAfterReset, err := r.store.CheckCurrentTenantExists(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to check current tenant exists after reset: %w", err)
	}
	fmt.Printf("   ✓ Current tenant exists after reset: %t\n", existsAfterReset)

	if existsAfterReset {
		return fmt.Errorf("current tenant should not exist after reset")
	}

	// Cleanup: Delete test tenant
	fmt.Println("9. Cleaning up test tenant...")
	err = r.store.SoftDeleteTenant(r.ctx, sqlcTenant.ID)
	if err != nil {
		return fmt.Errorf("failed to delete test tenant: %w", err)
	}
	fmt.Printf("   ✓ Deleted test tenant: %s\n", sqlcTenant.ID)

	fmt.Println("=== All tenant context operations completed successfully! ===")
	return nil
}

// TestSessionPersistence demonstrates session persistence across multiple queries
func (r *DatabaseTestRunner) TestSessionPersistence() error {
	fmt.Println("\n=== Testing Session Persistence ===")

	// Create test tenant
	params := db.CreateTenantParams{
		Name:      "Session Test Tenant",
		Slug:      "session-test",
		Email:     "session@test.com",
		Subdomain: stringPtr("session-test"),
		Status:    "active",
		Industry:  stringPtr("testing"),
	}

	sqlcTenant, err := r.store.CreateTenant(r.ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create test tenant: %w", err)
	}
	defer r.store.SoftDeleteTenant(r.ctx, sqlcTenant.ID)

	// Set tenant context
	err = r.store.SetTenantContext(r.ctx, sqlcTenant.ID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	fmt.Printf("Set tenant context to: %s\n", sqlcTenant.ID)

	// Test persistence across multiple queries
	fmt.Println("Testing session persistence across 10 queries...")
	for i := 1; i <= 10; i++ {
		currentID, err := r.store.GetCurrentTenantID(r.ctx)
		if err != nil {
			return fmt.Errorf("failed to get current tenant ID on query %d: %w", i, err)
		}

		if currentID != sqlcTenant.ID {
			return fmt.Errorf("session lost on query %d: expected %s, got %s", i, sqlcTenant.ID, currentID)
		}

		fmt.Printf("   Query %d: ✓ Session persisted (%s)\n", i, currentID)
	}

	fmt.Println("=== Session persistence test completed successfully! ===")
	return nil
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}