# Chapter 6: Configuration Definitions

[← Data Model](./05-data-model.md) | [Next: Tenant Configuration →](./07-tenant-configuration.md)

---

## What Are Configuration Definitions?

Configuration definitions are the schema registry for the Settings Module. They declare what is configurable, what type it must be, what the system-level default is, and any validation rules. Without a definition, a configuration key cannot be used.

Think of `config_definitions` as the "contract" that all configuration values must satisfy.

---

## Configuration Types

| Type | Description | Example Value |
|------|-------------|---------------|
| `string` | Text value | `"INV-"`, `"USD"`, `"NET30"` |
| `integer` | Whole number | `40`, `100`, `1000` |
| `boolean` | True/false flag | `true`, `false` |
| `decimal` | Decimal number | `1.5`, `1000.00` |
| `json` | Arbitrary JSON object/array | `{"method": "FIFO", "auto": true}` |

---

## Creating a Configuration Definition

```sql
-- name: CreateConfigDefinition :one
INSERT INTO config_definitions (
    module_name, config_key, config_type, default_value,
    validation_rules, description, required_feature_flag
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;
```

Example — registering an invoice prefix configuration:

```go
def, err := store.CreateConfigDefinition(ctx, db.CreateConfigDefinitionParams{
    ModuleName:   "finance",
    ConfigKey:    "invoice_prefix",
    ConfigType:   "string",
    DefaultValue: []byte(`"INV-"`),
    ValidationRules: []byte(`{
        "type": "string",
        "minLength": 2,
        "maxLength": 20,
        "pattern": "^[A-Z0-9-]+$"
    }`),
    Description: "Prefix applied to all invoice document numbers",
    RequiredFeatureFlag: pgtype.Text{Valid: false},
})
```

---

## Listing Definitions

```sql
-- name: ListConfigDefinitions :many
SELECT * FROM config_definitions
WHERE (module_name IS NULL OR module_name = $1)
ORDER BY module_name, config_key;
```

Filter by module to get all keys for a specific module:

```go
defs, err := store.ListConfigDefinitions(ctx, db.ListConfigDefinitionsParams{
    ModuleName: pgtype.Text{String: "finance", Valid: true},
})
```

---

## Validation Rules

Validation rules are stored as JSONB and support JSON Schema-style constraints. The service layer validates configuration values against these rules before persisting:

```json
{
  "type": "integer",
  "minimum": 0,
  "maximum": 1000000,
  "description": "Approval limit in base currency units"
}
```

For string enums:

```json
{
  "type": "string",
  "enum": ["FIFO", "LIFO", "AVCO", "SPECIFIC"],
  "description": "Inventory valuation method"
}
```

---

## Feature Flag Gating

If `required_feature_flag` is set, the configuration key is only accessible to tenants that have that feature flag enabled. This allows progressive rollout of new configuration capabilities:

```sql
-- Only tenants with "inventory.lot_tracking" flag can use this key
INSERT INTO config_definitions (module_name, config_key, ..., required_feature_flag)
VALUES ('inventory', 'lot_tracking_mode', ..., 'inventory.lot_tracking');
```

---

## Standard Module Configuration Keys

### Finance Module

| Config Key | Type | Default | Description |
|---|---|---|---|
| `invoice_prefix` | string | `"INV-"` | Invoice document prefix |
| `quote_prefix` | string | `"QUO-"` | Quotation document prefix |
| `po_prefix` | string | `"PO-"` | Purchase order prefix |
| `default_currency` | string | `"USD"` | Default transaction currency |
| `auto_approval_limit` | decimal | `1000.00` | Auto-approve below this amount |
| `payment_terms_default` | string | `"NET30"` | Default payment terms |
| `multi_currency_enabled` | boolean | `false` | Enable multi-currency support |
| `fiscal_year_start` | integer | `1` | Month fiscal year starts (1=Jan) |

### HR Module

| Config Key | Type | Default | Description |
|---|---|---|---|
| `default_pay_frequency` | string | `"biweekly"` | Payroll frequency |
| `overtime_threshold` | integer | `40` | Weekly hours before overtime |
| `overtime_multiplier` | decimal | `1.5` | Overtime pay multiplier |
| `tax_jurisdiction` | string | `"US-FED"` | Default tax jurisdiction |
| `vacation_accrual_method` | string | `"monthly"` | Vacation accrual method |
| `sick_leave_annual_hours` | integer | `40` | Annual sick leave hours |

### Inventory Module

| Config Key | Type | Default | Description |
|---|---|---|---|
| `valuation_method` | string | `"FIFO"` | Cost valuation method |
| `allow_negative_stock` | boolean | `false` | Allow stock to go negative |
| `auto_reorder_enabled` | boolean | `true` | Enable automatic reordering |
| `cycle_count_frequency` | string | `"monthly"` | Cycle count schedule |
| `lot_tracking_required` | boolean | `false` | Require lot numbers |
| `serial_tracking_required` | boolean | `false` | Require serial numbers |

---

[← Data Model](./05-data-model.md) | [Next: Tenant Configuration →](./07-tenant-configuration.md)
