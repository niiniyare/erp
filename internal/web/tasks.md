# UI Architecture Evolution — Task Tracker

> **How to use:** Mark `[ ]` → `[x]` when a task is complete. Add a date if you like:
> `[x] (2026-01-15)`. New developers: start at the first unchecked task, read the
> **Cross-Cutting Foundations** section first, then the task spec.

---

## Cross-Cutting Foundations

Two production-ready packages already exist. Every task in this tracker **must** use
them. Do not introduce new error types, new cache wrappers, or in-process data stores.

### 1 · Error Handling — `awo.so/internal/shared/errors`

Use `BusinessError` for every error that crosses a layer boundary.
Never use bare `fmt.Errorf` or `errors.New` for errors surfaced to callers.

```go
// Constructing a UI-layer error
return sharedErrors.NewBusinessError("SCHEMA_COMPILE_FAILED", "Failed to compile UI schema").
    WithHTTPStatus(http.StatusInternalServerError).
    WithCategory(sharedErrors.CategorySystem).
    WithDetail("page", pageID).
    WithDetail("stage", "compile").
    WithSuggestion("Check that all required node fields are populated")

// Checking errors in pipeline stages
if sharedErrors.IsBusinessErrorCode(err, "SCHEMA_COMPILE_FAILED") { ... }

// HTTP response — use the existing converter, never write your own
httpErr := sharedErrors.ToHTTPError(err)
c.Status(httpErr.Status).JSON(httpErr)
```

Domain-specific UI error codes belong in `internal/web/errors.go`, following the same
pattern as `internal/core/finance/errors.go` — define constants, predefined sentinels,
and contextual constructors. Do not add UI error codes to the shared package itself.

Each phase owns a code prefix:

| Phase / layer       | Code prefix    |
|---------------------|----------------|
| AST compilation     | `AST_*`        |
| Pipeline validation | `VALIDATE_*`   |
| DAG                 | `DAG_*`        |
| Page registry       | `REGISTRY_*`   |
| DSL block construction | `DSL_*`     |
| Cache / versioning  | `CACHE_*`      |

### 2 · Caching — `awo.so/internal/platform/cache`

All UI pipeline caching goes through `cache.Service`. The implementation already
provides multi-tenant key isolation (tenant context is injected by `TenantMiddleware`
before any handler runs), gzip compression, circuit-breaking, distributed locking,
and OpenTelemetry tracing — all for free.

```go
// Store a compiled schema — tenant isolation is automatic from ctx
cacheService.Set(ctx, "ui:schema:"+pageID, compiled, 10*time.Minute)

// Standard cache-aside pattern
var s CompiledSchema
if err := cacheService.Get(ctx, "ui:schema:"+pageID, &s); errors.Is(err, cache.ErrCacheMiss) {
    s = compileSchema(page)
    cacheService.Set(ctx, "ui:schema:"+pageID, s, 10*time.Minute)
}

// Invalidate all compiled schemas for this tenant (e.g. after a config change)
cacheService.DeletePattern(ctx, "ui:schema:*")

// Global (non-tenant) cache for static, permission-free building blocks
cacheService.SetGlobalMemory("ui:dsl:block:line-items", compiled, 0) // no expiry
```

**Rules:**
- Static blocks (no per-request data, no permissions) → `SetGlobalMemory` at startup
- Permission-filtered blocks (vary by role) → `SetMemory` per tenant, short TTL
- Never use `sync.Map`, `ristretto`, or any other in-process cache in `internal/web/`

---

## The Core Insight: All ERP UI Is Patterns

Before a single line of code is written, every developer on this project must internalise
one idea: **ERP screens are not unique snowflakes — they are combinations of a small set
of recurring patterns.**

An ERP system at scale has hundreds of screens. Without building blocks, each screen
re-implements the same concerns slightly differently. With building blocks, each screen
is 10–20 lines of composition. The blocks carry all the complexity.

There are five pattern families. Every screen in AWO ERP belongs to one or more of them:

```
┌──────────────────────────┬──────────────────────────────────────────────┐
│  PATTERN FAMILY          │  EXAMPLES                                    │
├──────────────────────────┼──────────────────────────────────────────────┤
│  1. Document Forms       │  Invoice, PO, Bill, Credit Note, Expense,    │
│                          │  Journal Entry, Employee Record, Contract     │
├──────────────────────────┼──────────────────────────────────────────────┤
│  2. Listing Pages        │  Invoice List, Customer List, Product List,  │
│  (filter bar + table)    │  Transaction History, Audit Log, Stock List  │
├──────────────────────────┼──────────────────────────────────────────────┤
│  3. Data Display         │  Detail views, read-only summaries,          │
│  (cards, boards, trees)  │  entity hierarchy, approval queue, kanban    │
├──────────────────────────┼──────────────────────────────────────────────┤
│  4. Reports              │  P&L, Balance Sheet, Ageing, Cash Flow,      │
│                          │  Stock Valuation, Tax Report, Payroll        │
├──────────────────────────┼──────────────────────────────────────────────┤
│  5. Dashboards           │  Finance overview, Operations overview,      │
│                          │  HR summary, per-module landing pages        │
└──────────────────────────┴──────────────────────────────────────────────┘
```

The goal of Phase 7 (DSL Building Blocks) is to implement one canonical block per
recurring concern, so that assembling a new screen is composition, not construction.

---

## Phase 1: Typed UI AST (Foundation)

- [x] **TASK 1 — Core AST Node Interfaces** (2026-05-18)
  - Files: `internal/web/ast/node.go`, `internal/web/ast/errors.go`,
    `internal/web/ast/compile.go`
  - Define `Node` and `ImmutableNode` interfaces
  - Implement `CompileTree` with validation-before-compilation
  - All compile/validation errors must be `sharedErrors.BusinessError` with code `AST_*`
  - Status: _Complete_

- [x] **TASK 2 — Layout Container Nodes** (2026-05-18)
  - Files: `internal/web/ast/layout.go`
  - Implement: `PageNode`, `GridNode`, `FlexNode`, `TabsNode`, `SplitPaneNode`,
    `SectionNode` (collapsible labelled group — used extensively in document forms)
  - Status: _Complete_

- [x] **TASK 3 — Data Display Nodes** (2026-05-18)
  - Files: `internal/web/ast/display.go`
  - Implement: `TableNode`, `CRUDNode`, `ChartNode`, `CardNode`,
    `StatNode` (single KPI with label/value/trend), `TimelineNode`, `TreeNode`, `APISpec`
  - Critical: `syncLocation:true` structural invariant must be enforced at compile time —
    return `BusinessError` with code `AST_SYNC_LOCATION_MISSING` from `CompileTree`,
    not as a runtime render failure
  - Status: _Complete_ — syncLocation always emitted by CRUDNode.Compile(); transparent bg always emitted by ChartNode.Compile()

- [x] **TASK 4 — Form & Input Nodes** (2026-05-18)
  - Files: `internal/web/ast/form.go`
  - Implement: `FormNode`, `FilterBarNode` (distinct from `FormNode` — see Phase 7),
    `InputTextNode`, `InputNumberNode`, `InputDateNode`, `InputDateRangeNode`,
    `SelectNode`, `ComboNode`, `MultiSelectNode`, `CheckboxNode`, `ToggleNode`,
    `ActionNode`, `DialogNode`, `DrawerNode`
  - Status: _Complete_ — ActionNode canonical definition in display.go (used by CRUDNode); form.go imports it naturally within same package

- [x] **TASK 5 — AST Compiler Integration** (2026-05-18)
  - Files: `internal/web/ui/types.go`, `internal/web/ui/pipeline.go`,
    `internal/web/stages/compile.go`, `internal/web/stages/normalize.go`
  - Add `ASTPageFn` type; dual dispatch in `CompileStage` so legacy `PageFn` continues
    to work during migration
  - `DataKeyASTPageFn` + `DataKeyASTCompiled` keys added to pipeline.go
  - NormalizeStage skips structural checks when `DataKeyASTCompiled` is true
  - Cache path unchanged — AST-compiled schemas go through existing CacheStoreStage
  - Status: _Complete_

---

## Phase 2: Stage DAG Validation

- [x] **TASK 6 — DAG Validator at Pipeline Startup** (2026-05-18)
  - Files: `internal/pipeline/dag.go`, `internal/web/wire.go`
  - `ValidateDAG` added to StageRegistry; DFS cycle detection + missing-dep checks
  - Returns `*DAGError` aggregating all violations
  - Called in `wire.go` after registration; panics on invalid graph
  - Validates `ui.OperationKey` only (AppOperationKey has different stage set)
  - Status: _Complete_

- [x] **TASK 7 — Wire `DependsOn` into All UI Stages** (2026-05-18)
  - Files: `authz.go`, `cache.go`, `registry.go`, `compile.go`, `normalize.go`,
    `response.go`
  - session(root) → authz → cache_lookup, registry → compile → normalize → validate → cache_store
  - response has no DependsOn (priority-ordered last; cache-hit uses NextStageID)
  - Status: _Complete_

---

## Phase 3: Cache Generation Versioning

- [x] **TASK 8 — Generation-Aware Cache Key** (2026-05-18)
  - Files: `internal/web/cache/version.go`, `internal/web/cache/invalidation.go`,
    `internal/web/authz/service.go`, `internal/web/wire.go`,
    `internal/web/handler/schema.go`
  - `CacheVersions` struct (CompilerVersion, ASTVersion, PolicyGeneration, SchemaGeneration);
    8-component key via `uicache.Key`; 4 invalidation scopes via `InvalidateSchemaCache`
  - `UIPipeline` wrapper auto-injects `DataKeyCacheVersions` into every `opCtx.Data`
  - `SchemaHandler` now accepts `PipelineRunner` interface — satisfied by both `*UIPipeline`
    and `*pipeline.PipelineBuilder`; `NewUIPipeline` returns `*UIPipeline`
  - `authz.CacheKey` and `authz.InvalidationPattern` deprecated; point to `internal/web/cache`
  - All invalidation uses `cache.Service.DeletePattern` — never direct Redis
  - Status: _Complete_

---

## Phase 4: Normalisation / Validation Separation

- [x] **TASK 9 — Strict `NormalizeStage` Scope Reduction** (2026-05-18)
  - Files: `internal/web/stages/normalize.go`
  - `NormalizeStage`: pure canonicalization (lowercase type, trim api whitespace);
    Execute always returns nil error; skips schema missing gracefully
  - `ValidateStage`: absorbs all structural rules (CRUD syncLocation, chart bg,
    API method prefix) + security rules (no IAM expressions); all errors are
    `*sharedErrors.BusinessError` with `VALIDATE_*` codes + `WithCause(ui.ErrSchemaInvalid)`
    so `errors.Is(err, ui.ErrSchemaInvalid)` in SchemaHandler continues to work
  - Structural rules still skip on `DataKeyASTCompiled=true`; security rules always run
  - Large invariant comment block at top of file explains Normalize vs Validate boundary
  - Status: _Complete_

---

## Phase 5: Registry Dependency Graph

- [ ] **TASK 10 — Page Registry with Module Declarations**
  - Files: `internal/web/registry/registry.go`, `internal/web/wire.go`,
    all `page/*/init()` functions
  - Add `PageRegistration` struct with metadata; implement `ValidateRegistry` startup check
  - Update all 6 existing pages; resolution errors use code prefix `REGISTRY_*`
  - Status: _Not started_

---

## Phase 6: Distributed Observability

- [ ] **TASK 11 — Stage Span Attributes and Metrics**
  - Files: `internal/web/stages/instrument.go`, `internal/web/metrics/ui.go`,
    `internal/web/wire.go`
  - Add `UIStageAttributes` for trace spans via the existing `tracing.Service` interface
  - Define 6 metrics via the existing `metrics.MetricsProvider` (not a new library):
    `ui_compile_duration_ms`, `stage_execution_duration_ms`,
    `schema_validation_failures_total`, `cache_generation_mismatch_total`,
    `invalidation_events_total`, `registry_resolution_failures_total`
  - Status: _Not started_

---

## Phase 7: ERP UI Domain DSL — Building-Block Architecture

> **Read this entire section before writing any DSL code.**

### Directory layout

```
internal/web/dsl/
│
├── blocks/                         # Primitive building blocks
│   │                               # Each file = one reusable concern.
│   │                               # Zero application logic; pure UI composition.
│   │
│   ├── -- Document blocks --
│   │   ├── line_items.go           # ProductServiceLineBlock  ← most widely shared
│   │   ├── document_header.go      # DocumentHeaderBlock  (ref, date, currency, status)
│   │   ├── party.go                # PartyBlock  (customer / supplier / employee)
│   │   ├── address.go              # AddressBlock  (billing / shipping / both)
│   │   ├── tax_summary.go          # TaxSummaryBlock
│   │   ├── totals.go               # TotalsSummaryBlock
│   │   ├── payment_terms.go        # PaymentTermsBlock
│   │   ├── attachments.go          # AttachmentsBlock
│   │   ├── approval.go             # ApprovalWorkflowBlock
│   │   └── notes.go                # InternalNotesBlock
│   │
│   ├── -- Listing blocks --
│   │   ├── filter_bar.go           # FilterBarBlock  ← shared by all listing pages
│   │   ├── data_table.go           # DataTableBlock  (sortable, paginated, selectable)
│   │   ├── bulk_actions.go         # BulkActionsBlock  (delete, export, status change)
│   │   └── empty_state.go          # EmptyStateBlock  (no results / first-use prompt)
│   │
│   ├── -- Data display blocks --
│   │   ├── stat_card.go            # StatCardBlock  (KPI: label + value + trend)
│   │   ├── detail_card.go          # DetailCardBlock  (read-only field group)
│   │   ├── activity_feed.go        # ActivityFeedBlock  (audit trail, comments)
│   │   ├── status_badge.go         # StatusBadgeBlock  (colour-mapped enum display)
│   │   └── entity_breadcrumb.go    # EntityBreadcrumbBlock  (hierarchy path)
│   │
│   ├── -- Report blocks --
│   │   ├── report_header.go        # ReportHeaderBlock  (title, period, run-by)
│   │   ├── report_filter.go        # ReportFilterBlock  (period, entity, currency)
│   │   ├── report_table.go         # ReportTableBlock   (subtotals, grouping, totals row)
│   │   └── report_chart.go         # ReportChartBlock   (bar / line / pie variants)
│   │
│   └── -- Dashboard blocks --
│       ├── kpi_row.go              # KPIRowBlock         (row of StatCards)
│       ├── chart_panel.go          # ChartPanelBlock     (titled chart + period picker)
│       ├── activity_panel.go       # ActivityPanelBlock  (recent transactions / events)
│       └── quick_actions.go        # QuickActionsBlock   (shortcut buttons)
│
├── screens/                        # Composed screen layouts (thin wrappers)
│   │                               # Config in, ast.Node out.
│   │                               # Target: every file under 60 lines.
│   │
│   ├── -- Document screens --
│   │   ├── invoice.go              # InvoiceScreen     (sales + purchase via config)
│   │   ├── purchase_order.go       # POScreen
│   │   ├── credit_note.go          # CreditNoteScreen
│   │   ├── debit_note.go           # DebitNoteScreen
│   │   ├── expense_claim.go        # ExpenseClaimScreen
│   │   ├── goods_receipt.go        # GoodsReceiptScreen
│   │   ├── delivery_note.go        # DeliveryNoteScreen
│   │   ├── journal_entry.go        # JournalEntryScreen
│   │   └── payment_voucher.go      # PaymentVoucherScreen
│   │
│   ├── -- Listing screens --
│   │   ├── list_screen.go          # GenericListScreen  (covers ~80% of listing pages)
│   │   └── approval_queue.go       # ApprovalQueueScreen
│   │
│   ├── -- Report screens --
│   │   ├── profit_loss.go          # ProfitLossScreen
│   │   ├── balance_sheet.go        # BalanceSheetScreen
│   │   ├── trial_balance.go        # TrialBalanceScreen
│   │   ├── cash_flow.go            # CashFlowScreen
│   │   ├── ageing.go               # AgeingScreen  (AR + AP variant)
│   │   └── tax_report.go           # TaxReportScreen
│   │
│   └── -- Dashboard screens --
│       ├── finance_dashboard.go    # FinanceDashboardScreen
│       ├── operations_dashboard.go # OperationsDashboardScreen
│       └── module_overview.go      # ModuleOverviewScreen (per-module landing page)
│
└── builders/                       # Domain-specific convenience builders
    ├── finance.go                  # Account pickers, GL code selectors
    ├── inventory.go                # Product/service catalogue pickers
    ├── hr.go                       # Employee pickers, leave type selectors
    ├── approval.go                 # Approval chain builders
    └── tenant.go                   # Tenant admin builders
```

---

### Family 1 — Document Forms

Every document form in AWO ERP is built from the same set of blocks.
The differences between an Invoice and a Purchase Order are **configuration**, not code.

**Shared block matrix:**

| Block                | Invoice | PO  | Bill | Credit Note | Expense | GRN | Journal |
|----------------------|:-------:|:---:|:----:|:-----------:|:-------:|:---:|:-------:|
| DocumentHeader       | ✓       | ✓   | ✓    | ✓           | ✓       | ✓   | ✓       |
| Party                | ✓       | ✓   | ✓    | ✓           | ✓       | ✓   | —       |
| Address              | ✓       | ✓   | ✓    | ✓           | —       | ✓   | —       |
| **ProductServiceLine** | **✓** | **✓** | **✓** | **✓**    | **✓**   | **✓** | —    |
| TaxSummary           | ✓       | ✓   | ✓    | ✓           | ✓       | —   | —       |
| TotalsSummary        | ✓       | ✓   | ✓    | ✓           | ✓       | ✓   | ✓       |
| PaymentTerms         | ✓       | ✓   | ✓    | —           | —       | —   | —       |
| Attachments          | ✓       | ✓   | ✓    | ✓           | ✓       | ✓   | ✓       |
| ApprovalWorkflow     | ✓       | ✓   | ✓    | ✓           | ✓       | ✓   | ✓       |
| InternalNotes        | ✓       | ✓   | ✓    | ✓           | ✓       | ✓   | ✓       |

**`ProductServiceLineBlock` is the most important block in the entire DSL.** It is
the block that, if done wrong, will be copy-pasted into every document. Design it
first, and design it carefully. Every document that has line items must use it.

```go
// internal/web/dsl/blocks/line_items.go

// LineItemConfig controls columns and behaviour per document type.
// Documents opt-in to the columns they need; unused columns are hidden.
type LineItemConfig struct {
    ShowProductCode   bool
    ShowDescription   bool
    ShowQty           bool
    ShowUnitOfMeasure bool
    ShowUnitPrice     bool
    ShowDiscount      bool   // invoices + credit notes
    ShowTaxRate       bool   // not needed on GRN, delivery note
    ShowSubtotal      bool   // always true — not configurable
    ShowAccount       bool   // journal entry and expense lines use GL account

    AllowFreeTextItem bool   // false = must select from product catalogue
    DefaultCurrency   string
    MaxLines          int    // 0 = unlimited
    ReadOnly          bool   // for view/detail screens
}

func DefaultLineItemConfig() LineItemConfig { ... }   // standard commercial document
func GRNLineItemConfig()     LineItemConfig { ... }   // no tax, no price editing
func JournalLineItemConfig() LineItemConfig { ... }   // account + debit/credit columns

// ProductServiceLineBlock returns the shared line items node.
// Every document form with line items MUST use this function.
// Never define a custom line item table in a screen file.
func ProductServiceLineBlock(sess domain.UISessionContext, cfg LineItemConfig) ast.Node { ... }
```

Composing a document screen is then straightforward:

```go
// internal/web/dsl/screens/invoice.go  (~30 lines)

type InvoiceScreenConfig struct {
    IsPurchase        bool
    ShowPaymentTerms  bool
    ShowAttachments   bool
    ShowInternalNotes bool
    ReadOnly          bool
}

func InvoiceScreen(sess domain.UISessionContext, cfg InvoiceScreenConfig) ast.Node {
    lineCfg := blocks.DefaultLineItemConfig()
    lineCfg.ShowDiscount = true
    lineCfg.ReadOnly = cfg.ReadOnly

    return ast.Page(
        blocks.DocumentHeaderBlock(sess, ...),
        blocks.PartyBlock(sess, ...),
        blocks.AddressBlock(sess, ...),
        blocks.ProductServiceLineBlock(sess, lineCfg),
        blocks.TaxSummaryBlock(sess),
        blocks.TotalsSummaryBlock(sess),
        blocks.ApprovalWorkflowBlock(sess),
        conditional(cfg.ShowPaymentTerms,  blocks.PaymentTermsBlock(sess)),
        conditional(cfg.ShowAttachments,   blocks.AttachmentsBlock(sess)),
        conditional(cfg.ShowInternalNotes, blocks.InternalNotesBlock(sess)),
    )
}
```

---

### Family 2 — Listing Pages (Filter Bar + Table)

Every listing page in AWO ERP follows the same structure:

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

**`FilterBarBlock` shared concerns across listing pages:**

| Filter          | Invoice List | Customer List | Product List | Txn History | Audit Log |
|-----------------|:---:|:---:|:---:|:---:|:---:|
| Date range      | ✓   | —   | —   | ✓   | ✓   |
| Status          | ✓   | ✓   | ✓   | ✓   | —   |
| Entity / tenant | ✓   | ✓   | —   | ✓   | ✓   |
| Amount range    | ✓   | —   | ✓   | ✓   | —   |
| Free text search | ✓  | ✓   | ✓   | ✓   | ✓   |
| Currency        | ✓   | —   | ✓   | ✓   | —   |
| Category / type | ✓   | —   | ✓   | ✓   | ✓   |

```go
// internal/web/dsl/blocks/filter_bar.go

type FilterBarConfig struct {
    ShowDateRange     bool
    ShowStatus        bool
    StatusOptions     []SelectOption
    ShowEntityPicker  bool
    ShowAmountRange   bool
    ShowCurrency      bool
    ShowTypeFilter    bool
    TypeOptions       []SelectOption
    ShowSearch        bool
    SearchPlaceholder string
    Collapsible       bool   // on mobile, collapses to a [Filter] button
}

// FilterBarBlock is the shared filter header used on every listing page.
// Emits an filter-change event that DataTableBlock listens to.
// Never write a custom filter form in a screen file.
func FilterBarBlock(sess domain.UISessionContext, cfg FilterBarConfig) ast.Node { ... }
```

Most listing pages need nothing beyond `GenericListScreen`:

```go
// internal/web/dsl/screens/list_screen.go

type ListScreenConfig struct {
    Title          string
    Resource       string           // "invoices", "customers" — used in API calls
    Columns        []ColumnDef
    FilterBar      FilterBarConfig
    AllowCreate    bool
    AllowExport    bool
    BulkActions    []BulkActionDef
    RowClickAction string           // "navigate" | "drawer" | "dialog"
}

// GenericListScreen composes FilterBarBlock + DataTableBlock + BulkActionsBlock.
// This single function covers the invoice list, customer list, product list,
// transaction history, audit log, employee list, user list, and more.
func GenericListScreen(sess domain.UISessionContext, cfg ListScreenConfig) ast.Node { ... }
```

---

### Family 3 — Data Display (Cards, Detail Views, Boards)

Read-only views of a single record — accessed by clicking a row, or rendered as the
left-hand panel of a split detail view.

`StatCardBlock` and `DetailCardBlock` are the two workhorses. They appear in dashboards,
document detail headers, and entity profile pages alike.

```go
// StatCardBlock — one KPI figure with label and optional trend arrow.
// Used in dashboards AND in document detail headers
// (e.g. "Total Due: KES 45,200  ↑ from last month").
blocks.StatCardBlock(sess, StatCardConfig{
    Label:    "Outstanding AR",
    ValueKey: "ar_balance",
    Currency: "KES",
    Trend:    TrendDown,  // green when falling (you owe less), red when rising
})

// DetailCardBlock — a read-only field group (label: value pairs).
// Used in document view mode, customer profile, product detail page.
blocks.DetailCardBlock(sess, DetailCardConfig{
    Title:  "Customer Details",
    Fields: []FieldDef{
        {Label: "Name",    Key: "customer_name"},
        {Label: "Phone",   Key: "phone"},
        {Label: "Email",   Key: "email"},
        {Label: "Balance", Key: "balance", Format: "currency"},
    },
})

// ActivityFeedBlock — chronological list of events (status changes, comments, edits).
// Used on every document detail page and on the operations dashboard.
blocks.ActivityFeedBlock(sess, ActivityFeedConfig{
    Resource:     "invoice",
    ResourceID:   invoiceID,
    ShowComments: true,
})
```

---

### Family 4 — Reports

Reports are the most formulaic family. Every report screen is the same four blocks
in sequence. The only variation is filter options, column definitions, and grouping.

```
ReportHeaderBlock → ReportFilterBlock → ReportTableBlock → ReportChartBlock
```

```go
// internal/web/dsl/blocks/report_table.go

type ReportTableConfig struct {
    Columns        []ReportColumnDef
    GroupBy        []string      // column keys to group on (e.g. account_type, section)
    ShowSubtotals  bool
    ShowGrandTotal bool
    Exportable     bool          // adds CSV / PDF export buttons
    Currency       string
}

// ReportTableBlock handles subtotals, grand total row, and multi-level grouping.
// Every financial report (P&L, Balance Sheet, Ageing, Trial Balance) MUST use this.
// Never define a custom total row in a report screen file.
func ReportTableBlock(sess domain.UISessionContext, cfg ReportTableConfig) ast.Node { ... }
```

A complete report screen:

```go
// internal/web/dsl/screens/profit_loss.go  (~25 lines)

func ProfitLossScreen(sess domain.UISessionContext) ast.Node {
    return ast.Page(
        blocks.ReportHeaderBlock(sess, blocks.ReportHeaderConfig{
            Title: "Profit & Loss Statement",
        }),
        blocks.ReportFilterBlock(sess, blocks.ReportFilterConfig{
            ShowPeriod:       true,
            ShowEntityPicker: true,
            ShowCurrency:     true,
            ShowComparison:   true,   // vs prior period
        }),
        blocks.ReportTableBlock(sess, blocks.ReportTableConfig{
            Columns:        plColumns(),
            GroupBy:        []string{"section", "account_type"},
            ShowSubtotals:  true,
            ShowGrandTotal: true,
            Exportable:     true,
        }),
        blocks.ReportChartBlock(sess, blocks.ReportChartConfig{
            Type:   ChartTypeBar,
            Series: []string{"income", "expenses", "net_profit"},
        }),
    )
}
```

---

### Family 5 — Dashboards

Dashboards compose blocks from all other families. The insight here is that
`StatCardBlock`, `ChartPanelBlock`, and `ActivityFeedBlock` are **already defined**
in families 3 and 4. Dashboards add no new building blocks — only new compositions.

```go
// internal/web/dsl/screens/finance_dashboard.go  (~35 lines)

func FinanceDashboardScreen(sess domain.UISessionContext) ast.Node {
    return ast.Page(
        blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
            {Label: "Total Revenue",  ValueKey: "revenue",     Trend: TrendUp},
            {Label: "Outstanding AR", ValueKey: "ar_balance",  Trend: TrendDown},
            {Label: "Outstanding AP", ValueKey: "ap_balance",  Trend: TrendDown},
            {Label: "Cash & Bank",    ValueKey: "cash_balance"},
        }),

        ast.Grid(cols(2),
            blocks.ChartPanelBlock(sess, blocks.ChartPanelConfig{
                Title:        "Revenue vs Expenses",
                ChartType:    ChartTypeLine,
                PeriodPicker: true,
            }),
            blocks.ActivityPanelBlock(sess, blocks.ActivityPanelConfig{
                Title:    "Recent Transactions",
                Resource: "transactions",
                Limit:    10,
            }),
        ),

        blocks.QuickActionsBlock(sess, []blocks.QuickAction{
            {Label: "New Invoice",    Permission: "finance.invoices.create"},
            {Label: "Record Payment", Permission: "finance.payments.create"},
            {Label: "View Reports",   Permission: "finance.reports.read"},
        }),
    )
}
```

---

### The Permission Rule (applies to all families)

Every block function accepts `domain.UISessionContext`. The block decides internally
what to render based on the session's permissions. Callers must never gate-keep before
calling a block — that belongs inside the block.

```go
// Correct — the block decides internally what to show
blocks.ProductServiceLineBlock(sess, lineCfg)

// Wrong — caller should not check permissions before calling a block
if sess.HasPermission("finance.invoices.write") {
    blocks.ProductServiceLineBlock(sess, lineCfg)
}
```

---

- [ ] **TASK 12 — Implement All DSL Building Blocks and Screen Layouts**
  - Files: `internal/web/dsl/blocks/*.go`, `internal/web/dsl/screens/*.go`,
    `internal/web/dsl/builders/*.go`
  - **Step 1 — Document blocks:** Start with `ProductServiceLineBlock`
    (design in parallel with T3/T4 since it affects TableNode and FormNode types),
    then `DocumentHeaderBlock`, `PartyBlock`, `TaxSummaryBlock`, `TotalsSummaryBlock`.
    These five blocks unblock all document screens.
  - **Step 2 — Listing blocks:** `FilterBarBlock`, `DataTableBlock`, `BulkActionsBlock`.
    Then `GenericListScreen` — this single function unblocks all listing pages.
  - **Step 3 — Data display blocks:** `StatCardBlock`, `DetailCardBlock`,
    `ActivityFeedBlock`, `StatusBadgeBlock`. These unblock dashboards and detail views.
  - **Step 4 — Report blocks:** `ReportHeaderBlock`, `ReportFilterBlock`,
    `ReportTableBlock`, `ReportChartBlock`. Then implement all report screens.
  - **Step 5 — Dashboard blocks:** `KPIRowBlock`, `ChartPanelBlock`,
    `ActivityPanelBlock`, `QuickActionsBlock`. Then implement dashboard screens.
  - **Step 6 — Screen composition:** Implement all files in `screens/` by composing
    blocks. Each screen file must be under 60 lines. If it is not, logic has leaked
    from a block into the screen layer — extract it.
  - **Step 7 — Domain builders:** Finance, inventory, HR, approval, tenant in
    `builders/`.
  - **Hard constraints:**
    - Zero `map[string]any` anywhere in `dsl/`
    - Zero duplicated block logic between `screens/` files — if two screens share
      a visual concern, it must be a block
    - All block-construction errors use `sharedErrors.BusinessError` with prefix `DSL_*`
    - Static blocks (no permissions) cached via `cache.Service.SetGlobalMemory` at startup
    - Permission-filtered blocks cached via `cache.Service.SetMemory` per tenant, short TTL
  - Status: _Not started_

---

## Phase 8: Strict Architectural Enforcement (CI Guards)

- [ ] **TASK 13 — CI Architecture Guards**
  - Files: `scripts/check-arch.sh`, `.github/workflows/ci.yml`
  - Guard 1: No `map[string]any` in `internal/web/` — zero tolerance
  - Guard 2: No direct IAM imports from `dsl/` — only via `domain.UISessionContext`
  - Guard 3: No permission string checks in `visibleOn` predicates
  - Guard 4: No schema generated outside the pipeline
  - Guard 5: Every stage has a non-empty `DependsOn`
  - Guard 6: No `sync.Map` or third-party in-process cache in `internal/web/`
  - Guard 7: No block-level concerns (line item columns, tax rows, filter inputs,
    KPI stat cards, report total rows) defined in more than one `screens/` file —
    if the grep finds the same structural pattern in two places, extract a block
  - Guard 8: Every `screens/` file is under 60 lines — a screen over 60 lines is
    a build warning; it means logic escaped the block layer
  - Status: _Not started_

---

## Phase 9: Documentation

- [ ] **TASK 14 — Architecture and DSL Docs**
  - Files: `docs/reference/modules/ui/architecture.md`, `ast.md`, `pipeline.md`,
    `cache.md`, `dsl.md`, `contributing.md`
  - No references to `map[string]any` as a schema type anywhere
  - Include the pipeline DAG diagram
  - Include the five-family pattern table from this task file
  - Include the block directory tree from Phase 7
  - Document the cache key naming convention and invalidation scopes
  - Document the permission rule: blocks own permission checks, screens do not
  - Document the 60-line screen size guideline as a smell detector
  - Status: _Not started_

---

## Parallel Execution Map

| Track              | Tasks              | Can Start When              |
|--------------------|--------------------|-----------------------------|
| AST                | T1–5               | Now                         |
| DAG                | T6–7               | Now                         |
| Cache versioning   | T8                 | Now                         |
| Normalise          | T9                 | T5 done                     |
| Registry           | T10                | T1 done                     |
| Observability      | T11                | T8 done                     |
| DSL blocks Steps 1–5 | T12            | T1 done (needs AST types)   |
| DSL screens Step 6 | T12                | T12 Steps 1–5 done          |
| CI guards          | T13                | All above done              |
| Docs               | T14                | Last                        |

**Important:** `ProductServiceLineBlock` design affects `TableNode` and `FormNode`
in Tasks 3 and 4. Plan T12 Step 1 alongside T3/T4 — do not finalise those AST node
types before the line item block column model is agreed.

---

## Current Blockers

> _Edit this section when something is blocking progress._

- None yet

---

## Screen Inventory and Migration Status

### Document Screens

| Screen               | Blocks Used                                              | Legacy | AST | Done |
|----------------------|----------------------------------------------------------|:------:|:---:|:----:|
| Sales Invoice        | Header, Party, Address, LineItems, Tax, Totals, Approval, PayTerms | ✓ | ☐ | ☐ |
| Purchase Invoice (Bill) | same as above                                         | ☐      | ☐   | ☐   |
| Purchase Order       | Header, Party, Address, LineItems, Tax, Totals, PayTerms, Approval | ☐ | ☐ | ☐ |
| Sales Order          | same as PO                                               | ☐      | ☐   | ☐   |
| Credit Note          | Header, Party, LineItems, Tax, Totals, Approval          | ☐      | ☐   | ☐   |
| Debit Note           | same as Credit Note                                      | ☐      | ☐   | ☐   |
| Expense Claim        | Header, Party, LineItems, Tax, Totals, Attachments, Approval | ☐  | ☐   | ☐   |
| Goods Receipt        | Header, Party, LineItems (GRN cfg), Totals, Attachments, Approval | ☐ | ☐ | ☐ |
| Delivery Note        | Header, Party, Address, LineItems (no price), Attachments | ☐     | ☐   | ☐   |
| Journal Entry        | Header, LineItems (journal cfg), Totals, Attachments, Approval | ✓ | ☐  | ☐   |
| Payment Voucher      | Header, Party, PayTerms, Attachments, Approval           | ☐      | ☐   | ☐   |

### Listing Screens (all use `GenericListScreen`)

| Screen               | Legacy | AST | Done |
|----------------------|:------:|:---:|:----:|
| Invoice List         | ✓      | ☐   | ☐    |
| Customer List        | ☐      | ☐   | ☐    |
| Supplier List        | ☐      | ☐   | ☐    |
| Product / Service List | ☐    | ☐   | ☐    |
| Transaction History  | ✓      | ☐   | ☐    |
| Audit Log            | ☐      | ☐   | ☐    |
| Approval Queue       | ☐      | ☐   | ☐    |
| HR / Employees       | ✓      | ☐   | ☐    |
| IAM / Users          | ✓      | ☐   | ☐    |

### Report Screens

| Screen               | Legacy | AST | Done |
|----------------------|:------:|:---:|:----:|
| Profit & Loss        | ☐      | ☐   | ☐    |
| Balance Sheet        | ☐      | ☐   | ☐    |
| Trial Balance        | ☐      | ☐   | ☐    |
| Cash Flow            | ☐      | ☐   | ☐    |
| AR Ageing            | ☐      | ☐   | ☐    |
| AP Ageing            | ☐      | ☐   | ☐    |
| Tax Report           | ☐      | ☐   | ☐    |
| Reports Hub          | ✓      | ☐   | ☐    |

### Dashboard Screens

| Screen                 | Legacy | AST | Done |
|------------------------|:------:|:---:|:----:|
| Main Dashboard         | ✓      | ☐   | ☐    |
| Finance Dashboard      | ☐      | ☐   | ☐    |
| Operations Dashboard   | ☐      | ☐   | ☐    |

---

## Quick Resume Guide

1. Find the first unchecked `[ ]` task above
2. Re-read **Cross-Cutting Foundations** before every session
3. For DSL tasks, re-read the relevant family section in Phase 7
4. When done: `[ ]` → `[x]`, optionally add a date

**Current priority:** Phase 1 (AST) is the foundation. Start with TASK 1.
Discuss `ProductServiceLineBlock` column model (T12 Step 1) in parallel with
T3/T4 — the AST table and form node types must accommodate it before those
tasks are signed off.
