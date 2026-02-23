-- =============================================================================
-- ROLLBACK: Remove config_definition_id FK from configuration_audit
-- =============================================================================
DROP INDEX IF EXISTS idx_config_audit_tenant_time;
DROP INDEX IF EXISTS idx_config_audit_definition;

ALTER TABLE configuration_audit
  DROP COLUMN IF EXISTS config_definition_id;
