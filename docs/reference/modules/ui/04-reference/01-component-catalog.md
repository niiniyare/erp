# Component Catalog

> Last verified: 2026-05-18 | Code pointer: `internal/web/amis/`, `internal/web/ast/`

Quick-reference for all Go UI components. Two APIs exist: **legacy `amis.*` builders** (returns `map[string]any`) and **modern `ast.*` nodes** (typed, compiled by `ast.CompileTree`). Prefer `ast.*` for new code.

---

## Layout Components

### Page (`amis.Page` / `ast.PageNode`)

Root container for every page schema. SchemaHandler wraps the compiled page in `{status:0, data:{...}}`.

**Legacy:**
```go
amis.Page("Invoice List").
    Data(amis.M{"can_create": sess.Can("create", "invoice")}).
    Body(crudSchema).
    Build()
```

**Modern (`ast.PageNode`):**
```go
ast.PageNode{
    Title:    "Invoice List",
    InitAPI:  ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/${id}"},  // optional: load initial data
    SubmitAPI: ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices"},      // optional: submit on save
    Redirect:  "#invoices/${id}",  // navigate after successful submit
    Data: ui.M{
        "can_create": sess.Can("create", "invoice"),
    },
    Body: someNode,
}
```

| Field | Purpose |
|---|---|
| `Title` | Page heading |
| `InitAPI` | GET call at mount — populates data scope |
| `SubmitAPI` | POST/PUT endpoint for form-style pages |
| `Redirect` | URL to navigate after submit success (supports `${field}` tokens) |
| `Data` | Initial data scope (permissions, flags, etc.) |
| `Body` | Primary content node |
| `Aside` | Right aside panel |
| `Toolbar` | Top toolbar items |

---

### Grid (`amis.Grid` / `ast.GridNode`)

12-column responsive grid. Each column takes `md` = 1–12.

**Legacy:**
```go
amis.Grid(
    amis.Col(8, mainContent),
    amis.Col(4, sidePanel),
)
```

**Modern:**
```go
ast.GridNode{
    Columns: []ast.GridColumn{
        {MD: 8, Body: []ast.Node{mainContent}},
        {MD: 4, Body: []ast.Node{sidePanel}},
    },
}
```

Common column widths: 12 (full), 8+4, 6+6, 4+4+4, 3+3+3+3.

---

### Panel (`amis.Panel`)

Card with header and optional footer.

```go
amis.Panel("Totals").
    Body(totalsTable).
    Footer(submitButton).
    Build()
```

No `ast.PanelNode` yet — use `ast.CardNode` from dsl blocks, or raw `amis.Panel`.

---

### Tabs (`amis.Tabs` / `ast.TabsNode`)

**Legacy:**
```go
amis.Tabs().
    Mode("line").     // "line"|"card"|"radio"|"vertical"
    Tab("General", generalFields).
    Tab("Billing", billingFields).
    Build()
```

**Modern:**
```go
ast.TabsNode{
    Tabs: []ast.TabItem{
        {Title: "General", Body: generalSection},
        {Title: "Billing", Body: billingSection, VisibleOn: "${can_view_billing}"},
    },
}
```

---

### Service (`amis.Service`)

Loads data from an API and makes it available to children via the AMIS data chain. Use when you need to load data that is not part of the page's primary API call.

```go
amis.Service("get:/api/v1/dashboard/summary").
    Body(
        kpiRow,
        chartGrid,
    ).
    Build()
```

No modern `ast.ServiceNode` yet. Use `amis.Service` for multi-source pages.

---

### Flex (`ast.FlexNode`)

Flexible row/column layout without 12-column constraints.

```go
ast.FlexNode{
    Direction: "row",   // "row"|"column"
    Gap:       "sm",    // "xs"|"sm"|"md"|"lg"
    Wrap:      true,
    Items:     []ast.Node{btn1, btn2, btn3},
}
```

Used by `QuickActionsBlock`. No `amis.Flex` equivalent — use `amis.Grid` or raw `map[string]any`.

---

## Data Display Components

### CRUD (`amis.CRUD` / `ast.CRUDNode`)

Paginated list/table. Always sets `syncLocation: true` automatically.

**Legacy:**
```go
amis.CRUD("get:/api/v1/finance/invoices").
    Columns(
        amis.Column("number", "Invoice #").Sortable(),
        amis.Column("status", "Status").Type("status"),
        amis.Column("amount", "Amount").Type("currency").Align("right"),
        amis.Column("actions", "").Buttons(
            amis.EditBtn("put:/api/v1/finance/invoices/${id}", fields...),
            amis.DeleteBtn("delete:/api/v1/finance/invoices/${id}"),
        ),
    ).
    Toolbar(amis.CreateBtn("New Invoice", "post:/api/v1/finance/invoices", fields...)).
    PerPage(20).
    DefaultSort("created_at", "desc").
    Build()
```

**Modern (`ast.CRUDNode`):**
```go
ast.CRUDNode{
    API:        ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices"},
    PrimaryKey: "id",
    PageSize:   20,
    Columns: []ast.TableColumn{
        {Name: "number",   Label: "Invoice #", Sortable: true},
        {Name: "status",   Label: "Status",    Type: "status"},
        {Name: "amount",   Label: "Amount",    Type: "currency"},
    },
    Toolbar: []ast.Node{
        ast.ActionNode{Label: "New", ActionType: "link", Target: "#invoices/new", Level: "primary"},
    },
    RowActions: []ast.ActionNode{
        {Label: "View", ActionType: "link", Target: "#invoices/${id}"},
    },
    BulkActions: []ast.Node{...},
    Filter:      filterNode,
}
```

**Column types:**

| Type | Display |
|---|---|
| `"text"` (default) | Plain string |
| `"date"` | Formatted date |
| `"datetime"` | Formatted date+time |
| `"number"` | Locale-formatted number |
| `"currency"` | Currency with symbol |
| `"status"` | Status badge (AMIS mapping) |
| `"mapping"` | Value → label lookup |
| `"tag"` | Colored tag |
| `"image"` | Thumbnail |
| `"link"` | Anchor |
| `"tpl"` | Custom template |
| `"operation"` | Action buttons column |

**Column modifiers (legacy):**
```go
amis.Column("status", "Status").
    Type("mapping").
    Map(amis.M{"active": "Active", "suspended": "Suspended"}).
    Width(120).
    Align("center").
    Fixed("right").          // pin to right edge
    VisibleOn("${is_admin}") // conditional column
```

---

### Descriptions (`amis.Descriptions`)

Read-only key-value display. Use for record detail views.

```go
amis.Descriptions("Invoice Details").
    Item("Invoice #", "number").
    Item("Supplier",  "supplier_name").
    Item("Amount",    "amount").
    Columns(2). // 2-column layout
    Build()
```

---

### Timeline (`amis.Timeline`)

Activity feed / audit log.

```go
amis.Timeline("get:/api/v1/finance/invoices/${id}/history").Build()

// Static items:
amis.Timeline("").Items(
    amis.M{"time": "2026-05-01", "title": "Created",  "detail": "Draft created"},
    amis.M{"time": "2026-05-02", "title": "Approved", "detail": "Approved by Alice"},
).Build()
```

---

### Chart (`amis.Chart` / `ast.ChartNode`)

ECharts visualization. **Always** transparent background.

**Legacy:**
```go
amis.Chart("get:/api/v1/dashboard/revenue-chart").
    Height(300).
    Build()
```

**Modern:**
```go
ast.ChartNode{
    API:    ast.APISpec{Method: "get", URL: "/api/v1/dashboard/revenue-chart"},
    Height: 300,
    // backgroundColor: transparent enforced automatically
}
```

Chart endpoints return ECharts `option` JSON directly:
```json
{ "status": 0, "data": { "xAxis": {...}, "series": [...] } }
```

---

### Alert (`amis.Alert`)

Inline info/warning/error banner.

```go
amis.Alert("warning", "This invoice is overdue.").
    VisibleOn("${is_overdue}").
    Build()
```

Levels: `"info"` | `"success"` | `"warning"` | `"danger"`

---

### Stat / KPI Card (`amis.Stat`)

KPI number display (legacy inline version):

```go
amis.Stat("Total Revenue", "${total_revenue}").Build()
```

For modern KPI cards with trend, use `blocks.StatCardBlock` or `blocks.KPIRowBlock`.

---

### Tpl (Template)

Inline HTML template with AMIS expression support:

```go
amis.Tpl(`<span class="badge badge-${status_class}">${status_label}</span>`)
```

Use CSS variables for colours. No hardcoded hex.

---

## Form Components

### Form (`amis.Form` / `ast.FormNode`)

**Legacy:**
```go
amis.Form("post:/api/v1/finance/invoices").
    Title("New Invoice").
    Mode("normal").
    Fields(
        amis.Required(amis.TextField("ref_number", "Reference #")),
        amis.DateField("document_date", "Date"),
        amis.SelectAPIField("supplier_id", "Supplier", "get:/api/v1/procurement/suppliers/options"),
    ).
    Redirect("#invoices/${id}").
    Build()
```

Form modes: `"normal"` (default, single column) | `"horizontal"` | `"inline"`

---

### Wizard (`amis.Wizard`)

Multi-step form. **Always end with a review step.**

```go
amis.Wizard("post:/api/v1/finance/invoices").
    Step("Basic Info",
        amis.Required(amis.TextField("ref_number", "Reference #")),
        amis.DateField("document_date", "Date"),
    ).
    Step("Supplier",
        amis.Required(amis.SelectAPIField("supplier_id", "Supplier", "get:/api/v1/.../options")),
        amis.AddressBlock(sess, ...), // blocks compose here
    ).
    ReviewStep(
        amis.M{"label": "Reference", "name": "ref_number"},
        amis.M{"label": "Supplier",  "name": "supplier_name"},
    ).
    Build()
```

---

## Field Components

### Text Fields

| Function | AMIS type | Notes |
|---|---|---|
| `amis.TextField(name, label)` | `input-text` | Standard single-line text |
| `amis.TextAreaField(name, label)` | `textarea` | Multi-line text |
| `amis.HiddenField(name)` | `hidden` | Carries data, no display |

**Modern equivalents:**
```go
ast.InputTextNode{Name: "ref_number", Label: "Reference #", Required: true, MaxLength: 100}
```

---

### Numeric Fields

| Function | AMIS type | Notes |
|---|---|---|
| `amis.NumberField(name, label)` | `input-number` | Integer default |

**Modern:**
```go
ast.InputNumberNode{Name: "qty", Label: "Qty", Required: true, Precision: 0, Min: 0}
ast.InputNumberNode{Name: "amount", Label: "Amount", Precision: 2} // currency-scale
```

---

### Date Fields

| Function | AMIS type |
|---|---|
| `amis.DateField(name, label)` | `input-date` |
| `amis.DateTimeField(name, label)` | `input-datetime` |
| `amis.DateRangeField(name, label)` | `input-date-range` |

**Modern:**
```go
ast.InputDateNode{Name: "document_date", Label: "Date", Required: true, Format: "YYYY-MM-DD"}
```

---

### Select / Dropdown

```go
// Static options (legacy)
amis.SelectField("status", "Status",
    amis.SelectOpt("Active",    "active"),
    amis.SelectOpt("Suspended", "suspended"),
)

// API-sourced options (legacy)
amis.SelectAPIField("supplier_id", "Supplier", "get:/api/v1/procurement/suppliers/options")

// Modern — static
ast.SelectNode{
    Name: "status",
    Label: "Status",
    Options: []ast.SelectOption{
        {Label: "Active",    Value: "active"},
        {Label: "Suspended", Value: "suspended"},
    },
}

// Modern — API-sourced
ast.SelectNode{
    Name:       "supplier_id",
    Label:      "Supplier",
    Source:     &ast.APISpec{Method: "get", URL: "/api/v1/procurement/suppliers/options"},
    Searchable: true,
}
```

Select options API format:
```json
{ "status": 0, "data": [{"label": "Supplier A", "value": "uuid-1"}, ...] }
```

---

### Boolean Fields

| Function | AMIS type | Notes |
|---|---|---|
| `amis.SwitchField(name, label)` | `switch` | Toggle — value: `true`/`false` |
| `amis.CheckboxField(name, label)` | `checkbox` | Single checkbox |
| `amis.RadioField(name, label, opts...)` | `radios` | Exclusive choice |

---

### File / Image Upload

```go
amis.FileField("attachment", "Attachment", "post:/api/v1/files/upload")
amis.ImageField("logo", "Logo", "post:/api/v1/files/upload")
```

Upload API must return:
```json
{ "status": 0, "data": { "value": "https://..." } }
```

---

## Field Modifiers (legacy only)

Apply to any field `M`:

```go
field := amis.TextField("email", "Email")
field = amis.Required(field)                     // marks required
field = amis.Validate(field, "isEmail")          // validation rule
field = amis.Placeholder(field, "user@example.com")
field = amis.Default(field, "user@example.com")
field = amis.Desc(field, "Business email only")  // description below field
field = amis.VisibleOn(field, "${show_email}")    // conditional + clearValueOnHidden
field = amis.DisabledOn(field, "${read_only}")   // conditional disabled
field = amis.Optional(field)                     // adds "(optional)" label remark
```

**Validation rules:** `"isEmail"` | `"isUrl"` | `"isInt"` | `"minimum:N"` | `"maximum:N"` | `"minLength:N"` | `"maxLength:N"` | `"regex:pattern"`

**Modern equivalents (on AST nodes):** Most modifiers are fields on the node struct directly:

```go
ast.InputTextNode{
    Name:        "email",
    Label:       "Email",
    Required:    true,
    Placeholder: "user@example.com",
    MaxLength:   254,
    DisabledOn:  "${read_only}",
    VisibleOn:   "${show_email}",
}
```

---

## Action Components

### Action Button (`ast.ActionNode` / raw `M`)

```go
// Modern
ast.ActionNode{
    Label:       "Approve",
    ActionType:  "ajax",
    API:         &ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices/${id}/approve"},
    Level:       "success",   // "default"|"primary"|"success"|"warning"|"danger"
    Icon:        "fa fa-check",
    ConfirmText: "Approve this invoice?",
    VisibleOn:   "${can_approve}",
}

// Link action
ast.ActionNode{
    Label:      "View",
    ActionType: "link",
    Target:     "#invoices/${id}",
}
```

### Row Action Shortcuts (legacy, `amis/crud.go`)

```go
amis.CreateBtn("New Invoice", "post:/api/v1/finance/invoices", fields...)  // toolbar button → dialog
amis.EditBtn("put:/api/v1/finance/invoices/${id}", fields...)               // row edit → dialog
amis.ViewBtn(detailSchema)                                                  // row view → drawer
amis.DeleteBtn("delete:/api/v1/finance/invoices/${id}")                     // row delete with confirm
```

---

## Utility Components

### Breadcrumb

```go
amis.Breadcrumb(
    amis.BC("Finance",  "#finance"),
    amis.BC("Invoices", "#invoices"),
    amis.BC("INV-001"), // current page — no href
)
```

### Dividers

```go
amis.Divider()   // form divider
amis.HDivider()  // page-level horizontal rule
amis.Section("Address Details") // section header inside a form
```

### Section (`ast.SectionNode`)

Collapsible section container:

```go
ast.SectionNode{
    Title:     "Internal Notes",
    Collapsed: true,   // collapsed by default
    Body:      []ast.Node{notesField},
}
```

---

## Choosing: `amis.*` vs `ast.*`

| Situation | Use |
|---|---|
| New page, standard components | `ast.*` nodes |
| New page, component not in ast package | `amis.*` builder + raw `map[string]any` |
| Migrating existing legacy page | `amis.*` until migrated, then switch to `ast.*` |
| Composing inside a DSL block | `ast.*` nodes only (blocks return `ast.Node`) |
| Quick one-off admin page | Either — choose what's faster |

See [Migration Guide](../appendices/B-migration-guide.md) for migrating `amis.*` → `ast.*` pages.

---

## AMIS Expression Reference

Expressions used in `visibleOn`, `disabledOn`, `requiredOn`:

| Expression | Meaning |
|---|---|
| `"${can_create}"` | True when `can_create` is truthy in data scope |
| `"!${can_create}"` | Negation |
| `"${status === 'draft'}"` | Equality check |
| `"${amount > 0}"` | Numeric comparison |
| `"${can_edit && status !== 'paid'}"` | Logical AND |
| `"${role === 'admin' \|\| is_platform}"` | Logical OR |
| `"true"` | Always (used for read-only `disabledOn`) |

Always expose boolean permission values in `Data:` first:
```go
Data: ui.M{
    "can_create": sess.Can("create", "invoice"),
    "can_approve": sess.Can("approve", "invoice"),
}
```
Then reference as `"${can_create}"` in expressions.
