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

// PostgreSQLSessionTestSuite tests PostgreSQL session management functions directly
type PostgreSQLSessionTestSuite struct {
	suite.Suite
	pool  *pgxpool.Pool
	store db.Store
	ctx   context.Context
}

func (suite *PostgreSQLSessionTestSuite) SetupSuite() {
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
}

func (suite *PostgreSQLSessionTestSuite) TearDownSuite() {
	if suite.pool != nil {
		suite.pool.Close()
	}
}

// TestPostgreSQLSetConfig tests PostgreSQL set_config function directly
func (suite *PostgreSQLSessionTestSuite) TestPostgreSQLSetConfig() {
	tests := []struct {
		name     string
		test     func()
	}{
		{
			name: "SetConfig_BasicString",
			test: func() {
				// Test setting a basic string value
				testValue := "test_value_123"
				
				// Execute set_config directly
				_, err := suite.pool.Exec(suite.ctx, 
					"SELECT set_config('app.test_key', $1, false)", testValue)
				assert.NoError(suite.T(), err)
				
				// Verify the value was set using current_setting
				var retrievedValue string
				err = suite.pool.QueryRow(suite.ctx, 
					"SELECT current_setting('app.test_key', true)").Scan(&retrievedValue)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), testValue, retrievedValue)
			},
		},
		{
			name: "SetConfig_UUID",
			test: func() {
				// Test setting a UUID value
				testUUID := uuid.New()
				
				// Execute set_config with UUID
				_, err := suite.pool.Exec(suite.ctx,
					"SELECT set_config('app.test_uuid', $1, false)", testUUID.String())
				assert.NoError(suite.T(), err)
				
				// Verify the UUID was set correctly
				var retrievedUUIDStr string
				err = suite.pool.QueryRow(suite.ctx,
					"SELECT current_setting('app.test_uuid', true)").Scan(&retrievedUUIDStr)
				assert.NoError(suite.T(), err)
				
				// Parse back to UUID to verify format
				retrievedUUID, err := uuid.Parse(retrievedUUIDStr)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), testUUID, retrievedUUID)
			},
		},
		{
			name: "SetConfig_EmptyValue", 
			test: func() {
				// Test setting empty value (equivalent to NULL/reset)
				_, err := suite.pool.Exec(suite.ctx,
					"SELECT set_config('app.empty_test', '', false)")
				assert.NoError(suite.T(), err)
				
				// Verify empty value is returned
				var retrievedValue string
				err = suite.pool.QueryRow(suite.ctx,
					"SELECT current_setting('app.empty_test', true)").Scan(&retrievedValue)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "", retrievedValue)
			},
		},
		{
			name: "CurrentSetting_NonExistent",
			test: func() {
				// Test current_setting with missing_ok=true for non-existent key
				var retrievedValue string
				err := suite.pool.QueryRow(suite.ctx,
					"SELECT current_setting('app.nonexistent_key', true)").Scan(&retrievedValue)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "", retrievedValue) // Should return empty string
			},
		},
		{
			name: "CurrentSetting_NonExistent_StrictMode",
			test: func() {
				// Test current_setting with missing_ok=false for non-existent key
				var retrievedValue string
				err := suite.pool.QueryRow(suite.ctx,
					"SELECT current_setting('app.nonexistent_strict', false)").Scan(&retrievedValue)
				// Should get an error when missing_ok=false and key doesn't exist
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "unrecognized configuration parameter")
			},
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.test()
		})
	}
}

// TestTenantContextSpecificFunctions tests our specific tenant context functions
func (suite *PostgreSQLSessionTestSuite) TestTenantContextSpecificFunctions() {
	testTenantID := uuid.New()
	
	tests := []struct {
		name string
		test func()
	}{
		{
			name: "AppCurrentTenantID_SetAndGet",
			test: func() {
				// Set tenant ID using our specific app.current_tenant_id key
				_, err := suite.pool.Exec(suite.ctx,
					"SELECT set_config('app.current_tenant_id', $1, false)", testTenantID.String())
				assert.NoError(suite.T(), err)
				
				// Get using current_setting
				var retrievedUUIDStr string
				err = suite.pool.QueryRow(suite.ctx,
					"SELECT current_setting('app.current_tenant_id', true)").Scan(&retrievedUUIDStr)
				assert.NoError(suite.T(), err)
				
				// Verify UUID format and value
				retrievedUUID, err := uuid.Parse(retrievedUUIDStr)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), testTenantID, retrievedUUID)
			},
		},
		{
			name: "AppCurrentTenantID_CaseHandling",
			test: func() {
				// Test our CASE statement logic for handling empty/NULL values
				
				// First, set empty value
				_, err := suite.pool.Exec(suite.ctx,
					"SELECT set_config('app.current_tenant_id', '', false)")
				assert.NoError(suite.T(), err)
				
				// Test our CASE logic (matches GetCurrentTenantID query)
				var result interface{}
				err = suite.pool.QueryRow(suite.ctx, `
					SELECT CASE 
						WHEN current_setting('app.current_tenant_id', true) = '' THEN NULL
						ELSE current_setting('app.current_tenant_id')::uuid
					END`).Scan(&result)
				assert.NoError(suite.T(), err)
				assert.Nil(suite.T(), result) // Should be NULL for empty string
				
				// Now set a real UUID
				_, err = suite.pool.Exec(suite.ctx,
					"SELECT set_config('app.current_tenant_id', $1, false)", testTenantID.String())
				assert.NoError(suite.T(), err)
				
				// Test CASE logic with real UUID
				var resultUUID uuid.UUID
				err = suite.pool.QueryRow(suite.ctx, `
					SELECT CASE 
						WHEN current_setting('app.current_tenant_id', true) = '' THEN NULL
						ELSE current_setting('app.current_tenant_id')::uuid
					END`).Scan(&resultUUID)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), testTenantID, resultUUID)
			},
		},
		{
			name: "SessionScopeVsTransaction",
			test: func() {
				// Test that our session-scoped settings persist across queries
				// but not across different connections
				
				// Set value in current session
				_, err := suite.pool.Exec(suite.ctx,
					"SELECT set_config('app.session_test', $1, false)", "session_value")
				assert.NoError(suite.T(), err)
				
				// Verify it persists in same session across multiple queries
				for i := 0; i < 3; i++ {
					var value string
					err = suite.pool.QueryRow(suite.ctx,
						"SELECT current_setting('app.session_test', true)").Scan(&value)
					assert.NoError(suite.T(), err)
					assert.Equal(suite.T(), "session_value", value)
				}
				
				// Test with transaction-scoped setting (is_local=true)
				tx, err := suite.pool.Begin(suite.ctx)
				require.NoError(suite.T(), err)
				
				// Set transaction-scoped value
				_, err = tx.Exec(suite.ctx,
					"SELECT set_config('app.tx_test', $1, true)", "tx_value")
				assert.NoError(suite.T(), err)
				
				// Verify it's set within transaction
				var txValue string
				err = tx.QueryRow(suite.ctx,
					"SELECT current_setting('app.tx_test', true)").Scan(&txValue)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "tx_value", txValue)
				
				// Commit transaction
				err = tx.Commit(suite.ctx)
				assert.NoError(suite.T(), err)
				
				// Verify transaction-scoped value is gone after commit
				var afterTxValue string
				err = suite.pool.QueryRow(suite.ctx,
					"SELECT current_setting('app.tx_test', true)").Scan(&afterTxValue)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "", afterTxValue) // Should be empty after transaction
				
				// But session-scoped value should still be there
				var sessionValue string
				err = suite.pool.QueryRow(suite.ctx,
					"SELECT current_setting('app.session_test', true)").Scan(&sessionValue)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "session_value", sessionValue)
			},
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.test()
		})
	}
}

// TestSQLCGeneratedFunctions tests our SQLC-generated tenant context functions
func (suite *PostgreSQLSessionTestSuite) TestSQLCGeneratedFunctions() {
	testTenantID := uuid.New()
	
	tests := []struct {
		name string
		test func()
	}{
		{
			name: "SetTenantContext_SQLC",
			test: func() {
				// Use SQLC-generated SetTenantContext function
				err := suite.store.SetTenantContext(suite.ctx, testTenantID)
				assert.NoError(suite.T(), err)
				
				// Verify using SQLC-generated GetCurrentTenantID
				currentID, err := suite.store.GetCurrentTenantID(suite.ctx)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), testTenantID, currentID)
			},
		},
		{
			name: "GetCurrentTenantID_SQLC_NoContext",
			test: func() {
				// Reset context first
				err := suite.store.ResetTenantContext(suite.ctx)
				require.NoError(suite.T(), err)
				
				// Get current tenant ID with no context
				currentID, err := suite.store.GetCurrentTenantID(suite.ctx)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), uuid.Nil, currentID)
			},
		},
		{
			name: "ResetTenantContext_SQLC",
			test: func() {
				// Set context first
				err := suite.store.SetTenantContext(suite.ctx, testTenantID)
				require.NoError(suite.T(), err)
				
				// Verify it's set
				currentID, err := suite.store.GetCurrentTenantID(suite.ctx)
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), testTenantID, currentID)
				
				// Reset context
				err = suite.store.ResetTenantContext(suite.ctx)
				assert.NoError(suite.T(), err)
				
				// Verify it's reset
				resetID, err := suite.store.GetCurrentTenantID(suite.ctx)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), uuid.Nil, resetID)
			},
		},
		{
			name: "MultipleContextSwitches_SQLC",
			test: func() {
				tenant1ID := uuid.New()
				tenant2ID := uuid.New()
				
				// Switch between multiple tenant contexts
				for i := 0; i < 5; i++ {
					// Set first tenant
					err := suite.store.SetTenantContext(suite.ctx, tenant1ID)
					assert.NoError(suite.T(), err)
					
					currentID, err := suite.store.GetCurrentTenantID(suite.ctx)
					assert.NoError(suite.T(), err)
					assert.Equal(suite.T(), tenant1ID, currentID)
					
					// Set second tenant
					err = suite.store.SetTenantContext(suite.ctx, tenant2ID)
					assert.NoError(suite.T(), err)
					
					currentID, err = suite.store.GetCurrentTenantID(suite.ctx)
					assert.NoError(suite.T(), err)
					assert.Equal(suite.T(), tenant2ID, currentID)
				}
			},
		},
	}
	
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.test()
		})
	}
}

// TestSessionPersistenceAcrossQueries tests that session context persists properly
func (suite *PostgreSQLSessionTestSuite) TestSessionPersistenceAcrossQueries() {
	testTenantID := uuid.New()
	
	suite.Run("SessionPersistence", func() {
		// Set tenant context
		err := suite.store.SetTenantContext(suite.ctx, testTenantID)
		require.NoError(suite.T(), err)
		
		// Execute multiple queries and verify context persists
		for i := 0; i < 10; i++ {
			currentID, err := suite.store.GetCurrentTenantID(suite.ctx)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), testTenantID, currentID)
			
			// Also test with direct SQL
			var directID uuid.UUID
			err = suite.pool.QueryRow(suite.ctx, `
				SELECT CASE 
					WHEN current_setting('app.current_tenant_id', true) = '' THEN NULL
					ELSE current_setting('app.current_tenant_id')::uuid
				END`).Scan(&directID)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), testTenantID, directID)
		}
	})
}

// TestPostgreSQLSession runs the test suite
func TestPostgreSQLSession(t *testing.T) {
	suite.Run(t, new(PostgreSQLSessionTestSuite))
}