> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Event Outbox Backlog Runbook
portal: 9 — Operations
section: 09-operations
audience: [devops, sre, backend-engineer]
related:
  - "[Operations Overview](01-operations-overview.md)"
  - "[Outbox Pattern](../03-platform-architecture/04-event-architecture/03-outbox-pattern.md)"
  - "[Event Bus Internals](../03-platform-architecture/04-event-architecture/02-event-bus-internals.md)"
---

# Event Outbox Backlog Runbook

**Trigger**: `event_outbox_undelivered > 1000` alert, or consumers not receiving events.

## 1. Check Current Backlog

```sql
-- Total undelivered
SELECT COUNT(*) AS undelivered
FROM event_outbox
WHERE delivered_at IS NULL;

-- Breakdown by topic + oldest message
SELECT topic,
       COUNT(*) AS count,
       MIN(created_at) AS oldest,
       MAX(created_at) AS newest
FROM event_outbox
WHERE delivered_at IS NULL
GROUP BY topic
ORDER BY oldest;
```

A backlog growing steadily means the relay goroutine has stopped or is falling behind.

## 2. Check Relay Goroutine Logs

```bash
kubectl logs -l app=awoerp-server -n production --since=30m \
  | grep '"module":"outbox"'
```

Look for:
- `"msg":"relay error"` → relay goroutine hitting errors
- No outbox logs at all → relay goroutine may have panicked and not restarted
- `"msg":"relay paused"` → too many consecutive errors, relay is backing off

## 3. Check Redis Streams (Production)

The outbox relay publishes to Redis Streams. If Redis is unavailable, messages queue in the DB:

```bash
# Check Redis connectivity
kubectl exec -it deploy/awoerp-server -n production -- \
  redis-cli -u "$REDIS_URL" ping

# Check stream lengths
kubectl exec -it deploy/awoerp-server -n production -- \
  redis-cli -u "$REDIS_URL" \
  XLEN awoerp.contracts.submitted
```

If Redis is down → fix Redis first (see Redis section of [Database Issues](03-database-issues.md)), then the backlog will drain automatically once relay reconnects.

## 4. Check Consumer Lag

If relay is publishing to Redis but consumers are lagging:

```bash
# Check consumer group pending count
kubectl exec -it deploy/awoerp-server -n production -- \
  redis-cli -u "$REDIS_URL" \
  XPENDING awoerp.contracts.submitted awoerp-audit-workers - + 10
```

High pending count with no active consumers → consumer worker is down. Check consumer pod status.

## 5. Replay Dead-Letter Events

Events that failed all retry attempts go to the DLQ:

```bash
# List topics with DLQ messages
awoctl events dlq list

# Replay a specific topic's DLQ from a date
awoctl events replay --dlq contracts.submitted --from 2025-01-15

# Replay all DLQ topics
awoctl events replay --all-dlq --from 2025-01-15
```

## 6. Manually Mark Events as Delivered

**Use only as last resort** — if events are definitely stale and replaying would cause issues:

```sql
-- Mark specific old events as delivered (skip them)
-- First: verify what you're skipping
SELECT id, topic, payload->>'contract_id', created_at
FROM event_outbox
WHERE delivered_at IS NULL
  AND created_at < now() - interval '7 days'
  AND topic = 'contracts.submitted'
LIMIT 10;

-- Then mark delivered (add WHERE clause — never mark all)
UPDATE event_outbox
SET delivered_at = now()
WHERE delivered_at IS NULL
  AND created_at < now() - interval '7 days'
  AND topic = 'contracts.submitted';
```

Get DBA sign-off before running UPDATE on `event_outbox`.

## 7. Restart Relay (if relay is stuck)

The relay runs inside the API server process. Restart the pods:

```bash
kubectl rollout restart deployment/awoerp-server -n production
kubectl rollout status deployment/awoerp-server -n production
```

Monitor backlog drain:

```sql
-- Run every 30 seconds to watch drain rate
SELECT COUNT(*) AS undelivered, now() AS checked_at
FROM event_outbox
WHERE delivered_at IS NULL;
```

## 8. Prevent Accumulation

Large backlogs indicate the relay is consistently behind. Long-term fixes:
- Increase relay batch size (config: `OUTBOX_RELAY_BATCH_SIZE`)
- Increase relay goroutine poll interval (make it more frequent)
- Move high-volume topics to direct async (no outbox) if at-least-once DB durability isn't needed

## Resolution Criteria

- `SELECT COUNT(*) FROM event_outbox WHERE delivered_at IS NULL` < 100
- No new messages older than 60 seconds in the backlog
- Relay logs show successful deliveries (`"msg":"events relayed"`)
- Consumer pending counts in Redis are draining
