> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Platform Module: Plugin & Module Registry"
part: "Part VI — Platform Entities"
chapter: 48
section: "platform-module-registry"
related:
  - "[Chapter 42: Tenant & Organisation](./platform-tenant-module.md)"
  - "[Chapter 44: Feature Flags](./platform-feature-flags-module.md)"
  - "[Chapter 45: Settings](./platform-settings-module.md)"
  - "[Chapter 41: Platform Module Overview](./platform-module-overview.md)"
---

# Chapter 48 — Plugin & Module Registry

> **Primary source for:** the `modules` and `tenant_modules` tables, how business modules self-register, the two levels of module enablement (deployment vs tenant), how the SDUI sidebar is driven by active modules, trial module management, and module versioning.
>
> **Audience:** platform engineers deploying a new module, business stakeholders who want to understand the "module marketplace" concept, and any developer whose module needs to interact with the registry.

---

## 48.1 What Is the Module Registry?

The Awo ERP platform ships with a core framework and a collection of **business modules**: Finance, HR, Inventory, CRM, Forecourt, Payroll, etc. Not every deployment needs every module. A fuel station chain needs Forecourt but not a complex CRM. A hospital needs a custom module for patient billing but not Forecourt at all.

The **Module Registry** tracks two things:

1. **Which modules are installed** on this Awo server binary (the `modules` table — global).
2. **Which modules each tenant has activated** (the `tenant_modules` table — per-tenant).

Think of it as a two-layer system:

```
Layer 1 — Deployment level (modules table)
  The server binary includes: Finance, HR, Inventory, CRM, Forecourt
  "These modules are available on this server."

        ┌──────────────────┐
        │ Acme Petroleum   │ active: Finance, Forecourt
        ├──────────────────┤
        │ Beta Hospital    │ active: Finance, HR, [custom billing module]
        ├──────────────────┤
        │ Gamma Wholesale  │ active: Finance, Inventory, CRM
        └──────────────────┘

Layer 2 — Tenant level (tenant_modules table)
  Each tenant activates the modules they have subscribed to.
  "This tenant can use these modules."
```

---

## 48.2 The Analogy: Electrical Circuit Panel

Think of the modules as electrical circuits in an office building:

- The **electrician** (platform engineer) wires the circuits into the building at construction time. The circuits exist whether tenants use them or not.
- Each **tenant's floor** has its own circuit breaker panel. The tenant (via their admin) decides which circuits to turn on.
- A circuit that isn't turned on for a tenant simply doesn't appear — no navigation menu, no API routes accessible, no data.

The key insight: the circuit (module code) exists on the server regardless. The breaker (tenant_modules activation) determines visibility and access.

---

## 48.3 Directory Layout

```
framework/platform/registry/
├── registry.go    ← init() — registers ModuleDef and TenantModuleDef
│                    RegisterModule() called by business module init() functions
├── definition.go  ← ModuleDef, TenantModuleDef
├── policy.go      ← policy helpers
├── hooks.go       ← reloadModuleNavigation
├── service.go     ← RegistryService: IsActive(), ActiveModules(), Activate()
└── migrations/
    └── 20240101000060_create_module_registry.up.sql
```

---

## 48.4 ModuleDef EntityDefinition (Deployment-Level Catalogue)

The `modules` table is the global catalogue of every module installed on this server. It is populated at startup by each module's `init()` function calling `registry.RegisterModule()`.

```go
// framework/platform/registry/definition.go
package registry

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

var ModuleDef = definition.EntityDefinition{
    Name:     "module",
    Label:    "Module",
    Module:   "Platform",
    Table:    "modules",

    // GLOBAL: the module catalogue is the same for all tenants.
    // What's installed on the server is a deployment decision, not a per-tenant one.
    OrgScope: org.ScopeLevelGlobal,

    // Module catalogue rows don't need auditing.
    // They change only on server deployments, which are tracked separately.
    Audited: false,

    Fields: []*definition.FieldDef{
        {
            Name:     "name",
            Type:     definition.FieldTypeData,
            Label:    "Module Name",
            Required: true,
            Description: "Stable identifier. Used in feature flag keys, config keys, " +
                "and navigation generation. Never rename after first deployment. " +
                "Examples: 'finance', 'hr', 'inventory', 'forecourt'.",
        },
        {
            Name:     "label",
            Type:     definition.FieldTypeData,
            Label:    "Display Name",
            Required: true,
            Description: "Human-readable name shown in navigation and admin UI.",
        },
        {
            Name:     "version",
            Type:     definition.FieldTypeData,
            Label:    "Version",
            Required: true,
            Description: "Semantic version of this module. Updated by RegisterModule() " +
                "on server startup. Shown in admin panel for support purposes.",
        },
        {Name: "description", Type: definition.FieldTypeSmallText, Label: "Description"},
        {Name: "repo_url",    Type: definition.FieldTypeData,      Label: "Repository URL"},
        {
            Name:    "status",
            Type:    definition.FieldTypeSelect,
            Label:   "Stability",
            Options: []string{"stable", "beta", "deprecated"},
            Description: "stable: production-ready. beta: early access, may have breaking changes. " +
                "deprecated: will be removed in a future release.",
        },
    },

    Policies: []definition.PolicyDef{
        // System callers write (module init() calls registry.RegisterModule()).
        definition.Policy(definition.OpAll, definition.AllowSystem),

        // All authenticated users read the module catalogue.
        // Needed for the "available modules" list in the admin activation UI.
        definition.Policy(definition.OpRead, allowTenantViewer),

        // Nobody can create/delete/edit modules via the user-facing API.
        // Module registration is a developer action, not an operator action.
        definition.Policy(definition.OpAll, definition.DenyAll),
    },
}
```

---

## 48.5 TenantModuleDef EntityDefinition (Per-Tenant Activation)

```go
var TenantModuleDef = definition.EntityDefinition{
    Name:     "tenant_module",
    Label:    "Module Activation",
    Module:   "Platform",
    Table:    "tenant_modules",

    // TENANT-SCOPED: each activation row belongs to one tenant.
    // RLS ensures Tenant A cannot see or modify Tenant B's module activations.
    OrgScope: org.ScopeLevelTenant,
    Audited:  true, // every activation/deactivation is audited

    Fields: []*definition.FieldDef{
        {
            Name:         "module_name",
            Type:         definition.FieldTypeLink,
            Label:        "Module",
            TargetEntity: "module",
            Required:     true,
        },
        {
            Name:     "status",
            Type:     definition.FieldTypeSelect,
            Label:    "Status",
            Options:  []string{"active", "suspended", "trial"},
            Required: true,
            Description: "active: tenant has full access. " +
                "trial: time-limited access; reverts to suspended when trial_ends_at passes. " +
                "suspended: access removed (payment failure, plan downgrade).",
        },
        {Name: "activated_at", Type: definition.FieldTypeDateTime, Label: "Activated At"},
        {
            Name:        "trial_ends_at",
            Type:        definition.FieldTypeDateTime,
            Label:       "Trial Ends",
            Description: "Only relevant when status=trial. A daily Temporal schedule " +
                "checks this and moves status to 'suspended' when expired.",
        },
        {
            Name:        "config",
            Type:        definition.FieldTypeJSON,
            Label:       "Module Configuration",
            Description: "Module-specific configuration JSON. Structure defined by each module. " +
                "Finance might store {etims_device_serial, etims_taxpayer_pin}. " +
                "Forecourt might store {pump_protocol, tank_monitoring_enabled}. " +
                "Prefer the Settings module for user-editable config; use this for " +
                "module-wide tenant config that is set once at activation.",
        },
    },

    Policies: []definition.PolicyDef{
        definition.Policy(definition.OpAll, definition.AllowSystem),

        // Platform admins activate/deactivate modules for tenants.
        // Tenant admins cannot activate modules — only a platform admin (who controls billing)
        // should enable access to a paid module.
        definition.Policy(definition.OpAll, requireRole("platform_admin")),

        // All authenticated tenant users can READ which modules are active.
        // The SDUI navigation builder needs this to know which sections to show.
        definition.Policy(definition.OpRead, allowTenantViewer),

        definition.Policy(definition.OpAll, definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            // After any activation status change, rebuilds the SDUI navigation
            // sidebar for this tenant. A newly activated Finance module immediately
            // appears in the nav; a deactivated Forecourt module immediately disappears.
            Name: "reload_module_navigation",
            Ops:  definition.OpCreate | definition.OpUpdate,
            When: definition.HookAfter,
            Fn:   reloadModuleNavigation,
        },
    },
}
```

---

## 48.6 Module Self-Registration — The Full Pattern

Business modules register themselves in their package's `init()` function. Here is the complete registration for the Finance module:

```go
// In awo.so/module/finance — init()
// This file is the single entry point for everything the Finance module
// contributes to the framework.
package finance

import (
    "awo.so/framework/definition"
    "awo.so/framework/platform/registry"
    "awo.so/framework/platform/settings"
    "awo.so/framework/platform/featureflag"
)

func init() {
    // ── 1. Register all EntityDefinitions ─────────────────────────────────
    // The framework mounts REST CRUD routes and SDUI pages for each.
    definition.Register(&AccountDef)
    definition.Register(&JournalEntryDef)
    definition.Register(&InvoiceDef)
    definition.Register(&InvoiceItemDef)
    definition.Register(&PaymentDef)
    definition.Register(&TaxRateDef)
    definition.Register(&CostCentreDef)

    // ── 2. Register the module with the platform registry ─────────────────
    // Upserts a row in the modules table on server startup.
    // Version is updated on every startup from the module's version constant.
    registry.RegisterModule(registry.ModuleInfo{
        Name:        "finance",
        Label:       "Finance",
        Version:     FinanceModuleVersion, // constant: "1.4.2"
        Description: "General ledger, chart of accounts, AP/AR, invoicing, payments, tax.",
        RepoURL:     "https://github.com/awoerp/module-finance",
        Status:      "stable",
    })

    // ── 3. Register module-owned configuration keys ────────────────────────
    // These appear in the Settings admin UI under "Finance" section.
    settings.RegisterKey(settings.KeyDef{
        Module: "finance", Key: "finance.invoice_prefix",
        Label: "Invoice Number Prefix", ValueType: "string", DefaultValue: "INV-",
        IsTenantEditable: true,
    })
    settings.RegisterKey(settings.KeyDef{
        Module: "finance", Key: "finance.vat_rate",
        Label: "Default VAT Rate (%)", ValueType: "integer", DefaultValue: "16",
        IsTenantEditable: true,
    })
    settings.RegisterKey(settings.KeyDef{
        Module: "finance", Key: "finance.etims_api_key",
        Label: "KRA eTIMS API Key", ValueType: "string",
        IsSensitive: true, IsTenantEditable: true,
    })

    // ── 4. Register module-owned feature flags ─────────────────────────────
    featureflag.RegisterFlag(featureflag.FlagDef{
        Key: "finance.etims_integration", Name: "KRA eTIMS Real-Time Invoicing",
        Type: "boolean", DefaultValue: "false", Module: "finance", Status: "active",
    })
    featureflag.RegisterFlag(featureflag.FlagDef{
        Key: "finance.new_invoice_pdf", Name: "New Invoice PDF Layout",
        Type: "percentage", DefaultValue: "0", Module: "finance", Status: "active",
    })
}
```

### Why Everything in `init()`?

Go's `init()` functions run before `main()`, in import order, exactly once. By the time `bootstrap.Mount()` is called in `main()`:

- All EntityDefinitions are registered and validated.
- All module metadata rows are upserted.
- All settings keys are registered.
- All feature flags are registered.

`bootstrap.Mount()` reads `definition.All()` and builds the complete server in a single pass. No per-module bootstrap call. Adding a new module to a deployment is one import line in `main.go`.

---

## 48.7 How Navigation Is Built from the Registry

The SDUI sidebar is generated dynamically from:
1. The tenant's active modules (`tenant_modules WHERE status='active'`).
2. The EntityDefinitions registered to each active module (`definition.All()` filtered by `Module == module_name`).

```go
// framework/framework/sdui/nav.go
func buildNavigation(ctx context.Context, tenantID string) (*NavSchema, error) {
    // Get this tenant's active modules
    activeModules, err := tenantModuleStore.List(ctx, pgstore.ListOptions{
        Filter: map[string]any{"status": "active"},
    })
    if err != nil {
        return nil, err
    }

    nav := &NavSchema{}

    for _, row := range activeModules.Items {
        moduleName, _ := row.Get("module_name")
        name := moduleName.(string)

        // Get all EntityDefinitions for this module, sorted by NavOrder
        entityDefs := definition.AllForModule(name)

        section := &NavSection{Label: moduleLabel(name), Icon: moduleIcon(name)}
        for _, def := range entityDefs {
            section.Items = append(section.Items, &NavItem{
                Label: def.Label,
                URL:   "/app/" + def.Name,
                Icon:  def.Icon,
            })
        }

        if len(section.Items) > 0 {
            nav.Sections = append(nav.Sections, section)
        }
    }

    return nav, nil
}
```

Result: a tenant with Finance + Forecourt active sees:
- **Finance** → Accounts, Journal Entries, Invoices, Payments, Tax Rates
- **Forecourt** → Pumps, Tanks, Fuel Deliveries, Wetstock Reports

A tenant with only Finance active sees only the Finance section. The Forecourt section does not appear — not just hidden in UI, but absent from the generated navigation JSON.

---

## 48.8 Trial Module Management

The `trial` status enables a "try before you buy" workflow:

```
Platform admin activates Finance for Acme Petroleum with status='trial', trial_ends_at='2024-04-30'
    │
    ▼
Acme Petroleum sees Finance in their navigation
Acme Petroleum can create invoices, run reports, set up chart of accounts
    │
    ▼
2024-04-30 00:00 UTC — Daily Temporal schedule runs:
SELECT * FROM tenant_modules WHERE status='trial' AND trial_ends_at <= now()
    │
    ▼
For each expired trial:
  UPDATE tenant_modules SET status='suspended' WHERE ...
  → reloadModuleNavigation hook fires
  → Finance disappears from Acme Petroleum's navigation
  → Data is PRESERVED (invoices, chart of accounts all still exist in DB)
  → Tenant admin gets a notification: "Your Finance trial has expired"
    │
    ▼
Acme Petroleum subscribes. Platform admin sets status='active'.
  → Finance reappears in navigation
  → All their trial data is immediately accessible again
```

Data is never deleted on trial expiry. This is critical — a business that entered real data during a trial and then let the trial lapse should be able to recover it when they subscribe.

---

## 48.9 Module Versioning

Each module row stores its `version`. `RegisterModule()` upserts the version on every server startup:

```go
func RegisterModule(info ModuleInfo) {
    // Runs at startup (system context, no tenant).
    moduleStore.Upsert(systemCtx, map[string]any{
        "name":    info.Name,
        "label":   info.Label,
        "version": info.Version,      // updated to current binary version
        "status":  info.Status,
        // description and repo_url also updated
    }, onConflict: "name")
}
```

The admin panel shows the current installed version of every module. This helps support teams identify version mismatches in production.

### Module Compatibility Checks

Modules can declare a minimum framework version requirement:

```go
registry.RegisterModule(registry.ModuleInfo{
    Name:                "forecourt",
    MinFrameworkVersion: "2.0.0", // requires framework v2 or later
    // ...
})
```

At startup, `RegisterModule()` compares against `framework.Version`. If the requirement is not met, the server logs an error and refuses to start:

```
FATAL: module "forecourt" v3.1.0 requires framework >= 2.0.0, but current framework is 1.9.4
       Update the framework or downgrade the forecourt module.
```

This prevents silent incompatibilities where a module ships new hook signatures that the framework doesn't know about.

---

## 48.10 Migration

```sql
-- framework/platform/registry/migrations/20240101000060_create_module_registry.up.sql

-- ── Module Catalogue ────────────────────────────────────────────────────────
-- Global: no tenant_id, no RLS.
-- Populated by each module's RegisterModule() call at startup.
CREATE TABLE modules (
    id          uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text    NOT NULL UNIQUE,   -- stable identifier; never rename
    label       text    NOT NULL,
    version     text    NOT NULL,
    description text,
    repo_url    text,
    status      text    NOT NULL DEFAULT 'stable'
                        CHECK (status IN ('stable', 'beta', 'deprecated')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON modules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Seed platform modules (they register themselves at startup;
-- this seed is a safety net for fresh installs before first startup).
INSERT INTO modules (name, label, version, status) VALUES
    ('tenant',   'Tenant Management', '1.0.0', 'stable'),
    ('iam',      'Identity & Access', '1.0.0', 'stable'),
    ('featureflag', 'Feature Flags',  '1.0.0', 'stable'),
    ('settings', 'Settings',          '1.0.0', 'stable'),
    ('audit',    'Audit Log',         '1.0.0', 'stable'),
    ('metadata', 'Custom Fields',     '1.0.0', 'stable'),
    ('registry', 'Module Registry',   '1.0.0', 'stable')
ON CONFLICT (name) DO UPDATE
    SET version = EXCLUDED.version,
        updated_at = now();


-- ── Tenant Module Activations ───────────────────────────────────────────────
CREATE TABLE tenant_modules (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL REFERENCES tenants(id),
    module_name  text        NOT NULL REFERENCES modules(name),
    status       text        NOT NULL DEFAULT 'active'
                             CHECK (status IN ('active', 'suspended', 'trial')),
    activated_at timestamptz NOT NULL DEFAULT now(),
    trial_ends_at timestamptz,           -- NULL for non-trial activations
    config       jsonb       NOT NULL DEFAULT '{}',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    -- A tenant can have only one activation row per module.
    UNIQUE (tenant_id, module_name)
);

ALTER TABLE tenant_modules ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_modules_isolation ON tenant_modules
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Index for the "what modules does this tenant have active?" query.
-- This runs on every navigation build (cached, but cache misses hit here).
CREATE INDEX tenant_modules_active_idx ON tenant_modules (tenant_id, status)
    WHERE status = 'active';

-- Index for the daily trial expiry job.
CREATE INDEX tenant_modules_trial_expiry_idx ON tenant_modules (trial_ends_at)
    WHERE status = 'trial' AND trial_ends_at IS NOT NULL;

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON tenant_modules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

---

## 48.11 Common Mistakes

**"Can a tenant activate a module they haven't paid for?"**

Not through the API. The `OpAll` policy requires `platform_admin` role. Tenant users can only read `tenant_modules`. Activation is always a platform-admin action that follows a billing event.

**"What if a module is in the server binary but has no row in the `modules` table?"**

`RegisterModule()` runs at startup and upserts the row. On a fresh install, the first server startup creates all module rows. The seed SQL in the migration is a safety fallback.

**"Should I use `tenant_modules.config` or the Settings module for module configuration?"**

Use `tenant_modules.config` for configuration that is set once at activation time and rarely changes (device serial numbers, API environment). Use the Settings module for configuration that tenant admins regularly adjust (tax rates, invoice prefixes, thresholds). The Settings module has a full admin UI, audit trail, and type validation. `tenant_modules.config` is a raw JSONB blob — convenient for setup, not for day-to-day configuration.

**"Can a business write their own module and add it to the registry?"**

Yes. Custom modules (built by the customer's own developers or third-party developers) follow the same pattern: a Go package that imports `awo.so/framework`, registers EntityDefinitions, and calls `registry.RegisterModule()`. They are imported in `main.go` like any other module. There is no distinction between "official" and "custom" modules at runtime.
