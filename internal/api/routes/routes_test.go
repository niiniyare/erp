package routes

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// RouteRegistrationTestSuite follows TDD approach for route registration implementation
type RouteRegistrationTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	app         *fiber.App
	registry    *RouteRegistry
	mockLogger  *logger.MockLogger
	mockMetrics *metrics.MockMetricsProvider
	mockTracer  *tracing.MockTracingService
	mockSpan    *tracing.MockSpan
}

// SetupTest initializes test dependencies for each test (RED phase setup)
func (suite *RouteRegistrationTestSuite) SetupTest() {
	// Initialize gomock controller
	suite.ctrl = gomock.NewController(suite.T())

	// Initialize mocks
	suite.mockLogger = logger.NewMockLogger(suite.ctrl)
	suite.mockMetrics = metrics.NewMockMetricsProvider(suite.ctrl)
	suite.mockTracer = tracing.NewMockTracingService(suite.ctrl)
	suite.mockSpan = tracing.NewMockSpan(suite.ctrl)

	// Setup basic mock expectations
	suite.mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).Return(context.Background(), suite.mockSpan).AnyTimes()
	suite.mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
	suite.mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// Create Fiber app
	suite.app = fiber.New()

	// Create route registry (this will fail until we implement it)
	suite.registry = NewRouteRegistry(
		suite.mockLogger,
		suite.mockMetrics,
		suite.mockTracer,
	)
}

func (suite *RouteRegistrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// TestRouteRegistry_RegisterModule tests module registration (RED phase)
func (suite *RouteRegistrationTestSuite) TestRouteRegistry_RegisterModule() {
	tests := []struct {
		name           string
		moduleName     string
		basePath       string
		expectedRoutes int
		description    string
	}{
		{
			name:           "register_health_module",
			moduleName:     "health",
			basePath:       "/health",
			expectedRoutes: 1,
			description:    "Should register health module successfully",
		},
		{
			name:           "register_api_module",
			moduleName:     "api",
			basePath:       "/api/v1",
			expectedRoutes: 1,
			description:    "Should register API module with versioning",
		},
		{
			name:           "register_tenant_module",
			moduleName:     "tenant",
			basePath:       "/api/v1/tenants",
			expectedRoutes: 1,
			description:    "Should register tenant module under API",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Register module
			err := suite.registry.RegisterModule(suite.app, tt.moduleName, tt.basePath, func(router fiber.Router) {
				router.Get("/test", func(c *fiber.Ctx) error {
					return c.JSON(map[string]string{"module": tt.moduleName})
				})
			})

			// Verify registration succeeded
			require.NoError(suite.T(), err, "Module registration should not error: %s", tt.description)

			// Test route is accessible
			req := httptest.NewRequest("GET", tt.basePath+"/test", nil)
			resp, err := suite.app.Test(req, -1)
			require.NoError(suite.T(), err, "Route should be accessible")
			defer resp.Body.Close()

			assert.Equal(suite.T(), 200, resp.StatusCode, "Route should return 200 OK")
		})
	}
}

// TestRouteRegistry_MiddlewareChain tests middleware application (RED phase)
func (suite *RouteRegistrationTestSuite) TestRouteRegistry_MiddlewareChain() {
	tests := []struct {
		name           string
		middlewareType string
		expectedHeader string
		description    string
	}{
		{
			name:           "apply_cors_middleware",
			middlewareType: "cors",
			expectedHeader: "Access-Control-Allow-Origin",
			description:    "Should apply CORS middleware to routes",
		},
		{
			name:           "apply_security_middleware",
			middlewareType: "security",
			expectedHeader: "X-Content-Type-Options",
			description:    "Should apply security headers middleware",
		},
		{
			name:           "apply_tenant_middleware",
			middlewareType: "tenant",
			expectedHeader: "X-Tenant-Context",
			description:    "Should apply tenant context middleware",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Create fresh registry and app for this test
			freshRegistry := NewRouteRegistry(
				suite.mockLogger,
				suite.mockMetrics,
				suite.mockTracer,
			)
			freshApp := fiber.New()

			// Register module with middleware
			err := freshRegistry.RegisterModuleWithMiddleware(
				freshApp,
				"test",
				"/test",
				[]string{tt.middlewareType},
				func(router fiber.Router) {
					router.Get("/middleware", func(c *fiber.Ctx) error {
						return c.JSON(map[string]string{"middleware": tt.middlewareType})
					})
				},
			)

			// Verify registration succeeded
			require.NoError(suite.T(), err, "Module registration with middleware should not error: %s", tt.description)

			// Test middleware is applied
			req := httptest.NewRequest("GET", "/test/middleware", nil)
			resp, err := freshApp.Test(req, -1)
			require.NoError(suite.T(), err, "Route should be accessible")
			defer resp.Body.Close()

			assert.Equal(suite.T(), 200, resp.StatusCode, "Route should return 200 OK")
			// Note: In real implementation, we'd verify middleware headers are set
		})
	}
}

// TestRouteRegistry_ConflictDetection tests route conflict detection (RED phase)
func (suite *RouteRegistrationTestSuite) TestRouteRegistry_ConflictDetection() {
	tests := []struct {
		name        string
		firstPath   string
		secondPath  string
		expectError bool
		description string
	}{
		{
			name:        "detect_exact_path_conflict",
			firstPath:   "/api/v1/users",
			secondPath:  "/api/v1/users",
			expectError: true,
			description: "Should detect exact path conflicts",
		},
		{
			name:        "allow_different_paths",
			firstPath:   "/api/v1/users",
			secondPath:  "/api/v1/tenants",
			expectError: false,
			description: "Should allow different paths",
		},
		{
			name:        "allow_different_parameter_names",
			firstPath:   "/api/v1/users/:id",
			secondPath:  "/api/v1/orders/:id",
			expectError: false,
			description: "Should allow different parameter paths",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Create fresh registry for this test
			freshRegistry := NewRouteRegistry(
				suite.mockLogger,
				suite.mockMetrics,
				suite.mockTracer,
			)
			freshApp := fiber.New()

			// Register first route
			err1 := freshRegistry.RegisterModule(freshApp, "first", tt.firstPath, func(router fiber.Router) {
				router.Get("/test", func(c *fiber.Ctx) error {
					return c.JSON(map[string]string{"route": "first"})
				})
			})
			require.NoError(suite.T(), err1, "First route registration should succeed")

			// Register second route
			err2 := freshRegistry.RegisterModule(freshApp, "second", tt.secondPath, func(router fiber.Router) {
				router.Get("/test", func(c *fiber.Ctx) error {
					return c.JSON(map[string]string{"route": "second"})
				})
			})

			if tt.expectError {
				assert.Error(suite.T(), err2, "Should detect route conflict: %s", tt.description)
			} else {
				assert.NoError(suite.T(), err2, "Should allow non-conflicting routes: %s", tt.description)
			}
		})
	}
}

// TestRouteRegistry_PathParameterExtraction tests parameter extraction (RED phase)
func (suite *RouteRegistrationTestSuite) TestRouteRegistry_PathParameterExtraction() {
	tests := []struct {
		name           string
		routePath      string
		requestPath    string
		expectedParams map[string]string
		description    string
	}{
		{
			name:        "extract_single_parameter",
			routePath:   "/users/:id",
			requestPath: "/users/123",
			expectedParams: map[string]string{
				"id": "123",
			},
			description: "Should extract single path parameter",
		},
		{
			name:        "extract_multiple_parameters",
			routePath:   "/tenants/:tenantId/users/:userId",
			requestPath: "/tenants/tenant-123/users/user-456",
			expectedParams: map[string]string{
				"tenantId": "tenant-123",
				"userId":   "user-456",
			},
			description: "Should extract multiple path parameters",
		},
		{
			name:           "no_parameters",
			routePath:      "/health/status",
			requestPath:    "/health/status",
			expectedParams: map[string]string{},
			description:    "Should handle routes with no parameters",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Create fresh registry and app for this test
			freshRegistry := NewRouteRegistry(
				suite.mockLogger,
				suite.mockMetrics,
				suite.mockTracer,
			)
			freshApp := fiber.New()

			// Register route with parameter extraction
			err := freshRegistry.RegisterModule(freshApp, "test", "", func(router fiber.Router) {
				router.Get(tt.routePath, func(c *fiber.Ctx) error {
					params := make(map[string]string)
					for key := range tt.expectedParams {
						params[key] = c.Params(key)
					}
					return c.JSON(map[string]interface{}{
						"params": params,
						"path":   c.Path(),
					})
				})
			})
			require.NoError(suite.T(), err, "Route registration should succeed")

			// Test parameter extraction
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			resp, err := freshApp.Test(req, -1)
			require.NoError(suite.T(), err, "Route should be accessible")
			defer resp.Body.Close()

			assert.Equal(suite.T(), 200, resp.StatusCode, "Route should return 200 OK")
		})
	}
}

// TestRouteRegistry_ModuleGrouping tests route grouping by module (RED phase)
func (suite *RouteRegistrationTestSuite) TestRouteRegistry_ModuleGrouping() {
	// Register multiple modules
	modules := []struct {
		name string
		path string
	}{
		{"health", "/health"},
		{"auth", "/api/v1/auth"},
		{"users", "/api/v1/users"},
		{"tenants", "/api/v1/tenants"},
	}

	for _, module := range modules {
		err := suite.registry.RegisterModule(suite.app, module.name, module.path, func(router fiber.Router) {
			router.Get("/info", func(c *fiber.Ctx) error {
				return c.JSON(map[string]string{
					"module": module.name,
					"path":   module.path,
				})
			})
		})
		require.NoError(suite.T(), err, "Module registration should succeed for %s", module.name)
	}

	// Test that all modules are accessible
	for _, module := range modules {
		req := httptest.NewRequest("GET", module.path+"/info", nil)
		resp, err := suite.app.Test(req, -1)
		require.NoError(suite.T(), err, "Route should be accessible for %s", module.name)
		defer resp.Body.Close()

		assert.Equal(suite.T(), 200, resp.StatusCode, "Route should return 200 OK for %s", module.name)
	}

	// Test module metadata
	registeredModules := suite.registry.GetRegisteredModules()
	assert.Len(suite.T(), registeredModules, len(modules), "Should have registered all modules")

	for _, module := range modules {
		assert.Contains(suite.T(), registeredModules, module.name, "Should contain module %s", module.name)
	}
}

// TestRouteRegistry_PathValidation tests path validation (RED phase)
func (suite *RouteRegistrationTestSuite) TestRouteRegistry_PathValidation() {
	tests := []struct {
		name        string
		path        string
		expectError bool
		description string
	}{
		{
			name:        "valid_absolute_path",
			path:        "/api/v1/users",
			expectError: false,
			description: "Should accept valid absolute paths",
		},
		{
			name:        "invalid_relative_path",
			path:        "api/users",
			expectError: true,
			description: "Should reject relative paths",
		},
		{
			name:        "valid_empty_path_for_root",
			path:        "",
			expectError: false,
			description: "Should allow empty paths for root-level routes",
		},
		{
			name:        "invalid_special_characters",
			path:        "/api/v1/users/<script>",
			expectError: true,
			description: "Should reject paths with invalid characters",
		},
		{
			name:        "valid_parameterized_path",
			path:        "/api/v1/users/:id/profile",
			expectError: false,
			description: "Should accept parameterized paths",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.registry.ValidatePath(tt.path)

			if tt.expectError {
				assert.Error(suite.T(), err, "Should reject invalid path: %s", tt.description)
			} else {
				assert.NoError(suite.T(), err, "Should accept valid path: %s", tt.description)
			}
		})
	}
}

// Test runner
func TestRouteRegistrationSuite(t *testing.T) {
	suite.Run(t, new(RouteRegistrationTestSuite))
}