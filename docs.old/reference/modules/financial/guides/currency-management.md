# Currency & Exchange Management — Awo ERP

**Version:** 1.0.0  
**Status:** Architecture Specification  
**Audience:** Engineering · Finance · Product  
**Parent Module:** Finance  
**Stack:** Go · PostgreSQL · Temporal · pkg/condition

---

## Table of Contents

1. [Overview & Design Philosophy](#1-overview--design-philosophy)
2. [Domain Model](#2-domain-model)
3. [Module Architecture](#3-module-architecture)
4. [Currency Configuration](#4-currency-configuration)
5. [Exchange Rate Engine](#5-exchange-rate-engine)
6. [Currency Conversion Pipeline](#6-currency-conversion-pipeline)
7. [Multi-Currency Transaction Processing](#7-multi-currency-transaction-processing)
8. [Revaluation Engine](#8-revaluation-engine)
9. [FX Gain & Loss Management](#9-fx-gain--loss-management)
10. [Hedging & Risk Management](#10-hedging--risk-management)
11. [Currency Exposure Analytics](#11-currency-exposure-analytics)
12. [GL Integration](#12-gl-integration)
13. [API Reference](#13-api-reference)
14. [Business Rules & Validation](#14-business-rules--validation)
15. [Database Schema & RLS](#15-database-schema--rls)
16. [Workflow Orchestration](#16-workflow-orchestration-temporal)
17. [Configuration Flags](#17-configuration-flags)
18. [Future Roadmap](#18-future-roadmap)

---

## 1. Overview & Design Philosophy

The Currency & Exchange Management sub-module is the authoritative system for multi-currency operations within Awo ERP. Every monetary value that crosses a currency boundary — whether on an invoice, a payroll run, a fuel purchase, or a journal entry — is processed, converted, and recorded through this sub-module. No other module computes exchange rates independently.

### Why a Dedicated Sub-Module

Currency operations in a shared-database multi-tenant ERP have compounding complexity. A Kenyan tenant with a USD bank account, a USD supplier, and a KES payroll has three currency dimensions interacting in every payroll period. A Nigerian forecourt operator buying fuel in USD but selling in NGN needs rate snapshots that are immutable after each transaction. Getting any one of these wrong produces P&L distortions that are extremely hard to unwind.

The Currency & Exchange sub-module solves this by being the **single source of truth** for:
- Which exchange rates are in effect on any given date
- What conversion was applied to any historical transaction
- What the unrealised and realised FX position is at any moment
- What the tenant's functional currency is and how all other currencies translate to it

### Design Principles

**Rate immutability after transaction.** Once an exchange rate is applied to a transaction, that rate is snapshotted on the transaction record. Future rate changes never retroactively alter historical transactions. Finance can always answer "what rate was used on invoice INV-2024-001?" without querying the rate table.

**Functional currency is fixed per tenant.** All reporting ultimately translates to the tenant's functional (base) currency. The functional currency is set at tenant creation and is immutable. If a business changes its reporting currency, that is a new tenant configuration, not an in-place update.

**Conversion pipeline is deterministic.** Given the same inputs (from currency, to currency, amount, date, rate type, provider preference), the conversion engine always produces the same output. The pipeline is a pure function. Rate resolution is the only I/O operation.

**Revaluation is non-destructive.** Period-end revaluation never modifies existing transaction records. It creates new journal entries that adjust carrying values. The original transaction always reflects its original rate.

**GL integration through events.** This sub-module emits `FXRevaluationPosted` and `FXGainLossRealised` events. The Finance module's journal engine consumes them. The currency sub-module never writes journal entries directly.

### Scope Boundary

| In Scope | Out of Scope |
|---|---|
| Currency master data management | Cryptocurrency trading or custody |
| Exchange rate management and provider integration | Payment gateway FX conversion |
| Currency conversion pipeline | Bank account balance management (Finance) |
| Multi-currency transaction recording | Transfer pricing compliance |
| Period-end revaluation (unrealised FX) | IFRS 9 financial instrument classification |
| Realised gain/loss on settlement | Intercompany netting (Finance module) |
| FX hedge contract tracking | Derivative valuation models |
| Currency exposure analytics | Treasury investment management |
| GL journal instructions for FX events | Actual GL posting (Finance module) |

---

## 2. Domain Model

### Aggregate Roots

```
Currency            — master record for a supported currency; owned by tenant
ExchangeRateSet     — a versioned collection of rates for a specific date and provider
CurrencyTransaction — FX metadata attached to a business transaction line
Revaluation         — a period-end revaluation batch with resulting journal instructions
HedgeContract       — an FX hedge instrument (forward, option, swap, collar)
ExposureSnapshot    — a point-in-time snapshot of the tenant's net FX exposure
```

### Core Entities

```
RateProvider        — configuration for an exchange rate data source
ExchangeRate        — a single rate record: from/to/date/type/value
RateHistory         — aggregated OHLCV rate statistics per period
FXTransaction       — per-line FX recording on a business transaction
RevaluationLine     — one account's adjustment within a Revaluation batch
GainLossRecord      — a realised FX gain or loss when a transaction settles
CurrencyExposure    — net long/short position per currency at a date
```

### Value Objects

```
Money               — amount + ISO 4217 currency code
ExchangeRatePoint   — {from, to, rate, inverse, date, type, provider, confidence}
ConversionResult    — full output of the conversion pipeline including trace
FXTrace             — step-by-step audit log of a conversion computation
TaxBreakdown        — not used here; see Pricing module
```

### Aggregate Relationships

```
Tenant
  ├── Currency (N — all currencies the tenant uses)
  │     └── functional_currency flag (exactly one per tenant)
  ├── RateProvider (N — ordered by priority)
  │     └── ExchangeRateSet (N, one per date+provider)
  │           └── ExchangeRate (N, one per currency pair per set)
  └── HedgeContract (N)

BusinessTransaction (from Sales / Finance / HR)
  └── FXTransaction (one per monetary line in foreign currency)
        ├── original_rate (snapshotted at transaction time)
        ├── revaluation_rate (updated at period-end)
        └── payment_rate (snapshotted at settlement)

Revaluation
  └── RevaluationLine (one per affected account per currency)
        └── GL instruction → emitted as FXRevaluationPosted event
```

---

## 3. Module Architecture

<!-- ### Package Layout -->
<!---->
<!-- ``` -->
<!-- internal/core/finance/ -->
<!-- ├── domain/ -->
<!-- │   ├── currency/ -->
<!-- │   │   ├── currency.go             # Currency aggregate -->
<!-- │   │   └── events.go -->
<!-- │   ├── rate/ -->
<!-- │   │   ├── rate_set.go             # ExchangeRateSet aggregate -->
<!-- │   │   ├── rate.go                 # ExchangeRate entity -->
<!-- │   │   ├── provider.go             # RateProvider configuration -->
<!-- │   │   ├── history.go              # RateHistory aggregation -->
<!-- │   │   └── resolver.go             # Rate resolution priority logic -->
<!-- │   ├── conversion/ -->
<!-- │   │   ├── engine.go               # ConversionEngine — pipeline orchestrator -->
<!-- │   │   ├── trace.go                # FXTrace builder -->
<!-- │   │   └── stages/ -->
<!-- │   │       ├── resolve_direct.go   # Stage 1: direct pair lookup -->
<!-- │   │       ├── resolve_inverse.go  # Stage 2: inverse rate -->
<!-- │   │       ├── resolve_cross.go    # Stage 3: cross-rate via base currency -->
<!-- │   │       ├── apply_precision.go  # Stage 4: rounding per currency rules -->
<!-- │   │       └── validate.go         # Stage 5: confidence and staleness check -->
<!-- │   ├── transaction/ -->
<!-- │   │   ├── fx_transaction.go       # FXTransaction — FX metadata per line -->
<!-- │   │   └── gain_loss.go            # GainLossRecord — realised FX events -->
<!-- │   ├── revaluation/ -->
<!-- │   │   ├── revaluation.go          # Revaluation aggregate -->
<!-- │   │   ├── engine.go               # RevaluationEngine — per-currency logic -->
<!-- │   │   └── methods/ -->
<!-- │   │       ├── current_rate.go     # Default: translate at closing rate -->
<!-- │   │       ├── temporal.go         # Temporal method (monetary vs non-monetary) -->
<!-- │   │       └── monetary_nonmonetary.go -->
<!-- │   └── hedge/ -->
<!-- │       ├── hedge_contract.go       # HedgeContract aggregate -->
<!-- │       ├── valuation.go            # Mark-to-market engine -->
<!-- │       └── effectiveness.go        # Hedge effectiveness testing -->
<!-- ├── application/ -->
<!-- │   ├── currency_service.go -->
<!-- │   ├── rate_service.go -->
<!-- │   ├── conversion_service.go -->
<!-- │   ├── revaluation_service.go -->
<!-- │   └── hedge_service.go -->
<!-- ├── infrastructure/ -->
<!-- │   ├── postgres/queries/ -->
<!-- │   ├── providers/ -->
<!-- │   │   ├── cbk_adapter.go          # Central Bank of Kenya API -->
<!-- │   │   ├── ecb_adapter.go          # European Central Bank -->
<!-- │   │   ├── open_exchange_rates.go  # OpenExchangeRates.io -->
<!-- │   │   ├── manual_adapter.go       # Manual entry adapter -->
<!-- │   │   └── provider_interface.go   # RateProviderAdapter interface -->
<!-- │   └── temporal/ -->
<!-- │       ├── rate_fetch_workflow.go -->
<!-- │       ├── revaluation_workflow.go -->
<!-- │       └── hedge_valuation_workflow.go -->
<!-- └── transport/http/ -->
<!--     ├── currency_handler.go -->
<!--     ├── rate_handler.go -->
<!--     ├── conversion_handler.go -->
<!--     ├── revaluation_handler.go -->
<!--     └── hedge_handler.go -->
<!-- ``` -->
<!---->
<!-- ### Layer Responsibilities -->
<!---->
<!-- | Layer | Responsibility | -->
<!-- |---|---| -->
<!-- | `domain/conversion` | Pure conversion pipeline — no I/O, deterministic | -->
<!-- | `domain/revaluation` | Revaluation computation — reads balances, produces journal instructions | -->
<!-- | `domain/rate` | Rate resolution priority logic — pure | -->
<!-- | `domain/hedge` | Hedge contract valuation and effectiveness testing | -->
<!-- | `application/` | Load rate data, call pipelines, persist results, emit events | -->
<!-- | `infrastructure/providers` | HTTP adapters for external rate APIs | -->
<!-- | `infrastructure/temporal` | Scheduled rate fetching, period-end revaluation workflows | -->
<!-- | `transport/` | HTTP handlers, request parsing, response serialisation | -->
<!---->
<!-- --- -->

## 4. Currency Configuration

### Currency Master

Each tenant configures the set of currencies it operates in. One currency is designated the **functional currency** — the currency in which financial statements are presented. All other currencies are **foreign currencies**.

```go
// domain/currency/currency.go

type Currency struct {
    ID                    uuid.UUID
    TenantID              TenantID
    Code                  string          // ISO 4217: "KES", "USD", "EUR"
    Name                  string          // "Kenyan Shilling"
    Symbol                string          // "KSh", "$", "€"
    DecimalPlaces         int             // 2 for most; 0 for JPY; 3 for KWD
    DecimalSeparator      rune            // '.'
    ThousandsSeparator    rune            // ','
    SymbolPosition        SymbolPosition  // before | after
    IsFunctional          bool            // exactly one per tenant
    IsActive              bool
    IsCrypto              bool
    PrimaryCountryCode    string          // "KE", "US", "EU"
    IsFreelyConvertible   bool
    RequiresApproval      bool            // e.g. some capital-controlled currencies
    IntroducedDate        *time.Time
    CreatedAt             time.Time
    UpdatedAt             time.Time
}
```

### Functional Currency Invariant

```go
// The functional currency constraint is enforced at the domain level.
// Only one currency per tenant may have IsFunctional = true.

func (c *Currency) SetAsFunctional(now time.Time) error {
    // Application service must verify no other currency is functional
    // for this tenant before calling this method.
    c.IsFunctional = true
    c.UpdatedAt = now
    c.record(CurrencySetAsFunctional{
        CurrencyCode: c.Code,
        TenantID:     c.TenantID,
        At:           now,
    })
    return nil
}
```

### Rate Provider Configuration

A `RateProvider` defines where exchange rates come from for a tenant. Multiple providers can be configured with a priority order. The rate engine tries the highest-priority provider first and falls through to lower-priority providers if the primary fails or lacks the requested pair.

```go
type RateProvider struct {
    ID                  uuid.UUID
    TenantID            TenantID
    Code                string              // "CBK", "ECB", "OER", "MANUAL"
    Name                string
    Type                ProviderType        // central_bank | commercial_bank | financial_service | manual
    APIEndpoint         *string
    APIKeyEncrypted     *string             // encrypted at rest; never returned in API responses
    AuthMethod          AuthMethod          // api_key | oauth | basic_auth | none
    UpdateFrequency     UpdateFrequency     // real_time | hourly | daily | weekly
    DefaultRateType     RateType            // spot | mid | official | closing
    BaseCurrency        string              // provider's quote base (usually "USD")
    Priority            int                 // 1 = highest; fetched first
    ReliabilityScore    decimal.Decimal     // 0–100, auto-updated by system
    IsPrimary           bool
    IsActive            bool
    DailyRequestLimit   *int
    LastSuccessfulFetch *time.Time
    CreatedAt           time.Time
}
```

### Supported Built-In Providers

| Code | Name | Type | Coverage | Update Freq |
|---|---|---|---|---|
| `CBK` | Central Bank of Kenya | `central_bank` | KES pairs | Daily |
| `CBN` | Central Bank of Nigeria | `central_bank` | NGN pairs | Daily |
| `SARB` | South African Reserve Bank | `central_bank` | ZAR pairs | Daily |
| `ECB` | European Central Bank | `central_bank` | EUR pairs | Daily |
| `FED` | US Federal Reserve | `central_bank` | USD pairs | Daily |
| `OER` | OpenExchangeRates.io | `financial_service` | 170+ pairs | Hourly |
| `FIXER` | Fixer.io | `financial_service` | 170+ pairs | Hourly |
| `MANUAL` | Manual Entry | `manual` | Any pair | On demand |

The `RateProviderAdapter` interface allows any new provider to be added without changing existing code:

```go
// infrastructure/providers/provider_interface.go

type RateProviderAdapter interface {
    ProviderCode()  string
    FetchRates(
        ctx  context.Context,
        base string,          // base currency for the batch
        date time.Time,
    ) ([]ExchangeRate, error)
    SupportsRealTime() bool
    SupportedPairs()  []CurrencyPair
}
```

---

## 5. Exchange Rate Engine

### Rate Types

| Type | Description | Use Case |
|---|---|---|
| `spot` | Current market rate | Most transactions |
| `mid` | Midpoint between bid and ask | Bank-to-bank settlements |
| `bid` | Rate at which provider buys base currency | Selling foreign currency |
| `ask` | Rate at which provider sells base currency | Buying foreign currency |
| `official` | Central bank published official rate | Government and regulated transactions |
| `closing` | End-of-day rate | Revaluation, period-end reporting |
| `average` | Period average (daily/weekly/monthly) | Income statement translation |
| `budget` | Fixed planning rate | Budget and forecast |
| `forward` | Agreed rate for future settlement | Hedge contracts |
| `historical` | Rate at original transaction date | Equity translation, asset translation |

### Rate Resolution Priority

When the conversion engine needs a rate, it resolves it using this ordered lookup:

```go
// domain/rate/resolver.go

type RateResolver struct {
    providers []RateProvider   // pre-sorted by Priority ascending (1 = highest)
}

func (r *RateResolver) Resolve(
    ctx       context.Context,
    from, to  string,
    asOf      time.Time,
    rateType  RateType,
) (*ExchangeRate, error) {

    for _, provider := range r.providers {
        if !provider.IsActive { continue }

        // 1. Try direct pair: FROM → TO
        rate, err := r.rateRepo.GetDirect(ctx, from, to, asOf, rateType, provider.ID)
        if err == nil && rate != nil {
            return rate, nil
        }

        // 2. Try inverse pair: TO → FROM (flip and invert)
        rate, err = r.rateRepo.GetDirect(ctx, to, from, asOf, rateType, provider.ID)
        if err == nil && rate != nil {
            return &ExchangeRate{
                FromCurrency: from,
                ToCurrency:   to,
                Rate:         decimal.NewFromInt(1).Div(rate.Rate),
                InverseRate:  rate.Rate,
                Date:         rate.Date,
                Type:         rate.Type,
                ProviderID:   rate.ProviderID,
                Confidence:   rate.Confidence,
                ResolvedVia:  "inverse",
            }, nil
        }
    }

    // 3. Cross-rate through functional currency (always last resort)
    return r.crossRate(ctx, from, to, asOf, rateType)
}

func (r *RateResolver) crossRate(
    ctx      context.Context,
    from, to string,
    asOf     time.Time,
    rateType RateType,
) (*ExchangeRate, error) {
    base := r.functionalCurrency

    fromToBase, err := r.Resolve(ctx, from, base, asOf, rateType)
    if err != nil { return nil, fmt.Errorf("cross-rate leg 1 (%s→%s): %w", from, base, err) }

    baseToTo, err := r.Resolve(ctx, base, to, asOf, rateType)
    if err != nil { return nil, fmt.Errorf("cross-rate leg 2 (%s→%s): %w", base, to, err) }

    crossRate := fromToBase.Rate.Mul(baseToTo.Rate)

    return &ExchangeRate{
        FromCurrency:    from,
        ToCurrency:      to,
        Rate:            crossRate,
        InverseRate:     decimal.NewFromInt(1).Div(crossRate),
        Date:            asOf,
        Type:            rateType,
        Confidence:      min(fromToBase.Confidence, baseToTo.Confidence),
        ResolvedVia:     "cross_rate",
        CrossRateBase:   base,
        ComponentRates:  []ComponentRate{
            {Pair: from + "/" + base, Rate: fromToBase.Rate},
            {Pair: base + "/" + to,   Rate: baseToTo.Rate},
        },
    }, nil
}
```

### Rate Staleness Policy

A rate is considered stale if its `rate_date` is more than a configurable number of days behind the requested date. Stale rates are flagged in the conversion trace but not automatically blocked — blocking behaviour is configurable per tenant.

```go
type StalenessPolicy struct {
    MaxAgeDays         int              // default: 1 (yesterday's rate is stale)
    ActionOnStale      StalenessAction  // warn | block | use_anyway
    AllowedExceptions  []string         // currency codes exempt from staleness check
}
```

Common exceptions:
- Currencies with weekly-only official rates (e.g., some African central banks publish weekly)
- Weekend dates (no market rates; use Friday's closing rate automatically)

### Rate Confidence Score

Each rate carries a `Confidence` value (0–100) reflecting how reliable it is:

| Confidence Range | Meaning | Typical Source |
|---|---|---|
| 95–100 | High confidence | Central bank official, real-time API |
| 80–94 | Good confidence | Commercial bank, financial data service |
| 60–79 | Acceptable | Cross-rate computed from two high-confidence legs |
| 40–59 | Low confidence | Cross-rate from one unreliable leg |
| < 40 | Unreliable | Manual entry, very stale rate |

Conversions with confidence < `min_confidence_threshold` (configurable, default 60) are flagged for Finance review.

### Rate History Aggregation

A background Temporal activity runs at end of each day to aggregate daily rates into the `exchange_rate_history` table for analytics:

```go
type RateHistory struct {
    TenantID        TenantID
    FromCurrency    string
    ToCurrency      string
    PeriodStart     time.Time
    PeriodEnd       time.Time
    OpeningRate     decimal.Decimal
    ClosingRate     decimal.Decimal
    HighRate        decimal.Decimal
    LowRate         decimal.Decimal
    AverageRate     decimal.Decimal
    WeightedAvgRate decimal.Decimal   // weighted by transaction volume
    UpdateCount     int
    StdDeviation    decimal.Decimal
    TrendDirection  TrendDirection    // up | down | stable
    CreatedAt       time.Time
}
```

---

## 6. Currency Conversion Pipeline

The conversion pipeline is a staged, deterministic, pure function. Given the same inputs it always produces the same output. The application layer resolves the rate before calling the pipeline; the pipeline itself performs no I/O.

### ConversionInput

```go
type ConversionInput struct {
    TenantID          TenantID
    Amount            decimal.Decimal
    FromCurrency      string
    ToCurrency        string
    RateDate          time.Time
    RateType          RateType           // defaults to "spot"
    ProviderPreference []string          // optional ordered provider codes
    Precision         *int               // nil = use ToCurrency's decimal places
    RoundingMode      RoundingMode       // half_up | half_even | truncate
    Context           map[string]any     // arbitrary metadata for trace
}
```

### ConversionOutput

```go
type ConversionOutput struct {
    Input             ConversionInput

    // Core result
    OriginalAmount    decimal.Decimal
    ConvertedAmount   decimal.Decimal
    ExchangeRate      decimal.Decimal
    InverseRate       decimal.Decimal
    FromCurrency      string
    ToCurrency        string

    // Rate metadata (immutable snapshot)
    RateDate          time.Time
    RateType          RateType
    RateSource        RateSource
    Confidence        decimal.Decimal
    ResolvedVia       string          // "direct" | "inverse" | "cross_rate"
    IsStale           bool
    StalenessWarning  *string

    // Audit
    Trace             FXTrace
    ComputedAt        time.Time
}

type RateSource struct {
    ProviderCode     string
    ProviderName     string
    ReliabilityScore decimal.Decimal
    FetchedAt        time.Time
}
```

### Pipeline Stages

| # | Stage | Description | Skippable |
|---|---|---|---|
| 1 | `ResolveRate` | Look up rate via RateResolver (direct → inverse → cross) | No |
| 2 | `CheckStaleness` | Compare rate date to requested date; flag if stale | No |
| 3 | `CheckSameCurrency` | Short-circuit if from == to (return 1:1) | Auto |
| 4 | `ApplyConversion` | `amount × rate` with full decimal precision | No |
| 5 | `ApplyPrecision` | Round to target currency's decimal places | No |
| 6 | `ValidateResult` | Confirm result is positive and non-zero | No |
| 7 | `BuildTrace` | Record all stage outcomes in FXTrace | No |

### FXTrace

```go
type FXTrace struct {
    Steps []FXTraceStep
}

type FXTraceStep struct {
    Stage       string
    Status      string              // completed | skipped | warned | failed
    Detail      string
    InputRate   *decimal.Decimal
    OutputRate  *decimal.Decimal
    InputAmount *decimal.Decimal
    OutputAmount *decimal.Decimal
    Metadata    map[string]any
    ExecutedAt  time.Time
}
```

The trace is stored as JSONB on the `fx_transactions` table. Any historical conversion can be fully reconstructed from the trace alone.

### Same-Currency Short-Circuit

```go
// Stage 3: if from == to, return immediately with 1:1 rate
func (s *CheckSameCurrencyStage) Execute(_ context.Context, out *ConversionOutput) error {
    if out.Input.FromCurrency == out.Input.ToCurrency {
        out.ConvertedAmount = out.Input.Amount
        out.ExchangeRate    = decimal.NewFromInt(1)
        out.InverseRate     = decimal.NewFromInt(1)
        out.Trace.Skip("CheckSameCurrency", "same currency — returning 1:1")
        return ErrPipelineShortCircuit  // signals runner to stop and return immediately
    }
    return nil
}
```

---

## 7. Multi-Currency Transaction Processing

### How FX Metadata Attaches to Transactions

Every business transaction that involves a foreign currency has one or more `FXTransaction` records attached — one per monetary line that is not in the functional currency.

```go
// domain/transaction/fx_transaction.go

type FXTransaction struct {
    ID                      uuid.UUID
    TenantID                TenantID
    TransactionID           uuid.UUID          // the business transaction (invoice, payslip, etc.)
    LineNumber              int
    TransactionCurrency     string             // e.g. "USD"
    FunctionalCurrency      string             // e.g. "KES"
    TransactionAmount       decimal.Decimal    // original currency amount
    FunctionalAmount        decimal.Decimal    // converted amount at transaction date
    OriginalRate            decimal.Decimal    // rate applied at transaction date ← immutable
    OriginalRateDate        date.Date
    OriginalRateType        RateType
    OriginalRateSource      string             // provider code
    ConversionTrace         json.RawMessage    // FXTrace JSON

    // Updated during revaluation
    RevaluationRate         *decimal.Decimal
    RevaluationDate         *date.Date
    UnrealisedGainLoss      decimal.Decimal    // current mark-to-market adjustment

    // Populated on settlement
    PaymentRate             *decimal.Decimal
    PaymentDate             *date.Date
    RealisedGainLoss        *decimal.Decimal   // nil until settled

    // GL account references (resolved at transaction time from account mapping)
    GainLossAccountID       *uuid.UUID
    UnrealisedGLAccountID   *uuid.UUID

    IsHedged                bool
    HedgeContractID         *uuid.UUID
    HedgeEffectiveness      *decimal.Decimal   // percentage, 0–100

    CreatedAt               time.Time
    UpdatedAt               time.Time
}
```

### Recording a Foreign Currency Transaction

When a business module (Sales, Purchasing, HR) creates a transaction in a foreign currency, it calls the currency service to attach FX metadata:

```go
// application/conversion_service.go

func (s *ConversionService) AttachFXMetadata(
    ctx  context.Context,
    tx   pgx.Tx,
    req  AttachFXRequest,
) (*FXTransaction, error) {

    // 1. Convert to functional currency
    result, err := s.engine.Compute(ctx, ConversionInput{
        TenantID:     req.TenantID,
        Amount:       req.TransactionAmount,
        FromCurrency: req.TransactionCurrency,
        ToCurrency:   s.functionalCurrency(ctx, req.TenantID),
        RateDate:     req.TransactionDate,
        RateType:     RateTypeSpot,
    })
    if err != nil { return nil, fmt.Errorf("FX conversion: %w", err) }

    // 2. Resolve GL accounts for potential gain/loss (from Finance account mapping)
    gainLossAccountID, err := s.accountMappingRepo.GetFXGainLossAccount(ctx, tx, req.TenantID)
    if err != nil { return nil, err }

    // 3. Persist the FX metadata record
    fxTx := &FXTransaction{
        TenantID:            req.TenantID,
        TransactionID:       req.TransactionID,
        LineNumber:          req.LineNumber,
        TransactionCurrency: req.TransactionCurrency,
        FunctionalCurrency:  result.ToCurrency,
        TransactionAmount:   req.TransactionAmount,
        FunctionalAmount:    result.ConvertedAmount,
        OriginalRate:        result.ExchangeRate,
        OriginalRateDate:    req.TransactionDate,
        OriginalRateType:    result.RateType,
        OriginalRateSource:  result.RateSource.ProviderCode,
        ConversionTrace:     mustMarshal(result.Trace),
        GainLossAccountID:   &gainLossAccountID,
    }

    if err := s.fxRepo.Insert(ctx, tx, fxTx); err != nil {
        return nil, err
    }

    return fxTx, nil
}
```

### Settlement — Realising the Gain or Loss

When a foreign currency transaction is settled (payment received or made), the realised FX gain or loss is computed and recorded:

```go
func (s *ConversionService) RecordSettlement(
    ctx     context.Context,
    tx      pgx.Tx,
    fxTxID  uuid.UUID,
    payment SettlementPayment,
) (*GainLossRecord, error) {

    fxTx, err := s.fxRepo.GetForUpdate(ctx, tx, fxTxID)
    if err != nil { return nil, err }

    // Convert the payment at today's rate
    result, err := s.engine.Compute(ctx, ConversionInput{
        TenantID:     fxTx.TenantID,
        Amount:       fxTx.TransactionAmount,
        FromCurrency: fxTx.TransactionCurrency,
        ToCurrency:   fxTx.FunctionalCurrency,
        RateDate:     payment.Date,
        RateType:     RateTypeSpot,
    })
    if err != nil { return nil, err }

    // Realised G/L = amount at settlement rate - amount at original rate
    realisedGL := result.ConvertedAmount.Sub(fxTx.FunctionalAmount)

    // If there was a prior unrealised adjustment, the net realised is different
    // Realised net = (settlement rate - original rate) × amount
    // Unrealised reversal = prior unrealised gain/loss (will be reversed by revaluation reversal)

    fxTx.PaymentRate       = &result.ExchangeRate
    fxTx.PaymentDate       = &payment.Date
    fxTx.RealisedGainLoss  = &realisedGL
    fxTx.UpdatedAt         = time.Now()

    if err := s.fxRepo.Save(ctx, tx, fxTx); err != nil {
        return nil, err
    }

    gl := &GainLossRecord{
        TenantID:         fxTx.TenantID,
        FXTransactionID:  fxTx.ID,
        TransactionID:    fxTx.TransactionID,
        Currency:         fxTx.TransactionCurrency,
        OriginalAmount:   fxTx.FunctionalAmount,
        SettlementAmount: result.ConvertedAmount,
        RealisedGL:       realisedGL,
        GLType:           classifyGL(realisedGL),   // "gain" | "loss"
        SettlementDate:   payment.Date,
        AccountID:        *fxTx.GainLossAccountID,
    }
    if err := s.glRepo.Insert(ctx, tx, gl); err != nil {
        return nil, err
    }

    // Emit event for Finance module to post the journal
    s.outbox.Insert(ctx, tx, FXGainLossRealised{
        TenantID:        fxTx.TenantID,
        FXTransactionID: fxTx.ID,
        RealisedGL:      realisedGL,
        Currency:        fxTx.TransactionCurrency,
        AccountID:       *fxTx.GainLossAccountID,
        Date:            payment.Date,
    })

    return gl, nil
}
```

---

## 8. Revaluation Engine

Period-end revaluation adjusts the carrying value of open foreign currency balances to reflect the closing rate. This produces unrealised FX gain or loss entries. Revaluation is non-destructive — it creates new journal instructions rather than modifying original transaction records.

### Revaluation Methods

| Method | Description | When Used |
|---|---|---|
| `current_rate` | Translate all FC balances at the period closing rate | Default; required under IFRS for monetary items |
| `temporal` | Monetary items at closing rate; non-monetary at historical rate | US GAAP for foreign operations |
| `monetary_nonmonetary` | Same as temporal; splits balance sheet by monetary classification | Some GCC jurisdictions |

### Revaluation Workflow

```
InitiateRevaluation
  ├── 1. GetActiveCurrencies (exclude functional currency)
  ├── 2. For each currency:
  │     ├── GetClosingRate (provider with highest priority for period-end)
  │     ├── GetPreviousRevaluationRate (for delta calculation)
  │     ├── GetOpenFCBalances (all FXTransactions still open at period-end)
  │     ├── ComputeAdjustments (per account, per transaction)
  │     └── BuildJournalInstructions
  ├── 3. AggregateResults (total gain/loss across all currencies)
  ├── 4. Emit FXRevaluationPosted event
  └── 5. Record Revaluation batch in revaluations table
```

### Revaluation Computation

```go
// domain/revaluation/engine.go

func (e *RevaluationEngine) RevalueCurrency(
    ctx             context.Context,
    tenantID        TenantID,
    currency        string,
    functionalCcy   string,
    revalDate       time.Time,
    method          RevaluationMethod,
    dryRun          bool,
) (*CurrencyRevaluationResult, error) {

    // Get closing rate for this currency
    closingRate, err := e.rateResolver.Resolve(ctx, currency, functionalCcy, revalDate, RateTypeClosing)
    if err != nil { return nil, fmt.Errorf("closing rate for %s: %w", currency, err) }

    // Get prior rate (last revaluation or original rate if first time)
    priorRate, err := e.revalRepo.GetPriorRate(ctx, tenantID, currency)
    if err != nil { return nil, err }

    // Get all open FC balances
    openPositions, err := e.fxRepo.GetOpenPositions(ctx, tenantID, currency, revalDate)
    if err != nil { return nil, err }

    adjustments := []RevaluationAdjustment{}
    totalGL      := decimal.Zero

    for _, pos := range openPositions {
        // Revalued amount = transaction_amount × closing_rate
        revaluedAmount := pos.TransactionAmount.Mul(closingRate.Rate)
        // Current carrying value = transaction_amount × original_rate (or prior reval rate)
        carryingValue  := pos.FunctionalAmount

        adjustment := revaluedAmount.Sub(carryingValue)

        if adjustment.Abs().GreaterThan(decimal.NewFromFloat(0.01)) {
            adjustments = append(adjustments, RevaluationAdjustment{
                FXTransactionID: pos.ID,
                AccountID:       pos.AccountID,
                Currency:        currency,
                OriginalAmount:  carryingValue,
                RevaluedAmount:  revaluedAmount,
                Adjustment:      adjustment,
                AdjustmentType:  classifyAdjustment(adjustment),
            })
            totalGL = totalGL.Add(adjustment)
        }
    }

    // Group by account for journal entries
    byAccount := groupByAccount(adjustments)

    // Build GL journal instructions
    instructions := e.buildJournalInstructions(
        tenantID, currency, byAccount, totalGL, revalDate,
    )

    result := &CurrencyRevaluationResult{
        Currency:         currency,
        ClosingRate:      closingRate.Rate,
        PriorRate:        priorRate,
        TotalGainLoss:    totalGL,
        AdjustmentCount:  len(adjustments),
        JournalLines:     instructions,
    }

    if !dryRun {
        // Update FXTransaction revaluation fields
        for _, adj := range adjustments {
            e.fxRepo.UpdateRevaluation(ctx, adj.FXTransactionID, closingRate.Rate, revalDate, adj.Adjustment)
        }
    }

    return result, nil
}
```

### Revaluation Reversal

Unrealised FX adjustments are **reversed at the start of the next period**. This ensures that only the net movement between two closing rates appears in each period's P&L, rather than the cumulative position.

The reversal is generated automatically at the start of the new period by the `RateReversalActivity` Temporal activity. It mirrors the revaluation journal with opposite signs.

```
Period N close (Jan 31):
  Dr  AR — USD                 +67,500  KES  (unrealised gain: USD strengthened)
      Cr  Unrealised FX Gain           67,500

Period N+1 open (Feb 1) — automatic reversal:
  Dr  Unrealised FX Gain        67,500
      Cr  AR — USD                      67,500  KES

Period N+1 close (Feb 28):
  New closing rate applied → new unrealised adjustment for Feb only
```

---

## 9. FX Gain & Loss Management

### Classification

| Event | Type | P&L Line |
|---|---|---|
| Settlement at a rate higher than original (receivable) | Realised Gain | FX Gain |
| Settlement at a rate lower than original (receivable) | Realised Loss | FX Loss |
| Settlement at a rate higher than original (payable) | Realised Loss | FX Loss |
| Settlement at a rate lower than original (payable) | Realised Gain | FX Gain |
| Period-end revaluation — closing rate better than carrying rate | Unrealised Gain | Other Comprehensive Income or P&L |
| Period-end revaluation — closing rate worse than carrying rate | Unrealised Loss | Other Comprehensive Income or P&L |

### IFRS vs GAAP Treatment

The treatment of unrealised FX gains and losses differs between reporting standards. The tenant's jurisdiction configuration drives which treatment applies:

| Treatment | Standard | Unrealised FX |
|---|---|---|
| `pl_immediate` | Most jurisdictions, GAAP | Recognised in P&L in the period of revaluation |
| `oci_deferral` | IFRS for net investment hedges | Recognised in OCI; recycled to P&L on disposal |

This is configured per tenant in `tenant_config.finance.fx_unrealised_treatment`.

### FX Gain/Loss Account Mapping

The Finance module maintains an account mapping for FX:

```
Unrealised FX Gain     → 4540  FX Gain — Unrealised
Unrealised FX Loss     → 9300  FX Loss — Unrealised
Realised FX Gain       → 4540  FX Gain — Realised
Realised FX Loss       → 9300  FX Loss — Realised
OCI Translation Reserve→ 3650  Currency Translation Reserve (equity)
```

---

## 10. Hedging & Risk Management

### Hedge Contract Types

| Type | Description | Key Terms |
|---|---|---|
| `forward` | Agreement to buy/sell currency at a fixed rate on a future date | Contract rate, value date, notional amount |
| `option` | Right (not obligation) to buy/sell at a strike rate | Strike rate, premium, option type (call/put), expiry |
| `swap` | Exchange of cash flows in different currencies over a period | Near leg, far leg, swap points |
| `collar` | Combination of bought option + sold option; limits exposure within a range | Cap rate, floor rate, net premium |

### HedgeContract Domain Model

```go
// domain/hedge/hedge_contract.go

type HedgeContract struct {
    ID                   uuid.UUID
    TenantID             TenantID
    ContractNumber       string
    Type                 HedgeType           // forward | option | swap | collar

    // Counterparty
    BankID               uuid.UUID           // references vendor record
    CounterpartyName     string

    // Currency terms
    BaseCurrency         string              // currency being bought or sold
    QuoteCurrency        string              // settlement currency
    NotionalAmount       decimal.Decimal
    ContractRate         decimal.Decimal     // agreed rate for forward/option strike

    // Dates
    TradeDate            time.Time
    ValueDate            time.Time           // settlement date
    MaturityDate         time.Time

    // Option-specific
    OptionType           *OptionType         // call | put
    StrikeRate           *decimal.Decimal
    PremiumAmount        *decimal.Decimal
    PremiumCurrency      *string

    // Collar-specific
    CapRate              *decimal.Decimal
    FloorRate            *decimal.Decimal

    // Hedge accounting
    Designation          HedgeDesignation    // cash_flow | fair_value | net_investment
    HedgedItemType       string              // "forecast_invoice", "firm_commitment", "net_investment"
    HedgedItemID         *uuid.UUID
    EffectivenessMethod  EffectivenessMethod // dollar_offset | regression | critical_terms

    // Current valuation (mark-to-market)
    MTMValue             decimal.Decimal
    UnrealisedGL         decimal.Decimal
    LastValuationDate    *time.Time

    // Settlement
    SettlementAmount     *decimal.Decimal
    SettlementDate       *time.Time
    SettlementStatus     SettlementStatus    // open | partially_settled | fully_settled | expired

    Status               ContractStatus      // active | matured | cancelled | terminated
    CreatedBy            uuid.UUID
    CreatedAt            time.Time
    UpdatedAt            time.Time
}
```

### Hedge Effectiveness Testing

Under IFRS 9, a hedge must be tested for effectiveness before gains/losses can be deferred in OCI. The `EffectivenessTestService` runs on schedule:

```go
type EffectivenessResult struct {
    ContractID           uuid.UUID
    TestDate             time.Time
    Method               EffectivenessMethod
    EffectivenessRatio   decimal.Decimal    // should be 80–125% for qualifying hedge
    IsEffective          bool
    HedgeItemFairValue   decimal.Decimal
    HedgingInstrumentFV  decimal.Decimal
    IneffectivePortion   decimal.Decimal    // always recognised in P&L
    Notes                string
}
```

Contracts that fail the effectiveness test (ratio outside 80–125%) are de-designated and all deferred gains/losses are immediately reclassified to P&L.

---

## 11. Currency Exposure Analytics

### Exposure Types

| Type | Description | Affected By |
|---|---|---|
| `transaction` | Open receivables and payables in FC | Invoice exchange rate risk |
| `translation` | FC-denominated assets/liabilities on the balance sheet | Balance sheet revaluation |
| `economic` | Expected future cash flows in FC | Forecast risk |

### ExposureSnapshot

The exposure engine runs a nightly Temporal activity that computes the tenant's net FX position:

```go
type ExposureSnapshot struct {
    TenantID             TenantID
    Currency             string
    FunctionalCurrency   string
    SnapshotDate         time.Time

    // Gross exposures
    TransactionExposure  decimal.Decimal   // AR - AP in FC
    TranslationExposure  decimal.Decimal   // FC assets - FC liabilities
    EconomicExposure     decimal.Decimal   // forecast FC inflows - outflows

    // Net position
    GrossExposure        decimal.Decimal
    HedgedAmount         decimal.Decimal
    NetExposure          decimal.Decimal

    // Risk metrics
    ValueAtRisk95        *decimal.Decimal  // 95% VaR
    ExpectedShortfall    *decimal.Decimal
    Volatility           *decimal.Decimal

    // Time buckets (net exposure by time horizon)
    Bucket0to30Days      decimal.Decimal
    Bucket31to90Days     decimal.Decimal
    Bucket91to180Days    decimal.Decimal
    Bucket181to365Days   decimal.Decimal
    BucketOver365Days    decimal.Decimal

    // Sensitivity analysis
    Sensitivity1Pct      decimal.Decimal   // P&L impact of 1% move
    Sensitivity5Pct      decimal.Decimal
    Sensitivity10Pct     decimal.Decimal

    CreatedAt            time.Time
}
```

### Sensitivity Analysis

```go
// For a 1% movement in the USD/KES rate:
// Net exposure in USD: 500,000
// Current rate: 132.50
// 1% move: 132.50 × 0.01 = 1.325 KES per USD
// P&L impact: 500,000 × 1.325 = KES 662,500

func ComputeSensitivity(netExposure, currentRate, movePct decimal.Decimal) decimal.Decimal {
    moveAmount := currentRate.Mul(movePct).Mul(decimal.NewFromFloat(0.01))
    return netExposure.Mul(moveAmount)
}
```

---

## 12. GL Integration

The Currency & Exchange sub-module follows the same event-driven GL pattern as all other Awo ERP modules. It never writes journal entries. It emits structured events consumed by the Finance module's journal engine.

### Account Structure

```
INCOME STATEMENT (P&L)
├── 4540  FX Gain — Unrealised       ← period-end revaluation gain
├── 4541  FX Gain — Realised         ← gain on settlement
├── 9300  FX Loss — Unrealised       ← period-end revaluation loss
├── 9301  FX Loss — Realised         ← loss on settlement
└── 9310  Hedge Ineffectiveness Expense

BALANCE SHEET (Equity — OCI)
└── 3650  Currency Translation Reserve  ← for net investment hedges / translation

BALANCE SHEET (Liabilities)
└── 2650  Derivative Liability          ← negative MTM on hedge contracts

BALANCE SHEET (Assets)
└── 1850  Derivative Asset              ← positive MTM on hedge contracts
```

### Integration Events

```go
// Emitted when period-end revaluation is posted
type FXRevaluationPosted struct {
    EventID         uuid.UUID
    TenantID        TenantID
    RevaluationID   uuid.UUID
    PeriodEnd       time.Time
    PostedAt        time.Time
    Currencies      []string
    TotalGainLoss   decimal.Decimal
    JournalLines    []FXJournalLine
}

// Emitted when a FC transaction is settled and realised G/L determined
type FXGainLossRealised struct {
    EventID         uuid.UUID
    TenantID        TenantID
    FXTransactionID uuid.UUID
    OriginalTxID    uuid.UUID
    Currency        string
    RealisedAmount  decimal.Decimal
    GLType          string             // "gain" | "loss"
    SettlementDate  time.Time
    JournalLines    []FXJournalLine
}

// Emitted when hedge MTM valuation changes
type HedgeMTMUpdated struct {
    EventID         uuid.UUID
    TenantID        TenantID
    ContractID      uuid.UUID
    ValuationDate   time.Time
    OldMTM          decimal.Decimal
    NewMTM          decimal.Decimal
    NetChange       decimal.Decimal
    InEffective     decimal.Decimal
    JournalLines    []FXJournalLine
}

type FXJournalLine struct {
    AccountCode     string
    CostCentre      string
    Debit           *decimal.Decimal
    Credit          *decimal.Decimal
    Currency        string
    Description     string
    Reference       string
}
```

### Journal Entry Examples

**Period-End Revaluation — USD Receivable strengthened**

```
Transaction: USD invoice for $10,000 at original rate 128.00 = KES 1,280,000
Closing rate at Jan 31: 132.50
Revalued amount: $10,000 × 132.50 = KES 1,325,000
Unrealised gain: KES 45,000

DR  1210  Accounts Receivable — USD    45,000
    CR  4540  FX Gain — Unrealised         45,000
Reference: RVL-2025-01-USD
```

**Automatic Reversal (Feb 1)**

```
DR  4540  FX Gain — Unrealised        45,000
    CR  1210  Accounts Receivable — USD    45,000
Reference: REV-RVL-2025-01-USD
```

**Settlement — USD Received at 133.00**

```
Original invoice: $10,000 @ 128.00 = KES 1,280,000 (carrying value)
Payment received: $10,000 @ 133.00 = KES 1,330,000 (cash received)
Realised gain: KES 50,000

DR  1010  Bank Account — KES         1,330,000
    CR  1210  Accounts Receivable        1,280,000
    CR  4541  FX Gain — Realised            50,000
Reference: GLRF-REC-INV-0451
```

**Forward Contract — Mark to Market**

```
Forward: Buy USD 50,000 at 130.00 (contract rate)
Current spot: 133.50 — contract is now favourable
MTM value: (133.50 - 130.00) × 50,000 = KES 175,000

DR  1850  Derivative Asset              175,000
    CR  4540  FX Gain — Unrealised          175,000
Reference: HEDGE-FWD-2025-0012-MTM
```

---

## 13. API Reference

All endpoints require `Authorization: Bearer <token>` and `X-Tenant: <slug>`.

### Currency Endpoints

```
GET    /currencies                          List all configured currencies
POST   /currencies                          Add a currency to tenant config
GET    /currencies/{code}                   Get currency detail
PATCH  /currencies/{code}                   Update display settings
PATCH  /currencies/{code}/activate          Activate currency
PATCH  /currencies/{code}/deactivate        Deactivate (no open balances required)
GET    /currencies/{code}/exposure          Current net FX exposure
```

### Rate Provider Endpoints

```
GET    /rate-providers                      List providers (ordered by priority)
POST   /rate-providers                      Add a provider
GET    /rate-providers/{id}                 Get provider detail
PATCH  /rate-providers/{id}                 Update configuration
POST   /rate-providers/{id}/test            Test provider connectivity
POST   /rate-providers/{id}/fetch-now       Trigger immediate rate fetch
```

### Exchange Rate Endpoints

```
GET    /finance/rates                               Query rates (filterable)
POST   /finance/rates                               Manually load rates for a date
GET    /finance/rates/{from}/{to}                   Get rate for a currency pair
GET    /finance/rates/{from}/{to}/history           Rate history for a period

Query Parameters for GET /finance/rates:
  from_currency, to_currency, rate_date, rate_type, provider_code
```

#### `POST /finance/rates` — Manual Rate Load

```json
{
  "date": "2025-01-31",
  "rate_type": "closing",
  "provider_code": "MANUAL",
  "rates": [
    { "from": "USD", "to": "KES", "rate": 132.50, "bid": 132.10, "ask": 132.90 },
    { "from": "GBP", "to": "KES", "rate": 164.75 },
    { "from": "EUR", "to": "KES", "rate": 143.20 }
  ]
}
```

**Response** `201 Created`

```json
{
  "rate_set_id": "uuid",
  "date": "2025-01-31",
  "provider": "MANUAL",
  "rates_loaded": 3,
  "rates_rejected": 0,
  "warnings": []
}
```

### Conversion Endpoint

```
POST /finance/convert
```

**Request:**

```json
{
  "amount": 10000.00,
  "from_currency": "USD",
  "to_currency": "KES",
  "rate_date": "2025-01-31",
  "rate_type": "spot",
  "include_trace": true
}
```

**Response:**

```json
{
  "original_amount": 10000.00,
  "converted_amount": 1325000.00,
  "exchange_rate": 132.50,
  "inverse_rate": 0.00754717,
  "from_currency": "USD",
  "to_currency": "KES",
  "rate_date": "2025-01-31",
  "rate_type": "spot",
  "rate_source": {
    "provider_code": "OER",
    "provider_name": "OpenExchangeRates",
    "reliability_score": 95.0,
    "fetched_at": "2025-01-31T09:00:00Z"
  },
  "resolved_via": "direct",
  "confidence": 95.0,
  "is_stale": false,
  "computed_at": "2025-01-31T14:22:33Z",
  "trace": [
    { "stage": "ResolveRate",      "status": "completed", "detail": "Direct pair USD/KES from OER" },
    { "stage": "CheckStaleness",   "status": "completed", "detail": "Rate is current" },
    { "stage": "CheckSameCurrency","status": "skipped",   "detail": "Different currencies" },
    { "stage": "ApplyConversion",  "status": "completed", "detail": "10000 × 132.50 = 1325000" },
    { "stage": "ApplyPrecision",   "status": "completed", "detail": "Rounded to 2dp: 1325000.00" },
    { "stage": "ValidateResult",   "status": "completed", "detail": "Result is positive" }
  ]
}
```

### Revaluation Endpoints

```
GET    /finance/revaluations                        List revaluation runs
POST   /finance/revaluations                        Initiate a revaluation run
GET    /finance/revaluations/{id}                   Get revaluation detail and results
GET    /finance/revaluations/{id}/lines             Per-currency, per-account lines
POST   /finance/revaluations/dry-run                Preview revaluation without committing
```

#### `POST /finance/revaluations`

```json
{
  "period_end_date": "2025-01-31",
  "currencies": [],
  "method": "current_rate",
  "rate_type": "closing",
  "dry_run": false
}
```

**Response** `202 Accepted`

```json
{
  "revaluation_id": "uuid",
  "status": "processing",
  "workflow_id": "reval-uuid",
  "currencies_queued": ["USD", "GBP", "EUR"],
  "estimated_completion_seconds": 30
}
```

### Hedge Contract Endpoints

```
GET    /finance/hedges                              List hedge contracts
POST   /finance/hedges                              Create a new hedge contract
GET    /finance/hedges/{id}                         Get contract detail + MTM
PATCH  /finance/hedges/{id}                         Update contract terms (pre-settlement)
POST   /finance/hedges/{id}/settle                  Record settlement
POST   /finance/hedges/{id}/terminate               Early termination
GET    /finance/hedges/{id}/effectiveness           Latest effectiveness test result
GET    /finance/hedges/summary                      Portfolio MTM summary
```

### Exposure Analytics Endpoints

```
GET    /finance/exposure                            Current exposure snapshot (all currencies)
GET    /finance/exposure/{currency}                 Exposure for a specific currency
GET    /finance/exposure/{currency}/sensitivity     Sensitivity analysis
GET    /finance/exposure/history                    Historical snapshots (trend)
```

### Rate Analytics Endpoints (readonly_role)

```
GET    /finance/rates/{from}/{to}/trend             Rate trend chart data
GET    /finance/rates/volatility                    Volatility by currency pair
GET    /finance/rates/divergence                    Multi-provider rate divergence
GET    /finance/fx-summary                          FX gain/loss P&L summary
```

---

## 14. Business Rules & Validation

### Currency Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `FX-CCY-001` | Exactly one functional currency per tenant; cannot be changed after first transaction | Error |
| `FX-CCY-002` | Functional currency cannot be deactivated | Error |
| `FX-CCY-003` | A currency cannot be deactivated while it has open FX transaction positions | Error |
| `FX-CCY-004` | Currency code must be a valid ISO 4217 code | Error |
| `FX-CCY-005` | `decimal_places` must be 0, 2, or 3 (matching the ISO 4217 definition) | Error |

### Rate Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `FX-RTE-001` | Exchange rate must be positive and non-zero | Error |
| `FX-RTE-002` | From currency must not equal to currency | Error |
| `FX-RTE-003` | Rate date cannot be more than 1 calendar day in the future | Error |
| `FX-RTE-004` | If a rate for the same pair, date, type, and provider already exists, it is updated not duplicated | Automatic |
| `FX-RTE-005` | A rate change of more than the configured `rate_alert_threshold_pct` (default 5%) triggers an alert to Finance | Warning |
| `FX-RTE-006` | If no rate is found for the requested date, the most recent available rate within `max_rate_age_days` (default 1) may be used, flagged as stale | Warning |
| `FX-RTE-007` | Rates from `MANUAL` provider always require `finance.currencies.load-rates` permission | Error |
| `FX-RTE-008` | API-key-encrypted fields for rate providers are never returned in API responses | Error |

### Conversion Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `FX-CNV-001` | A conversion with no available rate must fail explicitly — never return zero | Error |
| `FX-CNV-002` | A stale rate conversion is allowed by default but flagged in the trace | Warning |
| `FX-CNV-003` | A cross-rate computation must record both component rates in the trace | Error |
| `FX-CNV-004` | Converted amounts are rounded to the target currency's `decimal_places` | Automatic |
| `FX-CNV-005` | Confidence below `min_confidence_threshold` (default 60) is flagged for Finance review | Warning |
| `FX-CNV-006` | Same-currency conversions always return 1:1 with no rate lookup | Automatic |

### Revaluation Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `FX-RVL-001` | A revaluation run cannot be created if the period is already hard-closed | Error |
| `FX-RVL-002` | Only one revaluation run per currency per period-end date may be active; re-running supersedes the prior run | Automatic |
| `FX-RVL-003` | Dry-run mode must not persist any records or emit any events | Error |
| `FX-RVL-004` | A revaluation run must complete the reversal for the next period's opening before the period is hard-closed | Warning |
| `FX-RVL-005` | Adjustments below KES 1 (or functional currency equivalent) are not posted (de minimis threshold — configurable) | Configurable |
| `FX-RVL-006` | Revaluation method (`current_rate`, `temporal`, `monetary_nonmonetary`) must be consistent within a fiscal year | Error |

### Hedge Rules

| Rule ID | Rule | Severity |
|---|---|---|
| `FX-HDG-001` | Hedge contract notional must be positive | Error |
| `FX-HDG-002` | Maturity date must be after trade date | Error |
| `FX-HDG-003` | Base currency must not equal quote currency | Error |
| `FX-HDG-004` | A collar contract must have `cap_rate` > `floor_rate` | Error |
| `FX-HDG-005` | Hedge accounting designation (`cash_flow`, `fair_value`, `net_investment`) requires a linked hedged item ID | Error |
| `FX-HDG-006` | Effectiveness ratio outside 80–125% results in automatic de-designation and P&L reclassification of all deferred gains/losses | Automatic |
| `FX-HDG-007` | A settled or expired contract cannot be modified | Error |

---

## 15. Database Schema & RLS

```sql
-- ── currencies ────────────────────────────────────────────────────────────────
CREATE TABLE currencies (
    tenant_id               UUID        NOT NULL REFERENCES tenants(id),
    id                      UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by              UUID,
    updated_by              UUID,

    code                    CHAR(3)     NOT NULL,
    name                    TEXT        NOT NULL,
    symbol                  TEXT        NOT NULL,
    decimal_places          SMALLINT    NOT NULL DEFAULT 2
                                CHECK (decimal_places IN (0, 2, 3)),
    decimal_separator       CHAR(1)     NOT NULL DEFAULT '.',
    thousands_separator     CHAR(1)     NOT NULL DEFAULT ',',
    symbol_position         TEXT        NOT NULL DEFAULT 'before'
                                CHECK (symbol_position IN ('before', 'after')),
    is_functional           BOOLEAN     NOT NULL DEFAULT FALSE,
    is_active               BOOLEAN     NOT NULL DEFAULT TRUE,
    is_crypto               BOOLEAN     NOT NULL DEFAULT FALSE,
    is_freely_convertible   BOOLEAN     NOT NULL DEFAULT TRUE,
    requires_approval       BOOLEAN     NOT NULL DEFAULT FALSE,
    primary_country_code    CHAR(2),
    secondary_countries     JSONB       NOT NULL DEFAULT '[]',
    introduced_date         DATE,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, code)
);

-- Enforce single functional currency per tenant at DB level
CREATE UNIQUE INDEX idx_currencies_one_functional
    ON currencies (tenant_id)
    WHERE is_functional = TRUE;

CREATE INDEX idx_currencies_tenant ON currencies (tenant_id);

ALTER TABLE currencies ENABLE ROW LEVEL SECURITY;
ALTER TABLE currencies FORCE  ROW LEVEL SECURITY;

CREATE POLICY currencies_app_all ON currencies
    FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY currencies_ro_select ON currencies
    FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── rate_providers ────────────────────────────────────────────────────────────
CREATE TABLE rate_providers (
    tenant_id               UUID        NOT NULL REFERENCES tenants(id),
    id                      UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    code                    TEXT        NOT NULL,
    name                    TEXT        NOT NULL,
    provider_type           TEXT        NOT NULL
                                CHECK (provider_type IN ('central_bank','commercial_bank','financial_service','manual')),
    api_endpoint            TEXT,
    api_key_encrypted       TEXT,       -- never returned in API; encrypted at column level
    auth_method             TEXT        DEFAULT 'api_key'
                                CHECK (auth_method IN ('api_key','oauth','basic_auth','none')),
    update_frequency        TEXT        NOT NULL DEFAULT 'daily'
                                CHECK (update_frequency IN ('real_time','hourly','daily','weekly')),
    default_rate_type       TEXT        NOT NULL DEFAULT 'mid'
                                CHECK (default_rate_type IN ('spot','mid','bid','ask','official','closing')),
    base_currency           CHAR(3)     NOT NULL DEFAULT 'USD',
    priority                INT         NOT NULL DEFAULT 100,
    reliability_score       NUMERIC(5,2) NOT NULL DEFAULT 100.0
                                CHECK (reliability_score BETWEEN 0 AND 100),
    is_primary              BOOLEAN     NOT NULL DEFAULT FALSE,
    is_active               BOOLEAN     NOT NULL DEFAULT TRUE,
    daily_request_limit     INT,
    last_successful_fetch   TIMESTAMPTZ,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, code)
);

CREATE INDEX idx_rate_providers_tenant ON rate_providers (tenant_id);

ALTER TABLE rate_providers ENABLE ROW LEVEL SECURITY;
ALTER TABLE rate_providers FORCE  ROW LEVEL SECURITY;

CREATE POLICY rate_providers_app_all ON rate_providers
    FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY rate_providers_ro_select ON rate_providers
    FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── exchange_rates ────────────────────────────────────────────────────────────
CREATE TABLE exchange_rates (
    tenant_id               UUID        NOT NULL REFERENCES tenants(id),
    id                      UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    from_currency           CHAR(3)     NOT NULL,
    to_currency             CHAR(3)     NOT NULL,
    rate_date               DATE        NOT NULL,
    rate_time               TIME,
    exchange_rate           NUMERIC(18,8) NOT NULL CHECK (exchange_rate > 0),
    inverse_rate            NUMERIC(18,8),
    rate_type               TEXT        NOT NULL DEFAULT 'spot'
                                CHECK (rate_type IN ('spot','forward','budget','historical','average','closing','mid','bid','ask','official')),
    rate_source             TEXT        NOT NULL DEFAULT 'mid'
                                CHECK (rate_source IN ('bid','ask','mid','official','closing')),
    provider_id             UUID        REFERENCES rate_providers(id),
    source_reference        TEXT,

    effective_from          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_to            TIMESTAMPTZ,

    bid_rate                NUMERIC(18,8),
    ask_rate                NUMERIC(18,8),
    spread_percentage       NUMERIC(8,6),
    forward_points          NUMERIC(10,6),
    maturity_date           DATE,

    confidence_level        NUMERIC(5,2) NOT NULL DEFAULT 100.0
                                CHECK (confidence_level BETWEEN 0 AND 100),
    liquidity_indicator     TEXT        NOT NULL DEFAULT 'high'
                                CHECK (liquidity_indicator IN ('high','medium','low')),
    is_official_rate        BOOLEAN     NOT NULL DEFAULT FALSE,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, from_currency, to_currency, rate_date, rate_type, provider_id),
    CONSTRAINT different_currencies CHECK (from_currency != to_currency)
);

CREATE INDEX idx_exchange_rates_tenant      ON exchange_rates (tenant_id);
CREATE INDEX idx_exchange_rates_pair_date   ON exchange_rates (tenant_id, from_currency, to_currency, rate_date DESC);
CREATE INDEX idx_exchange_rates_effective   ON exchange_rates (tenant_id, effective_from, effective_to);

ALTER TABLE exchange_rates ENABLE ROW LEVEL SECURITY;
ALTER TABLE exchange_rates FORCE  ROW LEVEL SECURITY;

CREATE POLICY exchange_rates_app_all ON exchange_rates
    FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY exchange_rates_ro_select ON exchange_rates
    FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── exchange_rate_history ─────────────────────────────────────────────────────
CREATE TABLE exchange_rate_history (
    tenant_id               UUID        NOT NULL REFERENCES tenants(id),
    id                      UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    from_currency           CHAR(3)     NOT NULL,
    to_currency             CHAR(3)     NOT NULL,
    period_start            DATE        NOT NULL,
    period_end              DATE        NOT NULL,

    opening_rate            NUMERIC(18,8),
    closing_rate            NUMERIC(18,8),
    high_rate               NUMERIC(18,8),
    low_rate                NUMERIC(18,8),
    average_rate            NUMERIC(18,8),
    weighted_avg_rate       NUMERIC(18,8),
    update_count            INT         NOT NULL DEFAULT 0,
    std_deviation           NUMERIC(10,8),
    trend_direction         TEXT        CHECK (trend_direction IN ('up','down','stable')),

    PRIMARY KEY (id),
    UNIQUE (tenant_id, from_currency, to_currency, period_start, period_end),
    CONSTRAINT valid_period CHECK (period_end >= period_start)
);

CREATE INDEX idx_rate_history_tenant ON exchange_rate_history (tenant_id);
CREATE INDEX idx_rate_history_pair   ON exchange_rate_history (tenant_id, from_currency, to_currency, period_start DESC);

ALTER TABLE exchange_rate_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE exchange_rate_history FORCE  ROW LEVEL SECURITY;

CREATE POLICY rate_history_app_all ON exchange_rate_history
    FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY rate_history_ro_select ON exchange_rate_history
    FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── fx_transactions ───────────────────────────────────────────────────────────
-- Append-only. One record per monetary line in a foreign currency.
CREATE TABLE fx_transactions (
    tenant_id               UUID        NOT NULL REFERENCES tenants(id),
    id                      UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    transaction_id          UUID        NOT NULL,
    line_number             INT         NOT NULL DEFAULT 1,
    transaction_currency    CHAR(3)     NOT NULL,
    functional_currency     CHAR(3)     NOT NULL,
    transaction_amount      NUMERIC(18,4) NOT NULL,
    functional_amount       NUMERIC(18,4) NOT NULL,

    original_rate           NUMERIC(18,8) NOT NULL CHECK (original_rate > 0),
    original_rate_date      DATE        NOT NULL,
    original_rate_type      TEXT        NOT NULL,
    original_rate_source    TEXT        NOT NULL,
    conversion_trace        JSONB       NOT NULL DEFAULT '{}',

    revaluation_rate        NUMERIC(18,8),
    revaluation_date        DATE,
    unrealised_gain_loss    NUMERIC(18,4) NOT NULL DEFAULT 0,

    payment_rate            NUMERIC(18,8),
    payment_date            DATE,
    realised_gain_loss      NUMERIC(18,4),

    gain_loss_account_id    UUID,
    unrealised_gl_account_id UUID,

    is_hedged               BOOLEAN     NOT NULL DEFAULT FALSE,
    hedge_contract_id       UUID        REFERENCES fx_hedge_contracts(id),
    hedge_effectiveness     NUMERIC(5,2),

    revaluation_frequency   TEXT        NOT NULL DEFAULT 'monthly'
                                CHECK (revaluation_frequency IN ('daily','weekly','monthly','quarterly')),

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, transaction_id, line_number)
);

CREATE INDEX idx_fx_transactions_tenant      ON fx_transactions (tenant_id);
CREATE INDEX idx_fx_transactions_transaction ON fx_transactions (tenant_id, transaction_id);
CREATE INDEX idx_fx_transactions_open        ON fx_transactions (tenant_id, transaction_currency)
    WHERE payment_date IS NULL;

ALTER TABLE fx_transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE fx_transactions FORCE  ROW LEVEL SECURITY;

CREATE POLICY fx_transactions_app_all ON fx_transactions
    FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY fx_transactions_ro_select ON fx_transactions
    FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── revaluations ──────────────────────────────────────────────────────────────
CREATE TABLE revaluations (
    tenant_id               UUID        NOT NULL REFERENCES tenants(id),
    id                      UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by              UUID,

    revaluation_date        DATE        NOT NULL,
    method                  TEXT        NOT NULL DEFAULT 'current_rate'
                                CHECK (method IN ('current_rate','temporal','monetary_nonmonetary')),
    status                  TEXT        NOT NULL DEFAULT 'processing'
                                CHECK (status IN ('processing','completed','failed','dry_run')),
    currencies_processed    TEXT[]      NOT NULL DEFAULT '{}',
    total_gain_loss         NUMERIC(18,4) NOT NULL DEFAULT 0,
    unrealised_gain_loss    NUMERIC(18,4) NOT NULL DEFAULT 0,
    affected_account_count  INT         NOT NULL DEFAULT 0,
    affected_tx_count       INT         NOT NULL DEFAULT 0,
    journal_ref             TEXT,
    posted_to_gl            BOOLEAN     NOT NULL DEFAULT FALSE,
    processing_ms           INT,
    error_detail            TEXT,

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id)
);

CREATE INDEX idx_revaluations_tenant ON revaluations (tenant_id);
CREATE INDEX idx_revaluations_date   ON revaluations (tenant_id, revaluation_date DESC);

ALTER TABLE revaluations ENABLE ROW LEVEL SECURITY;
ALTER TABLE revaluations FORCE  ROW LEVEL SECURITY;

CREATE POLICY revaluations_app_all ON revaluations
    FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY revaluations_ro_select ON revaluations
    FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── fx_hedge_contracts ────────────────────────────────────────────────────────
CREATE TABLE fx_hedge_contracts (
    tenant_id               UUID        NOT NULL REFERENCES tenants(id),
    id                      UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by              UUID,
    updated_by              UUID,

    contract_number         TEXT        NOT NULL,
    contract_type           TEXT        NOT NULL
                                CHECK (contract_type IN ('forward','option','swap','collar')),
    bank_id                 UUID,
    counterparty_name       TEXT        NOT NULL,

    base_currency           CHAR(3)     NOT NULL,
    quote_currency          CHAR(3)     NOT NULL,
    notional_amount         NUMERIC(18,4) NOT NULL CHECK (notional_amount > 0),
    contract_rate           NUMERIC(18,8) NOT NULL,

    trade_date              DATE        NOT NULL,
    value_date              DATE        NOT NULL,
    maturity_date           DATE        NOT NULL,

    option_type             TEXT        CHECK (option_type IN ('call','put')),
    strike_rate             NUMERIC(18,8),
    premium_amount          NUMERIC(18,4),
    premium_currency        CHAR(3),
    cap_rate                NUMERIC(18,8),
    floor_rate              NUMERIC(18,8),

    designation             TEXT        CHECK (designation IN ('cash_flow','fair_value','net_investment')),
    hedged_item_type        TEXT,
    hedged_item_id          UUID,
    effectiveness_method    TEXT        CHECK (effectiveness_method IN ('dollar_offset','regression','critical_terms')),

    mtm_value               NUMERIC(18,4) NOT NULL DEFAULT 0,
    unrealised_gl           NUMERIC(18,4) NOT NULL DEFAULT 0,
    last_valuation_date     DATE,

    settlement_amount       NUMERIC(18,4),
    settlement_date         DATE,
    settlement_status       TEXT        NOT NULL DEFAULT 'open'
                                CHECK (settlement_status IN ('open','partially_settled','fully_settled','expired')),
    status                  TEXT        NOT NULL DEFAULT 'active'
                                CHECK (status IN ('active','matured','cancelled','terminated')),

    PRIMARY KEY (id),
    UNIQUE (tenant_id, id),
    UNIQUE (tenant_id, contract_number),
    CONSTRAINT different_hedge_currencies CHECK (base_currency != quote_currency),
    CONSTRAINT valid_maturity CHECK (maturity_date > trade_date),
    CONSTRAINT collar_rates CHECK (
        contract_type != 'collar'
        OR (cap_rate IS NOT NULL AND floor_rate IS NOT NULL AND cap_rate > floor_rate)
    )
);

CREATE INDEX idx_hedge_contracts_tenant ON fx_hedge_contracts (tenant_id);
CREATE INDEX idx_hedge_contracts_status ON fx_hedge_contracts (tenant_id, status)
    WHERE status = 'active';

ALTER TABLE fx_hedge_contracts ENABLE ROW LEVEL SECURITY;
ALTER TABLE fx_hedge_contracts FORCE  ROW LEVEL SECURITY;

CREATE POLICY hedge_contracts_app_all ON fx_hedge_contracts
    FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY hedge_contracts_ro_select ON fx_hedge_contracts
    FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── currency_exposures ────────────────────────────────────────────────────────
CREATE TABLE currency_exposures (
    tenant_id               UUID        NOT NULL REFERENCES tenants(id),
    id                      UUID        NOT NULL DEFAULT gen_random_uuid(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    currency_code           CHAR(3)     NOT NULL,
    functional_currency     CHAR(3)     NOT NULL,
    snapshot_date           DATE        NOT NULL,

    transaction_exposure    NUMERIC(18,4) NOT NULL DEFAULT 0,
    translation_exposure    NUMERIC(18,4) NOT NULL DEFAULT 0,
    economic_exposure       NUMERIC(18,4) NOT NULL DEFAULT 0,

    gross_exposure          NUMERIC(18,4) NOT NULL DEFAULT 0,
    hedged_amount           NUMERIC(18,4) NOT NULL DEFAULT 0,
    net_exposure            NUMERIC(18,4) NOT NULL DEFAULT 0,

    value_at_risk_95        NUMERIC(18,4),
    expected_shortfall      NUMERIC(18,4),
    volatility              NUMERIC(8,6),

    bucket_0_30             NUMERIC(18,4) NOT NULL DEFAULT 0,
    bucket_31_90            NUMERIC(18,4) NOT NULL DEFAULT 0,
    bucket_91_180           NUMERIC(18,4) NOT NULL DEFAULT 0,
    bucket_181_365          NUMERIC(18,4) NOT NULL DEFAULT 0,
    bucket_over_365         NUMERIC(18,4) NOT NULL DEFAULT 0,

    sensitivity_1pct        NUMERIC(18,4),
    sensitivity_5pct        NUMERIC(18,4),
    sensitivity_10pct       NUMERIC(18,4),

    PRIMARY KEY (id),
    UNIQUE (tenant_id, currency_code, snapshot_date)
);

CREATE INDEX idx_currency_exposures_tenant ON currency_exposures (tenant_id);
CREATE INDEX idx_currency_exposures_date   ON currency_exposures (tenant_id, snapshot_date DESC);

ALTER TABLE currency_exposures ENABLE ROW LEVEL SECURITY;
ALTER TABLE currency_exposures FORCE  ROW LEVEL SECURITY;

CREATE POLICY currency_exposures_app_all ON currency_exposures
    FOR ALL TO application_role
    USING     (tenant_id = current_setting('app.current_tenant')::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY currency_exposures_ro_select ON currency_exposures
    FOR SELECT TO readonly_role
    USING (tenant_id = current_setting('app.current_tenant')::UUID);


-- ── Standard Views ────────────────────────────────────────────────────────────

-- Current open FX positions (receivables and payables in FC)
CREATE OR REPLACE VIEW v_open_fx_positions AS
SELECT
    ft.tenant_id,
    ft.id,
    ft.transaction_id,
    ft.transaction_currency,
    ft.functional_currency,
    ft.transaction_amount,
    ft.functional_amount,
    ft.original_rate,
    ft.original_rate_date,
    COALESCE(ft.revaluation_rate, ft.original_rate) AS current_rate,
    ft.unrealised_gain_loss,
    ft.is_hedged,
    ft.hedge_contract_id
FROM fx_transactions ft
WHERE ft.payment_date IS NULL;   -- only open (unsettled) positions

ALTER VIEW v_open_fx_positions SET (security_invoker = true);
GRANT SELECT ON v_open_fx_positions TO application_role;
GRANT SELECT ON v_open_fx_positions TO readonly_role;


-- FX gain/loss P&L summary
CREATE OR REPLACE VIEW v_fx_gainloss_summary AS
SELECT
    ft.tenant_id,
    ft.transaction_currency,
    ft.functional_currency,
    DATE_TRUNC('month', ft.payment_date)                    AS month,
    COUNT(*)                                                 AS settled_count,
    SUM(ft.realised_gain_loss)                              AS total_realised_gl,
    SUM(CASE WHEN ft.realised_gain_loss > 0
             THEN ft.realised_gain_loss ELSE 0 END)         AS total_realised_gain,
    SUM(CASE WHEN ft.realised_gain_loss < 0
             THEN ft.realised_gain_loss ELSE 0 END)         AS total_realised_loss
FROM fx_transactions ft
WHERE ft.payment_date IS NOT NULL
  AND ft.realised_gain_loss IS NOT NULL
GROUP BY ft.tenant_id, ft.transaction_currency, ft.functional_currency,
         DATE_TRUNC('month', ft.payment_date);

ALTER VIEW v_fx_gainloss_summary SET (security_invoker = true);
GRANT SELECT ON v_fx_gainloss_summary TO application_role;
GRANT SELECT ON v_fx_gainloss_summary TO readonly_role;
```

---

## 16. Workflow Orchestration (Temporal)

### Rate Fetch Workflow

Triggered on a schedule configured per provider (`update_frequency`). Fetches rates from the external provider and loads them into `exchange_rates`.

```go
// infrastructure/temporal/rate_fetch_workflow.go

func RateFetchWorkflow(ctx workflow.Context, input RateFetchInput) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 2 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts:        5,
            InitialInterval:        30 * time.Second,
            BackoffCoefficient:     2.0,
            MaximumInterval:        5 * time.Minute,
            NonRetryableErrorTypes: []string{"ErrProviderNotFound", "ErrInvalidAPIKey"},
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Fetch rates from provider
    var rates []ExchangeRate
    if err := workflow.ExecuteActivity(ctx, FetchRatesActivity, input).Get(ctx, &rates); err != nil {
        // On failure, record in provider's reliability score
        workflow.ExecuteActivity(ctx, RecordProviderFailureActivity, input.ProviderID, err.Error())
        return err
    }

    // Persist rates
    if err := workflow.ExecuteActivity(ctx, PersistRatesActivity, input.TenantID, rates).Get(ctx, nil); err != nil {
        return err
    }

    // Check for unusual movements and alert if needed
    return workflow.ExecuteActivity(ctx, CheckRateAlertsActivity, input.TenantID, rates).Get(ctx, nil)
}
```

### Revaluation Workflow

Triggered at period-end, either manually or by the period-close checklist workflow.

```go
func RevaluationWorkflow(ctx workflow.Context, input RevaluationInput) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 3},
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // 1. Get all active foreign currencies
    var currencies []string
    if err := workflow.ExecuteActivity(ctx, GetActiveFCurrenciesActivity, input.TenantID).Get(ctx, &currencies); err != nil {
        return err
    }

    // 2. Process each currency independently (parallel)
    var futures []workflow.Future
    for _, ccy := range currencies {
        f := workflow.ExecuteActivity(ctx, RevalueCurrencyActivity, RevalueCurrencyInput{
            TenantID:        input.TenantID,
            Currency:        ccy,
            RevaluationDate: input.PeriodEndDate,
            Method:          input.Method,
            DryRun:          input.DryRun,
        })
        futures = append(futures, f)
    }

    results := []CurrencyRevaluationResult{}
    for _, f := range futures {
        var r CurrencyRevaluationResult
        if err := f.Get(ctx, &r); err != nil {
            return err
        }
        results = append(results, r)
    }

    // 3. Aggregate and emit journal instructions
    if !input.DryRun {
        return workflow.ExecuteActivity(ctx, EmitRevaluationEventActivity, input.TenantID, results).Get(ctx, nil)
    }
    return nil
}
```

### Daily Rate History Aggregation Cron

```go
// Runs at 00:05 daily. Aggregates yesterday's rates into exchange_rate_history.
func RateHistoryAggregationWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 15 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)
    yesterday := time.Now().AddDate(0, 0, -1)
    return workflow.ExecuteActivity(ctx, AggregateRateHistoryActivity, yesterday).Get(ctx, nil)
}
```

### Currency Exposure Snapshot Cron

```go
// Runs at 01:00 daily. Computes and stores the exposure snapshot for yesterday.
func ExposureSnapshotWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 20 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)
    yesterday := time.Now().AddDate(0, 0, -1)
    return workflow.ExecuteActivity(ctx, ComputeExposureSnapshotActivity, yesterday).Get(ctx, nil)
}
```

### Hedge MTM Valuation Cron

```go
// Runs daily at 09:30 (after rates are loaded). Marks all active hedge contracts to market.
func HedgeMTMWorkflow(ctx workflow.Context) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)
    return workflow.ExecuteActivity(ctx, MarkHedgesToMarketActivity, time.Now()).Get(ctx, nil)
}
```

---

## 17. Configuration Flags

Stored per tenant in `tenant_config.finance.currency`.

| Flag | Default | Description |
|---|---|---|
| `base_currency` | `KES` | Tenant's functional currency. Immutable after first transaction. |
| `max_rate_age_days` | `1` | Maximum number of days a rate may be used before it is considered stale |
| `stale_rate_action` | `warn` | `warn` \| `block` — what to do when a stale rate is used |
| `rate_alert_threshold_pct` | `5` | Percentage change between consecutive rates that triggers an alert |
| `min_confidence_threshold` | `60` | Below this confidence score, flag for Finance review |
| `revaluation_method` | `current_rate` | `current_rate` \| `temporal` \| `monetary_nonmonetary` |
| `fx_unrealised_treatment` | `pl_immediate` | `pl_immediate` \| `oci_deferral` — IFRS vs GAAP treatment |
| `revaluation_auto_run` | `true` | Automatically trigger revaluation on period-end date |
| `revaluation_day_of_month` | `last` | Day on which the revaluation Temporal workflow is triggered |
| `de_minimis_threshold` | `1.00` | Revaluation adjustments below this amount (in functional currency) are not posted |
| `cross_rate_base` | `USD` | Base currency for cross-rate calculations when direct pair is unavailable |
| `hedge_effectiveness_min_pct` | `80` | Minimum effectiveness ratio (%) before de-designation |
| `hedge_effectiveness_max_pct` | `125` | Maximum effectiveness ratio (%) before de-designation |
| `exposure_snapshot_enabled` | `true` | Run daily FX exposure snapshot activity |
| `preferred_rate_providers` | `[]` | Ordered list of provider codes; overrides DB priority for this tenant |

---

## 18. Future Roadmap

### Real-Time Rate Streaming

For tenants with very high FX transaction volumes (e.g. cross-border wholesale operations, large forecourt networks buying fuel in USD), a WebSocket-based real-time rate stream from a financial data provider (Refinitiv, Bloomberg, or the CBK live API) can replace the scheduled fetch model. The `RateProviderAdapter` interface supports this without changing existing code.

### AI-Assisted Hedging Recommendations

The combination of the `ExposureSnapshot` historical data and the `RateHistory` series provides enough signal for a simple ML recommendation layer:
- "Your USD net exposure has exceeded your historical hedge threshold for 3 consecutive days"
- "Based on current volatility, a 3-month forward at today's rate would save an estimated KES 45,000 vs spot settlement"

### Netting Module Integration

When the Finance module's Intercompany Netting feature is extended, the Currency sub-module will need to compute net settlement amounts across currency pairs and apply the correct cross rates at the netting date, including handling partial settlement and residual FX differences.

### Central Bank API Adapters

As Awo ERP expands across African markets, each central bank has its own rate publication format and API. Planned adapters:
- Bank of Uganda (BOU) XLSX daily publication
- Bank of Tanzania (BOT) RSS feed
- Ghana Central Bank (BOG) API
- Reserve Bank of Zimbabwe (RBZ) daily gazette

Each follows the `RateProviderAdapter` interface — no changes to the conversion engine.

---

*This document is the authoritative specification for the Awo ERP Currency & Exchange Management sub-module. All multi-currency operations across Sales, Finance, HR & Payroll, Inventory, and Pricing modules must route through the conversion engine and FX transaction recording mechanism defined here. No module may independently compute or store an exchange rate.*



<!-- # AWO ERP Currency Management Module -->
<!---->
<!-- **Version**: 1.0   -->
<!-- **Date**: January 2025   -->
<!-- **Status**: Technical Specification   -->
<!---->
<!-- --- -->
<!---->
<!-- ##  Overview -->
<!---->
<!-- The Currency Management module provides  multi-currency support for global business operations. It handles exchange rate management, currency conversion, hedging operations, and compliance with international financial reporting standards. -->
<!---->
<!-- > **️ Core Database Schema**: For the main financial database schema (accounts, transactions, entries), see [Technical Architecture - Database Schema](../technical-architecture.md#database-schema-architecture). This guide covers currency-specific extensions. -->
<!---->
<!-- ##  Currency Configuration & Master Data -->
<!---->
<!-- ### Currency Master Data -->
<!---->
<!-- ```sql -->
<!-- -- Supported currencies with  details -->
<!-- CREATE TABLE currencies ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE, -->
<!--     entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE, -->
<!---->
<!--     -- Currency identification -->
<!--     currency_code CHAR(3) UNIQUE NOT NULL, -- ISO 4217 code (USD, EUR, GBP) -->
<!--     currency_name VARCHAR(100) NOT NULL, -->
<!--     currency_symbol VARCHAR(10), -->
<!---->
<!--     -- Currency formatting -->
<!--     decimal_places INTEGER DEFAULT 2, -->
<!--     decimal_separator CHAR(1) DEFAULT '.', -->
<!--     thousands_separator CHAR(1) DEFAULT ',', -->
<!--     symbol_position VARCHAR(10) DEFAULT 'before', -- before, after -->
<!---->
<!--     -- Currency attributes -->
<!--     is_base_currency BOOLEAN DEFAULT false, -->
<!--     is_active BOOLEAN DEFAULT true, -->
<!--     is_crypto_currency BOOLEAN DEFAULT false, -->
<!---->
<!--     -- Regional information -->
<!--     primary_country_code CHAR(2), -->
<!--     secondary_countries JSONB, -- Array of country codes where used -->
<!---->
<!--     -- Trading characteristics -->
<!--     is_freely_convertible BOOLEAN DEFAULT true, -->
<!--     trading_start_time TIME DEFAULT '00:00:00', -->
<!--     trading_end_time TIME DEFAULT '23:59:59', -->
<!--     trading_timezone VARCHAR(50), -->
<!---->
<!--     -- Compliance and regulations -->
<!--     requires_central_bank_approval BOOLEAN DEFAULT false, -->
<!--     exchange_control_restrictions JSONB, -->
<!---->
<!--     -- Historical tracking -->
<!--     introduced_date DATE, -->
<!--     last_revaluation_date DATE, -->
<!---->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!--     updated_at TIMESTAMPTZ DEFAULT NOW(), -->
<!---->
<!--     CONSTRAINT valid_symbol_position CHECK (symbol_position IN ('before', 'after')) -->
<!-- ); -->
<!---->
<!-- -- Exchange rate providers and sources -->
<!-- CREATE TABLE exchange_rate_providers ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE, -->
<!--     entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE, -->
<!---->
<!---->
<!--     -- Provider identification -->
<!--     provider_code VARCHAR(20) UNIQUE NOT NULL, -->
<!--     provider_name VARCHAR(255) NOT NULL, -->
<!--     provider_type VARCHAR(30) NOT NULL, -- central_bank, commercial_bank, financial_service, manual -->
<!---->
<!--     -- API configuration -->
<!--     api_endpoint VARCHAR(500), -->
<!--     api_key_encrypted VARCHAR(500), -- Encrypted API credentials -->
<!--     authentication_method VARCHAR(30), -- api_key, oauth, basic_auth -->
<!---->
<!--     -- Data characteristics -->
<!--     update_frequency VARCHAR(20) DEFAULT 'daily', -- real_time, hourly, daily, weekly -->
<!--     rate_type VARCHAR(20) DEFAULT 'mid', -- bid, ask, mid, official -->
<!--     base_currency CHAR(3) DEFAULT 'USD', -->
<!---->
<!--     -- Reliability and priority -->
<!--     priority_order INTEGER DEFAULT 100, -- Lower = higher priority -->
<!--     reliability_score DECIMAL(3,1) DEFAULT 100.0, -- 0-100 reliability rating -->
<!---->
<!--     -- Operational settings -->
<!--     is_active BOOLEAN DEFAULT true, -->
<!--     is_primary BOOLEAN DEFAULT false, -->
<!--     last_successful_update TIMESTAMPTZ, -->
<!---->
<!--     -- Rate limits and costs -->
<!--     daily_request_limit INTEGER, -->
<!--     cost_per_request DECIMAL(8,4), -->
<!---->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!--     updated_at TIMESTAMPTZ DEFAULT NOW(), -->
<!---->
<!--     CONSTRAINT valid_provider_type CHECK (provider_type IN ('central_bank', 'commercial_bank', 'financial_service', 'manual')), -->
<!--     CONSTRAINT valid_update_frequency CHECK (update_frequency IN ('real_time', 'hourly', 'daily', 'weekly')), -->
<!--     CONSTRAINT valid_rate_type CHECK (rate_type IN ('bid', 'ask', 'mid', 'official')) -->
<!-- ); -->
<!---->
<!-- -- Exchange rates with  tracking -->
<!-- CREATE TABLE exchange_rates ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE, -->
<!--     entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE, -->
<!---->
<!--     -- Rate identification -->
<!--     from_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code), -->
<!--     to_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code), -->
<!--     rate_date DATE NOT NULL, -->
<!--     rate_time TIME DEFAULT CURRENT_TIME, -->
<!---->
<!--     -- Rate values -->
<!--     exchange_rate DECIMAL(18,8) NOT NULL, -- High precision for crypto and exotic pairs -->
<!--     inverse_rate DECIMAL(18,8), -- Calculated inverse rate -->
<!---->
<!--     -- Rate types -->
<!--     rate_type VARCHAR(20) DEFAULT 'spot', -- spot, forward, budget, historical, average -->
<!--     rate_source VARCHAR(30) DEFAULT 'mid', -- bid, ask, mid, official, closing -->
<!---->
<!--     -- Rate provider -->
<!--     provider_id UUID REFERENCES exchange_rate_providers(id), -->
<!--     source_reference VARCHAR(100), -- Provider's reference ID -->
<!---->
<!--     -- Rate validity -->
<!--     effective_from TIMESTAMPTZ DEFAULT NOW(), -->
<!--     effective_to TIMESTAMPTZ, -->
<!---->
<!--     -- Rate spreads (for trading) -->
<!--     bid_rate DECIMAL(18,8), -->
<!--     ask_rate DECIMAL(18,8), -->
<!--     spread_percentage DECIMAL(8,6), -->
<!---->
<!--     -- Forward rate specific fields -->
<!--     forward_points DECIMAL(10,6), -->
<!--     maturity_date DATE, -- For forward contracts -->
<!---->
<!--     -- Volatility and analytics -->
<!--     volatility_measure DECIMAL(8,6), -->
<!--     daily_change_percentage DECIMAL(8,4), -->
<!---->
<!--     -- Quality indicators -->
<!--     confidence_level DECIMAL(5,2) DEFAULT 100.0, -- Confidence in rate accuracy -->
<!--     liquidity_indicator VARCHAR(10) DEFAULT 'high', -- high, medium, low -->
<!---->
<!--     -- Audit and compliance -->
<!--     is_official_rate BOOLEAN DEFAULT false, -- Central bank or official rate -->
<!--     regulatory_approved BOOLEAN DEFAULT true, -->
<!---->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!---->
<!--     CONSTRAINT valid_rate_type CHECK (rate_type IN ('spot', 'forward', 'budget', 'historical', 'average')), -->
<!--     CONSTRAINT valid_rate_source CHECK (rate_source IN ('bid', 'ask', 'mid', 'official', 'closing')), -->
<!--     CONSTRAINT valid_liquidity CHECK (liquidity_indicator IN ('high', 'medium', 'low')), -->
<!--     CONSTRAINT positive_exchange_rate CHECK (exchange_rate > 0), -->
<!--     CONSTRAINT different_currencies CHECK (from_currency != to_currency), -->
<!---->
<!--     UNIQUE(from_currency, to_currency, rate_date, rate_type, provider_id) -->
<!-- ); -->
<!---->
<!-- -- Exchange rate history for analytical purposes -->
<!-- CREATE TABLE exchange_rate_history ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE, -->
<!--     entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE, -->
<!---->
<!--     -- Historical rate data -->
<!--     from_currency CHAR(3) NOT NULL, -->
<!--     to_currency CHAR(3) NOT NULL, -->
<!--     rate_period_start DATE NOT NULL, -->
<!--     rate_period_end DATE NOT NULL, -->
<!---->
<!--     -- Aggregated rates -->
<!--     opening_rate DECIMAL(18,8), -->
<!--     closing_rate DECIMAL(18,8), -->
<!--     high_rate DECIMAL(18,8), -->
<!--     low_rate DECIMAL(18,8), -->
<!--     average_rate DECIMAL(18,8), -->
<!--     weighted_average_rate DECIMAL(18,8), -->
<!---->
<!--     -- Volume and activity -->
<!--     trading_volume DECIMAL(18,2), -->
<!--     number_of_updates INTEGER DEFAULT 0, -->
<!---->
<!--     -- Statistical measures -->
<!--     standard_deviation DECIMAL(10,8), -->
<!--     coefficient_of_variation DECIMAL(8,4), -->
<!---->
<!--     -- Trend analysis -->
<!--     trend_direction VARCHAR(10), -- up, down, stable -->
<!--     momentum_indicator DECIMAL(8,4), -->
<!---->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!---->
<!--     CONSTRAINT valid_trend_direction CHECK (trend_direction IN ('up', 'down', 'stable')), -->
<!--     CONSTRAINT valid_period CHECK (rate_period_end >= rate_period_start) -->
<!-- ); -->
<!-- ``` -->
<!---->
<!-- ### Currency Conversion Engine -->
<!---->
<!-- ```typescript -->
<!-- interface CurrencyConversionEngine { -->
<!--   convertAmount(request: ConversionRequest): Promise<ConversionResult>; -->
<!--   getExchangeRate(fromCurrency: string, toCurrency: string, rateDate?: Date, rateType?: string): Promise<ExchangeRate>; -->
<!--   calculateCrossRate(baseCurrency: string, fromCurrency: string, toCurrency: string): Promise<CrossRateResult>; -->
<!--   getBulkRates(request: BulkRateRequest): Promise<BulkRateResponse>; -->
<!-- } -->
<!---->
<!-- interface ConversionRequest { -->
<!--   amount: number; -->
<!--   from_currency: string; -->
<!--   to_currency: string; -->
<!--   rate_date?: Date; -->
<!--   rate_type?: 'spot' | 'budget' | 'forward' | 'average'; -->
<!--   provider_preference?: string[]; -->
<!--   precision?: number; -->
<!-- } -->
<!---->
<!-- interface ConversionResult { -->
<!--   original_amount: number; -->
<!--   converted_amount: number; -->
<!--   exchange_rate: number; -->
<!--   inverse_rate: number; -->
<!--   from_currency: string; -->
<!--   to_currency: string; -->
<!--   conversion_date: Date; -->
<!--   rate_source: ExchangeRateSource; -->
<!--   confidence_level: number; -->
<!--   calculation_method: string; -->
<!-- } -->
<!---->
<!-- interface ExchangeRateSource { -->
<!--   provider_name: string; -->
<!--   rate_type: string; -->
<!--   rate_time: Date; -->
<!--   reliability_score: number; -->
<!-- } -->
<!---->
<!-- class CurrencyService implements CurrencyConversionEngine { -->
<!--   async convertAmount(request: ConversionRequest): Promise<ConversionResult> { -->
<!--     // Validate currencies -->
<!--     await this.validateCurrencies(request.from_currency, request.to_currency); -->
<!---->
<!--     // Handle same currency conversion -->
<!--     if (request.from_currency === request.to_currency) { -->
<!--       return this.createSameCurrencyResult(request); -->
<!--     } -->
<!---->
<!--     // Get exchange rate -->
<!--     const exchangeRate = await this.getExchangeRate( -->
<!--       request.from_currency, -->
<!--       request.to_currency, -->
<!--       request.rate_date, -->
<!--       request.rate_type -->
<!--     ); -->
<!---->
<!--     // Apply precision settings -->
<!--     const precision = request.precision || await this.getCurrencyPrecision(request.to_currency); -->
<!---->
<!--     // Perform conversion -->
<!--     const convertedAmount = this.applyPrecision( -->
<!--       request.amount * exchangeRate.rate, -->
<!--       precision -->
<!--     ); -->
<!---->
<!--     return { -->
<!--       original_amount: request.amount, -->
<!--       converted_amount: convertedAmount, -->
<!--       exchange_rate: exchangeRate.rate, -->
<!--       inverse_rate: 1 / exchangeRate.rate, -->
<!--       from_currency: request.from_currency, -->
<!--       to_currency: request.to_currency, -->
<!--       conversion_date: exchangeRate.rate_date, -->
<!--       rate_source: { -->
<!--         provider_name: exchangeRate.provider_name, -->
<!--         rate_type: exchangeRate.rate_type, -->
<!--         rate_time: exchangeRate.created_at, -->
<!--         reliability_score: exchangeRate.reliability_score -->
<!--       }, -->
<!--       confidence_level: exchangeRate.confidence_level, -->
<!--       calculation_method: 'direct_conversion' -->
<!--     }; -->
<!--   } -->
<!---->
<!--   async getExchangeRate( -->
<!--     fromCurrency: string,  -->
<!--     toCurrency: string,  -->
<!--     rateDate: Date = new Date(), -->
<!--     rateType: string = 'spot' -->
<!--   ): Promise<ExchangeRate> { -->
<!---->
<!--     // Try direct rate first -->
<!--     let rate = await this.getDirectRate(fromCurrency, toCurrency, rateDate, rateType); -->
<!---->
<!--     if (rate) { -->
<!--       return rate; -->
<!--     } -->
<!---->
<!--     // Try inverse rate -->
<!--     rate = await this.getInverseRate(fromCurrency, toCurrency, rateDate, rateType); -->
<!---->
<!--     if (rate) { -->
<!--       return rate; -->
<!--     } -->
<!---->
<!--     // Calculate cross rate through base currency -->
<!--     const baseCurrency = await this.getBaseCurrency(); -->
<!--     return await this.calculateCrossRate(baseCurrency, fromCurrency, toCurrency); -->
<!--   } -->
<!---->
<!--   async calculateCrossRate( -->
<!--     baseCurrency: string,  -->
<!--     fromCurrency: string,  -->
<!--     toCurrency: string -->
<!--   ): Promise<CrossRateResult> { -->
<!---->
<!--     // Get rates: FROM -> BASE and BASE -> TO -->
<!--     const fromToBase = await this.getDirectRate(fromCurrency, baseCurrency); -->
<!--     const baseToTo = await this.getDirectRate(baseCurrency, toCurrency); -->
<!---->
<!--     if (!fromToBase || !baseToTo) { -->
<!--       throw new Error(`Cannot calculate cross rate for ${fromCurrency}/${toCurrency} through ${baseCurrency}`); -->
<!--     } -->
<!---->
<!--     // Calculate cross rate -->
<!--     const crossRate = fromToBase.rate * baseToTo.rate; -->
<!---->
<!--     return { -->
<!--       from_currency: fromCurrency, -->
<!--       to_currency: toCurrency, -->
<!--       cross_rate: crossRate, -->
<!--       base_currency: baseCurrency, -->
<!--       component_rates: [ -->
<!--         { pair: `${fromCurrency}/${baseCurrency}`, rate: fromToBase.rate }, -->
<!--         { pair: `${baseCurrency}/${toCurrency}`, rate: baseToTo.rate } -->
<!--       ], -->
<!--       calculation_date: new Date(), -->
<!--       confidence_level: Math.min(fromToBase.confidence_level, baseToTo.confidence_level) -->
<!--     }; -->
<!--   } -->
<!---->
<!--   private async getDirectRate( -->
<!--     fromCurrency: string,  -->
<!--     toCurrency: string,  -->
<!--     rateDate: Date = new Date(), -->
<!--     rateType: string = 'spot' -->
<!--   ): Promise<ExchangeRate | null> { -->
<!---->
<!--     const query = ` -->
<!--       SELECT er.*, erp.provider_name, erp.reliability_score -->
<!--       FROM exchange_rates er -->
<!--       JOIN exchange_rate_providers erp ON er.provider_id = erp.id -->
<!--       WHERE er.from_currency = $1  -->
<!--         AND er.to_currency = $2 -->
<!--         AND er.rate_date <= $3 -->
<!--         AND er.rate_type = $4 -->
<!--         AND er.effective_from <= NOW() -->
<!--         AND (er.effective_to IS NULL OR er.effective_to >= NOW()) -->
<!--       ORDER BY er.rate_date DESC, erp.priority_order ASC -->
<!--       LIMIT 1 -->
<!--     `; -->
<!---->
<!--     const result = await this.db.query(query, [fromCurrency, toCurrency, rateDate, rateType]); -->
<!--     return result.rows[0] || null; -->
<!--   } -->
<!---->
<!--   private async getInverseRate( -->
<!--     fromCurrency: string,  -->
<!--     toCurrency: string,  -->
<!--     rateDate: Date = new Date(), -->
<!--     rateType: string = 'spot' -->
<!--   ): Promise<ExchangeRate | null> { -->
<!---->
<!--     const directRate = await this.getDirectRate(toCurrency, fromCurrency, rateDate, rateType); -->
<!---->
<!--     if (!directRate) { -->
<!--       return null; -->
<!--     } -->
<!---->
<!--     return { -->
<!--       ...directRate, -->
<!--       from_currency: fromCurrency, -->
<!--       to_currency: toCurrency, -->
<!--       rate: 1 / directRate.rate, -->
<!--       inverse_rate: directRate.rate, -->
<!--       calculation_method: 'inverse' -->
<!--     }; -->
<!--   } -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- ## ️ Multi-Currency Transaction Processing -->
<!---->
<!-- ### Currency Transaction Management -->
<!---->
<!-- ```sql -->
<!-- -- Currency-aware transactions -->
<!-- CREATE TABLE currency_transactions ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE, -->
<!---->
<!--     -- Transaction identification -->
<!--     transaction_id UUID NOT NULL REFERENCES finance_transactions(id), -->
<!--     line_number INTEGER NOT NULL DEFAULT 1, -->
<!---->
<!--     -- Currency details -->
<!--     transaction_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code), -->
<!--     functional_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code), -- Company's reporting currency -->
<!---->
<!--     -- Original amounts -->
<!--     transaction_amount DECIMAL(15,2) NOT NULL, -->
<!---->
<!--     -- Converted amounts -->
<!--     functional_amount DECIMAL(15,2) NOT NULL, -- Amount in functional currency -->
<!--     exchange_rate DECIMAL(18,8) NOT NULL, -->
<!--     rate_date DATE NOT NULL, -->
<!--     rate_type VARCHAR(20) DEFAULT 'spot', -->
<!---->
<!--     -- Exchange differences -->
<!--     realized_gain_loss DECIMAL(15,2) DEFAULT 0, -- Realized when payment occurs -->
<!--     unrealized_gain_loss DECIMAL(15,2) DEFAULT 0, -- Mark-to-market revaluation -->
<!---->
<!--     -- Rate tracking -->
<!--     original_exchange_rate DECIMAL(18,8), -- Rate at transaction date -->
<!--     payment_exchange_rate DECIMAL(18,8), -- Rate at payment date -->
<!--     revaluation_exchange_rate DECIMAL(18,8), -- Rate at revaluation date -->
<!---->
<!--     -- Currency hedging -->
<!--     is_hedged BOOLEAN DEFAULT false, -->
<!--     hedge_contract_id UUID REFERENCES hedge_contracts(id), -->
<!--     hedge_effectiveness_percentage DECIMAL(5,2), -->
<!---->
<!--     -- Revaluation tracking -->
<!--     last_revaluation_date DATE, -->
<!--     revaluation_frequency VARCHAR(20) DEFAULT 'monthly', -- daily, weekly, monthly, quarterly -->
<!---->
<!--     -- GL impact -->
<!--     gain_loss_account_id UUID REFERENCES accounts(id), -->
<!--     unrealized_gain_loss_account_id UUID REFERENCES accounts(id), -->
<!---->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!--     updated_at TIMESTAMPTZ DEFAULT NOW(), -->
<!---->
<!--     CONSTRAINT valid_rate_type CHECK (rate_type IN ('spot', 'forward', 'budget', 'average')), -->
<!--     CONSTRAINT valid_revaluation_frequency CHECK (revaluation_frequency IN ('daily', 'weekly', 'monthly', 'quarterly')), -->
<!--     CONSTRAINT positive_exchange_rate CHECK (exchange_rate > 0), -->
<!---->
<!--     UNIQUE(transaction_id, line_number) -->
<!-- ); -->
<!---->
<!-- -- Currency revaluation history -->
<!-- CREATE TABLE currency_revaluations ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE, -->
<!---->
<!--     -- Revaluation identification -->
<!--     revaluation_date DATE NOT NULL, -->
<!--     revaluation_batch_id UUID NOT NULL, -->
<!---->
<!--     -- Currency scope -->
<!--     currency_code CHAR(3) NOT NULL REFERENCES currencies(currency_code), -->
<!--     functional_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code), -->
<!---->
<!--     -- Revaluation parameters -->
<!--     revaluation_rate DECIMAL(18,8) NOT NULL, -->
<!--     previous_rate DECIMAL(18,8) NOT NULL, -->
<!--     rate_change_percentage DECIMAL(8,4), -->
<!---->
<!--     -- Account balances -->
<!--     original_functional_balance DECIMAL(15,2), -->
<!--     revalued_functional_balance DECIMAL(15,2), -->
<!--     revaluation_adjustment DECIMAL(15,2), -->
<!---->
<!--     -- Impact summary -->
<!--     total_gain_loss DECIMAL(15,2), -->
<!--     unrealized_gain_loss DECIMAL(15,2), -->
<!---->
<!--     -- Affected accounts -->
<!--     affected_account_count INTEGER DEFAULT 0, -->
<!--     affected_transaction_count INTEGER DEFAULT 0, -->
<!---->
<!--     -- GL posting -->
<!--     journal_entry_id UUID REFERENCES finance_journal_entries(id), -->
<!--     posted_to_gl BOOLEAN DEFAULT false, -->
<!---->
<!--     -- Process metadata -->
<!--     revaluation_method VARCHAR(30) DEFAULT 'current_rate', -- current_rate, temporal, monetary_nonmonetary -->
<!--     processed_by UUID REFERENCES users(id), -->
<!--     processing_duration_ms INTEGER, -->
<!---->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!---->
<!--     CONSTRAINT valid_revaluation_method CHECK (revaluation_method IN ('current_rate', 'temporal', 'monetary_nonmonetary')) -->
<!-- ); -->
<!-- ``` -->
<!---->
<!-- ### Automated Currency Revaluation -->
<!---->
<!-- ```typescript -->
<!-- interface CurrencyRevaluationEngine { -->
<!--   performRevaluation(request: RevaluationRequest): Promise<RevaluationResult>; -->
<!--   calculateUnrealizedGainLoss(currency: string, asOfDate: Date): Promise<GainLossCalculation>; -->
<!--   schedulePeriodicRevaluations(tenantId: string): Promise<ScheduleResult>; -->
<!--   generateRevaluationReport(revaluationId: string): Promise<RevaluationReport>; -->
<!-- } -->
<!---->
<!-- interface RevaluationRequest { -->
<!--   tenant_id: string; -->
<!--   revaluation_date: Date; -->
<!--   currencies: string[]; // Empty array = all currencies -->
<!--   revaluation_method: 'current_rate' | 'temporal' | 'monetary_nonmonetary'; -->
<!--   rate_source?: string; -->
<!--   accounts_filter?: AccountFilter; -->
<!--   dry_run?: boolean; -->
<!-- } -->
<!---->
<!-- interface RevaluationResult { -->
<!--   revaluation_batch_id: string; -->
<!--   processed_currencies: string[]; -->
<!--   total_gain_loss: number; -->
<!--   unrealized_gain_loss: number; -->
<!--   affected_accounts: number; -->
<!--   affected_transactions: number; -->
<!--   journal_entries_created: string[]; -->
<!--   processing_summary: ProcessingSummary; -->
<!--   warnings: string[]; -->
<!--   errors: string[]; -->
<!-- } -->
<!---->
<!-- class CurrencyRevaluationService implements CurrencyRevaluationEngine { -->
<!--   async performRevaluation(request: RevaluationRequest): Promise<RevaluationResult> { -->
<!--     const batchId = this.generateBatchId(); -->
<!--     const result: RevaluationResult = { -->
<!--       revaluation_batch_id: batchId, -->
<!--       processed_currencies: [], -->
<!--       total_gain_loss: 0, -->
<!--       unrealized_gain_loss: 0, -->
<!--       affected_accounts: 0, -->
<!--       affected_transactions: 0, -->
<!--       journal_entries_created: [], -->
<!--       processing_summary: { -->
<!--         start_time: new Date(), -->
<!--         end_time: null, -->
<!--         duration_ms: 0 -->
<!--       }, -->
<!--       warnings: [], -->
<!--       errors: [] -->
<!--     }; -->
<!---->
<!--     try { -->
<!--       // Get currencies to revalue -->
<!--       const currencies = request.currencies.length > 0  -->
<!--         ? request.currencies  -->
<!--         : await this.getActiveCurrencies(request.tenant_id); -->
<!---->
<!--       // Get functional currency -->
<!--       const functionalCurrency = await this.getFunctionalCurrency(request.tenant_id); -->
<!---->
<!--       for (const currency of currencies) { -->
<!--         if (currency === functionalCurrency) { -->
<!--           continue; // Skip functional currency -->
<!--         } -->
<!---->
<!--         try { -->
<!--           const currencyResult = await this.revalueCurrency( -->
<!--             request.tenant_id, -->
<!--             currency, -->
<!--             functionalCurrency, -->
<!--             request.revaluation_date, -->
<!--             request.revaluation_method, -->
<!--             batchId, -->
<!--             request.dry_run -->
<!--           ); -->
<!---->
<!--           result.processed_currencies.push(currency); -->
<!--           result.total_gain_loss += currencyResult.total_gain_loss; -->
<!--           result.unrealized_gain_loss += currencyResult.unrealized_gain_loss; -->
<!--           result.affected_accounts += currencyResult.affected_accounts; -->
<!--           result.affected_transactions += currencyResult.affected_transactions; -->
<!---->
<!--           if (currencyResult.journal_entry_id && !request.dry_run) { -->
<!--             result.journal_entries_created.push(currencyResult.journal_entry_id); -->
<!--           } -->
<!---->
<!--           if (currencyResult.warnings.length > 0) { -->
<!--             result.warnings.push(...currencyResult.warnings.map(w => `${currency}: ${w}`)); -->
<!--           } -->
<!---->
<!--         } catch (error) { -->
<!--           result.errors.push(`Failed to revalue ${currency}: ${error.message}`); -->
<!--           continue; -->
<!--         } -->
<!--       } -->
<!---->
<!--       result.processing_summary.end_time = new Date(); -->
<!--       result.processing_summary.duration_ms =  -->
<!--         result.processing_summary.end_time.getTime() - result.processing_summary.start_time.getTime(); -->
<!---->
<!--       return result; -->
<!---->
<!--     } catch (error) { -->
<!--       result.errors.push(`Revaluation failed: ${error.message}`); -->
<!--       return result; -->
<!--     } -->
<!--   } -->
<!---->
<!--   private async revalueCurrency( -->
<!--     tenantId: string, -->
<!--     currency: string, -->
<!--     functionalCurrency: string, -->
<!--     revaluationDate: Date, -->
<!--     method: string, -->
<!--     batchId: string, -->
<!--     dryRun: boolean = false -->
<!--   ): Promise<CurrencyRevaluationResult> { -->
<!---->
<!--     // Get current exchange rate -->
<!--     const currentRate = await this.currencyService.getExchangeRate( -->
<!--       currency,  -->
<!--       functionalCurrency,  -->
<!--       revaluationDate -->
<!--     ); -->
<!---->
<!--     // Get previous revaluation rate -->
<!--     const previousRate = await this.getPreviousRevaluationRate( -->
<!--       tenantId,  -->
<!--       currency,  -->
<!--       functionalCurrency -->
<!--     ); -->
<!---->
<!--     // Get all open balances in this currency -->
<!--     const openBalances = await this.getOpenCurrencyBalances( -->
<!--       tenantId,  -->
<!--       currency,  -->
<!--       revaluationDate -->
<!--     ); -->
<!---->
<!--     let totalGainLoss = 0; -->
<!--     let affectedAccounts = 0; -->
<!--     let affectedTransactions = 0; -->
<!--     const revaluationEntries: RevaluationEntry[] = []; -->
<!---->
<!--     for (const balance of openBalances) { -->
<!--       // Calculate revaluation adjustment -->
<!--       const originalFunctionalAmount = balance.transaction_amount * balance.original_rate; -->
<!--       const revaluedFunctionalAmount = balance.transaction_amount * currentRate.rate; -->
<!--       const adjustment = revaluedFunctionalAmount - originalFunctionalAmount; -->
<!---->
<!--       if (Math.abs(adjustment) > 0.01) { // Only process significant adjustments -->
<!--         totalGainLoss += adjustment; -->
<!--         affectedTransactions++; -->
<!---->
<!--         revaluationEntries.push({ -->
<!--           transaction_id: balance.transaction_id, -->
<!--           account_id: balance.account_id, -->
<!--           currency: currency, -->
<!--           transaction_amount: balance.transaction_amount, -->
<!--           original_rate: balance.original_rate, -->
<!--           revaluation_rate: currentRate.rate, -->
<!--           adjustment_amount: adjustment, -->
<!--           adjustment_type: adjustment >= 0 ? 'gain' : 'loss' -->
<!--         }); -->
<!--       } -->
<!--     } -->
<!---->
<!--     // Group adjustments by account -->
<!--     const accountAdjustments = this.groupAdjustmentsByAccount(revaluationEntries); -->
<!--     affectedAccounts = Object.keys(accountAdjustments).length; -->
<!---->
<!--     let journalEntryId: string | null = null; -->
<!---->
<!--     if (!dryRun && totalGainLoss !== 0) { -->
<!--       // Create revaluation journal entry -->
<!--       journalEntryId = await this.createRevaluationJournalEntry( -->
<!--         tenantId, -->
<!--         currency, -->
<!--         functionalCurrency, -->
<!--         revaluationDate, -->
<!--         accountAdjustments, -->
<!--         totalGainLoss, -->
<!--         batchId -->
<!--       ); -->
<!---->
<!--       // Update currency transaction records -->
<!--       await this.updateCurrencyTransactionRates(revaluationEntries, currentRate.rate); -->
<!--     } -->
<!---->
<!--     // Record revaluation history -->
<!--     if (!dryRun) { -->
<!--       await this.recordRevaluationHistory({ -->
<!--         tenant_id: tenantId, -->
<!--         revaluation_date: revaluationDate, -->
<!--         revaluation_batch_id: batchId, -->
<!--         currency_code: currency, -->
<!--         functional_currency: functionalCurrency, -->
<!--         revaluation_rate: currentRate.rate, -->
<!--         previous_rate: previousRate, -->
<!--         total_gain_loss: totalGainLoss, -->
<!--         affected_account_count: affectedAccounts, -->
<!--         affected_transaction_count: affectedTransactions, -->
<!--         journal_entry_id: journalEntryId -->
<!--       }); -->
<!--     } -->
<!---->
<!--     return { -->
<!--       currency: currency, -->
<!--       total_gain_loss: totalGainLoss, -->
<!--       unrealized_gain_loss: totalGainLoss, // All unrealized for open positions -->
<!--       affected_accounts: affectedAccounts, -->
<!--       affected_transactions: affectedTransactions, -->
<!--       journal_entry_id: journalEntryId, -->
<!--       warnings: [] -->
<!--     }; -->
<!--   } -->
<!---->
<!--   private async createRevaluationJournalEntry( -->
<!--     tenantId: string, -->
<!--     currency: string, -->
<!--     functionalCurrency: string, -->
<!--     revaluationDate: Date, -->
<!--     accountAdjustments: Record<string, number>, -->
<!--     totalGainLoss: number, -->
<!--     batchId: string -->
<!--   ): Promise<string> { -->
<!---->
<!--     const journalEntries: JournalEntryLine[] = []; -->
<!---->
<!--     // Create adjustment entries for each affected account -->
<!--     for (const [accountId, adjustment] of Object.entries(accountAdjustments)) { -->
<!--       if (adjustment >= 0) { -->
<!--         // Gain - debit the original account -->
<!--         journalEntries.push({ -->
<!--           account_id: accountId, -->
<!--           debit_amount: Math.abs(adjustment), -->
<!--           credit_amount: 0, -->
<!--           description: `Currency revaluation gain - ${currency}` -->
<!--         }); -->
<!--       } else { -->
<!--         // Loss - credit the original account -->
<!--         journalEntries.push({ -->
<!--           account_id: accountId, -->
<!--           debit_amount: 0, -->
<!--           credit_amount: Math.abs(adjustment), -->
<!--           description: `Currency revaluation loss - ${currency}` -->
<!--         }); -->
<!--       } -->
<!--     } -->
<!---->
<!--     // Create offsetting entry to unrealized gain/loss account -->
<!--     const unrealizedGLAccount = await this.getUnrealizedGainLossAccount(tenantId); -->
<!---->
<!--     if (totalGainLoss >= 0) { -->
<!--       // Net gain - credit unrealized gain account -->
<!--       journalEntries.push({ -->
<!--         account_id: unrealizedGLAccount, -->
<!--         debit_amount: 0, -->
<!--         credit_amount: totalGainLoss, -->
<!--         description: `Unrealized currency gain - ${currency} revaluation` -->
<!--       }); -->
<!--     } else { -->
<!--       // Net loss - debit unrealized loss account -->
<!--       journalEntries.push({ -->
<!--         account_id: unrealizedGLAccount, -->
<!--         debit_amount: Math.abs(totalGainLoss), -->
<!--         credit_amount: 0, -->
<!--         description: `Unrealized currency loss - ${currency} revaluation` -->
<!--       }); -->
<!--     } -->
<!---->
<!--     // Create journal entry -->
<!--     const journalEntry = await this.journalService.createJournalEntry({ -->
<!--       tenant_id: tenantId, -->
<!--       entry_date: revaluationDate, -->
<!--       reference: `CRV-${batchId}-${currency}`, -->
<!--       description: `Currency revaluation - ${currency}`, -->
<!--       entries: journalEntries, -->
<!--       entry_type: 'currency_revaluation', -->
<!--       source_module: 'currency_management' -->
<!--     }); -->
<!---->
<!--     return journalEntry.id; -->
<!--   } -->
<!-- } -->
<!-- ``` -->
<!---->
<!-- ##  Currency Risk Management -->
<!---->
<!-- ### Hedging and Risk Management -->
<!---->
<!-- ```sql -->
<!-- -- Foreign exchange hedge contracts -->
<!-- CREATE TABLE fx_hedge_contracts ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE, -->
<!---->
<!--     -- Contract identification -->
<!--     contract_number VARCHAR(50) UNIQUE NOT NULL, -->
<!--     contract_type VARCHAR(30) NOT NULL, -- forward, option, swap, collar -->
<!---->
<!--     -- Counterparty -->
<!--     bank_counterparty_id UUID REFERENCES vendors(id), -->
<!--     counterparty_name VARCHAR(255), -->
<!---->
<!--     -- Currency details -->
<!--     base_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code), -->
<!--     quote_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code), -->
<!---->
<!--     -- Contract terms -->
<!--     notional_amount DECIMAL(18,2) NOT NULL, -->
<!--     contract_rate DECIMAL(18,8) NOT NULL, -->
<!---->
<!--     -- Contract dates -->
<!--     trade_date DATE NOT NULL, -->
<!--     value_date DATE NOT NULL, -- Settlement date -->
<!--     maturity_date DATE NOT NULL, -->
<!---->
<!--     -- Option-specific fields -->
<!--     option_type VARCHAR(10), -- call, put -->
<!--     strike_rate DECIMAL(18,8), -->
<!--     premium_amount DECIMAL(15,2), -->
<!--     premium_currency CHAR(3) REFERENCES currencies(currency_code), -->
<!---->
<!--     -- Collar-specific fields -->
<!--     cap_rate DECIMAL(18,8), -- Upper bound -->
<!--     floor_rate DECIMAL(18,8), -- Lower bound -->
<!---->
<!--     -- Hedge accounting -->
<!--     hedge_designation VARCHAR(30), -- cash_flow, fair_value, net_investment -->
<!--     hedged_item_type VARCHAR(50), -- forecast_transaction, commitment, recognized_asset, net_investment -->
<!--     hedge_effectiveness_method VARCHAR(30), -- dollar_offset, regression, critical_terms -->
<!---->
<!--     -- Risk metrics -->
<!--     delta DECIMAL(8,6), -- Price sensitivity -->
<!--     gamma DECIMAL(8,6), -- Delta sensitivity   -->
<!--     vega DECIMAL(8,6), -- Volatility sensitivity -->
<!--     theta DECIMAL(8,6), -- Time decay -->
<!---->
<!--     -- Current valuation -->
<!--     mark_to_market_value DECIMAL(15,2) DEFAULT 0, -->
<!--     unrealized_gain_loss DECIMAL(15,2) DEFAULT 0, -->
<!--     last_valuation_date DATE, -->
<!---->
<!--     -- Settlement -->
<!--     settlement_amount DECIMAL(15,2), -->
<!--     settlement_date DATE, -->
<!--     settlement_status VARCHAR(20) DEFAULT 'open', -- open, partially_settled, fully_settled, expired -->
<!---->
<!--     -- Status -->
<!--     contract_status VARCHAR(20) DEFAULT 'active', -- active, matured, cancelled, terminated -->
<!---->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!--     updated_at TIMESTAMPTZ DEFAULT NOW(), -->
<!--     created_by UUID REFERENCES users(id), -->
<!---->
<!--     CONSTRAINT valid_contract_type CHECK (contract_type IN ('forward', 'option', 'swap', 'collar')), -->
<!--     CONSTRAINT valid_option_type CHECK (option_type IS NULL OR option_type IN ('call', 'put')), -->
<!--     CONSTRAINT valid_hedge_designation CHECK (hedge_designation IN ('cash_flow', 'fair_value', 'net_investment')), -->
<!--     CONSTRAINT valid_settlement_status CHECK (settlement_status IN ('open', 'partially_settled', 'fully_settled', 'expired')), -->
<!--     CONSTRAINT valid_contract_status CHECK (contract_status IN ('active', 'matured', 'cancelled', 'terminated')), -->
<!--     CONSTRAINT positive_notional CHECK (notional_amount > 0), -->
<!--     CONSTRAINT different_hedge_currencies CHECK (base_currency != quote_currency) -->
<!-- ); -->
<!---->
<!-- -- Currency exposure analysis -->
<!-- CREATE TABLE currency_exposures ( -->
<!--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -->
<!--     tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE, -->
<!---->
<!--     -- Exposure identification -->
<!--     currency_code CHAR(3) NOT NULL REFERENCES currencies(currency_code), -->
<!--     functional_currency CHAR(3) NOT NULL REFERENCES currencies(currency_code), -->
<!--     exposure_date DATE NOT NULL, -->
<!---->
<!--     -- Exposure types -->
<!--     transaction_exposure DECIMAL(18,2) DEFAULT 0, -- Existing receivables/payables -->
<!--     translation_exposure DECIMAL(18,2) DEFAULT 0, -- Foreign subsidiary assets/liabilities   -->
<!--     economic_exposure DECIMAL(18,2) DEFAULT 0, -- Future cash flows -->
<!---->
<!--     -- Net exposure -->
<!--     gross_exposure DECIMAL(18,2) DEFAULT 0, -->
<!--     hedged_amount DECIMAL(18,2) DEFAULT 0, -->
<!--     net_exposure DECIMAL(18,2) DEFAULT 0, -->
<!---->
<!--     -- Risk metrics -->
<!--     value_at_risk DECIMAL(15,2), -- VaR at 95% confidence -->
<!--     expected_shortfall DECIMAL(15,2), -- Expected loss beyond VaR -->
<!--     volatility DECIMAL(8,6), -- Historical volatility -->
<!---->
<!--     -- Breakdown by time horizon -->
<!--     exposure_0_30_days DECIMAL(18,2) DEFAULT 0, -->
<!--     exposure_31_90_days DECIMAL(18,2) DEFAULT 0, -->
<!--     exposure_91_180_days DECIMAL(18,2) DEFAULT 0, -->
<!--     exposure_181_365_days DECIMAL(18,2) DEFAULT 0, -->
<!--     exposure_over_1_year DECIMAL(18,2) DEFAULT 0, -->
<!---->
<!--     -- Sensitivity analysis -->
<!--     sensitivity_1_percent DECIMAL(15,2), -- P&L impact of 1% currency move -->
<!--     sensitivity_5_percent DECIMAL(15,2), -->
<!--     sensitivity_10_percent DECIMAL(15,2), -->
<!---->
<!--     created_at TIMESTAMPTZ DEFAULT NOW(), -->
<!---->
<!--     UNIQUE(tenant_id, currency_code, exposure_date) -->
<!-- ); -->
<!-- ``` -->
<!---->
<!-- This  currency management module provides enterprise-level multi-currency capabilities with sophisticated risk management, automated revaluation, and compliance features essential for global financial operations. -->
<!---->
<!-- <function_calls> -->
<!-- <invoke name="TodoWrite"> -->
<!-- <parameter name="todos">[{"id": "analyze-current-schema", "content": "Analyze current financial.sql schema and identify core vs extended modules", "status": "completed", "priority": "high"}, {"id": "restructure-architecture", "content": "Design modular architecture with Finance, Buy, Sell, and misc modules", "status": "completed", "priority": "high"}, {"id": "update-documentation", "content": "Update financial documentation to reflect new modular structure", "status": "in_progress", "priority": "high"}, {"id": "create-currency-docs", "content": "Create  currency management documentation", "status": "completed", "priority": "medium"}] -->
