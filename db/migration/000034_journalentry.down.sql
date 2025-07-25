-- =====================================================================
-- JOURNALENTRY DOWN MIGRATION
-- =====================================================================

DROP POLICY IF EXISTS admin_full_access_policy ON journalentry;
DROP POLICY IF EXISTS tenant_isolation_policy ON journalentry;
ALTER TABLE IF EXISTS journalentry DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS journalentry;
