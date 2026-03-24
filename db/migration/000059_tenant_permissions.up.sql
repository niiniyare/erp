-- ------------------------------------------------------------------------------------------------
-- TENANT FOUNDATION — ROLE PERMISSIONS (GRANTs)
-- ------------------------------------------------------------------------------------------------
-- Grants the minimum necessary privileges to each role for the tenant foundation tables
-- and context functions. GRANTs live in their own migration for two reasons:
--   1. They can be re-applied safely (GRANT is idempotent).
--   2. They change more frequently than structural DDL — separating them lets the
--      permissions model evolve without touching table/function migrations.
--
-- Principle of least privilege:
--   application_role receives only what it needs to serve requests:
--     • SELECT + INSERT + UPDATE + DELETE on tenants (filtered by RLS)
--     • SELECT on reserved_subdomains (for UI "subdomain unavailable" messages)
--     • SELECT on config_definitions (to resolve configuration hierarchy)
--     • EXECUTE on all four context functions
--   It does NOT receive:
--     • TRUNCATE on any table
--     • INSERT/UPDATE/DELETE on reserved_subdomains (ops-only table)
--     • INSERT/UPDATE/DELETE on config_definitions (Settings module service account only)
--
--   readonly_role receives SELECT only. Period.
--   admin_role receives ALL PRIVILEGES on tables and functions.
--   These are narrowly scoped to this schema — not database-wide.
-- ------------------------------------------------------------------------------------------------

-- ------------------------------------------------------------------------------------------------
-- TENANTS TABLE
-- ------------------------------------------------------------------------------------------------
GRANT SELECT, INSERT, UPDATE, DELETE  ON tenants TO application_role;  -- full DML, filtered by RLS
GRANT ALL PRIVILEGES                  ON tenants TO admin_role;         -- unrestricted (RLS policy is USING(TRUE))
GRANT SELECT                          ON tenants TO readonly_role;      -- SELECT only, filtered by RLS

-- ------------------------------------------------------------------------------------------------
-- RESERVED_SUBDOMAINS TABLE
-- ------------------------------------------------------------------------------------------------
-- application_role reads this to surface "subdomain unavailable" messages
-- in the UI before the trigger fires.
GRANT SELECT           ON reserved_subdomains TO application_role;
GRANT SELECT           ON reserved_subdomains TO readonly_role;
GRANT SELECT, INSERT, DELETE ON reserved_subdomains TO admin_role;  -- ops manages the list

-- ------------------------------------------------------------------------------------------------
-- CONFIG_DEFINITIONS TABLE
-- ------------------------------------------------------------------------------------------------
-- application_role reads definitions to resolve the configuration hierarchy.
-- Writes are performed via the Settings module's admin service account
-- (dedicated DB user, not application_role). application_role cannot modify definitions.
GRANT SELECT                         ON config_definitions TO application_role;
GRANT SELECT                         ON config_definitions TO readonly_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON config_definitions TO admin_role;

-- ------------------------------------------------------------------------------------------------
-- CONTEXT FUNCTIONS
-- ------------------------------------------------------------------------------------------------
-- application_role needs all four context functions.
GRANT EXECUTE ON FUNCTION current_tenant_id()        TO application_role;
GRANT EXECUTE ON FUNCTION set_tenant_context(UUID)   TO application_role;
GRANT EXECUTE ON FUNCTION clear_tenant_context()     TO application_role;
GRANT EXECUTE ON FUNCTION validate_tenant_context()  TO application_role;

-- readonly_role needs to read the context (RLS evaluation) and validate it mid-job,
-- but must never set or clear it.
GRANT EXECUTE ON FUNCTION current_tenant_id()       TO readonly_role;
GRANT EXECUTE ON FUNCTION validate_tenant_context() TO readonly_role;

-- admin_role: all context functions for tooling and diagnostics.
GRANT EXECUTE ON FUNCTION current_tenant_id()        TO admin_role;
GRANT EXECUTE ON FUNCTION set_tenant_context(UUID)   TO admin_role;
GRANT EXECUTE ON FUNCTION clear_tenant_context()     TO admin_role;
GRANT EXECUTE ON FUNCTION validate_tenant_context()  TO admin_role;

-- ------------------------------------------------------------------------------------------------
-- OPERATIONS NOTES
-- ------------------------------------------------------------------------------------------------
-- 1. The Settings module background worker (bulk config imports, template application)
--    should connect as a DEDICATED DB user granted application_role PLUS
--    INSERT/UPDATE on config_definitions. Do not use application_role directly
--    for write operations on config_definitions from background jobs.
--
-- 2. For read replicas, connect as readonly_role. set_tenant_context() must be
--    called on read replicas too — GRANT above covers current_tenant_id() for
--    readonly_role. If readonly_role needs to set context (replica after context-
--    setting sequence), grant set_tenant_context(UUID) to readonly_role explicitly.
--
-- 3. Sequence GRANTs: if sequences are added for id columns (not needed for
--    gen_random_uuid() based PKs), grant USAGE and SELECT to application_role
--    and NEXTVAL to the specific writer role.
-- ------------------------------------------------------------------------------------------------
