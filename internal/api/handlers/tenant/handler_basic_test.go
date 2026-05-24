package tenant

import (
	"context"
	"encoding/json"
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

	"awo.so/internal/core/tenant"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// TestTenantHandler_CreateBasic tests basic tenant creation with proper mocks
func TestTenantHandler_CreateBasic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)
	mockSvc := tenant.NewMockService(ctrl)

	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().RecordError(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetStatus(gomock.Any(), gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().DebugContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WarnContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

	tenantID := uuid.New()
	now := time.Now()
	mockSvc.EXPECT().CreateTenant(gomock.Any(), gomock.Any()).Return(&tenant.Tenant{
		ID:           tenantID,
		Name:         "Test Company",
		Email:        "test@company.com",
		Status:       tenant.StatusPending,
		Slug:         "test-company",
		Timezone:     "UTC",
		CurrencyCode: "USD",
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil).AnyTimes()

	handler := NewTenantHandler(mockSvc, mockLogger, mockMetrics, mockTracer, nil)

	app := fiber.New()
	app.Post("/tenants", handler.Create)

	payload := `{
		"name": "Test Company",
		"email": "test@company.com",
		"country_code": "US",
		"currency_code": "USD"
	}`

	req := httptest.NewRequest("POST", "/tenants", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 201, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]any
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

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

	mockLogger := logger.NewMockLogger(ctrl)
	mockMetrics := metrics.NewMockMetricsProvider(ctrl)
	mockTracer := tracing.NewMockService(ctrl)
	mockSpan := tracing.NewMockSpan(ctrl)
	mockSvc := tenant.NewMockService(ctrl)

	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(context.Background(), mockSpan).AnyTimes()
	mockSpan.EXPECT().End(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()

	mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().DebugContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WarnContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()

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

	mockSvc.EXPECT().ListTenants(gomock.Any(), gomock.Any()).Return(testTenants, int64(2), nil).AnyTimes()

	handler := NewTenantHandler(mockSvc, mockLogger, mockMetrics, mockTracer, nil)

	app := fiber.New()
	app.Get("/tenants", handler.List)

	req := httptest.NewRequest("GET", "/tenants", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]any
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "pagination")

	data, ok := response["data"].([]any)
	require.True(t, ok)
	assert.Len(t, data, 2)
}
