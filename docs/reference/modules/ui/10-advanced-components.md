[<-- Back to Index](README.md)

## Advanced Components

Patterns from `web/component.md` — power-user features used across Awo ERP modules.

### Select with "Add New"

Let users add a new option to a dropdown without leaving the form. Essential for ERP: add a new supplier while creating a PO, add a new GL account while posting a journal.

```json
{
  "type":           "select",
  "name":           "supplier_id",
  "label":          "Supplier",
  "source":         "/api/v1/suppliers/options",
  "searchable":     true,
  "clearable":      true,
  "creatable":      true,
  "createBtnLabel": "Add New Supplier",
  "addDialog": {
    "title": "Create New Supplier",
    "size":  "md",
    "body": {
      "type": "form",
      "api":  "post:/api/v1/suppliers/quick-add",
      "body": [
        { "type": "input-text", "name": "name",  "label": "Supplier Name", "required": true },
        { "type": "input-text", "name": "email", "label": "Email" },
        { "type": "input-text", "name": "phone", "label": "Phone" }
      ]
    }
  }
}
```

Go handler for `quick-add` must return `{ status: 0, data: { label: "...", value: "uuid" } }` — AMIS uses this to auto-select the newly created option.

### Inline Table Editing (quickEdit)

Edit cell values directly in the table row without opening a dialog:

```json
{
  "type": "crud",
  "api":  "/api/v1/inventory/products",
  "quickSaveApi": "put:/api/v1/inventory/products/${id}",
  "columns": [
    {
      "name":  "price",
      "label": "Price",
      "type":  "tpl",
      "tpl":   "KES ${price|number:2}",
      "quickEdit": {
        "type":          "input-number",
        "prefix":        "KES",
        "precision":     2,
        "min":           0,
        "saveImmediately": true
      }
    },
    {
      "name":  "is_active",
      "label": "Active",
      "type":  "mapping",
      "map":   { "1": "Active", "0": "Inactive" },
      "quickEdit": {
        "type":          "switch",
        "trueValue":     1,
        "falseValue":    0,
        "saveImmediately": true
      }
    }
  ]
}
```

`saveImmediately: true` → saves on blur without a "save row" button. Use for simple edits. For multi-field edits, omit it — all changes save together when the user clicks Save.

### Expandable Rows

Show additional detail inline without leaving the list:

```json
{
  "type":       "crud",
  "api":        "/api/v1/sales/orders",
  "expandable": {
    "expandableOn": "true",
    "component": {
      "type": "service",
      "api":  "get:/api/v1/sales/orders/${id}/lines",
      "body": {
        "type": "table",
        "source": "${items}",
        "columns": [
          { "name": "product_name", "label": "Product" },
          { "name": "quantity",     "label": "Qty" },
          { "name": "unit_price",   "label": "Unit Price", "type": "number" }
        ]
      }
    }
  },
  "columns": [ "...regular columns..." ]
}
```

### Cascading Selects

Second select filters based on first:

```json
[
  {
    "type":     "select",
    "name":     "country_id",
    "label":    "Country",
    "source":   "/api/v1/countries",
    "required": true
  },
  {
    "type":       "select",
    "name":       "region_id",
    "label":      "Region",
    "source":     "/api/v1/regions?country_id=${country_id}",
    "required":   true,
    "visibleOn":  "${country_id}",
    "clearValueOnHidden": true
  }
]
```

The `source` re-fetches whenever `country_id` changes. `clearValueOnHidden: true` resets the region selection when country changes.

### Tree Select — Department Hierarchy

```json
{
  "type":         "tree-select",
  "name":         "department_id",
  "label":        "Department",
  "source":       "/api/v1/departments/tree",
  "searchable":   true,
  "clearable":    true,
  "valueField":   "id",
  "labelField":   "name",
  "childrenField": "children"
}
```

Go handler returns hierarchical JSON:
```json
[
  { "id": "1", "name": "Operations", "children": [
    { "id": "2", "name": "Warehouse",   "children": [] },
    { "id": "3", "name": "Forecourt",   "children": [] }
  ]}
]
```

### Transfer (Dual-List) — Multi-Assignment

Used for assigning users to roles, products to categories, employees to shifts:

```json
{
  "type":        "transfer",
  "name":        "assigned_users",
  "label":       "Assign Users",
  "source":      "/api/v1/users?available=true",
  "searchable":  true,
  "sortable":    true,
  "valueField":  "id",
  "labelField":  "name"
}
```

### Custom Field Validators

```json
{
  "type":     "input-text",
  "name":     "kra_pin",
  "label":    "KRA PIN",
  "required": true,
  "validations": {
    "matchRegexp": "^[A-Z]\\d{9}[A-Z]$"
  },
  "validationErrors": {
    "matchRegexp": "KRA PIN must be in format A000000000Z"
  }
}
```

For async validation (e.g. check unique email):

```json
{
  "type":            "input-email",
  "name":            "email",
  "label":           "Email",
  "required":        true,
  "validateApi":     "get:/api/v1/users/check-email?email=${email}",
  "validateOnChange": false
}
```

Go handler for validate endpoint: return `{ status: 0 }` if valid, `{ status: 422, msg: "Email already taken" }` if not.

### Draggable List (Priority Ordering)

For reordering items like business rule priorities:

```json
{
  "type":         "crud",
  "api":          "get:/api/v1/settings/rules",
  "draggable":    true,
  "saveOrderApi": "post:/api/v1/settings/rules/reorder",
  "columns": [
    { "name": "name",     "label": "Rule Name" },
    { "name": "priority", "label": "Priority"  }
  ]
}
```

Go handler for reorder receives `{ ids: ["uuid1", "uuid2", ...] }` in the order the user dragged them.

---
