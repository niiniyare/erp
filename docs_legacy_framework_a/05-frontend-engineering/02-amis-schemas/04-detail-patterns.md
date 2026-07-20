> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Detail View Patterns
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer, full-stack-engineer]
related:
  - "[CRUD Patterns](02-crud-patterns.md)"
  - "[Form Patterns](03-form-patterns.md)"
  - "[AMIS Schema Overview](01-amis-schema-overview.md)"
---

# Detail View Patterns

Detail (show) screens display a single record with full field set, audit info, and related sub-data.

## Basic Detail Page

```json
{
  "type": "page",
  "title": "Contract Detail",
  "toolbar": [
    {
      "type": "button",
      "label": "Edit",
      "icon": "fa fa-edit",
      "actionType": "link",
      "link": "/contracts/${id}/edit",
      "visibleOn": "${status === 'draft'}"
    },
    {
      "type": "button",
      "label": "Submit for Review",
      "level": "primary",
      "actionType": "ajax",
      "api": "POST /api/v1/contracts/${id}/submit",
      "confirmText": "Submit this contract for review?",
      "visibleOn": "${status === 'draft'}",
      "onEvent": {
        "success": { "actions": [{ "actionType": "reload", "componentId": "contract-detail" }] }
      }
    },
    {
      "type": "button",
      "label": "Back",
      "actionType": "link",
      "link": "/contracts"
    }
  ],
  "body": {
    "type": "service",
    "id": "contract-detail",
    "api": "GET /api/v1/contracts/${id}",
    "body": [
      {
        "type": "panel",
        "title": "Contract Information",
        "body": {
          "type": "property",
          "items": [
            { "label": "Contract Number", "content": "${contract_number}" },
            { "label": "Title", "content": "${title}" },
            { "label": "Status", "content": {
              "type": "tag",
              "label": "${status | upperFirst}",
              "color": "${status === 'active' ? 'success' : status === 'rejected' ? 'danger' : status === 'approved' ? 'info' : 'default'}"
            }},
            { "label": "Total Value", "content": "${currency} ${total_value}" },
            { "label": "Start Date", "content": "${start_date}" },
            { "label": "End Date", "content": "${end_date}" },
            { "label": "Vendor", "content": "${vendor_name}" },
            { "label": "Assigned To", "content": "${assigned_to_name}" }
          ]
        }
      }
    ]
  }
}
```

## Two-Column Layout

```json
{
  "type": "grid",
  "columns": [
    {
      "md": 8,
      "body": [
        {
          "type": "panel",
          "title": "Details",
          "body": {
            "type": "property",
            "column": 2,
            "items": [
              { "label": "Contract Number", "content": "${contract_number}" },
              { "label": "Type", "content": "${contract_type | upperFirst}" },
              { "label": "Start Date", "content": "${start_date}" },
              { "label": "End Date", "content": "${end_date}" },
              { "label": "Currency", "content": "${currency}" },
              { "label": "Total Value", "content": "${total_value}" }
            ]
          }
        },
        {
          "type": "panel",
          "title": "Description",
          "body": "${description | default('No description provided.')}"
        }
      ]
    },
    {
      "md": 4,
      "body": [
        {
          "type": "panel",
          "title": "Status & Activity",
          "body": {
            "type": "property",
            "column": 1,
            "items": [
              { "label": "Status", "content": "${status | upperFirst}" },
              { "label": "Version", "content": "${version}" },
              { "label": "Created By", "content": "${created_by_name}" },
              { "label": "Created At", "content": "${created_at | date:'YYYY-MM-DD HH:mm'}" },
              { "label": "Updated At", "content": "${updated_at | date:'YYYY-MM-DD HH:mm'}" }
            ]
          }
        }
      ]
    }
  ]
}
```

## Tabbed Detail with Sub-Lists

For records with related child data (contract lines, audit trail):

```json
{
  "type": "tabs",
  "tabs": [
    {
      "title": "Overview",
      "body": {
        "type": "property",
        "column": 2,
        "items": [
          { "label": "Contract Number", "content": "${contract_number}" },
          { "label": "Status", "content": "${status | upperFirst}" },
          { "label": "Total Value", "content": "${currency} ${total_value}" },
          { "label": "Vendor", "content": "${vendor_name}" }
        ]
      }
    },
    {
      "title": "Line Items",
      "body": {
        "type": "crud",
        "api": "GET /api/v1/contracts/${id}/lines",
        "columns": [
          { "name": "description", "label": "Description" },
          { "name": "quantity", "label": "Qty", "width": 80 },
          { "name": "unit_price", "label": "Unit Price", "width": 120 },
          {
            "name": "total_price",
            "label": "Total",
            "width": 120,
            "tpl": "${currency} ${total_price}"
          }
        ],
        "footerToolbar": ["statistics"]
      }
    },
    {
      "title": "Audit Trail",
      "body": {
        "type": "crud",
        "api": "GET /api/v1/contracts/${id}/audit",
        "syncLocation": false,
        "columns": [
          {
            "name": "created_at",
            "label": "Time",
            "width": 160,
            "tpl": "${created_at | date:'YYYY-MM-DD HH:mm'}"
          },
          { "name": "actor_name", "label": "User", "width": 140 },
          { "name": "action", "label": "Action" },
          {
            "name": "details",
            "label": "Details",
            "type": "json",
            "levelExpand": 1
          }
        ]
      }
    }
  ]
}
```

## Status Timeline

Show lifecycle progress as a visual step indicator:

```json
{
  "type": "steps",
  "steps": [
    {
      "title": "Draft",
      "subTitle": "${status_history.draft_at | date:'YYYY-MM-DD' | default:'-'}"
    },
    {
      "title": "Under Review",
      "subTitle": "${status_history.submitted_at | date:'YYYY-MM-DD' | default:'-'}"
    },
    {
      "title": "Approved",
      "subTitle": "${status_history.approved_at | date:'YYYY-MM-DD' | default:'-'}"
    },
    {
      "title": "Active",
      "subTitle": "${status_history.activated_at | date:'YYYY-MM-DD' | default:'-'}"
    }
  ],
  "value": "${['draft','submitted','under_review','approved','active'].indexOf(status)}"
}
```

## Conditional Action Buttons

Map each status to the allowed actions:

```json
{
  "type": "button-group",
  "buttons": [
    {
      "label": "Submit",
      "level": "primary",
      "actionType": "ajax",
      "api": "POST /api/v1/contracts/${id}/submit",
      "confirmText": "Submit for review?",
      "visibleOn": "${status === 'draft'}"
    },
    {
      "label": "Approve",
      "level": "success",
      "actionType": "ajax",
      "api": "POST /api/v1/contracts/${id}/approve",
      "confirmText": "Approve this contract?",
      "visibleOn": "${status === 'under_review'}"
    },
    {
      "label": "Reject",
      "level": "danger",
      "actionType": "dialog",
      "dialog": {
        "title": "Reject Contract",
        "body": {
          "type": "form",
          "api": "POST /api/v1/contracts/${id}/reject",
          "body": [
            {
              "type": "textarea",
              "name": "reason",
              "label": "Reason",
              "required": true,
              "placeholder": "Explain why this contract is being rejected"
            }
          ]
        }
      },
      "visibleOn": "${status === 'under_review'}"
    },
    {
      "label": "Terminate",
      "level": "danger",
      "actionType": "ajax",
      "api": "POST /api/v1/contracts/${id}/terminate",
      "confirmText": "Permanently terminate this contract? This cannot be undone.",
      "visibleOn": "${status === 'active'}"
    }
  ]
}
```

## Alert Banners for State

Show contextual warnings based on record state:

```json
{
  "type": "each",
  "value": "${[
    status === 'rejected' ? {type:'warning', body:'Contract was rejected. Reason: ' + rejection_reason} : null,
    days_until_expiry !== null && days_until_expiry <= 30 ? {type:'danger', body:'Contract expires in ' + days_until_expiry + ' days.'} : null
  ] | filter}",
  "items": {
    "type": "alert",
    "alertType": "${type}",
    "body": "${body}"
  }
}
```

## Detail Page Checklist

- Load data via `service` component — not page-level `initApi` (allows refresh without page reload)
- Put action buttons in page `toolbar`, not inside panels
- Use `property` for field display — not a `form` in read-only mode
- Use `visibleOn` to show/hide action buttons based on status — never disable them
- Audit trail tab: `syncLocation: false` so tab pagination doesn't change URL params
