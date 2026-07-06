-- platform_organization: arbitrary-depth org tree node.
-- Materialized path (path column) enables O(1) ancestor/descendant queries.
-- path format: "/root-id/parent-id/self-id/" (UUID strings, no braces).
-- depth is redundant with path but stored for cheap depth-limit enforcement.
-- code is immutable after creation — embedded in workflow IDs and audit records.

CREATE TABLE IF NOT EXISTS platform_organization (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES platform_tenant(id),
    parent_id   UUID REFERENCES platform_organization(id),
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(50)  NOT NULL,
    type        VARCHAR(50),
    path        VARCHAR(4096) NOT NULL DEFAULT '/',
    depth       INTEGER NOT NULL DEFAULT 0,
    description VARCHAR(1024),
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    custom_fields JSONB NOT NULL DEFAULT '{}',

    CONSTRAINT platform_organization_code_tenant_unique UNIQUE (tenant_id, code),
    CONSTRAINT platform_organization_depth_nonneg CHECK (depth >= 0)
);

-- B-tree index for tenant-scoped code lookups.
CREATE INDEX IF NOT EXISTS idx_platform_organization_tenant_code
    ON platform_organization (tenant_id, code);

-- Path-prefix queries: "WHERE path LIKE '/root-id/%'" benefit from this index.
CREATE INDEX IF NOT EXISTS idx_platform_organization_path
    ON platform_organization USING btree (path text_pattern_ops);

-- Parent traversal.
CREATE INDEX IF NOT EXISTS idx_platform_organization_parent
    ON platform_organization (parent_id)
    WHERE parent_id IS NOT NULL;

-- Active filter (most queries include active = true).
CREATE INDEX IF NOT EXISTS idx_platform_organization_tenant_active
    ON platform_organization (tenant_id, active);

-- GIN index for custom_fields JSONB queries.
CREATE INDEX IF NOT EXISTS idx_platform_organization_custom_fields
    ON platform_organization USING GIN (custom_fields);

-- RLS: every tenant-scoped table must have row-level security.
ALTER TABLE platform_organization ENABLE ROW LEVEL SECURITY;
ALTER TABLE platform_organization FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON platform_organization
    USING (tenant_id = current_tenant_id());

-- Migration compatibility note:
-- The old platform_org_unit and platform_branch tables are superseded by this
-- table. A data migration script (separate from schema migrations) should
-- transform existing rows if upgrading from a version that had those tables.
-- platform_org_unit rows map to platform_organization with type preserved.
-- platform_branch rows map to platform_organization with type = 'branch'.
