-- 000001_core.down.sql — reverse of 000001_core.up.sql
-- WARNING: removing these functions will break all tenant isolation.
-- Only run this in a full teardown scenario.

DROP FUNCTION IF EXISTS apply_global_rls(regclass);
DROP FUNCTION IF EXISTS apply_tenant_rls(regclass);
DROP FUNCTION IF EXISTS set_tenant_context(uuid);
DROP FUNCTION IF EXISTS current_tenant_id();
