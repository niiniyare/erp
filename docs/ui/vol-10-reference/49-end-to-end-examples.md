---
title: "49 – End-to-End Examples"
volume: "vol-10-reference"
chapter: 49
section: "Reference"
status: "implemented"
---

# Chapter 49 – End-to-End Examples

Three complete, working examples.  Each shows the full path from registration to
rendered JSON.

## Table of Contents
- [Example 1: Finance Dashboard (existing, fully working)](#example-1-finance-dashboard)
- [Example 2: Adding a New Listing Page](#example-2-adding-a-new-listing-page)
- [Example 3: Adding a Document Form (Invoice Pattern)](#example-3-adding-a-document-form-invoice-pattern)

---

## Example 1: Finance Dashboard

The Finance Dashboard is the canonical reference implementation.  It demonstrates
all conventions: `init()` registration, `ASTPageFn`, `InitAPI`, and block
composition.

### Registration (in init)

```go
// internal/web/dsl/screens/finance_dashboard.go
package screens

import (
    "github.com/your-org/erp/internal/web/dsl/ast"
    "github.com/your-org/erp/internal/web/dsl/blocks"
    "github.com/your-org/erp/internal/web/dsl/registry"
    ui "github.com/your-org/erp/internal/web/ui"
)

func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/finance/dashboard",
        Module: "finance",
        Title:  "Finance Dashboard",
        ASTFn:  FinanceDashboardScreen,
    })
}
```

### ASTPageFn implementation

```go
func FinanceDashboardScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "Finance Dashboard",
        InitAPI: &ast.APISpec{
            Method: "get",
            URL:    "/api/v1/finance/dashboard/summary",
        },
        Body: []ast.Node{
            blocks.KPIRowBlock(sess),
            blocks.RevenueChartBlock(sess),
            blocks.QuickActionsBlock(sess, []blocks.QuickAction{
                {Label: "New Invoice",    Permission: "finance.invoices.create"},
                {Label: "Record Payment", Permission: "finance.payments.create"},
            }),
            blocks.RecentInvoicesBlock(sess),
        },
    }
}
```

### Block: QuickActionsBlock (abbreviated)

```go
// internal/web/dsl/blocks/finance_quick_actions.go
type QuickAction struct {
    Label      string
    Permission string
}

func QuickActionsBlock(sess ui.UISessionContext, actions []QuickAction) ast.Node {
    var buttons []ast.ActionNode
    for _, a := range actions {
        if sess.HasPermission(a.Permission) {
            buttons = append(buttons, ast.ActionNode{
                Label:      a.Label,
                ActionType: "link",
            })
        }
    }
    if len(buttons) == 0 {
        return ast.NullNode{}
    }
    return ast.ToolbarNode{Buttons: buttons}
}
```

### Resulting AMIS JSON (abbreviated)

```json
{
  "type": "page",
  "title": "Finance Dashboard",
  "initApi": {
    "method": "get",
    "url": "/api/v1/finance/dashboard/summary"
  },
  "body": [
    {
      "type": "grid",
      "columns": [
        {"type": "stat", "label": "Revenue (KES)", "value": "${revenue}"},
        {"type": "stat", "label": "Expenses (KES)", "value": "${expenses}"},
        {"type": "stat", "label": "Net (KES)", "value": "${net}"}
      ]
    },
    {
      "type": "chart",
      "chartType": "bar",
      "data": "${revenue_by_month}",
      "style": {"height": "300px"}
    },
    {
      "type": "toolbar",
      "buttons": [
        {"type": "action", "label": "New Invoice", "actionType": "link"},
        {"type": "action", "label": "Record Payment", "actionType": "link"}
      ]
    },
    {
      "type": "crud",
      "api": "/api/v1/finance/invoices?recent=true",
      "columns": [
        {"name": "number", "label": "Invoice #"},
        {"name": "amount", "label": "Amount"},
        {"name": "status", "label": "Status", "type": "mapping", "map": {"PAID": "Paid", "PENDING": "Pending"}}
      ]
    }
  ]
}
```

---

## Example 2: Adding a New Listing Page

Walkthrough for adding a new listing page — Supplier List under the Procurement
module.

### Step 1: Define the screen config (optional typed config)

```go
// internal/web/dsl/screens/procurement_supplier_list.go
package screens
```

### Step 2: Register with RegisterPage in init

```go
func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/procurement/suppliers",
        Module: "procurement",
        Title:  "Suppliers",
        ASTFn:  SupplierListScreen,
    })
}
```

### Step 3: Implement the ASTPageFn

```go
func SupplierListScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "Suppliers",
        Body: []ast.Node{
            ast.CRUDNode{
                API: &ast.APISpec{
                    Method: "get",
                    URL:    "/api/v1/procurement/suppliers",
                },
                Columns: []ast.TableColumnNode{
                    {Name: "code",        Label: "Code"},
                    {Name: "name",        Label: "Supplier Name"},
                    {Name: "contact",     Label: "Contact"},
                    {Name: "currency",    Label: "Currency"},
                    {Name: "status",      Label: "Status",
                     Type: "mapping",
                     Map: map[string]string{
                         "ACTIVE":   "Active",
                         "INACTIVE": "Inactive",
                     }},
                },
                Filter: &ast.FilterNode{
                    Controls: []ast.FormControlNode{
                        {Name: "q",      Label: "Search", Type: "input-text",
                         Placeholder: "Search by name or code"},
                        {Name: "status", Label: "Status",  Type: "select",
                         Options: []ast.SelectOption{
                             {Label: "Active",   Value: "ACTIVE"},
                             {Label: "Inactive", Value: "INACTIVE"},
                         }},
                    },
                },
                HeaderToolbar: blocks.ListToolbar(sess, blocks.ListToolbarConfig{
                    CreatePermission: "procurement.suppliers.create",
                    CreateLabel:      "New Supplier",
                    CreateURL:        "/procurement/suppliers/new",
                    ExportPermission: "procurement.suppliers.export",
                }),
                RowActions: []ast.ActionNode{
                    {
                        Label:      "View",
                        ActionType: "link",
                        Link:       "/procurement/suppliers/${id}",
                    },
                },
            },
        },
    }
}
```

### Step 4: Verify registration

```go
// Test in register_test.go — ValidateRegistry catches bad registrations
func TestMain(m *testing.M) {
    if err := registry.ValidateRegistry(); err != nil {
        fmt.Fprintf(os.Stderr, "registry: %v\n", err)
        os.Exit(1)
    }
    os.Exit(m.Run())
}
```

The new route `/procurement/suppliers` is immediately available after the
package is imported by `main`.  No wiring changes are required elsewhere.

---

## Example 3: Adding a Document Form (Invoice Pattern)

An invoice creation form demonstrating permission gating inside blocks, a
multi-section form layout, and a submit action.

### Screen

```go
// internal/web/dsl/screens/finance_invoice_new.go
func init() {
    registry.RegisterPage(registry.PageRegistration{
        Route:  "/finance/invoices/new",
        Module: "finance",
        Title:  "New Invoice",
        ASTFn:  InvoiceNewScreen,
    })
}

func InvoiceNewScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "New Invoice",
        Body: []ast.Node{
            blocks.InvoiceFormBlock(sess),
        },
    }
}
```

### Block: InvoiceFormBlock

```go
// internal/web/dsl/blocks/invoice_form.go
func InvoiceFormBlock(sess ui.UISessionContext) ast.Node {
    // Permission gate — block owns this check
    if !sess.HasPermission("finance.invoices.create") {
        return ast.AlertNode{
            Level:   "warning",
            Message: "You do not have permission to create invoices.",
        }
    }

    return ast.FormNode{
        API: &ast.APISpec{
            Method: "post",
            URL:    "/api/v1/finance/invoices",
        },
        Controls: []ast.FormControlNode{
            // Header section
            {Name: "customer_id", Label: "Customer",    Type: "select",
             Source: "/api/v1/entities?type=customer",  Required: true},
            {Name: "inv_date",    Label: "Invoice Date", Type: "input-date", Required: true},
            {Name: "due_date",    Label: "Due Date",     Type: "input-date", Required: true},
            {Name: "currency",    Label: "Currency",     Type: "select",
             Value: sess.Currency,
             Options: []ast.SelectOption{
                 {Label: "KES", Value: "KES"},
                 {Label: "USD", Value: "USD"},
                 {Label: "EUR", Value: "EUR"},
             }},
            // Line items — composed from nested block
            invoiceLineItemsControl(sess),
            // Notes
            {Name: "notes", Label: "Notes", Type: "textarea"},
        },
        Actions: []ast.ActionNode{
            {
                Label:      "Save Draft",
                ActionType: "submit",
                Level:      "default",
                API: &ast.APISpec{
                    Method: "post",
                    URL:    "/api/v1/finance/invoices?status=DRAFT",
                },
            },
            {
                Label:      "Submit for Approval",
                ActionType: "submit",
                Level:      "primary",
                VisibleOn:  "${sess_can_submit}",
            },
            {
                Label:      "Cancel",
                ActionType: "link",
                Link:        "/finance/invoices",
                Level:      "link",
            },
        },
    }
}
```

### Permission gating inside blocks

The block above demonstrates two permission patterns:

1. **Hard gate** — return an `AlertNode` when the user has no create permission.
   The screen does not know about this; it always calls `InvoiceFormBlock`.
2. **Conditional action visibility** — the "Submit for Approval" button is
   gated via an AMIS expression `${sess_can_submit}`, a boolean injected into
   the page data by the `InitAPI` response based on the user's permissions.
   This avoids exposing permission keys in the browser while still gating the
   action at render time.

### Resulting AMIS JSON (abbreviated)

```json
{
  "type": "page",
  "title": "New Invoice",
  "body": [{
    "type": "form",
    "api": {"method": "post", "url": "/api/v1/finance/invoices"},
    "controls": [
      {"type": "select",     "name": "customer_id", "label": "Customer",
       "source": "/api/v1/entities?type=customer", "required": true},
      {"type": "input-date", "name": "inv_date",    "label": "Invoice Date", "required": true},
      {"type": "input-date", "name": "due_date",    "label": "Due Date",     "required": true},
      {"type": "select",     "name": "currency",    "label": "Currency",
       "value": "KES",
       "options": [{"label": "KES", "value": "KES"}, {"label": "USD", "value": "USD"}]},
      {"type": "textarea",   "name": "notes",       "label": "Notes"}
    ],
    "actions": [
      {"type": "action", "label": "Save Draft",           "actionType": "submit", "level": "default"},
      {"type": "action", "label": "Submit for Approval",  "actionType": "submit", "level": "primary",
       "visibleOn": "${sess_can_submit}"},
      {"type": "action", "label": "Cancel",               "actionType": "link",   "level": "link",
       "link": "/finance/invoices"}
    ]
  }]
}
```
