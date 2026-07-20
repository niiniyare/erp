> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Observability Overview
portal: 3 — Platform Architecture
section: 05-observability
audience: [architect, backend-engineer, sre]
related:
  - "[Logging](02-logging.md)"
  - "[Tracing](03-tracing.md)"
  - "[Metrics](04-metrics.md)"
  - "[Instrumentation Guide](../../04-backend-engineering/00-module-development-guide/08-instrumentation/01-instrumentation-overview.md)"
---

# Observability Overview

AwoERP implements the three pillars of observability: structured logs (Zerolog), distributed traces (OpenTelemetry), and metrics (Prometheus).

## Golden Rule

**Instrumentation never fails a request.** Span recording, metric increments, audit writes, and notification sends are always fire-and-forget. A broken observability pipeline must not take down the application.

## Three Pillars

| Pillar | Technology | Export |
|--------|-----------|--------|
| Logs | Zerolog (structured JSON) | stdout → log aggregator |
| Traces | OpenTelemetry SDK | OTLP → Jaeger / Tempo |
| Metrics | Prometheus client | HTTP scrape → Prometheus |

## Correlation

All three pillars are correlated via:
- `trace_id`: from the active OTel span, injected into log fields
- `request_id`: from `X-Request-ID` header, in logs and response
- `tenant_id`: in all logs, traces, and metric labels

## Service Skeleton Pattern

Every service method follows this instrumentation pattern:

```go
func (s *contractService) Create(ctx context.Context, req CreateContractRequest) (*domain.Contract, error) {
    // 1. Start span
    ctx, span := s.tracer.Start(ctx, "contract.service.create")
    defer span.End()

    // 2. Start metric timer
    timer := s.metrics.StartTimer("contract_service_create")
    defer timer.ObserveDuration()

    // 3. Log entry (debug level)
    s.log.Debug().
        Str("tenant_id", req.TenantID.String()).
        Msg("contract.service.create started")

    // ... business logic ...

    // 4. On error: record to span
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        s.metrics.IncrementCounter("contract_create_errors_total")
        return nil, err
    }

    // 5. On success: increment counter
    s.metrics.IncrementCounter("contract_create_total")
    return contract, nil
}
```

## Log Levels

| Level | When |
|-------|------|
| `ERROR` | Unexpected failures, panics, 5xx responses |
| `WARN` | Non-fatal issues: async failures, deprecated endpoints |
| `INFO` | Server start/stop, migration complete, key lifecycle events |
| `DEBUG` | Per-request detail, service entry/exit |

Production default: `INFO`. Debug enabled per-tenant via feature flag for troubleshooting.

## Span Naming

```
{module}.{layer}.{operation}
```

Examples:
```
contract.service.create
contract.repository.get_by_id
contract.handler.submit
iam.session.resolve
```

## Metric Naming

```
{module}_{noun}_{unit}_total   (counters)
{module}_{noun}_{unit}         (gauges, histograms)
```

Examples:
```
contract_requests_total
contract_request_duration_seconds
contract_active_count
finance_ledger_postings_total
```
