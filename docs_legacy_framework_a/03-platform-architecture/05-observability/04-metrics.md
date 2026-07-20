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
portal: 3 — Platform Architecture
section: 05-observability
audience: [architect, backend-engineer, sre]
related:
  - "[Observability Overview](01-observability-overview.md)"
  - "[Tracing](03-tracing.md)"
  - "[Metrics Guide](../../04-backend-engineering/00-module-development-guide/08-instrumentation/04-metrics.md)"
---

# Metrics

## Prometheus Setup

```go
// internal/platform/metrics/metrics.go
var (
    httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests by method, path, and status",
    }, []string{"method", "path", "status"})

    httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "path"})
)
```

## Module Metrics Interface

```go
// internal/platform/metrics/interface.go
type MetricsProvider interface {
    IncrementCounter(name string, labels ...string)
    ObserveHistogram(name string, value float64, labels ...string)
    SetGauge(name string, value float64, labels ...string)
    StartTimer(name string, labels ...string) Timer
}

type Timer interface {
    ObserveDuration()
}
```

## Module Registration

Each module registers its metrics at init time:

```go
// internal/core/contracts/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    ContractCreateTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "contract_create_total",
        Help: "Total contracts created",
    }, []string{"tenant_id", "contract_type"})

    ContractRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "contract_request_duration_seconds",
        Help:    "Contract service operation duration",
        Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5},
    }, []string{"operation"})

    ContractActiveGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "contract_active_count",
        Help: "Active contracts per tenant",
    }, []string{"tenant_id"})
)
```

## Scrape Endpoint

Prometheus scrapes `/metrics` on port 9090 (separate from HTTP server):

```go
func startMetricsServer(port int) {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
```

## Key Metrics to Alert On

| Metric | Alert condition |
|--------|---------------|
| `http_requests_total{status="5xx"}` | > 1% of requests |
| `http_request_duration_seconds{p99}` | > 2 seconds |
| `contract_create_total` error rate | > 5% |
| `event_outbox_pending_count` | > 1000 (relay stuck) |
| `db_pool_acquire_duration_seconds{p99}` | > 500ms (pool exhaustion) |
| `temporal_workflow_failures_total` | > 0 for critical workflows |
