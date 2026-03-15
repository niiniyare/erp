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
