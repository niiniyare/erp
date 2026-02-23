-- =============================================================================
-- ROLLBACK: Remove entity is_active / deleted_at consistency trigger
-- =============================================================================
DROP TRIGGER IF EXISTS enforce_entity_active_deleted ON entities;
DROP FUNCTION IF EXISTS enforce_entity_active_deleted_consistency();
