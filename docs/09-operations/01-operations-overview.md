---
title: Operations Overview
portal: 9 — Operations
section: 09-operations
audience: [sre, devops, backend-engineer]
related:
  - "[Service Down Runbook](02-service-down.md)"
  - "[Database Issues Runbook](03-database-issues.md)"
  - "[Migration Failure Runbook](05-migration-failure.md)"
  - "[Kubernetes Deployment](../06-devops/03-kubernetes-deployment.md)"
---

# Operations Overview

This portal contains runbooks for diagnosing and resolving production incidents on AwoERP.

## Runbook Index

| Runbook | When to Use |
|---------|-------------|
| [Service Down](02-service-down.md) | Pod crashes, 5xx errors, service unreachable |
| [Database Issues](03-database-issues.md) | Slow queries, pool exhaustion, replication lag |
| [High Error Rate](04-high-error-rate.md) | Error rate spike on any endpoint |
| [Migration Failure](05-migration-failure.md) | Failed deployment due to DB migration error |
| [Temporal Worker Down](06-temporal-worker.md) | Workflows stuck, workers not polling |
| [Event Outbox Stuck](07-event-outbox.md) | Events not being delivered, outbox backlog |

## Incident Severity Classification

| Level | Definition | Response SLA |
|-------|-----------|--------------|
| P0 — Critical | All tenants down or data loss risk | 15 minutes |
| P1 — High | Single tenant down or major feature broken | 1 hour |
| P2 — Medium | Degraded performance or non-critical feature broken | 4 hours |
| P3 — Low | Minor issue, cosmetic, or workaround available | 1 business day |

## On-Call Responsibilities

1. Acknowledge incident within 5 minutes of alert
2. Open incident channel: `#incidents`
3. Post status update every 15 minutes for P0/P1
4. Engage second responder for P0 after 15 minutes without resolution
5. Write post-mortem within 48 hours of P0/P1 resolution

## Common First Steps

For any incident:

```bash
# Check pod status
kubectl get pods -n awoerp

# Check recent events
kubectl get events -n awoerp --sort-by='.lastTimestamp' | tail -20

# Check application logs (last 100 lines)
kubectl logs -n awoerp deployment/awoerp-api --tail=100

# Check error rate in last 5 minutes
kubectl exec -n awoerp deployment/awoerp-api -- \
  curl -s localhost:9090/metrics | grep http_requests_total
```

## Health Check Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /health/live` | Liveness probe — process up |
| `GET /health/ready` | Readiness probe — DB + Redis connected |
| `GET /health/startup` | Startup probe — migrations applied |

A failing `/health/ready` while `/health/live` passes usually indicates DB/Redis connectivity issues.

## Database Connection

```bash
# Check DB pool status
kubectl exec -n awoerp deployment/awoerp-api -- \
  curl -s localhost:9090/metrics | grep db_pool

# Test DB connectivity
kubectl exec -n awoerp deployment/postgres-0 -- \
  psql -U awoerp -c "SELECT 1"
```

## Key Metrics

| Metric | Alert Threshold |
|--------|----------------|
| `http_request_duration_p99` | > 2s |
| `http_error_rate_5xx` | > 1% |
| `db_pool_idle` | 0 (pool exhaustion) |
| `temporal_workflow_backlog` | > 100 |
| `event_outbox_pending` | > 500 |
