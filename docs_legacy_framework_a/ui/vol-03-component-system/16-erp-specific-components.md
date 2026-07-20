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
chapter: 16
title: "ERP-Specific Components"
audience: "Backend engineers building any document or listing screen"
prerequisites:
  - "Vol 01 Ch 4 — AST node types"
  - "Ch 12 — Screen composition"
  - "Ch 15 — Workflow and Approval Components"
section: "vol-03-component-system"
related:
  - "[Ch 13 — Dashboard Framework](13-dashboard-framework.md)"
  - "[Ch 15 — Workflow and Approval Components](15-workflow-and-approval-components.md)"
---

# Chapter 16 — ERP-Specific Components

## Table of Contents

1. [Why ERP needs domain components](#161-why-erp-needs-domain-components)
2. [Line item grid](#162-line-item-grid)
3. [Status badge](#163-status-badge)
4. [Detail card](#164-detail-card)
5. [Document header block](#165-document-header-block)
6. [Party block](#166-party-block)
7. [Totals summary block](#167-totals-summary-block)
8. [Block matrix — which blocks appear in which document types](#168-block-matrix)
9. [Zero map rule — typed AST nodes only](#169-zero-map-rule)

---

## 16.1 Why ERP needs domain components

Generic UI platforms provide form inputs, tables, and charts. ERP software layers highly specific domain semantics on top of those primitives:

- A line item grid is not just a repeating form — it must compute subtotals using a formula, handle unit-of-measure selectors sourced from inventory, and branch its columns based on document type (commercial vs. journal).
- A status badge is not a colour indicator — it must map business workflow states (draft, approved, posted, reversed) to visual feedback consistently across every document type.
- A detail card is not a label-value table — it must apply currency and date filters correctly so financial amounts always render with the right precision.

Domain components encode these rules once and expose them via typed configuration structs. Without them, every screen would reinvent the same patterns and diverge over time.

---

## 16.2 Line item grid

### ProductServiceLineBlock

`ProductServiceLineBlock` is the shared line items component for all document forms that contain rows of goods or services. It is defined in `internal/web/dsl/blocks/line_items.go`.

```go
func ProductServiceLineBlock(sess ui.UISessionContext, cfg LineItemConfig) ast.Node
```

The block emits an `ast.ComboNode` — AMIS's repeating-row control. Each row renders the columns specified by the `LineItemConfig`.

### LineItemConfig

```go
type LineItemConfig struct {
    ShowProductCode   bool
    ShowDescription   bool
    ShowQty           bool
    ShowUnitOfMeasure bool
    ShowUnitPrice     bool
    ShowDiscount      bool
    ShowTaxRate       bool
    // ShowSubtotal: computed subtotal column (qty × unit_price × discount).
    // Use for commercial documents.
    ShowSubtotal      bool
    // ShowDebit / ShowCredit: explicit amount inputs for journal entry lines.
    // Do not combine with ShowSubtotal.
    ShowDebit         bool
    ShowCredit        bool
    ShowAccount       bool
    AllowFreeTextItem bool
    DefaultCurrency   string
    MaxLines          int
    ReadOnly          bool
}
```

### Pre-built configs

Three factory functions cover the main document types:

```go
// Standard commercial documents: invoice, PO, bill, SO
lineCfg := blocks.DefaultLineItemConfig()
// Sets: ShowProductCode, ShowDescription, ShowQty, ShowUnitOfMeasure,
//       ShowUnitPrice, ShowSubtotal

// Goods receipt note: quantity received, no pricing
lineCfg := blocks.GRNLineItemConfig()
// Sets: ShowProductCode, ShowDescription, ShowQty, ShowUnitOfMeasure

// Journal entry: double-entry debit/credit, no product columns
lineCfg := blocks.JournalLineItemConfig()
// Sets: ShowDescription, ShowAccount, ShowDebit, ShowCredit
```

Start from the factory and override individual fields:

```go
lineCfg := blocks.DefaultLineItemConfig()
lineCfg.ShowDiscount = true      // sales invoices only
lineCfg.ShowTaxRate  = true      // when tax module is enabled
lineCfg.ReadOnly     = cfg.ReadOnly
```

### FormulaNode for subtotal computation

When `ShowSubtotal` is true, the block adds a `FormulaNode` and a read-only `InputNumberNode` sharing the same field name `"subtotal"`:

```go
ast.FormulaNode{
    Name:      "subtotal",
    Formula:   "qty * unit_price * (1 - (discount_pct || 0) / 100)",
    InitSet:   true,
    Condition: "${qty && unit_price}",
}
ast.InputNumberNode{Name: "subtotal", Label: "Subtotal", Precision: 2, DisabledOn: "true"}
```

**Formula details:**
- `(discount_pct || 0)` guards against undefined — if the discount column is not visible or not entered, it defaults to 0.
- `Condition: "${qty && unit_price}"` prevents computing before both required fields are filled.
- `InitSet: true` computes the value on page load, not just when the user edits fields.

The read-only `InputNumberNode` displays the computed value. The `FormulaNode` itself is invisible — it only writes to the scope variable.

### AMIS output shape

The combo node serialises to:

```json
{
  "type": "combo",
  "name": "line_items",
  "label": "Line Items",
  "multiple": true,
  "addButtonLabel": "Add Line",
  "items": [
    {"type": "input-text",   "name": "product_code",  "label": "Code"},
    {"type": "input-text",   "name": "description",   "label": "Description", "required": true},
    {"type": "input-number", "name": "qty",            "label": "Qty", "required": true},
    {"type": "select",       "name": "uom",            "label": "UOM", "source": "..."},
    {"type": "input-number", "name": "unit_price",     "label": "Unit Price", "required": true, "precision": 2},
    {"type": "formula",      "name": "subtotal",       "formula": "qty * unit_price * (1 - (discount_pct || 0) / 100)", "initSet": true, "condition": "${qty && unit_price}"},
    {"type": "input-number", "name": "subtotal",       "label": "Subtotal", "precision": 2, "disabledOn": "true"}
  ]
}
```

### Rule: never define a custom line item table

Every document form with line items **must** use `ProductServiceLineBlock`. Defining a custom combo or table in a screen file is prohibited. Customise via `LineItemConfig` fields; if a required configuration is not expressible via those fields, extend `LineItemConfig` and update the block.

---

## 16.3 Status badge

### StatusBadgeBlock

`StatusBadgeBlock` renders a standalone colour-coded status indicator for use in document detail views and cards. It is defined in `internal/web/dsl/blocks/status_badge.go`.

```go
func StatusBadgeBlock(_ ui.UISessionContext, cfg StatusBadgeConfig) ast.Node
```

**Important:** `StatusBadgeBlock` emits an `ast.MappingNode` with AMIS `type: "mapping"`. It does **not** use AMIS `type: "status"`. The `"status"` type has a fixed set of colours that cannot be customised. The `"mapping"` type renders arbitrary HTML, giving full control over badge appearance.

```go
type StatusBadgeConfig struct {
    FieldName string
    Label     string
    Mappings  []StatusMapping
}

type StatusMapping struct {
    Value string
    Label string
    Color string // "success" | "warning" | "danger" | "info" | "default"
}
```

### Usage

```go
blocks.StatusBadgeBlock(sess, blocks.StatusBadgeConfig{
    FieldName: "status",
    Label:     "Status",
    Mappings: []blocks.StatusMapping{
        {Value: "draft",     Label: "Draft",     Color: "default"},
        {Value: "sent",      Label: "Sent",      Color: "info"},
        {Value: "paid",      Label: "Paid",      Color: "success"},
        {Value: "overdue",   Label: "Overdue",   Color: "danger"},
        {Value: "cancelled", Label: "Cancelled", Color: "warning"},
    },
})
```

### Color mapping

| Color value | CSS class | Visual meaning |
|---|---|---|
| `"success"` | `badge-success` | Positive / complete |
| `"warning"` | `badge-warning` | Caution / at risk |
| `"danger"` | `badge-danger` | Error / overdue |
| `"info"` | `badge-info` | Neutral / in progress |
| `"default"` | `badge-default` | Unstarted / draft |

A `"*"` fallback mapping is always added automatically and renders the raw field value in a default badge when no explicit mapping matches.

### StatusBadgeColumn for tables

For listing pages, use `StatusBadgeColumn` (or `StatusBadgeColumnDef`) instead of `StatusBadgeBlock`. These return a table column configured as `type: "mapping"`:

```go
// For use inside DataTableConfig.Columns (returns ColumnDef)
blocks.StatusBadgeColumnDef(blocks.StatusBadgeConfig{
    FieldName: "status",
    Label:     "Status",
    Mappings:  invoiceStatusMappings,
})

// For use in ast.TableColumn slice directly (returns ast.TableColumn)
blocks.StatusBadgeColumn(blocks.StatusBadgeConfig{
    FieldName: "status",
    Label:     "Status",
    Mappings:  invoiceStatusMappings,
})
```

### AMIS output shape

```json
{
  "type": "mapping",
  "name": "status",
  "label": "Status",
  "map": {
    "draft":     "<span class=\"badge badge-default\">Draft</span>",
    "sent":      "<span class=\"badge badge-info\">Sent</span>",
    "paid":      "<span class=\"badge badge-success\">Paid</span>",
    "overdue":   "<span class=\"badge badge-danger\">Overdue</span>",
    "cancelled": "<span class=\"badge badge-warning\">Cancelled</span>",
    "*":         "<span class=\"badge badge-default\">${value}</span>"
  }
}
```

---

## 16.4 Detail card

### DetailCardBlock

`DetailCardBlock` renders a read-only field group — a description list of label-value pairs. It is defined in `internal/web/dsl/blocks/detail_card.go`.

```go
func DetailCardBlock(_ ui.UISessionContext, cfg DetailCardConfig) ast.Node
```

**Important:** `DetailCardBlock` emits a `CardNode` containing a `PropertyNode` with AMIS `type: "property"`. It does **not** use a single-row table. The `"property"` type is a semantic description list, not a grid — it respects the `Column` count for layout but remains accessible and correctly styled in both light and dark themes.

```go
type DetailCardConfig struct {
    Title  string
    Fields []FieldDef
    Column int // label-value pairs per row (default: 3)
}

type FieldDef struct {
    Label  string
    Key    string
    Format string // "" | "currency" | "date" | "percent"
}
```

### Format filters

| Format | AMIS expression | Output |
|---|---|---|
| `""` (default) | `"${key}"` | Raw string value |
| `"currency"` | `"${key\|number}"` | Thousand separators, 2 decimal places |
| `"date"` | `"${key\|date:YYYY-MM-DD}"` | ISO date format |
| `"percent"` | `"${key\|percent}"` | Percentage with `%` suffix |

### Usage

```go
blocks.DetailCardBlock(sess, blocks.DetailCardConfig{
    Title:  "Invoice Summary",
    Column: 3,
    Fields: []blocks.FieldDef{
        {Label: "Invoice Number", Key: "reference"},
        {Label: "Issue Date",     Key: "issue_date",   Format: "date"},
        {Label: "Due Date",       Key: "due_date",     Format: "date"},
        {Label: "Currency",       Key: "currency"},
        {Label: "Subtotal",       Key: "subtotal",     Format: "currency"},
        {Label: "Tax",            Key: "tax_total",    Format: "currency"},
        {Label: "Total",          Key: "grand_total",  Format: "currency"},
        {Label: "Discount",       Key: "discount_pct", Format: "percent"},
    },
})
```

### AMIS output shape

```json
{
  "type": "card",
  "header": {"title": "Invoice Summary"},
  "body": [
    {
      "type": "property",
      "column": 3,
      "items": [
        {"label": "Invoice Number", "content": "${reference}"},
        {"label": "Issue Date",     "content": "${issue_date|date:YYYY-MM-DD}"},
        {"label": "Due Date",       "content": "${due_date|date:YYYY-MM-DD}"},
        {"label": "Currency",       "content": "${currency}"},
        {"label": "Subtotal",       "content": "${subtotal|number}"},
        {"label": "Tax",            "content": "${tax_total|number}"},
        {"label": "Total",          "content": "${grand_total|number}"},
        {"label": "Discount",       "content": "${discount_pct|percent}"}
      ]
    }
  ]
}
```

---

## 16.5 Document header block

`DocumentHeaderBlock` renders the top section of a document form containing the reference number, date, currency, and status fields. It is defined in `internal/web/dsl/blocks/`.

```go
type DocumentHeaderConfig struct {
    ShowCurrency  bool
    ShowStatus    bool
    StatusOptions []ast.SelectOption
    ReadOnly      bool
}
```

```go
blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
    ShowCurrency: true,
    ShowStatus:   true,
    ReadOnly:     cfg.ReadOnly,
    StatusOptions: []ast.SelectOption{
        {Label: "Draft",     Value: "draft"},
        {Label: "Sent",      Value: "sent"},
        {Label: "Paid",      Value: "paid"},
        {Label: "Overdue",   Value: "overdue"},
        {Label: "Cancelled", Value: "cancelled"},
    },
})
```

The block always renders: reference number (auto-generated display), document date, and a currency selector sourced from `/api/v1/platform/currencies/options`. `ShowStatus` adds the status dropdown. `ReadOnly` disables all fields.

---

## 16.6 Party block

`PartyBlock` renders the customer or supplier selection section. The block is configurable for all counterparty types.

```go
// For sales documents (invoices, SO, credit notes)
partyCfg := blocks.DefaultCustomerConfig()

// For purchase documents (bills, PO, GRN, debit notes)
partyCfg := blocks.DefaultSupplierConfig()

blocks.PartyBlock(sess, partyCfg)
```

`PartyBlock` renders a searchable select for the party entity, plus fields for billing contact and (for sales documents) shipping contact. The party options are sourced from the appropriate API endpoint based on the config type.

---

## 16.7 Totals summary block

`TotalsSummaryBlock` renders the read-only totals footer of a document: subtotal, discount, tax, and grand total. It reads from page scope variables populated by the form's line items and tax calculation API.

```go
blocks.TotalsSummaryBlock(sess)
```

No configuration is needed. The block reads `subtotal`, `discount_amount`, `tax_total`, and `grand_total` from page scope. These must be returned by the `InitAPI` endpoint or computed server-side after line item changes.

`TaxSummaryBlock` renders a breakdown of tax by tax rate, for use between the line items and the totals:

```go
blocks.TaxSummaryBlock(sess)
```

The canonical ordering in a document form is:

```
ProductServiceLineBlock  ← line items with per-row formula
TaxSummaryBlock          ← tax breakdown by rate
TotalsSummaryBlock       ← grand total row
```

---

## 16.8 Block matrix

The following table maps document types to which blocks they include. Ticked cells indicate the block is present in the screen's `Body` slice.

| Block | Invoice | Bill (PO) | Credit Note | Expense | GRN | Journal Entry |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| `DocumentHeaderBlock` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `PartyBlock` (customer) | ✓ | | ✓ | | | |
| `PartyBlock` (supplier) | | ✓ | | ✓ | ✓ | |
| `AddressBlock` | ✓ | ✓ | ✓ | | | |
| `ProductServiceLineBlock` (default) | ✓ | ✓ | ✓ | ✓ | | |
| `ProductServiceLineBlock` (GRN) | | | | | ✓ | |
| `ProductServiceLineBlock` (journal) | | | | | | ✓ |
| `TaxSummaryBlock` | ✓ | ✓ | ✓ | ✓ | | |
| `TotalsSummaryBlock` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `ApprovalWorkflowBlock` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `PaymentTermsBlock` | opt | opt | | | | |
| `AttachmentsBlock` | opt | opt | opt | opt | opt | opt |
| `InternalNotesBlock` | opt | opt | opt | opt | | opt |

**Key:**
- ✓ = always present
- opt = optional, controlled by screen config
- blank = not applicable

The `ShowDiscount` flag on `DefaultLineItemConfig()` is set only for sales-side documents (Invoice, Credit Note). Purchase documents and expense claims do not show a discount column.

---

## 16.9 Zero map rule — typed AST nodes only

The DSL block layer must use typed `ast.*` nodes exclusively. Raw `map[string]any` values are prohibited in block and screen files.

### Why the rule exists

1. **Compile-time safety.** Struct field names are checked by the Go compiler. A mistyped `map[string]any` key silently produces a missing field in the schema.
2. **ValidateStage.** The pipeline's validate pass works on typed nodes. It can check invariants like the chart transparency requirement, required `Dialog` fields on dialog action nodes, and `MappingNode` fallback presence. Raw maps bypass these checks.
3. **Consistency.** The `blocks.*` functions exist specifically so each domain concept is expressed the same way everywhere. A raw map in a screen file is a local redefinition that will diverge from the block implementation over time.

### The rule in practice

```go
// Correct — typed node
ast.MappingNode{
    Name:  "status",
    Label: "Status",
    Map: map[string]string{
        "draft":   `<span class="badge badge-default">Draft</span>`,
        "paid":    `<span class="badge badge-success">Paid</span>`,
        "*":       `<span class="badge badge-default">${value}</span>`,
    },
}

// Wrong — raw map in a screen file
map[string]any{
    "type":  "mapping",
    "name":  "status",
    "label": "Status",
    "map": map[string]string{
        "draft": "Draft",
        "paid":  "Paid",
    },
}
```

The wrong form also has a subtle bug: the `"*"` fallback is missing. `StatusBadgeBlock` adds it automatically; the raw map does not.

### When Config maps are acceptable

The only legitimate `map[string]any` in the codebase is the `Config` field of `ast.ChartNode`. ECharts options are an open schema — there is no typed representation for the full ECharts API. The chart `Config` map is the single sanctioned exception.

```go
// Acceptable — ChartNode.Config is a documented exception
ast.ChartNode{
    Config: map[string]any{
        "backgroundColor": "transparent",
        "series": []map[string]any{...},
    },
}
```

All other use of `map[string]any` in block or screen files should be treated as a bug.

---

### ERP component checklist

- [ ] Line items: use `ProductServiceLineBlock` — never define a custom combo or table
- [ ] Line item config: start from `DefaultLineItemConfig()`, `GRNLineItemConfig()`, or `JournalLineItemConfig()` and override only what differs
- [ ] `ShowSubtotal` and `ShowDebit`/`ShowCredit` are mutually exclusive — do not set both
- [ ] Status badge: use `StatusBadgeBlock` for detail views, `StatusBadgeColumnDef` for listing tables
- [ ] Status badge: AMIS node type is `"mapping"`, not `"status"` — the block handles this automatically
- [ ] Detail card: use `DetailCardBlock` with `Format` for currency/date/percent fields — not raw `${var}` strings
- [ ] Detail card: AMIS node type is `"property"` inside a `"card"` — not a table
- [ ] Document header: always first block in a document `Body`
- [ ] Approval: always last substantive block before optional attachments/notes
- [ ] No raw `map[string]any` in block or screen files (except `ChartNode.Config`)
- [ ] `"*"` fallback is present in every mapping node — `StatusBadgeBlock` adds it automatically; manual maps must add it explicitly
