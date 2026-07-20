> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Form Patterns
portal: 5 — Frontend Engineering
section: 05-frontend-engineering
audience: [frontend-engineer, full-stack-engineer]
related:
  - "[AMIS Schema Overview](01-amis-schema-overview.md)"
  - "[CRUD Patterns](02-crud-patterns.md)"
---

# Form Patterns

## Standard Create Form

```json
{
  "type": "form",
  "title": "Create Contract",
  "api": {
    "method": "post",
    "url": "/api/v1/contracts"
  },
  "redirect": "/contracts/${id}",
  "body": [
    {
      "type": "input-text",
      "name": "title",
      "label": "Title",
      "required": true,
      "maxLength": 255,
      "placeholder": "Enter contract title"
    },
    {
      "type": "select",
      "name": "contract_type",
      "label": "Contract Type",
      "required": true,
      "options": [
        { "label": "Service", "value": "service" },
        { "label": "Supply", "value": "supply" },
        { "label": "License", "value": "license" },
        { "label": "Framework", "value": "framework" }
      ]
    },
    {
      "type": "select",
      "name": "vendor_id",
      "label": "Vendor",
      "required": true,
      "source": "/api/v1/entities?entity_type=COMPANY",
      "labelField": "name",
      "valueField": "id",
      "searchable": true
    },
    {
      "type": "input-number",
      "name": "total_value",
      "label": "Total Value",
      "required": true,
      "precision": 6,
      "min": 0
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
    },
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
      "format": "YYYY-MM-DD",
      "validations": { "isAfter": "${start_date}" },
      "validationErrors": { "isAfter": "End date must be after start date" }
    },
    {
      "type": "textarea",
      "name": "description",
      "label": "Description",
      "maxLength": 2000
    }
  ]
}
```

## Edit Form (pre-populate)

Wrap in a `service` to load existing data:

```json
{
  "type": "service",
  "api": "/api/v1/contracts/${id}",
  "body": {
    "type": "form",
    "title": "Edit Contract",
    "api": {
      "method": "put",
      "url": "/api/v1/contracts/${id}"
    },
    "initApi": "/api/v1/contracts/${id}",
    "body": []
  }
}
```

`initApi` pre-fills form fields from the API response. Field `name` must match API response keys exactly.

## Field Visibility and Linkage

Show/hide fields based on other field values with `visibleOn`:

```json
{
  "type": "input-number",
  "name": "discount_pct",
  "label": "Discount %",
  "visibleOn": "${contract_type === 'framework'}"
}
```

Computed fields with `disabledOn` to show derived values:

```json
{
  "type": "static",
  "name": "net_value",
  "label": "Net Value",
  "tpl": "${currency} ${(total_value * (1 - discount_pct / 100)).toFixed(2)}"
}
```

## Dynamic Options from API

```json
{
  "type": "select",
  "name": "entity_id",
  "label": "Business Unit",
  "source": {
    "url": "/api/v1/entities",
    "method": "get",
    "data": { "entity_type": "DEPARTMENT", "status": "active" }
  },
  "labelField": "name",
  "valueField": "id",
  "searchable": true,
  "clearable": true
}
```

## Validation Patterns

```json
{
  "type": "input-text",
  "name": "contract_number",
  "label": "Contract Number",
  "required": true,
  "validations": {
    "matchRegexp": "/^CONT-\\d{4}-\\d{4}$/"
  },
  "validationErrors": {
    "matchRegexp": "Format must be CONT-YYYY-NNNN"
  }
}
```

Server-side validation errors (422 response) are automatically displayed under the relevant field if the error `field` key matches the form field `name`.

## Multi-Step Form (Wizard)

For complex forms with logical grouping:

```json
{
  "type": "wizard",
  "api": {
    "method": "post",
    "url": "/api/v1/contracts"
  },
  "steps": [
    {
      "title": "Basic Info",
      "body": [
        { "type": "input-text", "name": "title", "label": "Title", "required": true }
      ]
    },
    {
      "title": "Financial Terms",
      "body": [
        { "type": "input-number", "name": "total_value", "label": "Value", "required": true }
      ]
    },
    {
      "title": "Review",
      "body": [
        { "type": "static", "name": "title", "label": "Title" },
        { "type": "static", "name": "total_value", "label": "Value" }
      ]
    }
  ]
}
```

## Form in Dialog

Forms inside dialogs close and reload the parent CRUD on submit:

```json
{
  "type": "button",
  "label": "Add Line",
  "actionType": "dialog",
  "dialog": {
    "title": "Add Contract Line",
    "body": {
      "type": "form",
      "api": {
        "method": "post",
        "url": "/api/v1/contracts/${contractId}/lines"
      },
      "body": [
        { "type": "input-text", "name": "description", "label": "Description", "required": true },
        { "type": "input-number", "name": "quantity", "label": "Qty", "required": true },
        { "type": "input-number", "name": "unit_price", "label": "Unit Price", "required": true, "precision": 6 }
      ]
    },
    "actions": [
      { "type": "button", "actionType": "cancel", "label": "Cancel" },
      { "type": "submit", "label": "Add", "level": "primary" }
    ]
  }
}
```
