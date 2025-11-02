# Conditional Logic

## Show/Hide Fields

### Simple Condition

```json
{
  "name": "shipping_address",
  "type": "text",
  "label": "Shipping Address",
  "showIf": "needs_shipping === true"
}
```

Uses Alpine.js on client:
```html
<div x-show="needs_shipping === true">
  <input name="shipping_address" />
</div>
```

### Complex Condition

```json
{
  "showIf": "order_total > 100 && country === 'US'"
}
```

## Permission-Based Visibility

```json
{
  "name": "salary",
  "type": "currency",
  "label": "Salary",
  "requirePermission": "hr.view_salary"
}
```

Server sets runtime visibility:
```go
field.Runtime = &FieldRuntime{
    Visible: user.HasPermission("hr.view_salary"),
}
```

## Conditional Required

```json
{
  "name": "tax_id",
  "type": "text",
  "label": "Tax ID",
  "requiredIf": "is_business === true"
}
```

## Example: Shipping Form

```json
{
  "fields": [
    {
      "name": "needs_shipping",
      "type": "checkbox",
      "label": "Ship to different address"
    },
    {
      "name": "shipping_street",
      "type": "text",
      "label": "Street",
      "showIf": "needs_shipping === true",
      "requiredIf": "needs_shipping === true"
    },
    {
      "name": "shipping_city",
      "type": "text",
      "label": "City",
      "showIf": "needs_shipping === true",
      "requiredIf": "needs_shipping === true"
    }
  ]
}
```

[← Back](07-layout.md) | [Next: Storage →](09-storage-interface.md)