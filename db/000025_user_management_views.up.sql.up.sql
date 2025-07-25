-- =====================================================================
-- USER FUNCTIONS AND TRIGGERS DOWN MIGRATION
-- =====================================================================

-- Drop triggers in reverse order of creation
DROP TRIGGER IF EXISTS validate_role_hierarchy_trigger ON roles;
DROP TRIGGER IF EXISTS enforce_persons_tenant_isolation ON persons;
DROP TRIGGER IF EXISTS update_access_requests_updated_at ON access_requests;
DROP TRIGGER IF EXISTS update_policies_updated_at ON policies;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;
DROP TRIGGER IF EXISTS update_persons_updated_at ON persons;

-- Drop functions
DROP FUNCTION IF EXISTS validate_role_hierarchy() CASCADE;
DROP FUNCTION IF EXISTS enforce_tenant_isolation() CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS cleanup_expired_data(UUID) CASCADE;
DROP FUNCTION IF EXISTS user_has_permission(UUID, VARCHAR, VARCHAR, UUID, UUID, JSONB) CASCADE;
DROP FUNCTION IF EXISTS create_default_system_data(UUID) CASCADE;

-- =====================================================================
-- MIGRATION COMPLETION MESSAGE
-- =====================================================================

DO $$
BEGIN
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'USER FUNCTIONS AND TRIGGERS DOWN MIGRATION COMPLETED';
    RAISE NOTICE '===================================================================';
    RAISE NOTICE 'Functions and triggers dropped.';
    RAISE NOTICE '===================================================================';
END;
$$;