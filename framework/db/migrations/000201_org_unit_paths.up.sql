-- ============================================================
-- 000201_org_unit_paths — closure table for ancestor lookups
-- ============================================================
-- Stores every ancestor–descendant pair in the org hierarchy,
-- including self-references at depth 0. Maintained by
-- platform/org/pgorg via InsertPaths and RebuildPaths.
-- ============================================================

CREATE TABLE IF NOT EXISTS org_unit_paths (
    ancestor_id   uuid    NOT NULL REFERENCES org_units (id) ON DELETE CASCADE,
    descendant_id uuid    NOT NULL REFERENCES org_units (id) ON DELETE CASCADE,
    depth         integer NOT NULL CHECK (depth >= 0),
    PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE INDEX IF NOT EXISTS org_unit_paths_descendant_idx ON org_unit_paths (descendant_id);

SELECT apply_tenant_rls('org_unit_paths');
