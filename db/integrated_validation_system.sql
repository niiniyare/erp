-- =====================================================
-- INTEGRATED VALIDATION SYSTEM
-- =====================================================
-- Combines structural validation with business logic validation
-- for ERP/Accounting system validation
-- =====================================================

-- =====================================================
-- MASTER VALIDATION ORCHESTRATOR
-- =====================================================

CREATE OR REPLACE FUNCTION run_comprehensive_validation()
RETURNS TABLE(
    validation_phase TEXT,
    check_name TEXT,
    status TEXT,
    critical_issues INT,
    high_issues INT,
    medium_issues INT,
    low_issues INT,
    total_issues INT,
    compliance_percentage NUMERIC(5,2)
) AS $$
BEGIN
    -- Create temporary table to store all results
    DROP TABLE IF EXISTS temp_validation_results;
    CREATE TEMP TABLE temp_validation_results (
        phase TEXT,
        check_name TEXT,
        status TEXT,
        critical_count INT DEFAULT 0,
        high_count INT DEFAULT 0,
        medium_count INT DEFAULT 0,
        low_count INT DEFAULT 0,
        total_count INT DEFAULT 0,
        compliance_pct NUMERIC(5,2) DEFAULT 100.00
    );

    -- Phase 1: Structural Validations
    INSERT INTO temp_validation_results (phase, check_name, status, total_count, compliance_pct)
    SELECT 
        'STRUCTURAL',
        check_category,
        status,
        non_compliant_items,
        compliance_percentage
    FROM generate_migration_validation_report();

    -- Phase 2: Business Logic Validations  
    INSERT INTO temp_validation_results (phase, check_name, status, critical_count, high_count, medium_count, low_count, total_count)
    SELECT 
        'BUSINESS_LOGIC',
        validation_category,
        status,
        critical_issues,
        high_issues,
        medium_issues,
        low_issues,
        total_issues
    FROM generate_business_logic_report()
    WHERE validation_category != 'OVERALL_SUMMARY';

    -- Return consolidated results
    RETURN QUERY
    SELECT 
        tvr.phase::TEXT as validation_phase,
        tvr.check_name::TEXT,
        tvr.status::TEXT,
        tvr.critical_count::INT as critical_issues,
        tvr.high_count::INT as high_issues,
        tvr.medium_count::INT as medium_issues,
        tvr.low_count::INT as low_issues,
        tvr.total_count::INT as total_issues,
        tvr.compliance_pct::NUMERIC(5,2) as compliance_percentage
    FROM temp_validation_results tvr
    ORDER BY 
        CASE tvr.phase 
            WHEN 'STRUCTURAL' THEN 1 
            WHEN 'BUSINESS_LOGIC' THEN 2 
            ELSE 3 
        END,
        tvr.critical_count DESC,
        tvr.high_count DESC;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- DEPLOYMENT READINESS CHECK
-- =====================================================

CREATE OR REPLACE FUNCTION check_deployment_readiness()
RETURNS TABLE(
    overall_status TEXT,
    deployment_decision TEXT,
    blockers TEXT[],
    warnings TEXT[],
    recommendations TEXT[]
) AS $$
DECLARE
    critical_structural INT := 0;
    critical_business INT := 0;
    high_structural INT := 0;
    high_business INT := 0;
    blocker_list TEXT[] := ARRAY[]::TEXT[];
    warning_list TEXT[] := ARRAY[]::TEXT[];
    recommendation_list TEXT[] := ARRAY[]::TEXT[];
    deployment_status TEXT;
    decision TEXT;
BEGIN
    -- Count critical issues
    SELECT 
        COUNT(*) FILTER (WHERE validation_phase = 'STRUCTURAL' AND status = 'FAIL'),
        COUNT(*) FILTER (WHERE validation_phase = 'BUSINESS_LOGIC' AND critical_issues > 0),
        COUNT(*) FILTER (WHERE validation_phase = 'STRUCTURAL' AND status = 'WARNING'),
        COUNT(*) FILTER (WHERE validation_phase = 'BUSINESS_LOGIC' AND high_issues > 0)
    INTO critical_structural, critical_business, high_structural, high_business
    FROM run_comprehensive_validation();

    -- Determine blockers
    IF critical_structural > 0 THEN
        blocker_list := array_append(blocker_list, 'Critical structural issues: Missing required columns, RLS policies, or permissions');
    END IF;

    IF critical_business > 0 THEN
        blocker_list := array_append(blocker_list, 'Critical business logic issues: Data integrity violations or financial inconsistencies');
    END IF;

    -- Determine warnings
    IF high_structural > 0 THEN
        warning_list := array_append(warning_list, 'Structural warnings: Performance indexes or documentation missing');
    END IF;

    IF high_business > 0 THEN
        warning_list := array_append(warning_list, 'Business logic warnings: Workflow violations or compliance issues');
    END IF;

    -- Determine overall status and decision
    IF array_length(blocker_list, 1) > 0 THEN
        deployment_status := 'BLOCKED';
        decision := 'DO NOT DEPLOY - Critical issues must be resolved first';
    ELSIF array_length(warning_list, 1) > 2 THEN
        deployment_status := 'CONDITIONAL';
        decision := 'DEPLOY WITH CAUTION - Monitor closely and plan fixes';
    ELSE
        deployment_status := 'READY';
        decision := 'APPROVED FOR DEPLOYMENT';
    END IF;

    -- Generate recommendations
    recommendation_list := ARRAY[
        'Run validation checks daily in production',
        'Monitor business logic reports weekly',
        'Address warnings in next maintenance window',
        'Implement automated alerting for critical issues'
    ];

    RETURN QUERY
    SELECT 
        deployment_status,
        decision,
        blocker_list,
        warning_list,
        recommendation_list;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- AUTOMATED MONITORING SETUP
-- =====================================================

-- Create monitoring views for production use
CREATE OR REPLACE VIEW v_daily_health_check AS
SELECT 
    CURRENT_DATE as check_date,
    validation_phase,
    check_name,
    critical_issues,
    high_issues,
    total_issues,
    CASE 
        WHEN critical_issues > 0 THEN 'ALERT'
        WHEN high_issues > 5 THEN 'WARNING'
        WHEN total_issues = 0 THEN 'HEALTHY'
        ELSE 'MONITOR'
    END as health_status
FROM run_comprehensive_validation()
WHERE total_issues > 0 OR validation_phase = 'STRUCTURAL';

-- Grant permissions on monitoring view
GRANT SELECT ON v_daily_health_check TO application_role;

-- =====================================================
-- CUSTOM BUSINESS RULES ENGINE
-- =====================================================

-- Table to store custom business rules
CREATE TABLE IF NOT EXISTS business_validation_rules (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    rule_name VARCHAR(100) NOT NULL,
    rule_description TEXT,
    validation_sql TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'MEDIUM'
        CHECK (severity IN ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW')),
    
    is_active BOOLEAN DEFAULT true,
    rule_category VARCHAR(50) DEFAULT 'CUSTOM',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE (tenant_id, rule_name)
);

-- Enable RLS on business rules
ALTER TABLE business_validation_rules ENABLE ROW LEVEL SECURITY;
CREATE POLICY business_validation_rules_tenant_policy ON business_validation_rules
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- Function to execute custom business rules
CREATE OR REPLACE FUNCTION run_custom_business_rules()
RETURNS TABLE(
    validation_type TEXT,
    entity_type TEXT,
    entity_id UUID,
    issue_description TEXT,
    severity TEXT,
    fix_recommendation TEXT
) AS $$
DECLARE
    rule_rec RECORD;
    dynamic_sql TEXT;
BEGIN
    -- Loop through active custom rules for current tenant
    FOR rule_rec IN 
        SELECT * FROM business_validation_rules 
        WHERE tenant_id = current_tenant_id() AND is_active = true
    LOOP
        -- Execute the custom validation SQL
        -- Note: In production, you'd want additional security checks here
        BEGIN
            dynamic_sql := rule_rec.validation_sql;
            
            -- Return results from custom rule
            RETURN QUERY EXECUTE format('
                SELECT 
                    %L::TEXT as validation_type,
                    ''CUSTOM''::TEXT as entity_type,
                    NULL::UUID as entity_id,
                    %L::TEXT as issue_description,
                    %L::TEXT as severity,
                    ''Review custom rule: %s''::TEXT as fix_recommendation
                WHERE EXISTS (%s)',
                rule_rec.rule_category,
                rule_rec.rule_description,
                rule_rec.severity,
                rule_rec.rule_name,
                dynamic_sql
            );
        EXCEPTION
            WHEN OTHERS THEN
                -- Log error and continue with next rule
                RETURN QUERY
                SELECT 
                    'CUSTOM_RULE_ERROR'::TEXT,
                    'VALIDATION_RULE'::TEXT,
                    rule_rec.id,
                    ('Error executing rule: ' || rule_rec.rule_name || ' - ' || SQLERRM)::TEXT,
                    'HIGH'::TEXT,
                    'Fix SQL syntax in custom validation rule'::TEXT;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- ENHANCED BUSINESS LOGIC WITH CUSTOM RULES
-- =====================================================

-- Enhanced business logic validation including custom rules
CREATE OR REPLACE FUNCTION check_all_business_logic_enhanced()
RETURNS TABLE(
    validation_type TEXT,
    entity_type TEXT,
    entity_id UUID,
    issue_description TEXT,
    severity TEXT,
    fix_recommendation TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Standard business logic validations
    SELECT * FROM check_all_business_logic()
    
    UNION ALL
    
    -- Custom business rules
    SELECT * FROM run_custom_business_rules()
    
    ORDER BY 
        CASE severity 
            WHEN 'CRITICAL' THEN 1 
            WHEN 'HIGH' THEN 2 
            WHEN 'MEDIUM' THEN 3 
            ELSE 4 
        END,
        validation_type,
        entity_type;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- SAMPLE CUSTOM BUSINESS RULES
-- =====================================================

-- Function to create sample custom business rules
CREATE OR REPLACE FUNCTION create_sample_business_rules(p_tenant_id UUID)
RETURNS VOID AS $$
BEGIN
    INSERT INTO business_validation_rules (
        tenant_id, rule_name, rule_description, validation_sql, severity, rule_category
    ) VALUES
    (
        p_tenant_id,
        'Large Invoice Approval Required',
        'Invoices over $10,000 require approval documentation',
        'SELECT 1 FROM sales_invoices WHERE tenant_id = current_tenant_id() AND total_amount > 10000 AND (notes IS NULL OR notes NOT LIKE ''%approved%'')',
        'HIGH',
        'APPROVAL_WORKFLOW'
    ),
    (
        p_tenant_id,
        'Vendor Payment Terms Consistency',
        'Vendor payment terms should match invoice payment terms',
        'SELECT 1 FROM purchase_invoices pi JOIN vendors v ON pi.vendor_id = v.id WHERE pi.tenant_id = current_tenant_id() AND pi.payment_terms != v.payment_terms',
        'MEDIUM',
        'DATA_CONSISTENCY'
    ),
    (
        p_tenant_id,
        'Duplicate Customer Check',
        'Check for potential duplicate customers by name similarity',
        'SELECT 1 FROM customers c1 JOIN customers c2 ON c1.tenant_id = c2.tenant_id AND c1.id != c2.id WHERE c1.tenant_id = current_tenant_id() AND similarity(c1.name, c2.name) > 0.8',
        'LOW',
        'DATA_QUALITY'
    )
    ON CONFLICT (tenant_id, rule_name) DO NOTHING;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- VALIDATION SCHEDULING AND AUTOMATION
-- =====================================================

-- Table to track validation runs
CREATE TABLE IF NOT EXISTS validation_run_history (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    run_type VARCHAR(50) NOT NULL, -- 'MANUAL', 'SCHEDULED', 'DEPLOYMENT'
    run_status VARCHAR(20) NOT NULL, -- 'RUNNING', 'COMPLETED', 'FAILED'
    
    structural_issues_count INT DEFAULT 0,
    business_logic_issues_count INT DEFAULT 0,
    critical_issues_count INT DEFAULT 0,
    
    execution_time_ms INT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Enable RLS on validation history
ALTER TABLE validation_run_history ENABLE ROW LEVEL SECURITY;
CREATE POLICY validation_run_history_tenant_policy ON validation_run_history
    FOR ALL TO application_role
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- Function to log validation runs
CREATE OR REPLACE FUNCTION log_validation_run(
    p_run_type TEXT DEFAULT 'MANUAL'
)
RETURNS UUID AS $$
DECLARE
    run_id UUID;
    start_time TIMESTAMPTZ;
    structural_count INT;
    business_count INT;
    critical_count INT;
BEGIN
    start_time := NOW();
    
    -- Create run record
    INSERT INTO validation_run_history (tenant_id, run_type, run_status)
    VALUES (current_tenant_id(), p_run_type, 'RUNNING')
    RETURNING id INTO run_id;
    
    -- Count issues
    SELECT 
        COUNT(*) FILTER (WHERE validation_phase = 'STRUCTURAL' AND total_issues > 0),
        COUNT(*) FILTER (WHERE validation_phase = 'BUSINESS_LOGIC' AND total_issues > 0),
        COUNT(*) FILTER (WHERE critical_issues > 0)
    INTO structural_count, business_count, critical_count
    FROM run_comprehensive_validation();
    
    -- Update run record
    UPDATE validation_run_history
    SET 
        run_status = 'COMPLETED',
        structural_issues_count = structural_count,
        business_logic_issues_count = business_count,
        critical_issues_count = critical_count,
        execution_time_ms = EXTRACT(EPOCH FROM (NOW() - start_time)) * 1000,
        completed_at = NOW()
    WHERE id = run_id;
    
    RETURN run_id;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- ALERTING AND NOTIFICATION SYSTEM
-- =====================================================

-- Function to generate alerts for critical issues
CREATE OR REPLACE FUNCTION generate_validation_alerts()
RETURNS TABLE(
    alert_level TEXT,
    alert_message TEXT,
    affected_entities INT,
    recommended_action TEXT
) AS $$
DECLARE
    critical_count INT;
    high_count INT;
BEGIN
    -- Count critical and high issues
    SELECT 
        COUNT(*) FILTER (WHERE critical_issues > 0),
        COUNT(*) FILTER (WHERE high_issues > 0)
    INTO critical_count, high_count
    FROM run_comprehensive_validation();
    
    -- Generate alerts based on issue counts
    IF critical_count > 0 THEN
        RETURN QUERY
        SELECT 
            'CRITICAL'::TEXT,
            'Critical validation issues detected requiring immediate attention'::TEXT,
            critical_count,
            'Review and fix critical issues before proceeding with any operations'::TEXT;
    END IF;
    
    IF high_count > 5 THEN
        RETURN QUERY
        SELECT 
            'WARNING'::TEXT,
            'Multiple high-priority validation issues detected'::TEXT,
            high_count,
            'Schedule maintenance window to address high-priority issues'::TEXT;
    END IF;
    
    -- If no issues, return success message
    IF critical_count = 0 AND high_count <= 5 THEN
        RETURN QUERY
        SELECT 
            'INFO'::TEXT,
            'Validation checks passed successfully'::TEXT,
            0,
            'Continue with normal operations'::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- COMPREHENSIVE USAGE EXAMPLES
-- =====================================================

/*
COMPLETE VALIDATION WORKFLOW:

1. BASIC VALIDATION:
   SELECT * FROM run_comprehensive_validation();

2. DEPLOYMENT READINESS:
   SELECT * FROM check_deployment_readiness();

3. DAILY HEALTH CHECK:
   SELECT * FROM v_daily_health_check;

4. LOG VALIDATION RUN:
   SELECT log_validation_run('SCHEDULED');

5. GENERATE ALERTS:
   SELECT * FROM generate_validation_alerts();

6. CUSTOM BUSINESS RULES:
   -- Add custom rule
   INSERT INTO business_validation_rules (
       tenant_id, rule_name, rule_description, validation_sql, severity
   ) VALUES (
       current_tenant_id(),
       'Weekend Invoice Check',
       'Invoices should not be dated on weekends',
       'SELECT 1 FROM sales_invoices WHERE EXTRACT(DOW FROM invoice_date) IN (0,6)',
       'LOW'
   );
   
   -- Run enhanced validation
   SELECT * FROM check_all_business_logic_enhanced();

7. FULL INTEGRATION EXAMPLE:
   DO $$
   DECLARE
       run_id UUID;
       deployment_status RECORD;
   BEGIN
       -- Log the validation run
       run_id := log_validation_run('DEPLOYMENT');
       
       -- Check deployment readiness
       SELECT * INTO deployment_status FROM check_deployment_readiness();
       
       -- Output results
       RAISE NOTICE 'Validation Run ID: %', run_id;
       RAISE NOTICE 'Deployment Status: %', deployment_status.overall_status;
       RAISE NOTICE 'Decision: %', deployment_status.deployment_decision;
       
       -- Show alerts
       FOR rec IN SELECT * FROM generate_validation_alerts() LOOP
           RAISE NOTICE 'ALERT [%]: %', rec.alert_level, rec.alert_message;
       END LOOP;
   END;
   $$;

MONITORING QUERIES:

-- Weekly validation trend
SELECT 
    DATE_TRUNC('week', created_at) as week,
    AVG(critical_issues_count) as avg_critical,
    AVG(execution_time_ms) as avg_execution_time
FROM validation_run_history
WHERE tenant_id = current_tenant_id()
GROUP BY week
ORDER BY week DESC;

-- Most common validation issues
SELECT 
    validation_type,
    COUNT(*) as occurrence_count,
    severity
FROM check_all_business_logic_enhanced()
GROUP BY validation_type, severity
ORDER BY occurrence_count DESC;

CI/CD INTEGRATION:

#!/bin/bash
# deployment-validation.sh

echo "Running validation..."
psql -d $DATABASE_URL -c "
DO \$\$
DECLARE
    deployment_ready BOOLEAN := false;
    critical_issues INT := 0;
BEGIN
    SELECT 
        overall_status = 'READY' OR overall_status = 'CONDITIONAL'
    INTO deployment_ready
    FROM check_deployment_readiness();
    
    SELECT COUNT(*) INTO critical_issues
    FROM run_comprehensive_validation()
    WHERE critical_issues > 0;
    
    IF NOT deployment_ready OR critical_issues > 0 THEN
        RAISE EXCEPTION 'Deployment blocked: Critical validation failures detected';
    ELSE
        RAISE NOTICE 'Validation passed: Deployment approved';
    END IF;
END;
\$\$;
"

echo "Validation completed successfully"
*/

-- Grant permissions for all new functions
GRANT EXECUTE ON FUNCTION run_comprehensive_validation() TO application_role;
GRANT EXECUTE ON FUNCTION check_deployment_readiness() TO application_role;
GRANT EXECUTE ON FUNCTION run_custom_business_rules() TO application_role;
GRANT EXECUTE ON FUNCTION check_all_business_logic_enhanced() TO application_role;
GRANT EXECUTE ON FUNCTION create_sample_business_rules(UUID) TO application_role;
GRANT EXECUTE ON FUNCTION log_validation_run(TEXT) TO application_role;
GRANT EXECUTE ON FUNCTION generate_validation_alerts() TO application_role;

-- Grant table permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON business_validation_rules TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON validation_run_history TO application_role;
