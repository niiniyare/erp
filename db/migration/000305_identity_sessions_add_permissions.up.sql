-- ================================================================================================
-- USER SESSIONS - Add permissions JSONB and principal_id columns
-- ================================================================================================
--
-- permissions: pre-computed Casbin policy map stored at login time for O(1) per-request
--              lookups via ResolvedSession.Can(permission).
-- principal_id: for portal users, the contact/employee UUID they represent.
-- ================================================================================================

ALTER TABLE user_sessions
    ADD COLUMN IF NOT EXISTS permissions  JSONB    NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS principal_id UUID     REFERENCES users(id) ON DELETE SET NULL;

COMMENT ON COLUMN user_sessions.permissions  IS 'Pre-computed permission map {"finance.invoices.read": true, ...} stored at login; used for O(1) authz on every request.';
COMMENT ON COLUMN user_sessions.principal_id IS 'For portal users: the contact/employee UUID they act as. NULL for platform and tenant users.';

-- ================================================================================================
-- API KEYS — Third-party / integration access keys
-- ================================================================================================
-- API key sessions are built at request time (not stored in user_sessions) to avoid DB bloat
-- from high-frequency API calls. Scopes are a ceiling on permissions.
CREATE TABLE IF NOT EXISTS api_keys (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name         VARCHAR(100) NOT NULL,
  key_hash     TEXT        UNIQUE NOT NULL,  -- SHA-256 of the raw key; raw key never stored
  scopes       TEXT[]      NOT NULL DEFAULT '{}',  -- subset of permission keys
  created_by   UUID        NOT NULL REFERENCES users(id),
  expires_at   TIMESTAMPTZ,                 -- NULL = no expiry
  revoked_at   TIMESTAMPTZ,                 -- NULL = active
  last_used_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  api_keys            IS 'Third-party / integration API keys. Key is hashed (SHA-256) — raw key returned once at creation and never stored.';
COMMENT ON COLUMN api_keys.key_hash   IS 'SHA-256 hex of the raw bearer token. Raw token returned once at creation, never stored.';
COMMENT ON COLUMN api_keys.scopes     IS 'Allowed permission keys for this key. Acts as a ceiling — cannot exceed the creating user''s own permissions.';
COMMENT ON COLUMN api_keys.created_by IS 'User who created this key. Key inherits at most the creator''s permissions at time of creation.';
COMMENT ON COLUMN api_keys.revoked_at IS 'Set to revoke the key immediately. Checked on every request before permission evaluation.';

CREATE INDEX idx_api_keys_tenant    ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_hash      ON api_keys(key_hash);
CREATE INDEX idx_api_keys_active    ON api_keys(tenant_id, revoked_at)
    WHERE revoked_at IS NULL;
CREATE INDEX idx_api_keys_expiry    ON api_keys(expires_at)
    WHERE expires_at IS NOT NULL;

ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
CREATE POLICY api_keys_tenant_isolation ON api_keys FOR ALL TO application_role
    USING  (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id())
    WITH CHECK (current_tenant_id() IS NOT NULL AND tenant_id = current_tenant_id());
CREATE POLICY api_keys_admin_access ON api_keys FOR ALL TO admin_role USING (TRUE) WITH CHECK (TRUE);

GRANT SELECT, INSERT, UPDATE, DELETE ON api_keys TO application_role;
GRANT ALL ON api_keys TO admin_role;
