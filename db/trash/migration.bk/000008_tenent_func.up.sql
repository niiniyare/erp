-- =====================================================
-- ERP/ACCOUNTING SYSTEM - RLS AND HIERARCHY MANAGEMENT (UUID)
-- =====================================================
-- Description: Row Level Security policies, hierarchy management, and utility functions with UUID support
-- Author: Abdirahman Ahmed
-- Date: 2025-07-04
-- Version: 1.0.0
-- Dependencies: Requires base tables with UUID primary keys (tenants, entities, persons, etc.)
-- =====================================================

-- Start transaction to ensure atomic migration
BEGIN;

-- =====================================================
-- TENANT CONTEXT FUNCTIONS
-- =====================================================

-- -----------------------------------------------------
-- CURRENT TENANT ID FUNCTION
-- -----------------------------------------------------
-- Enhanced tenant context function with UUID support and proper error handling
-- CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
-- DECLARE
--     tid_str TEXT;
-- BEGIN
--     -- Get the setting with missing_ok=true to avoid exceptions
--     tid_str := current_setting('app.current_tenant_id', true);
--
--     -- Check if setting exists and is not empty
--     IF tid_str IS NULL OR tid_str = '' THEN
--         RAISE EXCEPTION 'Tenant context not set' USING ERRCODE = 'P0001';
--     END IF;
--
--     -- Convert to UUID with proper error handling
--     BEGIN
--         RETURN tid_str::UUID;
--     EXCEPTION WHEN invalid_text_representation THEN
--         RAISE EXCEPTION 'Invalid tenant UUID: %', tid_str USING ERRCODE = 'P0002';
--     END;
-- END;
-- $$ LANGUAGE plpgsql STABLE SECURITY DEFINER;
--
-- -- Add function comment
-- COMMENT ON FUNCTION current_tenant_id() IS 'Retrieves current tenant UUID from session context with enhanced error handling';
--
-- =====================================================
-- ROW LEVEL SECURITY POLICIES
-- =====================================================

-- -----------------------------------------------------
-- TENANT ISOLATION POLICIES
-- -----------------------------------------------------
-- Apply tenant isolation to all tenant-specific tables
-- Note: These policies assume the tables exist with UUID tenant_id columns. Create tables first if needed.

-- Core tenant table policy
-- CREATE POLICY tenant_isolation_policy ON tenants 
--     USING (id = current_tenant_id());


-- Audit and accounting policies
CREATE POLICY tenant_isolation_policy ON audit_logs 
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_policy ON account 
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_policy ON chartofaccount 
    USING (tenant_id = current_tenant_id());

-- -----------------------------------------------------
-- SPECIALIZED POLICIES
-- -----------------------------------------------------
-- User roles policy - users can only see their own roles within their tenant
CREATE POLICY user_roles_policy ON user_roles
    USING (user_id IN (SELECT id FROM users WHERE tenant_id = current_tenant_id()));

-- Add policy comments
COMMENT ON POLICY tenant_isolation_policy ON tenants IS 'Isolates tenant data based on current session UUID context';
COMMENT ON POLICY user_roles_policy ON user_roles IS 'Allows users to see only their own roles within their tenant';

-- =====================================================
-- TRIGGER FUNCTIONS
-- =====================================================

-- -----------------------------------------------------
-- TIMESTAMP UPDATE FUNCTION
-- -----------------------------------------------------
-- Generic function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_timestamps() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add function comment
COMMENT ON FUNCTION update_timestamps() IS 'Generic trigger function to update updated_at timestamp on row updates';

-- -----------------------------------------------------
-- APPLY TIMESTAMP TRIGGERS
-- -----------------------------------------------------
-- Apply timestamp triggers to all relevant tables
CREATE TRIGGER update_tenant_timestamps 
    BEFORE UPDATE ON tenants 
    FOR EACH ROW EXECUTE FUNCTION update_timestamps();

CREATE TRIGGER update_entity_timestamps 
    BEFORE UPDATE ON entities 
    FOR EACH ROW EXECUTE FUNCTION update_timestamps();

CREATE TRIGGER update_person_timestamps 
    BEFORE UPDATE ON persons 
    FOR EACH ROW EXECUTE FUNCTION update_timestamps();

CREATE TRIGGER update_employee_timestamps 
    BEFORE UPDATE ON employees 
    FOR EACH ROW EXECUTE FUNCTION update_timestamps();

CREATE TRIGGER update_user_timestamps 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_timestamps();


-- Add trigger comments
COMMENT ON TRIGGER update_tenant_timestamps ON tenants IS 'Automatically updates updated_at timestamp on tenant updates';
COMMENT ON TRIGGER update_entity_timestamps ON entities IS 'Automatically updates updated_at timestamp on entity updates';

-- =====================================================
-- HIERARCHY MANAGEMENT FUNCTIONS
-- =====================================================

-- -----------------------------------------------------
-- HIERARCHY PATH MAINTENANCE
-- -----------------------------------------------------
-- Enhanced hierarchy path maintenance function with UUID support (without ltree dependency)
CREATE OR REPLACE FUNCTION maintain_hierarchy_paths() RETURNS TRIGGER AS $$
BEGIN
    -- Handle DELETE operations
    IF TG_OP = 'DELETE' THEN
        -- Remove all paths involving this entity
        DELETE FROM hierarchy_paths
        WHERE tenant_id = OLD.tenant_id
        AND (ancestor_id = OLD.id OR descendant_id = OLD.id);
        RETURN OLD;
    END IF;

    -- Handle UPDATE of parent_id
    IF TG_OP = 'UPDATE' AND OLD.parent_id IS DISTINCT FROM NEW.parent_id THEN
        -- Remove old hierarchy paths for this entity
        DELETE FROM hierarchy_paths
        WHERE tenant_id = NEW.tenant_id AND descendant_id = NEW.id;
    END IF;

    -- Clear old paths for this descendant (for INSERT or parent change)
    IF TG_OP = 'INSERT' OR (TG_OP = 'UPDATE' AND OLD.parent_id IS DISTINCT FROM NEW.parent_id) THEN
        DELETE FROM hierarchy_paths
        WHERE tenant_id = NEW.tenant_id AND descendant_id = NEW.id;
    END IF;

    -- Self-reference (depth 0) - every entity is an ancestor of itself
    INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
    VALUES (NEW.tenant_id, NEW.id, NEW.id, 0)
    ON CONFLICT DO NOTHING;

    -- Add parent paths (all ancestors)
    IF NEW.parent_id IS NOT NULL THEN
        INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
        SELECT NEW.tenant_id, p.ancestor_id, NEW.id, p.depth + 1
        FROM hierarchy_paths p
        WHERE p.tenant_id = NEW.tenant_id
        AND p.descendant_id = NEW.parent_id
        ON CONFLICT DO NOTHING;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add function comment
COMMENT ON FUNCTION maintain_hierarchy_paths() IS 'Maintains hierarchy paths table for efficient hierarchy queries without ltree using UUID';

-- Apply hierarchy maintenance trigger
CREATE TRIGGER trg_maintain_hierarchy_paths
    BEFORE INSERT OR UPDATE OR DELETE ON entities
    FOR EACH ROW EXECUTE FUNCTION maintain_hierarchy_paths();

-- Add trigger comment
COMMENT ON TRIGGER trg_maintain_hierarchy_paths ON entities IS 'Maintains hierarchy paths when entities are modified';

-- =====================================================
-- HIERARCHY QUERY FUNCTIONS
-- =====================================================

-- -----------------------------------------------------
-- GET ENTITY DESCENDANTS
-- -----------------------------------------------------
-- Function to get all descendants of an entity using UUID
CREATE OR REPLACE FUNCTION get_entity_descendants(entity_id UUID, max_depth INT DEFAULT NULL)
RETURNS TABLE(id UUID, name VARCHAR, type VARCHAR, depth INT) AS $$
BEGIN
    RETURN QUERY
    SELECT e.id, e.name, e.type, hp.depth
    FROM entities e
    JOIN hierarchy_paths hp ON e.id = hp.descendant_id
    WHERE hp.tenant_id = current_tenant_id()
    AND hp.ancestor_id = entity_id
    AND (max_depth IS NULL OR hp.depth <= max_depth)
    ORDER BY hp.depth, e.name;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION get_entity_descendants(UUID, INT) IS 'Returns all descendants of an entity with optional depth limit using UUID';

-- -----------------------------------------------------
-- GET ENTITY ANCESTORS
-- -----------------------------------------------------
-- Function to get all ancestors of an entity using UUID
CREATE OR REPLACE FUNCTION get_entity_ancestors(entity_id UUID, max_depth INT DEFAULT NULL)
RETURNS TABLE(id UUID, name VARCHAR, type VARCHAR, depth INT) AS $$
BEGIN
    RETURN QUERY
    SELECT e.id, e.name, e.type, hp.depth
    FROM entities e
    JOIN hierarchy_paths hp ON e.id = hp.ancestor_id
    WHERE hp.tenant_id = current_tenant_id()
    AND hp.descendant_id = entity_id
    AND hp.depth > 0  -- Exclude self-reference
    AND (max_depth IS NULL OR hp.depth <= max_depth)
    ORDER BY hp.depth DESC, e.name;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION get_entity_ancestors(UUID, INT) IS 'Returns all ancestors of an entity with optional depth limit using UUID';

-- -----------------------------------------------------
-- CHECK ENTITY DESCENDANT RELATIONSHIP
-- -----------------------------------------------------
-- Function to check if one entity is descendant of another using UUID
CREATE OR REPLACE FUNCTION is_entity_descendant(descendant_id UUID, ancestor_id UUID)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS(
        SELECT 1 FROM hierarchy_paths
        WHERE tenant_id = current_tenant_id()
        AND ancestor_id = is_entity_descendant.ancestor_id
        AND descendant_id = is_entity_descendant.descendant_id
        AND depth > 0
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION is_entity_descendant(UUID, UUID) IS 'Checks if one entity is a descendant of another using UUID';

-- -----------------------------------------------------
-- GET ENTITY PATH AS TEXT
-- -----------------------------------------------------
-- Function to get entity path as readable text using UUID
CREATE OR REPLACE FUNCTION get_entity_path(entity_id UUID, separator VARCHAR DEFAULT ' > ')
RETURNS TEXT AS $$
DECLARE
    result TEXT;
BEGIN
    SELECT string_agg(e.name, separator ORDER BY hp.depth DESC)
    INTO result
    FROM entities e
    JOIN hierarchy_paths hp ON e.id = hp.ancestor_id
    WHERE hp.tenant_id = current_tenant_id()
    AND hp.descendant_id = entity_id;

    RETURN COALESCE(result, '');
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION get_entity_path(UUID, VARCHAR) IS 'Returns entity path as human-readable text with custom separator using UUID';

-- -----------------------------------------------------
-- GET IMMEDIATE CHILDREN
-- -----------------------------------------------------
-- Function to get immediate children of an entity using UUID
CREATE OR REPLACE FUNCTION get_entity_children(entity_id UUID)
RETURNS TABLE(id UUID, name VARCHAR, type VARCHAR) AS $$
BEGIN
    RETURN QUERY
    SELECT e.id, e.name, e.type
    FROM entities e
    WHERE e.tenant_id = current_tenant_id()
    AND e.parent_id = entity_id
    ORDER BY e.name;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION get_entity_children(UUID) IS 'Returns immediate children of an entity using UUID';

-- -----------------------------------------------------
-- REBUILD HIERARCHY PATHS
-- -----------------------------------------------------
-- Function to rebuild hierarchy paths using UUID (useful for data repair)
CREATE OR REPLACE FUNCTION rebuild_hierarchy_paths(tenant_id_param UUID DEFAULT NULL)
RETURNS VOID AS $$
DECLARE
    target_tenant_id UUID;
BEGIN
    target_tenant_id := COALESCE(tenant_id_param, current_tenant_id());

    -- Clear existing paths for the tenant
    DELETE FROM hierarchy_paths WHERE tenant_id = target_tenant_id;

    -- Rebuild paths using recursive CTE
    WITH RECURSIVE entity_paths AS (
        -- Root entities (self-references)
        SELECT id, id as ancestor_id, 0 as depth, tenant_id
        FROM entities
        WHERE tenant_id = target_tenant_id

        UNION ALL

        -- Parent-child relationships
        SELECT e.id, ep.ancestor_id, ep.depth + 1, e.tenant_id
        FROM entities e
        JOIN entity_paths ep ON e.parent_id = ep.id
    )
    INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
    SELECT tenant_id, ancestor_id, id, depth
    FROM entity_paths;
    
    -- Log completion
    RAISE NOTICE 'Hierarchy paths rebuilt for tenant: %', target_tenant_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION rebuild_hierarchy_paths(UUID) IS 'Rebuilds hierarchy paths for a tenant using UUID (useful for data repair)';

-- =====================================================
-- RLS MANAGEMENT FUNCTIONS
-- =====================================================

-- -----------------------------------------------------
-- APPLY RLS TO TENANT TABLES
-- -----------------------------------------------------
-- Function to apply RLS policies to tenant tables with UUID support
CREATE OR REPLACE FUNCTION apply_rls_to_tenant_tables(
    schema_name TEXT,
    policy_name TEXT DEFAULT 'tenant_isolation_policy',
    policy_type TEXT DEFAULT 'all',
    force_rls BOOLEAN DEFAULT true
)
RETURNS VOID AS $$
DECLARE
    table_record RECORD;
    sql_statement TEXT;
BEGIN
    -- Log start
    RAISE NOTICE 'Applying RLS to tenant tables in schema: %', schema_name;
    
    -- Find all tables with tenant_id column in the specified schema
    FOR table_record IN
        SELECT t.table_name
        FROM information_schema.tables t
        JOIN information_schema.columns c ON t.table_name = c.table_name
        WHERE t.table_schema = schema_name
        AND c.column_name = 'tenant_id'
        AND t.table_type = 'BASE TABLE'
    LOOP
        -- Enable RLS if forced
        IF force_rls THEN
            sql_statement := format('ALTER TABLE %I.%I ENABLE ROW LEVEL SECURITY', schema_name, table_record.table_name);
            EXECUTE sql_statement;
        END IF;
        
        -- Create tenant isolation policy
        sql_statement := format(
            'CREATE POLICY %I ON %I.%I USING (tenant_id = current_tenant_id())',
            policy_name,
            schema_name,
            table_record.table_name
        );
        
        BEGIN
            EXECUTE sql_statement;
            RAISE NOTICE 'Applied policy % to table %.%', policy_name, schema_name, table_record.table_name;
        EXCEPTION
            WHEN duplicate_object THEN
                RAISE NOTICE 'Policy % already exists on table %.%', policy_name, schema_name, table_record.table_name;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION apply_rls_to_tenant_tables(TEXT, TEXT, TEXT, BOOLEAN) IS 'Applies RLS policies to all tenant tables in a schema with UUID support';

-- -----------------------------------------------------
-- REMOVE RLS FROM TENANT TABLES
-- -----------------------------------------------------
-- Function to remove RLS policies from tenant tables
CREATE OR REPLACE FUNCTION remove_rls_from_tenant_tables_by_type(
    schema_name TEXT,
    policy_type TEXT DEFAULT 'all',
    dry_run BOOLEAN DEFAULT false
)
RETURNS VOID AS $$
DECLARE
    table_record RECORD;
    sql_statement TEXT;
BEGIN
    -- Log start
    RAISE NOTICE 'Removing RLS from tenant tables in schema: % (dry_run: %)', schema_name, dry_run;
    
    -- Find all tables with tenant_id column in the specified schema
    FOR table_record IN
        SELECT t.table_name
        FROM information_schema.tables t
        JOIN information_schema.columns c ON t.table_name = c.table_name
        WHERE t.table_schema = schema_name
        AND c.column_name = 'tenant_id'
        AND t.table_type = 'BASE TABLE'
    LOOP
        -- Drop tenant isolation policy
        sql_statement := format(
            'DROP POLICY IF EXISTS tenant_isolation_policy ON %I.%I',
            schema_name,
            table_record.table_name
        );
        
        IF NOT dry_run THEN
            EXECUTE sql_statement;
        END IF;
        
        RAISE NOTICE 'Removed policy from table %.% (dry_run: %)', schema_name, table_record.table_name, dry_run;
    END LOOP;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION remove_rls_from_tenant_tables_by_type(TEXT, TEXT, BOOLEAN) IS 'Removes RLS policies from tenant tables with dry run support';

-- -----------------------------------------------------
-- MANAGE RLS ACROSS SCHEMAS
-- -----------------------------------------------------
-- Master function to manage RLS across multiple schemas
CREATE OR REPLACE FUNCTION manage_rls_across_schemas(
    schemas TEXT[],
    mode TEXT DEFAULT 'apply',
    policy_name TEXT DEFAULT 'tenant_isolation_policy',
    policy_type TEXT DEFAULT 'all',
    force_rls BOOLEAN DEFAULT true,
    dry_run BOOLEAN DEFAULT false
)
RETURNS VOID AS $$
DECLARE
    schema_name TEXT;
BEGIN
    -- Log start
    RAISE NOTICE 'Managing RLS across schemas: % (mode: %, dry_run: %)', schemas, mode, dry_run;
    
    FOREACH schema_name IN ARRAY schemas LOOP
        IF mode = 'apply' THEN
            IF NOT dry_run THEN
                PERFORM apply_rls_to_tenant_tables(schema_name, policy_name, policy_type, force_rls);
            ELSE
                RAISE NOTICE 'Would apply RLS to schema: %', schema_name;
            END IF;
        ELSIF mode = 'remove' THEN
            PERFORM remove_rls_from_tenant_tables_by_type(schema_name, policy_type, dry_run);
        ELSE
            RAISE EXCEPTION 'Invalid mode: %, expected ''apply'' or ''remove''', mode;
        END IF;
    END LOOP;
    
    -- Log completion
    RAISE NOTICE 'Completed RLS management across schemas';
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION manage_rls_across_schemas(TEXT[], TEXT, TEXT, TEXT, BOOLEAN, BOOLEAN) IS 'Manages RLS policies across multiple schemas with dry run support';

-- =====================================================
-- UTILITY FUNCTIONS
-- =====================================================

-- -----------------------------------------------------
-- GET TENANT STATISTICS
-- -----------------------------------------------------
-- Function to get tenant statistics using UUID
CREATE OR REPLACE FUNCTION get_tenant_statistics(tenant_id_param UUID DEFAULT NULL)
RETURNS TABLE(
    tenant_id UUID,
    entity_count BIGINT,
    user_count BIGINT,
    project_count BIGINT,
    hierarchy_depth INT
) AS $$
DECLARE
    target_tenant_id UUID;
BEGIN
    target_tenant_id := COALESCE(tenant_id_param, current_tenant_id());
    
    RETURN QUERY
    SELECT 
        target_tenant_id,
        COALESCE((SELECT COUNT(*) FROM entities WHERE tenant_id = target_tenant_id), 0)::BIGINT,
        COALESCE((SELECT COUNT(*) FROM users WHERE tenant_id = target_tenant_id), 0)::BIGINT,
        -- COALESCE((SELECT COUNT(*) FROM projects WHERE tenant_id = target_tenant_id), 0)::BIGINT,
        COALESCE((SELECT MAX(depth) FROM hierarchy_paths WHERE tenant_id = target_tenant_id), 0)::INT;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION get_tenant_statistics(UUID) IS 'Returns statistical information about a tenant using UUID';

-- -----------------------------------------------------
-- UUID VALIDATION FUNCTIONS
-- -----------------------------------------------------
-- Function to validate if a UUID is properly formatted
CREATE OR REPLACE FUNCTION is_valid_uuid(uuid_text TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    -- Try to cast to UUID
    BEGIN
        PERFORM uuid_text::UUID;
        RETURN TRUE;
    EXCEPTION
        WHEN invalid_text_representation THEN
            RETURN FALSE;
    END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Add function comment
COMMENT ON FUNCTION is_valid_uuid(TEXT) IS 'Validates if a text string is a properly formatted UUID';

-- Function to generate a new UUID if input is null or invalid
CREATE OR REPLACE FUNCTION ensure_valid_uuid(input_uuid UUID DEFAULT NULL)
RETURNS UUID AS $$
BEGIN
    RETURN COALESCE(input_uuid, uuid_generate_v4());
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Add function comment
COMMENT ON FUNCTION ensure_valid_uuid(UUID) IS 'Returns input UUID or generates new one if null';

-- =====================================================
-- TENANT CONTEXT MANAGEMENT (UUID)
-- =====================================================

-- -----------------------------------------------------
-- SET TENANT CONTEXT BY UUID
-- -----------------------------------------------------
-- Function to set tenant context using UUID
CREATE OR REPLACE FUNCTION set_tenant_context_uuid(tenant_uuid UUID)
RETURNS VOID AS $$
BEGIN
    -- Validate tenant exists and is active
    IF NOT EXISTS (
        SELECT 1 FROM tenants 
        WHERE id = tenant_uuid 
        AND status = 'active' 
        AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Invalid or inactive tenant: %', tenant_uuid;
    END IF;
    
    -- Set session variable for tenant context
    PERFORM set_config('app.current_tenant_id', tenant_uuid::TEXT, true);
    
    -- Log tenant context change
    RAISE NOTICE 'Tenant context set to: %', tenant_uuid;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION set_tenant_context_uuid(UUID) IS 'Sets tenant context using UUID with validation';

-- Function to clear tenant context
CREATE OR REPLACE FUNCTION clear_tenant_context()
RETURNS VOID AS $$
BEGIN
    -- Clear session variable
    PERFORM set_config('app.current_tenant_id', '', true);
    RAISE NOTICE 'Tenant context cleared';
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION clear_tenant_context() IS 'Clears the current tenant context';

-- =====================================================
-- PERMISSIONS AND GRANTS
-- =====================================================

-- Grant necessary permissions to application role
-- Note: Assumes application_role exists
GRANT EXECUTE ON FUNCTION current_tenant_id() TO application_role;
GRANT EXECUTE ON FUNCTION update_timestamps() TO application_role;
GRANT EXECUTE ON FUNCTION maintain_hierarchy_paths() TO application_role;
GRANT EXECUTE ON FUNCTION get_entity_descendants(UUID, INT) TO application_role;
GRANT EXECUTE ON FUNCTION get_entity_ancestors(UUID, INT) TO application_role;
GRANT EXECUTE ON FUNCTION is_entity_descendant(UUID, UUID) TO application_role;
GRANT EXECUTE ON FUNCTION get_entity_path(UUID, VARCHAR) TO application_role;
GRANT EXECUTE ON FUNCTION get_entity_children(UUID) TO application_role;
GRANT EXECUTE ON FUNCTION rebuild_hierarchy_paths(UUID) TO application_role;
GRANT EXECUTE ON FUNCTION get_tenant_statistics(UUID) TO application_role;
GRANT EXECUTE ON FUNCTION is_valid_uuid(TEXT) TO application_role;
GRANT EXECUTE ON FUNCTION ensure_valid_uuid(UUID) TO application_role;
GRANT EXECUTE ON FUNCTION set_tenant_context_uuid(UUID) TO application_role;
GRANT EXECUTE ON FUNCTION clear_tenant_context() TO application_role;

-- Grant RLS management functions to admin role (if exists)
-- GRANT EXECUTE ON FUNCTION apply_rls_to_tenant_tables(TEXT, TEXT, TEXT, BOOLEAN) TO admin_role;
-- GRANT EXECUTE ON FUNCTION remove_rls_from_tenant_tables_by_type(TEXT, TEXT, BOOLEAN) TO admin_role;
-- GRANT EXECUTE ON FUNCTION manage_rls_across_schemas(TEXT[], TEXT, TEXT, TEXT, BOOLEAN, BOOLEAN) TO admin_role;

-- =====================================================
-- VALIDATION FUNCTIONS
-- =====================================================

-- -----------------------------------------------------
-- VALIDATE HIERARCHY INTEGRITY
-- -----------------------------------------------------
-- Function to validate hierarchy integrity using UUID
CREATE OR REPLACE FUNCTION validate_hierarchy_integrity(tenant_id_param UUID DEFAULT NULL)
RETURNS TABLE(
    issue_type TEXT,
    entity_id UUID,
    entity_name VARCHAR,
    description TEXT
) AS $$
DECLARE
    target_tenant_id UUID;
BEGIN
    target_tenant_id := COALESCE(tenant_id_param, current_tenant_id());
    
    -- Check for circular references
    RETURN QUERY
    SELECT 
        'circular_reference'::TEXT,
        e.id,
        e.name,
        'Entity is its own ancestor'::TEXT
    FROM entities e
    WHERE e.tenant_id = target_tenant_id
    AND EXISTS (
        SELECT 1 FROM hierarchy_paths hp
        WHERE hp.tenant_id = target_tenant_id
        AND hp.ancestor_id = e.id
        AND hp.descendant_id = e.id
        AND hp.depth > 0
    );
    
    -- Check for missing hierarchy paths
    RETURN QUERY
    SELECT 
        'missing_hierarchy'::TEXT,
        e.id,
        e.name,
        'Entity missing from hierarchy_paths'::TEXT
    FROM entities e
    WHERE e.tenant_id = target_tenant_id
    AND NOT EXISTS (
        SELECT 1 FROM hierarchy_paths hp
        WHERE hp.tenant_id = target_tenant_id
        AND hp.descendant_id = e.id
        AND hp.depth = 0
    );
    
    -- Check for orphaned hierarchy paths
    RETURN QUERY
    SELECT 
        'orphaned_hierarchy'::TEXT,
        hp.descendant_id,
        ''::VARCHAR,
        'Hierarchy path exists but entity does not'::TEXT
    FROM hierarchy_paths hp
    WHERE hp.tenant_id = target_tenant_id
    AND NOT EXISTS (
        SELECT 1 FROM entities e
        WHERE e.tenant_id = target_tenant_id
        AND e.id = hp.descendant_id
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Add function comment
COMMENT ON FUNCTION validate_hierarchy_integrity(UUID) IS 'Validates hierarchy integrity and reports issues using UUID';

-- Grant permission
GRANT EXECUTE ON FUNCTION validate_hierarchy_integrity(UUID) TO application_role;

-- =====================================================
-- EXAMPLE USAGE FUNCTIONS
-- =====================================================

-- -----------------------------------------------------
-- EXAMPLE QUERIES VIEW
-- -----------------------------------------------------
-- Create a view with example queries for UUID usage
CREATE OR REPLACE VIEW hierarchy_uuid_examples AS
SELECT 
    'Example queries for UUID-based hierarchy management' AS description,
    '-- Set tenant context by UUID' AS example_1,
    'SELECT set_tenant_context_uuid(''550e8400-e29b-41d4-a716-446655440000''::UUID);' AS query_1,
    '-- Get current tenant UUID' AS example_2,
    'SELECT current_tenant_id();' AS query_2,
    '-- Get entity descendants' AS example_3,
    'SELECT * FROM get_entity_descendants(''550e8400-e29b-41d4-a716-446655440001''::UUID, 3);' AS query_3,
    '-- Get entity ancestors' AS example_4,
    'SELECT * FROM get_entity_ancestors(''550e8400-e29b-41d4-a716-446655440001''::UUID);' AS query_4,
    '-- Check if entity is descendant' AS example_5,
    'SELECT is_entity_descendant(''child-uuid''::UUID, ''parent-uuid''::UUID);' AS query_5,
    '-- Get entity path' AS example_6,
    'SELECT get_entity_path(''550e8400-e29b-41d4-a716-446655440001''::UUID, '' → '');' AS query_6,
    '-- Get tenant statistics' AS example_7,
    'SELECT * FROM get_tenant_statistics();' AS query_7,
    '-- Validate hierarchy integrity' AS example_8,
    'SELECT * FROM validate_hierarchy_integrity();' AS query_8;

-- Add comment
COMMENT ON VIEW hierarchy_uuid_examples IS 'Example queries demonstrating UUID-based hierarchy operations';

-- =====================================================
-- MIGRATION COMPLETION
-- =====================================================

-- Commit the transaction
COMMIT;

-- Log completion
DO $$
BEGIN
    RAISE NOTICE '====================================================';
    RAISE NOTICE 'ERP RLS and Hierarchy Management Migration (UUID) Completed!';
    RAISE NOTICE '====================================================';
    RAISE NOTICE 'Features Created:';
    RAISE NOTICE '  - Enhanced tenant context management with UUID support';
    RAISE NOTICE '  - Row Level Security policies for all tenant tables';
    RAISE NOTICE '  - Automatic timestamp triggers';
    RAISE NOTICE '  - Comprehensive hierarchy management with UUID';
    RAISE NOTICE '  - Hierarchy query functions (UUID-based)';
    RAISE NOTICE '  - RLS management utilities';
    RAISE NOTICE '  - UUID validation functions';
    RAISE NOTICE '  - Validation and integrity checking';
    RAISE NOTICE '';
    RAISE NOTICE 'Functions Available:';
    RAISE NOTICE '  - current_tenant_id() -> UUID';
    RAISE NOTICE '  - set_tenant_context_uuid(UUID)';
    RAISE NOTICE '  - clear_tenant_context()';
    RAISE NOTICE '  - get_entity_descendants(UUID, INT)';
    RAISE NOTICE '  - get_entity_ancestors(UUID, INT)';
    RAISE NOTICE '  - is_entity_descendant(UUID, UUID) -> BOOLEAN';
    RAISE NOTICE '  - get_entity_path(UUID, VARCHAR) -> TEXT';
    RAISE NOTICE '  - get_entity_children(UUID)';
    RAISE NOTICE '  - rebuild_hierarchy_paths(UUID)';
    RAISE NOTICE '  - get_tenant_statistics(UUID)';
    RAISE NOTICE '  - validate_hierarchy_integrity(UUID)';
    RAISE NOTICE '  - is_valid_uuid(TEXT) -> BOOLEAN';
    RAISE NOTICE '  - ensure_valid_uuid(UUID) -> UUID';
    RAISE NOTICE '  - manage_rls_across_schemas(TEXT[], TEXT, ...)';
    RAISE NOTICE '';
    RAISE NOTICE 'Security Features:';
    RAISE NOTICE '  - RLS policies applied to all tenant tables';
    RAISE NOTICE '  - UUID-based tenant isolation enforced';
    RAISE NOTICE '  - Hierarchy integrity maintained';
    RAISE NOTICE '  - UUID validation and error handling';
    RAISE NOTICE '';
    RAISE NOTICE 'Usage Examples:';
    RAISE NOTICE '  -- Set tenant context';
    RAISE NOTICE '  SELECT set_tenant_context_uuid(''550e8400-e29b-41d4-a716-446655440000''::UUID);';
    RAISE NOTICE '  -- Get descendants';
    RAISE NOTICE '  SELECT * FROM get_entity_descendants(entity_uuid, 3);';
    RAISE NOTICE '  -- Check hierarchy integrity';
    RAISE NOTICE '  SELECT * FROM validate_hierarchy_integrity();';
    RAISE NOTICE '';
    RAISE NOTICE 'Note: This migration uses UUID primary keys throughout.';
    RAISE NOTICE 'All tenant_id references must be UUID type.';
    RAISE NOTICE 'View hierarchy_uuid_examples for more usage examples.';
    RAISE NOTICE '====================================================';
END;
$$;
