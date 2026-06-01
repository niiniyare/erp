---
title: Logging
portal: 3 — Platform Architecture
section: 05-observability
audience: [architect, backend-engineer, sre]
related:
  - "[Observability Overview](01-observability-overview.md)"
  - "[Tracing](03-tracing.md)"
  - "[Logging Guide](../../04-backend-engineering/00-module-development-guide/08-instrumentation/02-logging.md)"
---

# Logging

## Logger Initialization

```go
// cmd/server/main.go
log.Logger = zerolog.New(os.Stdout).
    With().
    Timestamp().
    Str("service", "awoerp").
    Str("version", buildVersion).
    Logger()

// Respect LOG_LEVEL env var
level, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
if err != nil {
    level = zerolog.InfoLevel
}
zerolog.SetGlobalLevel(level)
```

## Module Child Logger

Each service creates a child logger scoped to its module:

```go
func NewContractService(log zerolog.Logger, ...) ContractService {
    return &contractService{
        log: log.With().Str("module", "contracts").Logger(),
    }
}
```

Handler child logger:
```go
log.With().Str("module", "contracts").Str("layer", "handler").Logger()
```

## Standard Log Fields

Every log entry in a request handler context should include:

```go
s.log.Info().
    Str("tenant_id", tenantID.String()).
    Str("user_id", userID.String()).
    Str("request_id", requestID).
    Str("trace_id", span.SpanContext().TraceID().String()).
    Msg("contract created")
```

## Request Log Format

The Logger middleware emits one structured log per request:

```json
{
  "level": "info",
  "service": "awoerp",
  "method": "POST",
  "path": "/api/v1/contracts",
  "status": 201,
  "latency_ms": 23,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "remote_ip": "10.0.0.1",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

## Error Log Format

```json
{
  "level": "error",
  "service": "awoerp",
  "module": "contracts",
  "error": "contract not found",
  "contract_id": "...",
  "tenant_id": "...",
  "trace_id": "...",
  "timestamp": "2025-01-15T10:30:00Z",
  "message": "GetByID failed"
}
```

## What NOT to Log

- Passwords, tokens, API keys — ever
- Full request/response bodies (may contain PII)
- SQL query text with parameters (parameterized queries only)
- User PII (email, phone, name) in non-PII log streams

## Log Aggregation

Logs go to stdout. The container runtime forwards to the log aggregator (Loki, Elasticsearch, CloudWatch). No file-based logging.

## Sampling

High-throughput paths (list endpoints) log at DEBUG by default. Enable per-tenant DEBUG logging via feature flag `observability.debug_logging` for targeted troubleshooting without noise from other tenants.
