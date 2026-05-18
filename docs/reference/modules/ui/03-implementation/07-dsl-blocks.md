# DSL Blocks Reference

> Last verified: 2026-05-18 | Code pointer: `internal/web/dsl/blocks/`, `internal/web/ast/`

> **Partial listing.** This file documents 12 confirmed blocks. Run `ls internal/web/dsl/blocks/*.go` for the full list (~26 total). Add new blocks here when you add new files.

---

## What Is a UIBlock?

A `UIBlock` is a reusable schema fragment:

```go
type UIBlock func(sess UISessionContext) M
```

Blocks take a `UISessionContext` (and usually a config struct) and return an `ast.Node`. They are not full pages — they compose into pages via `ast.GridNode`, `ast.SectionNode`, etc.

**Rule:** Never inline ERP domain logic (line items, addresses, approval workflows) directly in a screen file. Use the block. This ensures consistency across documents and makes domain changes affect all pages at once.

---

## Document Blocks

These blocks appear in transactional document forms (invoices, purchase orders, bills, GRNs).

### `DocumentHeaderBlock`

Source: `internal/web/dsl/blocks/document_header.go`

The top section of any document form. Renders reference number, date, due date, and optional currency/status fields.

```go
blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
    ResourceURL:  "/api/v1/finance/invoices",
    ShowCurrency: true,
    ShowStatus:   true,
    StatusOptions: []ast.SelectOption{
        {Label: "Draft",    Value: "draft"},
        {Label: "Sent",     Value: "sent"},
        {Label: "Paid",     Value: "paid"},
    },
    ReadOnly: !sess.Can("update", "invoice"),
})
```

| Config Field | Type | Purpose |
|---|---|---|
| `ResourceURL` | `string` | API URL — used for future deep-link support |
| `ShowCurrency` | `bool` | Adds currency selector, defaults to `sess.Currency` |
| `ShowStatus` | `bool` | Adds status selector if `StatusOptions` non-empty |
| `StatusOptions` | `[]ast.SelectOption` | Enumerated status values |
| `ReadOnly` | `bool` | All fields disabled when true |

Rendered fields: `ref_number`, `document_date`, `due_date`, `currency` (optional), `status` (optional).

---

### `ProductServiceLineBlock`

Source: `internal/web/dsl/blocks/line_items.go`

Line items combo for commercial documents. **Every document with line items must use this block.** Never define a custom line item table in a screen file — changes to line item logic must be made here once.

```go
// Standard invoice / PO / bill:
blocks.ProductServiceLineBlock(sess, blocks.DefaultLineItemConfig())

// Goods receipt note (no pricing):
blocks.ProductServiceLineBlock(sess, blocks.GRNLineItemConfig())

// Journal entry (accounts + debit/credit):
blocks.ProductServiceLineBlock(sess, blocks.JournalLineItemConfig())

// Custom:
blocks.ProductServiceLineBlock(sess, blocks.LineItemConfig{
    ShowProductCode:   true,
    ShowDescription:   true,
    ShowQty:           true,
    ShowUnitPrice:     true,
    ShowDiscount:      true,
    ShowTaxRate:       true,
    ShowSubtotal:      true,
    MaxLines:          50,
    ReadOnly:          !sess.Can("update", "invoice"),
})
```

| Config Preset | Visible Columns |
|---|---|
| `DefaultLineItemConfig()` | Code, Description, Qty, UOM, Unit Price, Subtotal |
| `GRNLineItemConfig()` | Code, Description, Qty, UOM |
| `JournalLineItemConfig()` | Description, Account, Subtotal |

The `line_items` field name is fixed — backend expects this key.

---

### `TotalsSummaryBlock`

Source: `internal/web/dsl/blocks/totals.go`

Footer card showing subtotal / tax / total. Reads from `${totals}` in the AMIS data scope. Backend must populate `totals` as an array of `{label, amount}` objects.

```go
blocks.TotalsSummaryBlock(sess)
```

No config. Always reads `${totals}` from page data.

---

### `AddressBlock`

Source: `internal/web/dsl/blocks/address.go`

Billing and/or shipping address sections with country selector backed by `/api/v1/platform/countries/options`.

```go
blocks.AddressBlock(sess, blocks.AddressConfig{
    ShowBilling:  true,
    ShowShipping: true,
    ReadOnly:     !sess.Can("update", "invoice"),
})
```

Field names: `billing_street`, `billing_city`, `billing_state`, `billing_postal_code`, `billing_country` (and `shipping_*` variants).

---

### `ApprovalWorkflowBlock`

Source: `internal/web/dsl/blocks/approval.go`

Approval status + note section. Gated on feature flag `approval_workflow`. When flag is disabled, renders a collapsed placeholder section (not nil — always safe to include).

```go
blocks.ApprovalWorkflowBlock(sess)
// Visibility gated on feature flag internally.
// Approval action gated on AMIS expression: !${can_approve}
// Set can_approve in page data: "can_approve": sess.Can("approve", "invoice")
```

---

### `InternalNotesBlock`

Source: `internal/web/dsl/blocks/notes.go`

Collapsed rich-text internal notes section. Field name: `internal_notes`. Max 2000 chars. Collapsed by default — does not inflate the form for normal users.

```go
blocks.InternalNotesBlock(sess)
```

No config. Always adds `internal_notes` field.

---

## Dashboard / Analytics Blocks

### `StatCardBlock`

Source: `internal/web/dsl/blocks/stat_card.go`

Single KPI metric card: label, value, optional trend arrow.

```go
blocks.StatCardBlock(sess, blocks.StatCardConfig{
    Label:     "Total Revenue",
    ValueKey:  "total_revenue",     // key in AMIS page data scope
    Format:    "currency",          // "number"|"currency"|"percent"
    Currency:  "",                  // defaults to sess.Currency when Format=="currency"
    TrendKey:  "revenue_trend_pct", // optional — shows up/down arrow
    Trend:     blocks.TrendUp,      // TrendUp | TrendDown | TrendNeutral
    IconClass: "fa fa-coins",       // optional Font Awesome icon
})
```

| `TrendDirection` | Meaning |
|---|---|
| `TrendUp` | Green when positive, red when negative |
| `TrendDown` | Red when positive, green when negative |
| `TrendNeutral` (default) | No colour coding |

---

### `KPIRowBlock`

Source: `internal/web/dsl/blocks/kpi_row.go`

Horizontal row of `StatCardBlock` KPIs, each in an equal-width grid column.

```go
blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
    {Label: "Revenue",   ValueKey: "revenue",   Format: "currency"},
    {Label: "Invoices",  ValueKey: "inv_count",  Format: "number"},
    {Label: "Overdue",   ValueKey: "overdue_pct", Format: "percent", Trend: blocks.TrendDown},
})
```

Column width = `12 / len(cards)`. Max 6 cards (2 columns each). More than 6 wraps to 1 column each.

---

### `StatusBadgeColumn` / `StatusBadgeBlock`

Source: `internal/web/dsl/blocks/status_badge.go`

**`StatusBadgeColumn`** — use inside `DataTableConfig.Columns`:

```go
blocks.StatusBadgeColumn(blocks.StatusBadgeConfig{
    FieldName: "status",
    Label:     "Status",
    Mappings: []blocks.StatusMapping{
        {Value: "active",    Label: "Active",    Color: "success"},
        {Value: "suspended", Label: "Suspended", Color: "warning"},
        {Value: "archived",  Label: "Archived",  Color: "default"},
    },
})
// Returns ast.TableColumn{Type: "status", ...}
```

**`StatusBadgeBlock`** — standalone status display (outside a table):

```go
blocks.StatusBadgeBlock(sess, blocks.StatusBadgeConfig{...})
// Returns a disabled select node — read-only status display
```

`Color` values: `"success"` (green), `"warning"` (amber), `"danger"` (red), `"info"` (blue), `"default"` (grey).

---

### `EmptyStateBlock`

Source: `internal/web/dsl/blocks/empty_state.go`

"No results" prompt with optional call-to-action link.

```go
blocks.EmptyStateBlock(sess, blocks.EmptyStateConfig{
    Title:       "No invoices yet",
    Description: "Create your first invoice to get started.",
    ActionLabel: "Create Invoice",
    ActionURL:   "#invoices/new",
})
```

All fields optional — defaults to generic "No results" / "No records found matching your filters."

---

## Navigation / Actions Blocks

### `QuickActionsBlock`

Source: `internal/web/dsl/blocks/quick_actions.go`

Row of shortcut buttons. Actions without a `Permission` are always visible. Actions with `Permission` are excluded if the session lacks it — this is structural exclusion (not AMIS expression-based).

```go
blocks.QuickActionsBlock(sess, []blocks.QuickAction{
    {Label: "New Invoice", URL: "#invoices/new", Permission: "invoice.create", Icon: "fa fa-plus"},
    {Label: "Reports",     URL: "#reports",       Permission: "",               Icon: "fa fa-chart-bar"},
})
```

`Permission` format: `"resource.action"` — same format as `AllUIPermissions`. Empty string = always visible.

---

## List / Table Blocks

### `DataTableBlock`

Source: `internal/web/dsl/blocks/data_table.go`

Paginated, filterable CRUD table. Use this for listing pages instead of constructing `ast.CRUDNode` directly — it handles toolbar, bulk actions, and permission gating.

```go
blocks.DataTableBlock(sess, blocks.DataTableConfig{
    APIURL:  "/api/v1/finance/invoices",
    Title:   "Invoices",
    PageSize: 20,
    PrimaryKey: "id",
    Columns: []blocks.ColumnDef{
        {Name: "number",   Label: "Invoice #",  Sortable: true},
        {Name: "supplier", Label: "Supplier"},
        {Name: "amount",   Label: "Amount",     Type: "currency"},
        {Name: "due_date", Label: "Due",         Type: "date", Sortable: true},
    },
    // Optional status badge column — add to Columns as:
    // blocks.StatusBadgeColumn(statusCfg) → cast to blocks.ColumnDef (note: use DataTableBlock's Columns []ColumnDef OR compose with ast.CRUDNode directly)
    AllowCreate: sess.Can("create", "invoice"),
    CreateURL:   "#invoices/new",
    AllowExport: sess.Can("export", "invoice"),
    BulkActions: []blocks.BulkActionDef{
        {
            Label:      "Bulk Archive",
            Permission: "invoice.delete",
            APIURL:     "/api/v1/finance/invoices/bulk-archive",
            APIMethod:  "post",
            Level:      "warning",
            Confirm:    "Archive selected invoices?",
        },
    },
    RowActions: []ast.ActionNode{
        {Label: "View", ActionType: "link", Target: "#invoices/${id}"},
    },
})
```

| Column `Type` | Renders as |
|---|---|
| `"text"` (default) | Plain string |
| `"date"` | Formatted date |
| `"number"` | Numeric with locale formatting |
| `"currency"` | Currency with symbol |
| `"status"` | Status badge (use `StatusBadgeColumn` helper) |
| `"link"` | Hyperlink |
| `"image"` | Thumbnail image |

`BulkActionDef.Permission` is checked structurally in Go — users without the permission never see the button in the schema.

---

## Block Helper Functions

These are package-level helpers in `quick_actions.go`, available to all blocks within the `blocks` package:

| Function | Purpose |
|---|---|
| `canPerm(sess, "resource.action")` | Checks raw dot-notation permission |
| `boolExpr(bool) string` | Returns `"true"` or `""` for AMIS `disabledOn` expressions |
| `resourceFromURL(url) string` | Extracts `"module.resource"` from an API URL |

These are unexported — internal use within the `blocks` package only.

---

## Composing Blocks into a Page

Typical document form structure:

```go
func Schema(sess ui.UISessionContext) any {
    readOnly := !sess.Can("update", "invoice")
    return ast.PageNode{
        Title: "Invoice #${ref_number}",
        InitAPI: ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/${id}"},
        Data: ui.M{
            "can_approve": sess.Can("approve", "invoice"),
        },
        Body: ast.GridNode{
            Columns: []ast.GridColumn{
                {MD: 8, Body: []ast.Node{
                    blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
                        ShowCurrency: true,
                        ShowStatus:   true,
                        StatusOptions: invoiceStatuses(),
                        ReadOnly: readOnly,
                    }),
                    blocks.ProductServiceLineBlock(sess, blocks.DefaultLineItemConfig()),
                    blocks.InternalNotesBlock(sess),
                }},
                {MD: 4, Body: []ast.Node{
                    blocks.TotalsSummaryBlock(sess),
                    blocks.AddressBlock(sess, blocks.AddressConfig{ShowBilling: true}),
                    blocks.ApprovalWorkflowBlock(sess),
                }},
            },
        },
    }
}
```

---

## Adding a New Block

1. Create `internal/web/dsl/blocks/<name>.go`
2. Define a `*Config` struct (exported, zero value must be safe)
3. Define `*Block(sess ui.UISessionContext, cfg *Config) ast.Node`
4. Add the block to this doc under the relevant section
5. Write a unit test verifying the returned node compiles without error via `ast.CompileTree(block(sess, cfg))`

**Do not** add side effects, I/O, or IAM calls inside a block. Blocks are pure functions of their inputs. Permission checks must be done on `sess` before building the node, not by calling out to any service.
