package handlers

//
// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http/httptest"
// 	"sync"
// 	"testing"
//
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"
// 	"go.uber.org/mock/gomock"
//
// 	"awo.so/internal/core/tenant"
// 	tenant_repo "awo.so/internal/core/tenant/repository"
// 	"awo.so/internal/platform/cache"
// 	"awo.so/internal/shared/errors"
// 	"awo.so/internal/shared/logger"
// 	"awo.so/internal/shared/metrics"
// 	"awo.so/internal/shared/tracing"
// )
//
// // RouterTestSuite tests the Router implementation with centralized route registration
// type RouterTestSuite struct {
// 	suite.Suite
// 	ctrl        *gomock.Controller
// 	app         *fiber.App
// 	deps        *Dependencies
// 	mockLogger  *logger.MockLogger
// 	mockMetrics *metrics.MockMetricsProvider
// 	mockTracer  *tracing.MockService
// 	mockSpan    *tracing.MockSpan
// 	mockRepo    *tenant_repo.MockRepository
// 	mockCache   *cache.MockService
// }
//
// func (suite *RouterTestSuite) SetupTest() {
// 	// Initialize gomock controller
// 	suite.ctrl = gomock.NewController(suite.T())
//
// 	// Initialize generated mocks
// 	suite.mockLogger = logger.NewMockLogger(suite.ctrl)
// 	suite.mockMetrics = metrics.NewMockMetricsProvider(suite.ctrl)
// 	suite.mockTracer = tracing.NewMockService(suite.ctrl)
// 	suite.mockSpan = tracing.NewMockSpan(suite.ctrl)
// 	suite.mockRepo = tenant_repo.NewMockRepository(suite.ctrl)
// 	suite.mockCache = cache.NewMockService(suite.ctrl)
//
// 	// Setup default mock expectations for common operations
// 	suite.setupDefaultMockExpectations()
//
// 	// Create tenant service for testing
// 	tenantService := tenant.NewService(tenant.Dependencies{
// 		Store:  nil,
// 		Cache:  suite.mockCache,
// 		Tracer: suite.mockTracer,
// 		Logger: suite.mockLogger,
// 	})
//
// 	// Create dependencies
// 	suite.deps = &Dependencies{
// 		Logger:        suite.mockLogger,
// 		Metrics:       suite.mockMetrics,
// 		Tracer:        suite.mockTracer,
// 		TenantService: tenantService,
// 	}
//
// 	// Create Fiber app with proper error handler
// 	suite.app = fiber.New(fiber.Config{
// 		ErrorHandler: func(c *fiber.Ctx, err error) error {
// 			// Handle Fiber's 404 errors properly
// 			code := fiber.StatusInternalServerError
// 			if e, ok := err.(*fiber.Error); ok {
// 				code = e.Code
// 			}
// 			return c.Status(code).JSON(fiber.Map{
// 				"error": err.Error(),
// 			})
// 		},
// 	})
// }
//
// func (suite *RouterTestSuite) setupDefaultMockExpectations() {
// 	// Tracing expectations
// 	suite.mockTracer.EXPECT().
// 		StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).
// 		Return(context.Background(), suite.mockSpan).
// 		AnyTimes()
// 	suite.mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
// 	suite.mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
//
// 	// Logger expectations
// 	suite.mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
// 	suite.mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
// 	suite.mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
// 	suite.mockLogger.EXPECT().WarnContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
// 	suite.mockLogger.EXPECT().ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
// 	suite.mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
//
// 	// Metrics expectations
// 	suite.mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
// }
//
// func (suite *RouterTestSuite) TearDownTest() {
// 	suite.ctrl.Finish()
// }
//
// // TestNewRouter tests router creation with validation
// func (suite *RouterTestSuite) TestNewRouter() {
// 	tests := []struct {
// 		name        string
// 		deps        *Dependencies
// 		expectError bool
// 		errorCode   string
// 		description string
// 	}{
// 		{
// 			name:        "valid_dependencies",
// 			deps:        suite.deps,
// 			expectError: false,
// 			description: "Should create router with valid dependencies",
// 		},
// 		{
// 			name:        "nil_dependencies",
// 			deps:        nil,
// 			expectError: true,
// 			errorCode:   "INVALID_DEPENDENCIES",
// 			description: "Should fail with nil dependencies",
// 		},
// 		{
// 			name: "missing_logger",
// 			deps: &Dependencies{
// 				Metrics: suite.mockMetrics,
// 				Tracer:  suite.mockTracer,
// 			},
// 			expectError: true,
// 			errorCode:   "MISSING_LOGGER",
// 			description: "Should fail with missing logger",
// 		},
// 		{
// 			name: "missing_metrics",
// 			deps: &Dependencies{
// 				Logger: suite.mockLogger,
// 				Tracer: suite.mockTracer,
// 			},
// 			expectError: true,
// 			errorCode:   "MISSING_METRICS",
// 			description: "Should fail with missing metrics",
// 		},
// 		{
// 			name: "missing_tracer",
// 			deps: &Dependencies{
// 				Logger:  suite.mockLogger,
// 				Metrics: suite.mockMetrics,
// 			},
// 			expectError: true,
// 			errorCode:   "MISSING_TRACER",
// 			description: "Should fail with missing tracer",
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		suite.Run(tt.name, func() {
// 			router, err := NewRouter(tt.deps)
//
// 			if tt.expectError {
// 				require.Error(suite.T(), err, tt.description)
// 				require.Nil(suite.T(), router)
//
// 				// Verify it's a BusinessError with correct properties
// 				be := assertBusinessError(suite.T(), err)
// 				if be != nil {
// 					require.Equal(suite.T(), tt.errorCode, be.Code)
// 					require.Equal(suite.T(), errors.CategorySystem, be.Category)
// 					require.Equal(suite.T(), errors.SeverityCritical, be.Severity)
// 					require.NotEmpty(suite.T(), be.Suggestions, "Should include suggestions")
// 				}
// 			} else {
// 				require.NoError(suite.T(), err, tt.description)
// 				require.NotNil(suite.T(), router)
// 			}
// 		})
// 	}
// }
//
// // TestDependencyValidation tests the Validate method
// func (suite *RouterTestSuite) TestDependencyValidation() {
// 	tests := []struct {
// 		name     string
// 		deps     *Dependencies
// 		wantErr  bool
// 		errCode  string
// 		severity errors.Severity
// 	}{
// 		{
// 			name: "valid_dependencies",
// 			deps: &Dependencies{
// 				Logger:  suite.mockLogger,
// 				Metrics: suite.mockMetrics,
// 				Tracer:  suite.mockTracer,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:     "nil_dependencies",
// 			deps:     nil,
// 			wantErr:  true,
// 			errCode:  "INVALID_DEPENDENCIES",
// 			severity: errors.SeverityCritical,
// 		},
// 		{
// 			name: "missing_logger",
// 			deps: &Dependencies{
// 				Metrics: suite.mockMetrics,
// 				Tracer:  suite.mockTracer,
// 			},
// 			wantErr:  true,
// 			errCode:  "MISSING_LOGGER",
// 			severity: errors.SeverityCritical,
// 		},
// 		{
// 			name: "missing_metrics",
// 			deps: &Dependencies{
// 				Logger: suite.mockLogger,
// 				Tracer: suite.mockTracer,
// 			},
// 			wantErr:  true,
// 			errCode:  "MISSING_METRICS",
// 			severity: errors.SeverityCritical,
// 		},
// 		{
// 			name: "missing_tracer",
// 			deps: &Dependencies{
// 				Logger:  suite.mockLogger,
// 				Metrics: suite.mockMetrics,
// 			},
// 			wantErr:  true,
// 			errCode:  "MISSING_TRACER",
// 			severity: errors.SeverityCritical,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		suite.Run(tt.name, func() {
// 			err := tt.deps.Validate()
//
// 			if tt.wantErr {
// 				require.Error(suite.T(), err)
//
// 				be := assertBusinessError(suite.T(), err)
// 				if be != nil {
// 					require.Equal(suite.T(), tt.errCode, be.Code)
// 					require.Equal(suite.T(), errors.CategorySystem, be.Category)
// 					require.Equal(suite.T(), tt.severity, be.Severity)
// 				}
// 			} else {
// 				require.NoError(suite.T(), err)
// 			}
// 		})
// 	}
// }
//
// // TestRegisterAll tests the complete route registration flow
// func (suite *RouterTestSuite) TestRegisterAll() {
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err, "RegisterAll should succeed")
//
// 	// Verify health endpoint is accessible
// 	req := httptest.NewRequest("GET", "/health", nil)
// 	resp, err := suite.app.Test(req, -1)
// 	require.NoError(suite.T(), err)
// 	defer resp.Body.Close()
//
// 	require.Equal(suite.T(), 200, resp.StatusCode, "Health endpoint should be accessible")
// }
//
// // TestHealthRouteRegistration tests health route specifically
// func (suite *RouterTestSuite) TestHealthRouteRegistration() {
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err)
//
// 	tests := []struct {
// 		name           string
// 		path           string
// 		acceptHeader   string
// 		expectedStatus int
// 		expectedFields []string
// 		description    string
// 	}{
// 		{
// 			name:           "health_check_json",
// 			path:           "/health",
// 			acceptHeader:   "application/json",
// 			expectedStatus: 200,
// 			expectedFields: []string{"status", "timestamp", "uptime", "version"},
// 			description:    "Should return JSON health status",
// 		},
// 		{
// 			name:           "health_check_html",
// 			path:           "/health",
// 			acceptHeader:   "text/html",
// 			expectedStatus: 200,
// 			description:    "Should return HTML health status",
// 		},
// 		{
// 			name:           "health_check_root",
// 			path:           "/health/",
// 			acceptHeader:   "application/json",
// 			expectedStatus: 200,
// 			expectedFields: []string{"status", "timestamp", "uptime", "version"},
// 			description:    "Should handle trailing slash",
// 		},
// 		{
// 			name:           "health_check_default",
// 			path:           "/health",
// 			acceptHeader:   "",
// 			expectedStatus: 200,
// 			expectedFields: []string{"status", "timestamp", "uptime", "version"},
// 			description:    "Should default to JSON without Accept header",
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		suite.Run(tt.name, func() {
// 			req := httptest.NewRequest("GET", tt.path, nil)
// 			if tt.acceptHeader != "" {
// 				req.Header.Set("Accept", tt.acceptHeader)
// 			}
//
// 			resp, err := suite.app.Test(req, -1)
// 			require.NoError(suite.T(), err, "Request should not error")
// 			defer resp.Body.Close()
//
// 			require.Equal(suite.T(), tt.expectedStatus, resp.StatusCode, tt.description)
//
// 			// Only validate JSON response structure
// 			if tt.acceptHeader == "application/json" || tt.acceptHeader == "" {
// 				body, err := io.ReadAll(resp.Body)
// 				require.NoError(suite.T(), err)
//
// 				var healthResponse map[string]any
// 				err = json.Unmarshal(body, &healthResponse)
// 				require.NoError(suite.T(), err, "Response should be valid JSON")
//
// 				for _, field := range tt.expectedFields {
// 					require.Contains(suite.T(), healthResponse, field,
// 						"Response should contain field %s", field)
// 				}
//
// 				require.Equal(suite.T(), "healthy", healthResponse["status"],
// 					"Health status should be 'healthy'")
// 			}
// 		})
// 	}
// }
//
// // TestListRoutes tests the route listing functionality
// func (suite *RouterTestSuite) TestListRoutes() {
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	// Before registration - should be empty
// 	routes := router.ListRoutes()
// 	require.Empty(suite.T(), routes, "Should have no routes before registration")
//
// 	// After registration - should contain health module
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err)
//
// 	routes = router.ListRoutes()
// 	require.NotEmpty(suite.T(), routes, "Should have registered routes")
// 	require.Contains(suite.T(), routes, ModuleHealth, "Should contain health module")
// }
//
// // TestGetModuleInfo tests retrieving specific module information
// func (suite *RouterTestSuite) TestGetModuleInfo() {
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err)
//
// 	// Test existing module
// 	info, err := router.GetModuleInfo(ModuleHealth)
// 	require.NoError(suite.T(), err, "Should retrieve existing module info")
// 	require.NotNil(suite.T(), info)
//
// 	// Test non-existent module
// 	_, err = router.GetModuleInfo("nonexistent")
// 	require.Error(suite.T(), err, "Should error for non-existent module")
//
// 	// Verify it's a BusinessError with correct code
// 	be := assertBusinessError(suite.T(), err)
// 	if be != nil {
// 		require.Equal(suite.T(), "MODULE_NOT_FOUND", be.Code)
// 		require.Equal(suite.T(), errors.CategoryBusiness, be.Category)
// 		require.Contains(suite.T(), be.Details, "module_name")
// 		require.Equal(suite.T(), "nonexistent", be.Details["module_name"])
// 	}
// }
//
// // TestHealthCheck tests the router health check functionality
// func (suite *RouterTestSuite) TestHealthCheck() {
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	// Before registration - should warn about no routes
// 	err = router.HealthCheck()
// 	require.Error(suite.T(), err, "Should error with no routes")
//
// 	be := assertBusinessError(suite.T(), err)
// 	if be != nil {
// 		require.Equal(suite.T(), "HEALTH_CHECK_FAILED", be.Code)
// 		require.Equal(suite.T(), errors.SeverityWarning, be.Severity)
// 	}
//
// 	// After registration - should pass
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err)
//
// 	err = router.HealthCheck()
// 	require.NoError(suite.T(), err, "Should pass health check after registration")
// }
//
// // TestPrintRoutes tests that PrintRoutes doesn't panic
// func (suite *RouterTestSuite) TestPrintRoutes() {
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	// Should not panic even with no routes
// 	require.NotPanics(suite.T(), func() {
// 		router.PrintRoutes()
// 	}, "PrintRoutes should not panic with empty routes")
//
// 	// Register routes and print again
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err)
//
// 	require.NotPanics(suite.T(), func() {
// 		router.PrintRoutes()
// 	}, "PrintRoutes should not panic with registered routes")
// }
//
// // TestConcurrentAccess tests concurrent access to router methods
// func (suite *RouterTestSuite) TestConcurrentAccess() {
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err)
//
// 	// Concurrent ListRoutes calls should be safe
// 	var wg sync.WaitGroup
// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			routes := router.ListRoutes()
// 			require.NotEmpty(suite.T(), routes)
// 		}()
// 	}
// 	wg.Wait()
//
// 	// Concurrent GetModuleInfo calls should be safe
// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			info, err := router.GetModuleInfo(ModuleHealth)
// 			require.NoError(suite.T(), err)
// 			require.NotNil(suite.T(), info)
// 		}()
// 	}
// 	wg.Wait()
// }
//
// // TestRouteMiddleware tests that middleware is properly applied
// func (suite *RouterTestSuite) TestRouteMiddleware() {
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err)
//
// 	// Test CORS middleware on health endpoint
// 	req := httptest.NewRequest("OPTIONS", "/health", nil)
// 	req.Header.Set("Origin", "http://localhost:3000")
// 	req.Header.Set("Access-Control-Request-Method", "GET")
//
// 	resp, err := suite.app.Test(req, -1)
// 	require.NoError(suite.T(), err)
// 	defer resp.Body.Close()
//
// 	// Should have CORS headers
// 	require.NotEmpty(suite.T(), resp.Header.Get("Access-Control-Allow-Origin"),
// 		"Should set CORS allow origin header")
// }
//
// // TestContentNegotiation tests content type negotiation in routes
// func (suite *RouterTestSuite) TestContentNegotiation() {
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err)
//
// 	tests := []struct {
// 		acceptHeader     string
// 		expectedMimeType string
// 		description      string
// 	}{
// 		{
// 			acceptHeader:     "application/json",
// 			expectedMimeType: "application/json",
// 			description:      "Should return JSON for JSON Accept header",
// 		},
// 		{
// 			acceptHeader:     "text/html",
// 			expectedMimeType: "text/html",
// 			description:      "Should return HTML for HTML Accept header",
// 		},
// 		{
// 			acceptHeader:     "*/*",
// 			expectedMimeType: "application/json",
// 			description:      "Should default to JSON for wildcard Accept header",
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		suite.Run(tt.description, func() {
// 			req := httptest.NewRequest("GET", "/health", nil)
// 			req.Header.Set("Accept", tt.acceptHeader)
//
// 			resp, err := suite.app.Test(req, -1)
// 			require.NoError(suite.T(), err)
// 			defer resp.Body.Close()
//
// 			contentType := resp.Header.Get("Content-Type")
// 			require.Contains(suite.T(), contentType, tt.expectedMimeType, tt.description)
// 		})
// 	}
// }
//
// // TestInvalidRoutes tests 404 scenarios
// func (suite *RouterTestSuite) TestInvalidRoutes() {
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err)
//
// 	tests := []struct {
// 		path string
// 		code int
// 	}{
// 		{"/invalid-path", 404},
// 		{"/api/v1/unknown", 404},
// 		{"/health/invalid", 404},
// 	}
//
// 	for _, tt := range tests {
// 		suite.Run("status_"+tt.path, func() {
// 			req := httptest.NewRequest("GET", tt.path, nil)
// 			resp, err := suite.app.Test(req, -1)
// 			require.NoError(suite.T(), err)
// 			defer resp.Body.Close()
//
// 			// Fiber returns "Cannot GET /path" for 404s
// 			require.Equal(suite.T(), tt.code, resp.StatusCode,
// 				"Should return %d for path: %s", tt.code, tt.path)
// 		})
// 	}
// }
//
// // TestModuleRegistrationFailure tests error handling during module registration
// func (suite *RouterTestSuite) TestModuleRegistrationFailure() {
// 	// This test would need a way to force a registration failure
// 	// For now, we test that errors are properly wrapped
// 	router, err := NewRouter(suite.deps)
// 	require.NoError(suite.T(), err)
//
// 	// Attempting to register on a nil app should fail gracefully
// 	// Note: This might panic in Fiber, so we test with a valid app
// 	err = router.RegisterAll(suite.app)
// 	require.NoError(suite.T(), err)
// }
//
// // Test runner
// func TestRouterTestSuite(t *testing.T) {
// 	suite.Run(t, new(RouterTestSuite))
// }
//
// // BenchmarkRouterCreation benchmarks router creation
// func BenchmarkRouterCreation(b *testing.B) {
// 	ctrl := gomock.NewController(b)
// 	defer ctrl.Finish()
//
// 	mockLogger := logger.NewMockLogger(ctrl)
// 	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
// 	mockTracer := tracing.NewMockService(ctrl)
//
// 	deps := &Dependencies{
// 		Logger:  mockLogger,
// 		Metrics: mockMetrics,
// 		Tracer:  mockTracer,
// 	}
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_, _ = NewRouter(deps)
// 	}
// }
//
// // BenchmarkRouteRegistration benchmarks route registration
// func BenchmarkRouteRegistration(b *testing.B) {
// 	ctrl := gomock.NewController(b)
// 	defer ctrl.Finish()
//
// 	mockLogger := logger.NewMockLogger(ctrl)
// 	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
// 	mockTracer := tracing.NewMockService(ctrl)
// 	mockSpan := tracing.NewMockSpan(ctrl)
//
// 	// Setup expectations
// 	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).
// 		Return(context.Background(), mockSpan).AnyTimes()
// 	mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
// 	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
// 	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
// 	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
// 	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
//
// 	deps := &Dependencies{
// 		Logger:  mockLogger,
// 		Metrics: mockMetrics,
// 		Tracer:  mockTracer,
// 	}
//
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		app := fiber.New()
// 		router, _ := NewRouter(deps)
// 		_ = router.RegisterAll(app)
// 	}
// }
//
// // Helper functions for debugging tests
//
// // assertBusinessError checks if error is BusinessError and returns it
// func assertBusinessError(t *testing.T, err error) *errors.BusinessError {
// 	t.Helper()
//
// 	if err == nil {
// 		require.Fail(t, "Expected error but got nil")
// 		return nil
// 	}
//
// 	be, ok := err.(*errors.BusinessError)
// 	if !ok {
// 		require.Fail(t,
// 			fmt.Sprintf("Expected BusinessError but got %T: %v", err, err))
// 		return nil
// 	}
// 	return be
// }
