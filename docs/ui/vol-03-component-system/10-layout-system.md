# Chapter 10 — Layout System

> **Volume:** III — Component System
> **Audience:** Backend Engineers
> **Prerequisites:** Chapter 09 — Component System

---

## Table of Contents

- [10.1 Layout Principles](#101-layout-principles)
- [10.2 PageNode — The Root Container](#102-pagenode--the-root-container)
- [10.3 GridNode — Column Layout](#103-gridnode--column-layout)
- [10.4 FlexNode — Flex Container](#104-flexnode--flex-container)
- [10.5 TabsNode — Tabbed Content](#105-tabsnode--tabbed-content)
- [10.6 SplitPaneNode — Resizable Panes](#106-splitpanenode--resizable-panes)
- [10.7 SectionNode — Collapsible Sections](#107-sectionnode--collapsible-sections)
- [10.8 Layout Composition Patterns](#108-layout-composition-patterns)

---

## 10.1 Layout Principles

All layout decisions belong in the backend, expressed through layout nodes. The AMIS SDK renders layout components using its own CSS. Page functions do not set pixel widths, CSS classes (except for specific AMIS-documented props), or positioning values.

The layout vocabulary:
- **`PageNode`** — every page has exactly one root `PageNode`
- **`GridNode`** — divide horizontal space into columns (12-column grid)
- **`FlexNode`** — flex row or column for fine-grained alignment
- **`TabsNode`** — group content into switchable tabs
- **`SplitPaneNode`** — resizable horizontal split
- **`SectionNode`** — collapsible labeled section

---

## 10.2 PageNode — The Root Container

Every page compiled by the pipeline has a `PageNode` as its root. `PageNode` compiles to AMIS type `"page"`.

```go
type PageNode struct {
    Title   string
    SubTitle string
    Remark  string           // tooltip next to the title
    InitAPI *APISpec         // API called on mount; result merged into page data scope
    Toolbar []Node           // buttons/actions in the page header toolbar
    Body    []Node           // main content
    AsideBody []Node         // optional sidebar content
}
```

`InitAPI` is optional. When set, the AMIS SDK makes the API call when the page mounts and merges the response into the page's data scope. All child nodes can reference response fields using `${fieldName}` expressions.

```go
ast.PageNode{
    Title:   "Invoice Detail",
    SubTitle: "View and manage invoice",
    InitAPI: &ast.APISpec{
        Method: "get",
        URL:    "/api/v1/finance/invoices/${id}",
    },
    Toolbar: []ast.Node{
        ast.ActionNode{Label: "Back", ActionType: "link", Target: "/finance/invoices"},
    },
    Body: []ast.Node{
        ast.PropertyNode{
            Title: "Invoice Information",
            Items: []ast.PropertyItem{
                {Label: "Invoice #", Content: "${invoice_number}"},
                {Label: "Amount", Content: "${amount}", Type: "currency"},
            },
        },
    },
}
```

`PageNode` compiles to:

```json
{
  "type": "page",
  "title": "Invoice Detail",
  "subTitle": "View and manage invoice",
  "initApi": {"method": "get", "url": "/api/v1/finance/invoices/${id}"},
  "toolbar": [...],
  "body": [...]
}
```

---

## 10.3 GridNode — Column Layout

`GridNode` implements AMIS's 12-column grid. Each column specifies its width using `MD` (medium breakpoint, 1–12).

```go
type GridNode struct {
    Columns []GridColumn
    Gap     int // column gap in pixels
}

type GridColumn struct {
    Body []Node
    MD   int    // 1–12, defaults to equal split
    SM   int    // optional: small breakpoint override
    XS   int    // optional: extra-small breakpoint override
}
```

```go
ast.GridNode{
    Columns: []ast.GridColumn{
        // 2/3 width chart, 1/3 width activity panel
        {Body: []ast.Node{chartNode}, MD: 8},
        {Body: []ast.Node{activityNode}, MD: 4},
    },
}
```

```go
ast.GridNode{
    Columns: []ast.GridColumn{
        // Four equal-width KPI cards
        {Body: []ast.Node{ast.StatNode{Label: "Revenue", Value: "${revenue}"}}, MD: 3},
        {Body: []ast.Node{ast.StatNode{Label: "AR", Value: "${ar}"}}, MD: 3},
        {Body: []ast.Node{ast.StatNode{Label: "AP", Value: "${ap}"}}, MD: 3},
        {Body: []ast.Node{ast.StatNode{Label: "Cash", Value: "${cash}"}}, MD: 3},
    },
}
```

Compiled JSON:

```json
{
  "type": "grid",
  "columns": [
    {"body": [...], "md": 8},
    {"body": [...], "md": 4}
  ]
}
```

**MD values must sum to 12** for a full-width row. For unequal splits, use values like 6+6, 8+4, 4+4+4, 3+3+3+3, etc.

---

## 10.4 FlexNode — Flex Container

`FlexNode` provides CSS flexbox layout for fine-grained alignment of small elements.

```go
type FlexNode struct {
    Body      []Node
    Direction string  // "row"|"column", default "row"
    Justify   string  // "flex-start"|"flex-end"|"center"|"space-between"|"space-around"
    Align     string  // "flex-start"|"flex-end"|"center"|"stretch"
    Gap       int     // gap between items in pixels
}
```

```go
ast.FlexNode{
    Direction: "row",
    Justify:   "space-between",
    Align:     "center",
    Body: []ast.Node{
        ast.ActionNode{Label: "Approve", ActionType: "ajax", Level: "success"},
        ast.ActionNode{Label: "Reject", ActionType: "ajax", Level: "danger"},
        ast.ActionNode{Label: "Defer", ActionType: "ajax", Level: "default"},
    },
}
```

Use `FlexNode` for action bars and small aligned groups. Use `GridNode` for page-level column layouts.

---

## 10.5 TabsNode — Tabbed Content

`TabsNode` groups content into switchable tabs.

```go
type TabsNode struct {
    Tabs          []TabItem
    MountOnEnter  bool  // defer rendering until tab is first selected
    UnmountOnExit bool  // destroy content when tab is deselected
    DefaultKey    string // key of the initially active tab
}

type TabItem struct {
    Title   string
    Key     string  // unique identifier for the tab
    Body    []Node
    Icon    string  // optional FontAwesome class
    Badge   string  // optional badge text
}
```

```go
ast.TabsNode{
    MountOnEnter:  true,   // defer CRUD API calls until the tab is selected
    UnmountOnExit: false,  // keep state when switching tabs
    Tabs: []ast.TabItem{
        {
            Title: "Invoices",
            Key:   "invoices",
            Icon:  "fa fa-file-invoice",
            Body:  []ast.Node{invoiceCRUD},
        },
        {
            Title: "Payments",
            Key:   "payments",
            Body:  []ast.Node{paymentCRUD},
        },
    },
}
```

> **Critical:** Use `MountOnEnter` and `UnmountOnExit`. Do not use `Mountable` — it is not an AMIS tabs field. `Mountable` on a `TabsNode` struct is silently ignored by AMIS.

**`MountOnEnter: true`** is important for tabs containing `CRUDNode` components. Without it, all CRUD components make their API calls when the page loads, even for hidden tabs. With `MountOnEnter: true`, the API call is deferred until the user selects the tab.

Compiled JSON:

```json
{
  "type": "tabs",
  "mountOnEnter": true,
  "unmountOnExit": false,
  "tabs": [
    {"title": "Invoices", "key": "invoices", "body": [...]},
    {"title": "Payments", "key": "payments", "body": [...]}
  ]
}
```

---

## 10.6 SplitPaneNode — Resizable Panes

`SplitPaneNode` creates a horizontally split view where the user can drag the divider to resize the panes.

```go
type SplitPaneNode struct {
    Left     []Node
    Right    []Node
    MinLeft  int   // minimum left pane width in pixels
    MinRight int   // minimum right pane width in pixels
}
```

```go
ast.SplitPaneNode{
    MinLeft:  300,
    MinRight: 400,
    Left:  []ast.Node{invoiceListCRUD},
    Right: []ast.Node{invoiceDetailProperty},
}
```

Use for master-detail layouts where the list and detail are shown side by side.

---

## 10.7 SectionNode — Collapsible Sections

`SectionNode` creates a collapsible, labeled section within a page or form.

```go
type SectionNode struct {
    Title       string
    Body        []Node
    Collapsible bool
    Collapsed   bool  // initial collapsed state
}
```

```go
ast.SectionNode{
    Title:       "Line Items",
    Collapsible: true,
    Collapsed:   false,
    Body:        []ast.Node{lineItemsTable},
}
```

Compiles to AMIS type `"fieldset"`:

```json
{
  "type": "fieldset",
  "title": "Line Items",
  "collapsible": true,
  "collapsed": false,
  "body": [...]
}
```

---

## 10.8 Layout Composition Patterns

### Pattern 1: Dashboard with KPI Row and Tabbed Tables

```go
ast.PageNode{
    Title:   "Finance Dashboard",
    InitAPI: &ast.APISpec{Method: "get", URL: "/api/v1/finance/dashboard/summary"},
    Body: []ast.Node{
        // KPI row
        ast.GridNode{Columns: []ast.GridColumn{
            {Body: []ast.Node{ast.StatNode{Label: "Revenue", Value: "${revenue}"}}, MD: 3},
            {Body: []ast.Node{ast.StatNode{Label: "AR Balance", Value: "${ar}"}}, MD: 3},
            {Body: []ast.Node{ast.StatNode{Label: "AP Balance", Value: "${ap}"}}, MD: 3},
            {Body: []ast.Node{ast.StatNode{Label: "Cash", Value: "${cash}"}}, MD: 3},
        }},
        // Chart + activity side by side
        ast.GridNode{Columns: []ast.GridColumn{
            {Body: []ast.Node{revenueChart}, MD: 8},
            {Body: []ast.Node{activityTimeline}, MD: 4},
        }},
        // Tabbed data tables
        ast.TabsNode{
            MountOnEnter: true,
            Tabs: []ast.TabItem{
                {Title: "Invoices", Body: []ast.Node{invoiceCRUD}},
                {Title: "Expenses", Body: []ast.Node{expenseCRUD}},
            },
        },
    },
}
```

### Pattern 2: Form with Sections

```go
ast.PageNode{
    Title: "New Invoice",
    Body: []ast.Node{
        ast.FormNode{
            API: &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices"},
            Body: []ast.Node{
                ast.SectionNode{Title: "Header Information", Body: []ast.Node{
                    ast.InputTextNode{Name: "invoice_number", Label: "Invoice #", Required: true},
                    ast.SelectNode{Name: "vendor_id", Label: "Vendor",
                        Source: "/api/v1/vendors?fields=id,name"},
                    ast.InputDateNode{Name: "due_date", Label: "Due Date"},
                }},
                ast.SectionNode{Title: "Line Items", Collapsible: false, Body: []ast.Node{
                    lineItemsGrid,
                }},
            },
            Actions: []ast.Node{
                ast.ActionNode{Label: "Submit", ActionType: "submit", Level: "primary"},
                ast.ActionNode{Label: "Save Draft", ActionType: "ajax", Level: "default",
                    API: &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/draft"}},
                ast.ActionNode{Label: "Cancel", ActionType: "link", Target: "/finance/invoices"},
            },
        },
    },
}
```

### Pattern 3: Master-Detail Split

```go
ast.PageNode{
    Title: "Purchase Orders",
    Body: []ast.Node{
        ast.SplitPaneNode{
            MinLeft:  350,
            MinRight: 500,
            Left: []ast.Node{
                ast.CRUDNode{
                    API:     ast.APISpec{Method: "get", URL: "/api/v1/procurement/purchase-orders"},
                    Columns: []ast.TableColumn{
                        {Name: "po_number", Label: "PO #"},
                        {Name: "vendor_name", Label: "Vendor"},
                    },
                },
            },
            Right: []ast.Node{
                ast.PropertyNode{Title: "Order Details", Items: []ast.PropertyItem{
                    {Label: "PO Number", Content: "${po_number}"},
                    {Label: "Vendor", Content: "${vendor_name}"},
                    {Label: "Total", Content: "${total_amount}", Type: "currency"},
                }},
            },
        },
    },
}
```

---

*End of Chapter 10*

**Previous:** [Chapter 09 — Component System](./09-component-system.md)
**Next:** [Chapter 11 — Forms Framework](./11-forms-framework.md)
