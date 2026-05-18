# UI DSL — Building-Block Architecture

The `internal/web/dsl` package provides a three-layer composition model for ERP
screens. Every screen is composed entirely of reusable blocks. No screen defines
its own tables, forms, or charts.

---

## Three-Layer Model

```
builders/   Domain-specific node constructors (account picker, employee picker, …)
            ↓  used by
blocks/     Reusable UI concerns (line items, filter bar, stat card, …)
            ↓  composed by
screens/    Thin screen layouts — config in, ast.Node out, ≤ 60 lines
```

### Hard Constraints

- **Zero `map[string]any`** in `dsl/` (except `ChartNode.Config` — the sole ECharts escape hatch)
- **No IAM imports** in `dsl/` — interact with identity only through `ui.UISessionContext`
- **Blocks own permission checks** — callers must never gate before calling a block
- **Screens ≤ 60 lines** — a larger file means logic escaped the block layer
- All block-construction errors use `*sharedErrors.BusinessError` with `DSL_*` prefix

---

## Block Directory

```
internal/web/dsl/blocks/
│
├── Document blocks
│   ├── line_items.go        ProductServiceLineBlock  — shared by all documents with line items
│   ├── document_header.go   DocumentHeaderBlock      — ref, date, currency, status
│   ├── party.go             PartyBlock               — customer / supplier / employee
│   ├── address.go           AddressBlock             — billing / shipping
│   ├── tax_summary.go       TaxSummaryBlock
│   ├── totals.go            TotalsSummaryBlock
│   ├── payment_terms.go     PaymentTermsBlock
│   ├── attachments.go       AttachmentsBlock
│   ├── approval.go          ApprovalWorkflowBlock    — gated on approval_workflow flag
│   └── notes.go             InternalNotesBlock
│
├── Listing blocks
│   ├── filter_bar.go        FilterBarBlock
│   ├── data_table.go        DataTableBlock           — wraps CRUDNode
│   ├── bulk_actions.go      BulkActionsBlock
│   └── empty_state.go       EmptyStateBlock
│
├── Data display blocks
│   ├── stat_card.go         StatCardBlock
│   ├── detail_card.go       DetailCardBlock
│   ├── activity_feed.go     ActivityFeedBlock
│   ├── status_badge.go      StatusBadgeBlock / StatusBadgeColumn
│   └── entity_breadcrumb.go EntityBreadcrumbBlock
│
├── Report blocks
│   ├── report_header.go     ReportHeaderBlock
│   ├── report_filter.go     ReportFilterBlock
│   ├── report_table.go      ReportTableBlock         — subtotals, grouping, grand total
│   └── report_chart.go      ReportChartBlock
│
└── Dashboard blocks
    ├── kpi_row.go           KPIRowBlock              — row of StatCards in a GridNode
    ├── chart_panel.go       ChartPanelBlock
    ├── activity_panel.go    ActivityPanelBlock
    └── quick_actions.go     QuickActionsBlock        — permission-filtered shortcuts
```

---

## Five-Family Pattern

### Family 1 — Document Forms

Every document form shares the same set of blocks. Differences are
**configuration**, not code.

| Block | Invoice | PO | Bill | Credit Note | Expense | GRN | Journal |
|-------|:-------:|:--:|:----:|:-----------:|:-------:|:---:|:-------:|
| DocumentHeader | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Party | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | — |
| Address | ✓ | ✓ | ✓ | ✓ | — | ✓ | — |
| **ProductServiceLine** | **✓** | **✓** | **✓** | **✓** | **✓** | **✓** | — |
| TaxSummary | ✓ | ✓ | ✓ | ✓ | ✓ | — | — |
| TotalsSummary | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| PaymentTerms | ✓ | ✓ | ✓ | — | — | — | — |
| Attachments | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| ApprovalWorkflow | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| InternalNotes | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

`ProductServiceLineBlock` is the most critical block — never define a custom
line item table in a screen file.

### Family 2 — Listing Pages

Every listing page: `FilterBarBlock` + `DataTableBlock` + optional `BulkActionsBlock`.
`GenericListScreen` covers ~80% of listing pages.

```
┌─────────────────────────────────────────────────────────┐
│  Page title                             [+ New]  [⋮]   │
├─────────────────────────────────────────────────────────┤
│  FilterBarBlock                                         │
│  date range │ status │ entity │ search  │ [Filter]      │
├─────────────────────────────────────────────────────────┤
│  BulkActionsBlock  (visible only when rows selected)    │
├─────────────────────────────────────────────────────────┤
│  DataTableBlock                                         │
│  col │ col │ col │ col │ status │ actions               │
│  ...                              pagination            │
└─────────────────────────────────────────────────────────┘
```

### Family 3 — Data Display

`StatCardBlock` and `DetailCardBlock` appear in dashboards, document headers,
and entity profile pages. `ActivityFeedBlock` appears on every document detail
page.

### Family 4 — Reports

Every report screen is the same four blocks in sequence:

```
ReportHeaderBlock → ReportFilterBlock → ReportTableBlock → ReportChartBlock
```

The only variation is filter options, column definitions, and grouping.

### Family 5 — Dashboards

Dashboards compose blocks from all other families. No new building blocks are
introduced for dashboards — only new compositions of `KPIRowBlock`,
`ChartPanelBlock`, `ActivityPanelBlock`, and `QuickActionsBlock`.

---

## The Permission Rule

Every block function accepts `ui.UISessionContext`. The block decides internally
what to render based on permissions.

```go
// Correct — the block decides what to show
blocks.ApprovalWorkflowBlock(sess)

// Wrong — the caller should not check permissions before calling a block
if sess.Can("read", "finance.approvals") {
    blocks.ApprovalWorkflowBlock(sess)
}
```

The second pattern is banned. It leaks permission logic out of the block layer,
making it impossible to audit and easy to miss. The CI guard
(`scripts/check-arch.sh` Guard 2 + Guard 3) enforces the boundaries.

---

## The 60-Line Screen Rule

Every `screens/*.go` file must be under 60 lines. The CI guard enforces this.

A screen file over 60 lines is a smell — it means logic escaped the block layer
into the screen. The fix is always to extract the logic into a block:

1. Identify the repeated or complex UI concern in the screen file.
2. Create (or extend) a block in `blocks/` that encapsulates it.
3. Replace the inline logic with a single block call.

---

## Domain Builders

`dsl/builders/` provides convenience constructors for domain-specific picker
nodes that appear in many blocks:

| File | Builders |
|------|---------|
| `finance.go` | `AccountPickerNode`, `CurrencyPickerNode`, `TaxRatePickerNode`, `PaymentMethodPickerNode` |
| `inventory.go` | `ProductPickerNode`, `UOMPickerNode`, `WarehousePickerNode` |
| `hr.go` | `EmployeePickerNode`, `DepartmentPickerNode`, `JobPositionPickerNode` |
| `approval.go` | `ApprovalStatusSelect`, `ApproverPickerNode` |
| `tenant.go` | `TenantStatusSelect`, `PlanPickerNode` |

Builders return `ast.Node`, never raw maps.
