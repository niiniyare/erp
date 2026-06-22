### Chapter 10 — Custom Fields — Runtime Schema Extension

Custom fields allow tenant administrators to extend both system entities and custom entities with additional attributes at runtime, without database migrations or code deployments. A petrol station operator can add a `nema_compliance_note` text field to `ServiceRequest`. A fleet management company can add a `fleet_tier` selection field to `Customer`. These extensions are defined through the admin UI or the management API, stored as `CustomFieldDef` records, and applied automatically to API responses, form pages, and validation pipelines. This chapter documents the storage model, definition lifecycle, validation, querying, SDUI integration, and promotion path for custom fields.

---

#### 10.1. The Custom Field Model

##### 10.1.1. How custom fields extend both system entities and custom entities

Custom fields extend entities in two different storage locations depending on the entity type. For system entities (whose data lives in typed SQL tables), custom field values are stored in a dedicated `custom_fields jsonb` column appended to the entity's table. For custom entities (which are entirely JSONB-backed), custom field definitions simply add keys to the existing JSONB document structure.

In both cases, the `CustomFieldDef` records are the authoritative schema for these extensions. The framework reads `CustomFieldDef` records at tenant boot and merges their field descriptions into the in-memory `EntityDefinition` for the targeted entity. From the perspective of any code that interacts with the entity through `EntityRepository`, a custom field is indistinguishable from a field declared in Go code — it appears in the `EntityRecord`, participates in filtering via the Filter DSL's JSONB path predicates, and is subject to the same validation pipeline.

##### 10.1.2. Storage — the `custom_fields JSONB` column pattern on system entities

When a module developer declares that a system entity supports custom field extension, the entity's migration includes a `custom_fields jsonb DEFAULT '{}'::jsonb` column. This column stores the values for all custom fields defined by the tenant for that entity, keyed by the custom field's stable key.

```sql
-- Example: custom_fields column on a system entity table
ALTER TABLE service_requests
    ADD COLUMN custom_fields jsonb NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX idx_service_requests_custom_fields
    ON service_requests USING gin (custom_fields);
```

The GIN index enables efficient filtering on custom field values. When a tenant defines a `CustomFieldDef` for `ServiceRequest` with key `vehicle_type`, the value for a specific record is stored as `custom_fields->'vehicle_type'` and queried as `custom_fields->>'vehicle_type' = 'truck'`.

The `custom_fields` column is declared on a system entity by adding the `entity.AllowCustomFields()` option to the `EntityDefinition`:

```go
// Example: Declaring custom field support on a system entity
var ServiceRequestDefinition = entity.Define("ServiceRequest",
    entity.Fields( /* ... */ ),
    entity.AllowCustomFields(), // adds custom_fields JSONB column to migration
)
```

##### 10.1.3. The `CustomFieldDef` system entity — metadata table

`CustomFieldDef` is a system entity defined in the framework core. Each row describes one custom field: the target entity name, a stable key, a human-readable label, the field type, validation rules, display order, and UI placement metadata. `CustomFieldDef` records are tenant-scoped — one tenant's custom fields do not affect any other tenant's data or schema.

The `CustomFieldDef` entity has these primary fields:

```
entity_name     — the target EntityDefinition name ("ServiceRequest", "Customer")
field_key       — stable snake_case identifier ("vehicle_type", "fleet_tier")
label           — human-readable display name ("Vehicle Type", "Fleet Tier")
field_type      — one of the supported types (see §10.2.4)
is_required     — boolean
options         — JSON array of option strings for Select/MultiSelect
regex_pattern   — optional validation regex
max_length      — optional max character length
display_order   — integer controlling form placement
section_label   — optional section heading for grouping in the form
active          — boolean, false soft-deletes the field without removing data
```

##### 10.1.4. Scope — custom fields are per-tenant, per-entity

Every `CustomFieldDef` belongs to exactly one tenant and applies to exactly one entity type. Tenant A's `vehicle_type` field on `ServiceRequest` is completely independent of Tenant B's `vehicle_type` field on `ServiceRequest`, even if they have the same key. The `EntityRegistry`'s per-tenant custom entity layer (§2.5.3) ensures that each tenant's merged `EntityDefinition` reflects only that tenant's `CustomFieldDef` records.

This per-tenant scope means that a query executed in Tenant A's context that filters on `custom_fields.vehicle_type` will never return or affect Tenant B's records, regardless of whether Tenant B has a custom field with the same key.

---

#### 10.2. Defining Custom Fields

##### 10.2.1. Via the admin UI — field type, label, options, validation rules

Tenant administrators define custom fields through the admin UI at `/app/settings/custom-fields`. The UI presents a form for each `CustomFieldDef` attribute: entity selector, field type dropdown, label input, options input (for Select/MultiSelect), and validation rule configuration. On save, the UI posts to `POST /api/v1/custom-field-defs`, which creates the `CustomFieldDef` record and triggers a per-tenant `EntityRegistry` cache invalidation.

The new field becomes available immediately after the cache invalidation propagates — typically within one second. No server restart, no migration, and no code change is required.

##### 10.2.2. Via the API — `POST /api/v1/custom-field-defs`

The management API accepts the same payload as the admin UI form. Programmatic definition is appropriate for: automated tenant onboarding scripts that provision a standard set of custom fields for all new tenants, integration tests that need specific custom fields on test tenants, and bulk field creation tools.

```go
// Example: Creating a custom field definition via the management API
payload := map[string]any{
    "entity_name":    "ServiceRequest",
    "field_key":      "vehicle_type",
    "label":          "Vehicle Type",
    "field_type":     "select",
    "options":        []string{"sedan", "suv", "truck", "motorcycle", "tuk_tuk"},
    "is_required":    false,
    "display_order":  10,
    "section_label":  "Vehicle Details",
}
// POST /api/v1/custom-field-defs with this payload
```

The API validates that `entity_name` refers to an entity that has `AllowCustomFields()` declared and that `field_key` does not conflict with an existing field (either system field or existing custom field) on that entity. Attempting to create a custom field with the same key as a system field (e.g., `field_key = "status"`) returns a 422 validation error.

##### 10.2.3. Via fixture files — for module developers shipping pre-configured fields

Module developers who want to ship a default set of custom fields with their module use fixture files loaded by `awo entity seed --module={module_name}`. The fixture file is a JSON array of `CustomFieldDef` payloads. The seed command creates the records for the target tenant, skipping records whose `field_key` already exists for that entity on that tenant (idempotent upsert).

```json
[
  {
    "entity_name": "ServiceRequest",
    "field_key": "service_bay_number",
    "label": "Service Bay",
    "field_type": "select",
    "options": ["Bay 1", "Bay 2", "Bay 3", "Bay 4"],
    "display_order": 20,
    "section_label": "Workshop Details"
  },
  {
    "entity_name": "ServiceRequest",
    "field_key": "technician_notes",
    "label": "Technician Notes",
    "field_type": "long_text",
    "display_order": 30,
    "section_label": "Workshop Details"
  }
]
```

##### 10.2.4. Supported field types for custom fields — subset of the full field type list

Custom fields support a subset of the field types available to system entities. The excluded types are those that require schema-level support that the JSONB storage model cannot provide:

**Supported:** `data`, `small_text`, `long_text`, `int`, `float`, `currency`, `bool`, `date`, `datetime`, `select`, `multi_select`, `link`, `attach`, `attach_image`.

**Not supported:** `table` (child table editing requires a separate entity table), `dynamic_link` (polymorphic FK integrity requires application-layer guards that are not yet wired for JSONB-backed custom fields), `uuid` (UUIDs as custom fields can be approximated by `data` with a regex validator).

The `link` type for custom fields works with a reduced feature set: the amis page builder renders it as a search-as-you-type picker, but referential integrity is validated at the application layer (existence check in the async validator) rather than at the database level. There is no FK constraint on the JSONB column.

---

#### 10.3. Validation Rules on Custom Fields

##### 10.3.1. Required, max length, regex — declared in `CustomFieldDef`

Custom field validation rules are declared in the `CustomFieldDef` record rather than in Go code. The framework's validation pipeline reads the active `CustomFieldDef` records at tenant boot and dynamically constructs the validator chain for each custom field.

The available declarative validators are:
- `is_required: true` — the field must be present and non-empty.
- `max_length: N` — for `data` and `small_text` fields, maximum character count.
- `regex_pattern: "pattern"` — a Go-compatible regular expression; the value must match.
- For `select` and `multi_select`: the value must be one of the declared `options`.
- For `currency`, `int`, `float`: `min_value` and `max_value` bounds.

These rules are evaluated at exactly the same point in the validation pipeline as static field validators declared in Go code (§5.3): after type coercion, before any `before_save` hook.

##### 10.3.2. How custom field validators are loaded and executed

During tenant boot, the framework loads all active `CustomFieldDef` records and constructs a per-field validator function from the declared rules. These validator functions are stored in the per-tenant `EntityRegistry` slice alongside the `EntityDefinition` extension they belong to. When the validation pipeline runs for a record of the target entity, it iterates over both the system field validators (compiled from Go declarations) and the custom field validators (constructed from `CustomFieldDef` records) and executes them in the declared order (`display_order` determines custom field execution order).

When a `CustomFieldDef` is created or modified via the API or admin UI, the validator cache for that tenant is invalidated. The next request for any entity of the affected type rebuilds the validator chain from the updated `CustomFieldDef` records.

##### 10.3.3. Validation error reporting — same field-level error format as system fields

Custom field validation errors are reported in exactly the same format as system field errors: `{ name: "field_key", errors: ["message"] }`. The `name` is the `field_key` of the `CustomFieldDef` record. The amis form component highlights the corresponding custom field input when it receives an error with a matching `name`. There is no special handling or distinct error shape for custom field failures.

---

#### 10.4. Querying Custom Fields

##### 10.4.1. JSONB path predicates in the Filter DSL

Custom field values are accessed in `Filter` predicates using the JSONB path notation: `custom_fields.{field_key}`. The Filter DSL translates this to the appropriate PostgreSQL JSONB expression.

```go
// Example: Filtering on custom field values
tc, err := tenant.FromContext(ctx)
if err != nil {
    return err
}
repo, err := entity.Resolve(ctx, tc, "ServiceRequest")
if err != nil {
    return err
}
filter := entity.NewFilter().
    Eq("custom_fields.vehicle_type", "truck").
    IsNotNull("custom_fields.service_bay_number")
records, _, err := repo.Query(ctx, filter)
if err != nil {
    return fmt.Errorf("querying custom field filter: %w", err)
}
```

The `Eq` predicate on a JSONB path generates `custom_fields->>'vehicle_type' = 'truck'`. Numeric comparison predicates generate a cast: `(custom_fields->>'credit_limit')::numeric >= 50000`. The framework handles the type cast automatically based on the field type declared in the `CustomFieldDef`.

##### 10.4.2. GIN index strategy for custom fields — which paths to index

The GIN index created on the `custom_fields` column (§10.1.2) supports the `@>` containment operator efficiently. However, the path-extraction operators (`->>`, `->`) used by the Filter DSL's comparison predicates require a more specific index strategy to avoid sequential scans.

For `field_key` values that are frequently used in filters, declare a functional index on the extracted path:

```sql
-- Example: Functional index for a high-frequency custom field filter
CREATE INDEX idx_service_requests_custom_vehicle_type
    ON service_requests ((custom_fields->>'vehicle_type'));

CREATE INDEX idx_service_requests_custom_bay_number
    ON service_requests ((custom_fields->>'service_bay_number'));
```

These indexes are not generated automatically because the framework cannot know which custom fields tenants will filter on frequently. Tenant-specific functional indexes should be added via a custom migration for tenants that report slow custom field queries. The `awo entity migrate --diff` command with a custom Atlas schema annotation can generate these targeted indexes.

##### 10.4.3. Performance characteristics — when JSONB queries are acceptable, when to promote to a column

JSONB filtering is acceptable for custom fields when: the table has fewer than roughly 100,000 rows per tenant, the query is not in a hot path (executed on every page load or on every API call from a batch process), and the filter selectivity is moderate (the custom field filter eliminates more than 10% of rows, making the sequential scan less likely). For most ERP use cases involving tenant-specific attributes, these conditions hold.

JSONB filtering becomes unacceptable when: the field is used in a financial report query that must aggregate millions of rows, the field participates in a JOIN with another entity's column (impossible without a functional index), or the field must be sorted and the sort order needs to be stable across large result sets under concurrent writes. In these cases, the custom field should be promoted to a system field (§10.6.5).

---

#### 10.5. Surfacing Custom Fields in the SDUI Layer

##### 10.5.1. Page builder reads `CustomFieldDef` records at render time

Page builder functions receive the full merged `EntityDefinition` — including fields contributed by `CustomFieldDef` records — via the `entity.PageContext`. When a page builder calls `ctx.EntityDef().Fields()`, the returned list includes both system-declared fields and active custom fields for the current tenant. Page builders do not need to query `CustomFieldDef` directly; the framework has already merged them.

```go
// Example: Page builder using the merged field list including custom fields
func BuildServiceRequestFormPage(ctx entity.PageContext) (entity.PageDefinition, error) {
    fields := ctx.EntityDef().Fields() // includes custom fields
    var formFields []entity.AmisFormField
    for _, f := range fields {
        formFields = append(formFields, entity.FieldToAmisComponent(f))
    }
    return entity.NewFormPage().
        Title("Service Request").
        API("/api/v1/service-requests").
        Fields(formFields).
        Build(), nil
}
```

`entity.FieldToAmisComponent` converts a field descriptor (system or custom) into the appropriate amis component configuration. For a `select` custom field, it generates an amis `select` component with the options from `CustomFieldDef.options`. For a `link` custom field, it generates an amis picker component wired to the linked entity's list endpoint.

##### 10.5.2. Automatic form field injection — ordering and section placement

Custom fields are injected into the form at the position specified by `display_order` relative to other custom fields in the same section. System fields occupy fixed positions determined by the page builder function. Custom fields are appended after the system fields in the section specified by `section_label`, or in a default "Additional Fields" section if no `section_label` is set.

The framework's `entity.FieldToAmisComponent` function checks `section_label` and groups custom fields under amis `Fieldset` components. A tenant that defines five custom fields with `section_label = "Compliance"` will see those five fields grouped under a "Compliance" fieldset in the form, separate from the system fields.

##### 10.5.3. Tenant control over field placement via UI metadata on `CustomFieldDef`

In addition to `display_order` and `section_label`, `CustomFieldDef` records support a `ui_metadata` JSONB column for storing additional amis component hints: placeholder text, help text, conditional visibility expressions, and field width in multi-column layouts.

```json
{
  "placeholder": "e.g. KCC 123A",
  "help_text": "Enter the full Kenya vehicle registration number",
  "visible_when": "vehicle_type != 'motorcycle'",
  "column_span": 6
}
```

The `visible_when` expression uses amis's built-in expression syntax. It is included verbatim in the generated amis form field definition; the amis client evaluates it at render time. The framework does not validate `visible_when` expressions at definition time — an invalid expression will cause the field to always be visible (the amis default when a visibility expression fails to evaluate).

---

#### 10.6. Custom Field Lifecycle

##### 10.6.1. Adding a custom field — no migration required for system entities

Adding a `CustomFieldDef` record for a system entity that has `AllowCustomFields()` declared requires no database migration. The `custom_fields jsonb` column already exists on the table; adding a new key to the JSONB document does not require a schema change. The new field appears in API responses and forms immediately after the `CustomFieldDef` record is created and the per-tenant `EntityRegistry` cache is invalidated.

For custom entities (where all fields are JSONB-backed), adding a new `CustomFieldDef` is similarly migration-free. The custom entity's JSONB document grows a new key on first use.

##### 10.6.2. Renaming a custom field — key stability vs label changes

The `field_key` of a `CustomFieldDef` is immutable after the first record with a value for that key is created. Changing the `field_key` would orphan all existing data stored under the old key. The correct renaming procedure is: add a new `CustomFieldDef` with the new key, run a data migration that copies values from the old key to the new key in the JSONB column, verify the data migration is complete, and then mark the old `CustomFieldDef` as `active: false`.

Changing the `label` (the human-readable display name) is safe at any time and takes effect immediately. The `label` is a presentation concern; it does not affect storage or filtering.

##### 10.6.3. Changing a field type — data migration implications

Changing the `field_type` of an existing `CustomFieldDef` from `data` to `currency`, for example, is a dangerous operation if existing records have values that are not valid for the new type. The framework does not perform automatic type coercion of existing JSONB values when a `CustomFieldDef`'s type is changed. Existing values remain as-is in the JSONB column; only new writes are validated against the new type.

The safe procedure for a type change is: create a new `CustomFieldDef` with the new type, migrate existing values to the new field, mark the old field inactive. For `select` fields where option sets are growing (adding new options), direct modification of the `options` array on the existing `CustomFieldDef` is safe — new options become valid for new writes, and existing values with old options remain valid.

> **Danger:** Never change a `CustomFieldDef`'s `field_type` directly on an entity that has existing records with values for that field. Changing the type without migrating the data will cause validation failures for every existing record that is subsequently updated, because the old values fail the new type's validator. This can make existing records effectively un-editable until the data is corrected.

##### 10.6.4. Deprecating and removing a custom field — soft removal first

To deprecate a custom field, set `active: false` on its `CustomFieldDef` record. An inactive custom field: stops appearing in forms and API responses, stops being validated on new writes, and stops being included in the merged `EntityDefinition`. Existing data remains in the `custom_fields` JSONB column under the field's key — it is not deleted.

To permanently remove a custom field and its data, execute a data migration that removes the key from the JSONB column for all affected records, then delete the `CustomFieldDef` record:

```sql
-- Example: Removing a deprecated custom field from all records
UPDATE service_requests
    SET custom_fields = custom_fields - 'old_field_key'
    WHERE custom_fields ? 'old_field_key';
```

This migration must be run as a reviewed Atlas migration file (§11), not as an ad-hoc SQL command, so it is version-controlled and applied consistently across all tenant schemas.

##### 10.6.5. Promoting a custom field to a system field — when and how

Promotion from custom field to system field is warranted when any of the conditions in §2.3.5 become true. The procedure is:

**Step 1 — Define the system field.** Add the field to the Go `EntityDefinition` with the same name and a compatible type. Generate and review the Atlas migration that adds the typed column to the table. Do not apply the migration yet.

**Step 2 — Dual-write phase.** Deploy a code change that writes to both the new typed column and the old JSONB key on every create and update. This ensures no data is lost during the transition window.

**Step 3 — Backfill migration.** Apply the Atlas migration (adding the typed column), then run a backfill migration that copies values from `custom_fields->>'{field_key}'` to the new typed column for all existing records. Review and test this migration against a staging dataset before production.

**Step 4 — Cutover.** Remove the dual-write code. Update any filters that use the JSONB path notation to use the typed column name. Mark the `CustomFieldDef` as `active: false`.

**Step 5 — Cleanup.** After confirming the new column is correct and all reads use it, remove the JSONB key from existing records (§10.6.4 data migration) and delete the `CustomFieldDef` record.

> **Note:** Steps 2 and 3 can be reversed (backfill first, then dual-write) if downtime is acceptable. The dual-write-first approach is safer for zero-downtime production systems because it ensures no write is lost between migration application and code deployment.

---

#### Chapter summary

Chapter 10 documents the custom field system across its full lifecycle: the `custom_fields JSONB` column storage model for system entities (§10.1.2), the `CustomFieldDef` metadata entity that drives runtime schema extension (§10.1.3), the three definition paths (admin UI, management API, fixture files) and the subset of supported field types (§10.2), the dynamic validator chain constructed from `CustomFieldDef` records (§10.3), JSONB path predicates and GIN index strategy for efficient filtering (§10.4), automatic form field injection via the merged `EntityDefinition` in page builders (§10.5), and the complete lifecycle management including the five-step system field promotion procedure (§10.6). The three most critical concepts are the `AllowCustomFields()` declaration required on system entities before custom fields can be added (§10.1.2), the `field_key` immutability constraint after data is written (§10.6.2), and the danger of changing `field_type` without migrating existing data (§10.6.3 Danger callout).

**Next chapters to read:**

- §11 — Database Migrations (the Atlas migration pipeline that creates the `custom_fields` column and any targeted functional indexes for custom field performance; the promotion procedure in §10.6.5 generates new migration files)
- §5 — Field System (the full field type reference that determines which types are available for custom fields and what their storage, validation, and serialisation behaviour is)
- §21 — Server-Driven UI Philosophy (the page builder pipeline that surfaces custom fields in forms, including the `entity.FieldToAmisComponent` helper and the `section_label`/`ui_metadata` placement mechanism)
