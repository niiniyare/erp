-- Phase 9: IAM — permissions table.
--
-- Binds a dotted permission string to a role.
-- granted=TRUE = grant. granted=FALSE = explicit deny (highest priority).
-- tenant_id IS NULL for system-role permissions; otherwise tenant-specific.
--
-- Permission format: <module>.<resource>.<action>
-- Examples: fms.shifts.read, finance.transactions.submit, iam.users.write

CREATE TABLE permissions (
    id          UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id   UUID        REFERENCES tenants(id) ON DELETE CASCADE,  -- NULL = system permission
    role_id     UUID        NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission  VARCHAR(200) NOT NULL,
    granted     BOOLEAN     NOT NULL DEFAULT TRUE,   -- FALSE = explicit deny
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX permissions_role_id   ON permissions (role_id);
CREATE INDEX permissions_tenant_id ON permissions (tenant_id) WHERE tenant_id IS NOT NULL;
CREATE UNIQUE INDEX permissions_role_perm ON permissions (role_id, permission);

ALTER TABLE permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE permissions FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON permissions
    USING (tenant_id IS NULL OR tenant_id = current_tenant_id());

-- ── Seed permissions for built-in system roles ────────────────────────────────
-- tenant-admin: wildcard across all modules.
INSERT INTO permissions (role_id, permission, granted)
SELECT r.id, '*.*.*', TRUE
FROM roles r WHERE r.slug = 'tenant-admin' AND r.tenant_id IS NULL;

-- tenant-manager: write on all business modules, read-only on IAM/billing.
INSERT INTO permissions (role_id, permission, granted)
SELECT r.id, p.permission, TRUE
FROM roles r
CROSS JOIN (VALUES
    ('*.*.read'),
    ('*.*.write'),
    ('*.*.submit'),
    ('*.*.cancel')
) AS p(permission)
WHERE r.slug = 'tenant-manager' AND r.tenant_id IS NULL;

-- Deny IAM write for tenant-manager (managers cannot manage users/roles).
INSERT INTO permissions (role_id, permission, granted)
SELECT r.id, 'iam.*.write', FALSE
FROM roles r WHERE r.slug = 'tenant-manager' AND r.tenant_id IS NULL;

-- tenant-staff: read-only on all modules.
INSERT INTO permissions (role_id, permission, granted)
SELECT r.id, '*.*.read', TRUE
FROM roles r WHERE r.slug = 'tenant-staff' AND r.tenant_id IS NULL;

-- tenant-api-readonly: read on all modules.
INSERT INTO permissions (role_id, permission, granted)
SELECT r.id, '*.*.read', TRUE
FROM roles r WHERE r.slug = 'tenant-api-readonly' AND r.tenant_id IS NULL;
