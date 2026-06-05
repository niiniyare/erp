# Chapter 11 — Forms Framework

> **Volume:** III — Component System
> **Audience:** Backend Engineers
> **Prerequisites:** Chapter 10 — Layout System

---

## Table of Contents

- [11.1 Form Architecture](#111-form-architecture)
- [11.2 FormNode — The Form Container](#112-formnode--the-form-container)
- [11.3 Field Types](#113-field-types)
- [11.4 Field Modifiers](#114-field-modifiers)
- [11.5 Form Actions](#115-form-actions)
- [11.6 FilterBarNode — CRUD Filters](#116-filterbarnde--crud-filters)
- [11.7 The Legacy amis.Form Builder](#117-the-legacy-amisform-builder)
- [11.8 Wizard Forms](#118-wizard-forms)
- [11.9 Form Patterns for ERP Workflows](#119-form-patterns-for-erp-workflows)

---

## 11.1 Form Architecture

Forms in AwoERP are defined entirely in the backend, as part of the page schema. The browser's AMIS SDK renders them using its built-in form components and handles:
- Client-side validation (based on `required`, `validations` fields)
- Form state management
- API submission
- Success/failure navigation

The backend defines:
- Which fields appear
- Their types, labels, and initial values
- Required/optional markers
- Validation rules
- The submit API endpoint
- Post-submit behavior (redirect, reload)
- Which actions are available (submit, save-draft, cancel)

Importantly, **form field visibility is controlled by the page function based on `UISessionContext`**. A field that a user cannot see is simply not included in the schema — there is no `visible: false` hiding mechanism.

---

## 11.2 FormNode — The Form Container

```go
type FormNode struct {
    API        *APISpec    // submit endpoint; nil for display-only forms
    InitAPI    *APISpec    // load initial values; nil if values come from page scope
    Mode       string      // "normal"|"horizontal"|"inline"
    Title      string      // panel title
    WrapPanel  bool        // wrap form in a panel (default true)
    Body       []Node      // field nodes
    Actions    []Node      // form action buttons
    Redirect   string      // URL to navigate after successful submit
    ReloadOn   string      // component name to reload after submit
}
```

Simple create form:

```go
ast.FormNode{
    API: &ast.APISpec{
        Method: "post",
        URL:    "/api/v1/finance/invoices",
    },
    Redirect: "/finance/invoices",  // navigate back to list on success
    Body: []ast.Node{
        ast.InputTextNode{Name: "invoice_number", Label: "Invoice #", Required: true},
        ast.SelectNode{Name: "vendor_id", Label: "Vendor", Required: true,
            Source: "/api/v1/vendors?fields=id,name"},
        ast.InputDateNode{Name: "due_date", Label: "Due Date"},
        ast.InputNumberNode{Name: "amount", Label: "Amount", Required: true, Min: 0},
    },
    Actions: []ast.Node{
        ast.ActionNode{Label: "Create Invoice", ActionType: "submit", Level: "primary"},
        ast.ActionNode{Label: "Cancel", ActionType: "link", Target: "/finance/invoices"},
    },
}
```

Edit form (loads existing data from API):

```go
ast.FormNode{
    InitAPI: &ast.APISpec{
        Method: "get",
        URL:    "/api/v1/finance/invoices/${id}",
    },
    API: &ast.APISpec{
        Method: "put",
        URL:    "/api/v1/finance/invoices/${id}",
    },
    Body: []ast.Node{
        ast.InputTextNode{Name: "invoice_number", Label: "Invoice #", Required: true},
        // ...
    },
    Actions: []ast.Node{
        ast.ActionNode{Label: "Update", ActionType: "submit", Level: "primary"},
    },
}
```

---

## 11.3 Field Types

### Text Fields

```go
// Single-line text
ast.InputTextNode{
    Name:        "invoice_number",
    Label:       "Invoice #",
    Required:    true,
    Placeholder: "e.g. INV-2026-001",
    MaxLength:   50,
}

// Multi-line textarea: use raw M (no typed node yet)
ui.M{
    "type":  "textarea",
    "name":  "notes",
    "label": "Notes",
    "minRows": 3,
}
```

### Number Fields

```go
ast.InputNumberNode{
    Name:     "quantity",
    Label:    "Quantity",
    Required: true,
    Min:      1,
    Max:      10000,
    Step:     1,
}

// Currency amount: use raw M for currency-specific formatting
ui.M{
    "type":       "input-number",
    "name":       "amount",
    "label":      "Amount",
    "precision":  2,
    "min":        0,
    "prefix":     "${currency}",  // from page scope
}
```

### Date Fields

```go
// Single date
ast.InputDateNode{
    Name:    "due_date",
    Label:   "Due Date",
    Format:  "YYYY-MM-DD",
    MinDate: "${today}",  // AMIS expression
}

// Date range
ast.InputDateRangeNode{
    Name:      "period",
    Label:     "Period",
    StartName: "start_date",
    EndName:   "end_date",
    Format:    "YYYY-MM-DD",
}
```

### Select Fields

```go
// Static options
ast.SelectNode{
    Name:     "status",
    Label:    "Status",
    Required: true,
    Options: []ast.SelectOption{
        {Label: "Draft", Value: "DRAFT"},
        {Label: "Pending", Value: "PENDING"},
        {Label: "Approved", Value: "APPROVED"},
    },
}

// Dynamic options from API
ast.SelectNode{
    Name:       "vendor_id",
    Label:      "Vendor",
    Required:   true,
    Source:     "/api/v1/vendors?fields=id,name",
    LabelField: "name",
    ValueField: "id",
    Searchable: true,
}

// Multi-select
ast.SelectNode{
    Name:       "category_ids",
    Label:      "Categories",
    Multiple:   true,
    Source:     "/api/v1/categories",
    LabelField: "name",
    ValueField: "id",
}
```

### Legacy Field Helpers (amis package)

The `amis` package provides functional field helpers for use in legacy `PageFn` schemas. These return raw `ui.M`:

```go
amis.TextField("invoice_number", "Invoice #")
amis.NumberField("amount", "Amount")
amis.DateField("due_date", "Due Date")
amis.DateRangeField("period", "Period")
amis.SelectField("status", "Status",
    amis.SelectOpt("Draft", "DRAFT"),
    amis.SelectOpt("Pending", "PENDING"),
)
amis.SelectAPIField("vendor_id", "Vendor", "/api/v1/vendors")
amis.SwitchField("is_recurring", "Recurring")
amis.TextAreaField("notes", "Notes")
```

---

## 11.4 Field Modifiers

### In Typed AST Nodes

Modifiers are struct fields on the node:

```go
ast.InputTextNode{
    Name:        "gl_code",
    Label:       "GL Code",
    Required:    true,
    ReadOnly:    false,
    Placeholder: "e.g. 4000-001",
    // VisibleOn and DisabledOn are AMIS expression strings
    VisibleOn:   "${account_type === 'EXPENSE'}",
    DisabledOn:  "${status !== 'DRAFT'}",
}
```

### In Legacy amis Package

The `amis` package provides modifier functions that mutate a field `M`:

```go
field := amis.TextField("gl_code", "GL Code")
field = amis.Required(field)
field = amis.Placeholder(field, "e.g. 4000-001")
field = amis.VisibleOn(field, "${account_type === 'EXPENSE'}")
field = amis.DisabledOn(field, "${status !== 'DRAFT'}")
field = amis.Desc(field, "Enter the GL account code for this expense")
field = amis.Validate(field, "regex:^[0-9]{4}-[0-9]{3}$")
```

Available modifiers:
- `Required(field M) M` — marks required
- `Optional(field M) M` — shows "(optional)" label remark
- `Placeholder(field M, text string) M`
- `Default(field M, val any) M` — sets default value
- `VisibleOn(field M, expr string) M` — conditional visibility (also sets `clearValueOnHidden: true`)
- `DisabledOn(field M, expr string) M`
- `Desc(field M, text string) M` — description below field
- `Validate(field M, rules ...string) M` — e.g., `"isEmail"`, `"minimum:0"`, `"regex:..."`

---

## 11.5 Form Actions

Form actions appear below the form fields (or in the form toolbar, depending on `wrapPanel` setting).

Standard action pattern:

```go
Actions: []ast.Node{
    // Primary action: submit the form
    ast.ActionNode{Label: "Create Invoice", ActionType: "submit", Level: "primary"},

    // Secondary: save without submitting
    ast.ActionNode{
        Label:      "Save Draft",
        ActionType: "ajax",
        Level:      "default",
        API:        &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/draft"},
    },

    // Tertiary: cancel (no confirmation needed for new forms)
    ast.ActionNode{Label: "Cancel", ActionType: "link", Target: "/finance/invoices"},
}
```

For edit forms where cancelling discards unsaved changes:

```go
ast.ActionNode{
    Label:       "Cancel",
    ActionType:  "link",
    Target:      "/finance/invoices",
    ConfirmText: "Discard unsaved changes?",
}
```

For conditional actions (based on workflow state):

```go
var actions []ast.Node
actions = append(actions, ast.ActionNode{Label: "Update", ActionType: "submit", Level: "primary"})

if sess.Can("approve", "finance.invoices") {
    actions = append(actions, ast.ActionNode{
        Label:      "Approve",
        ActionType: "ajax",
        Level:      "success",
        DisabledOn: "${status !== 'PENDING'}",
        API:        &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/${id}/approve"},
    })
}
```

---

## 11.6 FilterBarNode — CRUD Filters

`FilterBarNode` provides the filter form above a `CRUDNode`. The filter's field values are appended as query parameters to the CRUD's API URL.

```go
ast.CRUDNode{
    API:    ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices"},
    Filter: ast.FilterBarNode{
        Body: []ast.Node{
            ast.InputTextNode{Name: "invoice_number", Label: "Invoice #"},
            ast.SelectNode{
                Name:  "status",
                Label: "Status",
                Options: []ast.SelectOption{
                    {Label: "All", Value: ""},
                    {Label: "Pending", Value: "PENDING"},
                    {Label: "Approved", Value: "APPROVED"},
                },
            },
            ast.InputDateRangeNode{
                Name:      "date_range",
                Label:     "Date Range",
                StartName: "start_date",
                EndName:   "end_date",
            },
        },
    },
    Columns: []ast.TableColumn{...},
}
```

The AMIS SDK submits the filter form when the user clicks "Search" (AMIS renders this button automatically). The filtered values are appended to the CRUD API URL as query parameters, e.g., `/api/v1/finance/invoices?status=PENDING&start_date=2026-01-01&end_date=2026-06-30`.

---

## 11.7 The Legacy amis.Form Builder

The `amis.Form` builder is available for legacy `PageFn` pages.

```go
// Legacy pattern — PageFn with amis.Form builder
func InvoiceFormLegacy(sess ui.UISessionContext) ui.Schema {
    fields := []ui.M{
        amis.Required(amis.TextField("invoice_number", "Invoice #")),
        amis.Required(amis.SelectAPIField("vendor_id", "Vendor", "/api/v1/vendors")),
        amis.DateField("due_date", "Due Date"),
        amis.Required(amis.NumberField("amount", "Amount")),
    }

    if sess.Can("approve", "finance.invoices") {
        fields = append(fields, amis.SelectField("status", "Status",
            amis.SelectOpt("Draft", "DRAFT"),
            amis.SelectOpt("Pending Approval", "PENDING"),
        ))
    }

    return ui.M{
        "type": "page",
        "title": "New Invoice",
        "body": amis.Form("post:/api/v1/finance/invoices").
            Fields(fields...).
            Redirect("/finance/invoices").
            Build(),
    }
}
```

`amis.Form` returns `*FormBuilder` which implements `json.Marshaler`. `.Build()` returns the raw `M` for embedding.

---

## 11.8 Wizard Forms

Multi-step forms use `WizardBuilder` from the `amis` package. The typed AST does not have a `WizardNode` yet.

```go
// Wizard — amis builder (no typed AST equivalent yet)
func TenantOnboardingForm(sess ui.UISessionContext) ui.Schema {
    return ui.M{
        "type":  "page",
        "title": "New Tenant Setup",
        "body": amis.Wizard("post:/api/v1/platform/tenants").
            Step("Organization",
                amis.Required(amis.TextField("name", "Organization Name")),
                amis.Required(amis.TextField("domain", "Domain")),
            ).
            Step("Contact",
                amis.Required(amis.TextField("contact_email", "Admin Email")),
                amis.TextField("contact_phone", "Phone"),
            ).
            Step("Configuration",
                amis.SelectField("plan", "Plan",
                    amis.SelectOpt("Starter", "starter"),
                    amis.SelectOpt("Professional", "professional"),
                    amis.SelectOpt("Enterprise", "enterprise"),
                ),
                amis.SelectAPIField("currency", "Default Currency", "/api/v1/currencies"),
            ).
            ReviewStep(
                ui.M{"label": "Organization", "name": "name"},
                ui.M{"label": "Domain", "name": "domain"},
                ui.M{"label": "Admin Email", "name": "contact_email"},
                ui.M{"label": "Plan", "name": "plan"},
            ).
            Build(),
    }
}
```

---

## 11.9 Form Patterns for ERP Workflows

### Pattern: Create-then-Submit Workflow

Many ERP documents have a two-step flow: create as draft, then submit for approval.

```go
func NewInvoiceScreen(sess ui.UISessionContext) ast.Node {
    var submitAction ast.Node
    if sess.Can("submit", "finance.invoices") {
        submitAction = ast.ActionNode{
            Label:      "Submit for Approval",
            ActionType: "ajax",
            Level:      "primary",
            DisabledOn: "${status !== 'DRAFT'}",
            API:        &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/${id}/submit"},
            ConfirmText: "Submit this invoice for approval?",
        }
    }

    return ast.PageNode{
        Title:   "New Invoice",
        InitAPI: &ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/${id}"},
        Body: []ast.Node{
            ast.FormNode{
                API: &ast.APISpec{Method: "put", URL: "/api/v1/finance/invoices/${id}"},
                DisabledOn: "${status !== 'DRAFT'}", // form is read-only after submission
                Body: []ast.Node{
                    // ...fields...
                },
                Actions: func() []ast.Node {
                    actions := []ast.Node{
                        ast.ActionNode{Label: "Save Draft", ActionType: "submit", Level: "default",
                            DisabledOn: "${status !== 'DRAFT'}"},
                    }
                    if submitAction != nil {
                        actions = append(actions, submitAction)
                    }
                    return actions
                }(),
            },
        },
    }
}
```

### Pattern: Read-Only Detail with Conditional Actions

```go
func InvoiceDetailScreen(sess ui.UISessionContext) ast.Node {
    toolbar := []ast.Node{
        ast.ActionNode{Label: "Back", ActionType: "link", Target: "/finance/invoices"},
    }

    if sess.Can("approve", "finance.invoices") {
        toolbar = append(toolbar, ast.ActionNode{
            Label:      "Approve",
            ActionType: "dialog",
            Level:      "success",
            VisibleOn:  "${status === 'PENDING'}",
            Dialog: &ast.DialogNode{
                Title: "Approve Invoice",
                Body: []ast.Node{
                    ast.FormNode{
                        API: &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/${id}/approve"},
                        Body: []ast.Node{
                            ast.InputTextNode{Name: "note", Label: "Note", Required: false},
                        },
                    },
                },
            },
        })
    }

    return ast.PageNode{
        Title:   "Invoice Detail",
        InitAPI: &ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/${id}"},
        Toolbar: toolbar,
        Body: []ast.Node{
            ast.PropertyNode{
                Items: []ast.PropertyItem{
                    {Label: "Invoice #", Content: "${invoice_number}"},
                    {Label: "Vendor", Content: "${vendor_name}"},
                    {Label: "Amount", Content: "${amount}", Type: "currency"},
                    {Label: "Status", Content: "${status}"},
                },
            },
        },
    }
}
```

---

*End of Chapter 11*

**Previous:** [Chapter 10 — Layout System](./10-layout-system.md)
**Next:** [Chapter 12 — Tables and Data Grids](./12-tables-and-data-grids.md)
