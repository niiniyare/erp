-- =============================================================================
-- MIGRATION 007 DOWN: Triggers
-- =============================================================================
-- Drop triggers before functions — a function cannot be dropped while a
-- trigger still references it.
-- =============================================================================

-- Triggers
DROP TRIGGER IF EXISTS tenant_hierarchy_depth_check         ON tenants;
DROP TRIGGER IF EXISTS tenant_subdomain_reserved_check      ON tenants;
DROP TRIGGER IF EXISTS enforce_tenants_slug_immutability     ON tenants;
DROP TRIGGER IF EXISTS tenant_slug_trigger                  ON tenants;
DROP TRIGGER IF EXISTS update_tenants_last_activity         ON tenants;
DROP TRIGGER IF EXISTS update_tenants_updated_at            ON tenants;

-- Trigger functions
DROP FUNCTION IF EXISTS check_tenant_hierarchy_depth();
DROP FUNCTION IF EXISTS check_subdomain_not_reserved();
DROP FUNCTION IF EXISTS enforce_slug_immutability();
DROP FUNCTION IF EXISTS generate_unique_slug_from_name();
DROP FUNCTION IF EXISTS update_last_activity_at_column();
DROP FUNCTION IF EXISTS update_updated_at_column();
