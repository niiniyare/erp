//go:build integration
// +build integration

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/api/handlers"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	goahttp "goa.design/goa/v3/http"

	// Import Goa-generated packages
	featureflaggen "github.com/niiniyare/erp/internal/api/gen/featureflag"
	featureflagsvr "github.com/niiniyare/erp/internal/api/gen/http/featureflag/server"
)

// FeatureFlagAPITestSuite provides integration tests for the feature flag API
type FeatureFlagAPITestSuite struct {
	suite.Suite

	// Infrastructure
	ctx         context.Context
	pool        *pgxpool.Pool
	store       db.Store
	redisClient cache.Service

	// Services
	tenantService      tenant.Service
	featureFlagService featureflag.SimpleService

	// API components
	server  *httptest.Server
	handler http.Handler

	// Test data
	testTenantID uuid.UUID
	testUserID   uuid.UUID
}

// SetupSuite runs once before all tests in the suite
func (s *FeatureFlagAPITestSuite) SetupSuite() {
	// Check if integration tests should run
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		s.T().Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	s.ctx = context.Background()

	// Setup database connection
	s.setupDatabase(databaseURL)

	// Setup Redis (mock for now)
	s.setupRedis()

	// Setup services
	s.setupServices()

	// Setup API server
	s.setupAPIServer()

	// Setup test data
	s.setupTestData()
}

// TearDownSuite runs once after all tests in the suite
func (s *FeatureFlagAPITestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	if s.pool != nil {
		s.pool.Close()
	}
}

// SetupTest runs before each test
func (s *FeatureFlagAPITestSuite) SetupTest() {
	// Clean up any existing feature flags
	s.cleanupFeatureFlags()
}

func (s *FeatureFlagAPITestSuite) setupDatabase(databaseURL string) {
	config, err := pgxpool.ParseConfig(databaseURL)
	s.Require().NoError(err, "Failed to parse database URL")

	config.MaxConns = 5
	config.MinConns = 1

	s.pool, err = pgxpool.NewWithConfig(s.ctx, config)
	s.Require().NoError(err, "Failed to create connection pool")

	err = s.pool.Ping(s.ctx)
	s.Require().NoError(err, "Failed to ping database")

	s.store = db.NewStore(s.pool)
}

func (s *FeatureFlagAPITestSuite) setupRedis() {
	// For integration tests, we'll use a mock Redis client
	s.redisClient = &MockRedisClient{}
}

func (s *FeatureFlagAPITestSuite) setupServices() {
	// Create logger
	testLogger := logger.WithFields(logger.Fields{"test": "featureflag_api"})

	// Create metrics service (mock)
	metricsService := &MockMetricsService{}

	// Create tracing service (mock)
	tracingService := &MockTracingService{}

	// Create audit service (mock)
	auditService := &MockAuditService{}

	// Setup tenant service
	tenantRepo := tenant.NewRepository(s.store, tracingService)
	s.tenantService = tenant.NewService(tenantRepo, s.redisClient, tracingService)

	// Setup feature flag service
	featureFlagRepo := featureflag.NewSimpleRepository(s.store)
	s.featureFlagService = featureflag.NewSimpleService(featureFlagRepo, s.tenantService, s.store, auditService)

	// Create GOA handler
	goaHandler := handlers.NewFeatureFlagService(
		s.featureFlagService,
		testLogger,
		metricsService,
		tracingService,
	)

	// Create endpoints
	endpoints := featureflaggen.NewEndpoints(goaHandler)

	// Create HTTP mux
	mux := goahttp.NewMuxer()

	// Create server
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	eh := func(ctx context.Context, w http.ResponseWriter, err error) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	server := featureflagsvr.New(endpoints, mux, dec, enc, eh, nil)
	featureflagsvr.Mount(mux, server)

	s.handler = mux
}

func (s *FeatureFlagAPITestSuite) setupAPIServer() {
	s.server = httptest.NewServer(s.handler)
}

func (s *FeatureFlagAPITestSuite) setupTestData() {
	s.testTenantID = uuid.New()
	s.testUserID = uuid.New()

	// Create test tenant in database
	_, err := s.store.CreateTenant(s.ctx, db.CreateTenantParams{
		ID:     s.testTenantID,
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: "active",
	})
	s.Require().NoError(err, "Failed to create test tenant")
}

func (s *FeatureFlagAPITestSuite) cleanupFeatureFlags() {
	// This would clean up any feature flags created during tests
	// For now, we'll rely on test isolation through unique names
}

// Test Cases

func (s *FeatureFlagAPITestSuite) TestCreateFeatureFlag() {
	payload := map[string]interface{}{
		"name":               "test-create-flag",
		"description":        "A test feature flag for creation",
		"flag_type":          "boolean",
		"default_value":      true,
		"rollout_percentage": 50,
		"target_audience": map[string]interface{}{
			"roles": []string{"admin", "user"},
		},
		"metadata": map[string]interface{}{
			"category": "test",
			"owner":    "engineering",
		},
	}

	body, err := json.Marshal(payload)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(
		s.contextWithTenant(),
		"POST",
		s.server.URL+"/api/v1/feature-flags",
		bytes.NewReader(body),
	)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)

	s.Assert().Equal("test-create-flag", result["name"])
	s.Assert().Equal("boolean", result["flag_type"])
	s.Assert().Equal(true, result["default_value"])
	s.Assert().NotEmpty(result["id"])
}

func (s *FeatureFlagAPITestSuite) TestGetFeatureFlag() {
	// First create a feature flag
	flagName := "test-get-flag"
	s.createTestFlag(flagName, "Test flag for retrieval")

	req, err := http.NewRequestWithContext(
		s.contextWithTenant(),
		"GET",
		s.server.URL+"/api/v1/feature-flags/"+flagName,
		nil,
	)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)

	s.Assert().Equal(flagName, result["name"])
	s.Assert().Equal("Test flag for retrieval", result["description"])
}

func (s *FeatureFlagAPITestSuite) TestListFeatureFlags() {
	// Create multiple test flags
	s.createTestFlag("test-list-flag-1", "First test flag")
	s.createTestFlag("test-list-flag-2", "Second test flag")

	req, err := http.NewRequestWithContext(
		s.contextWithTenant(),
		"GET",
		s.server.URL+"/api/v1/feature-flags?page_size=10&page=1",
		nil,
	)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)

	data, ok := result["data"].([]interface{})
	s.Require().True(ok, "Expected data to be an array")
	s.Assert().GreaterOrEqual(len(data), 2, "Should have at least 2 flags")

	pagination, ok := result["pagination"].(map[string]interface{})
	s.Require().True(ok, "Expected pagination metadata")
	s.Assert().Equal(float64(1), pagination["current_page"])
	s.Assert().Equal(float64(10), pagination["page_size"])
}

func (s *FeatureFlagAPITestSuite) TestUpdateFeatureFlag() {
	// First create a feature flag
	flagName := "test-update-flag"
	flagID := s.createTestFlag(flagName, "Original description")

	payload := map[string]interface{}{
		"description":        "Updated description",
		"default_value":      false,
		"rollout_percentage": 75,
	}

	body, err := json.Marshal(payload)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(
		s.contextWithTenant(),
		"PUT",
		s.server.URL+"/api/v1/feature-flags/"+flagID.String(),
		bytes.NewReader(body),
	)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)

	s.Assert().Equal("Updated description", result["description"])
	s.Assert().Equal(false, result["default_value"])
}

func (s *FeatureFlagAPITestSuite) TestEvaluateFeatureFlag() {
	// Create a test flag with rollout percentage
	flagName := "test-evaluate-flag"
	s.createTestFlag(flagName, "Test flag for evaluation")

	evaluationPayload := map[string]interface{}{
		"name": flagName,
		"context": map[string]interface{}{
			"tenant_id":   s.testTenantID.String(),
			"user_id":     s.testUserID.String(),
			"environment": "test",
			"attributes": map[string]string{
				"role": "admin",
				"plan": "enterprise",
			},
			"client_info": map[string]interface{}{
				"version":    "1.0.0",
				"platform":   "web",
				"ip_address": "127.0.0.1",
				"user_agent": "test-agent",
			},
		},
	}

	body, err := json.Marshal(evaluationPayload)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(
		s.contextWithTenant(),
		"POST",
		s.server.URL+"/api/v1/feature-flags/"+flagName+"/evaluate",
		bytes.NewReader(body),
	)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)

	s.Assert().Equal(flagName, result["flag_name"])
	s.Assert().Contains([]interface{}{true, false}, result["value"])
	s.Assert().Contains([]interface{}{true, false}, result["enabled"])
	s.Assert().NotEmpty(result["reason"])

	metadata, ok := result["metadata"].(map[string]interface{})
	s.Require().True(ok, "Expected metadata object")
	s.Assert().NotEmpty(metadata["evaluated_at"])
	s.Assert().NotNil(metadata["cache_hit"])
}

func (s *FeatureFlagAPITestSuite) TestBulkEvaluateFeatureFlags() {
	// Create multiple test flags
	flag1 := "test-bulk-flag-1"
	flag2 := "test-bulk-flag-2"
	s.createTestFlag(flag1, "First bulk test flag")
	s.createTestFlag(flag2, "Second bulk test flag")

	bulkPayload := map[string]interface{}{
		"flag_names": []string{flag1, flag2},
		"context": map[string]interface{}{
			"tenant_id":   s.testTenantID.String(),
			"user_id":     s.testUserID.String(),
			"environment": "test",
			"attributes": map[string]string{
				"role": "user",
			},
		},
	}

	body, err := json.Marshal(bulkPayload)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(
		s.contextWithTenant(),
		"POST",
		s.server.URL+"/api/v1/feature-flags/evaluate/bulk",
		bytes.NewReader(body),
	)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)

	results, ok := result["results"].(map[string]interface{})
	s.Require().True(ok, "Expected results object")
	s.Assert().Contains(results, flag1)
	s.Assert().Contains(results, flag2)
}

func (s *FeatureFlagAPITestSuite) TestGetFeatureFlagStats() {
	// Create some test flags to generate stats
	s.createTestFlag("stats-flag-1", "Stats test flag 1")
	s.createTestFlag("stats-flag-2", "Stats test flag 2")

	req, err := http.NewRequestWithContext(
		s.contextWithTenant(),
		"GET",
		s.server.URL+"/api/v1/feature-flags/stats",
		nil,
	)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)

	s.Assert().NotNil(result["total_flags"])
	s.Assert().NotNil(result["enabled_flags"])
	s.Assert().NotNil(result["rollout_flags"])
	s.Assert().NotNil(result["flags_by_type"])
	s.Assert().NotNil(result["recent_activity"])
}

func (s *FeatureFlagAPITestSuite) TestDeleteFeatureFlag() {
	// Create a flag to delete
	flagName := "test-delete-flag"
	flagID := s.createTestFlag(flagName, "Flag to be deleted")

	req, err := http.NewRequestWithContext(
		s.contextWithTenant(),
		"DELETE",
		s.server.URL+"/api/v1/feature-flags/"+flagID.String(),
		nil,
	)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusNoContent, resp.StatusCode)

	// Verify the flag is deleted by trying to get it
	req, err = http.NewRequestWithContext(
		s.contextWithTenant(),
		"GET",
		s.server.URL+"/api/v1/feature-flags/"+flagName,
		nil,
	)
	s.Require().NoError(err)

	resp, err = http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *FeatureFlagAPITestSuite) TestHealthCheck() {
	req, err := http.NewRequestWithContext(
		s.ctx,
		"GET",
		s.server.URL+"/api/v1/feature-flags/health",
		nil,
	)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	s.Require().NoError(err)

	s.Assert().Equal("healthy", result["status"])
	s.Assert().NotEmpty(result["timestamp"])
	s.Assert().NotEmpty(result["version"])
}

// Helper methods

func (s *FeatureFlagAPITestSuite) contextWithTenant() context.Context {
	// Add tenant context to the request
	// This would typically be done by middleware in the real application
	return context.WithValue(s.ctx, "tenant_id", s.testTenantID)
}

func (s *FeatureFlagAPITestSuite) createTestFlag(name, description string) uuid.UUID {
	flagID := uuid.New()

	// Create flag directly in database for testing
	_, err := s.store.CreateFeatureFlag(s.ctx, db.CreateFeatureFlagParams{
		ID:             flagID,
		TenantID:       s.testTenantID,
		Name:           name,
		Description:    description,
		FlagType:       "boolean",
		DefaultValue:   true,
		TargetAudience: []byte(`{"roles": ["admin", "user"]}`),
		Metadata:       []byte(`{"category": "test", "owner": "engineering"}`),
	})
	s.Require().NoError(err)

	return flagID
}

// Mock implementations

type MockRedisClient struct{}

func (m *MockRedisClient) Get(ctx context.Context, key string, dest any) error {
	return cache.ErrCacheMiss
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return nil
}

func (m *MockRedisClient) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *MockRedisClient) Flush(ctx context.Context) error {
	return nil
}

func (m *MockRedisClient) MGet(ctx context.Context, keys []string, dest interface{}) error {
	return cache.ErrCacheMiss
}

func (m *MockRedisClient) MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
	return nil
}

func (m *MockRedisClient) MDelete(ctx context.Context, keys []string) error {
	return nil
}

func (m *MockRedisClient) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *MockRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	return nil
}

func (m *MockRedisClient) Stats() *cache.CacheStats {
	return &cache.CacheStats{}
}

func (m *MockRedisClient) Close() error {
	return nil
}

type MockMetricsService struct{}

func (m *MockMetricsService) IncrementCounter(name string, labels map[string]string)               {}
func (m *MockMetricsService) RecordHistogram(name string, value float64, labels map[string]string) {}
func (m *MockMetricsService) SetGauge(name string, value float64, labels map[string]string)        {}

type MockTracingService struct{}

func (m *MockTracingService) StartSpan(ctx context.Context, name string, opts ...tracing.SpanOption) (context.Context, tracing.Span) {
	return ctx, &MockSpan{}
}

func (m *MockTracingService) SpanFromContext(ctx context.Context) tracing.Span {
	return &MockSpan{}
}

func (m *MockTracingService) InjectHTTPHeaders(ctx context.Context, headers http.Header) {}

func (m *MockTracingService) ExtractHTTPHeaders(ctx context.Context, headers http.Header) context.Context {
	return ctx
}

func (m *MockTracingService) RecordError(ctx context.Context, err error, opts ...tracing.ErrorOption) {
}

type MockSpan struct{}

func (m *MockSpan) End()                                       {}
func (m *MockSpan) SetAttributes(attrs ...interface{})         {}
func (m *MockSpan) AddEvent(name string, attrs ...interface{}) {}
func (m *MockSpan) RecordError(err error, opts ...interface{}) {}
func (m *MockSpan) SetStatus(code int, description string)     {}
func (m *MockSpan) SetName(name string)                        {}

type MockAuditService struct{}

func (m *MockAuditService) Record(ctx context.Context, event audit.AuditEvent) error {
	// Mock audit service does nothing but could log for testing
	return nil
}

// Run the test suite
func TestFeatureFlagAPIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(FeatureFlagAPITestSuite))
}
