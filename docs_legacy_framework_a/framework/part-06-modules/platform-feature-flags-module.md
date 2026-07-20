> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Platform Module — Feature Flags"
part: "Part VI — Platform Entities"
chapter: 44
section: "platform-feature-flags-module"
related:
  - "[Chapter 41: Platform Modules Overview](./platform-module-overview.md)"
  - "[Chapter 45: Settings Module](./platform-settings-module.md)"
  - "[Chapter 43: IAM Module](./platform-iam-module.md)"
---

# Chapter 44 — Feature Flags Module

> **Who should read this?** Developers rolling out new functionality gradually, operators managing tenant access to beta features, and anyone who wants to understand the difference between a feature flag and a setting.

---

## 44.1 The Mental Model: A Dimmer Switch Panel

Settings (Chapter 45) are like a **thermostat** — they hold a value that business logic reads and acts on. Feature flags are like a **dimmer switch** — they control whether a feature is on, off, or partially on for a given population.

The distinction matters:

| Concept | Feature Flag | Setting |
|---|---|---|
| **Value type** | Boolean or percentage (0–100) | Arbitrary: string, int, bool, JSON |
| **Who changes it** | Platform team | Tenant admin (for their own settings) |
| **Granularity** | Per-tenant override of a global default | Per-tenant value |
| **Purpose** | Gradual rollout, kill-switch, A/B test | Configuration (tax rate, invoice prefix) |
| **UI** | Platform admin panel | Tenant admin panel |

---

## 44.2 Three-Tier Evaluation

Every flag evaluation resolves through three tiers, checked in order:

```
1. Tenant override   — does this tenant have an explicit on/off/percentage?
2. System default    — what is the flag's global default?
3. Built-in fallback — false (safe default if flag doesn't exist yet)
```

This means:
- A flag can be **off globally** but **on for one tenant** (early access program).
- A flag can be **on globally** but **off for one tenant** (exclusion from beta).
- A flag can be **deployed in code** before being activated anywhere — the built-in fallback is always `false`.

---

## 44.3 EntityDefinition: FeatureFlag

```go
// internal/platform/featureflag/definition.go
package featureflag

import "awo.so/framework/definition"

var FeatureFlagDef = definition.EntityDefinition{
    Name:        "feature_flag",
    Label:       "Feature Flag",
    Description: "A global on/off switch for a product feature.",
    Module:      "Platform",
    Table:       "feature_flags",
    OrgScope:    org.ScopeLevelGlobal, // flags are global; overrides are per-tenant

    Fields: []*definition.FieldDef{
        definition.Field("key").
            OfType(definition.FieldTypeData).
            WithLabel("Flag Key").
            RequiredField().
            UniqueField().
            ImmutableField().    // key is referenced in code; changing breaks things
            SearchableField(),

        definition.Field("description").
            OfType(definition.FieldTypeSmallText).
            WithLabel("Description"),

        definition.Field("enabled").
            OfType(definition.FieldTypeBool).
            WithLabel("Enabled by Default").
            WithDefault(false), // safe default: new flags start OFF globally

        definition.Field("rollout_pct").
            OfType(definition.FieldTypeInt).
            WithLabel("Rollout Percentage").
            WithDefault(0),    // 0 = off for all, 100 = on for all

        definition.Field("status").
            OfType(definition.FieldTypeSelect).
            WithLabel("Status").
            WithOptions("draft", "active", "deprecated", "removed").
            WithDefault("draft"),
    },

    Policies: []definition.PolicyDef{
        {Op: definition.OpCreate, Fn: requirePlatformAdmin},
        {Op: definition.OpRead,   Fn: definition.AllowAll},  // any authenticated user can read flags
        {Op: definition.OpUpdate, Fn: requirePlatformAdmin},
        {Op: definition.OpDelete, Fn: requirePlatformAdmin},
    },

    Audited: true,
}
```

### Why `OrgScope: ScopeLevelGlobal`?

Feature flags are global catalogue entries — they exist once and apply to all tenants. The per-tenant customisation lives in `FeatureFlagOverrideDef`. Separating the global definition from tenant overrides means a flag can be updated globally without touching any tenant row.

### Why `enabled` defaults to `false`?

New code that checks `EvalService.Bool("new-billing-flow")` will return `false` until the flag is explicitly enabled. This is the **dark deployment** pattern: ship the code gated behind a flag, verify it deploys cleanly, then enable the flag. If something goes wrong, flip the flag off. No code rollback needed.

---

## 44.4 EntityDefinition: FeatureFlagOverride

```go
var FeatureFlagOverrideDef = definition.EntityDefinition{
    Name:        "feature_flag_override",
    Label:       "Feature Flag Override",
    Description: "A per-tenant override of a global feature flag.",
    Module:      "Platform",
    Table:       "feature_flag_overrides",
    OrgScope:    org.ScopeLevelTenant,

    Fields: []*definition.FieldDef{
        definition.Field("flag_key").
            OfType(definition.FieldTypeData).
            WithLabel("Flag Key").
            RequiredField().
            ImmutableField(),

        definition.Field("enabled").
            OfType(definition.FieldTypeBool).
            WithLabel("Enabled").
            RequiredField(),

        definition.Field("rollout_pct").
            OfType(definition.FieldTypeInt).
            WithLabel("Rollout Percentage").
            WithDefault(100), // when an override enables a flag, default to 100% of tenant
    },

    Policies: []definition.PolicyDef{
        {Op: definition.OpCreate, Fn: requirePlatformAdmin},
        {Op: definition.OpRead,   Fn: allowWithinTenant},
        {Op: definition.OpUpdate, Fn: requirePlatformAdmin},
        {Op: definition.OpDelete, Fn: requirePlatformAdmin},
    },

    Audited: true,
}
```

### The `UNIQUE NULLS NOT DISTINCT` Constraint

There must be at most one override per (tenant, flag_key) pair. The migration uses PostgreSQL 15's `UNIQUE NULLS NOT DISTINCT` syntax:

```sql
CREATE UNIQUE INDEX feature_flag_overrides_unique_idx
    ON feature_flag_overrides (tenant_id, flag_key);
```

If a platform admin tries to create a second override for the same tenant+flag, PostgreSQL returns error code `23505` (unique_violation), which `pgstore.mapPgError` maps to `persistence.ErrConflict`. The API handler translates that to HTTP 409.

---

## 44.5 The Evaluation Service

```go
// internal/platform/featureflag/service.go
type EvalService struct {
    store persistence.TenantStore
    cache *redis.Client
}

// Bool evaluates a boolean feature flag for a tenant.
// Returns false if the flag doesn't exist or the tenant has no override.
func (s *EvalService) Bool(ctx context.Context, tenantID uuid.UUID, flagKey string) bool {
    pct := s.resolve(ctx, tenantID, flagKey)
    return pct == 100
}

// Percentage returns the effective rollout percentage (0–100) for a tenant.
func (s *EvalService) Percentage(ctx context.Context, tenantID uuid.UUID, flagKey string) int {
    return s.resolve(ctx, tenantID, flagKey)
}

// IsEnabled checks percentage rollout for a specific entity (e.g. a specific invoice).
// Deterministic: same entityID always returns the same result for a given percentage.
func (s *EvalService) IsEnabled(ctx context.Context, tenantID uuid.UUID, flagKey string, entityID uuid.UUID) bool {
    pct := s.resolve(ctx, tenantID, flagKey)
    if pct == 0 {
        return false
    }
    if pct == 100 {
        return true
    }
    hash := fnv32(tenantID.String() + ":" + flagKey + ":" + entityID.String())
    return int(hash%100) < pct
}

func (s *EvalService) resolve(ctx context.Context, tenantID uuid.UUID, flagKey string) int {
    // Fast path: Redis cache.
    cacheKey := fmt.Sprintf("fflag:%s:%s", tenantID, flagKey)
    if cached, err := s.cache.Get(ctx, cacheKey).Int(); err == nil {
        return cached
    }

    // Slow path: DB lookup.
    // 1. Check tenant override.
    overrides, _ := s.store.ForEntity(ctx, tenantID, "feature_flag_override")
    page, _ := overrides.List(ctx, persistence.ListOptions{
        Filter: map[string]any{"flag_key": flagKey},
        Limit:  1,
    })
    if len(page.Records) > 0 {
        if !page.Records[0].Get("enabled").(bool) {
            s.cache.Set(ctx, cacheKey, 0, 5*time.Minute)
            return 0
        }
        pct := int(page.Records[0].Get("rollout_pct").(int64))
        s.cache.Set(ctx, cacheKey, pct, 5*time.Minute)
        return pct
    }

    // 2. Fall back to global flag default.
    globalStore, _ := s.store.ForEntity(ctx, uuid.Nil, "feature_flag")
    flags, _ := globalStore.List(ctx, persistence.ListOptions{
        Filter: map[string]any{"flag_key": flagKey},
        Limit:  1,
    })
    if len(flags.Records) == 0 {
        return 0 // flag doesn't exist → safe default
    }
    flag := flags.Records[0]
    if !flag.Get("enabled").(bool) {
        s.cache.Set(ctx, cacheKey, 0, 5*time.Minute)
        return 0
    }
    pct := int(flag.Get("rollout_pct").(int64))
    s.cache.Set(ctx, cacheKey, pct, 5*time.Minute)
    return pct
}

func fnv32(s string) uint32 {
    h := fnv.New32a()
    h.Write([]byte(s))
    return h.Sum32()
}
```

### Why FNV Hash for Percentage Rollout?

The `IsEnabled` method must be **deterministic** — the same `(tenantID, flagKey, entityID)` tuple must always return the same result within a given rollout percentage. If it were random, an invoice would be "enabled" on one request and "disabled" on the next.

FNV-32a is fast, deterministic, and distributes reasonably uniformly across `[0, 100)`. The `entityID` in the hash ensures different entities get different outcomes at the same percentage — you don't enable/disable the same 50% of entities for every flag.

---

## 44.6 The `RegisterFlag` Pattern

Business modules declare their flags at init time:

```go
// module/finance/finance.go
package finance

import (
    "awo.so/internal/platform/featureflag"
    "awo.so/framework/definition"
)

func init() {
    definition.Register(&InvoiceDef)
    definition.Register(&AccountDef)

    featureflag.RegisterFlag(featureflag.Flag{
        Key:         "finance.new-tax-engine",
        Description: "Use the redesigned VAT calculation engine (Kenya eTIMS v2)",
        Enabled:     false,
        Status:      "draft",
    })
}
```

`RegisterFlag` is a no-op if the flag already exists in the database (idempotent). On first deploy, it inserts the flag row. This ensures every flag referenced in code exists in the database from the moment the binary starts — no manual database seeding required.

---

## 44.7 Flag Lifecycle

```
draft → active → deprecated → removed
```

| Status | Meaning | Code still checks it? |
|---|---|---|
| `draft` | Defined in code, not yet enabled anywhere | Yes |
| `active` | Live; may have tenant overrides | Yes |
| `deprecated` | Being phased out; new code should not check it | Yes (legacy paths) |
| `removed` | Deleted from code; row kept for audit trail | No |

The `EvalService` does not enforce lifecycle — it always evaluates. Lifecycle is a documentation signal and an operational gate: platform admins can see which flags are deprecated and plan removal.

---

## 44.8 Migration

```sql
-- internal/platform/featureflag/migrations/20240101000001_create_feature_flags.up.sql

CREATE TABLE feature_flags (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    updated_at   timestamptz NOT NULL DEFAULT NOW(),
    key          text NOT NULL UNIQUE,
    description  text,
    enabled      boolean NOT NULL DEFAULT false,
    rollout_pct  int NOT NULL DEFAULT 0 CHECK (rollout_pct BETWEEN 0 AND 100),
    status       text NOT NULL DEFAULT 'draft'
                     CHECK (status IN ('draft','active','deprecated','removed'))
);

-- No RLS: feature_flags is global (no tenant_id column).

-- internal/platform/featureflag/migrations/20240101000002_create_overrides.up.sql

CREATE TABLE feature_flag_overrides (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id),
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    updated_at   timestamptz NOT NULL DEFAULT NOW(),
    flag_key     text NOT NULL REFERENCES feature_flags(key),
    enabled      boolean NOT NULL,
    rollout_pct  int NOT NULL DEFAULT 100 CHECK (rollout_pct BETWEEN 0 AND 100),
    UNIQUE (tenant_id, flag_key)
);

CREATE INDEX feature_flag_overrides_tenant_idx ON feature_flag_overrides (tenant_id);

ALTER TABLE feature_flag_overrides ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON feature_flag_overrides
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

---

## 44.9 FAQ

**Q: Should I use a feature flag or a setting for "enable dark mode"?**
A: Setting. Dark mode is a user preference with a specific value per user/tenant — it has no rollout semantics. Feature flags are for controlling code path availability during releases.

**Q: What happens if the flag key in code doesn't match the database?**
A: `EvalService.Bool` returns `false` (the safe default). The `RegisterFlag` pattern prevents this: flags are self-registering at startup, so a missing database row is created automatically.

**Q: Can I use feature flags at the org unit level (per branch)?**
A: Not directly. The current model is per-tenant only. For per-branch variations, use a Setting scoped to org unit level, or add a `unit_id` column to `feature_flag_overrides` in a future migration.

**Q: How long does the Redis cache hold flag values?**
A: 5 minutes. This means a flag change takes up to 5 minutes to propagate to all running servers. For emergency kill-switches, platform admins can invalidate the cache key directly via the admin panel.
