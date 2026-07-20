> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "39 – Telemetry"
volume: "vol-08-dx"
chapter: 39
section: "Developer Experience"
status: "implemented"
related: "[38 – Observability](38-observability.md)"
---

# Chapter 39 – Telemetry

Chapter 38 covers the metrics and tracing instrumentation wired into
`InstrumentedStage`.  This chapter documents the structured log fields emitted
throughout the UI pipeline — the third pillar of observability.

## Table of Contents
- [39.1 Tenant ID in Every Log Entry](#391-tenant-id-in-every-log-entry)
- [39.2 Standard Log Fields](#392-standard-log-fields)
- [39.3 Correlating Logs With Traces](#393-correlating-logs-with-traces)
- [39.4 Log Sampling Guidance](#394-log-sampling-guidance)

---

## 39.1 Tenant ID in Every Log Entry

The UI pipeline is multi-tenant by construction.  Every log entry emitted during
a pipeline execution includes `tenant_id` as a structured field so that log
queries can be scoped to a single tenant.

This is enforced by threading `UISessionContext` (which carries `TenantID`)
through every stage's `Execute(ctx, pipelineCtx)` call.  Stages extract it when
building log records:

```go
log.InfoContext(ctx, "stage complete",
    slog.String("stage",     s.Name()),
    slog.String("tenant_id", pipelineCtx.Session.TenantID),
    slog.String("route",     pipelineCtx.Route),
    slog.Int64("duration_ms", elapsed.Milliseconds()),
)
```

---

## 39.2 Standard Log Fields

The table below lists all structured fields emitted across the pipeline.
Fields marked **every entry** appear on all log lines from `InstrumentedStage`.

| Field | Type | Present | Description |
|---|---|---|---|
| `tenant_id` | string | every entry | Tenant UUID from `UISessionContext` |
| `route` | string | every entry | URL route / surface ID |
| `stage` | string | every entry | Name of the executing stage |
| `duration_ms` | int64 | every entry | Wall-clock time for `Execute` |
| `cache_hit` | bool | CacheStage | Whether the schema was served from cache |
| `operation_key` | string | CompileStage | Compiled page cache key |
| `ast_compiled` | bool | CompileStage | True when ASTPageFn path was taken |
| `error` | string | on error | Error message (never stack trace) |
| `user_id` | string | AuthStage | Authenticated user UUID |
| `perm_fingerprint` | string | CacheStage | Permission fingerprint used for cache lookup |

### Slow stage entry

When a stage exceeds 50 ms (§38.5), an additional `"slow UI pipeline stage"`
warning is emitted with `duration_ms` and `stage` fields — see §40.2 for log
level guidance.

---

## 39.3 Correlating Logs With Traces

`InstrumentedStage` injects the active OTel trace context into every log record
using `slog`'s context-aware methods (`log.InfoContext`, `log.WarnContext`,
`log.ErrorContext`).  When the logging backend supports OTel correlation (e.g.
OpenObserve, Loki with Tempo, Datadog), the `trace_id` and `span_id` fields are
automatically extracted from the context and appended to log records.

This means a slow-stage warning in the log can be directly linked to the
matching OTel span in the trace backend — no manual correlation required.

**Ensure the logger is configured with an OTel log bridge** in application
startup:

```go
// internal/shared/logger/slog.go — OTel log bridge configuration
handler := otelslog.NewHandler("awoerp.ui", otelslog.WithLoggerProvider(lp))
logger  := slog.New(handler)
```

With this bridge active, `trace_id` appears automatically on every log line
emitted inside a traced request.

---

## 39.4 Log Sampling Guidance

Not every successful schema generation needs a log entry in production.  The
recommended approach:

| Scenario | Log level | Sampling |
|---|---|---|
| Successful cache hit | `Debug` | Sample at 1 % in production |
| Successful compile (cache miss) | `Info` | Log every occurrence |
| Slow stage (> 50 ms) | `Warn` | Log every occurrence |
| Validation failure | `Warn` | Log every occurrence |
| Pipeline error | `Error` | Log every occurrence, alert |
| Registry resolution failure | `Error` | Log every occurrence, alert |

Configure the application logger's minimum level to `Info` in production.
`Debug` entries (cache hits) are emitted only when the logger minimum level is
`Debug`, which should be reserved for local development or targeted
troubleshooting sessions.
