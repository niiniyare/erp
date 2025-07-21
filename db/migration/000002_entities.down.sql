-- Down migration for Entities Module

-- Drop triggers first
DROP TRIGGER IF EXISTS entities_maintain_entity_id ON entities;
DROP TRIGGER IF EXISTS hierarchy_paths_maintain_entity_id ON hierarchy_paths;
DROP FUNCTION IF EXISTS maintain_entity_id();

-- Drop RLS policies
DROP POLICY IF EXISTS admin_full_access_policy ON entitystate;
DROP POLICY IF EXISTS admin_full_access_policy ON hierarchy_paths;
DROP POLICY IF EXISTS admin_full_access_policy ON entities;
DROP POLICY IF EXISTS tenant_isolation_policy ON entitystate;
DROP POLICY IF EXISTS tenant_isolation_policy ON hierarchy_paths;
DROP POLICY IF EXISTS tenant_isolation_policy ON entities;

-- Disable RLS
ALTER TABLE entitystate DISABLE ROW LEVEL SECURITY;
ALTER TABLE hierarchy_paths DISABLE ROW LEVEL SECURITY;
ALTER TABLE entities DISABLE ROW LEVEL SECURITY;

-- Drop foreign key constraints from dependent tables first
-- ALTER TABLE uom DROP CONSTRAINT IF EXISTS uom_entity_id_fkey;
-- ALTER TABLE uom_conversion DROP CONSTRAINT IF EXISTS uom_conversion_entity_id_fkey;
ALTER TABLE chartofaccount DROP CONSTRAINT IF EXISTS chartofaccount_entity_id_fkey;
ALTER TABLE account DROP CONSTRAINT IF EXISTS account_entity_id_fkey;

-- Drop entitystate table
DROP TABLE IF EXISTS entitystate;

-- Drop hierarchy_paths table
DROP TABLE IF EXISTS hierarchy_paths;

-- Finally, drop entities table and its associated index
DROP INDEX IF EXISTS tenant_code_unique_idx;
DROP TABLE IF EXISTS entities;
