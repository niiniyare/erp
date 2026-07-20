> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Metrics
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Instrumentation Overview](01-instrumentation-overview.md)"
  - "[Tracing](03-tracing.md)"
  - "[Metrics Architecture](../../../03-platform-architecture/05-observability/04-metrics.md)"
---

# Metrics

Prometheus metrics track aggregate behaviour: how many contracts are created per minute, how often approval fails, what the 95th percentile latency is for list queries.

## Metric Types

| Type | Use for |
|------|---------|
| Counter | Events that only increase: creates, errors, approvals |
| Gauge | Current state: active contracts, pending approvals |
| Histogram | Distribution of values: request latency, contract values |

## Standard Metrics for Every Module

```go
// Increment counter on each create
s.metrics.IncrementCounter("contracts.created_total", 1)

// Increment counter on each status transition
s.metrics.IncrementCounter("contracts.submitted_total", 1)
s.metrics.IncrementCounter("contracts.approved_total", 1)
s.metrics.IncrementCounter("contracts.terminated_total", 1)

// Increment on errors
s.metrics.IncrementCounter("contracts.errors_total", 1)

// Observe request duration (called at start and end of method)
start := time.Now()
defer s.metrics.ObserveDuration("contracts.create_duration_seconds", time.Since(start))
```

## Metric Naming Convention

Format: `<module>.<operation>_<unit>` — lowercase, underscores, Prometheus-style.

| Metric | Type | Description |
|--------|------|-------------|
| `contracts.created_total` | Counter | Contracts created |
| `contracts.updated_total` | Counter | Contracts updated |
| `contracts.submitted_total` | Counter | Contracts submitted for review |
| `contracts.approved_total` | Counter | Contracts approved |
| `contracts.activated_total` | Counter | Contracts activated |
| `contracts.terminated_total` | Counter | Contracts terminated |
| `contracts.deleted_total` | Counter | Contracts soft-deleted |
| `contracts.errors_total` | Counter | Service errors |
| `contracts.list_duration_seconds` | Histogram | List query duration |
| `contracts.create_duration_seconds` | Histogram | Create operation duration |

## MetricsProvider Interface

```go
// awo.so/internal/shared/metrics
type MetricsProvider interface {
	IncrementCounter(name string, value float64)
	ObserveHistogram(name string, value float64)
	SetGauge(name string, value float64)
	// ... other methods
}
```

All calls are safe to ignore errors — the interface implementations never return errors. If the metrics backend is unavailable, calls are silently dropped.

## Latency Histogram Pattern

```go
func (s *contractService) List(ctx context.Context, ...) (*ListContractsResult, error) {
	start := time.Now()
	defer func() {
		s.metrics.ObserveHistogram(
			"contracts.list_duration_seconds",
			time.Since(start).Seconds(),
		)
	}()

	// ... method body ...
}
```

The deferred observation fires even when the method returns an error, so the histogram includes both successful and failed requests. This accurately represents true latency distribution.

## Labels / Tags

Some metrics providers support labels/tags for filtering. If the provider supports them:

```go
s.metrics.IncrementCounterWithLabels("contracts.status_transitions_total", 1, map[string]string{
	"from_status": string(current.Status),
	"to_status":   string(newStatus),
})
```

This enables breakdown by transition type in dashboards (e.g., "how many draft → submitted vs approved → active transitions happened today?").
