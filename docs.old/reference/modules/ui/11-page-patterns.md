[<-- Back to Index](README.md)

## Page Patterns

Standard schema structures used repeatedly across all Awo ERP modules. Reuse these instead of inventing new layouts.

### Pattern 1 — Standard CRUD List

Every module's primary list view (purchase orders, employees, invoices, etc.):

```markdown
page
  └── crud
        ├── filter (form with status/date/keyword)
        ├── headerToolbar [search-box, bulkActions, columns-toggler, reload]
        ├── columns [..., operation column with buttons]
        └── footerToolbar [statistics, pagination]
```

Example: `web/schemas/pages/purchase-orders.json`

```json
{
  "type": "page",
  "title": "Purchase Orders",
  "toolbar": [
    { "type": "button", "label": "New Order", "level": "primary",
      "actionType": "link", "link": "#purchase-orders-new" }
  ],
  "body": {
    "type": "crud",
    "id":   "po-list",
    "api":  "get:/api/v1/purchase-orders",
    "syncLocation": true,
    "headerToolbar": ["search-box", "bulkActions",
      {"type":"columns-toggler"}, {"type":"reload"}],
    "filter": {
      "body": [
        {"type":"input-text",       "name":"keywords",  "label":"Search"},
        {"type":"select",           "name":"status",    "label":"Status", "clearable":true,
         "options":[{"label":"Draft","value":"draft"},{"label":"Confirmed","value":"confirmed"}]},
        {"type":"input-date-range", "name":"date_range","label":"Date Range"}
      ]
    },
    "columns": [
      {"name":"reference",     "label":"Reference",  "sortable":true},
      {"name":"supplier_name", "label":"Supplier",   "sortable":true},
      {"name":"total",         "label":"Total",      "type":"number", "prefix":"KES ", "sortable":true},
      {"name":"status",        "label":"Status",     "type":"tag",
       "colorMap":{"draft":"default","submitted":"processing","confirmed":"success","cancelled":"error"}},
      {"name":"created_at",    "label":"Date",       "type":"date", "sortable":true},
      {"type":"operation",     "label":"Actions",
       "buttons":[
         {"type":"button","label":"View",   "level":"link","actionType":"link","link":"#po-detail?id=${id}"},
         {"type":"button","label":"Submit", "level":"link","visibleOn":"${status==='draft'}",
          "actionType":"ajax","api":"post:/api/v1/purchase-orders/${id}/submit",
          "confirmText":"Submit this order for approval?"}
       ]}
    ]
  }
}
```

### Pattern 2 — Record Detail Page

Full information on one entity with tabs for related data:

```markdown
page
  ├── toolbar [action buttons — vary by status]
  └── tabs
        ├── "Overview"   → descriptions
        ├── "Line Items" → crud (no filter, paginated)
        ├── "Documents"  → crud (file attachments)
        ├── "Approvals"  → service + timeline
        └── "Activity"   → service + timeline (audit log)
```

Action buttons in toolbar should use `visibleOn` based on status:

```json
"toolbar": [
  { "type":"button","label":"Submit for Approval","level":"primary",
    "visibleOn":"${status==='draft'}",
    "actionType":"ajax","api":"post:/api/v1/purchase-orders/${id}/submit" },
  { "type":"button","label":"Cancel Order","level":"danger",
    "visibleOn":"${status!=='cancelled' && status!=='completed'}",
    "actionType":"ajax","api":"post:/api/v1/purchase-orders/${id}/cancel",
    "confirmText":"Cancel this order? This cannot be undone." }
]
```

### Pattern 3 — Dashboard

Situational awareness — what needs attention right now:

```json
{
  "type": "page",
  "title": "Finance Dashboard",
  "body": {
    "type": "grid",
    "columns": [
      {
        "columnClassName": "col-sm-4",
        "body": {
          "type": "service",
          "api": "get:/api/v1/analytics/receivables/summary",
          "body": [
            {"type":"stat","source":"${total}","label":"Total Receivables","prefix":"KES "},
            {"type":"stat","source":"${overdue}","label":"Overdue","prefix":"KES ","className":"text-danger"}
          ]
        }
      },
      {
        "columnClassName": "col-sm-8",
        "body": {
          "type": "chart",
          "api":  "get:/api/v1/analytics/revenue/monthly",
          "config": {
            "backgroundColor": "transparent",
            "xAxis":  {"type":"category","data":"${months}"},
            "yAxis":  {"type":"value"},
            "series": [{"type":"bar","data":"${values}","name":"Revenue"}]
          }
        }
      }
    ]
  }
}
```

Dashboard rules (from UX spec):
- Maximum 6 widgets
- Every widget has a direct action button — no "go to the list to find it"
- No raw tables — tables belong in the List surface
- All data scoped to the current user automatically

### Pattern 4 — Approval Inbox

Pending approvals page for any user with approval responsibilities:

```json
{
  "type": "page",
  "title": "Approvals Inbox",
  "body": {
    "type": "crud",
    "id":   "approvals-inbox",
    "api":  "get:/api/v1/approvals/pending",
    "columns": [
      {"name":"record_type",  "label":"Type"},
      {"name":"record_ref",   "label":"Reference"},
      {"name":"requested_by", "label":"Requested By"},
      {"name":"requested_at", "label":"Date", "type":"date"},
      {
        "type":"operation","label":"Actions",
        "buttons":[
          {
            "label":"View Details","type":"button","level":"link",
            "actionType":"drawer",
            "drawer":{
              "title":"Approve ${record_type} #${record_ref}",
              "body":{
                "type":"service",
                "schemaApi":"get:/schema/approvals/${record_type}/${record_id}"
              }
            }
          },
          {
            "label":"Approve","type":"button","level":"success",
            "actionType":"ajax",
            "api":"post:/api/v1/approvals/${request_id}/approve",
            "onEvent":{"success":{"actions":[
              {"actionType":"reload","componentId":"approvals-inbox"},
              {"actionType":"toast","args":{"msg":"Approved","msgType":"success"}}
            ]}}
          },
          {
            "label":"Reject","type":"button","level":"danger",
            "actionType":"ajax",
            "api":"post:/api/v1/approvals/${request_id}/reject",
            "confirmText":"Reject this request?",
            "onEvent":{"success":{"actions":[
              {"actionType":"reload","componentId":"approvals-inbox"}
            ]}}
          }
        ]
      }
    ]
  }
}
```

### Pattern 5 — Business Rules Builder

The most complex UI. Uses `condition-builder` and `crud` with draggable rows:

```markdown
page
  └── crud (rules list — draggable for priority ordering)
        ├── column: name
        ├── column: active (switch — saves immediately)
        ├── column: conditions (condition-builder in drawer)
        ├── column: actions (combo)
        └── column: operations → [Edit drawer] [Test dialog] [History drawer]
```

See the full Go implementation in `awo/web/schema/settings/rules.go`.

### Pattern 6 — Settings Page

```markdown
page
  └── tabs
        ├── "General"      → form (tenant config key/values)
        ├── "Rules"        → crud (rule sets) + nested rule editor
        └── "Integrations" → crud (webhooks/API keys)
```

---
