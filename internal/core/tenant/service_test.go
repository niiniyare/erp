//go:build unit
// +build unit

package tenant

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

// MockCache implements the cache.Service interface for testing
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) Flush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCache) MGet(ctx context.Context, keys []string, dest interface{}) error {
	args := m.Called(ctx, keys, dest)
	return args.Error(0)
}

func (m *MockCache) MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
	args := m.Called(ctx, pairs, expiration)
	return args.Error(0)
}

func (m *MockCache) MDelete(ctx context.Context, keys []string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockCache) DeletePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCache) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCache) Stats() *cache.CacheStats {
	args := m.Called()
	return args.Get(0).(*cache.CacheStats)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

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

func (m *MockSpan) AddEvent(name string, attrs ...attribute.KeyValue) {
	m.Called(name, attrs)
}

func (m *MockSpan) SetAttributes(attrs ...attribute.KeyValue) {
	m.Called(attrs)
}

func (m *MockSpan) SetStatus(code codes.Code, description string) {
	m.Called(code, description)
}

func (m *MockSpan) SetName(name string) {
	m.Called(name)
}

func (m *MockSpan) RecordError(err error, opts ...trace.EventOption) {
	m.Called(err, opts)
}

func (m *MockSpan) IsRecording() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSpan) SpanContext() trace.SpanContext {
	args := m.Called()
	return args.Get(0).(trace.SpanContext)
}

// TenantServiceTestSuite is the main test suite for the tenant service
type TenantServiceTestSuite struct {
	suite.Suite
	service      Service
	mockRepo     *MockRepository
	mockCache    *MockCache
	mockTracer   *MockTracingService
	mockSpan     *MockSpan
	ctx          context.Context
	testTenantID uuid.UUID
	testTenant   *Tenant
	gomockCtrl   *gomock.Controller
}

func (suite *TenantServiceTestSuite) SetupSuite() {
	// Suite-level setup - use a fixed UUID for consistency across tests
	suite.testTenantID = uuid.MustParse("12345678-1234-1234-1234-123456789012")
	subdomain := "test-company"
	industry := "Technology"

	suite.testTenant = &Tenant{
		ID:        suite.testTenantID,
		Name:      "Test Company",
		Slug:      "test-company",
		Email:     "admin@test-company.com",
		Subdomain: &subdomain,
		Status:    StatusActive,
		Industry:  &industry,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (suite *TenantServiceTestSuite) SetupTest() {
	// Test-level setup
	suite.gomockCtrl = gomock.NewController(suite.T())
	suite.mockRepo = NewMockRepository(suite.gomockCtrl)
	suite.mockCache = new(MockCache)
	suite.mockTracer = new(MockTracingService)
	suite.mockSpan = new(MockSpan)

	// Setup default tracing mocks
	suite.mockTracer.On("StartSpan", mock.Anything, mock.Anything, mock.Anything).Return(
		context.Background(), suite.mockSpan)
	suite.mockSpan.On("End", mock.Anything).Return()
	suite.mockSpan.On("SetAttributes", mock.Anything).Return()
	suite.mockSpan.On("RecordError", mock.Anything, mock.Anything).Return()
	suite.mockSpan.On("SetStatus", mock.Anything, mock.Anything).Return()

	suite.service = NewService(suite.mockRepo, suite.mockCache, suite.mockTracer)
	suite.ctx = context.Background()
}

func (suite *TenantServiceTestSuite) TearDownTest() {
	// Test-level cleanup
	if suite.mockCache != nil {
		suite.mockCache.AssertExpectations(suite.T())
	}
	if suite.gomockCtrl != nil {
		suite.gomockCtrl.Finish()
	}
}

// ===============================
// CRUD Operation Tests - Table Driven
// ===============================

func (suite *TenantServiceTestSuite) TestCreateTenant() {
	industry := "Finance"

	tests := []struct {
		name          string
		request       CreateTenantRequest
		setupMocks    func()
		expectedError string
		validate      func(*Tenant)
	}{
		{
			name: "Success",
			request: CreateTenantRequest{
				Name:      "New Company",
				Slug:      "new-company",
				Email:     "admin@new-company.com",
				Subdomain: strPtr("new-company"),
				Industry:  &industry,
				Status:    StatusActive,
			},
			setupMocks: func() {
				suite.mockRepo.EXPECT().GetBySubdomain(suite.ctx, "new-company").Return(
					nil, sharedErrors.ErrTenantNotFound)
				suite.mockRepo.EXPECT().Create(suite.ctx, gomock.Any()).Return(nil)
				// Cache expectations for tenant object and UUID mapping
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*tenant.Tenant"), 30*time.Minute).Return(nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"), 30*time.Minute).Return(nil)
			},
			validate: func(tenant *Tenant) {
				suite.Equal("New Company", tenant.Name)
				suite.Equal("admin@new-company.com", tenant.Email)
				suite.NotEqual(uuid.Nil, tenant.ID)
			},
		},
		{
			name: "SubdomainExists",
			request: CreateTenantRequest{
				Name:      "New Company",
				Email:     "admin@new-company.com",
				Subdomain: strPtr("existing-company"),
			},
			setupMocks: func() {
				existingTenant := &Tenant{ID: uuid.New(), Name: "Existing Company"}
				suite.mockRepo.EXPECT().GetBySubdomain(suite.ctx, "existing-company").Return(
					existingTenant, nil)
			},
			expectedError: "Subdomain already exists",
		},
		{
			name: "ValidationError",
			request: CreateTenantRequest{
				Name:  "", // Invalid
				Email: "admin@test.com",
			},
			setupMocks:    func() {},
			expectedError: "Tenant name is required",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupMocks()

			tenant, err := suite.service.CreateTenant(suite.ctx, tt.request)

			if tt.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tt.expectedError)
				suite.Nil(tenant)
			} else {
				suite.NoError(err)
				suite.NotNil(tenant)
				if tt.validate != nil {
					tt.validate(tenant)
				}
			}
		})
	}
}

func (suite *TenantServiceTestSuite) TestGetTenantByID() {
	tests := []struct {
		name          string
		tenantID      uuid.UUID
		setupMocks    func()
		expectedError string
		validate      func(*Tenant)
	}{
		{
			name:     "Success_CacheMiss",
			tenantID: suite.testTenantID,
			setupMocks: func() {
				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*tenant.Tenant")).Return(errors.New("cache miss"))
				suite.mockRepo.EXPECT().GetByID(suite.ctx, suite.testTenantID).Return(
					suite.testTenant, nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					suite.testTenant, 30*time.Minute).Return(nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"), 30*time.Minute).Return(nil)
			},
			validate: func(tenant *Tenant) {
				suite.Equal(suite.testTenant.ID, tenant.ID)
				suite.Equal(suite.testTenant.Name, tenant.Name)
			},
		},
		{
			name:     "Success_CacheHit",
			tenantID: suite.testTenantID,
			setupMocks: func() {
				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*tenant.Tenant")).Run(func(args mock.Arguments) {
					tenant := args.Get(2).(*Tenant)
					*tenant = *suite.testTenant
				}).Return(nil)
			},
			validate: func(tenant *Tenant) {
				suite.Equal(suite.testTenant.ID, tenant.ID)
			},
		},
		{
			name:     "NotFound",
			tenantID: uuid.New(),
			setupMocks: func() {
				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*tenant.Tenant")).Return(errors.New("cache miss"))
				suite.mockRepo.EXPECT().GetByID(suite.ctx, gomock.Any()).Return(
					nil, sharedErrors.ErrTenantNotFound)
			},
			expectedError: "Tenant not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Create fresh mocks for each test case
			suite.gomockCtrl = gomock.NewController(suite.T())
			suite.mockRepo = NewMockRepository(suite.gomockCtrl)
			suite.mockCache = new(MockCache)
			suite.service = NewService(suite.mockRepo, suite.mockCache, suite.mockTracer)

			tt.setupMocks()

			tenant, err := suite.service.GetTenantByID(suite.ctx, tt.tenantID)

			if tt.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tt.expectedError)
				suite.Nil(tenant)
			} else {
				suite.NoError(err)
				suite.NotNil(tenant)
				if tt.validate != nil {
					tt.validate(tenant)
				}
			}

			// Cleanup for this test case
			suite.mockCache.AssertExpectations(suite.T())
			suite.gomockCtrl.Finish()
		})
	}
}

// ===============================
// Tenant Context Management Tests - Table Driven
// ===============================

func (suite *TenantServiceTestSuite) TestSetTenant() {
	tests := []struct {
		name          string
		tenantID      uuid.UUID
		setupMocks    func(tenantID uuid.UUID)
		expectedError string
	}{
		{
			name:     "Success",
			tenantID: suite.testTenantID,
			setupMocks: func(tenantID uuid.UUID) {
				// ValidateTenantAccess calls GetTenantByID, not Exists
				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*tenant.Tenant")).Return(errors.New("cache miss"))
				suite.mockRepo.EXPECT().GetByID(suite.ctx, tenantID).Return(suite.testTenant, nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					suite.testTenant, 30*time.Minute).Return(nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"), 30*time.Minute).Return(nil)
				suite.mockRepo.EXPECT().SetTenant(suite.ctx, tenantID).Return(nil)
			},
		},
		{
			name:     "TenantNotFound",
			tenantID: uuid.New(),
			setupMocks: func(tenantID uuid.UUID) {
				// Mock cache miss for GetTenantByID call
				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*tenant.Tenant")).Return(errors.New("cache miss"))
				// Mock repository call that will return tenant not found
				suite.mockRepo.EXPECT().GetByID(suite.ctx, tenantID).Return(
					nil, sharedErrors.ErrTenantNotFound)
			},
			expectedError: "failed to validate tenant",
		},
		{
			name:     "InactiveTenant",
			tenantID: suite.testTenantID,
			setupMocks: func(tenantID uuid.UUID) {
				inactiveTenant := *suite.testTenant
				inactiveTenant.Status = StatusSuspended

				// ValidateTenantAccess calls GetTenantByID, not Exists
				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*tenant.Tenant")).Return(errors.New("cache miss"))
				suite.mockRepo.EXPECT().GetByID(suite.ctx, tenantID).Return(&inactiveTenant, nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					&inactiveTenant, 30*time.Minute).Return(nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"), 30*time.Minute).Return(nil)
			},
			expectedError: "failed to validate tenant",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Create fresh mocks for each test case
			suite.gomockCtrl = gomock.NewController(suite.T())
			suite.mockRepo = NewMockRepository(suite.gomockCtrl)
			suite.mockCache = new(MockCache)
			suite.service = NewService(suite.mockRepo, suite.mockCache, suite.mockTracer)

			tt.setupMocks(tt.tenantID)

			err := suite.service.SetTenant(suite.ctx, tt.tenantID)

			if tt.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tt.expectedError)
			} else {
				suite.NoError(err)
			}

			// Cleanup for this test case
			suite.mockCache.AssertExpectations(suite.T())
			suite.gomockCtrl.Finish()
		})
	}
}

func (suite *TenantServiceTestSuite) TestGetCurrentTenant() {
	tests := []struct {
		name          string
		setupMocks    func()
		expectedError string
		validate      func(*Tenant)
	}{
		{
			name: "Success",
			setupMocks: func() {
				suite.mockRepo.EXPECT().GetCurrentTenant(suite.ctx).Return(suite.testTenant, nil)
			},
			validate: func(tenant *Tenant) {
				suite.Equal(suite.testTenant.ID, tenant.ID)
				suite.Equal(suite.testTenant.Name, tenant.Name)
			},
		},
		{
			name: "NoContextSet",
			setupMocks: func() {
				suite.mockRepo.EXPECT().GetCurrentTenant(suite.ctx).Return(
					nil, sharedErrors.ErrTenantIDNotInContext)
			},
			expectedError: "Tenant ID not found in request context",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupMocks()

			tenant, err := suite.service.GetCurrentTenant(suite.ctx)

			if tt.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tt.expectedError)
				suite.Nil(tenant)
			} else {
				suite.NoError(err)
				suite.NotNil(tenant)
				if tt.validate != nil {
					tt.validate(tenant)
				}
			}
		})
	}
}

// ===============================
// Utility Tests - Table Driven
// ===============================

func (suite *TenantServiceTestSuite) TestResolveTenantID() {
	tests := []struct {
		name          string
		subdomain     string
		setupMocks    func()
		expectedError string
		validate      func(uuid.UUID)
	}{
		{
			name:      "Success_CacheMiss",
			subdomain: "test-company",
			setupMocks: func() {
				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*uuid.UUID")).Return(errors.New("cache miss"))
				suite.mockRepo.EXPECT().ResolveSubdomainToID(suite.ctx, "test-company").Return(
					suite.testTenantID, nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					suite.testTenantID, 30*time.Minute).Return(nil)
			},
			validate: func(tenantID uuid.UUID) {
				suite.Equal(suite.testTenantID, tenantID)
			},
		},
		{
			name:      "Success_CacheHit",
			subdomain: "test-company",
			setupMocks: func() {
				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*uuid.UUID")).Run(func(args mock.Arguments) {
					tenantID := args.Get(2).(*uuid.UUID)
					*tenantID = suite.testTenantID
				}).Return(nil)
			},
			validate: func(tenantID uuid.UUID) {
				suite.Equal(suite.testTenantID, tenantID)
			},
		},
		{
			name:      "NotFound",
			subdomain: "nonexistent",
			setupMocks: func() {
				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*uuid.UUID")).Return(errors.New("cache miss"))
				suite.mockRepo.EXPECT().ResolveSubdomainToID(suite.ctx, "nonexistent").Return(
					uuid.Nil, sharedErrors.ErrTenantNotFound)
			},
			expectedError: "Tenant not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Create fresh mocks for each test case
			suite.gomockCtrl = gomock.NewController(suite.T())
			suite.mockRepo = NewMockRepository(suite.gomockCtrl)
			suite.mockCache = new(MockCache)
			suite.service = NewService(suite.mockRepo, suite.mockCache, suite.mockTracer)

			tt.setupMocks()

			tenantID, err := suite.service.ResolveTenantID(suite.ctx, tt.subdomain)

			if tt.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tt.expectedError)
				suite.Equal(uuid.Nil, tenantID)
			} else {
				suite.NoError(err)
				if tt.validate != nil {
					tt.validate(tenantID)
				}
			}

			// Cleanup for this test case
			suite.mockCache.AssertExpectations(suite.T())
			suite.gomockCtrl.Finish()
		})
	}
}

func (suite *TenantServiceTestSuite) TestValidateTenantAccess() {
	tests := []struct {
		name          string
		tenantID      uuid.UUID
		setupMocks    func()
		expectedError string
	}{
		{
			name:     "Success",
			tenantID: suite.testTenantID,
			setupMocks: func() {
				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*tenant.Tenant")).Return(errors.New("cache miss"))
				suite.mockRepo.EXPECT().GetByID(suite.ctx, suite.testTenantID).Return(
					suite.testTenant, nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					suite.testTenant, 30*time.Minute).Return(nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"), 30*time.Minute).Return(nil)
			},
		},
		{
			name:     "InactiveTenant",
			tenantID: suite.testTenantID,
			setupMocks: func() {
				inactiveTenant := *suite.testTenant
				inactiveTenant.Status = StatusSuspended

				suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("*tenant.Tenant")).Return(errors.New("cache miss"))
				suite.mockRepo.EXPECT().GetByID(suite.ctx, suite.testTenantID).Return(
					&inactiveTenant, nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					&inactiveTenant, 30*time.Minute).Return(nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"), 30*time.Minute).Return(nil)
			},
			expectedError: "Tenant is not active",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupMocks()

			err := suite.service.ValidateTenantAccess(suite.ctx, tt.tenantID)

			if tt.expectedError != "" {
				suite.Error(err)
				suite.Contains(err.Error(), tt.expectedError)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// ===============================
// Cache Management Tests
// ===============================

func (suite *TenantServiceTestSuite) TestCacheOperations() {
	tests := []struct {
		name       string
		operation  func() error
		setupMocks func()
	}{
		{
			name: "ClearTenantCache",
			operation: func() error {
				return suite.service.ClearTenantCache(suite.ctx, suite.testTenantID)
			},
			setupMocks: func() {
				suite.mockRepo.EXPECT().GetByID(suite.ctx, suite.testTenantID).Return(
					suite.testTenant, nil)
				suite.mockCache.On("Delete", suite.ctx, mock.AnythingOfType("string")).Return(nil).Times(3)
			},
		},
		{
			name: "ClearSubdomainCache",
			operation: func() error {
				return suite.service.ClearSubdomainCache(suite.ctx, "test-company")
			},
			setupMocks: func() {
				suite.mockCache.On("Delete", suite.ctx, mock.AnythingOfType("string")).Return(nil).Times(2)
			},
		},
		{
			name: "WarmupCache",
			operation: func() error {
				return suite.service.WarmupCache(suite.ctx, suite.testTenantID)
			},
			setupMocks: func() {
				suite.mockRepo.EXPECT().GetByID(suite.ctx, suite.testTenantID).Return(
					suite.testTenant, nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					suite.testTenant, 30*time.Minute).Return(nil)
				suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"),
					mock.AnythingOfType("uuid.UUID"), 30*time.Minute).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupMocks()

			err := tt.operation()
			suite.NoError(err)
		})
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}

// TestTenantService runs the test suite
func TestTenantService(t *testing.T) {
	suite.Run(t, new(TenantServiceTestSuite))
}
