-- ------------------------------------------------------------------------------------------------
-- HIERARCHY_PATHS
-- ------------------------------------------------------------------------------------------------
-- Closure table for efficient entity hierarchy queries. Stores every ancestor-descendant pair
-- (ancestor_id, descendant_id, depth) — including self-references at depth 0.
-- This is the complete information set; no redundant entity_id column is needed.
--
-- NOTE: Depends on entities(uuid) and tenants(id) from migration 000201.
--       The no_self_parent CHECK in 000201 prevents immediate self-reference; this table
--       supplements that with full transitive closure.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE hierarchy_paths (
  tenant_id     UUID        NOT NULL REFERENCES tenants(id)  ON DELETE CASCADE,
  ancestor_id   UUID        NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
  descendant_id UUID        NOT NULL REFERENCES entities(uuid) ON DELETE CASCADE,
  depth         INT         NOT NULL CHECK (depth >= 0),       -- 0 = self, 1 = direct parent-child, 2+ = deeper
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (tenant_id, ancestor_id, descendant_id)
);

COMMENT ON TABLE   hierarchy_paths              IS 'Closure table for efficient entity hierarchy queries. Stores all ancestor-descendant relationships with depth information. Enables fast retrieval of entity trees, subtrees, and hierarchy levels without recursive queries.';
COMMENT ON COLUMN  hierarchy_paths.tenant_id    IS 'Tenant identifier — partitions hierarchy data by tenant for multi-tenancy.';
COMMENT ON COLUMN  hierarchy_paths.ancestor_id  IS 'Ancestor entity in the relationship — references entities.uuid.';
COMMENT ON COLUMN  hierarchy_paths.descendant_id IS 'Descendant entity in the relationship — references entities.uuid.';
COMMENT ON COLUMN  hierarchy_paths.depth        IS 'Hierarchical distance: 0 = self-reference, 1 = direct parent-child, 2+ = deeper.';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------
CREATE INDEX idx_hierarchy_paths_descendant  ON hierarchy_paths(descendant_id);         -- Reverse traversal — find all ancestors of an entity
CREATE INDEX idx_hierarchy_paths_depth       ON hierarchy_paths(tenant_id, depth);       -- Queries filtered by hierarchy depth
CREATE INDEX idx_hierarchy_paths_tenant      ON hierarchy_paths(tenant_id);              -- Tenant-level scans
CREATE INDEX idx_hierarchy_paths_ancestor    ON hierarchy_paths(ancestor_id);            -- Forward traversal — find all descendants
CREATE INDEX idx_hierarchy_paths_desc_depth  ON hierarchy_paths(descendant_id, depth);  -- Find typed ancestor (e.g. nearest COMPANY above entity X)

COMMENT ON INDEX idx_hierarchy_paths_descendant IS 'Enables efficient reverse hierarchy traversal — finds all ancestors of a given entity.';
COMMENT ON INDEX idx_hierarchy_paths_depth      IS 'Optimises queries filtering by hierarchy depth — useful for organisation-level reports.';

-- ------------------------------------------------------------------------------------------------
-- CHECK_ENTITY_HIERARCHY_DEPTH
-- ------------------------------------------------------------------------------------------------
-- Trigger function that prevents A→B→A cycles and enforces a maximum hierarchy depth of 8 levels.
-- The no_self_parent CHECK in 000201 handles the immediate self-reference case; this trigger
-- catches multi-hop cycles and depth violations on every INSERT or UPDATE of parent_id.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION check_entity_hierarchy_depth()
RETURNS TRIGGER AS $$
DECLARE
  v_depth   INTEGER := 0;
  v_current UUID    := NEW.parent_id;
  v_seen    UUID[]  := ARRAY[NEW.uuid];
BEGIN
  WHILE v_current IS NOT NULL LOOP
    -- Cycle detection
    IF v_current = ANY(v_seen) THEN
      RAISE EXCEPTION
        'Circular reference in entity hierarchy: entity % creates a cycle at ancestor %',
        NEW.uuid, v_current
        USING ERRCODE = '23000';
    END IF;

    v_depth := v_depth + 1;
    v_seen  := v_seen || v_current;

    IF v_depth > 8 THEN
      RAISE EXCEPTION
        'Entity hierarchy exceeds the maximum depth of 8 levels (entity %)',
        NEW.uuid
        USING ERRCODE = '23000';
    END IF;

    SELECT parent_id INTO v_current
      FROM entities
     WHERE uuid = v_current;
  END LOOP;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_entity_hierarchy_depth IS
  'Prevents circular references and enforces max hierarchy depth of 8 levels on the entities table. '
  'Walks the parent chain on every INSERT/UPDATE that sets parent_id.';

CREATE TRIGGER entities_check_hierarchy_depth
  BEFORE INSERT OR UPDATE OF parent_id ON entities
  FOR EACH ROW
  WHEN (NEW.parent_id IS NOT NULL)
  EXECUTE FUNCTION check_entity_hierarchy_depth();

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE hierarchy_paths ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON hierarchy_paths
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY admin_full_access_policy ON hierarchy_paths
  FOR ALL TO admin_role USING (TRUE);

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON hierarchy_paths TO application_role;
