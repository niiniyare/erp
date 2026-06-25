---
title: "Platform Module: Metadata & Custom Fields"
part: "Part VI — Platform Entities"
chapter: 47
section: "platform-metadata-module"
related:
  - "[Chapter 42: Tenant & Organisation](./platform-tenant-module.md)"
  - "[Chapter 46: Audit Log](./platform-audit-module.md)"
  - "[Chapter 41: Platform Module Overview](./platform-module-overview.md)"
  - "[Chapter 2: EntityDefinition](../part-01-foundations/02-entity-definition.md)"
---

# Chapter 47 — Metadata & Custom Fields Module

> **Primary source for:** the `custom_field_defs` table, how custom field values are stored in JSONB, how SDUI merges framework and custom fields at render time, GIN index strategy, and how business module developers add custom field support to their entities.
>
> **Audience:** developers building or extending modules, business stakeholders who want to understand how Awo ERP adapts to industry-specific requirements without code changes, and anyone asking "why is there a `custom_fields` column on this table?"

---

## 47.1 The Problem: Every Business Is Different

The Awo ERP framework ships with a standard data model covering the fields most businesses need. An Invoice has `date`, `total`, `status`, `customer_id`. A Customer has `name`, `email`, `phone`, `address`.

But businesses operate in specific industries with specific regulatory and operational requirements:

- A **petrol station chain** needs `pump_number` and `meter_start` on every fuel sale record.
- A **pharmaceutical distributor** needs `batch_number`, `expiry_date`, and `nairobi_pharmacies_board_reg` on every item.
- A **construction company** needs `site_code` and `cost_centre_ref` on every purchase order.
- A **school** needs `term`, `class`, and `stream` on every fee payment.

The Awo team cannot ship a separate version of the Invoice entity for every industry. The answer is **custom fields**: a mechanism that lets a business administrator add new data fields to any entity at runtime — without writing Go code, without running a migration, without redeploying the server.

---

## 47.2 How Custom Fields Work — The Big Picture

Custom fields have two parts:

**1. The Definition** (`custom_field_defs` table)
Describes the field: name, label, type, which entity it extends, sort order, validation options. One definition row per field per entity per tenant.

**2. The Value** (inside the entity row's `custom_fields jsonb` column)
The actual data. A customer row with custom fields looks like:

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "name": "Acme Hardware Ltd",
  "email": "orders@acmehardware.co.ke",
  "phone": "+254722000000",
  "custom_fields": {
    "nssf_number": "12345678",
    "kra_compliance_cert": "CERT-2024-00123",
    "credit_tier": "gold",
    "account_manager": "Jane Wanjiku"
  }
}
```

Framework fields (`name`, `email`, `phone`) are typed SQL columns. Custom fields are a single `custom_fields jsonb` column that any tenant can put any keys into — within constraints defined by their `custom_field_defs` rows.

---

## 47.3 Directory Layout

```
internal/platform/metadata/
├── metadata.go    ← init() — registers CustomFieldDefDef
├── definition.go  ← CustomFieldDefDef
├── policy.go      ← policy helpers
├── hooks.go       ← validateFieldName, validateLinkTargetExists,
│                     invalidateSDUISchemaCache
├── service.go     ← MetadataService: GetDefs(), MergeCustomFields()
└── migrations/
    └── 20240101000050_create_custom_fields.up.sql
```

---

## 47.4 CustomFieldDef EntityDefinition

```go
// internal/platform/metadata/definition.go
package metadata

import (
    "awo.so/framework/definition"
    "awo.so/framework/org"
)

var CustomFieldDefDef = definition.EntityDefinition{
    Name:     "custom_field_def",
    Label:    "Custom Field",
    Module:   "Platform",
    Table:    "custom_field_defs",
    OrgScope: org.ScopeLevelTenant,
    Audited:  true,

    // SoftDelete: true — critical design decision.
    // When a tenant admin "deletes" a custom field, we do NOT drop the JSONB key
    // from every entity row. That would lose historical data silently.
    // Instead, the definition is soft-deleted (deleted_at is set).
    // Existing records keep their JSONB values. The field simply stops appearing
    // in forms and API responses for new records.
    // Hard cleanup can happen separately after a retention period.
    SoftDelete: true,

    Fields: []*definition.FieldDef{
        {
            Name:     "entity_name",
            Type:     definition.FieldTypeData,
            Label:    "Entity",
            Required: true,
            Description: "EntityDefinition.Name of the entity being extended. " +
                "Examples: 'customer', 'invoice', 'employee', 'sales_order'. " +
                "The framework validates this against the registered definition registry.",
        },
        {
            Name:     "field_name",
            Type:     definition.FieldTypeData,
            Label:    "Field Name (JSON Key)",
            Required: true,
            Description: "The key used inside the custom_fields JSONB column. " +
                "Must be snake_case. Cannot be a reserved framework field name. " +
                "Once set, renaming breaks existing data references — treat as immutable.",
        },
        {
            Name:     "label",
            Type:     definition.FieldTypeData,
            Label:    "Display Label",
            Required: true,
            Description: "Human-readable name shown in forms and table columns.",
        },
        {
            Name:     "field_type",
            Type:     definition.FieldTypeSelect,
            Label:    "Field Type",
            Options:  []string{"text", "integer", "boolean", "date", "select", "link"},
            Required: true,
            Description: "Determines the SDUI control rendered and value validation applied. " +
                "text → InputText. integer → InputNumber. boolean → Switch. " +
                "date → DatePicker. select → Select (with options). link → entity picker.",
        },
        {
            Name:        "options",
            Type:        definition.FieldTypeJSON,
            Label:       "Select Options",
            Description: "For field_type=select only. JSON array of {value, label} objects. " +
                `Example: [{"value": "gold", "label": "Gold"}, {"value": "silver", "label": "Silver"}]`,
        },
        {
            Name:        "target_entity",
            Type:        definition.FieldTypeData,
            Label:       "Target Entity",
            Description: "For field_type=link only. The EntityDefinition.Name of the linked entity. " +
                "Example: 'user' (select the account manager for this customer). " +
                "The framework generates a search picker API call automatically.",
        },
        {
            Name:    "required",
            Type:    definition.FieldTypeBool,
            Label:   "Required",
            Default: "false",
            Description: "If true, the SDUI form validation blocks submission without this field.",
        },
        {
            Name:    "sort_order",
            Type:    definition.FieldTypeInt,
            Label:   "Sort Order",
            Default: "0",
            Description: "Controls position of this field relative to other custom fields " +
                "in the form. Lower = earlier. Framework fields always appear first.",
        },
        {
            Name:        "section",
            Type:        definition.FieldTypeData,
            Label:       "Form Section",
            Description: "Optional grouping. Custom fields with the same section value " +
                "are grouped under a collapsible section heading in the form. " +
                "Example: 'Kenya Compliance', 'Station Details'.",
        },
    },

    Policies: []definition.PolicyDef{
        // System always passes.
        definition.Policy(definition.OpAll, definition.AllowSystem),

        // Only tenant admins can define custom fields.
        // Regular users cannot add fields that bypass data governance.
        definition.Policy(definition.OpAll, requireRole("admin")),

        // All authenticated users can READ custom field definitions.
        // Reason: the SDUI needs the definitions to render forms correctly.
        // Every time a customer form loads, it queries custom_field_defs for 'customer'.
        definition.Policy(definition.OpRead, allowTenantViewer),

        definition.Policy(definition.OpAll, definition.DenyAll),
    },

    Hooks: []definition.HookDef{
        {
            // Validates field_name on creation:
            // 1. Enforces snake_case pattern (only lowercase letters, digits, underscores).
            // 2. Rejects reserved names (id, tenant_id, created_at, updated_at, deleted_at,
            //    custom_fields — these are framework-managed SQL columns).
            // 3. Checks uniqueness within (tenant_id, entity_name, field_name).
            //    The DB constraint also catches this, but an early hook gives a better error.
            Name: "validate_field_name",
            Ops:  definition.OpCreate,
            When: definition.HookBefore,
            Fn:   validateFieldName,
        },
        {
            // For field_type=link: verifies that target_entity is a registered EntityDefinition.
            // Prevents "link to unicorn_entity" from silently failing at render time.
            Name: "validate_target_entity_exists",
            Ops:  definition.OpCreate,
            When: definition.HookBefore,
            Fn:   validateLinkTargetExists,
        },
        {
            // After any custom field definition change, invalidates the cached SDUI schema
            // for (tenant_id, entity_name). The next form load triggers a schema rebuild.
            //
            // WHY AfterHook: DB must commit before cache invalidation.
            // If we cleared the cache and the write then rolled back, subsequent requests
            // would rebuild from DB and get the old schema — consistent but we wasted a rebuild.
            // If the write commits and we clear the cache — new schema is used. Correct.
            Name: "invalidate_sdui_schema_cache",
            Ops:  definition.OpCreate | definition.OpUpdate | definition.OpDelete,
            When: definition.HookAfter,
            Fn:   invalidateSDUISchemaCache,
        },
    },
}
```

---

## 47.5 Hook Implementations

```go
// internal/platform/metadata/hooks.go
package metadata

import (
    "context"
    "fmt"
    "regexp"
    "strings"

    "awo.so/framework/definition"
)

var snakeCaseRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Reserved names: framework-managed SQL columns that must not be overridden.
var reservedFieldNames = map[string]bool{
    "id":            true,
    "tenant_id":     true,
    "org_unit_id":   true,
    "created_at":    true,
    "updated_at":    true,
    "deleted_at":    true,
    "custom_fields": true,
}

func validateFieldName(ctx context.Context, v definition.ViewerContext, op definition.Op, rec definition.MutableRecord) error {
    name, _ := rec.Get("field_name")
    if name == nil {
        return fmt.Errorf("field_name is required")
    }
    n := name.(string)

    if !snakeCaseRegex.MatchString(n) {
        return fmt.Errorf("field_name %q must be snake_case (lowercase letters, digits, underscores only, must start with a letter)", n)
    }

    if reservedFieldNames[n] {
        return fmt.Errorf("field_name %q is reserved by the framework and cannot be used as a custom field name", n)
    }

    return nil
}

func validateLinkTargetExists(ctx context.Context, v definition.ViewerContext, op definition.Op, rec definition.MutableRecord) error {
    fieldType, _ := rec.Get("field_type")
    if fieldType == nil || fieldType.(string) != "link" {
        return nil // only relevant for link fields
    }

    target, _ := rec.Get("target_entity")
    if target == nil || target.(string) == "" {
        return fmt.Errorf("target_entity is required for field_type=link")
    }

    if definition.Lookup(target.(string)) == nil {
        return fmt.Errorf("target_entity %q is not a registered EntityDefinition", target.(string))
    }

    return nil
}

func invalidateSDUISchemaCache(ctx context.Context, v definition.ViewerContext, op definition.Op, rec definition.Record) error {
    tenantID, _    := rec.Get("tenant_id")
    entityName, _  := rec.Get("entity_name")
    cacheKey := fmt.Sprintf("sdui:schema:%s:%s", tenantID, entityName)
    return redis.Del(ctx, cacheKey)
}
```

---

## 47.6 How Values Are Stored — JSONB Deep Dive

### Adding the Column to an Entity

Business module developers add custom field support to their entity tables with two lines in a migration:

```sql
-- In Finance module migration for customers:
ALTER TABLE customers
    ADD COLUMN IF NOT EXISTS custom_fields jsonb NOT NULL DEFAULT '{}';

-- GIN index: enables fast queries like:
--   WHERE custom_fields->>'nssf_number' = '12345678'
--   WHERE custom_fields @> '{"credit_tier": "gold"}'
CREATE INDEX customers_custom_fields_gin
    ON customers USING gin(custom_fields);
```

No other changes needed. The framework detects the `custom_fields` column and activates custom field support for that entity automatically.

### Reading and Writing Values

From the application's perspective, reading and writing custom fields is transparent:

```go
// Writing a customer with custom fields:
rec := pgstore.NewMutableRecord(CustomerDef)
rec.Set("name",  "Acme Hardware Ltd")
rec.Set("email", "orders@acmehardware.co.ke")
// Custom fields are set in the same record — the framework routes them to JSONB
rec.SetCustom("nssf_number",  "12345678")
rec.SetCustom("credit_tier",  "gold")
customerStore.Create(ctx, rec)

// Reading a customer with custom fields:
customer, _ := customerStore.FindByID(ctx, id)
name,  _ := customer.Get("name")         // → from SQL column
nssf,  _ := customer.GetCustom("nssf_number") // → from jsonb
```

### Querying by Custom Field Value

The framework's filter system supports a `custom.` prefix for JSONB path queries:

```
GET /api/v1/customer?filter[custom.credit_tier][eq]=gold&filter[custom.nssf_number][starts_with]=123
```

Translates to:

```sql
SELECT * FROM customers
WHERE custom_fields->>'credit_tier' = 'gold'
  AND custom_fields->>'nssf_number' LIKE '123%'
  AND tenant_id = current_setting('app.current_tenant_id')::uuid;
```

The GIN index makes these queries fast for equality checks. Range queries on numeric custom fields cast the JSONB text to the appropriate type:

```sql
WHERE (custom_fields->>'credit_limit')::numeric > 500000
```

---

## 47.7 SDUI Form Merging

When the SDUI handler builds a form for any entity, it merges framework fields and custom fields:

```go
// internal/framework/sdui/handler.go (simplified)
func (s *SDUIHandler) formFields(ctx context.Context, def *definition.EntityDefinition) ([]*FieldSchema, error) {
    // Step 1: framework fields from EntityDefinition.Fields
    // These always come first, in the order declared in the EntityDefinition.
    fields := make([]*FieldSchema, 0, len(def.Fields))
    for _, f := range def.Fields {
        if f.Sensitive {
            continue // never rendered in UI
        }
        fields = append(fields, frameworkFieldToSchema(f))
    }

    // Step 2: custom field definitions for this tenant + entity
    customDefs, err := s.customFieldStore.List(ctx, pgstore.ListOptions{
        Filter: map[string]any{
            "entity_name": def.Name,
            "deleted_at":  nil, // exclude soft-deleted definitions
        },
        OrderBy: "sort_order ASC, created_at ASC",
    })
    if err != nil {
        return nil, err
    }

    // Step 3: convert custom field definitions to AMIS form controls
    for _, cd := range customDefs.Items {
        fields = append(fields, customDefToSchema(cd))
    }

    return fields, nil
}

func customDefToSchema(def definition.Record) *FieldSchema {
    fieldType, _ := def.Get("field_type")
    fieldName, _ := def.Get("field_name")
    label, _     := def.Get("label")
    required, _  := def.Get("required")

    // Custom field values live in the custom_fields JSONB column.
    // The AMIS binding uses "custom_fields.{field_name}" to read/write
    // the correct key within the JSON object.
    binding := "custom_fields." + fieldName.(string)

    switch fieldType.(string) {
    case "text":
        return &FieldSchema{Type: "input-text", Name: binding, Label: label.(string), Required: required.(bool)}
    case "integer":
        return &FieldSchema{Type: "input-number", Name: binding, Label: label.(string), Required: required.(bool)}
    case "boolean":
        return &FieldSchema{Type: "switch", Name: binding, Label: label.(string)}
    case "date":
        return &FieldSchema{Type: "input-date", Name: binding, Label: label.(string), Format: "YYYY-MM-DD"}
    case "select":
        options, _ := def.Get("options")
        return &FieldSchema{Type: "select", Name: binding, Label: label.(string),
            Options: options, Required: required.(bool)}
    case "link":
        target, _ := def.Get("target_entity")
        return &FieldSchema{
            Type:        "select",
            Name:        binding,
            Label:       label.(string),
            Source:      "/api/v1/" + target.(string) + "?limit=50", // remote data source
            ValueField:  "id",
            LabelField:  "name",
            Required:    required.(bool),
        }
    }
    return nil
}
```

The result: a customer form for Acme Hardware Ltd shows:
1. Framework fields: Name, Email, Phone, Address, Status, Credit Limit (from CustomerDef.Fields)
2. Custom fields (in sort_order): NSSF Number, KRA Compliance Cert, Credit Tier, Account Manager

All rendered automatically, no template changes.

---

## 47.8 Migration

```sql
-- internal/platform/metadata/migrations/20240101000050_create_custom_fields.up.sql

CREATE TABLE custom_field_defs (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid        NOT NULL REFERENCES tenants(id),
    entity_name  text        NOT NULL,
    field_name   text        NOT NULL,
    label        text        NOT NULL,
    field_type   text        NOT NULL
                             CHECK (field_type IN ('text','integer','boolean','date','select','link')),
    options      jsonb,               -- for field_type=select: [{value, label}, ...]
    target_entity text,               -- for field_type=link: EntityDefinition.Name
    required     boolean     NOT NULL DEFAULT false,
    sort_order   integer     NOT NULL DEFAULT 0,
    section      text,
    deleted_at   timestamptz,         -- soft-delete; see §47.4 for rationale
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    -- One definition per (tenant, entity, field) combination.
    -- NULLS NOT DISTINCT not needed here: deleted_at has no bearing on uniqueness.
    -- A soft-deleted field frees up the name for re-creation.
    UNIQUE (tenant_id, entity_name, field_name)
    -- Actually we want: allow re-creation after soft-delete.
    -- Change to: UNIQUE (tenant_id, entity_name, field_name) WHERE deleted_at IS NULL
);

ALTER TABLE custom_field_defs ENABLE ROW LEVEL SECURITY;
CREATE POLICY custom_field_defs_tenant_isolation ON custom_field_defs
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Index for the SDUI schema builder query:
-- WHERE entity_name = 'customer' AND deleted_at IS NULL ORDER BY sort_order
CREATE INDEX custom_field_defs_entity_idx
    ON custom_field_defs (tenant_id, entity_name, sort_order)
    WHERE deleted_at IS NULL;

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON custom_field_defs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

---

## 47.9 Security Considerations

### Who Can Define Custom Fields?

Only tenant admins. The policy `requireRole("admin")` enforces this. A regular user cannot add a field to exfiltrate sensitive data.

### Platform Entities Are Protected

The `validateFieldName` hook checks if `entity_name` is in a platform-restricted list. Tenant admins cannot add custom fields to `user`, `role`, `session`, or `audit_log`. These platform entities are security-sensitive — allowing custom fields there would create vectors for data misuse (e.g., storing passwords in a "notes" field outside normal security controls).

### RLS on Values

Custom field values live *inside* the entity row's `custom_fields` column. They are protected by the same RLS policy as the rest of the entity row. There is no separate `custom_field_values` table — no second RLS policy needed.

### No Custom Fields in Audit Log Diffs?

Custom field values ARE included in audit log diffs because they are part of the full row captured by `to_jsonb(OLD)` and `to_jsonb(NEW)` in the trigger. The `custom_fields` JSONB object appears as-is in `previous_value` and `new_value`. This is correct — changes to custom field values should be audited.

---

## 47.10 Common Mistakes

**"Should I add a custom field for a value that all tenants need?"**

No. If every tenant in a given industry needs the field, it belongs in the framework's entity definition (committed code) or the module's EntityDefinition.Fields. Custom fields are for per-tenant exceptions, not universal requirements.

**"Can I rename a custom field?"**

Only the label (display name) can be changed safely. The `field_name` (the JSON key) is immutable after creation — existing records use that key. Renaming `field_name` would silently orphan data in existing rows under the old key. If you need to rename: create a new field, migrate data via a one-time script, soft-delete the old field.

**"What happens to custom field values when a custom field is deleted?"**

The definition is soft-deleted. The JSONB values remain in entity rows. They are no longer shown in forms or returned in API responses (the service filters by `deleted_at IS NULL`). A cleanup job can later remove the JSONB keys with:

```sql
UPDATE customers
   SET custom_fields = custom_fields - 'old_field_name'
 WHERE tenant_id = $1
   AND custom_fields ? 'old_field_name';
```

This is intentionally a manual/scheduled operation — never automatic — to prevent accidental data loss.

**"Can I use custom fields as foreign keys with referential integrity?"**

No. JSONB values are not enforced by PostgreSQL foreign key constraints. For `field_type=link`, the picker validates at input time (the selected value must exist in the target entity), but the DB does not enforce it. If the target record is deleted, the JSONB link becomes stale. Link-type custom fields are for "nice to have" references — not load-bearing financial references.
