-- Down migration to revert the changes

-- Drop the tenant_configurations table
DROP TABLE IF EXISTS tenant_configurations;

-- Drop the tenants table
DROP TABLE IF EXISTS tenants;

-- Drop extensions (if they were created solely for this schema and are no longer needed)
-- Be cautious when dropping extensions, as they might be used by other parts of your database.
-- Only drop them if you are certain they are not used elsewhere.
DROP EXTENSION IF EXISTS "ltree";
DROP EXTENSION IF EXISTS "uuid-ossp";


