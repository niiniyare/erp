# Infrastructure Guide

A practical and conceptual reference for the shared infrastructure packages. This guide explains not just *how* to use each package, but *why* it was designed the way it was and *when* you should reach for each tool.

---

## Table of Contents

1. [The Big Picture](#0-the-big-picture)
2. [Logger](#1-logger)
3. [Tracing](#2-tracing)
4. [Metrics](#3-metrics)
5. [Errors](#4-errors)
6. [Cache](#5-cache)
7. [Putting It All Together](#6-putting-it-all-together)
8. [Testing with Mocks](#7-testing-with-mocks)

---

## 0. The Big Picture

### Why Do These Packages Exist?

When you build a multi-tenant ERP system that will eventually run at scale, three problems emerge very quickly:

1. **You can't see what's happening inside a running system.** Without structured logs, distributed traces, and metrics, debugging a production issue means reading guesswork. These packages solve the *observability* problem.

2. **Errors carry no meaning across layers.** A raw `pgx` error at the database layer means nothing to an HTTP handler. Without a common error vocabulary, every developer invents their own error shapes and the API becomes unpredictable. The errors package solves the *communication* problem.

3. **Every database call is slower than it needs to be.** Without caching, the same data gets fetched from Postgres repeatedly for every request. The cache package solves the *performance* problem, and does it in a multi-tenant-safe way.

These four packages — logger, tracing, metrics, errors — together form the **observability triad plus error contract**. The cache is the **performance layer**. They are intentionally designed to work together.

### The Three Pillars of Observability

Observability is the ability to understand what a system is doing from the outside, without modifying it. The three pillars are:

```
┌─────────────────────────────────────────────────────────────────┐
│                        OBSERVABILITY                            │
│                                                                 │
│  LOGS          TRACES            METRICS                        │
│  ─────         ──────            ───────                        │
│  What          Why & Where       How much / How fast            │
│  happened      it happened       it's happening                 │
│                                                                 │
│  Discrete      Causal chain      Aggregated numbers             │
│  events        across services   over time                      │
│                                                                 │
│  "User login   "This request     "99th percentile               │
│  failed at     touched 4         login latency is               │
│  14:32:01"     services and      340ms this hour"               │
│                took 450ms"                                      │
└─────────────────────────────────────────────────────────────────┘
```

Each pillar answers a different question. You need all three to fully understand a production incident:
- Metrics tell you *something is wrong* (latency spike)
- Traces tell you *where it went wrong* (slow DB query in tenant service)
- Logs tell you *what exactly happened* (the SQL query, the error message, the tenant ID)

### How the Packages Relate to Each Other

```
HTTP Request
     │
     ▼
┌─────────────────────────────────────────┐
│  Handler                                │
│  • ExtractHTTPHeaders (tracing)         │
│  • WithTenantID (cache ctx)             │
│  • errors.ToHTTPError on failure        │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Service                                │
│  • StartSpan (tracing)                  │
│  • Timer (metrics)                      │
│  • cache.Get → cache miss → repo        │
│  • returns *BusinessError on failure    │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Repository                             │
│  • Runs SQL                             │
│  • converts pgx errors → *RepoError     │
│  • converts sentinel → *BusinessError   │
└─────────────────────────────────────────┘
```

Errors flow *up*, observability signals flow *out* (to Jaeger, Prometheus, log aggregator). The `context.Context` is the thread that carries both the trace span and the tenant identity across all layers.

### Interface-First Design

Every package exposes an **interface**, not a concrete type:

```go
// You depend on this interface, not on a specific implementation
var log  logger.Logger
var tr   tracing.Service
var m    metrics.MetricsProvider
var c    cache.Service
```

This is intentional. It means:
- You can swap backends (zerolog → zap, Prometheus → OTel) without touching any business code
- Unit tests inject mocks instead of real infrastructure
- Services remain decoupled from infrastructure concerns
- The `noopService` / disabled variants let you run code without any infra at all (useful in tests and local dev)

---

## 1. Logger

**Package:** `awo.so/internal/shared/logger`

### Why This Design?

Logging has two consumers with very different needs:

- **Developers** in local dev want readable, coloured console output
- **Operations** in production want structured JSON that a log aggregator (Loki, Datadog, CloudWatch) can parse and query

The logger abstracts over three backends — **zerolog** (default, fastest), **zap** (common in Go ecosystem), and **slog** (stdlib, no extra dependency) — so you can pick the right one for your deployment without changing any calling code.

The interface uses `Fields map[string]any` instead of positional arguments. This forces structured logging from day one. Unstructured logs (`fmt.Sprintf("user %s not found", id)`) are a dead end — they are impossible to query in aggregators. Structured fields are queryable: `filter by user_id = "abc123"`.

### When to Use Each Method

| Method | Use when |
|--------|----------|
| `Debug` | Detailed internal state: SQL queries, cache keys, computed values. Disabled in production. |
| `Info` | Things that happened and are worth knowing: entity created, job completed, config loaded. |
| `Warn` | Something unexpected but recoverable: rate limit approaching, retry attempt, missing optional config. |
| `Error` | Something failed that shouldn't have: DB query failed, external API down, unexpected nil. |
| `Fatal` | Unrecoverable startup failure. Calls `os.Exit(1)`. Never use inside request handlers. |

`DebugContext`, `InfoContext`, etc. are the same but accept a `context.Context`. Prefer these inside service/handler methods because some backends can extract trace IDs from the context automatically.

### Global vs Injected Logger

The package exposes both a **global logger** (via `logger.Initialize`) and an **injectable interface** (via `LoggerFactory`). Use the global for main/startup code. Inject the interface into services so they are testable:

```go
// main.go — use global
logger.Initialize(logger.Config{...})

// service — inject interface
type TenantService struct {
    log logger.Logger // injected, mockable
}
```

### Initialization

```go
// Option A: explicit config
err := logger.Initialize(logger.Config{
    Type:        logger.ZerologLogger, // "zap" | "zerolog" | "slog"
    Level:       logger.InfoLevel,
    Output:      os.Stdout,
    Format:      "console",            // "json" in prod, "console" in dev
    Development: true,
    ServiceName: "my-service",
    Version:     "1.0.0",
})

// Option B: read from environment variables
// LOG_TYPE, LOG_LEVEL, LOG_FORMAT, SERVICE_NAME, SERVICE_VERSION
err := logger.InitializeFromEnv()

// Option C: inject as a dependency (for testability)
factory := &logger.LoggerFactory{}
log, err := factory.NewLogger(logger.DefaultConfig())
```

### Basic Logging

```go
logger.Debug("cache miss", logger.Fields{"key": "user:123"})
logger.Info("user created", logger.Fields{"user_id": userID, "email": email})
logger.Warn("rate limit approaching", logger.Fields{"ip": clientIP, "count": 95})
logger.Error("db query failed", logger.Fields{"error": err.Error(), "table": "users"})
logger.Fatal("startup failed", logger.Fields{"reason": err.Error()})
```

### Context-Aware Logging

```go
// Preferred inside service/handler methods
logger.InfoContext(ctx, "request processed", logger.Fields{"path": r.URL.Path, "ms": elapsed})
logger.ErrorContext(ctx, "handler failed", logger.Fields{"error": err.Error()})
```

### Scoped Logger (Recommended Pattern)

Rather than repeating `module`, `tenant_id` etc. on every call, create a child logger with those fields baked in. This is the most common pattern inside services:

```go
log := logger.WithFields(logger.Fields{
    "module":    "tenant",
    "tenant_id": tenantID,
})

log.Info("tenant provisioned")
log.Error("provisioning failed", logger.Fields{"error": err.Error()})
```

### Cleanup

```go
defer logger.Close() // flush buffered entries — critical for zap
```

### Log Levels (lowest to highest)

`Debug` → `Info` → `Warn` → `Error` → `Fatal`

Only messages at or above the configured level are emitted. In production, run at `Info`. Enable `Debug` temporarily when investigating an issue.

---

## 2. Tracing

**Package:** `awo.so/internal/shared/tracing`

### Why This Design?

Logs are great for understanding what happened on a single machine. But in a system where a single user request touches multiple services, logs alone don't help you understand *the full journey*. You need distributed tracing.

A **trace** is a tree of **spans**, where each span represents one operation (HTTP handler, service call, DB query). Spans carry a **trace ID** that is the same across all services involved in a single request. This lets you reconstruct the full execution path after the fact in tools like Jaeger or Grafana Tempo.

The package wraps **OpenTelemetry**, the industry standard. OTel is vendor-neutral — you can send traces to Jaeger, Zipkin, Datadog, or any OTLP-compatible backend without changing your code.

The key design insight: **the `context.Context` is the carrier**. You pass `ctx` through every function call, and the current span travels with it. Child spans automatically know their parent. When you call `StartSpan(ctx, "name")`, you get back a new `ctx` with the child span inside it — always use that new ctx for the rest of the operation.

### When to Create a Span

Create a span at every meaningful boundary:

| Boundary | Example |
|----------|---------|
| Incoming HTTP request | Every handler |
| Service method | `TenantService.Create`, `UserService.Authenticate` |
| Outgoing HTTP call | Calling an external API |
| Database query | Each repo method |
| Message queue operation | Publishing or consuming an event |

Do not create spans for trivial helper functions — it creates noise with no signal.

### The Span Lifecycle

```
StartSpan → ctx with span attached
    │
    ├── SetAttributes  (describe what this op is doing)
    ├── AddEvent       (timestamped breadcrumbs within the op)
    ├── RecordError    (if something goes wrong)
    └── span.End()     (always — use defer)
```

If you forget `span.End()`, the span is never exported. Use `defer span.End()` immediately after `StartSpan`.

### Initialization

```go
cfg := tracing.Config{
    ServiceName:        "my-service",
    ServiceVersion:     "1.0.0",
    Environment:        "production",
    Endpoint:           "localhost:4317",      // OTLP collector address
    Protocol:           tracing.ProtocolGRPC,  // "grpc" | "http" | "stdout"
    Insecure:           true,                  // set false in production
    SamplingRate:       1.0,                   // 1.0 = trace everything
    Enabled:            true,
    BatchTimeout:       5 * time.Second,
    MaxExportBatchSize: 512,
    MaxQueueSize:       2048,
}

// Or use defaults — reads from env: SERVICE_NAME, OTEL_ENDPOINT, OTEL_PROTOCOL, etc.
cfg = tracing.DefaultConfig()

svc, err := tracing.NewService(cfg)
defer svc.Shutdown(context.Background()) // flush pending spans on exit

// Disabled — all methods are no-ops, safe to use in tests
svc = tracing.NewNoOpService()
```

### Creating Spans

```go
ctx, span := svc.StartSpan(ctx, "TenantService.Create",
    tracing.WithSpanKind(tracing.SpanKindInternal),
    tracing.WithAttributes(
        attribute.String("tenant.slug", slug),
        attribute.String("tenant.plan", plan),
    ),
)
defer span.End() // always

// Add attributes discovered mid-operation
svc.SetAttributes(ctx, attribute.Int("tenant.user_count", count))

// Record an error — WithErrorStatus marks the span red in Jaeger
if err != nil {
    svc.RecordError(ctx, err, tracing.WithErrorStatus())
    return err
}

// Breadcrumb events — timestamped notes within the span
svc.AddEvent(ctx, "validation_passed")
svc.AddEvent(ctx, "db_insert_complete", attribute.String("table", "tenants"))
```

### Span Kinds

The kind tells the tracing backend how to draw the relationship between spans:

| Kind | When to use |
|------|------------|
| `SpanKindServer` | Handling an incoming HTTP/gRPC request |
| `SpanKindClient` | Making an outgoing HTTP/gRPC call |
| `SpanKindInternal` | Internal function call (default) |
| `SpanKindProducer` | Publishing a message to a queue |
| `SpanKindConsumer` | Consuming a message from a queue |

### HTTP Propagation

This is how distributed tracing works across services. The trace ID and span ID are injected into HTTP headers on outgoing requests and extracted from headers on incoming requests. Without this, traces break at service boundaries.

```go
// In your HTTP server middleware — extract from incoming request
ctx = svc.ExtractHTTPHeaders(ctx, r.Header)

// When making an outgoing HTTP call — inject into request
req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
svc.InjectHTTPHeaders(ctx, req.Header)
```

### Correlating Traces with Logs

Always include the trace ID in your error logs so you can jump from a log entry to the full trace:

```go
traceID := svc.GetTraceID(ctx)

log.ErrorContext(ctx, "payment failed", logger.Fields{
    "error":    err.Error(),
    "trace_id": traceID, // paste this into Jaeger search
})
```

### Sampling Rate

`SamplingRate: 1.0` means trace every request. This is fine in development. In high-traffic production, it can be expensive — set it to `0.1` (10%) or use head-based sampling at the collector level. For critical paths (auth, payment), keep it at 1.0.

### Semantic Convention Helpers

OTel defines standard attribute names for common systems. Use these so your traces are compatible with dashboards and tools out of the box:

```go
tracing.HTTPAttributes("POST", "/api/tenants", userAgent, 201)
tracing.DBAttributes("postgresql", "erp_db", "INSERT INTO tenants ...")
tracing.RPCAttributes("grpc", "TenantService", "CreateTenant")
tracing.MessagingAttributes("kafka", "events.tenant.created", "publish")
```

---

## 3. Metrics

**Package:** `awo.so/internal/shared/metrics`

### Why This Design?

Traces show you individual requests. Logs show you individual events. Neither is good for answering questions like: *"How many requests per second are we handling? What is the 99th percentile latency? How many errors happened in the last 5 minutes?"*

These are aggregate, time-series questions. Metrics are the answer. A metric is a number that changes over time and gets scraped/pushed to a time-series database (Prometheus) where you can query and graph it.

The package supports both **Prometheus** (pull model — Prometheus scrapes your `/metrics` endpoint) and **OpenTelemetry** (push model — your service pushes to a collector). Prometheus is the simpler choice unless you already have an OTel pipeline.

When `Enabled: false`, every method becomes a no-op with zero cost. This lets you add instrumentation everywhere without worrying about it breaking non-production environments that don't have Prometheus.

### The Three Metric Types and When to Use Each

**Counter** — only goes up. Use for counting things that happened.
> "How many requests did we serve?" "How many errors occurred?" "How many emails were sent?"

**Gauge** — goes up and down. Use for current state measurements.
> "How many active connections right now?" "How deep is the job queue?" "How many tenants are provisioned?"

**Histogram** — records the distribution of values. Use for latency and sizes.
> "What is the 95th percentile request duration?" "What is the distribution of payload sizes?"

Never use a counter for something that can decrease (use gauge). Never use a gauge for something you want to compute percentiles on (use histogram).

### Labels (Dimensions)

Every metric can have labels — key-value pairs that let you slice the data. For example, instead of separate metrics for each HTTP route, have one metric with a `route` label:

```go
// This lets you query: sum(requests_total{method="GET"}) by (route)
counter := svc.Counter("requests_total", "...", "method", "route", "status")
counter.Inc(metrics.Fields{"method": "GET", "route": "/api/tenants", "status": "200"})
```

**Cardinality warning:** never use a high-cardinality value as a label (user IDs, tenant IDs, request IDs). This creates millions of time series and kills Prometheus. Labels should have a small, bounded set of possible values.

### Initialization

```go
svc, err := metrics.NewMetricsService(metrics.MetricsConfig{
    Provider:  "prometheus", // "prometheus" | "otel"
    Namespace: "erp",        // metric name prefix: erp_tenant_requests_total
    Subsystem: "tenant",     // second prefix component
    Enabled:   true,
})
defer svc.Close()

// Expose the Prometheus scrape endpoint
http.Handle("/metrics", svc.Handler())
```

### Counter

```go
// Register with label key names, use with label values
counter := svc.Counter("requests_total", "Total HTTP requests", "method", "status")
counter.Inc(metrics.Fields{"method": "POST", "status": "201"})
counter.Add(5, metrics.Fields{"method": "GET", "status": "200"})

// Shorthand — auto-registers, infers label keys from fields
svc.IncrementCounter("errors_total", metrics.Fields{"type": "validation"})
```

### Gauge

```go
gauge := svc.Gauge("active_sessions", "Active user sessions", "tenant_plan")
gauge.Set(42, metrics.Fields{"tenant_plan": "pro"})
gauge.Inc(metrics.Fields{"tenant_plan": "pro"})
gauge.Dec(metrics.Fields{"tenant_plan": "pro"})

svc.SetGauge("queue_depth", 100, metrics.Fields{"queue": "email"})
```

### Histogram

```go
hist := svc.Histogram(
    "request_duration_seconds",
    "HTTP request duration in seconds",
    metrics.StandardHTTPDurationBuckets(), // [0.001, 0.005, 0.01, 0.025, ... 10]
    "route", "method",
)
hist.Observe(0.042, metrics.Fields{"route": "/api/tenants", "method": "GET"})

// Shorthand — uses Prometheus default buckets
svc.ObserveHistogram("db_query_seconds", 0.005, metrics.Fields{"table": "users"})
```

### Timer — The Most Common Pattern

Use this for any operation you want to measure. It starts a stopwatch and records to a histogram when stopped:

```go
// Explicit stop
timer := svc.Timer("db_query", metrics.Fields{"table": "tenants", "op": "insert"})
result, err := db.Exec(query)
elapsed := timer.Stop() // records duration, returns it too

// Wrap a function — cleaner for synchronous work
elapsed := svc.TimerFunc("send_email", metrics.Fields{"type": "welcome"}, func() {
    emailClient.Send(msg)
})
```

### Predefined Buckets

```go
metrics.StandardHTTPDurationBuckets() // 1ms → 10s  — for HTTP handlers
metrics.StandardSizeBuckets()         // 100B → 100MB — for request/response sizes
```

---

## 4. Errors

**Package:** `awo.so/internal/shared/errors`

### Why This Design?

In a layered architecture (handler → service → repository), errors need to travel upward through layers that have different concerns:

- The **repository** knows about database errors (constraint violations, connection failures)
- The **service** knows about business rules (user already exists, tenant suspended)
- The **handler** needs to know what HTTP status to return and what message is safe to show to the client

A raw `pgx` error is meaningless to a handler. A `fmt.Errorf("user not found")` string is unmatchable without string comparison. Without a structured error vocabulary, each developer invents their own patterns and the codebase becomes inconsistent.

This package solves it with three ideas:

**1. Rich error types** that carry everything needed at every layer: HTTP status, machine-readable code, human-readable message, debug details, user-facing suggestions, category, and severity.

**2. Error codes** (strings like `"USER_NOT_FOUND"`) as the stable identity of an error. Code can be checked with `errors.As` through any wrapping chain — even through `fmt.Errorf("%w", err)`.

**3. A layered architecture** where each layer converts errors to the appropriate type:

```
Repository:  pgx error  →  *RepositoryError  (internal, not shown to users)
Service:     sentinel   →  *BusinessError    (domain meaning, HTTP status)
Handler:     any error  →  HTTPError         (safe JSON for the client)
```

### Error Flow Through Layers

```
pgx constraint violation
        │
        ▼ (repo parses it)
*RepositoryError{Code: "QUERY_FAILED", ...}
        │  or
*BusinessError{Code: "USER_EXISTS", HTTPStatus: 409}
        │
        ▼ (service receives it, may wrap or re-raise)
*BusinessError propagates up unchanged
        │
        ▼ (handler calls ToHTTPError)
HTTPError{Status: 409, Code: "USER_EXISTS", Message: "...", Suggestions: [...]}
        │
        ▼
JSON response to client
```

### Error Types

| Type | Layer | Contains |
|------|-------|---------|
| `*BusinessError` | service → handler | code, message, HTTP status, details, suggestions, category, severity |
| `*RepositoryError` | repository only | code, message, operation, table — never exposed to client |
| `ValidationErrors` | anywhere | slice of field-level errors with codes |
| `*ErrorCollection` | service | multiple errors from a batch operation |

### Module-Specific Errors (Important Architecture Decision)

The shared package defines the *types and building blocks*. Each module defines its *own errors* in a local `errors.go`. This keeps domain concerns separated:

```
internal/shared/errors/        ← types, constructors, HTTP conversion, common codes
internal/core/tenant/errors.go ← tenant-domain errors using shared types
internal/core/finance/errors.go← finance-domain errors using shared types
internal/core/iam/errors.go    ← IAM-domain errors using shared types
```

Never add business errors for a specific module into the shared package. The shared package is for cross-cutting concerns only.

### Using Predefined Errors

```go
return errors.ErrTenantNotFound
return errors.ErrTenantExists
return errors.ErrSubdomainAlreadyExists
return errors.ErrFeatureNotEnabled
return errors.ErrUserNotFound
return errors.ErrUserExists
return errors.ErrInvalidCredentials
return errors.ErrUnauthorized
return errors.ErrForbidden
return errors.ErrAccountLocked
return errors.ErrEntityNotFound
return errors.ErrRoleNotFound
return errors.ErrCircularReference
```

### Contextual Constructors (Preferred Over Sentinels)

Sentinels are convenient but carry no context. Prefer contextual constructors — they include the specific ID or value that caused the error, which makes debugging vastly faster:

```go
// Instead of: return errors.ErrUserNotFound
return errors.NewUserNotFoundError(userID) // includes user_id in details

return errors.NewUserNotFoundByEmailError(email)
return errors.NewEntityNotFoundError(entityID)
return errors.NewRoleNotFoundError(roleID)
return errors.NewInvalidCredentialsError(email, attemptCount)
return errors.NewEntityNameExistsError(name)
return errors.NewFeatureNotEnabledError("advanced_reports", "Professional")
return errors.NewCircularReferenceError(entityID, parentID)
```

### Custom Business Errors

For domain errors not already defined:

```go
return errors.NewBusinessError("ORDER_EXPIRED", "Order has expired").
    WithHTTPStatus(http.StatusGone).
    WithCategory(errors.CategoryBusiness).
    WithDetail("order_id", orderID).
    WithDetail("expired_at", expiredAt).
    WithSuggestion("Place a new order to continue")
```

### Repository Errors (in repo layer only)

Wrap raw database errors so they carry context but don't leak DB internals upward:

```go
return errors.NewRepositoryError("QUERY_FAILED", "Failed to fetch tenant", pgErr).
    WithOperation("SELECT").
    WithTable("tenants")
```

### Validation Errors

Collect all field errors before returning — never fail on the first one:

```go
var ve errors.ValidationErrors

if !isValidEmail(req.Email) {
    ve.Add("email", "Invalid email format")
}
if len(req.Name) < 2 {
    ve.AddWithCode("name", "Name is too short", errors.CodeTooShort)
}
if req.Age < 18 {
    ve.AddWithValue("age", "Must be 18 or older", errors.CodeInvalidRange, req.Age)
}

if ve.HasErrors() {
    return ve // HTTP 400 with {"validation_errors": {"email": [...], "name": [...]}}
}
```

### Checking Errors

Always use `errors.As` or the provided helpers — never compare error strings:

```go
// Convenience helpers (use these)
errors.IsUserNotFound(err)
errors.IsTenantNotFound(err)
errors.IsUnauthorized(err)
errors.IsForbidden(err)
errors.IsConflict(err)           // any "already exists" error
errors.IsNotFoundError(err)      // any "not found" error
errors.IsValidationError(err)
errors.IsAuthenticationError(err)
errors.IsTemporaryError(err)     // retryable — good for retry logic

// For custom error codes
errors.IsBusinessErrorCode(err, "ORDER_EXPIRED")

// Direct unwrapping — safe through fmt.Errorf("%w", ...) chains
var be *errors.BusinessError
if errors.As(err, &be) {
    // be.Code, be.HTTPStatus, be.Category, be.Details, be.Suggestions
}
```

### HTTP Conversion (handlers only)

`ToHTTPError` uses `errors.As` internally, so it correctly unwraps any error chain:

```go
func handleErr(w http.ResponseWriter, err error) {
    httpErr := errors.ToHTTPError(err)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(httpErr.Status)
    json.NewEncoder(w).Encode(httpErr)
}

// Response shape:
// {
//   "status": 404,
//   "code": "USER_NOT_FOUND",
//   "message": "User not found",
//   "details": { "user_id": "abc123" },
//   "suggestions": ["Verify the user ID is correct"],
//   "timestamp": "2026-03-24T12:00:00Z"
// }
```

### Error Categories and Severity — Why They Matter

Categories let you route errors to different alert channels without checking error codes:

```go
errors.CategoryValidation   // user made a mistake — no alert, 400 response
errors.CategoryBusiness     // domain rule violated — log warn, 4xx response
errors.CategorySecurity     // auth/authz failure — may alert security team
errors.CategoryTenant       // tenant-level issue — notify tenant admin
errors.CategoryRepository   // database failure — alert on-call
errors.CategoryIntegration  // external service down — alert + circuit break
errors.CategorySystem       // infra failure — critical alert, 500 response
```

Severity helps prioritize:

```go
errors.SeverityInfo      // informational, no action needed
errors.SeverityWarning   // worth knowing, investigate eventually
errors.SeverityError     // something failed, investigate soon
errors.SeverityCritical  // system in danger, act immediately
```

### Creating Module-Specific Errors (Best Practice)

```go
// internal/core/finance/errors.go
package finance

import (
    "net/http"
    "awo.so/internal/shared/errors"
)

const CodeInsufficientFunds = "INSUFFICIENT_FUNDS"

var ErrInsufficientFunds = errors.NewBusinessError(CodeInsufficientFunds, "Insufficient funds").
    WithHTTPStatus(http.StatusConflict).
    WithCategory(errors.CategoryBusiness).
    WithSuggestion("Check your account balance before proceeding")

func NewInsufficientFundsError(accountID string, required, available float64) *errors.BusinessError {
    return errors.NewBusinessError(CodeInsufficientFunds, "Insufficient funds").
        WithHTTPStatus(http.StatusConflict).
        WithCategory(errors.CategoryBusiness).
        WithDetail("account_id", accountID).
        WithDetail("required", required).
        WithDetail("available", available)
}

func IsInsufficientFunds(err error) bool {
    return errors.IsBusinessErrorCode(err, CodeInsufficientFunds)
}
```

---

## 5. Cache

**Package:** `awo.so/internal/platform/cache`

### Why This Design?

Every database read has latency. For an ERP system where the same tenant config, user permissions, or entity tree might be read hundreds of times per second, hitting Postgres every time is wasteful and slow. Caching solves this.

But caching in a multi-tenant system introduces a critical problem: **tenant data isolation**. If two tenants' data can share a cache key (even accidentally), you have a data leak. This package eliminates that risk by **automatically namespacing every key with the tenant identity** extracted from the context. Developers never think about tenant isolation in cache keys — the infrastructure enforces it.

### The Two-Layer Cache Architecture

```
Request
   │
   ▼
┌──────────────────────────────┐
│  L1: In-Memory Cache         │  Fast (nanoseconds), no network
│  (per process, tenant-aware) │  Limited size (LRU eviction)
└──────────────┬───────────────┘  Lost on restart
               │ miss
               ▼
┌──────────────────────────────┐
│  L2: Redis Cache             │  Slower (milliseconds), network hop
│  (shared, tenant-aware)      │  Large capacity, survives restarts
└──────────────┬───────────────┘  Shared across all service instances
               │ miss
               ▼
┌──────────────────────────────┐
│  Database (PostgreSQL)       │  Slowest (tens of milliseconds)
└──────────────────────────────┘  Source of truth
```

For the hottest data (feature flags, tenant config, user permissions), use L1 (memory). For data that needs to survive restarts or be shared across instances, use L2 (Redis). The database is only hit on a miss at both layers.

### How Tenant Namespacing Works

You set tenant identity in the context **once** — typically in HTTP middleware — and every cache operation from that point forward is automatically scoped:

```
You pass key:         "user:123"
Context contains:     tenant_id = "a1b2c3d4-..."
Redis key becomes:    erp:tenant:a1b2c3d4-...:user:123
```

This is enforced at the infrastructure level. No developer can accidentally read another tenant's data. Two tenants can have the same logical key and they will never collide.

### When to Use Cache vs Not

**Cache when:**
- Data is read frequently but changes infrequently (tenant config, user roles, entity tree)
- The same data is read multiple times per request
- You can tolerate slightly stale data (within the TTL window)

**Do not cache when:**
- Data must be perfectly up-to-date every time (financial balances, audit logs)
- Data changes on every write and is only read once
- The data is very large and varies per request (search results with dynamic filters)

### The Circuit Breaker

The cache has a built-in circuit breaker. After 5 consecutive Redis errors, it "opens" and all Redis operations immediately return `ErrCircuitOpen` for 30 seconds, instead of waiting to time out each time. This prevents a Redis outage from cascading into your service being completely unresponsive. Your service degrades gracefully (slower, hitting DB more) instead of failing completely.

### Initialization

```go
import "awo.so/internal/platform/cache"

baseRedis := &config.RedisConfig{Host: "localhost", Port: 6379, DB: 0}
cfg := cache.DefaultRedisConfig(baseRedis)

cfg.EnableTracing  = true  // trace each cache operation
cfg.EnableMetrics  = true  // record hit/miss/latency metrics
cfg.EnableLogging  = true  // log errors and debug info

// Simple (no observability)
client, err := cache.NewRedisClient(cfg)

// With observability — any arg can be nil
client, err := cache.NewRedisClientWithObservability(cfg, log, tracingSvc, metricsSvc)
defer client.Close()
```

### Setting Tenant Context

Set this once in middleware, before the request reaches your service:

```go
// Use whichever identifier you have available
ctx = cache.WithTenantID(ctx, tenantUUID)       // from parsed UUID
ctx = cache.WithTenantSlug(ctx, "acme-corp")    // from subdomain or header
ctx = cache.WithTenantSubdomain(ctx, "acme")    // from host header
```

If tenant context is missing and `RequireTenantContext = true` (the default), every cache call will return `ErrNoTenantContext`. This is intentional — it is a programming error, not a runtime error. Fix your middleware.

### Core Operations

```go
// Set — value is JSON-serialized, compressed if > 1KB
err = client.Set(ctx, "user:123", userObj, 10*time.Minute)

// Get — pass a pointer, JSON is deserialized into it
var user UserDTO
err = client.Get(ctx, "user:123", &user)
if errors.Is(err, cache.ErrCacheMiss) {
    // normal — load from database, then Set to warm cache
}

err = client.Delete(ctx, "user:123")
err = client.Flush(ctx) // delete all keys for this tenant
```

### Batch Operations

Use batch operations when you need multiple keys — one network round trip instead of N:

```go
// Write many at once
err = client.MSet(ctx, map[string]any{
    "user:1": user1,
    "user:2": user2,
    "user:3": user3,
}, 10*time.Minute)

// Read many at once
results, err := client.MGet(ctx, []string{"user:1", "user:2", "user:3"})
for _, r := range results {
    if r.Err == nil {
        // r.Key — the key you passed
        // r.Value — the raw value
    }
}

err = client.MDelete(ctx, []string{"user:1", "user:2"})
```

### Pattern Operations

Useful for cache invalidation (e.g. invalidate all cached data for an entity):

```go
// Find keys matching a glob pattern (tenant-scoped automatically)
keys, err := client.Keys(ctx, "user:*")

// Invalidate all cached users for this tenant
err = client.DeletePattern(ctx, "user:*")

exists, err := client.Exists(ctx, "user:123")
ttl, err    := client.TTL(ctx, "user:123")
err          = client.Expire(ctx, "user:123", 30*time.Minute)
```

### In-Memory Cache (L1 — no network hop)

For the very hottest data — things read on every request:

```go
// Tenant-scoped memory cache — isolated per tenant, same namespacing as Redis
err = client.SetMemory(ctx, "config:features", featureFlags, 5*time.Minute)
err = client.GetMemory(ctx, "config:features", &featureFlags)
err = client.DeleteMemory(ctx, "config:features")

// Global memory cache — shared across ALL tenants, no context needed
// Use only for system-wide immutable data: tax formulas, currency codes, lookup tables
err = client.SetGlobalMemory("formula:vat_standard", "amount * 0.20", 1*time.Hour)
err = client.GetGlobalMemory("formula:vat_standard", &formula)
err = client.DeleteGlobalMemory("formula:vat_standard")
```

> **Rule:** If the data differs per tenant → use tenant-scoped memory. If it is the same for every tenant and rarely changes → use global memory.

### Health and Statistics

```go
if err := client.Ping(ctx); err != nil {
    // Redis unreachable — return 503 from health endpoint
}

stats := client.Stats()
// stats.HitRatio          — if this drops below ~0.7, review your TTLs or caching strategy
// stats.AverageLatency    — if this spikes, Redis may be under pressure
// stats.ConnectionsActive — watch for connection pool exhaustion
// stats.Errors            — rising errors may indicate Redis instability
// stats.MemoryCacheSize   — watch for eviction pressure (hitting MemoryCacheMaxSize)

client.Reset() // reset counters — useful at start of each metrics window
```

### Cache Errors

```go
cache.ErrCacheMiss        // expected — key expired or was never set, fetch from DB
cache.ErrNoTenantContext  // bug — middleware did not set tenant context
cache.ErrCircuitOpen      // Redis is unhealthy, circuit breaker activated
cache.ErrInvalidData      // data corruption — log, treat as miss, do not panic
cache.ErrNilValue         // nil passed as Set value — programming error
```

### Default Configuration

| Setting | Default | Notes |
|---------|---------|-------|
| `PoolSize` | 10 | Redis connection pool — increase for high concurrency |
| `MaxRetries` | 3 | Retry transient failures automatically |
| `EnableCompression` | true | Gzip values > 1KB — reduces Redis memory usage |
| `RequireTenantContext` | true | Enforces tenant isolation — do not disable |
| `AllowGlobalOperations` | false | If false, ops without tenant ctx return error |
| `EnableMemoryCache` | true | L1 cache enabled |
| `MemoryCacheMaxSize` | 1000 items | LRU eviction above this — tune per service |
| `EnableCircuitBreaker` | true | Opens after 5 consecutive Redis errors |
| `CircuitBreakerTimeout` | 30s | Time before retrying Redis after circuit opens |

---

## 6. Putting It All Together

### How Context Flows

The `context.Context` is the backbone of the whole system. It carries:
- The current **trace span** (tracing propagation)
- The **tenant identity** (cache namespacing, error enrichment)
- The **request cancellation signal** (timeouts, client disconnects)

Middleware sets all of this up once, and every layer downstream just reads from ctx:

```
Middleware:
  ctx = tracingSvc.ExtractHTTPHeaders(ctx, r.Header) // attach incoming trace
  ctx = cache.WithTenantID(ctx, tenantID)             // attach tenant
  ctx = context.WithTimeout(ctx, 30*time.Second)      // attach deadline

Handler → Service → Repository: all receive the same ctx
```

### A Complete Service Method

```go
func (s *TenantService) GetTenant(ctx context.Context, id string) (*Tenant, error) {
    // Tracing — start a span, it will be a child of whatever span is in ctx
    ctx, span := s.tracer.StartSpan(ctx, "TenantService.GetTenant",
        tracing.WithSpanKind(tracing.SpanKindInternal),
        tracing.WithAttributes(attribute.String("tenant.id", id)),
    )
    defer span.End()

    // Metrics — measure how long this method takes
    timer := s.metrics.Timer("get_tenant_duration", metrics.Fields{"layer": "service"})
    defer timer.Stop()

    // Cache — pass only the logical key, tenant prefix is automatic
    var tenant Tenant
    if err := s.cache.Get(ctx, "tenant:"+id, &tenant); err == nil {
        s.metrics.IncrementCounter("cache_hits", metrics.Fields{"entity": "tenant"})
        s.log.DebugContext(ctx, "tenant cache hit", logger.Fields{"tenant_id": id})
        return &tenant, nil
    }

    s.metrics.IncrementCounter("cache_misses", metrics.Fields{"entity": "tenant"})

    // Database — error coming back is already a *BusinessError or *RepositoryError
    t, err := s.repo.GetByID(ctx, id)
    if err != nil {
        s.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
        s.log.ErrorContext(ctx, "failed to load tenant", logger.Fields{
            "tenant_id": id,
            "error":     err.Error(),
            "trace_id":  s.tracer.GetTraceID(ctx),
        })
        return nil, err
    }

    // Warm the cache for the next request
    _ = s.cache.Set(ctx, "tenant:"+id, t, 10*time.Minute)

    return t, nil
}
```

### A Complete Handler

```go
func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")

    tenant, err := h.service.GetTenant(r.Context(), id)
    if err != nil {
        // ToHTTPError handles every error type correctly:
        // *BusinessError → uses its HTTPStatus and Code
        // *RepositoryError → 500, hides internal details
        // ValidationErrors → 400 with field-level breakdown
        httpErr := errors.ToHTTPError(err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(httpErr.Status)
        json.NewEncoder(w).Encode(httpErr)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tenant)
}
```

---

## 7. Testing with Mocks

### Why Mocks for Infrastructure?

You don't want your unit tests to actually hit Redis or Jaeger. Infrastructure is slow, stateful, and makes tests flaky. Every package exposes an interface, and every interface has a generated mock (`*_mock.go`) using `gomock`. Inject the mock in tests, inject the real implementation in production — your service code never knows the difference.

### Using Mocks

```go
import (
    "testing"
    "awo.so/internal/shared/logger"
    "awo.so/internal/shared/tracing"
    "awo.so/internal/shared/metrics"
    "awo.so/internal/platform/cache"
    "go.uber.org/mock/gomock"
)

func TestTenantService_GetTenant_CacheHit(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockLog     := logger.NewMockLogger(ctrl)
    mockTracer  := tracing.NewMockService(ctrl)
    mockSpan    := tracing.NewMockSpan(ctrl)
    mockMetrics := metrics.NewMockMetricsProvider(ctrl)
    mockCache   := cache.NewMockService(ctrl)

    ctx := context.Background()

    // Define what we expect to happen
    mockTracer.EXPECT().
        StartSpan(gomock.Any(), "TenantService.GetTenant", gomock.Any()).
        Return(ctx, mockSpan)

    mockSpan.EXPECT().End()

    mockCache.EXPECT().
        Get(gomock.Any(), "tenant:abc123", gomock.Any()).
        Return(nil) // nil error = cache hit

    mockLog.EXPECT().
        DebugContext(gomock.Any(), "tenant cache hit", gomock.Any())

    // ... build service with mocks and run the test
}

func TestTenantService_GetTenant_NotFound(t *testing.T) {
    // ...
    mockCache.EXPECT().
        Get(gomock.Any(), "tenant:abc123", gomock.Any()).
        Return(cache.ErrCacheMiss)

    mockRepo.EXPECT().
        GetByID(gomock.Any(), "abc123").
        Return(nil, errors.ErrTenantNotFound)

    mockTracer.EXPECT().
        RecordError(gomock.Any(), errors.ErrTenantNotFound, gomock.Any())

    // ...assert the service returns ErrTenantNotFound
}
```

### No-Op Implementations for Simple Tests

When you only care about testing business logic and don't want to set up full mock expectations for every infra call:

```go
// Tracing — built-in no-op, zero overhead
tracingSvc := tracing.NewNoOpService()

// Metrics — disabled provider, all calls are no-ops
metricsSvc, _ := metrics.NewMetricsService(metrics.MetricsConfig{Enabled: false})

// Logger — mock with AnyTimes() so any log call is silently accepted
mockLog := logger.NewMockLogger(ctrl)
mockLog.EXPECT().DebugContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
mockLog.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
mockLog.EXPECT().ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
```

### Testing Error Handling

The errors package makes assertions clean and precise:

```go
_, err := svc.GetTenant(ctx, "nonexistent")

assert.True(t, errors.IsTenantNotFound(err))

// If you need to check details
var be *errors.BusinessError
require.True(t, errors.As(err, &be))
assert.Equal(t, http.StatusNotFound, be.HTTPStatus)
assert.Equal(t, "TENANT_NOT_FOUND", be.Code)

// HTTP response shape
httpErr := errors.ToHTTPError(err)
assert.Equal(t, 404, httpErr.Status)
```
