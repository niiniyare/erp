-- =====================================================
-- BUSINESS LOGIC VALIDATION RULES
-- =====================================================
-- Domain-specific validation rules for accounting/ERP system
-- Validates business rules, data integrity, and compliance requirements
-- =====================================================

-- =====================================================
-- 1. FINANCIAL DATA INTEGRITY VALIDATIONS
-- =====================================================

-- Validate invoice totals match line items
CREATE OR REPLACE FUNCTION check_invoice_line_totals()
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
    -- Check sales invoice totals
    WITH sales_discrepancies AS (
        SELECT 
            si.id,
            si.invoice_number,
            si.total_amount as header_total,
            COALESCE(SUM(sil.line_total), 0) as calculated_total,
            ABS(si.total_amount - COALESCE(SUM(sil.line_total), 0)) as discrepancy
        FROM sales_invoices si
        LEFT JOIN sales_invoice_lines sil ON si.id = sil.invoice_id
        WHERE si.tenant_id = current_tenant_id()
          AND si.status != 'draft'
        GROUP BY si.id, si.invoice_number, si.total_amount
        HAVING ABS(si.total_amount - COALESCE(SUM(sil.line_total), 0)) > 0.01
    )
    SELECT 
        'FINANCIAL_INTEGRITY'::TEXT,
        'SALES_INVOICE'::TEXT,
        sd.id,
        'Invoice total (' || sd.header_total || ') does not match line items total (' || sd.calculated_total || '). Discrepancy: ' || sd.discrepancy,
        CASE 
            WHEN sd.discrepancy > 100 THEN 'CRITICAL'
            WHEN sd.discrepancy > 1 THEN 'HIGH'
            ELSE 'MEDIUM'
        END::TEXT,
        'Recalculate invoice total or review line items for invoice: ' || sd.invoice_number
    FROM sales_discrepancies sd
    
    UNION ALL
    
    -- Check purchase invoice totals (when purchase_invoice_lines exists)
    SELECT 
        'FINANCIAL_INTEGRITY'::TEXT,
        'PURCHASE_INVOICE'::TEXT,
        pi.id,
        'Purchase invoice total validation needed - line items table may be missing',
        'MEDIUM'::TEXT,
        'Implement purchase_invoice_lines table and validation'
    FROM purchase_invoices pi
    WHERE pi.tenant_id = current_tenant_id()
      AND NOT EXISTS (
          SELECT 1 FROM information_schema.tables 
          WHERE table_name = 'purchase_invoice_lines'
      )
    LIMIT 1; -- Just show the issue once
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 2. PAYMENT APPLICATION VALIDATIONS
-- =====================================================

-- Validate payment applications don't exceed invoice amounts
CREATE OR REPLACE FUNCTION check_payment_application_integrity()
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
    -- Check for over-applied payments to sales invoices
    WITH sales_overpayments AS (
        SELECT 
            si.id,
            si.invoice_number,
            si.total_amount,
            COALESCE(SUM(pa.applied_amount), 0) as total_applied,
            COALESCE(SUM(pa.applied_amount), 0) - si.total_amount as overpayment
        FROM sales_invoices si
        LEFT JOIN payment_applications pa ON si.id = pa.sales_invoice_id
        WHERE si.tenant_id = current_tenant_id()
        GROUP BY si.id, si.invoice_number, si.total_amount
        HAVING COALESCE(SUM(pa.applied_amount), 0) > si.total_amount
    )
    SELECT 
        'PAYMENT_INTEGRITY'::TEXT,
        'SALES_INVOICE'::TEXT,
        so.id,
        'Invoice ' || so.invoice_number || ' is over-applied by ' || so.overpayment || 
        '. Invoice total: ' || so.total_amount || ', Applied: ' || so.total_applied,
        'CRITICAL'::TEXT,
        'Review payment applications and adjust amounts'
    FROM sales_overpayments so
    
    UNION ALL
    
    -- Check for payments with invalid amounts
    SELECT 
        'PAYMENT_INTEGRITY'::TEXT,
        'PAYMENT'::TEXT,
        p.id,
        'Payment ' || p.payment_number || ' has invalid amount: ' || p.amount,
        'HIGH'::TEXT,
        'Payment amounts must be positive'
    FROM payments p
    WHERE p.tenant_id = current_tenant_id()
      AND p.amount <= 0;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 3. BUSINESS WORKFLOW VALIDATIONS
-- =====================================================

-- Validate status transitions and business rules
CREATE OR REPLACE FUNCTION check_business_workflow_integrity()
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
    -- Check for invoices with invalid status progressions
    SELECT 
        'WORKFLOW_INTEGRITY'::TEXT,
        'SALES_INVOICE'::TEXT,
        si.id,
        'Invoice ' || si.invoice_number || ' has status "' || si.status || 
        '" but amount_due (' || si.amount_due || ') does not match expected value',
        'HIGH'::TEXT,
        CASE 
            WHEN si.status = 'paid' AND si.amount_due > 0 THEN 'Set amount_due to 0 for paid invoices'
            WHEN si.status = 'draft' AND si.amount_due != si.total_amount THEN 'Set amount_due = total_amount for draft invoices'
            ELSE 'Review invoice status and amount_due consistency'
        END
    FROM sales_invoices si
    WHERE si.tenant_id = current_tenant_id()
      AND (
          (si.status = 'paid' AND si.amount_due > 0) OR
          (si.status = 'draft' AND si.amount_due != si.total_amount) OR
          (si.status = 'void' AND si.amount_due > 0)
      )
    
    UNION ALL
    
    -- Check for invoices with future due dates in overdue status
    SELECT 
        'WORKFLOW_INTEGRITY'::TEXT,
        'SALES_INVOICE'::TEXT,
        si.id,
        'Invoice ' || si.invoice_number || ' has status "overdue" but due date (' || 
        si.due_date || ') is in the future',
        'MEDIUM'::TEXT,
        'Update status to "issued" or correct due date'
    FROM sales_invoices si
    WHERE si.tenant_id = current_tenant_id()
      AND si.status = 'overdue'
      AND si.due_date > CURRENT_DATE
    
    UNION ALL
    
    -- Check for customers with credit limit violations
    SELECT 
        'BUSINESS_RULES'::TEXT,
        'CUSTOMER'::TEXT,
        c.id,
        'Customer ' || c.name || ' has outstanding balance (' || 
        COALESCE(SUM(si.amount_due), 0) || ') exceeding credit limit (' || 
        COALESCE(c.credit_limit, 0) || ')',
        'HIGH'::TEXT,
        'Review credit limit or collect outstanding payments'
    FROM customers c
    LEFT JOIN sales_invoices si ON c.id = si.customer_id AND si.status IN ('issued', 'partial', 'overdue')
    WHERE c.tenant_id = current_tenant_id()
      AND c.credit_limit IS NOT NULL
      AND c.credit_limit > 0
    GROUP BY c.id, c.name, c.credit_limit
    HAVING COALESCE(SUM(si.amount_due), 0) > c.credit_limit;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 4. REFERENTIAL INTEGRITY VALIDATIONS
-- =====================================================

-- Check for orphaned records and broken relationships
CREATE OR REPLACE FUNCTION check_referential_integrity()
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
    -- Check for orphaned invoice lines
    SELECT 
        'REFERENTIAL_INTEGRITY'::TEXT,
        'SALES_INVOICE_LINE'::TEXT,
        sil.id,
        'Invoice line references non-existent invoice: ' || sil.invoice_id,
        'CRITICAL'::TEXT,
        'Delete orphaned line or restore missing invoice'
    FROM sales_invoice_lines sil
    WHERE sil.tenant_id = current_tenant_id()
      AND NOT EXISTS (
          SELECT 1 FROM sales_invoices si 
          WHERE si.id = sil.invoice_id AND si.tenant_id = current_tenant_id()
      )
    
    UNION ALL
    
    -- Check for orphaned payment applications
    SELECT 
        'REFERENTIAL_INTEGRITY'::TEXT,
        'PAYMENT_APPLICATION'::TEXT,
        pa.id,
        'Payment application references non-existent payment or invoice',
        'CRITICAL'::TEXT,
        'Delete orphaned application or restore missing references'
    FROM payment_applications pa
    WHERE pa.tenant_id = current_tenant_id()
      AND (
          (pa.sales_invoice_id IS NOT NULL AND NOT EXISTS (
              SELECT 1 FROM sales_invoices si 
              WHERE si.id = pa.sales_invoice_id AND si.tenant_id = current_tenant_id()
          )) OR
          NOT EXISTS (
              SELECT 1 FROM payments p 
              WHERE p.id = pa.payment_id AND p.tenant_id = current_tenant_id()
          )
      )
    
    UNION ALL
    
    -- Check for customers/vendors without proper tenant isolation
    SELECT 
        'REFERENTIAL_INTEGRITY'::TEXT,
        'CUSTOMER'::TEXT,
        c.id,
        'Customer ' || c.name || ' may have cross-tenant data contamination',
        'CRITICAL'::TEXT,
        'Verify tenant_id is correct and RLS policies are working'
    FROM customers c
    WHERE c.tenant_id != current_tenant_id()
      AND EXISTS (
          SELECT 1 FROM sales_invoices si 
          WHERE si.customer_id = c.id AND si.tenant_id = current_tenant_id()
      );
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 5. CURRENCY AND EXCHANGE RATE VALIDATIONS
-- =====================================================

-- Validate currency consistency and exchange rates
CREATE OR REPLACE FUNCTION check_currency_integrity()
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
    -- Check for currency mismatches between customer and invoices
    SELECT 
        'CURRENCY_INTEGRITY'::TEXT,
        'SALES_INVOICE'::TEXT,
        si.id,
        'Invoice ' || si.invoice_number || ' currency (' || si.currency_code || 
        ') differs from customer default (' || c.currency_code || ')',
        'MEDIUM'::TEXT,
        'Verify currency is intentional or update to match customer default'
    FROM sales_invoices si
    JOIN customers c ON si.customer_id = c.id
    WHERE si.tenant_id = current_tenant_id()
      AND si.currency_code != c.currency_code
    
    UNION ALL
    
    -- Check for invalid exchange rates
    SELECT 
        'CURRENCY_INTEGRITY'::TEXT,
        'SALES_INVOICE'::TEXT,
        si.id,
        'Invoice ' || si.invoice_number || ' has invalid exchange rate: ' || si.exchange_rate,
        'HIGH'::TEXT,
        'Update exchange rate to positive value'
    FROM sales_invoices si
    WHERE si.tenant_id = current_tenant_id()
      AND (si.exchange_rate IS NULL OR si.exchange_rate <= 0)
      AND si.currency_code != (
          SELECT currency_code FROM tenants WHERE id = current_tenant_id()
      )
    
    UNION ALL
    
    -- Check for missing exchange rates for foreign currency transactions
    SELECT 
        'CURRENCY_INTEGRITY'::TEXT,
        'PAYMENT'::TEXT,
        p.id,
        'Payment ' || p.payment_number || ' in foreign currency missing exchange rate',
        'HIGH'::TEXT,
        'Add exchange rate for foreign currency payments'
    FROM payments p
    WHERE p.tenant_id = current_tenant_id()
      AND p.currency_code != (
          SELECT currency_code FROM tenants WHERE id = current_tenant_id()
      )
      AND (p.exchange_rate IS NULL OR p.exchange_rate = 1);
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 6. DATE AND TEMPORAL LOGIC VALIDATIONS
-- =====================================================

-- Validate date logic and temporal consistency
CREATE OR REPLACE FUNCTION check_date_logic_integrity()
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
    -- Check for invoices with due dates before invoice dates
    SELECT 
        'DATE_LOGIC'::TEXT,
        'SALES_INVOICE'::TEXT,
        si.id,
        'Invoice ' || si.invoice_number || ' has due date (' || si.due_date || 
        ') before invoice date (' || si.invoice_date || ')',
        'HIGH'::TEXT,
        'Correct due date to be on or after invoice date'
    FROM sales_invoices si
    WHERE si.tenant_id = current_tenant_id()
      AND si.due_date < si.invoice_date
    
    UNION ALL
    
    -- Check for payments dated before related invoices
    SELECT 
        'DATE_LOGIC'::TEXT,
        'PAYMENT'::TEXT,
        p.id,
        'Payment ' || p.payment_number || ' dated (' || p.payment_date || 
        ') before related invoice date',
        'MEDIUM'::TEXT,
        'Verify payment date or invoice date is correct'
    FROM payments p
    JOIN payment_applications pa ON p.id = pa.payment_id
    JOIN sales_invoices si ON pa.sales_invoice_id = si.id
    WHERE p.tenant_id = current_tenant_id()
      AND p.payment_date < si.invoice_date
    
    UNION ALL
    
    -- Check for future-dated transactions
    SELECT 
        'DATE_LOGIC'::TEXT,
        'SALES_INVOICE'::TEXT,
        si.id,
        'Invoice ' || si.invoice_number || ' has future invoice date: ' || si.invoice_date,
        'MEDIUM'::TEXT,
        'Verify invoice date is correct'
    FROM sales_invoices si
    WHERE si.tenant_id = current_tenant_id()
      AND si.invoice_date > CURRENT_DATE + INTERVAL '1 day'
      AND si.status != 'draft';
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 7. PERFORMANCE AND REPORTING VALIDATIONS
-- =====================================================

-- Check for potential performance issues in business data
CREATE OR REPLACE FUNCTION check_performance_business_logic()
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
    -- Check for customers with excessive open invoices
    SELECT 
        'PERFORMANCE_RISK'::TEXT,
        'CUSTOMER'::TEXT,
        c.id,
        'Customer ' || c.name || ' has ' || COUNT(si.id) || ' open invoices (performance risk)',
        'MEDIUM'::TEXT,
        'Consider consolidating or collecting older invoices'
    FROM customers c
    JOIN sales_invoices si ON c.id = si.customer_id
    WHERE c.tenant_id = current_tenant_id()
      AND si.status IN ('issued', 'partial', 'overdue')
    GROUP BY c.id, c.name
    HAVING COUNT(si.id) > 100
    
    UNION ALL
    
    -- Check for invoices with excessive line items
    SELECT 
        'PERFORMANCE_RISK'::TEXT,
        'SALES_INVOICE'::TEXT,
        si.id,
        'Invoice ' || si.invoice_number || ' has ' || COUNT(sil.id) || ' line items (performance risk)',
        'LOW'::TEXT,
        'Consider breaking into multiple invoices for better performance'
    FROM sales_invoices si
    JOIN sales_invoice_lines sil ON si.id = sil.invoice_id
    WHERE si.tenant_id = current_tenant_id()
    GROUP BY si.id, si.invoice_number
    HAVING COUNT(sil.id) > 50;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 8. COMPLIANCE AND AUDIT VALIDATIONS
-- =====================================================

-- Check for compliance and audit requirements
CREATE OR REPLACE FUNCTION check_compliance_requirements()
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
    -- Check for missing audit trails on financial transactions
    SELECT 
        'COMPLIANCE'::TEXT,
        'SALES_INVOICE'::TEXT,
        si.id,
        'Invoice ' || si.invoice_number || ' is missing audit information (modified_by is NULL)',
        'MEDIUM'::TEXT,
        'Implement user tracking in application layer'
    FROM sales_invoices si
    WHERE si.tenant_id = current_tenant_id()
      AND si.status != 'draft'
      AND si.modified_by IS NULL
    
    UNION ALL
    
    -- Check for transactions without proper documentation
    SELECT 
        'COMPLIANCE'::TEXT,
        'PAYMENT'::TEXT,
        p.id,
        'Payment ' || p.payment_number || ' lacks reference number for audit trail',
        'MEDIUM'::TEXT,
        'Add reference numbers to all payments for audit compliance'
    FROM payments p
    WHERE p.tenant_id = current_tenant_id()
      AND p.reference_number IS NULL
      AND p.amount > 1000 -- Significant amounts
    
    UNION ALL
    
    -- Check for void transactions without proper documentation
    SELECT 
        'COMPLIANCE'::TEXT,
        'SALES_INVOICE'::TEXT,
        si.id,
        'Void invoice ' || si.invoice_number || ' lacks documentation in notes field',
        'HIGH'::TEXT,
        'Add reason for voiding in notes field for audit compliance'
    FROM sales_invoices si
    WHERE si.tenant_id = current_tenant_id()
      AND si.status = 'void'
      AND (si.notes IS NULL OR LENGTH(TRIM(si.notes)) < 10);
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 9. MASTER BUSINESS LOGIC VALIDATION
-- =====================================================

-- Comprehensive business logic validation function
CREATE OR REPLACE FUNCTION check_all_business_logic()
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
    SELECT * FROM check_invoice_line_totals()
    UNION ALL
    SELECT * FROM check_payment_application_integrity()
    UNION ALL
    SELECT * FROM check_business_workflow_integrity()
    UNION ALL
    SELECT * FROM check_referential_integrity()
    UNION ALL
    SELECT * FROM check_currency_integrity()
    UNION ALL
    SELECT * FROM check_date_logic_integrity()
    UNION ALL
    SELECT * FROM check_performance_business_logic()
    UNION ALL
    SELECT * FROM check_compliance_requirements()
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
-- 10. BUSINESS LOGIC REPORTING
-- =====================================================

-- Generate business logic validation summary
CREATE OR REPLACE FUNCTION generate_business_logic_report()
RETURNS TABLE(
    validation_category TEXT,
    critical_issues INT,
    high_issues INT,
    medium_issues INT,
    low_issues INT,
    total_issues INT,
    status TEXT
) AS $$
BEGIN
    RETURN QUERY
    WITH issue_summary AS (
        SELECT 
            validation_type,
            COUNT(*) FILTER (WHERE severity = 'CRITICAL') as critical_count,
            COUNT(*) FILTER (WHERE severity = 'HIGH') as high_count,
            COUNT(*) FILTER (WHERE severity = 'MEDIUM') as medium_count,
            COUNT(*) FILTER (WHERE severity = 'LOW') as low_count,
            COUNT(*) as total_count
        FROM check_all_business_logic()
        GROUP BY validation_type
        
        UNION ALL
        
        SELECT 
            'OVERALL_SUMMARY' as validation_type,
            COUNT(*) FILTER (WHERE severity = 'CRITICAL') as critical_count,
            COUNT(*) FILTER (WHERE severity = 'HIGH') as high_count,
            COUNT(*) FILTER (WHERE severity = 'MEDIUM') as medium_count,
            COUNT(*) FILTER (WHERE severity = 'LOW') as low_count,
            COUNT(*) as total_count
        FROM check_all_business_logic()
    )
    SELECT 
        validation_type::TEXT,
        critical_count::INT,
        high_count::INT,
        medium_count::INT,
        low_count::INT,
        total_count::INT,
        CASE 
            WHEN critical_count > 0 THEN 'CRITICAL - IMMEDIATE ACTION REQUIRED'
            WHEN high_count > 0 THEN 'HIGH - ACTION REQUIRED'
            WHEN medium_count > 5 THEN 'MEDIUM - REVIEW RECOMMENDED'
            WHEN total_count = 0 THEN 'CLEAN - NO ISSUES FOUND'
            ELSE 'LOW - MONITOR'
        END::TEXT
    FROM issue_summary
    ORDER BY critical_count DESC, high_count DESC, total_count DESC;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- USAGE INSTRUCTIONS AND EXAMPLES
-- =====================================================

/*
BUSINESS LOGIC VALIDATION USAGE:

1. Run comprehensive business logic check:
   SELECT * FROM check_all_business_logic();

2. Get summary report:
   SELECT * FROM generate_business_logic_report();

3. Run specific validations:
   SELECT * FROM check_invoice_line_totals();
   SELECT * FROM check_payment_application_integrity();
   SELECT * FROM check_business_workflow_integrity();

4. Filter by severity:
   SELECT * FROM check_all_business_logic() WHERE severity = 'CRITICAL';

5. Filter by entity type:
   SELECT * FROM check_all_business_logic() WHERE entity_type = 'SALES_INVOICE';

6. Generate daily/weekly business health reports:
   SELECT 
       'Business Logic Health Check - ' || CURRENT_DATE as report_title,
       * 
   FROM generate_business_logic_report();

INTEGRATION WITH MAIN VALIDATION:
Add this to your main migration validation script:

-- Run business logic validations
SELECT 'BUSINESS LOGIC VALIDATION RESULTS:' as section_header;
SELECT * FROM generate_business_logic_report();

-- Show critical issues only
SELECT 'CRITICAL BUSINESS ISSUES:' as section_header;
SELECT * FROM check_all_business_logic() WHERE severity = 'CRITICAL';
*/

-- Grant permissions for business logic validation functions
GRANT EXECUTE ON FUNCTION check_invoice_line_totals() TO application_role;
GRANT EXECUTE ON FUNCTION check_payment_application_integrity() TO application_role;
GRANT EXECUTE ON FUNCTION check_business_workflow_integrity() TO application_role;
GRANT EXECUTE ON FUNCTION check_referential_integrity() TO application_role;
GRANT EXECUTE ON FUNCTION check_currency_integrity() TO application_role;
GRANT EXECUTE ON FUNCTION check_date_logic_integrity() TO application_role;
GRANT EXECUTE ON FUNCTION check_performance_business_logic() TO application_role;
GRANT EXECUTE ON FUNCTION check_compliance_requirements() TO application_role;
GRANT EXECUTE ON FUNCTION check_all_business_logic() TO application_role;
GRANT EXECUTE ON FUNCTION generate_business_logic_report() TO application_role;
