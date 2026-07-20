> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Instrumentation Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Logging](02-logging.md)"
  - "[Tracing](03-tracing.md)"
  - "[Metrics](04-metrics.md)"
  - "[Observability Architecture](../../../03-platform-architecture/05-observability/01-observability-overview.md)"
---

# Instrumentation Overview

Every AwoERP service method is instrumented with three pillars: structured logging (Zerolog), distributed tracing (OpenTelemetry), and metrics (Prometheus). Instrumentation never fails the request — it is always wrapped in error-ignoring paths.

## The Three Pillars

| Pillar | Library | Purpose |
|--------|---------|---------|
| Structured logging | Zerolog | Human-readable + machine-parseable event log |
| Distributed tracing | OpenTelemetry | Request flow across service boundaries |
| Metrics | Prometheus | Aggregated counters, histograms, gauges |

## Injected Dependencies

```go
type contractService struct {
	// ...
	logger  logger.Logger           // awo.so/internal/shared/logger
	tracer  tracing.Service         // awo.so/internal/shared/tracing
	metrics metrics.MetricsProvider // awo.so/internal/shared/metrics
}
```

All three are injected via Wire. No global state, no `log.Printf`. Never import `fmt` for logging.

## The Golden Rule: Instrumentation Never Fails the Request

```go
// CORRECT — ignore span errors, never propagate them
ctx, span := s.tracer.Start(ctx, "ContractService.Create")
defer span.End()  // End is safe to call even if Start failed

// CORRECT — log errors but don't return them
if err := s.metrics.IncrementCounter("contracts.created", 1); err != nil {
	// Silently ignore — metric loss is acceptable
}

// WRONG — propagating instrumentation error
ctx, span, err := s.tracer.Start(ctx, "ContractService.Create")
if err != nil {
	return nil, err  // NEVER do this — blocks the operation for a tracing failure
}
```

## Instrumentation Skeleton for Every Service Method

```go
func (s *contractService) <MethodName>(ctx context.Context, ...) (<ReturnType>, error) {
	// 1. Start span — always first line
	ctx, span := s.tracer.Start(ctx, "ContractService.<MethodName>")
	defer span.End()

	// 2. ... business logic ...

	// 3. On error — record in span before returning
	if err != nil {
		span.RecordError(err)
		s.logger.Error().
			Err(err).
			Str("method", "<MethodName>").
			Str("tenant_id", tenantID.String()).
			Msg("<method name> failed")
		return <zero>, err
	}

	// 4. On success — set attributes, increment counter
	span.SetAttributes(
		attribute.String("contract.id", result.ID.String()),
	)
	s.metrics.IncrementCounter("contracts.<method_name>d", 1)

	return result, nil
}
```
