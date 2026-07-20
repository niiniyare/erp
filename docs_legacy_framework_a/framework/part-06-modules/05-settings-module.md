> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Platform Module: Settings & Configuration"
part: "Part VI — Platform Entities"
chapter: 45
section: "platform-settings-module"
related:
  - "[Chapter 44: Feature Flags](./platform-feature-flags-module.md)"
  - "[Chapter 42: Tenant & Organisation](./platform-tenant-module.md)"
  - "[Chapter 41: Platform Module Overview](./platform-module-overview.md)"
---

# Chapter 45 — Settings & Configuration Module

> **Primary source for:** the `config_keys` and `config_values` tables, three-tier configuration resolution, module key registration, the SDUI auto-generated settings pages, and Kenya-specific configuration.
>
> **Audience:** developers adding configurable behaviour to a module, business stakeholders who want to know how a tenant customises their ERP, and anyone asking "where does this value come from?"

---

## 45.1 What Are Settings?

Settings are the configuration knobs that control *how* the system behaves for a specific business. If feature flags answer "is this feature ON?", settings answer "how does this feature work?".

Examples:

| Setting Key | What it controls | Example value |
|---|---|---|
| `finance.invoice_prefix` | Prefix for auto-generated invoice numbers | `"ACM-"` |
| `finance.vat_rate` | Default VAT percentage applied to invoices | `"16"` |
| `finance.inventory_valuation` | Stock cost method | `"FIFO"` |
| `hr.overtime_multiplier` | Pay rate for hours over 8/day | `"1.5"` |
| `locale.timezone` | Default timezone for date display | `"Africa/Nairobi"` |
| `locale.date_format` | How dates are displayed in UI | `"DD/MM/YYYY"` |
| `forecourt.tank_variance_threshold` | % variance before NEMA alert triggers | `"0.5"` |

### Settings vs Feature Flags

| | Settings | Feature Flags (Ch. 44) |
|---|---|---|
| **Question** | "What value should I use?" | "Should this run at all?" |
| **Changed by** | Tenant admin (usually) | Platform admin |
| **Scope** | Hierarchical (system/tenant/branch) | Tenant or user override |
| **Example** | `finance.invoice_prefix = "ACM-"` | `finance.etims_integration = true` |

The two systems are complementary. A module might use a flag to gate an integration on/off, then use settings to supply the integration's credentials and configuration.

---

## 45.2 Three-Level Hierarchy

Settings resolve through three levels. Most specific wins.

```
Entity Level  (org node / branch override)   ← most specific
      │
      ▼ if not set
Tenant Level  (business-wide override)
      │
      ▼ if not set
System Level  (platform default shipped with the module)   ← least specific
```

**Why three levels?**

Two levels (system/tenant) covers 90% of cases. The third level (entity/org-node) is needed for large enterprises where different branches have genuinely different configuration. A petrol station chain might have:

- System default: `forecourt.tank_variance_threshold = 0.5%`
- Tenant override for all stations: `0.3%` (their internal standard)
- Nairobi Airport station override: `0.8%` (aviation fuel tanks have looser tolerance)

Without the entity level, the Nairobi Airport station would either get the wrong threshold or require a code change.

---

## 45.3 Directory Layout

```
framework/platform/settings/
├── settings.go    ← init() — registers ConfigKeyDef and ConfigValueDef
├── definition.go  ← EntityDefinition declarations
├── policy.go      ← policy helpers
├── hooks.go       ← validateValueMatchesKeyType, assertKeyIsTenantEditable,
│                     invalidateSettingsCache
├── service.go     ← Service: Get(), GetInt(), GetBool(), Set()
├── register.go    ← RegisterKey() called by module init() functions
└── migrations/
    └── 20240101000030_create_settings.up.sql
```

---

## 45.4 ConfigKey EntityDefinition (the Catalogue)

The `config_keys` table is the **catalogue** of every known configuration key. It is global — the key definitions are the same across all tenants. Module teams register their keys here.

```go
// framework/platform/settings/definition.go
package settings

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

var ConfigKeyDef = definition.EntityDefinition{
    Name:     "config_key",
    Label:    "Configuration Key",
    Module:   "Platform",
    Table:    "config_keys",

    // GLOBAL: the key catalogue is the same for every tenant.
    // Every deployment that includes the Finance module has 'finance.invoice_prefix'
    // as a key, regardless of whether any tenant has set an override.
    OrgScope: org.ScopeLevelGlobal,

    // Configuration key definitions don't need auditing.
    // They change only on deployment (when modules register new keys).
    Audited: false,

    Fields: []*definition.FieldDef{
        {
            Name:     "module",
            Type:     definition.FieldTypeData,
            Label:    "Module",
            Required: true,
            Description: "Owning module name. Used for admin UI grouping. " +
                "Examples: 'finance', 'hr', 'forecourt', 'locale'.",
        },
        {
            Name:     "key",
            Type:     definition.FieldTypeData,
            Label:    "Key",
            Required: true,
            Description: "Stable dot-namespaced identifier. " +
                "Format: '{module}.{name}'. Never change after creation — " +
                "existing code references break if the key is renamed.",
        },
        {Name: "label",       Type: definition.FieldTypeData,      Label: "Display Label", Required: true},
        {Name: "description", Type: definition.FieldTypeSmallText, Label: "Description",
            Description: "Shown in the admin settings UI to explain what this setting controls."},
        {
            Name:    "value_type",
            Type:    definition.FieldTypeSelect,
            Label:   "Value Type",
            Options: []string{"string", "integer", "boolean", "json", "date"},
            Description: "Used to: (a) validate submitted values, " +
                "(b) generate the correct SDUI input control, " +
                "(c) coerce the stored text value on read.",
        },
        {
            Name:        "default_value",
            Type:        definition.FieldTypeData,
            Label:       "System Default",
            Description: "Returned when no tenant or entity override exists. " +
                "All values stored as text; coerced to value_type on read.",
        },
        {
            Name:    "is_sensitive",
            Type:    definition.FieldTypeBool,
            Label:   "Sensitive",
            Default: "false",
            Description: "If true: value is masked in admin UI (shown as ••••••). " +
                "Not excluded from the DB — only from display. " +
                "Use for API keys, credentials, secret tokens.",
        },
        {
            Name:    "is_tenant_editable",
            Type:    definition.FieldTypeBool,
            Label:   "Tenant Editable",
            Default: "true",
            Description: "If false: only platform admins (system callers) can set this value. " +
                "Tenant admins see it in UI but it renders as read-only. " +
                "Use for security-critical settings (max API rate, allowed IPs).",
        },
    },

    Policies: []definition.PolicyDef{
        // System callers manage the catalogue (module init() RegisterKey calls).
        definition.Policy(definition.OpAll, definition.AllowSystem),

        // All authenticated users READ the catalogue.
        // Reason: the Settings admin UI reads config_keys to know which settings
        // exist and how to render them. Restricting reads would break the UI.
        definition.Policy(definition.OpRead, allowTenantViewer),

        // Nobody else can create/modify/delete config keys via the API.
        // Key registration is a developer-only action via RegisterKey().
        definition.Policy(definition.OpAll, definition.DenyAll),
    },
}
```

---

## 45.5 ConfigValue EntityDefinition (the Per-Tenant Values)

```go
var ConfigValueDef = definition.EntityDefinition{
    Name:     "config_value",
    Label:    "Configuration Value",
    Module:   "Platform",
    Table:    "config_values",
    OrgScope: org.ScopeLevelTenant,
    Audited:  true, // every settings change is audited
    SoftDelete: false,

    Fields: []*definition.FieldDef{
        {
            Name:         "config_key",
            Type:         definition.FieldTypeLink,
            Label:        "Setting",
            TargetEntity: "config_key",
            Required:     true,
        },
        {
            Name:     "scope",
            Type:     definition.FieldTypeSelect,
            Label:    "Scope Level",
            Options:  []string{"system", "tenant", "entity"},
            Required: true,
            Description: "system: platform-wide value (tenant_id ignored in resolution). " +
                "tenant: applies to all org units in this tenant. " +
                "entity: applies to a specific org node (and overrides tenant-level).",
        },
        {
            Name:         "entity_id",
            Type:         definition.FieldTypeLink,
            Label:        "Org Node",
            TargetEntity: "org_node",
            Description:  "Set only when scope=entity. Which branch/division this applies to.",
        },
        {
            Name:     "value",
            Type:     definition.FieldTypeData,
            Label:    "Value",
            Required: true,
            Description: "Always stored as text. The Settings service coerces to the " +
                "config_key's value_type on read (integer → int64, boolean → bool, etc.).",
        },
        {Name: "set_by", Type: definition.FieldTypeLink,     Label: "Set By", TargetEntity: "user"},
        {Name: "set_at", Type: definition.FieldTypeDateTime, Label: "Set At"},
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll, definition.AllowSystem),

        // Tenant admins can set tenant-level and entity-level values.
        // The assertKeyIsTenantEditable hook blocks changes to platform-only settings.
        definition.Policy(definition.OpAll, requireRole("admin")),

        // All authenticated users can READ configuration values.
        // Reason: module code reads settings on every request to drive behaviour.
        // Invoice handlers read 'finance.invoice_prefix'. Payroll reads 'hr.overtime_multiplier'.
        definition.Policy(definition.OpRead, allowTenantViewer),

        definition.Policy(definition.OpAll, definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            // Validates that the submitted value matches the config_key's value_type.
            // Rejects "abc" for an integer key. Rejects "maybe" for a boolean key.
            Name: "validate_value_type",
            Ops:  definition.OpCreate | definition.OpUpdate,
            When: definition.HookBefore,
            Fn:   validateValueMatchesKeyType,
        },
        {
            // Blocks tenant admins from changing settings where is_tenant_editable=false.
            // System callers (AllowSystem) bypass this hook.
            Name: "check_tenant_editable",
            Ops:  definition.OpCreate | definition.OpUpdate,
            When: definition.HookBefore,
            Fn:   assertKeyIsTenantEditable,
        },
        {
            // After any settings change, clears the Redis cache for this tenant's settings.
            // New reads will re-query PostgreSQL and rebuild the cache.
            // Also publishes a Redis pub/sub event so other server instances
            // invalidate their in-process cache immediately.
            Name: "invalidate_settings_cache",
            Ops:  definition.OpCreate | definition.OpUpdate | definition.OpDelete,
            When: definition.HookAfter,
            Fn:   invalidateSettingsCache,
        },
    },
}
```

---

## 45.6 Hook Implementations

```go
// framework/platform/settings/hooks.go
package settings

import (
    "context"
    "fmt"
    "strconv"

    "awo.so/framework/definition"
)

// validateValueMatchesKeyType rejects values that don't match the key's declared type.
func validateValueMatchesKeyType(ctx context.Context, v definition.ViewerContext, op definition.Op, rec definition.MutableRecord) error {
    keyName, _ := rec.Get("config_key")
    value, _    := rec.Get("value")
    if keyName == nil || value == nil {
        return nil
    }

    // Look up the config key's value_type
    key, err := configKeyStore.FindBy(ctx, map[string]any{"key": keyName})
    if err != nil {
        return fmt.Errorf("unknown config key %q: %w", keyName, err)
    }
    valueType, _ := key.Get("value_type")

    strVal := value.(string)
    switch valueType.(string) {
    case "integer":
        if _, err := strconv.ParseInt(strVal, 10, 64); err != nil {
            return fmt.Errorf("setting %q requires an integer value, got %q", keyName, strVal)
        }
    case "boolean":
        if strVal != "true" && strVal != "false" {
            return fmt.Errorf("setting %q requires 'true' or 'false', got %q", keyName, strVal)
        }
    case "date":
        if _, err := time.Parse("2006-01-02", strVal); err != nil {
            return fmt.Errorf("setting %q requires YYYY-MM-DD date, got %q", keyName, strVal)
        }
    // "string" and "json" accept any text value.
    }
    return nil
}

// assertKeyIsTenantEditable blocks writes to platform-only settings.
func assertKeyIsTenantEditable(ctx context.Context, v definition.ViewerContext, op definition.Op, rec definition.MutableRecord) error {
    if v.IsSystem() {
        return nil // system callers always pass
    }
    keyName, _ := rec.Get("config_key")
    key, err := configKeyStore.FindBy(ctx, map[string]any{"key": keyName})
    if err != nil {
        return err
    }
    editable, _ := key.Get("is_tenant_editable")
    if !editable.(bool) {
        return fmt.Errorf("setting %q can only be changed by platform administrators", keyName)
    }
    return nil
}
```

---

## 45.7 The Settings Service

```go
// framework/platform/settings/service.go
package settings

import (
    "context"
    "strconv"
    "time"

    "github.com/google/uuid"
    "awo.so/framework/contextutil"
    "awo.so/framework/persistence/pgstore"
)

type Service struct {
    values     pgstore.EntityStore // ConfigValueDef
    keys       pgstore.EntityStore // ConfigKeyDef
    redis      RedisClient
}

// Get resolves a setting through the three-tier cascade.
// entityID: pass the org_unit_id the viewer is operating as, or nil for tenant-level.
func (s *Service) Get(ctx context.Context, key string, entityID *uuid.UUID) (string, error) {
    tenantID := contextutil.TenantID(ctx)

    // ── Fast path: Redis cache ───────────────────────────────────────────
    // Cache key: "settings:{tenant_id}" as a Redis hash of {config_key → value}.
    // Built on first miss; cleared by invalidateSettingsCache hook.
    if val, err := s.redis.HGet(ctx, "settings:"+tenantID, key); err == nil {
        return val, nil
    }

    // ── Slow path: DB three-tier cascade ─────────────────────────────────

    // Tier 1: entity-level (most specific)
    if entityID != nil {
        if val, err := s.lookup(ctx, key, "entity", tenantID, entityID); err == nil {
            s.cache(ctx, tenantID, key, val)
            return val, nil
        }
    }

    // Tier 2: tenant-level
    if val, err := s.lookup(ctx, key, "tenant", tenantID, nil); err == nil {
        s.cache(ctx, tenantID, key, val)
        return val, nil
    }

    // Tier 3: system default from config_keys catalogue
    val, err := s.systemDefault(ctx, key)
    if err == nil {
        s.cache(ctx, tenantID, key, val)
    }
    return val, err
}

// GetInt is a typed convenience wrapper.
func (s *Service) GetInt(ctx context.Context, key string, entityID *uuid.UUID) (int64, error) {
    str, err := s.Get(ctx, key, entityID)
    if err != nil {
        return 0, err
    }
    return strconv.ParseInt(str, 10, 64)
}

// GetBool is a typed convenience wrapper.
func (s *Service) GetBool(ctx context.Context, key string, entityID *uuid.UUID) (bool, error) {
    str, err := s.Get(ctx, key, entityID)
    if err != nil {
        return false, err
    }
    return str == "true", nil
}

func (s *Service) lookup(ctx context.Context, key, scope, tenantID string, entityID *uuid.UUID) (string, error) {
    filter := map[string]any{"config_key": key, "scope": scope}
    if entityID != nil {
        filter["entity_id"] = *entityID
    } else {
        filter["entity_id"] = nil
    }

    result, err := s.values.List(ctx, pgstore.ListOptions{Filter: filter, Limit: 1})
    if err != nil || len(result.Items) == 0 {
        return "", pgstore.ErrNotFound
    }
    val, _ := result.Items[0].Get("value")
    return val.(string), nil
}

func (s *Service) systemDefault(ctx context.Context, key string) (string, error) {
    result, err := s.keys.List(ctx, pgstore.ListOptions{
        Filter: map[string]any{"key": key},
        Limit:  1,
    })
    if err != nil || len(result.Items) == 0 {
        return "", fmt.Errorf("unknown config key %q", key)
    }
    val, _ := result.Items[0].Get("default_value")
    if val == nil {
        return "", nil
    }
    return val.(string), nil
}

func (s *Service) cache(ctx context.Context, tenantID, key, value string) {
    s.redis.HSet(ctx, "settings:"+tenantID, key, value)
    s.redis.Expire(ctx, "settings:"+tenantID, 10*time.Minute)
}
```

---

## 45.8 Module Developer Integration

Module developers register their settings keys in `init()`, then call the service at runtime:

```go
// In awo.so/module/finance — init()
package finance

import "awo.so/framework/platform/settings"

func init() {
    // Register all settings this module uses.
    // Called once at startup; upserts rows in config_keys.
    settings.RegisterKey(settings.KeyDef{
        Module:           "finance",
        Key:              "finance.invoice_prefix",
        Label:            "Invoice Number Prefix",
        Description:      "Prepended to every auto-generated invoice number.",
        ValueType:        "string",
        DefaultValue:     "INV-",
        IsTenantEditable: true,
    })

    settings.RegisterKey(settings.KeyDef{
        Module:           "finance",
        Key:              "finance.vat_rate",
        Label:            "Default VAT Rate (%)",
        Description:      "Applied to new invoices unless overridden per item.",
        ValueType:        "integer",
        DefaultValue:     "16",
        IsTenantEditable: true,
    })

    settings.RegisterKey(settings.KeyDef{
        Module:           "finance",
        Key:              "finance.inventory_valuation",
        Label:            "Inventory Valuation Method",
        ValueType:        "string",
        DefaultValue:     "FIFO",
        IsTenantEditable: true,
    })

    settings.RegisterKey(settings.KeyDef{
        Module:           "finance",
        Key:              "finance.etims_api_key",
        Label:            "KRA eTIMS API Key",
        Description:      "Provided by Kenya Revenue Authority after device registration.",
        ValueType:        "string",
        IsSensitive:      true,
        IsTenantEditable: true,
    })
}
```

Then in a handler:

```go
// In the Finance invoice service:
func (s *InvoiceService) nextInvoiceNumber(ctx context.Context) (string, error) {
    prefix, err := s.settings.Get(ctx, "finance.invoice_prefix", nil)
    if err != nil {
        prefix = "INV-" // fallback; should not happen if key is registered
    }
    seq, err := s.db.NextSequenceValue(ctx, "invoice_number")
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%s%06d", prefix, seq), nil
}
```

---

## 45.9 SDUI Auto-Generated Settings Pages

The framework automatically generates a Settings admin page for every module. It reads `config_keys WHERE module = '{name}'` and renders the appropriate input control per `value_type`:

| value_type | SDUI control |
|---|---|
| `string` | InputText |
| `integer` | InputNumber |
| `boolean` | Switch (toggle) |
| `date` | DatePicker |
| `json` | CodeEditor (JSON syntax) |

Sensitive settings (`is_sensitive=true`) render as InputPassword (masked). Non-editable settings (`is_tenant_editable=false`) render as disabled fields with a "Managed by platform" tooltip.

Module developers get this settings page for free — no UI code needed.

---

## 45.10 Migration

```sql
-- framework/platform/settings/migrations/20240101000030_create_settings.up.sql

-- ── Config Keys Catalogue ───────────────────────────────────────────────────
-- Global: no tenant_id, no RLS. Key definitions are platform-wide.
CREATE TABLE config_keys (
    id                  uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
    module              text    NOT NULL,
    key                 text    NOT NULL UNIQUE,
    label               text    NOT NULL,
    description         text,
    value_type          text    NOT NULL DEFAULT 'string'
                                CHECK (value_type IN ('string','integer','boolean','json','date')),
    default_value       text,
    is_sensitive        boolean NOT NULL DEFAULT false,
    is_tenant_editable  boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX config_keys_module_idx ON config_keys (module);

-- Seed platform-level locale defaults (module = 'locale').
-- Business modules seed their own keys via RegisterKey() at startup.
INSERT INTO config_keys (module, key, label, value_type, default_value) VALUES
    ('locale', 'locale.timezone',    'Default Timezone',    'string',  'Africa/Nairobi'),
    ('locale', 'locale.currency',    'Base Currency',       'string',  'KES'),
    ('locale', 'locale.date_format', 'Date Display Format', 'string',  'DD/MM/YYYY'),
    ('locale', 'locale.language',    'Language',            'string',  'en');


-- ── Config Values ───────────────────────────────────────────────────────────
CREATE TABLE config_values (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    config_key  text        NOT NULL REFERENCES config_keys(key) ON DELETE CASCADE,
    scope       text        NOT NULL DEFAULT 'tenant'
                            CHECK (scope IN ('system', 'tenant', 'entity')),
    -- NULL when scope = 'tenant' or 'system'.
    -- Set to an org_nodes.id when scope = 'entity'.
    entity_id   uuid        REFERENCES org_nodes(id),
    value       text        NOT NULL,
    set_by      uuid        REFERENCES users(id),
    set_at      timestamptz NOT NULL DEFAULT now(),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),

    -- Enforce one value per (tenant, key, scope, entity).
    -- NULLS NOT DISTINCT: two NULLs for entity_id are equal —
    -- prevents duplicate tenant-level overrides for the same key.
    UNIQUE NULLS NOT DISTINCT (tenant_id, config_key, scope, entity_id)
);

ALTER TABLE config_values ENABLE ROW LEVEL SECURITY;
CREATE POLICY config_values_tenant_isolation ON config_values
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Index for the three-tier resolution cascade.
-- Query: WHERE tenant_id=? AND config_key=? AND scope=? AND entity_id IS [NOT] NULL
CREATE INDEX config_values_lookup_idx ON config_values
    (tenant_id, config_key, scope, entity_id NULLS FIRST);

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON config_values
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

---

## 45.11 Kenya-Specific Configuration

The Settings module is the home for Kenya-specific regulatory configuration. These are managed through the standard settings system — not hardcoded — because rates and thresholds change with Finance Acts and regulatory updates.

| Key | Module | Description | Default |
|---|---|---|---|
| `tax.kra_etims_env` | `tax` | eTIMS environment: `"sandbox"` or `"production"` | `"sandbox"` |
| `tax.vat_rate` | `finance` | Standard VAT rate per Finance Act | `"16"` |
| `tax.withholding_rate` | `finance` | Withholding tax rate | `"5"` |
| `hr.nhif_employee_rate` | `hr` | NHIF employee contribution % | varies by income |
| `hr.nssf_tier1_limit` | `hr` | NSSF Tier I upper earning limit (KES) | `"7000"` |
| `hr.nssf_tier2_limit` | `hr` | NSSF Tier II upper earning limit (KES) | `"36000"` |
| `hr.paye_table` | `hr` | PAYE tax bands (JSON array of rate/threshold pairs) | JSON |
| `forecourt.nema_variance_limit` | `forecourt` | Fuel tank variance % before NEMA report | `"0.5"` |

When Kenya Revenue Authority updates PAYE bands at the start of a new fiscal year, the HR module admin updates `hr.paye_table` in the settings UI. No code deployment needed. The change propagates to all tenants on their next payroll run.

---

## 45.12 Common Mistakes

**"Should I store a lookup table (e.g., tax bands) in settings?"**

Only if it changes infrequently and is the same across all records (a global rate). If tax bands are date-ranged or complex, create a proper entity table (`tax_bands` with `effective_from` / `effective_to` columns) — that data belongs in a queryable, auditable table, not a JSON blob in settings.

**"Can I read settings with a direct SQL query?"**

Never. Always use `settings.Service.Get()`. Direct SQL bypasses the cache, skips the three-tier resolution (you'd only get the direct row, missing the cascade), and skips type coercion. The service is the contract; the DB is the implementation.

**"Should I add a fallback default in my code if the setting is missing?"**

No. Register a `DefaultValue` in `RegisterKey`. The service returns it automatically. Hardcoded fallbacks in service code mean two places define the same default — they drift apart. One source of truth: the `config_keys` row.

**"What if I need to read settings in a Temporal activity?"**

Temporal activities run without a user session context. Use `settings.ServiceWithSystem(ctx, tenantID)` — a variant of `Get()` that accepts an explicit `tenantID` rather than reading from the request context. The activity still goes through the three-tier cascade and Redis cache.
