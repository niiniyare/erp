// Package tracing provides distributed tracing capabilities using OpenTelemetry.
// It offers a simplified, enterprise-ready interface for instrumenting Go applications
// with support for OTLP exporters and common observability patterns.
package tracing

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Common errors.
var (
	ErrServiceClosed       = errors.New("tracing service is closed")
	ErrInvalidConfig       = errors.New("invalid configuration")
	ErrInvalidSamplingRate = errors.New("sampling rate must be between 0.0 and 1.0")
	ErrEmptyServiceName    = errors.New("service name cannot be empty")
	ErrUnsupportedProtocol = errors.New("unsupported protocol")
)

// Service provides distributed tracing functionality for applications.
// It wraps OpenTelemetry's tracing capabilities with a simplified interface
// suitable for enterprise use cases.
//
// This interface supports two usage patterns:
//
// Pattern 1 (Recommended): Service-level methods
//
//	ctx, span := tracingService.StartSpan(ctx, "operation")
//	defer span.End()
//	tracingService.SetAttributes(ctx, attr1, attr2)
//	tracingService.RecordError(ctx, err)
//
// Pattern 2 (Legacy/Direct): Span-level methods
//
//	ctx, span := tracer.StartSpan(ctx, "operation")
//	defer span.End()
//	span.SetAttributes(attr1, attr2)
//	span.RecordError(err)
type Service interface {
	// StartSpan creates a new span with the given name and options.
	// Returns a new context containing the span and the span itself.
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// SpanFromContext retrieves the current span from the context.
	// Returns a no-op span if no span exists in the context.
	SpanFromContext(ctx context.Context) Span

	// InjectHTTPHeaders injects tracing context into HTTP headers for distributed tracing.
	InjectHTTPHeaders(ctx context.Context, headers http.Header)

	// ExtractHTTPHeaders extracts tracing context from HTTP headers.
	// Returns a new context containing the extracted trace information.
	ExtractHTTPHeaders(ctx context.Context, headers http.Header) context.Context

	// SetAttributes sets attributes on the current span in the context.
	SetAttributes(ctx context.Context, attrs ...attribute.KeyValue)

	// RecordError records an error on the current span in the context.
	RecordError(ctx context.Context, err error, opts ...ErrorOption)

	// AddEvent adds a timed event to the current span in the context.
	AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)

	// GetTraceID returns the trace ID from the current span in the context.
	GetTraceID(ctx context.Context) string

	// GetSpanID returns the span ID from the current span in the context.
	GetSpanID(ctx context.Context) string

	// Shutdown gracefully shuts down the tracing service, flushing any pending spans.
	// Should be called before application exit.
	Shutdown(ctx context.Context) error
}

// Span represents a single operation within a trace.
// Spans can be used directly for fine-grained control over tracing.
type Span interface {
	// End completes the span. Should be called when the operation finishes.
	End(opts ...SpanEndOption)

	// SetAttributes sets attributes describing the operation.
	SetAttributes(attrs ...attribute.KeyValue)

	// SetStatus sets the status of the span.
	SetStatus(code codes.Code, description string)

	// RecordError records an error that occurred during the span.
	RecordError(err error, opts ...trace.EventOption)

	// AddEvent adds a timed event with optional attributes.
	AddEvent(name string, attrs ...attribute.KeyValue)

	// IsRecording returns true if the span is recording.
	IsRecording() bool

	// SpanContext returns the span context for linking spans.
	SpanContext() trace.SpanContext

	// SetName updates the span name.
	SetName(name string)
}

// SpanKind represents the relationship between the span and its parent.
type SpanKind int

const (
	// SpanKindInternal indicates an internal operation.
	SpanKindInternal SpanKind = iota
	// SpanKindServer indicates a server handling a request.
	SpanKindServer
	// SpanKindClient indicates a client making a request.
	SpanKindClient
	// SpanKindProducer indicates a message producer.
	SpanKindProducer
	// SpanKindConsumer indicates a message consumer.
	SpanKindConsumer
)

// Protocol specifies the transport protocol for trace data.
type Protocol string

const (
	// ProtocolGRPC uses gRPC for trace export.
	ProtocolGRPC Protocol = "grpc"
	// ProtocolHTTP uses HTTP/protobuf for trace export.
	ProtocolHTTP Protocol = "http"
	// ProtocolStdout writes traces to stdout (useful for development).
	ProtocolStdout Protocol = "stdout"
)

// ExporterType specifies the type of trace exporter.
type ExporterType string

const (
	// GRPCExporter uses gRPC for trace export.
	GRPCExporter ExporterType = "grpc"
	// HTTPExporter uses HTTP/protobuf for trace export.
	HTTPExporter ExporterType = "http"
	// StdoutExporter writes traces to stdout (useful for development).
	StdoutExporter ExporterType = "stdout"
)

// Config holds configuration for the tracing service.
type Config struct {
	// ServiceName identifies the service in traces (required).
	ServiceName string

	// ServiceVersion is the version of the service.
	ServiceVersion string

	// Environment specifies the deployment environment (e.g., production, staging).
	Environment string

	// Endpoint is the OTLP collector endpoint (e.g., "localhost:4317").
	Endpoint string

	// Protocol specifies the transport protocol (grpc, http, stdout).
	Protocol Protocol

	// ExporterType specifies the type of trace exporter.
	ExporterType ExporterType

	// Insecure disables TLS for the connection.
	Insecure bool

	// Headers are custom headers to include in trace exports.
	Headers map[string]string

	// SamplingRate determines the percentage of traces to sample (0.0 to 1.0).
	SamplingRate float64

	// SamplingRatio is an alias for SamplingRate for backward compatibility.
	SamplingRatio float64

	// Enabled controls whether tracing is active.
	// When false, all operations become no-ops.
	Enabled bool

	// BatchTimeout is the maximum time between batch exports.
	BatchTimeout time.Duration

	// MaxExportBatchSize is the maximum number of spans per batch.
	MaxExportBatchSize int

	// MaxQueueSize is the maximum queue size for pending spans.
	MaxQueueSize int
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.ServiceName == "" {
		return ErrEmptyServiceName
	}

	samplingRate := c.getSamplingRate()
	if samplingRate < 0.0 || samplingRate > 1.0 {
		return ErrInvalidSamplingRate
	}

	if c.Protocol != ProtocolGRPC && c.Protocol != ProtocolHTTP && c.Protocol != ProtocolStdout {
		return fmt.Errorf("%w: %s", ErrUnsupportedProtocol, c.Protocol)
	}

	if c.BatchTimeout <= 0 {
		return fmt.Errorf("%w: batch timeout must be positive", ErrInvalidConfig)
	}

	if c.MaxExportBatchSize <= 0 {
		return fmt.Errorf("%w: max export batch size must be positive", ErrInvalidConfig)
	}

	if c.MaxQueueSize <= 0 {
		return fmt.Errorf("%w: max queue size must be positive", ErrInvalidConfig)
	}

	return nil
}

// getSamplingRate returns the effective sampling rate, preferring SamplingRatio over SamplingRate for backward compatibility.
func (c *Config) getSamplingRate() float64 {
	if c.SamplingRatio > 0 {
		return c.SamplingRatio
	}
	return c.SamplingRate
}

// DefaultConfig returns a configuration with sensible defaults.
// Values can be overridden via environment variables:
//   - SERVICE_NAME
//   - SERVICE_VERSION
//   - ENVIRONMENT
//   - OTEL_ENDPOINT
//   - OTEL_PROTOCOL
//   - OTEL_INSECURE
//   - TRACING_ENABLED
func DefaultConfig() Config {
	return Config{
		ServiceName:        getEnv("SERVICE_NAME", "unknown-service"),
		ServiceVersion:     getEnv("SERVICE_VERSION", "0.0.0"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		Endpoint:           getEnv("OTEL_ENDPOINT", "localhost:4317"),
		Protocol:           Protocol(getEnv("OTEL_PROTOCOL", string(ProtocolGRPC))),
		Insecure:           getEnv("OTEL_INSECURE", "true") == "true",
		SamplingRate:       1.0,
		Enabled:            getEnv("TRACING_ENABLED", "true") == "true",
		BatchTimeout:       5 * time.Second,
		MaxExportBatchSize: 512,
		MaxQueueSize:       2048,
	}
}

// NewService creates a new tracing service with the provided configuration.
// Returns an error if the configuration is invalid or initialization fails.
func NewService(cfg Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if !cfg.Enabled {
		return &noopService{}, nil
	}

	res, err := createResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	exporter, err := createExporter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.getSamplingRate())),
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(cfg.BatchTimeout),
			sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatchSize),
			sdktrace.WithMaxQueueSize(cfg.MaxQueueSize),
		),
	)

	// Set global providers
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	tracer := provider.Tracer(
		cfg.ServiceName,
		trace.WithInstrumentationVersion(cfg.ServiceVersion),
	)

	return &service{
		tracer:   tracer,
		provider: provider,
	}, nil
}

// service is the concrete implementation of Service.
type service struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
	mu       sync.RWMutex
	closed   bool
}

// StartSpan implements Service.
func (s *service) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	if closed {
		return ctx, &noopSpan{}
	}

	options := defaultSpanOptions()
	for _, opt := range opts {
		opt.apply(&options)
	}

	startOpts := []trace.SpanStartOption{
		trace.WithSpanKind(otelSpanKind(options.kind)),
	}

	if len(options.attributes) > 0 {
		startOpts = append(startOpts, trace.WithAttributes(options.attributes...))
	}

	if len(options.links) > 0 {
		startOpts = append(startOpts, trace.WithLinks(options.links...))
	}

	ctx, span := s.tracer.Start(ctx, name, startOpts...)
	return ctx, &spanWrapper{span: span}
}

// SpanFromContext implements Service.
func (s *service) SpanFromContext(ctx context.Context) Span {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	if closed {
		return &noopSpan{}
	}

	return &spanWrapper{span: trace.SpanFromContext(ctx)}
}

// InjectHTTPHeaders implements Service.
func (s *service) InjectHTTPHeaders(ctx context.Context, headers http.Header) {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	if closed {
		return
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(headers))
}

// ExtractHTTPHeaders implements Service.
func (s *service) ExtractHTTPHeaders(ctx context.Context, headers http.Header) context.Context {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	if closed {
		return ctx
	}

	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(headers))
}

// SetAttributes implements Service.
func (s *service) SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	if closed {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// RecordError implements Service.
func (s *service) RecordError(ctx context.Context, err error, opts ...ErrorOption) {
	if err == nil {
		return
	}

	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	if closed {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	options := defaultErrorOptions()
	for _, opt := range opts {
		opt.apply(&options)
	}

	span.RecordError(err, trace.WithAttributes(options.attributes...))

	if options.setStatus {
		span.SetStatus(codes.Error, err.Error())
	}
}

// AddEvent implements Service.
func (s *service) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	if closed {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// GetTraceID implements Service.
func (s *service) GetTraceID(ctx context.Context) string {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	if closed {
		return ""
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID implements Service.
func (s *service) GetSpanID(ctx context.Context) string {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	if closed {
		return ""
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// Shutdown implements Service.
func (s *service) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrServiceClosed
	}

	s.closed = true

	if s.provider == nil {
		return nil
	}

	if err := s.provider.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown provider: %w", err)
	}

	return nil
}

// spanWrapper wraps an OpenTelemetry span.
type spanWrapper struct {
	span trace.Span
}

// End implements Span.
func (w *spanWrapper) End(opts ...SpanEndOption) {
	options := defaultSpanEndOptions()
	for _, opt := range opts {
		opt.apply(&options)
	}

	var endOpts []trace.SpanEndOption
	if !options.timestamp.IsZero() {
		endOpts = append(endOpts, trace.WithTimestamp(options.timestamp))
	}

	w.span.End(endOpts...)
}

// SetAttributes implements Span.
func (w *spanWrapper) SetAttributes(attrs ...attribute.KeyValue) {
	w.span.SetAttributes(attrs...)
}

// SetStatus implements Span.
func (w *spanWrapper) SetStatus(code codes.Code, description string) {
	w.span.SetStatus(code, description)
}

// RecordError implements Span.
func (w *spanWrapper) RecordError(err error, opts ...trace.EventOption) {
	w.span.RecordError(err, opts...)
}

// AddEvent implements Span.
func (w *spanWrapper) AddEvent(name string, attrs ...attribute.KeyValue) {
	w.span.AddEvent(name, trace.WithAttributes(attrs...))
}

// IsRecording implements Span.
func (w *spanWrapper) IsRecording() bool {
	return w.span.IsRecording()
}

// SpanContext implements Span.
func (w *spanWrapper) SpanContext() trace.SpanContext {
	return w.span.SpanContext()
}

// SetName implements Span.
func (w *spanWrapper) SetName(name string) {
	w.span.SetName(name)
}

// noopService is a no-op implementation used when tracing is disabled.
type noopService struct{}

func (n *noopService) StartSpan(ctx context.Context, _ string, _ ...SpanOption) (context.Context, Span) {
	return ctx, &noopSpan{}
}

func (n *noopService) SpanFromContext(_ context.Context) Span {
	return &noopSpan{}
}

func (n *noopService) InjectHTTPHeaders(_ context.Context, _ http.Header) {}

func (n *noopService) ExtractHTTPHeaders(ctx context.Context, _ http.Header) context.Context {
	return ctx
}

func (n *noopService) SetAttributes(_ context.Context, _ ...attribute.KeyValue) {}

func (n *noopService) RecordError(_ context.Context, _ error, _ ...ErrorOption) {}

func (n *noopService) AddEvent(_ context.Context, _ string, _ ...attribute.KeyValue) {}

func (n *noopService) GetTraceID(_ context.Context) string { return "" }

func (n *noopService) GetSpanID(_ context.Context) string { return "" }

func (n *noopService) Shutdown(_ context.Context) error { return nil }

// noopSpan is a no-op implementation of Span.
type noopSpan struct{}

func (n *noopSpan) End(_ ...SpanEndOption)                      {}
func (n *noopSpan) SetAttributes(_ ...attribute.KeyValue)       {}
func (n *noopSpan) SetStatus(_ codes.Code, _ string)            {}
func (n *noopSpan) RecordError(_ error, _ ...trace.EventOption) {}
func (n *noopSpan) AddEvent(_ string, _ ...attribute.KeyValue)  {}
func (n *noopSpan) IsRecording() bool                           { return false }
func (n *noopSpan) SpanContext() trace.SpanContext              { return trace.SpanContext{} }
func (n *noopSpan) SetName(_ string)                            {}

// SpanOption configures a span during creation.
type SpanOption interface {
	apply(*spanOptions)
}

type spanOptionFunc func(*spanOptions)

func (f spanOptionFunc) apply(opts *spanOptions) {
	f(opts)
}

type spanOptions struct {
	kind       SpanKind
	attributes []attribute.KeyValue
	links      []trace.Link
}

func defaultSpanOptions() spanOptions {
	return spanOptions{
		kind: SpanKindInternal,
	}
}

// WithSpanKind sets the kind of span being created.
func WithSpanKind(kind SpanKind) SpanOption {
	return spanOptionFunc(func(opts *spanOptions) {
		opts.kind = kind
	})
}

// WithAttributes adds attributes to the span at creation time.
func WithAttributes(attrs ...attribute.KeyValue) SpanOption {
	return spanOptionFunc(func(opts *spanOptions) {
		opts.attributes = append(opts.attributes, attrs...)
	})
}

// WithLinks adds links to other spans.
func WithLinks(links ...trace.Link) SpanOption {
	return spanOptionFunc(func(opts *spanOptions) {
		opts.links = append(opts.links, links...)
	})
}

// SpanEndOption configures span termination.
type SpanEndOption interface {
	apply(*spanEndOptions)
}

type spanEndOptionFunc func(*spanEndOptions)

func (f spanEndOptionFunc) apply(opts *spanEndOptions) {
	f(opts)
}

type spanEndOptions struct {
	timestamp time.Time
}

func defaultSpanEndOptions() spanEndOptions {
	return spanEndOptions{}
}

// WithTimestamp sets a custom end timestamp for the span.
func WithTimestamp(t time.Time) SpanEndOption {
	return spanEndOptionFunc(func(opts *spanEndOptions) {
		opts.timestamp = t
	})
}

// ErrorOption configures error recording.
type ErrorOption interface {
	apply(*errorOptions)
}

type errorOptionFunc func(*errorOptions)

func (f errorOptionFunc) apply(opts *errorOptions) {
	f(opts)
}

type errorOptions struct {
	attributes []attribute.KeyValue
	setStatus  bool
}

func defaultErrorOptions() errorOptions {
	return errorOptions{}
}

// WithErrorAttributes adds attributes to the error event.
func WithErrorAttributes(attrs ...attribute.KeyValue) ErrorOption {
	return errorOptionFunc(func(opts *errorOptions) {
		opts.attributes = append(opts.attributes, attrs...)
	})
}

// WithErrorStatus marks the span as failed when recording an error.
func WithErrorStatus() ErrorOption {
	return errorOptionFunc(func(opts *errorOptions) {
		opts.setStatus = true
	})
}

// createResource creates an OpenTelemetry resource with service information.
func createResource(cfg Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(cfg.ServiceName),
		semconv.ServiceVersionKey.String(cfg.ServiceVersion),
	}

	if cfg.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironmentKey.String(cfg.Environment))
	}

	return resource.New(
		context.Background(),
		resource.WithAttributes(attrs...),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
}

// createExporter creates an OpenTelemetry span exporter based on the protocol or exporter type.
func createExporter(cfg Config) (sdktrace.SpanExporter, error) {
	// Use ExporterType if provided, otherwise fall back to Protocol
	var exporterType string
	if cfg.ExporterType != "" {
		exporterType = string(cfg.ExporterType)
	} else {
		exporterType = string(cfg.Protocol)
	}

	switch exporterType {
	case "grpc":
		return createGRPCExporter(cfg)
	case "http":
		return createHTTPExporter(cfg)
	case "stdout":
		// Check if we should suppress output for integration tests
		if os.Getenv("INTEGRATION_TEST_QUIET") == "true" {
			return stdouttrace.New(stdouttrace.WithWriter(io.Discard))
		}
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProtocol, exporterType)
	}
}

func createGRPCExporter(cfg Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
	}

	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
	}

	client := otlptracegrpc.NewClient(opts...)
	return otlptrace.New(context.Background(), client)
}

func createHTTPExporter(cfg Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
	}

	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
	}

	client := otlptracehttp.NewClient(opts...)
	return otlptrace.New(context.Background(), client)
}

// otelSpanKind converts our SpanKind to OpenTelemetry's SpanKind.
func otelSpanKind(kind SpanKind) trace.SpanKind {
	switch kind {
	case SpanKindServer:
		return trace.SpanKindServer
	case SpanKindClient:
		return trace.SpanKindClient
	case SpanKindProducer:
		return trace.SpanKindProducer
	case SpanKindConsumer:
		return trace.SpanKindConsumer
	default:
		return trace.SpanKindInternal
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewNoOpService creates a tracing service with tracing disabled.
// Useful for testing or when tracing should be explicitly disabled.
func NewNoOpService() Service {
	return &noopService{}
}

// ============================================================================
// HELPER FUNCTIONS FOR COMMON SEMANTIC CONVENTIONS
// ============================================================================

// HTTPAttributes creates standard HTTP semantic convention attributes.
func HTTPAttributes(method, url, userAgent string, statusCode int) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.HTTPMethodKey.String(method),
		semconv.HTTPURLKey.String(url),
	}

	if userAgent != "" {
		attrs = append(attrs, semconv.HTTPUserAgentKey.String(userAgent))
	}

	if statusCode > 0 {
		attrs = append(attrs, semconv.HTTPStatusCodeKey.Int(statusCode))
	}

	return attrs
}

// DBAttributes creates standard database semantic convention attributes.
func DBAttributes(system, name, statement string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}

	if system != "" {
		attrs = append(attrs, semconv.DBSystemKey.String(system))
	}

	if name != "" {
		attrs = append(attrs, semconv.DBNameKey.String(name))
	}

	if statement != "" {
		attrs = append(attrs, semconv.DBStatementKey.String(statement))
	}

	return attrs
}

// RPCAttributes creates standard RPC semantic convention attributes.
func RPCAttributes(system, service, method string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}

	if system != "" {
		attrs = append(attrs, semconv.RPCSystemKey.String(system))
	}

	if service != "" {
		attrs = append(attrs, semconv.RPCServiceKey.String(service))
	}

	if method != "" {
		attrs = append(attrs, semconv.RPCMethodKey.String(method))
	}

	return attrs
}

// MessagingAttributes creates standard messaging semantic convention attributes.
func MessagingAttributes(system, destination, operation string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}

	if system != "" {
		attrs = append(attrs, semconv.MessagingSystemKey.String(system))
	}

	if destination != "" {
		attrs = append(attrs, semconv.MessagingDestinationNameKey.String(destination))
	}

	if operation != "" {
		attrs = append(attrs, semconv.MessagingOperationKey.String(operation))
	}

	return attrs
}
