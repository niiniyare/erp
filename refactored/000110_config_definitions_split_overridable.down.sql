-- =============================================================================
-- ROLLBACK: Remove the two-axis override columns
-- =============================================================================
ALTER TABLE config_definitions
  DROP COLUMN IF EXISTS is_entity_overridable,
  DROP COLUMN IF EXISTS is_tenant_overridable;
