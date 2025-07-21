package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
)

// MockTracingService is a mock implementation of the TracingService interface.
type MockTracingService struct{}

// NewMockTracingService creates and returns a new instance of MockTracingService.
func NewMockTracingService() *MockTracingService {
	return &MockTracingService{}
}

// StartSpan implements TracingService.StartSpan.
func (m *MockTracingService) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	// In a mock, we can just return the original context and a no-op span.
	return ctx, &noopSpan{}
}

// SpanFromContext implements TracingService.SpanFromContext.
func (m *MockTracingService) SpanFromContext(ctx context.Context) Span {
	return &noopSpan{}
}

// InjectHTTPHeaders implements TracingService.InjectHTTPHeaders.
func (m *MockTracingService) InjectHTTPHeaders(ctx context.Context, headers http.Header) {
	// No-op for mock
}

// ExtractHTTPHeaders implements TracingService.ExtractHTTPHeaders.
func (m *MockTracingService) ExtractHTTPHeaders(ctx context.Context, headers http.Header) context.Context {
	return ctx
}

// RecordError implements TracingService.RecordError.
func (m *MockTracingService) RecordError(ctx context.Context, err error, opts ...ErrorOption) {
	// No-op for mock
}

// AddEvent implements TracingService.AddEvent.
func (m *MockTracingService) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	// No-op for mock
}

// SetAttributes implements TracingService.SetAttributes.
func (m *MockTracingService) SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	// No-op for mock
}

// GetTraceID implements TracingService.GetTraceID.
func (m *MockTracingService) GetTraceID(ctx context.Context) string {
	return "mock-trace-id"
}

// GetSpanID implements TracingService.GetSpanID.
func (m *MockTracingService) GetSpanID(ctx context.Context) string {
	return "mock-span-id"
}

// Shutdown implements TracingService.Shutdown.
func (m *MockTracingService) Shutdown(ctx context.Context) error {
	return nil
}
