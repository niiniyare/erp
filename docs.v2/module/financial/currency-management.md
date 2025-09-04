# AWO ERP Currency Management Module

**Version**: 1.0  
**Date**: January 2025  
**Status**: Technical Specification  

---

## 🌍 Overview

The Currency Management module provides  multi-currency support for global business operations. It handles exchange rate management, currency conversion, hedging operations, and compliance with international financial reporting standards.

## 💱 Currency Configuration & Master Data

### Currency Master Data

```sql
-- Supported currencies with  details
CREATE TABLE currencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Currency identification
    currency_code CHAR(3) UNIQUE NOT NULL, -- ISO 4217 code (USD, EUR, GBP)
    currency_name VARCHAR(100) NOT NULL,
    currency_symbol VARCHAR(10),
    
    -- Currency formatting
    decimal_places INTEGER DEFAULT 2,
    decimal_separator CHAR(1) DEFAULT '.',
    thousands_separator CHAR(1) DEFAULT ',',
    symbol_position VARCHAR(10) DEFAULT 'before', -- before, after
    
    -- Currency attributes
    is_base_currency BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    is_crypto_currency BOOLEAN DEFAULT false,
    
    -- Regional information
    primary_country_code CHAR(2),
    secondary_countries JSONB, -- Array of country codes where used
    
    -- Trading characteristics
    is_freely_convertible BOOLEAN DEFAULT true,
    trading_start_time TIME DEFAULT '00:00:00',
    trading_end_time TIME DEFAULT '23:59:59',
    trading_timezone VARCHAR(50),
    
    -- Compliance and regulations
    requires_central_bank_approval BOOLEAN DEFAULT false,
    exchange_control_restrictions JSONB,
    
    -- Historical tracking
    introduced_date DATE,
    last_revaluation_date DATE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_symbol_position CHECK (symbol_position IN ('before', 'after'))
);

-- Exchange rate providers and sources
CREATE TABLE exchange_rate_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Provider identification
    provider_code VARCHAR(20) UNIQUE NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    provider_type VARCHAR(30) NOT NULL, -- central_bank, commercial_bank, financial_service, manual
    
    -- API configuration
    api_endpoint VARCHAR(500),
    api_key_encrypted VARCHAR(500), -- Encrypted API credentials
    authentication_method VARCHAR(30), -- api_key, oauth, basic_auth
    
    -- Data characteristics
    update_frequency VARCHAR(20) DEFAULT 'daily', -- real_time, hourly, daily, weekly
    rate_type VARCHAR(20) DEFAULT 'mid', -- bid, ask, mid, official
    base_currency CHAR(3) DEFAULT 'USD',
    
    -- Reliability and priority
    priority_order INTEGER DEFAULT 100, -- Lower = higher priority
    reliability_score DECIMAL(3,1) DEFAULT 100.0, -- 0-100 reliability rating
    
    -- Operational settings
    is_active BOOLEAN DEFAULT true,
    is_primary BOOLEAN DEFAULT false,
    last_successful_update TIMESTAMPTZ,
    
    -- Rate limits and costs
    daily_request_limit INTEGER,
    cost_per_request DECIMAL(8,4),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_provider_type CHECK (provider_type IN ('central_bank', 'commercial_bank', 'financial_service', 'manual')),
    CONSTRAINT valid_update_frequency CHECK (update_frequency IN ('real_time', 'hourly', 'daily', 'weekly')),
    CONSTRAINT valid_rate_type CHECK (rate_type IN ('bid', 'ask', 'mid', 'official'))
);

-- Exchange rates with  tracking
CREATE TABLE exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Rate identification
    from_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code),
    to_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code),
    rate_date DATE NOT NULL,
    rate_time TIME DEFAULT CURRENT_TIME,
    
    -- Rate values
    exchange_rate DECIMAL(18,8) NOT NULL, -- High precision for crypto and exotic pairs
    inverse_rate DECIMAL(18,8), -- Calculated inverse rate
    
    -- Rate types
    rate_type VARCHAR(20) DEFAULT 'spot', -- spot, forward, budget, historical, average
    rate_source VARCHAR(30) DEFAULT 'mid', -- bid, ask, mid, official, closing
    
    -- Rate provider
    provider_id UUID REFERENCES exchange_rate_providers(id),
    source_reference VARCHAR(100), -- Provider's reference ID
    
    -- Rate validity
    effective_from TIMESTAMPTZ DEFAULT NOW(),
    effective_to TIMESTAMPTZ,
    
    -- Rate spreads (for trading)
    bid_rate DECIMAL(18,8),
    ask_rate DECIMAL(18,8),
    spread_percentage DECIMAL(8,6),
    
    -- Forward rate specific fields
    forward_points DECIMAL(10,6),
    maturity_date DATE, -- For forward contracts
    
    -- Volatility and analytics
    volatility_measure DECIMAL(8,6),
    daily_change_percentage DECIMAL(8,4),
    
    -- Quality indicators
    confidence_level DECIMAL(5,2) DEFAULT 100.0, -- Confidence in rate accuracy
    liquidity_indicator VARCHAR(10) DEFAULT 'high', -- high, medium, low
    
    -- Audit and compliance
    is_official_rate BOOLEAN DEFAULT false, -- Central bank or official rate
    regulatory_approved BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_rate_type CHECK (rate_type IN ('spot', 'forward', 'budget', 'historical', 'average')),
    CONSTRAINT valid_rate_source CHECK (rate_source IN ('bid', 'ask', 'mid', 'official', 'closing')),
    CONSTRAINT valid_liquidity CHECK (liquidity_indicator IN ('high', 'medium', 'low')),
    CONSTRAINT positive_exchange_rate CHECK (exchange_rate > 0),
    CONSTRAINT different_currencies CHECK (from_currency != to_currency),
    
    UNIQUE(from_currency, to_currency, rate_date, rate_type, provider_id)
);

-- Exchange rate history for analytical purposes
CREATE TABLE exchange_rate_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Historical rate data
    from_currency CHAR(3) NOT NULL,
    to_currency CHAR(3) NOT NULL,
    rate_period_start DATE NOT NULL,
    rate_period_end DATE NOT NULL,
    
    -- Aggregated rates
    opening_rate DECIMAL(18,8),
    closing_rate DECIMAL(18,8),
    high_rate DECIMAL(18,8),
    low_rate DECIMAL(18,8),
    average_rate DECIMAL(18,8),
    weighted_average_rate DECIMAL(18,8),
    
    -- Volume and activity
    trading_volume DECIMAL(18,2),
    number_of_updates INTEGER DEFAULT 0,
    
    -- Statistical measures
    standard_deviation DECIMAL(10,8),
    coefficient_of_variation DECIMAL(8,4),
    
    -- Trend analysis
    trend_direction VARCHAR(10), -- up, down, stable
    momentum_indicator DECIMAL(8,4),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_trend_direction CHECK (trend_direction IN ('up', 'down', 'stable')),
    CONSTRAINT valid_period CHECK (rate_period_end >= rate_period_start)
);
```

### Currency Conversion Engine

```typescript
interface CurrencyConversionEngine {
  convertAmount(request: ConversionRequest): Promise<ConversionResult>;
  getExchangeRate(fromCurrency: string, toCurrency: string, rateDate?: Date, rateType?: string): Promise<ExchangeRate>;
  calculateCrossRate(baseCurrency: string, fromCurrency: string, toCurrency: string): Promise<CrossRateResult>;
  getBulkRates(request: BulkRateRequest): Promise<BulkRateResponse>;
}

interface ConversionRequest {
  amount: number;
  from_currency: string;
  to_currency: string;
  rate_date?: Date;
  rate_type?: 'spot' | 'budget' | 'forward' | 'average';
  provider_preference?: string[];
  precision?: number;
}

interface ConversionResult {
  original_amount: number;
  converted_amount: number;
  exchange_rate: number;
  inverse_rate: number;
  from_currency: string;
  to_currency: string;
  conversion_date: Date;
  rate_source: ExchangeRateSource;
  confidence_level: number;
  calculation_method: string;
}

interface ExchangeRateSource {
  provider_name: string;
  rate_type: string;
  rate_time: Date;
  reliability_score: number;
}

class CurrencyService implements CurrencyConversionEngine {
  async convertAmount(request: ConversionRequest): Promise<ConversionResult> {
    // Validate currencies
    await this.validateCurrencies(request.from_currency, request.to_currency);
    
    // Handle same currency conversion
    if (request.from_currency === request.to_currency) {
      return this.createSameCurrencyResult(request);
    }
    
    // Get exchange rate
    const exchangeRate = await this.getExchangeRate(
      request.from_currency,
      request.to_currency,
      request.rate_date,
      request.rate_type
    );
    
    // Apply precision settings
    const precision = request.precision || await this.getCurrencyPrecision(request.to_currency);
    
    // Perform conversion
    const convertedAmount = this.applyPrecision(
      request.amount * exchangeRate.rate,
      precision
    );
    
    return {
      original_amount: request.amount,
      converted_amount: convertedAmount,
      exchange_rate: exchangeRate.rate,
      inverse_rate: 1 / exchangeRate.rate,
      from_currency: request.from_currency,
      to_currency: request.to_currency,
      conversion_date: exchangeRate.rate_date,
      rate_source: {
        provider_name: exchangeRate.provider_name,
        rate_type: exchangeRate.rate_type,
        rate_time: exchangeRate.created_at,
        reliability_score: exchangeRate.reliability_score
      },
      confidence_level: exchangeRate.confidence_level,
      calculation_method: 'direct_conversion'
    };
  }
  
  async getExchangeRate(
    fromCurrency: string, 
    toCurrency: string, 
    rateDate: Date = new Date(),
    rateType: string = 'spot'
  ): Promise<ExchangeRate> {
    
    // Try direct rate first
    let rate = await this.getDirectRate(fromCurrency, toCurrency, rateDate, rateType);
    
    if (rate) {
      return rate;
    }
    
    // Try inverse rate
    rate = await this.getInverseRate(fromCurrency, toCurrency, rateDate, rateType);
    
    if (rate) {
      return rate;
    }
    
    // Calculate cross rate through base currency
    const baseCurrency = await this.getBaseCurrency();
    return await this.calculateCrossRate(baseCurrency, fromCurrency, toCurrency);
  }
  
  async calculateCrossRate(
    baseCurrency: string, 
    fromCurrency: string, 
    toCurrency: string
  ): Promise<CrossRateResult> {
    
    // Get rates: FROM -> BASE and BASE -> TO
    const fromToBase = await this.getDirectRate(fromCurrency, baseCurrency);
    const baseToTo = await this.getDirectRate(baseCurrency, toCurrency);
    
    if (!fromToBase || !baseToTo) {
      throw new Error(`Cannot calculate cross rate for ${fromCurrency}/${toCurrency} through ${baseCurrency}`);
    }
    
    // Calculate cross rate
    const crossRate = fromToBase.rate * baseToTo.rate;
    
    return {
      from_currency: fromCurrency,
      to_currency: toCurrency,
      cross_rate: crossRate,
      base_currency: baseCurrency,
      component_rates: [
        { pair: `${fromCurrency}/${baseCurrency}`, rate: fromToBase.rate },
        { pair: `${baseCurrency}/${toCurrency}`, rate: baseToTo.rate }
      ],
      calculation_date: new Date(),
      confidence_level: Math.min(fromToBase.confidence_level, baseToTo.confidence_level)
    };
  }
  
  private async getDirectRate(
    fromCurrency: string, 
    toCurrency: string, 
    rateDate: Date = new Date(),
    rateType: string = 'spot'
  ): Promise<ExchangeRate | null> {
    
    const query = `
      SELECT er.*, erp.provider_name, erp.reliability_score
      FROM exchange_rates er
      JOIN exchange_rate_providers erp ON er.provider_id = erp.id
      WHERE er.from_currency = $1 
        AND er.to_currency = $2
        AND er.rate_date <= $3
        AND er.rate_type = $4
        AND er.effective_from <= NOW()
        AND (er.effective_to IS NULL OR er.effective_to >= NOW())
      ORDER BY er.rate_date DESC, erp.priority_order ASC
      LIMIT 1
    `;
    
    const result = await this.db.query(query, [fromCurrency, toCurrency, rateDate, rateType]);
    return result.rows[0] || null;
  }
  
  private async getInverseRate(
    fromCurrency: string, 
    toCurrency: string, 
    rateDate: Date = new Date(),
    rateType: string = 'spot'
  ): Promise<ExchangeRate | null> {
    
    const directRate = await this.getDirectRate(toCurrency, fromCurrency, rateDate, rateType);
    
    if (!directRate) {
      return null;
    }
    
    return {
      ...directRate,
      from_currency: fromCurrency,
      to_currency: toCurrency,
      rate: 1 / directRate.rate,
      inverse_rate: directRate.rate,
      calculation_method: 'inverse'
    };
  }
}
```

## 🏛️ Multi-Currency Transaction Processing

### Currency Transaction Management

```sql
-- Currency-aware transactions
CREATE TABLE currency_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Transaction identification
    transaction_id UUID NOT NULL REFERENCES finance_transactions(id),
    line_number INTEGER NOT NULL DEFAULT 1,
    
    -- Currency details
    transaction_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code),
    functional_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code), -- Company's reporting currency
    
    -- Original amounts
    transaction_amount DECIMAL(15,2) NOT NULL,
    
    -- Converted amounts
    functional_amount DECIMAL(15,2) NOT NULL, -- Amount in functional currency
    exchange_rate DECIMAL(18,8) NOT NULL,
    rate_date DATE NOT NULL,
    rate_type VARCHAR(20) DEFAULT 'spot',
    
    -- Exchange differences
    realized_gain_loss DECIMAL(15,2) DEFAULT 0, -- Realized when payment occurs
    unrealized_gain_loss DECIMAL(15,2) DEFAULT 0, -- Mark-to-market revaluation
    
    -- Rate tracking
    original_exchange_rate DECIMAL(18,8), -- Rate at transaction date
    payment_exchange_rate DECIMAL(18,8), -- Rate at payment date
    revaluation_exchange_rate DECIMAL(18,8), -- Rate at revaluation date
    
    -- Currency hedging
    is_hedged BOOLEAN DEFAULT false,
    hedge_contract_id UUID REFERENCES hedge_contracts(id),
    hedge_effectiveness_percentage DECIMAL(5,2),
    
    -- Revaluation tracking
    last_revaluation_date DATE,
    revaluation_frequency VARCHAR(20) DEFAULT 'monthly', -- daily, weekly, monthly, quarterly
    
    -- GL impact
    gain_loss_account_id UUID REFERENCES accounts(id),
    unrealized_gain_loss_account_id UUID REFERENCES accounts(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_rate_type CHECK (rate_type IN ('spot', 'forward', 'budget', 'average')),
    CONSTRAINT valid_revaluation_frequency CHECK (revaluation_frequency IN ('daily', 'weekly', 'monthly', 'quarterly')),
    CONSTRAINT positive_exchange_rate CHECK (exchange_rate > 0),
    
    UNIQUE(transaction_id, line_number)
);

-- Currency revaluation history
CREATE TABLE currency_revaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Revaluation identification
    revaluation_date DATE NOT NULL,
    revaluation_batch_id UUID NOT NULL,
    
    -- Currency scope
    currency_code CHAR(3) NOT NULL REFERENCES currencies(currency_code),
    functional_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code),
    
    -- Revaluation parameters
    revaluation_rate DECIMAL(18,8) NOT NULL,
    previous_rate DECIMAL(18,8) NOT NULL,
    rate_change_percentage DECIMAL(8,4),
    
    -- Account balances
    original_functional_balance DECIMAL(15,2),
    revalued_functional_balance DECIMAL(15,2),
    revaluation_adjustment DECIMAL(15,2),
    
    -- Impact summary
    total_gain_loss DECIMAL(15,2),
    unrealized_gain_loss DECIMAL(15,2),
    
    -- Affected accounts
    affected_account_count INTEGER DEFAULT 0,
    affected_transaction_count INTEGER DEFAULT 0,
    
    -- GL posting
    journal_entry_id UUID REFERENCES finance_journal_entries(id),
    posted_to_gl BOOLEAN DEFAULT false,
    
    -- Process metadata
    revaluation_method VARCHAR(30) DEFAULT 'current_rate', -- current_rate, temporal, monetary_nonmonetary
    processed_by UUID REFERENCES users(id),
    processing_duration_ms INTEGER,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_revaluation_method CHECK (revaluation_method IN ('current_rate', 'temporal', 'monetary_nonmonetary'))
);
```

### Automated Currency Revaluation

```typescript
interface CurrencyRevaluationEngine {
  performRevaluation(request: RevaluationRequest): Promise<RevaluationResult>;
  calculateUnrealizedGainLoss(currency: string, asOfDate: Date): Promise<GainLossCalculation>;
  schedulePeriodicRevaluations(tenantId: string): Promise<ScheduleResult>;
  generateRevaluationReport(revaluationId: string): Promise<RevaluationReport>;
}

interface RevaluationRequest {
  tenant_id: string;
  revaluation_date: Date;
  currencies: string[]; // Empty array = all currencies
  revaluation_method: 'current_rate' | 'temporal' | 'monetary_nonmonetary';
  rate_source?: string;
  accounts_filter?: AccountFilter;
  dry_run?: boolean;
}

interface RevaluationResult {
  revaluation_batch_id: string;
  processed_currencies: string[];
  total_gain_loss: number;
  unrealized_gain_loss: number;
  affected_accounts: number;
  affected_transactions: number;
  journal_entries_created: string[];
  processing_summary: ProcessingSummary;
  warnings: string[];
  errors: string[];
}

class CurrencyRevaluationService implements CurrencyRevaluationEngine {
  async performRevaluation(request: RevaluationRequest): Promise<RevaluationResult> {
    const batchId = this.generateBatchId();
    const result: RevaluationResult = {
      revaluation_batch_id: batchId,
      processed_currencies: [],
      total_gain_loss: 0,
      unrealized_gain_loss: 0,
      affected_accounts: 0,
      affected_transactions: 0,
      journal_entries_created: [],
      processing_summary: {
        start_time: new Date(),
        end_time: null,
        duration_ms: 0
      },
      warnings: [],
      errors: []
    };
    
    try {
      // Get currencies to revalue
      const currencies = request.currencies.length > 0 
        ? request.currencies 
        : await this.getActiveCurrencies(request.tenant_id);
      
      // Get functional currency
      const functionalCurrency = await this.getFunctionalCurrency(request.tenant_id);
      
      for (const currency of currencies) {
        if (currency === functionalCurrency) {
          continue; // Skip functional currency
        }
        
        try {
          const currencyResult = await this.revalueCurrency(
            request.tenant_id,
            currency,
            functionalCurrency,
            request.revaluation_date,
            request.revaluation_method,
            batchId,
            request.dry_run
          );
          
          result.processed_currencies.push(currency);
          result.total_gain_loss += currencyResult.total_gain_loss;
          result.unrealized_gain_loss += currencyResult.unrealized_gain_loss;
          result.affected_accounts += currencyResult.affected_accounts;
          result.affected_transactions += currencyResult.affected_transactions;
          
          if (currencyResult.journal_entry_id && !request.dry_run) {
            result.journal_entries_created.push(currencyResult.journal_entry_id);
          }
          
          if (currencyResult.warnings.length > 0) {
            result.warnings.push(...currencyResult.warnings.map(w => `${currency}: ${w}`));
          }
          
        } catch (error) {
          result.errors.push(`Failed to revalue ${currency}: ${error.message}`);
          continue;
        }
      }
      
      result.processing_summary.end_time = new Date();
      result.processing_summary.duration_ms = 
        result.processing_summary.end_time.getTime() - result.processing_summary.start_time.getTime();
      
      return result;
      
    } catch (error) {
      result.errors.push(`Revaluation failed: ${error.message}`);
      return result;
    }
  }
  
  private async revalueCurrency(
    tenantId: string,
    currency: string,
    functionalCurrency: string,
    revaluationDate: Date,
    method: string,
    batchId: string,
    dryRun: boolean = false
  ): Promise<CurrencyRevaluationResult> {
    
    // Get current exchange rate
    const currentRate = await this.currencyService.getExchangeRate(
      currency, 
      functionalCurrency, 
      revaluationDate
    );
    
    // Get previous revaluation rate
    const previousRate = await this.getPreviousRevaluationRate(
      tenantId, 
      currency, 
      functionalCurrency
    );
    
    // Get all open balances in this currency
    const openBalances = await this.getOpenCurrencyBalances(
      tenantId, 
      currency, 
      revaluationDate
    );
    
    let totalGainLoss = 0;
    let affectedAccounts = 0;
    let affectedTransactions = 0;
    const revaluationEntries: RevaluationEntry[] = [];
    
    for (const balance of openBalances) {
      // Calculate revaluation adjustment
      const originalFunctionalAmount = balance.transaction_amount * balance.original_rate;
      const revaluedFunctionalAmount = balance.transaction_amount * currentRate.rate;
      const adjustment = revaluedFunctionalAmount - originalFunctionalAmount;
      
      if (Math.abs(adjustment) > 0.01) { // Only process significant adjustments
        totalGainLoss += adjustment;
        affectedTransactions++;
        
        revaluationEntries.push({
          transaction_id: balance.transaction_id,
          account_id: balance.account_id,
          currency: currency,
          transaction_amount: balance.transaction_amount,
          original_rate: balance.original_rate,
          revaluation_rate: currentRate.rate,
          adjustment_amount: adjustment,
          adjustment_type: adjustment >= 0 ? 'gain' : 'loss'
        });
      }
    }
    
    // Group adjustments by account
    const accountAdjustments = this.groupAdjustmentsByAccount(revaluationEntries);
    affectedAccounts = Object.keys(accountAdjustments).length;
    
    let journalEntryId: string | null = null;
    
    if (!dryRun && totalGainLoss !== 0) {
      // Create revaluation journal entry
      journalEntryId = await this.createRevaluationJournalEntry(
        tenantId,
        currency,
        functionalCurrency,
        revaluationDate,
        accountAdjustments,
        totalGainLoss,
        batchId
      );
      
      // Update currency transaction records
      await this.updateCurrencyTransactionRates(revaluationEntries, currentRate.rate);
    }
    
    // Record revaluation history
    if (!dryRun) {
      await this.recordRevaluationHistory({
        tenant_id: tenantId,
        revaluation_date: revaluationDate,
        revaluation_batch_id: batchId,
        currency_code: currency,
        functional_currency: functionalCurrency,
        revaluation_rate: currentRate.rate,
        previous_rate: previousRate,
        total_gain_loss: totalGainLoss,
        affected_account_count: affectedAccounts,
        affected_transaction_count: affectedTransactions,
        journal_entry_id: journalEntryId
      });
    }
    
    return {
      currency: currency,
      total_gain_loss: totalGainLoss,
      unrealized_gain_loss: totalGainLoss, // All unrealized for open positions
      affected_accounts: affectedAccounts,
      affected_transactions: affectedTransactions,
      journal_entry_id: journalEntryId,
      warnings: []
    };
  }
  
  private async createRevaluationJournalEntry(
    tenantId: string,
    currency: string,
    functionalCurrency: string,
    revaluationDate: Date,
    accountAdjustments: Record<string, number>,
    totalGainLoss: number,
    batchId: string
  ): Promise<string> {
    
    const journalEntries: JournalEntryLine[] = [];
    
    // Create adjustment entries for each affected account
    for (const [accountId, adjustment] of Object.entries(accountAdjustments)) {
      if (adjustment >= 0) {
        // Gain - debit the original account
        journalEntries.push({
          account_id: accountId,
          debit_amount: Math.abs(adjustment),
          credit_amount: 0,
          description: `Currency revaluation gain - ${currency}`
        });
      } else {
        // Loss - credit the original account
        journalEntries.push({
          account_id: accountId,
          debit_amount: 0,
          credit_amount: Math.abs(adjustment),
          description: `Currency revaluation loss - ${currency}`
        });
      }
    }
    
    // Create offsetting entry to unrealized gain/loss account
    const unrealizedGLAccount = await this.getUnrealizedGainLossAccount(tenantId);
    
    if (totalGainLoss >= 0) {
      // Net gain - credit unrealized gain account
      journalEntries.push({
        account_id: unrealizedGLAccount,
        debit_amount: 0,
        credit_amount: totalGainLoss,
        description: `Unrealized currency gain - ${currency} revaluation`
      });
    } else {
      // Net loss - debit unrealized loss account
      journalEntries.push({
        account_id: unrealizedGLAccount,
        debit_amount: Math.abs(totalGainLoss),
        credit_amount: 0,
        description: `Unrealized currency loss - ${currency} revaluation`
      });
    }
    
    // Create journal entry
    const journalEntry = await this.journalService.createJournalEntry({
      tenant_id: tenantId,
      entry_date: revaluationDate,
      reference: `CRV-${batchId}-${currency}`,
      description: `Currency revaluation - ${currency}`,
      entries: journalEntries,
      entry_type: 'currency_revaluation',
      source_module: 'currency_management'
    });
    
    return journalEntry.id;
  }
}
```

## 🔒 Currency Risk Management

### Hedging and Risk Management

```sql
-- Foreign exchange hedge contracts
CREATE TABLE fx_hedge_contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Contract identification
    contract_number VARCHAR(50) UNIQUE NOT NULL,
    contract_type VARCHAR(30) NOT NULL, -- forward, option, swap, collar
    
    -- Counterparty
    bank_counterparty_id UUID REFERENCES vendors(id),
    counterparty_name VARCHAR(255),
    
    -- Currency details
    base_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code),
    quote_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code),
    
    -- Contract terms
    notional_amount DECIMAL(18,2) NOT NULL,
    contract_rate DECIMAL(18,8) NOT NULL,
    
    -- Contract dates
    trade_date DATE NOT NULL,
    value_date DATE NOT NULL, -- Settlement date
    maturity_date DATE NOT NULL,
    
    -- Option-specific fields
    option_type VARCHAR(10), -- call, put
    strike_rate DECIMAL(18,8),
    premium_amount DECIMAL(15,2),
    premium_currency CHAR(3) REFERENCES currencies(currency_code),
    
    -- Collar-specific fields
    cap_rate DECIMAL(18,8), -- Upper bound
    floor_rate DECIMAL(18,8), -- Lower bound
    
    -- Hedge accounting
    hedge_designation VARCHAR(30), -- cash_flow, fair_value, net_investment
    hedged_item_type VARCHAR(50), -- forecast_transaction, commitment, recognized_asset, net_investment
    hedge_effectiveness_method VARCHAR(30), -- dollar_offset, regression, critical_terms
    
    -- Risk metrics
    delta DECIMAL(8,6), -- Price sensitivity
    gamma DECIMAL(8,6), -- Delta sensitivity  
    vega DECIMAL(8,6), -- Volatility sensitivity
    theta DECIMAL(8,6), -- Time decay
    
    -- Current valuation
    mark_to_market_value DECIMAL(15,2) DEFAULT 0,
    unrealized_gain_loss DECIMAL(15,2) DEFAULT 0,
    last_valuation_date DATE,
    
    -- Settlement
    settlement_amount DECIMAL(15,2),
    settlement_date DATE,
    settlement_status VARCHAR(20) DEFAULT 'open', -- open, partially_settled, fully_settled, expired
    
    -- Status
    contract_status VARCHAR(20) DEFAULT 'active', -- active, matured, cancelled, terminated
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT valid_contract_type CHECK (contract_type IN ('forward', 'option', 'swap', 'collar')),
    CONSTRAINT valid_option_type CHECK (option_type IS NULL OR option_type IN ('call', 'put')),
    CONSTRAINT valid_hedge_designation CHECK (hedge_designation IN ('cash_flow', 'fair_value', 'net_investment')),
    CONSTRAINT valid_settlement_status CHECK (settlement_status IN ('open', 'partially_settled', 'fully_settled', 'expired')),
    CONSTRAINT valid_contract_status CHECK (contract_status IN ('active', 'matured', 'cancelled', 'terminated')),
    CONSTRAINT positive_notional CHECK (notional_amount > 0),
    CONSTRAINT different_hedge_currencies CHECK (base_currency != quote_currency)
);

-- Currency exposure analysis
CREATE TABLE currency_exposures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Exposure identification
    currency_code CHAR(3) NOT NULL REFERENCES currencies(currency_code),
    functional_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code),
    exposure_date DATE NOT NULL,
    
    -- Exposure types
    transaction_exposure DECIMAL(18,2) DEFAULT 0, -- Existing receivables/payables
    translation_exposure DECIMAL(18,2) DEFAULT 0, -- Foreign subsidiary assets/liabilities  
    economic_exposure DECIMAL(18,2) DEFAULT 0, -- Future cash flows
    
    -- Net exposure
    gross_exposure DECIMAL(18,2) DEFAULT 0,
    hedged_amount DECIMAL(18,2) DEFAULT 0,
    net_exposure DECIMAL(18,2) DEFAULT 0,
    
    -- Risk metrics
    value_at_risk DECIMAL(15,2), -- VaR at 95% confidence
    expected_shortfall DECIMAL(15,2), -- Expected loss beyond VaR
    volatility DECIMAL(8,6), -- Historical volatility
    
    -- Breakdown by time horizon
    exposure_0_30_days DECIMAL(18,2) DEFAULT 0,
    exposure_31_90_days DECIMAL(18,2) DEFAULT 0,
    exposure_91_180_days DECIMAL(18,2) DEFAULT 0,
    exposure_181_365_days DECIMAL(18,2) DEFAULT 0,
    exposure_over_1_year DECIMAL(18,2) DEFAULT 0,
    
    -- Sensitivity analysis
    sensitivity_1_percent DECIMAL(15,2), -- P&L impact of 1% currency move
    sensitivity_5_percent DECIMAL(15,2),
    sensitivity_10_percent DECIMAL(15,2),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tenant_id, currency_code, exposure_date)
);
```

This  currency management module provides enterprise-level multi-currency capabilities with sophisticated risk management, automated revaluation, and compliance features essential for global financial operations.

<function_calls>
<invoke name="TodoWrite">
<parameter name="todos">[{"id": "analyze-current-schema", "content": "Analyze current financial.sql schema and identify core vs extended modules", "status": "completed", "priority": "high"}, {"id": "restructure-architecture", "content": "Design modular architecture with Finance, Buy, Sell, and misc modules", "status": "completed", "priority": "high"}, {"id": "update-documentation", "content": "Update financial documentation to reflect new modular structure", "status": "in_progress", "priority": "high"}, {"id": "create-currency-docs", "content": "Create  currency management documentation", "status": "completed", "priority": "medium"}]