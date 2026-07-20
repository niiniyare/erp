> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Environment Variables Reference
portal: 6 — DevOps
section: 06-devops
audience: [devops, sre, backend-engineer]
related:
  - "[DevOps Overview](01-devops-overview.md)"
  - "[Server Startup](../04-backend-engineering/00-module-development-guide/21-server-startup/01-server-startup-overview.md)"
---

# Environment Variables Reference

All configuration is via environment variables. No config files are read at runtime.

## Required Variables

The server will not start if these are missing.

| Variable | Example | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://user:pass@host:5432/awoerp?sslmode=require` | PostgreSQL connection string (pgxpool DSN) |
| `REDIS_URL` | `redis://:{password}@host:6379/0` | Redis connection string |
| `SESSION_SECRET` | 64-char random hex | Used for session token signing |
| `TEMPORAL_HOST` | `temporal-frontend:7233` | Temporal gRPC endpoint |
| `TEMPORAL_NAMESPACE` | `awoerp-production` | Temporal namespace |

## Optional / Defaulted

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | `8080` | API server listen port |
| `HEALTH_PORT` | `8081` | Health check server port |
| `METRICS_PORT` | `9090` | Prometheus metrics port |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` (production), `text` (development) |
| `DATABASE_MAX_CONNS` | `20` | pgxpool max open connections |
| `DATABASE_MIN_CONNS` | `5` | pgxpool min idle connections |
| `DATABASE_CONN_LIFETIME` | `1h` | Max connection lifetime |
| `OUTBOX_RELAY_BATCH_SIZE` | `100` | Events per relay cycle |
| `OUTBOX_RELAY_INTERVAL` | `1s` | Poll interval for outbox relay |
| `SESSION_TTL` | `86400` | Session TTL in seconds (24h) |
| `RATE_LIMIT_AUTH` | `10` | Auth endpoint rate limit per minute per IP |
| `ENV` | `production` | `development`, `staging`, `production` — affects log format, panic behavior |

## Feature Flags

Feature flags are per-tenant (stored in DB), not environment variables. Use the Tenant API to enable/disable features.

## Temporal Worker Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TEMPORAL_MAX_CONCURRENT_WORKFLOW_TASKS` | `10` | Concurrent workflow task pollers per worker |
| `TEMPORAL_MAX_CONCURRENT_ACTIVITIES` | `50` | Concurrent activity executions per worker |
| `WORKER_TASK_QUEUE_PREFIX` | `awoerp` | Task queue name prefix (e.g., `awoerp.contracts`) |

## Observability

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | — | OTLP gRPC endpoint for traces (e.g., `http://otel-collector:4317`) |
| `OTEL_SERVICE_NAME` | `awoerp` | Service name in traces |
| `OTEL_TRACES_SAMPLER` | `parentbased_traceidratio` | Sampling strategy |
| `OTEL_TRACES_SAMPLER_ARG` | `0.1` | Sampling ratio (10%) |

## Secrets Management

In production, secrets are injected as environment variables from:
- **Kubernetes Secrets** — mounted as env vars in the pod spec
- **External Secrets Operator** — syncs from AWS Secrets Manager / Vault into Kubernetes Secrets

Never put secrets in:
- Container images
- Kubernetes ConfigMaps (not encrypted)
- Helm values files checked into git

### Example Kubernetes Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: awoerp-secrets
  namespace: production
type: Opaque
stringData:
  DATABASE_URL: "postgres://..."
  REDIS_URL: "redis://..."
  SESSION_SECRET: "..."
```

### Example Pod envFrom

```yaml
spec:
  containers:
  - name: awoerp-server
    envFrom:
    - secretRef:
        name: awoerp-secrets
    - configMapRef:
        name: awoerp-config
```

## Local Development (.env)

For local development, use a `.env` file (not committed):

```bash
# .env (gitignored)
DATABASE_URL=postgres://postgres:postgres@localhost:5432/awoerp?sslmode=disable
REDIS_URL=redis://localhost:6379/0
TEMPORAL_HOST=localhost:7233
TEMPORAL_NAMESPACE=default
SESSION_SECRET=dev-secret-not-for-production-use
ENV=development
LOG_FORMAT=text
LOG_LEVEL=debug
```

Load with `go run ./cmd/server` + `godotenv` or the project's `make run` target.
