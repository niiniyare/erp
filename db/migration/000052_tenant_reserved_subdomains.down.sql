-- =============================================================================
-- MIGRATION 002 DOWN: Reserved Subdomains
-- =============================================================================
-- Dropping this table requires that migration 007 (triggers) has already
-- been rolled back — otherwise the trigger referencing this table will
-- produce an error on the next tenant INSERT/UPDATE.
--
-- Order of DOWN execution: 009 → 008 → 007 → 006 → 005 → 004 → 003 → 002 → 001
-- =============================================================================

DROP TABLE IF EXISTS reserved_subdomains;
