---
title: Module Deployment Checklist
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, devops]
related:
  - "[Module Launch Checklist](../23-worked-example/10-launch-checklist.md)"
  - "[Kubernetes Deployment](../../../06-devops/03-kubernetes-deployment.md)"
  - "[Migration Strategy](../../../03-platform-architecture/03-data-architecture/03-migration-strategy.md)"
---

# Module Deployment Checklist

Use this checklist when deploying a new module or significant feature to production.

## Pre-Deploy: Code Review

- [ ] All PR reviewers approved
- [ ] No `TODO: before deploy` or `FIXME` comments left in code
- [ ] SQLC generated (`make sqlc`) — `wire_gen.go` and `db/sqlc/` committed
- [ ] Wire generated (`make wire`) — `wire_gen.go` committed
- [ ] All tests passing in CI (`go test ./... -race`)
- [ ] No new `govulncheck` findings
- [ ] `golangci-lint` clean

## Pre-Deploy: Database

- [ ] Migration files follow naming convention (`{group}{seq}_description.up.sql`)
- [ ] Migration tested on a copy of production DB
- [ ] Migration is reversible (`.down.sql` exists and tested)
- [ ] No `LOCK TABLE` or `ALTER TABLE ... ADD COLUMN NOT NULL` without default on large tables
- [ ] No `missing_ok` in `SET LOCAL app.tenant_id`
- [ ] New tables have RLS enabled and policy created
- [ ] New tables have `tenant_id`, `created_at`, `updated_at`, `deleted_at` as appropriate
- [ ] Indexes created for common query patterns (especially on `tenant_id` + filter columns)

## Pre-Deploy: API

- [ ] New endpoints documented in Portal 8 (API Reference)
- [ ] All new endpoints registered in Wire route registrar
- [ ] All new endpoints protected by `Authenticate` middleware
- [ ] Authorization middleware applied with correct permission string
- [ ] Rate limits appropriate for endpoint type

## Pre-Deploy: Feature Flags

- [ ] New feature flags registered in `feature_flags` table (via migration)
- [ ] Default value set correctly (`false` for new features, `true` for replacements)
- [ ] Feature flagged code has `sess.FeatureEnabled()` guard
- [ ] Plan for enabling flag per tenant post-deploy

## Deploy Steps

1. **Merge PR to main** — CI runs automatically
2. **Wait for staging deploy** — CD pipeline deploys to staging
3. **Smoke test staging** — exercise key endpoints with test tenant
4. **Run production migration**:
   ```bash
   kubectl apply -f k8s/jobs/migrate-{version}.yaml -n production
   kubectl wait --for=condition=complete job/awoerp-migrate-{version} -n production --timeout=300s
   ```
5. **Deploy to production**:
   ```bash
   kubectl set image deployment/awoerp-server awoerp-server=ghcr.io/org/awoerp:{sha} -n production
   kubectl rollout status deployment/awoerp-server -n production
   ```
6. **Monitor for 10 minutes**: check error rate, latency, DB pool

## Post-Deploy: Verify

- [ ] Readiness probe passes: `curl http://{host}:8081/health/ready`
- [ ] No new error spikes in Grafana
- [ ] P99 latency baseline unchanged
- [ ] DB pool not exhausted
- [ ] New endpoints return correct responses on smoke test
- [ ] Audit log entries being written for new operations
- [ ] Event outbox not backing up

## Post-Deploy: Enable Feature Flag

If feature is behind a flag:

1. Enable for internal/test tenant first
2. Monitor for 30 minutes
3. Enable for beta tenants
4. Enable broadly after validation

```bash
# Via API
curl -X PUT /api/v1/tenants/{id}/features/contracts.bulk_import \
  -H "Authorization: Bearer {admin-token}" \
  -d '{"enabled": true}'
```

## Rollback Plan

If post-deploy issues detected:

```bash
# Rollback image (no migration rollback needed if expand-contract followed)
kubectl rollout undo deployment/awoerp-server -n production

# If migration must be rolled back (rare — only for non-expand-contract migrations)
migrate -database "$DATABASE_URL" -path db/migration down 1
```

**Communicate**: post in `#incidents` with what happened, what was rolled back, ETA for fix.
