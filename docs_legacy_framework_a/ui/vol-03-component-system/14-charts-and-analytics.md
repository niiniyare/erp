> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
volume: "03 — Component System"
chapter: 14
title: "Charts and Analytics"
audience: "Backend engineers building report and dashboard screens"
prerequisites:
  - "Ch 13 — Dashboard Framework"
  - "Vol 01 Ch 4 — AST node types"
section: "vol-03-component-system"
related:
  - "[Ch 13 — Dashboard Framework](13-dashboard-framework.md)"
  - "[Ch 16 — ERP-Specific Components](16-erp-specific-components.md)"
---

# Chapter 14 — Charts and Analytics

## Table of Contents

1. [AMIS chart type overview](#141-amis-chart-type-overview)
2. [ChartNode fields](#142-chartnode-fields)
3. [Critical invariant — the transparency requirement](#143-critical-invariant--the-transparency-requirement)
4. [Common ECharts config patterns](#144-common-echarts-config-patterns)
5. [Trial balance chart example](#145-trial-balance-chart-example)
6. [Dashboard chart panels](#146-dashboard-chart-panels)
7. [Legacy ChartBuilder usage](#147-legacy-chartbuilder-usage)

---

## 14.1 AMIS chart type overview

The AMIS `"chart"` type is a thin wrapper around [Apache ECharts](https://echarts.apache.org/). When the browser renders a page, the AMIS runtime hands the `config` object inside the chart node directly to ECharts as its option tree. This means:

- Any ECharts option documented in the ECharts API reference is valid inside `Config`.
- Dynamic data is expressed as AMIS expressions within config values — for example `"data": "${revenue_series}"` — and the runtime substitutes the page-scope value before passing to ECharts.
- Theming is handled via the two mandatory transparency fields (see section 14.3). Do not set chart background colours via ECharts theme objects.

The AMIS `"chart"` type has no built-in analytics logic. Aggregation, period filtering, and data transformation happen on the backend. The frontend only renders what the API returns.

---

## 14.2 ChartNode fields

`ast.ChartNode` has the following fields:

| Field | Type | Required | Purpose |
|---|---|---|---|
| `Title` | `string` | No | Section heading above the chart |
| `Config` | `map[string]any` | Yes | Full ECharts option object |
| `Style` | `map[string]string` | Yes | CSS applied to the wrapping div |
| `Height` | `int` | No | Pixel height (default: 300) |
| `API` | `*ast.APISpec` | No | If set, the chart fetches its own data independently of the page `InitAPI` |

### Config

`Config` is the ECharts option object. Every ECharts option is valid here. The minimum required key is `"backgroundColor": "transparent"` (see section 14.3).

```go
ast.ChartNode{
    Title:  "Monthly Revenue",
    Height: 280,
    Style: map[string]string{
        "background": "transparent",
    },
    Config: map[string]any{
        "backgroundColor": "transparent",
        "tooltip": map[string]any{"trigger": "axis"},
        "xAxis":   map[string]any{"type": "category", "data": "${months}"},
        "yAxis":   map[string]any{"type": "value"},
        "series": []map[string]any{
            {"name": "Revenue", "type": "bar", "data": "${monthly_revenue}"},
        },
    },
}
```

### Style

`Style` is a CSS map applied to the `<div>` that wraps the ECharts canvas. It must always contain `"background": "transparent"`. Additional style overrides (e.g., `"margin": "0"`) are permitted.

### API (independent chart data)

When `API` is set, the chart fetches its data independently from the page `InitAPI`. Use this for charts with period-picker controls that trigger re-fetches:

```go
ast.ChartNode{
    // ...
    API: &ast.APISpec{
        Method: "get",
        URL:    "/api/v1/finance/dashboard/revenue-chart",
        // SendOn can limit when the fetch fires, e.g. "${period_start && period_end}"
    },
}
```

When `API` is nil, the chart reads data from the page scope populated by the page `InitAPI`. This is the common case for dashboard summary charts.

---

## 14.3 Critical invariant — the transparency requirement

The ValidateStage pipeline checks every `ChartNode` serialisation and asserts:

```
schema["style"]["background"] == "transparent"
```

If this assertion fails, the page render pipeline returns HTTP 500 with error code `COMPILE_CHART_INVALID_BACKGROUND`. This check exists because a non-transparent background breaks the platform dark theme — opaque white or grey backgrounds remain visible as mis-coloured rectangles when the rest of the page inverts.

**Both fields must be set. They are not equivalent to each other:**

| Field | What it controls |
|---|---|
| `Style["background"] = "transparent"` | The HTML `<div>` surrounding the ECharts canvas |
| `Config["backgroundColor"] = "transparent"` | The ECharts canvas itself |

Setting only one still fails dark mode:

- Only `Style` set: the canvas background stays white inside a transparent div.
- Only `Config` set: the div has an opaque background that ignores the dark theme.

Both must always be set. Copy this pattern exactly:

```go
Style: map[string]string{
    "background": "transparent",
},
Config: map[string]any{
    "backgroundColor": "transparent",
    // ... rest of ECharts options
},
```

---

## 14.4 Common ECharts config patterns

The following snippets show the minimum `Config` maps for the three most common chart types. Add `"backgroundColor": "transparent"` to each (shown in section 14.3's pattern above).

### Bar chart

```go
Config: map[string]any{
    "backgroundColor": "transparent",
    "tooltip": map[string]any{"trigger": "axis"},
    "legend":  map[string]any{"data": []string{"Revenue", "Cost"}},
    "xAxis": map[string]any{
        "type": "category",
        "data": "${period_labels}", // AMIS expression — array from page scope
    },
    "yAxis": map[string]any{"type": "value"},
    "series": []map[string]any{
        {"name": "Revenue", "type": "bar", "data": "${revenue_series}"},
        {"name": "Cost",    "type": "bar", "data": "${cost_series}"},
    },
}
```

### Line chart

```go
Config: map[string]any{
    "backgroundColor": "transparent",
    "tooltip":  map[string]any{"trigger": "axis"},
    "legend":   map[string]any{"data": []string{"Revenue", "Expenses"}},
    "xAxis":    map[string]any{"type": "category", "data": "${months}"},
    "yAxis":    map[string]any{"type": "value"},
    "series": []map[string]any{
        {"name": "Revenue",  "type": "line", "smooth": true, "data": "${revenue_series}"},
        {"name": "Expenses", "type": "line", "smooth": true, "data": "${expense_series}"},
    },
}
```

### Pie / Donut chart

```go
Config: map[string]any{
    "backgroundColor": "transparent",
    "tooltip": map[string]any{"trigger": "item"},
    "legend":  map[string]any{"orient": "vertical", "left": "left"},
    "series": []map[string]any{
        {
            "name":      "Expense Breakdown",
            "type":      "pie",
            "radius":    []string{"40%", "70%"}, // donut variant
            "data":      "${expense_breakdown}",  // array of {name, value}
            "emphasis":  map[string]any{"itemStyle": map[string]any{"shadowBlur": 10}},
        },
    },
}
```

For pie series, the backend must return `expense_breakdown` as an array of objects with `name` and `value` keys:

```json
{
  "expense_breakdown": [
    {"name": "Salaries", "value": 45000},
    {"name": "Rent",     "value": 12000},
    {"name": "Utilities","value": 3200}
  ]
}
```

### Dynamic colours using AMIS expressions

AMIS expressions can be used inside `Config` for dynamic colour logic. The most common case is a trend colour on a single-series stat chart:

```go
"itemStyle": map[string]any{
    "color": "${total_change > 0 ? '#52c41a' : '#f5222d'}",
},
```

The AMIS runtime evaluates this expression before passing `config` to ECharts, so ECharts receives a plain string like `"#52c41a"`.

---

## 14.5 Trial balance chart example

`TrialBalanceScreen` in `internal/web/dsl/screens/trial_balance.go` uses report blocks rather than raw `ChartNode` calls. The screen is included here to show how chart output fits into a report page:

```go
func TrialBalanceScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "Trial Balance",
        Body: []ast.Node{
            blocks.ReportHeaderBlock(sess, blocks.ReportHeaderConfig{
                Title: "Trial Balance",
            }),
            blocks.ReportFilterBlock(sess, blocks.ReportFilterConfig{
                ShowPeriod:   true,
                ShowCurrency: true,
            }),
            blocks.ReportTableBlock(sess, blocks.ReportTableConfig{
                Columns: []blocks.ReportColumnDef{
                    {Name: "account_code", Label: "Code",    Type: "text"},
                    {Name: "account_name", Label: "Account", Type: "text"},
                    {Name: "debit",        Label: "Debit",   Type: "currency"},
                    {Name: "credit",       Label: "Credit",  Type: "currency"},
                },
                ShowGrandTotal: true,
                Exportable:     true,
            }),
        },
    }
}
```

Note that `TrialBalanceScreen` does not include a `ReportChartBlock` in the implemented version. The tasks design file documents a `ReportChartBlock` but it is not confirmed present in `trial_balance.go` as shipped.

> **Status: Planned — Not Yet Implemented**
> `ReportChartBlock` is designed (tasks.md) to add a bar chart of debit vs credit by account group below the trial balance table. It has not been added to `TrialBalanceScreen` yet. When added, it will use `ChartNode` with the `"bar"` series type and reference `${chart_debit_series}` / `${chart_credit_series}` from the report API response.

When `ReportChartBlock` is implemented, the usage pattern will be:

```go
// Planned usage — not yet available
blocks.ReportChartBlock(sess, blocks.ReportChartConfig{
    ChartType: "bar",
    Title:     "Debit vs Credit by Account Group",
    APIURL:    "/api/v1/finance/reports/trial-balance/chart",
})
```

---

## 14.6 Dashboard chart panels

Dashboard chart panels are `ChartNode` instances wrapped inside a titled card with an optional period picker. The `ChartPanelBlock` in `blocks/` provides this pattern.

> **Status: Implemented** — `ChartPanelBlock` is called in `FinanceDashboardScreen` and is confirmed present.

`ChartPanelConfig` fields:

| Field | Type | Purpose |
|---|---|---|
| `Title` | `string` | Card heading |
| `ChartType` | `blocks.ChartType` | One of `ChartTypeLine`, `ChartTypeBar`, `ChartTypePie` |
| `PeriodPicker` | `bool` | Adds a date-range picker above the chart |
| `APIURL` | `string` | Endpoint for chart data |

```go
blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
    Title:        "Revenue vs Expenses",
    ChartType:    blocks.ChartTypeLine,
    PeriodPicker: true,
    APIURL:       "/api/v1/finance/dashboard/revenue-chart",
})
```

When `PeriodPicker` is true, the rendered schema includes a date range control wired to the chart's `API` reload mechanism. The backend receives `period_start` and `period_end` query parameters and returns updated series data.

The `ChartPanelBlock` internally constructs a `ChartNode` with both mandatory transparency fields set. You do not need to set them manually when using `ChartPanelBlock`.

### Placing charts in a grid

Charts rarely occupy the full page width on a dashboard. Use `ast.GridNode` to create a two-column layout:

```go
ast.GridNode{Columns: []ast.GridColumn{
    {
        Body: []ast.Node{
            blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
                Title:     "Revenue vs Expenses",
                ChartType: blocks.ChartTypeLine,
                APIURL:    "/api/v1/finance/dashboard/revenue-chart",
            }),
        },
        MD: 8,
    },
    {
        Body: []ast.Node{
            blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
                Title:     "Expense Breakdown",
                ChartType: blocks.ChartTypePie,
                APIURL:    "/api/v1/finance/dashboard/expense-breakdown",
            }),
        },
        MD: 4,
    },
}}
```

---

## 14.7 Legacy ChartBuilder usage

Before the typed AST was introduced, pages used a `ChartBuilder` helper that constructed `map[string]any` directly. This approach is now deprecated.

> **Legacy pattern — do not use in new code**

```go
// Old approach — PageFn returning map[string]any
func Schema(sess ui.UISessionContext) map[string]any {
    return map[string]any{
        "type": "page",
        "body": []map[string]any{
            {
                "type":   "chart",
                "height": 300,
                "config": map[string]any{
                    "backgroundColor": "transparent",
                    // ...
                },
            },
        },
    }
}
```

**Problems with the legacy approach:**
- The ValidateStage transparency check still runs, but it finds the raw `map[string]any` rather than an `ast.ChartNode`. Depending on registry version, this may silently pass or fail.
- No compile-time type safety on chart fields.
- `"style"` key is frequently omitted because there is no struct to enforce it.

**Migration path:** Replace the `Fn` field in `PageRegistration` with `ASTFn` and return an `ast.PageNode` containing `ast.ChartNode` structs. The `ASTFn` path runs through CompileStage, which applies the ValidateStage checks correctly.

```go
// Migrated — use ASTFn with typed ChartNode
registry.RegisterPage(registry.PageRegistration{
    Route:  "/finance/reports/revenue",
    Module: "finance",
    Title:  "Revenue Report",
    ASTFn: func(sess ui.UISessionContext) any {
        return ast.PageNode{
            Title: "Revenue Report",
            Body: []ast.Node{
                ast.ChartNode{
                    Title:  "Revenue",
                    Height: 300,
                    Style:  map[string]string{"background": "transparent"},
                    Config: map[string]any{
                        "backgroundColor": "transparent",
                        // ...
                    },
                },
            },
        }
    },
})
```

---

### Chart node checklist

- [ ] `Style["background"] = "transparent"` is set
- [ ] `Config["backgroundColor"] = "transparent"` is set (separate from Style)
- [ ] Page registered with `ASTFn`, not legacy `Fn`
- [ ] Dynamic data uses AMIS expression syntax (`"${variable}"`) not Go string concatenation
- [ ] If `API` is set on the chart, `SendOn` limits fetches to when required params are present
- [ ] Pie series data is `[]map[string]any` with `name` and `value` keys
- [ ] No ECharts `color` or `theme` options reference hard-coded values that break dark mode
