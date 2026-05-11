[<-- Back to Index](README.md)

## Resource/Action Ownership Model

> **Implementation status**: Module-owned resource naming is [IMPLEMENTED] in existing policies (see `seed.go`).
> The MRA (Module/Resource/Action) runtime registry is [IMPLEMENTED] as a DB catalogue.
> Feature flags, tenant settings, and user preferences are [IMPLEMENTED] — see schema in migration `000601` and `000801`.

---

### 1. The Core Question

**Should resource/action definitions live inside each module package, or in a shared common registry?**

**Answer: Modules own their resources and actions.**

Each business module defines its own resources and actions. There is no giant centralized permission registry hardcoded at compile time. Instead:

- The **Module/Resource/Action (MRA) registry** is a runtime DB catalogue (`modules`, `resources`, `actions` tables) seeded by each module at deployment time.
- Casbin policies reference **module-qualified resource names** (e.g., `finance.receivables.invoices`).
- Shared contracts exist only for **cross-cutting concerns** (the IAM authorization interface, standard action verbs, naming conventions).
- New modules register themselves without modifying core IAM code.

The existing seed.go already demonstrates this: finance module resources (`finance.receivables.invoices`, `finance.accounts`, `finance.transactions`) and people module resources (`people.employees`, `people.persons`) are registered by their respective modules, not by a central registry.

---

### 2. Why Module Ownership

#### Avoids Tight Coupling

A change to finance module resources (adding `finance.budget_reforecast`) does not require modifying HR module code, IAM core code, or any other module. The only thing that changes is the finance module's seed data and policies.

#### Supports Independent Enable/Disable

Modules are gated by feature flags (`feature_flag_definitions` table, auto-seeded when a `modules` row is inserted). When a module is disabled:
- Its feature flag resolves to `false` in `session.Configuration.Flags`
- Handlers check the flag and return early
- The Casbin policies still exist but are never evaluated because the handler gate stops the request first

This means enabling/disabling a module requires no Casbin policy changes.

#### Consistent with Tenant Isolation

Each tenant's Casbin domain has its own copy of module policies. Enabling the finance module for Tenant A (adding finance policies) does not affect Tenant B's domain. Module policies are tenant-scoped.

#### Extensible Without Core Changes

Third-party modules (or future Awo modules like travel management, forecourt management) can register their resources in the `modules`/`resources`/`actions` tables and seed Casbin policies without any changes to the IAM core package.

---

### 3. How Consistency Is Maintained

#### Resource Naming Convention

Resource names use **dot-separated module qualification**:

```
{module}.{submodule?}.{resource}
```

Examples from current seed data:
```
finance.receivables.invoices    — invoice management under AR
finance.payables.bills          — bill management under AP
finance.accounts                — chart of accounts
finance.transactions            — general ledger transactions
people.employees                — employee records
people.persons                  — person/contact records
settings.iam                    — IAM settings
settings.general                — general tenant settings
```

**Rules:**
- Always use the module slug as the first segment
- Use dot notation, never slashes (slashes are reserved for Casbin path matching with `keyMatch2`)
- Use plural nouns for resources (invoices, not invoice)
- Use lowercase snake_case

#### Action Naming Convention

Standard action verbs (from existing seed data and domain conventions):

| Verb | Meaning |
|---|---|
| `read` | View/list/query |
| `create` | Create new records |
| `update` | Modify existing records |
| `delete` | Delete/archive records |
| `approve` | Business approval workflow |
| `export` | Export to file/external system |
| `void` | Void/cancel a financial document |
| `reconcile` | Mark as reconciled |
| `transfer` | Move inventory/funds between locations |
| `assign` | Assign a record to a user/team |
| `*` | All actions (wildcard — admin roles only) |

Use the `*` wildcard action **only** for admin-level roles. Tenant-user roles should always use explicit action verbs.

#### Module Registration at Boot

When a module row is inserted into the `modules` table, a DB trigger (`trg_seed_module_flag`) automatically creates a `feature_flag_definitions` row for the module-level flag. Similarly for resource rows (`trg_seed_resource_flag`).

This means module-level feature flag definitions are auto-generated from the MRA registry — no manual flag registration needed.

---

### 4. Module Examples

#### Finance Module [IMPLEMENTED - PARTIAL]

Current seed data from `seed.go`:

```
Resources and actions:
  finance.receivables.invoices  → read, create, update, delete, approve, export
  finance.payables.bills        → read, create, update, approve
  finance.accounts              → read, create, update
  finance.transactions          → read, create
```

Planned additions [NOT IN v1.0]:
```
  finance.budget                → read, create, update, approve
  finance.fiscal_year           → read, update
  finance.reconciliation        → read, reconcile, export
  finance.reports               → read, export
```

#### Inventory Module [PLANNED - NOT IN v1.0]

```
Module: inventory

Resources and planned actions:
  inventory.product         → create, read, update, delete
  inventory.warehouse       → create, read, update
  inventory.transfer        → create, read, approve
  inventory.stock_movement  → read, receive, adjust, export
  inventory.purchase_order  → create, read, update, approve
```

Casbin policy examples:
```
(role:warehouse_manager, tenantID, inventory.warehouse, read,   allow)
(role:warehouse_manager, tenantID, inventory.transfer,  create, allow)
(role:warehouse_manager, tenantID, inventory.transfer,  approve, allow)
(role:stock_viewer,      tenantID, inventory.*,          read,   allow)
```

#### CRM Module [PLANNED - NOT IN v1.0]

```
Module: crm

Resources and planned actions:
  crm.contact        → create, read, update, delete, assign
  crm.opportunity    → create, read, update, delete, assign, convert
  crm.activity       → create, read, update, delete
  crm.pipeline       → read, update
```

#### Airline / Travel Module [PLANNED - NOT IN v1.0]

```
Module: travel

Resources and planned actions:
  travel.booking     → create, read, update, cancel
  travel.flight      → read, update
  travel.passenger   → create, read, update, check_in, board
  travel.manifest    → read, export
  travel.seat        → read, assign
```

Domain-specific actions like `check_in`, `board` are module-defined extensions to the standard verb set.

#### Forecourt / Gas Station Module [PLANNED - NOT IN v1.0]

```
Module: forecourt

Resources and planned actions:
  forecourt.pump           → read, authorize, close
  forecourt.nozzle         → read
  forecourt.transaction    → read, authorize, reconcile, export
  forecourt.tank           → read, update
  forecourt.price          → read, update
```

Note the absence of `create` / `delete` for hardware resources like pumps and nozzles — those are managed via provisioning, not end-user actions.

---

### 5. Feature Flags vs Tenant Settings vs User Preferences

These three configuration concepts are distinct. Mixing them up leads to incorrect architecture decisions.

#### Feature Flags

**Purpose**: Platform-controlled on/off switches for entire features, capabilities, or modules.

**Storage**: `feature_flag_definitions` (system catalogue) + `feature_flags` (per-tenant overrides)

**Who controls**:
- Platform operators control `is_system = true` flags (e.g., whether a beta feature is available at all)
- Tenant admins can override `is_system = false` flags within platform-permitted bounds

**Resolved into session**: `session.Configuration.Flags["flag.key"]` — a boolean

**Examples**:
```
enable_finance_module           → true/false (is module available at all)
enable_multi_currency           → true/false (is multi-currency feature on)
finance.payables.bills          → true/false (is AP module available)
hr.payroll_v2.enabled           → true/false (new payroll engine beta flag)
```

**Flag key format**: Derived from MRA slugs. Module-level: `{module.slug}`. Resource-level: `{module.slug}.{resource.slug}`. Never write free-form flag keys — derive them from the MRA registry.

**Default values**: Module-level flags default to `false` (module is off until explicitly enabled). Resource-level flags default to `true` (resource is available once its module is enabled).

**Usage in code**:
```go
if !sess.FeatureEnabled("finance") {
    return ErrFeatureDisabled
}
```

#### Tenant Settings

**Purpose**: Per-tenant configuration values that affect business logic behavior within enabled features.

**Storage**: `setting_definitions` (system catalogue, developer-seeded) + `tenant_settings` (per-tenant values)

**Who controls**: Tenant admins can configure settings within the defined value bounds. System settings (`is_system = true`) require platform operator access.

**Resolved into session**: `session.Configuration.Settings["setting.key"]` — a string (parsed by typed accessors)

**Examples**:
```
iam.session_ttl_hours              → "8" (integer — session lifetime)
finance.transactions.approval_threshold → "10000.00" (decimal — approval required above this amount)
finance.budget_control_mode        → "soft" (enum — warn or block on budget overrun)
iam.mfa.required                   → "true" (bool — enforce MFA for all users)
```

**Setting key format**: Same dot-notation as flags: `{module}.{resource?}.{name}`. Defined in `setting_definitions.setting_key`.

**Usage in code**:
```go
ttl := sess.SettingInt("iam.session_ttl_hours", 8)
threshold := sess.SettingDecimal("finance.transactions.approval_threshold", 0)
```

**Value types**: `bool`, `int`, `decimal`, `text`, `enum` — defined in `setting_definitions.value_type`.

#### User Preferences

**Purpose**: Per-user UI and workflow preferences. Do not affect authorization or business logic.

**Storage**: `user_preferences` table (per-user, key-value)

**Who controls**: Each user controls their own preferences. Admins cannot set preferences for other users via the normal path (RLS enforces `user_id = current_setting('app.user_id')`).

**Resolved into session**: `session.Configuration.Prefs["pref.key"]` — a string

**Examples**:
```
ui.theme                    → "dark" or "light"
finance.entry_mode          → "spreadsheet" or "form"
finance.show_account_codes  → "true"
ui.date_format              → "YYYY-MM-DD"
ui.language                 → "en-GB"
notifications.email_digest  → "daily"
```

**Key point**: User preferences **do not affect authorization**. They are purely cosmetic or workflow conveniences. A user preference cannot grant or restrict any permission.

#### Summary Table

| Concept | Controlled By | Scope | Authorization Impact | Example |
|---|---|---|---|---|
| Feature Flag | Platform (system) or tenant (non-system) | Module/resource on-off | Yes — gates entire features | `enable_finance_module` |
| Tenant Setting | Tenant admin (non-system) or platform (system) | Business behavior values | Indirect (affects thresholds, modes) | `finance.approval_threshold` |
| User Preference | User themselves | UI/UX display choices | None | `ui.theme` |

---

### 6. Naming Conventions Summary

| Element | Convention | Example |
|---|---|---|
| Module resource | `{module}.{sub?}.{resource}` | `finance.receivables.invoices` |
| Feature flag key | `{module.slug}` or `{module.slug}.{resource.slug}` | `finance`, `finance.invoices` |
| Setting key | `{module}.{resource?}.{name}` | `iam.session_ttl_hours` |
| Preference key | `{module}.{name}` | `finance.entry_mode` |
| Casbin role | `role:{name}` | `role:accounts_payable_clerk` |
| Standard actions | create, read, update, delete, approve, export, void, reconcile, assign | — |
| Admin wildcard | `*` for subject and/or action | Admin roles only |

**Critical rule**: Never use wildcards (`*`) in Casbin policies for tenant-user roles. Wildcards are only appropriate for the `tenant_admin` and platform admin roles. All other roles should have explicit, minimal permission grants.

---

See also:
- [Casbin Policy Engine](./05-casbin-policy-engine.md)
- [Role Management](./07-role-management.md)
- [Tenant Administration](./23-tenant-administration.md)
- [Feature Flags](./04b-feature-flags.md)
- [Tenant Settings](./05b-tenant-settings.md)
