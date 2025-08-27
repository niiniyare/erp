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
    memo TEXT,
    
    -- Financial information
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(18,8) DEFAULT 1.0 CHECK (exchange_rate > 0),
    total_debit_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00 CHECK (total_debit_amount >= 0),
    total_credit_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00 CHECK (total_credit_amount >= 0),
    
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
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
        validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
    ),
    validation_errors JSONB DEFAULT '[]'::jsonb,
    
    -- Metadata and attachments
    transaction_attributes JSONB DEFAULT '{}'::jsonb,
    attachment_ids TEXT[],
    tags VARCHAR(25)[],
    
    -- Standard timestamps and audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    posted_by UUID REFERENCES users(id),
    posted_at TIMESTAMPTZ,
    
    -- Business rule constraints
    CONSTRAINT balanced_transaction CHECK (
        CASE 
            WHEN transaction_status IN ('POSTED', 'APPROVED') 
            THEN total_debit_amount = total_credit_amount 
            ELSE true 
        END
    ),
    CONSTRAINT posting_date_logic CHECK (
        CASE 
            WHEN transaction_status = 'POSTED' 
            THEN posting_date IS NOT NULL AND posted_by IS NOT NULL AND posted_at IS NOT NULL
            ELSE true 
        END
    ),
    CONSTRAINT approval_logic CHECK (
        CASE 
            WHEN approval_required = true AND transaction_status IN ('APPROVED', 'POSTED')
            THEN approved_by IS NOT NULL AND approved_at IS NOT NULL
            ELSE true 
        END
    ),
    CONSTRAINT recurring_logic CHECK (
        CASE 
            WHEN is_recurring = true 
            THEN recurring_frequency IS NOT NULL 
            ELSE recurring_frequency IS NULL 
        END
    ),
    CONSTRAINT reversal_logic CHECK (
        CASE 
            WHEN is_reversed = true 
            THEN reversed_by_transaction_id IS NOT NULL 
            ELSE reversed_by_transaction_id IS NULL 
        END
    ),
    
    -- Unique constraints
    UNIQUE (tenant_id, transaction_number)
);

-- =====================================================================
-- TABLE AND COLUMN COMMENTS
-- =====================================================================

COMMENT ON TABLE finance_transactions IS 
'Header table for all financial transactions. Contains transaction metadata, approval workflow, and summary amounts.';

-- Transaction identification
COMMENT ON COLUMN finance_transactions.id IS 
'Primary key - UUID for the transaction';

COMMENT ON COLUMN finance_transactions.tenant_id IS 
'Foreign key to tenant - ensures data isolation in multi-tenant environment';

COMMENT ON COLUMN finance_transactions.entity_id IS 
'Foreign key to entities table - links transaction to specific business entity/company';

COMMENT ON COLUMN finance_transactions.transaction_number IS 
'Unique transaction number within tenant - Auto-generated or user-provided';

COMMENT ON COLUMN finance_transactions.transaction_type IS 
'Type of transaction - determines behavior and validation rules';

COMMENT ON COLUMN finance_transactions.transaction_status IS 
'Current status in transaction lifecycle - controls what operations are allowed';

-- Transaction dates
COMMENT ON COLUMN finance_transactions.transaction_date IS 
'Date when the transaction occurred - business date for accounting purposes';

COMMENT ON COLUMN finance_transactions.posting_date IS 
'Date when transaction was posted to the general ledger - required when status is POSTED';

COMMENT ON COLUMN finance_transactions.due_date IS 
'Due date for payment transactions - used for AP/AR and cash management';

-- Transaction details
COMMENT ON COLUMN finance_transactions.description IS 
'Main description of the transaction - required field for audit trail';

COMMENT ON COLUMN finance_transactions.reference_number IS 
'Internal reference number - invoice number, check number, etc.';

COMMENT ON COLUMN finance_transactions.external_reference IS 
'External reference from third-party systems - bank reference, vendor invoice number';

COMMENT ON COLUMN finance_transactions.memo IS 
'Additional notes or memo about the transaction - free text field for additional context';

-- Financial information
COMMENT ON COLUMN finance_transactions.currency_code IS 
'ISO 4217 currency code - defaults to USD but supports multi-currency';

COMMENT ON COLUMN finance_transactions.exchange_rate IS 
'Exchange rate from transaction currency to functional currency - defaults to 1.0 for same currency';

COMMENT ON COLUMN finance_transactions.total_debit_amount IS 
'Sum of all debit entries - Must equal total_credit_amount for balanced transactions';

COMMENT ON COLUMN finance_transactions.total_credit_amount IS 
'Sum of all credit entries - Must equal total_debit_amount for balanced transactions';

-- Source and traceability
COMMENT ON COLUMN finance_transactions.source_module IS 
'Source module that created this transaction - AP, AR, GL, PAYROLL, etc.';

COMMENT ON COLUMN finance_transactions.source_document_type IS 
'Type of source document - INVOICE, PAYMENT, JOURNAL_ENTRY, etc.';

COMMENT ON COLUMN finance_transactions.source_document_id IS 
'ID of the source document that generated this transaction';

COMMENT ON COLUMN finance_transactions.batch_id IS 
'Batch ID for grouping related transactions - useful for imports and bulk operations';

-- Approval workflow
COMMENT ON COLUMN finance_transactions.approval_required IS 
'Whether this transaction requires approval before posting';

COMMENT ON COLUMN finance_transactions.approval_status IS 
'Current approval status - tracks approval workflow progress';

COMMENT ON COLUMN finance_transactions.approved_by IS 
'User who approved the transaction - required if approval_required is true';

COMMENT ON COLUMN finance_transactions.approved_at IS 
'Timestamp when transaction was approved';

COMMENT ON COLUMN finance_transactions.approval_notes IS 
'Notes from the approver - can include reasons for approval or rejection';

-- Recurring transaction
COMMENT ON COLUMN finance_transactions.is_recurring IS 
'Whether this is a recurring transaction template';

COMMENT ON COLUMN finance_transactions.recurring_frequency IS 
'Frequency for recurring transactions - DAILY, WEEKLY, MONTHLY, QUARTERLY, YEARLY';

COMMENT ON COLUMN finance_transactions.next_recurring_date IS 
'Next date when this recurring transaction should be generated';

-- Reversal tracking
COMMENT ON COLUMN finance_transactions.is_reversed IS 
'Whether this transaction has been reversed';

COMMENT ON COLUMN finance_transactions.reversed_by_transaction_id IS 
'ID of the reversing transaction - creates audit trail for reversals';

COMMENT ON COLUMN finance_transactions.reversal_reason IS 
'Reason for reversing the transaction - required for compliance';

-- Audit and validation
COMMENT ON COLUMN finance_transactions.version IS 
'Version number for optimistic locking - prevents concurrent modifications';

COMMENT ON COLUMN finance_transactions.validation_status IS 
'Status of transaction validation - PENDING, VALID, WARNING, ERROR';

COMMENT ON COLUMN finance_transactions.validation_errors IS 
'JSON array of validation errors and warnings - helps with troubleshooting';

-- Metadata and attachments
COMMENT ON COLUMN finance_transactions.transaction_attributes IS 
'JSON object for additional transaction attributes - flexible extension point';

COMMENT ON COLUMN finance_transactions.attachment_ids IS 
'Array of attachment/document IDs - links to supporting documents';

COMMENT ON COLUMN finance_transactions.tags IS 
'Array of tags for categorization and filtering - max 25 chars each';

-- Standard audit fields
COMMENT ON COLUMN finance_transactions.created_at IS 
'Timestamp when record was created - automatic timestamp';

COMMENT ON COLUMN finance_transactions.updated_at IS 
'Timestamp when record was last updated - updated by triggers';

COMMENT ON COLUMN finance_transactions.deleted_at IS 
'Soft delete timestamp - NULL means record is active';

COMMENT ON COLUMN finance_transactions.created_by IS 
'User who created the transaction - required for audit trail';

COMMENT ON COLUMN finance_transactions.updated_by IS 
'User who last updated the transaction';

COMMENT ON COLUMN finance_transactions.posted_by IS 
'User who posted the transaction to the general ledger';

COMMENT ON COLUMN finance_transactions.posted_at IS 
'Timestamp when transaction was posted - required when status is POSTED';

-- =====================================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================================

-- Primary access patterns
CREATE INDEX idx_finance_transactions_tenant_date ON finance_transactions (tenant_id, transaction_date);
CREATE INDEX idx_finance_transactions_tenant_number ON finance_transactions (tenant_id, transaction_number);
CREATE INDEX idx_finance_transactions_tenant_status ON finance_transactions (tenant_id, transaction_status);
CREATE INDEX idx_finance_transactions_entity_date ON finance_transactions (entity_id, transaction_date) WHERE entity_id IS NOT NULL;

-- Workflow and approval indexes
CREATE INDEX idx_finance_transactions_approval ON finance_transactions (tenant_id, approval_status) WHERE approval_required = true;
CREATE INDEX idx_finance_transactions_validation ON finance_transactions (tenant_id, validation_status);

-- Source document tracking
CREATE INDEX idx_finance_transactions_source ON finance_transactions (source_document_type, source_document_id) WHERE source_document_id IS NOT NULL;
CREATE INDEX idx_finance_transactions_batch ON finance_transactions (batch_id) WHERE batch_id IS NOT NULL;

-- Recurring transactions
CREATE INDEX idx_finance_transactions_recurring ON finance_transactions (tenant_id, next_recurring_date) WHERE is_recurring = true;

-- Audit and reporting
CREATE INDEX idx_finance_transactions_created ON finance_transactions (tenant_id, created_at);
CREATE INDEX idx_finance_transactions_posted ON finance_transactions (tenant_id, posted_at) WHERE posted_at IS NOT NULL;

-- Tag search (GIN index for array operations)
CREATE INDEX idx_finance_transactions_tags ON finance_transactions USING GIN (tags) WHERE tags IS NOT NULL;

-- =====================================================================
-- ROW LEVEL SECURITY
-- =====================================================================

-- Enable Row Level Security
ALTER TABLE finance_transactions ENABLE ROW LEVEL SECURITY;

-- RLS policy for tenant isolation
CREATE POLICY tenant_isolation_policy ON finance_transactions
    FOR ALL TO application_role
    USING (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
        AND deleted_at IS NULL
    )
    WITH CHECK (
        current_tenant_id() IS NOT NULL 
        AND tenant_id = current_tenant_id()
    );

-- Admin bypass policy
CREATE POLICY admin_full_access_policy ON finance_transactions
    FOR ALL TO admin_role
    USING (true);

-- =====================================================================
-- PERMISSIONS
-- =====================================================================

GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transactions TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transactions TO admin_role;

-- =====================================================================
-- TRIGGERS
-- =====================================================================

-- Updated timestamp trigger
CREATE OR REPLACE FUNCTION update_finance_transactions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_finance_transactions_updated_at
    BEFORE UPDATE ON finance_transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_finance_transactions_updated_at();

-- Version increment trigger
CREATE OR REPLACE FUNCTION increment_finance_transaction_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_finance_transaction_version
    BEFORE UPDATE ON finance_transactions
    FOR EACH ROW
    EXECUTE FUNCTION increment_finance_transaction_version();

--
-- -- =====================================================================
-- -- FINANCE MODULE - TRANSACTIONS TABLE
-- -- Core transaction header table for all financial transactions
-- -- =====================================================================
--
-- -- Finance transaction headers
-- CREATE TABLE finance_transactions (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
--     entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
--
--     -- Transaction identification
--     transaction_number VARCHAR(50) NOT NULL,
--     transaction_type VARCHAR(30) NOT NULL CHECK (
--         transaction_type IN ('MANUAL', 'SYSTEM', 'IMPORTED', 'RECURRING', 'ADJUSTMENT', 'CLOSING')
--     ),
--     transaction_status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (
--         transaction_status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'POSTED', 'CANCELLED', 'REVERSED')
--     ),
--
--     -- Transaction dates
--     transaction_date DATE NOT NULL,
--     posting_date DATE,
--     due_date DATE,
--
--     -- Transaction details
--     description TEXT NOT NULL,
--     reference_number VARCHAR(100),
--     external_reference VARCHAR(100),
--
--     -- Financial information
--     currency_code CHAR(3) NOT NULL DEFAULT 'USD',
--     exchange_rate DECIMAL(18,8) DEFAULT 1.0,
--     total_debit_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
--     total_credit_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
--
--     -- Source and traceability
--     source_module VARCHAR(50),
--     source_document_type VARCHAR(50),
--     source_document_id UUID,
--     batch_id UUID,
--
--     -- Approval workflow
--     approval_required BOOLEAN DEFAULT false,
--     approval_status VARCHAR(20) DEFAULT 'NOT_REQUIRED' CHECK (
--         approval_status IN ('NOT_REQUIRED', 'PENDING', 'APPROVED', 'REJECTED')
--     ),
--     approved_by UUID REFERENCES users(id),
--     approved_at TIMESTAMPTZ,
--     approval_notes TEXT,
--
--     -- Recurring transaction
--     is_recurring BOOLEAN DEFAULT false,
--     recurring_frequency VARCHAR(20) CHECK (
--         recurring_frequency IS NULL OR 
--         recurring_frequency IN ('DAILY', 'WEEKLY', 'MONTHLY', 'QUARTERLY', 'YEARLY')
--     ),
--     next_recurring_date DATE,
--
--     -- Reversal tracking
--     is_reversed BOOLEAN DEFAULT false,
--     reversed_by_transaction_id UUID REFERENCES finance_transactions(id),
--     reversal_reason TEXT,
--
--     -- Audit and validation
--     version INTEGER NOT NULL DEFAULT 1,
--     validation_status VARCHAR(20) DEFAULT 'PENDING' CHECK (
--         validation_status IN ('PENDING', 'VALID', 'WARNING', 'ERROR')
--     ),
--     validation_errors JSONB DEFAULT '[]'::jsonb,
--
--     -- Metadata
--     transaction_attributes JSONB DEFAULT '{}'::jsonb,
--     memo TEXT,
--     attachment_ids []TEXT,
--     tags []VARCHAR(25),
--     -- Standard timestamps
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     deleted_at TIMESTAMPTZ,
--     created_by UUID NOT NULL REFERENCES users(id),
--     updated_by UUID REFERENCES users(id),
--     posted_by UUID REFERENCES users(id),
--     posted_at TIMESTAMPTZ,
--
--     -- Unique constraints
--     UNIQUE (tenant_id, transaction_number)
-- );
--
-- -- Table comments
-- COMMENT ON TABLE finance_transactions IS 
-- 'Header table for all financial transactions. Contains transaction metadata, approval workflow, and summary amounts.';
--
-- COMMENT ON COLUMN finance_transactions.transaction_number IS 
-- 'Unique transaction number within tenant - Auto-generated or user-provided';
--
-- COMMENT ON COLUMN finance_transactions.transaction_type IS 
-- 'Type of transaction - determines behavior and validation rules';
--
-- COMMENT ON COLUMN finance_transactions.transaction_status IS 
-- 'Current status in transaction lifecycle - controls what operations are allowed';
--
-- COMMENT ON COLUMN finance_transactions.exchange_rate IS 
-- 'Exchange rate from transaction currency to functional currency';
--
-- COMMENT ON COLUMN finance_transactions.total_debit_amount IS 
-- 'Sum of all debit entries - Must equal total_credit_amount for balanced transactions';
--
-- COMMENT ON COLUMN finance_transactions.total_credit_amount IS 
-- 'Sum of all credit entries - Must equal total_debit_amount for balanced transactions';
--
-- -- Enable Row Level Security
-- ALTER TABLE finance_transactions ENABLE ROW LEVEL SECURITY;
--
-- -- RLS policy for tenant isolation
-- CREATE POLICY tenant_isolation_policy ON finance_transactions
--     FOR ALL TO application_role
--     USING (
--         current_tenant_id() IS NOT NULL 
--         AND tenant_id = current_tenant_id()
--     )
--     WITH CHECK (
--         current_tenant_id() IS NOT NULL 
--         AND tenant_id = current_tenant_id()
--     );
--
-- -- Admin bypass policy
-- CREATE POLICY admin_full_access_policy ON finance_transactions
--     FOR ALL TO admin_role
--     USING (true);
--
-- -- Grant permissions
-- GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transactions TO application_role;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON finance_transactions TO admin_role;
