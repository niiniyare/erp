-- Down migration for Entities Module

-- Drop foreign key constraints from dependent tables first
-- ALTER TABLE uom DROP CONSTRAINT IF EXISTS uom_entity_id_fkey;
-- ALTER TABLE uom_conversion DROP CONSTRAINT IF EXISTS uom_conversion_entity_id_fkey;
ALTER TABLE chartofaccount DROP CONSTRAINT IF EXISTS chartofaccount_entity_id_fkey;
ALTER TABLE account DROP CONSTRAINT IF EXISTS account_entity_id_fkey;

-- Now, drop tables that depend on 'entities' if they are also being managed in this migration
-- (If these tables are managed in other migration files, you might only need to drop the FK constraints)

-- Drop entitystate table
DROP TABLE IF EXISTS entitystate;

-- Drop hierarchy_paths table
DROP TABLE IF EXISTS hierarchy_paths;

-- Finally, drop entities table and its associated index
DROP INDEX IF EXISTS tenant_code_unique_idx;
DROP TABLE IF EXISTS entities;
