-- Phase 9: IAM — roles table.
--
-- System roles have tenant_id IS NULL and are visible to all tenants.
-- Tenant-defined roles have tenant_id set and are scoped by RLS.
-- parent_role_id enables role inheritance (permissions accumulate upward).

CREATE TABLE roles (
    id             UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id      UUID        REFERENCES tenants(id) ON DELETE CASCADE,  -- NULL = system role
    name           VARCHAR(100) NOT NULL,
    slug           VARCHAR(100) NOT NULL,
    description    TEXT,
    is_system      BOOLEAN     NOT NULL DEFAULT FALSE,   -- TRUE = immutable; cannot be modified
    is_active      BOOLEAN     NOT NULL DEFAULT TRUE,
    parent_role_id UUID        REFERENCES roles(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- System roles: globally unique slug (tenant_id IS NULL).
CREATE UNIQUE INDEX roles_global_slug  ON roles (slug)             WHERE tenant_id IS NULL;
-- Tenant roles: unique slug per tenant.
CREATE UNIQUE INDEX roles_tenant_slug  ON roles (tenant_id, slug)  WHERE tenant_id IS NOT NULL;
CREATE INDEX        roles_tenant_id    ON roles (tenant_id)        WHERE tenant_id IS NOT NULL;

ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles FORCE ROW LEVEL SECURITY;

-- System roles (tenant_id IS NULL) visible to everyone.
-- Tenant roles visible only to their tenant.
CREATE POLICY tenant_isolation ON roles
    USING (tenant_id IS NULL OR tenant_id = current_tenant_id());

-- ── Seed built-in system roles ───────────────────────────────────────────────
-- These are seeded once at migration time and never removed.

INSERT INTO roles (id, tenant_id, name, slug, description, is_system, is_active)
VALUES
    (gen_random_uuid(), NULL, 'Tenant Admin',       'tenant-admin',
     'Full access within one tenant — manages users, roles, settings.',                 TRUE, TRUE),
    (gen_random_uuid(), NULL, 'Tenant Manager',     'tenant-manager',
     'Write access to business modules; no IAM or billing.',                            TRUE, TRUE),
    (gen_random_uuid(), NULL, 'Tenant Staff',       'tenant-staff',
     'Read-only by default; additional permissions added per custom role.',              TRUE, TRUE),
    (gen_random_uuid(), NULL, 'Tenant API Readonly','tenant-api-readonly',
     'Read-only API key role for machine-to-machine integrations.',                     TRUE, TRUE);
