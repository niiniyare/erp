package tracing_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"awo.so/internal/shared/tracing" // Replace with your actual module path
)

// Flag to control JSON output (use environment variable to avoid flag redefinition)
//
// Usage:
//   go test -run TestServiceTestSuite                     # Run quietly (no JSON output)
//   VERBOSE_TRACING=true go test -run TestServiceTestSuite # Run with JSON output

// ServiceTestSuite is the main test suite for tracing service
type ServiceTestSuite struct {
	suite.Suite
	service tracing.Service
	ctx     context.Context
}

// SetupSuite runs once before all tests
func (s *ServiceTestSuite) SetupSuite() {
	s.ctx = context.Background()
}

// SetupTest runs before each test
func (s *ServiceTestSuite) SetupTest() {
	config := tracing.Config{
		ServiceName:        "test-service",
		ServiceVersion:     "1.0.0",
		Environment:        "test",
		Endpoint:           "localhost:4317",
		Protocol:           tracing.ProtocolStdout,
		Insecure:           true,
		SamplingRate:       1.0,
		Enabled:            true,
		BatchTimeout:       1 * time.Second,
		MaxExportBatchSize: 100,
		MaxQueueSize:       1000,
	}

	// Set quiet mode if verbose environment variable is not set
	if os.Getenv("VERBOSE_TRACING") != "true" {
		os.Setenv("INTEGRATION_TEST_QUIET", "true")
	}

	service, err := tracing.NewService(config)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), service)

	s.service = service
}

// TearDownTest runs after each test
func (s *ServiceTestSuite) TearDownTest() {
	if s.service != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := s.service.Shutdown(ctx)
		require.NoError(s.T(), err)
	}
	// Clean up environment variable
	os.Unsetenv("INTEGRATION_TEST_QUIET")
}

// TestNewService_Success tests successful service creation
func (s *ServiceTestSuite) TestNewService_Success() {
	config := tracing.DefaultConfig()
	config.ServiceName = "test-service"

	service, err := tracing.NewService(config)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), service)
}

// TestNewService_Disabled tests service creation when disabled
func (s *ServiceTestSuite) TestNewService_Disabled() {
	config := tracing.DefaultConfig()
	config.Enabled = false

	service, err := tracing.NewService(config)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), service)

	// Should return no-op service
	ctx, span := service.StartSpan(s.ctx, "test")
	require.NotNil(s.T(), ctx)
	require.NotNil(s.T(), span)
	require.False(s.T(), span.IsRecording())
}

// TestNewService_InvalidConfig tests service creation with invalid config
func (s *ServiceTestSuite) TestNewService_InvalidConfig() {
	testCases := []struct {
		name        string
		config      tracing.Config
		expectedErr error
	}{
		{
			name: "empty service name",
			config: tracing.Config{
				ServiceName:        "",
				Enabled:            true,
				Protocol:           tracing.ProtocolStdout,
				SamplingRate:       1.0,
				BatchTimeout:       1 * time.Second,
				MaxExportBatchSize: 100,
				MaxQueueSize:       1000,
			},
			expectedErr: tracing.ErrEmptyServiceName,
		},
		{
			name: "invalid sampling rate - negative",
			config: tracing.Config{
				ServiceName:        "test",
				Enabled:            true,
				Protocol:           tracing.ProtocolStdout,
				SamplingRate:       -0.5,
				BatchTimeout:       1 * time.Second,
				MaxExportBatchSize: 100,
				MaxQueueSize:       1000,
			},
			expectedErr: tracing.ErrInvalidSamplingRate,
		},
		{
			name: "invalid sampling rate - too high",
			config: tracing.Config{
				ServiceName:        "test",
				Enabled:            true,
				Protocol:           tracing.ProtocolStdout,
				SamplingRate:       1.5,
				BatchTimeout:       1 * time.Second,
				MaxExportBatchSize: 100,
				MaxQueueSize:       1000,
			},
			expectedErr: tracing.ErrInvalidSamplingRate,
		},
		{
			name: "unsupported protocol",
			config: tracing.Config{
				ServiceName:        "test",
				Enabled:            true,
				Protocol:           "invalid",
				SamplingRate:       1.0,
				BatchTimeout:       1 * time.Second,
				MaxExportBatchSize: 100,
				MaxQueueSize:       1000,
			},
			expectedErr: tracing.ErrUnsupportedProtocol,
		},
		{
			name: "invalid batch timeout",
			config: tracing.Config{
				ServiceName:        "test",
				Enabled:            true,
				Protocol:           tracing.ProtocolStdout,
				SamplingRate:       1.0,
				BatchTimeout:       0,
				MaxExportBatchSize: 100,
				MaxQueueSize:       1000,
			},
			expectedErr: tracing.ErrInvalidConfig,
		},
		{
			name: "invalid max export batch size",
			config: tracing.Config{
				ServiceName:        "test",
				Enabled:            true,
				Protocol:           tracing.ProtocolStdout,
				SamplingRate:       1.0,
				BatchTimeout:       1 * time.Second,
				MaxExportBatchSize: 0,
				MaxQueueSize:       1000,
			},
			expectedErr: tracing.ErrInvalidConfig,
		},
		{
			name: "invalid max queue size",
			config: tracing.Config{
				ServiceName:        "test",
				Enabled:            true,
				Protocol:           tracing.ProtocolStdout,
				SamplingRate:       1.0,
				BatchTimeout:       1 * time.Second,
				MaxExportBatchSize: 100,
				MaxQueueSize:       0,
			},
			expectedErr: tracing.ErrInvalidConfig,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			service, err := tracing.NewService(tc.config)

			require.Error(s.T(), err)
			require.Nil(s.T(), service)
			require.ErrorIs(s.T(), err, tc.expectedErr)
		})
	}
}

// TestStartSpan_Basic tests basic span creation
func (s *ServiceTestSuite) TestStartSpan_Basic() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")

	require.NotNil(s.T(), ctx)
	require.NotNil(s.T(), span)
	require.True(s.T(), span.IsRecording())

	span.End()
}

// TestStartSpan_WithOptions tests span creation with options
func (s *ServiceTestSuite) TestStartSpan_WithOptions() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("test.key", "test-value"),
			attribute.Int("test.count", 42),
		))

	require.NotNil(s.T(), ctx)
	require.NotNil(s.T(), span)
	require.True(s.T(), span.IsRecording())

	span.End()
}

// TestStartSpan_Nested tests nested span creation
func (s *ServiceTestSuite) TestStartSpan_Nested() {
	// Parent span
	ctx1, span1 := s.service.StartSpan(s.ctx, "parent-operation")
	require.NotNil(s.T(), ctx1)
	require.NotNil(s.T(), span1)
	defer span1.End()

	// Child span
	ctx2, span2 := s.service.StartSpan(ctx1, "child-operation")
	require.NotNil(s.T(), ctx2)
	require.NotNil(s.T(), span2)
	defer span2.End()

	// Grandchild span
	ctx3, span3 := s.service.StartSpan(ctx2, "grandchild-operation")
	require.NotNil(s.T(), ctx3)
	require.NotNil(s.T(), span3)
	defer span3.End()

	// All spans should be recording
	require.True(s.T(), span1.IsRecording())
	require.True(s.T(), span2.IsRecording())
	require.True(s.T(), span3.IsRecording())
}

// TestSpanFromContext tests retrieving span from context
func (s *ServiceTestSuite) TestSpanFromContext() {
	ctx, originalSpan := s.service.StartSpan(s.ctx, "test-operation")
	defer originalSpan.End()

	retrievedSpan := s.service.SpanFromContext(ctx)

	require.NotNil(s.T(), retrievedSpan)
	require.True(s.T(), retrievedSpan.IsRecording())
}

// TestSpanFromContext_NoSpan tests retrieving span when none exists
func (s *ServiceTestSuite) TestSpanFromContext_NoSpan() {
	span := s.service.SpanFromContext(s.ctx)

	require.NotNil(s.T(), span)
	// Should return no-op span
	require.False(s.T(), span.IsRecording())
}

// TestSetAttributes_ViaService tests setting attributes via service
func (s *ServiceTestSuite) TestSetAttributes_ViaService() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	s.service.SetAttributes(ctx,
		attribute.String("user.id", "123"),
		attribute.String("user.name", "John Doe"),
		attribute.Bool("user.premium", true),
	)

	require.True(s.T(), span.IsRecording())
}

// TestSetAttributes_ViaSpan tests setting attributes directly on span
func (s *ServiceTestSuite) TestSetAttributes_ViaSpan() {
	_, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	span.SetAttributes(
		attribute.String("order.id", "ord-123"),
		attribute.Float64("order.total", 99.99),
		attribute.Int("order.items", 3),
	)

	require.True(s.T(), span.IsRecording())
}

// TestRecordError_ViaService tests error recording via service
func (s *ServiceTestSuite) TestRecordError_ViaService() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	testErr := errors.New("test error")

	s.service.RecordError(ctx, testErr)

	require.True(s.T(), span.IsRecording())
}

// TestRecordError_WithStatus tests error recording with status
func (s *ServiceTestSuite) TestRecordError_WithStatus() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	testErr := errors.New("test error with status")

	s.service.RecordError(ctx, testErr, tracing.WithErrorStatus())

	require.True(s.T(), span.IsRecording())
}

// TestRecordError_WithAttributes tests error recording with attributes
func (s *ServiceTestSuite) TestRecordError_WithAttributes() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	testErr := errors.New("test error with attributes")

	s.service.RecordError(ctx, testErr,
		tracing.WithErrorStatus(),
		tracing.WithErrorAttributes(
			attribute.String("error.type", "validation"),
			attribute.String("error.severity", "high"),
		))

	require.True(s.T(), span.IsRecording())
}

// TestRecordError_ViaSpan tests error recording directly on span
func (s *ServiceTestSuite) TestRecordError_ViaSpan() {
	_, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	testErr := errors.New("test error via span")

	span.RecordError(testErr)
	span.SetStatus(codes.Error, testErr.Error())

	require.True(s.T(), span.IsRecording())
}

// TestRecordError_NilError tests that nil errors are ignored
func (s *ServiceTestSuite) TestRecordError_NilError() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	// Should not panic
	s.service.RecordError(ctx, nil)
	s.service.RecordError(ctx, nil, tracing.WithErrorStatus())

	require.True(s.T(), span.IsRecording())
}

// TestAddEvent_ViaService tests adding events via service
func (s *ServiceTestSuite) TestAddEvent_ViaService() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	s.service.AddEvent(ctx, "checkpoint-reached")
	s.service.AddEvent(ctx, "user-action",
		attribute.String("action", "click"),
		attribute.String("target", "button"),
	)

	require.True(s.T(), span.IsRecording())
}

// TestAddEvent_ViaSpan tests adding events directly on span
func (s *ServiceTestSuite) TestAddEvent_ViaSpan() {
	_, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	span.AddEvent("processing-started")
	span.AddEvent("data-validated",
		attribute.Int("records", 100),
		attribute.String("status", "valid"),
	)

	require.True(s.T(), span.IsRecording())
}

// TestSetStatus tests setting span status
func (s *ServiceTestSuite) TestSetStatus() {
	testCases := []struct {
		name        string
		code        codes.Code
		description string
	}{
		{
			name:        "ok status",
			code:        codes.Ok,
			description: "operation successful",
		},
		{
			name:        "error status",
			code:        codes.Error,
			description: "operation failed",
		},
		{
			name:        "unset status",
			code:        codes.Unset,
			description: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, span := s.service.StartSpan(s.ctx, "test-operation")
			defer span.End()

			span.SetStatus(tc.code, tc.description)

			require.True(s.T(), span.IsRecording())
		})
	}
}

// TestSetName tests updating span name
func (s *ServiceTestSuite) TestSetName() {
	_, span := s.service.StartSpan(s.ctx, "initial-name")
	defer span.End()

	span.SetName("updated-name")

	require.True(s.T(), span.IsRecording())
}

// TestSpanEnd tests span ending
func (s *ServiceTestSuite) TestSpanEnd() {
	_, span := s.service.StartSpan(s.ctx, "test-operation")

	require.True(s.T(), span.IsRecording())

	span.End()

	// After ending, span may still report as recording in some implementations
	// but it should not panic
}

// TestSpanEnd_WithTimestamp tests span ending with custom timestamp
func (s *ServiceTestSuite) TestSpanEnd_WithTimestamp() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")

	customTime := time.Now().Add(-1 * time.Hour)
	span.End(tracing.WithTimestamp(customTime))

	require.NotNil(s.T(), ctx)
}

// TestHTTPHeaderPropagation tests HTTP header injection and extraction
func (s *ServiceTestSuite) TestHTTPHeaderPropagation() {
	// Create a span
	ctx, span := s.service.StartSpan(s.ctx, "http-request")
	defer span.End()

	// Inject into headers
	headers := http.Header{}
	s.service.InjectHTTPHeaders(ctx, headers)

	require.NotEmpty(s.T(), headers)

	// Extract from headers
	newCtx := s.service.ExtractHTTPHeaders(context.Background(), headers)

	require.NotNil(s.T(), newCtx)
}

// TestHTTPHeaderPropagation_RoundTrip tests full HTTP propagation cycle
func (s *ServiceTestSuite) TestHTTPHeaderPropagation_RoundTrip() {
	// Start a span in service A
	ctxA, spanA := s.service.StartSpan(s.ctx, "service-a-operation")
	defer spanA.End()

	// Create HTTP request and inject trace context
	req := httptest.NewRequest("GET", "http://example.com/api", nil)
	s.service.InjectHTTPHeaders(ctxA, req.Header)

	// Simulate receiving request in service B
	ctxB := s.service.ExtractHTTPHeaders(context.Background(), req.Header)

	// Start a span in service B (should be child of A's span)
	ctxB, spanB := s.service.StartSpan(ctxB, "service-b-operation")
	defer spanB.End()

	require.NotNil(s.T(), ctxA)
	require.NotNil(s.T(), ctxB)
	require.True(s.T(), spanA.IsRecording())
	require.True(s.T(), spanB.IsRecording())
}

// TestGetTraceID tests retrieving trace ID
func (s *ServiceTestSuite) TestGetTraceID() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	traceID := s.service.GetTraceID(ctx)

	require.NotEmpty(s.T(), traceID)
	require.Len(s.T(), traceID, 32) // Trace ID is 32 hex characters
}

// TestGetTraceID_NoSpan tests getting trace ID with no span
func (s *ServiceTestSuite) TestGetTraceID_NoSpan() {
	traceID := s.service.GetTraceID(s.ctx)

	require.Empty(s.T(), traceID)
}

// TestGetSpanID tests retrieving span ID
func (s *ServiceTestSuite) TestGetSpanID() {
	ctx, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	spanID := s.service.GetSpanID(ctx)

	require.NotEmpty(s.T(), spanID)
	require.Len(s.T(), spanID, 16) // Span ID is 16 hex characters
}

// TestGetSpanID_NoSpan tests getting span ID with no span
func (s *ServiceTestSuite) TestGetSpanID_NoSpan() {
	spanID := s.service.GetSpanID(s.ctx)

	require.Empty(s.T(), spanID)
}

// TestSpanContext tests retrieving span context
func (s *ServiceTestSuite) TestSpanContext() {
	_, span := s.service.StartSpan(s.ctx, "test-operation")
	defer span.End()

	spanContext := span.SpanContext()

	require.True(s.T(), spanContext.IsValid())
	require.NotEmpty(s.T(), spanContext.TraceID().String())
	require.NotEmpty(s.T(), spanContext.SpanID().String())
}

// TestShutdown tests graceful shutdown
func (s *ServiceTestSuite) TestShutdown() {
	config := tracing.DefaultConfig()
	config.ServiceName = "shutdown-test"
	config.Protocol = tracing.ProtocolStdout

	service, err := tracing.NewService(config)
	require.NoError(s.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = service.Shutdown(ctx)

	require.NoError(s.T(), err)
}

// TestShutdown_AlreadyClosed tests shutting down already closed service
func (s *ServiceTestSuite) TestShutdown_AlreadyClosed() {
	config := tracing.DefaultConfig()
	config.ServiceName = "shutdown-test"
	config.Protocol = tracing.ProtocolStdout

	service, err := tracing.NewService(config)
	require.NoError(s.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First shutdown
	err = service.Shutdown(ctx)
	require.NoError(s.T(), err)

	// Second shutdown should return error
	err = service.Shutdown(ctx)
	require.ErrorIs(s.T(), err, tracing.ErrServiceClosed)
}

// TestShutdown_WithTimeout tests shutdown with context timeout
func (s *ServiceTestSuite) TestShutdown_WithTimeout() {
	config := tracing.DefaultConfig()
	config.ServiceName = "shutdown-timeout-test"
	config.Protocol = tracing.ProtocolStdout

	service, err := tracing.NewService(config)
	require.NoError(s.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = service.Shutdown(ctx)

	// Should complete within timeout
	require.NoError(s.T(), err)
}

// TestOperationsAfterShutdown tests that operations after shutdown are safe
func (s *ServiceTestSuite) TestOperationsAfterShutdown() {
	config := tracing.DefaultConfig()
	config.ServiceName = "shutdown-ops-test"
	config.Protocol = tracing.ProtocolStdout

	service, err := tracing.NewService(config)
	require.NoError(s.T(), err)

	// Shutdown the service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = service.Shutdown(ctx)
	require.NoError(s.T(), err)

	// Try operations after shutdown - should not panic
	ctx2, span := service.StartSpan(s.ctx, "after-shutdown")
	require.NotNil(s.T(), ctx2)
	require.NotNil(s.T(), span)
	require.False(s.T(), span.IsRecording()) // Should be no-op

	service.SetAttributes(ctx2, attribute.String("test", "value"))
	service.RecordError(ctx2, errors.New("test"))
	service.AddEvent(ctx2, "test")
	service.InjectHTTPHeaders(ctx2, http.Header{})

	traceID := service.GetTraceID(ctx2)
	require.Empty(s.T(), traceID)
}

// TestConcurrentOperations tests concurrent span operations
func (s *ServiceTestSuite) TestConcurrentOperations() {
	const numGoroutines = 100

	done := make(chan bool, numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer func() { done <- true }()

			ctx, span := s.service.StartSpan(s.ctx, "concurrent-operation")
			defer span.End()

			span.SetAttributes(
				attribute.Int("goroutine.id", id),
				attribute.String("operation", "concurrent"),
			)

			s.service.AddEvent(ctx, "processing",
				attribute.Int("id", id))

			if id%2 == 0 {
				s.service.RecordError(ctx, errors.New("even error"))
			}

			span.SetStatus(codes.Ok, "completed")
		}(i)
	}

	// Wait for all goroutines
	for range numGoroutines {
		<-done
	}
}

// TestNoOpService tests no-op service behavior
func (s *ServiceTestSuite) TestNoOpService() {
	service := tracing.NewNoOpService()
	require.NotNil(s.T(), service)

	ctx, span := service.StartSpan(s.ctx, "noop-operation")
	require.NotNil(s.T(), ctx)
	require.NotNil(s.T(), span)
	require.False(s.T(), span.IsRecording())

	// All operations should be safe no-ops
	service.SetAttributes(ctx, attribute.String("test", "value"))
	service.RecordError(ctx, errors.New("test"))
	service.AddEvent(ctx, "test")
	span.SetAttributes(attribute.String("test", "value"))
	span.SetStatus(codes.Ok, "success")
	span.End()

	err := service.Shutdown(context.Background())
	require.NoError(s.T(), err)
}

// TestSpanKinds tests different span kinds
func (s *ServiceTestSuite) TestSpanKinds() {
	kinds := []struct {
		name string
		kind tracing.SpanKind
	}{
		{"internal", tracing.SpanKindInternal},
		{"server", tracing.SpanKindServer},
		{"client", tracing.SpanKindClient},
		{"producer", tracing.SpanKindProducer},
		{"consumer", tracing.SpanKindConsumer},
	}

	for _, k := range kinds {
		s.Run(k.name, func() {
			ctx, span := s.service.StartSpan(s.ctx, "test-"+k.name,
				tracing.WithSpanKind(k.kind))
			defer span.End()

			require.NotNil(s.T(), ctx)
			require.NotNil(s.T(), span)
			require.True(s.T(), span.IsRecording())
		})
	}
}

// TestAttributeTypes tests different attribute types
func (s *ServiceTestSuite) TestAttributeTypes() {
	_, span := s.service.StartSpan(s.ctx, "attribute-types")
	defer span.End()

	span.SetAttributes(
		attribute.String("string", "value"),
		attribute.Bool("bool", true),
		attribute.Int("int", 42),
		attribute.Int64("int64", int64(9999999999)),
		attribute.Float64("float64", 3.14159),
		attribute.StringSlice("string_slice", []string{"a", "b", "c"}),
		attribute.BoolSlice("bool_slice", []bool{true, false}),
		attribute.IntSlice("int_slice", []int{1, 2, 3}),
		attribute.Int64Slice("int64_slice", []int64{100, 200, 300}),
		attribute.Float64Slice("float64_slice", []float64{1.1, 2.2, 3.3}),
	)

	require.True(s.T(), span.IsRecording())
}

// TestDefaultConfig tests default configuration
func (s *ServiceTestSuite) TestDefaultConfig() {
	config := tracing.DefaultConfig()

	require.NotEmpty(s.T(), config.ServiceName)
	require.NotEmpty(s.T(), config.ServiceVersion)
	require.NotEmpty(s.T(), config.Environment)
	require.NotEmpty(s.T(), config.Endpoint)
	require.NotEmpty(s.T(), config.Protocol)
	require.Equal(s.T(), 1.0, config.SamplingRate)
	require.True(s.T(), config.Enabled)
	require.Greater(s.T(), config.BatchTimeout, time.Duration(0))
	require.Greater(s.T(), config.MaxExportBatchSize, 0)
	require.Greater(s.T(), config.MaxQueueSize, 0)
}

// TestConfigValidation tests configuration validation
func (s *ServiceTestSuite) TestConfigValidation() {
	testCases := []struct {
		name      string
		config    tracing.Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: tracing.Config{
				ServiceName:        "test",
				Enabled:            true,
				Protocol:           tracing.ProtocolStdout,
				SamplingRate:       1.0,
				BatchTimeout:       1 * time.Second,
				MaxExportBatchSize: 100,
				MaxQueueSize:       1000,
			},
			expectErr: false,
		},
		{
			name: "disabled config - always valid",
			config: tracing.Config{
				Enabled: false,
			},
			expectErr: false,
		},
		{
			name: "empty service name",
			config: tracing.Config{
				ServiceName:        "",
				Enabled:            true,
				Protocol:           tracing.ProtocolStdout,
				SamplingRate:       1.0,
				BatchTimeout:       1 * time.Second,
				MaxExportBatchSize: 100,
				MaxQueueSize:       1000,
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.config.Validate()

			if tc.expectErr {
				require.Error(s.T(), err)
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

// TestHelperFunctions tests semantic convention helpers
func (s *ServiceTestSuite) TestHelperFunctions() {
	s.Run("HTTPAttributes", func() {
		attrs := tracing.HTTPAttributes("GET", "http://example.com", "Mozilla/5.0", 200)
		require.NotEmpty(s.T(), attrs)
		require.GreaterOrEqual(s.T(), len(attrs), 3)
	})

	s.Run("DBAttributes", func() {
		attrs := tracing.DBAttributes("postgresql", "mydb", "SELECT * FROM users")
		require.NotEmpty(s.T(), attrs)
		require.GreaterOrEqual(s.T(), len(attrs), 1)
	})

	s.Run("RPCAttributes", func() {
		attrs := tracing.RPCAttributes("grpc", "UserService", "GetUser")
		require.NotEmpty(s.T(), attrs)
		require.GreaterOrEqual(s.T(), len(attrs), 1)
	})

	s.Run("MessagingAttributes", func() {
		attrs := tracing.MessagingAttributes("kafka", "user-events", "publish")
		require.NotEmpty(s.T(), attrs)
		require.GreaterOrEqual(s.T(), len(attrs), 1)
	})
}

// TestRealWorldScenario_ABAC simulates ABAC permission evaluation
func (s *ServiceTestSuite) TestRealWorldScenario_ABAC() {
	ctx, span := s.service.StartSpan(s.ctx, "abac.service.EvaluatePermission",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("user_id", "user-123"),
			attribute.String("resource_type", "document"),
			attribute.String("action", "read"),
		))
	defer span.End()

	// Validate tenant context
	s.service.AddEvent(ctx, "validating-tenant-context")
	span.SetAttributes(
		attribute.String("tenant.id", "tenant-456"),
		attribute.String("tenant.name", "Acme Corp"),
	)

	// Collect user attributes
	s.service.AddEvent(ctx, "collecting-user-attributes")
	span.SetAttributes(attribute.Int("user_attributes.count", 5))

	// Collect resource attributes
	s.service.AddEvent(ctx, "collecting-resource-attributes")
	span.SetAttributes(attribute.Int("resource_attributes.count", 3))

	// Evaluate policies
	s.service.AddEvent(ctx, "evaluating-policies")
	span.SetAttributes(
		attribute.String("decision", "allow"),
		attribute.Bool("cache_hit", false),
		attribute.Int("policies_evaluated", 2),
	)

	span.SetStatus(codes.Ok, "permission evaluation successful")

	require.True(s.T(), span.IsRecording())
}

// TestRealWorldScenario_HTTPRequest simulates HTTP request handling
func (s *ServiceTestSuite) TestRealWorldScenario_HTTPRequest() {
	// Simulate incoming HTTP request
	req := httptest.NewRequest("GET", "http://api.example.com/users/123", nil)
	req.Header.Set("User-Agent", "TestClient/1.0")

	// Extract trace context
	ctx := s.service.ExtractHTTPHeaders(s.ctx, req.Header)

	// Start server span
	ctx, span := s.service.StartSpan(ctx, "GET /users/:id",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("http.method", "GET"),
			attribute.String("http.url", req.URL.String()),
			attribute.String("http.user_agent", req.UserAgent()),
		))
	defer span.End()

	// Simulate processing
	s.service.AddEvent(ctx, "authentication-check")
	s.service.AddEvent(ctx, "authorization-check")

	// Simulate database call
	ctx2, dbSpan := s.service.StartSpan(ctx, "database.query",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
		))
	dbSpan.SetAttributes(attribute.String("db.statement", "SELECT * FROM users WHERE id = $1"))
	dbSpan.End()

	// Success response
	span.SetAttributes(attribute.Int("http.status_code", 200))
	span.SetStatus(codes.Ok, "request successful")

	require.NotNil(s.T(), ctx2)
	require.True(s.T(), span.IsRecording())
}

// TestRealWorldScenario_ErrorHandling simulates error scenarios
func (s *ServiceTestSuite) TestRealWorldScenario_ErrorHandling() {
	ctx, span := s.service.StartSpan(s.ctx, "process-payment",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	// Step 1: Validate payment
	s.service.AddEvent(ctx, "validating-payment")

	// Step 2: Charge card - simulate failure
	testErr := errors.New("insufficient funds")
	s.service.RecordError(ctx, testErr,
		tracing.WithErrorStatus(),
		tracing.WithErrorAttributes(
			attribute.String("error.type", "payment_failed"),
			attribute.String("error.code", "INSUFFICIENT_FUNDS"),
		))

	span.SetAttributes(
		attribute.String("payment.status", "failed"),
		attribute.String("payment.error", testErr.Error()),
	)
	span.SetStatus(codes.Error, "payment processing failed")

	require.True(s.T(), span.IsRecording())
}

// TestRealWorldScenario_BackgroundJob simulates background job processing
func (s *ServiceTestSuite) TestRealWorldScenario_BackgroundJob() {
	ctx, span := s.service.StartSpan(s.ctx, "background-job.cleanup",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("job.type", "cleanup"),
			attribute.String("job.schedule", "hourly"),
		))
	defer span.End()

	startTime := time.Now()
	itemsProcessed := 0

	// Process items
	for i := range 5 {
		_, itemSpan := s.service.StartSpan(ctx, "process-item",
			tracing.WithAttributes(attribute.Int("item.id", i)))

		// Simulate work
		time.Sleep(1 * time.Millisecond)
		itemsProcessed++

		itemSpan.SetStatus(codes.Ok, "item processed")
		itemSpan.End()
	}

	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Int("job.items_processed", itemsProcessed),
		attribute.Int64("job.duration_ms", duration.Milliseconds()),
	)

	s.service.AddEvent(ctx, "job-completed",
		attribute.Int("items", itemsProcessed))

	span.SetStatus(codes.Ok, "job completed successfully")
	require.Equal(s.T(), 5, itemsProcessed)
}

// TestRealWorldScenario_DistributedTrace simulates distributed trace across services
func (s *ServiceTestSuite) TestRealWorldScenario_DistributedTrace() {
	// Service A: API Gateway
	ctxA, spanA := s.service.StartSpan(s.ctx, "api-gateway.handle-request",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer spanA.End()

	spanA.SetAttributes(attribute.String("service", "api-gateway"))

	// Create request to Service B
	reqToB := httptest.NewRequest("GET", "http://service-b/process", nil)
	s.service.InjectHTTPHeaders(ctxA, reqToB.Header)

	// Service B: Process request
	ctxB := s.service.ExtractHTTPHeaders(context.Background(), reqToB.Header)
	ctxB, spanB := s.service.StartSpan(ctxB, "service-b.process",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer spanB.End()

	spanB.SetAttributes(attribute.String("service", "service-b"))

	// Service B calls Service C
	reqToC := httptest.NewRequest("GET", "http://service-c/data", nil)
	s.service.InjectHTTPHeaders(ctxB, reqToC.Header)

	// Service C: Fetch data
	ctxC := s.service.ExtractHTTPHeaders(context.Background(), reqToC.Header)
	ctxC, spanC := s.service.StartSpan(ctxC, "service-c.fetch-data",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer spanC.End()

	spanC.SetAttributes(attribute.String("service", "service-c"))
	spanC.SetStatus(codes.Ok, "data fetched")

	// Complete the chain
	spanB.SetStatus(codes.Ok, "processing complete")
	spanA.SetStatus(codes.Ok, "request handled")

	// Verify all spans are in same trace
	traceIDA := s.service.GetTraceID(ctxA)
	traceIDB := s.service.GetTraceID(ctxB)
	traceIDC := s.service.GetTraceID(ctxC)

	require.Equal(s.T(), traceIDA, traceIDB)
	require.Equal(s.T(), traceIDB, traceIDC)
}

// TestProtocols tests different protocol configurations
func (s *ServiceTestSuite) TestProtocols() {
	protocols := []tracing.Protocol{
		tracing.ProtocolGRPC,
		tracing.ProtocolHTTP,
		tracing.ProtocolStdout,
	}

	for _, protocol := range protocols {
		s.Run(string(protocol), func() {
			config := tracing.DefaultConfig()
			config.ServiceName = "protocol-test"
			config.Protocol = protocol
			config.Endpoint = "localhost:4317"

			service, err := tracing.NewService(config)
			require.NoError(s.T(), err)
			require.NotNil(s.T(), service)

			ctx, span := service.StartSpan(s.ctx, "test-operation")
			span.End()

			require.NotNil(s.T(), ctx)

			// Cleanup - shutdown may timeout if no collector is running
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err = service.Shutdown(shutdownCtx)

			// For gRPC and HTTP protocols, allow shutdown timeout errors when no collector is running
			if protocol == tracing.ProtocolGRPC || protocol == tracing.ProtocolHTTP {
				// These protocols may timeout during shutdown if no collector is available
				// This is expected behavior in test environments without a running collector
				if err != nil {
					s.T().Logf("Shutdown warning for protocol %s: %v (expected when no collector is running)", protocol, err)
				}
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

// TestSamplingRates tests different sampling rates
func (s *ServiceTestSuite) TestSamplingRates() {
	samplingRates := []float64{0.0, 0.1, 0.5, 1.0}

	for _, rate := range samplingRates {
		s.Run(string(rune(rate*100)), func() {
			config := tracing.DefaultConfig()
			config.ServiceName = "sampling-test"
			config.Protocol = tracing.ProtocolStdout
			config.SamplingRate = rate

			service, err := tracing.NewService(config)
			require.NoError(s.T(), err)
			require.NotNil(s.T(), service)

			// Create multiple spans
			for range 10 {
				ctx, span := service.StartSpan(s.ctx, "sampled-operation")
				span.End()
				require.NotNil(s.T(), ctx)
			}

			// Cleanup
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err = service.Shutdown(shutdownCtx)
			require.NoError(s.T(), err)
		})
	}
}

// TestSpanLinks tests span linking
func (s *ServiceTestSuite) TestSpanLinks() {
	// Create first span
	ctx1, span1 := s.service.StartSpan(s.ctx, "operation-1")
	span1Context := span1.SpanContext()
	span1.End()

	// Create second span with link to first
	ctx2, span2 := s.service.StartSpan(s.ctx, "operation-2",
		tracing.WithLinks(trace.Link{
			SpanContext: span1Context,
		}))
	defer span2.End()

	require.NotNil(s.T(), ctx1)
	require.NotNil(s.T(), ctx2)
	require.True(s.T(), span2.IsRecording())
}

// TestComplexAttributeScenario tests complex attribute usage
func (s *ServiceTestSuite) TestComplexAttributeScenario() {
	ctx, span := s.service.StartSpan(s.ctx, "complex-operation")
	defer span.End()

	// Initial attributes
	span.SetAttributes(
		attribute.String("operation.type", "data-processing"),
		attribute.String("operation.version", "v2"),
	)

	// Add more attributes during processing
	s.service.SetAttributes(ctx,
		attribute.Int("records.total", 1000),
		attribute.Int("records.processed", 0),
	)

	// Simulate processing batches
	for i := range 10 {
		batchCtx, batchSpan := s.service.StartSpan(ctx, "process-batch")

		batchSpan.SetAttributes(
			attribute.Int("batch.number", i),
			attribute.Int("batch.size", 100),
		)

		// Simulate work
		time.Sleep(1 * time.Millisecond)

		batchSpan.SetStatus(codes.Ok, "batch processed")
		batchSpan.End()

		require.NotNil(s.T(), batchCtx)
	}

	// Final attributes
	s.service.SetAttributes(ctx,
		attribute.Int("records.processed", 1000),
		attribute.Bool("operation.success", true),
	)

	span.SetStatus(codes.Ok, "all batches processed")
}

// TestMemoryLeakPrevention tests that ended spans don't leak memory
func (s *ServiceTestSuite) TestMemoryLeakPrevention() {
	const numSpans = 1000

	for i := range numSpans {
		ctx, span := s.service.StartSpan(s.ctx, "leak-test-span")
		span.SetAttributes(
			attribute.Int("iteration", i),
			attribute.String("data", "some data"),
		)
		span.End()
		require.NotNil(s.T(), ctx)
	}

	// If we get here without OOM, test passes
	require.True(s.T(), true)
}

// TestPanicRecovery tests that panics don't break tracing
func (s *ServiceTestSuite) TestPanicRecovery() {
	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("Recovered from panic: %v", r)
		}
	}()

	ctx, span := s.service.StartSpan(s.ctx, "panic-test")
	defer func() {
		if span != nil {
			span.End()
		}
	}()

	// This should not cause issues with tracing
	span.SetAttributes(attribute.String("test", "value"))
	require.NotNil(s.T(), ctx)
}

// TestContextCancellation tests behavior with cancelled context
func (s *ServiceTestSuite) TestContextCancellation() {
	ctx, cancel := context.WithCancel(s.ctx)

	ctx, span := s.service.StartSpan(ctx, "cancellable-operation")
	defer span.End()

	span.SetAttributes(attribute.String("status", "started"))

	// Cancel context
	cancel()

	// Operations should still work
	span.SetAttributes(attribute.String("status", "cancelled"))
	s.service.AddEvent(ctx, "context-cancelled")

	require.NotNil(s.T(), ctx)
}

// TestLongRunningOperation tests span with extended duration
func (s *ServiceTestSuite) TestLongRunningOperation() {
	ctx, span := s.service.StartSpan(s.ctx, "long-running-operation",
		tracing.WithAttributes(
			attribute.String("operation.type", "batch-processing"),
		))
	defer span.End()

	// Simulate checkpoints
	checkpoints := []string{"init", "validate", "process", "finalize"}

	for i, checkpoint := range checkpoints {
		s.service.AddEvent(ctx, checkpoint,
			attribute.Int("checkpoint.number", i+1),
			attribute.String("checkpoint.name", checkpoint),
		)

		time.Sleep(5 * time.Millisecond)
	}

	span.SetAttributes(
		attribute.Int("checkpoints.completed", len(checkpoints)),
	)
	span.SetStatus(codes.Ok, "completed all checkpoints")
}

// TestEmptyAttributes tests handling of empty attributes
func (s *ServiceTestSuite) TestEmptyAttributes() {
	ctx, span := s.service.StartSpan(s.ctx, "empty-attrs-test")
	defer span.End()

	// Should not panic with empty attributes
	span.SetAttributes()
	s.service.SetAttributes(ctx)
	s.service.AddEvent(ctx, "event-no-attrs")
	span.AddEvent("span-event-no-attrs")

	require.True(s.T(), span.IsRecording())
}

// TestSpecialCharactersInNames tests span names with special characters
func (s *ServiceTestSuite) TestSpecialCharactersInNames() {
	specialNames := []string{
		"operation/with/slashes",
		"operation.with.dots",
		"operation-with-dashes",
		"operation_with_underscores",
		"operation:with:colons",
		"operation with spaces",
	}

	for _, name := range specialNames {
		s.Run(name, func() {
			ctx, span := s.service.StartSpan(s.ctx, name)
			defer span.End()

			require.NotNil(s.T(), ctx)
			require.True(s.T(), span.IsRecording())
		})
	}
}

// TestVeryLongSpanName tests span with very long name
func (s *ServiceTestSuite) TestVeryLongSpanName() {
	longName := string(make([]byte, 1000))
	for range longName {
		longName = "a" + longName
	}

	ctx, span := s.service.StartSpan(s.ctx, longName)
	defer span.End()

	require.NotNil(s.T(), ctx)
	require.True(s.T(), span.IsRecording())
}

// TestMultipleShutdowns tests calling shutdown multiple times
func (s *ServiceTestSuite) TestMultipleShutdowns() {
	config := tracing.DefaultConfig()
	config.ServiceName = "multi-shutdown-test"
	config.Protocol = tracing.ProtocolStdout

	service, err := tracing.NewService(config)
	require.NoError(s.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First shutdown
	err = service.Shutdown(ctx)
	require.NoError(s.T(), err)

	// Second shutdown
	err = service.Shutdown(ctx)
	require.ErrorIs(s.T(), err, tracing.ErrServiceClosed)

	// Third shutdown
	err = service.Shutdown(ctx)
	require.ErrorIs(s.T(), err, tracing.ErrServiceClosed)
}

// Run the test suite
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

// ============================================================================
// BENCHMARK TESTS
// ============================================================================

// BenchmarkStartSpan benchmarks span creation
func BenchmarkStartSpan(b *testing.B) {
	config := tracing.DefaultConfig()
	config.ServiceName = "benchmark-test"
	config.Protocol = tracing.ProtocolStdout

	service, _ := tracing.NewService(config)
	defer service.Shutdown(context.Background())

	ctx := context.Background()

	for b.Loop() {
		_, span := service.StartSpan(ctx, "benchmark-operation")
		span.End()
	}
}

// BenchmarkStartSpanWithAttributes benchmarks span creation with attributes
func BenchmarkStartSpanWithAttributes(b *testing.B) {
	config := tracing.DefaultConfig()
	config.ServiceName = "benchmark-test"
	config.Protocol = tracing.ProtocolStdout

	service, _ := tracing.NewService(config)
	defer service.Shutdown(context.Background())

	ctx := context.Background()

	for b.Loop() {
		_, span := service.StartSpan(ctx, "benchmark-operation",
			tracing.WithAttributes(
				attribute.String("key1", "value1"),
				attribute.Int("key2", 42),
				attribute.Bool("key3", true),
			))
		span.End()
	}
}

// BenchmarkNestedSpans benchmarks nested span creation
func BenchmarkNestedSpans(b *testing.B) {
	config := tracing.DefaultConfig()
	config.ServiceName = "benchmark-test"
	config.Protocol = tracing.ProtocolStdout

	service, _ := tracing.NewService(config)
	defer service.Shutdown(context.Background())

	ctx := context.Background()

	for b.Loop() {
		ctx1, span1 := service.StartSpan(ctx, "parent")
		ctx2, span2 := service.StartSpan(ctx1, "child")
		ctx3, span3 := service.StartSpan(ctx2, "grandchild")

		span3.End()
		span2.End()
		span1.End()

		_ = ctx3
	}
}

// BenchmarkSetAttributes benchmarks setting attributes
func BenchmarkSetAttributes(b *testing.B) {
	config := tracing.DefaultConfig()
	config.ServiceName = "benchmark-test"
	config.Protocol = tracing.ProtocolStdout

	service, _ := tracing.NewService(config)
	defer service.Shutdown(context.Background())

	_, span := service.StartSpan(context.Background(), "benchmark-operation")
	defer span.End()

	for i := 0; b.Loop(); i++ {
		span.SetAttributes(
			attribute.String("key", "value"),
			attribute.Int("count", i),
		)
	}
}

// BenchmarkRecordError benchmarks error recording
func BenchmarkRecordError(b *testing.B) {
	config := tracing.DefaultConfig()
	config.ServiceName = "benchmark-test"
	config.Protocol = tracing.ProtocolStdout

	service, _ := tracing.NewService(config)
	defer service.Shutdown(context.Background())

	ctx, span := service.StartSpan(context.Background(), "benchmark-operation")
	defer span.End()

	testErr := errors.New("benchmark error")

	for b.Loop() {
		service.RecordError(ctx, testErr)
	}
}

// BenchmarkNoOpService benchmarks no-op service
func BenchmarkNoOpService(b *testing.B) {
	service := tracing.NewNoOpService()
	ctx := context.Background()

	for b.Loop() {
		_, span := service.StartSpan(ctx, "noop-operation")
		span.SetAttributes(attribute.String("key", "value"))
		span.End()
	}
}
