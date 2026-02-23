-- =============================================================================
-- FIX: Enforce is_active / deleted_at consistency on entities
-- =============================================================================
-- WHY: An entity can currently have is_active = TRUE and deleted_at IS NOT NULL
--      simultaneously. This is contradictory — a deleted entity should not be
--      active. Application code that filters by is_active = TRUE will include
--      soft-deleted entities; code that filters by deleted_at IS NULL will
--      miss entities that are active but incorrectly not soft-deleted.
--
-- RULE: When deleted_at is set, is_active must be FALSE.
--       When is_active is set to TRUE, deleted_at must be NULL.
--
-- HOW: Trigger that enforces the invariant on INSERT and UPDATE.
--      Also reports existing inconsistencies as WARNINGs (not errors)
--      so this migration can be applied to a live database without
--      breaking existing data — fix the data separately.
-- =============================================================================

-- Report existing inconsistencies without blocking the migration
DO $$
DECLARE
  inconsistent_count INTEGER;
BEGIN
  SELECT COUNT(*) INTO inconsistent_count
    FROM entities
   WHERE is_active = TRUE AND deleted_at IS NOT NULL;

  IF inconsistent_count > 0 THEN
    RAISE WARNING
      'entities_enforce_active_deleted_consistency: found % entities with '
      'is_active=TRUE and deleted_at IS NOT NULL. '
      'Run: UPDATE entities SET is_active = FALSE WHERE deleted_at IS NOT NULL; '
      'before this invariant is fully enforced.',
      inconsistent_count;
  END IF;
END;
$$;

CREATE OR REPLACE FUNCTION enforce_entity_active_deleted_consistency()
  RETURNS TRIGGER
  LANGUAGE plpgsql
AS $$
BEGIN
  -- A deleted entity cannot be active
  IF NEW.deleted_at IS NOT NULL AND NEW.is_active = TRUE THEN
    RAISE EXCEPTION
      'Entity consistency violation: is_active cannot be TRUE when deleted_at is set. '
      'Set is_active = FALSE before or alongside setting deleted_at. '
      'Entity: %, name: %', NEW.uuid, NEW.name
      USING ERRCODE = 'check_violation';
  END IF;

  -- Reactivating an entity must also clear deleted_at
  IF NEW.is_active = TRUE AND OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NOT NULL THEN
    RAISE EXCEPTION
      'Entity consistency violation: cannot set is_active = TRUE while deleted_at remains set. '
      'Clear deleted_at alongside setting is_active = TRUE. '
      'Entity: %, name: %', NEW.uuid, NEW.name
      USING ERRCODE = 'check_violation';
  END IF;

  RETURN NEW;
END;
$$;

COMMENT ON FUNCTION enforce_entity_active_deleted_consistency() IS
  'BEFORE INSERT OR UPDATE trigger: prevents is_active=TRUE + deleted_at IS NOT NULL '
  'from existing simultaneously on the same entity row. '
  'To restore a deleted entity: UPDATE entities SET is_active=TRUE, deleted_at=NULL WHERE uuid=$1.';

DROP TRIGGER IF EXISTS enforce_entity_active_deleted ON entities;
CREATE TRIGGER enforce_entity_active_deleted
  BEFORE INSERT OR UPDATE OF is_active, deleted_at ON entities
  FOR EACH ROW
  EXECUTE FUNCTION enforce_entity_active_deleted_consistency();
