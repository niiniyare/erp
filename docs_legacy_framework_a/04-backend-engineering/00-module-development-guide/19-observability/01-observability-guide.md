> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Observability Guide
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Observability Architecture](../../../03-platform-architecture/05-observability/01-observability-overview.md)"
  - "[Logging](../../../03-platform-architecture/05-observability/02-logging.md)"
  - "[Tracing](../../../03-platform-architecture/05-observability/03-tracing.md)"
  - "[Metrics](../../../03-platform-architecture/05-observability/04-metrics.md)"
---

# Observability Guide

## Module Logger

Every module gets a child logger scoped to the module name:

```go
func NewContractService(
    // ... other deps
    logger *slog.Logger,
) *ContractService {
    return &ContractService{
        logger: logger.With("module", "contracts"),
    }
}
```

Use child loggers for operations:

```go
func (s *ContractService) Create(ctx context.Context, ...) (*Contract, error) {
    log := s.logger.With("operation", "Create")

    contract, err := s.repo.Create(ctx, tenantID, req)
    if err != nil {
        log.ErrorContext(ctx, "repo create failed",
            "tenant_id", tenantID,
            "error", err,
        )
        return nil, err
    }

    log.InfoContext(ctx, "contract created",
        "contract_id", contract.ID,
        "tenant_id", tenantID,
    )
    return contract, nil
}
```

## Log Levels

| Level | When to use |
|-------|------------|
| `Debug` | Detailed flow tracing — dev only, never in prod by default |
| `Info` | Successful state changes (`contract created`, `payment posted`) |
| `Warn` | Recoverable issues, permission denials, retries |
| `Error` | Operation failed — requires investigation |

Do not log at `Info` for read operations (GET, List) — too noisy.

## Standard Log Fields

Always include these fields in error logs:

```go
s.logger.ErrorContext(ctx, "operation failed",
    "tenant_id", sess.TenantID,
    "user_id",   sess.UserID,
    "contract_id", contract.ID,
    "error",     err,
)
```

The `request_id` is injected automatically by the request logger middleware — no need to pass it manually in service logs.

## Tracing: Span Per Operation

Create a span for each significant operation:

```go
func (s *ContractService) Submit(ctx context.Context, sess ResolvedSession, id uuid.UUID, version int) error {
    ctx, span := s.tracer.Start(ctx, "contracts.ContractService.Submit")
    defer span.End()

    span.SetAttributes(
        attribute.String("contract.id", id.String()),
        attribute.String("tenant.id", sess.TenantID.String()),
    )

    contract, err := s.repo.GetByID(ctx, id, sess.TenantID)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    // ...
}
```

Span naming: `{module}.{Type}.{Method}` — e.g., `contracts.ContractService.Submit`.

## Tracing: Async Goroutines

Async goroutines (audit, events, notifications) must link to the parent span, not inherit the cancelled context:

```go
go func() {
    // Link to parent span — don't propagate the cancellable request context
    ctx := context.Background()
    ctx, span := tracer.Start(ctx, "contracts.ContractService.publishAsync",
        trace.WithLinks(trace.LinkFromContext(parentCtx)),
    )
    defer span.End()
    // ...
}()
```

## Metrics: Module Registration

Register metrics once in the module constructor:

```go
type ContractMetrics struct {
    contractsCreated  prometheus.Counter
    contractsApproved prometheus.Counter
    operationDuration *prometheus.HistogramVec
}

func NewContractMetrics(reg prometheus.Registerer) *ContractMetrics {
    return &ContractMetrics{
        contractsCreated: promauto.With(reg).NewCounter(prometheus.CounterOpts{
            Name: "contracts_created_total",
            Help: "Total contracts created",
        }),
        contractsApproved: promauto.With(reg).NewCounter(prometheus.CounterOpts{
            Name: "contracts_approved_total",
            Help: "Total contracts approved",
        }),
        operationDuration: promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
            Name:    "contracts_operation_duration_seconds",
            Help:    "Contract operation durations",
            Buckets: prometheus.DefBuckets,
        }, []string{"operation", "status"}),
    }
}
```

## Metrics: Record in Service

```go
func (s *ContractService) Create(ctx context.Context, ...) (*Contract, error) {
    start := time.Now()
    contract, err := s.repo.Create(ctx, tenantID, req)

    status := "success"
    if err != nil {
        status = "error"
    }
    s.metrics.operationDuration.WithLabelValues("create", status).
        Observe(time.Since(start).Seconds())

    if err == nil {
        s.metrics.contractsCreated.Inc()
    }

    return contract, err
}
```

## Health Check Integration

If your module has a dependency that can fail (e.g., a required external service), register a health check:

```go
// In NewContractService
func (s *ContractService) HealthCheck(ctx context.Context) error {
    // Simple DB connectivity check
    return s.repo.Ping(ctx)
}
```

Register with the health server:

```go
healthServer.Register("contracts-db", contractSvc.HealthCheck)
```

## What Not to Log

```go
// ❌ Never log sensitive data
log.Info("user authenticated", "password", req.Password)
log.Info("session", "token", sessionToken)
log.Info("request body", "body", string(rawBody))  // may contain PII

// ❌ Don't log every read operation — too noisy
log.Info("contract fetched", "id", id)  // skip for GET operations

// ❌ Don't duplicate error context already in the error message
log.Error("error: contract not found", "error", domain.ErrContractNotFound)
// ✅ Just log the error once with context
log.Error("GetByID failed", "contract_id", id, "error", err)
```
