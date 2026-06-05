---
chapter: 29
title: "Feature Flag Architecture"
volume: "vol-06-platform"
section: "Platform"
description: "allKnownFlags registry, sess.Flag() in page functions, flag fingerprint cache isolation, and feature-gated UI patterns."
status: implemented
---

# Chapter 29 — Feature Flag Architecture

## Table of Contents

- [29.1 allKnownFlags — The Flag Registry](#291-allknownflags)
- [29.2 sess.Flag() in Page Functions](#292-sessflag-in-page-functions)
- [29.3 Flag Fingerprint for Cache Isolation](#293-flag-fingerprint-for-cache-isolation)
- [29.4 Feature-Gated UI Patterns](#294-feature-gated-ui-patterns)

---

## 29.1 allKnownFlags

All feature flags known to the UI platform are listed statically in `types.go` as `allKnownFlags`. This list drives flag pre-resolution at `AuthzStage`, similar to how `allUIPermissions` drives permission pre-resolution.

```go
// types.go
var allKnownFlags = []string{
    "advanced_reporting",
    "bulk_import",
    "multi_currency",
    "approval_workflow",
    "ai_assist",
    "billing.autopay",
    "hr.payroll_v2",
    "inventory.lot_tracking",
    "finance.auto_reconcile",
}
```

At `AuthzStage`, each flag in `allKnownFlags` is evaluated via `contract.SessionContext.FeatureEnabled(flagName)` and the result is snapshotted into `UISessionContext.featureFlags` as an immutable `map[string]bool`.

### Adding a New Flag

To add a new flag:

1. Add the flag name string to `allKnownFlags` in `types.go`.
2. Enable the flag for the appropriate tenants in the feature flag service.
3. Use `sess.Flag("your_flag_name")` in the relevant page functions or block functions.
4. Bump `SchemaGeneration` in `CacheVersions` to invalidate stale cached schemas that were compiled before the flag existed.

If a flag name is not in `allKnownFlags`, `sess.Flag()` will always return `false` — it is not an error, but the flag will never activate.

---

## 29.2 sess.Flag() in Page Functions

`sess.Flag(name)` returns whether the named flag is active for the current user's session. Like `sess.Can()`, it is a pure map lookup with no I/O.

```go
func (s *UISessionContext) Flag(name string) bool {
    return s.featureFlags[name]
}
```

Flags can be checked at any level — page functions, block functions, or individual node builders. The common patterns:

```go
// Entire section gated by a flag
func FinanceReportsPage(ctx context.Context, sess *UISessionContext) Schema {
    body := []Node{BasicReportsBlock(sess)}

    if sess.Flag("advanced_reporting") {
        body = append(body, AdvancedChartsBlock(sess))
        body = append(body, PivotTableBlock(sess))
    }

    return PageSchema{Title: "Reports", Body: body}
}

// Single field gated by a flag
func InvoiceFormBlock(sess *UISessionContext) FormBlock {
    fields := []Field{
        {Name: "number", Label: "Invoice No.", Type: "text"},
        {Name: "date", Label: "Date", Type: "date"},
        {Name: "amount", Label: "Amount", Type: "number"},
    }

    if sess.Flag("multi_currency") {
        fields = append(fields, Field{
            Name:  "currency",
            Label: "Currency",
            Type:  "select",
            Options: CurrencyOptions(),
        })
    }

    return FormBlock{Fields: fields}
}

// AI feature gated
func InvoiceDetailBlock(sess *UISessionContext) []Node {
    nodes := []Node{InvoiceSummaryNode()}

    if sess.Flag("ai_assist") {
        nodes = append(nodes, AIInsightPanel())
    }

    return nodes
}
```

Flags and permissions can be combined:

```go
// Show AI suggestions only if user has read access AND AI is enabled for the tenant
if sess.Can("read", "finance.invoices") && sess.Flag("ai_assist") {
    nodes = append(nodes, AIInvoiceInsights())
}
```

---

## 29.3 Flag Fingerprint for Cache Isolation

The flag fingerprint is the 8th component of the cache key:

```
route + tenant_id + compiler_version + ast_version + policy_generation + schema_generation + perm_fingerprint + flag_fingerprint
```

The flag fingerprint is a hash of the `featureFlags` map values at the time `UISessionContext` is constructed. Two users with different flag states will have different fingerprints, resulting in separate cache entries.

### Why This Matters

Consider two users in the same tenant:
- User A: `advanced_reporting=true, multi_currency=false`
- User B: `advanced_reporting=false, multi_currency=false`

Without flag fingerprinting, both users would share a cache entry. If User A's schema is cached first (with the advanced analytics block), User B would incorrectly receive the advanced analytics block.

With flag fingerprinting:
```
User A cache key: "finance/reports:tenant-x:v2:v1:pg3:sg1:perm-abc:flag-xyz1"
User B cache key: "finance/reports:tenant-x:v2:v1:pg3:sg1:perm-abc:flag-xyz2"
```

The two cache entries are independent. Each user always receives the schema that matches their flag state.

### Flag State Changes

When a flag is toggled for a user or tenant, the fingerprint changes on the next request — the new request is a cache miss and triggers a fresh compilation. The old cache entry expires via TTL. No explicit invalidation is needed for individual flag changes.

For a global flag rollout (enabling `bulk_import` for all tenants), the tenant-level invalidation scope can be used to proactively clear stale entries across all tenants.

---

## 29.4 Feature-Gated UI Patterns

### Pattern 1: Beta Feature Section

A new module or feature section is hidden behind a flag during development or limited rollout.

```go
func InventoryPage(ctx context.Context, sess *UISessionContext) Schema {
    body := []Node{
        InventoryListBlock(sess),
        InventoryToolbarBlock(sess),
    }

    // Lot tracking is a beta feature
    if sess.Flag("inventory.lot_tracking") {
        body = append(body, LotTrackingPanel(sess))
    }

    return PageSchema{Title: "Inventory", Body: body}
}
```

When `inventory.lot_tracking` is disabled, the panel is completely absent from the compiled schema — not hidden with CSS, not disabled. It does not exist in the JSON.

### Pattern 2: V2 Replacement

A flag gates a new implementation of an existing feature, replacing it entirely when enabled.

```go
func PayrollRunBlock(sess *UISessionContext) Node {
    if sess.Flag("hr.payroll_v2") {
        return PayrollRunV2Block(sess)  // new multi-step wizard
    }
    return PayrollRunV1Block(sess)  // original single-form page
}
```

When the flag is fully rolled out and V1 is removed, the flag check and `PayrollRunV1Block` can both be deleted.

### Pattern 3: Flag + Permission Combination

An advanced feature that requires both the feature flag and a specific permission.

```go
func FinancialReconcileBlock(sess *UISessionContext) []Node {
    // Finance auto-reconcile requires both:
    // 1. The feature to be enabled for the tenant
    // 2. The user to have the manage permission on accounts
    if sess.Flag("finance.auto_reconcile") && sess.Can("manage", "finance.accounts") {
        return []Node{AutoReconcileButton(), ReconcileHistoryLink()}
    }
    return []Node{}
}
```

### Pattern 4: Approval Workflow Toggle

Some workflows are optional and enabled per tenant.

```go
func PurchaseOrderActionsBlock(sess *UISessionContext) []Node {
    nodes := []Node{SaveDraftButton()}

    if sess.Flag("approval_workflow") {
        // Multi-step approval — submit goes to approval queue
        nodes = append(nodes, SubmitForApprovalButton())
    } else {
        // No approval workflow — direct approval
        if sess.Can("approve", "buy.purchase-orders") {
            nodes = append(nodes, ApproveDirectlyButton())
        }
    }

    return nodes
}
```

This pattern allows the same codebase to serve both tenants with and without approval workflows, with the flag controlling which UI path is rendered.
