---
title: High Error Rate Runbook
portal: 9 — Operations
section: 09-operations
audience: [devops, sre]
related:
  - "[Operations Overview](01-operations-overview.md)"
  - "[Service Down](02-service-down.md)"
  - "[Database Issues](03-database-issues.md)"
---

# High Error Rate Runbook

**Trigger**: 5xx error rate > 1% over 5 minutes. PagerDuty alert: `awoerp-high-error-rate`.

## 1. Quantify the Problem

```bash
# Count errors in last 15 minutes
kubectl logs -l app=awoerp-server -n production --since=15m \
  | grep '"level":"error"' | wc -l

# Get error breakdown by path
kubectl logs -l app=awoerp-server -n production --since=15m \
  | grep '"status":5' \
  | jq -r '"\(.status) \(.path)"' \
  | sort | uniq -c | sort -rn | head -20
```

## 2. Identify Error Type

```bash
# Show full error messages
kubectl logs -l app=awoerp-server -n production --since=15m \
  | grep '"level":"error"' \
  | jq '{time: .time, path: .path, error: .error, request_id: .request_id}'
```

### 500 Internal Server Error

Application panic or unhandled error. Look for:
- `"error":"pq: ..."` → database error
- `"error":"context deadline exceeded"` → timeout
- `"error":"connection refused"` → upstream service down (Redis, Temporal)
- Stack trace with `goroutine` → panic recovered by Fiber middleware

### 503 Service Unavailable

Overload or upstream unavailable:
- Check Redis connectivity (session lookups failing → all requests fail auth)
- Check Temporal worker connectivity

### 504 Gateway Timeout

Request exceeded timeout:
- Check p99 latency in Prometheus
- Look for slow queries (see [Database Issues](03-database-issues.md))

## 3. Check Specific Errors

### Database Errors

```bash
kubectl logs -l app=awoerp-server -n production --since=15m \
  | grep -E '"error":".*pq:|pgx:|context deadline'
```

If DB errors → follow [Database Issues](03-database-issues.md).

### Redis / Session Errors

```bash
kubectl logs -l app=awoerp-server -n production --since=15m \
  | grep -E '"error":".*redis|session'

# Test Redis from pod
kubectl exec -it deploy/awoerp-server -n production -- \
  redis-cli -u "$REDIS_URL" ping
```

### Validation Errors Spike (422)

422s are not 5xx but a spike indicates a client-side issue (bad deploy of frontend, API schema change):

```bash
kubectl logs -l app=awoerp-server -n production --since=15m \
  | grep '"status":422' \
  | jq -r '.path' | sort | uniq -c | sort -rn | head -10
```

## 4. Check Recent Deploys

```bash
kubectl rollout history deployment/awoerp-server -n production

# Check when the error rate started vs deploy time
kubectl describe deployment awoerp-server -n production \
  | grep -A5 "Events:"
```

If errors started immediately after a deploy → rollback is likely the right action.

## 5. Rollback Decision

Rollback if:
- Error rate > 5% and still increasing
- Errors started with a specific deploy
- Root cause not identifiable within 10 minutes

```bash
kubectl rollout undo deployment/awoerp-server -n production
kubectl rollout status deployment/awoerp-server -n production
```

Do **not** rollback if:
- Errors are from an external dependency (DB, Redis) — fix the dependency
- Error rate is steady and < 2% — investigate without rollback

## 6. Correlate with Traces

If OTel tracing is active, find the trace IDs from error logs:

```bash
kubectl logs -l app=awoerp-server -n production --since=15m \
  | grep '"level":"error"' \
  | jq -r '.trace_id' | head -10
```

Look up trace IDs in Jaeger/Tempo to see the full request path and which span failed.

## 7. Notify Affected Users

If error rate is > 5% for > 5 minutes and affecting end users:
1. Post in `#incidents` Slack channel: service name, impact, start time, who is investigating
2. Update status page if public-facing
3. Notify tenant admins via operations team for B2B impact

## Resolution Criteria

- 5xx error rate < 0.1% over a 5-minute window
- No new panics in logs
- Liveness and readiness probes passing
- P99 latency back to baseline (< 500ms)
