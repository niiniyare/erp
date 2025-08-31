-- Comprehensive account view with group information
CREATE VIEW v_finance_accounts_with_groups AS
SELECT 
    a.id,
    a.tenant_id,
    a.account_code,
    a.account_name,
    a.account_description,
    a.root_type,
    a.account_type,
    a.account_category,
    a.normal_balance,
    a.current_balance,
    a.is_active,
    
    -- Group information
    g.group_code,
    g.group_name as group_name,
    g.group_category,
    
    -- Header information  
    h.group_code as header_code,
    h.group_name as header_name,
    h.financial_statement_section,
    
    -- Hierarchy information
    a.parent_account_id,
    a.account_level,

    -- Compute hierarchy flags dynamically (will be faster once trigger-maintained columns exist)
    EXISTS (
        SELECT 1 FROM finance_accounts children 
        WHERE children.parent_account_id = a.id 
        AND children.deleted_at IS NULL
    ) as has_children,
    NOT EXISTS (
        SELECT 1 FROM finance_accounts children 
        WHERE children.parent_account_id = a.id 
        AND children.deleted_at IS NULL
    ) as is_leaf_account,
    
    -- Display information
    COALESCE(a.display_order, g.statement_order, 999) as effective_display_order,
    a.cash_flow_type,
    a.show_in_reports
    
FROM finance_accounts a
LEFT JOIN finance_account_groups g ON a.account_group_id = g.id
LEFT JOIN finance_account_groups h ON a.account_header_id = h.id
WHERE a.deleted_at IS NULL;

COMMENT ON VIEW v_finance_accounts_with_groups IS 
'Complete account view with group and header information for reporting';

-- Financial statement structure view
CREATE VIEW v_financial_statement_structure AS
SELECT 
    g.id,
    g.tenant_id,
    g.group_code,
    g.group_name,
    g.root_type,
    g.financial_statement_section,
    g.statement_order,
    g.group_level,
    g.parent_group_id,
    g.show_in_summary,
    g.group_path,
    COUNT(a.id) as account_count,
    SUM(CASE WHEN a.is_active = true THEN 1 ELSE 0 END) as active_account_count,
    COALESCE(SUM(a.current_balance), 0) as group_balance
FROM finance_account_groups g
LEFT JOIN finance_accounts a ON (a.account_group_id = g.id OR a.account_header_id = g.id)
    AND a.deleted_at IS NULL
WHERE g.deleted_at IS NULL
GROUP BY g.id, g.tenant_id, g.group_code, g.group_name, g.root_type,
         g.financial_statement_section, g.statement_order, g.group_level,
         g.parent_group_id, g.show_in_summary, g.group_path
ORDER BY g.financial_statement_section, g.statement_order, g.group_code;

COMMENT ON VIEW v_financial_statement_structure IS 
'Financial statement structure with account groups, hierarchies, and balances';

-- =====================================================================
-- STANDARD CHART OF ACCOUNTS GROUPS SETUP FUNCTION
-- =====================================================================

-- Function to create standard account groups for new tenants
CREATE OR REPLACE FUNCTION create_standard_account_groups(p_tenant_id UUID)
RETURNS VOID AS $$
DECLARE
    v_assets_id UUID;
    v_liabilities_id UUID;
    v_equity_id UUID;
    v_revenue_id UUID;
    v_expenses_id UUID;
BEGIN
    -- Create root level groups
    INSERT INTO finance_account_groups (
        tenant_id, group_code, group_name, root_type, 
        financial_statement_section, statement_order, is_system_group
    ) VALUES 
    (p_tenant_id, 'ASSETS', 'Assets', 'ASSET', 'Balance Sheet', 100, true),
    (p_tenant_id, 'LIABILITIES', 'Liabilities', 'LIABILITY', 'Balance Sheet', 200, true),
    (p_tenant_id, 'EQUITY', 'Equity', 'EQUITY', 'Balance Sheet', 300, true),
    (p_tenant_id, 'REVENUE', 'Revenue', 'REVENUE', 'Income Statement', 400, true),
    (p_tenant_id, 'EXPENSES', 'Expenses', 'EXPENSE', 'Income Statement', 500, true)
    RETURNING id INTO v_assets_id, v_liabilities_id, v_equity_id, v_revenue_id, v_expenses_id;
    
    -- Create second level groups for Assets
    SELECT id INTO v_assets_id FROM finance_account_groups 
    WHERE tenant_id = p_tenant_id AND group_code = 'ASSETS';
    
    INSERT INTO finance_account_groups (
        tenant_id, group_code, group_name, root_type, parent_group_id,
        group_category, financial_statement_section, statement_order, group_level, is_system_group
    ) VALUES 
    (p_tenant_id, 'CURRENT_ASSETS', 'Current Assets', 'ASSET', v_assets_id, 'CURRENT_ASSETS', 'Balance Sheet', 110, 2, true),
    (p_tenant_id, 'FIXED_ASSETS', 'Fixed Assets', 'ASSET', v_assets_id, 'FIXED_ASSETS', 'Balance Sheet', 120, 2, true),
    (p_tenant_id, 'OTHER_ASSETS', 'Other Assets', 'ASSET', v_assets_id, 'OTHER_ASSETS', 'Balance Sheet', 130, 2, true);
    
    -- Create second level groups for Liabilities
    SELECT id INTO v_liabilities_id FROM finance_account_groups 
    WHERE tenant_id = p_tenant_id AND group_code = 'LIABILITIES';
    
    INSERT INTO finance_account_groups (
        tenant_id, group_code, group_name, root_type, parent_group_id,
        group_category, financial_statement_section, statement_order, group_level, is_system_group
    ) VALUES 
    (p_tenant_id, 'CURRENT_LIABILITIES', 'Current Liabilities', 'LIABILITY', v_liabilities_id, 'CURRENT_LIABILITIES', 'Balance Sheet', 210, 2, true),
    (p_tenant_id, 'LONG_TERM_LIABILITIES', 'Long-term Liabilities', 'LIABILITY', v_liabilities_id, 'LONG_TERM_LIABILITIES', 'Balance Sheet', 220, 2, true);
    
    -- Create expense groups
    SELECT id INTO v_expenses_id FROM finance_account_groups 
    WHERE tenant_id = p_tenant_id AND group_code = 'EXPENSES';
    
    INSERT INTO finance_account_groups (
        tenant_id, group_code, group_name, root_type, parent_group_id,
        group_category, financial_statement_section, statement_order, group_level, is_system_group
    ) VALUES 
    (p_tenant_id, 'OPERATING_EXPENSES', 'Operating Expenses', 'EXPENSE', v_expenses_id, 'OPERATING_EXPENSES', 'Income Statement', 510, 2, true),
    (p_tenant_id, 'ADMINISTRATIVE_EXPENSES', 'Administrative Expenses', 'EXPENSE', v_expenses_id, 'ADMINISTRATIVE_EXPENSES', 'Income Statement', 520, 2, true),
    (p_tenant_id, 'FINANCIAL_EXPENSES', 'Financial Expenses', 'EXPENSE', v_expenses_id, 'FINANCIAL_EXPENSES', 'Income Statement', 530, 2, true);
    
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_standard_account_groups IS 
'Creates standard account group structure for new tenants';

-- =====================================================================
-- GROUP HIERARCHY MAINTENANCE
-- =====================================================================

-- Function to maintain group hierarchy path (similar to accounts)
CREATE OR REPLACE FUNCTION maintain_group_hierarchy_path()
RETURNS TRIGGER AS $$
DECLARE
    parent_path VARCHAR(500);
BEGIN
    -- Build the group path based on parent hierarchy
    IF NEW.parent_group_id IS NULL THEN
        NEW.group_path := '/' || NEW.group_code || '/';
        NEW.group_level := 1;
    ELSE
        -- Get parent path and level
        SELECT group_path, group_level 
        INTO parent_path, NEW.group_level
        FROM finance_account_groups 
        WHERE id = NEW.parent_group_id;
        
        NEW.group_path := parent_path || NEW.group_code || '/';
        NEW.group_level := NEW.group_level + 1;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_maintain_group_hierarchy
    BEFORE INSERT OR UPDATE OF parent_group_id, group_code ON finance_account_groups
    FOR EACH ROW
    EXECUTE FUNCTION maintain_group_hierarchy_path();

-- =====================================================================
-- ENHANCED CHART OF ACCOUNTS VIEW WITH GROUPS
-- =====================================================================

CREATE VIEW v_chart_of_accounts_complete AS
SELECT 
    -- Account information
    a.id as account_id,
    a.tenant_id,
    a.account_code,
    a.account_name,
    a.account_description,
    a.root_type,
    a.account_type,
    a.account_subtype,
    a.account_category,
    a.normal_balance,
    a.current_balance,
    a.is_active,
    -- a.is_leaf_account,
    NOT EXISTS (
        SELECT 1 FROM finance_accounts children 
        WHERE children.parent_account_id = a.id 
        AND children.deleted_at IS NULL
    ) as is_leaf_account,

    -- Account hierarchy
    a.parent_account_id,
    a.account_level,
    a.account_path,
    
    -- Primary group information
    g.id as group_id,
    g.group_code,
    g.group_name,
    g.group_category,
    g.group_path,
    
    -- Header information
    h.id as header_id,
    h.group_code as header_code,
    h.group_name as header_name,
    h.financial_statement_section,
    h.statement_order as header_order,
    
    -- Combined display information
    COALESCE(a.display_order, g.statement_order, h.statement_order, 999) as display_order,
    COALESCE(h.financial_statement_section, g.financial_statement_section, 
             CASE a.root_type 
                WHEN 'ASSET' THEN 'Balance Sheet'
                WHEN 'LIABILITY' THEN 'Balance Sheet' 
                WHEN 'EQUITY' THEN 'Balance Sheet'
                ELSE 'Income Statement'
             END) as statement_section,
    
    -- Cash flow information
    COALESCE(a.cash_flow_type, g.cash_flow_category, h.cash_flow_category) as cash_flow_classification,
    
    -- Reporting flags
    a.show_in_reports AND COALESCE(g.show_in_summary, true) as include_in_reports
    
FROM finance_accounts a
LEFT JOIN finance_account_groups g ON a.account_group_id = g.id
LEFT JOIN finance_account_groups h ON a.account_header_id = h.id
WHERE a.deleted_at IS NULL
ORDER BY 
    COALESCE(h.statement_order, g.statement_order, 999),
    COALESCE(g.statement_order, 999),
    a.display_order,
    a.account_code;

COMMENT ON VIEW v_chart_of_accounts_complete IS 
'Complete chart of accounts with full group and header hierarchy for reporting';

-- =====================================================================
-- FINANCIAL STATEMENT BUILDER VIEW
-- =====================================================================

CREATE VIEW v_financial_statement_builder AS
WITH grouped_balances AS (
    SELECT 
        coa.tenant_id,
        coa.statement_section,
        coa.header_id,
        coa.header_code,
        coa.header_name,
        coa.header_order,
        coa.group_id,
        coa.group_code,
        coa.group_name,
        coa.group_category,
        COUNT(coa.account_id) as account_count,
        SUM(coa.current_balance) as group_balance,
        SUM(CASE WHEN coa.is_active THEN coa.current_balance ELSE 0 END) as active_balance
    FROM v_chart_of_accounts_complete coa
    WHERE coa.include_in_reports = true
    GROUP BY 
        coa.tenant_id, coa.statement_section, coa.header_id, coa.header_code, 
        coa.header_name, coa.header_order, coa.group_id, coa.group_code, 
        coa.group_name, coa.group_category
)
SELECT 
    tenant_id,
    statement_section,
    header_code,
    header_name,
    header_order,
    group_code,
    group_name,
    group_category,
    account_count,
    group_balance,
    active_balance,
    -- Running totals by statement section
    SUM(group_balance) OVER (
        PARTITION BY tenant_id, statement_section 
        ORDER BY header_order, group_code
    ) as running_total,
    -- Percentage of section total
    ROUND(
        (group_balance / NULLIF(SUM(group_balance) OVER (PARTITION BY tenant_id, statement_section), 0)) * 100, 
        2
    ) as percentage_of_section
FROM grouped_balances
ORDER BY statement_section, header_order, group_code;

COMMENT ON VIEW v_financial_statement_builder IS 
'Pre-aggregated data for building financial statements with grouping and totals';


-- =====================================================================
-- MAINTENANCE FUNCTIONS
-- =====================================================================

-- Function to update group totals (can be run periodically)
CREATE OR REPLACE FUNCTION update_group_balances(p_tenant_id UUID DEFAULT NULL)
RETURNS VOID AS $$
BEGIN
    -- This is informational since groups don't store balances directly
    -- But useful for validation and reporting
    WITH group_totals AS (
        SELECT 
            COALESCE(a.account_group_id, a.account_header_id) as group_id,
            SUM(a.current_balance) as total_balance,
            COUNT(a.id) as account_count
        FROM finance_accounts a
        WHERE a.deleted_at IS NULL 
        AND a.tenant_id = COALESCE(p_tenant_id, a.tenant_id)
        AND (a.account_group_id IS NOT NULL OR a.account_header_id IS NOT NULL)
        GROUP BY COALESCE(a.account_group_id, a.account_header_id)
    )
    UPDATE finance_account_groups g
    SET 
        group_attributes = COALESCE(g.group_attributes, '{}'::jsonb) || 
                          jsonb_build_object(
                              'calculated_balance', gt.total_balance,
                              'account_count', gt.account_count,
                              'last_calculated', NOW()
                          ),
        updated_at = NOW()
    FROM group_totals gt
    WHERE g.id = gt.group_id;
    
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_group_balances IS 
'Updates group metadata with calculated balances and account counts';

GRANT SELECT ON v_finance_accounts_with_groups TO application_role, admin_role;
GRANT SELECT ON v_chart_of_accounts_complete TO application_role, admin_role;
GRANT SELECT ON v_financial_statement_builder TO application_role, admin_role;

