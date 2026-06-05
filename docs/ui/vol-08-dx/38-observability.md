---
title: "38 – Observability"
volume: "vol-08-dx"
chapter: 38
section: "Developer Experience"
status: "implemented"
---

# Chapter 38 – Observability

The UI pipeline exposes OpenTelemetry traces and Prometheus-compatible metrics
for every schema generation request.  All instrumentation is wired automatically
when the pipeline is constructed via `NewUIPipeline`.

## Table of Contents
- [38.1 InstrumentedStage](#381-instrumentedstage)
- [38.2 OTel Span Per Stage](#382-otel-span-per-stage)
- [38.3 The Six Metrics](#383-the-six-metrics)
- [38.4 RegisterUIMetrics](#384-registeruimetrics)
- [38.5 Slow Stage Warnings](#385-slow-stage-warnings)
- [38.6 Dashboard Signals](#386-dashboard-signals)

---

## 38.1 InstrumentedStage

`InstrumentedStage` is a decorator that wraps any `Stage` implementation.  It
adds tracing and metrics around the wrapped stage's `Execute` method without
modifying the stage logic itself.

```go
// stages.Instrument wraps s with OTel + metrics + structured logging.
// Called for every stage in NewUIPipeline:
func Instrument(s Stage, tracer trace.Tracer, mp metric.MeterProvider, log *slog.Logger) Stage {
    return &InstrumentedStage{
        inner:  s,
        tracer: tracer,
        mp:     mp,
        log:    log,
    }
}
```

`NewUIPipeline` calls `stages.Instrument` on every stage before adding it to
the pipeline chain, so no stage is ever unobserved.

```go
func NewUIPipeline(tracer trace.Tracer, mp metric.MeterProvider, log *slog.Logger, ...) *Pipeline {
    RegisterUIMetrics(mp) // pre-register histograms

    pipeline := &Pipeline{}
    for _, s := range rawStages {
        pipeline.Add(stages.Instrument(s, tracer, mp, log))
    }
    return pipeline
}
```

---

## 38.2 OTel Span Per Stage

Each call to `InstrumentedStage.Execute` opens an OTel child span named after
the stage.  The span carries the `UIStageAttributes` attribute set:

```go
type UIStageAttributes struct {
    Stage          string // stage name, e.g. "CompileStage"
    Route          string // surface_id / URL route, e.g. "/finance/dashboard"
    TenantID       string // tenant UUID
    OperationKey   string // compiled page key
    CacheHit       bool   // true if schema served from cache
    ASTCompiled    bool   // true if ASTPageFn path was used
}
```

Span attributes emitted:
- `ui.stage` — stage name
- `surface_id` — route (used as surface identifier in dashboards)
- `tenant_id` — tenant UUID
- `cache_hit` — bool
- `error` — set on span if `Execute` returns a non-nil error

The spans nest under the parent request span created by `SchemaHandler`, giving
a full waterfall of stage timings in any OTel-compatible backend (Tempo, Jaeger,
Honeycomb).

---

## 38.3 The Six Metrics

`RegisterUIMetrics` pre-registers six metrics.  All metric names are constants
in the `stages` package:

| Constant | Metric Name | Type | Description |
|---|---|---|---|
| `MetricCompileDuration` | `ui_compile_duration_ms` | Histogram | Time in ms to compile a page (CompileStage only) |
| `MetricStageExecutionDuration` | `ui_stage_execution_duration_ms` | Histogram | Time in ms per stage execution; label: `stage` |
| `MetricSchemaValidationFailures` | `ui_schema_validation_failures_total` | Counter | ValidateStage rejections; labels: `route`, `tenant_id` |
| `MetricCacheGenerationMismatch` | `ui_cache_generation_mismatch_total` | Counter | Cache entries rejected due to version mismatch |
| `MetricInvalidationEvents` | `ui_invalidation_events_total` | Counter | Cache invalidation events fired (permission/flag changes) |
| `MetricRegistryResolutionFailures` | `ui_registry_resolution_failures_total` | Counter | Registry lookup failures (unknown route) |

### Histogram Buckets

`RegisterUIMetrics` pre-registers the histograms with buckets tuned for UI
response times:

```
[1, 5, 10, 25, 50, 100, 250, 500, 1000] ms
```

These buckets surface the P50/P95/P99 compile latency distribution at useful
resolution for the 1–500 ms range.

---

## 38.4 RegisterUIMetrics

`RegisterUIMetrics(mp metric.MeterProvider)` must be called exactly once before
any stage executes.  `NewUIPipeline` calls it automatically — **do not call it
again** in application startup code.

```go
func RegisterUIMetrics(mp metric.MeterProvider) {
    meter := mp.Meter("awoerp.ui")

    compileDurationHist, _ = meter.Float64Histogram(
        MetricCompileDuration,
        metric.WithUnit("ms"),
        metric.WithExplicitBucketBoundaries(1, 5, 10, 25, 50, 100, 250, 500, 1000),
    )
    stageExecHist, _ = meter.Float64Histogram(
        MetricStageExecutionDuration,
        metric.WithUnit("ms"),
        metric.WithExplicitBucketBoundaries(1, 5, 10, 25, 50, 100, 250, 500, 1000),
    )
    // ... counters registered similarly
}
```

If `NewUIPipeline` is not used (e.g. in unit tests), call `RegisterUIMetrics`
with a no-op `MeterProvider` to avoid nil-pointer panics:

```go
RegisterUIMetrics(metric.NewNoopMeterProvider())
```

---

## 38.5 Slow Stage Warnings

`InstrumentedStage` logs a structured warning for any stage that exceeds the
slow-stage threshold of **50 ms**:

```go
const slowStageThresholdMs = 50

if elapsed.Milliseconds() > slowStageThresholdMs {
    log.WarnContext(ctx, "slow UI pipeline stage",
        slog.String("stage",     s.inner.Name()),
        slog.String("route",     attrs.Route),
        slog.String("tenant_id", attrs.TenantID),
        slog.Int64("duration_ms", elapsed.Milliseconds()),
    )
}
```

A slow stage warning does **not** abort the request — it is a signal for
investigation.  Repeated slow-stage warnings on `CompileStage` indicate that
the cache is not being hit (check invalidation logic or permission fingerprint
churn).

---

## 38.6 Dashboard Signals

The following metric combinations are the most useful for an operational
dashboard:

| Signal | Query (PromQL sketch) | Alert threshold |
|---|---|---|
| Cache hit rate | `1 - rate(ui_cache_generation_mismatch_total[5m]) / rate(ui_stage_execution_duration_ms_count{stage="CacheStage"}[5m])` | < 80% → investigate invalidation churn |
| P95 compile duration | `histogram_quantile(0.95, rate(ui_compile_duration_ms_bucket[5m]))` | > 200 ms → look for large page functions |
| Validation failure rate | `rate(ui_schema_validation_failures_total[5m])` | > 0 sustained → schema bug in a recent deploy |
| Registry resolution failures | `increase(ui_registry_resolution_failures_total[1m])` | > 0 → broken route registration |
| Invalidation storm | `rate(ui_invalidation_events_total[1m])` | Spike → permission or flag update loop |

Traces complement these metrics: when P95 compile duration spikes, open a trace
waterfall to identify which stage is the bottleneck.
