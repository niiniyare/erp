# Chapter 2: Why Configuration Management in ERP

[← Executive Summary](./01-executive-summary.md) | [Next: Architecture Overview →](./03-architecture-overview.md)

---

## The Problem Without a Settings Module

ERP systems serve organizations that span multiple industries, geographies, and operational models. Without centralized configuration management, teams face:

**Hardcoded Business Rules** — Developers embed `"INV-"` prefixes and `40` overtime thresholds directly in code. Changing them requires a deployment.

**Scattered Config Files** — Environment variables, YAML files, and database flags owned by different modules drift out of sync. No single source of truth.

**No Inheritance** — A 200-branch retail chain has to configure each branch independently. There's no concept of "inherit from tenant unless overridden."

**Silent Changes** — When someone changes a tax calculation method, there's no audit record of who changed it, from what, to what, and when.

**Risky Onboarding** — Setting up a new tenant means manually copying configuration from a reference tenant. Templates are impossible, so mistakes are common.

---

## What the Settings Module Solves

### Inheritance Eliminates Redundancy

```
Tenant default: invoice_prefix = "INV-"
  ├── Branch A (no override) → resolves "INV-"
  ├── Branch B (overrides)   → resolves "BRANCH-B-INV-"
  └── Branch C (no override) → resolves "INV-"
```

Set once at the tenant level; override only where needed. Changes at the tenant level propagate automatically to all non-overriding entities.

### Templates Enable Fast Onboarding

A new manufacturing tenant can have 45+ standard configuration values applied in seconds:

```http
POST /api/v1/settings/templates/{manufacturing-template-id}/apply
{
  "target_type": "tenant",
  "target_id": "new-tenant-uuid"
}
```

Industry templates live as `SYSTEM` scope and are available to all tenants. Tenants can create their own `TENANT`-scope templates for internal rollouts.

### Audit Provides Accountability

Every configuration change — whether via API, template application, or bulk operation — creates an immutable audit record:

```sql
-- configuration_audit table
module_name    | config_key_name       | old_value | new_value | user_id | applied_at
finance        | invoice_prefix        | "INV-"    | "MFG-INV-"| usr-456 | 2025-09-01
inventory      | valuation_method      | "FIFO"    | "AVCO"    | usr-789 | 2025-09-02
```

### Multi-Tenant Isolation Is Non-Negotiable

Row-Level Security on all configuration tables means that even if a query doesn't filter by `tenant_id`, the database enforces isolation. Tenant A's configurations are invisible to Tenant B — always.

---

## Business Value

| Scenario | Without Settings Module | With Settings Module |
|---|---|---|
| Onboard new tenant | Manual config, hours of work | Template apply, minutes |
| Change approval limit for 150 branches | 150 separate updates | Single bulk operation |
| Audit who changed tax method | Grep through logs, maybe | Instant query on audit table |
| Regional config override | Code change + deploy | Entity-level override via API |
| Roll back bad config | Manual reverse, no record | Audit shows previous value |

---

[← Executive Summary](./01-executive-summary.md) | [Next: Architecture Overview →](./03-architecture-overview.md)
