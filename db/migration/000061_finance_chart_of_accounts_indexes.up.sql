-- =====================================================================
-- FINANCE CHART OF ACCOUNTS - PERFORMANCE INDEXES
-- =====================================================================

-- Primary lookup indexes
CREATE INDEX idx_chart_of_accounts_tenant_code 
    ON finance_accounts(tenant_id, account_code) 
    WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_chart_of_accounts_tenant_code IS 
'Optimizes account lookup by code within tenant - Most common query pattern';

CREATE INDEX idx_chart_of_accounts_tenant_type 
    ON finance_accounts(tenant_id, account_type, root_type) 
    WHERE deleted_at IS NULL;

-- Hierarchy traversal indexes
CREATE INDEX idx_chart_of_accounts_parent 
    ON finance_accounts(parent_account_id) 
    WHERE deleted_at IS NULL;

-- Financial reporting indexes
CREATE INDEX idx_chart_of_accounts_statement_line 
    ON finance_accounts(tenant_id, financial_statement_line, report_order) 
    WHERE deleted_at IS NULL AND is_active = true;

CREATE INDEX idx_chart_of_accounts_balance_tracking 
    ON finance_accounts(tenant_id, last_transaction_date DESC, current_balance) 
    WHERE deleted_at IS NULL AND current_balance != 0;

-- Control account relationships
CREATE INDEX idx_chart_of_accounts_control 
    ON finance_accounts(control_account_id) 
    WHERE is_control_account = false AND deleted_at IS NULL;

-- Multi-currency indexes
CREATE INDEX idx_chart_of_accounts_currency 
    ON finance_accounts(tenant_id, currency_code, is_multi_currency) 
    WHERE deleted_at IS NULL;

-- Additional performance indexes
CREATE INDEX idx_chart_of_accounts_entity 
    ON finance_accounts(entity_id) 
    WHERE deleted_at IS NULL;

CREATE INDEX idx_chart_of_accounts_active 
    ON finance_accounts(is_active) 
    WHERE is_active = true AND deleted_at IS NULL;

-- JSON attribute search
CREATE INDEX idx_chart_of_accounts_attributes_gin 
    ON finance_accounts USING gin(account_attributes);