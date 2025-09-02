//go:build database
// +build database

package tenant

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/platform/config"
)

// DatabaseTestRunner provides utilities for running database tests
type DatabaseTestRunner struct {
	pool  *pgxpool.Pool
	store db.Store
	ctx   context.Context
}

// NewDatabaseTestRunner creates a new database test runner
func NewDatabaseTestRunner() (*DatabaseTestRunner, error) {
	ctx := context.Background()

	// Check if TEST_DATABASE_URL is set for testing override first
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		// Try to load configuration with timeout protection
		// If config loading takes too long, fall back to defaults
		configChan := make(chan *config.Config, 1)
		errorChan := make(chan error, 1)
		
		go func() {
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("config loading panicked: %v", r)
				}
			}()
			cfg := config.Load()
			configChan <- cfg
		}()
		
		// Wait for config with timeout
		timeout := time.After(2 * time.Second)
		select {
		case cfg := <-configChan:
			databaseURL = cfg.Database.GetDatabaseURL()
		case err := <-errorChan:
			return nil, fmt.Errorf("config loading error: %w", err)
		case <-timeout:
			// Config loading took too long, use fallback defaults
			databaseURL = getDefaultTestDatabaseURL()
		}
	}

	// Create database connection
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure for testing
	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
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

// GetStore returns the database store for testing
func (r *DatabaseTestRunner) GetStore() db.Store {
	return r.store
}

// GetPool returns the connection pool for testing
func (r *DatabaseTestRunner) GetPool() *pgxpool.Pool {
	return r.pool
}

// CreateTestTenant creates a test tenant for testing purposes
func (r *DatabaseTestRunner) CreateTestTenant(ctx context.Context, name string) (*db.Tenant, error) {
	params := db.CreateTenantParams{
		Name:      name,
		Slug:      name + "-slug",
		Email:     name + "@test.com",
		Subdomain: stringPtr(name + "-sub"),
		Status:    "active",
		Industry:  stringPtr("testing"),
	}

	return r.store.CreateTenant(ctx, params)
}

// getDefaultTestDatabaseURL returns a default database URL for testing
func getDefaultTestDatabaseURL() string {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "admin")
	password := getEnvOrDefault("DB_PASSWORD", "admin")
	dbName := getEnvOrDefault("DB_NAME", "ledger")
	sslMode := getEnvOrDefault("DB_SSL_MODE", "disable")
	
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", 
		user, password, host, port, dbName, sslMode)
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
