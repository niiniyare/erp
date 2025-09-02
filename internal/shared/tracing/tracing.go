package tracing

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
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

// TracingConfig holds configuration for the tracing service
type TracingConfig struct {
	// Service information
	ServiceName    string
	ServiceVersion string
	Environment    string

	// Exporter configuration
	ExporterType   ExporterType
	ExporterConfig ExporterConfig

	// Sampling configuration
	SamplingRatio float64

	// Resource attributes
	ResourceAttributes map[string]string

	// Propagation configuration
	Propagators []string

	// Batch processor configuration
	BatchTimeout   time.Duration
	BatchSize      int
	MaxExportBatch int
	MaxQueueSize   int

	// Enable/disable tracing
	Enabled bool
}

// ExporterType defines the type of trace exporter
type ExporterType string

const (
	JaegerExporter   ExporterType = "jaeger"
	OTLPGRPCExporter ExporterType = "otlp-grpc"
	OTLPHTTPExporter ExporterType = "otlp-http"
	StdoutExporter   ExporterType = "stdout"
	NoopExporter     ExporterType = "noop"
)

// ExporterConfig holds exporter-specific configuration
type ExporterConfig struct {
	// Jaeger configuration
	JaegerEndpoint string
	JaegerUser     string
	JaegerPassword string

	// OTLP configuration
	OTLPEndpoint string
	OTLPHeaders  map[string]string
	OTLPInsecure bool

	// Stdout configuration
	StdoutPrettyPrint bool
}

// SpanKind represents the kind of span
type SpanKind trace.SpanKind

const (
	SpanKindInternal SpanKind = SpanKind(trace.SpanKindInternal)
	SpanKindServer   SpanKind = SpanKind(trace.SpanKindServer)
	SpanKindClient   SpanKind = SpanKind(trace.SpanKindClient)
	SpanKindProducer SpanKind = SpanKind(trace.SpanKindProducer)
	SpanKindConsumer SpanKind = SpanKind(trace.SpanKindConsumer)
)

// TracingService provides distributed tracing functionality
type TracingService interface {
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
	SpanFromContext(ctx context.Context) Span
	InjectHTTPHeaders(ctx context.Context, headers http.Header)
	ExtractHTTPHeaders(ctx context.Context, headers http.Header) context.Context
	RecordError(ctx context.Context, err error, opts ...ErrorOption)
	AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)
	SetAttributes(ctx context.Context, attrs ...attribute.KeyValue)
	GetTraceID(ctx context.Context) string
	GetSpanID(ctx context.Context) string
	Shutdown(ctx context.Context) error
}

// tracingService implements TracingService
type tracingService struct {
	config   TracingConfig
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
	mu       sync.RWMutex
	closed   bool
}

// NewTracingService creates a new tracing service instance
func NewTracingService(config TracingConfig) (TracingService, error) {
	if !config.Enabled {
		return &tracingService{
			config: config,
			tracer: otel.Tracer(config.ServiceName),
		}, nil
	}

	// Create resource with service information
	res, err := createResource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace exporter
	exporter, err := createExporter(config.ExporterType, config.ExporterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create trace provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SamplingRatio)),
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(config.BatchTimeout),
			sdktrace.WithMaxExportBatchSize(config.BatchSize),
			sdktrace.WithMaxQueueSize(config.MaxQueueSize),
		),
	)

	// Set global trace provider
	otel.SetTracerProvider(provider)

	// Configure propagators
	otel.SetTextMapPropagator(createPropagator(config.Propagators))

	// Create tracer
	tracer := provider.Tracer(
		config.ServiceName,
		trace.WithInstrumentationVersion(config.ServiceVersion),
	)

	return &tracingService{
		config:   config,
		tracer:   tracer,
		provider: provider,
	}, nil
}

// StartSpan starts a new span
func (ts *tracingService) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	if !ts.config.Enabled {
		return ctx, &noopSpan{}
	}

	spanOpts := &SpanOptions{}
	for _, opt := range opts {
		opt.apply(spanOpts)
	}

	// Convert to OpenTelemetry options
	otelOpts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKind(spanOpts.Kind)),
	}

	if len(spanOpts.Attributes) > 0 {
		otelOpts = append(otelOpts, trace.WithAttributes(spanOpts.Attributes...))
	}

	if len(spanOpts.Links) > 0 {
		otelOpts = append(otelOpts, trace.WithLinks(spanOpts.Links...))
	}

	otelCtx, otelSpan := ts.tracer.Start(ctx, name, otelOpts...)

	return otelCtx, &otelSpanWrapper{span: otelSpan}
}

// SpanFromContext returns the span from the context
func (ts *tracingService) SpanFromContext(ctx context.Context) Span {
	if !ts.config.Enabled {
		return &noopSpan{}
	}

	span := trace.SpanFromContext(ctx)
	return &otelSpanWrapper{span: span}
}

// InjectHTTPHeaders injects tracing headers into HTTP headers
func (ts *tracingService) InjectHTTPHeaders(ctx context.Context, headers http.Header) {
	if !ts.config.Enabled {
		return
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(headers))
}

// ExtractHTTPHeaders extracts tracing context from HTTP headers
func (ts *tracingService) ExtractHTTPHeaders(ctx context.Context, headers http.Header) context.Context {
	if !ts.config.Enabled {
		return ctx
	}

	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(headers))
}

// RecordError records an error in the current span
func (ts *tracingService) RecordError(ctx context.Context, err error, opts ...ErrorOption) {
	if !ts.config.Enabled || err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	errorOpts := &ErrorOptions{}
	for _, opt := range opts {
		opt.apply(errorOpts)
	}

	// Record the error
	span.RecordError(err, trace.WithAttributes(errorOpts.Attributes...))

	// Set status if requested
	if errorOpts.SetStatus {
		span.SetStatus(codes.Error, err.Error())
	}
}

// AddEvent adds an event to the current span
func (ts *tracingService) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if !ts.config.Enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetAttributes sets attributes on the current span
func (ts *tracingService) SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	if !ts.config.Enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// GetTraceID returns the trace ID from the context
func (ts *tracingService) GetTraceID(ctx context.Context) string {
	if !ts.config.Enabled {
		return ""
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from the context
func (ts *tracingService) GetSpanID(ctx context.Context) string {
	if !ts.config.Enabled {
		return ""
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// Shutdown gracefully shuts down the tracing service
func (ts *tracingService) Shutdown(ctx context.Context) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.closed || ts.provider == nil {
		return nil
	}

	ts.closed = true
	return ts.provider.Shutdown(ctx)
}

// Span interface wraps OpenTelemetry span functionality
type Span interface {
	End(opts ...SpanEndOption)
	AddEvent(name string, attrs ...attribute.KeyValue)
	SetAttributes(attrs ...attribute.KeyValue)
	SetStatus(code codes.Code, description string)
	SetName(name string)
	RecordError(err error, opts ...trace.EventOption)
	IsRecording() bool
	SpanContext() trace.SpanContext
}

// SpanOptions holds options for creating spans
type SpanOptions struct {
	Kind       SpanKind
	Attributes []attribute.KeyValue
	Links      []trace.Link
}

// SpanOption is a function that configures SpanOptions
type SpanOption interface {
	apply(*SpanOptions)
}

type spanOptionFunc func(*SpanOptions)

func (f spanOptionFunc) apply(opts *SpanOptions) {
	f(opts)
}

// WithSpanKind sets the span kind
func WithSpanKind(kind SpanKind) SpanOption {
	return spanOptionFunc(func(opts *SpanOptions) {
		opts.Kind = kind
	})
}

// WithAttributes sets span attributes
func WithAttributes(attrs ...attribute.KeyValue) SpanOption {
	return spanOptionFunc(func(opts *SpanOptions) {
		opts.Attributes = append(opts.Attributes, attrs...)
	})
}

// WithLinks adds span links
func WithLinks(links ...trace.Link) SpanOption {
	return spanOptionFunc(func(opts *SpanOptions) {
		opts.Links = append(opts.Links, links...)
	})
}

// SpanEndOptions holds options for ending spans
type SpanEndOptions struct {
	Timestamp time.Time
}

// SpanEndOption is a function that configures SpanEndOptions
type SpanEndOption interface {
	apply(*SpanEndOptions)
}

type spanEndOptionFunc func(*SpanEndOptions)

func (f spanEndOptionFunc) apply(opts *SpanEndOptions) {
	f(opts)
}

// WithTimestamp sets the end timestamp
func WithTimestamp(t time.Time) SpanEndOption {
	return spanEndOptionFunc(func(opts *SpanEndOptions) {
		opts.Timestamp = t
	})
}

// ErrorOptions holds options for recording errors
type ErrorOptions struct {
	Attributes []attribute.KeyValue
	SetStatus  bool
}

// ErrorOption is a function that configures ErrorOptions
type ErrorOption interface {
	apply(*ErrorOptions)
}

type errorOptionFunc func(*ErrorOptions)

func (f errorOptionFunc) apply(opts *ErrorOptions) {
	f(opts)
}

// WithErrorAttributes adds attributes to error recording
func WithErrorAttributes(attrs ...attribute.KeyValue) ErrorOption {
	return errorOptionFunc(func(opts *ErrorOptions) {
		opts.Attributes = append(opts.Attributes, attrs...)
	})
}

// WithErrorStatus sets the span status on error
func WithErrorStatus() ErrorOption {
	return errorOptionFunc(func(opts *ErrorOptions) {
		opts.SetStatus = true
	})
}

// otelSpanWrapper wraps OpenTelemetry span
type otelSpanWrapper struct {
	span trace.Span
}

func (w *otelSpanWrapper) End(opts ...SpanEndOption) {
	endOpts := &SpanEndOptions{}
	for _, opt := range opts {
		opt.apply(endOpts)
	}

	var otelOpts []trace.SpanEndOption
	if !endOpts.Timestamp.IsZero() {
		otelOpts = append(otelOpts, trace.WithTimestamp(endOpts.Timestamp))
	}

	w.span.End(otelOpts...)
}

func (w *otelSpanWrapper) AddEvent(name string, attrs ...attribute.KeyValue) {
	w.span.AddEvent(name, trace.WithAttributes(attrs...))
}

func (w *otelSpanWrapper) SetAttributes(attrs ...attribute.KeyValue) {
	w.span.SetAttributes(attrs...)
}

func (w *otelSpanWrapper) SetStatus(code codes.Code, description string) {
	w.span.SetStatus(code, description)
}

func (w *otelSpanWrapper) SetName(name string) {
	w.span.SetName(name)
}

func (w *otelSpanWrapper) RecordError(err error, opts ...trace.EventOption) {
	w.span.RecordError(err, opts...)
}

func (w *otelSpanWrapper) IsRecording() bool {
	return w.span.IsRecording()
}

func (w *otelSpanWrapper) SpanContext() trace.SpanContext {
	return w.span.SpanContext()
}

// noopSpan is a no-op implementation of Span
type noopSpan struct{}

func (n *noopSpan) End(opts ...SpanEndOption)                         {}
func (n *noopSpan) AddEvent(name string, attrs ...attribute.KeyValue) {}
func (n *noopSpan) SetAttributes(attrs ...attribute.KeyValue)         {}
func (n *noopSpan) SetStatus(code codes.Code, description string)     {}
func (n *noopSpan) SetName(name string)                               {}
func (n *noopSpan) RecordError(err error, opts ...trace.EventOption)  {}
func (n *noopSpan) IsRecording() bool                                 { return false }
func (n *noopSpan) SpanContext() trace.SpanContext                    { return trace.SpanContext{} }

// Helper functions

func createResource(config TracingConfig) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(config.ServiceName),
		semconv.ServiceVersionKey.String(config.ServiceVersion),
	}

	if config.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironmentKey.String(config.Environment))
	}

	// Add custom resource attributes
	for k, v := range config.ResourceAttributes {
		attrs = append(attrs, attribute.String(k, v))
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

func createExporter(exporterType ExporterType, config ExporterConfig) (sdktrace.SpanExporter, error) {
	switch exporterType {
	case JaegerExporter:
		return jaeger.New(jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(config.JaegerEndpoint),
			jaeger.WithUsername(config.JaegerUser),
			jaeger.WithPassword(config.JaegerPassword),
		))

	case OTLPGRPCExporter:
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
		}
		if config.OTLPInsecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		if len(config.OTLPHeaders) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(config.OTLPHeaders))
		}
		return otlptrace.New(context.Background(), otlptracegrpc.NewClient(opts...))

	case OTLPHTTPExporter:
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(config.OTLPEndpoint),
		}
		if config.OTLPInsecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		if len(config.OTLPHeaders) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(config.OTLPHeaders))
		}
		return otlptrace.New(context.Background(), otlptracehttp.NewClient(opts...))

	case StdoutExporter:
		opts := []stdouttrace.Option{}
		if config.StdoutPrettyPrint {
			opts = append(opts, stdouttrace.WithPrettyPrint())
		}
		return stdouttrace.New(opts...)

	case NoopExporter:
		return &noopExporter{}, nil

	default:
		return nil, fmt.Errorf("unsupported exporter type: %s", exporterType)
	}
}

func createPropagator(propagators []string) propagation.TextMapPropagator {
	var props []propagation.TextMapPropagator

	for _, p := range propagators {
		switch p {
		case "tracecontext":
			props = append(props, propagation.TraceContext{})
		case "baggage":
			props = append(props, propagation.Baggage{})
		case "b3":
			// Note: B3 propagator would need to be imported separately
			// props = append(props, b3.New())
		}
	}

	if len(props) == 0 {
		// Default propagators
		props = []propagation.TextMapPropagator{
			propagation.TraceContext{},
			propagation.Baggage{},
		}
	}

	return propagation.NewCompositeTextMapPropagator(props...)
}

// noopExporter is a no-op trace exporter
type noopExporter struct{}

func (n *noopExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (n *noopExporter) Shutdown(ctx context.Context) error {
	return nil
}

// DefaultConfig returns a default tracing configuration
func DefaultConfig() TracingConfig {
	return TracingConfig{
		ServiceName:    getEnvOrDefault("SERVICE_NAME", "unknown-service"),
		ServiceVersion: getEnvOrDefault("SERVICE_VERSION", "unknown"),
		Environment:    getEnvOrDefault("ENVIRONMENT", "development"),
		ExporterType:   StdoutExporter,
		SamplingRatio:  1.0,
		Propagators:    []string{"tracecontext", "baggage"},
		BatchTimeout:   5 * time.Second,
		BatchSize:      512,
		MaxExportBatch: 512,
		MaxQueueSize:   2048,
		Enabled:        true,
		ExporterConfig: ExporterConfig{
			StdoutPrettyPrint: true,
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Common attribute helpers
func HTTPAttributes(method, url, userAgent string, statusCode int) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.HTTPMethodKey.String(method),
		semconv.HTTPURLKey.String(url),
		semconv.HTTPUserAgentKey.String(userAgent),
		semconv.HTTPStatusCodeKey.Int(statusCode),
	}
}

func DBAttributes(system, name, statement string) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.DBSystemKey.String(system),
		semconv.DBNameKey.String(name),
		semconv.DBStatementKey.String(statement),
	}
}

func RPCAttributes(system, service, method string) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.RPCSystemKey.String(system),
		semconv.RPCServiceKey.String(service),
		semconv.RPCMethodKey.String(method),
	}
}

func MessagingAttributes(system, destination, operation string) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.MessagingSystemKey.String(system),
		semconv.MessagingDestinationNameKey.String(destination),
		semconv.MessagingOperationKey.String(operation),
	}
}

// NewNoOpTracingService creates a tracing service with tracing disabled for testing
func NewNoOpTracingService() TracingService {
	return &tracingService{
		config: TracingConfig{
			Enabled: false,
		},
		tracer: otel.Tracer("noop"),
	}
}
