---
title: amis Web UI Schemas
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, frontend-engineer]
related:
  - "[UI Schema Design](../17-ui-schema/01-ui-schema-overview.md)"
  - "[AMIS Schema Overview](../../../../05-frontend-engineering/02-amis-schemas/01-amis-schema-overview.md)"
  - "[List Patterns](../../../../05-frontend-engineering/02-amis-schemas/02-list-patterns.md)"
  - "[Detail Patterns](../../../../05-frontend-engineering/02-amis-schemas/04-detail-patterns.md)"
---

# amis Web UI Schemas

Every module needs four AMIS JSON schemas for its primary entity. This section covers the schema files for the Contracts module as the working example.

## Schema File Location

```
web/schemas/pages/
└── contracts/
    ├── list.json       ← CRUD list with filters and toolbar
    ├── detail.json     ← Read-only detail view
    ├── create.json     ← New contract form
    └── edit.json       ← Edit existing contract form
```

## list.json

```json
{
  "type": "page",
  "title": "Contracts",
  "body": {
    "type": "crud",
    "api": "/api/v1/contracts",
    "autoFetchSchema": false,
    "columns": [
      { "name": "contract_number", "label": "Contract No.", "width": 160 },
      { "name": "title", "label": "Title", "flex": 1 },
      {
        "name": "status",
        "label": "Status",
        "type": "mapping",
        "map": {
          "draft":        "<span class='label label-default'>Draft</span>",
          "submitted":    "<span class='label label-info'>Submitted</span>",
          "under_review": "<span class='label label-warning'>Under Review</span>",
          "approved":     "<span class='label label-primary'>Approved</span>",
          "active":       "<span class='label label-success'>Active</span>",
          "suspended":    "<span class='label label-warning'>Suspended</span>",
          "terminated":   "<span class='label label-danger'>Terminated</span>"
        }
      },
      { "name": "total_value", "label": "Value", "type": "number", "prefix": "$" },
      { "name": "end_date", "label": "Expires", "type": "date" },
      {
        "type": "operation",
        "label": "Actions",
        "buttons": [
          {
            "label": "View",
            "type": "button",
            "actionType": "link",
            "link": "/contracts/${id}"
          },
          {
            "label": "Delete",
            "type": "button",
            "actionType": "ajax",
            "confirmText": "Delete contract ${contract_number}?",
            "api": {
              "method": "DELETE",
              "url": "/api/v1/contracts/${id}"
            },
            "hiddenOn": "${status !== 'draft'}"
          }
        ]
      }
    ],
    "filter": {
      "title": "Filter",
      "body": [
        {
          "type": "select",
          "name": "status",
          "label": "Status",
          "clearable": true,
          "options": [
            { "label": "Draft",        "value": "draft" },
            { "label": "Submitted",    "value": "submitted" },
            { "label": "Under Review", "value": "under_review" },
            { "label": "Approved",     "value": "approved" },
            { "label": "Active",       "value": "active" },
            { "label": "Terminated",   "value": "terminated" }
          ]
        },
        {
          "type": "input-text",
          "name": "title",
          "label": "Title",
          "placeholder": "Search by title"
        }
      ]
    },
    "toolbar": [
      {
        "type": "button",
        "label": "New Contract",
        "level": "primary",
        "actionType": "link",
        "link": "/contracts/create"
      }
    ]
  }
}
```

## detail.json

```json
{
  "type": "page",
  "title": "Contract Detail",
  "initApi": "/api/v1/contracts/${id}",
  "toolbar": [
    {
      "type": "button",
      "label": "Submit for Review",
      "level": "primary",
      "visibleOn": "${status === 'draft'}",
      "actionType": "ajax",
      "api": {
        "method": "POST",
        "url": "/api/v1/contracts/${id}/submit"
      },
      "confirmText": "Submit contract ${contract_number} for review?"
    },
    {
      "type": "button",
      "label": "Approve",
      "level": "success",
      "visibleOn": "${status === 'under_review'}",
      "actionType": "dialog",
      "dialog": {
        "title": "Approve Contract",
        "body": {
          "type": "form",
          "api": {
            "method": "POST",
            "url": "/api/v1/contracts/${id}/approve"
          },
          "body": [
            {
              "type": "textarea",
              "name": "comment",
              "label": "Comments",
              "required": true,
              "placeholder": "Provide approval comments..."
            }
          ]
        }
      }
    }
  ],
  "body": {
    "type": "panel",
    "title": "Contract Information",
    "body": {
      "type": "property",
      "column": 3,
      "items": [
        { "label": "Contract No.",  "content": "${contract_number}" },
        { "label": "Title",         "content": "${title}", "span": 2 },
        { "label": "Status",        "content": "${status}" },
        { "label": "Type",          "content": "${contract_type}" },
        { "label": "Total Value",   "content": "$${total_value}" },
        { "label": "Currency",      "content": "${currency}" },
        { "label": "Start Date",    "content": "${start_date | date:YYYY-MM-DD}" },
        { "label": "End Date",      "content": "${end_date | date:YYYY-MM-DD}" },
        { "label": "Version",       "content": "${version}" }
      ]
    }
  }
}
```

## create.json

```json
{
  "type": "page",
  "title": "New Contract",
  "body": {
    "type": "form",
    "api": {
      "method": "POST",
      "url": "/api/v1/contracts"
    },
    "redirect": "/contracts/${id}",
    "body": [
      {
        "type": "group",
        "body": [
          {
            "type": "input-text",
            "name": "contract_number",
            "label": "Contract No.",
            "required": true,
            "maxLength": 50
          },
          {
            "type": "select",
            "name": "contract_type",
            "label": "Type",
            "required": true,
            "options": [
              { "label": "Service",   "value": "service" },
              { "label": "Supply",    "value": "goods" },
              { "label": "Lease",     "value": "lease" },
              { "label": "Other",     "value": "other" }
            ]
          }
        ]
      },
      {
        "type": "input-text",
        "name": "title",
        "label": "Title",
        "required": true,
        "maxLength": 255
      },
      {
        "type": "group",
        "body": [
          {
            "type": "input-number",
            "name": "total_value",
            "label": "Total Value",
            "required": true,
            "min": 0,
            "precision": 2
          },
          {
            "type": "select",
            "name": "currency",
            "label": "Currency",
            "required": true,
            "value": "USD",
            "options": [
              { "label": "USD", "value": "USD" },
              { "label": "EUR", "value": "EUR" },
              { "label": "GBP", "value": "GBP" }
            ]
          }
        ]
      },
      {
        "type": "group",
        "body": [
          {
            "type": "input-date",
            "name": "start_date",
            "label": "Start Date",
            "required": true,
            "format": "YYYY-MM-DD"
          },
          {
            "type": "input-date",
            "name": "end_date",
            "label": "End Date",
            "required": true,
            "format": "YYYY-MM-DD"
          }
        ]
      },
      {
        "type": "textarea",
        "name": "description",
        "label": "Description",
        "maxLength": 2000
      }
    ],
    "actions": [
      {
        "type": "submit",
        "label": "Create Contract",
        "level": "primary"
      },
      {
        "type": "button",
        "label": "Cancel",
        "actionType": "link",
        "link": "/contracts"
      }
    ]
  }
}
```

## edit.json

The edit schema is identical to create.json but:
- Uses `PUT /api/v1/contracts/${id}`
- Includes `initApi: "/api/v1/contracts/${id}"` to pre-fill current values
- Includes a hidden `version` field for optimistic locking

```json
{
  "type": "page",
  "title": "Edit Contract",
  "initApi": "/api/v1/contracts/${id}",
  "body": {
    "type": "form",
    "api": {
      "method": "PUT",
      "url": "/api/v1/contracts/${id}"
    },
    "redirect": "/contracts/${id}",
    "body": [
      { "type": "hidden", "name": "version" },
      "...same fields as create.json..."
    ]
  }
}
```

**The `version` hidden field is mandatory in edit forms** — the API rejects updates without a version match (optimistic locking).

## Schema Checklist

For every module entity, ensure:

- [ ] `list.json` — CRUD with filter and toolbar "Create" button
- [ ] `detail.json` — property panel with all fields, action buttons gated by `visibleOn`
- [ ] `create.json` — form posting to `POST /api/v1/{module}`
- [ ] `edit.json` — form with `initApi` and `PUT` endpoint, hidden `version` field
- [ ] Status `mapping` in list uses colored labels, not raw strings
- [ ] Monetary values display with currency prefix
- [ ] Date fields formatted as `YYYY-MM-DD`
- [ ] Delete/destructive actions have `confirmText`
- [ ] Form redirects to detail page after create/edit success

## Connecting Schemas to Navigation

Register the module in the sidebar schema (`web/schemas/app.json`):

```json
{
  "type": "nav",
  "links": [
    {
      "label": "Contracts",
      "icon": "fa fa-file-contract",
      "children": [
        { "label": "All Contracts", "to": "/contracts" },
        { "label": "New Contract",  "to": "/contracts/create" }
      ]
    }
  ]
}
```

Page routes in the app schema map URL patterns to schema files:

```json
{
  "pages": [
    { "url": "/contracts",          "schema": "/schemas/pages/contracts/list.json" },
    { "url": "/contracts/create",   "schema": "/schemas/pages/contracts/create.json" },
    { "url": "/contracts/:id",      "schema": "/schemas/pages/contracts/detail.json" },
    { "url": "/contracts/:id/edit", "schema": "/schemas/pages/contracts/edit.json" }
  ]
}
```
