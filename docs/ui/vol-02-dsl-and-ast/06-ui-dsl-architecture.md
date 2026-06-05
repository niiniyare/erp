# Chapter 06 — UI DSL Architecture

> **Volume:** II — DSL and AST
> **Audience:** Backend Engineers, Platform Engineers
> **Prerequisites:** Chapter 05 — SDUI Fundamentals

---

## Table of Contents

- [6.1 Three Layers of the DSL](#61-three-layers-of-the-dsl)
- [6.2 DSL Blocks](#62-dsl-blocks)
- [6.3 DSL Screens](#63-dsl-screens)
- [6.4 DSL Builders](#64-dsl-builders)
- [6.5 The AMIS Builder Package (Legacy)](#65-the-amis-builder-package-legacy)
- [6.6 Choosing the Right Layer](#66-choosing-the-right-layer)
- [6.7 Registering a New Page](#67-registering-a-new-page)
- [6.8 Worked Example: Invoice Listing Page](#68-worked-example-invoice-listing-page)

---

## 6.1 Three Layers of the DSL

The Go UI DSL has three layers, each building on the one below:

```
┌──────────────────────────────────────────────────┐
│  DSL Screens  (dsl/screens/)                     │
│  Complete page functions for registered routes   │
├──────────────────────────────────────────────────┤
│  DSL Blocks   (dsl/blocks/)                      │
│  Reusable, session-aware schema fragments        │
├──────────────────────────────────────────────────┤
│  AST Package  (ast/)                             │
│  Typed node structs; CompileTree for validation  │
└──────────────────────────────────────────────────┘
```

**AST nodes** are the primitive vocabulary: `PageNode`, `CRUDNode`, `FormNode`, `GridNode`, `TabsNode`, `ActionNode`, etc. They map directly to AMIS component types.

**Blocks** compose AST nodes into ERP-domain fragments. A `DataTableBlock` builds a complete, paginated, permission-filtered CRUD table from a `DataTableConfig`. A `StatusBadgeBlock` emits a correctly-typed mapping node. Authors call blocks, not individual AST nodes.

**Screens** compose blocks into complete pages. A screen is exactly a `PageFn` or `ASTPageFn` — it receives a `UISessionContext` and returns a root `ast.Node`. Screens are the unit that gets registered in the page registry.

---

## 6.2 DSL Blocks

**Location:** `internal/web/dsl/blocks/`

Each block is a function with signature `func(sess ui.UISessionContext, cfg XxxConfig) ast.Node` (or `ast.TableColumn`, etc. for sub-components).

### Implemented Blocks

**DataTableBlock** — Paginated, filterable CRUD table.

```go
// blocks/data_table.go
func DataTableBlock(sess ui.UISessionContext, cfg DataTableConfig) ast.Node

type DataTableConfig struct {
    APIURL           string
    Columns          []ColumnDef
    RowActions       []ast.ActionNode
    BulkActions      []BulkActionDef
    Filter           ast.Node
    PrimaryKey       string
    PageSize         int
    Title            string
    AllowCreate      bool
    CreateURL        string
    CreatePermission string  // e.g. "finance.invoices.create" — required when AllowCreate is true
    AllowExport      bool
}
```

> **Note:** `CreatePermission` is a required field when `AllowCreate` is true. Without it, the "New" button is never shown, regardless of the user's actual permissions. This is intentional — explicit is better than derived.

**StatusBadgeBlock** — Colour-mapped status indicator.

```go
// blocks/status_badge.go
func StatusBadgeBlock(_ ui.UISessionContext, cfg StatusBadgeConfig) ast.Node
func StatusBadgeColumn(cfg StatusBadgeConfig) ast.TableColumn
func StatusBadgeColumnDef(cfg StatusBadgeConfig) ColumnDef

type StatusBadgeConfig struct {
    FieldName string
    Label     string
    Mappings  []StatusMapping
}

type StatusMapping struct {
    Value string
    Label string
    Color string // "success"|"warning"|"danger"|"info"|"default"
}
```

> **Note:** `StatusBadgeBlock` emits a `MappingNode` (AMIS type `"mapping"`). Do not use AMIS type `"status"` — it is not a valid AMIS component type.

**DetailCardBlock** — Property list for read-only record detail views. Emits a `PropertyNode` (AMIS type `"property"`), not a single-row table.

**LineItemsBlock** — Editable line-item grid for invoices, purchase orders, journals.

**ApprovalBlock** — Approval action bar with configurable workflow actions (approve, reject, delegate, escalate).

**QuickActionsBlock** — Permission-filtered grid of quick-action buttons for dashboard screens.

### Example: Using StatusBadgeColumnDef in a DataTableBlock

```go
func InvoiceListScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "Invoices",
        Body: []ast.Node{
            blocks.DataTableBlock(sess, blocks.DataTableConfig{
                APIURL:           "/api/v1/finance/invoices",
                PrimaryKey:       "id",
                PageSize:         20,
                AllowCreate:      true,
                CreateURL:        "/finance/invoices/new",
                CreatePermission: "finance.invoices.create",
                AllowExport:      sess.Can("export", "finance.invoices"),
                Columns: []blocks.ColumnDef{
                    {Name: "invoice_number", Label: "Invoice #", Sortable: true},
                    {Name: "vendor_name", Label: "Vendor"},
                    {Name: "amount", Label: "Amount", Type: "currency"},
                    blocks.StatusBadgeColumnDef(blocks.StatusBadgeConfig{
                        FieldName: "status",
                        Label:     "Status",
                        Mappings: []blocks.StatusMapping{
                            {Value: "DRAFT", Label: "Draft", Color: "default"},
                            {Value: "PENDING", Label: "Pending", Color: "warning"},
                            {Value: "APPROVED", Label: "Approved", Color: "success"},
                            {Value: "REJECTED", Label: "Rejected", Color: "danger"},
                        },
                    }),
                },
                RowActions: []ast.ActionNode{
                    {Label: "View", ActionType: "link", Target: "/finance/invoices/${id}", Level: "default"},
                },
            }),
        },
    }
}
```

---

## 6.3 DSL Screens

**Location:** `internal/web/dsl/screens/`

Screens are the page functions that get registered in the registry. Each screen corresponds to exactly one route. Screens compose blocks from `dsl/blocks/` and AST nodes from `ast/`.

### Implemented Screens

| Screen | Route |
|--------|-------|
| `FinanceDashboardScreen` | `/finance/dashboard` |
| `InvoiceScreen` (sales) | `/finance/invoices/new` |
| `InvoiceScreen` (purchase) | `/finance/bills/new` |
| `JournalEntryScreen` | `/finance/journal-entries/new` |
| `TrialBalanceScreen` | `/finance/reports/trial-balance` |
| `RegisterScreen` | (registration route) |

Example — `FinanceDashboardScreen`:

```go
// internal/web/dsl/screens/finance_dashboard.go

func FinanceDashboardScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "Finance Dashboard",
        InitAPI: &ast.APISpec{
            Method: "get",
            URL:    "/api/v1/finance/dashboard/summary",
        },
        Body: []ast.Node{
            blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
                {Label: "Total Revenue", ValueKey: "revenue", Format: "currency", Trend: blocks.TrendUp},
                {Label: "Outstanding AR", ValueKey: "ar_balance", Format: "currency", Trend: blocks.TrendDown},
                {Label: "Outstanding AP", ValueKey: "ap_balance", Format: "currency", Trend: blocks.TrendDown},
                {Label: "Cash & Bank", ValueKey: "cash_balance", Format: "currency"},
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
                {Label: "New Invoice", URL: "/finance/invoices/new",
                    Permission: "finance.invoices.create", Icon: "fa fa-file-invoice"},
                {Label: "Record Payment", URL: "/finance/payments/new",
                    Permission: "finance.payments.create", Icon: "fa fa-credit-card"},
            }),
        },
    }
}
```

---

## 6.4 DSL Builders

**Location:** `internal/web/dsl/builders/`

Builders are module-domain helpers that produce common AST patterns for a specific ERP domain. Unlike blocks (which are generic ERP patterns), builders encode domain-specific business knowledge.

Implemented builders:
- `finance.go` — Finance module patterns (ledger account selectors, period pickers)
- `inventory.go` — Inventory patterns (warehouse selectors, lot tracking)
- `hr.go` — HR patterns (employee selectors, leave calendars)
- `approval.go` — Approval workflow action bars
- `tenant.go` — Tenant administration forms

---

## 6.5 The AMIS Builder Package (Legacy)

**Location:** `internal/web/amis/`

The `amis` package predates the typed AST. It provides fluent Go builders that produce raw `map[string]any`. These builders implement `json.Marshaler` so they can be used directly with `c.JSON()` or embedded in larger schemas.

```go
// Legacy pattern — amis package
type Ctx struct {
    Flags map[string]bool
    User  CtxUser
    Can   func(action, resource string) bool
}

type SchemaFn func(ctx Ctx) Schema

// Builders: all implement json.Marshaler — no .Build() call needed
// PageBuilder, CRUDBuilder, FormBuilder, GridBuilder, PanelBuilder,
// TabsBuilder, StatBuilder, ChartBuilder, AlertBuilder, WizardBuilder, ...
```

**This path is deprecated for new development.** The `amis` package remains for existing `PageFn` registrations. New pages use `ASTPageFn` with the typed AST.

---

## 6.6 Choosing the Right Layer

| Situation | Use |
|-----------|-----|
| New page for a registered route | `ASTPageFn` + DSL Screens + DSL Blocks |
| Reusable ERP widget (e.g., status badge column) | DSL Block in `dsl/blocks/` |
| Domain-specific helper (e.g., GL account selector pattern) | DSL Builder in `dsl/builders/` |
| Quick prototype, ad-hoc JSON | `PageFn` + raw `ui.M{}` literals (not for production) |
| Existing page registered with `PageFn` | Do not refactor unless migrating |

---

## 6.7 Registering a New Page

Every page registration happens in an `init()` function, typically in the same file as the screen or in a `register.go` file in the screen's package.

```go
package screens

import "awo.so/internal/web/registry"

func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:       "/finance/invoices",
        Module:      "finance",
        Title:       "Invoice List",
        Description: "Paginated, filterable list of sales invoices",
        ASTFn:       InvoiceListScreen,
    })
}
```

`RegisterPage` panics on duplicate routes. This is intentional — it surfaces configuration bugs at startup before any traffic is served.

At startup, after all `init()` functions run, call:

```go
if err := registry.ValidateRegistry(); err != nil {
    panic(err)
}
```

This validates that every registration has a non-empty `Route`, `Module`, `Title`, and at least one of `Fn` or `ASTFn`. The error code is `REGISTRY_VALIDATION_FAILED`.

---

## 6.8 Worked Example: Invoice Listing Page

Full end-to-end: define the screen, register it, test it.

### Screen Definition

```go
// internal/web/dsl/screens/invoice_list.go
package screens

import (
    "awo.so/internal/web/ast"
    "awo.so/internal/web/dsl/blocks"
    "awo.so/internal/web/registry"
    "awo.so/internal/web/ui"
)

func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/finance/invoices",
        Module: "finance",
        Title:  "Invoices",
        ASTFn:  InvoiceListScreen,
    })
}

func InvoiceListScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "Invoices",
        Body: []ast.Node{
            blocks.DataTableBlock(sess, blocks.DataTableConfig{
                APIURL:           "/api/v1/finance/invoices",
                PrimaryKey:       "id",
                PageSize:         25,
                AllowCreate:      true,
                CreateURL:        "/finance/invoices/new",
                CreatePermission: "finance.invoices.create",
                Columns: []blocks.ColumnDef{
                    {Name: "invoice_number", Label: "Invoice #", Sortable: true, Width: 150},
                    {Name: "vendor_name", Label: "Vendor"},
                    {Name: "amount", Label: "Amount", Type: "currency"},
                    {Name: "due_date", Label: "Due Date", Type: "date"},
                    blocks.StatusBadgeColumnDef(blocks.StatusBadgeConfig{
                        FieldName: "status",
                        Label:     "Status",
                        Mappings: []blocks.StatusMapping{
                            {Value: "DRAFT", Label: "Draft", Color: "default"},
                            {Value: "PENDING", Label: "Pending Approval", Color: "warning"},
                            {Value: "APPROVED", Label: "Approved", Color: "success"},
                            {Value: "PAID", Label: "Paid", Color: "success"},
                            {Value: "OVERDUE", Label: "Overdue", Color: "danger"},
                        },
                    }),
                },
                RowActions: []ast.ActionNode{
                    {Label: "View", ActionType: "link", Target: "/finance/invoices/${id}"},
                },
            }),
        },
    }
}
```

### What the Compiled Schema Looks Like

After `CompileTree`, the schema becomes:

```json
{
  "type": "page",
  "title": "Invoices",
  "body": [
    {
      "type": "crud",
      "api": {"method": "get", "url": "/api/v1/finance/invoices"},
      "syncLocation": false,
      "primaryField": "id",
      "perPage": 25,
      "columns": [
        {"name": "invoice_number", "label": "Invoice #", "sortable": true, "width": 150},
        {"name": "vendor_name", "label": "Vendor"},
        {"name": "amount", "label": "Amount", "type": "number"},
        {"name": "due_date", "label": "Due Date", "type": "date"},
        {
          "name": "status",
          "label": "Status",
          "type": "mapping",
          "map": {
            "DRAFT": "<span class='badge badge-default'>Draft</span>",
            "PENDING": "<span class='badge badge-warning'>Pending Approval</span>",
            "APPROVED": "<span class='badge badge-success'>Approved</span>",
            "PAID": "<span class='badge badge-success'>Paid</span>",
            "OVERDUE": "<span class='badge badge-danger'>Overdue</span>",
            "*": "<span class='badge badge-default'>${value}</span>"
          }
        }
      ],
      "toolbar": [
        {"type": "button", "label": "New", "actionType": "link",
         "to": "/finance/invoices/new", "level": "primary", "icon": "fa fa-plus"}
      ],
      "rowActions": [
        {"type": "button", "label": "View", "actionType": "link", "to": "/finance/invoices/${id}"}
      ]
    }
  ]
}
```

The "New" button is present because the user has the `finance.invoices.create` permission. For a user without that permission, `DataTableBlock` omits it from the toolbar before the schema is compiled.

---

*End of Chapter 06*

**Previous:** [Chapter 05 — SDUI Fundamentals](./05-sdui-fundamentals.md)
**Next:** [Chapter 07 — AST Design](./07-ast-design.md)
