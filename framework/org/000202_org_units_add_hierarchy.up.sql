-- ------------------------------------------------------------------------------------------------
-- ORG_UNIT_PATHS  (closure table)
-- ------------------------------------------------------------------------------------------------
-- Stores every ancestor–descendant pair for the org_units hierarchy, including self-references
-- at depth 0. Together with the materialized path column on org_units, this enables:
--
--   • O(1) depth lookup          — SELECT depth FROM org_unit_paths WHERE ancestor_id = $a AND descendant_id = $d
--   • O(1) subtree membership    — org_units.org_unit_path LIKE '/root_uuid/%'
--   • Efficient nearest-ancestor — ORDER BY depth ASC LIMIT 1 (used in v_active_org_units)
--   • Ancestor list              — WHERE descendant_id = $x ORDER BY depth
--   • Descendant list            — WHERE ancestor_id   = $x ORDER BY depth
--
-- MAINTENANCE:
--   This table is NOT self-maintaining via triggers. The application layer must call
--   fn_insert_org_unit_paths(org_unit_uuid) on INSERT and fn_rebuild_org_unit_paths(org_unit_uuid)
--   on reparent. On soft-delete, rows are left in place so historical reports remain correct;
--   hard-deletes cascade automatically via the FK ON DELETE CASCADE.
--
--   Row count = Σ (depth_of_node + 1) across all org units. For an 8-level tree with
--   100 nodes the worst case is ~450 rows — well within practical limits.
--
-- NOTE: Depends on org_units(uuid) from migration 000201.
-- ------------------------------------------------------------------------------------------------
CREATE TABLE org_unit_paths (
  tenant_id     UUID        NOT NULL REFERENCES tenants(id)    ON DELETE CASCADE,
  ancestor_id   UUID        NOT NULL REFERENCES org_units(uuid) ON DELETE CASCADE,
  descendant_id UUID        NOT NULL REFERENCES org_units(uuid) ON DELETE CASCADE,
  depth         INTEGER     NOT NULL CHECK (depth >= 0),        -- 0 = self, 1 = direct parent–child, 2+ = deeper
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (tenant_id, ancestor_id, descendant_id)
);

COMMENT ON TABLE  org_unit_paths               IS 'Closure table for the org_units hierarchy. Stores every ancestor–descendant pair with depth, enabling efficient tree queries without recursive CTEs. A row with ancestor_id = descendant_id at depth 0 is the self-reference every node carries.';
COMMENT ON COLUMN org_unit_paths.tenant_id     IS 'Tenant identifier — partitions closure data by tenant for multi-tenancy. Redundant with the FK chain but required for RLS and for cross-table JOINs that filter by tenant.';
COMMENT ON COLUMN org_unit_paths.ancestor_id   IS 'The ancestor org unit in this relationship.';
COMMENT ON COLUMN org_unit_paths.descendant_id IS 'The descendant org unit in this relationship.';
COMMENT ON COLUMN org_unit_paths.depth         IS 'Hierarchical distance between ancestor and descendant: 0 = same node (self-reference), 1 = direct parent–child, 2+ = transitive.';
COMMENT ON COLUMN org_unit_paths.created_at    IS 'Row creation timestamp — set once when the relationship is first recorded.';
COMMENT ON COLUMN org_unit_paths.updated_at    IS 'Last modification timestamp — updated when an ancestor is reparented (depth values change).';

-- ------------------------------------------------------------------------------------------------
-- INDEXES
-- ------------------------------------------------------------------------------------------------

-- Reverse traversal — find all ancestors of a given descendant
CREATE INDEX org_unit_paths_descendant_idx
  ON org_unit_paths (descendant_id);

-- Find nearest typed ancestor (e.g. nearest COMPANY above node X): join + ORDER BY depth ASC LIMIT 1
CREATE INDEX org_unit_paths_desc_depth_idx
  ON org_unit_paths (descendant_id, depth);

-- Forward traversal — find all descendants of a given ancestor
CREATE INDEX org_unit_paths_ancestor_idx
  ON org_unit_paths (ancestor_id);

-- Depth-scoped queries within a tenant (e.g. all level-2 units across the tenant)
CREATE INDEX org_unit_paths_tenant_depth_idx
  ON org_unit_paths (tenant_id, depth);

COMMENT ON INDEX org_unit_paths_descendant_idx   IS 'Enables reverse traversal — finds all ancestors of a given org unit.';
COMMENT ON INDEX org_unit_paths_desc_depth_idx   IS 'Supports nearest-typed-ancestor queries: join org_units on ancestor_id, filter by type, ORDER BY depth ASC LIMIT 1.';
COMMENT ON INDEX org_unit_paths_ancestor_idx     IS 'Enables forward traversal — finds all descendants of a given org unit.';
COMMENT ON INDEX org_unit_paths_tenant_depth_idx IS 'Optimises tenant-wide queries filtered by depth — e.g. all direct children of root units.';

-- ------------------------------------------------------------------------------------------------
-- CHECK_ORG_UNIT_HIERARCHY_DEPTH  (trigger function)
-- ------------------------------------------------------------------------------------------------
-- Fires BEFORE INSERT OR UPDATE OF parent_id on org_units.
--
-- Responsibilities:
--   1. Cycle detection — walks the parent chain collecting visited UUIDs; raises if any UUID
--      appears twice. (The no_self_parent CHECK constraint on org_units catches the trivial
--      self-loop; this trigger catches multi-hop cycles such as A → B → A.)
--   2. Depth enforcement — raises if the chain exceeds 8 levels (root = level 1).
--
-- Concurrency caveat: the walk reads committed org_units rows. A concurrent transaction that
-- is also reparenting can cause this walk to see a stale intermediate state. For reparent
-- operations that must be race-free, the caller should acquire an advisory lock or use
-- SELECT … FOR UPDATE on the affected ancestor chain before calling.
-- ------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION check_org_unit_hierarchy_depth()
RETURNS TRIGGER AS $$
DECLARE
  v_depth   INTEGER := 0;
  v_current UUID    := NEW.parent_id;
  v_seen    UUID[]  := ARRAY[NEW.uuid];
BEGIN
  WHILE v_current IS NOT NULL LOOP

    IF v_current = ANY(v_seen) THEN
      RAISE EXCEPTION
        'Circular reference detected in org unit hierarchy: unit % creates a cycle at ancestor %',
        NEW.uuid, v_current
        USING ERRCODE = '23000';
    END IF;

    v_depth := v_depth + 1;
    v_seen  := v_seen || v_current;

    IF v_depth > 8 THEN
      RAISE EXCEPTION
        'Org unit hierarchy exceeds the maximum depth of 8 levels (unit %)',
        NEW.uuid
        USING ERRCODE = '23000';
    END IF;

    SELECT parent_id
      INTO v_current
      FROM org_units
     WHERE uuid = v_current;

  END LOOP;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_org_unit_hierarchy_depth IS
  'Trigger function — prevents circular parent references and enforces a maximum hierarchy depth '
  'of 8 levels on org_units. Walks the parent chain on every INSERT or UPDATE that sets '
  'parent_id. The trivial self-loop case is caught earlier by the no_self_parent CHECK constraint; '
  'this function catches multi-hop cycles (A → B → A). '
  'Concurrency caveat: reads committed rows only — callers must acquire advisory locks for '
  'race-free reparent operations.';

CREATE TRIGGER org_units_check_hierarchy_depth
  BEFORE INSERT OR UPDATE OF parent_id ON org_units
  FOR EACH ROW
  WHEN (NEW.parent_id IS NOT NULL)
  EXECUTE FUNCTION check_org_unit_hierarchy_depth();

-- ------------------------------------------------------------------------------------------------
-- ROW LEVEL SECURITY
-- ------------------------------------------------------------------------------------------------
ALTER TABLE org_unit_paths ENABLE ROW LEVEL SECURITY;
ALTER TABLE org_unit_paths FORCE  ROW LEVEL SECURITY;

CREATE POLICY org_unit_paths_tenant_isolation ON org_unit_paths
  FOR ALL TO application_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  )
  WITH CHECK (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

CREATE POLICY org_unit_paths_admin_access ON org_unit_paths
  FOR ALL TO admin_role
  USING (TRUE)
  WITH CHECK (TRUE);

CREATE POLICY org_unit_paths_readonly_select ON org_unit_paths
  FOR SELECT TO readonly_role
  USING (
    current_tenant_id() IS NOT NULL
    AND tenant_id = current_tenant_id()
  );

-- ------------------------------------------------------------------------------------------------
-- PERMISSIONS
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE ON org_unit_paths TO application_role;
