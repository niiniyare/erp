---
title: AMIS Schema Guide
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer, fullstack-engineer]
related:
  - "[Frontend Overview](01-frontend-overview.md)"
  - "[Dark Theme](03-dark-theme.md)"
---

# AMIS Schema Guide

## Basic Page Schema

```json
{
  "type": "page",
  "title": "Contracts",
  "body": {
    "type": "crud",
    "api": {
      "method": "get",
      "url": "/api/v1/contracts",
      "headers": { "Authorization": "Bearer ${token}" },
      "responseData": {
        "items": "${data.data}",
        "total": "${data.pagination.total}"
      }
    },
    "columns": [
      { "name": "contract_number", "label": "Number" },
      { "name": "title", "label": "Title" },
      { "name": "status", "label": "Status", "type": "mapping",
        "map": {
          "draft": "<span class='label label-default'>Draft</span>",
          "submitted": "<span class='label label-info'>Submitted</span>",
          "active": "<span class='label label-success'>Active</span>",
          "terminated": "<span class='label label-danger'>Terminated</span>"
        }
      },
      { "name": "total_value", "label": "Value" },
      { "name": "end_date", "label": "End Date" }
    ],
    "footerToolbar": ["statistics", "pagination"]
  }
}
```

## Create Form

```json
{
  "type": "dialog",
  "title": "Create Contract",
  "body": {
    "type": "form",
    "api": {
      "method": "post",
      "url": "/api/v1/contracts",
      "headers": { "Authorization": "Bearer ${token}" }
    },
    "body": [
      { "type": "input-text", "name": "contract_number", "label": "Contract Number", "required": true },
      { "type": "input-text", "name": "title", "label": "Title", "required": true },
      { "type": "select", "name": "contract_type", "label": "Type",
        "options": [
          {"label": "Service", "value": "service"},
          {"label": "Goods", "value": "goods"},
          {"label": "Lease", "value": "lease"}
        ]
      },
      { "type": "input-date", "name": "start_date", "label": "Start Date", "format": "YYYY-MM-DD" },
      { "type": "input-date", "name": "end_date", "label": "End Date", "format": "YYYY-MM-DD" },
      { "type": "input-number", "name": "total_value", "label": "Value" },
      { "type": "input-text", "name": "currency", "label": "Currency", "value": "USD" }
    ]
  }
}
```

## Status Transition Actions

```json
{
  "type": "button",
  "label": "Submit for Review",
  "visibleOn": "${status === 'draft'}",
  "actionType": "ajax",
  "api": {
    "method": "post",
    "url": "/api/v1/contracts/${id}/submit",
    "data": { "version": "${version}" }
  },
  "confirmText": "Submit contract ${contract_number} for review?"
}
```

## Error Handling

AMIS expects `{"status": 0}` for success and `{"status": 1, "msg": "error message"}` for errors.

The backend returns `{"error": {"code": "...", "message": "..."}}`. Use AMIS `responseData` transform or an API adapter middleware to remap:

```json
{
  "api": {
    "url": "/api/v1/contracts",
    "responseData": {
      "&": "$$",
      "status": "${status >= 200 && status < 300 ? 0 : 1}",
      "msg": "${error.message || ''}",
      "data": "${data}"
    }
  }
}
```

## Token Injection

Store the session token in `localStorage` and inject via AMIS global data:

```javascript
amis.embed('#root', schema, {
  data: {
    token: localStorage.getItem('session_token')
  }
})
```

All API schemas reference `${token}` in the `Authorization` header.
