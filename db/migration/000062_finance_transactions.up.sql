-- =====================================================================
-- FINANCE MODULE - TRANSACTIONS TABLE
-- Core transaction header table for all financial transactions
-- =====================================================================

-- Finance transaction headers
CREATE TABLE finance_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
    
    -- Transaction identification
    transaction_number VARCHAR(50) NOT NULL,
    transaction_type VARCHAR(30) NOT NULL CHECK (
        transaction_type IN ('MANUAL', 'SYSTEM', 'IMPORTED', 'RECURRING', 'ADJUSTMENT', 'CLOSING')
    ),
    transaction_status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (
        transaction_status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'POSTED', 'CANCELLED', 'REVERSED')
    ),
    
    -- Transaction dates
    transaction_date DATE NOT NULL,
    posting_date DATE,
    due_date DATE,
    
    -- Transaction details
    description TEXT NOT NULL,
    reference_number VARCHAR(100),
    external_reference VARCHAR(100),
    
    -- Financial information
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(18,8) DEFAULT 1.0,
    total_debit_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_credit_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    
    -- Source and traceability
    source_module VARCHAR(50),
    source_document_type VARCHAR(50),
    source_document_id UUID,
    batch_id UUID,
    
    -- Approval workflow
    approval_required BOOLEAN DEFAULT false,
    approval_status VARCHAR(20) DEFAULT 'NOT_REQUIRED' CHECK (
        approval_status IN ('NOT_REQUIRED', 'PENDING', 'APPROVED', 'REJECTED')
    ),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    approval_notes TEXT,
    
    -- Recurring transaction
    is_recurring BOOLEAN DEFAULT false,
    recurring_frequency VARCHAR(20) CHECK (
        recurring_frequency IS NULL OR 
        recurring_frequency IN ('DAILY', 'WEEKLY', 'MONTHLY', 'QUARTERLY', 'YEARLY')
    ),
    next_recurring_date DATE,
    
    -- Reversal tracking
    is_reversed BOOLEAN DEFAULT false,
    reversed_by_transaction_id UUID REFERENCES finance_transactions(id),
    reversal_reason TEXT,
    
    -- Audit and validation
    version INTEGER NOT NULL DEFAULT 1,
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    -- Metadata
    transaction_attributes JSONB DEFAULT '{}'::jsonb,
    
    -- Standard timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    posted_by UUID REFERENCES users(id),
    posted_at TIMESTAMPTZ,
    
    -- Unique constraints
    UNIQUE (tenant_id, transaction_number)
);

-- Table comments
COMMENT ON TABLE finance_transactions IS 
'Header table for all financial transactions. Contains transaction metadata, approval workflow, and summary amounts.';

COMMENT ON COLUMN finance_transactions.transaction_number IS 
'Unique transaction number within tenant - Auto-generated or user-provided';

COMMENT ON COLUMN finance_transactions.transaction_type IS 
'Type of transaction - determines behavior and validation rules';

COMMENT ON COLUMN finance_transactions.transaction_status IS 
'Current status in transaction lifecycle - controls what operations are allowed';

COMMENT ON COLUMN finance_transactions.exchange_rate IS 
'Exchange rate from transaction currency to functional currency';

COMMENT ON COLUMN finance_transactions.total_debit_amount IS 
'Sum of all debit entries - Must equal total_credit_amount for balanced transactions';

COMMENT ON COLUMN finance_transactions.total_credit_amount IS 
'Sum of all credit entries - Must equal total_debit_amount for balanced transactions';

-- Enable Row Level Security
ALTER TABLE finance_transactions ENABLE ROW LEVEL SECURITY;

-- RLS policy for tenant isolation
CREATE POLICY tenant_isolation_policy ON finance_transactions
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
CREATE POLICY admin_full_access_policy ON finance_transactions
    FOR ALL TO admin_role
    USING (true);

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transactions TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transactions TO admin_role;