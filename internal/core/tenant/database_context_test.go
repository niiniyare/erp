//go:build database
// +build database

package tenant

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// MockTracingService implements the tracing.TracingService interface for testing
type MockTracingService struct {
	mock.Mock
}

func (m *MockTracingService) StartSpan(ctx context.Context, name string, opts ...tracing.SpanOption) (context.Context, tracing.Span) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(context.Context), args.Get(1).(tracing.Span)
}

func (m *MockTracingService) SpanFromContext(ctx context.Context) tracing.Span {
	args := m.Called(ctx)
	return args.Get(0).(tracing.Span)
}

func (m *MockTracingService) InjectHTTPHeaders(ctx context.Context, headers http.Header) {
	m.Called(ctx, headers)
}

func (m *MockTracingService) ExtractHTTPHeaders(ctx context.Context, headers http.Header) context.Context {
	args := m.Called(ctx, headers)
	return args.Get(0).(context.Context)
}

func (m *MockTracingService) RecordError(ctx context.Context, err error, opts ...tracing.ErrorOption) {
	m.Called(ctx, err, opts)
}

func (m *MockTracingService) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	m.Called(ctx, name, attrs)
}

func (m *MockTracingService) SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	m.Called(ctx, attrs)
}

func (m *MockTracingService) GetTraceID(ctx context.Context) string {
	args := m.Called(ctx)
	return args.String(0)
}

func (m *MockTracingService) GetSpanID(ctx context.Context) string {
	args := m.Called(ctx)
	return args.String(0)
}

func (m *MockTracingService) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockSpan implements tracing.Span for testing
type MockSpan struct {
	mock.Mock
}

func (m *MockSpan) End(opts ...tracing.SpanEndOption) {
	m.Called(opts)
}

func (m *MockSpan) SetAttributes(attrs ...attribute.KeyValue) {
	m.Called(attrs)
}

func (m *MockSpan) SetStatus(code codes.Code, description string) {
	m.Called(code, description)
}

func (m *MockSpan) RecordError(err error, opts ...trace.EventOption) {
	m.Called(err, opts)
}

func (m *MockSpan) AddEvent(name string, attrs ...attribute.KeyValue) {
	m.Called(name, attrs)
}

func (m *MockSpan) SetName(name string) {
	m.Called(name)
}

func (m *MockSpan) IsRecording() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSpan) SpanContext() trace.SpanContext {
	args := m.Called()
	return args.Get(0).(trace.SpanContext)
}

// TenantDatabaseContextTestSuite tests tenant context management with real PostgreSQL database
type TenantDatabaseContextTestSuite struct {
	suite.Suite
	pool         *pgxpool.Pool
	store        db.Store
	repository   Repository
	testTenantID uuid.UUID
	testTenant   *Tenant
	ctx          context.Context
}

func (suite *TenantDatabaseContextTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Use improved database test runner approach
	dbRunner, err := NewDatabaseTestRunner()
	if err != nil {
		suite.T().Skipf("Database not available for testing: %v", err)
		return
	}

	// Get the pool from the database runner
	suite.pool = dbRunner.GetPool()

	// Test connection to ensure it's working
	err = suite.pool.Ping(suite.ctx)
	if err != nil {
		suite.T().Skipf("Database connection failed: %v", err)
		return
	}

	// Create store and repository
	suite.store = db.NewStore(suite.pool)

	// Create a mock tracer for testing
	mockTracer := &MockTracingService{}
	mockSpan := &MockSpan{}

	// Setup basic mock expectations - we don't care about tracing in database tests
	mockTracer.On("StartSpan", mock.Anything, mock.Anything, mock.Anything).Return(
		suite.ctx, mockSpan).Maybe()
	mockSpan.On("End", mock.Anything).Return().Maybe()
	mockSpan.On("SetAttributes", mock.Anything).Return().Maybe()
	mockSpan.On("RecordError", mock.Anything, mock.Anything).Return().Maybe()
	mockSpan.On("SetStatus", mock.Anything, mock.Anything).Return().Maybe()

	suite.repository = NewRepository(suite.store, mockTracer)
}

func (suite *TenantDatabaseContextTestSuite) TearDownSuite() {
	if suite.pool != nil {
		suite.pool.Close()
	}
}

func (suite *TenantDatabaseContextTestSuite) SetupTest() {
	// Create a test tenant for each test with unique identifiers to avoid collisions
	uniqueID := uuid.New().String()
	params := db.CreateTenantParams{
		Name:      fmt.Sprintf("DB Test Tenant %s", uniqueID[0:8]),
		Slug:      fmt.Sprintf("db-test-%s", uniqueID[0:13]),
		Email:     fmt.Sprintf("test-%s@db-tenant.com", uniqueID[0:8]),
		Subdomain: stringPtr(fmt.Sprintf("db-test-%s", uniqueID[0:13])),
		Status:    "active",
		Industry:  stringPtr("technology"),
	}

	sqlcTenant, err := suite.store.CreateTenant(suite.ctx, params)
	require.NoError(suite.T(), err, "Failed to create test tenant")

	suite.testTenant, err = FromSQLCTenant(sqlcTenant)
	require.NoError(suite.T(), err, "Failed to convert SQLC tenant")
	suite.testTenantID = suite.testTenant.ID
}

func (suite *TenantDatabaseContextTestSuite) TearDownTest() {
	// Clean up test tenant
	if suite.testTenantID != uuid.Nil {
		err := suite.store.SoftDeleteTenant(suite.ctx, suite.testTenantID)
		if err != nil {
			suite.T().Logf("Warning: Failed to cleanup test tenant %s: %v", suite.testTenantID, err)
		}
	}
}

// TestPostgreSQLSessionContext tests PostgreSQL session context management
func (suite *TenantDatabaseContextTestSuite) TestPostgreSQLSessionContext() {
	tests := []struct {
		name string
		test func()
	}{
		{
			name: "SetTenantContext_Success",
			test: func() {
				// Set tenant context
				err := suite.store.SetTenantContext(suite.ctx, suite.testTenantID)
				assert.NoError(suite.T(), err)

				// Verify context was set
				currentID, err := suite.store.GetCurrentTenantID(suite.ctx)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), suite.testTenantID, currentID)
			},
		},
		{
			name: "ResetTenantContext_Success",
			test: func() {
				// First set a tenant context
				err := suite.store.SetTenantContext(suite.ctx, suite.testTenantID)
				require.NoError(suite.T(), err)

				// Verify it's set
				currentID, err := suite.store.GetCurrentTenantID(suite.ctx)
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), suite.testTenantID, currentID)

				// Reset the context
				err = suite.store.ResetTenantContext(suite.ctx)
				assert.NoError(suite.T(), err)

				// Verify context was reset (should return uuid.Nil)
				resetID, err := suite.store.GetCurrentTenantID(suite.ctx)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), uuid.Nil, resetID)
			},
		},
		{
			name: "GetCurrentTenantID_NoContext",
			test: func() {
				// Ensure no context is set first
				err := suite.store.ResetTenantContext(suite.ctx)
				require.NoError(suite.T(), err)

				// Try to get current tenant ID with no context
				tenantID, err := suite.store.GetCurrentTenantID(suite.ctx)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), uuid.Nil, tenantID)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.test()
		})
	}
}

// TestTenantContextSwitching tests switching between different tenant contexts
func (suite *TenantDatabaseContextTestSuite) TestTenantContextSwitching() {
	// Create a second test tenant with unique identifiers
	uniqueID2 := uuid.New().String()
	params2 := db.CreateTenantParams{
		Name:      fmt.Sprintf("Second DB Test Tenant %s", uniqueID2[0:8]),
		Slug:      fmt.Sprintf("db-test-tenant-2-%s", uniqueID2[0:8]),
		Email:     fmt.Sprintf("test2-%s@database-tenant.com", uniqueID2[0:8]),
		Subdomain: stringPtr(fmt.Sprintf("db-test-2-%s", uniqueID2[0:8])),
		Status:    "active",
		Industry:  stringPtr("finance"),
	}

	sqlcTenant2, err := suite.store.CreateTenant(suite.ctx, params2)
	require.NoError(suite.T(), err)

	testTenant2, err := FromSQLCTenant(sqlcTenant2)
	require.NoError(suite.T(), err)
	defer func() {
		// Clean up second tenant
		err := suite.store.SoftDeleteTenant(suite.ctx, testTenant2.ID)
		require.NoError(suite.T(), err)
	}()

	// Test switching between tenant contexts
	suite.Run("SwitchBetweenTenants", func() {
		// Start with first tenant
		err := suite.store.SetTenantContext(suite.ctx, suite.testTenantID)
		require.NoError(suite.T(), err)

		currentID, err := suite.store.GetCurrentTenantID(suite.ctx)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.testTenantID, currentID)

		// Switch to second tenant
		err = suite.store.SetTenantContext(suite.ctx, testTenant2.ID)
		require.NoError(suite.T(), err)

		currentID, err = suite.store.GetCurrentTenantID(suite.ctx)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), testTenant2.ID, currentID)

		// Switch back to first tenant
		err = suite.store.SetTenantContext(suite.ctx, suite.testTenantID)
		require.NoError(suite.T(), err)

		currentID, err = suite.store.GetCurrentTenantID(suite.ctx)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.testTenantID, currentID)
	})
}

// TestRepositorySessionContext tests repository methods with session context
func (suite *TenantDatabaseContextTestSuite) TestRepositorySessionContext() {
	tests := []struct {
		name string
		test func()
	}{
		{
			name: "SetTenant_Success",
			test: func() {
				err := suite.repository.SetTenant(suite.ctx, suite.testTenantID)
				assert.NoError(suite.T(), err)

				// Verify tenant was set
				tenantID, err := suite.repository.GetTenant(suite.ctx)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), suite.testTenantID, tenantID)
			},
		},
		{
			name: "GetTenant_AfterSet",
			test: func() {
				// Set tenant first
				err := suite.repository.SetTenant(suite.ctx, suite.testTenantID)
				require.NoError(suite.T(), err)

				// Get tenant ID
				tenantID, err := suite.repository.GetTenant(suite.ctx)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), suite.testTenantID, tenantID)
			},
		},
		{
			name: "ResetTenant_Success",
			test: func() {
				// Set tenant context first
				err := suite.repository.SetTenant(suite.ctx, suite.testTenantID)
				require.NoError(suite.T(), err)

				// Verify it's set
				tenantID, err := suite.repository.GetTenant(suite.ctx)
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), suite.testTenantID, tenantID)

				// Reset tenant context
				err = suite.repository.ResetTenant(suite.ctx)
				assert.NoError(suite.T(), err)

				// Verify it was reset - should return error about no tenant context
				_, err = suite.repository.GetTenant(suite.ctx)
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "no tenant context set")
			},
		},
		{
			name: "GetTenant_NoContext",
			test: func() {
				// Ensure no context is set
				err := suite.repository.ResetTenant(suite.ctx)
				require.NoError(suite.T(), err)

				// Try to get tenant with no context
				_, err = suite.repository.GetTenant(suite.ctx)
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "no tenant context set")
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.test()
		})
	}
}

// TestTenantContextIsolation tests that tenant context is properly isolated
func (suite *TenantDatabaseContextTestSuite) TestTenantContextIsolation() {
	// Create a second test tenant with unique identifiers
	uniqueID3 := uuid.New().String()
	params2 := db.CreateTenantParams{
		Name:      fmt.Sprintf("Isolation Test Tenant %s", uniqueID3[0:8]),
		Slug:      fmt.Sprintf("isolation-test-tenant-%s", uniqueID3[0:8]),
		Email:     fmt.Sprintf("isolation-%s@database-tenant.com", uniqueID3[0:8]),
		Subdomain: stringPtr(fmt.Sprintf("isolation-test-%s", uniqueID3[0:8])),
		Status:    "active",
		Industry:  stringPtr("retail"),
	}

	sqlcTenant2, err := suite.store.CreateTenant(suite.ctx, params2)
	require.NoError(suite.T(), err)

	testTenant2, err := FromSQLCTenant(sqlcTenant2)
	require.NoError(suite.T(), err)
	defer func() {
		// Clean up second tenant
		err := suite.store.SoftDeleteTenant(suite.ctx, testTenant2.ID)
		require.NoError(suite.T(), err)
	}()

	suite.Run("TenantContextIsolation", func() {
		// Set first tenant context
		err := suite.repository.SetTenant(suite.ctx, suite.testTenantID)
		require.NoError(suite.T(), err)

		// Verify first tenant is active
		tenantID, err := suite.repository.GetTenant(suite.ctx)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.testTenantID, tenantID)

		// Switch to second tenant
		err = suite.repository.SetTenant(suite.ctx, testTenant2.ID)
		require.NoError(suite.T(), err)

		// Verify second tenant is now active (and first is not)
		tenantID, err = suite.repository.GetTenant(suite.ctx)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), testTenant2.ID, tenantID)
		assert.NotEqual(suite.T(), suite.testTenantID, tenantID)

		// Switch back to first tenant
		err = suite.repository.SetTenant(suite.ctx, suite.testTenantID)
		require.NoError(suite.T(), err)

		// Verify first tenant is active again
		tenantID, err = suite.repository.GetTenant(suite.ctx)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.testTenantID, tenantID)
		assert.NotEqual(suite.T(), testTenant2.ID, tenantID)
	})
}

// TestEndToEndTenantWorkflow tests complete tenant context workflow
func (suite *TenantDatabaseContextTestSuite) TestEndToEndTenantWorkflow() {
	suite.Run("CompleteWorkflow", func() {
		// 1. Start with no tenant context
		err := suite.repository.ResetTenant(suite.ctx)
		require.NoError(suite.T(), err)

		// 2. Verify no context is set
		_, err = suite.repository.GetTenant(suite.ctx)
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "no tenant context set")

		// 3. Set tenant context
		err = suite.repository.SetTenant(suite.ctx, suite.testTenantID)
		require.NoError(suite.T(), err)

		// 4. Verify tenant context is set correctly
		tenantID, err := suite.repository.GetTenant(suite.ctx)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.testTenantID, tenantID)

		// 5. Test repository operations with tenant context
		tenant, err := suite.repository.GetByID(suite.ctx, suite.testTenantID)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.testTenantID, tenant.ID)
		assert.Equal(suite.T(), suite.testTenant.Name, tenant.Name)

		// 6. Test subdomain resolution
		resolvedID, err := suite.repository.ResolveSubdomainToID(suite.ctx, *suite.testTenant.Subdomain)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), suite.testTenantID, resolvedID)

		// 7. Reset tenant context
		err = suite.repository.ResetTenant(suite.ctx)
		require.NoError(suite.T(), err)

		// 8. Verify context was reset
		_, err = suite.repository.GetTenant(suite.ctx)
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "no tenant context set")
	})
}

// TestPerformanceWithTenantContext tests performance of tenant context operations
func (suite *TenantDatabaseContextTestSuite) TestPerformanceWithTenantContext() {
	suite.Run("PerformanceTest", func() {
		// Measure time for multiple context switches
		iterations := 100
		start := time.Now()

		for i := 0; i < iterations; i++ {
			// Set tenant context
			err := suite.repository.SetTenant(suite.ctx, suite.testTenantID)
			require.NoError(suite.T(), err)

			// Get tenant context
			tenantID, err := suite.repository.GetTenant(suite.ctx)
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), suite.testTenantID, tenantID)

			// Reset tenant context
			err = suite.repository.ResetTenant(suite.ctx)
			require.NoError(suite.T(), err)
		}

		elapsed := time.Since(start)
		avgTime := elapsed / time.Duration(iterations)

		suite.T().Logf("Performed %d tenant context operations in %v (avg: %v per operation)",
			iterations*3, elapsed, avgTime)

		// Assert reasonable performance (should be much faster than 1ms per operation)
		assert.Less(suite.T(), avgTime, 10*time.Millisecond,
			"Tenant context operations should be fast")
	})
}

// TestTenantDatabaseContext runs the test suite
func TestTenantDatabaseContext(t *testing.T) {
	suite.Run(t, new(TenantDatabaseContextTestSuite))
}
