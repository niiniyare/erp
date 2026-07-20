> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
chapter: 30
title: "Customization Framework"
volume: "vol-06-platform"
section: "Platform"
description: "What is implemented today (tenant locale/currency), and what is planned for tenant-specific form overlays, label overrides, and module control."
status: partially-implemented
---

# Chapter 30 — Customization Framework

## Table of Contents

- [30.1 What Is Implemented](#301-what-is-implemented)
- [30.2 PLANNED: Tenant-Specific Field Overlays](#302-planned-tenant-specific-field-overlays)
- [30.3 PLANNED: Label and Copy Overrides](#303-planned-label-and-copy-overrides)
- [30.4 PLANNED: Module Enable/Disable](#304-planned-module-enabledisable)
- [30.5 PLANNED: No-Code Customization UI](#305-planned-no-code-customization-ui)

---

## 30.1 What Is Implemented

The customization framework is in its first phase. Tenant-specific display settings are available through `UISessionContext` and are already factored into the schema pipeline.

### Tenant Currency

Every block that renders monetary values should use `sess.Currency()` to format output correctly for the tenant.

```go
currency := sess.Currency()
// e.g., "USD", "EUR", "SAR", "GBP"

// Use in display nodes
AmountDisplayNode{
    Value:    "${amount}",
    Currency: currency,
    Locale:   sess.Locale(),
}
```

The currency is sourced from `contract.SessionContext.Setting("tenant.currency")` during `InjectSessionContext`.

### Tenant Locale

`sess.Locale()` returns the BCP-47 locale tag for the tenant (e.g., `en-US`, `ar-SA`, `fr-FR`). This controls:

- Number formatting (decimal separators, grouping separators)
- Date formatting
- Right-to-left layout when the locale requires it

Block functions should propagate the locale to any node that renders formatted values.

```go
DateFieldNode{
    Name:   "invoice_date",
    Label:  "Date",
    Format: localeDateFormat(sess.Locale()),
}
```

### Tenant Timezone

`sess.Timezone()` returns the IANA timezone name (e.g., `America/New_York`, `Asia/Dubai`). This is used for timestamp display and for date-picker defaults.

```go
DateTimePickerNode{
    Name:     "scheduled_at",
    Label:    "Scheduled Date",
    Timezone: sess.Timezone(),
}
```

### Feature Flags as Customization

Feature flags (Chapter 29) are the current mechanism for tenant-level feature customization. A tenant can have `multi_currency`, `approval_workflow`, or `advanced_reporting` enabled independently. This provides per-tenant UI variation within a shared codebase.

---

## 30.2 PLANNED: Tenant-Specific Field Overlays

> **Status: PLANNED — not implemented.**

The goal is to allow tenants to add, remove, or reorder fields on existing forms without modifying the shared schema code.

### Intended Design

A `TenantOverlayStage` would run after the base schema is compiled (or retrieved from cache) and apply a patch to the compiled schema:

```
CacheStoreStage (priority 80)
      ↓
TenantOverlayStage (priority 90) ← PLANNED
      ↓
Response
```

The overlay would be stored as a JSON patch document in the tenant configuration service:

```json
{
  "route": "buy/purchase-orders",
  "tenant_id": "tenant-x",
  "patches": [
    {
      "op": "add",
      "path": "/body/0/fields/-",
      "value": {
        "type": "text",
        "name": "internal_ref",
        "label": "Internal Reference",
        "required": false
      }
    },
    {
      "op": "remove",
      "path": "/body/0/fields/5"
    }
  ]
}
```

### Open Design Questions

- How should overlays interact with the cache? A separate overlay version component in the cache key would be needed.
- How are field positions (path indices) maintained when the base schema changes?
- Should overlays apply before or after permission pruning?

These questions need resolution before implementation can begin.

---

## 30.3 PLANNED: Label and Copy Overrides

> **Status: PLANNED — not implemented.**

Different business domains use different terminology for the same concepts. One tenant calls it "Invoice," another calls it "Tax Invoice," a third calls it "Bill."

The planned label override system would allow tenants to provide a translation/terminology table that gets applied to all rendered schemas:

```json
{
  "tenant_id": "tenant-x",
  "overrides": {
    "en-US": {
      "Invoice": "Tax Invoice",
      "Customer": "Client",
      "Vendor": "Supplier"
    }
  }
}
```

The implementation would require a post-compilation text substitution pass that is locale-aware and does not break AMIS schema structure.

---

## 30.4 PLANNED: Module Enable/Disable

> **Status: PLANNED — not implemented.**

Tenants on lower-tier plans should not see modules they have not purchased. The planned mechanism is tenant-level module registration:

- Each tenant has a set of enabled modules.
- The navigation tree compilation filters modules against the tenant's enabled set.
- Route-level schema requests for disabled modules return 404 rather than 403 (the module does not exist for this tenant, not "you don't have permission").

This is distinct from permission-based access: a user might have the `finance.invoices.read` permission in Casbin, but if the finance module is disabled for the tenant, no invoice pages appear.

Feature flags partially address this today (e.g., disabling `hr.payroll_v2` hides the new payroll UI), but there is no clean module-level on/off switch tied to subscription tier.

---

## 30.5 PLANNED: No-Code Customization UI

> **Status: PLANNED — not implemented.**

The end goal of the customization framework is an admin UI that allows tenant administrators to:

1. Add custom fields to any form.
2. Reorder fields using drag-and-drop.
3. Override labels.
4. Enable or disable modules.
5. Preview changes before publishing.

This requires:
- A metadata-driven schema for "what can be customized on this page."
- A visual editor that produces overlay patch documents.
- A preview pipeline that applies overlays without saving them.
- Version control for overlay changes (who changed what, rollback capability).

No part of this admin UI is implemented. The overlay data model and `TenantOverlayStage` (§30.2) are prerequisites.
