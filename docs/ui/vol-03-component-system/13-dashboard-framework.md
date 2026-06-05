---
volume: "03 — Component System"
chapter: 13
title: "Dashboard Framework"
audience: "Backend engineers building dashboard screens"
prerequisites:
  - "Vol 01 — DSL fundamentals"
  - "Vol 02 — Block library overview"
  - "Ch 12 — Screen composition"
section: "vol-03-component-system"
related:
  - "[Ch 14 — Charts and Analytics](14-charts-and-analytics.md)"
  - "[Ch 15 — Workflow and Approval Components](15-workflow-and-approval-components.md)"
---

# Chapter 13 — Dashboard Framework

## Table of Contents

1. [Dashboard pattern overview](#131-dashboard-pattern-overview)
2. [KPI Row](#132-kpi-row)
3. [Chart panels](#133-chart-panels)
4. [Activity panels](#134-activity-panels)
5. [Quick actions](#135-quick-actions)
6. [Finance dashboard screen](#136-finance-dashboard-screen)
7. [Building a new dashboard](#137-building-a-new-dashboard)

---

## 13.1 Dashboard pattern overview

The UI platform organises every screen into one of five pattern families:

| Family | Example screens |
|---|---|
| **Document Forms** | Invoice, Purchase Order, Bill, Journal Entry |
| **Listing Pages** | Invoices list, Suppliers list, Inventory items |
| **Data Display** | Detail cards, Kanban boards, Tree views |
| **Reports** | Trial Balance, P&L, Balance Sheet, Cash Flow |
| **Dashboards** | Finance Overview, Operations Overview |

Dashboards are the most compositional family. A dashboard screen does not introduce any patterns unique to itself — it assembles blocks from all other families into a single page. The finance dashboard, for example, combines:

- **KPI Row** (Data Display family — `StatNode`)
- **Chart Panel** (Reports family — `ChartNode`)
- **Activity Panel** (Listing family — `CRUDNode` or `TableNode`)
- **Quick Actions** (Document Form family — `ActionNode`)

This means every technique documented in other chapters applies directly inside a dashboard. The dashboard is just the page that hosts all of them together.

### Design principles

**One `InitAPI` per page.** A dashboard page uses a single `PageNode.InitAPI` call to load all KPI summary data into page scope. Individual blocks then reference `${revenue}`, `${ar_balance}`, and so on as simple variable interpolations. Blocks must not issue their own independent data fetches unless they have genuinely independent lifecycles (e.g., a chart that needs period filtering).

**Blocks own permission checks.** Dashboard blocks — including `QuickActionsBlock` — filter their own output based on the session. The dashboard screen function does not wrap blocks in permission guards. See section 13.5.

**No raw `map[string]any`.** Even inside a dashboard, all nodes must be typed AST structs. See Vol 01, Ch 4 for the zero-map rule.

---

## 13.2 KPI Row

A KPI row is a horizontal band of stat cards at the top of a dashboard. Each card shows one metric with a label, a formatted value, and an optional trend indicator.

### StatNode and CSS custom properties

Stat cards are rendered via `ast.StatNode`. The node uses AMIS CSS custom properties for colour — **never hard-coded hex values and never custom class names**. The required properties are:

| Property | Purpose |
|---|---|
| `--Panel-bg-color` | Card background |
| `--colors-neutral-text-2` | Primary label and value text |
| `--colors-neutral-text-4` | Secondary/subtext |
| `--colors-neutral-line-8` | Card border |

These properties are defined on `:root` in the platform theme and automatically invert when `html.dark` is active. Using them means stat cards inherit dark mode for free.

### Trend colours

Trend indicators use inline AMIS expression syntax, not class names:

```
${trendKey > 0 ? '#52c41a' : '#f5222d'}
```

- `#52c41a` — green (positive / up trend)
- `#f5222d` — red (negative / down trend)

The expression is evaluated client-side by the AMIS runtime against the page-scope data loaded by `InitAPI`.

### KPIRowBlock usage

> **Status: Implemented** — `KPIRowBlock` is called directly in `FinanceDashboardScreen` and accepts `[]StatCardConfig`.

```go
blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
    {Label: "Total Revenue",   ValueKey: "revenue",    Format: "currency", Trend: blocks.TrendUp},
    {Label: "Outstanding AR",  ValueKey: "ar_balance", Format: "currency", Trend: blocks.TrendDown},
    {Label: "Outstanding AP",  ValueKey: "ap_balance", Format: "currency", Trend: blocks.TrendDown},
    {Label: "Cash & Bank",     ValueKey: "cash_balance", Format: "currency"},
})
```

`ValueKey` is the field name in the `InitAPI` response. `Format` controls number rendering — `"currency"` adds thousand separators and two decimal places. `Trend` is one of `blocks.TrendUp`, `blocks.TrendDown`, or omitted.

### AMIS output shape

A `StatNode` serialises to approximately:

```json
{
  "type": "stat",
  "label": "Total Revenue",
  "value": "${revenue|number}",
  "style": {
    "background": "var(--Panel-bg-color)",
    "color": "var(--colors-neutral-text-2)",
    "borderColor": "var(--colors-neutral-line-8)"
  },
  "trendColor": "${revenue_trend > 0 ? '#52c41a' : '#f5222d'}"
}
```

The exact field names depend on the AST serialiser version. The critical invariant is that `style` must reference CSS custom properties, not hard-coded colours.

---

## 13.3 Chart panels

Chart panels display time-series or categorical data, typically placed in the main body of a dashboard beside the activity panel.

### ChartNode requirements

Every `ast.ChartNode` has two mandatory transparency fields. They serve different purposes and both are required:

| Field | Purpose |
|---|---|
| `Style["background"]` = `"transparent"` | CSS applied to the wrapping `<div>` |
| `Config["backgroundColor"]` = `"transparent"` | ECharts canvas background option |

Missing either field causes a ValidateStage failure (HTTP 500). The platform's `ValidateStage` checks `schema["style"]["background"] == "transparent"` explicitly. See Ch 14 for the full ValidateStage contract.

```go
ast.ChartNode{
    Title:  "Revenue vs Expenses",
    Height: 300,
    Style: map[string]string{
        "background": "transparent",   // required — wrapping div
    },
    Config: map[string]any{
        "backgroundColor": "transparent", // required — ECharts canvas
        "tooltip":  map[string]any{"trigger": "axis"},
        "legend":   map[string]any{"data": []string{"Revenue", "Expenses"}},
        "xAxis":    map[string]any{"type": "category", "data": "${chart_labels}"},
        "yAxis":    map[string]any{"type": "value"},
        "series": []map[string]any{
            {"name": "Revenue",  "type": "line", "data": "${revenue_series}"},
            {"name": "Expenses", "type": "line", "data": "${expense_series}"},
        },
    },
}
```

### ChartPanelBlock

`ChartPanelBlock` wraps `ChartNode` inside a titled card with an optional period picker control. It is called in `FinanceDashboardScreen`:

```go
blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
    Title:        "Revenue vs Expenses",
    ChartType:    blocks.ChartTypeLine,
    PeriodPicker: true,
    APIURL:       "/api/v1/finance/dashboard/revenue-chart",
})
```

`PeriodPicker: true` adds a date-range selector above the chart. Changing the range triggers a reload of `APIURL`, which returns updated series data. The chart node references `${revenue_series}` from that API response.

---

## 13.4 Activity panels

An activity panel is a compact recent-transactions feed, typically placed in the right-hand column of a dashboard grid.

### ActivityPanelBlock

`ActivityPanelBlock` wraps a `CRUDNode` or `TableNode` configured for read-only, compact display:

```go
blocks.ActivityPanelBlock(sess, blocks.ActivityPanelConfig{
    Title:    "Recent Transactions",
    Resource: "finance/transactions",
    Limit:    10,
})
```

`Resource` is the API path segment used to construct `/api/v1/{Resource}`. `Limit` sets the page size. The activity panel does not show a filter bar, bulk actions, or a create button — it is display-only.

The underlying AMIS node type is either `"crud"` or `"table"` (depending on whether server-side paging is needed). For dashboard activity feeds, `"table"` with a static `api` is sufficient.

### Layout with GridNode

Dashboard layouts combine chart and activity panels using `ast.GridNode`:

```go
ast.GridNode{Columns: []ast.GridColumn{
    {
        Body: []ast.Node{blocks.ChartPanelBlock(sess, chartCfg)},
        MD:   8,  // 8 of 12 columns at medium breakpoint
    },
    {
        Body: []ast.Node{blocks.ActivityPanelBlock(sess, activityCfg)},
        MD:   4,
    },
}}
```

`MD` follows a 12-column grid. Use `MD: 12` for full-width sections.

---

## 13.5 Quick actions

Quick actions are shortcut buttons at the bottom of a dashboard, linking to common new-document or report pages.

### QuickActionsBlock and QuickAction struct

```go
// QuickAction is defined in internal/web/dsl/blocks/quick_actions.go
type QuickAction struct {
    Label      string
    URL        string
    Permission string // dot-notation: "resource.action" — empty = always visible
    Icon       string
}
```

`QuickActionsBlock` filters actions based on the session before building any nodes. Actions whose `Permission` the session cannot satisfy are excluded entirely — they do not appear in the rendered schema at all.

```go
blocks.QuickActionsBlock(sess, []blocks.QuickAction{
    {
        Label:      "New Invoice",
        URL:        "/finance/invoices/new",
        Permission: "finance.invoices.create",
        Icon:       "fa fa-file-invoice",
    },
    {
        Label:      "Record Payment",
        URL:        "/finance/payments/new",
        Permission: "finance.payments.create",
        Icon:       "fa fa-credit-card",
    },
    {
        Label:      "View Reports",
        URL:        "/finance/reports",
        Permission: "finance.reports.read",
        Icon:       "fa fa-chart-bar",
    },
})
```

### Permission check implementation

Internally, `canPerm` in `quick_actions.go` splits a dot-notation string at the last `.` to derive the action and resource:

```go
func canPerm(sess ui.UISessionContext, perm string) bool {
    i := strings.LastIndex(perm, ".")
    if i < 0 {
        return false
    }
    return sess.Can(perm[i+1:], perm[:i])
    // e.g. "finance.invoices.create" → sess.Can("create", "finance.invoices")
}
```

**The screen function must not gate-keep.** Wrapping `QuickActionsBlock` in an `if sess.Can(...)` check is an anti-pattern — the block already handles it internally. The calling screen should pass all desired actions unconditionally and let the block filter:

```go
// Correct
blocks.QuickActionsBlock(sess, allActions)

// Wrong — defeats the block's own filtering
if sess.Can("create", "finance.invoices") {
    blocks.QuickActionsBlock(sess, filteredActions)
}
```

### Fallback behaviour

If no actions pass the permission filter, `QuickActionsBlock` emits a single "Dashboard" link button rather than an empty node. This guarantees the section is never visually empty.

### AMIS output shape

Each permitted action becomes an `ActionNode` with `ActionType: "link"`:

```json
{
  "type": "action",
  "label": "New Invoice",
  "actionType": "link",
  "link": "/finance/invoices/new",
  "level": "default",
  "icon": "fa fa-file-invoice"
}
```

The outer container is a `FlexNode` with `direction: "row"`, `gap: "sm"`, and `wrap: true`.

---

## 13.6 Finance dashboard screen

`FinanceDashboardScreen` in `internal/web/dsl/screens/finance_dashboard.go` is the reference implementation of the dashboard pattern. It is reproduced here in full for study:

```go
func FinanceDashboardScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "Finance Dashboard",
        // InitAPI loads KPI summary data into page scope.
        // Child StatNodes reference ${revenue}, ${ar_balance}, etc. directly.
        InitAPI: &ast.APISpec{
            Method: "get",
            URL:    "/api/v1/finance/dashboard/summary",
        },
        Body: []ast.Node{
            blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
                {Label: "Total Revenue",  ValueKey: "revenue",     Format: "currency", Trend: blocks.TrendUp},
                {Label: "Outstanding AR", ValueKey: "ar_balance",  Format: "currency", Trend: blocks.TrendDown},
                {Label: "Outstanding AP", ValueKey: "ap_balance",  Format: "currency", Trend: blocks.TrendDown},
                {Label: "Cash & Bank",    ValueKey: "cash_balance", Format: "currency"},
            }),
            ast.GridNode{Columns: []ast.GridColumn{
                {Body: []ast.Node{blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
                    Title:        "Revenue vs Expenses",
                    ChartType:    blocks.ChartTypeLine,
                    PeriodPicker: true,
                    APIURL:       "/api/v1/finance/dashboard/revenue-chart",
                })}, MD: 8},
                {Body: []ast.Node{blocks.ActivityPanelBlock(sess, blocks.ActivityPanelConfig{
                    Title:    "Recent Transactions",
                    Resource: "finance/transactions",
                    Limit:    10,
                })}, MD: 4},
            }},
            blocks.QuickActionsBlock(sess, []blocks.QuickAction{
                {Label: "New Invoice",    URL: "/finance/invoices/new",  Permission: "finance.invoices.create",  Icon: "fa fa-file-invoice"},
                {Label: "Record Payment", URL: "/finance/payments/new",  Permission: "finance.payments.create",  Icon: "fa fa-credit-card"},
                {Label: "View Reports",   URL: "/finance/reports",       Permission: "finance.reports.read",     Icon: "fa fa-chart-bar"},
            }),
        },
    }
}
```

Key observations:

1. **Single `InitAPI`** at the page level. All KPI cards read from that one response.
2. **`GridNode` controls layout.** The 8+4 column split puts the chart in the majority and the activity feed in a narrower right rail.
3. **`QuickActionsBlock` is unconditional.** The screen passes all three actions and the block filters them.
4. **No raw maps.** Every node is a typed struct.

---

## 13.7 Building a new dashboard

Follow these steps to add a new dashboard screen.

### Step 1 — Write the screen function

Create a new file in `internal/web/dsl/screens/`, for example `operations_dashboard.go`:

```go
package screens

import (
    "awo.so/internal/web/ast"
    "awo.so/internal/web/dsl/blocks"
    "awo.so/internal/web/ui"
)

func OperationsDashboardScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "Operations Dashboard",
        InitAPI: &ast.APISpec{
            Method: "get",
            URL:    "/api/v1/operations/dashboard/summary",
        },
        Body: []ast.Node{
            blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
                {Label: "Open POs",          ValueKey: "open_po_count",   Format: ""},
                {Label: "Pending Shipments", ValueKey: "pending_shipments", Format: ""},
                {Label: "Low Stock Items",   ValueKey: "low_stock_count", Format: ""},
            }),
            blocks.QuickActionsBlock(sess, []blocks.QuickAction{
                {Label: "New Purchase Order", URL: "/buy/purchase-orders/new", Permission: "buy.purchase-orders.create", Icon: "fa fa-shopping-cart"},
                {Label: "Receive Goods",      URL: "/buy/grn/new",             Permission: "buy.grn.create",             Icon: "fa fa-truck"},
            }),
        },
    }
}
```

### Step 2 — Register the route

Add the registration in `register.go` (or a module-specific `register_operations.go`):

```go
func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:       "/operations/dashboard",
        Module:      "operations",
        Title:       "Operations Dashboard",
        Description: "Open POs, pending shipments, and quick actions for operations",
        ASTFn: func(sess ui.UISessionContext) any {
            return OperationsDashboardScreen(sess)
        },
    })
}
```

**`ASTFn` is required for new screens.** The legacy `Fn` field is kept only for pages that have not yet migrated to the typed AST.

### Step 3 — Validate at startup

`registry.ValidateRegistry()` is called at application startup after all `init()` functions run. If `Route`, `Module`, or `Title` is missing, or neither `Fn` nor `ASTFn` is set, the validator returns a `*BusinessError` with code `REGISTRY_VALIDATION_FAILED` and the server panics before accepting traffic.

### Step 4 — Implement the backend summary endpoint

The `InitAPI` URL must return a flat JSON object whose keys match the `ValueKey` values referenced by stat cards and chart series. For the operations dashboard above:

```json
{
  "open_po_count": 14,
  "pending_shipments": 3,
  "low_stock_count": 7
}
```

### Step 5 — (Optional) blank-import the screens package

If the new dashboard is in a package not yet imported by `cmd/server`, add a blank import to trigger `init()`:

```go
import _ "awo.so/internal/web/dsl/screens"
```

---

### Dashboard checklist

- [ ] Screen function returns `ast.PageNode` with `InitAPI`
- [ ] `KPIRowBlock` `ValueKey` fields match `InitAPI` response keys
- [ ] Chart nodes have both `Style["background"]` and `Config["backgroundColor"]` set to `"transparent"`
- [ ] `QuickActionsBlock` called unconditionally with all candidate actions
- [ ] Route registered via `registry.RegisterPage` with `ASTFn`
- [ ] `registry.ValidateRegistry()` passes at startup (run the server to confirm)
- [ ] Backend summary endpoint returns flat JSON matching all `ValueKey` references
