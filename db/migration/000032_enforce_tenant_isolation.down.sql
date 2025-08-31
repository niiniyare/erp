-- =====================================================================
-- ACCOUNT DOWN MIGRATION
-- =====================================================================
DROP POLICY IF EXISTS admin_full_access_policy ON account;

DROP POLICY IF EXISTS tenant_isolation_policy ON account;

ALTER TABLE
  IF EXISTS account DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS account;
