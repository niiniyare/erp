-- =============================================================================
-- FIX: Add config JSONB to entitystate for document sequence formatting
-- =============================================================================
-- WHY: The PRD Phase 2 requires configurable document number formatting
--      per entity: custom prefix (BRANCH-INV-), padding length, reset
--      frequency (yearly/monthly/never), and format templates.
--      Currently all entities share the same implicit format with no
--      per-entity customisation possible.
--
-- Schema for config JSONB:
-- {
--   "prefix":          "INV-",          -- document number prefix
--   "suffix":          "",              -- document number suffix
--   "pad_length":      6,               -- zero-padding width
--   "reset_frequency": "yearly",        -- yearly | monthly | never
--   "format_template": "{prefix}{year:2d}{number:06d}{suffix}"
-- }
--
-- NULL config = use tenant-level default from Settings module.
-- '{}' config = entity explicitly uses system defaults with no customisation.
-- =============================================================================

ALTER TABLE entitystate
  ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'::jsonb;

COMMENT ON COLUMN entitystate.config IS
  'Document sequence formatting config for this entity+document_type. '
  'Keys: prefix, suffix, pad_length, reset_frequency, format_template. '
  'NULL = inherit from tenant Settings module defaults. '
  '{} = use system defaults explicitly.';

-- Backfill existing rows with empty object (not NULL) so they use system defaults
UPDATE entitystate
   SET config = '{}'::jsonb
 WHERE config IS NULL;

-- GIN index for sequence config lookups by prefix or format
CREATE INDEX IF NOT EXISTS idx_entitystate_config_gin
  ON entitystate USING gin(config)
  WHERE config IS NOT NULL AND config <> '{}'::jsonb;
