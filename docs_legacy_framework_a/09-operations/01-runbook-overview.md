> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Operations Runbook Overview
portal: 9 — Operations
section: 09-operations
audience: [devops, sre]
related:
  - "[Health Checks](../03-platform-architecture/05-observability/01-observability-overview.md)"
  - "[DevOps Overview](../06-devops/01-devops-overview.md)"
  - "[Metrics](../03-platform-architecture/05-observability/04-metrics.md)"
---

# Operations Runbook Overview

## Service Endpoints

| Endpoint | Port | Purpose |
|----------|------|---------|
| HTTP API | 8080 | All API traffic |
| Health probes | 8081 | Kubernetes liveness/readiness |
| Metrics | 9090 | Prometheus scrape |
| Temporal UI | 8088 | Workflow monitoring (internal) |

## Runbooks Index

| Runbook | When to use |
|---------|------------|
| [Service Down](02-service-down.md) | HTTP server not responding |
| [Database Issues](03-database-issues.md) | DB connection errors, slow queries |
| [High Error Rate](04-high-error-rate.md) | 5xx rate above threshold |
| [Migration Failure](05-migration-failure.md) | Migration job fails on deploy |
| [Temporal Worker Stuck](06-temporal-worker.md) | Workflows not progressing |
| [Event Outbox Backlog](07-event-outbox.md) | Outbox relay behind |

## Common Diagnostic Commands

### Check server health

```bash
curl http://{host}:8081/health/live
curl http://{host}:8081/health/ready
```

### View recent logs

```bash
kubectl logs -l app=awoerp-server --tail=100 -n production
kubectl logs -l app=awoerp-server --tail=100 -n production | grep '"level":"error"'
```

### Check database connections

```bash
# Pool status
kubectl exec -it deploy/awoerp-server -- curl localhost:9090/metrics | grep db_pool
```

### Check event outbox backlog

```sql
SELECT COUNT(*) FROM event_outbox WHERE delivered_at IS NULL;
SELECT topic, COUNT(*), MIN(created_at) AS oldest
FROM event_outbox
WHERE delivered_at IS NULL
GROUP BY topic
ORDER BY oldest;
```

### Replay dead-letter events

```bash
awoctl events replay --dlq contracts.submitted --from 2025-01-15
```

### Force session revoke (security incident)

```bash
# Revoke all sessions for a user
awoctl sessions revoke-all --user-id {uuid}

# Revoke all sessions for a tenant
awoctl sessions revoke-tenant --tenant-id {uuid}
```

## Alerting Thresholds

| Alert | Threshold | Severity |
|-------|-----------|---------|
| 5xx error rate | > 1% over 5 min | P2 |
| p99 latency | > 2s over 5 min | P2 |
| DB pool exhaustion | < 2 idle connections | P1 |
| Readiness probe failing | > 2 consecutive | P1 |
| Event outbox backlog | > 1000 undelivered | P3 |
| Temporal workflow failures | > 0 for critical workflows | P2 |

## Escalation Path

1. On-call SRE (PagerDuty rotation)
2. Backend lead
3. Database admin (for data issues)
4. CTO (for data breach or extended P1)
