# DSL Blocks Reference

> Last verified: 2026-05-19 | Code pointer: `internal/web/dsl/blocks/`, `internal/web/ast/`

> **Complete listing.** All 27 blocks documented. Add new blocks here when you add new files.

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

### `PartyBlock`

Source: `internal/web/dsl/blocks/party.go`

Party (customer / supplier / employee) selector section. Three presets provided — use these rather than passing a raw `PartyConfig`.

```go
// Customer
blocks.PartyBlock(sess, blocks.DefaultCustomerConfig())

// Supplier
blocks.PartyBlock(sess, blocks.DefaultSupplierConfig())

// Employee
blocks.PartyBlock(sess, blocks.DefaultEmployeeConfig())

// Custom
blocks.PartyBlock(sess, blocks.PartyConfig{
    Label:      "Vendor",
    FieldName:  "vendor_id",
    OptionsURL: "/api/v1/procurement/vendors/options",
    Required:   true,
    ReadOnly:   !sess.Can("update", "purchase_order"),
})
```

| Preset | `FieldName` | `OptionsURL` |
|---|---|---|
| `DefaultCustomerConfig()` | `customer_id` | `/api/v1/crm/customers/options` |
| `DefaultSupplierConfig()` | `supplier_id` | `/api/v1/procurement/suppliers/options` |
| `DefaultEmployeeConfig()` | `employee_id` | `/api/v1/hr/employees/options` |

---

### `AttachmentsBlock`

Source: `internal/web/dsl/blocks/attachments.go`

Collapsible file attachment section. Collapsed by default. Files are uploaded via multipart to the document's attachments endpoint. Field name: `attachments`.

```go
blocks.AttachmentsBlock(sess)
```

No config. Always renders collapsed.

---

### `PaymentTermsBlock`

Source: `internal/web/dsl/blocks/payment_terms.go`

Payment terms selector + due date inputs. Terms options sourced from `/api/v1/finance/payment-terms/options`.

```go
blocks.PaymentTermsBlock(sess, blocks.PaymentTermsConfig{
    ReadOnly: !sess.Can("update", "invoice"),
})
```

Field names: `payment_terms_id`, `payment_due_date`.

---

### `TaxSummaryBlock`

Source: `internal/web/dsl/blocks/tax_summary.go`

Read-only tax breakdown card. Sourced from `${tax_lines}` in the AMIS data scope — no extra API call. Backend must populate `tax_lines` as an array of `{tax_name, taxable_amount, tax_amount, rate_pct}`.

```go
blocks.TaxSummaryBlock(sess)
```

No config. Always reads `${tax_lines}` from page data.

---

### `ActivityFeedBlock`

Source: `internal/web/dsl/blocks/activity_feed.go`

Chronological audit trail for a specific document / resource. Calls `/api/v1/audit/{resource}/activity` or `/api/v1/audit/activity` if resource is empty.

```go
blocks.ActivityFeedBlock(sess, blocks.ActivityFeedConfig{
    Title:        "Invoice History",
    Resource:     "invoices",   // → /api/v1/audit/invoices/activity
    ResourceID:   "${id}",      // available in AMIS scope
    ShowComments: true,
    Limit:        20,
})
```

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

### `ActivityPanelBlock`

Source: `internal/web/dsl/blocks/activity_panel.go`

Recent activity card for dashboard widgets. Calls `/api/v1/{resource}/recent` or `/api/v1/audit/recent` if resource is empty.

```go
blocks.ActivityPanelBlock(sess, blocks.ActivityPanelConfig{
    Title:    "Recent Transactions",
    Resource: "transactions",   // → /api/v1/transactions/recent
    Limit:    10,
})
```

---

### `ChartPanelBlock`

Source: `internal/web/dsl/blocks/chart_panel.go`

Card-wrapped ECharts chart with optional period picker. Use `ChartType` constants: `ChartTypeBar`, `ChartTypeLine`, `ChartTypePie`.

```go
blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
    Title:        "Revenue by Month",
    ChartType:    blocks.ChartTypeLine,
    APIURL:       "/api/v1/finance/reports/revenue-trend",
    PeriodPicker: true,    // adds Month/Quarter/Year selector above chart
    Height:       300,     // default 300
})
```

When `PeriodPicker: true`, adds `chart_period` select with options: `"month"` (default), `"quarter"`, `"year"`.

---

### `DetailCardBlock`

Source: `internal/web/dsl/blocks/detail_card.go`

Read-only label:value field group card. Backed by `${record}` in AMIS data scope.

```go
blocks.DetailCardBlock(sess, blocks.DetailCardConfig{
    Title: "Invoice Details",
    Fields: []blocks.FieldDef{
        {Label: "Customer",   Key: "customer_name"},
        {Label: "Issued",     Key: "document_date", Format: "date"},
        {Label: "Total",      Key: "total_amount",  Format: "currency"},
        {Label: "Tax",        Key: "tax_amount",    Format: "currency"},
    },
})
```

`Format` values: `""` / `"text"` (default), `"date"`, `"currency"`, `"percent"`, `"number"`.

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

### `EntityBreadcrumbBlock`

Source: `internal/web/dsl/blocks/entity_breadcrumb.go`

Hierarchy path for an entity. Reads `${breadcrumbs}` from AMIS data scope as an array of `{label, url}`. Backend must populate `breadcrumbs` in the page init API response.

```go
blocks.EntityBreadcrumbBlock(sess)
```

No config.

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

### `FilterBarBlock`

Source: `internal/web/dsl/blocks/filter_bar.go`

Filter header for every listing page. **Never write a custom filter form in a screen file** — always use this block.

```go
blocks.FilterBarBlock(sess, blocks.FilterBarConfig{
    ShowSearch:        true,
    SearchPlaceholder: "Search invoices…",
    ShowDateRange:     true,
    ShowStatus:        true,
    StatusOptions: []ast.SelectOption{
        {Label: "Draft",  Value: "draft"},
        {Label: "Sent",   Value: "sent"},
        {Label: "Paid",   Value: "paid"},
        {Label: "Overdue", Value: "overdue"},
    },
    ShowEntityPicker:  true,
    EntityURL:         "/api/v1/crm/customers/options",
    EntityFieldName:   "customer_id",
    EntityLabel:       "Customer",
    ShowCurrency:      false,
    ShowAmountRange:   true,
    Collapsible:       true,
})
```

| Config Field | Type | Default | Purpose |
|---|---|---|---|
| `ShowSearch` | `bool` | `false` | Keyword search input (`keywords` field) |
| `SearchPlaceholder` | `string` | `"Search…"` | Input placeholder |
| `ShowDateRange` | `bool` | `false` | Date range picker (`date_range` field) |
| `ShowStatus` | `bool` | `false` | Multi-select status filter |
| `StatusOptions` | `[]ast.SelectOption` | — | Required when `ShowStatus: true` |
| `ShowEntityPicker` | `bool` | `false` | Searchable select for related entity |
| `EntityURL` | `string` | — | Options API URL |
| `EntityFieldName` | `string` | `"entity_id"` | Filter field name |
| `ShowCurrency` | `bool` | `false` | Currency select (options from platform API) |
| `ShowAmountRange` | `bool` | `false` | `amount_min` + `amount_max` number inputs |
| `ShowTypeFilter` | `bool` | `false` | Type select (configure `TypeOptions`, `TypeFieldName`) |
| `Collapsible` | `bool` | `false` | Wraps in collapsible section |

When all flags are false, falls back to a single keyword search input.

---

### `BulkActionsBlock`

Source: `internal/web/dsl/blocks/bulk_actions.go`

Permission-filtered bulk action buttons. Returns `[]ast.Node` (not a single node). Pass directly to `DataTableConfig.BulkActions` or a CRUD toolbar.

```go
actions := blocks.BulkActionsBlock(sess, []blocks.BulkActionDef{
    {
        Label:      "Approve Selected",
        Permission: "invoice.approve",
        APIURL:     "/api/v1/finance/invoices/bulk-approve",
        APIMethod:  "post",
        Level:      "primary",
        Confirm:    "Approve all selected invoices?",
    },
    {
        Label:      "Void Selected",
        Permission: "invoice.delete",
        APIURL:     "/api/v1/finance/invoices/bulk-void",
        APIMethod:  "post",
        Level:      "danger",
        Confirm:    "Void selected invoices? This cannot be undone.",
    },
})
```

Actions with `Permission` not held by the session are **structurally excluded** — never sent to the browser. Empty `Permission` = always included.

Returns empty slice (never nil) if no actions pass the permission check.

---

## Report Blocks

Use these blocks on financial report pages (P&L, balance sheet, ledger reports). Never use `DataTableBlock` on a report page — use `ReportTableBlock`.

### `ReportHeaderBlock`

Source: `internal/web/dsl/blocks/report_header.go`

Report title and metadata card.

```go
blocks.ReportHeaderBlock(sess, blocks.ReportHeaderConfig{
    Title:       "Profit & Loss Statement",
    Description: "Net income for the selected period",
})
```

---

### `ReportFilterBlock`

Source: `internal/web/dsl/blocks/report_filter.go`

Filter form for report pages. Always include `ShowPeriod: true` — period is required for all financial reports.

```go
blocks.ReportFilterBlock(sess, blocks.ReportFilterConfig{
    ShowPeriod:       true,
    ShowEntityPicker: true,
    EntityURL:        "/api/v1/crm/customers/options",
    ShowCurrency:     true,
    ShowComparison:   true,   // adds "Compare With" prior period / prior year
})
```

`ShowComparison` adds `compare_period` select: `"prior"`, `"prior_year"`, `"none"` (default).

---

### `ReportTableBlock`

Source: `internal/web/dsl/blocks/report_table.go`

Data table for financial reports. Supports grouping, subtotals, grand total, and CSV/PDF export.

```go
blocks.ReportTableBlock(sess, blocks.ReportTableConfig{
    Columns: []blocks.ReportColumnDef{
        {Name: "account_name", Label: "Account",  Type: "text",     Sortable: true},
        {Name: "debit",        Label: "Debit",    Type: "currency", Sortable: false},
        {Name: "credit",       Label: "Credit",   Type: "currency", Sortable: false},
        {Name: "balance",      Label: "Balance",  Type: "currency", Sortable: true},
    },
    GroupBy:        []string{"account_type"},
    ShowSubtotals:  true,
    ShowGrandTotal: true,
    Exportable:     sess.Can("export", "report"),
})
```

Export buttons use `${api_url}/export?format=csv` and `${api_url}/export?format=pdf` — `api_url` must be in the page `Data`.

`Column.Type` values: `"text"`, `"number"`, `"currency"`.

---

### `ReportChartBlock`

Source: `internal/web/dsl/blocks/report_chart.go`

Bare ECharts node for embedding inside a report page. Lighter than `ChartPanelBlock` — no card wrapper, no period picker.

```go
blocks.ReportChartBlock(sess, blocks.ReportChartConfig{
    Type:   blocks.ChartTypeLine,
    Height: 250,
})
```

`ChartType` constants: `ChartTypeBar`, `ChartTypeLine`, `ChartTypePie`.

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
