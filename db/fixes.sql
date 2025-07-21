-- =====================================================
-- AUTO-FIX GENERATOR SCRIPT
-- =====================================================
-- Generates SQL statements to fix common migration issues
-- Run this after the validation checklist to get auto-fix commands
-- =====================================================

-- =====================================================
-- 1. GENERATE MISSING COLUMN FIXES
-- =====================================================

CREATE OR REPLACE FUNCTION generate_missing_column_fixes()
RETURNS TEXT AS $$
DECLARE
    fix_sql TEXT := '';
    rec RECORD;
    col_name TEXT;
BEGIN
    fix_sql := '-- =====================================================
-- AUTO-GENERATED FIXES FOR MISSING COLUMNS
-- =====================================================

';
    
    FOR rec IN SELECT * FROM check_standard_columns() WHERE compliance_status = 'NON-COMPLIANT' LOOP
        fix_sql := fix_sql || '-- Fix for table: ' || rec.table_name || E'\n';
        fix_sql := fix_sql || 'ALTER TABLE ' || rec.table_name || E'\n';
        
        -- Add each missing column
        FOREACH col_name IN ARRAY rec.missing_columns LOOP
            CASE col_name
                WHEN 'tenant_id' THEN
                    fix_sql := fix_sql || '    ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,' || E'\n';
                WHEN 'entity_id' THEN
                    fix_sql := fix_sql || '    ADD COLUMN IF NOT EXISTS entity_id UUID NOT NULL DEFAULT uuid_generate_v4(),' || E'\n';
                WHEN 'created_at' THEN
                    fix_sql := fix_sql || '    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),' || E'\n';
                WHEN 'updated_at' THEN
                    fix_sql := fix_sql || '    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),' || E'\n';
                WHEN 'deleted_at' THEN
                    fix_sql := fix_sql || '    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ,' || E'\n';
                WHEN 'modified_by' THEN
                    fix_sql := fix_sql || '    ADD COLUMN IF NOT EXISTS modified_by UUID,' || E'\n';
            END CASE;
        END LOOP;
        
        -- Remove trailing comma and add semicolon
        fix_sql := rtrim(fix_sql, ',' || E'\n') || ';' || E'\n\n';
    END LOOP;
    
    RETURN fix_sql;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 2. GENERATE RLS POLICY FIXES
-- =====================================================

CREATE OR REPLACE FUNCTION generate_rls_policy_fixes()
RETURNS TEXT AS $$
DECLARE
    fix_sql TEXT := '';
    rec RECORD;
BEGIN
    fix_sql := '-- =====================================================
-- AUTO-GENERATED RLS POLICY FIXES
-- =====================================================

';
    
    FOR rec IN SELECT * FROM check_rls_compliance() WHERE compliance_status != 'COMPLIANT' LOOP
        fix_sql := fix_sql || '-- Fix RLS for table: ' || rec.table_name || E'\n';
        
        IF NOT rec.rls_enabled THEN
            fix_sql := fix_sql || 'ALTER TABLE ' || rec.table_name || ' ENABLE ROW LEVEL SECURITY;' || E'\n';
        END IF;
        
        IF NOT rec.has_tenant_policy THEN
            fix_sql := fix_sql || 'CREATE POLICY ' || rec.table_name || '_tenant_isolation_policy ON ' || rec.table_name || E'\n';
            fix_sql := fix_sql || '    FOR ALL TO application_role' || E'\n';
            fix_sql := fix_sql || '    USING (' || E'\n';
            fix_sql := fix_sql || '        current_tenant_id() IS NOT NULL' || E'\n';
            fix_sql := fix_sql || '        AND tenant_id = current_tenant_id()' || E'\n';
            fix_sql := fix_sql || '    )' || E'\n';
            fix_sql := fix_sql || '    WITH CHECK (' || E'\n';
            fix_sql := fix_sql || '        current_tenant_id() IS NOT NULL' || E'\n';
            fix_sql := fix_sql || '        AND tenant_id = current_tenant_id()' || E'\n';
            fix_sql := fix_sql || '    );' || E'\n';
        END IF;
        
        fix_sql := fix_sql || E'\n';
    END LOOP;
    
    RETURN fix_sql;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 3. GENERATE INDEX FIXES
-- =====================================================

CREATE OR REPLACE FUNCTION generate_index_fixes()
RETURNS TEXT AS $$
DECLARE
    fix_sql TEXT := '';
    rec RECORD;
BEGIN
    fix_sql := '-- =====================================================
-- AUTO-GENERATED INDEX FIXES
-- =====================================================

';
    
    FOR rec IN SELECT * FROM check_performance_indexes() 
               WHERE status LIKE '%MISSING%' AND recommendation != 'OK' LOOP
        fix_sql := fix_sql || '-- ' || rec.index_type || ' index for ' || rec.table_name || E'\n';
        fix_sql := fix_sql || rec.recommendation || E'\n\n';
    END LOOP;
    
    RETURN fix_sql;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 4. GENERATE PERMISSION FIXES
-- =====================================================

CREATE OR REPLACE FUNCTION generate_permission_fixes()
RETURNS TEXT AS $$
DECLARE
    fix_sql TEXT := '';
    rec RECORD;
BEGIN
    fix_sql := '-- =====================================================
-- AUTO-GENERATED PERMISSION FIXES
-- =====================================================

';
    
    FOR rec IN SELECT * FROM check_permissions_compliance() WHERE status != 'COMPLIANT' LOOP
        fix_sql := fix_sql || '-- Fix permissions for: ' || rec.object_name || E'\n';
        
        IF rec.object_type = 'TABLE' THEN
            fix_sql := fix_sql || 'GRANT SELECT, INSERT, UPDATE, DELETE ON ' || rec.object_name || ' TO application_role;' || E'\n';
        ELSIF rec.object_type = 'FUNCTION' THEN
            fix_sql := fix_sql || 'GRANT EXECUTE ON FUNCTION ' || rec.object_name || '() TO application_role;' || E'\n';
        END IF;
        
        fix_sql := fix_sql || E'\n';
    END LOOP;
    
    RETURN fix_sql;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 5. GENERATE TRIGGER FIXES
-- =====================================================

CREATE OR REPLACE FUNCTION generate_trigger_fixes()
RETURNS TEXT AS $$
DECLARE
    fix_sql TEXT := '';
    rec RECORD;
BEGIN
    fix_sql := '-- =====================================================
-- AUTO-GENERATED TRIGGER FIXES
-- =====================================================

';
    
    FOR rec IN SELECT * FROM check_required_triggers() WHERE NOT exists LOOP
        fix_sql := fix_sql || '-- ' || rec.trigger_type || ' for ' || rec.table_name || E'\n';
        fix_sql := fix_sql || rec.recommendation || E'\n\n';
    END LOOP;
    
    RETURN fix_sql;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 6. GENERATE COMMENT FIXES
-- =====================================================

CREATE OR REPLACE FUNCTION generate_comment_fixes()
RETURNS TEXT AS $$
DECLARE
    fix_sql TEXT := '';
    rec RECORD;
BEGIN
    fix_sql := '-- =====================================================
-- AUTO-GENERATED COMMENT TEMPLATES
-- =====================================================

';
    
    FOR rec IN SELECT * FROM check_documentation_compliance() WHERE comment_status = 'MISSING COMMENT' LOOP
        IF rec.object_type = 'TABLE' THEN
            fix_sql := fix_sql || 'COMMENT ON TABLE ' || rec.object_name || ' IS ''TODO: Add table description'';' || E'\n';
        ELSIF rec.object_type = 'COLUMN' THEN
            fix_sql := fix_sql || 'COMMENT ON COLUMN ' || rec.object_name || ' IS ''TODO: Add column description'';' || E'\n';
        END IF;
    END LOOP;
    
    RETURN fix_sql;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 7. MASTER FIX GENERATOR
-- =====================================================

CREATE OR REPLACE FUNCTION generate_all_fixes()
RETURNS TEXT AS $$
DECLARE
    complete_fix_sql TEXT := '';
BEGIN
    complete_fix_sql := '-- =====================================================
-- COMPREHENSIVE AUTO-GENERATED MIGRATION FIXES
-- Generated: ' || NOW() || '
-- =====================================================

' || generate_missing_column_fixes() || 
    generate_rls_policy_fixes() || 
    generate_index_fixes() || 
    generate_permission_fixes() || 
    generate_trigger_fixes() || 
    generate_comment_fixes();
    
    RETURN complete_fix_sql;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- USAGE INSTRUCTIONS
-- =====================================================

/*
USAGE:

1. Generate and save all fixes to a file:
   \o fixes.sql
   SELECT generate_all_fixes();
   \o

2. Review the generated fixes before applying

3. Apply the fixes:
   \i fixes.sql

4. Re-run validation to confirm fixes:
   \i migration_validation_checklist.sql

INDIVIDUAL FIX GENERATORS:

-- Just missing columns:
SELECT generate_missing_column_fixes();

-- Just RLS policies:
SELECT generate_rls_policy_fixes();

-- Just indexes:
SELECT generate_index_fixes();

-- Just permissions:
SELECT generate_permission_fixes();

-- Just triggers:
SELECT generate_trigger_fixes();

-- Just comments:
SELECT generate_comment_fixes();
*/

-- =====================================================
-- EXECUTE AUTO-FIX GENERATION
-- =====================================================

-- Uncomment the following lines to automatically generate and display fixes:

/*
\echo '========================================'
\echo 'GENERATING AUTO-FIXES...'
\echo '========================================'

SELECT generate_all_fixes();
*/

-- =====================================================
-- SAFE EXECUTION MODE
-- =====================================================

-- Create a function to safely apply fixes with confirmation
CREATE OR REPLACE FUNCTION apply_fixes_safely(fix_sql TEXT)
RETURNS VOID AS $$
BEGIN
    -- Log what we're about to do
    RAISE NOTICE 'About to apply fixes. Review carefully before proceeding.';
    RAISE NOTICE 'Fix SQL: %', fix_sql;
    
    -- In a real implementation, you might want to:
    -- 1. Create a backup
    -- 2. Apply in a transaction
    -- 3. Validate after each fix
    -- 4. Rollback on any errors
    
    RAISE NOTICE 'Execute the generated SQL manually after review.';
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- CLEANUP
-- =====================================================

-- Uncomment to remove the fix generator functions after use:
/*
DROP FUNCTION IF EXISTS generate_missing_column_fixes();
DROP FUNCTION IF EXISTS generate_rls_policy_fixes();
DROP FUNCTION IF EXISTS generate_index_fixes();
DROP FUNCTION IF EXISTS generate_permission_fixes();
DROP FUNCTION IF EXISTS generate_trigger_fixes();
DROP FUNCTION IF EXISTS generate_comment_fixes();
DROP FUNCTION IF EXISTS generate_all_fixes();
DROP FUNCTION IF EXISTS apply_fixes_safely(TEXT);
*/
