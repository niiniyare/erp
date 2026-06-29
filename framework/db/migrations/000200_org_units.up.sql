-- ============================================================
-- 000200_org_units — hierarchical org unit tree
-- ============================================================

CREATE TABLE IF NOT EXISTS org_units (
    id         uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id  uuid        NOT NULL DEFAULT current_tenant_id(),
    parent_id  uuid        REFERENCES org_units (id) ON DELETE RESTRICT,
    name       text        NOT NULL,
    code       text,
    path       text        NOT NULL,  -- materialized ltree-style path for descendant queries
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS org_units_tenant_path_idx ON org_units (tenant_id, path);
CREATE INDEX IF NOT EXISTS org_units_parent_idx      ON org_units (parent_id);

SELECT apply_tenant_rls('org_units');
