-- =====================================================
-- JSON MIGRATION VALIDATION CHECKLIST
-- =====================================================
-- JSON-formatted output version of migration validation
-- Perfect for API integration, automation, and monitoring
-- =====================================================

-- =====================================================
-- 1. JSON STANDARD COLUMNS VALIDATION
-- =====================================================

CREATE OR REPLACE FUNCTION check_standard_columns_json()
RETURNS JSON AS $$
DECLARE
    result JSON;
BEGIN
    WITH compliance_data AS (
        SELECT 
            t.table_name,
            array_agg(rc.column_name) FILTER (WHERE rc.column_name != ALL(tc.existing_columns)) AS missing_columns,
            CASE 
                WHEN array_agg(rc.column_name) FILTER (WHERE rc.column_name != ALL(tc.existing_columns)) IS NULL 
                  OR array_length(array_agg(rc.column_name) FILTER (WHERE rc.column_name != ALL(tc.existing_columns)), 1) = 0 
                THEN 'COMPLIANT'
                ELSE 'NON-COMPLIANT'
            END AS compliance_status
        FROM information_schema.tables t
        LEFT JOIN information_schema.columns c ON t.table_name = c.table_name
        CROSS JOIN unnest(ARRAY['tenant_id', 'entity_id', 'created_at', 'updated_at', 'deleted_at', 'modified_by']) AS rc(column_name)
        LEFT JOIN (
            SELECT 
                table_name,
                array_agg(column_name) AS existing_columns
            FROM information_schema.columns
            WHERE table_schema = 'public'
            GROUP BY table_name
        ) tc ON t.table_name = tc.table_name
        WHERE t.table_schema = 'public'
          AND t.table_type = 'BASE TABLE'
          AND t.table_name NOT IN ('tenants', 'tenant_configurations', 'tenant_usage_stats')
        GROUP BY t.table_name, tc.existing_columns
    ),
    summary AS (
        SELECT 
            COUNT(*) as total_tables,
            COUNT(*) FILTER (WHERE compliance_status = 'COMPLIANT') as compliant_tables,
            COUNT(*) FILTER (WHERE compliance_status = 'NON-COMPLIANT') as non_compliant_tables,
            ROUND(
                (COUNT(*) FILTER (WHERE compliance_status = 'COMPLIANT')::NUMERIC / COUNT(*)) * 100, 
                2
            ) as compliance_percentage
        FROM compliance_data
    ),
    detailed_issues AS (
        SELECT json_agg(
            json_build_object(
                'table_name', table_name,
                'missing_columns', COALESCE(missing_columns, ARRAY[]::TEXT[]),
                'compliance_status', compliance_status
            )
        ) as tables
        FROM compliance_data
    )
    SELECT json_build_object(
        'check_type', 'standard_columns',
        'timestamp', NOW(),
        'summary', json_build_object(
            'total_tables', s.total_tables,
            'compliant_tables', s.compliant_tables,
            'non_compliant_tables', s.non_compliant_tables,
            'compliance_percentage', s.compliance_percentage,
            'status', CASE 
                WHEN s.compliance_percentage = 100 THEN 'PASS'
                WHEN s.compliance_percentage >= 80 THEN 'WARNING'
                ELSE 'FAIL'
            END
        ),
        'details', d.tables
    ) INTO result
    FROM summary s, detailed_issues d;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 2. JSON RLS COMPLIANCE VALIDATION
-- =====================================================

CREATE OR REPLACE FUNCTION check_rls_compliance_json()
RETURNS JSON AS $$
DECLARE
    result JSON;
BEGIN
    WITH rls_data AS (
        SELECT 
            c.relname as table_name,
            c.relrowsecurity as rls_enabled,
            EXISTS(
                SELECT 1 FROM pg_policy p 
                WHERE p.polrelid = c.oid 
                  AND p.polname LIKE '%tenant%'
            ) as has_tenant_policy,
            CASE 
                WHEN c.relrowsecurity AND EXISTS(
                    SELECT 1 FROM pg_policy p2 
                    WHERE p2.polrelid = c.oid 
                      AND p2.polname LIKE '%tenant%'
                ) THEN 'COMPLIANT'
                WHEN NOT c.relrowsecurity THEN 'RLS_NOT_ENABLED'
                ELSE 'MISSING_TENANT_POLICY'
            END as compliance_status,
            array_agg(p.polname) FILTER (WHERE p.polname IS NOT NULL) as policies
        FROM pg_class c
        LEFT JOIN pg_policy p ON p.polrelid = c.oid
        WHERE c.relkind = 'r' 
          AND c.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
          AND c.relname NOT IN ('tenants', 'tenant_configurations', 'tenant_usage_stats')
        GROUP BY c.relname, c.relrowsecurity
    ),
    summary AS (
        SELECT 
            COUNT(*) as total_tables,
            COUNT(*) FILTER (WHERE compliance_status = 'COMPLIANT') as compliant_tables,
            COUNT(*) FILTER (WHERE compliance_status != 'COMPLIANT') as non_compliant_tables,
            ROUND(
                (COUNT(*) FILTER (WHERE compliance_status = 'COMPLIANT')::NUMERIC / COUNT(*)) * 100, 
                2
            ) as compliance_percentage
        FROM rls_data
    ),
    detailed_issues AS (
        SELECT json_agg(
            json_build_object(
                'table_name', table_name,
                'rls_enabled', rls_enabled,
                'has_tenant_policy', has_tenant_policy,
                'compliance_status', compliance_status,
                'policies', COALESCE(policies, ARRAY[]::TEXT[])
            )
        ) as tables
        FROM rls_data
    )
    SELECT json_build_object(
        'check_type', 'rls_compliance',
        'timestamp', NOW(),
        'summary', json_build_object(
            'total_tables', s.total_tables,
            'compliant_tables', s.compliant_tables,
            'non_compliant_tables', s.non_compliant_tables,
            'compliance_percentage', s.compliance_percentage,
            'status', CASE 
                WHEN s.compliance_percentage = 100 THEN 'PASS'
                WHEN s.compliance_percentage >= 80 THEN 'WARNING'
                ELSE 'FAIL'
            END
        ),
        'details', d.tables
    ) INTO result
    FROM summary s, detailed_issues d;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 3. JSON PERFORMANCE INDEXES VALIDATION
-- =====================================================

CREATE OR REPLACE FUNCTION check_performance_indexes_json()
RETURNS JSON AS $$
DECLARE
    result JSON;
BEGIN
    WITH required_indexes AS (
        SELECT 
            t.table_name,
            c.column_name,
            CASE 
                WHEN c.column_name = 'tenant_id' THEN 'tenant_isolation'
                WHEN c.column_name = 'entity_id' THEN 'entity_lookup'
                WHEN c.column_name LIKE '%_date' THEN 'date_range'
                WHEN c.column_name = 'status' THEN 'status_filter'
                WHEN c.column_name LIKE '%_id' AND c.column_name != 'tenant_id' AND c.column_name != 'entity_id' THEN 'foreign_key'
                ELSE 'other'
            END as index_type,
            CASE 
                WHEN c.column_name IN ('tenant_id', 'entity_id') THEN 'CRITICAL'
                WHEN c.column_name = 'status' OR c.column_name LIKE '%_date' THEN 'HIGH'
                WHEN c.column_name LIKE '%_id' THEN 'MEDIUM'
                ELSE 'LOW'
            END as priority
        FROM information_schema.tables t
        JOIN information_schema.columns c ON t.table_name = c.table_name
        WHERE t.table_schema = 'public' 
          AND t.table_type = 'BASE TABLE'
          AND (c.column_name IN ('tenant_id', 'entity_id', 'created_at', 'updated_at', 'status')
            OR c.column_name LIKE '%_id'
            OR c.column_name LIKE '%_date')
    ),
    existing_indexes AS (
        SELECT 
            i.tablename as table_name,
            i.indexname as index_name,
            array_to_string(ARRAY(
                SELECT a.attname
                FROM pg_attribute a
                WHERE a.attrelid = ix.indrelid
                  AND a.attnum = ANY(ix.indkey)
                ORDER BY array_position(ix.indkey, a.attnum)
            ), ', ') as columns
        FROM pg_indexes i
        JOIN pg_class c ON c.relname = i.tablename
        JOIN pg_index ix ON ix.indexrelid = (SELECT oid FROM pg_class WHERE relname = i.indexname)
        WHERE i.schemaname = 'public'
    ),
    index_analysis AS (
        SELECT 
            ri.table_name,
            ri.column_name,
            ri.index_type,
            ri.priority,
            CASE 
                WHEN ei.index_name IS NOT NULL THEN 'EXISTS'
                WHEN ri.priority = 'CRITICAL' THEN 'CRITICAL_MISSING'
                WHEN ri.priority = 'HIGH' THEN 'HIGH_MISSING'
                ELSE 'RECOMMENDED_MISSING'
            END as status,
            ei.index_name,
            CASE 
                WHEN ei.index_name IS NULL AND ri.index_type = 'tenant_isolation' THEN 
                    'CREATE INDEX idx_' || ri.table_name || '_tenant ON ' || ri.table_name || '(tenant_id);'
                WHEN ei.index_name IS NULL AND ri.index_type = 'entity_lookup' THEN 
                    'CREATE INDEX idx_' || ri.table_name || '_entity ON ' || ri.table_name || '(entity_id);'
                WHEN ei.index_name IS NULL AND ri.index_type = 'status_filter' THEN 
                    'CREATE INDEX idx_' || ri.table_name || '_status ON ' || ri.table_name || '(status);'
                WHEN ei.index_name IS NULL AND ri.index_type = 'foreign_key' THEN 
                    'CREATE INDEX idx_' || ri.table_name || '_' || replace(ri.column_name, '_id', '') || ' ON ' || ri.table_name || '(' || ri.column_name || ');'
                ELSE NULL
            END as recommended_sql
        FROM required_indexes ri
        LEFT JOIN existing_indexes ei ON ri.table_name = ei.table_name 
            AND ei.columns LIKE '%' || ri.column_name || '%'
    ),
    summary AS (
        SELECT 
            COUNT(*) as total_indexes_needed,
            COUNT(*) FILTER (WHERE status = 'EXISTS') as existing_indexes,
            COUNT(*) FILTER (WHERE status LIKE '%MISSING%') as missing_indexes,
            COUNT(*) FILTER (WHERE status = 'CRITICAL_MISSING') as critical_missing,
            COUNT(*) FILTER (WHERE status = 'HIGH_MISSING') as high_missing,
            ROUND(
                (COUNT(*) FILTER (WHERE status = 'EXISTS')::NUMERIC / COUNT(*)) * 100, 
                2
            ) as compliance_percentage
        FROM index_analysis
    ),
    detailed_analysis AS (
        SELECT json_agg(
            json_build_object(
                'table_name', table_name,
                'column_name', column_name,
                'index_type', index_type,
                'priority', priority,
                'status', status,
                'existing_index', index_name,
                'recommended_sql', recommended_sql
            )
        ) as indexes
        FROM index_analysis
    )
    SELECT json_build_object(
        'check_type', 'performance_indexes',
        'timestamp', NOW(),
        'summary', json_build_object(
            'total_indexes_needed', s.total_indexes_needed,
            'existing_indexes', s.existing_indexes,
            'missing_indexes', s.missing_indexes,
            'critical_missing', s.critical_missing,
            'high_missing', s.high_missing,
            'compliance_percentage', s.compliance_percentage,
            'status', CASE 
                WHEN s.critical_missing = 0 AND s.compliance_percentage >= 90 THEN 'PASS'
                WHEN s.critical_missing = 0 AND s.compliance_percentage >= 70 THEN 'WARNING'
                ELSE 'FAIL'
            END
        ),
        'details', d.indexes
    ) INTO result
    FROM summary s, detailed_analysis d;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 4. JSON PERMISSIONS VALIDATION
-- =====================================================

CREATE OR REPLACE FUNCTION check_permissions_compliance_json()
RETURNS JSON AS $$
DECLARE
    result JSON;
BEGIN
    WITH permission_data AS (
        -- Check table permissions for application_role
        SELECT 
            'TABLE' as object_type,
            t.table_name as object_name,
            'application_role' as role_name,
            COALESCE(array_agg(tp.privilege_type) FILTER (WHERE tp.privilege_type IS NOT NULL), ARRAY[]::TEXT[]) as privileges,
            CASE 
                WHEN 'SELECT' = ANY(array_agg(tp.privilege_type)) 
                 AND 'INSERT' = ANY(array_agg(tp.privilege_type))
                 AND 'UPDATE' = ANY(array_agg(tp.privilege_type))
                 AND 'DELETE' = ANY(array_agg(tp.privilege_type)) THEN 'COMPLIANT'
                ELSE 'MISSING_PERMISSIONS'
            END as status
        FROM information_schema.tables t
        LEFT JOIN information_schema.table_privileges tp ON t.table_name = tp.table_name
            AND tp.grantee = 'application_role'
        WHERE t.table_schema = 'public' 
          AND t.table_type = 'BASE TABLE'
        GROUP BY t.table_name
        
        UNION ALL
        
        -- Check function permissions
        SELECT 
            'FUNCTION' as object_type,
            p.proname as object_name,
            'application_role' as role_name,
            ARRAY[CASE 
                WHEN has_function_privilege('application_role', p.oid, 'EXECUTE') THEN 'EXECUTE'
                ELSE 'NONE'
            END] as privileges,
            CASE 
                WHEN has_function_privilege('application_role', p.oid, 'EXECUTE') THEN 'COMPLIANT'
                ELSE 'MISSING_EXECUTE'
            END as status
        FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public'
          AND p.proname LIKE '%tenant%'
    ),
    summary AS (
        SELECT 
            COUNT(*) as total_objects,
            COUNT(*) FILTER (WHERE status LIKE 'COMPLIANT%') as compliant_objects,
            COUNT(*) FILTER (WHERE status LIKE 'MISSING%') as non_compliant_objects,
            COUNT(*) FILTER (WHERE object_type = 'TABLE') as total_tables,
            COUNT(*) FILTER (WHERE object_type = 'FUNCTION') as total_functions,
            ROUND(
                (COUNT(*) FILTER (WHERE status LIKE 'COMPLIANT%')::NUMERIC / COUNT(*)) * 100, 
                2
            ) as compliance_percentage
        FROM permission_data
    ),
    detailed_permissions AS (
        SELECT json_agg(
            json_build_object(
                'object_type', object_type,
                'object_name', object_name,
                'role_name', role_name,
                'privileges', privileges,
                'status', status
            )
        ) as permissions
        FROM permission_data
    )
    SELECT json_build_object(
        'check_type', 'permissions_compliance',
        'timestamp', NOW(),
        'summary', json_build_object(
            'total_objects', s.total_objects,
            'compliant_objects', s.compliant_objects,
            'non_compliant_objects', s.non_compliant_objects,
            'total_tables', s.total_tables,
            'total_functions', s.total_functions,
            'compliance_percentage', s.compliance_percentage,
            'status', CASE 
                WHEN s.compliance_percentage = 100 THEN 'PASS'
                WHEN s.compliance_percentage >= 90 THEN 'WARNING'
                ELSE 'FAIL'
            END
        ),
        'details', d.permissions
    ) INTO result
    FROM summary s, detailed_permissions d;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 5. JSON DOCUMENTATION COMPLIANCE
-- =====================================================

CREATE OR REPLACE FUNCTION check_documentation_compliance_json()
RETURNS JSON AS $$
DECLARE
    result JSON;
BEGIN
    WITH documentation_data AS (
        -- Check table comments
        SELECT 
            'TABLE' as object_type,
            t.table_name as object_name,
            CASE 
                WHEN obj_description(c.oid) IS NOT NULL THEN 'HAS_COMMENT'
                ELSE 'MISSING_COMMENT'
            END as comment_status,
            obj_description(c.oid) as comment_text
        FROM information_schema.tables t
        JOIN pg_class c ON c.relname = t.table_name
        WHERE t.table_schema = 'public' 
          AND t.table_type = 'BASE TABLE'
        
        UNION ALL
        
        -- Check critical column comments
        SELECT 
            'COLUMN' as object_type,
            (cols.table_name || '.' || cols.column_name) as object_name,
            CASE 
                WHEN col_description(c.oid, cols.ordinal_position) IS NOT NULL THEN 'HAS_COMMENT'
                ELSE 'MISSING_COMMENT'
            END as comment_status,
            col_description(c.oid, cols.ordinal_position) as comment_text
        FROM information_schema.columns cols
        JOIN pg_class c ON c.relname = cols.table_name
        WHERE cols.table_schema = 'public'
          AND cols.column_name IN ('entity_id', 'status', 'metadata', 'settings', 'amount', 'total_amount')
    ),
    summary AS (
        SELECT 
            COUNT(*) as total_objects,
            COUNT(*) FILTER (WHERE comment_status = 'HAS_COMMENT') as documented_objects,
            COUNT(*) FILTER (WHERE comment_status = 'MISSING_COMMENT') as undocumented_objects,
            COUNT(*) FILTER (WHERE object_type = 'TABLE') as total_tables,
            COUNT(*) FILTER (WHERE object_type = 'COLUMN') as total_columns,
            ROUND(
                (COUNT(*) FILTER (WHERE comment_status = 'HAS_COMMENT')::NUMERIC / COUNT(*)) * 100, 
                2
            ) as documentation_percentage
        FROM documentation_data
    ),
    detailed_documentation AS (
        SELECT json_agg(
            json_build_object(
                'object_type', object_type,
                'object_name', object_name,
                'comment_status', comment_status,
                'comment_text', comment_text
            )
        ) as documentation
        FROM documentation_data
    )
    SELECT json_build_object(
        'check_type', 'documentation_compliance',
        'timestamp', NOW(),
        'summary', json_build_object(
            'total_objects', s.total_objects,
            'documented_objects', s.documented_objects,
            'undocumented_objects', s.undocumented_objects,
            'total_tables', s.total_tables,
            'total_columns', s.total_columns,
            'documentation_percentage', s.documentation_percentage,
            'status', CASE 
                WHEN s.documentation_percentage >= 80 THEN 'PASS'
                WHEN s.documentation_percentage >= 60 THEN 'WARNING'
                ELSE 'FAIL'
            END
        ),
        'details', d.documentation
    ) INTO result
    FROM summary s, detailed_documentation d;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 6. JSON TRIGGERS VALIDATION
-- =====================================================

CREATE OR REPLACE FUNCTION check_required_triggers_json()
RETURNS JSON AS $$
DECLARE
    result JSON;
BEGIN
    WITH trigger_data AS (
        -- Check for updated_at triggers
        SELECT 
            t.table_name,
            'updated_at_trigger' as trigger_type,
            ('update_' || t.table_name || '_updated_at') as expected_trigger_name,
            EXISTS(
                SELECT 1 FROM information_schema.triggers tr
                WHERE tr.event_object_table = t.table_name
                  AND tr.trigger_name LIKE '%updated_at%'
            ) as exists,
            CASE 
                WHEN EXISTS(
                    SELECT 1 FROM information_schema.triggers tr
                    WHERE tr.event_object_table = t.table_name
                      AND tr.trigger_name LIKE '%updated_at%'
                ) THEN 'EXISTS'
                ELSE 'MISSING'
            END as status,
            'CREATE TRIGGER update_' || t.table_name || '_updated_at BEFORE UPDATE ON ' || t.table_name || ' FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();' as recommended_sql
        FROM information_schema.tables t
        JOIN information_schema.columns c ON t.table_name = c.table_name
        WHERE t.table_schema = 'public' 
          AND t.table_type = 'BASE TABLE'
          AND c.column_name = 'updated_at'
    ),
    summary AS (
        SELECT 
            COUNT(*) as total_triggers_needed,
            COUNT(*) FILTER (WHERE status = 'EXISTS') as existing_triggers,
            COUNT(*) FILTER (WHERE status = 'MISSING') as missing_triggers,
            ROUND(
                (COUNT(*) FILTER (WHERE status = 'EXISTS')::NUMERIC / COUNT(*)) * 100, 
                2
            ) as compliance_percentage
        FROM trigger_data
    ),
    detailed_triggers AS (
        SELECT json_agg(
            json_build_object(
                'table_name', table_name,
                'trigger_type', trigger_type,
                'expected_trigger_name', expected_trigger_name,
                'exists', exists,
                'status', status,
                'recommended_sql', recommended_sql
            )
        ) as triggers
        FROM trigger_data
    )
    SELECT json_build_object(
        'check_type', 'required_triggers',
        'timestamp', NOW(),
        'summary', json_build_object(
            'total_triggers_needed', s.total_triggers_needed,
            'existing_triggers', s.existing_triggers,
            'missing_triggers', s.missing_triggers,
            'compliance_percentage', s.compliance_percentage,
            'status', CASE 
                WHEN s.compliance_percentage = 100 THEN 'PASS'
                WHEN s.compliance_percentage >= 80 THEN 'WARNING'
                ELSE 'FAIL'
            END
        ),
        'details', d.triggers
    ) INTO result
    FROM summary s, detailed_triggers d;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 7. MASTER JSON VALIDATION REPORT
-- =====================================================

CREATE OR REPLACE FUNCTION generate_migration_validation_report_json()
RETURNS JSON AS $$
DECLARE
    result JSON;
    overall_status TEXT;
    total_checks INT;
    passed_checks INT;
    failed_checks INT;
    warning_checks INT;
BEGIN
    WITH all_checks AS (
        SELECT 'standard_columns' as check_name, check_standard_columns_json() as check_result
        UNION ALL
        SELECT 'rls_compliance' as check_name, check_rls_compliance_json() as check_result
        UNION ALL
        SELECT 'performance_indexes' as check_name, check_performance_indexes_json() as check_result
        UNION ALL
        SELECT 'permissions_compliance' as check_name, check_permissions_compliance_json() as check_result
        UNION ALL
        SELECT 'documentation_compliance' as check_name, check_documentation_compliance_json() as check_result
        UNION ALL
        SELECT 'required_triggers' as check_name, check_required_triggers_json() as check_result
    ),
    check_summary AS (
        SELECT 
            COUNT(*) as total,
            COUNT(*) FILTER (WHERE check_result->>'summary'->>'status' = 'PASS') as passed,
            COUNT(*) FILTER (WHERE check_result->>'summary'->>'status' = 'WARNING') as warnings,
            COUNT(*) FILTER (WHERE check_result->>'summary'->>'status' = 'FAIL') as failed
        FROM all_checks
    )
    SELECT 
        cs.total,
        cs.passed,
        cs.warnings,
        cs.failed
    INTO total_checks, passed_checks, warning_checks, failed_checks
    FROM check_summary cs;
    
    -- Determine overall status
    IF failed_checks > 0 THEN
        overall_status := 'FAIL';
    ELSIF warning_checks > 0 THEN
        overall_status := 'WARNING';
    ELSE
        overall_status := 'PASS';
    END IF;
    
    SELECT json_build_object(
        'migration_validation_report', json_build_object(
            'timestamp', NOW(),
            'tenant_id', current_tenant_id(),
            'overall_status', overall_status,
            'summary', json_build_object(
                'total_checks', total_checks,
                'passed_checks', passed_checks,
                'warning_checks', warning_checks,
                'failed_checks', failed_checks,
                'success_rate', ROUND((passed_checks::NUMERIC / total_checks) * 100, 2)
            ),
            'individual_checks', json_object_agg(
                ac.check_name, 
                ac.check_result
            ),
            'recommendations', json_build_array(
                CASE 
                    WHEN failed_checks > 0 THEN 'Critical issues found - deployment should be blocked'
                    WHEN warning_checks > 0 THEN 'Warnings found - review and plan fixes'
                    ELSE 'All checks passed - deployment approved'
                END,
                'Run validation checks regularly in development',
                'Address critical issues before production deployment',
                'Monitor compliance trends over time'
            )
        )
    ) INTO result
    FROM all_checks ac;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 8. COMPREHENSIVE JSON VALIDATION EXECUTOR
-- =====================================================

CREATE OR REPLACE FUNCTION run_json_validation_suite()
RETURNS JSON AS $$
DECLARE
    start_time TIMESTAMPTZ;
    end_time TIMESTAMPTZ;
    execution_time_ms NUMERIC;
    validation_result JSON;
BEGIN
    start_time := NOW();
    
    -- Run the comprehensive validation
    SELECT generate_migration_validation_report_json() INTO validation_result;
    
    end_time := NOW();
    execution_time_ms := EXTRACT(EPOCH FROM (end_time - start_time)) * 1000;
    
    -- Add execution metadata
    SELECT json_build_object(
        'validation_suite_result', validation_result,
        'execution_metadata', json_build_object(
            'start_time', start_time,
            'end_time', end_time,
            'execution_time_ms', execution_time_ms,
            'database_version', version(),
            'current_user', current_user,
            'current_database', current_database()
        )
    ) INTO validation_result;
    
    RETURN validation_result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 9. JSON OUTPUT HELPER FUNCTIONS
-- =====================================================

-- Pretty print JSON with proper formatting
CREATE OR REPLACE FUNCTION pretty_print_validation_json()
RETURNS TEXT AS $$
DECLARE
    json_result JSON;
    formatted_output TEXT;
BEGIN
    SELECT run_json_validation_suite() INTO json_result;
    
    -- Format JSON with proper indentation (PostgreSQL 12+)
    SELECT jsonb_pretty(json_result::jsonb) INTO formatted_output;
    
    RETURN formatted_output;
END;
$$ LANGUAGE plpgsql;

-- Get only failed checks in JSON format
CREATE OR REPLACE FUNCTION get_failed_checks_json()
RETURNS JSON AS $$
DECLARE
    result JSON;
BEGIN
    WITH validation_data AS (
        SELECT run_json_validation_suite() as full_result
    ),
    failed_checks AS (
        SELECT 
            key as check_name,
            value as check_details
        FROM validation_data,
        json_each(full_result->'validation_suite_result'->'migration_validation_report'->'individual_checks')
        WHERE value->>'summary'->>'status' IN ('FAIL', 'WARNING')
    )
    SELECT json_build_object(
        'failed_checks_report', json_build_object(
            'timestamp', NOW(),
            'tenant_id', current_tenant_id(),
            'failed_checks', json_object_agg(fc.check_name, fc.check_details)
        )
    ) INTO result
    FROM failed_checks fc;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 10. USAGE EXAMPLES AND OUTPUT
-- =====================================================

/*
JSON VALIDATION USAGE EXAMPLES:

1. FULL VALIDATION SUITE (PRETTY FORMATTED):
   SELECT pretty_print_validation_json();

2. FULL VALIDATION SUITE (COMPACT JSON):
   SELECT run_json_validation_suite();

3. INDIVIDUAL CHECK RESULTS:
   SELECT check_standard_columns_json();
   SELECT check_rls_compliance_json();
   SELECT check_performance_indexes_json();

4. ONLY FAILED/WARNING CHECKS:
   SELECT get_failed_checks_json();

5. SAVE TO FILE (psql):
   \o validation_report.json
   SELECT run_json_validation_suite();
   \o

6. API INTEGRATION EXAMPLE (Node.js):
   const result = await db.query('SELECT run_json_validation_suite() as report');
   const validationReport = result.rows[0].report;
   
   if (validationReport.validation_suite_result.migration_validation_report.overall_status === 'FAIL') {
       throw new Error('Migration validation failed');
   }

7. CI/CD PIPELINE USAGE:
   psql -d $DATABASE_URL -t -c "SELECT run_json_validation_suite();" > validation_report.json
   
   # Parse with jq
   jq '.validation_suite_result.migration_validation_report.overall_status' validation_report.json

8. MONITORING INTEGRATION:
   # Send to monitoring system
   curl -X POST https://monitoring.example.com/webhooks/validation \
        -H "Content-Type: application/json" \
        -d "$(psql -d $DATABASE_URL -t -c 'SELECT run_json_validation_suite();')"

SAMPLE OUTPUT STRUCTURE:
{
  "validation_suite_result": {
    "migration_validation_report": {
      "timestamp": "2025-01-21T10:30:00Z",
      "tenant_id": "123e4567-e89b-12d3-a456-426614174000",
      "overall_status": "PASS",
      "summary": {
        "total_checks": 6,
        "passed_checks": 5,
        "warning_checks": 1,
        "failed_checks": 0,
        "success_rate": 83.33
      },
      "individual_checks": {
        "standard_columns": {
          "check_type": "standard_columns",
          "summary": {
            "compliance_percentage": 100.00,
            "status": "PASS"
          },
          "details": [...]
        }
      }
    }
  },
  "execution_metadata": {
    "execution_time_ms": 45.23
  }
}
*/

-- Grant permissions for JSON validation functions
GRANT EXECUTE ON FUNCTION check_standard_columns_json() TO application_role;
GRANT EXECUTE ON FUNCTION check_rls_compliance_json() TO application_role;
GRANT EXECUTE ON FUNCTION check_performance_indexes_json() TO application_role;
GRANT EXECUTE ON FUNCTION check_permissions_compliance_json() TO application_role;
GRANT EXECUTE ON FUNCTION check_documentation_compliance_json() TO application_role;
GRANT EXECUTE ON FUNCTION check_required_triggers_json() TO application_role;
GRANT EXECUTE ON FUNCTION generate_migration_validation_report_json() TO application_role;
GRANT EXECUTE ON FUNCTION run_json_validation_suite() TO application_role;
GRANT EXECUTE ON FUNCTION pretty_print_validation_json() TO application_role;
GRANT EXECUTE ON FUNCTION get_failed_checks_json() TO application_role;

-- =====================================================
-- EXECUTE JSON VALIDATION
-- =====================================================

-- Uncomment to run validation and see JSON output:
-- SELECT pretty_print_validation_json();
