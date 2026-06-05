---
title: "40 – Logging"
volume: "vol-08-dx"
chapter: 40
section: "Developer Experience"
status: "implemented"
related: "[39 – Telemetry](39-telemetry.md)"
---

# Chapter 40 – Logging

This chapter documents the specific log fields, log levels, and logging patterns
used in the UI pipeline's HTTP boundary — `SchemaHandler` — and how those
conventions extend into DSL code.

## Table of Contents
- [40.1 SchemaHandler Log Fields](#401-schemahandler-log-fields)
- [40.2 Log Levels and When to Use Them](#402-log-levels-and-when-to-use-them)
- [40.3 The log.ErrorContext / log.WarnContext Pattern](#403-the-logerrorcontext--logwarncontext-pattern)
- [40.4 What DSL Code Must Not Log](#404-what-dsl-code-must-not-log)

---

## 40.1 SchemaHandler Log Fields

`SchemaHandler` is the HTTP handler that invokes the pipeline and serialises the
resulting schema to JSON.  It is the outermost observability boundary for a
schema request.

Fields emitted on every request:

| Field | Value |
|---|---|
| `route` | URL path matched by the router, e.g. `/finance/dashboard` |
| `tenant_id` | Tenant UUID extracted from the session token |
| `user_id` | Authenticated user UUID |
| `status` | HTTP status code returned |
| `duration_ms` | Total handler time (includes full pipeline) |

Fields emitted on error:

| Field | Value |
|---|---|
| `error` | Error message string (never a raw stack trace) |
| `stage` | Name of the pipeline stage that failed (if determinable) |

Example log record (JSON output):

```json
{
  "time": "2026-06-05T10:23:11Z",
  "level": "ERROR",
  "msg": "schema pipeline failed",
  "route": "/finance/invoices",
  "tenant_id": "t_9f3a2c",
  "user_id": "u_7b1d44",
  "status": 500,
  "stage": "ValidateStage",
  "error": "schema validation failed: chart node missing style",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7"
}
```

---

## 40.2 Log Levels and When to Use Them

The UI pipeline uses four log levels with the following semantics:

### Error — pipeline failures

Use `log.ErrorContext` when the pipeline cannot produce a schema and the request
must return HTTP 500.  Examples:
- `ValidateStage` rejects the schema
- `CompileStage` catches a panic in a PageFn
- `AuthStage` cannot validate the session token (as opposed to 401 — that is a
  Warn)

```go
log.ErrorContext(ctx, "schema pipeline failed",
    slog.String("route",     route),
    slog.String("tenant_id", tenantID),
    slog.String("error",     err.Error()),
    slog.String("stage",     failedStage),
)
```

### Warn — expected negative outcomes

Use `log.WarnContext` for outcomes that are correct but worth noting:
- Route not found in the registry (HTTP 404)
- Session has expired (HTTP 401)
- Schema serves from cache but cache generation mismatched (invalidated, recompiled)
- A stage took longer than 50 ms (§38.5)

```go
log.WarnContext(ctx, "page not found in registry",
    slog.String("route",     route),
    slog.String("tenant_id", tenantID),
)
```

### Info — normal pipeline completion (cache miss)

Use `log.InfoContext` for successful schema compilations where a cache miss
occurred (i.e., the schema was freshly compiled).

### Debug — cache hits

Use `log.DebugContext` for successful schema responses served from cache.
Cache hits are the hot path and would flood logs at Info level in production.

---

## 40.3 The log.ErrorContext / log.WarnContext Pattern

All logging in the pipeline must use the `Context`-aware variants
(`ErrorContext`, `WarnContext`, `InfoContext`, `DebugContext`) so that OTel
trace correlation fields (`trace_id`, `span_id`) are automatically included.

**Correct:**

```go
log.WarnContext(ctx, "page not found in registry",
    slog.String("route",     route),
    slog.String("tenant_id", tenantID),
)
```

**Incorrect (loses trace correlation):**

```go
log.Warn("page not found in registry")            // no context
log.Warn("page not found: " + route)              // unstructured
fmt.Printf("page not found: %s\n", route)         // not a log entry at all
```

Structured fields must always be passed as `slog.String`, `slog.Int64`,
`slog.Bool`, etc. — never interpolated into the message string.

---

## 40.4 What DSL Code Must Not Log

Block functions and screen functions **must not** call `log` directly.  DSL code
is pure: it receives `UISessionContext`, computes schema nodes, and returns them.
Side effects — including logging — belong in pipeline stages.

If a block function encounters an unexpected condition, it should return a
degraded-but-valid schema (e.g. an empty panel) and let `ValidateStage` surface
the issue.  Structural errors in DSL code are caught by `CompileStage`'s
panic-recovery and reported through the standard pipeline error path.

```go
// Wrong — DSL code should not log
func InvoiceSummaryBlock(sess ui.UISessionContext) ast.Node {
    log.InfoContext(context.Background(), "rendering invoice summary") // NO
    return ast.PanelNode{...}
}

// Correct — pure function, no side effects
func InvoiceSummaryBlock(sess ui.UISessionContext) ast.Node {
    return ast.PanelNode{
        Title: "Invoice Summary",
        // ...
    }
}
```
