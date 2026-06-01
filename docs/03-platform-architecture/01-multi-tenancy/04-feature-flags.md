---
title: Feature Flags
portal: 3 — Platform Architecture
section: 03-platform-architecture
audience: [architect, backend-engineer]
related:
  - "[Tenancy Model](01-tenancy-model.md)"
  - "[Session Architecture](../02-iam/02-session-architecture.md)"
  - "[Configuration Overview](../../04-backend-engineering/00-module-development-guide/20-configuration/01-configuration-overview.md)"
---

# Feature Flags

## Architecture

Feature flags are per-tenant boolean toggles stored in PostgreSQL. They are pre-loaded into `ResolvedSession` at login.

```
feature_flags table
  (name, description, default_enabled)

tenant_feature_overrides table
  (tenant_id, feature_name, enabled)
```

At login:
1. Load `feature_flags` (defaults)
2. Merge with `tenant_feature_overrides` (overrides per tenant)
3. Store result in `ResolvedSession.FeatureFlags`

After session creation, flag changes don't take effect until the user's next login (or forced session invalidation).

## Schema

```sql
CREATE TABLE feature_flags (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name            text NOT NULL UNIQUE,
    description     text NOT NULL,
    default_enabled boolean NOT NULL DEFAULT false,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tenant_feature_overrides (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id),
    feature_name text NOT NULL REFERENCES feature_flags(name),
    enabled      boolean NOT NULL,
    updated_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, feature_name)
);
```

## Using Feature Flags

In service code:

```go
if !sess.FeatureEnabled("contracts.bulk_import") {
    return nil, &domain.BusinessError{
        Code:    "FEATURE_NOT_ENABLED",
        Message: "bulk import is not available for your account",
        Status:  403,
    }
}
```

In AMIS schema (frontend):

```json
{
  "type": "button",
  "label": "Import",
  "visibleOn": "${features.contracts_bulk_import}"
}
```

Features are injected into AMIS `initialData.features` from the session.

## Registering a New Flag

Create a migration:

```sql
INSERT INTO feature_flags (name, description, default_enabled)
VALUES ('contracts.bulk_import', 'Enable CSV bulk import for contracts', false);
```

Naming: `{module}.{feature_description}` — lowercase, dot separator.

New flags default to `false` (opt-in). Rarely use `true` unless replacing a feature that was previously always-on for all tenants.

## Enabling for Specific Tenants

Via Tenant API:

```bash
curl -X PUT /api/v1/tenants/{id}/features/contracts.bulk_import \
  -H "Authorization: Bearer {admin-token}" \
  -d '{"enabled": true}'
```

Via direct DB (for emergency/migration):

```sql
INSERT INTO tenant_feature_overrides (tenant_id, feature_name, enabled)
VALUES ($1, 'contracts.bulk_import', true)
ON CONFLICT (tenant_id, feature_name)
DO UPDATE SET enabled = true, updated_at = now();
```

After enabling, all active sessions for that tenant still have the old value. Invalidate sessions to force reload:

```bash
awoctl sessions revoke-tenant --tenant-id {uuid}
```

## Rollout Strategy

1. New flag registered as `default_enabled: false`
2. Enable for internal/test tenant — verify behavior
3. Enable for beta tenants (10%)
4. Enable for all tenants after 2 weeks of validation
5. Set `default_enabled: true` — remove from `tenant_feature_overrides` for those tenants
6. After all tenants have default true, remove the flag gate from code

## Materialized View for Performance

`mv_tenant_feature_flags_cache` provides a pre-joined view of effective flags per tenant:

```sql
CREATE MATERIALIZED VIEW mv_tenant_feature_flags_cache AS
SELECT
    t.id AS tenant_id,
    ff.name AS feature_name,
    COALESCE(tfo.enabled, ff.default_enabled) AS enabled
FROM tenants t
CROSS JOIN feature_flags ff
LEFT JOIN tenant_feature_overrides tfo
  ON tfo.tenant_id = t.id AND tfo.feature_name = ff.name;
```

Session loading queries this view instead of joining the two tables on every login.

Refresh after flag changes:

```sql
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_tenant_feature_flags_cache;
```

The `CONCURRENTLY` option allows reads to continue during refresh without locking.
