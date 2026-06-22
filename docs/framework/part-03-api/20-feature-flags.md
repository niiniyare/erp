---
title: "Chapter 20: Feature Flags"
part: "Part III — The API Layer"
chapter: 20
section: "20-feature-flags"
related:
  - "[Chapter 14: Multi-Tenancy Middleware](14-multitenancy-middleware.md)"
  - "[Chapter 41: Redis Usage](../part-07-multitenancy/41-redis-usage.md)"
---

# Chapter 20: Feature Flags

Feature flags in Awo control which features are available to which tenants, enable percentage rollouts, and gate SDUI components without code changes. The implementation uses Redis for caching with a PostgreSQL source of truth.

---

## 20.1. Feature Flag Service

### 20.1.1. Flag Definition

```go
type FeatureFlag struct {
    Name         string          // stable key: "advanced_reporting"
    Type         FlagType        // boolean | string | percentage
    DefaultValue interface{}     // system-wide default
    Description  string
    Status       string          // draft | active | deprecated | removed
}

type TenantFlagOverride struct {
    TenantID   uuid.UUID
    FlagName   string
    Value      interface{}     // overrides DefaultValue for this tenant
    RolloutPct int             // 0-100; for percentage flags
}
```

### 20.1.2. Redis Storage Layout

```
flag:{tenant_id}:{flag_name}         → serialised FlagValue (TTL: 15 min)
eval:{hash(tenant_id+flag_name+ctx)} → evaluation result (TTL: 5 min)
bulk_eval:{hash(tenant_id)}          → all flags for tenant (TTL: 2 min)
ff_stats:{tenant_id}:{flag_name}     → evaluation count stats (TTL: 30 sec)
```

The `CachedFeatureFlagService` (implemented in `internal/core/featureflag/service_cached.go`) handles cache population and invalidation.

### 20.1.3. Flag Loading at Tenant Boot

During tenant middleware resolution, all flags for the tenant are loaded in a single bulk fetch and stored in the `TenantContext.Features` field. Route handlers and page builders access flags from context — no additional DB or Redis calls per-flag:

```go
// In tenant middleware:
flags, err := ffService.BulkEvaluate(ctx, tenantID)
if err != nil {
    return err
}
tenantCtx.Features = flags
```

---

## 20.2. Flag Evaluation

### 20.2.1. Evaluation Order

For any given flag evaluation `(tenant, flag_name)`:
1. Check for a **user-level override** (rare, used for individual beta testers)
2. Check for a **tenant-level override** in `TenantFlagOverride`
3. Check the **system default** in `FeatureFlag.DefaultValue`

The first value found wins.

### 20.2.2. Percentage Rollout — Stable Hashing

For `percentage` type flags, the evaluation is deterministic per tenant — the same tenant always gets the same result for a given rollout percentage. This prevents a tenant from getting different UI behaviour on different page loads.

```go
func evaluatePercentage(tenantID uuid.UUID, flagName string, pct int) bool {
    // Deterministic hash: same tenant + flag = same bucket
    h := sha256.New()
    h.Write([]byte(tenantID.String()))
    h.Write([]byte(flagName))
    hash := h.Sum(nil)

    // Map first 4 bytes to 0-9999
    bucket := int(binary.BigEndian.Uint32(hash[:4])) % 10000
    return bucket < pct*100  // pct=50 → bucket < 5000 → 50% of tenants get true
}
```

### 20.2.3. Accessing Flags in Handlers and Page Builders

```go
// In a route handler:
func (h *InvoiceHandler) List(c *fiber.Ctx) error {
    tenantCtx := middleware.GetTenant(c)
    if tenantCtx.Features.Bool("advanced_reporting") {
        // Include advanced analytics in response
    }
    return c.Next()
}

// In a page builder:
func BuildInvoicePage(ctx context.Context, tenantCtx *TenantContext) *amis.Page {
    flags := tenantCtx.Features
    page := amis.NewPage()

    if flags.Bool("invoice_batch_submit") {
        page.AddToolbarButton(batchSubmitButton())
    }

    return page
}
```

---

## 20.3. Flag Lifecycle

### 20.3.1. Draft → Active → Deprecated → Removed

```
draft:      flag defined, not evaluated (returns default)
active:     flag evaluated normally
deprecated: evaluated but logs a warning; will be removed
removed:    evaluation returns the final value (100% or 0%); code references should be cleaned up
```

Flags in `removed` status return a constant value without hitting Redis — the flag is effectively hardcoded at cleanup time.

### 20.3.2. Dead Flag Cleanup

A Temporal workflow runs weekly to detect dead flags:
- Flags with `rollout_pct = 100` → feature is universally available, remove the flag check
- Flags with `rollout_pct = 0` → feature is disabled for all, remove the code branch
- Flags in `deprecated` status for >30 days → auto-generate a cleanup PR

### 20.3.3. Flag Dependency Graph

Some features depend on others being enabled:

```go
type FeatureFlag struct {
    // ...
    Requires []string `json:"requires"` // flag names that must be enabled
}
```

If flag `advanced_reporting` requires `analytics_module`, evaluating `advanced_reporting` when `analytics_module` is false returns false regardless of `advanced_reporting`'s own value.

---

## 20.4. Feature Flags in the SDUI Layer

### 20.4.1. Page Builders Receive the Resolved Flag Set

Page builder functions receive the `TenantContext` which includes the fully resolved flag set. No additional flag evaluation is needed inside the page builder:

```go
type PageBuilderContext struct {
    Tenant      *TenantContext
    Actor       *Actor
    Permissions *Permissions
    Flags       FeatureFlagSet  // already evaluated
}
```

### 20.4.2. Conditionally Including UI Blocks Based on Flags

```go
func BuildDashboard(ctx PageBuilderContext) *amis.Page {
    page := amis.NewPage()

    // Always include standard metrics
    page.AddBlock(standardMetrics())

    // Only include advanced analytics if enabled
    if ctx.Flags.Bool("advanced_analytics_dashboard") {
        page.AddBlock(advancedAnalytics())
    }

    // String flag: control which chart library to use
    chartType := ctx.Flags.String("dashboard_chart_library", "echarts")
    page.AddBlock(revenueChart(chartType))

    return page
}
```

### 20.4.3. Flag-Gated Menu Items

The sidebar navigation is generated server-side per-tenant. Menu items for disabled features are simply not included in the amis navigation schema:

```go
func BuildNavigation(ctx PageBuilderContext) *amis.Nav {
    nav := amis.NewNav()
    nav.AddItem("Invoices", "/invoices")
    nav.AddItem("Customers", "/customers")

    if ctx.Flags.Bool("payroll_module") {
        nav.AddItem("Payroll", "/payroll")
    }
    if ctx.Flags.Bool("crm_module") {
        nav.AddItem("CRM", "/crm")
    }

    return nav
}
```

The tenant sees only the navigation items for their enabled modules. There are no disabled/greyed-out menu items pointing to pages they cannot access — the items don't exist in their rendered UI at all.
