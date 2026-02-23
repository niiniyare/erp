-- =============================================================================
-- ROLLBACK: Remove config column from entitystate
-- =============================================================================
DROP INDEX IF EXISTS idx_entitystate_config_gin;

ALTER TABLE entitystate
  DROP COLUMN IF EXISTS config;
