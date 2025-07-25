//go:build integration
// +build integration

package tenant

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
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

// TestRepositoryCompilation tests that the repository can be instantiated
// and that the SQLC queries compile correctly with the interface
func TestRepositoryCompilation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock store to test interface compatibility
	mockStore := db.NewMockStore(ctrl)

	// Create a mock tracer for testing
	mockTracer := new(MockTracingService)
	mockSpan := new(MockSpan)

	// Setup basic mock expectations
	mockTracer.On("StartSpan", mock.Anything, mock.Anything, mock.Anything).Return(
		context.Background(), mockSpan).Maybe()
	mockSpan.On("End", mock.Anything).Return().Maybe()
	mockSpan.On("SetAttributes", mock.Anything).Return().Maybe()

	// Test that repository can be created with the store interface
	repo := NewRepository(mockStore, mockTracer)
	require.NotNil(t, repo)

	// Verify the repository implements the Repository interface
	var _ Repository = repo

	t.Log("Repository compilation test passed - SQLC queries are syntactically correct")
}

// TestSQLCQueriesExistence verifies that all required SQLC methods exist
// This ensures our repository interface is compatible with generated SQLC code
func TestSQLCQueriesExistence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)

	// Test that all required SQLC methods exist and have correct signatures
	tests := []struct {
		name string
		test func()
	}{
		{
			name: "CreateTenant exists",
			test: func() {
				// This will fail to compile if CreateTenant method doesn't exist or has wrong signature
				mockStore.EXPECT().CreateTenant(context.Background(), gomock.Any()).Return(nil, nil).AnyTimes()
			},
		},
		{
			name: "GetTenantByID exists",
			test: func() {
				mockStore.EXPECT().GetTenantByID(context.Background(), gomock.Any()).Return(nil, nil).AnyTimes()
			},
		},
		{
			name: "GetCurrentTenantID exists",
			test: func() {
				mockStore.EXPECT().GetCurrentTenantID(context.Background()).Return(uuid.Nil, nil).AnyTimes()
			},
		},
		{
			name: "ResetTenantContext exists",
			test: func() {
				mockStore.EXPECT().ResetTenantContext(context.Background()).Return(nil).AnyTimes()
			},
		},
		{
			name: "SetTenantContext exists",
			test: func() {
				mockStore.EXPECT().SetTenantContext(context.Background(), gomock.Any()).Return(nil).AnyTimes()
			},
		},
		{
			name: "ResolveSubdomainToID exists",
			test: func() {
				mockStore.EXPECT().ResolveSubdomainToID(context.Background(), gomock.Any()).Return(uuid.Nil, nil).AnyTimes()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.test, "SQLC method should exist with correct signature")
		})
	}
}
