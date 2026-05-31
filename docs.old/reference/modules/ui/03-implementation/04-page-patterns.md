# Page Patterns

> Last verified: 2026-05-18 | Code pointer: `internal/web/dsl/blocks/`, `internal/web/ast/`

Three recurring page shapes cover ~90% of ERP pages. Pick the right pattern before writing any code.

---

## Pattern 1: List (CRUD)

**When to use:** Any paginated list of records. Invoices, users, products, tenants.

**Key components:** `DataTableBlock` (or `ast.CRUDNode` directly), `StatusBadgeColumn`, toolbar with create button.

### Minimal example

```go
func Schema(sess ui.UISessionContext) any {
    return blocks.DataTableBlock(sess, blocks.DataTableConfig{
        APIURL:      "/api/v1/finance/invoices",
        Title:       "Invoices",
        PrimaryKey:  "id",
        PageSize:    20,
        AllowCreate: sess.Can("create", "invoice"),
        CreateURL:   "#invoices/new",
        AllowExport: sess.Can("export", "invoice"),
        Columns: []blocks.ColumnDef{
            {Name: "number",   Label: "Invoice #", Sortable: true},
            {Name: "supplier", Label: "Supplier"},
            {Name: "amount",   Label: "Amount",    Type: "currency"},
            {Name: "status",   Label: "Status",    Type: "status"},
            {Name: "due_date", Label: "Due",        Type: "date", Sortable: true},
        },
        RowActions: []ast.ActionNode{
            {Label: "View", ActionType: "link", Target: "#invoices/${id}"},
        },
    })
}
```

### With KPI row above the table

```go
func Schema(sess ui.UISessionContext) any {
    return ast.PageNode{
        Title: "Invoices",
        InitAPI: ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/summary"},
        Body: ast.GridNode{
            Columns: []ast.GridColumn{
                {MD: 12, Body: []ast.Node{
                    blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
                        {Label: "Total Outstanding", ValueKey: "outstanding", Format: "currency"},
                        {Label: "Overdue",           ValueKey: "overdue_count", Format: "number"},
                        {Label: "This Month",        ValueKey: "month_total",  Format: "currency"},
                    }),
                }},
                {MD: 12, Body: []ast.Node{
                    blocks.DataTableBlock(sess, blocks.DataTableConfig{
                        APIURL:     "/api/v1/finance/invoices",
                        // ...
                    }),
                }},
            },
        },
    }
}
```

### Backend requirements

```go
// List endpoint
GET /api/v1/finance/invoices?offset=0&limit=20
→ { "success": true, "data": [...], "meta": { "pagination": { "total_records": 142 } } }

// Summary endpoint (for KPI row)
GET /api/v1/finance/invoices/summary
→ { "success": true, "data": { "outstanding": 48200.00, "overdue_count": 3, "month_total": 12500.00 } }
```

### Common mistakes

- Forgetting `PrimaryKey` → bulk actions and row selection break
- Using raw `ast.CRUDNode` without `syncLocation: true` → URL params don't persist (set automatically when using `DataTableBlock`)
- Creating a new `CreateURL` route that's not registered → 404

---

## Pattern 2: Document Form (Create / Edit)

**When to use:** Single-record create or edit with line items. Invoices, purchase orders, bills, journal entries.

**Key components:** `DocumentHeaderBlock`, `ProductServiceLineBlock`, `TotalsSummaryBlock`, optional `AddressBlock`, `ApprovalWorkflowBlock`, `InternalNotesBlock`.

### Full document form example

```go
func Schema(sess ui.UISessionContext) any {
    isEdit    := false // or detect from route param
    readOnly  := isEdit && !sess.Can("update", "invoice")

    return ast.PageNode{
        Title:   "New Invoice",
        SubmitAPI: ast.APISpec{Method: "post", URL: "/api/v1/finance/invoices"},
        Redirect: "#invoices/${id}", // navigate to view page after save
        Data: ui.M{
            "can_approve": sess.Can("approve", "invoice"),
        },
        Body: ast.GridNode{
            Columns: []ast.GridColumn{
                {MD: 8, Body: []ast.Node{
                    blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{
                        ShowCurrency: true,
                        ShowStatus:   true,
                        StatusOptions: []ast.SelectOption{
                            {Label: "Draft", Value: "draft"},
                            {Label: "Sent",  Value: "sent"},
                        },
                        ReadOnly: readOnly,
                    }),
                    blocks.ProductServiceLineBlock(sess, blocks.DefaultLineItemConfig()),
                    blocks.InternalNotesBlock(sess),
                }},
                {MD: 4, Body: []ast.Node{
                    blocks.TotalsSummaryBlock(sess),
                    blocks.AddressBlock(sess, blocks.AddressConfig{
                        ShowBilling:  true,
                        ShowShipping: true,
                        ReadOnly:     readOnly,
                    }),
                    blocks.ApprovalWorkflowBlock(sess),
                }},
            },
        },
    }
}
```

### Edit form (load existing record)

```go
return ast.PageNode{
    Title:   "Edit Invoice #${ref_number}",
    InitAPI: ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices/${id}"},   // ← load existing
    SubmitAPI: ast.APISpec{Method: "put", URL: "/api/v1/finance/invoices/${id}"}, // ← update
    // ...
}
```

The `InitAPI` populates the form's data scope. AMIS merges the response `data` object into the page scope — all fields map by name.

### Backend requirements

```go
// Create
POST /api/v1/finance/invoices
Body: { "ref_number": "...", "document_date": "...", "line_items": [...], ... }
→ { "status": 0, "msg": "Invoice created", "data": { "id": "uuid" } }

// Edit load
GET /api/v1/finance/invoices/{id}
→ { "status": 0, "data": { "id": "uuid", "ref_number": "INV-001", "line_items": [...], ... } }

// Edit save
PUT /api/v1/finance/invoices/{id}
→ { "status": 0, "msg": "Invoice updated" }
```

### Common mistakes

- `SubmitAPI` and `InitAPI` using same method: `InitAPI` must be `GET`
- `line_items` field name changed in schema — backend always expects `line_items`
- `Redirect` pointing to a route not registered → 404 after save

---

## Pattern 3: Settings / Config Form

**When to use:** Single-record settings — no line items, no document header. Tenant settings, user profile, module config.

**Key components:** `ast.FormNode` or raw `ast.SectionNode` groups, standard field nodes.

### Example

```go
func Schema(sess ui.UISessionContext) any {
    canEdit := sess.Can("update", "tenant.settings")

    return ast.PageNode{
        Title:    "Tenant Settings",
        InitAPI:  ast.APISpec{Method: "get", URL: "/api/v1/platform/tenants/${tenant_id}/settings"},
        SubmitAPI: ast.APISpec{Method: "put", URL: "/api/v1/platform/tenants/${tenant_id}/settings"},
        Body: ast.SectionNode{
            Title: "General",
            Body: []ast.Node{
                ast.InputTextNode{Name: "display_name", Label: "Display Name", Required: true, DisabledOn: boolExprSess(canEdit)},
                ast.SelectNode{
                    Name:  "timezone",
                    Label: "Timezone",
                    Source: &ast.APISpec{Method: "get", URL: "/api/v1/platform/timezones/options"},
                },
                ast.SelectNode{
                    Name:  "default_currency",
                    Label: "Default Currency",
                    Source: &ast.APISpec{Method: "get", URL: "/api/v1/platform/currencies/options"},
                },
            },
        },
    }
}

func boolExprSess(allowed bool) string {
    if allowed { return "" }
    return "true"
}
```

### With tabbed sections

```go
Body: ast.TabsNode{
    Tabs: []ast.TabItem{
        {
            Title: "General",
            Body:  generalSection(sess),
        },
        {
            Title: "Billing",
            Body:  billingSection(sess),
        },
        {
            Title: "Integrations",
            Body:  integrationsSection(sess),
            VisibleOn: "${is_platform}", // only platform admins
        },
    },
},
```

---

## Pattern 4: Dashboard

**When to use:** Overview pages with KPIs, charts, quick actions. Not a form, not a list.

**Key components:** `KPIRowBlock`, `StatCardBlock`, `QuickActionsBlock`, `ast.ChartNode`.

### Example

```go
func Schema(sess ui.UISessionContext) any {
    return ast.PageNode{
        Title:   "Dashboard",
        InitAPI: ast.APISpec{Method: "get", URL: "/api/v1/dashboard/summary"},
        Body: ast.GridNode{
            Columns: []ast.GridColumn{
                // Row 1: KPIs
                {MD: 12, Body: []ast.Node{
                    blocks.KPIRowBlock(sess, []blocks.StatCardConfig{
                        {Label: "Revenue (MTD)", ValueKey: "revenue_mtd",   Format: "currency", Trend: blocks.TrendUp, TrendKey: "revenue_trend"},
                        {Label: "Open Invoices", ValueKey: "open_invoices", Format: "number"},
                        {Label: "Payables Due",  ValueKey: "payables_due",  Format: "currency", Trend: blocks.TrendDown},
                    }),
                }},
                // Row 2: Chart + Quick Actions
                {MD: 8, Body: []ast.Node{
                    ast.ChartNode{
                        // transparent background enforced automatically
                        API: ast.APISpec{Method: "get", URL: "/api/v1/dashboard/revenue-chart"},
                    },
                }},
                {MD: 4, Body: []ast.Node{
                    blocks.QuickActionsBlock(sess, []blocks.QuickAction{
                        {Label: "New Invoice", URL: "#invoices/new", Permission: "invoice.create", Icon: "fa fa-file-invoice"},
                        {Label: "New Bill",    URL: "#bills/new",    Permission: "bill.create",    Icon: "fa fa-file-alt"},
                        {Label: "Reports",     URL: "#reports",      Permission: "",               Icon: "fa fa-chart-bar"},
                    }),
                }},
            },
        },
    }
}
```

### Backend requirements

```go
GET /api/v1/dashboard/summary
→ {
    "status": 0,
    "data": {
        "revenue_mtd":   84200.00,
        "revenue_trend": 12.4,       // percent change — drives trend arrow
        "open_invoices": 7,
        "payables_due":  22100.00
    }
}
```

---

## Choosing the Right Pattern

| Page type | Pattern | Key block |
|-----------|---------|-----------|
| Record list (invoices, users) | List (CRUD) | `DataTableBlock` |
| Create document with line items | Document Form | `ProductServiceLineBlock` |
| Edit existing document | Document Form + `InitAPI` | `DocumentHeaderBlock` |
| Settings / profile | Settings Form | `ast.SectionNode` groups |
| KPI overview | Dashboard | `KPIRowBlock` + `ast.ChartNode` |
| Nested list inside detail view | List (CRUD) inside grid column | `DataTableBlock` with scoped API |

---

## Permissions Checklist for Any Pattern

1. **Expose as AMIS variable in `Data:`** — for element-level visibility (`visibleOn: "${can_edit}"`)
2. **Gate structurally in Go** — for entire sections or blocks only admin roles should see
3. **Add to `AllUIPermissions`** — every permission used in `sess.Can()` must be registered or it always returns false

```go
Data: ui.M{
    "can_create": sess.Can("create", "invoice"),  // → visibleOn / disabledOn in schema
    "can_edit":   sess.Can("update", "invoice"),
    "can_delete": sess.Can("delete", "invoice"),
    "is_platform": sess.IsPlatform,               // mode gate
},
```

See [IAM Integration](../02-architecture/03-iam-integration.md) for AllUIPermissions registration.
