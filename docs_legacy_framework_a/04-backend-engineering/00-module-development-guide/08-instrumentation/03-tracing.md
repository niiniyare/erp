> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Tracing
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Instrumentation Overview](01-instrumentation-overview.md)"
  - "[Metrics](04-metrics.md)"
  - "[Tracing Architecture](../../../03-platform-architecture/05-observability/03-tracing.md)"
---

# Tracing

Distributed tracing tracks a request as it flows through service methods, database calls, and external service calls. Each unit of work is a **span**. Spans form a tree — child spans nest inside parent spans.

## Starting a Span

```go
func (s *contractService) Create(ctx context.Context, req CreateContractRequest) (*domain.Contract, error) {
	// Start a span — ctx carries the parent span from the HTTP middleware
	ctx, span := s.tracer.Start(ctx, "ContractService.Create")
	defer span.End()  // End is always deferred — runs even on error return

	// ...
}
```

The span name convention: `<ServiceName>.<MethodName>`.

## Setting Span Attributes

Attributes are key-value pairs attached to the span. They appear in the trace viewer (Jaeger, Tempo):

```go
// After successful operation
span.SetAttributes(
	attribute.String("contract.id",     contract.ID.String()),
	attribute.String("tenant.id",       contract.TenantID.String()),
	attribute.String("contract.number", contract.ContractNumber),
	attribute.String("contract.status", string(contract.Status)),
)
```

Standard attribute keys:

| Key | When to set |
|-----|------------|
| `contract.id` | After create/update/fetch |
| `tenant.id` | On every span |
| `contract.status` | After status change |
| `contract.number` | On create |
| `db.rows_affected` | After bulk operations |

## Recording Errors

```go
if err != nil {
	span.RecordError(err)                    // attaches error to span
	span.SetStatus(codes.Error, err.Error()) // marks span as failed in the trace
	return nil, err
}
```

`RecordError` creates an error event on the span. `SetStatus` marks the span red in the trace UI. Both should be called when returning an error.

## Passing Context Through

The `ctx` carrying the span must be passed to all downstream calls:

```go
// CORRECT — ctx carries the span
contract, err := s.repo.GetByID(ctx, id, tenantID)  // repo sees parent span

// WRONG — background ctx loses the trace
contract, err := s.repo.GetByID(context.Background(), id, tenantID)
```

The repository's `WithTenant` call also creates a child span for the DB transaction. The trace shows: `ContractService.Create` → `WithTenant` → `CreateContract (sql)`.

## Async Goroutine Tracing

For async goroutines (audit, event publish), create a new span from a **copy** of the parent context:

```go
// Capture the span context from the request context
spanCtx := trace.SpanContextFromContext(ctx)

go func() {
	// Create a new background context with the parent span context linked
	asyncCtx := trace.ContextWithRemoteSpanContext(context.Background(), spanCtx)
	asyncCtx, asyncSpan := s.tracer.Start(asyncCtx, "ContractService.publishContractCreated")
	defer asyncSpan.End()

	if err := s.eventBus.Publish(asyncCtx, evt); err != nil {
		asyncSpan.RecordError(err)
	}
}()
```

This links the async span to the parent trace without using the cancelled request context.

## Import

```go
import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)
```
