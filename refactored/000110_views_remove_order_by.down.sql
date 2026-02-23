-- =============================================================================
-- ROLLBACK: Restore ORDER BY in views (functionally identical — ORDER BY is ignored)
-- =============================================================================
-- Note: restoring ORDER BY has no actual effect on query results.
-- This rollback exists only for strict migration reversibility.

CREATE OR REPLACE VIEW v_entity_changes AS
SELECT t.name AS tenant_name, e.uuid AS entity_id, e.name AS entity_name,
  e.type AS entity_type, e.validation_status, e.created_at, e.updated_at, e.deleted_at,
  CASE WHEN e.deleted_at IS NOT NULL THEN 'DELETED'
       WHEN e.updated_at > e.created_at + INTERVAL '1 minute' THEN 'MODIFIED'
       ELSE 'CREATED' END AS change_type
FROM entities e JOIN tenants t ON e.tenant_id = t.id
ORDER BY e.updated_at DESC;
