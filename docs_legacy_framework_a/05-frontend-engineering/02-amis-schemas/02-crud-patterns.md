> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: CRUD Patterns
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer, full-stack-engineer]
related:
  - "[AMIS Schema Overview](01-amis-schema-overview.md)"
  - "[Form Patterns](03-form-patterns.md)"
  - "[Contracts API](../../08-api-reference/03-contracts-api.md)"
---

# CRUD Patterns

The `crud` component handles the list/search/paginate/act pattern. This page shows the standard patterns used across all AwoERP list screens.

## Standard List Screen

```json
{
  "type": "page",
  "title": "Contracts",
  "body": {
    "type": "crud",
    "syncLocation": false,
    "api": "/api/v1/contracts",
    "defaultParams": {
      "page_size": 20
    },
    "headerToolbar": [
      {
        "type": "button",
        "label": "New Contract",
        "icon": "fa fa-plus",
        "level": "primary",
        "actionType": "dialog",
        "dialog": {
          "title": "Create Contract",
          "body": { "$ref": "#/definitions/createForm" }
        }
      },
      "bulkActions",
      "export-csv",
      "columns-toggler",
      "pagination"
    ],
    "footerToolbar": ["statistics", "pagination"],
    "filterTogglable": true,
    "filter": {
      "title": "Search Contracts",
      "body": [
        {
          "type": "input-text",
          "name": "q",
          "label": "Search",
          "placeholder": "Contract number or title"
        },
        {
          "type": "select",
          "name": "status",
          "label": "Status",
          "clearable": true,
          "options": [
            { "label": "Draft", "value": "draft" },
            { "label": "Under Review", "value": "under_review" },
            { "label": "Approved", "value": "approved" },
            { "label": "Active", "value": "active" },
            { "label": "Terminated", "value": "terminated" }
          ]
        },
        {
          "type": "input-date-range",
          "name": "start_date",
          "label": "Start Date"
        }
      ]
    },
    "columns": [
      {
        "name": "contract_number",
        "label": "Number",
        "sortable": true,
        "type": "link",
        "href": "/contracts/${id}"
      },
      {
        "name": "title",
        "label": "Title",
        "sortable": true
      },
      {
        "name": "status",
        "label": "Status",
        "type": "mapping",
        "map": {
          "draft":        "<span class='label label-default'>Draft</span>",
          "under_review": "<span class='label label-warning'>Under Review</span>",
          "approved":     "<span class='label label-info'>Approved</span>",
          "active":       "<span class='label label-success'>Active</span>",
          "terminated":   "<span class='label label-danger'>Terminated</span>"
        }
      },
      {
        "name": "total_value",
        "label": "Value",
        "type": "number",
        "precision": 2,
        "prefix": "${currency} "
      },
      {
        "name": "end_date",
        "label": "Expires",
        "type": "date",
        "format": "YYYY-MM-DD"
      },
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
            "label": "Edit",
            "type": "button",
            "actionType": "dialog",
            "visibleOn": "${status === 'draft'}",
            "dialog": {
              "title": "Edit Contract",
              "body": { "type": "service", "api": "/api/v1/contracts/${id}", "body": { "$ref": "#/definitions/editForm" } }
            }
          },
          {
            "label": "Delete",
            "type": "button",
            "level": "danger",
            "actionType": "ajax",
            "api": { "method": "delete", "url": "/api/v1/contracts/${id}" },
            "confirmText": "Delete contract ${contract_number}? This cannot be undone.",
            "visibleOn": "${status === 'draft'}"
          }
        ]
      }
    ]
  }
}
```

## Pagination

AMIS `crud` auto-paginates. The backend must return:

```json
{
  "data": [...],
  "meta": {
    "total": 142,
    "limit": 20,
    "offset": 0
  }
}
```

AMIS maps `meta.total` → total count. Configure `perPageAvailable` to match backend limits:

```json
{
  "type": "crud",
  "perPageAvailable": [10, 20, 50, 100],
  "api": "/api/v1/contracts"
}
```

## Bulk Actions

```json
{
  "type": "crud",
  "bulkActions": [
    {
      "label": "Bulk Archive",
      "actionType": "ajax",
      "api": {
        "method": "post",
        "url": "/api/v1/contracts/bulk-archive",
        "data": { "ids": "${ids}" }
      },
      "confirmText": "Archive ${selectedItems.length} contracts?"
    }
  ]
}
```

## Inline Editing

For simple fields, use `quickEdit` on the column:

```json
{
  "name": "title",
  "label": "Title",
  "quickEdit": {
    "type": "input-text",
    "saveImmediately": {
      "api": { "method": "patch", "url": "/api/v1/contracts/${id}", "data": { "title": "${title}" } }
    }
  }
}
```

Use sparingly — prefer dialog forms for multi-field edits to avoid partial-save confusion.

## Empty State

```json
{
  "type": "crud",
  "placeholder": "No contracts found. Create one to get started.",
  "api": "/api/v1/contracts"
}
```

## Nested Detail Panel

Use `expandable` for master-detail without navigation:

```json
{
  "type": "crud",
  "expandable": {
    "expandableOn": "${lines_count > 0}",
    "keyField": "id",
    "type": "service",
    "api": "/api/v1/contracts/${id}/lines",
    "body": {
      "type": "table",
      "columns": [
        { "name": "description", "label": "Description" },
        { "name": "quantity", "label": "Qty" },
        { "name": "unit_price", "label": "Unit Price" }
      ]
    }
  }
}
```
