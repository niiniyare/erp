# Chapter 12 — Tables and Data Grids

> **Volume:** III — Component System
> **Audience:** Backend Engineers
> **Prerequisites:** Chapter 11 — Forms Framework

---

## Table of Contents

- [12.1 CRUDNode vs TableNode](#121-crudnode-vs-tablenode)
- [12.2 CRUDNode In Depth](#122-crudnode-in-depth)
- [12.3 Column Types](#123-column-types)
- [12.4 Toolbar Buttons](#124-toolbar-buttons)
- [12.5 Row Actions](#125-row-actions)
- [12.6 Bulk Actions](#126-bulk-actions)
- [12.7 DataTableBlock — The DSL Shortcut](#127-datatable-block--the-dsl-shortcut)
- [12.8 StatusBadge Columns](#128-statusbadge-columns)
- [12.9 Pagination and Sorting](#129-pagination-and-sorting)
- [12.10 The amis.CRUD Builder (Legacy)](#1210-the-amiscrud-builder-legacy)

---

## 12.1 CRUDNode vs TableNode

Two node types render tabular data:

| | `CRUDNode` | `TableNode` |
|---|---|---|
| **AMIS type** | `"crud"` | `"table"` |
| **Data source** | Server API (paginated) | Inline static data or page scope |
| **Pagination** | Yes (server-side) | No |
| **Sorting** | Yes (sends `orderBy` to API) | No |
| **Filter** | Yes (sends filter params to API) | No |
| **Use case** | All listing pages | Small static datasets, embedded previews |

Use `CRUDNode` for all server-backed listing pages. Use `TableNode` for small, pre-loaded datasets (e.g., a summary table embedded in a detail view).

---

## 12.2 CRUDNode In Depth

```go
type CRUDNode struct {
    API          APISpec         // required: list endpoint
    SyncLocation bool            // must be false — ValidateStage enforces this
    PrimaryKey   string          // default "id"
    PageSize     int             // default 10
    Columns      []TableColumn   // required
    Filter       Node            // optional: filter bar above the table
    Toolbar      []Node          // optional: buttons above the table (right side)
    BulkActions  []Node          // optional: actions for selected rows
    RowActions   []ActionNode    // optional: per-row action buttons
    Title        string          // optional: panel title
    DefaultSort  string          // optional: field name for default sort
    DefaultOrder string          // optional: "asc"|"desc"
}
```

**`SyncLocation` must always be `false`.**

AMIS CRUD syncs its filter state to the browser URL by default (`syncLocation: true`). This causes:
- Unexpected browser history entries (every filter change pushes to history)
- Back-button confusion
- URL pollution in single-page app navigation

ValidateStage enforces `syncLocation: false` on all CRUD nodes. The AST `SyncLocation` field defaults to `false`; set it explicitly for clarity.

---

## 12.3 Column Types

Each `TableColumn` has a `Type` field that controls how the AMIS SDK renders the cell.

```go
type TableColumn struct {
    Name     string            // field name in the API response
    Label    string            // column header text
    Type     string            // renderer type (see below)
    Map      map[string]string // for "mapping" type
    Width    int               // column width in pixels
    Sortable bool
    Fixed    string            // ""|"left"|"right"
}
```

**Supported column types:**

| Type | Renders | Notes |
|------|---------|-------|
| `""` or `"text"` | Plain string | Default |
| `"date"` | Formatted date | Uses browser locale |
| `"datetime"` | Formatted date+time | |
| `"number"` | Formatted number | |
| `"currency"` | Currency-formatted number | Add `prefix`/`suffix` for symbol |
| `"mapping"` | Value mapped to HTML | Used for status badges |
| `"link"` | Clickable link | Requires `href` or `to` field in column config |
| `"image"` | Image thumbnail | |
| `"tpl"` | Template with expressions | Use `${fieldName}` in the label |

> **Warning:** `"status"` is not a valid AMIS column type. It appears in some AMIS examples but does not render correctly. Use `"mapping"` for status badges.

### Examples

```go
// Plain text
ast.TableColumn{Name: "invoice_number", Label: "Invoice #", Sortable: true}

// Currency (number formatted as currency)
ast.TableColumn{Name: "amount", Label: "Amount", Type: "currency"}

// Date
ast.TableColumn{Name: "due_date", Label: "Due Date", Type: "date"}

// Mapping (status badge)
ast.TableColumn{
    Name:  "status",
    Label: "Status",
    Type:  "mapping",
    Map: map[string]string{
        "DRAFT":    `<span class="badge badge-default">Draft</span>`,
        "PENDING":  `<span class="badge badge-warning">Pending</span>`,
        "APPROVED": `<span class="badge badge-success">Approved</span>`,
        "*":        `<span class="badge badge-default">${value}</span>`,
    },
}

// Fixed left column (for wide tables)
ast.TableColumn{Name: "invoice_number", Label: "Invoice #", Fixed: "left", Width: 150}
```

---

## 12.4 Toolbar Buttons

The `Toolbar` field places buttons at the top of the CRUD component (header toolbar). Typical usage: "New" button.

```go
ast.CRUDNode{
    API: ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices"},
    Toolbar: []ast.Node{
        ast.ActionNode{
            Label:      "New Invoice",
            ActionType: "link",
            Target:     "/finance/invoices/new",
            Level:      "primary",
            Icon:       "fa fa-plus",
        },
        ast.ActionNode{
            Label:      "Export",
            ActionType: "ajax",
            Level:      "default",
            Icon:       "fa fa-download",
            API:        &ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/export"},
        },
    },
    Columns: []ast.TableColumn{...},
}
```

Permission-controlled toolbar — only include the "New" button if the user can create:

```go
var toolbar []ast.Node
if sess.Can("create", "finance.invoices") {
    toolbar = append(toolbar, ast.ActionNode{
        Label:      "New Invoice",
        ActionType: "link",
        Target:     "/finance/invoices/new",
        Level:      "primary",
        Icon:       "fa fa-plus",
    })
}
```

---

## 12.5 Row Actions

Row actions are per-row buttons rendered in a dedicated column. Each row typically has 1–3 actions.

```go
ast.CRUDNode{
    RowActions: []ast.ActionNode{
        // View: navigate to detail page
        {Label: "View", ActionType: "link", Target: "/finance/invoices/${id}"},

        // Edit: open inline dialog
        {
            Label:      "Edit",
            ActionType: "dialog",
            Icon:       "fa fa-pencil",
            VisibleOn:  "${status === 'DRAFT'}",
            Dialog: &ast.DialogNode{
                Title: "Edit Invoice",
                Body: []ast.Node{
                    ast.FormNode{
                        InitAPI: &ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/${id}"},
                        API:     &ast.APISpec{Method: "put", URL: "/api/v1/finance/invoices/${id}"},
                        Body: []ast.Node{
                            ast.InputTextNode{Name: "invoice_number", Label: "Invoice #"},
                        },
                    },
                },
            },
        },

        // Delete: with confirmation
        {
            Label:       "Delete",
            ActionType:  "ajax",
            Level:       "danger",
            Icon:        "fa fa-trash",
            VisibleOn:   "${status === 'DRAFT'}",
            ConfirmText: "Delete this invoice? This cannot be undone.",
            API:         &ast.APISpec{Method: "delete", URL: "/api/v1/finance/invoices/${id}"},
        },
    },
}
```

**`RowActions` are included in `CRUDNode.Children()`** for `CompileTree` validation. `DialogNode` and `DrawerNode` inside row actions are validated along with the rest of the tree.

---

## 12.6 Bulk Actions

Bulk actions appear when the user selects one or more rows via checkboxes.

```go
ast.CRUDNode{
    BulkActions: []ast.Node{
        ast.ActionNode{
            Label:       "Delete Selected",
            ActionType:  "ajax",
            Level:       "danger",
            ConfirmText: "Delete all selected invoices?",
            API:         &ast.APISpec{Method: "delete", URL: "/api/v1/finance/invoices/bulk"},
        },
        ast.ActionNode{
            Label:      "Export Selected",
            ActionType: "ajax",
            Level:      "default",
            API:        &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/bulk-export"},
        },
    },
}
```

AMIS includes the selected row IDs in the bulk action request body automatically.

For permission-filtered bulk actions, filter in the page function:

```go
var bulkActions []ast.Node
if sess.Can("delete", "finance.invoices") {
    bulkActions = append(bulkActions, ast.ActionNode{
        Label:       "Delete Selected",
        ActionType:  "ajax",
        Level:       "danger",
        ConfirmText: "Delete selected invoices?",
        API:         &ast.APISpec{Method: "delete", URL: "/api/v1/finance/invoices/bulk"},
    })
}
```

---

## 12.7 DataTableBlock — The DSL Shortcut

`DataTableBlock` from `internal/web/dsl/blocks/data_table.go` composes all of the above into a single call. It handles permission filtering for the "New" button automatically.

```go
blocks.DataTableBlock(sess, blocks.DataTableConfig{
    APIURL:           "/api/v1/finance/invoices",
    PrimaryKey:       "id",
    PageSize:         25,
    Title:            "Invoices",
    AllowCreate:      true,
    CreateURL:        "/finance/invoices/new",
    CreatePermission: "finance.invoices.create", // required if AllowCreate is true
    AllowExport:      true,
    Columns: []blocks.ColumnDef{
        {Name: "invoice_number", Label: "Invoice #", Sortable: true, Width: 150},
        {Name: "vendor_name", Label: "Vendor"},
        {Name: "amount", Label: "Amount", Type: "currency"},
        {Name: "due_date", Label: "Due Date", Type: "date"},
        blocks.StatusBadgeColumnDef(blocks.StatusBadgeConfig{
            FieldName: "status",
            Label:     "Status",
            Mappings:  invoiceStatusMappings,
        }),
    },
    RowActions: []ast.ActionNode{
        {Label: "View", ActionType: "link", Target: "/finance/invoices/${id}"},
    },
    BulkActions: []blocks.BulkActionDef{
        {Label: "Delete", Permission: "finance.invoices.delete",
            APIURL: "/api/v1/finance/invoices/bulk", APIMethod: "delete",
            Level: "danger", Confirm: "Delete selected?"},
    },
})
```

`DataTableBlock` returns an `ast.CRUDNode` with:
- `SyncLocation: false` (always)
- "New" button shown only when `sess.Can(...)` for `CreatePermission`
- Bulk actions filtered by their individual `Permission` fields
- Export button included when `AllowExport` is true

> **Important:** `CreatePermission` is required when `AllowCreate` is true. Without it, the button is never shown — the function checks `cfg.CreatePermission == ""` and treats an empty permission as "not authorized." This is intentional: explicit permission declaration is required, not derived from the URL.

---

## 12.8 StatusBadge Columns

Status badge columns use AMIS `"mapping"` type to render colored HTML badges.

```go
// StatusBadgeColumnDef — returns ColumnDef for DataTableConfig.Columns
blocks.StatusBadgeColumnDef(blocks.StatusBadgeConfig{
    FieldName: "status",
    Label:     "Status",
    Mappings: []blocks.StatusMapping{
        {Value: "DRAFT", Label: "Draft", Color: "default"},
        {Value: "PENDING", Label: "Pending Approval", Color: "warning"},
        {Value: "APPROVED", Label: "Approved", Color: "success"},
        {Value: "PAID", Label: "Paid", Color: "success"},
        {Value: "OVERDUE", Label: "Overdue", Color: "danger"},
        {Value: "CANCELLED", Label: "Cancelled", Color: "danger"},
    },
})
```

This produces a `ColumnDef` with:
- `Type: "mapping"`
- `Map: map[string]string{...}` — each value maps to an HTML `<span class="badge badge-{color}">{label}</span>`
- `"*"` fallback: `<span class="badge badge-default">${value}</span>`

Available colors: `"success"`, `"warning"`, `"danger"`, `"info"`, `"default"`.

The underlying `buildMap` function generates:

```json
{
  "DRAFT": "<span class=\"badge badge-default\">Draft</span>",
  "PENDING": "<span class=\"badge badge-warning\">Pending Approval</span>",
  "*": "<span class=\"badge badge-default\">${value}</span>"
}
```

---

## 12.9 Pagination and Sorting

AMIS CRUD handles pagination and sorting automatically. The SDK appends standard query parameters to the API URL:

- `page=1` — current page (1-indexed)
- `perPage=25` — items per page
- `orderBy=invoice_number` — sort column (when user clicks a sortable column header)
- `orderDir=asc` — sort direction
- Plus any active filter values

The business API must read these parameters and return a paginated response:

```json
{
  "items": [...],
  "total": 347,
  "page": 1,
  "perPage": 25
}
```

AMIS reads `total` to calculate the page count and render the pagination control.

For sortable columns, set `Sortable: true` on the column definition. AMIS sends `orderBy=<column.name>` to the API when the user clicks the header.

---

## 12.10 The amis.CRUD Builder (Legacy)

The `amis.CRUD` builder is available for legacy `PageFn` pages.

```go
// Legacy amis.CRUD builder
func InvoiceListLegacy(sess ui.UISessionContext) ui.Schema {
    toolbar := []any{}
    if sess.Can("create", "finance.invoices") {
        toolbar = append(toolbar, amis.CreateBtn("New Invoice", "post:/api/v1/finance/invoices",
            amis.Required(amis.TextField("invoice_number", "Invoice #")),
        ))
    }

    return ui.M{
        "type": "page",
        "title": "Invoices",
        "body": amis.CRUD("get:/api/v1/finance/invoices").
            Columns(
                amis.Column("invoice_number", "Invoice #").Sortable(),
                amis.Column("vendor_name", "Vendor"),
                amis.Column("amount", "Amount").Type("currency"),
                amis.Column("status", "Status").Map(ui.M{
                    "DRAFT":    `<span class="badge">Draft</span>`,
                    "PENDING":  `<span class="badge badge-warning">Pending</span>`,
                    "APPROVED": `<span class="badge badge-success">Approved</span>`,
                }),
            ).
            Toolbar(toolbar...).
            PerPage(25).
            Build(),
    }
}
```

Note: `amis.CRUD` sets `syncLocation: true` by default. ValidateStage will reject this. Legacy pages using `amis.CRUD` must explicitly override it:

```go
// Override syncLocation in the amis builder output
schema := amis.CRUD("get:/api/v1/finance/invoices").Build()
schema["syncLocation"] = false  // required: ValidateStage enforces this
```

This is one reason `DataTableBlock` and the typed AST are preferred — `DataTableBlock` always sets `SyncLocation: false`.

---

*End of Chapter 12*

**Previous:** [Chapter 11 — Forms Framework](./11-forms-framework.md)
