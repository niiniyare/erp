-- Down script for entities and hierarchy_paths tables

-- First, drop the foreign key constraint that depends on the 'entities' table
-- ALTER TABLE customer DROP CONSTRAINT IF EXISTS customer_entity_id_fkey; -- Assuming 'customer' is the table and 'customer_entity_id_fkey' is the constraint name

DROP TABLE IF EXISTS hierarchy_paths;

-- DROP INDEX IF EXISTS tenant_code_unique_idx; -- Explicitly drop the index

DROP TABLE IF EXISTS entities;
