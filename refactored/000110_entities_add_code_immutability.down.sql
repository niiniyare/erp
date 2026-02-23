-- =============================================================================
-- ROLLBACK: Remove entity code immutability trigger
-- =============================================================================
DROP TRIGGER IF EXISTS enforce_entity_code_immutability ON entities;
DROP FUNCTION IF EXISTS enforce_entity_code_immutability();
