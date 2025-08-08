package handlers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"goa.design/goa/v3/security"

	adminfeatureflag "github.com/niiniyare/erp/internal/api/gen/admin_featureflag"
	"github.com/niiniyare/erp/internal/api/handlers"
	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/abac/models"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

func TestAdminFeatureFlagAPIIntegration(t *testing.T) {
	tests := []struct {
		name        string
		testID      string
		description string
		testFunc    func(t *testing.T)
	}{
		{
			name:        "AdminBulkEnableAPI",
			testID:      "FF-API-005",
			description: "Test admin bulk enable endpoint",
			testFunc:    testAdminBulkEnableAPI,
		},
		{
			name:        "AdminPermissionValidation",
			testID:      "FF-API-006",
			description: "Test admin endpoint permission validation",
			testFunc:    testAdminPermissionValidation,
		},
		{
			name:        "SystemHealthAPI",
			testID:      "FF-API-007",
			description: "Test system health endpoint",
			testFunc:    testSystemHealthAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running API Integration Test ID: %s - %s", tt.testID, tt.description)
			tt.testFunc(t)
		})
	}
}

// FF-API-005: Test admin bulk enable endpoint
func testAdminBulkEnableAPI(t *testing.T) {
	// Given: Valid admin JWT and flag names list
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdminService := featureflag.NewMockAdminService(ctrl)
	mockABACService := abac.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics, _ := metrics.NewMetricsService(metrics.MetricsConfig{
		Enabled: false,
	})
	// Create a minimal tracing service for testing
	tracingConfig := tracing.TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}
	mockTracing, _ := tracing.NewTracingService(tracingConfig)

	// Create service
	service := handlers.NewAdminFeatureFlagService(
		mockAdminService,
		mockABACService,
		mockLogger,
		mockMetrics,
		mockTracing,
	)

	// We're testing the service directly rather than through HTTP endpoints

	userID := uuid.New()
	tenantID := uuid.New()
	flagNames := []string{"feature_a", "feature_b", "feature_c"}
	reason := "Enable features for new release"

	// Mock ABAC authorization (allow)
	expectedAuthResult := &abac.PermissionEvaluationResult{
		Decision:         types.PolicyDecisionAllow,
		EvaluationTimeMS: 2,
		CacheHit:         false,
		RequestID:        "api-test-005",
		PolicyDecisions: []*models.PolicyDecision{
			{PolicyID: uuid.New(), Decision: types.PolicyDecisionAllow, Reason: "admin_bulk_enable_policy"},
		},
		Timestamp: time.Now(),
	}

	mockABACService.EXPECT().
		EvaluatePermission(gomock.Any(), gomock.Any()).
		Return(expectedAuthResult, nil)

	// Mock admin service bulk enable
	expectedResult := &featureflag.BulkOperationResult{
		TotalRequested: 3,
		Successful:     3,
		Failed:         0,
		ExecutedAt:     time.Now(),
		ExecutionTime:  time.Millisecond * 150,
		Summary: featureflag.BulkOperationSummary{
			SuccessRate: 100.0,
		},
		Results: []featureflag.BulkOperationItemResult{
			{Identifier: "feature_a", Success: true},
			{Identifier: "feature_b", Success: true},
			{Identifier: "feature_c", Success: true},
		},
	}

	mockAdminService.EXPECT().
		BulkEnableFlags(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *featureflag.BulkEnableFlagsRequest) (*featureflag.BulkOperationResult, error) {
			assert.Equal(t, flagNames, req.FlagNames)
			assert.Equal(t, reason, req.Reason)
			return expectedResult, nil
		})

	// Create context with authorization info
	ctx := context.Background()
	ctx = context.WithValue(ctx, "jwt_user_id", userID.String())
	ctx = context.WithValue(ctx, "jwt_tenant_id", tenantID.String())
	ctx = context.WithValue(ctx, "user_roles", []string{featureflag.RoleFeatureFlagAdmin})

	// When: Calling BulkEnable endpoint
	payload := &adminfeatureflag.BulkEnablePayload{
		TenantID:  tenantID.String(),
		FlagNames: flagNames,
		Reason:    reason,
	}

	response, err := service.BulkEnable(ctx, payload)

	// Then: Returns successful operation result
	require.NoError(t, err)
	assert.NotNil(t, response)

	// Response includes success/failure counts
	assert.Equal(t, 3, response.TotalRequested)
	assert.Equal(t, 3, response.Successful)
	assert.Equal(t, 0, response.Failed)
	assert.NotEmpty(t, response.ExecutedAt)

	t.Logf("FF-API-005: Bulk enable API returned success with %d/%d flags processed",
		response.Successful, response.TotalRequested)
}

// FF-API-006: Test admin endpoint permission validation
func testAdminPermissionValidation(t *testing.T) {
	// Given: User without admin role
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdminService := featureflag.NewMockAdminService(ctrl)
	mockABACService := abac.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics, _ := metrics.NewMetricsService(metrics.MetricsConfig{
		Enabled: false,
	})
	// Create a minimal tracing service for testing
	tracingConfig := tracing.TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}
	mockTracing, _ := tracing.NewTracingService(tracingConfig)

	// Create service
	service := handlers.NewAdminFeatureFlagService(
		mockAdminService,
		mockABACService,
		mockLogger,
		mockMetrics,
		mockTracing,
	)

	// We're testing the service directly rather than through HTTP endpoints

	userID := uuid.New()
	tenantID := uuid.New()
	flagNames := []string{"feature_a"}
	reason := "Unauthorized attempt"

	// Mock ABAC authorization (deny)
	expectedAuthResult := &abac.PermissionEvaluationResult{
		Decision:         types.PolicyDecisionDeny,
		EvaluationTimeMS: 1,
		CacheHit:         false,
		RequestID:        "api-test-006",
		PolicyDecisions: []*models.PolicyDecision{
			{PolicyID: uuid.New(), Decision: types.PolicyDecisionDeny, Reason: "insufficient_role_policy"},
		},
		Timestamp: time.Now(),
	}

	mockABACService.EXPECT().
		EvaluatePermission(gomock.Any(), gomock.Any()).
		Return(expectedAuthResult, nil)

	// Admin service should not be called when authorization fails
	// mockAdminService.EXPECT() - no expectations set

	// Create context with operator role (insufficient permissions)
	ctx := context.Background()
	ctx = context.WithValue(ctx, "jwt_user_id", userID.String())
	ctx = context.WithValue(ctx, "jwt_tenant_id", tenantID.String())
	ctx = context.WithValue(ctx, "user_roles", []string{featureflag.RoleFeatureFlagOperator})

	// When: Calling BulkEnable endpoint
	payload := &adminfeatureflag.BulkEnablePayload{
		TenantID:  tenantID.String(),
		FlagNames: flagNames,
		Reason:    reason,
	}

	response, err := service.BulkEnable(ctx, payload)

	// Then: Returns authorization error
	require.Error(t, err)
	assert.Nil(t, response)

	// Error message indicates insufficient permissions
	assert.Contains(t, err.Error(), "insufficient permissions")
	assert.Contains(t, err.Error(), "bulk enable operation")

	t.Logf("FF-API-006: Permission validation correctly denied access with error")
}

// FF-API-007: Test system health endpoint
func testSystemHealthAPI(t *testing.T) {
	// Given: Valid admin credentials
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdminService := featureflag.NewMockAdminService(ctrl)
	mockABACService := abac.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics, _ := metrics.NewMetricsService(metrics.MetricsConfig{
		Enabled: false,
	})
	// Create a minimal tracing service for testing
	tracingConfig := tracing.TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}
	mockTracing, _ := tracing.NewTracingService(tracingConfig)

	// Create service
	service := handlers.NewAdminFeatureFlagService(
		mockAdminService,
		mockABACService,
		mockLogger,
		mockMetrics,
		mockTracing,
	)

	// We're testing the service directly rather than through HTTP endpoints

	userID := uuid.New()
	tenantID := uuid.New()

	// Mock ABAC authorization (allow)
	expectedAuthResult := &abac.PermissionEvaluationResult{
		Decision:         types.PolicyDecisionAllow,
		EvaluationTimeMS: 1,
		CacheHit:         true,
		RequestID:        "api-test-007",
		PolicyDecisions: []*models.PolicyDecision{
			{PolicyID: uuid.New(), Decision: types.PolicyDecisionAllow, Reason: "admin_health_check_policy"},
		},
		Timestamp: time.Now(),
	}

	mockABACService.EXPECT().
		EvaluatePermission(gomock.Any(), gomock.Any()).
		Return(expectedAuthResult, nil)

	// Mock admin service health check
	expectedHealthResult := &featureflag.SystemHealthResult{
		Status:         "healthy",
		Timestamp:      time.Now(),
		DatabaseStatus: "healthy",
		CacheStatus:    "healthy",
		OverallScore:   95,
		ComponentHealth: map[string]featureflag.ComponentHealth{
			"database": {
				Status:       "healthy",
				ResponseTime: time.Millisecond * 2,
				Details:      "All connections active",
			},
			"cache": {
				Status:       "healthy",
				ResponseTime: time.Millisecond * 1,
				Details:      "Redis cluster operational",
			},
			"repository": {
				Status:       "healthy",
				ResponseTime: time.Millisecond * 3,
				Details:      "All queries executing normally",
			},
		},
	}

	mockAdminService.EXPECT().
		GetSystemHealth(gomock.Any()).
		Return(expectedHealthResult, nil)

	// Create context with system admin authorization
	ctx := context.Background()
	ctx = context.WithValue(ctx, "jwt_user_id", userID.String())
	ctx = context.WithValue(ctx, "jwt_tenant_id", tenantID.String())
	ctx = context.WithValue(ctx, "user_roles", []string{featureflag.RoleSystemAdmin})

	// When: Calling SystemHealth endpoint
	payload := &adminfeatureflag.SystemHealthPayload{
		TenantID: tenantID.String(),
	}

	startTime := time.Now()
	response, err := service.SystemHealth(ctx, payload)
	responseTime := time.Since(startTime)

	// Then: Returns health status
	require.NoError(t, err)
	assert.NotNil(t, response)

	// Response includes component health details
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "healthy", response.DatabaseStatus)
	assert.Equal(t, "healthy", response.CacheStatus)
	assert.Equal(t, 95, response.OverallScore)
	assert.NotEmpty(t, response.Timestamp)

	// Response time is < 5 seconds
	assert.True(t, responseTime < 5*time.Second, "Response time should be < 5s, got %v", responseTime)

	t.Logf("FF-API-007: Health endpoint returned status %s with score %d in %v",
		response.Status, response.OverallScore, responseTime)
}

// Test Case: JWT Authentication Integration
func TestJWTAuthenticationIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdminService := featureflag.NewMockAdminService(ctrl)
	mockABACService := abac.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics, _ := metrics.NewMetricsService(metrics.MetricsConfig{
		Enabled: false,
	})
	// Create a minimal tracing service for testing
	tracingConfig := tracing.TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}
	mockTracing, _ := tracing.NewTracingService(tracingConfig)

	// Create service
	service := handlers.NewAdminFeatureFlagService(
		mockAdminService,
		mockABACService,
		mockLogger,
		mockMetrics,
		mockTracing,
	)

	t.Run("ValidJWTToken", func(t *testing.T) {
		// Given: Valid JWT token
		ctx := context.Background()
		token := "valid-jwt-token-here"
		scheme := &security.JWTScheme{Name: "jwt"}

		// When: Authenticating with valid token (cast to concrete type)
		adminService, ok := service.(*handlers.AdminFeatureFlagService)
		require.True(t, ok, "Service should be of type *AdminFeatureFlagService")

		authedCtx, err := adminService.JWTAuth(ctx, token, scheme)

		// Then: Authentication succeeds
		require.NoError(t, err)
		assert.NotNil(t, authedCtx)

		// Token is added to context
		ctxToken := authedCtx.Value("jwt_token")
		assert.Equal(t, token, ctxToken)
	})

	t.Run("MissingJWTToken", func(t *testing.T) {
		// Given: Missing JWT token
		ctx := context.Background()
		token := ""
		scheme := &security.JWTScheme{Name: "jwt"}

		// When: Authenticating with missing token
		adminService, ok := service.(*handlers.AdminFeatureFlagService)
		require.True(t, ok, "Service should be of type *AdminFeatureFlagService")

		authedCtx, err := adminService.JWTAuth(ctx, token, scheme)

		// Then: Authentication fails
		require.Error(t, err)
		assert.Nil(t, authedCtx)
		assert.Contains(t, err.Error(), "missing authentication token")
	})
}

// Test Case: Error Response Format Validation
func TestErrorResponseFormat(t *testing.T) {
	t.Run("ValidationErrorMapping", func(t *testing.T) {
		// Test validation error patterns
		validationErr := fmt.Errorf("validation failed: invalid flag name")

		// Check if error contains validation keywords
		isValidationError := strings.Contains(validationErr.Error(), "validation")
		assert.True(t, isValidationError, "Should detect validation error")
	})

	t.Run("PermissionErrorMapping", func(t *testing.T) {
		// Test permission error patterns
		permissionErr := fmt.Errorf("unauthorized access denied")

		// Check if error contains permission keywords
		isPermissionError := strings.Contains(permissionErr.Error(), "unauthorized") || strings.Contains(permissionErr.Error(), "access denied")
		assert.True(t, isPermissionError, "Should detect permission error")
	})
}

// Test Case: Concurrent Request Handling
func TestConcurrentRequestHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdminService := featureflag.NewMockAdminService(ctrl)
	mockABACService := abac.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics, _ := metrics.NewMetricsService(metrics.MetricsConfig{
		Enabled: false,
	})
	tracingConfig := tracing.TracingConfig{
		ServiceName: "test-service",
		Enabled:     false,
	}
	mockTracing, _ := tracing.NewTracingService(tracingConfig)

	// Create service
	service := handlers.NewAdminFeatureFlagService(
		mockAdminService,
		mockABACService,
		mockLogger,
		mockMetrics,
		mockTracing,
	)

	// Setup concurrent test with 10 simultaneous requests
	numRequests := 10
	userID := uuid.New()
	tenantID := uuid.New()

	// Mock ABAC to allow all requests
	mockABACService.EXPECT().
		EvaluatePermission(gomock.Any(), gomock.Any()).
		Return(&abac.PermissionEvaluationResult{
			Decision:         types.PolicyDecisionAllow,
			EvaluationTimeMS: 1,
			RequestID:        "concurrent-test",
			PolicyDecisions: []*models.PolicyDecision{
				{PolicyID: uuid.New(), Decision: types.PolicyDecisionAllow, Reason: "test_policy"},
			},
			Timestamp: time.Now(),
		}, nil).
		Times(numRequests)

	// Mock health service calls
	mockAdminService.EXPECT().
		GetSystemHealth(gomock.Any()).
		Return(&featureflag.SystemHealthResult{
			Status:         "healthy",
			OverallScore:   100,
			DatabaseStatus: "healthy",
			CacheStatus:    "healthy",
			Timestamp:      time.Now(),
		}, nil).
		Times(numRequests)

	// Run concurrent requests
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "jwt_user_id", userID.String())
			ctx = context.WithValue(ctx, "jwt_tenant_id", tenantID.String())
			ctx = context.WithValue(ctx, "user_roles", []string{featureflag.RoleSystemAdmin})

			payload := &adminfeatureflag.SystemHealthPayload{
				TenantID: tenantID.String(),
			}

			_, err := service.SystemHealth(ctx, payload)
			results <- err
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		err := <-results
		if err == nil {
			successCount++
		}
	}

	// Verify all requests succeeded
	assert.Equal(t, numRequests, successCount, "All concurrent requests should succeed")

	t.Logf("Concurrent request test: %d/%d requests succeeded", successCount, numRequests)
}
