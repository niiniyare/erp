-- =============================================================================
-- MIGRATION 004 DOWN: Indexes
-- =============================================================================
-- Dropping indexes is non-destructive (data is preserved).
-- Safe to run without taking a backup first.
-- =============================================================================

DROP INDEX IF EXISTS idx_tenants_name_trgm;
DROP INDEX IF EXISTS idx_tenants_email_trgm;
DROP INDEX IF EXISTS idx_tenants_settings_gin;
DROP INDEX IF EXISTS idx_tenants_metadata_gin;
DROP INDEX IF EXISTS idx_tenants_active;
DROP INDEX IF EXISTS idx_tenants_deleted;
DROP INDEX IF EXISTS idx_tenants_last_activity;
DROP INDEX IF EXISTS idx_tenants_parent;
DROP INDEX IF EXISTS idx_tenants_subdomain;
DROP INDEX IF EXISTS idx_tenants_billing_email;
DROP INDEX IF EXISTS idx_tenants_email;
DROP INDEX IF EXISTS idx_tenants_plan_tier;
DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_slug;

