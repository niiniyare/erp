---
title: Migration Failure Runbook
portal: 9 — Operations
section: 09-operations
audience: [devops, sre, backend-engineer]
related:
  - "[Operations Overview](01-operations-overview.md)"
  - "[Migration Strategy](../03-platform-architecture/03-data-architecture/03-migration-strategy.md)"
  - "[Service Down](02-service-down.md)"
---

# Migration Failure Runbook

**Trigger**: Migration job fails during deploy. App pods may be in `Init:Error` or may have started on old schema.

## 1. Identify the Failure

```bash
# Find the migration job
kubectl get jobs -n production | grep migrate

# View migration job logs
kubectl logs job/awoerp-migrate -n production

# If job has multiple pods (retries)
kubectl logs -l job-name=awoerp-migrate -n production
```

## 2. Classify the Error

### Syntax Error in SQL

```
ERROR: syntax error at or near "CONCURRENTLY" (SQLSTATE 42601)
```

The migration file has a SQL error. Must fix source and redeploy — cannot auto-recover.

**Immediate action**: roll back the deploy, fix SQL, redeploy.

```bash
kubectl rollout undo deployment/awoerp-server -n production
```

### Migration Already Applied (Dirty State)

```
error: Dirty database version {N}. Fix and force version.
```

A previous migration run failed partway through. `migrate` library marks the DB as "dirty".

```bash
# Connect to DB and check state
psql "$DATABASE_URL" -c "SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5;"

# Force-mark the version as clean (only if you've verified the migration DID partially apply)
migrate -database "$DATABASE_URL" -path db/migration force {N}

# Then apply remaining migrations
migrate -database "$DATABASE_URL" -path db/migration up
```

**Warning**: `force` skips re-applying the failed migration. Manually verify the schema is in the expected state before continuing.

### Lock Timeout (concurrent migration)

```
ERROR: deadlock detected
ERROR: canceling statement due to lock timeout
```

Another process holds a lock on the table being migrated (e.g., autovacuum, long-running query).

```sql
-- Find blocking processes
SELECT pid, query, state, wait_event_type, wait_event
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY query_start;

-- Cancel the blocking process if safe
SELECT pg_cancel_backend({pid});
```

Then re-run the migration job:

```bash
kubectl delete job awoerp-migrate -n production
kubectl apply -f k8s/jobs/migrate.yaml -n production
```

### Constraint Violation

```
ERROR: insert or update on table "X" violates foreign key constraint
ERROR: column "X" of relation "Y" contains null values
```

Existing data violates the new constraint. The migration needs a data fix step before adding the constraint.

**Do not force this** — it indicates a schema design issue. Roll back deploy, fix migration with a data migration step, redeploy.

## 3. Check What Applied

```sql
-- View applied migrations in order
SELECT version, dirty FROM schema_migrations ORDER BY version;

-- Compare to files in db/migration/
-- Any version in files but NOT in table = not yet applied
-- Any version in table but NOT in files = orphaned
```

## 4. Manual Recovery

If automated migration is stuck and app needs to be running (emergency):

```bash
# Run migration manually from a pod
kubectl run migrate-manual \
  --image=awoerp:latest \
  --restart=Never \
  --env DATABASE_URL="$DATABASE_URL" \
  -n production \
  -- /app/awoerp migrate up
```

## 5. Zero-Downtime Migration Principles

All production migrations must follow expand-contract:

1. **Expand**: add new column nullable (no lock required)
2. **Deploy app**: writes to both old and new columns
3. **Backfill**: `UPDATE ... WHERE new_col IS NULL` (in batches)
4. **Contract**: add NOT NULL constraint, drop old column

If a migration violated this pattern and caused an outage, document it as a post-mortem item.

## 6. Post-Incident

After recovery:
1. Verify row counts match expectations
2. Run a smoke test against critical endpoints
3. Check migration table — all versions clean (not dirty)
4. Write post-mortem if downtime > 5 minutes

```sql
-- Final check: no dirty state
SELECT COUNT(*) FROM schema_migrations WHERE dirty = true;
-- Should return 0
```
