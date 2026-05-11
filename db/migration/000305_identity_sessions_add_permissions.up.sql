-- ------------------------------------------------------------------------------------------------
-- USER_SESSIONS — ADD PRINCIPAL_ID COLUMN
-- ------------------------------------------------------------------------------------------------
-- principal_id: for portal users, the contact/employee UUID they act on behalf of.
--
-- NOTE: Depends on user_sessions (000304) and users(id).
--
-- SES-4 (IAM v1.0): permissions column removed from active queries — authorization is
-- Casbin-only. The DB column is retained for rollback safety; it will be dropped in a
-- dedicated migration (DB-1) once v1.0 is fully stable.
-- ------------------------------------------------------------------------------------------------
ALTER TABLE user_sessions
    -- ADD COLUMN IF NOT EXISTS permissions  JSONB NOT NULL DEFAULT '{}'::jsonb,  -- SES-4: removed; Casbin-only authz
    ADD COLUMN IF NOT EXISTS principal_id UUID  REFERENCES users(id) ON DELETE SET NULL;

-- COMMENT ON COLUMN user_sessions.permissions  IS '...';  -- SES-4: column removed from queries
COMMENT ON COLUMN user_sessions.principal_id IS 'For portal users: the contact/employee UUID they act as. NULL for platform and tenant users.';

-- ------------------------------------------------------------------------------------------------
-- API_KEYS TABLE
-- ------------------------------------------------------------------------------------------------
-- Third-party / integration access keys. Sessions are built at request time (not stored in
-- user_sessions) to avoid DB bloat from high-frequency API calls. Scopes are a ceiling on
-- permissions — a key can never exceed the creating user's own grants.
--
-- NOTE: Depends on tenants(id) and users(id). RLS uses current_tenant_id().
-- ------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS api_keys (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name         VARCHAR(100) NOT NULL,
  key_hash     TEXT         UNIQUE NOT NULL,  -- SHA-256 of the raw key; raw key never stored
  scopes       TEXT[]       NOT NULL DEFAULT '{}',  -- subset of permission keys
  created_by   UUID         NOT NULL REFERENCES users(id),
  expires_at   TIMESTAMPTZ,                          -- NULL = no expiry
  revoked_at   TIMESTAMPTZ,                          -- NULL = active
  last_used_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  api_keys            IS 'Third-party / integration API keys. Key is hashed (SHA-256) — raw key returned once at creation and never stored.';
COMMENT ON COLUMN api_keys.key_hash   IS 'SHA-256 hex of the raw bearer token. Raw token returned once at creation, never stored.';
COMMENT ON COLUMN api_keys.scopes     IS 'Allowed permission keys for this key. Acts as a ceiling — cannot exceed the creating user''s own permissions.';
COMMENT ON COLUMN api_keys.created_by IS 'User who created this key. Key inherits at most the creator''s permissions at time of creation.';
COMMENT ON COLUMN api_keys.revoked_at IS 'Set to revoke the key immediately. Checked on every request before permission evaluation.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_api_keys_tenant ON api_keys(tenant_id);                        -- tenant-based queries
CREATE INDEX idx_api_keys_hash   ON api_keys(key_hash);                         -- key lookup on every request
CREATE INDEX idx_api_keys_active ON api_keys(tenant_id, revoked_at)
    WHERE revoked_at IS NULL;                                                    -- active key queries
CREATE INDEX idx_api_keys_expiry ON api_keys(expires_at)
    WHERE expires_at IS NOT NULL;                                                -- expiry cleanup

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys FORCE  ROW LEVEL SECURITY;

CREATE POLICY api_keys_tenant_isolation ON api_keys FOR ALL TO application_role
    USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
    WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

CREATE POLICY api_keys_admin_access ON api_keys FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

CREATE POLICY api_keys_ro_select ON api_keys
    FOR SELECT TO readonly_role
    USING (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON api_keys TO application_role;
GRANT ALL ON api_keys TO admin_role;
