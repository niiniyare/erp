> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# Chapter 09 — Component System

> **Volume:** III — Component System
> **Audience:** Backend Engineers, Platform Engineers
> **Prerequisites:** Chapter 05 — SDUI Fundamentals, Chapter 07 — AST Design

---

## Table of Contents

- [9.1 Component Categories](#91-component-categories)
- [9.2 The Two Builder Paths](#92-the-two-builder-paths)
- [9.3 AMIS Type Strings](#93-amis-type-strings)
- [9.4 Display Components](#94-display-components)
- [9.5 Data Components](#95-data-components)
- [9.6 Action Components](#96-action-components)
- [9.7 Overlay Components](#97-overlay-components)
- [9.8 Composing Components in Practice](#98-composing-components-in-practice)
- [9.9 Common Mistakes](#99-common-mistakes)

---

## 9.1 Component Categories

AwoERP's component system is built on AMIS component types. Every component has a `"type"` field that the AMIS SDK uses to select its renderer.

The component categories are:

| Category | Examples | Go Representation |
|----------|----------|------------------|
| **Layout** | `page`, `grid`, `flex`, `tabs` | AST: `PageNode`, `GridNode`, `TabsNode` |
| **Data Display** | `stat`, `table`, `timeline`, `property`, `mapping` | AST: `StatNode`, `TableNode`, `PropertyNode`, `MappingNode` |
| **Data** | `crud`, `chart` | AST: `CRUDNode`, `ChartNode` |
| **Form** | `form`, `input-text`, `input-date`, `select` | AST: `FormNode`, `InputTextNode`, `SelectNode` |
| **Actions** | `button` | AST: `ActionNode` |
| **Overlays** | `dialog`, `drawer` | AST: `DialogNode`, `DrawerNode` |
| **Content** | `tpl`, `alert`, `divider` | Raw `ui.M{}` literals |

---

## 9.2 The Two Builder Paths

### Path 1: Typed AST (Preferred)

Use AST node structs for all new pages. The typed AST gives compile-time correctness and `CompileTree` validation.

```go
// Typed AST — preferred
ast.PageNode{
    Title: "My Page",
    Body: []ast.Node{
        ast.StatNode{Label: "Revenue", Value: "${revenue}"},
    },
}
```

### Path 2: AMIS Builder Package (Legacy)

Use fluent builders from `internal/web/amis/` for existing `PageFn` pages during migration.

```go
// Legacy amis builders — for PageFn migrations only
amis.Page("My Page").
    Body(
        amis.Stat("Revenue", "${revenue}"),
    )
```

Builders implement `json.Marshaler` — no `.Build()` call needed. `Build()` exists for embedding into larger `M{}` literals.

### Mixing Paths

During migration, a screen may use the AMIS builder package inside a `PageFn` while the rest of the codebase uses AST. This is acceptable. Do not mix both inside a single `ASTPageFn`.

---

## 9.3 AMIS Type Strings

AMIS renders components based on the `"type"` field. NormalizeStage lowercases all type values, so both `"Page"` and `"page"` will work — but write lowercase at the source.

Validated type strings (from the AMIS SDK):

```
Layout:      page, grid, flex, tabs, splitpane, fieldset
Display:     stat, statistic, table, card, timeline, tree, mapping, property
Data:        crud, chart
Form:        form, wizard, filter
Inputs:      input-text, input-number, textarea, input-date, input-datetime,
             input-date-range, select, multi-select, switch, checkbox, radios,
             input-file, input-image, hidden
Actions:     button
Overlays:    dialog, drawer
Content:     tpl, alert, remark, divider, static, iframe
```

> **Warning:** `"status"` is not a valid AMIS type. Use `"mapping"` for status badges. This is a common mistake — see §9.9.

---

## 9.4 Display Components

### StatNode / Stat Builder

KPI number display with optional trend indicator. Used for dashboards.

```go
// AST
ast.StatNode{
    Label:  "Total Revenue",
    Value:  "${revenue}",     // AMIS expression from page data scope
    Format: "currency",
    Trend:  "up",             // "up"|"down"|"" — optional trend arrow
}
```

```go
// Legacy amis builder
amis.Stat("Total Revenue", "${revenue}")
```

> **Note:** StatNode uses AMIS CSS custom property tokens for styling, not custom class names. Do not set `className` on stat nodes — AMIS manages their appearance.

### MappingNode

Maps data values to display labels or HTML. The correct component for status badges.

```go
ast.MappingNode{
    Name:  "status",
    Label: "Status",
    Map: map[string]string{
        "DRAFT":    `<span class="badge badge-default">Draft</span>`,
        "PENDING":  `<span class="badge badge-warning">Pending</span>`,
        "APPROVED": `<span class="badge badge-success">Approved</span>`,
        "*":        `<span class="badge badge-default">${value}</span>`,
    },
}
```

The `"*"` key is a fallback rendered when no other key matches.

### PropertyNode

Key-value property list for record detail views.

```go
ast.PropertyNode{
    Title: "Invoice Details",
    Items: []ast.PropertyItem{
        {Label: "Invoice Number", Content: "${invoice_number}"},
        {Label: "Vendor", Content: "${vendor_name}"},
        {Label: "Amount", Content: "${amount}", Type: "currency"},
        {Label: "Status", Content: "${status}"},
    },
}
```

> **Note:** Use `PropertyNode` (AMIS type `"property"`) for detail cards, not a single-row table. `PropertyNode` provides proper label/value layout; a single-row `TableNode` does not.

### TimelineNode

Chronological event list. Used for audit trails and workflow history.

```go
ast.TimelineNode{
    Source: "${timeline_items}",
    Items: []ast.TimelineItem{
        // Static items if not loaded from source
        {Time: "2026-06-01", Label: "Invoice Created", Icon: "fa fa-file"},
        {Time: "2026-06-02", Label: "Submitted for Approval", Icon: "fa fa-paper-plane"},
    },
}
```

---

## 9.5 Data Components

### CRUDNode

The primary component for listing, filtering, sorting, and paginating server-side data.

```go
ast.CRUDNode{
    API: ast.APISpec{
        Method: "get",
        URL:    "/api/v1/finance/invoices",
    },
    SyncLocation: false,      // always false — required by ValidateStage
    PrimaryKey:   "id",
    PageSize:     25,
    Columns: []ast.TableColumn{
        {Name: "invoice_number", Label: "Invoice #", Sortable: true},
        {Name: "amount", Label: "Amount", Type: "currency"},
        {Name: "status", Label: "Status", Type: "mapping", Map: statusMap},
    },
    Toolbar: []ast.Node{
        ast.ActionNode{Label: "New", ActionType: "link", Target: "/finance/invoices/new", Level: "primary"},
    },
    RowActions: []ast.ActionNode{
        {Label: "View", ActionType: "link", Target: "/finance/invoices/${id}"},
        {Label: "Delete", ActionType: "ajax", Level: "danger",
            API: &ast.APISpec{Method: "delete", URL: "/api/v1/finance/invoices/${id}"},
            ConfirmText: "Delete this invoice?"},
    },
}
```

**`SyncLocation: false` is mandatory.** AMIS CRUD syncs filter state to the URL by default. This causes unexpected browser history entries and breaks the back button. ValidateStage enforces `SyncLocation: false` on all CRUD nodes.

**`RowActions` must be included in `Children()`** for tree validation. `CRUDNode` satisfies this by including all `RowActions` as children.

### ChartNode

ECharts chart. Requires transparent background in two places.

```go
ast.ChartNode{
    Height: 300,
    API: &ast.APISpec{
        Method: "get",
        URL:    "/api/v1/finance/dashboard/revenue-chart",
    },
    // Both Style and Config background are required — ValidateStage checks both
    Style: map[string]any{
        "background": "transparent",
    },
    Config: map[string]any{
        "backgroundColor": "transparent",
        "xAxis": map[string]any{"type": "category", "data": "${months}"},
        "yAxis": map[string]any{"type": "value"},
        "series": []any{
            map[string]any{"type": "line", "data": "${revenue}"},
        },
    },
}
```

> **Warning:** Both `style.background` and `config.backgroundColor` must be `"transparent"`. AMIS applies background from both locations. Missing either causes a white box in dark-mode environments. ValidateStage returns `AST_CHART_OPAQUE_BG` if either is missing.

---

## 9.6 Action Components

### ActionNode

Covers all interactive buttons: links, AJAX calls, dialogs, drawers, URL navigation.

```go
// Link to another route
ast.ActionNode{Label: "View", ActionType: "link", Target: "/finance/invoices/${id}"}

// AJAX call
ast.ActionNode{
    Label:      "Delete",
    ActionType: "ajax",
    Level:      "danger",
    API:        &ast.APISpec{Method: "delete", URL: "/api/v1/finance/invoices/${id}"},
    ConfirmText: "Delete this invoice?",
}

// Open a dialog
ast.ActionNode{
    Label:      "Approve",
    ActionType: "dialog",
    Level:      "success",
    DisabledOn: "${!can_approve}",  // AMIS expression — ! inside ${}
    Dialog: &ast.DialogNode{
        Title: "Approve Invoice",
        Body: []ast.Node{
            ast.FormNode{
                API: &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/${id}/approve"},
                Body: []ast.Node{
                    ast.InputTextNode{Name: "note", Label: "Approval Note"},
                },
            },
        },
    },
}
```

**Level values:** `"primary"`, `"success"`, `"warning"`, `"danger"`, `"default"`

---

## 9.7 Overlay Components

### DialogNode

Modal dialog. Always referenced from an `ActionNode.Dialog` field, never used standalone.

```go
ast.DialogNode{
    Title: "Edit Invoice",
    Size:  "lg", // "sm"|"md"|"lg"|"xl"
    Body: []ast.Node{
        ast.FormNode{
            API: &ast.APISpec{Method: "put", URL: "/api/v1/finance/invoices/${id}"},
            Body: []ast.Node{
                ast.InputTextNode{Name: "invoice_number", Label: "Invoice #"},
                ast.SelectNode{
                    Name: "status",
                    Label: "Status",
                    Options: []ast.SelectOption{
                        {Label: "Draft", Value: "DRAFT"},
                        {Label: "Pending", Value: "PENDING"},
                    },
                },
            },
        },
    },
}
```

### DrawerNode

Side panel drawer. Used for detail views alongside the list.

```go
ast.DrawerNode{
    Title:    "Invoice Detail",
    Position: "right", // "right"|"left"|"top"|"bottom"
    Size:     "md",
    Body: []ast.Node{
        ast.PropertyNode{...},
    },
}
```

---

## 9.8 Composing Components in Practice

Components compose by nesting `ast.Node` values inside `Body`, `Columns`, `Toolbar`, etc.

A typical dashboard page:

```go
ast.PageNode{
    Title: "Finance Dashboard",
    InitAPI: &ast.APISpec{Method: "get", URL: "/api/v1/finance/dashboard/summary"},
    Body: []ast.Node{
        // Row of KPI stats
        ast.GridNode{Columns: []ast.GridColumn{
            {Body: []ast.Node{ast.StatNode{Label: "Revenue", Value: "${revenue}"}}, MD: 3},
            {Body: []ast.Node{ast.StatNode{Label: "AR Balance", Value: "${ar_balance}"}}, MD: 3},
            {Body: []ast.Node{ast.StatNode{Label: "AP Balance", Value: "${ap_balance}"}}, MD: 3},
            {Body: []ast.Node{ast.StatNode{Label: "Cash", Value: "${cash}"}}, MD: 3},
        }},
        // Tabbed content below
        ast.TabsNode{
            MountOnEnter:  true,
            UnmountOnExit: false,
            Tabs: []ast.TabItem{
                {Title: "Invoices", Body: []ast.Node{ast.CRUDNode{...}}},
                {Title: "Payments", Body: []ast.Node{ast.CRUDNode{...}}},
            },
        },
    },
}
```

**`TabsNode` uses `MountOnEnter`/`UnmountOnExit`**, not `Mountable`. `Mountable` is not an AMIS tabs field. Setting it has no effect. `MountOnEnter: true` defers rendering the tab's content until the user first selects it — important for performance when tabs contain CRUD components with API calls.

---

## 9.9 Common Mistakes

| Mistake | Correct Approach |
|---------|-----------------|
| Using `"status"` as column type | Use `"mapping"` — `"status"` is not a valid AMIS type |
| Using `"property"` type incorrectly | Use `PropertyNode` or `ast.M{"type": "property"}` with correct `items` structure |
| `DisabledOn: "!${can_approve}"` | Use `DisabledOn: "${!can_approve}"` — `!` must be inside the expression |
| `ChartNode` missing `config.backgroundColor` | Set both `style.background` and `config.backgroundColor` to `"transparent"` |
| `CRUDNode` with `SyncLocation: true` | Always use `SyncLocation: false` |
| `TabsNode` with `Mountable` field | Use `MountOnEnter` and `UnmountOnExit` |
| `DataTableBlock` with no `CreatePermission` | Provide `CreatePermission` when `AllowCreate` is true, or the button is always hidden |
| `DetailCardBlock` with single-row table | Use `PropertyNode` / `DetailCardBlock` which emits `"property"` type |
| `ActionNode` dialog type with no `Dialog` field | Always set `ActionNode.Dialog` when `ActionType` is `"dialog"` |

---

*End of Chapter 09*

**Previous:** [Chapter 08 — The Compilation Pipeline](../vol-02-dsl-and-ast/08-compilation-pipeline.md)
**Next:** [Chapter 10 — Layout System](./10-layout-system.md)
