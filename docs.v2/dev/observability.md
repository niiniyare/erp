# Observability Integration Guide

 observability through structured logging, distributed tracing, and metrics collection for better monitoring, debugging, and performance optimization.

## 📊 Observability Stack Overview

Our observability system includes three pillars integrated throughout all application layers:

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layers                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ API Layer   │  │ Service     │  │ Repository  │        │
│  │ 📝🔍📊     │  │ 📝🔍📊     │  │ 📝🔍📊     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Observability Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  📝 Logging │  │ 🔍 Tracing  │  │ 📊 Metrics  │        │
│  │   (Zap/     │  │ (OpenTel)   │  │ (Prometheus)│        │
│  │  Zerolog)   │  │             │  │             │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

**Legend**: 📝 Logging | 🔍 Tracing | 📊 Metrics

## 📝 Structured Logging

### Logger Architecture

We support multiple logging backends with a unified interface:

```go
// internal/shared/logger/logger.go
type Logger interface {
    Debug(msg string, fields ...Fields)
    Info(msg string, fields ...Fields)
    Warn(msg string, fields ...Fields)
    Error(msg string, fields ...Fields)
    Fatal(msg string, fields ...Fields)
    
    // Context-aware logging
    DebugContext(ctx context.Context, msg string, fields ...Fields)
    InfoContext(ctx context.Context, msg string, fields ...Fields)
    WarnContext(ctx context.Context, msg string, fields ...Fields)
    ErrorContext(ctx context.Context, msg string, fields ...Fields)
    
    // Logger chaining
    WithFields(fields Fields) Logger
    WithContext(ctx context.Context) Logger
}

type Fields map[string]interface{}
```

### Supported Logger Types

#### 1. **Zap Logger** (High Performance)
```go
// Configuration
zapConfig := zap.Config{
    Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
    Development: false,
    Sampling: &zap.SamplingConfig{
        Initial:    100,
        Thereafter: 100,
    },
    Encoding: "json",
    EncoderConfig: zapcore.EncoderConfig{
        TimeKey:        "timestamp",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "caller",
        MessageKey:     "message",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeLevel:    zapcore.LowercaseLevelEncoder,
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeDuration: zapcore.StringDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    },
}
```

#### 2. **Zerolog** (Zero Allocation)
```go
// Configuration
zerolog.TimeFieldFormat = time.RFC3339
log.Logger = zerolog.New(os.Stdout).With().
    Timestamp().
    Str("service", "erp-api").
    Str("version", "1.0.0").
    Logger()
```

#### 3. **Slog** (Go Standard - Default)
```go
// Configuration
opts := &slog.HandlerOptions{
    Level: slog.LevelInfo,
    AddSource: true,
}
handler := slog.NewJSONHandler(os.Stdout, opts)
logger := slog.New(handler)
```

### Logger Initialization

#### Environment-Based Setup
```go
// Initialize from environment variables
func InitializeFromEnv() error {
    logType := os.Getenv("LOG_TYPE")       // zap, zerolog, slog
    logLevel := os.Getenv("LOG_LEVEL")     // debug, info, warn, error
    logFormat := os.Getenv("LOG_FORMAT")   // json, text, console
    
    config := Config{
        Type:        ParseLoggerType(logType),
        Level:       ParseLogLevel(logLevel),
        Format:      logFormat,
        ServiceName: os.Getenv("SERVICE_NAME"),
        Version:     os.Getenv("SERVICE_VERSION"),
        Development: os.Getenv("LOG_DEVELOPMENT") == "true",
    }
    
    return Initialize(config)
}
```

#### Programmatic Setup
```go
// Custom configuration
config := logger.Config{
    Type:        logger.ZapLogger,
    Level:       logger.InfoLevel,
    Format:      "json",
    ServiceName: "erp-api",
    Version:     "1.0.0",
    Development: false,
}

if err := logger.Initialize(config); err != nil {
    log.Fatal("Failed to initialize logger:", err)
}
```

### Logging Patterns

#### Context-Aware Logging
```go
// Extract context information and log
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    // Extract trace and tenant information from context
    logger.InfoContext(ctx, "Starting tenant creation", logger.Fields{
        "tenant_name": req.Name,
        "tenant_slug": req.Slug,
        "operation":   "create_tenant",
    })
    
    // Business logic...
    tenant, err := s.processCreation(ctx, req)
    if err != nil {
        logger.ErrorContext(ctx, "Failed to create tenant", logger.Fields{
            "tenant_name": req.Name,
            "error":       err.Error(),
            "operation":   "create_tenant",
        })
        return nil, err
    }
    
    logger.InfoContext(ctx, "Tenant created successfully", logger.Fields{
        "tenant_id":   tenant.ID.String(),
        "tenant_name": tenant.Name,
        "operation":   "create_tenant",
    })
    
    return tenant, nil
}
```

#### Structured Field Logging
```go
// Use consistent field names across the application
var StandardFields = struct {
    TenantID    string
    UserID      string
    Operation   string
    Duration    string
    ErrorType   string
    RequestID   string
}{
    TenantID:  "tenant_id",
    UserID:    "user_id", 
    Operation: "operation",
    Duration:  "duration_ms",
    ErrorType: "error_type",
    RequestID: "request_id",
}

// Usage
logger.InfoContext(ctx, "Processing user request", logger.Fields{
    StandardFields.TenantID:  tenantID.String(),
    StandardFields.UserID:    userID.String(),
    StandardFields.Operation: "get_user_profile",
    StandardFields.RequestID: getRequestID(ctx),
})
```

## 🔍 Distributed Tracing

### Tracing Architecture

OpenTelemetry-based distributed tracing for request correlation and performance monitoring:

```go
// internal/shared/tracing/tracing.go
type TracingService struct {
    config   TracingConfig
    tracer   trace.Tracer
    provider *sdktrace.TracerProvider
}

type TracingConfig struct {
    ServiceName    string
    ServiceVersion string
    Environment    string
    ExporterType   ExporterType
    SamplingRatio  float64
    Enabled        bool
    ExporterConfig ExporterConfig
}

type ExporterType string

const (
    JaegerExporter     ExporterType = "jaeger"
    OTLPGRPCExporter  ExporterType = "otlp_grpc"
    OTLPHTTPExporter  ExporterType = "otlp_http"
    StdoutExporter    ExporterType = "stdout"
)
```

### Tracing Service Interface
```go
type Tracer interface {
    // Span management
    StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
    
    // HTTP header injection/extraction for distributed tracing
    InjectHTTPHeaders(ctx context.Context, headers http.Header)
    ExtractHTTPHeaders(ctx context.Context, headers http.Header) context.Context
    
    // Span utilities
    RecordError(ctx context.Context, err error, opts ...ErrorOption)
    SetAttribute(ctx context.Context, key string, value interface{})
    AddEvent(ctx context.Context, name string, opts ...EventOption)
    
    // Shutdown
    Shutdown(ctx context.Context) error
}

type Span interface {
    SetAttributes(attrs ...attribute.KeyValue)
    SetStatus(code codes.Code, description string)
    RecordError(err error, opts ...trace.EventOption)
    AddEvent(name string, opts ...trace.EventOption)
    End()
}
```

### Tracing Configuration

#### Jaeger Configuration
```go
tracingConfig := tracing.TracingConfig{
    ServiceName:    "erp-api",
    ServiceVersion: "1.0.0",
    Environment:    "production",
    ExporterType:   tracing.JaegerExporter,
    SamplingRatio:  0.1, // 10% sampling
    Enabled:        true,
    ExporterConfig: tracing.ExporterConfig{
        JaegerEndpoint: "http://jaeger:14268/api/traces",
    },
}
```

#### OTLP Configuration
```go
tracingConfig := tracing.TracingConfig{
    ServiceName:    "erp-api",
    ServiceVersion: "1.0.0",
    Environment:    "production",
    ExporterType:   tracing.OTLPGRPCExporter,
    SamplingRatio:  0.05, // 5% sampling for production
    Enabled:        true,
    ExporterConfig: tracing.ExporterConfig{
        OTLPEndpoint: "http://otel-collector:4317",
    },
}
```

### Tracing Integration Patterns

#### HTTP Request Tracing
```go
// API Layer - Extract and inject trace context
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    // Extract tracing context from HTTP headers
    ctx := h.tracing.ExtractHTTPHeaders(c.Request.Context(), c.Request.Header)
    
    // Start HTTP span
    ctx, span := h.tracing.StartSpan(ctx, "http.create_tenant", 
        tracing.WithSpanKind(tracing.SpanKindServer),
        tracing.WithAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
            attribute.String("http.route", "/api/v1/tenants"),
            attribute.String("http.user_agent", c.Request.UserAgent()),
        ))
    defer span.End()
    
    // Parse request
    var req CreateTenantRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // Record error in span
        h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
        span.SetAttributes(attribute.String("error.type", "validation_error"))
        
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Add request attributes to span
    span.SetAttributes(
        attribute.String("tenant.name", req.Name),
        attribute.String("tenant.slug", req.Slug),
    )
    
    // Call service layer
    tenant, err := h.service.CreateTenant(ctx, req)
    if err != nil {
        // Record error with proper classification
        h.tracing.RecordError(ctx, err)
        
        // Classify error type
        if errors.Is(err, ErrSubdomainAlreadyExists) {
            span.SetAttributes(attribute.String("error.type", "business_error"))
            c.JSON(http.StatusConflict, gin.H{"error": "Subdomain already exists"})
        } else {
            span.SetAttributes(attribute.String("error.type", "internal_error"))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        }
        
        span.SetStatus(codes.Error, err.Error())
        return
    }
    
    // Success attributes
    span.SetAttributes(
        attribute.String("tenant.id", tenant.ID.String()),
        attribute.String("response.status", "created"),
    )
    span.SetStatus(codes.Ok, "Tenant created successfully")
    
    c.JSON(http.StatusCreated, tenant)
}
```

#### Service Layer Tracing
```go
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    // Start service span
    ctx, span := s.tracing.StartSpan(ctx, "service.create_tenant",
        tracing.WithSpanKind(tracing.SpanKindInternal))
    defer span.End()
    
    // Add business context to span
    span.SetAttributes(
        attribute.String("tenant.name", req.Name),
        attribute.String("tenant.slug", req.Slug),
    )
    
    // Business validation with sub-spans
    if req.Subdomain != nil {
        ctx, validationSpan := s.tracing.StartSpan(ctx, "service.validate_subdomain")
        validationSpan.SetAttributes(attribute.String("subdomain", *req.Subdomain))
        
        exists, err := s.repo.Exists(ctx, *req.Subdomain)
        if err != nil {
            validationSpan.RecordError(err)
            validationSpan.SetStatus(codes.Error, "Validation failed")
            validationSpan.End()
            return nil, fmt.Errorf("failed to validate subdomain: %w", err)
        }
        
        validationSpan.SetAttributes(attribute.Bool("subdomain.exists", exists))
        validationSpan.End()
        
        if exists {
            span.SetAttributes(attribute.String("error.type", "subdomain_conflict"))
            span.SetStatus(codes.Error, "Subdomain already exists")
            return nil, ErrSubdomainAlreadyExists
        }
    }
    
    // Create tenant entity
    tenant := &Tenant{
        ID:   uuid.New(),
        Name: req.Name,
        Slug: req.Slug,
        // ... other fields
    }
    
    span.SetAttributes(attribute.String("tenant.id", tenant.ID.String()))
    
    // Repository operation
    if err := s.repo.Create(ctx, tenant); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to create tenant")
        return nil, fmt.Errorf("failed to create tenant: %w", err)
    }
    
    span.SetStatus(codes.Ok, "Tenant created successfully")
    return tenant, nil
}
```

#### Repository Layer Tracing
```go
func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
    // Start database span
    ctx, span := r.tracing.StartSpan(ctx, "repository.create_tenant",
        tracing.WithSpanKind(tracing.SpanKindClient),
        tracing.WithAttributes(
            attribute.String("db.system", "postgresql"),
            attribute.String("db.operation", "create"),
            attribute.String("db.table", "tenants"),
        ))
    defer span.End()
    
    // Add entity attributes
    span.SetAttributes(
        attribute.String("tenant.id", tenant.ID.String()),
        attribute.String("tenant.name", tenant.Name),
    )
    
    // Convert and execute
    params, err := ToSQLCCreateParams(tenant)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to convert tenant")
        return fmt.Errorf("failed to convert tenant: %w", err)
    }
    
    _, err = r.store.CreateTenant(ctx, params)
    if err != nil {
        // Classify database errors
        span.SetAttributes(attribute.String("db.error.type", classifyDBError(err)))
        span.RecordError(err)
        span.SetStatus(codes.Error, "Database operation failed")
        return fmt.Errorf("failed to create tenant: %w", err)
    }
    
    span.SetStatus(codes.Ok, "Tenant created in database")
    return nil
}
```

## 📊 Metrics Collection

### Metrics Architecture

Prometheus and OpenTelemetry metrics for performance and business monitoring:

```go
// internal/shared/metrics/metrics.go
type MetricsProvider interface {
    // Counter metrics
    Counter(name, help string, labelKeys ...string) Counter
    IncrementCounter(name string, labels Fields)
    
    // Gauge metrics
    Gauge(name, help string, labelKeys ...string) Gauge
    SetGauge(name string, value float64, labels Fields)
    
    // Histogram metrics
    Histogram(name, help string, buckets []float64, labelKeys ...string) Histogram
    ObserveHistogram(name string, value float64, labels Fields)
    
    // Timer utilities
    Timer(name string, labels Fields) Timer
}

type Counter interface {
    Inc(labels ...string)
    Add(value float64, labels ...string)
}

type Gauge interface {
    Set(value float64, labels ...string)
    Inc(labels ...string)
    Dec(labels ...string)
    Add(value float64, labels ...string)
}

type Histogram interface {
    Observe(value float64, labels ...string)
}

type Timer interface {
    Stop() time.Duration
}
```

### Metrics Configuration

#### Prometheus Configuration
```go
metricsConfig := metrics.MetricsConfig{
    Provider:  "prometheus",
    Namespace: "erp",
    Subsystem: "api",
    Enabled:   true,
    Labels: map[string]string{
        "service": "erp-api",
        "version": "1.0.0",
        "env":     "production",
    },
}

metricsService, err := metrics.NewMetricsService(metricsConfig)
```

#### OpenTelemetry Metrics Configuration
```go
metricsConfig := metrics.MetricsConfig{
    Provider: "opentelemetry",
    OTLP: metrics.OTLPConfig{
        Endpoint: "http://otel-collector:4317",
        Headers:  map[string]string{"api-key": "secret"},
    },
    Enabled: true,
}
```

### Core Metrics

#### HTTP Metrics
```go
// Request duration histogram
httpDuration := metrics.Histogram(
    "http_request_duration_seconds",
    "HTTP request duration in seconds",
    []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
    "method", "endpoint", "status",
)

// Request count counter
httpRequests := metrics.Counter(
    "http_requests_total",
    "Total number of HTTP requests",
    "method", "endpoint", "status",
)

// Active requests gauge
activeRequests := metrics.Gauge(
    "http_active_requests",
    "Number of active HTTP requests",
    "method", "endpoint",
)
```

#### Business Metrics
```go
// Tenant operations
tenantOps := metrics.Counter(
    "tenant_operations_total",
    "Total tenant operations",
    "operation", "status",
)

// Database operations
dbOps := metrics.Counter(
    "database_operations_total", 
    "Total database operations",
    "operation", "table", "status",
)

// Cache operations
cacheOps := metrics.Counter(
    "cache_operations_total",
    "Total cache operations", 
    "operation", "result",
)
```

### Metrics Integration Examples

#### HTTP Handler Metrics
```go
func (h *TenantHandler) CreateTenant(c *gin.Context) {
    // Start request timer
    timer := h.metrics.Timer("http_request_duration", metrics.Fields{
        "method":   c.Request.Method,
        "endpoint": "/api/v1/tenants",
    })
    defer timer.Stop()
    
    // Increment active requests
    h.metrics.SetGauge("http_active_requests", 1, metrics.Fields{
        "method":   c.Request.Method,
        "endpoint": "/api/v1/tenants",
    })
    defer h.metrics.SetGauge("http_active_requests", -1, metrics.Fields{
        "method":   c.Request.Method,
        "endpoint": "/api/v1/tenants",
    })
    
    // Process request
    tenant, err := h.service.CreateTenant(c.Request.Context(), req)
    if err != nil {
        // Error metrics
        h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
            "method":   c.Request.Method,
            "endpoint": "/api/v1/tenants",
            "status":   "error",
        })
        
        h.metrics.IncrementCounter("http_errors_total", metrics.Fields{
            "method":     c.Request.Method,
            "endpoint":   "/api/v1/tenants",
            "error_type": classifyError(err),
        })
        
        handleError(c, err)
        return
    }
    
    // Success metrics
    h.metrics.IncrementCounter("http_requests_total", metrics.Fields{
        "method":   c.Request.Method,
        "endpoint": "/api/v1/tenants",
        "status":   "success",
    })
    
    c.JSON(http.StatusCreated, tenant)
}
```

#### Service Layer Metrics
```go
func (s *service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
    // Business operation timer
    timer := s.metrics.Timer("tenant_operation_duration", metrics.Fields{
        "operation": "create",
    })
    defer timer.Stop()
    
    // Business logic...
    tenant, err := s.processTenantCreation(ctx, req)
    if err != nil {
        // Business error metrics
        s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
            "operation": "create",
            "status":    "error",
            "error_type": classifyBusinessError(err),
        })
        return nil, err
    }
    
    // Success metrics
    s.metrics.IncrementCounter("tenant_operations_total", metrics.Fields{
        "operation": "create",
        "status":    "success",
    })
    
    s.metrics.IncrementCounter("tenants_created_total", metrics.Fields{})
    
    return tenant, nil
}
```

#### Repository Layer Metrics
```go
func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
    // Database operation timer
    timer := r.metrics.Timer("database_operation_duration", metrics.Fields{
        "operation": "create",
        "table":     "tenants",
    })
    defer timer.Stop()
    
    // Execute operation
    err := r.store.CreateTenant(ctx, params)
    if err != nil {
        // Database error metrics
        r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
            "operation": "create",
            "table":     "tenants", 
            "status":    "error",
        })
        
        r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
            "operation":  "create",
            "table":      "tenants",
            "error_type": classifyDBError(err),
        })
        
        return fmt.Errorf("failed to create tenant: %w", err)
    }
    
    // Success metrics
    r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
        "operation": "create",
        "table":     "tenants",
        "status":    "success",
    })
    
    return nil
}
```

## 🔧 Environment Configuration

### Environment Variables
```bash
# Logging Configuration
LOG_TYPE=zap                    # zap, zerolog, slog
LOG_LEVEL=info                  # debug, info, warn, error, fatal
LOG_FORMAT=json                 # json, text, console
LOG_DEVELOPMENT=false
SERVICE_NAME=erp-api
SERVICE_VERSION=1.0.0

# Tracing Configuration
OTEL_SERVICE_NAME=erp-api
OTEL_SERVICE_VERSION=1.0.0
OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:14268
OTEL_RESOURCE_ATTRIBUTES=environment=production
TRACING_ENABLED=true
TRACING_SAMPLING_RATIO=0.1

# Metrics Configuration
METRICS_PROVIDER=prometheus
METRICS_ENABLED=true
METRICS_NAMESPACE=erp
METRICS_SUBSYSTEM=api
PROMETHEUS_PORT=8090
```

### Configuration Initialization
```go
func initializeObservability() error {
    // Initialize logger
    if err := logger.InitializeFromEnv(); err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    
    // Initialize tracing
    tracingConfig := tracing.TracingConfig{
        ServiceName:    os.Getenv("OTEL_SERVICE_NAME"),
        ServiceVersion: os.Getenv("OTEL_SERVICE_VERSION"),
        Environment:    os.Getenv("ENVIRONMENT"),
        ExporterType:   tracing.OTLPGRPCExporter,
        SamplingRatio:  0.1,
        Enabled:        os.Getenv("TRACING_ENABLED") == "true",
        ExporterConfig: tracing.ExporterConfig{
            OTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
        },
    }
    
    tracingService, err := tracing.NewTracingService(tracingConfig)
    if err != nil {
        return fmt.Errorf("failed to initialize tracing: %w", err)
    }
    
    // Initialize metrics
    metricsConfig := metrics.MetricsConfig{
        Provider:  os.Getenv("METRICS_PROVIDER"),
        Namespace: os.Getenv("METRICS_NAMESPACE"),
        Subsystem: os.Getenv("METRICS_SUBSYSTEM"),
        Enabled:   os.Getenv("METRICS_ENABLED") == "true",
    }
    
    metricsService, err := metrics.NewMetricsService(metricsConfig)
    if err != nil {
        return fmt.Errorf("failed to initialize metrics: %w", err)
    }
    
    return nil
}
```

## 🎯 Best Practices

### 1. **Consistent Field Names**
```go
// Use standardized field names across the application
const (
    FieldTenantID   = "tenant_id"
    FieldUserID     = "user_id" 
    FieldOperation  = "operation"
    FieldDuration   = "duration_ms"
    FieldErrorType  = "error_type"
    FieldRequestID  = "request_id"
)
```

### 2. **Context Propagation**
```go
// Always pass context through layers
func (h *Handler) Process(c *gin.Context) {
    ctx := c.Request.Context()
    // Pass context to all downstream calls
    result, err := h.service.Process(ctx, request)
}
```

### 3. **Error Classification**
```go
// Classify errors for better metrics and debugging
func classifyError(err error) string {
    switch {
    case errors.Is(err, ErrTenantNotFound):
        return "not_found"
    case errors.Is(err, ErrSubdomainAlreadyExists):
        return "conflict"
    case errors.Is(err, ErrValidation):
        return "validation"
    default:
        return "internal"
    }
}
```

### 4. **Sampling Strategy**
```go
// Use different sampling rates for different environments
func getSamplingRatio() float64 {
    switch os.Getenv("ENVIRONMENT") {
    case "development":
        return 1.0  // 100% sampling
    case "staging":
        return 0.5  // 50% sampling
    case "production":
        return 0.1  // 10% sampling
    default:
        return 0.01 // 1% sampling
    }
}
```

### 5. **Performance Considerations**
```go
// Use appropriate logging levels
if logger.IsDebugEnabled() {
    logger.Debug("Expensive debug operation", expensiveFields())
}

// Lazy field evaluation
logger.Info("Operation completed", logger.Fields{
    "result": func() interface{} {
        return expensiveComputation()
    }(),
})
```

---

📚 **Next Steps**:
- [Code Examples](./code-examples.md) - See complete observability integration
- [Best Practices](./best-practices.md) - Development guidelines with observability
- [Error Handling](./error-handling.md) - Error tracking and monitoring
