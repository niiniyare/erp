-- ------------------------------------------------------------------------------------------------
-- ROLE_ASSIGNMENTS
-- ------------------------------------------------------------------------------------------------
-- Role assignment metadata: who holds which role in which domain, with expiry and revocation
-- tracking. Each row is paired with a casbin_rule row written atomically in AssignRole.
-- is_active=false signals either expiry or explicit revocation; use revoked_at/expires_at
-- to distinguish between the two.
--
-- NOTE: granted_by FK to users(id) is deferred — added in migration 000412 after the users
--       table exists.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE role_assignments (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  subject       VARCHAR(256) NOT NULL,
  role_name     VARCHAR(100) NOT NULL,
  role_slug     VARCHAR(100),                -- denormalized slug for display/audit without JOIN
  domain        VARCHAR(256) NOT NULL,
  assigned_by   VARCHAR(256),                -- subject string of granter (e.g. 'tenant:uuid')
  granted_by    UUID,                        -- UUID of granting user; FK to users(id) added in 000412
  delegated_by  VARCHAR(256),
  expires_at    TIMESTAMPTZ,                 -- NULL = permanent
  revoked_at    TIMESTAMPTZ,                 -- set on explicit revocation
  revoke_reason TEXT,                        -- required when revoked_at is set
  is_active     BOOLEAN      DEFAULT TRUE,
  created_at    TIMESTAMPTZ  DEFAULT NOW(),
  CONSTRAINT role_assignments_unique UNIQUE (subject, role_name, domain)
);

COMMENT ON TABLE  role_assignments              IS 'Role assignment metadata: who holds which role in which domain, with expiry and revocation tracking. Paired with casbin_rule rows written atomically in AssignRole.';
COMMENT ON COLUMN role_assignments.role_slug    IS 'Denormalized role slug for display in audit logs and UI without a roles JOIN.';
COMMENT ON COLUMN role_assignments.granted_by   IS 'UUID FK to the user who granted this assignment. Used for delegation chain audit.';
COMMENT ON COLUMN role_assignments.expires_at   IS 'When the assignment lapses automatically. NULL = permanent. Lazy expiry — checked on first request after expiry.';
COMMENT ON COLUMN role_assignments.revoked_at   IS 'Explicit revocation timestamp. Set alongside is_active=false. Distinguishes revocation from expiry.';
COMMENT ON COLUMN role_assignments.revoke_reason IS 'Human-readable revocation reason. e.g. ''termination'', ''role restructure''.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_role_assignments_subject ON role_assignments(subject, domain);
CREATE INDEX idx_role_assignments_expires ON role_assignments(expires_at) WHERE expires_at IS NOT NULL; -- filter active/lapsed
CREATE INDEX idx_role_assignments_revoked ON role_assignments(revoked_at) WHERE revoked_at IS NOT NULL; -- revocation audit
CREATE INDEX idx_role_assignments_granter ON role_assignments(granted_by) WHERE granted_by IS NOT NULL; -- delegation chain lookup

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE role_assignments ENABLE ROW LEVEL SECURITY;
CREATE POLICY ra_tenant ON role_assignments FOR ALL TO application_role
  USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
  WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
CREATE POLICY ra_admin ON role_assignments FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON role_assignments TO application_role;
GRANT ALL ON role_assignments TO admin_role;
