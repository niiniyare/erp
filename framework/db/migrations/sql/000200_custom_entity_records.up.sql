-- Custom entity records: shared JSONB table for EntityTypeCustom entities.
-- All field values are stored in the `data` jsonb column.
-- The `entity_type` column acts as a discriminator (EntityDefinition.Name).
-- RLS is enforced via set_tenant_context() + tenant_id column, matching the
-- pattern used by all system entity tables.

CREATE TABLE IF NOT EXISTS custom_entity_records (
    id          uuid         NOT NULL DEFAULT uuid_generate_v4(),
    tenant_id   uuid         NOT NULL,
    entity_type text         NOT NULL,
    data        jsonb        NOT NULL DEFAULT '{}',
    created_at  timestamptz  NOT NULL DEFAULT NOW(),
    updated_at  timestamptz  NOT NULL DEFAULT NOW(),
    deleted_at  timestamptz,

    CONSTRAINT custom_entity_records_pkey PRIMARY KEY (id)
);

-- GIN index for JSONB containment and key-exists queries (@>, ?, ?|, ?&).
-- Used by CustomStore filter/search operations.
CREATE INDEX IF NOT EXISTS custom_entity_records_data_gin_idx
    ON custom_entity_records USING GIN (data);

-- Composite index for the most common query pattern: tenant + entity type,
-- optionally with soft-delete filter (WHERE deleted_at IS NULL).
CREATE INDEX IF NOT EXISTS custom_entity_records_tenant_entity_idx
    ON custom_entity_records (tenant_id, entity_type)
    WHERE deleted_at IS NULL;

-- Index to support ORDER BY created_at / updated_at within a tenant+entity scope.
CREATE INDEX IF NOT EXISTS custom_entity_records_created_at_idx
    ON custom_entity_records (tenant_id, entity_type, created_at DESC)
    WHERE deleted_at IS NULL;

-- ── Row Level Security ────────────────────────────────────────────────────────

ALTER TABLE custom_entity_records ENABLE ROW LEVEL SECURITY;

-- Application role: only rows belonging to the current tenant context.
-- Uses set_tenant_context($1) → current_tenant_id() pattern from existing tables.
CREATE POLICY tenant_isolation_policy
    ON custom_entity_records
    FOR ALL
    TO application_role
    USING (tenant_id = current_tenant_id());

-- Admin role: unrestricted access (for migrations, support tooling).
CREATE POLICY admin_full_access_policy
    ON custom_entity_records
    FOR ALL
    TO admin_role
    USING (TRUE);

-- Read-only role: SELECT only, non-deleted rows.
CREATE POLICY readonly_select_policy
    ON custom_entity_records
    FOR SELECT
    TO readonly_role
    USING (deleted_at IS NULL);

-- ── Grants ────────────────────────────────────────────────────────────────────

GRANT SELECT, INSERT, UPDATE, DELETE ON custom_entity_records TO application_role;
GRANT ALL ON custom_entity_records TO admin_role;
GRANT SELECT ON custom_entity_records TO readonly_role;
