---
title: "Platform Module — Settings & Configuration"
part: "Part VI — Platform Entities"
chapter: 45
section: "platform-settings-module"
related:
  - "[Chapter 41: Platform Modules Overview](./platform-module-overview.md)"
  - "[Chapter 44: Feature Flags Module](./platform-feature-flags-module.md)"
  - "[Chapter 42: Tenant Module](./platform-tenant-module.md)"
---

# Chapter 45 — Settings & Configuration Module

> **Who should read this?** Developers adding configuration to a business module, tenant admins customising their deployment, and anyone trying to understand why a setting isn't taking effect.

---

## 45.1 Settings vs Feature Flags — When to Use Each

Both Settings and Feature Flags are stored per-tenant and change behaviour at runtime. They serve different purposes:

| Concern | Settings | Feature Flags |
|---|---|---|
| **What it holds** | A typed value (string, int, bool, JSON) | On/off or percentage |
| **Who configures it** | Tenant admin (for their own), platform admin (for system) | Platform team only |
| **Typical use** | Invoice prefix, tax rate, locale, SMTP host | New-feature rollout, kill-switch |
| **Hierarchy** | Three levels: system → tenant → (future: unit) | Two levels: global → tenant override |
| **SDUI** | Auto-generated settings page per module | Platform admin panel only |

**Rule of thumb**: if a value needs to be configured by the business (tax rate, payment terms), use Settings. If it controls whether a code path is active during a release, use Feature Flags.

---

## 45.2 Three-Level Hierarchy

Settings cascade through three levels, evaluated in order:

```
1. Tenant value    — has this tenant set a value for this key?
2. System default  — what is the system-wide default for this key?
3. Built-in value  — the Go default declared by RegisterKey()
```

Most tenants never touch most settings. The system default means a sensible value is always available. Tenant admins override only what is specific to their business.

---

## 45.3 EntityDefinition: ConfigKey

```go
// internal/platform/settings/definition.go
package settings

import "awo.so/framework/definition"

var ConfigKeyDef = definition.EntityDefinition{
    Name:        "config_key",
    Label:       "Configuration Key",
    Description: "A typed, named configuration parameter with a system default value.",
    Module:      "Platform",
    Table:       "config_keys",
    OrgScope:    org.ScopeLevelGlobal, // keys are a global catalogue

    Fields: []*definition.FieldDef{
        definition.Field("key").
            OfType(definition.FieldTypeData).
            WithLabel("Key").
            RequiredField().
            UniqueField().
            ImmutableField(). // referenced in code; changing breaks things
            SearchableField(),

        definition.Field("label").
            OfType(definition.FieldTypeData).
            WithLabel("Display Label").
            RequiredField(),

        definition.Field("description").
            OfType(definition.FieldTypeSmallText).
            WithLabel("Description"),

        definition.Field("value_type").
            OfType(definition.FieldTypeSelect).
            WithLabel("Value Type").
            WithOptions("string", "int", "bool", "json").
            RequiredField().
            ImmutableField(), // type cannot change after data exists

        definition.Field("default_value").
            OfType(definition.FieldTypeSmallText).
            WithLabel("Default Value"),

        definition.Field("module").
            OfType(definition.FieldTypeData).
            WithLabel("Module").
            RequiredField().
            ImmutableField(), // which module owns this key

        definition.Field("tenant_editable").
            OfType(definition.FieldTypeBool).
            WithLabel("Tenant Editable").
            WithDefault(true), // most settings are tenant-configurable
    },

    Policies: []definition.PolicyDef{
        {Ops: definition.OpCreate, Fn: requirePlatformAdmin},
        {Ops: definition.OpRead,   Fn: definition.AllowAll}, // any user can read key metadata
        {Ops: definition.OpUpdate, Fn: requirePlatformAdmin},
        {Ops: definition.OpDelete, Fn: requirePlatformAdmin},
    },

    Audited: true,
}
```

### Why `OrgScope: ScopeLevelGlobal`?

Config keys are a dictionary — they describe what settings *exist*, not what any particular tenant has chosen. The `config_values` table (below) is where per-tenant values live. This two-table design means adding a new key requires no per-tenant migration — the default_value handles all existing tenants immediately.

### Why `tenant_editable`?

Some keys are internal infrastructure that should only be changed by the platform team (e.g. `platform.db_pool_size`, `platform.redis_cluster`). Setting `tenant_editable: false` hides them from the tenant admin UI and rejects tenant write attempts via the `assertKeyIsTenantEditable` hook.

---

## 45.4 EntityDefinition: ConfigValue

```go
var ConfigValueDef = definition.EntityDefinition{
    Name:        "config_value",
    Label:       "Configuration Value",
    Description: "A per-tenant override of a configuration key.",
    Module:      "Platform",
    Table:       "config_values",
    OrgScope:    org.ScopeLevelTenant,

    Fields: []*definition.FieldDef{
        definition.Field("key").
            OfType(definition.FieldTypeData).
            WithLabel("Key").
            RequiredField().
            ImmutableField(), // which key this overrides

        definition.Field("value").
            OfType(definition.FieldTypeSmallText).
            WithLabel("Value").
            RequiredField(),
    },

    Hooks: []definition.HookDef{
        definition.BeforeSaveHook("validateType",    definition.OpWrite, validateValueMatchesKeyType),
        definition.BeforeSaveHook("assertEditable",  definition.OpWrite, assertKeyIsTenantEditable),
        definition.AfterSaveHook("invalidateCache",  definition.OpWrite, invalidateSettingsCache),
    },

    Policies: []definition.PolicyDef{
        {Ops: definition.OpCreate, Fn: requireRole("tenant_admin")},
        {Ops: definition.OpRead,   Fn: allowWithinTenant},
        {Ops: definition.OpUpdate, Fn: requireRole("tenant_admin")},
        {Ops: definition.OpDelete, Fn: requireRole("tenant_admin")},
    },

    Audited: true,
}
```

### The `validateValueMatchesKeyType` Hook

Every value is stored as a string (`smalltext`) for schema flexibility, but the system enforces that the string is valid for the key's declared `value_type`:

```go
// internal/platform/settings/hooks.go
func validateValueMatchesKeyType(ctx context.Context, m *definition.Mutation) error {
    if m.Op == definition.OpDelete {
        return nil
    }
    key   := m.After.Get("key").(string)
    value := m.After.Get("value").(string)

    keyDef, err := lookupConfigKey(ctx, key)
    if err != nil {
        return fmt.Errorf("unknown config key %q: %w", key, err)
    }

    switch keyDef.Get("value_type").(string) {
    case "int":
        if _, err := strconv.ParseInt(value, 10, 64); err != nil {
            return definition.NewFieldError("value", "must be a valid integer for key "+key)
        }
    case "bool":
        if value != "true" && value != "false" {
            return definition.NewFieldError("value", "must be 'true' or 'false' for key "+key)
        }
    case "json":
        if !json.Valid([]byte(value)) {
            return definition.NewFieldError("value", "must be valid JSON for key "+key)
        }
    }
    // string type: any value is valid
    return nil
}
```

### The `invalidateSettingsCache` Hook

After any value change, the Redis cache for that tenant+key must be cleared so the next read picks up the new value:

```go
func invalidateSettingsCache(ctx context.Context, m *definition.Mutation) error {
    rec := m.After
    if rec == nil {
        rec = m.Before // on delete, use before
    }
    key      := rec.Get("key").(string)
    tenantID := m.TenantID
    cacheKey := fmt.Sprintf("setting:%s:%s", tenantID, key)
    return redisClient.Del(ctx, cacheKey).Err()
}
```

---

## 45.5 The Settings Service

```go
// internal/platform/settings/service.go
type Service struct {
    store persistence.TenantStore
    cache *redis.Client
}

// Get returns the effective value for key in the given tenant.
// Resolution order: tenant value → system default → built-in default.
func (s *Service) Get(ctx context.Context, tenantID uuid.UUID, key string) string {
    // Fast path: Redis cache.
    cacheKey := fmt.Sprintf("setting:%s:%s", tenantID, key)
    if v, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
        return v
    }

    // Slow path: DB lookup.
    // 1. Tenant override.
    if v, ok := s.tenantValue(ctx, tenantID, key); ok {
        s.cache.Set(ctx, cacheKey, v, 10*time.Minute)
        return v
    }

    // 2. System default from config_keys table.
    if v, ok := s.systemDefault(ctx, key); ok {
        s.cache.Set(ctx, cacheKey, v, 10*time.Minute)
        return v
    }

    // 3. Built-in default registered by RegisterKey().
    if reg, ok := registeredKeys[key]; ok {
        return reg.Default
    }

    return ""
}

// GetInt returns the setting as int64. Panics on parse error — keys with
// value_type "int" should always hold valid integers.
func (s *Service) GetInt(ctx context.Context, tenantID uuid.UUID, key string) int64 {
    v := s.Get(ctx, tenantID, key)
    n, err := strconv.ParseInt(v, 10, 64)
    if err != nil {
        panic(fmt.Sprintf("settings: key %q value %q is not an integer", key, v))
    }
    return n
}

// GetBool returns the setting as bool.
func (s *Service) GetBool(ctx context.Context, tenantID uuid.UUID, key string) bool {
    return s.Get(ctx, tenantID, key) == "true"
}
```

### Why Panic on Bad Int?

`GetInt` panics rather than returning an error because the `validateValueMatchesKeyType` hook guarantees that any stored value is valid for its declared type. If `GetInt` returns an error at call time, every caller needs error handling for a condition that should never occur — that's noise that hides real errors. If a bad value somehow reaches the DB (e.g. manual SQL edit), a panic surfaces the corruption immediately rather than silently returning zero.

---

## 45.6 The `RegisterKey` Pattern

Business modules declare their settings at init time:

```go
// module/finance/finance.go
package finance

import (
    "awo.so/internal/platform/settings"
    "awo.so/framework/definition"
)

func init() {
    definition.Register(&InvoiceDef)

    settings.RegisterKey(settings.Key{
        Key:            "finance.invoice_prefix",
        Label:          "Invoice Number Prefix",
        Description:    "Prefix for auto-generated invoice numbers (e.g. 'INV', 'ACME-INV')",
        ValueType:      "string",
        Default:        "INV",
        Module:         "Finance",
        TenantEditable: true,
    })

    settings.RegisterKey(settings.Key{
        Key:            "finance.vat_rate",
        Label:          "Standard VAT Rate (%)",
        Description:    "Default VAT rate applied to taxable line items",
        ValueType:      "int",
        Default:        "16", // Kenya standard VAT rate
        Module:         "Finance",
        TenantEditable: true,
    })

    settings.RegisterKey(settings.Key{
        Key:            "finance.etims_enabled",
        Label:          "KRA eTIMS Integration Enabled",
        ValueType:      "bool",
        Default:        "false",
        Module:         "Finance",
        TenantEditable: false, // set by platform team per tenant activation
    })
}
```

`RegisterKey` is idempotent — if the key already exists in the database with the same type, it does nothing. On first deploy, it inserts the key row with the declared default. This means adding a new setting to a module requires no manual database work.

---

## 45.7 SDUI Auto-Generated Settings Pages

The SDUI layer generates a settings page per module from all `config_keys` where `module = 'Finance'` and `tenant_editable = true`. Each key maps to a form control based on `value_type`:

| value_type | AMIS control |
|---|---|
| `string` | `input-text` |
| `int` | `input-number` with `precision: 0` |
| `bool` | `switch` |
| `json` | `json-editor` |

Tenant admins see a settings page that looks like a regular form — no knowledge of keys, types, or the underlying two-table structure required.

---

## 45.8 Kenya-Specific Settings Reference

These keys are registered by the Finance and HR modules for Kenya compliance:

| Key | Default | Notes |
|---|---|---|
| `finance.vat_rate` | `16` | Standard VAT per KRA |
| `finance.vat_exempt_categories` | `[]` | JSON array of product category codes |
| `finance.etims_enabled` | `false` | KRA eTIMS electronic invoicing |
| `finance.etims_pin` | `""` | Trader KRA PIN for eTIMS |
| `hr.nhif_rate_employee` | `150` | NHIF employee contribution (KES) |
| `hr.nssf_tier1_employee` | `72` | NSSF Tier I employee (KES) |
| `hr.nssf_tier2_rate` | `6` | NSSF Tier II rate (%) |
| `hr.paye_enabled` | `true` | PAYE deduction from payroll |
| `hr.personal_relief` | `2400` | Monthly personal relief (KES) |

These defaults match the 2025 KRA and NSSF regulations. Tenants with different rates (NGOs, Export Processing Zones) override them via their settings page.

---

## 45.9 Migration

```sql
-- internal/platform/settings/migrations/20240101000001_create_config_keys.up.sql

CREATE TABLE config_keys (
    id             uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at     timestamptz NOT NULL DEFAULT NOW(),
    updated_at     timestamptz NOT NULL DEFAULT NOW(),
    key            text NOT NULL UNIQUE,
    label          text NOT NULL,
    description    text,
    value_type     text NOT NULL CHECK (value_type IN ('string','int','bool','json')),
    default_value  text,
    module         text NOT NULL,
    tenant_editable boolean NOT NULL DEFAULT true
);

-- No RLS: config_keys is a global catalogue (no tenant_id column).

-- internal/platform/settings/migrations/20240101000002_create_config_values.up.sql

CREATE TABLE config_values (
    id         uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id  uuid NOT NULL REFERENCES tenants(id),
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    key        text NOT NULL REFERENCES config_keys(key),
    value      text NOT NULL,
    UNIQUE (tenant_id, key)
);

CREATE INDEX config_values_tenant_idx ON config_values (tenant_id);

ALTER TABLE config_values ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON config_values
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

The `UNIQUE (tenant_id, key)` constraint ensures at most one value per tenant per key. An attempt to create a second value for the same pair returns `23505` → `persistence.ErrConflict` → HTTP 409, which the UI handles as "update instead of create."

---

## 45.10 FAQ

**Q: What happens to existing tenants when I add a new setting key?**
A: Nothing — they inherit the system default declared in `RegisterKey`. No migration, no backfill needed. The cascade means new tenants and existing tenants both get the default until they override it.

**Q: Can a tenant admin delete a config value?**
A: Yes. Deleting a `config_value` row reverts the tenant to the system default. The `invalidateSettingsCache` hook clears the Redis entry on delete as well.

**Q: How do I read a setting inside a hook or service?**
A: Inject `*settings.Service` and call `svc.Get(ctx, tenantID, "finance.vat_rate")`. Do not read from `config_values` directly — bypass the cache and the cascade resolution.

**Q: Can settings be scoped to an org unit (per-branch)?**
A: Not in the current implementation. `config_values` is `ScopeLevelTenant`. Adding unit-level overrides would require a `org_unit_id` column on `config_values` and a fourth resolution level. This is a planned enhancement.
