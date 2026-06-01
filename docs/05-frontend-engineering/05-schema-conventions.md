---
title: Schema Conventions
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer]
related:
  - "[AMIS Schema Guide](02-amis-schema-guide.md)"
  - "[Component Patterns](04-component-patterns.md)"
  - "[API Reference](../08-api-reference/01-api-reference-overview.md)"
---

# Schema Conventions

## File Organization

```
web/schemas/pages/
  contracts/
    list.json          ← CRUD list page
    create.json        ← create form (or inline in list)
    edit.json          ← edit form
    detail.json        ← read-only detail view
  finance/
    accounts/
      list.json
    transactions/
      list.json
  iam/
    users/
      list.json
```

One file per page view. Large schemas should be split by section and composed via `service` component.

## Page Root Schema

Every page schema has a root `page` component:

```json
{
  "type": "page",
  "title": "Contracts",
  "body": [...]
}
```

Do not use nested `page` components — only one per schema file.

## API URLs

All API URLs are relative (no hostname):

```json
{ "api": "/api/v1/contracts" }
```

Not:
```json
{ "api": "https://acme.awoerp.com/api/v1/contracts" }
```

The AMIS SDK sends requests relative to the current origin. Using absolute URLs breaks multi-tenant subdomains.

## Authentication Header

AMIS must inject the `Authorization` header on every API request. Configure globally when embedding:

```javascript
amis.embed('#app', schema, initialData, {
    theme: 'cxd',
    fetcher({ url, method, data, headers }) {
        const token = localStorage.getItem('session_token');
        return fetch(url, {
            method: method.toUpperCase(),
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token ? `Bearer ${token}` : '',
                ...headers,
            },
            body: data ? JSON.stringify(data) : undefined,
        }).then(r => r.json());
    }
});
```

## Naming in Schemas

| Convention | Example |
|------------|---------|
| Component IDs | `kebab-case`, unique per page | `contracts-list`, `create-form` |
| Field names | Match API response field names | `contract_number`, `total_value` |
| Action names | verb-noun | `create-contract`, `submit-for-review` |

## Error Response Format

AMIS expects errors in a specific format. The backend sends:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "contract_number", "message": "contract_number is required" }
    ]
  }
}
```

AMIS reads `details` and maps field errors to form inputs automatically when `name` matches the `field` value.

## Pagination

AMIS CRUD expects the API to return:

```json
{
  "data": [...],
  "pagination": {
    "total": 87
  }
}
```

Configure CRUD:

```json
{
  "type": "crud",
  "api": "/api/v1/contracts",
  "perPage": 20,
  "perPageAvailable": [10, 20, 50, 100],
  "defaultParams": { "page": 1, "page_size": 20 },
  "pageField": "page",
  "perPageField": "page_size",
  "totalField": "pagination.total"
}
```

## Form Submit Data

By default AMIS submits all form fields. Restrict with `data` on the API:

```json
{
  "api": {
    "method": "post",
    "url": "/api/v1/contracts",
    "data": {
      "contract_number": "${contract_number}",
      "title": "${title}",
      "total_value": "${total_value}",
      "currency": "${currency}",
      "start_date": "${start_date}",
      "end_date": "${end_date}"
    }
  }
}
```

This prevents unexpected fields being sent when form components accumulate hidden state.

## Refresh After Mutation

After create/update/delete actions, reload the list component:

```json
{
  "onEvent": {
    "submitSucc": {
      "actions": [
        {
          "actionType": "toast",
          "args": { "msgType": "success", "msg": "Contract created." }
        },
        {
          "actionType": "reload",
          "componentId": "contracts-list"
        },
        {
          "actionType": "closeDialog"
        }
      ]
    }
  }
}
```

## Schema Validation Before Commit

Before committing new schemas:
1. Open in browser and exercise all interactions manually
2. Check network tab — confirm no 4xx/5xx on initial load
3. Test with a user that has restricted permissions — confirm `visibleOn` conditions work
4. Test dark mode — toggle `html.dark` class in devtools

## Large Schema Organization

For complex pages (>200 lines), split into composable pieces using `service` component:

```json
{
  "type": "page",
  "title": "Contracts",
  "body": [
    {
      "type": "service",
      "schemaApi": "/static/schemas/contracts/filter-bar.json",
      "id": "filter-bar"
    },
    {
      "type": "service",
      "schemaApi": "/static/schemas/contracts/list-table.json",
      "id": "contracts-list"
    }
  ]
}
```

Sub-schemas are loaded independently, reducing initial schema parse time.
