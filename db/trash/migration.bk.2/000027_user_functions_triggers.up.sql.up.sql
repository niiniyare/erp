-- =====================================================================
-- BUDGETS DOWN MIGRATION
-- =====================================================================

DROP TRIGGER IF EXISTS update_budgets_updated_at ON budgets;
DROP POLICY IF EXISTS admin_full_access_policy ON budgets;
DROP POLICY IF EXISTS tenant_isolation_policy ON budgets;
ALTER TABLE IF EXISTS budgets DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS budgets;
