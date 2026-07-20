> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
chapter: 28
title: "Multi-Tenant Architecture"
volume: "vol-06-platform"
section: "Platform"
description: "TenantID in UISessionContext, cache key tenant isolation, tenant locale/currency, per-tenant cache invalidation, and flag fingerprinting."
status: implemented
---

# Chapter 28 — Multi-Tenant Architecture

## Table of Contents

- [28.1 TenantID in UISessionContext](#281-tenantid-in-uisessioncontext)
- [28.2 Cache Key Tenant Isolation](#282-cache-key-tenant-isolation)
- [28.3 Tenant Currency and Locale](#283-tenant-currency-and-locale)
- [28.4 Cache Invalidation by Tenant](#284-cache-invalidation-by-tenant)
- [28.5 Per-Tenant Feature Flag Overrides](#285-per-tenant-feature-flag-overrides)
- [28.6 PLANNED: Tenant Form Customization Overlays](#286-planned-tenant-form-customization-overlays)

---

## 28.1 TenantID in UISessionContext

Every request is scoped to a single tenant. The `TenantID` is extracted from the validated JWT during `InjectSessionContext` and flows through the entire pipeline as part of `UISessionContext`.

```go
// Accessing tenant identity inside a block or page function
tenantID := sess.TenantID()

// Used for data API calls that require tenant scoping
DataTableBlock{
    Resource: fmt.Sprintf("/api/tenants/%s/invoices", sess.TenantID()),
}
```

The TenantID is treated as a trusted, immutable value after JWT validation. No code in the schema pipeline makes assumptions about which tenant is "current" — it always reads from `sess.TenantID()`.

### Multi-Tenancy Guarantee

The schema pipeline provides a strict per-tenant guarantee: a schema compiled for tenant A can never be served to tenant B. This is enforced by the cache key design (§28.2) and by the fact that all API URLs baked into schemas include tenant-specific paths.

---

## 28.2 Cache Key Tenant Isolation

The 8-component cache key includes `tenant_id` as its second component:

```
route + tenant_id + compiler_version + ast_version + policy_generation + schema_generation + perm_fingerprint + flag_fingerprint
```

Two requests for the same route from different tenants will always produce different cache keys, even if those tenants have identical permissions and feature flags. There is no cross-tenant cache sharing.

```
// User A, Tenant X
key = "finance/invoices:tenant-x:v2:v1:pg3:sg1:perm-abc123:flag-def456"

// User B, Tenant X (same tenant, same permissions)
key = "finance/invoices:tenant-x:v2:v1:pg3:sg1:perm-abc123:flag-def456"
// → Cache HIT — safe because same tenant

// User C, Tenant Y (different tenant, same permissions)
key = "finance/invoices:tenant-y:v2:v1:pg3:sg1:perm-abc123:flag-def456"
// → Cache MISS (different tenant) → separate compilation → separate cached entry
```

The tenant component ensures that even if two tenants have identical permission fingerprints and flag fingerprints, their schemas are stored and served independently.

---

## 28.3 Tenant Currency and Locale

Tenant-specific display settings are available through `UISessionContext`. These are sourced from the `contract.SessionContext` that was hydrated by `InjectSessionContext`.

```go
// Currency — e.g. "USD", "EUR", "GBP"
currency := sess.Currency()
// Sourced from: contract.SessionContext.Setting("tenant.currency")

// Locale — e.g. "en-US", "ar-SA", "fr-FR"
locale := sess.Locale()

// Timezone — e.g. "America/New_York", "Asia/Dubai"
tz := sess.Timezone()
```

Block functions use these values to configure display formatting:

```go
func InvoiceTotalBlock(sess *UISessionContext) Node {
    return NumberDisplay{
        Value:    "${totalAmount}",
        Prefix:   currencySymbol(sess.Currency()),
        Locale:   sess.Locale(),
        Decimals: 2,
    }
}
```

Currency and locale are part of the user's session settings, not hardcoded. A tenant configured for Arabic locale will receive right-to-left formatted schemas automatically if block functions propagate `sess.Locale()` to their output nodes.

---

## 28.4 Cache Invalidation by Tenant

Cache invalidation is scoped. When a tenant's configuration changes, only that tenant's cached schemas need to be invalidated — not the entire cache.

The invalidation API is `cache.Service.DeletePattern`. It accepts a glob pattern that matches against cache keys.

```go
// Invalidate all cached schemas for a specific tenant
cache.DeletePattern(ctx, "finance/invoices:tenant-x:*")  // route + tenant scope

// Invalidate all routes for a tenant (module-level)
cache.DeletePattern(ctx, "*:tenant-x:*")  // full tenant scope

// Invalidate one route across all tenants (schema update)
cache.DeletePattern(ctx, "finance/invoices:*")  // route-level global

// Invalidate everything (version bump — use sparingly)
cache.DeletePattern(ctx, "*")
```

The four invalidation scopes and their triggers:

| Scope        | Pattern                     | Triggered by                              |
|--------------|-----------------------------|-------------------------------------------|
| Route        | `<route>:*`                 | Schema code change for a specific page    |
| Module       | `<module>/*:*`              | Module-wide schema or permission changes  |
| Tenant       | `*:<tenant_id>:*`           | Tenant config or permission assignment    |
| Global       | `*`                         | Compiler version bump, platform upgrades  |

Direct Redis operations are prohibited. All invalidation must go through `cache.Service.DeletePattern` so that key construction logic remains centralized.

---

## 28.5 Per-Tenant Feature Flag Overrides

Feature flags can be overridden per tenant. A flag enabled globally might be disabled for a specific tenant, or a flag might be enabled for a specific tenant before it is rolled out globally.

The flag fingerprint component of the cache key captures the exact set of flags active for the current request:

```
... :flag-<hash of active flags>
```

Two users in the same tenant with the same permissions but different flag assignments will have different flag fingerprints — and thus separate cache entries. This is the correct behavior: the schema for a user with `advanced_reporting: true` may differ from the schema for a user without it.

```go
// In a page function
func FinanceReportsPage(ctx context.Context, sess *UISessionContext) Schema {
    body := []Node{StandardReportsBlock(sess)}

    if sess.Flag("advanced_reporting") {
        body = append(body, AdvancedAnalyticsBlock(sess))
    }

    return PageSchema{Title: "Reports", Body: body}
}
```

The flag fingerprint means that:
- Tenants with `advanced_reporting` off get a schema without the analytics block, cached under their flag fingerprint.
- Tenants with `advanced_reporting` on get a different schema, cached under their flag fingerprint.
- These two cache entries coexist independently.

When a flag's value changes for a tenant, invalidating the tenant's cache entries (scope: tenant) causes both variants to be recompiled on the next request.

---

## 28.6 PLANNED: Tenant Form Customization Overlays

> **Status: PLANNED — not implemented.**

The intent is to allow tenants to customize forms without modifying shared schema code. This would enable:

- **Field additions**: A tenant adds custom fields specific to their business (e.g., an extra reference number field on purchase orders).
- **Field removals**: A tenant hides fields that are irrelevant to their workflow.
- **Label overrides**: A tenant renames "Invoice" to "Tax Invoice" throughout the UI.
- **Module enable/disable**: A tenant can suppress entire modules from their navigation tree.

The planned mechanism is a `TenantOverlayStage` in the pipeline that would apply a tenant-specific overlay to the base schema after compilation. The overlay would be stored in the tenant configuration service and cached alongside the base schema.

This feature requires:
1. A tenant overlay data model and admin UI.
2. A merge/patch algorithm for AMIS schema nodes.
3. A separate cache key component for the overlay version.
4. A no-code customization UI (§30.2).

None of these are implemented. The current multi-tenancy support is limited to currency, locale, timezone, and feature flags via `UISessionContext`.
