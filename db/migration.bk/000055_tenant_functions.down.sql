-- =============================================================================
-- MIGRATION 005 DOWN: Tenant Context Functions
-- =============================================================================
-- Dropping functions is safe only after RLS policies (006) and
-- permissions (009) that reference them have been rolled back.
-- =============================================================================

DROP FUNCTION IF EXISTS validate_tenant_context();
DROP FUNCTION IF EXISTS clear_tenant_context();
DROP FUNCTION IF EXISTS set_tenant_context(UUID);
DROP FUNCTION IF EXISTS current_tenant_id();
