# Tax Management

##  Overview

The Tax Management module provides  tax calculation, compliance, and reporting capabilities for multiple jurisdictions and tax types. It supports various tax regimes including VAT/GST, sales tax, income tax, withholding tax, and custom tax structures with automated calculations and regulatory compliance.

## ️ Tax Configuration & Setup

### Tax Authorities and Jurisdictions

```sql
-- Tax authorities (government entities)
CREATE TABLE tax_authorities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Authority identification
    authority_code VARCHAR(20) UNIQUE NOT NULL,
    authority_name VARCHAR(255) NOT NULL,
    authority_type VARCHAR(50) NOT NULL, -- federal, state, local, municipal, vat, customs
    
    -- Jurisdiction information
    country_code CHAR(2) NOT NULL,
    state_province_code VARCHAR(10),
    city_name VARCHAR(100),
    jurisdiction_level VARCHAR(20) NOT NULL, -- national, state, local
    
    -- Authority details
    registration_number VARCHAR(100),
    tax_identification_number VARCHAR(100),
    website_url VARCHAR(255),
    contact_information JSONB,
    
    -- Filing requirements
    filing_frequency VARCHAR(20) DEFAULT 'monthly', -- weekly, monthly, quarterly, annual
    filing_due_day INTEGER DEFAULT 15, -- Day of month when filing is due
    payment_due_day INTEGER DEFAULT 15,
    
    -- Electronic filing
    supports_electronic_filing BOOLEAN DEFAULT false,
    electronic_filing_endpoint VARCHAR(255),
    api_credentials JSONB, -- Encrypted API keys and tokens
    
    -- Compliance settings
    requires_registration BOOLEAN DEFAULT true,
    minimum_threshold_amount DECIMAL(15,2),
    penalty_rate DECIMAL(5,4),
    interest_rate DECIMAL(5,4),
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    registration_date DATE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_authority_type CHECK (authority_type IN ('federal', 'state', 'local', 'municipal', 'vat', 'customs')),
    CONSTRAINT valid_jurisdiction_level CHECK (jurisdiction_level IN ('national', 'state', 'local')),
    CONSTRAINT valid_filing_frequency CHECK (filing_frequency IN ('weekly', 'monthly', 'quarterly', 'annual'))
);

-- Tax codes and rates
CREATE TABLE tax_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Tax code identification
    tax_code VARCHAR(20) UNIQUE NOT NULL,
    tax_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Tax classification
    tax_type VARCHAR(50) NOT NULL, -- sales_tax, vat, gst, income_tax, withholding_tax, excise_tax, customs_duty
    tax_category VARCHAR(100), -- standard, reduced, zero, exempt, reverse_charge
    
    -- Tax authority
    tax_authority_id UUID NOT NULL REFERENCES tax_authorities(id),
    
    -- Tax calculation
    calculation_method VARCHAR(30) DEFAULT 'percentage', -- percentage, fixed_amount, progressive, lookup_table
    tax_rate DECIMAL(8,5) NOT NULL DEFAULT 0, -- Tax rate as decimal (e.g., 0.15 for 15%)
    
    -- Tax on tax
    compound_tax BOOLEAN DEFAULT false, -- Tax calculated on price including other taxes
    cascade_order INTEGER DEFAULT 1, -- Order of calculation for multiple taxes
    
    -- Applicability rules
    effective_date DATE NOT NULL,
    expiry_date DATE,
    minimum_amount DECIMAL(15,2),
    maximum_amount DECIMAL(15,2),
    
    -- GL account mapping
    tax_payable_account_id UUID NOT NULL REFERENCES accounts(id),
    tax_expense_account_id UUID REFERENCES accounts(id),
    tax_receivable_account_id UUID REFERENCES accounts(id),
    
    -- Reporting configuration
    reporting_code VARCHAR(50), -- For tax authority reporting
    return_line_number VARCHAR(20), -- Line number in tax returns
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_tax_type CHECK (tax_type IN ('sales_tax', 'vat', 'gst', 'income_tax', 'withholding_tax', 'excise_tax', 'customs_duty')),
    CONSTRAINT valid_calculation_method CHECK (calculation_method IN ('percentage', 'fixed_amount', 'progressive', 'lookup_table')),
    CONSTRAINT valid_tax_rate CHECK (tax_rate >= 0 AND tax_rate <= 1)
);

-- Progressive tax brackets (for income tax, etc.)
CREATE TABLE tax_brackets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tax_code_id UUID NOT NULL REFERENCES tax_codes(id) ON DELETE CASCADE,
    
    -- Bracket definition
    bracket_number INTEGER NOT NULL,
    minimum_amount DECIMAL(15,2) NOT NULL,
    maximum_amount DECIMAL(15,2),
    tax_rate DECIMAL(8,5) NOT NULL,
    
    -- Bracket calculation
    marginal_calculation BOOLEAN DEFAULT true, -- True for marginal, false for flat rate on entire amount
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tax_code_id, bracket_number)
);

-- Tax exemptions and special cases
CREATE TABLE tax_exemptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Exemption identification
    exemption_code VARCHAR(20) UNIQUE NOT NULL,
    exemption_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Exemption scope
    tax_code_id UUID REFERENCES tax_codes(id),
    exemption_type VARCHAR(50) NOT NULL, -- total, partial, temporary, conditional
    
    -- Exemption conditions
    customer_category VARCHAR(100), -- government, nonprofit, resale, export
    product_category VARCHAR(100),
    transaction_type VARCHAR(100),
    minimum_order_amount DECIMAL(15,2),
    
    -- Geographic scope
    applicable_countries JSONB, -- Array of country codes
    applicable_states JSONB, -- Array of state codes
    
    -- Time validity
    effective_date DATE NOT NULL,
    expiry_date DATE,
    
    -- Exemption rate (for partial exemptions)
    exemption_percentage DECIMAL(5,2) DEFAULT 100, -- 100% = full exemption
    
    -- Documentation requirements
    requires_certificate BOOLEAN DEFAULT false,
    certificate_types JSONB, -- Array of acceptable certificate types
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_exemption_type CHECK (exemption_type IN ('total', 'partial', 'temporary', 'conditional'))
);
```

### Tax Rules Engine

```typescript
interface TaxCalculationEngine {
  calculateTax(transaction: TaxableTransaction): Promise<TaxCalculationResult>;
  validateTaxConfiguration(tenantId: string): Promise<ValidationResult[]>;
  determineTaxCodes(transaction: TaxableTransaction): Promise<ApplicableTaxCode[]>;
  applyTaxExemptions(transaction: TaxableTransaction, taxCodes: ApplicableTaxCode[]): Promise<ApplicableTaxCode[]>;
}

interface TaxableTransaction {
  transaction_id: string;
  transaction_type: 'sales' | 'purchase' | 'expense' | 'payroll';
  transaction_date: Date;
  
  // Parties involved
  customer_id?: string;
  vendor_id?: string;
  employee_id?: string;
  
  // Geographic information
  ship_from_address: Address;
  ship_to_address: Address;
  bill_to_address: Address;
  
  // Transaction lines
  lines: TaxableTransactionLine[];
  
  // Transaction-level discounts and adjustments
  discounts: TransactionDiscount[];
  freight_amount?: number;
  
  // Currency
  currency_code: string;
  exchange_rate: number;
}

interface TaxableTransactionLine {
  line_id: string;
  item_id?: string;
  item_category: string;
  description: string;
  
  quantity: number;
  unit_price: number;
  line_amount: number;
  discount_amount: number;
  net_amount: number;
  
  // Tax classification
  tax_category?: string;
  exempt_reason?: string;
  
  // Special handling
  is_gift: boolean;
  is_return: boolean;
  is_taxable: boolean;
}

interface TaxCalculationResult {
  transaction_id: string;
  calculation_date: Date;
  
  // Tax breakdown by jurisdiction
  tax_details: TaxDetail[];
  
  // Summary totals
  subtotal_amount: number;
  total_tax_amount: number;
  total_amount: number;
  
  // Tax by type
  sales_tax: number;
  use_tax: number;
  vat_tax: number;
  excise_tax: number;
  
  // Compliance information
  tax_registration_numbers: { [authority: string]: string };
  exemption_certificates: ExemptionCertificate[];
  
  // Calculation metadata
  calculation_method: string;
  confidence_score: number;
  warnings: string[];
  errors: string[];
}

interface TaxDetail {
  tax_code: string;
  tax_name: string;
  tax_authority: string;
  jurisdiction: string;
  
  taxable_amount: number;
  tax_rate: number;
  tax_amount: number;
  
  calculation_method: string;
  is_compound: boolean;
  
  // GL accounts for posting
  tax_payable_account: string;
  tax_expense_account?: string;
}

class TaxCalculationService implements TaxCalculationEngine {
  async calculateTax(transaction: TaxableTransaction): Promise<TaxCalculationResult> {
    // Step 1: Determine applicable tax codes based on jurisdictions
    const applicableTaxCodes = await this.determineTaxCodes(transaction);
    
    // Step 2: Apply exemptions and special rules
    const effectiveTaxCodes = await this.applyTaxExemptions(transaction, applicableTaxCodes);
    
    // Step 3: Calculate taxes for each line and jurisdiction
    const taxDetails: TaxDetail[] = [];
    let totalTaxAmount = 0;
    
    for (const line of transaction.lines) {
      const lineTaxDetails = await this.calculateLineTax(line, effectiveTaxCodes, transaction);
      taxDetails.push(...lineTaxDetails);
      totalTaxAmount += lineTaxDetails.reduce((sum, detail) => sum + detail.tax_amount, 0);
    }
    
    // Step 4: Handle freight and other charges
    if (transaction.freight_amount && transaction.freight_amount > 0) {
      const freightTaxDetails = await this.calculateFreightTax(transaction.freight_amount, effectiveTaxCodes, transaction);
      taxDetails.push(...freightTaxDetails);
      totalTaxAmount += freightTaxDetails.reduce((sum, detail) => sum + detail.tax_amount, 0);
    }
    
    // Step 5: Apply compound tax calculations
    const compoundTaxDetails = await this.calculateCompoundTaxes(taxDetails, effectiveTaxCodes);
    taxDetails.push(...compoundTaxDetails);
    totalTaxAmount += compoundTaxDetails.reduce((sum, detail) => sum + detail.tax_amount, 0);
    
    // Step 6: Aggregate results
    const subtotalAmount = transaction.lines.reduce((sum, line) => sum + line.net_amount, 0);
    
    return {
      transaction_id: transaction.transaction_id,
      calculation_date: new Date(),
      tax_details: taxDetails,
      subtotal_amount: subtotalAmount,
      total_tax_amount: totalTaxAmount,
      total_amount: subtotalAmount + totalTaxAmount,
      sales_tax: this.aggregateTaxByType(taxDetails, 'sales_tax'),
      use_tax: this.aggregateTaxByType(taxDetails, 'use_tax'),
      vat_tax: this.aggregateTaxByType(taxDetails, 'vat'),
      excise_tax: this.aggregateTaxByType(taxDetails, 'excise_tax'),
      tax_registration_numbers: await this.getTaxRegistrationNumbers(transaction),
      exemption_certificates: await this.getApplicableExemptions(transaction),
      calculation_method: 'standard',
      confidence_score: 100,
      warnings: [],
      errors: []
    };
  }
  
  async determineTaxCodes(transaction: TaxableTransaction): Promise<ApplicableTaxCode[]> {
    const applicableCodes: ApplicableTaxCode[] = [];
    
    // Determine jurisdiction based on shipping addresses
    const jurisdictions = await this.determineJurisdictions(transaction);
    
    for (const jurisdiction of jurisdictions) {
      // Get active tax codes for this jurisdiction
      const taxCodes = await this.getTaxCodesForJurisdiction(
        jurisdiction.country_code,
        jurisdiction.state_code,
        jurisdiction.city_code,
        transaction.transaction_date
      );
      
      for (const taxCode of taxCodes) {
        // Check if tax code applies to this transaction type
        if (this.isTaxCodeApplicable(taxCode, transaction)) {
          applicableCodes.push({
            tax_code: taxCode,
            jurisdiction: jurisdiction,
            applicability_score: this.calculateApplicabilityScore(taxCode, transaction)
          });
        }
      }
    }
    
    // Sort by applicability score (highest first)
    return applicableCodes.sort((a, b) => b.applicability_score - a.applicability_score);
  }
  
  private async calculateLineTax(
    line: TaxableTransactionLine, 
    taxCodes: ApplicableTaxCode[], 
    transaction: TaxableTransaction
  ): Promise<TaxDetail[]> {
    const lineTaxDetails: TaxDetail[] = [];
    
    if (!line.is_taxable) {
      return lineTaxDetails;
    }
    
    for (const applicableCode of taxCodes) {
      const taxCode = applicableCode.tax_code;
      
      // Check if this tax code applies to this line item
      if (!this.doesTaxCodeApplyToLine(taxCode, line)) {
        continue;
      }
      
      let taxableAmount = line.net_amount;
      let taxAmount = 0;
      
      switch (taxCode.calculation_method) {
        case 'percentage':
          taxAmount = taxableAmount * taxCode.tax_rate;
          break;
          
        case 'fixed_amount':
          taxAmount = taxCode.tax_rate; // In this case, tax_rate stores the fixed amount
          break;
          
        case 'progressive':
          taxAmount = await this.calculateProgressiveTax(taxableAmount, taxCode.id);
          break;
          
        case 'lookup_table':
          taxAmount = await this.calculateLookupTableTax(taxableAmount, line, taxCode.id);
          break;
      }
      
      // Apply minimum and maximum limits
      if (taxCode.minimum_amount && taxAmount < taxCode.minimum_amount) {
        taxAmount = taxCode.minimum_amount;
      }
      if (taxCode.maximum_amount && taxAmount > taxCode.maximum_amount) {
        taxAmount = taxCode.maximum_amount;
      }
      
      lineTaxDetails.push({
        tax_code: taxCode.tax_code,
        tax_name: taxCode.tax_name,
        tax_authority: applicableCode.jurisdiction.authority_name,
        jurisdiction: applicableCode.jurisdiction.jurisdiction_name,
        taxable_amount: taxableAmount,
        tax_rate: taxCode.tax_rate,
        tax_amount: taxAmount,
        calculation_method: taxCode.calculation_method,
        is_compound: taxCode.compound_tax,
        tax_payable_account: taxCode.tax_payable_account_id,
        tax_expense_account: taxCode.tax_expense_account_id
      });
    }
    
    return lineTaxDetails;
  }
  
  private async calculateProgressiveTax(taxableAmount: number, taxCodeId: string): Promise<number> {
    const brackets = await this.getTaxBrackets(taxCodeId);
    let totalTax = 0;
    let remainingAmount = taxableAmount;
    
    for (const bracket of brackets.sort((a, b) => a.minimum_amount - b.minimum_amount)) {
      if (remainingAmount <= 0) break;
      
      const bracketMin = bracket.minimum_amount;
      const bracketMax = bracket.maximum_amount || Infinity;
      const bracketRate = bracket.tax_rate;
      
      if (taxableAmount > bracketMin) {
        const taxableInBracket = Math.min(remainingAmount, bracketMax - bracketMin);
        
        if (bracket.marginal_calculation) {
          // Marginal tax calculation
          totalTax += taxableInBracket * bracketRate;
        } else {
          // Flat rate on entire amount
          totalTax = taxableAmount * bracketRate;
          break;
        }
        
        remainingAmount -= taxableInBracket;
      }
    }
    
    return totalTax;
  }
}
```

##  Tax Transactions & Records

### Tax Transaction Tracking

```sql
-- Tax transactions for all tax-related activities
CREATE TABLE tax_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Transaction identification
    transaction_number VARCHAR(50) UNIQUE NOT NULL,
    source_transaction_id UUID NOT NULL, -- Reference to source document
    source_transaction_type VARCHAR(50) NOT NULL, -- sales_invoice, purchase_invoice, expense, payroll
    
    -- Tax calculation details
    tax_calculation_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tax_period_start DATE NOT NULL,
    tax_period_end DATE NOT NULL,
    
    -- Geographic information
    origin_country_code CHAR(2),
    origin_state_code VARCHAR(10),
    destination_country_code CHAR(2),
    destination_state_code VARCHAR(10),
    
    -- Transaction amounts
    subtotal_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    -- Currency
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1,
    base_currency_tax_amount DECIMAL(15,2),
    
    -- Tax status
    tax_status VARCHAR(20) DEFAULT 'calculated', -- calculated, filed, paid, adjusted, reversed
    calculation_method VARCHAR(30) DEFAULT 'automatic', -- automatic, manual, imported
    
    -- Compliance tracking
    requires_filing BOOLEAN DEFAULT true,
    filed_date DATE,
    payment_due_date DATE,
    payment_date DATE,
    
    -- Audit trail
    original_calculation JSONB, -- Store original tax calculation details
    adjustments JSONB, -- Store any subsequent adjustments
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_tax_status CHECK (tax_status IN ('calculated', 'filed', 'paid', 'adjusted', 'reversed')),
    CONSTRAINT valid_calculation_method CHECK (calculation_method IN ('automatic', 'manual', 'imported'))
);

-- Tax transaction lines (detailed breakdown)
CREATE TABLE tax_transaction_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tax_transaction_id UUID NOT NULL REFERENCES tax_transactions(id) ON DELETE CASCADE,
    
    -- Line identification
    line_number INTEGER NOT NULL,
    source_line_id UUID, -- Reference to source document line
    
    -- Tax code and authority
    tax_code_id UUID NOT NULL REFERENCES tax_codes(id),
    tax_authority_id UUID NOT NULL REFERENCES tax_authorities(id),
    
    -- Taxable amount and calculation
    taxable_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    tax_rate DECIMAL(8,5) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    -- Tax classification
    tax_type VARCHAR(50) NOT NULL,
    tax_category VARCHAR(100),
    jurisdiction_level VARCHAR(20),
    
    -- GL posting accounts
    tax_payable_account_id UUID REFERENCES accounts(id),
    tax_expense_account_id UUID REFERENCES accounts(id),
    
    -- Exemption information
    is_exempt BOOLEAN DEFAULT false,
    exemption_code VARCHAR(20),
    exemption_certificate_number VARCHAR(100),
    
    -- Line description
    description TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tax_transaction_id, line_number)
);

-- Tax exemption certificates
CREATE TABLE tax_exemption_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Certificate identification
    certificate_number VARCHAR(100) UNIQUE NOT NULL,
    certificate_type VARCHAR(50) NOT NULL, -- resale, government, nonprofit, manufacturing, export
    
    -- Certificate holder (customer/vendor)
    customer_id UUID REFERENCES customers(id),
    vendor_id UUID REFERENCES vendors(id),
    
    -- Certificate details
    issuing_authority VARCHAR(255),
    issue_date DATE NOT NULL,
    expiry_date DATE,
    
    -- Applicability
    applicable_tax_types JSONB, -- Array of tax types this certificate covers
    applicable_jurisdictions JSONB, -- Array of jurisdictions where valid
    applicable_products JSONB, -- Array of product categories (if limited)
    
    -- Certificate file
    certificate_file_url VARCHAR(500),
    
    -- Verification
    verification_status VARCHAR(20) DEFAULT 'pending', -- pending, verified, invalid, expired
    verified_by UUID REFERENCES users(id),
    verified_at TIMESTAMPTZ,
    verification_notes TEXT,
    
    -- Status
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, expired, revoked
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_certificate_type CHECK (certificate_type IN ('resale', 'government', 'nonprofit', 'manufacturing', 'export')),
    CONSTRAINT valid_verification_status CHECK (verification_status IN ('pending', 'verified', 'invalid', 'expired')),
    CONSTRAINT valid_certificate_status CHECK (status IN ('active', 'inactive', 'expired', 'revoked'))
);
```

##  Tax Returns & Compliance

### Tax Return Management

```sql
-- Tax return periods
CREATE TABLE tax_return_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Period identification
    tax_authority_id UUID NOT NULL REFERENCES tax_authorities(id),
    return_type VARCHAR(50) NOT NULL, -- vat, sales_tax, income_tax, withholding_tax
    
    -- Period dates
    period_start_date DATE NOT NULL,
    period_end_date DATE NOT NULL,
    filing_due_date DATE NOT NULL,
    payment_due_date DATE NOT NULL,
    
    -- Period status
    status VARCHAR(20) DEFAULT 'open', -- open, prepared, filed, paid, closed, amended
    
    -- Automatic generation
    auto_generated BOOLEAN DEFAULT false,
    next_period_id UUID REFERENCES tax_return_periods(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_return_type CHECK (return_type IN ('vat', 'sales_tax', 'income_tax', 'withholding_tax')),
    CONSTRAINT valid_period_status CHECK (status IN ('open', 'prepared', 'filed', 'paid', 'closed', 'amended')),
    UNIQUE(tenant_id, tax_authority_id, return_type, period_start_date)
);

-- Tax returns
CREATE TABLE tax_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Return identification
    return_number VARCHAR(50) UNIQUE NOT NULL,
    tax_return_period_id UUID NOT NULL REFERENCES tax_return_periods(id),
    
    -- Return details
    return_type VARCHAR(50) NOT NULL,
    tax_authority_id UUID NOT NULL REFERENCES tax_authorities(id),
    
    -- Filing information
    filing_method VARCHAR(20) DEFAULT 'electronic', -- electronic, paper, agent
    filing_date DATE,
    confirmation_number VARCHAR(100),
    
    -- Financial summary
    total_sales DECIMAL(15,2) DEFAULT 0,
    taxable_sales DECIMAL(15,2) DEFAULT 0,
    exempt_sales DECIMAL(15,2) DEFAULT 0,
    
    tax_collected DECIMAL(15,2) DEFAULT 0,
    tax_paid DECIMAL(15,2) DEFAULT 0,
    tax_due DECIMAL(15,2) DEFAULT 0,
    tax_refund DECIMAL(15,2) DEFAULT 0,
    
    penalties DECIMAL(15,2) DEFAULT 0,
    interest DECIMAL(15,2) DEFAULT 0,
    
    net_amount_due DECIMAL(15,2) DEFAULT 0,
    
    -- Payment information
    payment_method VARCHAR(50),
    payment_reference VARCHAR(100),
    payment_date DATE,
    
    -- Return status
    status VARCHAR(20) DEFAULT 'draft', -- draft, prepared, filed, paid, accepted, rejected, amended
    
    -- Amendment information
    amended_return_id UUID REFERENCES tax_returns(id),
    amendment_reason TEXT,
    
    -- Attachments and documentation
    return_file_url VARCHAR(500),
    supporting_documents JSONB,
    
    -- Preparation details
    prepared_by UUID REFERENCES users(id),
    reviewed_by UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    
    prepared_at TIMESTAMPTZ,
    filed_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_filing_method CHECK (filing_method IN ('electronic', 'paper', 'agent')),
    CONSTRAINT valid_return_status CHECK (status IN ('draft', 'prepared', 'filed', 'paid', 'accepted', 'rejected', 'amended'))
);

-- Tax return line items
CREATE TABLE tax_return_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tax_return_id UUID NOT NULL REFERENCES tax_returns(id) ON DELETE CASCADE,
    
    -- Line identification
    line_number VARCHAR(20) NOT NULL, -- Official line number from tax form
    line_description TEXT NOT NULL,
    line_category VARCHAR(100), -- sales, purchases, adjustments, calculations
    
    -- Line amounts
    line_amount DECIMAL(15,2) DEFAULT 0,
    
    -- Data source
    calculated_amount DECIMAL(15,2) DEFAULT 0, -- System calculated amount
    manual_override BOOLEAN DEFAULT false,
    override_reason TEXT,
    
    -- Supporting data
    supporting_transactions JSONB, -- Array of transaction IDs that contribute to this line
    calculation_details JSONB, -- Breakdown of how amount was calculated
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tax_return_id, line_number)
);

-- Tax payment tracking
CREATE TABLE tax_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Payment identification
    payment_number VARCHAR(50) UNIQUE NOT NULL,
    tax_return_id UUID REFERENCES tax_returns(id),
    tax_authority_id UUID NOT NULL REFERENCES tax_authorities(id),
    
    -- Payment details
    payment_type VARCHAR(30) NOT NULL, -- estimated, final, penalty, interest, refund
    payment_date DATE NOT NULL,
    payment_method VARCHAR(50) NOT NULL, -- ach, wire, check, credit_card, online
    
    -- Payment amounts
    tax_amount DECIMAL(15,2) DEFAULT 0,
    penalty_amount DECIMAL(15,2) DEFAULT 0,
    interest_amount DECIMAL(15,2) DEFAULT 0,
    total_payment_amount DECIMAL(15,2) NOT NULL,
    
    -- Payment processing
    payment_reference VARCHAR(100),
    bank_account_id UUID REFERENCES bank_accounts(id),
    confirmation_number VARCHAR(100),
    
    -- Status
    status VARCHAR(20) DEFAULT 'pending', -- pending, processed, failed, cancelled, refunded
    processed_date DATE,
    
    -- GL integration
    journal_entry_id UUID REFERENCES journal_entries(id),
    posted_to_gl BOOLEAN DEFAULT false,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_payment_type CHECK (payment_type IN ('estimated', 'final', 'penalty', 'interest', 'refund')),
    CONSTRAINT valid_payment_method CHECK (payment_method IN ('ach', 'wire', 'check', 'credit_card', 'online')),
    CONSTRAINT valid_payment_status CHECK (status IN ('pending', 'processed', 'failed', 'cancelled', 'refunded'))
);
```

### Automated Tax Return Generation

```typescript
interface TaxReturnGenerator {
  generateTaxReturn(periodId: string): Promise<TaxReturn>;
  calculateReturnLines(returnId: string): Promise<TaxReturnLine[]>;
  validateTaxReturn(returnId: string): Promise<ValidationResult>;
  submitTaxReturn(returnId: string): Promise<SubmissionResult>;
}

interface TaxReturnCalculator {
  period_id: string;
  tax_authority: TaxAuthority;
  return_type: string;
  
  sales_summary: {
    total_sales: number;
    taxable_sales: number;
    exempt_sales: number;
    export_sales: number;
  };
  
  tax_summary: {
    output_tax: number; // Tax on sales
    input_tax: number;  // Tax on purchases
    net_tax_due: number;
    adjustments: number;
  };
  
  transaction_details: TaxTransactionSummary[];
}

class VATReturnService implements TaxReturnGenerator {
  async generateTaxReturn(periodId: string): Promise<TaxReturn> {
    const period = await this.getTaxReturnPeriod(periodId);
    const authority = await this.getTaxAuthority(period.tax_authority_id);
    
    // Gather all tax transactions for the period
    const taxTransactions = await this.getTaxTransactionsForPeriod(
      period.tenant_id,
      period.period_start_date,
      period.period_end_date,
      authority.id
    );
    
    // Calculate return summary
    const calculator = await this.calculateReturnSummary(taxTransactions, period);
    
    // Create tax return record
    const taxReturn = await this.createTaxReturn({
      tenant_id: period.tenant_id,
      return_number: await this.generateReturnNumber(period),
      tax_return_period_id: periodId,
      return_type: period.return_type,
      tax_authority_id: authority.id,
      total_sales: calculator.sales_summary.total_sales,
      taxable_sales: calculator.sales_summary.taxable_sales,
      exempt_sales: calculator.sales_summary.exempt_sales,
      tax_collected: calculator.tax_summary.output_tax,
      tax_paid: calculator.tax_summary.input_tax,
      tax_due: Math.max(0, calculator.tax_summary.net_tax_due),
      tax_refund: Math.max(0, -calculator.tax_summary.net_tax_due),
      net_amount_due: calculator.tax_summary.net_tax_due + calculator.tax_summary.adjustments,
      status: 'draft'
    });
    
    // Generate return lines based on authority requirements
    await this.generateReturnLines(taxReturn.id, calculator, authority);
    
    return taxReturn;
  }
  
  private async calculateReturnSummary(
    transactions: TaxTransaction[], 
    period: TaxReturnPeriod
  ): Promise<TaxReturnCalculator> {
    
    // Separate sales and purchase transactions
    const salesTransactions = transactions.filter(t => t.source_transaction_type.includes('sales'));
    const purchaseTransactions = transactions.filter(t => t.source_transaction_type.includes('purchase'));
    
    // Calculate sales summary
    const salesSummary = {
      total_sales: salesTransactions.reduce((sum, t) => sum + t.subtotal_amount, 0),
      taxable_sales: salesTransactions.reduce((sum, t) => sum + this.getTaxableAmount(t), 0),
      exempt_sales: salesTransactions.reduce((sum, t) => sum + this.getExemptAmount(t), 0),
      export_sales: salesTransactions.reduce((sum, t) => sum + this.getExportAmount(t), 0)
    };
    
    // Calculate tax summary
    const outputTax = salesTransactions.reduce((sum, t) => sum + t.total_tax_amount, 0);
    const inputTax = purchaseTransactions.reduce((sum, t) => sum + t.total_tax_amount, 0);
    
    const taxSummary = {
      output_tax: outputTax,
      input_tax: inputTax,
      net_tax_due: outputTax - inputTax,
      adjustments: await this.calculateAdjustments(period.tenant_id, period)
    };
    
    return {
      period_id: period.id,
      tax_authority: await this.getTaxAuthority(period.tax_authority_id),
      return_type: period.return_type,
      sales_summary: salesSummary,
      tax_summary: taxSummary,
      transaction_details: await this.summarizeTransactions(transactions)
    };
  }
  
  private async generateReturnLines(
    returnId: string, 
    calculator: TaxReturnCalculator, 
    authority: TaxAuthority
  ): Promise<void> {
    
    // Get return line template for this authority
    const lineTemplate = await this.getReturnLineTemplate(authority.id, calculator.return_type);
    
    const returnLines: TaxReturnLine[] = [];
    
    for (const templateLine of lineTemplate.lines) {
      let lineAmount = 0;
      
      switch (templateLine.calculation_method) {
        case 'total_sales':
          lineAmount = calculator.sales_summary.total_sales;
          break;
          
        case 'taxable_sales':
          lineAmount = calculator.sales_summary.taxable_sales;
          break;
          
        case 'exempt_sales':
          lineAmount = calculator.sales_summary.exempt_sales;
          break;
          
        case 'output_tax':
          lineAmount = calculator.tax_summary.output_tax;
          break;
          
        case 'input_tax':
          lineAmount = calculator.tax_summary.input_tax;
          break;
          
        case 'net_tax':
          lineAmount = calculator.tax_summary.net_tax_due;
          break;
          
        case 'formula':
          lineAmount = await this.evaluateFormula(templateLine.formula, calculator);
          break;
      }
      
      returnLines.push({
        tax_return_id: returnId,
        line_number: templateLine.line_number,
        line_description: templateLine.description,
        line_category: templateLine.category,
        line_amount: lineAmount,
        calculated_amount: lineAmount,
        manual_override: false,
        calculation_details: {
          method: templateLine.calculation_method,
          source_data: this.getSourceDataReferences(templateLine, calculator)
        }
      });
    }
    
    await this.saveTaxReturnLines(returnLines);
  }
}
```

##  Multi-Jurisdiction Tax Support

### International Tax Features

```sql
-- Tax treaties between countries
CREATE TABLE tax_treaties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Treaty identification
    treaty_name VARCHAR(255) NOT NULL,
    country_a_code CHAR(2) NOT NULL,
    country_b_code CHAR(2) NOT NULL,
    
    -- Treaty details
    effective_date DATE NOT NULL,
    expiry_date DATE,
    treaty_type VARCHAR(50) DEFAULT 'double_taxation', -- double_taxation, trade, customs
    
    -- Tax rate reductions
    dividend_tax_rate DECIMAL(5,4),
    interest_tax_rate DECIMAL(5,4),
    royalty_tax_rate DECIMAL(5,4),
    capital_gains_exemption BOOLEAN DEFAULT false,
    
    -- Treaty provisions
    permanent_establishment_rules TEXT,
    tie_breaker_rules TEXT,
    mutual_agreement_procedure TEXT,
    
    -- Documentation requirements
    required_certificates JSONB,
    
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_treaty_type CHECK (treaty_type IN ('double_taxation', 'trade', 'customs')),
    UNIQUE(country_a_code, country_b_code, treaty_type)
);

-- Withholding tax rates
CREATE TABLE withholding_tax_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Rate identification
    source_country_code CHAR(2) NOT NULL, -- Country where income is sourced
    resident_country_code CHAR(2) NOT NULL, -- Country of tax residence
    
    -- Income type and rates
    income_type VARCHAR(50) NOT NULL, -- dividend, interest, royalty, service_fee, rent
    standard_rate DECIMAL(5,4) NOT NULL, -- Standard withholding rate
    treaty_rate DECIMAL(5,4), -- Reduced rate under tax treaty
    
    -- Applicability
    effective_date DATE NOT NULL,
    expiry_date DATE,
    
    -- Exemptions and conditions
    minimum_ownership_percentage DECIMAL(5,2), -- For dividend exemptions
    minimum_holding_period_days INTEGER, -- For capital gains
    exemption_conditions TEXT,
    
    -- Related treaty
    tax_treaty_id UUID REFERENCES tax_treaties(id),
    
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_income_type CHECK (income_type IN ('dividend', 'interest', 'royalty', 'service_fee', 'rent', 'capital_gain')),
    UNIQUE(source_country_code, resident_country_code, income_type, effective_date)
);

-- Transfer pricing documentation
CREATE TABLE transfer_pricing_docs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Documentation identification
    document_type VARCHAR(50) NOT NULL, -- master_file, local_file, cbcr, economic_analysis
    reporting_year INTEGER NOT NULL,
    
    -- Entities covered
    legal_entity_id UUID REFERENCES legal_entities(id),
    jurisdictions_covered JSONB, -- Array of country codes
    
    -- Filing requirements
    filing_deadline DATE,
    filed_date DATE,
    filing_status VARCHAR(20) DEFAULT 'pending', -- pending, filed, accepted, rejected
    
    -- Document content
    document_file_url VARCHAR(500),
    summary_description TEXT,
    
    -- Compliance tracking
    requires_update BOOLEAN DEFAULT false,
    last_updated_date DATE,
    next_review_date DATE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_document_type CHECK (document_type IN ('master_file', 'local_file', 'cbcr', 'economic_analysis')),
    CONSTRAINT valid_filing_status CHECK (filing_status IN ('pending', 'filed', 'accepted', 'rejected'))
);
```

This  tax management system provides robust multi-jurisdiction tax handling, automated calculations, compliance tracking, and reporting capabilities to meet complex international tax requirements while maintaining accuracy and regulatory compliance.
