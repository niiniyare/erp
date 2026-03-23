-- =============================================================================
-- MIGRATION 009 DOWN: Permissions (REVOKEs)
-- =============================================================================
-- Revoking GRANTs does not drop any data or objects.
-- Safe to run independently of other migrations.
-- Run this BEFORE rolling back migrations 005–008 if those functions/tables
-- need to be dropped (you cannot drop an object that still has live grants
-- in some PostgreSQL configurations).
-- =============================================================================

-- CONTEXT FUNCTIONS
REVOKE EXECUTE ON FUNCTION validate_tenant_context()  FROM admin_role;
REVOKE EXECUTE ON FUNCTION clear_tenant_context()     FROM admin_role;
REVOKE EXECUTE ON FUNCTION set_tenant_context(UUID)   FROM admin_role;
REVOKE EXECUTE ON FUNCTION current_tenant_id()        FROM admin_role;

REVOKE EXECUTE ON FUNCTION validate_tenant_context()  FROM readonly_role;
REVOKE EXECUTE ON FUNCTION current_tenant_id()        FROM readonly_role;

REVOKE EXECUTE ON FUNCTION validate_tenant_context()  FROM application_role;
REVOKE EXECUTE ON FUNCTION clear_tenant_context()     FROM application_role;
REVOKE EXECUTE ON FUNCTION set_tenant_context(UUID)   FROM application_role;
REVOKE EXECUTE ON FUNCTION current_tenant_id()        FROM application_role;

-- CONFIG_DEFINITIONS
REVOKE ALL PRIVILEGES ON config_definitions FROM admin_role;
REVOKE SELECT ON config_definitions FROM readonly_role;
REVOKE SELECT ON config_definitions FROM application_role;

-- RESERVED_SUBDOMAINS
REVOKE ALL PRIVILEGES ON reserved_subdomains FROM admin_role;
REVOKE SELECT ON reserved_subdomains FROM readonly_role;
REVOKE SELECT ON reserved_subdomains FROM application_role;

-- TENANTS
REVOKE ALL PRIVILEGES ON tenants FROM admin_role;
REVOKE SELECT ON tenants FROM readonly_role;
REVOKE SELECT, INSERT, UPDATE, DELETE ON tenants FROM application_role;
