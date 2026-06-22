---
title: "Chapter 23: Composite Blocks Reference"
part: "Part IV — The SDUI Layer"
chapter: 23
section: "23-composite-blocks"
related:
  - "[Chapter 22: Foundation Components](22-foundation-components.md)"
  - "[Chapter 21: SDUI Philosophy](21-sdui-philosophy.md)"
---

# Chapter 23: Composite Blocks Reference

Composite blocks are pre-assembled amis schema patterns for common ERP view types. Using them ensures visual consistency, reduces page builder code, and makes future amis version upgrades easier (update the composite block, not each page builder).

---

## 23.1. `AmisCRUD` — The Full List + Form Block

`AmisCRUD` combines a data table with create/edit forms in a single block. It is the most frequently used composite block — virtually every master data entity uses it.

### 23.1.1. When to Use AmisCRUD

Use `AmisCRUD` for master data management: Customers, Products, Vendors, Cost Centres. Do not use it for transactional documents (Invoices, Purchase Orders) where the form is complex and the detail view needs a master-detail layout — use `AmisDetailPage` for those.

### 23.1.2. Configuration

```go
func BuildCustomerListPage(ctx PageBuilderContext) *amis.Schema {
    crud := amis.NewCRUD().
        API("GET /api/v1/customers").
        CreateAPI("POST /api/v1/customers").
        UpdateAPI("PUT /api/v1/customers/${id}").
        DeleteAPI("DELETE /api/v1/customers/${id}").
        PrimaryField("id").
        DefaultParams(map[string]interface{}{"sort": "-created_at"})

    // Columns
    crud.AddColumn(amis.Column("name").Label("Customer Name").Searchable(true).Sortable(true))
    crud.AddColumn(amis.Column("email").Label("Email").Width(200))
    crud.AddColumn(amis.Column("phone").Label("Phone").Width(150))
    crud.AddColumn(amis.Column("status").Label("Status").Type("tag").Width(100).
        Map(map[string]string{"active": "success", "inactive": "warning"}))

    // Create/Edit form
    crud.FormFields(
        amis.TextField("name").Label("Customer Name").Required(true),
        amis.TextField("email").Label("Email").Validator("email"),
        amis.TextField("phone").Label("Phone").Validator("phone"),
        amis.SelectField("status").Label("Status").
            Options([]amis.Option{{Value: "active", Label: "Active"}, {Value: "inactive", Label: "Inactive"}}).
            Default("active"),
    )

    // Toolbar
    if ctx.Permissions.CanCreate("Customer") {
        crud.AddToolbar(amis.NewCreateButton("New Customer"))
    }

    return amis.NewPage().Title("Customers").Body(crud).Schema()
}
```

### 23.1.3. Inline Editing vs Modal Form vs Separate Page

| Option | Use case | amis config |
|---|---|---|
| Modal form | Standard — most entities | `.Mode("dialog")` (default) |
| Inline editing | Simple table data (tags, flags, single fields) | `.Mode("inline")` |
| Separate page | Complex forms, wizard flows | `.RowAction("link", "/customers/${id}/edit")` |

### 23.1.4. Toolbar Configuration

```go
crud.AddToolbar(amis.NewCreateButton("New Customer"))
crud.AddToolbar(amis.ExportButton("Export CSV").API("GET /api/v1/customers/export"))
if ctx.Permissions.CanDelete("Customer") {
    crud.AddBulkAction(amis.BulkDeleteAction())
}
```

---

## 23.2. `AmisForm` — Standalone Form

Used for entity creation/edit pages where the form is too complex for `AmisCRUD`'s inline modal.

### 23.2.1. Sections and Tabs

Use **sections** (fieldsets) when all fields are related and visible simultaneously:

```go
amis.NewForm().
    Section("Basic Information",
        amis.TextField("name").Required(true),
        amis.TextField("kra_pin").Label("KRA PIN"),
    ).
    Section("Contact Details",
        amis.TextField("email"),
        amis.TextField("phone"),
        amis.TextAreaField("address").Rows(3),
    ).
    SubmitAPI("POST /api/v1/customers")
```

Use **tabs** when fields are grouped into distinct concerns and seeing all at once would be overwhelming:

```go
amis.NewForm().
    Tab("General", generalFields...).
    Tab("Financial", financialFields...).
    Tab("Custom Fields", customFields...).
    SubmitAPI("PUT /api/v1/customers/${id}")
```

### 23.2.2. Conditional Field Visibility

amis expression syntax for conditional visibility:

```go
// Show 'insurer_name' only when 'requires_insurance' is true
amis.TextField("insurer_name").
    Label("Insurer Name").
    VisibleOn("${requires_insurance === true}")

// Show 'rejection_reason' only when status is 'rejected'
amis.TextAreaField("rejection_reason").
    Label("Rejection Reason").
    VisibleOn("${status === 'rejected'}").
    Required(true)
```

### 23.2.3. Field Dependencies

React to another field's change:

```go
// When customer_id changes, reload the credit_limit field from the customer API
amis.NumberField("credit_available").
    Label("Available Credit").
    ReadOnly(true).
    Source("GET /api/v1/customers/${customer_id}/credit").
    AutoLoad(true).
    ReloadOn("${customer_id}")
```

### 23.2.4. Submit Behaviour

```go
form.SubmitAPI("POST /api/v1/invoices").
    SuccessMessage("Invoice created successfully").
    SuccessRedirect("/invoices/${id}").
    ErrorMessage("Failed to create invoice — please check the fields above")
```

---

## 23.3. `AmisList` — Read-Only Table

For views where data is displayed but not edited inline.

```go
amis.NewList().
    API("GET /api/v1/gl-entries").
    Columns(
        amis.Column("entry_date").Label("Date").Type("date"),
        amis.Column("account_code").Label("Account"),
        amis.Column("debit").Label("Debit (KES)").Type("number").Precision(2).ThousandSeparator(true),
        amis.Column("credit").Label("Credit (KES)").Type("number").Precision(2).ThousandSeparator(true),
        amis.Column("description").Label("Description"),
    ).
    DefaultSort("-entry_date").
    InlineFilter(
        amis.DateRangeField("date_range").StartName("date_from").EndName("date_to"),
        amis.SelectField("account_id").Source("GET /api/v1/chart-of-accounts"),
    )
```

### 23.3.2. Row-Level Actions

```go
list.RowActions(
    amis.RowAction("View").Type("link").Href("/gl-entries/${id}"),
    amis.RowAction("Reverse").Type("ajax").
        API("POST /api/v1/gl-entries/${id}/reverse").
        ConfirmText("Reverse this GL entry?").
        HideOn("${reversed_at !== null}"),
)
```

---

## 23.4. `AmisDetailPage` — Master + Related Tables

For transactional document detail views (Invoice, Purchase Order, Sales Order).

```go
func BuildInvoiceDetailPage(ctx PageBuilderContext) *amis.Schema {
    detail := amis.NewDetailPage().
        API("GET /api/v1/invoices/${id}").
        Header(
            amis.Column("invoice_number").Label("Invoice #").Type("text"),
            amis.Column("status").Label("Status").Type("tag"),
            amis.Column("total_amount").Label("Total").Type("currency"),
            amis.Column("due_date").Label("Due Date").Type("date"),
        )

    // Action toolbar
    toolbar := amis.NewActionToolbar()
    if ctx.Permissions.CanSubmit("Invoice") {
        toolbar.Add(amis.ActionButton("Submit").
            API("POST /api/v1/invoices/${id}/submit").
            VisibleOn("${status === 'draft'}").
            ConfirmText("Submit this invoice? It cannot be edited after submission."))
    }
    if ctx.Permissions.CanCancel("Invoice") {
        toolbar.Add(amis.ActionButton("Cancel").
            API("POST /api/v1/invoices/${id}/cancel").
            VisibleOn("${status === 'submitted'}").
            Level("danger"))
    }
    detail.Toolbar(toolbar)

    // Related tables as tabs
    detail.Tab("Line Items", amis.NewList().
        API("GET /api/v1/invoice-items?invoice_id=${id}").
        Columns(lineItemColumns()...))

    detail.Tab("Payments", amis.NewList().
        API("GET /api/v1/payments?invoice_id=${id}").
        Columns(paymentColumns()...))

    detail.Tab("Activity", amis.NewList().
        API("GET /api/v1/audit-log?entity=Invoice&entity_id=${id}").
        Columns(auditLogColumns()...))

    return amis.NewPage().Title("Invoice Detail").Body(detail).Schema()
}
```

---

## 23.5. `AmisDashboard` — Stat Cards and Charts

```go
func BuildFinanceDashboard(ctx PageBuilderContext) *amis.Schema {
    dashboard := amis.NewDashboard()

    // Date range filter — applies to all blocks
    dashboard.GlobalFilter(
        amis.DateRangeField("period").
            StartName("date_from").EndName("date_to").
            Default("thisMonth"),
    )

    // KPI stat cards
    dashboard.AddCard(amis.StatCard().
        API("GET /api/v1/finance/kpis?date_from=${date_from}&date_to=${date_to}").
        Label("Revenue").Field("total_revenue").
        Prefix("KES ").ThousandSeparator(true).
        Trend("revenue_trend"))

    dashboard.AddCard(amis.StatCard().
        API("GET /api/v1/finance/kpis").
        Label("Outstanding Invoices").Field("outstanding_count").
        ColorOn("${outstanding_count > 50}", "warning"))

    // Charts
    dashboard.AddChart(amis.BarChart().
        API("GET /api/v1/finance/revenue-by-month?year=${year}").
        XField("month").
        Series(amis.Series("revenue", "Revenue (KES)")))

    return amis.NewPage().Title("Finance Dashboard").Body(dashboard).Schema()
}
```

---

## 23.6. `AmisKanban` — Status Column Board

```go
amis.NewKanban().
    API("GET /api/v1/support-tickets").
    GroupBy("status").
    Columns([]amis.KanbanColumn{
        {Value: "open", Label: "Open", Color: "blue"},
        {Value: "in_progress", Label: "In Progress", Color: "orange"},
        {Value: "resolved", Label: "Resolved", Color: "green"},
        {Value: "closed", Label: "Closed", Color: "gray"},
    }).
    CardFields(
        amis.CardField("title"),
        amis.CardField("assignee_name").Icon("user"),
        amis.CardField("priority").Type("tag"),
    ).
    DragUpdate("PATCH /api/v1/support-tickets/${id}").
    DragPatchField("status")
```

Dragging a card from "Open" to "In Progress" triggers `PATCH /api/v1/support-tickets/{id}` with `{ "status": "in_progress" }`. The full lifecycle runs: `before_save` hook, status transition validation, `after_save` notifications.

---

## 23.7. `AmisWizard` — Multi-Step Form

```go
amis.NewWizard().
    Step("Company Details",
        amis.TextField("company_name").Required(true),
        amis.TextField("kra_pin").Required(true),
        amis.SelectField("company_size").Options(companySizeOptions()),
    ).
    Step("Contact Information",
        amis.TextField("contact_name").Required(true),
        amis.TextField("contact_email").Required(true).Validator("email"),
        amis.TextField("contact_phone").Required(true).Validator("phone"),
    ).
    Step("Plan Selection",
        amis.RadioGroupField("plan").Options(planOptions()).Required(true),
    ).
    SubmitAPI("POST /api/v1/tenant-onboarding").
    SuccessMessage("Account created! Check your email to set your password.").
    SuccessRedirect("/login")
```

Each step has its own validation — the user cannot proceed to the next step until the current step's required fields are valid. Data from all steps is merged and submitted in the final step.

---

## 23.8. `AmisReportPage` — Parameterised Report

```go
amis.NewReportPage().
    Title("Aged Receivables Report").
    Parameters(
        amis.DateField("as_of_date").Label("As of Date").Default("today").Required(true),
        amis.SelectField("branch_id").Label("Branch").
            Source("GET /api/v1/branches").Optional(true),
        amis.SelectField("currency").Default("KES"),
    ).
    ReportAPI("GET /api/v1/reports/aged-receivables").
    Columns(
        amis.Column("customer_name").Label("Customer"),
        amis.Column("current").Label("Current (KES)").Type("number"),
        amis.Column("days_30").Label("1-30 Days").Type("number"),
        amis.Column("days_60").Label("31-60 Days").Type("number"),
        amis.Column("days_90").Label("61-90 Days").Type("number"),
        amis.Column("over_90").Label("Over 90 Days").Type("number").Color("danger"),
        amis.Column("total").Label("Total Outstanding").Type("number").Bold(true),
    ).
    SummaryRow(amis.SummaryRow().TotalsFor("current", "days_30", "days_60", "days_90", "over_90", "total")).
    Export(amis.ExportOptions().CSV().Excel().FileName("aged-receivables-${as_of_date}"))
```
