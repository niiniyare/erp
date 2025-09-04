# Financial Management

## 💰 Overview

The Financial Management module provides  accounting capabilities including general ledger, accounts payable/receivable, financial reporting, budgeting, and multi-currency support. Built on double-entry accounting principles with support for multiple accounting standards (GAAP, IFRS).

## 🏗️ Chart of Accounts

### Account Structure

```sql
-- Chart of Accounts
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Account identification
    account_code VARCHAR(20) UNIQUE NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    parent_account_id UUID REFERENCES accounts(id),
    
    -- Account classification
    account_type VARCHAR(50) NOT NULL, -- asset, liability, equity, revenue, expense
    account_subtype VARCHAR(50), -- current_asset, fixed_asset, current_liability, etc.
    account_category VARCHAR(100), -- cash, inventory, accounts_receivable, etc.
    
    -- Account properties
    is_active BOOLEAN DEFAULT true,
    is_system_account BOOLEAN DEFAULT false,
    allow_manual_entries BOOLEAN DEFAULT true,
    require_cost_center BOOLEAN DEFAULT false,
    require_project BOOLEAN DEFAULT false,
    
    -- Reporting and analysis
    reporting_group VARCHAR(100),
    financial_statement_section VARCHAR(100),
    cash_flow_category VARCHAR(50),
    
    -- Balance information
    current_balance DECIMAL(20,4) DEFAULT 0,
    current_balance_date DATE DEFAULT CURRENT_DATE,
    opening_balance DECIMAL(20,4) DEFAULT 0,
    
    -- Currency and location
    default_currency_code CHAR(3),
    tax_code VARCHAR(20),
    cost_center_code VARCHAR(20),
    
    -- Metadata
    description TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_account_type CHECK (account_type IN ('asset', 'liability', 'equity', 'revenue', 'expense')),
    CONSTRAINT valid_account_subtype CHECK (account_subtype IN (
        'current_asset', 'fixed_asset', 'other_asset',
        'current_liability', 'long_term_liability',
        'equity', 'retained_earnings',
        'operating_revenue', 'other_revenue',
        'cost_of_goods_sold', 'operating_expense', 'other_expense'
    ))
);

-- Standard Chart of Accounts Template
INSERT INTO accounts (tenant_id, account_code, account_name, account_type, account_subtype) VALUES
-- Assets
('template', '1000', 'Assets', 'asset', NULL),
('template', '1100', 'Current Assets', 'asset', 'current_asset'),
('template', '1110', 'Cash and Cash Equivalents', 'asset', 'current_asset'),
('template', '1111', 'Petty Cash', 'asset', 'current_asset'),
('template', '1112', 'Checking Account', 'asset', 'current_asset'),
('template', '1113', 'Savings Account', 'asset', 'current_asset'),
('template', '1120', 'Accounts Receivable', 'asset', 'current_asset'),
('template', '1130', 'Inventory', 'asset', 'current_asset'),
('template', '1140', 'Prepaid Expenses', 'asset', 'current_asset'),

-- Fixed Assets
('template', '1200', 'Fixed Assets', 'asset', 'fixed_asset'),
('template', '1210', 'Property, Plant & Equipment', 'asset', 'fixed_asset'),
('template', '1220', 'Accumulated Depreciation', 'asset', 'fixed_asset'),

-- Liabilities
('template', '2000', 'Liabilities', 'liability', NULL),
('template', '2100', 'Current Liabilities', 'liability', 'current_liability'),
('template', '2110', 'Accounts Payable', 'liability', 'current_liability'),
('template', '2120', 'Accrued Expenses', 'liability', 'current_liability'),
('template', '2130', 'Short-term Debt', 'liability', 'current_liability'),

-- Equity
('template', '3000', 'Equity', 'equity', 'equity'),
('template', '3100', 'Share Capital', 'equity', 'equity'),
('template', '3200', 'Retained Earnings', 'equity', 'retained_earnings'),

-- Revenue
('template', '4000', 'Revenue', 'revenue', 'operating_revenue'),
('template', '4100', 'Sales Revenue', 'revenue', 'operating_revenue'),
('template', '4200', 'Service Revenue', 'revenue', 'operating_revenue'),

-- Expenses
('template', '5000', 'Cost of Goods Sold', 'expense', 'cost_of_goods_sold'),
('template', '6000', 'Operating Expenses', 'expense', 'operating_expense'),
('template', '6100', 'Salaries and Wages', 'expense', 'operating_expense'),
('template', '6200', 'Rent Expense', 'expense', 'operating_expense'),
('template', '6300', 'Utilities Expense', 'expense', 'operating_expense');
```

### Account Hierarchies and Reporting

```typescript
interface AccountHierarchy {
  account_id: string;
  account_code: string;
  account_name: string;
  level: number;
  parent_path: string[];
  children: AccountHierarchy[];
  balance: {
    debit: number;
    credit: number;
    net: number;
  };
}

// Account balance calculation with hierarchy rollup
class AccountService {
  async getAccountHierarchy(tenantId: string, asOfDate?: Date): Promise<AccountHierarchy[]> {
    const accounts = await this.getAccountsWithBalances(tenantId, asOfDate);
    return this.buildHierarchy(accounts);
  }
  
  private buildHierarchy(accounts: Account[]): AccountHierarchy[] {
    const accountMap = new Map<string, AccountHierarchy>();
    const roots: AccountHierarchy[] = [];
    
    // Create hierarchy nodes
    accounts.forEach(account => {
      accountMap.set(account.id, {
        ...account,
        level: 0,
        parent_path: [],
        children: [],
        balance: account.balance
      });
    });
    
    // Build parent-child relationships
    accounts.forEach(account => {
      const node = accountMap.get(account.id)!;
      
      if (account.parent_account_id) {
        const parent = accountMap.get(account.parent_account_id);
        if (parent) {
          parent.children.push(node);
          node.level = parent.level + 1;
          node.parent_path = [...parent.parent_path, parent.account_id];
        }
      } else {
        roots.push(node);
      }
    });
    
    // Calculate rollup balances
    this.calculateRollupBalances(roots);
    
    return roots;
  }
  
  private calculateRollupBalances(nodes: AccountHierarchy[]): void {
    nodes.forEach(node => {
      if (node.children.length > 0) {
        this.calculateRollupBalances(node.children);
        
        // Sum child balances for parent accounts
        node.balance = node.children.reduce((total, child) => ({
          debit: total.debit + child.balance.debit,
          credit: total.credit + child.balance.credit,
          net: total.net + child.balance.net
        }), { debit: 0, credit: 0, net: 0 });
      }
    });
  }
}
```

## 📊 General Ledger

### Journal Entries

```sql
-- Journal entry header
CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Entry identification
    journal_entry_number VARCHAR(50) UNIQUE NOT NULL,
    reference_number VARCHAR(100),
    description TEXT NOT NULL,
    
    -- Entry details
    entry_date DATE NOT NULL,
    posting_date DATE,
    period_id UUID REFERENCES accounting_periods(id),
    
    -- Entry classification
    entry_type VARCHAR(50) DEFAULT 'manual', -- manual, automatic, closing, adjustment
    source_module VARCHAR(50), -- sales, purchase, inventory, payroll, etc.
    source_document_type VARCHAR(50),
    source_document_id UUID,
    
    -- Entry status
    status VARCHAR(20) DEFAULT 'draft', -- draft, posted, cancelled, reversed
    posted_at TIMESTAMPTZ,
    posted_by UUID REFERENCES users(id),
    
    -- Totals for validation
    total_debit DECIMAL(20,4) DEFAULT 0,
    total_credit DECIMAL(20,4) DEFAULT 0,
    
    -- Approval workflow
    requires_approval BOOLEAN DEFAULT false,
    approved_at TIMESTAMPTZ,
    approved_by UUID REFERENCES users(id),
    
    -- Reversal handling
    reversed_entry_id UUID REFERENCES journal_entries(id),
    reversal_reason TEXT,
    
    -- Currency
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1.000000,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_status CHECK (status IN ('draft', 'posted', 'cancelled', 'reversed')),
    CONSTRAINT valid_entry_type CHECK (entry_type IN ('manual', 'automatic', 'closing', 'adjustment')),
    CONSTRAINT balanced_entry CHECK (total_debit = total_credit)
);

-- Journal entry line items
CREATE TABLE journal_entry_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    
    -- Account information
    account_id UUID NOT NULL REFERENCES accounts(id),
    
    -- Amount information
    debit_amount DECIMAL(20,4) DEFAULT 0,
    credit_amount DECIMAL(20,4) DEFAULT 0,
    
    -- Multi-currency support
    foreign_currency_code CHAR(3),
    foreign_debit_amount DECIMAL(20,4),
    foreign_credit_amount DECIMAL(20,4),
    exchange_rate DECIMAL(10,6),
    
    -- Additional dimensions
    cost_center_id UUID REFERENCES cost_centers(id),
    project_id UUID REFERENCES projects(id),
    department_id UUID REFERENCES organizations(id),
    
    -- Line description and reference
    description TEXT,
    reference VARCHAR(255),
    
    -- Tax information
    tax_code VARCHAR(20),
    tax_amount DECIMAL(20,4) DEFAULT 0,
    
    -- Analytics dimensions
    dimension1_value VARCHAR(100),
    dimension2_value VARCHAR(100),
    dimension3_value VARCHAR(100),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_amounts CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR 
        (debit_amount = 0 AND credit_amount > 0)
    ),
    UNIQUE(journal_entry_id, line_number)
);
```

### Automated Journal Entry Generation

```typescript
interface AutoJournalEntryRule {
  id: string;
  tenant_id: string;
  name: string;
  source_module: string;
  trigger_event: string;
  
  conditions: JournalCondition[];
  mapping_rules: AccountMappingRule[];
  
  is_active: boolean;
  priority: number;
}

interface AccountMappingRule {
  account_type: 'debit' | 'credit';
  account_determination: 'fixed' | 'dynamic' | 'lookup';
  account_id?: string;
  account_field?: string; // For dynamic determination
  amount_field: string;
  description_template: string;
  
  conditions?: JournalCondition[];
  cost_center_field?: string;
  project_field?: string;
}

// Example: Sales Invoice Auto-Journal Entry
const salesInvoiceJournalRule: AutoJournalEntryRule = {
  id: 'sales-invoice-journal',
  tenant_id: 'tenant-id',
  name: 'Sales Invoice Journal Entry',
  source_module: 'sales',
  trigger_event: 'invoice_posted',
  
  conditions: [
    {
      field: 'invoice.status',
      operator: 'equals',
      value: 'posted'
    }
  ],
  
  mapping_rules: [
    {
      account_type: 'debit',
      account_determination: 'dynamic',
      account_field: 'customer.receivables_account_id',
      amount_field: 'total_amount',
      description_template: 'Invoice {invoice_number} - {customer_name}'
    },
    {
      account_type: 'credit',
      account_determination: 'dynamic',
      account_field: 'item.revenue_account_id',
      amount_field: 'line_total',
      description_template: 'Sale of {item_name} - Invoice {invoice_number}'
    },
    {
      account_type: 'credit',
      account_determination: 'dynamic',
      account_field: 'tax_code.tax_account_id',
      amount_field: 'tax_amount',
      description_template: 'Sales Tax - Invoice {invoice_number}',
      conditions: [
        {
          field: 'tax_amount',
          operator: 'greater_than',
          value: 0
        }
      ]
    }
  ],
  
  is_active: true,
  priority: 1
};
```

## 🧮 Accounts Receivable

### Customer Management

```sql
-- Customer master data
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Customer identification
    customer_code VARCHAR(50) UNIQUE NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_type VARCHAR(50) DEFAULT 'individual', -- individual, business, government
    
    -- Contact information
    primary_contact_name VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(20),
    mobile VARCHAR(20),
    website VARCHAR(255),
    
    -- Address information
    billing_address JSONB,
    shipping_address JSONB,
    
    -- Financial settings
    credit_limit DECIMAL(15,2) DEFAULT 0,
    payment_terms_days INTEGER DEFAULT 30,
    currency_code CHAR(3) DEFAULT 'USD',
    tax_id VARCHAR(50),
    tax_exempt BOOLEAN DEFAULT false,
    
    -- Account classification
    customer_group VARCHAR(100),
    sales_territory VARCHAR(100),
    price_list_id UUID REFERENCES price_lists(id),
    
    -- Accounting integration
    receivables_account_id UUID REFERENCES accounts(id),
    revenue_account_id UUID REFERENCES accounts(id),
    
    -- Status and preferences
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, blocked
    preferred_payment_method VARCHAR(50),
    statement_cycle VARCHAR(20) DEFAULT 'monthly',
    
    -- Credit management
    credit_status VARCHAR(20) DEFAULT 'approved', -- approved, blocked, on_hold
    last_credit_review_date DATE,
    credit_review_notes TEXT,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_customer_type CHECK (customer_type IN ('individual', 'business', 'government')),
    CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'blocked')),
    CONSTRAINT valid_credit_status CHECK (credit_status IN ('approved', 'blocked', 'on_hold'))
);
```

### Sales Invoicing

```sql
-- Sales invoices
CREATE TABLE sales_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Invoice identification
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id),
    
    -- Invoice dates
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    service_period_start DATE,
    service_period_end DATE,
    
    -- Financial details
    subtotal DECIMAL(15,2) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    -- Multi-currency support
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1.000000,
    base_currency_total DECIMAL(15,2),
    
    -- Payment information
    payment_terms_days INTEGER DEFAULT 30,
    payment_status VARCHAR(20) DEFAULT 'unpaid', -- unpaid, partial, paid, overdue
    amount_paid DECIMAL(15,2) DEFAULT 0,
    amount_due DECIMAL(15,2),
    
    -- Invoice status
    status VARCHAR(20) DEFAULT 'draft', -- draft, sent, posted, cancelled, void
    posted_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    
    -- References
    sales_order_id UUID REFERENCES sales_orders(id),
    project_id UUID REFERENCES projects(id),
    
    -- Billing address
    billing_address JSONB,
    
    -- Terms and notes
    terms_and_conditions TEXT,
    notes TEXT,
    internal_notes TEXT,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_payment_status CHECK (payment_status IN ('unpaid', 'partial', 'paid', 'overdue')),
    CONSTRAINT valid_status CHECK (status IN ('draft', 'sent', 'posted', 'cancelled', 'void'))
);

-- Sales invoice line items
CREATE TABLE sales_invoice_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_invoice_id UUID NOT NULL REFERENCES sales_invoices(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    
    -- Item information
    item_id UUID REFERENCES inventory_items(id),
    item_code VARCHAR(100),
    item_description TEXT NOT NULL,
    
    -- Quantity and pricing
    quantity DECIMAL(12,4) NOT NULL DEFAULT 1,
    unit_of_measure VARCHAR(20),
    unit_price DECIMAL(15,4) NOT NULL,
    discount_percentage DECIMAL(5,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    line_total DECIMAL(15,2) NOT NULL,
    
    -- Tax information
    tax_code VARCHAR(20),
    tax_rate DECIMAL(5,4) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    
    -- Accounting integration
    revenue_account_id UUID REFERENCES accounts(id),
    cost_center_id UUID REFERENCES cost_centers(id),
    project_id UUID REFERENCES projects(id),
    
    -- References
    sales_order_line_id UUID REFERENCES sales_order_lines(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(sales_invoice_id, line_number)
);
```

### Aging and Collections

```typescript
interface AgingReport {
  customer_id: string;
  customer_name: string;
  total_outstanding: number;
  
  aging_buckets: {
    current: number;        // 0-30 days
    days_31_60: number;     // 31-60 days
    days_61_90: number;     // 61-90 days
    days_over_90: number;   // Over 90 days
  };
  
  last_payment_date: Date;
  last_payment_amount: number;
  credit_limit: number;
  credit_available: number;
  
  invoices: AgingInvoice[];
}

interface CollectionAction {
  id: string;
  customer_id: string;
  action_type: 'email_reminder' | 'phone_call' | 'letter' | 'collection_agency' | 'legal_action';
  action_date: Date;
  due_amount: number;
  notes: string;
  follow_up_date: Date;
  assigned_to: string;
  status: 'pending' | 'completed' | 'cancelled';
}

// Automated collection workflow
const collectionWorkflow = {
  triggers: [
    {
      condition: 'days_overdue >= 7',
      action: 'send_gentle_reminder_email',
      template: 'first_reminder'
    },
    {
      condition: 'days_overdue >= 15',
      action: 'send_firm_reminder_email',
      template: 'second_reminder'
    },
    {
      condition: 'days_overdue >= 30',
      action: 'phone_call_required',
      assign_to: 'collections_team'
    },
    {
      condition: 'days_overdue >= 60',
      action: 'collection_letter',
      template: 'formal_demand'
    },
    {
      condition: 'days_overdue >= 90',
      action: 'escalate_to_manager',
      severity: 'high'
    }
  ]
};
```

## 💳 Accounts Payable

### Vendor Management

```sql
-- Vendor master data
CREATE TABLE vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Vendor identification
    vendor_code VARCHAR(50) UNIQUE NOT NULL,
    vendor_name VARCHAR(255) NOT NULL,
    vendor_type VARCHAR(50) DEFAULT 'supplier', -- supplier, service_provider, contractor
    
    -- Contact information
    primary_contact_name VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(20),
    fax VARCHAR(20),
    website VARCHAR(255),
    
    -- Address information
    billing_address JSONB,
    remit_to_address JSONB,
    
    -- Financial settings
    payment_terms_days INTEGER DEFAULT 30,
    currency_code CHAR(3) DEFAULT 'USD',
    tax_id VARCHAR(50),
    
    -- Banking information
    bank_name VARCHAR(255),
    bank_account_number VARCHAR(50),
    bank_routing_number VARCHAR(20),
    iban VARCHAR(34),
    swift_code VARCHAR(11),
    
    -- Vendor classification
    vendor_group VARCHAR(100),
    expense_category VARCHAR(100),
    
    -- Accounting integration
    payables_account_id UUID REFERENCES accounts(id),
    expense_account_id UUID REFERENCES accounts(id),
    
    -- Status and approval
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, blocked
    approval_status VARCHAR(20) DEFAULT 'approved', -- pending, approved, rejected
    approved_at TIMESTAMPTZ,
    approved_by UUID REFERENCES users(id),
    
    -- 1099 reporting (US specific)
    is_1099_vendor BOOLEAN DEFAULT false,
    vendor_1099_type VARCHAR(20), -- NEC, MISC, INT, etc.
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_vendor_type CHECK (vendor_type IN ('supplier', 'service_provider', 'contractor')),
    CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'blocked')),
    CONSTRAINT valid_approval_status CHECK (approval_status IN ('pending', 'approved', 'rejected'))
);
```

### Purchase Invoicing and Three-Way Matching

```sql
-- Purchase invoices
CREATE TABLE purchase_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Invoice identification
    vendor_invoice_number VARCHAR(100) NOT NULL,
    internal_invoice_number VARCHAR(50) UNIQUE,
    vendor_id UUID NOT NULL REFERENCES vendors(id),
    
    -- Invoice dates
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    received_date DATE DEFAULT CURRENT_DATE,
    
    -- Financial details
    subtotal DECIMAL(15,2) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    -- Multi-currency support
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1.000000,
    base_currency_total DECIMAL(15,2),
    
    -- Payment information
    payment_terms_days INTEGER,
    payment_status VARCHAR(20) DEFAULT 'unpaid', -- unpaid, partial, paid
    amount_paid DECIMAL(15,2) DEFAULT 0,
    amount_due DECIMAL(15,2),
    
    -- Invoice status and workflow
    status VARCHAR(20) DEFAULT 'received', -- received, matched, approved, posted, paid, rejected
    approval_status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected
    matching_status VARCHAR(20) DEFAULT 'unmatched', -- unmatched, matched, exception
    
    -- Three-way matching
    purchase_order_id UUID REFERENCES purchase_orders(id),
    goods_receipt_id UUID REFERENCES goods_receipts(id),
    matching_tolerance_exceeded BOOLEAN DEFAULT false,
    matching_exceptions JSONB,
    
    -- Approval workflow
    approved_at TIMESTAMPTZ,
    approved_by UUID REFERENCES users(id),
    approval_notes TEXT,
    
    -- References
    project_id UUID REFERENCES projects(id),
    
    -- Terms and notes
    description TEXT,
    notes TEXT,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_payment_status CHECK (payment_status IN ('unpaid', 'partial', 'paid')),
    CONSTRAINT valid_status CHECK (status IN ('received', 'matched', 'approved', 'posted', 'paid', 'rejected')),
    CONSTRAINT valid_approval_status CHECK (approval_status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT valid_matching_status CHECK (matching_status IN ('unmatched', 'matched', 'exception')),
    
    UNIQUE(vendor_id, vendor_invoice_number)
);

-- Three-way matching service
CREATE OR REPLACE FUNCTION perform_three_way_match(
    p_invoice_id UUID,
    p_tolerance_percentage DECIMAL DEFAULT 5.0
) RETURNS TABLE (
    match_result VARCHAR(20),
    exceptions JSONB
) AS $$
DECLARE
    v_invoice purchase_invoices%ROWTYPE;
    v_po purchase_orders%ROWTYPE;
    v_gr goods_receipts%ROWTYPE;
    v_exceptions JSONB DEFAULT '[]'::JSONB;
    v_match_result VARCHAR(20) DEFAULT 'matched';
BEGIN
    -- Get invoice details
    SELECT * INTO v_invoice FROM purchase_invoices WHERE id = p_invoice_id;
    
    -- Get related PO and GR
    SELECT * INTO v_po FROM purchase_orders WHERE id = v_invoice.purchase_order_id;
    SELECT * INTO v_gr FROM goods_receipts WHERE id = v_invoice.goods_receipt_id;
    
    -- Check vendor match
    IF v_invoice.vendor_id != v_po.vendor_id THEN
        v_exceptions := v_exceptions || jsonb_build_object(
            'type', 'vendor_mismatch',
            'message', 'Invoice vendor does not match PO vendor'
        );
        v_match_result := 'exception';
    END IF;
    
    -- Check amount tolerance
    IF ABS(v_invoice.total_amount - v_po.total_amount) > (v_po.total_amount * p_tolerance_percentage / 100) THEN
        v_exceptions := v_exceptions || jsonb_build_object(
            'type', 'amount_tolerance_exceeded',
            'message', 'Invoice amount exceeds PO amount tolerance',
            'invoice_amount', v_invoice.total_amount,
            'po_amount', v_po.total_amount,
            'tolerance', p_tolerance_percentage
        );
        v_match_result := 'exception';
    END IF;
    
    -- Update invoice matching status
    UPDATE purchase_invoices 
    SET 
        matching_status = v_match_result,
        matching_exceptions = v_exceptions,
        matching_tolerance_exceeded = CASE WHEN v_match_result = 'exception' THEN true ELSE false END
    WHERE id = p_invoice_id;
    
    RETURN QUERY SELECT v_match_result, v_exceptions;
END;
$$ LANGUAGE plpgsql;
```

## 🏦 Cash Management

### Bank Account Management

```sql
-- Bank accounts
CREATE TABLE bank_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Account identification
    account_name VARCHAR(255) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    account_type VARCHAR(50) NOT NULL, -- checking, savings, money_market, credit_line
    
    -- Bank information
    bank_name VARCHAR(255) NOT NULL,
    bank_branch VARCHAR(255),
    routing_number VARCHAR(20),
    swift_code VARCHAR(11),
    iban VARCHAR(34),
    
    -- Account details
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    current_balance DECIMAL(15,2) DEFAULT 0,
    available_balance DECIMAL(15,2) DEFAULT 0,
    last_reconciled_balance DECIMAL(15,2) DEFAULT 0,
    last_reconciled_date DATE,
    
    -- Overdraft and limits
    overdraft_limit DECIMAL(15,2) DEFAULT 0,
    minimum_balance DECIMAL(15,2) DEFAULT 0,
    maximum_daily_withdrawal DECIMAL(15,2),
    
    -- Account settings
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    allow_online_payments BOOLEAN DEFAULT true,
    require_dual_approval BOOLEAN DEFAULT false,
    
    -- GL integration
    gl_account_id UUID NOT NULL REFERENCES accounts(id),
    
    -- Electronic banking
    online_banking_enabled BOOLEAN DEFAULT false,
    bank_api_integration BOOLEAN DEFAULT false,
    last_statement_import DATE,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_account_type CHECK (account_type IN ('checking', 'savings', 'money_market', 'credit_line')),
    UNIQUE(tenant_id, account_number, bank_name)
);
```

### Bank Reconciliation

```typescript
interface BankReconciliation {
  id: string;
  tenant_id: string;
  bank_account_id: string;
  
  reconciliation_date: Date;
  statement_begin_date: Date;
  statement_end_date: Date;
  
  statement_begin_balance: number;
  statement_end_balance: number;
  book_begin_balance: number;
  book_end_balance: number;
  
  reconciled_items: ReconciliationItem[];
  outstanding_deposits: ReconciliationItem[];
  outstanding_checks: ReconciliationItem[];
  bank_adjustments: ReconciliationItem[];
  book_adjustments: ReconciliationItem[];
  
  status: 'in_progress' | 'balanced' | 'unbalanced' | 'finalized';
  variance_amount: number;
  
  reconciled_by: string;
  reconciled_at: Date;
  notes: string;
}

interface ReconciliationItem {
  transaction_id: string;
  transaction_date: Date;
  description: string;
  amount: number;
  transaction_type: 'deposit' | 'withdrawal' | 'transfer' | 'fee' | 'adjustment';
  is_reconciled: boolean;
  reconciliation_notes: string;
}

class BankReconciliationService {
  async performAutoReconciliation(
    bankAccountId: string,
    statementItems: StatementItem[],
    reconciliationRules: ReconciliationRule[]
  ): Promise<AutoReconciliationResult> {
    
    const bookTransactions = await this.getUnreconciledTransactions(bankAccountId);
    const matches: ReconciliationMatch[] = [];
    const unmatchedStatement: StatementItem[] = [];
    const unmatchedBook: BookTransaction[] = [];
    
    // Apply matching rules
    for (const rule of reconciliationRules) {
      const ruleMatches = await this.applyRule(rule, statementItems, bookTransactions);
      matches.push(...ruleMatches);
    }
    
    // Identify unmatched items
    const matchedStatementIds = new Set(matches.map(m => m.statement_item_id));
    const matchedBookIds = new Set(matches.map(m => m.book_transaction_id));
    
    unmatchedStatement.push(
      ...statementItems.filter(item => !matchedStatementIds.has(item.id))
    );
    
    unmatchedBook.push(
      ...bookTransactions.filter(tx => !matchedBookIds.has(tx.id))
    );
    
    return {
      total_matches: matches.length,
      matched_amount: matches.reduce((sum, m) => sum + m.amount, 0),
      matches,
      unmatched_statement_items: unmatchedStatement,
      unmatched_book_transactions: unmatchedBook,
      confidence_score: this.calculateConfidenceScore(matches)
    };
  }
  
  private async applyRule(
    rule: ReconciliationRule,
    statementItems: StatementItem[],
    bookTransactions: BookTransaction[]
  ): Promise<ReconciliationMatch[]> {
    const matches: ReconciliationMatch[] = [];
    
    for (const statement of statementItems) {
      for (const book of bookTransactions) {
        if (this.evaluateRule(rule, statement, book)) {
          matches.push({
            statement_item_id: statement.id,
            book_transaction_id: book.id,
            amount: statement.amount,
            match_confidence: rule.confidence_weight,
            match_rule: rule.name
          });
        }
      }
    }
    
    return matches;
  }
}
```

## 📈 Financial Reporting

### Standard Financial Statements

```typescript
interface FinancialReportParams {
  tenant_id: string;
  report_type: 'balance_sheet' | 'income_statement' | 'cash_flow' | 'trial_balance';
  period_type: 'monthly' | 'quarterly' | 'yearly' | 'custom';
  start_date: Date;
  end_date: Date;
  comparison_period?: DateRange;
  
  filters?: {
    organization_ids?: string[];
    cost_center_ids?: string[];
    project_ids?: string[];
    currency_code?: string;
  };
  
  options?: {
    include_budget_comparison?: boolean;
    include_prior_year_comparison?: boolean;
    consolidate_subsidiaries?: boolean;
    show_accounts_with_zero_balance?: boolean;
  };
}

// Balance Sheet Generator
class BalanceSheetService {
  async generateBalanceSheet(params: FinancialReportParams): Promise<BalanceSheet> {
    const asOfDate = params.end_date;
    
    // Get account balances as of the report date
    const accountBalances = await this.getAccountBalances(
      params.tenant_id,
      asOfDate,
      params.filters
    );
    
    // Group accounts by balance sheet categories
    const assets = this.groupAccountsByCategory(accountBalances, 'asset');
    const liabilities = this.groupAccountsByCategory(accountBalances, 'liability');
    const equity = this.groupAccountsByCategory(accountBalances, 'equity');
    
    // Calculate totals
    const totalAssets = this.calculateTotal(assets);
    const totalLiabilities = this.calculateTotal(liabilities);
    const totalEquity = this.calculateTotal(equity);
    
    // Verify balance sheet equation: Assets = Liabilities + Equity
    const balanceCheck = Math.abs(totalAssets - (totalLiabilities + totalEquity));
    
    return {
      report_date: asOfDate,
      assets: {
        current_assets: this.filterBySubtype(assets, 'current_asset'),
        fixed_assets: this.filterBySubtype(assets, 'fixed_asset'),
        other_assets: this.filterBySubtype(assets, 'other_asset'),
        total: totalAssets
      },
      liabilities: {
        current_liabilities: this.filterBySubtype(liabilities, 'current_liability'),
        long_term_liabilities: this.filterBySubtype(liabilities, 'long_term_liability'),
        total: totalLiabilities
      },
      equity: {
        share_capital: this.filterBySubtype(equity, 'equity'),
        retained_earnings: this.filterBySubtype(equity, 'retained_earnings'),
        total: totalEquity
      },
      balance_verification: {
        is_balanced: balanceCheck < 0.01, // Allow for rounding differences
        variance: balanceCheck
      }
    };
  }
}

// Cash Flow Statement Generator
class CashFlowService {
  async generateCashFlowStatement(params: FinancialReportParams): Promise<CashFlowStatement> {
    const { start_date, end_date } = params;
    
    // Get cash flow activities
    const operatingActivities = await this.getOperatingCashFlows(params);
    const investingActivities = await this.getInvestingCashFlows(params);
    const financingActivities = await this.getFinancingCashFlows(params);
    
    // Calculate net cash flow
    const netOperatingCash = this.calculateTotal(operatingActivities);
    const netInvestingCash = this.calculateTotal(investingActivities);
    const netFinancingCash = this.calculateTotal(financingActivities);
    const netCashFlow = netOperatingCash + netInvestingCash + netFinancingCash;
    
    // Get beginning and ending cash balances
    const beginningCash = await this.getCashBalance(params.tenant_id, start_date);
    const endingCash = await this.getCashBalance(params.tenant_id, end_date);
    
    return {
      period: { start_date, end_date },
      operating_activities: operatingActivities,
      investing_activities: investingActivities,
      financing_activities: financingActivities,
      net_cash_flows: {
        from_operations: netOperatingCash,
        from_investing: netInvestingCash,
        from_financing: netFinancingCash,
        net_change: netCashFlow
      },
      cash_balances: {
        beginning_balance: beginningCash,
        ending_balance: endingCash,
        net_change: endingCash - beginningCash
      },
      reconciliation: {
        calculated_change: netCashFlow,
        actual_change: endingCash - beginningCash,
        variance: (endingCash - beginningCash) - netCashFlow
      }
    };
  }
}
```

This  financial management system provides robust accounting capabilities with proper controls, automation, and reporting to meet enterprise-level requirements while maintaining flexibility for different business needs and regulatory compliance.
