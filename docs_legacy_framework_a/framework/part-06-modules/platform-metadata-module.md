> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "Platform Module — Metadata & Custom Fields"
part: "Part VI — Platform Entities"
chapter: 47
section: "platform-metadata-module"
related:
  - "[Chapter 41: Platform Modules Overview](./platform-module-overview.md)"
  - "[Chapter 42: Tenant Module](./platform-tenant-module.md)"
  - "[Chapter 45: Settings & Configuration Module](./platform-settings-module.md)"
---

# Chapter 47 — Metadata & Custom Fields Module

> **Who should read this?** Developers enabling per-tenant schema extension, tenant admins adding fields to standard entities, and anyone building SDUI forms that surface custom data.

---

## 47.1 Why Custom Fields?

Awo's `EntityDefinition` is defined in Go at compile time. This gives strong typing, IDE support, and zero-cost query generation — but it means every tenant sees the same schema. Real businesses need per-tenant extensions:

- A logistics company needs a "delivery zone" field on Customer.
- An NGO needs a "donor category" on Contact.
- A manufacturing firm needs "machine serial number" on Asset.

The Metadata module provides a runtime extension layer: **custom field definitions** stored in the database, and **custom field values** stored per record per tenant. Core schema stays clean; tenant-specific data lives in the metadata tables.

---

## 47.2 Design Principles

1. **No schema migrations for new custom fields.** All values stored in a JSONB column on `entity_metadata`.
2. **Type-safe at write time.** A `validateCustomFieldValue` hook rejects values that don't match the declared type.
3. **Searchable.** JSONB GIN index enables `entity_metadata.data @> '{"zone": "Westlands"}'` queries.
4. **SDUI auto-rendered.** Custom fields appear on entity forms automatically, after the built-in fields.
5. **Tenant-isolated.** Custom field definitions are per-tenant; no cross-tenant leakage.

---

## 47.3 EntityDefinition: CustomFieldDef

```go
// internal/platform/metadata/definition.go
package metadata

import "awo.so/framework/definition"

// CustomFieldDef describes the schema for one custom field on one entity type.
var CustomFieldDef = definition.EntityDefinition{
    Name:        "custom_field_def",
    Label:       "Custom Field",
    Description: "A tenant-defined extension field for a standard entity type.",
    Module:      "Platform",
    Table:       "custom_field_defs",
    OrgScope:    org.ScopeLevelTenant,

    Fields: []*definition.FieldDef{
        definition.Field("entity_name").
            OfType(definition.FieldTypeData).
            WithLabel("Entity").
            RequiredField().
            ImmutableField().
            SearchableField(),

        definition.Field("key").
            OfType(definition.FieldTypeData).
            WithLabel("Field Key").
            RequiredField().
            ImmutableField(). // referenced in stored data; changing breaks lookups
            SearchableField(),

        definition.Field("label").
            OfType(definition.FieldTypeData).
            WithLabel("Display Label").
            RequiredField(),

        definition.Field("description").
            OfType(definition.FieldTypeSmallText).
            WithLabel("Description"),

        definition.Field("field_type").
            OfType(definition.FieldTypeSelect).
            WithLabel("Type").
            WithOptions("string", "int", "bool", "date", "select", "multiselect", "json").
            RequiredField().
            ImmutableField(),

        definition.Field("options").
            OfType(definition.FieldTypeJSON).
            WithLabel("Options").
            WithDescription("For select/multiselect: JSON array of {value, label} objects."),

        definition.Field("default_value").
            OfType(definition.FieldTypeSmallText).
            WithLabel("Default Value"),

        definition.Field("required").
            OfType(definition.FieldTypeBool).
            WithLabel("Required").
            WithDefault(false),

        definition.Field("sort_order").
            OfType(definition.FieldTypeInt).
            WithLabel("Sort Order").
            WithDefault(0),
    },

    Hooks: []definition.HookDef{
        definition.BeforeValidateHook("validateKey", definition.OpCreate, validateCustomFieldKey),
    },

    Policies: []definition.PolicyDef{
        {Ops: definition.OpCreate, Fn: requireRole("tenant_admin")},
        {Ops: definition.OpRead,   Fn: allowWithinTenant},
        {Ops: definition.OpUpdate, Fn: requireRole("tenant_admin")},
        {Ops: definition.OpDelete, Fn: requireRole("tenant_admin")},
    },

    Audited: true,
}
```

### Key Constraints

- `key` is immutable: once data is stored under `"zone"`, renaming the key would orphan existing values. Renaming is a new-key + migration workflow.
- `UNIQUE (tenant_id, entity_name, key)` enforced at DB level — one definition per key per entity per tenant.

---

## 47.4 EntityDefinition: EntityMetadata

```go
// EntityMetadata stores the actual custom field values for one entity record.
var EntityMetadataDef = definition.EntityDefinition{
    Name:        "entity_metadata",
    Label:       "Entity Metadata",
    Description: "Custom field values attached to a specific entity record.",
    Module:      "Platform",
    Table:       "entity_metadata",
    OrgScope:    org.ScopeLevelTenant,

    Fields: []*definition.FieldDef{
        definition.Field("entity_name").
            OfType(definition.FieldTypeData).
            WithLabel("Entity").
            RequiredField().
            ImmutableField(),

        definition.Field("entity_id").
            OfType(definition.FieldTypeUUID).
            WithLabel("Record ID").
            RequiredField().
            ImmutableField(),

        definition.Field("data").
            OfType(definition.FieldTypeJSON).
            WithLabel("Data").
            RequiredField(),
    },

    Hooks: []definition.HookDef{
        definition.BeforeValidateHook("validateFields", definition.OpWrite, validateCustomFieldValues),
    },

    Policies: []definition.PolicyDef{
        {Ops: definition.OpCreate, Fn: allowWithinTenant},
        {Ops: definition.OpRead,   Fn: allowWithinTenant},
        {Ops: definition.OpUpdate, Fn: allowWithinTenant},
        {Ops: definition.OpDelete, Fn: requireRole("tenant_admin")},
    },

    Audited: true,
}
```

---

## 47.5 The `validateCustomFieldValues` Hook

Before any save to `entity_metadata`, the hook loads all `custom_field_defs` for the entity type and validates each key in `data`:

```go
// internal/platform/metadata/hooks.go
func validateCustomFieldValues(ctx context.Context, m *definition.Mutation) error {
    if m.Op == definition.OpDelete {
        return nil
    }
    entityName := m.After.Get("entity_name").(string)
    raw        := m.After.Get("data")

    data, ok := raw.(map[string]any)
    if !ok {
        return definition.NewFieldError("data", "must be a JSON object")
    }

    defs, err := loadFieldDefs(ctx, m.TenantID, entityName)
    if err != nil {
        return err
    }

    defMap := make(map[string]*CustomFieldDef, len(defs))
    for _, d := range defs {
        defMap[d.Key] = d
    }

    // Check required fields.
    for _, d := range defs {
        if d.Required {
            if _, ok := data[d.Key]; !ok {
                return definition.NewFieldError("data."+d.Key, "required")
            }
        }
    }

    // Validate types for provided values.
    for key, val := range data {
        fd, ok := defMap[key]
        if !ok {
            return definition.NewFieldError("data."+key, "unknown custom field")
        }
        if err := validateType(key, val, fd.FieldType); err != nil {
            return err
        }
    }
    return nil
}

func validateType(key string, val any, fieldType string) error {
    switch fieldType {
    case "int":
        switch val.(type) {
        case float64, int, int64:
        default:
            return definition.NewFieldError("data."+key, "must be an integer")
        }
    case "bool":
        if _, ok := val.(bool); !ok {
            return definition.NewFieldError("data."+key, "must be true or false")
        }
    case "date":
        s, ok := val.(string)
        if !ok {
            return definition.NewFieldError("data."+key, "must be a date string")
        }
        if _, err := time.Parse("2006-01-02", s); err != nil {
            return definition.NewFieldError("data."+key, "must be YYYY-MM-DD")
        }
    case "select":
        // Validate value is one of the declared options — loaded separately if needed.
    }
    return nil
}
```

---

## 47.6 The Metadata Service

Application code and hooks read custom fields through `metadata.Service`, not directly from the table:

```go
// internal/platform/metadata/service.go
type Service struct {
    store persistence.TenantStore
    cache *redis.Client
}

// Get returns the custom field data for one record. Returns empty map if none.
func (s *Service) Get(ctx context.Context, tenantID uuid.UUID, entityName string, entityID uuid.UUID) (map[string]any, error) {
    cacheKey := fmt.Sprintf("meta:%s:%s:%s", tenantID, entityName, entityID)
    if v, err := s.cache.Get(ctx, cacheKey).Bytes(); err == nil {
        var m map[string]any
        _ = json.Unmarshal(v, &m)
        return m, nil
    }

    store, err := s.store.ForEntity(ctx, tenantID, "entity_metadata")
    if err != nil {
        return nil, err
    }
    // Find by entity_name + entity_id filter.
    page, err := store.List(ctx, persistence.ListOptions{
        Filter: map[string]any{
            "entity_name": entityName,
            "entity_id":   entityID,
        },
        Limit: 1,
    })
    if err != nil {
        return nil, err
    }
    if len(page.Records) == 0 {
        return map[string]any{}, nil
    }

    data := page.Records[0].Get("data").(map[string]any)
    b, _ := json.Marshal(data)
    s.cache.Set(ctx, cacheKey, b, 5*time.Minute)
    return data, nil
}

// Set writes (upsert) custom field data for one record.
func (s *Service) Set(ctx context.Context, tenantID uuid.UUID, entityName string, entityID uuid.UUID, data map[string]any) error {
    // Uses INSERT ... ON CONFLICT (tenant_id, entity_name, entity_id) DO UPDATE SET data = ...
    // Implemented via BulkUpdate or a raw upsert query.
    cacheKey := fmt.Sprintf("meta:%s:%s:%s", tenantID, entityName, entityID)
    defer s.cache.Del(ctx, cacheKey)
    return s.upsert(ctx, tenantID, entityName, entityID, data)
}
```

---

## 47.7 SDUI Integration

The SDUI layer queries `custom_field_defs` at page-render time and appends controls after the built-in fields:

```go
// framework/sdui/amis/custom_fields.go
func AppendCustomControls(def *definition.EntityDefinition, tenantID uuid.UUID, svc *metadata.Service) []any {
    defs, _ := svc.ListFieldDefs(ctx, tenantID, def.Name)

    controls := make([]any, 0, len(defs))
    for _, fd := range defs {
        controls = append(controls, customFieldControl(fd))
    }
    return controls
}

func customFieldControl(fd *CustomFieldDef) map[string]any {
    base := map[string]any{
        "name":     "metadata." + fd.Key, // namespaced to avoid clash with built-in fields
        "label":    fd.Label,
        "required": fd.Required,
    }
    switch fd.FieldType {
    case "string":
        base["type"] = "input-text"
    case "int":
        base["type"] = "input-number"
        base["precision"] = 0
    case "bool":
        base["type"] = "switch"
    case "date":
        base["type"] = "input-date"
        base["format"] = "YYYY-MM-DD"
    case "select":
        base["type"] = "select"
        base["options"] = fd.Options
    case "multiselect":
        base["type"] = "select"
        base["multiple"] = true
        base["options"] = fd.Options
    case "json":
        base["type"] = "json-editor"
    }
    return base
}
```

The API handler for entity forms receives `metadata.*` fields in the body and routes them to `metadata.Service.Set` rather than the entity store. This separation keeps the entity table clean.

---

## 47.8 Migration

```sql
-- internal/platform/metadata/migrations/20240101000001_create_custom_field_defs.up.sql

CREATE TABLE custom_field_defs (
    id           uuid    PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id    uuid    NOT NULL REFERENCES tenants(id),
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    updated_at   timestamptz NOT NULL DEFAULT NOW(),
    entity_name  text    NOT NULL,
    key          text    NOT NULL,
    label        text    NOT NULL,
    description  text,
    field_type   text    NOT NULL CHECK (field_type IN ('string','int','bool','date','select','multiselect','json')),
    options      jsonb,
    default_value text,
    required     boolean NOT NULL DEFAULT false,
    sort_order   int     NOT NULL DEFAULT 0,
    UNIQUE (tenant_id, entity_name, key)
);

ALTER TABLE custom_field_defs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON custom_field_defs
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- internal/platform/metadata/migrations/20240101000002_create_entity_metadata.up.sql

CREATE TABLE entity_metadata (
    id          uuid        PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   uuid        NOT NULL REFERENCES tenants(id),
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW(),
    entity_name text        NOT NULL,
    entity_id   uuid        NOT NULL,
    data        jsonb       NOT NULL DEFAULT '{}',
    UNIQUE (tenant_id, entity_name, entity_id)
);

-- GIN index for JSONB containment queries
CREATE INDEX entity_metadata_data_gin ON entity_metadata USING gin(data);
CREATE INDEX entity_metadata_entity_idx ON entity_metadata (tenant_id, entity_name, entity_id);

ALTER TABLE entity_metadata ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON entity_metadata
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

---

## 47.9 FAQ

**Q: Can custom fields be indexed for fast filtering?**
A: The GIN index on `data` supports containment queries (`data @> '{"zone":"Westlands"}'`) efficiently. For exact equality on one key used frequently, add a functional index: `CREATE INDEX ON entity_metadata ((data->>'zone')) WHERE entity_name = 'customer'`.

**Q: What happens to metadata when the entity record is deleted?**
A: The `entity_metadata` row is not automatically deleted — there is no FK to the entity table (which may live in a different module). A `BeforeDeleteHook` on each auditable entity should call `metadata.Service.Delete` to clean up. Alternatively, a nightly GC job prunes orphaned rows.

**Q: Can I query entity records by custom field value?**
A: Yes, but it requires a JOIN. The List API for the entity does not yet support metadata filters out of the box. Add a `metadata_filter` query param handled by a custom list handler that joins `entity_metadata` on `entity_name + entity_id`.

**Q: Are custom fields included in API responses?**
A: Not by default. Call `metadata.Service.Get` in an `AfterReadHook` (or in the response builder) to embed a `metadata` key in the record map. This is opt-in per entity to avoid unnecessary DB hits on list endpoints.
