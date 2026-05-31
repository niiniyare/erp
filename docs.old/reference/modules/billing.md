# Billing Module — Complete Technical & Business Guide

> **Package:** `awo/internal/billing`  
> **Audience:** Platform engineers, finance team, product managers, and business stakeholders.  
> **Scope:** SaaS billing lifecycle, Lago-inspired metering primitives, charge models, pricing engine, automated GL entries, revenue recognition, platform self-accounting, SaaS metrics, and the data access boundary between platform operators and Awo's own tenant books.

---

## Table of Contents

1. [Billing Philosophy](#1-billing-philosophy)
2. [Architecture Overview](#2-architecture-overview)
3. [Lago Inspiration — What We Adopted and Why](#3-lago-inspiration--what-we-adopted-and-why)
4. [Database Schema](#4-database-schema)
   - 4.1 Billable Metrics
   - 4.2 Usage Events (Raw Stream)
   - 4.3 Charge Models
   - 4.4 Module Pricing
   - 4.5 Geographical Pricing
   - 4.6 Volume Discounts & Pricing Rules
   - 4.7 Subscriptions
   - 4.8 Invoices and Payments
   - 4.9 Wallets and Prepaid Credits
   - 4.10 One-Off Invoices
5. [Pricing Engine](#5-pricing-engine)
   - 5.1 Billable Metric Aggregation
   - 5.2 Charge Model Implementations
   - 5.3 Full Calculation Algorithm
   - 5.4 Dynamic Pricing Rules
   - 5.5 Proration Algorithm
6. [Event Ingestion](#6-event-ingestion)
7. [Subscription Lifecycle](#7-subscription-lifecycle)
8. [Invoice Generation](#8-invoice-generation)
9. [Automated GL Entries](#9-automated-gl-entries)
10. [Platform as Its Own ERP Tenant](#10-platform-as-its-own-erp-tenant)
11. [Platform User vs Platform Tenant — Data Access Boundary](#11-platform-user-vs-platform-tenant--data-access-boundary)
12. [SaaS Metrics & Reporting](#12-saas-metrics--reporting)
13. [Billing Services & Actions](#13-billing-services--actions)
14. [Payment Gateway Integration](#14-payment-gateway-integration)
15. [Testing Strategy](#15-testing-strategy)
16. [Troubleshooting](#16-troubleshooting)

---

## 1. Billing Philosophy

### 1.1 Chosen Strategy: Module Marketplace with Metered Charges

Awo ERP uses a **module marketplace** pricing model layered on top of a **metered billing engine**. Every tenant pays for exactly what they enable and exactly what they use:

```
Monthly Bill = Base Platform Fee
             + Σ apply_charge_model(enabled_module, usage)
             + Σ apply_charge_model(billable_metric, aggregated_usage)
             ± geographical adjustment factor
             - wallet credits applied
             - applicable discounts
             + tax
```

This replaced the earlier flat-rate overage approach after studying Lago's architecture. The key insight from Lago: **separating what you measure (billable metrics) from how you price it (charge models)** lets you add new pricing dimensions with zero code changes. A new billable metric is a database row. A new charge model structure is a JSON configuration. Neither requires a deployment.

### 1.2 Three Pricing Principles

**Transparency:** A tenant can calculate their next invoice from publicly listed prices. The self-serve portal shows a live running total updated the moment a module is enabled or disabled.

**Alignment:** When a tenant grows and enables more modules or generates more usage, your revenue grows automatically. This is the correct SaaS incentive structure.

**Simplicity floor:** Every added pricing dimension must justify itself with meaningful revenue impact. Complexity is a cost paid on every support ticket and every billing dispute.

### 1.3 Pricing Model Components

| Component | Mechanism | Varies By |
|---|---|---|
| Base platform fee | Fixed monthly | Nothing — flat |
| Module access | Volume charge model | Module + module count |
| User seats | Graduated charge model | Active user headcount |
| Storage | Weighted sum + standard model | GB-months |
| Email/SMS delivery | Package charge model | Message blocks |
| Geographical adjustment | Multiplier on subtotal | Country/region |
| Annual billing discount | 15% off monthly subtotal | Billing cycle |
| Enterprise commitment | Spending minimum | Negotiated floor |
| Wallets/Credits | Deducted from invoice total | Prepaid or promotional |

---

## 2. Architecture Overview

### 2.1 Module Boundaries

```
awo/internal/billing/
├── service.go              # BillingService facade — single import for other modules
├── events.go               # EventIngestService — usage event ingestion with idempotency
├── aggregation.go          # UsageAggregator — aggregates events per metric per period
├── pricing.go              # PricingEngine — charge model evaluator, full bill calculator
├── subscription.go         # SubscriptionService — lifecycle management
├── invoice.go              # InvoiceService — generation, one-off invoices, PDF
├── payment.go              # PaymentService — recording receipts, wallet deductions
├── dunning.go              # DunningService — failed payment recovery schedule
├── recognition.go          # RevenueRecognitionService — deferred revenue monthly job
├── metrics.go              # SaaSMetricsService — MRR, ARR, churn, unit economics
├── gl_bridge.go            # GLBridge — posts all billing events to Awo's own GL
├── domain/
│   ├── billable_metric.go
│   ├── charge.go
│   ├── subscription.go
│   ├── invoice.go
│   ├── payment.go
│   ├── wallet.go
│   └── events.go
└── repo/
    ├── metric_repo.go
    ├── event_repo.go
    ├── charge_repo.go
    ├── subscription_repo.go
    ├── invoice_repo.go
    ├── payment_repo.go
    └── wallet_repo.go
```

### 2.2 Data Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│  Application (any module) emits usage events                        │
│  POST /internal/billing/events                                      │
│    { transaction_id, code, tenant_id, timestamp, properties }       │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│  EventIngestService                                                 │
│  ├── Deduplicates on (tenant_id, transaction_id) — idempotent       │
│  ├── Validates code against billable_metrics catalogue              │
│  └── Appends to usage_events (immutable raw stream)                 │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                           ▼ (at invoice time)
┌─────────────────────────────────────────────────────────────────────┐
│  PricingEngine.Calculate()                                          │
│  ├── Base fee lookup                                                │
│  ├── For each enabled module:                                       │
│  │     look up module_pricing → apply volume charge model           │
│  ├── For each plan_charge:                                          │
│  │     UsageAggregator.Aggregate(metric, period)                    │
│  │       → count / sum / unique_count / max / latest / weighted_sum │
│  │     apply charge model (standard/graduated/package/percentage/…)  │
│  ├── Geographical multiplier                                        │
│  ├── Pricing rules (loyalty, startup programme, etc.)               │
│  ├── Annual discount                                                │
│  └── Wallet credits                                                 │
└──────────────────────────┬──────────────────────────────────────────┘
                           │
                  ┌────────┴────────┐
                  ▼                 ▼
         Invoice record         GL Bridge
         (billing tables)    → finance_transactions
                                WHERE tenant_id = AwoTenantID
```

### 2.3 The GLBridge Pattern

The billing module never writes to `finance_transactions` directly. It calls a thin `GLBridge` adapter over `finance.TransactionService.CreateFromSource()`, always targeting `AwoTenantID`:

```go
// awo/internal/billing/gl_bridge.go

// GLBridge posts all billing financial events to Awo's own GL.
// It is the only code path creating transactions with source_module = "billing".
type GLBridge struct {
    financeSvc *finance.TransactionService
    accounts   *AwoChartOfAccounts
}
```

---

## 3. Lago Inspiration — What We Adopted and Why

[Lago](https://github.com/getlago/lago) is an open-source metering and usage-based billing engine (AGPL-3.0). We studied its architecture and adopted its core concepts natively in Go. We did not embed or integrate Lago itself for three reasons: it is Ruby on Rails (incompatible with our Go binary), AGPL-3.0 requires publishing modifications in a commercial product, and it has no concept of GL integration — which is essential for us.

### 3.1 What We Adopted

**Billable metrics with aggregation types.** Lago separates *what to measure* from *how to price it*. We adopted this: a `billable_metrics` table defines the measurement (what event, what field, what aggregation function). A `plan_charges` table defines the pricing rule (what charge model, what tiers). Adding a new billable dimension is a database row, not a code change.

**Six charge models.** Lago supports standard, graduated, package, percentage, volume, and graduated percentage charge models. We implement all six. The graduated model alone enables seat pricing with volume discounts that scale fairly — users 1–5 free, 6–20 at KES 300, 21+ at KES 250.

**Idempotent event ingestion.** Lago uses caller-supplied `transaction_id` keys to deduplicate events. We adopted this pattern. The same event can be sent twice (network retry, webhook replay) without being double-counted. The idempotency key is a unique constraint on `(tenant_id, transaction_id)`.

**Six aggregation types.** COUNT, SUM, UNIQUE_COUNT, MAX, LATEST, and WEIGHTED_SUM cover every realistic billing scenario. WEIGHTED_SUM is particularly valuable for storage: charging by GB-months (how much storage × how long) rather than peak storage is fairer and more accurate.

**Prepaid wallets.** Lago's wallet/credits system enables enterprise prepaid deals and promotional credits. We adopted the schema and deduction logic.

**One-off invoices.** Lago supports immediate charges outside the billing cycle. We adopted this for professional services, setup fees, and manual adjustments.

### 3.2 What We Did Not Adopt

Lago has no GL integration — it assumes an external accounting system. Our `GLBridge` is entirely original and is the most important part of the billing module for our purposes. Lago's dunning, revenue recognition, and SaaS metrics were independently designed to fit our architecture.

---

## 4. Database Schema

### 4.1 Billable Metrics

The catalogue of measurable quantities in the system. Shared across all tenants — no `tenant_id`. Similar to Lago's `billable_metrics` concept.

```sql
CREATE TABLE billable_metrics (
  id                uuid   PRIMARY KEY DEFAULT gen_random_uuid(),
  code              text   UNIQUE NOT NULL,
  -- 'active_users' | 'storage_gb' | 'gl_transactions_posted'
  -- 'api_calls' | 'email_sends' | 'sms_sends'
  name              text   NOT NULL,
  description       text,
  aggregation_type  text   NOT NULL,
  -- 'count'          → number of matching events in period
  -- 'sum'            → sum of a numeric field across events
  -- 'unique_count'   → count of distinct field values
  -- 'max'            → maximum field value in period
  -- 'latest'         → last field value before period end
  -- 'weighted_sum'   → time-weighted sum (value × days_held / period_days)
  field_name        text,
  -- which properties key to aggregate (NULL for 'count')
  -- e.g. "gb" for storage_gb, "user_id" for active_users
  recurring         bool   NOT NULL DEFAULT false,
  -- false: resets to 0 each period (API calls, email sends)
  -- true:  carries forward as high watermark (peak storage, peak users)
  event_code        text   NOT NULL,
  -- the code field value that usage_events must carry to match this metric
  is_active         bool   NOT NULL DEFAULT true
);

-- Standard metric definitions (seeded at deployment)
INSERT INTO billable_metrics (code, name, aggregation_type, field_name, recurring, event_code)
VALUES
  ('active_users',          'Active Users',              'unique_count', 'user_id',    true,  'user_session'),
  ('storage_gb',            'Storage Used (GB)',          'weighted_sum', 'gb',         true,  'storage_snapshot'),
  ('gl_transactions_posted','GL Transactions Posted',     'count',         NULL,         false, 'transaction_posted'),
  ('api_calls',             'API Calls',                  'count',         NULL,         false, 'api_request'),
  ('email_sends',           'Email Sends',                'count',         NULL,         false, 'email_sent'),
  ('sms_sends',             'SMS Sends',                  'count',         NULL,         false, 'sms_sent');
```

### 4.2 Usage Events (Raw Immutable Stream)

Every measurable event is appended here. Never updated or deleted. Idempotency is guaranteed by the unique constraint on `(tenant_id, transaction_id)`.

```sql
CREATE TABLE usage_events (
  id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           uuid        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  billable_metric_id  uuid        NOT NULL REFERENCES billable_metrics(id),
  transaction_id      text        NOT NULL,
  -- Caller-supplied idempotency key. Format: "{source}:{source_id}"
  -- e.g. "session:sess_abc123" | "storage:snapshot_2025_01_31" | "tx:txn_uuid"
  code                text        NOT NULL,
  timestamp           timestamptz NOT NULL,
  -- When the event occurred (may differ from ingested_at for late arrivals)
  properties          jsonb       NOT NULL DEFAULT '{}',
  -- flexible payload: { "user_id": "usr_abc", "gb": 12.5, "module": "finance" }
  ingested_at         timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, transaction_id)
);

-- Primary query pattern: get events for a tenant+metric within a time window
CREATE INDEX idx_usage_events_tenant_metric_time
  ON usage_events(tenant_id, billable_metric_id, timestamp);

-- For UNIQUE_COUNT aggregation queries on properties
CREATE INDEX idx_usage_events_properties_gin
  ON usage_events USING gin(properties);
```

**How events are emitted from other modules:**

```go
// The Finance module emits this when a transaction is posted
billingClient.IngestEvent(ctx, billing.EventParams{
    TenantID:      session.TenantID,
    TransactionID: "tx:" + transaction.ID.String(),  // idempotency key
    Code:          "transaction_posted",
    Timestamp:     time.Now(),
    Properties:    map[string]any{"module": "finance"},
})

// The IAM module emits this on session creation for active_users tracking
billingClient.IngestEvent(ctx, billing.EventParams{
    TenantID:      session.TenantID,
    TransactionID: "session:" + session.ID.String(),
    Code:          "user_session",
    Timestamp:     time.Now(),
    Properties:    map[string]any{"user_id": session.UserID.String()},
})

// The storage monitoring job emits this daily
billingClient.IngestEvent(ctx, billing.EventParams{
    TenantID:      tenantID,
    TransactionID: "storage:" + tenantID.String() + ":" + date.Format("2006-01-02"),
    Code:          "storage_snapshot",
    Timestamp:     time.Now(),
    Properties:    map[string]any{"gb": currentStorageGB},
})
```

### 4.3 Charge Models

Links a billable metric to a pricing structure. The `properties` JSONB column carries the charge model configuration — its shape varies by `charge_model`.

```sql
CREATE TABLE plan_charges (
  id                  uuid   PRIMARY KEY DEFAULT gen_random_uuid(),
  billable_metric_id  uuid   NOT NULL REFERENCES billable_metrics(id),
  charge_model        text   NOT NULL,
  -- 'standard'         → flat rate per unit above free_units
  -- 'graduated'        → cumulative tiers with different per-unit rates
  -- 'package'          → charge per block of N units
  -- 'percentage'       → percentage of a base amount field
  -- 'volume'           → single rate for ALL units based on total volume tier reached
  -- 'spending_minimum' → enterprise minimum commitment floor
  properties          jsonb  NOT NULL DEFAULT '{}',
  -- charge model configuration (see Section 5.2 for each shape)
  pay_in_advance      bool   NOT NULL DEFAULT false,
  -- true: charge at start of period; false: charge at end based on actual usage
  invoiceable         bool   NOT NULL DEFAULT true,
  currency_code       char(3) NOT NULL DEFAULT 'KES',
  is_active           bool   NOT NULL DEFAULT true,
  sort_order          int    NOT NULL DEFAULT 999
);

-- Standard plan charges (seeded at deployment)
-- Users: graduated model (0-5 free, 6-20 at 300, 21-50 at 250, 51+ at 200)
INSERT INTO plan_charges (billable_metric_id, charge_model, properties)
SELECT id, 'graduated', '{
  "graduated_ranges": [
    { "from": 0,  "to": 5,    "flat_amount": "0", "per_unit_amount": "0" },
    { "from": 6,  "to": 20,   "flat_amount": "0", "per_unit_amount": "300" },
    { "from": 21, "to": 50,   "flat_amount": "0", "per_unit_amount": "250" },
    { "from": 51, "to": null, "flat_amount": "0", "per_unit_amount": "200" }
  ]
}'::jsonb FROM billable_metrics WHERE code = 'active_users';

-- Storage: weighted_sum metric → standard charge above free tier
INSERT INTO plan_charges (billable_metric_id, charge_model, properties)
SELECT id, 'standard', '{"unit_amount": "150", "free_units": 5}'::jsonb
FROM billable_metrics WHERE code = 'storage_gb';

-- Email: package model (1000 included free, then KES 50 per 100)
INSERT INTO plan_charges (billable_metric_id, charge_model, properties)
SELECT id, 'package', '{"package_size": 100, "amount": "50", "free_units": 1000}'::jsonb
FROM billable_metrics WHERE code = 'email_sends';
```

### 4.4 Module Pricing

Separate from charge models. Module access is a flat fee per module per month. The volume model is applied at the module level (the more modules enabled, the lower the per-module price).

```sql
CREATE TABLE module_pricing (
  id              uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id       uuid          NOT NULL REFERENCES modules(id) ON DELETE RESTRICT,
  currency_code   char(3)       NOT NULL DEFAULT 'KES',
  price_monthly   numeric(10,2) NOT NULL,
  price_annual    numeric(10,2) NOT NULL,
  effective_from  date          NOT NULL DEFAULT CURRENT_DATE,
  effective_until date,
  is_active       bool          NOT NULL DEFAULT true,
  UNIQUE (module_id, currency_code, effective_from)
);

-- Base platform fee
CREATE TABLE platform_base_fee (
  id              uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  currency_code   char(3)       NOT NULL DEFAULT 'KES',
  fee_monthly     numeric(10,2) NOT NULL,
  fee_annual      numeric(10,2) NOT NULL,
  effective_from  date          NOT NULL DEFAULT CURRENT_DATE,
  effective_until date,
  UNIQUE (currency_code, effective_from)
);

-- Per-tenant price overrides (grandfathering, negotiated deals)
CREATE TABLE tenant_price_overrides (
  id              uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       uuid          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  module_id       uuid          REFERENCES modules(id),   -- NULL = applies to all modules
  override_type   text          NOT NULL, -- 'locked' | 'discount_pct' | 'custom_price'
  locked_price    numeric(10,2),
  discount_pct    numeric(5,2),
  reason          text,                   -- 'grandfathered' | 'early_adopter' | 'negotiated'
  valid_until     date,
  created_by      uuid          REFERENCES users(id),
  created_at      timestamptz   NOT NULL DEFAULT now()
);
```

### 4.5 Geographical Pricing

```sql
CREATE TABLE geo_pricing (
  id              uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
  region_code     text         NOT NULL,  -- 'KE' | 'UG' | 'TZ' | 'NG' | 'ZA' | 'GLOBAL'
  currency_code   char(3)      NOT NULL,
  multiplier      numeric(5,4) NOT NULL DEFAULT 1.0000,
  base_fee_override numeric(10,2),        -- NULL = standard base fee × multiplier
  effective_from  date         NOT NULL DEFAULT CURRENT_DATE,
  effective_until date,
  UNIQUE (region_code, currency_code, effective_from)
);

INSERT INTO geo_pricing (region_code, currency_code, multiplier) VALUES
  ('KE', 'KES', 1.0000),   -- Kenya: full price
  ('UG', 'KES', 0.7500),   -- Uganda: 25% cheaper
  ('TZ', 'KES', 0.7500),   -- Tanzania: 25% cheaper
  ('NG', 'NGN', 1.2000),   -- Nigeria: 20% premium (larger market)
  ('ZA', 'ZAR', 1.5000),   -- South Africa: 50% premium
  ('GLOBAL', 'USD', 2.0000); -- USD pricing
```

### 4.6 Volume Discounts and Dynamic Pricing Rules

```sql
-- Volume discount: the more modules enabled, the lower the per-module price
-- This is now expressed as a 'volume' charge model on module access
-- (replaced the old volume_discount_rules table)

-- Dynamic pricing rules evaluated in priority order
CREATE TABLE pricing_rules (
  id              uuid  PRIMARY KEY DEFAULT gen_random_uuid(),
  name            text  NOT NULL,
  description     text,
  priority        int   NOT NULL DEFAULT 100,  -- lower = evaluated first
  condition_type  text  NOT NULL,
  -- 'module_count_gte'       → volume-based discount
  -- 'active_since_days_gte'  → loyalty discount
  -- 'region_in'              → market-specific promotion
  -- 'user_count_gte'         → enterprise scale discount
  -- 'billing_cycle_eq'       → reward annual commitment
  -- 'tag_in'                 → custom tenant tags (startup programme, etc.)
  condition_value jsonb NOT NULL,
  action_type     text  NOT NULL,
  -- 'discount_pct'   → percentage off
  -- 'discount_fixed' → fixed amount off
  -- 'free_months'    → N months free (applied as credit)
  action_value    jsonb NOT NULL,
  -- stacking: false = only the best matching rule applies; true = rules combine
  stacking        bool  NOT NULL DEFAULT false,
  is_active       bool  NOT NULL DEFAULT true,
  effective_from  date  NOT NULL DEFAULT CURRENT_DATE,
  effective_until date
);

-- Example rules
INSERT INTO pricing_rules (name, priority, condition_type, condition_value,
                            action_type, action_value) VALUES
  ('Loyalty 24-month', 50,
   'active_since_days_gte', '{"days": 730}',
   'discount_pct', '{"percentage": 10, "applies_to": "modules"}'),
  ('Startup programme', 30,
   'tag_in', '{"tags": ["startup"]}',
   'discount_pct', '{"percentage": 50, "applies_to": "all", "max_months": 6}'),
  ('Annual commitment bonus', 20,
   'billing_cycle_eq', '{"cycle": "annual"}',
   'discount_pct', '{"percentage": 15, "applies_to": "all"}');
```

### 4.7 Subscriptions

```sql
CREATE TABLE tenant_subscriptions (
  id                    uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             uuid          NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  status                text          NOT NULL DEFAULT 'trialing',
  -- 'trialing' | 'active' | 'past_due' | 'suspended' | 'cancelled'
  billing_cycle         text          NOT NULL DEFAULT 'monthly',
  region_code           text          NOT NULL DEFAULT 'KE',
  currency_code         char(3)       NOT NULL DEFAULT 'KES',
  trial_ends_at         timestamptz,
  current_period_start  date          NOT NULL,
  current_period_end    date          NOT NULL,
  next_invoice_date     date          NOT NULL,
  cancelled_at          timestamptz,
  cancellation_reason   text,
  dunning_attempts      int           NOT NULL DEFAULT 0,
  -- Payment gateway
  payment_customer_id   text,
  payment_method_id     text,
  billing_email         text,
  billing_name          text,
  created_at            timestamptz   NOT NULL DEFAULT now(),
  updated_at            timestamptz   NOT NULL DEFAULT now(),
  UNIQUE (tenant_id)
);

CREATE TABLE subscription_events (
  id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       uuid        NOT NULL REFERENCES tenants(id),
  subscription_id uuid        NOT NULL REFERENCES tenant_subscriptions(id),
  event_type      text        NOT NULL,
  -- 'trial_started' | 'activated' | 'module_enabled' | 'module_disabled'
  -- 'upgraded' | 'downgraded' | 'payment_failed' | 'suspended'
  -- 'cancelled' | 'reactivated'
  event_data      jsonb       NOT NULL DEFAULT '{}',
  occurred_at     timestamptz NOT NULL DEFAULT now(),
  actor_user_id   uuid        REFERENCES users(id)
);

-- Pending proration lines added between invoices (settled on next invoice)
CREATE TABLE pending_proration_lines (
  id              uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       uuid          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  description     text          NOT NULL,
  amount          numeric(10,2) NOT NULL,
  period_start    date          NOT NULL,
  period_end      date          NOT NULL,
  module_id       uuid          REFERENCES modules(id),
  created_at      timestamptz   NOT NULL DEFAULT now(),
  invoiced_at     timestamptz   -- NULL until included on an invoice
);

-- Bundle presets (suggested starting configurations)
CREATE TABLE subscription_bundles (
  id                    uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  slug                  text          UNIQUE NOT NULL,
  name                  text          NOT NULL,
  description           text,
  included_module_ids   uuid[]        NOT NULL,
  bundle_discount_pct   numeric(5,2)  NOT NULL DEFAULT 0,
  is_featured           bool          NOT NULL DEFAULT false,
  sort_order            int           NOT NULL DEFAULT 999,
  is_active             bool          NOT NULL DEFAULT true
);
```

### 4.8 Invoices and Payments

```sql
CREATE TABLE subscription_invoices (
  id                uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id         uuid          NOT NULL REFERENCES tenants(id),
  subscription_id   uuid          NOT NULL REFERENCES tenant_subscriptions(id),
  invoice_number    text          UNIQUE NOT NULL,
  invoice_type      text          NOT NULL DEFAULT 'subscription',
  -- 'subscription' | 'one_off' | 'credit_note'
  status            text          NOT NULL DEFAULT 'draft',
  -- 'draft' | 'open' | 'paid' | 'void' | 'uncollectable'
  period_start      date          NOT NULL,
  period_end        date          NOT NULL,
  issue_date        date          NOT NULL,
  due_date          date          NOT NULL,
  subtotal          numeric(10,2) NOT NULL,
  discount_amount   numeric(10,2) NOT NULL DEFAULT 0,
  tax_amount        numeric(10,2) NOT NULL DEFAULT 0,
  total_amount      numeric(10,2) NOT NULL,
  amount_paid       numeric(10,2) NOT NULL DEFAULT 0,
  currency_code     char(3)       NOT NULL,
  gl_transaction_id uuid          REFERENCES finance_transactions(id),
  payment_intent_id text,
  paid_at           timestamptz,
  reminder_count    int           NOT NULL DEFAULT 0,
  last_reminder_at  timestamptz,
  created_at        timestamptz   NOT NULL DEFAULT now(),
  updated_at        timestamptz   NOT NULL DEFAULT now()
);

CREATE TABLE subscription_invoice_lines (
  id              uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id      uuid          NOT NULL REFERENCES subscription_invoices(id) ON DELETE CASCADE,
  line_type       text          NOT NULL,
  -- 'base_fee' | 'module' | 'usage_charge' | 'proration' | 'discount'
  -- 'credit' | 'spending_minimum' | 'wallet_credit'
  module_id       uuid          REFERENCES modules(id),
  metric_code     text,                    -- references billable_metrics.code
  charge_model    text,                    -- which model produced this line
  description     text          NOT NULL,
  quantity        numeric(15,4) NOT NULL DEFAULT 1,
  unit_price      numeric(10,4) NOT NULL,
  amount          numeric(10,2) NOT NULL,
  period_start    date,
  period_end      date,
  sort_order      int           NOT NULL DEFAULT 999
);

CREATE TABLE subscription_payments (
  id                uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id        uuid          NOT NULL REFERENCES subscription_invoices(id),
  tenant_id         uuid          NOT NULL REFERENCES tenants(id),
  amount            numeric(10,2) NOT NULL,
  currency_code     char(3)       NOT NULL,
  payment_method    text,
  payment_reference text,
  gateway_fee       numeric(10,2) NOT NULL DEFAULT 0,
  net_received      numeric(10,2) NOT NULL,
  paid_at           timestamptz   NOT NULL,
  gl_transaction_id uuid          REFERENCES finance_transactions(id),
  recorded_by       uuid          REFERENCES users(id)
);

CREATE TABLE credit_notes (
  id                    uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             uuid          NOT NULL REFERENCES tenants(id),
  invoice_id            uuid          REFERENCES subscription_invoices(id),
  credit_number         text          UNIQUE NOT NULL,
  reason                text          NOT NULL,
  amount                numeric(10,2) NOT NULL,
  currency_code         char(3)       NOT NULL,
  status                text          NOT NULL DEFAULT 'open',
  -- 'open' | 'applied' | 'voided'
  applied_to_invoice_id uuid          REFERENCES subscription_invoices(id),
  gl_transaction_id     uuid          REFERENCES finance_transactions(id),
  created_at            timestamptz   NOT NULL DEFAULT now()
);

CREATE SEQUENCE subscription_invoice_seq START 1;
CREATE OR REPLACE FUNCTION next_invoice_number() RETURNS text AS $$
  SELECT 'AWO-' || TO_CHAR(NOW(), 'YYYY') || '-' ||
         LPAD(nextval('subscription_invoice_seq')::text, 5, '0');
$$ LANGUAGE sql;
```

### 4.9 Wallets and Prepaid Credits

Inspired by Lago's wallet concept. Enables enterprise prepaid deals and promotional credits.

```sql
CREATE TABLE tenant_wallets (
  id              uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       uuid          NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  currency_code   char(3)       NOT NULL,
  balance         numeric(15,2) NOT NULL DEFAULT 0,
  expiry_at       timestamptz,             -- NULL = never expires
  status          text          NOT NULL DEFAULT 'active',
  -- 'active' | 'expired' | 'terminated'
  created_at      timestamptz   NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, currency_code)
);

CREATE TABLE wallet_transactions (
  id               uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  wallet_id        uuid          NOT NULL REFERENCES tenant_wallets(id),
  amount           numeric(15,2) NOT NULL,
  -- positive = credit added; negative = credit consumed
  transaction_type text          NOT NULL,
  -- 'purchase' | 'promo_credit' | 'invoice_deduction' | 'expiry_void' | 'manual_adjustment'
  source           text          NOT NULL,
  invoice_id       uuid          REFERENCES subscription_invoices(id),
  description      text,
  created_at       timestamptz   NOT NULL DEFAULT now()
);
```

### 4.10 One-Off Invoices

Lago calls these "one-off charges." Used for professional services, setup fees, manual adjustments, and any charge that should be invoiced immediately rather than waiting for the billing cycle.

```sql
CREATE TABLE one_off_invoices (
  id              uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       uuid          NOT NULL REFERENCES tenants(id),
  invoice_number  text          UNIQUE NOT NULL,
  status          text          NOT NULL DEFAULT 'draft',
  issue_date      date          NOT NULL,
  due_date        date          NOT NULL,
  description     text          NOT NULL,
  subtotal        numeric(10,2) NOT NULL,
  tax_amount      numeric(10,2) NOT NULL DEFAULT 0,
  total_amount    numeric(10,2) NOT NULL,
  gl_transaction_id uuid        REFERENCES finance_transactions(id),
  created_by      uuid          NOT NULL REFERENCES users(id),
  created_at      timestamptz   NOT NULL DEFAULT now()
);
```

---

## 5. Pricing Engine

### 5.1 Billable Metric Aggregation

The `UsageAggregator` computes the billable quantity for a given metric over a period. All aggregations query `usage_events`.

```go
// awo/internal/billing/aggregation.go

type UsageAggregator struct {
    repo EventRepository
}

func (a *UsageAggregator) Aggregate(ctx context.Context,
    params AggregationParams) (decimal.Decimal, error) {

    metric := params.Metric

    switch metric.AggregationType {

    case "count":
        // Number of matching events in the period
        n, err := a.repo.CountEvents(ctx, params.TenantID, metric.ID,
            params.PeriodStart, params.PeriodEnd)
        return decimal.NewFromInt(n), err

    case "sum":
        // Sum a numeric property across all events
        // e.g. sum of "gb" field in storage events
        return a.repo.SumEventField(ctx, params.TenantID, metric.ID,
            metric.FieldName, params.PeriodStart, params.PeriodEnd)

    case "unique_count":
        // Count distinct values of a property
        // e.g. count distinct "user_id" values for active user billing
        n, err := a.repo.CountUniqueField(ctx, params.TenantID, metric.ID,
            metric.FieldName, params.PeriodStart, params.PeriodEnd)
        return decimal.NewFromInt(n), err

    case "max":
        // Maximum value of a property in the period
        // e.g. peak concurrent sessions
        return a.repo.MaxEventField(ctx, params.TenantID, metric.ID,
            metric.FieldName, params.PeriodStart, params.PeriodEnd)

    case "latest":
        // Last value before period end
        // e.g. final storage snapshot of the month
        return a.repo.LatestEventField(ctx, params.TenantID, metric.ID,
            metric.FieldName, params.PeriodEnd)

    case "weighted_sum":
        // Time-weighted sum: value × (days held / total days in period)
        // Charges storage fairly based on how long data was held
        //
        // Example: 10 GB held from Jan 1, +5 GB added Jan 15 (31-day month)
        //   10 GB × 31 days = 310 GB-days
        //   5  GB × 17 days = 85 GB-days
        //   Total: 395 GB-days / 31 = 12.74 GB-months
        //   Charge: 12.74 × rate  (fairer than billing peak 15 GB)
        return a.repo.WeightedSumEventField(ctx, params.TenantID, metric.ID,
            metric.FieldName, params.PeriodStart, params.PeriodEnd)
    }

    return decimal.Zero, fmt.Errorf("unknown aggregation type: %s", metric.AggregationType)
}
```

**The weighted_sum SQL query:**

```sql
-- awo/db/queries/aggregation.sql
-- name: WeightedSumEventField :one
WITH events_with_duration AS (
  SELECT
    (properties ->> @field_name)::numeric AS value,
    timestamp AS started_at,
    LEAD(timestamp) OVER (ORDER BY timestamp)  AS ended_at,
    @period_end::timestamptz                   AS period_end
  FROM usage_events
  WHERE tenant_id          = @tenant_id
    AND billable_metric_id = @metric_id
    AND timestamp BETWEEN @period_start AND @period_end
)
SELECT
  COALESCE(
    SUM(
      value * EXTRACT(EPOCH FROM (COALESCE(ended_at, period_end) - started_at)) /
              EXTRACT(EPOCH FROM (@period_end::timestamptz - @period_start::timestamptz))
    ), 0
  ) AS weighted_sum
FROM events_with_duration;
```

### 5.2 Charge Model Implementations

All charge models are pure functions. They take the aggregated usage quantity and charge model properties, return a monetary amount and description. No DB calls inside these functions.

```go
// awo/internal/billing/pricing.go

// Standard: flat rate per unit above a free tier
// properties: { "unit_amount": "300", "free_units": 5 }
func applyStandard(props json.RawMessage, units decimal.Decimal) (decimal.Decimal, string) {
    var cfg struct {
        UnitAmount string `json:"unit_amount"`
        FreeUnits  int    `json:"free_units"`
    }
    json.Unmarshal(props, &cfg)

    free := decimal.NewFromInt(int64(cfg.FreeUnits))
    if units.LessThanOrEqual(free) {
        return decimal.Zero, fmt.Sprintf("%.0f units (within free tier)", units.InexactFloat64())
    }
    rate, _ := decimal.NewFromString(cfg.UnitAmount)
    chargeable := units.Sub(free)
    return rate.Mul(chargeable).Round(2),
        fmt.Sprintf("%.0f units (%.0f above free tier of %d)", units.InexactFloat64(),
            chargeable.InexactFloat64(), cfg.FreeUnits)
}

// Graduated: cumulative tiers — each tier has its own per-unit rate
// The first N units are charged at tier 1 rate, the next M at tier 2, etc.
// properties: { "graduated_ranges": [ { "from", "to", "flat_amount", "per_unit_amount" } ] }
//
// Example: users 0-5 free, 6-20 at KES 300, 21+ at KES 250
// For 25 users:
//   Tier 1 (0-5):   5 × 0     = 0
//   Tier 2 (6-20): 15 × 300   = 4,500
//   Tier 3 (21-25): 5 × 250   = 1,250
//   Total: KES 5,750
func applyGraduated(props json.RawMessage, units decimal.Decimal) (decimal.Decimal, string) {
    var cfg struct {
        Ranges []struct {
            From          int64   `json:"from"`
            To            *int64  `json:"to"`
            FlatAmount    string  `json:"flat_amount"`
            PerUnitAmount string  `json:"per_unit_amount"`
        } `json:"graduated_ranges"`
    }
    json.Unmarshal(props, &cfg)

    total, remaining := decimal.Zero, units
    for _, tier := range cfg.Ranges {
        if remaining.IsZero() { break }

        var tierSize decimal.Decimal
        if tier.To != nil {
            capacity := decimal.NewFromInt(*tier.To - tier.From + 1)
            tierSize = decimal.Min(remaining, capacity)
        } else {
            tierSize = remaining
        }

        flat, _   := decimal.NewFromString(tier.FlatAmount)
        perUnit, _ := decimal.NewFromString(tier.PerUnitAmount)
        total    = total.Add(flat).Add(perUnit.Mul(tierSize))
        remaining = remaining.Sub(tierSize)
    }
    return total.Round(2), fmt.Sprintf("%.0f units (graduated tiers)", units.InexactFloat64())
}

// Package: charge per block of N units
// First M units free, then KES X per block of N
// properties: { "package_size": 100, "amount": "50", "free_units": 1000 }
//
// Example: 1,000 free emails, then KES 50 per 100.
// For 1,350 emails: 350 above free → ceil(350/100) = 4 blocks → KES 200
func applyPackage(props json.RawMessage, units decimal.Decimal) (decimal.Decimal, string) {
    var cfg struct {
        PackageSize int    `json:"package_size"`
        Amount      string `json:"amount"`
        FreeUnits   int    `json:"free_units"`
    }
    json.Unmarshal(props, &cfg)

    free := decimal.NewFromInt(int64(cfg.FreeUnits))
    if units.LessThanOrEqual(free) { return decimal.Zero, "" }

    rate, _      := decimal.NewFromString(cfg.Amount)
    pkgSize      := decimal.NewFromInt(int64(cfg.PackageSize))
    chargeable   := units.Sub(free)
    blocks       := chargeable.Div(pkgSize).Ceil()
    return rate.Mul(blocks).Round(2),
        fmt.Sprintf("%.0f units → %s blocks of %d", units.InexactFloat64(),
            blocks.String(), cfg.PackageSize)
}

// Volume: single rate for ALL units based on which tier the total falls in
// Unlike graduated, the rate applies to every unit (not just units in that tier)
// properties: { "volume_ranges": [ { "from", "to", "per_unit_amount" } ] }
//
// Example: modules
//   1-3 modules:  KES 5,000 each
//   4-6 modules:  KES 4,500 each  ← activating 4th module reduces price of first 3 too
//   7+ modules:   KES 4,000 each
// For 5 modules: all 5 × KES 4,500 = KES 22,500
func applyVolume(props json.RawMessage, units decimal.Decimal) (decimal.Decimal, string) {
    var cfg struct {
        Ranges []struct {
            From          int64   `json:"from"`
            To            *int64  `json:"to"`
            PerUnitAmount string  `json:"per_unit_amount"`
        } `json:"volume_ranges"`
    }
    json.Unmarshal(props, &cfg)

    for _, tier := range cfg.Ranges {
        from := decimal.NewFromInt(tier.From)
        inTier := units.GreaterThanOrEqual(from)
        if tier.To != nil {
            inTier = inTier && units.LessThanOrEqual(decimal.NewFromInt(*tier.To))
        }
        if inTier {
            rate, _ := decimal.NewFromString(tier.PerUnitAmount)
            total   := rate.Mul(units).Round(2)
            return total, fmt.Sprintf("%.0f units at %s/unit (volume tier)",
                units.InexactFloat64(), tier.PerUnitAmount)
        }
    }
    return decimal.Zero, ""
}

// Percentage: charge a % of a base amount field in the event properties
// Useful for transaction-value based billing (charge 0.5% of invoice value posted)
// properties: { "percentage": "0.5", "base_field": "invoice_amount", "free_units": 0 }
func applyPercentage(props json.RawMessage, baseAmount decimal.Decimal) (decimal.Decimal, string) {
    var cfg struct {
        Percentage string `json:"percentage"`
    }
    json.Unmarshal(props, &cfg)
    pct, _ := decimal.NewFromString(cfg.Percentage)
    charge := baseAmount.Mul(pct).Div(decimal.NewFromInt(100)).Round(2)
    return charge, fmt.Sprintf("%s%% of %s", cfg.Percentage, baseAmount.String())
}

// SpendingMinimum: enterprise floor commitment
// If calculated charges are below the minimum, charge the minimum
// properties: { "minimum_amount": "50000", "invoice_display_name": "Monthly minimum" }
func applySpendingMinimum(props json.RawMessage,
    currentTotal decimal.Decimal) (decimal.Decimal, string) {
    var cfg struct {
        MinimumAmount      string `json:"minimum_amount"`
        InvoiceDisplayName string `json:"invoice_display_name"`
    }
    json.Unmarshal(props, &cfg)
    minimum, _ := decimal.NewFromString(cfg.MinimumAmount)
    if currentTotal.GreaterThanOrEqual(minimum) { return decimal.Zero, "" }
    shortfall := minimum.Sub(currentTotal)
    return shortfall, cfg.InvoiceDisplayName
}
```

### 5.3 Full Calculation Algorithm

```go
// awo/internal/billing/pricing.go

type BillCalculation struct {
    TenantID         uuid.UUID
    PeriodStart      time.Time
    PeriodEnd        time.Time
    Lines            []BillLine
    Subtotal         decimal.Decimal
    RuleDiscount     decimal.Decimal
    GeoAdjustment    decimal.Decimal
    AnnualDiscount   decimal.Decimal
    WalletCredit     decimal.Decimal
    TaxAmount        decimal.Decimal
    Total            decimal.Decimal
}

func (e *PricingEngine) Calculate(ctx context.Context,
    params PriceCalculationParams) (*BillCalculation, error) {

    asOf := params.PeriodEnd
    sub  := params.Subscription
    calc := &BillCalculation{TenantID: params.TenantID,
        PeriodStart: params.PeriodStart, PeriodEnd: params.PeriodEnd}

    // ── Step 1: Base fee ──────────────────────────────────────────────────
    baseFee, _ := e.pricingRepo.GetCurrentBaseFee(ctx, sub.CurrencyCode, asOf)
    e.addLine(calc, BillLine{
        Type: "base_fee", Description: "Awo ERP — Base Platform",
        Quantity: decimal.NewFromInt(1), UnitPrice: baseFee.FeeMonthly,
        Amount: baseFee.FeeMonthly,
    })
    calc.Subtotal = baseFee.FeeMonthly

    // ── Step 2: Module access fees (volume charge model) ──────────────────
    enabledModules, _ := e.flagRepo.ListEnabledModuleFlags(ctx, params.TenantID)
    moduleCount := len(enabledModules)
    for _, mod := range enabledModules {
        // Check for tenant-level price override (grandfathering)
        price := e.resolveModulePrice(ctx, mod.ModuleID, params.TenantID,
            sub.CurrencyCode, asOf, moduleCount)
        e.addLine(calc, BillLine{
            Type: "module", ModuleID: &mod.ModuleID,
            Description: mod.Label + " module",
            Quantity: decimal.NewFromInt(1), UnitPrice: price, Amount: price,
        })
        calc.Subtotal = calc.Subtotal.Add(price)
    }

    // ── Step 3: Usage-based charges (from plan_charges + usage_events) ────
    activeCharges, _ := e.chargeRepo.ListActive(ctx)
    for _, charge := range activeCharges {
        metric, _ := e.metricRepo.Get(ctx, charge.BillableMetricID)
        usage, _  := e.aggregator.Aggregate(ctx, AggregationParams{
            TenantID:    params.TenantID,
            Metric:      metric,
            PeriodStart: params.PeriodStart,
            PeriodEnd:   params.PeriodEnd,
        })

        if usage.IsZero() { continue }

        amount, desc := e.applyChargeModel(charge.ChargeModel, charge.Properties, usage)
        if amount.IsZero() { continue }

        e.addLine(calc, BillLine{
            Type: "usage_charge", MetricCode: metric.Code,
            Description: metric.Name + " — " + desc,
            Quantity: usage, UnitPrice: amount.Div(usage).Round(6), Amount: amount,
        })
        calc.Subtotal = calc.Subtotal.Add(amount)
    }

    // ── Step 4: Pending proration lines from mid-cycle changes ────────────
    prorations, _ := e.subRepo.GetPendingProratedLines(ctx, params.TenantID)
    for _, p := range prorations {
        e.addLine(calc, BillLine{
            Type: "proration", Description: p.Description,
            PeriodStart: &p.PeriodStart, PeriodEnd: &p.PeriodEnd, Amount: p.Amount,
        })
        calc.Subtotal = calc.Subtotal.Add(p.Amount)
    }

    // ── Step 5: Geographical multiplier ──────────────────────────────────
    adjusted := e.applyGeoMultiplier(calc.Subtotal, sub.RegionCode, asOf)
    if !adjusted.Equal(calc.Subtotal) {
        calc.GeoAdjustment = adjusted.Sub(calc.Subtotal)
        calc.Subtotal = adjusted
    }

    // ── Step 6: Dynamic pricing rules ─────────────────────────────────────
    ruleDiscount := e.evaluatePricingRules(ctx, params.TenantID, calc)
    if ruleDiscount.IsPositive() {
        calc.RuleDiscount = ruleDiscount
        calc.Subtotal = calc.Subtotal.Sub(ruleDiscount)
    }

    // ── Step 7: Annual billing discount (15%) ─────────────────────────────
    if sub.BillingCycle == "annual" {
        annualDiscount := calc.Subtotal.Mul(decimal.NewFromFloat(0.15)).Round(2)
        calc.AnnualDiscount = annualDiscount
        e.addLine(calc, BillLine{
            Type: "discount", Description: "Annual billing discount (15%)",
            Amount: annualDiscount.Neg(),
        })
        calc.Subtotal = calc.Subtotal.Sub(annualDiscount)
    }

    // ── Step 8: Spending minimum (enterprise floor) ───────────────────────
    if minimumCharge := e.checkSpendingMinimum(ctx, params.TenantID, calc.Subtotal); minimumCharge.IsPositive() {
        e.addLine(calc, BillLine{
            Type: "spending_minimum", Description: "Monthly minimum commitment",
            Amount: minimumCharge,
        })
        calc.Subtotal = calc.Subtotal.Add(minimumCharge)
    }

    // ── Step 9: Wallet credits ─────────────────────────────────────────────
    walletCredit := e.applyWalletCredits(ctx, params.TenantID, calc.Subtotal, sub.CurrencyCode)
    if walletCredit.IsPositive() {
        calc.WalletCredit = walletCredit
        e.addLine(calc, BillLine{
            Type: "wallet_credit", Description: "Prepaid credit applied",
            Amount: walletCredit.Neg(),
        })
        calc.Subtotal = calc.Subtotal.Sub(walletCredit)
        if calc.Subtotal.IsNegative() { calc.Subtotal = decimal.Zero }
    }

    // ── Step 10: Tax ──────────────────────────────────────────────────────
    calc.TaxAmount = e.computeTax(calc.Subtotal, sub.RegionCode)
    calc.Total = calc.Subtotal.Add(calc.TaxAmount)

    return calc, nil
}
```

### 5.4 Dynamic Pricing Rules

```go
func (e *PricingEngine) evaluatePricingRules(ctx context.Context,
    tenantID uuid.UUID, calc *BillCalculation) decimal.Decimal {

    rules, _ := e.ruleRepo.ListActive(ctx)
    sub, _   := e.subRepo.GetByTenantID(ctx, tenantID)

    var matchingRules []domain.PricingRule
    for _, rule := range rules {
        if e.ruleMatches(ctx, rule, tenantID, sub, calc) {
            matchingRules = append(matchingRules, rule)
        }
    }

    if len(matchingRules) == 0 { return decimal.Zero }

    // Sort by discount value descending; apply best non-stacking rule
    sort.Slice(matchingRules, func(i, j int) bool {
        return e.ruleDiscountValue(matchingRules[i], calc.Subtotal).
            GreaterThan(e.ruleDiscountValue(matchingRules[j], calc.Subtotal))
    })

    best := matchingRules[0]
    discount := e.applyRule(best, calc.Subtotal)
    e.addLine(calc, BillLine{
        Type: "discount", Description: "Promotional: " + best.Name,
        Amount: discount.Neg(),
    })
    return discount
}

func (e *PricingEngine) ruleMatches(ctx context.Context,
    rule domain.PricingRule, tenantID uuid.UUID,
    sub *domain.Subscription, calc *BillCalculation) bool {

    var cond map[string]any
    json.Unmarshal(rule.ConditionValue, &cond)

    switch rule.ConditionType {
    case "module_count_gte":
        threshold := int(cond["count"].(float64))
        return len(e.enabledModules(ctx, tenantID)) >= threshold

    case "active_since_days_gte":
        days := int(cond["days"].(float64))
        return time.Since(sub.CreatedAt) >= time.Duration(days)*24*time.Hour

    case "region_in":
        regions := cond["regions"].([]any)
        for _, r := range regions {
            if r.(string) == sub.RegionCode { return true }
        }
        return false

    case "billing_cycle_eq":
        return sub.BillingCycle == cond["cycle"].(string)

    case "tag_in":
        tags := e.tenantRepo.GetTags(ctx, tenantID)
        required := cond["tags"].([]any)
        for _, rt := range required {
            if slices.Contains(tags, rt.(string)) { return true }
        }
        return false
    }
    return false
}
```

### 5.5 Proration Algorithm

When a module is enabled mid-period, the charge for the partial period is computed on a daily basis and stored as a pending proration line for the next invoice:

```go
// Daily proration: (monthly_price / days_in_month) × days_remaining
func (e *PricingEngine) ProrateDays(
    monthlyPrice decimal.Decimal,
    daysRemaining int,
    daysInMonth int) decimal.Decimal {

    if daysInMonth == 0 { return monthlyPrice }
    dailyRate := monthlyPrice.Div(decimal.NewFromInt(int64(daysInMonth)))
    return dailyRate.Mul(decimal.NewFromInt(int64(daysRemaining))).Round(2)
}

// Example:
// Monthly price: KES 5,000
// Period: January (31 days)
// Module enabled Jan 20 → 12 days remaining
// Proration = (5,000 / 31) × 12 = KES 1,935.48
```

---

## 6. Event Ingestion

### 6.1 Idempotent Ingestion

The critical property: the same event sent twice produces exactly one record. The caller supplies a `transaction_id` that is globally unique within the tenant:

```go
// awo/internal/billing/events.go

func (s *EventIngestService) Ingest(ctx context.Context,
    params domain.IngestEventParams) error {

    // Validate the event code maps to a known billable metric
    metric, err := s.metricRepo.GetByCode(ctx, params.Code)
    if err != nil {
        return domain.ErrInvalid("unknown billable metric: " + params.Code)
    }

    // Idempotency: if we have seen this transaction_id for this tenant, skip silently
    // This is NOT an error — it is correct behaviour for retries
    exists, _ := s.repo.EventExists(ctx, params.TenantID, params.TransactionID)
    if exists { return nil }

    return s.repo.InsertEvent(ctx, domain.UsageEvent{
        TenantID:         params.TenantID,
        BillableMetricID: metric.ID,
        TransactionID:    params.TransactionID,
        Code:             params.Code,
        Timestamp:        params.Timestamp,
        Properties:       params.Properties,
    })
}
```

### 6.2 Transaction ID Convention

```
Format: "{source}:{source_id}"

Module emitting events:
  Finance transaction posted:  "tx:{transaction_uuid}"
  User session start:          "session:{session_uuid}"
  Storage snapshot:            "storage:{tenant_uuid}:{YYYY-MM-DD}"
  Email sent:                  "email:{message_id}"
  SMS sent:                    "sms:{message_id}"
  API call:                    "api:{request_id}"
```

This convention guarantees uniqueness per source without coordination. A finance transaction posting the same event twice (e.g., transient failure before commit) produces `"tx:same-uuid"` both times — the second is silently dropped.

### 6.3 Late Arrival Handling

Events with `timestamp` in a prior billing period (late arrivals) are handled based on whether that period's invoice has already been generated:

```go
func (s *EventIngestService) HandleLateArrival(ctx context.Context,
    event domain.UsageEvent) error {

    // Find the billing period this event belongs to
    period, _ := s.subRepo.FindPeriodForDate(ctx, event.TenantID, event.Timestamp)

    if period.InvoiceGenerated {
        // Generate a one-off adjustment invoice for this usage
        // rather than re-opening and regenerating the past invoice
        s.invoiceSvc.CreateAdjustmentInvoice(ctx, domain.AdjustmentParams{
            TenantID:   event.TenantID,
            MetricCode: event.Code,
            Events:     []domain.UsageEvent{event},
            Reason:     "Late arrival usage adjustment",
        })
        return nil
    }

    // Period not yet invoiced — insert normally; it will be included
    return s.repo.InsertEvent(ctx, event)
}
```

---

## 7. Subscription Lifecycle

### 7.1 Trial

New tenants start a 14-day trial with all modules enabled. No invoice generated, no payment method required:

```go
func (s *SubscriptionService) CreateTrial(ctx context.Context,
    params domain.CreateTrialParams) (*domain.Subscription, error) {

    trialEnds := time.Now().AddDate(0, 0, 14)
    sub := &domain.Subscription{
        TenantID:           params.TenantID,
        Status:             "trialing",
        BillingCycle:       "monthly",
        RegionCode:         params.RegionCode,
        CurrencyCode:       params.CurrencyCode,
        TrialEndsAt:        &trialEnds,
        CurrentPeriodStart: time.Now().Truncate(24 * time.Hour),
        CurrentPeriodEnd:   trialEnds,
        NextInvoiceDate:    trialEnds,
    }
    if err := s.repo.Create(ctx, sub); err != nil { return nil, err }
    s.platform.Flags.EnableAllModulesForTenant(ctx, params.TenantID)
    s.events.Emit(ctx, domain.TrialStarted{TenantID: params.TenantID, EndsAt: trialEnds})
    return sub, nil
}
```

### 7.2 Activation

When a tenant adds a payment method and selects their modules:

```go
func (s *SubscriptionService) Activate(ctx context.Context,
    params domain.ActivateSubscriptionParams) error {

    // 1. Sync module flags from selection
    allModules, _ := s.platform.Flags.ListAllModuleFlags(ctx)
    for _, mod := range allModules {
        isSelected := slices.Contains(params.SelectedModuleIDs, mod.ModuleID)
        s.platform.Flags.SetModuleForTenant(ctx, params.TenantID, mod.FlagKey, isSelected)
    }
    s.platform.IAM.InvalidateByTenant(ctx, params.TenantID)

    // 2. Update subscription
    now := time.Now()
    periodEnd := nextBillingDate(now, params.BillingCycle)
    s.repo.UpdateSubscription(ctx, domain.UpdateSubscriptionParams{
        TenantID: params.TenantID, Status: "active",
        BillingCycle: params.BillingCycle,
        CurrentPeriodStart: now, CurrentPeriodEnd: periodEnd,
        NextInvoiceDate: now,
    })

    // 3. Generate first invoice and collect payment
    s.invoiceSvc.GenerateForTenant(ctx, params.TenantID)
    s.recordEvent(ctx, params.TenantID, "activated",
        map[string]any{"modules": params.SelectedModuleIDs})
    return nil
}
```

### 7.3 Module Enable/Disable Mid-Cycle

```go
func (s *SubscriptionService) EnableModule(ctx context.Context,
    tenantID, moduleID uuid.UUID) error {

    module, _ := s.platform.Flags.EnableModuleForTenant(ctx, tenantID, moduleID)
    s.platform.IAM.InvalidateByTenant(ctx, tenantID)

    // Compute proration and store as pending line for next invoice
    sub, _ := s.repo.GetByTenantID(ctx, tenantID)
    modulePrice := s.pricingEngine.ResolveModulePrice(ctx, moduleID,
        tenantID, sub.CurrencyCode, time.Now(), s.enabledModuleCount(ctx, tenantID))
    daysRemaining := daysUntil(sub.CurrentPeriodEnd)
    daysInPeriod  := daysInMonth(sub.CurrentPeriodStart)
    proration := s.pricingEngine.ProrateDays(modulePrice, daysRemaining, daysInPeriod)

    s.repo.AddPendingProratedLine(ctx, domain.ProratedLine{
        TenantID:    tenantID,
        ModuleID:    moduleID,
        Amount:      proration,
        Description: fmt.Sprintf("%s — proration (%d days remaining)",
            module.Label, daysRemaining),
        PeriodStart: time.Now(),
        PeriodEnd:   sub.CurrentPeriodEnd,
    })
    s.recordEvent(ctx, tenantID, "module_enabled", map[string]any{"module_id": moduleID})
    return nil
}

func (s *SubscriptionService) DisableModule(ctx context.Context,
    tenantID, moduleID uuid.UUID) error {
    // Access continues until period end — no immediate disable
    s.platform.Flags.ScheduleModuleDisable(ctx, domain.ScheduledFlagChange{
        TenantID:    tenantID,
        ModuleID:    moduleID,
        EffectiveAt: s.getSubscription(ctx, tenantID).CurrentPeriodEnd,
    })
    s.recordEvent(ctx, tenantID, "module_disabled", map[string]any{"module_id": moduleID})
    return nil
}
```

### 7.4 Dunning Schedule

```
Day 0:  Invoice due — payment attempt 1
Day 3:  Failed → status: 'past_due' → email: "Payment failed, please update"
Day 7:  Attempt 2 → email: "Second attempt failed"
Day 14: Attempt 3 → email: "Final notice — service at risk"
Day 21: Attempt 4 → status: 'suspended' → disable non-Finance modules
        (Finance stays on so tenant can see invoices and update payment method)
Day 30: Status: 'cancelled' → all modules disabled
```

### 7.5 Cancellation

```go
func (s *SubscriptionService) Cancel(ctx context.Context,
    params domain.CancelSubscriptionParams) error {

    sub, _ := s.repo.GetByTenantID(ctx, params.TenantID)

    // Schedule end-of-period cancellation; access continues until period end
    s.repo.UpdateSubscription(ctx, domain.UpdateSubscriptionParams{
        TenantID: params.TenantID, CancellationReason: params.Reason,
        CancelledAt: ptr(time.Now()),
    })

    // Annual subscriptions: issue credit note for unused months
    if sub.BillingCycle == "annual" {
        refund := e.computeAnnualRefund(sub, time.Now())
        if refund.IsPositive() {
            s.issueCreditNote(ctx, params.TenantID, refund,
                "Cancellation refund — unused subscription period")
        }
    }
    s.recordEvent(ctx, params.TenantID, "cancelled", map[string]any{"reason": params.Reason})
    return nil
}
```

---

## 8. Invoice Generation

### 8.1 Monthly Invoice Job

```go
func (s *InvoiceService) GenerateForTenant(ctx context.Context,
    tenantID uuid.UUID) error {

    sub, _ := s.subRepo.GetByTenantID(ctx, tenantID)

    calc, _ := s.pricingEngine.Calculate(ctx, PriceCalculationParams{
        TenantID:     tenantID,
        Subscription: sub,
        PeriodStart:  sub.CurrentPeriodStart,
        PeriodEnd:    sub.CurrentPeriodEnd,
    })

    inv, _ := s.createInvoiceRecord(ctx, sub, calc)

    // Post to Awo's GL
    s.glBridge.PostInvoice(ctx, inv)

    // Deduct wallet credits (already in calc; confirm actual deduction)
    if calc.WalletCredit.IsPositive() {
        s.walletSvc.Deduct(ctx, tenantID, calc.WalletCredit, inv.ID)
    }

    // Mark pending proration lines as invoiced
    s.subRepo.MarkProrationsInvoiced(ctx, tenantID)

    // Initiate payment collection
    s.paymentSvc.CollectPayment(ctx, inv)

    // Advance billing period
    s.subRepo.AdvancePeriod(ctx, tenantID)

    s.notify.Dispatch(ctx, notify.InvoiceGenerated{
        TenantID: tenantID, InvoiceID: inv.ID,
        Amount: inv.TotalAmount, DueDate: inv.DueDate,
    })
    return nil
}
```

### 8.2 Invoice Line Display

```
Invoice AWO-2025-00341
────────────────────────────────────────────────────────────────
Description                            Qty     Unit      Amount
────────────────────────────────────────────────────────────────
Awo ERP — Base Platform                  1  1,500.00   1,500.00
Finance (GL, AR, AP)                     1  4,500.00   4,500.00  ← volume rate (5 modules)
Human Resources                          1  4,500.00   4,500.00
Selling & CRM                            1  4,500.00   4,500.00
Inventory Management                     1  4,500.00   4,500.00
Forecourt Management                     1  4,500.00   4,500.00
────────────────────────────────────────────────────────────────
Module subtotal                                        22,500.00  (all at 4,500 — volume tier)
Active Users: 18 (graduated)            18                         
  Tier 1: 0–5  (5 × 0)                                    0.00
  Tier 2: 6–20 (13 × 300)                             3,900.00
Storage: 12.74 GB-months weighted         1    150.00   1,911.00  ← weighted_sum
Email sends: 2,450 (package)              2     50.00     100.00  ← package (100-send blocks)
Forecourt module — proration (12 days)                 1,935.48
────────────────────────────────────────────────────────────────
Subtotal                                              31,846.48
Loyalty discount 10% (24+ months)                   (3,184.65)
────────────────────────────────────────────────────────────────
Net amount                                            28,661.83
VAT 16%                                               4,585.89
────────────────────────────────────────────────────────────────
TOTAL DUE                                             33,247.72
Wallet credit applied                                (5,000.00)
────────────────────────────────────────────────────────────────
AMOUNT DUE NOW                                        28,247.72
────────────────────────────────────────────────────────────────
Due: 31 January 2025
```

### 8.3 One-Off Invoices

```go
// Immediate charge for professional services, setup fees, manual adjustments
func (s *InvoiceService) CreateOneOff(ctx context.Context,
    params domain.OneOffInvoiceParams) (*domain.Invoice, error) {

    inv := &domain.Invoice{
        TenantID:     params.TenantID,
        InvoiceNumber: nextInvoiceNumber(),
        InvoiceType:  "one_off",
        Status:       "open",
        IssueDate:    time.Now(),
        DueDate:      time.Now().AddDate(0, 0, 14),  // 14-day payment terms
        Description:  params.Description,
        Subtotal:     params.Amount,
        TaxAmount:    e.computeTax(params.Amount, params.RegionCode),
    }
    inv.TotalAmount = inv.Subtotal.Add(inv.TaxAmount)

    s.repo.Create(ctx, inv)
    s.glBridge.PostOneOffInvoice(ctx, inv)
    s.paymentSvc.CollectPayment(ctx, inv)
    s.notify.Dispatch(ctx, notify.InvoiceGenerated{
        TenantID: params.TenantID, InvoiceID: inv.ID,
        Amount: inv.TotalAmount, DueDate: inv.DueDate,
    })
    return inv, nil
}
```

---

## 9. Automated GL Entries

All entries post to **Awo's own tenant** (`AwoTenantID`). The `GLBridge` always targets this tenant.

### 9.1 Invoice Posted (Subscription)

```
DEBIT   1210 Accounts Receivable — Subscriptions     [invoice total inc. tax]
CREDIT  4{module} {Module} Revenue                   [net per module line]
CREDIT  4900 Base Platform Fee Revenue               [base fee]
CREDIT  4{usage} Usage Charge Revenue                [usage-based lines]
CREDIT  2310 VAT Payable                             [tax amount]

source_module = "billing", source_document_type = "subscription_invoice"
```

```go
func (b *GLBridge) PostInvoice(ctx context.Context,
    inv *domain.SubscriptionInvoice) error {

    var entries []finance.EntryInput

    // AR debit
    entries = append(entries, finance.EntryInput{
        AccountCode: "1210", DebitAmount: inv.TotalAmount,
        Description:  "Invoice " + inv.Number + " — " + inv.TenantName,
        CostCenter:   inv.TenantID.String(),
    })

    // Revenue credits per line
    for _, line := range inv.Lines {
        switch line.LineType {
        case "module", "base_fee", "usage_charge", "spending_minimum":
            revenueAccount := b.accounts.RevenueAccountFor(line)
            entries = append(entries, finance.EntryInput{
                AccountCode: revenueAccount, CreditAmount: line.Amount,
                Description: line.Description,
            })
        case "proration":
            // Proration credited to the same module's revenue account
            entries = append(entries, finance.EntryInput{
                AccountCode: b.accounts.RevenueAccountFor(line),
                CreditAmount: line.Amount, Description: line.Description,
            })
        }
    }

    // Discounts reduce the corresponding revenue account
    for _, line := range inv.Lines {
        if line.LineType == "discount" && !line.Amount.IsZero() {
            entries = append(entries, finance.EntryInput{
                AccountCode: "4950", DebitAmount: line.Amount.Abs(),
                Description: line.Description,
            })
        }
    }

    // VAT
    if inv.TaxAmount.IsPositive() {
        entries = append(entries, finance.EntryInput{
            AccountCode: "2310", CreditAmount: inv.TaxAmount,
        })
    }

    return b.postTransaction(ctx, inv.IssueDate, entries,
        "billing", "subscription_invoice", inv.ID)
}
```

### 9.2 Payment Received

```
DEBIT   1120 Stripe Settlement / 1110 Bank      [net received]
DEBIT   5210 Payment Processing Fees            [gateway fee]
CREDIT  1210 Accounts Receivable                [invoice total]
```

### 9.3 Annual Subscription — Cash Received → Deferred Revenue

```
DEBIT   1120 Stripe Settlement                  [annual total received]
DEBIT   5210 Payment Processing Fees            [Stripe fee]
CREDIT  2220 Deferred Revenue — Subscriptions   [full annual amount]
```

### 9.4 Monthly Deferred Revenue Recognition

On the 1st of each month, 1/12 transfers from Deferred Revenue to earned Revenue:

```
DEBIT   2220 Deferred Revenue — Subscriptions   [1/12 of annual contract value]
CREDIT  4{module} {Module} Revenue              [proportional per enabled module]
CREDIT  4900 Base Platform Fee Revenue          [1/12 of annual base fee]
```

### 9.5 Wallet Credit Issuance

When a tenant purchases prepaid credits:

```
DEBIT   1120 Bank / Stripe Settlement           [credit amount]
CREDIT  2230 Deferred Revenue — Wallet Credits  [prepaid credit liability]
```

When wallet credits are applied to an invoice:

```
DEBIT   2230 Deferred Revenue — Wallet Credits  [credits consumed]
CREDIT  1210 Accounts Receivable                [reduces AR balance]
```

### 9.6 Churn — Annual Refund

```
DEBIT   2220 Deferred Revenue                   [remaining unrecognised balance]
CREDIT  1120 Bank (refund paid)                 [refund amount]
CREDIT  4{module} Revenue (accelerated)         [remainder recognised immediately]
```

### 9.7 Vendor Bills (Cloud Infrastructure)

```
On receipt of AWS/GCP bill (AP module):
DEBIT   5110 Compute Costs                      [EC2/GCE charges]
DEBIT   5120 Database Costs                     [RDS/Cloud SQL charges]
DEBIT   5130 Storage Costs                      [S3/GCS charges]
DEBIT   5140 Networking & CDN                   [egress/CDN charges]
CREDIT  2110 AP — Amazon Web Services / GCP     [total bill]

On payment:
DEBIT   2110 AP — AWS / GCP
CREDIT  1110 Operating Account
```

Cost center tags (e.g., `AWO-INFRA-FINANCE`, `AWO-INFRA-AIRLINE`) flow from AWS resource tags into GL entries, enabling per-module gross margin analysis.

---

## 10. Platform as Its Own ERP Tenant

### 10.1 Awo's Tenant Record

```sql
-- Seeded once in baseline migration — fixed UUID, never changes
INSERT INTO tenants (id, slug, name, status)
VALUES ('00000000-0000-0000-0000-000000000001', 'awo', 'Awo ERP Ltd', 'active')
ON CONFLICT DO NOTHING;
```

```go
// awo/internal/billing/constants.go
var AwoTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
```

### 10.2 Chart of Accounts (SaaS Structure)

```
ASSETS
  1000 Cash & Bank
    1110 Operating Account — KES
    1120 Stripe Settlement — USD
    1130 M-Pesa Float
  1200 Accounts Receivable
    1210 Subscription Receivables (AR control)
  1300 Prepaid & Other Current Assets
    1310 Prepaid Cloud Credits
  1400 Non-Current Assets
    1410 Computer Equipment
    1415 Accumulated Depreciation — Equipment

LIABILITIES
  2100 Accounts Payable
    2110 AP — Amazon Web Services
    2120 AP — Google Cloud Platform
    2130 AP — Other Cloud Vendors
  2200 Deferred Revenue
    2220 Deferred Revenue — Annual Subscriptions
    2230 Deferred Revenue — Wallet Credits
  2300 Tax Payable
    2310 VAT / Output Tax

EQUITY
  3100 Founder Capital
  3200 Retained Earnings / Accumulated Deficit

REVENUE (by module — enables per-module margin analysis)
  4100 Finance Module Revenue
  4200 HR Module Revenue
  4300 Selling Module Revenue
  4400 Inventory Module Revenue
  4500 Airline Module Revenue
  4600 Forecourt Module Revenue
  4700 Restaurant Module Revenue
  4800 Professional Services Revenue
  4900 Base Platform Fee Revenue
  4950 Discounts & Credits (contra-revenue)

COST OF REVENUE (direct cost of delivering the service)
  5100 Cloud Infrastructure
    5110 Compute (EC2/GCE)
    5120 Database (RDS/Cloud SQL)
    5130 Storage (S3/GCS)
    5140 Networking & CDN
  5200 Third-Party SaaS COGS
    5210 Payment Processing Fees (Stripe)
    5220 Email Delivery (SendGrid)
    5230 SMS Delivery
    5240 Monitoring & Observability

OPERATING EXPENSES
  6000 Research & Development
    6100 Engineering Salaries
    6200 Engineering Contractors
  7000 Sales & Marketing
    7100 Sales Salaries
    7300 Advertising
  8000 General & Administrative
    8100 Management Salaries
    8200 Finance & Legal

OTHER
  9100 Interest Income
  9200–9300 FX Gain/Loss
```

### 10.3 Revenue Recognition Policy

| Scenario | Policy |
|---|---|
| Monthly subscription | Recognise on invoice date |
| Annual subscription | Deferred Revenue → recognise 1/12 monthly |
| Trial → Paid conversion | No revenue during trial |
| Proration | Recognise in the period the proration covers |
| Wallet credit deduction | Recognise when deducted from invoice |
| Credit note | Debit original revenue account |

### 10.4 Unit Economics Per Module

Per-module gross margin is queryable directly from GL:

```go
func (s *MetricsService) ModuleGrossMargin(ctx context.Context,
    periodStart, periodEnd time.Time) ([]ModuleMargin, error) {

    // Revenue per module from GL accounts 4100–4800
    revenues, _ := s.glRepo.SumByAccountPrefix(ctx, AwoTenantID,
        []string{"4100","4200","4300","4400","4500","4600","4700"},
        periodStart, periodEnd)

    // COGS from 5xxx accounts tagged with cost centers
    cogsByModule, _ := s.glRepo.SumByCostCenter(ctx, AwoTenantID,
        "AWO-INFRA-", periodStart, periodEnd)

    var margins []ModuleMargin
    for moduleSlug, revenue := range revenues {
        cogs := cogsByModule[moduleSlug]
        margins = append(margins, ModuleMargin{
            ModuleSlug:  moduleSlug,
            Revenue:     revenue,
            COGS:        cogs,
            GrossMargin: revenue.Sub(cogs),
            MarginPct:   revenue.Sub(cogs).Div(revenue).Mul(decimal.NewFromInt(100)).Round(1),
        })
    }
    return margins, nil
}
```

---

## 11. Platform User vs Platform Tenant — Data Access Boundary

### 11.1 The Boundary

```
Platform operators CAN access:
  ├── All subscription metadata (plan, status, modules enabled, MRR)
  ├── All billing invoices and payment history (Awo's AR)
  ├── Usage metrics (storage, API calls, user counts)
  ├── SaaS analytics (MRR waterfall, churn, cohorts)
  ├── Awo's own GL data (via Awo's tenant login — not the platform session)
  └── System health (jobs, errors, performance)

Platform operators CANNOT access (without explicit tenant grant):
  ├── Tenant's chart of accounts content
  ├── Tenant's transaction data (their customer invoices, supplier bills)
  ├── Tenant's employee records and payroll
  ├── Tenant's customer/contact records
  └── Any business data belonging to the tenant
```

**The rule:** billing and operational metadata *about* a tenant is Awo's data. Business transactions *inside* a tenant are the tenant's data.

### 11.2 Database Enforcement

Two independent layers:

```
Layer 1: PostgreSQL role (awo_platform with BYPASSRLS)
  Platform operators use the awo_platform DB pool.
  BYPASSRLS bypasses Row-Level Security at the engine level.
  This is for operational needs — not a permission to read business data.
  Table-level GRANTs still apply.

Layer 2: Application service gates
  Every Finance service method checks the caller's identity:
  - Tenant user → session.TenantID enforces data scope automatically (RLS)
  - Platform user without tenant context → can only read platform tables
    (subscriptions, invoices, usage_events, tenant metadata)
  - Platform user inspecting a tenant → requires platform.support.tenant_data
    permission AND an active support ticket ID + triggers SensitiveAccessLog
```

### 11.3 The Dual-Role Pattern

A platform engineer at Awo has two separate accounts:

```
john@awoerp.com          (user_type='platform')
  → Platform operations: managing tenants, billing support, system monitoring
  → Session domain: _platform_
  → DB pool: awo_platform (BYPASSRLS)

john.doe@awoerp.com      (user_type='tenant', tenant_id=AwoTenantID)
  → Awo's internal finance, HR, operations
  → Session domain: <awo_tenant_uuid>
  → DB pool: awo_app (RLS active, scoped to Awo's tenant)
```

Separate credentials. Separate sessions. Separate audit logs. A compromised platform session does not expose Awo's accounting data. An accountant using the Finance module does not have platform operator capabilities.

### 11.4 Support Access Pattern (Break-Glass)

```go
// Tenant admin explicitly grants time-limited support access
func (s *SupportService) CreateSupportSession(ctx context.Context,
    params CreateSupportSessionParams) (*SupportSession, error) {

    // Tenant admin must have approved via portal
    if !s.tenantAdminGranted(ctx, params.TargetTenantID, params.TicketID) {
        return nil, domain.ErrForbidden.WithMessage(
            "tenant administrator must approve support access first")
    }

    session := &SupportSession{
        PlatformUserID: params.PlatformUserID,
        TargetTenantID: params.TargetTenantID,
        TicketID:       params.TicketID,
        ExpiresAt:      time.Now().Add(4 * time.Hour),
    }
    s.cache.Set(ctx, "support:"+session.TicketID, session, 4*time.Hour)

    // Notify tenant — they receive email every time their data is accessed
    s.notify.Send(ctx, notify.SupportAccessGranted{
        TenantID:     params.TargetTenantID,
        AccessorName: params.PlatformUserName,
        ExpiresAt:    session.ExpiresAt,
        TicketID:     params.TicketID,
    })
    return session, nil
}
```

---

## 12. SaaS Metrics & Reporting

### 12.1 MRR Waterfall

```go
type MRRMovement struct {
    AsOf           time.Time
    StartingMRR    decimal.Decimal
    NewMRR         decimal.Decimal     // new activations
    ExpansionMRR   decimal.Decimal     // module additions by existing tenants
    ContractionMRR decimal.Decimal     // module removals
    ChurnedMRR     decimal.Decimal     // cancellations
    ReactivatedMRR decimal.Decimal
    EndingMRR      decimal.Decimal
    NetNewMRR      decimal.Decimal     // New + Expansion - Contraction - Churned
}
```

MRR is computed from `subscription_events` joined to module prices: each `module_enabled` event adds the module's monthly price to MRR; each `module_disabled` or `cancelled` subtracts it.

### 12.2 Customer Metrics

```
Monthly churn rate = churned_tenants / active_at_start_of_period
Annual churn rate  = 1 - (1 - monthly_churn)^12

LTV  = ARPU / monthly_churn_rate
CAC  = sales_and_marketing_spend / new_customers_acquired
LTV/CAC target: > 3.0
Payback months = CAC / ARPU   target: < 12 months
```

### 12.3 Module Adoption Analytics

```sql
-- Which modules are generating the most revenue and how widely adopted are they?
SELECT
  m.slug, m.label,
  COUNT(DISTINCT tff.tenant_id) AS tenants_enabled,
  COUNT(DISTINCT ts.tenant_id)  AS total_active_tenants,
  ROUND(COUNT(DISTINCT tff.tenant_id)::numeric /
        NULLIF(COUNT(DISTINCT ts.tenant_id),0) * 100, 1) AS adoption_pct,
  COUNT(DISTINCT tff.tenant_id) * MAX(mp.price_monthly) AS module_mrr
FROM modules m
LEFT JOIN feature_flag_definitions ffd ON ffd.module_id = m.id AND ffd.resource_id IS NULL
LEFT JOIN tenant_feature_flags tff     ON tff.flag_id = ffd.id AND tff.enabled = true
LEFT JOIN tenant_subscriptions ts      ON ts.status = 'active'
LEFT JOIN module_pricing mp            ON mp.module_id = m.id AND mp.effective_until IS NULL
GROUP BY m.id, m.slug, m.label
ORDER BY module_mrr DESC;
```

### 12.4 Cohort Retention

```sql
-- Monthly cohort retention heatmap data
WITH cohorts AS (
  SELECT DATE_TRUNC('month', MIN(occurred_at)) AS cohort_month, tenant_id
  FROM subscription_events WHERE event_type = 'activated' GROUP BY tenant_id
),
monthly_active AS (
  SELECT c.cohort_month, c.tenant_id,
         DATE_TRUNC('month', ts.current_period_start) AS activity_month,
         EXTRACT(MONTH FROM AGE(
           DATE_TRUNC('month', ts.current_period_start), c.cohort_month
         )) AS months_since_start
  FROM cohorts c
  JOIN tenant_subscriptions ts ON ts.tenant_id = c.tenant_id
  WHERE ts.status IN ('active','past_due')
)
SELECT cohort_month, months_since_start, COUNT(DISTINCT tenant_id) AS retained
FROM monthly_active
GROUP BY cohort_month, months_since_start
ORDER BY cohort_month, months_since_start;
```

### 12.5 Report Types

| Report | Data Source | Audience |
|---|---|---|
| MRR Dashboard / Waterfall | `subscription_events` + `module_pricing` | All platform |
| Invoice Ageing | `subscription_invoices` | Finance |
| Churn Analysis | `subscription_events` | Product + Finance |
| Module Adoption | `tenant_feature_flags` + `modules` | Product |
| Cohort Retention | `subscription_events` | Product + Finance |
| Unit Economics | GL 5xxx + billing metrics | CEO + Finance |
| Gross Margin by Module | GL 4xxx + 5xxx by cost center | CEO + Finance |
| P&L Statement | GL `finance_transactions` WHERE tenant=AwoTenantID | Finance team (via Awo tenant login) |
| Balance Sheet | GL (Awo tenant) | Finance team |
| Cash Flow | GL (Awo tenant) | Finance team |

---

## 13. Billing Services & Actions

### 13.1 BillingService Facade

```go
type BillingService struct {
    Events        *EventIngestService
    Subscriptions *SubscriptionService
    Invoices      *InvoiceService
    Payments      *PaymentService
    Pricing       *PricingEngine
    Wallets       *WalletService
    Dunning       *DunningService
    Recognition   *RevenueRecognitionService
    Metrics       *SaaSMetricsService
    GLBridge      *GLBridge
}
```

### 13.2 Service Interfaces

```go
type EventIngestService interface {
    Ingest(ctx, params domain.IngestEventParams) error
    HandleLateArrival(ctx context.Context, event domain.UsageEvent) error
    GetUsageSummary(ctx, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*UsageSummary, error)
}

type PricingEngine interface {
    Calculate(ctx, params PriceCalculationParams)      (*BillCalculation, error)
    Preview(ctx, tenantID uuid.UUID)                   (*BillCalculation, error)
    ProrateDays(price decimal.Decimal, remaining, total int) decimal.Decimal
    ResolveModulePrice(ctx, moduleID, tenantID uuid.UUID,
                       currency string, asOf time.Time, moduleCount int) decimal.Decimal
}

type SubscriptionService interface {
    CreateTrial(ctx, params CreateTrialParams)         (*Subscription, error)
    Activate(ctx, params ActivateParams)               error
    EnableModule(ctx, tenantID, moduleID uuid.UUID)    error
    DisableModule(ctx, tenantID, moduleID uuid.UUID)   error
    Cancel(ctx, params CancelParams)                   error
    Reactivate(ctx, tenantID uuid.UUID)                error
    GetSubscription(ctx, tenantID uuid.UUID)           (*Subscription, error)
    PreviewNextInvoice(ctx, tenantID uuid.UUID)        (*BillCalculation, error)
}

type InvoiceService interface {
    GenerateForTenant(ctx, tenantID uuid.UUID)                      error
    RunMonthlyInvoicingJob(ctx context.Context)                      error
    CreateOneOff(ctx, params domain.OneOffInvoiceParams)            (*Invoice, error)
    GetInvoice(ctx, invoiceID uuid.UUID)                            (*Invoice, error)
    ListForTenant(ctx, params ListInvoicesParams)                   ([]*Invoice, int, error)
    VoidInvoice(ctx, invoiceID uuid.UUID, reason string)            error
    GeneratePDF(ctx, invoiceID uuid.UUID)                           ([]byte, error)
}

type WalletService interface {
    GetBalance(ctx, tenantID uuid.UUID, currency string)             (decimal.Decimal, error)
    AddCredit(ctx, params AddCreditParams)                           error
    Deduct(ctx, tenantID uuid.UUID, amount decimal.Decimal,
           invoiceID uuid.UUID)                                      error
    ListTransactions(ctx, tenantID uuid.UUID)                        ([]*WalletTransaction, error)
}
```

### 13.3 Scheduled Jobs

| Job | Schedule | Action |
|---|---|---|
| `MonthlyInvoicingJob` | Nightly 02:00 | Generate invoices for `next_invoice_date ≤ today` |
| `MonthlyRecognitionJob` | 1st of month 03:00 | Recognise 1/12 deferred revenue for annual subs |
| `TrialExpiryJob` | Nightly 01:00 | Notify trials ending in 3 days; downgrade expired |
| `DunningRetryJob` | Nightly 04:00 | Retry failed payments per dunning schedule |
| `ScheduledModuleDisableJob` | Nightly 00:30 | Apply end-of-period module disables |
| `EndOfPeriodCancellationJob` | Nightly 00:00 | Finalise pending cancellations |
| `UsageStorageSnapshotJob` | Daily 23:00 | Emit storage_snapshot events per tenant |
| `WalletExpiryJob` | Nightly 01:30 | Expire and void wallet credits past expiry_at |
| `SaaSMetricsSnapshotJob` | Nightly 05:00 | Snapshot MRR, ARR, churn into analytics table |

---

## 14. Payment Gateway Integration

### 14.1 Stripe

```go
func (s *PaymentService) CollectPayment(ctx context.Context,
    invoice *domain.SubscriptionInvoice) error {

    sub, _ := s.subRepo.GetByTenantID(ctx, invoice.TenantID)
    amountCents := invoice.TotalAmount.Mul(decimal.NewFromInt(100)).IntPart()

    pi, err := s.stripe.PaymentIntents.New(&stripe.PaymentIntentParams{
        Amount:        stripe.Int64(amountCents),
        Currency:      stripe.String(strings.ToLower(invoice.CurrencyCode)),
        Customer:      stripe.String(sub.PaymentCustomerID),
        PaymentMethod: stripe.String(sub.PaymentMethodID),
        Confirm:       stripe.Bool(true),
        OffSession:    stripe.Bool(true),
        Metadata: map[string]string{
            "invoice_id":     invoice.ID.String(),
            "invoice_number": invoice.Number,
            "tenant_id":      invoice.TenantID.String(),
        },
    })
    if err != nil {
        s.dunning.ProcessFailedPayment(ctx, invoice.ID)
        return nil  // dunning handles recovery
    }
    s.invoiceRepo.SetPaymentIntent(ctx, invoice.ID, pi.ID)
    return nil
}
```

### 14.2 Stripe Webhook (Idempotent Handler)

```go
func StripeWebhook(billing *billing.BillingService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        event, err := stripe.ConstructEvent(c.Body(),
            c.Get("Stripe-Signature"), webhookSecret)
        if err != nil { return c.Status(400).JSON(response.Err("invalid signature")) }

        switch event.Type {
        case "payment_intent.succeeded":
            pi := event.Data.Object.(*stripe.PaymentIntent)
            invoiceID, _ := uuid.Parse(pi.Metadata["invoice_id"])
            billing.Payments.RecordSuccess(c.Context(), invoiceID,
                pi.ID, decimal.NewFromInt(pi.AmountReceived).Div(decimal.NewFromInt(100)),
                decimal.NewFromInt(pi.ApplicationFeeAmount).Div(decimal.NewFromInt(100)))

        case "payment_intent.payment_failed":
            pi := event.Data.Object.(*stripe.PaymentIntent)
            invoiceID, _ := uuid.Parse(pi.Metadata["invoice_id"])
            billing.Dunning.ProcessFailedPayment(c.Context(), invoiceID)
        }
        return c.SendStatus(200)
    }
}
```

### 14.3 M-Pesa (Kenya)

```go
func (s *PaymentService) InitiateMPesaSTKPush(ctx context.Context,
    invoiceID uuid.UUID, phoneNumber string) error {

    invoice, _ := s.invoiceRepo.Get(ctx, invoiceID)
    resp, _ := s.mpesa.STKPush(ctx, mpesa.STKPushRequest{
        BusinessShortCode: config.MpesaPaybillNumber,
        Amount:            invoice.TotalAmount.IntPart(),
        PhoneNumber:       phoneNumber,
        AccountReference:  invoice.Number,
        TransactionDesc:   "Awo ERP — " + invoice.Number,
    })
    s.invoiceRepo.SetMpesaCheckoutID(ctx, invoiceID, resp.CheckoutRequestID)
    return nil
}
```

---

## 15. Testing Strategy

### 15.1 Aggregation Tests

```go
func TestAggregator_UniqueCount_ActiveUsers(t *testing.T) {
    // 3 unique users logged in during the period, one logged in twice
    ingestEvents(t, []domain.UsageEvent{
        {TransactionID: "session:a1", Code: "user_session",
         Properties: map[string]any{"user_id": "u1"}},
        {TransactionID: "session:a2", Code: "user_session",
         Properties: map[string]any{"user_id": "u2"}},
        {TransactionID: "session:a3", Code: "user_session",
         Properties: map[string]any{"user_id": "u1"}},  // u1 again — deduplicated
        {TransactionID: "session:a4", Code: "user_session",
         Properties: map[string]any{"user_id": "u3"}},
    })
    result, _ := aggregator.Aggregate(ctx, AggregationParams{
        Metric: activeUsersMetric, TenantID: tenantID,
        PeriodStart: jan1, PeriodEnd: jan31,
    })
    assert.Equal(t, "3", result.String())  // 3 unique users, not 4 events
}

func TestAggregator_WeightedSum_Storage(t *testing.T) {
    // 10 GB from Jan 1, +5 GB added Jan 15 (31-day month)
    // 10 GB × 31 days = 310; 5 GB × 17 days = 85; total 395 / 31 = 12.74
    ingestEvents(t, []domain.UsageEvent{
        {TransactionID: "storage:jan1",  Code: "storage_snapshot", Timestamp: jan1,
         Properties: map[string]any{"gb": 10.0}},
        {TransactionID: "storage:jan15", Code: "storage_snapshot", Timestamp: jan15,
         Properties: map[string]any{"gb": 15.0}},
    })
    result, _ := aggregator.Aggregate(ctx, AggregationParams{
        Metric: storageMetric, TenantID: tenantID,
        PeriodStart: jan1, PeriodEnd: jan31,
    })
    assert.Equal(t, "12.74", result.Round(2).String())
}
```

### 15.2 Charge Model Tests

```go
func TestGraduated_Users(t *testing.T) {
    // 25 users: 0-5 free, 6-20 at 300, 21-25 at 250
    // 5×0 + 15×300 + 5×250 = 0 + 4,500 + 1,250 = 5,750
    amount, _ := applyGraduated(userGraduatedProps, decimal.NewFromInt(25))
    assert.Equal(t, "5750.00", amount.String())
}

func TestVolume_Modules(t *testing.T) {
    // 5 modules falls in the 4-6 tier at KES 4,500 each
    // ALL 5 modules priced at 4,500 (not cumulative)
    amount, _ := applyVolume(moduleVolumeProps, decimal.NewFromInt(5))
    assert.Equal(t, "22500.00", amount.String())
}

func TestPackage_Emails(t *testing.T) {
    // 2,450 emails: 1,000 free, then ceil(1,450/100) = 15 blocks × 50 = 750
    amount, _ := applyPackage(emailPackageProps, decimal.NewFromInt(2450))
    assert.Equal(t, "750.00", amount.String())
}
```

### 15.3 Idempotency Test

```go
func TestEventIngestion_Idempotent(t *testing.T) {
    event := domain.IngestEventParams{
        TenantID:      tenantID,
        TransactionID: "tx:same-uuid",
        Code:          "transaction_posted",
        Timestamp:     time.Now(),
    }
    // Ingest same event twice — should not error and should not create duplicate
    err1 := svc.Ingest(ctx, event)
    err2 := svc.Ingest(ctx, event)
    assert.NoError(t, err1)
    assert.NoError(t, err2)

    count, _ := eventRepo.Count(ctx, tenantID, "transaction_posted")
    assert.Equal(t, 1, count)  // only one record created
}
```

### 15.4 GL Bridge Tests

```go
func TestGLBridge_PostInvoice_BalancedEntries(t *testing.T) {
    bridge := setupTestBridge(t)
    invoice := buildTestInvoice(t,
        []InvoiceLine{
            {LineType: "module", ModuleSlug: "finance", Amount: dec("4500")},
            {LineType: "usage_charge", MetricCode: "active_users", Amount: dec("3900")},
            {LineType: "usage_charge", MetricCode: "storage_gb", Amount: dec("1911")},
            {LineType: "discount", Description: "Loyalty discount", Amount: dec("-1031.10")},
        },
        dec("1476.64"),  // tax
    )

    err := bridge.PostInvoice(ctx, invoice)
    assert.NoError(t, err)

    tx, _ := financeRepo.GetBySourceDocument(ctx, AwoTenantID,
        "billing", "subscription_invoice", invoice.ID)
    assert.NotNil(t, tx)
    assert.Equal(t, "POSTED", tx.Status)
    // Double-entry must balance
    assert.Equal(t, tx.TotalDebitAmount, tx.TotalCreditAmount)
}
```

---

## 16. Troubleshooting

### 16.1 Aggregation Producing Wrong Number

```sql
-- Count raw events for this tenant/metric/period
SELECT COUNT(*), MIN(timestamp), MAX(timestamp)
FROM usage_events
WHERE tenant_id          = '<tenant_uuid>'
  AND billable_metric_id = (SELECT id FROM billable_metrics WHERE code = 'active_users')
  AND timestamp BETWEEN '2025-01-01' AND '2025-01-31';

-- Check for duplicate transaction_ids (should be 0)
SELECT transaction_id, COUNT(*)
FROM usage_events
WHERE tenant_id = '<tenant_uuid>'
GROUP BY transaction_id HAVING COUNT(*) > 1;

-- For unique_count: verify field_name values
SELECT properties ->> 'user_id' AS user_id, COUNT(*)
FROM usage_events
WHERE tenant_id = '<tenant_uuid>' AND code = 'user_session'
  AND timestamp BETWEEN '2025-01-01' AND '2025-01-31'
GROUP BY user_id;
```

### 16.2 Invoice Amount Unexpected

```sql
-- Check what plan_charges are active
SELECT pc.charge_model, pc.properties, bm.code, bm.aggregation_type
FROM plan_charges pc
JOIN billable_metrics bm ON pc.billable_metric_id = bm.id
WHERE pc.is_active = true;

-- Check if a pricing rule fired
SELECT * FROM pricing_rules WHERE is_active = true
ORDER BY priority;

-- Check tenant-level price overrides
SELECT * FROM tenant_price_overrides WHERE tenant_id = '<uuid>';

-- Replay the pricing calculation for debugging
-- (call PricingEngine.Preview with a specific date)
```

### 16.3 GL Entry Not Posted

```sql
-- Check if invoice has GL transaction linked
SELECT id, invoice_number, status, gl_transaction_id
FROM subscription_invoices WHERE id = '<uuid>';

-- If gl_transaction_id is NULL, check billing error log
SELECT * FROM billing_errors WHERE invoice_id = '<uuid>'
ORDER BY occurred_at DESC;

-- Manually re-trigger GL posting (platform admin action)
-- POST /platform/billing/invoices/<id>/post-to-gl
```

### 16.4 Event Not Being Counted

```
Cause: transaction_id already exists (appears as silent duplicate drop)

Diagnosis:
  SELECT * FROM usage_events
  WHERE transaction_id = 'tx:your-uuid' AND tenant_id = '<uuid>';
  -- If a row exists, the event was processed correctly on a prior attempt

Cause: event_code does not match any billable_metric.event_code
  SELECT * FROM billable_metrics WHERE event_code = 'your_code';
  -- If no rows, the metric needs to be defined

Cause: timestamp outside the period window
  -- Check that event.timestamp is within the billing period dates
```

### 16.5 Common Error Reference

| Error | Cause | Fix |
|---|---|---|
| `unknown billable metric code` | Metric not seeded | Insert row into `billable_metrics` |
| `no pricing found for module` | Module has no `module_pricing` row | Insert pricing for the module |
| `gl_bridge: imbalanced entries` | Revenue lines don't match AR debit | Check line totals; verify discount sign |
| `wallet: insufficient balance` | Credits exhausted before invoice | Let invoice charge normally; wallet deducts what is available |
| `stripe: no_payment_method` | Tenant has no card on file | Send payment method collection email |
| `recognition_job: negative deferred` | Over-recognition | Review annual subscription recognition run history |
| MRR lower than expected | Past-due tenants excluded | Check dunning status; include `past_due` in MRR count |
| Module still visible after cancellation | Scheduled disable job missed | Check `EndOfPeriodCancellationJob` logs; manually disable |
