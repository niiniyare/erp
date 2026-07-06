-- platform_organization: arbitrary-depth org tree node.
--
-- ISOLATION MODEL
-- ===============
-- platform_organization is NOT an RLS-enforced table. Organization hierarchy
-- is an application-layer authorization concern, not a database isolation
-- boundary. Tenant isolation is enforced exclusively on platform_tenant via
-- current_tenant_id(). Organization visibility is resolved by
-- OrganizationService.ResolveScope() and applied as explicit WHERE predicates
-- by the application service — never by RLS policies.
--
-- Materialized path (path column): "/root-id/parent-id/self-id/"
-- depth: redundant with path; stored for cheap depth-limit enforcement.
-- code: immutable after creation — embedded in workflow IDs and audit records.
-- type: free-form VARCHAR, not a CHECK constraint — metadata-driven.

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

-- Tenant-scoped code lookups.
CREATE INDEX IF NOT EXISTS idx_platform_organization_tenant_code
    ON platform_organization (tenant_id, code);

-- Path-prefix queries: "WHERE path LIKE '/root-id/%'".
-- text_pattern_ops makes LIKE prefix queries index-scannable.
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

-- NO RLS: organization visibility is an application concern.
-- Application services call OrganizationService.ResolveScope() and pass the
-- resulting org ID list as an explicit IN predicate. Never use RLS here.

-- platform_org_type: tenant-defined organization type registry.
-- Types are metadata-driven — tenant admins register valid types.
-- The framework never assumes predefined type values.
CREATE TABLE IF NOT EXISTS platform_org_type (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES platform_tenant(id),
    name        VARCHAR(50) NOT NULL,
    label       VARCHAR(255) NOT NULL,
    description VARCHAR(1024),
    sort_order  INTEGER NOT NULL DEFAULT 0,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT platform_org_type_name_tenant_unique UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_platform_org_type_tenant
    ON platform_org_type (tenant_id, active, sort_order);

-- NO RLS on platform_org_type for the same reason as platform_organization.

-- platform_org_assignment: user membership in one or more organizations.
-- Users may belong to multiple organizations with different roles in each.
-- primary_org: the user's home organization (at most one per user per tenant).
-- role: organization-level role name (e.g. "manager", "member", "viewer").
CREATE TABLE IF NOT EXISTS platform_org_assignment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES platform_tenant(id),
    user_id         UUID NOT NULL REFERENCES iam_user(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES platform_organization(id) ON DELETE CASCADE,
    role            VARCHAR(50) NOT NULL DEFAULT 'member',
    is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT platform_org_assignment_unique UNIQUE (tenant_id, user_id, organization_id)
);

-- Fast lookup: user's org memberships.
CREATE INDEX IF NOT EXISTS idx_platform_org_assignment_user
    ON platform_org_assignment (tenant_id, user_id);

-- Fast lookup: org members.
CREATE INDEX IF NOT EXISTS idx_platform_org_assignment_org
    ON platform_org_assignment (tenant_id, organization_id);

-- Enforce at most one primary org per user per tenant.
CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_org_assignment_primary
    ON platform_org_assignment (tenant_id, user_id)
    WHERE is_primary = TRUE;

-- NO RLS on platform_org_assignment — visibility resolved in application layer.

-- Migration compatibility note:
-- platform_org_unit and platform_branch (from earlier schema versions) are
-- superseded by platform_organization. A separate data migration script maps:
--   platform_org_unit rows → platform_organization (type preserved)
--   platform_branch rows   → platform_organization (type = 'branch')
