-- =============================================================================
-- FIX: Add config_definition_id FK to configuration_audit
-- =============================================================================
-- WHY: configuration_audit.config_key is a VARCHAR string with no FK.
--      After a config key is renamed or deleted, audit records become
--      unresolvable — you cannot JOIN audit to config_definitions.
--      Storing the definition UUID alongside the key string lets the
--      audit trail survive key renames (the UUID is stable, the key may change).
--
-- HOW: Add nullable config_definition_id. Nullable because:
--      - Existing audit rows predate this column.
--      - TEMPLATE_APPLY operations may not map 1:1 to a single definition.
--      Application code should populate this on all new writes.
-- =============================================================================

ALTER TABLE configuration_audit
  ADD COLUMN IF NOT EXISTS config_definition_id UUID
    REFERENCES config_definitions(id) ON DELETE SET NULL;

COMMENT ON COLUMN configuration_audit.config_definition_id IS
  'FK to config_definitions. Stable reference even if config_key is later renamed. '
  'NULL for historical records predating this column and for TEMPLATE_APPLY operations '
  'that touch multiple definitions atomically.';

-- Index to support "show all changes to this definition" queries
CREATE INDEX IF NOT EXISTS idx_config_audit_definition
  ON configuration_audit(config_definition_id, applied_at DESC)
  WHERE config_definition_id IS NOT NULL;

-- Index for time-range dashboard queries: "all changes in last 30 days for tenant"
-- This was missing from the original schema (tenant_id + config_key required knowing key upfront)
CREATE INDEX IF NOT EXISTS idx_config_audit_tenant_time
  ON configuration_audit(tenant_id, applied_at DESC);
