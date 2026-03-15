package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/niiniyare/erp/internal/core/tenant"
	tenant_repo "github.com/niiniyare/erp/internal/core/tenant/repository"
	"github.com/niiniyare/erp/internal/platform/cache"

	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TenantHandlerTestSuite follows TDD approach for tenant handler implementation
type TenantHandlerTestSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	app           *fiber.App
	handler       *TenantHandler
	mockLogger    *logger.MockLogger
	mockMetrics   *metrics.MockMetricsProvider
	mockTracer    *tracing.MockService
	mockSpan      *tracing.MockSpan
	tenantService tenant.Service
	mockRepo      *tenant_repo.MockRepository
	mockCache     *cache.MockService

	// Test data
	testTenantID uuid.UUID
	testTenant   *tenant.Tenant
}

func (suite *TenantHandlerTestSuite) SetupTest() {
	// Initialize gomock controller
	suite.ctrl = gomock.NewController(suite.T())

	// Initialize generated mocks
	suite.mockLogger = logger.NewMockLogger(suite.ctrl)
	suite.mockMetrics = metrics.NewMockMetricsProvider(suite.ctrl)
	suite.mockTracer = tracing.NewMockService(suite.ctrl)
	suite.mockSpan = tracing.NewMockSpan(suite.ctrl)
	suite.mockRepo = tenant_repo.NewMockRepository(suite.ctrl)
	suite.mockCache = cache.NewMockService(suite.ctrl)

	// Setup default mock expectations for common operations
	suite.setupDefaultMockExpectations()

	// Create real tenant service with mocked dependencies
	suite.tenantService = tenant.NewService(tenant.Dependencies{
		Store:  nil,
		Cache:  suite.mockCache,
		Tracer: suite.mockTracer,
		Logger: suite.mockLogger,
	})

	// Create handler with real service
	suite.handler = NewTenantHandler(
		suite.tenantService,
		suite.mockLogger,
		suite.mockMetrics,
		suite.mockTracer,
	)

	// Setup test data
	suite.testTenantID = uuid.New()
	suite.testTenant = &tenant.Tenant{
		ID:           suite.testTenantID,
		Slug:         "test-tenant",
		Name:         "Test Tenant",
		Email:        "test@example.com",
		Status:       tenant.StatusActive,
		Timezone:     "UTC",
		CurrencyCode: "USD",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create Fiber app with proper error handler
	suite.app = fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup routes
	// SetupTenantRoutes(suite.app, suite.handler)
}

func (suite *TenantHandlerTestSuite) setupDefaultMockExpectations() {
	// Tracing expectations
	suite.mockTracer.EXPECT().
		StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(context.Background(), suite.mockSpan).
		AnyTimes()
	suite.mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
	suite.mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	suite.mockSpan.EXPECT().RecordError(gomock.Any()).AnyTimes()
	suite.mockSpan.EXPECT().SetStatus(gomock.Any(), gomock.Any()).AnyTimes()

	// Logger expectations
	suite.mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().DebugContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockLogger.EXPECT().WarnContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Metrics expectations
	suite.mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// Cache expectations for common operations
	suite.mockCache.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(cache.ErrCacheMiss).AnyTimes()
	suite.mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	suite.mockCache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
}

func (suite *TenantHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// ============================================================================
// CRUD OPERATION TESTS (RED PHASE)
// ============================================================================

// TestTenantHandler_Create tests tenant creation with validation
func (suite *TenantHandlerTestSuite) TestTenantHandler_Create() {
	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		expectedFields []string
		description    string
		setupMocks     func()
	}{
		{
			name: "successful_tenant_creation",
			payload: `{
				"slug": "acme-corp",
				"name": "ACME Corporation",
				"email": "admin@acme.com",
				"timezone": "America/New_York",
				"currency_code": "USD"
			}`,
			expectedStatus: 201,
			expectedFields: []string{"id", "slug", "name", "email", "status", "created_at"},
			description:    "Should create tenant with valid payload",
			setupMocks: func() {
				suite.mockRepo.EXPECT().GetBySubdomain(gomock.Any(), "acme-corp").Return(nil, fmt.Errorf("tenant not found"))
				suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "minimal_tenant_creation",
			payload: `{
				"slug": "minimal-co",
				"name": "Minimal Company",
				"email": "contact@minimal.co"
			}`,
			expectedStatus: 201,
			expectedFields: []string{"id", "slug", "name", "email", "status"},
			description:    "Should create tenant with minimal required fields",
			setupMocks: func() {
				suite.mockRepo.EXPECT().GetBySubdomain(gomock.Any(), "minimal-co").Return(nil, fmt.Errorf("tenant not found"))
				suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "invalid_slug_format",
			payload: `{
				"slug": "Invalid Slug!",
				"name": "Test Company",
				"email": "test@company.com"
			}`,
			expectedStatus: 400,
			expectedFields: []string{"code", "message"},
			description:    "Should reject invalid slug format",
			setupMocks: func() {
				// No repository calls expected for validation failures
			},
		},
		{
			name: "duplicate_slug",
			payload: `{
				"slug": "existing-slug",
				"name": "Duplicate Company",
				"email": "duplicate@company.com"
			}`,
			expectedStatus: 409,
			expectedFields: []string{"code", "message"},
			description:    "Should reject duplicate slug",
			setupMocks: func() {
				existingTenant := &tenant.Tenant{ID: uuid.New(), Slug: "existing-slug"}
				suite.mockRepo.EXPECT().GetBySubdomain(gomock.Any(), "existing-slug").Return(existingTenant, nil)
			},
		},
		{
			name: "missing_required_fields",
			payload: `{
				"name": "Incomplete Company"
			}`,
			expectedStatus: 400,
			expectedFields: []string{"code", "message"},
			description:    "Should reject missing required fields",
			setupMocks: func() {
				// No repository calls expected for validation failures
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock expectations for this test
			tt.setupMocks()

			req := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := suite.app.Test(req, -1)
			require.NoError(suite.T(), err, "Request should not error")
			defer resp.Body.Close()

			assert.Equal(suite.T(), tt.expectedStatus, resp.StatusCode, tt.description)

			body, err := io.ReadAll(resp.Body)
			require.NoError(suite.T(), err)

			var response map[string]interface{}
			err = json.Unmarshal(body, &response)
			require.NoError(suite.T(), err, "Response should be valid JSON")

			for _, field := range tt.expectedFields {
				assert.Contains(suite.T(), response, field,
					"Response should contain field %s", field)
			}
		})
	}
}

// TestTenantHandler_Get tests tenant retrieval by ID
func (suite *TenantHandlerTestSuite) TestTenantHandler_Get() {
	tests := []struct {
		name           string
		tenantID       string
		view           string
		expectedStatus int
		expectedFields []string
		description    string
		setupMocks     func()
	}{
		{
			name:           "get_tenant_default_view",
			tenantID:       "550e8400-e29b-41d4-a716-446655440000",
			view:           "",
			expectedStatus: 200,
			expectedFields: []string{"id", "slug", "name", "email", "status", "created_at"},
			description:    "Should return tenant with default view",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
				tenantObj := &tenant.Tenant{
					ID:     tenantID,
					Slug:   "test-tenant",
					Name:   "Test Tenant",
					Email:  "test@example.com",
					Status: tenant.StatusActive,
				}
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil)
			},
		},
		{
			name:           "get_tenant_detailed_view",
			tenantID:       "550e8400-e29b-41d4-a716-446655440000",
			view:           "detailed",
			expectedStatus: 200,
			expectedFields: []string{"id", "slug", "name", "email"},
			description:    "Should return tenant with detailed view",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
				tenantObj := &tenant.Tenant{
					ID:     tenantID,
					Slug:   "test-tenant",
					Name:   "Test Tenant",
					Email:  "test@example.com",
					Status: tenant.StatusActive,
				}
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil)
			},
		},
		{
			name:           "get_tenant_summary_view",
			tenantID:       "550e8400-e29b-41d4-a716-446655440000",
			view:           "summary",
			expectedStatus: 200,
			expectedFields: []string{"id", "slug", "name", "status"},
			description:    "Should return tenant with summary view",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
				tenantObj := &tenant.Tenant{
					ID:     tenantID,
					Slug:   "test-tenant",
					Name:   "Test Tenant",
					Email:  "test@example.com",
					Status: tenant.StatusActive,
				}
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(tenantObj, nil)
			},
		},
		{
			name:           "tenant_not_found",
			tenantID:       "999e8400-e29b-41d4-a716-446655440000",
			view:           "",
			expectedStatus: 404,
			expectedFields: []string{"code", "message"},
			description:    "Should return 404 for non-existent tenant",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("999e8400-e29b-41d4-a716-446655440000")
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(nil, fmt.Errorf("tenant not found"))
			},
		},
		{
			name:           "invalid_tenant_id",
			tenantID:       "invalid-uuid",
			view:           "",
			expectedStatus: 400,
			expectedFields: []string{"code", "message"},
			description:    "Should return 400 for invalid UUID",
			setupMocks: func() {
				// No repository calls expected for invalid UUID
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock expectations for this test
			tt.setupMocks()

			url := fmt.Sprintf("/api/v1/tenants/%s", tt.tenantID)
			if tt.view != "" {
				url += "?view=" + tt.view
			}

			req := httptest.NewRequest("GET", url, nil)
			resp, err := suite.app.Test(req, -1)
			require.NoError(suite.T(), err, "Request should not error")
			defer resp.Body.Close()

			assert.Equal(suite.T(), tt.expectedStatus, resp.StatusCode, tt.description)

			body, err := io.ReadAll(resp.Body)
			require.NoError(suite.T(), err)

			var response map[string]interface{}
			err = json.Unmarshal(body, &response)
			require.NoError(suite.T(), err, "Response should be valid JSON")

			for _, field := range tt.expectedFields {
				assert.Contains(suite.T(), response, field,
					"Response should contain field %s", field)
			}
		})
	}
}

// TestTenantHandler_List tests tenant listing with pagination and filtering
func (suite *TenantHandlerTestSuite) TestTenantHandler_List() {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFields []string
		description    string
		setupMocks     func()
	}{
		{
			name:           "list_tenants_default",
			queryParams:    "",
			expectedStatus: 200,
			expectedFields: []string{"data", "pagination"},
			description:    "Should return paginated tenant list",
			setupMocks: func() {
				tenants := []*tenant.Tenant{suite.testTenant}
				suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(tenants, nil)
			},
		},
		{
			name:           "list_tenants_with_pagination",
			queryParams:    "?page=2&page_size=10",
			expectedStatus: 200,
			expectedFields: []string{"data", "pagination"},
			description:    "Should return second page of tenants",
			setupMocks: func() {
				tenants := []*tenant.Tenant{}
				suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(tenants, nil)
			},
		},
		{
			name:           "list_tenants_filter_by_status",
			queryParams:    "?status=ACTIVE",
			expectedStatus: 200,
			expectedFields: []string{"data", "pagination"},
			description:    "Should filter tenants by status",
			setupMocks: func() {
				tenants := []*tenant.Tenant{suite.testTenant}
				suite.mockRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(tenants, nil)
			},
		},
		{
			name:           "list_tenants_search",
			queryParams:    "?search=acme",
			expectedStatus: 200,
			expectedFields: []string{"data", "pagination"},
			description:    "Should search tenants by name/slug",
			setupMocks: func() {
				tenants := []*tenant.Tenant{suite.testTenant}
				suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(tenants, nil)
			},
		},
		{
			name:           "list_tenants_sort",
			queryParams:    "?sort_by=name&sort_order=asc",
			expectedStatus: 200,
			expectedFields: []string{"data", "pagination"},
			description:    "Should sort tenants by name ascending",
			setupMocks: func() {
				tenants := []*tenant.Tenant{suite.testTenant}
				suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(tenants, nil)
			},
		},
		{
			name:           "list_tenants_invalid_page_size",
			queryParams:    "?page_size=1000",
			expectedStatus: 400,
			expectedFields: []string{"code", "message"},
			description:    "Should reject invalid page size",
			setupMocks: func() {
				// No repository calls expected for validation failures
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock expectations for this test
			tt.setupMocks()

			url := "/api/v1/tenants" + tt.queryParams

			req := httptest.NewRequest("GET", url, nil)
			resp, err := suite.app.Test(req, -1)
			require.NoError(suite.T(), err, "Request should not error")
			defer resp.Body.Close()

			assert.Equal(suite.T(), tt.expectedStatus, resp.StatusCode, tt.description)

			body, err := io.ReadAll(resp.Body)
			require.NoError(suite.T(), err)

			var response map[string]interface{}
			err = json.Unmarshal(body, &response)
			require.NoError(suite.T(), err, "Response should be valid JSON")

			for _, field := range tt.expectedFields {
				assert.Contains(suite.T(), response, field,
					"Response should contain field %s", field)
			}
		})
	}
}

// TestTenantHandler_Update tests tenant updates with optimistic locking
func (suite *TenantHandlerTestSuite) TestTenantHandler_Update() {
	tests := []struct {
		name           string
		tenantID       string
		payload        string
		expectedStatus int
		expectedFields []string
		description    string
		setupMocks     func()
	}{
		{
			name:     "successful_tenant_update",
			tenantID: "550e8400-e29b-41d4-a716-446655440000",
			payload: `{
				"name": "ACME Corporation Updated",
				"email": "admin@acme-updated.com",
				"status": "ACTIVE"
			}`,
			expectedStatus: 200,
			expectedFields: []string{"id", "name", "email", "status", "updated_at"},
			description:    "Should update tenant with valid payload",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
				existingTenant := &tenant.Tenant{
					ID:     tenantID,
					Slug:   "test-tenant",
					Name:   "Test Tenant",
					Email:  "test@example.com",
					Status: tenant.StatusActive,
				}
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(existingTenant, nil)
				suite.mockRepo.EXPECT().Update(gomock.Any(), tenantID, gomock.Any()).Return(nil)
			},
		},
		{
			name:     "tenant_status_transition",
			tenantID: "550e8400-e29b-41d4-a716-446655440000",
			payload: `{
				"status": "SUSPENDED"
			}`,
			expectedStatus: 200,
			expectedFields: []string{"status"},
			description:    "Should update tenant status with reason",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
				existingTenant := &tenant.Tenant{
					ID:     tenantID,
					Slug:   "test-tenant",
					Name:   "Test Tenant",
					Email:  "test@example.com",
					Status: tenant.StatusActive,
				}
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(existingTenant, nil)
				suite.mockRepo.EXPECT().Update(gomock.Any(), tenantID, gomock.Any()).Return(nil)
			},
		},
		{
			name:     "partial_tenant_update",
			tenantID: "550e8400-e29b-41d4-a716-446655440000",
			payload: `{
				"timezone": "Europe/London"
			}`,
			expectedStatus: 200,
			expectedFields: []string{"timezone"},
			description:    "Should perform partial update",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
				existingTenant := &tenant.Tenant{
					ID:     tenantID,
					Slug:   "test-tenant",
					Name:   "Test Tenant",
					Email:  "test@example.com",
					Status: tenant.StatusActive,
				}
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(existingTenant, nil)
				suite.mockRepo.EXPECT().Update(gomock.Any(), tenantID, gomock.Any()).Return(nil)
			},
		},
		{
			name:     "tenant_not_found_update",
			tenantID: "999e8400-e29b-41d4-a716-446655440000",
			payload: `{
				"name": "Non-existent Tenant"
			}`,
			expectedStatus: 404,
			expectedFields: []string{"code", "message"},
			description:    "Should return 404 for non-existent tenant",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("999e8400-e29b-41d4-a716-446655440000")
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(nil, fmt.Errorf("tenant not found"))
			},
		},
		{
			name:     "invalid_status_transition",
			tenantID: "550e8400-e29b-41d4-a716-446655440000",
			payload: `{
				"status": "INVALID_STATUS"
			}`,
			expectedStatus: 400,
			expectedFields: []string{"code", "message"},
			description:    "Should reject invalid status values",
			setupMocks: func() {
				// No repository calls expected for validation failures
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock expectations for this test
			tt.setupMocks()

			url := fmt.Sprintf("/api/v1/tenants/%s", tt.tenantID)
			req := httptest.NewRequest("PUT", url, strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := suite.app.Test(req, -1)
			require.NoError(suite.T(), err, "Request should not error")
			defer resp.Body.Close()

			assert.Equal(suite.T(), tt.expectedStatus, resp.StatusCode, tt.description)

			body, err := io.ReadAll(resp.Body)
			require.NoError(suite.T(), err)

			var response map[string]interface{}
			err = json.Unmarshal(body, &response)
			require.NoError(suite.T(), err, "Response should be valid JSON")

			for _, field := range tt.expectedFields {
				assert.Contains(suite.T(), response, field,
					"Response should contain field %s", field)
			}
		})
	}
}

// TestTenantHandler_Delete tests tenant deletion with dependency checks
func (suite *TenantHandlerTestSuite) TestTenantHandler_Delete() {
	tests := []struct {
		name           string
		tenantID       string
		expectedStatus int
		description    string
		setupMocks     func()
	}{
		{
			name:           "successful_tenant_deletion",
			tenantID:       "550e8400-e29b-41d4-a716-446655440000",
			expectedStatus: 204,
			description:    "Should delete tenant successfully",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")
				existingTenant := &tenant.Tenant{
					ID:     tenantID,
					Slug:   "test-tenant",
					Name:   "Test Tenant",
					Email:  "test@example.com",
					Status: tenant.StatusActive,
				}
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(existingTenant, nil)
				suite.mockRepo.EXPECT().
					SoftDelete(gomock.Any(), tenantID).Return(nil)
			},
		},
		{
			name:           "tenant_not_found_deletion",
			tenantID:       "999e8400-e29b-41d4-a716-446655440000",
			expectedStatus: 404,
			description:    "Should return 404 for non-existent tenant",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("999e8400-e29b-41d4-a716-446655440000")
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(nil, fmt.Errorf("tenant not found"))
			},
		},
		{
			name:           "tenant_with_dependencies",
			tenantID:       "550e8400-e29b-41d4-a716-446655440001",
			expectedStatus: 409,
			description:    "Should prevent deletion of tenant with dependencies",
			setupMocks: func() {
				tenantID, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440001")
				existingTenant := &tenant.Tenant{
					ID:     tenantID,
					Slug:   "tenant-with-deps",
					Name:   "Tenant With Dependencies",
					Email:  "deps@example.com",
					Status: tenant.StatusActive,
				}
				suite.mockRepo.EXPECT().GetByID(gomock.Any(), tenantID).Return(existingTenant, nil)
				suite.mockRepo.EXPECT().SoftDelete(gomock.Any(), tenantID).Return(fmt.Errorf("cannot delete tenant with dependencies"))
			},
		},
		{
			name:           "invalid_tenant_id_deletion",
			tenantID:       "invalid-uuid",
			expectedStatus: 400,
			description:    "Should return 400 for invalid UUID",
			setupMocks: func() {
				// No repository calls expected for invalid UUID
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock expectations for this test
			tt.setupMocks()

			url := fmt.Sprintf("/api/v1/tenants/%s", tt.tenantID)
			req := httptest.NewRequest("DELETE", url, nil)

			resp, err := suite.app.Test(req, -1)
			require.NoError(suite.T(), err, "Request should not error")
			defer resp.Body.Close()

			assert.Equal(suite.T(), tt.expectedStatus, resp.StatusCode, tt.description)
		})
	}
}

// ============================================================================
// BUSINESS RULES TESTS (RED PHASE)
// ============================================================================

// TestTenantHandler_BusinessRules tests tenant business logic
func (suite *TenantHandlerTestSuite) TestTenantHandler_BusinessRules() {
	suite.Run("slug_uniqueness_validation", func() {
		// This test will fail until we implement slug uniqueness checking
		payload := `{
			"slug": "existing-slug",
			"name": "Test Company",
			"email": "test@company.com"
		}`

		req := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req, -1)
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), 409, resp.StatusCode, "Should reject duplicate slug")

		body, err := io.ReadAll(resp.Body)
		require.NoError(suite.T(), err)

		var response map[string]interface{}
		err = json.Unmarshal(body, &response)
		require.NoError(suite.T(), err)

		assert.Equal(suite.T(), "SLUG_ALREADY_EXISTS", response["code"])
	})

	suite.Run("tenant_status_transitions", func() {
		// Test valid status transitions
		validTransitions := []struct {
			from string
			to   string
		}{
			{"TRIAL", "ACTIVE"},
			{"ACTIVE", "SUSPENDED"},
			{"SUSPENDED", "ACTIVE"},
			{"PENDING_SETUP", "TRIAL"},
		}

		for _, transition := range validTransitions {
			payload := fmt.Sprintf(`{"status": "%s"}`, transition.to)
			req := httptest.NewRequest("PUT", "/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := suite.app.Test(req, -1)
			require.NoError(suite.T(), err)
			defer resp.Body.Close()

			assert.Equal(suite.T(), 200, resp.StatusCode,
				"Should allow transition from %s to %s", transition.from, transition.to)
		}
	})

	suite.Run("tenant_isolation_enforcement", func() {
		// Test that tenant isolation is properly enforced
		req := httptest.NewRequest("GET", "/api/v1/tenants/550e8400-e29b-41d4-a716-446655440000", nil)
		req.Header.Set("X-Tenant-ID", "different-tenant-id")

		resp, err := suite.app.Test(req, -1)
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), 403, resp.StatusCode, "Should prevent cross-tenant access")
	})
}

// ============================================================================
// CONTENT NEGOTIATION TESTS (RED PHASE)
// ============================================================================

// TestTenantHandler_ContentNegotiation tests JSON/HTML responses
func (suite *TenantHandlerTestSuite) TestTenantHandler_ContentNegotiation() {
	tests := []struct {
		acceptHeader     string
		expectedMimeType string
		description      string
	}{
		{
			acceptHeader:     "application/json",
			expectedMimeType: "application/json",
			description:      "Should return JSON for JSON Accept header",
		},
		{
			acceptHeader:     "text/html",
			expectedMimeType: "text/html",
			description:      "Should return HTML for HTML Accept header",
		},
		{
			acceptHeader:     "*/*",
			expectedMimeType: "application/json",
			description:      "Should default to JSON for wildcard Accept header",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.description, func() {
			req := httptest.NewRequest("GET", "/api/v1/tenants", nil)
			req.Header.Set("Accept", tt.acceptHeader)

			resp, err := suite.app.Test(req, -1)
			require.NoError(suite.T(), err)
			defer resp.Body.Close()

			contentType := resp.Header.Get("Content-Type")
			assert.Contains(suite.T(), contentType, tt.expectedMimeType, tt.description)
		})
	}
}

// ============================================================================
// OBSERVABILITY INTEGRATION TESTS (RED PHASE)
// ============================================================================

// TestTenantHandler_ObservabilityIntegration tests logging, metrics, and tracing
func (suite *TenantHandlerTestSuite) TestTenantHandler_ObservabilityIntegration() {
	req := httptest.NewRequest("GET", "/api/v1/tenants", nil)
	resp, err := suite.app.Test(req, -1)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	// Verify observability calls were made
	// This will fail until we implement proper observability integration
	assert.Equal(suite.T(), 200, resp.StatusCode)
}

// ============================================================================
// CONCURRENT ACCESS TESTS (RED PHASE)
// ============================================================================

// TestTenantHandler_ConcurrentAccess tests thread safety
func (suite *TenantHandlerTestSuite) TestTenantHandler_ConcurrentAccess() {
	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/api/v1/tenants", nil)
			resp, err := suite.app.Test(req, -1)
			require.NoError(suite.T(), err)
			defer resp.Body.Close()

			assert.Equal(suite.T(), 200, resp.StatusCode)
		}(i)
	}

	wg.Wait()
}

// ============================================================================
// ERROR HANDLING TESTS (RED PHASE)
// ============================================================================

// TestTenantHandler_ErrorHandling tests error scenarios
func (suite *TenantHandlerTestSuite) TestTenantHandler_ErrorHandling() {
	suite.Run("invalid_json_payload", func() {
		req := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.app.Test(req, -1)
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), 400, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(suite.T(), err)

		var response map[string]interface{}
		err = json.Unmarshal(body, &response)
		require.NoError(suite.T(), err)

		assert.Contains(suite.T(), response, "code")
		assert.Equal(suite.T(), "INVALID_JSON", response["code"])
	})
}

// Test runner
func TestTenantHandlerSuite(t *testing.T) {
	suite.Run(t, new(TenantHandlerTestSuite))
}

// ============================================================================
// BENCHMARK TESTS (RED PHASE)
// ============================================================================

// BenchmarkTenantHandler_Create benchmarks tenant creation
func BenchmarkTenantHandler_Create(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)

	// Setup expectations
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// TODO: Create proper mock for tenant.Service
	// For now, using nil service (tests will fail)
	var tenantService tenant.Service

	NewTenantHandler(tenantService, mockLogger, mockMetrics, mockTracer)
	app := fiber.New()
	// TODO: setupTenantRoutes implementation
	// setupTenantRoutes(app, handler)

	payload := `{
		"slug": "bench-tenant",
		"name": "Benchmark Tenant",
		"email": "bench@tenant.com"
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/v1/tenants", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req, -1)
		resp.Body.Close()
	}
}

// BenchmarkTenantHandler_List benchmarks tenant listing
func BenchmarkTenantHandler_List(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)

	// Setup expectations
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// TODO: Create proper mock for tenant.Service
	// For now, using nil service (tests will fail)
	var tenantService tenant.Service

	NewTenantHandler(tenantService, mockLogger, mockMetrics, mockTracer)
	app := fiber.New()

	// TODO: setupTenantRoutes implementation
	// setupTenantRoutes(app, handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/tenants", nil)
		resp, _ := app.Test(req, -1)
		resp.Body.Close()
	}
}
