---
title: "Chapter 40: Feature Flag Deep Dive"
part: "Part VII — Multi-Tenancy Operations"
chapter: 40
section: "40-feature-flag-deep-dive"
related:
  - "[Chapter 20: Feature Flags](../part-03-api/20-feature-flags.md)"
  - "[Chapter 41: Redis Usage Patterns](41-redis-usage.md)"
---

# Chapter 40: Feature Flag Deep Dive

Chapter 20 introduced the feature flag system from an API perspective. This chapter goes deeper: flag evaluation semantics, rollout strategies, gradual percentage rollouts, A/B testing, emergency kill switches, and the operational discipline required to keep flags from accumulating.

---

## 40.1. Flag Types

Not all flags are the same. Awo distinguishes four flag types:

| Type | Purpose | Example |
|---|---|---|
| `release` | Gate in-progress features | `new_invoice_editor` |
| `kill_switch` | Emergency disable of problematic feature | `mpesa_stkpush_enabled` |
| `experiment` | A/B test of two implementations | `checkout_flow_v2` |
| `operational` | Configuration that changes behaviour without A/B | `max_concurrent_payroll_runs` |

Kill switches default to `true` (feature enabled) and flip to `false` in emergencies. Release flags default to `false` and flip to `true` as features go live. This difference in default matters for safe deployments.

---

## 40.2. Rollout Targets

### 40.2.1. Global Default

The flag value applies to all tenants unless overridden:

```json
{
  "name": "new_invoice_editor",
  "type": "release",
  "global_default": false,
  "tenant_overrides": {
    "acme-petroleum": true,
    "beta-corp": true
  }
}
```

### 40.2.2. Percentage Rollout

Enable for a percentage of tenants, selected deterministically:

```go
func evaluatePercentageRollout(flagName string, tenantID uuid.UUID, percentage int) bool {
    if percentage >= 100 {
        return true
    }
    if percentage <= 0 {
        return false
    }
    // SHA-256 of "flag:{flagName}:tenant:{tenantID}" → first 4 bytes as uint32
    // Divides into 0-9999 bucket, compares against percentage * 100
    key := fmt.Sprintf("flag:%s:tenant:%s", flagName, tenantID.String())
    hash := sha256.Sum256([]byte(key))
    bucket := binary.BigEndian.Uint32(hash[:4]) % 10000
    return int(bucket) < percentage*100
}
```

**Deterministic assignment**: A tenant always lands in the same bucket for a given flag. If a tenant is in the 30% rollout, they stay there — they don't flip in and out on each request. This is critical for UX consistency.

**Sticky rollout**: Once the percentage increases from 30% to 50%, the 30% who were already enabled stay enabled, and 20% new tenants are added. The bucket assignment ensures monotonic rollout.

### 40.2.3. Plan-Based Targeting

Enable flags only for tenants on specific plans:

```json
{
  "name": "advanced_analytics",
  "type": "release",
  "plan_targets": ["growth", "enterprise"],
  "global_default": false
}
```

Evaluation order: tenant override → plan target → percentage rollout → global default. The first matching rule wins.

### 40.2.4. User-Level Flags

For A/B testing within a tenant, flags can target specific users (employees):

```go
type FlagEvalContext struct {
    TenantID uuid.UUID
    UserID   uuid.UUID  // optional — for user-level experiments
    Plan     string
    Tags     []string   // tenant tags e.g. "beta_tester", "enterprise_pilot"
}
```

The evaluation order becomes: user override → tenant override → user percentage → tenant percentage → plan target → global default.

---

## 40.3. SDUI Integration

### 40.3.1. Navigation Gating

Module navigation items are hidden for disabled flags:

```go
func buildNavigation(ctx PageBuilderContext) []amis.NavItem {
    items := []amis.NavItem{
        {Label: "Finance", URL: "/finance", Icon: "fa-calculator"},
        {Label: "Inventory", URL: "/inventory", Icon: "fa-boxes"},
    }

    if ctx.Flags.IsEnabled("forecourt_module") {
        items = append(items, amis.NavItem{
            Label: "Forecourt",
            URL:   "/forecourt",
            Icon:  "fa-gas-pump",
        })
    }

    if ctx.Flags.IsEnabled("crm_module") {
        items = append(items, amis.NavItem{
            Label: "CRM",
            URL:   "/crm",
            Icon:  "fa-users",
        })
    }

    return items
}
```

The navigation schema is cached per tenant+role+flag-hash. When flags change, the flag-hash changes, causing a cache miss and fresh navigation rebuild on next request.

### 40.3.2. Field-Level SDUI Flags

Individual form fields can be gated:

```go
// Show advanced pricing fields only for tenants with pricing_v2 enabled
if ctx.Flags.IsEnabled("pricing_v2") {
    fields = append(fields, amis.TextField{
        Name:  "contract_price",
        Label: "Contract Price (override)",
    })
}
```

---

## 40.4. Emergency Kill Switch Pattern

### 40.4.1. Designing Kill Switches

Every integration with an external service should have a kill switch:

```go
func (h *PaymentHandler) InitiateMPESAPayment(c *fiber.Ctx) error {
    if !h.flags.IsEnabled(c.Context(), "mpesa_stkpush_enabled") {
        return errs.NewBusinessError("FEATURE_DISABLED",
            "M-PESA payments are temporarily unavailable — please try again later or pay by bank transfer")
    }
    // proceed with STK Push
}
```

**Default true**: kill switches are enabled by default. The code path is active in production until the flag is flipped off. Flipping to `false` disables without a deployment.

### 40.4.2. Kill Switch Response Time

The Redis cache TTL for kill switches is set to 30 seconds (vs 15 minutes for regular flags):

```go
const (
    flagCacheTTLNormal    = 15 * time.Minute
    flagCacheTTLKillSwitch = 30 * time.Second
)

func cacheTTL(flag *FeatureFlag) time.Duration {
    if flag.Type == "kill_switch" {
        return flagCacheTTLKillSwitch
    }
    return flagCacheTTLNormal
}
```

When the M-PESA API starts returning errors, the on-call engineer flips the kill switch. Within 30 seconds, all new requests fail fast with the user-friendly error message instead of timing out against the external API.

---

## 40.5. Flag Lifecycle Management

### 40.5.1. The Flag Debt Problem

Feature flags accumulate. A codebase with 200 flags becomes hard to reason about — which flags are still relevant? Which branches are dead code? The solution is disciplined lifecycle management:

```
draft → active → deprecated → removed
```

- `draft`: defined but not yet evaluated — safe to deploy code with flag check before the flag is in the registry
- `active`: evaluated in production
- `deprecated`: globally enabled for all tenants — the `false` branch is dead code, ready for removal
- `removed`: flag deleted from registry; code using it should have been cleaned up

### 40.5.2. Flag Cleanup Policy

Each flag should have a planned removal date set at creation:

```json
{
  "name": "new_invoice_editor",
  "type": "release",
  "created_at": "2025-07-01",
  "planned_removal": "2025-10-01",
  "owner": "payments-team"
}
```

A nightly workflow scans for flags past their `planned_removal` date and posts a reminder to the owner's Slack channel. Flags that have been 100% enabled for >90 days are automatically moved to `deprecated`, generating a cleanup task.

### 40.5.3. Code Cleanup for Removed Flags

When a flag is removed, the code guard must be cleaned up:

**Before** (with flag):
```go
if flags.IsEnabled(ctx, "new_invoice_editor") {
    return newInvoiceEditorHandler(c)
}
return legacyInvoiceEditorHandler(c)
```

**After** (flag removed, old path deleted):
```go
return newInvoiceEditorHandler(c)
```

The `legacyInvoiceEditorHandler` function and all its dependencies can also be deleted. This is how flags prevent permanent code debt — but only if they are actually cleaned up.

---

## 40.6. Multi-Tenant Flag Audit

### 40.6.1. Flag Override Audit

Every change to a flag override is recorded:

```go
type FlagAuditEntry struct {
    TenantID    uuid.UUID
    FlagName    string
    OldValue    *bool   // nil = no previous override
    NewValue    bool
    ChangedBy   uuid.UUID
    ChangedAt   time.Time
    Reason      string  // free text — required for production flag changes
}
```

Requiring a `reason` for production flag changes creates accountability and makes incident post-mortems easier ("we enabled forecourt_live_pricing for acme-petroleum on 2025-07-15 to support their new pump controller").

### 40.6.2. Flag State Report

The platform admin dashboard shows flag state across all tenants:

```
Flag: mpesa_stkpush_enabled (kill_switch)
  Global default: true
  Tenant overrides: 0
  Status: ACTIVE

Flag: new_invoice_editor (release)
  Global default: false
  Percentage rollout: 25%
  Tenant overrides: 3 (all true)
  Status: ACTIVE
  Planned removal: 2025-10-01
  Days overdue: 0
```

This report flags (pun intended) any kill switches that have been `false` for more than 7 days — indicating either a prolonged outage that needs a proper fix, or a forgotten kill switch that should be removed.
