---
title: "Chapter 46: Observability Stack"
part: "Part VIII — Deployment and Operations"
chapter: 46
section: "46-observability-stack"
related:
  - "[Chapter 32: Workflow Observability](../part-05-workflow/32-workflow-observability.md)"
  - "[Chapter 13: Middleware Pipeline](../part-03-api/13-middleware-pipeline.md)"
---

# Chapter 46: Observability Stack

A running ERP must be observable — you must be able to understand what it is doing and diagnose what went wrong without attaching a debugger. Awo uses structured logging (slog), distributed tracing (OpenTelemetry → Tempo), metrics (Prometheus → Grafana), and an immutable audit log for business events.

---

## 46.1. Structured Logging with slog

### 46.1.1. Log Format

All logs are structured JSON, written to stdout. Log aggregators (Loki, CloudWatch) parse and index the JSON fields:

```json
{
  "time": "2025-07-15T10:23:45.123Z",
  "level": "INFO",
  "msg": "request completed",
  "request_id": "req_01j3abc",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "method": "POST",
  "path": "/api/v1/invoices",
  "status": 201,
  "latency_ms": 47,
  "entity": "Invoice",
  "entity_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479"
}
```

### 46.1.2. slog Setup

```go
func newLogger(env string) *slog.Logger {
    var handler slog.Handler

    if env == "development" {
        // Human-readable for local dev
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    } else {
        // JSON for production (parsed by Loki)
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })
    }

    return slog.New(handler)
}
```

### 46.1.3. Request-Scoped Logger

Each request carries a logger with request context pre-attached:

```go
// Middleware: attach request logger to context
func requestLoggerMiddleware(logger *slog.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        requestID := c.Locals("requestID").(string)

        // Attach request fields to logger
        reqLogger := logger.With(
            slog.String("request_id", requestID),
            slog.String("method", c.Method()),
            slog.String("path", c.Path()),
        )
        c.Locals("logger", reqLogger)

        err := c.Next()

        // Log completion
        reqLogger.Info("request completed",
            slog.Int("status", c.Response().StatusCode()),
            slog.Int64("latency_ms", time.Since(start).Milliseconds()),
        )
        return err
    }
}

// Usage in handler
func (h *Handler) Create(c *fiber.Ctx) error {
    logger := c.Locals("logger").(*slog.Logger)
    logger.Info("creating invoice", slog.String("customer_id", input.CustomerID.String()))
    // ...
}
```

### 46.1.4. Log Levels in Practice

| Level | When to use | Example |
|---|---|---|
| DEBUG | Detailed execution flow — only in development | "applying filter: account_id = ..." |
| INFO | Normal significant events | "invoice posted", "payroll run started" |
| WARN | Unexpected but handled situations | "tenant config cache miss", "rate limit approached" |
| ERROR | Failures that should be investigated | "failed to post GL entry", "external API timeout" |

Errors include the full stack trace via `slog.Any("error", err)` which logs the error chain via `errors.Unwrap`.

---

## 46.2. Distributed Tracing with OpenTelemetry

### 46.2.1. Trace Setup

```go
func initTracer(cfg Config) (func(), error) {
    exporter, err := otlptracehttp.New(context.Background(),
        otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
        otlptracehttp.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName("awo-api"),
            semconv.ServiceVersion(Version),
            semconv.DeploymentEnvironment(cfg.Env),
        )),
        sdktrace.WithSampler(sdktrace.ParentBased(
            sdktrace.TraceIDRatioBased(cfg.TraceSampleRate), // 0.1 in prod, 1.0 in dev
        )),
    )

    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return func() { tp.Shutdown(context.Background()) }, nil
}
```

### 46.2.2. Fiber OpenTelemetry Middleware

```go
app.Use(otelfiber.Middleware(
    otelfiber.WithSpanNameFormatter(func(ctx *fiber.Ctx) string {
        // Use route pattern, not actual URL (avoids high-cardinality)
        return ctx.Method() + " " + ctx.Route().Path
    }),
))
```

Using the route pattern (`GET /api/v1/invoices/:id`) rather than the actual URL (`GET /api/v1/invoices/550e8400-...`) prevents trace cardinality explosion.

### 46.2.3. Key Spans to Instrument

Manual spans for high-value operations:

```go
// Database query span
func (r *Repo) Query(ctx context.Context, filter Filter) ([]Record, error) {
    ctx, span := otel.Tracer("awo/db").Start(ctx, "db.query",
        trace.WithAttributes(
            attribute.String("db.entity", r.entityName),
            attribute.String("db.tenant", tenant.SlugFromContext(ctx)),
        ))
    defer span.End()
    // ...
}

// External API call span
func (s *MPESAService) STKPush(ctx context.Context, input STKPushInput) error {
    ctx, span := otel.Tracer("awo/mpesa").Start(ctx, "mpesa.stk_push",
        trace.WithAttributes(
            attribute.String("mpesa.phone", maskPhone(input.Phone)),
            attribute.Float64("mpesa.amount", input.Amount),
        ))
    defer span.End()
    // ...
}
```

---

## 46.3. Metrics with Prometheus

### 46.3.1. Custom Business Metrics

Beyond standard HTTP metrics (request count, latency, error rate), Awo exposes business metrics:

```go
var (
    invoicesPosted = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "awo_invoices_posted_total",
        Help: "Total number of invoices posted",
    }, []string{"tenant_plan"})

    payrollRunDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "awo_payroll_run_duration_seconds",
        Help:    "Time to complete a full payroll run",
        Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1s to ~17min
    }, []string{"employee_count_bucket"})

    activeTenantsGauge = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "awo_active_tenants",
        Help: "Number of active tenants",
    })
)
```

**Cardinality discipline**: metrics labels must have bounded cardinality. `tenant_id` as a label would create thousands of time series (one per tenant). Use `tenant_plan` (3-5 values) instead.

### 46.3.2. Grafana Dashboards

Key dashboards:

**API Health Dashboard**:
- Request rate by endpoint
- P50/P95/P99 latency
- Error rate (5xx) by endpoint
- Active database connections

**Tenant Activity Dashboard**:
- Active tenants (24h)
- API requests by plan tier
- Most active tenants (by request volume)
- Workflow execution rate

**Financial Operations Dashboard**:
- Journal entries posted per hour
- Payroll runs completed
- Invoice approval workflow queue depth
- Stuck workflows (running > 1 hour)

---

## 46.4. Audit Log

### 46.4.1. What the Audit Log Is For

The audit log is not an application log (which records technical events for debugging). It records business-meaningful events for compliance and accountability:
- Who changed what, when
- What was the value before and after
- What permission was used to authorise the change

Required by Kenya's Data Protection Act 2019 and best practice for financial systems.

### 46.4.2. Audit Log Entry

```go
type AuditEntry struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    TenantSlug  string
    Timestamp   time.Time
    ActorID     uuid.UUID    // user who performed the action
    ActorName   string       // denormalised for readability after user deletion
    ActorIP     string
    Action      string       // "create" | "update" | "delete" | "submit" | "cancel" | "login" | "logout"
    EntityType  string       // "Invoice" | "Employee" | "JournalEntry"
    EntityID    uuid.UUID
    EntityRef   string       // human-readable: "INV-0045"
    Before      *json.RawMessage  // previous values (for update/delete)
    After       *json.RawMessage  // new values (for create/update)
    Reason      *string      // for sensitive actions
}
```

### 46.4.3. Audit Log Storage

The audit log is written to a separate table (`audit_log` in the public schema) that crosses tenant boundaries — platform admins can query all tenants. It is append-only:

```sql
CREATE TABLE audit_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    timestamp timestamptz NOT NULL DEFAULT now(),
    actor_id uuid,
    actor_name text NOT NULL,
    actor_ip inet,
    action text NOT NULL,
    entity_type text NOT NULL,
    entity_id uuid,
    entity_ref text,
    before_val jsonb,
    after_val jsonb,
    reason text
);

-- No UPDATE or DELETE permissions granted to application role
REVOKE UPDATE, DELETE ON audit_log FROM awo_app;
```

The database role used by the application (`awo_app`) has only INSERT and SELECT on `audit_log`. Even if the application is compromised, it cannot delete audit entries.

### 46.4.4. High-Volume Audit Log

At 1,000 tenants with active users, the audit log can grow rapidly. Management:
- Partition by month (same as GL lines)
- Archive partitions older than 7 years to S3 cold storage
- `before_val` and `after_val` are nullable — for list views, omit them to reduce storage; include on detail view

### 46.4.5. Audit Log in the `after_save` Hook

```go
func auditAfterSave(ctx context.Context, op string, entity EntityRecord) error {
    actor := auth.ActorFromContext(ctx)

    before, after := diffRecord(entity)

    return auditRepo.Insert(ctx, AuditEntry{
        TenantID:   tenant.IDFromContext(ctx),
        ActorID:    actor.UserID,
        ActorName:  actor.Name,
        ActorIP:    actor.IP,
        Action:     op,
        EntityType: entity.Type(),
        EntityID:   entity.ID(),
        EntityRef:  entity.Reference(),
        Before:     before,
        After:      after,
    })
}
```

The `after_save` hook runs inside the same transaction as the entity change. If the audit write fails, the entire transaction rolls back — you cannot have a change without an audit record.
