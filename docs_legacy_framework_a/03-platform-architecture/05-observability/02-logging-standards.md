> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Logging Standards
portal: 3 — Platform Architecture
section: 05-observability
audience: [backend-engineer, sre]
related:
  - "[Observability Overview](01-observability-overview.md)"
  - "[Logging Guide](../../04-backend-engineering/00-module-development-guide/08-instrumentation/02-logging.md)"
  - "[Tracing](../../04-backend-engineering/00-module-development-guide/08-instrumentation/03-tracing.md)"
---

# Logging Standards

## Log Format

All logs are structured JSON via `zerolog`. Never use `fmt.Println` or `log.Printf`.

```json
{
  "level": "info",
  "time": "2025-05-31T14:30:00Z",
  "service": "awoerp",
  "trace_id": "abc123def456",
  "span_id": "789xyz",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "request_id": "req-7c9e6679",
  "module": "contracts",
  "message": "contract created"
}
```

## Required Fields

Every log entry must have:

| Field | Source | Description |
|-------|--------|-------------|
| `level` | zerolog | `debug`, `info`, `warn`, `error` |
| `time` | zerolog | RFC3339 UTC timestamp |
| `service` | startup | Always `"awoerp"` |
| `trace_id` | OTel context | Links to distributed trace |
| `request_id` | middleware | Per-request UUID from `X-Request-ID` header |

For tenant-scoped operations, also include:

| Field | Source |
|-------|--------|
| `tenant_id` | Session |
| `module` | Service name constant |

## Log Levels

| Level | Use for |
|-------|---------|
| `debug` | Development tracing — disabled in production by default |
| `info` | Normal business events: record created, state changed |
| `warn` | Recoverable anomaly: retry triggered, fallback used, deprecated path hit |
| `error` | Unrecoverable error that needs attention: DB failure, external service error |

Never log `error` for expected business errors (ErrNotFound, ErrForbidden) — those are `warn` at most, usually omitted entirely.

## Standard Logger Setup

```go
// cmd/server/main.go
log.Logger = zerolog.New(os.Stdout).
    With().
    Timestamp().
    Str("service", "awoerp").
    Logger()

// Set level from env
level, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
if err != nil {
    level = zerolog.InfoLevel
}
zerolog.SetGlobalLevel(level)
```

## Context Logger

Inject trace/request IDs into a context-bound logger, then pass it through the call stack:

```go
// middleware: attach logger to context
func requestLoggerMiddleware(next fiber.Handler) fiber.Handler {
    return func(c *fiber.Ctx) error {
        logger := log.With().
            Str("request_id", c.Locals("request_id").(string)).
            Str("trace_id", traceIDFromContext(c.UserContext())).
            Logger()
        c.Locals("logger", &logger)
        return next(c)
    }
}

// service: extract and use
func loggerFromCtx(ctx context.Context) *zerolog.Logger {
    if l, ok := ctx.Value(loggerKey{}).(*zerolog.Logger); ok {
        return l
    }
    return &log.Logger
}
```

## What to Log in Services

```go
func (s *contractService) Create(ctx context.Context, req CreateParams) (*domain.Contract, error) {
    logger := loggerFromCtx(ctx).With().
        Str("module", "contracts").
        Str("operation", "create").
        Str("tenant_id", req.TenantID.String()).
        Logger()

    contract, err := s.repo.Create(ctx, req)
    if err != nil {
        // Log infrastructure errors
        logger.Error().Err(err).Msg("failed to create contract")
        return nil, err
    }

    // Log business events at info
    logger.Info().
        Str("contract_id", contract.ID.String()).
        Str("contract_number", contract.ContractNumber).
        Msg("contract created")

    return contract, nil
}
```

## What NOT to Log

- **PII**: email, phone, full name, address — use IDs only
- **Secrets**: tokens, passwords, API keys — never log request bodies that may contain these
- **Large blobs**: avoid logging full JSON payloads; log only IDs and counts
- **Every DB query**: use query tracing in development only; SQLC query logs are debug-only

## HTTP Request Log Format

The request middleware logs one line per request on completion:

```json
{
  "level": "info",
  "time": "2025-05-31T14:30:00.123Z",
  "service": "awoerp",
  "request_id": "req-abc123",
  "trace_id": "xyz789",
  "tenant_id": "550e8400-...",
  "method": "POST",
  "path": "/api/v1/contracts",
  "status": 201,
  "latency_ms": 42,
  "message": "request completed"
}
```

Errors include the error message but never a stack trace in production.

## Error Log Fields

```go
logger.Error().
    Err(err).                          // error message
    Str("contract_id", id.String()).   // affected resource
    Int("attempt", attempt).           // retry context if applicable
    Msg("failed to notify reviewer")
```

## Sampling in High-Volume Paths

For endpoints called >1000 req/s (e.g., list endpoints), sample info logs:

```go
// Log 1% of successful list requests
if rand.Float64() < 0.01 {
    logger.Info().Int("result_count", len(results)).Msg("contracts listed")
}
// Always log errors
```

## Log Retention

| Environment | Retention |
|-------------|-----------|
| Development | Local disk, no retention policy |
| Staging | 7 days |
| Production | 90 days (compliance minimum) |

Audit logs (in `audit_log` DB table) are retained indefinitely regardless of log retention.

## Correlation with Traces

Every log entry in a request context includes `trace_id`. Use this to jump from a log line to the full distributed trace in your observability platform (Jaeger/Tempo):

```
trace_id: abc123def456  →  search in Jaeger  →  full request timeline
```

When `trace_id` is absent (background jobs, cron), use `job_id` or `workflow_id` instead.
