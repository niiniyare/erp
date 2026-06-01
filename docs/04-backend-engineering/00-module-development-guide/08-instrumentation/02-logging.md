---
title: Logging
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Instrumentation Overview](01-instrumentation-overview.md)"
  - "[Structured Logging Guide](../19-observability/02-structured-logging-guide.md)"
  - "[Tracing](03-tracing.md)"
---

# Logging

AwoERP uses Zerolog for structured logging. Every log event is JSON with consistent field names. Never use `fmt.Printf`, `log.Println`, or bare `fmt.Errorf` for observability.

## Log Levels

| Level | When to use |
|-------|------------|
| `TRACE` | Very high-volume debug info (query parameters, loop iterations) — disabled in production |
| `DEBUG` | Method entry with parameters — disabled in production |
| `INFO` | Significant business events (contract created, contract approved) |
| `WARN` | Recoverable problems (notification failed, non-critical service unavailable) |
| `ERROR` | Operation failed; client received an error response |
| `FATAL` | Unrecoverable startup failure — calls `os.Exit` |

## Standard Fields

Every log event should include these fields where relevant:

```go
s.logger.Error().
	Err(err).                               // error object
	Str("method", "ContractService.Create"). // which method
	Str("tenant_id", tenantID.String()).     // tenant isolation context
	Str("contract_id", contractID.String()). // resource being operated on
	Str("actor_id", actorID.String()).       // who triggered the operation
	Msg("create contract failed")            // short human-readable message
```

## Logging Patterns by Operation Type

### Successful Write

```go
s.logger.Info().
	Str("contract_id", contract.ID.String()).
	Str("tenant_id", contract.TenantID.String()).
	Str("contract_number", contract.ContractNumber).
	Str("status", string(contract.Status)).
	Msg("contract created")
```

### Failed Write

```go
s.logger.Error().
	Err(err).
	Str("tenant_id", tenantID.String()).
	Str("contract_number", req.ContractNumber).
	Msg("create contract failed")
```

### Permission Denied (INFO, not ERROR — expected event)

```go
s.logger.Info().
	Str("subject", principal.Subject).
	Str("domain", principal.Domain).
	Str("action", "contracts.contract.create").
	Msg("contract create denied: insufficient permissions")
```

### State Machine Violation (WARN)

```go
s.logger.Warn().
	Str("contract_id", id.String()).
	Str("from_status", string(current.Status)).
	Str("to_status", string(domain.ContractStatusSubmitted)).
	Msg("invalid contract status transition attempted")
```

### Infrastructure Warning (WARN — non-fatal)

```go
s.logger.Warn().
	Err(err).
	Str("type", "contract.submitted").
	Msg("notification send failed; continuing")
```

## Do Not Log

| Do not log | Reason |
|-----------|--------|
| Passwords, tokens, API keys | Security |
| PII (email, phone, national ID) | Privacy / GDPR |
| Contract content fields (title, description) | Privacy — may contain sensitive business info |
| Full request bodies | Size + privacy |
| Stack traces at INFO or above | Verbose; traces are in span |

Log stack traces only at DEBUG level during development.

## Logger Injection

```go
// Constructor
func NewContractService(
	// ...
	logger logger.Logger,
) ContractService {
	return &contractService{
		// ...
		logger: logger.With().Str("service", "contracts").Logger(),
	}
}
```

Using `.With().Str("service", "contracts").Logger()` creates a child logger with a permanent `service` field. All log events from this service automatically include `"service": "contracts"` without adding it to every line.
