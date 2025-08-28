-- =====================================================================
-- FINANCE MODULE - TRANSACTION ENTRIES TABLE  
-- Individual journal entries for double-entry bookkeeping
-- =====================================================================

-- Finance transaction entries (journal entries)
CREATE TABLE finance_transaction_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Transaction relationship
    transaction_id UUID NOT NULL REFERENCES finance_transactions(id) ON DELETE CASCADE,
    entry_number INTEGER NOT NULL,
    
    -- Account relationship
    account_id UUID NOT NULL REFERENCES finance_accounts(id) ON DELETE RESTRICT,
    
    -- Entry amounts
    debit_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    credit_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    
    -- Entry details
    description TEXT NOT NULL,
    reference VARCHAR(100),
    
    -- Dimensional analysis
    cost_center VARCHAR(20),
    department VARCHAR(50),
    project_id UUID,
    
    -- Multi-currency support
    original_currency CHAR(3),
    original_amount DECIMAL(15,2),
    exchange_rate DECIMAL(18,8),
    
    -- Tax information
    tax_code VARCHAR(20),
    tax_rate DECIMAL(5,2),
    tax_amount DECIMAL(15,2),
    
    -- Reconciliation
    reconciled BOOLEAN DEFAULT false,
    reconciled_date DATE,
    reconciliation_reference VARCHAR(100),
    
    -- Standard timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Constraints
    CHECK (debit_amount >= 0 AND credit_amount >= 0),
    CHECK (NOT (debit_amount > 0 AND credit_amount > 0)),
    CHECK (debit_amount > 0 OR credit_amount > 0),
    
    -- Unique entry number per transaction
    UNIQUE (transaction_id, entry_number)
);

-- Table comments
COMMENT ON TABLE finance_transaction_entries IS 
'Individual journal entries that make up financial transactions. Implements double-entry bookkeeping with debit and credit amounts.';

COMMENT ON COLUMN finance_transaction_entries.entry_number IS 
'Sequential entry number within transaction - Used for ordering and reference';

COMMENT ON COLUMN finance_transaction_entries.debit_amount IS 
'Debit amount in functional currency - Must be 0 if credit_amount > 0';

COMMENT ON COLUMN finance_transaction_entries.credit_amount IS 
'Credit amount in functional currency - Must be 0 if debit_amount > 0';

COMMENT ON COLUMN finance_transaction_entries.original_amount IS 
'Original transaction amount in original currency before conversion';

-- Enable Row Level Security
ALTER TABLE finance_transaction_entries ENABLE ROW LEVEL SECURITY;

-- RLS policy for tenant isolation
CREATE POLICY tenant_isolation_policy ON finance_transaction_entries
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy
CREATE POLICY admin_full_access_policy ON finance_transaction_entries
    FOR ALL TO admin_role
    USING (true);

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transaction_entries TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transaction_entries TO admin_role;