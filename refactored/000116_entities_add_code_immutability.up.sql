-- =============================================================================
-- FIX: Enforce entity code immutability after creation
-- =============================================================================
-- WHY: Entity codes appear in document numbers (BRANCH-INV-000001), API
--      paths, integrations, and stored references in ledger entries.
--      Changing a code silently breaks all stored references without any
--      FK cascade to alert the developer.
--      Immutability must be enforced at the DB level because application
--      guards are easier to accidentally bypass.
--
-- Escape hatch: An admin can disable this trigger temporarily for a
--   deliberate code change. That act produces a DDL audit event.
--   ALTER TABLE entities DISABLE TRIGGER enforce_entity_code_immutability;
--   UPDATE entities SET code = 'NEW-CODE' WHERE uuid = '...';
--   ALTER TABLE entities ENABLE TRIGGER enforce_entity_code_immutability;
-- =============================================================================

CREATE OR REPLACE FUNCTION enforce_entity_code_immutability()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  IF OLD.code IS NOT NULL AND OLD.code IS DISTINCT FROM NEW.code THEN
    RAISE EXCEPTION
      'Entity code is immutable after it has been set. '
      'Old: %, Attempted: %. '
      'Disable trigger temporarily for a deliberate code change (creates DDL audit event).',
      OLD.code, NEW.code
      USING ERRCODE = 'check_violation';
  END IF;
  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION enforce_entity_code_immutability() IS
  'BEFORE UPDATE trigger: prevents entity code changes after it has been set. '
  'Codes appear in document numbers and integration paths — changes break stored references. '
  'To change deliberately, disable the trigger (produces DDL audit event).';

DROP TRIGGER IF EXISTS enforce_entity_code_immutability ON entities;
CREATE TRIGGER enforce_entity_code_immutability
  BEFORE UPDATE ON entities
  FOR EACH ROW
  WHEN (OLD.code IS NOT NULL AND OLD.code IS DISTINCT FROM NEW.code)
  EXECUTE FUNCTION enforce_entity_code_immutability();
