package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"awo/internal/core/tenant"
	tenant_repo "awo/internal/core/tenant/repository"
	"awo/internal/platform/cache"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
)

// TestTenantHandler_CreateBasic tests basic tenant creation with proper mocks
func TestTenantHandler_CreateBasic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)
	mockRepo := tenant_repo.NewMockRepository(ctrl)
	mockCache := cache.NewMockService(ctrl)

	// Setup basic expectations
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().RecordError(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetStatus(gomock.Any(), gomock.Any()).AnyTimes()

	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().DebugContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WarnContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// Cache expectations (cache miss for new tenant)
	mockCache.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(cache.ErrCacheMiss).AnyTimes()
	mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Repository expectations - service will call these internally
	// We only set up what the service needs, don't call repository directly
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, t *tenant.Tenant) error {
		// Service calls this internally - simulate successful creation
		return nil
	}).AnyTimes()

	// Service may call these for subdomain validation
	mockRepo.EXPECT().GetBySubdomain(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("tenant not found")).AnyTimes()

	// Create service with mocks
	service := tenant.NewService(tenant.Dependencies{
		Store:  nil,
		Cache:  mockCache,
		Tracer: mockTracer,
		Logger: mockLogger,
	})
	handler := NewTenantHandler(service, mockLogger, mockMetrics, mockTracer)

	// Create Fiber app
	app := fiber.New()
	app.Post("/tenants", handler.Create)

	// Test data
	payload := `{
		"name": "Test Company",
		"email": "test@company.com"
	}`

	// Make request
	req := httptest.NewRequest("POST", "/tenants", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, 201, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	// Check response has expected fields
	assert.Contains(t, response, "id")
	assert.Contains(t, response, "name")
	assert.Contains(t, response, "email")
	assert.Equal(t, "Test Company", response["name"])
	assert.Equal(t, "test@company.com", response["email"])
}

// TestTenantHandler_ListBasic tests basic tenant listing
func TestTenantHandler_ListBasic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)
	mockRepo := tenant_repo.NewMockRepository(ctrl)
	mockCache := cache.NewMockService(ctrl)

	// Setup basic expectations
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()

	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().DebugContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	// Mock repository to return test tenants
	testTenants := []*tenant.Tenant{
		{
			ID:           uuid.New(),
			Name:         "Test Tenant 1",
			Email:        "test1@example.com",
			Status:       tenant.StatusActive,
			Slug:         "test-tenant-1",
			Timezone:     "UTC",
			CurrencyCode: "USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			Name:         "Test Tenant 2",
			Email:        "test2@example.com",
			Status:       tenant.StatusActive,
			Slug:         "test-tenant-2",
			Timezone:     "UTC",
			CurrencyCode: "USD",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Repository expectations - service will call these internally
	mockRepo.EXPECT().
		List(gomock.Any(), 20).
		Return(testTenants, nil).AnyTimes()

	// Create service with mocks
	service := tenant.NewService(tenant.Dependencies{Store: nil, Cache: mockCache, Tracer: mockTracer, Logger: mockLogger})
	handler := NewTenantHandler(service, mockLogger, mockMetrics, mockTracer)

	// Create Fiber app
	app := fiber.New()
	app.Get("/tenants", handler.List)

	// Make request
	req := httptest.NewRequest("GET", "/tenants", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	// Check response structure
	assert.Contains(t, response, "data")
	assert.Contains(t, response, "pagination")

	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
}
