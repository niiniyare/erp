-- =============================================================================
-- MIGRATION 008 DOWN: Configuration Definitions
-- =============================================================================
-- WARNING: This destroys all configuration definitions including any
-- module-specific definitions added after initial seeding.
-- Ensure the Settings module is fully disabled before rolling back.
--
-- The trigger is also removed since update_updated_at_column() is defined
-- in migration 007 — the trigger reference must be dropped before 007.down.
-- =============================================================================

DROP TRIGGER IF EXISTS update_config_definitions_updated_at ON config_definitions;

DROP TABLE IF EXISTS config_definitions;
