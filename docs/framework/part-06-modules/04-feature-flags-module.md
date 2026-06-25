---
title: "Platform Module: Feature Flags"
part: "Part VI — Platform Entities"
chapter: 44
section: "platform-feature-flags-module"
related:
  - "[Chapter 43: IAM Module](./platform-iam-module.md)"
  - "[Chapter 45: Settings Module](./platform-settings-module.md)"
  - "[Chapter 41: Platform Module Overview](./platform-module-overview.md)"
---

# Chapter 44 — Feature Flags Module

> **Primary source for:** the `feature_flags` and `feature_flag_overrides` tables, three-tier flag evaluation, Redis caching strategy, flag lifecycle, and how module developers gate code behind flags.
>
> **Audience:** new developers who need to understand how new features are safely rolled out, business stakeholders who want to know how Acme Petroleum can get a feature before Beta Gas Co, and anyone implementing a flag-gated feature.

---

## 44.1 What Is a Feature Flag?

A **feature flag** (also called a feature toggle or feature switch) is a named on/off switch for a piece of functionality in the system. The critical property: you can change it *without redeploying code*.

### Why Does This Matter?

Traditional software releases work like a light switch on a circuit breaker panel: you flip one switch and every room in the building changes at once. If something goes wrong, you flip it back.

Feature flags give you a dimmer switch per room. You can:

- Turn on "LPG cylinder deposit tracking" for Acme Petroleum only, while keeping it off for everyone else until it's proven in production.
- Gradually roll out "new invoice PDF layout" to 10% of tenants, watch for errors, then ramp to 25%, 50%, 100%.
- Give your own QA team access to "eTIMS integration" in production while keeping it off for real customers.
- If a feature causes database performance issues, turn it off instantly — without a hotfix deploy.

### Flags vs Settings

| | Feature Flag | Setting (Chapter 45) |
|---|---|---|
| **Answers** | "Is this feature ON?" | "How does this feature behave?" |
| **Type** | boolean / percentage / string | string / integer / boolean / json / date |
| **Changed by** | Platform admin or auto-rollout | Tenant admin |
| **Example** | `finance.etims_integration = true` | `finance.invoice_prefix = "ACM-"` |

A flag controls *whether* code runs. A setting controls *how* that code runs.

---

## 44.2 Three-Tier Evaluation

Every flag evaluation resolves through three levels, most specific wins:

```
User Override    ─── "Jane sees the new dashboard UI" (user_id set)
      │
      ▼ (if no user override)
Tenant Override  ─── "All of Acme Petroleum sees the new module" (user_id null)
      │
      ▼ (if no tenant override)
System Default   ─── "flag not yet enabled anywhere" (default_value on FeatureFlag row)
```

**Why this order?**

The same logic as CSS specificity: the most targeted rule wins. A platform engineer testing one user's experience sets a user override. A sales engineer enables a module for one paying customer via a tenant override. Unconfigured tenants get the system default.

---

## 44.3 Directory Layout

```
framework/platform/featureflag/
├── featureflag.go   ← init() — registers FeatureFlagDef and FeatureFlagOverrideDef
├── definition.go    ← EntityDefinition declarations
├── policy.go        ← policy helpers
├── hooks.go         ← invalidateFlagCache
├── service.go       ← EvalService: Bool(), String(), Percentage()
├── register.go      ← RegisterFlag() called by module init() functions
└── migrations/
    └── 20240101000020_create_feature_flags.up.sql
```

---

## 44.4 FeatureFlag EntityDefinition (the Catalogue)

The `feature_flags` table is the **catalogue** — it defines what flags exist and their system-wide defaults. It is a global table (no `tenant_id`) because the catalogue is the same for every deployment.

```go
// framework/platform/featureflag/definition.go
package featureflag

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

var FeatureFlagDef = definition.EntityDefinition{
    Name:     "feature_flag",
    Label:    "Feature Flag",
    Module:   "Platform",
    Table:    "feature_flags",

    // GLOBAL: the catalogue is shared across all tenants.
    // There is no per-tenant catalogue — only per-tenant overrides (see below).
    OrgScope: org.ScopeLevelGlobal,
    Audited:  true,

    Fields: []*definition.FieldDef{
        {
            Name:     "key",
            Type:     definition.FieldTypeData,
            Label:    "Flag Key",
            Required: true,
            Description: "Stable dot-namespaced identifier. Never change after creation. " +
                "Format: '{module}.{feature}'. " +
                "Examples: 'forecourt.wetstock_alerts', 'finance.etims_integration', " +
                "'hr.payroll_auto_submit'.",
        },
        {Name: "name",        Type: definition.FieldTypeData,      Label: "Display Name", Required: true},
        {Name: "description", Type: definition.FieldTypeSmallText, Label: "Description",
            Description: "Explain what enabling this flag changes. Helps admins understand impact."},
        {
            Name:     "type",
            Type:     definition.FieldTypeSelect,
            Label:    "Value Type",
            Options:  []string{"boolean", "string", "percentage"},
            Required: true,
            Description: "boolean: on/off. " +
                "string: a variable value (e.g. a colour theme name). " +
                "percentage: 0–100 for gradual rollout.",
        },
        {
            Name:     "default_value",
            Type:     definition.FieldTypeData,
            Label:    "System Default",
            Required: true,
            Description: "Value used when no tenant or user override exists. " +
                "For boolean: 'true' or 'false'. " +
                "For percentage: '0' to '100'. " +
                "For string: the default string value.",
        },
        {
            Name:    "status",
            Type:    definition.FieldTypeSelect,
            Label:   "Lifecycle Status",
            Options: []string{"draft", "active", "deprecated", "removed"},
            Required: true,
            Description: "draft: flag exists in code but not yet visible to operators. " +
                "active: available for overrides, evaluated in production. " +
                "deprecated: marked for removal; still evaluated; prompts cleanup. " +
                "removed: all code references deleted; safe to purge rows.",
        },
        {
            Name:        "module",
            Type:        definition.FieldTypeData,
            Label:       "Owning Module",
            Description: "Which module registered this flag. 'forecourt', 'hr', 'platform'. " +
                "Used for admin UI grouping.",
        },
    },

    Policies: []definition.PolicyDef{
        // System and platform admins manage the catalogue.
        definition.Policy(definition.OpAll, definition.AllowSystem),
        definition.Policy(definition.OpAll, requireRole("platform_admin")),

        // All authenticated users may READ the catalogue.
        // Reason: the EvalService needs to read defaults. Module code checks flags
        // on every request. Restricting reads would break flag evaluation.
        definition.Policy(definition.OpRead, allowTenantViewer),

        definition.Policy(definition.OpAll, definition.DenyAll),
    },
}
```

---

## 44.5 FeatureFlagOverride EntityDefinition (the Per-Tenant Values)

Overrides are where the per-tenant and per-user configuration lives. A single flag can have thousands of override rows — one per tenant that has configured it.

```go
var FeatureFlagOverrideDef = definition.EntityDefinition{
    Name:     "feature_flag_override",
    Label:    "Feature Flag Override",
    Module:   "Platform",
    Table:    "feature_flag_overrides",

    // TENANT-SCOPED: each override row belongs to one tenant.
    // RLS ensures Tenant A cannot see or modify Tenant B's overrides.
    OrgScope: org.ScopeLevelTenant,
    Audited:  true,

    Fields: []*definition.FieldDef{
        {
            Name:         "flag_key",
            Type:         definition.FieldTypeLink,
            Label:        "Flag",
            TargetEntity: "feature_flag",
            Required:     true,
        },
        {
            Name:         "user_id",
            Type:         definition.FieldTypeLink,
            Label:        "User",
            TargetEntity: "user",
            Description: "NULL = tenant-level override (applies to all users in this tenant). " +
                "Non-NULL = user-level override (applies only to this specific user).",
        },
        {
            Name:     "value",
            Type:     definition.FieldTypeData,
            Label:    "Override Value",
            Required: true,
            Description: "Stored as text. Interpreted according to the flag's 'type'. " +
                "For boolean: 'true' or 'false'. For percentage: '25'. For string: the value.",
        },
        {Name: "enabled_at", Type: definition.FieldTypeDateTime, Label: "Enabled At"},
        {Name: "created_by", Type: definition.FieldTypeLink,     Label: "Set By",
            TargetEntity: "user"},
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll, definition.AllowSystem),
        // Only tenant admins can set overrides.
        // A regular user cannot give themselves access to an unreleased feature.
        definition.Policy(definition.OpAll, requireRole("admin")),
        definition.Policy(definition.OpAll, definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            // After any override change, clear the Redis flag snapshot for this tenant.
            // The next EvalService.Bool() call will re-read from PostgreSQL and
            // repopulate the cache. Cache miss → DB hit → re-cache → serve.
            //
            // WHY AfterHook, not BeforeHook:
            //   The DB write must commit successfully first. If we cleared the cache
            //   BEFORE the write and the write then failed, we would have an empty
            //   cache and no updated override — incorrect state.
            Name: "invalidate_flag_cache",
            Ops:  definition.OpCreate | definition.OpUpdate | definition.OpDelete,
            When: definition.HookAfter,
            Fn:   invalidateFlagCache,
        },
    },
}
```

---

## 44.6 The EvalService

The EvalService is the public API that all module code uses to check flags. It hides the three-tier resolution behind simple typed methods.

```go
// framework/platform/featureflag/service.go
package featureflag

import (
    "context"
    "fmt"
    "strconv"

    "awo.so/framework/contextutil"
    "awo.so/framework/persistence/pgstore"
)

type EvalService struct {
    flags     pgstore.EntityStore // FeatureFlagDef
    overrides pgstore.EntityStore // FeatureFlagOverrideDef
    redis     RedisClient
}

// Bool evaluates a boolean flag for the current viewer context.
// Fast path: session snapshot in context (pre-loaded at login, O(1) map lookup).
// Slow path: direct DB cascade (used by background jobs, bootstrap, admin tools).
func (s *EvalService) Bool(ctx context.Context, key string) bool {
    val := s.resolve(ctx, key)
    return val == "true"
}

// Percentage evaluates a percentage flag and returns whether this tenant/user
// falls within the enabled percentage.
// Uses a deterministic hash so the same tenant always gets the same answer.
func (s *EvalService) Percentage(ctx context.Context, key string) bool {
    tenantID := contextutil.TenantID(ctx)
    val := s.resolve(ctx, key)
    threshold, err := strconv.Atoi(val)
    if err != nil || threshold <= 0 {
        return false
    }
    if threshold >= 100 {
        return true
    }
    // Deterministic hash: same tenant always in same bucket.
    // hash(tenant_id + ":" + flag_key) % 100 < threshold
    hash := fnv32(tenantID + ":" + key)
    return int(hash%100) < threshold
}

// resolve runs the three-tier lookup for any flag type.
func (s *EvalService) resolve(ctx context.Context, key string) string {
    tenantID := contextutil.TenantID(ctx)
    userID   := contextutil.UserID(ctx)

    // ── Fast path: session snapshot ──────────────────────────────────────
    // At login, SessionService pre-computes and stores the full flag set
    // for this user in their session context. Each request extracts it once.
    if flags, ok := contextutil.FlagsFromCtx(ctx); ok {
        if val, found := flags[key]; found {
            return val
        }
    }

    // ── Slow path: DB cascade ────────────────────────────────────────────

    // Tier 1: user-level override
    if userID != "" {
        if val, err := s.lookupOverride(ctx, key, tenantID, userID); err == nil {
            return val
        }
    }

    // Tier 2: tenant-level override
    if val, err := s.lookupOverride(ctx, key, tenantID, ""); err == nil {
        return val
    }

    // Tier 3: system default from catalogue
    return s.systemDefault(ctx, key)
}

func (s *EvalService) lookupOverride(ctx context.Context, key, tenantID, userID string) (string, error) {
    filter := map[string]any{"flag_key": key}
    if userID != "" {
        filter["user_id"] = userID
    } else {
        filter["user_id"] = nil // tenant-level = NULL user_id
    }

    result, err := s.overrides.List(ctx, pgstore.ListOptions{
        Filter: filter,
        Limit:  1,
    })
    if err != nil || len(result.Items) == 0 {
        return "", fmt.Errorf("not found")
    }
    val, _ := result.Items[0].Get("value")
    return val.(string), nil
}

func (s *EvalService) systemDefault(ctx context.Context, key string) string {
    result, err := s.flags.List(ctx, pgstore.ListOptions{
        Filter: map[string]any{"key": key},
        Limit:  1,
    })
    if err != nil || len(result.Items) == 0 {
        return "false" // safe default for unknown flags
    }
    val, _ := result.Items[0].Get("default_value")
    return val.(string)
}
```

---

## 44.7 Redis Caching

The EvalService's fast path uses a **session snapshot** — not a standalone Redis cache for flags. Here is why and how:

### Session Snapshot (Preferred)

When a user logs in, the `SessionService.Login()` call:

1. Resolves ALL flags for this tenant+user through the three-tier cascade (one DB query per flag, or a single query with JOIN).
2. Stores the resolved map in the session JSON: `{"flags": {"finance.etims": "true", "hr.payroll_auto_submit": "false", ...}}`.
3. Caches the session JSON in Redis.

On every subsequent request, middleware loads the session from Redis and attaches the flags map to the request context. Flag evaluation is an O(1) map lookup — no DB, no Redis flag-specific query.

### Standalone Redis Cache (Fallback)

For background jobs, Temporal activities, and anywhere a session context is not available, the EvalService checks a secondary Redis cache:

- Key: `flags:{tenant_id}` (a Redis hash of `{flag_key → value}`)
- TTL: 60 seconds
- Built on first miss; cleared by `invalidateFlagCache` AfterHook

```go
// Cache miss path:
func (s *EvalService) warmCache(ctx context.Context, tenantID string) {
    // Query all overrides for this tenant in one shot.
    overrides, _ := s.overrides.List(ctx, pgstore.ListOptions{
        Filter: map[string]any{"tenant_id": tenantID},
    })
    // Build a map and SET all fields in Redis with one HMSET call.
    flagMap := make(map[string]string)
    for _, row := range overrides.Items {
        key,  _ := row.Get("flag_key")
        val,  _ := row.Get("value")
        flagMap[key.(string)] = val.(string)
    }
    s.redis.HMSet(ctx, "flags:"+tenantID, flagMap)
    s.redis.Expire(ctx, "flags:"+tenantID, 60*time.Second)
}
```

---

## 44.8 Module Developer Integration

Module developers should never interact with the EvalService directly. They use the injected `flags` interface provided by the framework:

```go
// In a Finance module handler:
import "awo.so/framework/featureflag"

func (h *InvoiceHandler) Submit(ctx context.Context, id uuid.UUID) error {
    invoice, err := h.invoices.FindByID(ctx, id)
    if err != nil {
        return err
    }

    // Submit to KRA eTIMS only if the integration is enabled for this tenant.
    if featureflag.Bool(ctx, "finance.etims_integration") {
        if err := h.etimsService.Submit(ctx, invoice); err != nil {
            return fmt.Errorf("eTIMS submission failed: %w", err)
        }
    }

    return h.invoices.Update(ctx, invoice)
}
```

```go
// In a Forecourt SDUI page builder:
func buildForecourtDashboard(ctx context.Context) amis.Page {
    page := amis.NewPage("Forecourt Dashboard")

    if featureflag.Bool(ctx, "forecourt.wetstock_alerts") {
        page.AddSection(wetstockAlertsSection())
    }

    if pct := featureflag.Percentage(ctx, "forecourt.new_pump_layout"); pct {
        page.AddSection(newPumpLayoutSection())
    } else {
        page.AddSection(legacyPumpLayoutSection())
    }

    return page
}
```

### Registering a New Flag

Modules register their flags in `init()`:

```go
// In awo.so/module/finance — init()
import "awo.so/framework/platform/featureflag"

func init() {
    definition.Register(&InvoiceDef)
    // ... other entity registrations ...

    featureflag.RegisterFlag(featureflag.FlagDef{
        Key:          "finance.etims_integration",
        Name:         "KRA eTIMS Integration",
        Description:  "Submit VAT invoices to Kenya Revenue Authority eTIMS system in real-time.",
        Type:         "boolean",
        DefaultValue: "false",   // off by default; turned on per tenant when they're onboarded
        Module:       "finance",
        Status:       "active",
    })
}
```

`RegisterFlag` upserts the row into `feature_flags` at startup. If the row already exists, it updates the name/description but never changes the `default_value` of a flag that is already `active` (that would silently change behaviour for tenants relying on the default).

---

## 44.9 Migration

```sql
-- framework/platform/featureflag/migrations/20240101000020_create_feature_flags.up.sql

-- ── Feature Flags Catalogue ─────────────────────────────────────────────────
-- Global table: no tenant_id, no RLS.
-- Access controlled entirely at app layer via FeatureFlagDef.Policies.
CREATE TABLE feature_flags (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    key           text        NOT NULL UNIQUE,  -- the stable dot-namespaced key
    name          text        NOT NULL,
    description   text,
    type          text        NOT NULL DEFAULT 'boolean'
                              CHECK (type IN ('boolean', 'string', 'percentage')),
    default_value text        NOT NULL DEFAULT 'false',
    status        text        NOT NULL DEFAULT 'draft'
                              CHECK (status IN ('draft', 'active', 'deprecated', 'removed')),
    module        text,       -- owning module name for UI grouping
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON feature_flags
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();


-- ── Feature Flag Overrides ───────────────────────────────────────────────────
CREATE TABLE feature_flag_overrides (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    flag_key    text        NOT NULL REFERENCES feature_flags(key) ON DELETE CASCADE,
    user_id     uuid        REFERENCES users(id), -- NULL = tenant-level override
    value       text        NOT NULL,
    enabled_at  timestamptz NOT NULL DEFAULT now(),
    created_by  uuid        REFERENCES users(id),
    created_at  timestamptz NOT NULL DEFAULT now(),

    -- One override per (tenant, flag, user) combination.
    -- NULLS NOT DISTINCT (PostgreSQL 15+): two NULL user_ids are treated as equal
    -- in this constraint. Without it you could accidentally create two tenant-level
    -- overrides for the same flag (both have user_id = NULL).
    UNIQUE NULLS NOT DISTINCT (tenant_id, flag_key, user_id)
);

ALTER TABLE feature_flag_overrides ENABLE ROW LEVEL SECURITY;

-- Standard RLS: tenant sees only their own overrides.
CREATE POLICY ff_overrides_tenant_isolation ON feature_flag_overrides
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Fast lookup for the three-tier evaluation cascade.
CREATE INDEX ff_overrides_lookup_idx ON feature_flag_overrides
    (tenant_id, flag_key, user_id NULLS FIRST);
```

---

## 44.10 Flag Lifecycle

Flags have a four-stage lifecycle. Understanding it prevents orphan flags and stale code.

```
draft ──► active ──► deprecated ──► removed
```

| Stage | DB row exists? | Evaluated? | Code references? | Admin visible? |
|---|---|---|---|---|
| `draft` | ✅ | ✅ (uses default) | ✅ (gating code written) | Only platform admins |
| `active` | ✅ | ✅ | ✅ | All operators |
| `deprecated` | ✅ | ✅ | ⚠️ Removal in progress | All operators (warning shown) |
| `removed` | ✅ (archival) | ❌ (never reached) | ❌ | Hidden from UI |

### Developer Workflow: Adding a New Flag

1. Add `featureflag.RegisterFlag(...)` with `Status: "draft"` in your module `init()`.
2. Write code gated behind `featureflag.Bool(ctx, "module.feature")`.
3. Deploy. Flag exists but evaluates to default (usually `false`). Zero impact.
4. Set `Status: "active"` when you're ready to enable for tenants.
5. Create tenant overrides via admin panel for the tenants you want to enable it for.
6. Ramp to 100% (all tenants get tenant override `value = "true"`) or change `default_value` to `"true"` to enable for all new tenants going forward.

### Developer Workflow: Removing a Flag

1. Set `Status: "deprecated"`. This signals to the team that cleanup is needed.
2. Remove all `featureflag.Bool(ctx, "module.feature")` calls from code.
3. Remove the `featureflag.RegisterFlag(...)` call.
4. Set `Status: "removed"` in a migration (or let the next cleanup migration do it).
5. Optionally: delete override rows and the flag row (after 30-day archival window).

---

## 44.11 Frequently Asked Questions

**"Why not use environment variables for feature flags?"**

Environment variables require a process restart to change. They are the same for every tenant. They are invisible to non-developers. Feature flags solve all three problems.

**"Why not just have a settings key for every feature?"**

Settings are for "how does this work?" (configuration values). Flags are for "does this run at all?" (code path control). Mixing them makes settings hard to understand and flag evaluation awkward. Keep them separate.

**"If a flag doesn't exist in the catalogue, what does `Bool()` return?"**

`false`. Unknown flags default to `false` (safe off). This means if a developer checks `featureflag.Bool(ctx, "finance.etims_integration")` before calling `RegisterFlag`, the feature is simply off — not a crash, not an error.

**"What is the maximum number of flags?"**

No hard limit. The Redis session snapshot contains all resolved flags — if a tenant has 500 active flags, their session JSON is larger but still sub-kilobyte. Aim for fewer than 200 active flags per module to keep the system comprehensible. Archive flags aggressively.

**"Can a module check another module's flag?"**

Yes, by key string. But this creates a coupling between modules. Prefer instead: the other module exposes a function `IsEnabled(ctx) bool` that wraps the flag check. Module consumers call `forecourt.IsEnabled(ctx)`, not `featureflag.Bool(ctx, "forecourt.some_flag")`. This keeps the flag key internal to the module that owns it.
