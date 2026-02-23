-- =============================================================================
-- FIX: Remove ORDER BY from view definitions
-- =============================================================================
-- WHY: ORDER BY inside a CREATE VIEW definition is not honoured by the
--      PostgreSQL query planner. The spec says view output order is undefined
--      unless the outer query specifies ORDER BY. Worse, the planner inserts
--      a sort node for every query on the view — even when the caller adds
--      their own ORDER BY, you may get two sort passes.
--
-- HOW: Drop and recreate the affected views without the ORDER BY clause.
--      Callers that need ordered results must add ORDER BY themselves.
--      This is a best-practice change, not a behaviour change.
-- =============================================================================

-- v_entity_changes: had ORDER BY e.updated_at DESC
CREATE OR REPLACE VIEW v_entity_changes AS
SELECT
  t.name AS tenant_name,
  e.uuid AS entity_id,
  e.name AS entity_name,
  e.type AS entity_type,
  e.validation_status,
  e.created_at,
  e.updated_at,
  e.deleted_at,
  CASE
    WHEN e.deleted_at IS NOT NULL THEN 'DELETED'
    WHEN e.updated_at > e.created_at + INTERVAL '1 minute' THEN 'MODIFIED'
    ELSE 'CREATED'
  END AS change_type
  -- ORDER BY removed: callers must apply ORDER BY e.updated_at DESC themselves
FROM entities e
JOIN tenants t ON e.tenant_id = t.id;

COMMENT ON VIEW v_entity_changes IS
  'Entity lifecycle event view. '
  'ORDER BY was removed from view definition (not honoured by planner). '
  'Use: SELECT * FROM v_entity_changes ORDER BY updated_at DESC';

-- v_entity_paths: had ORDER BY hp.depth, a.name, d.name
CREATE OR REPLACE VIEW v_entity_paths AS
SELECT
  t.name AS tenant_name,
  a.uuid AS ancestor_id,
  a.name AS ancestor_name,
  a.type AS ancestor_type,
  d.uuid AS descendant_id,
  d.name AS descendant_name,
  d.type AS descendant_type,
  hp.depth
  -- ORDER BY removed: callers must apply ORDER BY hp.depth, a.name, d.name themselves
FROM hierarchy_paths hp
JOIN entities a ON hp.ancestor_id = a.uuid
JOIN entities d ON hp.descendant_id = d.uuid
JOIN tenants t ON hp.tenant_id = t.id
WHERE a.deleted_at IS NULL AND d.deleted_at IS NULL;

COMMENT ON VIEW v_entity_paths IS
  'Ancestor-descendant relationship view. '
  'ORDER BY was removed from view definition (not honoured by planner). '
  'Use: SELECT * FROM v_entity_paths ORDER BY depth, ancestor_name, descendant_name';
