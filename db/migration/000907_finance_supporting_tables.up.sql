-- Account balance history for audit trail
-- DECIMAL(19,4): supports 3-decimal currencies (KWD, IQD, OMR, JOD, BHD)
CREATE TABLE finance_account_balances (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  account_id UUID NOT NULL REFERENCES finance_accounts(id) ON DELETE CASCADE,
  entity_id UUID REFERENCES entities(uuid) ON DELETE CASCADE,
  -- Balance information
  balance_date DATE NOT NULL,
  opening_balance DECIMAL(19, 4) NOT NULL DEFAULT 0.0000,
  closing_balance DECIMAL(19, 4) NOT NULL DEFAULT 0.0000,
  period_debits DECIMAL(19, 4) NOT NULL DEFAULT 0.0000,
  period_credits DECIMAL(19, 4) NOT NULL DEFAULT 0.0000,
  -- Period information
  fiscal_year INTEGER NOT NULL,
  fiscal_period INTEGER NOT NULL,
  -- Audit trail
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by UUID REFERENCES users(id),
  -- Soft-delete: period-end balance snapshots are audit evidence; hard DELETE is prohibited
  deleted_at TIMESTAMPTZ,
  deleted_by UUID REFERENCES users(id),
  UNIQUE (tenant_id, account_id, balance_date)
);

COMMENT ON TABLE finance_account_balances IS 'Historical account balance tracking for audit and reporting purposes';

-- Account validation rules
CREATE TABLE finance_account_validation_rules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  -- Rule identification
  rule_name VARCHAR(100) NOT NULL,
  rule_description TEXT,
  -- Rule targeting
  account_type VARCHAR(50),
  root_type VARCHAR(20),
  account_pattern VARCHAR(100),  -- Regex pattern for account codes
  -- Validation rules
  min_amount DECIMAL(19, 4),
  max_amount DECIMAL(19, 4),
  required_reference BOOLEAN DEFAULT false,
  allowed_transaction_types VARCHAR(30) [],
  required_cost_center BOOLEAN DEFAULT false,
  -- Rule behavior
  is_active BOOLEAN DEFAULT TRUE,
  rule_severity VARCHAR(20) DEFAULT 'ERROR' CHECK (
    rule_severity IN ('INFO', 'WARNING', 'ERROR', 'BLOCKING')
  ),
  -- Custom validation
  custom_validation_function VARCHAR(100),
  validation_parameters JSONB DEFAULT '{}'::jsonb,
  -- Standard timestamps
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by UUID REFERENCES users(id),
  updated_by UUID REFERENCES users(id),
  UNIQUE (tenant_id, rule_name)
);

COMMENT ON TABLE finance_account_validation_rules IS 'Configurable validation rules for accounts and transactions - enables business rule enforcement';

-- COMPUTED COLUMNS (Backward Compatible)
ALTER TABLE
  finance_accounts
ADD
  COLUMN IF NOT EXISTS has_children BOOLEAN DEFAULT false;

ALTER TABLE
  finance_accounts
ADD
  COLUMN IF NOT EXISTS is_leaf_account BOOLEAN DEFAULT TRUE;

