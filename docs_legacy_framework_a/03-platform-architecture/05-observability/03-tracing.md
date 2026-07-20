> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Distributed Tracing
portal: 3 — Platform Architecture
section: 05-observability
audience: [architect, backend-engineer, sre]
related:
  - "[Observability Overview](01-observability-overview.md)"
  - "[Metrics](04-metrics.md)"
  - "[Tracing Guide](../../04-backend-engineering/00-module-development-guide/08-instrumentation/03-tracing.md)"
---

# Distributed Tracing

## Setup

```go
// internal/platform/telemetry/tracer.go
func InitTracer(cfg OTelConfig) (*sdktrace.TracerProvider, error) {
    exporter, err := otlptracehttp.New(context.Background(),
        otlptracehttp.WithEndpoint(cfg.Endpoint),
        otlptracehttp.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName(cfg.ServiceName),
            semconv.ServiceVersion(cfg.Version),
        )),
        sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
    )
    otel.SetTracerProvider(tp)
    return tp, nil
}
```

## Tracer per Module

Each module gets its own named tracer:

```go
tracer := otel.Tracer("awo.so/contracts")
```

## Span Lifecycle

```go
func (s *contractService) Create(ctx context.Context, req CreateContractRequest) (*domain.Contract, error) {
    ctx, span := s.tracer.Start(ctx, "contract.service.create",
        trace.WithAttributes(
            attribute.String("tenant.id", req.TenantID.String()),
            attribute.String("contract.type", string(req.ContractType)),
        ),
    )
    defer span.End()

    // ...

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    span.SetAttributes(attribute.String("contract.id", contract.ID.String()))
    return contract, nil
}
```

## HTTP Span (Fiber)

Fiber does not auto-instrument. Add OTel middleware:

```go
app.Use(otelfiber.Middleware(
    otelfiber.WithTracerProvider(tracerProvider),
    otelfiber.WithPropagators(propagation.TraceContext{}),
))
```

This creates a root span for each HTTP request and propagates `traceparent` headers for distributed tracing across services.

## Async Goroutine Tracing

Async goroutines (audit, events, notifications) must NOT inherit the request context — that context cancels when the request completes. Instead, link the goroutine span to the parent:

```go
func (s *contractService) recordAuditAsync(ctx context.Context, e audit.Event) {
    // Capture link to parent span before launching goroutine
    parentSpanCtx := trace.SpanFromContext(ctx).SpanContext()

    go func() {
        asyncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        asyncCtx, span := s.tracer.Start(asyncCtx, "contract.audit.record",
            trace.WithLinks(trace.Link{SpanContext: parentSpanCtx}),
        )
        defer span.End()

        if err := s.auditSvc.Record(asyncCtx, e); err != nil {
            span.RecordError(err)
        }
    }()
}
```

## Sampling

| Environment | Sample Rate |
|-------------|------------|
| Development | 100% |
| Staging | 100% |
| Production | 10% (configurable) |
| High-value paths (auth, payments) | 100% (always sample) |

Configure via `OTEL_TRACES_SAMPLER` and `OTEL_TRACES_SAMPLER_ARG` env vars.
