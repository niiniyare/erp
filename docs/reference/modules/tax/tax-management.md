# Tax Management Module — Comprehensive Documentation

**Version:** 1.0  
**Scope:** Multi-jurisdiction tax engine covering VAT/GST, Withholding Tax, Payroll Tax, Excise Duty, Import/Customs Duty, and Corporate Tax  
**Jurisdictions:** Kenya, Uganda, Tanzania, Nigeria, South Africa  

---

## Table of Contents

1. [Overview & Architecture](#1-overview--architecture)
2. [Tax Configuration UI — Adding New Taxes](#2-tax-configuration-ui--adding-new-taxes)
3. [VAT / GST](#3-vat--gst)
4. [Withholding Tax (WHT)](#4-withholding-tax-wht)
5. [Payroll Tax (PAYE, NSSF, NHIF/SHIF, SDL, NITA)](#5-payroll-tax-paye-nssf-nhifshif-sdl-nita)
6. [Excise Duty](#6-excise-duty)
7. [Import Duty / Customs](#7-import-duty--customs)
8. [Corporate Tax](#8-corporate-tax)
9. [e-Invoicing — KRA eTIMS Integration](#9-e-invoicing--kra-etims-integration)
10. [Multi-Jurisdiction Rate Tables](#10-multi-jurisdiction-rate-tables)
11. [Tax Return Generation](#11-tax-return-generation)
12. [WHT Certificates](#12-wht-certificates)
13. [Compliance Calendar & Alerts](#13-compliance-calendar--alerts)
14. [Data Model Reference](#14-data-model-reference)

---

## 1. Overview & Architecture

### 1.1 Purpose

The Tax Management module is a configurable, multi-jurisdiction engine that handles the full tax lifecycle — from transaction-time calculation through filing, payment, and audit. It is designed to support the complex and frequently changing tax environments of East and West African markets alongside international operations.

### 1.2 Core Design Principles

**Configuration over code.** Every tax type, rate, rule, exemption, and jurisdiction is stored as data. Introducing a new tax or updating a rate requires no code deployment — only a UI action by an authorized administrator.

**Jurisdiction layering.** Taxes are modelled at three levels — national, state/county, and local — and the engine stacks applicable taxes based on origin and destination addresses for every transaction.

**Immutable audit trail.** Every tax calculation is stored with its full input set, rule snapshot, and output. Rate changes are effective-dated; historical transactions always recalculate against the rate that was active at the transaction date.

**e-Invoicing native.** For Kenya, the engine integrates natively with KRA eTIMS at the point of invoice generation, not as a post-process batch.

### 1.3 Supported Tax Types

| Code | Tax Type | Primary Jurisdictions |
|---|---|---|
| `VAT` | Value Added Tax | Kenya, Uganda, Tanzania, South Africa |
| `GST` | Goods & Services Tax | (future) |
| `WHT` | Withholding Tax | All five jurisdictions |
| `PAYE` | Pay As You Earn (Payroll) | All five jurisdictions |
| `NSSF` | National Social Security Fund | Kenya, Uganda, Tanzania |
| `NHIF` | National Hospital Insurance Fund | Kenya |
| `SHIF` | Social Health Insurance Fund | Kenya (replaces NHIF from 2024) |
| `SDL` | Skills Development Levy | Kenya, Tanzania |
| `NITA` | National Industrial Training Authority levy | Kenya |
| `EXCISE` | Excise Duty | All five jurisdictions |
| `CUSTOMS` | Import Duty / Customs | All five jurisdictions |
| `CIT` | Corporate Income Tax | All five jurisdictions |

### 1.4 System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     ERP Core Modules                    │
│  Sales │ Purchasing │ Payroll │ Expenses │ Inventory     │
└────────────────────┬────────────────────────────────────┘
                     │ TaxableTransaction
                     ▼
┌─────────────────────────────────────────────────────────┐
│               Tax Calculation Engine                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Jurisdiction │  │  Rate Lookup │  │  Exemption   │  │
│  │  Resolver    │  │  Service     │  │  Checker     │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│  ┌──────────────────────────────────────────────────┐   │
│  │         Tax Rule Engine (per tax type)           │   │
│  │  VAT │ WHT │ PAYE │ Excise │ Customs │ CIT       │   │
│  └──────────────────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────┘
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
┌──────────────────┐   ┌──────────────────┐
│  Tax Ledger &    │   │  Compliance &    │
│  GL Posting      │   │  Return Engine   │
└──────────────────┘   └────────┬─────────┘
                                │
                    ┌───────────┼───────────┐
                    ▼           ▼           ▼
              ┌──────────┐ ┌──────┐ ┌──────────┐
              │ KRA eTIMS│ │ URA  │ │ FIRS/TRA │
              │(e-Invoice)│ │ API  │ │  portals │
              └──────────┘ └──────┘ └──────────┘
```

---

## 2. Tax Configuration UI — Adding New Taxes

### 2.1 Overview

All tax types, rates, rules, and jurisdictions are managed through the **Settings → Tax Management** section of the UI. Administrators can add entirely new taxes, update rates, set effective dates, map GL accounts, and configure filing rules — all without developer intervention.

### 2.2 Navigation

**Settings → Tax Management → Tax Codes → + New Tax Code**

### 2.3 Step-by-Step: Adding a New Tax

#### Step 1 — Select the Tax Authority

Before creating a tax code, the relevant tax authority must exist. Navigate to **Tax Authorities** and confirm the authority is listed. If not, create it first.

| Field | Description | Example |
|---|---|---|
| Authority Name | Full legal name of the authority | Kenya Revenue Authority |
| Authority Code | Short internal code | KRA |
| Country | ISO country code | KE |
| Authority Type | federal / state / local / vat / customs | federal |
| Filing Frequency | How often returns are filed | monthly |
| Filing Due Day | Day of month a return is due | 20 |
| Electronic Filing | Whether e-filing is supported | Yes |
| eTIMS / API Endpoint | Integration URL (if applicable) | https://etims-api.kra.go.ke |

#### Step 2 — Create the Tax Code

Navigate to **Tax Codes → + New Tax Code**.

**Basic Information tab:**

| Field | Description | Required |
|---|---|---|
| Tax Code | Unique short code (e.g., `KE-VAT-STD`) | Yes |
| Tax Name | Display name (e.g., "Kenya VAT Standard Rate") | Yes |
| Description | Long-form explanation | No |
| Tax Type | Dropdown: VAT, WHT, Excise, Customs, Payroll, CIT | Yes |
| Tax Category | Standard / Reduced / Zero-Rated / Exempt | Yes |
| Tax Authority | Link to authority created in Step 1 | Yes |

**Rate & Calculation tab:**

| Field | Description | Example |
|---|---|---|
| Calculation Method | Percentage / Fixed Amount / Progressive / Lookup Table | Percentage |
| Tax Rate | Decimal (e.g., enter 16 for 16%) | 16 |
| Effective From | Date this rate takes effect | 2024-01-01 |
| Expiry Date | Optional end date (leave blank for indefinite) | — |
| Minimum Taxable Amount | Threshold below which tax is not applied | — |
| Maximum Tax Cap | Cap on tax amount (if applicable) | — |
| Compound Tax | Whether tax is calculated on an already-taxed amount | No |
| Cascade Order | Calculation sequence when multiple taxes apply | 1 |

**GL Mapping tab:**

| Field | Account | Example |
|---|---|---|
| Tax Payable Account | Liability account for taxes collected | 2110 — VAT Payable |
| Tax Expense Account | Expense account for non-reclaimable input tax | 6310 — Tax Expense |
| Tax Receivable Account | Asset for reclaimable input VAT | 1310 — VAT Receivable |

**Applicability tab:**

| Rule | Options |
|---|---|
| Transaction Types | Sales / Purchases / Expenses / Payroll (multi-select) |
| Product Categories | Goods / Services / Digital / Exempt items |
| Customer Categories | All / Business / Government / Nonprofit / Export |
| Geographic Scope | Countries and states/counties where this code applies |

**Filing & Reporting tab:**

| Field | Description |
|---|---|
| Return Line Number | The line number on the official tax return form |
| Reporting Code | Authority's official box reference |
| Include in e-Invoice | Whether to transmit this tax line via eTIMS / API |

#### Step 3 — Add Rate History (if applicable)

For taxes with a history of rate changes (e.g., VAT moved from 16% to 18%), use the **Rate History** sub-table to add past rates with their effective and expiry dates. The engine automatically selects the correct rate for any historical transaction date.

#### Step 4 — Configure Exemptions

Navigate to **Tax Exemptions → + New Exemption** and link it to the tax code.

| Field | Description |
|---|---|
| Exemption Code | e.g., `EXM-EXPORT`, `EXM-GOV` |
| Exemption Type | Total / Partial / Temporary / Conditional |
| Customer Category | e.g., Government agencies, Exporters, NGOs |
| Product Category | e.g., Basic foodstuffs, Medical supplies |
| Exemption Percentage | 100% = fully exempt; 50% = 50% reduction |
| Certificate Required | Yes/No |
| Valid Jurisdictions | Countries and regions where exemption applies |

#### Step 5 — Activate & Test

1. Set the tax code status to **Active**.
2. Use **Test Tax Calculation** (available on the tax code record) to run a sample transaction and verify the calculation output before it goes live.
3. Review the audit log to confirm the tax code creation was recorded.

### 2.4 Rate Update Workflow

When a government announces a rate change:

1. Open the existing tax code.
2. Navigate to **Rate History → + Add Rate**.
3. Enter the new rate and its **Effective From** date.
4. Save. The old rate automatically gets an expiry date of one day before the new effective date.
5. All new transactions from the effective date will use the new rate. Historical transactions remain unchanged.

> **Important:** Never edit the current rate field directly. Always add a new rate history entry. This preserves the audit trail and ensures correct recalculation of historical periods if needed.

---

## 3. VAT / GST

### 3.1 How VAT Works in the System

VAT is a transactional tax calculated at the point of sale (output VAT) and purchase (input VAT). The net VAT payable to the authority is output VAT minus reclaimable input VAT.

The engine handles:
- **Output VAT** on sales invoices, credit notes, and debit notes
- **Input VAT** on supplier invoices and expenses
- **Reverse charge VAT** for cross-border digital services and B2B imports
- **Zero-rated supplies** with full input VAT reclaim
- **Exempt supplies** with no VAT and no input reclaim
- **Partial exemption** calculations where an entity makes both taxable and exempt supplies

### 3.2 VAT Rates by Jurisdiction

| Jurisdiction | Standard Rate | Reduced Rate | Zero Rate | Registration Threshold |
|---|---|---|---|---|
| Kenya | 16% | — | 0% | KES 5,000,000/year |
| Uganda | 18% | — | 0% | UGX 150,000,000/year |
| Tanzania | 18% | — | 0% | TZS 200,000,000/year |
| Nigeria | 7.5% (VAT) | — | 0% | NGN 25,000,000/year |
| South Africa | 15% | — | 0% | ZAR 1,000,000/year |

### 3.3 VAT Calculation Logic

**Standard rated transaction:**
```
Net Amount         =  Gross Amount / (1 + VAT Rate)     [for tax-inclusive prices]
Net Amount         =  Line Qty x Unit Price              [for tax-exclusive prices]
VAT Amount         =  Net Amount x VAT Rate
Total              =  Net Amount + VAT Amount
```

**Example — Kenya, KES 100,000 net, 16% VAT:**
```
VAT Amount  =  100,000 x 0.16  =  16,000
Total        =  100,000 + 16,000  =  116,000
```

**Reverse charge (imported services):**
```
Self-assess output VAT   =  Invoice Amount x Local VAT Rate
Input VAT claimable      =  Same amount (if registered and making taxable supplies)
Net cash effect          =  Zero (for fully taxable businesses)
```

### 3.4 Zero-Rated vs Exempt

| Characteristic | Zero-Rated | Exempt |
|---|---|---|
| VAT charged to customer | 0% | None |
| Input VAT recoverable | Yes (fully) | No |
| Included in VAT return | Yes | Usually in separate column |
| Examples (Kenya) | Exports, unprocessed food, medical | Financial services, residential rent |

### 3.5 Partial Exemption

Where a business makes both taxable and exempt supplies, input VAT must be apportioned:

```
Recoverable Input VAT  =  Total Input VAT x (Taxable Turnover / Total Turnover)
```

The partial exemption calculation runs automatically at period end and posts an adjustment entry. The apportionment method (turnover-based by default) is configurable per entity.

### 3.6 VAT on Imports

Import VAT is calculated on the customs value plus import duty:

```
Import VAT Base  =  CIF Value + Import Duty + Excise Duty (if applicable)
Import VAT       =  Import VAT Base x VAT Rate
```

Import VAT is typically paid at the border but is reclaimable as input VAT in the next VAT return (subject to supporting documentation).

### 3.7 VAT Return Filing

VAT returns are filed monthly in most East African jurisdictions (quarterly optional for small taxpayers in some countries). The return summarises:

- Box 1: Standard-rated sales and output VAT
- Box 2: Zero-rated sales
- Box 3: Exempt sales
- Box 4: Total output VAT
- Box 5: Input VAT on purchases
- Box 6: Input VAT on imports
- Box 7: Net VAT payable / (refundable)

Refer to Section 11 (Tax Return Generation) for how the system auto-populates these boxes from transaction data.

---

## 4. Withholding Tax (WHT)

### 4.1 Overview

Withholding Tax is deducted at source by the payer on certain payment types and remitted directly to the tax authority on behalf of the recipient. The recipient receives a net payment and a WHT certificate they can use to offset their own tax liability.

The system handles WHT in three directions:
- **Deducting WHT** from payments to suppliers and service providers
- **Suffering WHT** when customers deduct from payments to the company
- **Remitting WHT** to the relevant authority within the statutory deadline

### 4.2 WHT Rates by Jurisdiction and Income Type

#### Kenya (KRA)

| Payment Type | Resident Rate | Non-Resident Rate |
|---|---|---|
| Dividends | 5% | 15% (treaty may reduce) |
| Interest (bank) | 15% | 15% |
| Interest (other) | 15% | 15% |
| Royalties | 5% | 20% |
| Management / Professional Fees | 5% | 20% |
| Consultancy Fees | 5% | 20% |
| Construction contracts | 3% | 20% |
| Rent (commercial) | 10% | 30% |
| Winnings / gambling | 20% | 20% |
| Digital content monetisation | 5% | 20% |

#### Uganda (URA)

| Payment Type | Resident Rate | Non-Resident Rate |
|---|---|---|
| Dividends | 15% | 15% |
| Interest | 15% | 15% |
| Royalties | 15% | 15% |
| Professional fees | 6% | 15% |
| Construction | 6% | 15% |

#### Tanzania (TRA)

| Payment Type | Resident Rate | Non-Resident Rate |
|---|---|---|
| Dividends (listed company) | 5% | 5% |
| Dividends (other) | 10% | 10% |
| Interest | 10% | 10% |
| Royalties | 15% | 15% |
| Service fees | 5% | 15% |
| Insurance premiums | 5% | 5% |

#### Nigeria (FIRS)

| Payment Type | Rate |
|---|---|
| Dividends | 10% |
| Interest | 10% |
| Royalties | 10% |
| Director fees | 10% |
| Management fees (to non-resident) | 10% |
| Technical / consulting fees | 5% |
| Construction (over NGN 50M) | 5% |

#### South Africa (SARS)

| Payment Type | Resident Rate | Non-Resident Rate |
|---|---|---|
| Dividends tax | 20% | 20% (treaty may reduce) |
| Royalties (non-resident) | — | 15% |
| Interest (non-resident) | — | 15% |
| PAYE (resident employees) | Sliding scale | — |

### 4.3 WHT Calculation Logic

**On supplier payments:**
```
Gross Invoice Amount     =  Service / payment value
WHT Amount               =  Gross Amount x Applicable WHT Rate
Net Payment to Supplier  =  Gross Amount - WHT Amount
WHT Payable to Authority =  WHT Amount (due by deadline)
```

**Example — Kenya, professional fees KES 200,000:**
```
WHT Rate         =  5%
WHT Amount       =  200,000 x 0.05  =  10,000
Net to Supplier  =  200,000 - 10,000  =  190,000
```

### 4.4 WHT on Payments Received (Suffered WHT)

When a customer deducts WHT from a payment to the company:

1. The system records the gross invoice amount as revenue.
2. The net cash received is posted to the bank.
3. The difference is recorded as **WHT Receivable** (a recoverable tax credit).
4. At year end, WHT receivable offsets corporate income tax payable.

### 4.5 Double Tax Treaty (DTT) Rates

Where a double tax treaty is in force, reduced WHT rates may apply. The system stores treaties with effective-dated reduced rates. To apply treaty rates:

1. The payer must hold a **Tax Residency Certificate** from the recipient.
2. The recipient's country must be configured in the treaty table.
3. The income type must fall within the treaty's coverage.

The engine automatically applies the lower of the domestic rate or the treaty rate when a valid treaty residency certificate is on file for that vendor.

### 4.6 WHT Remittance Deadlines

| Jurisdiction | Deadline |
|---|---|
| Kenya | 20th of the following month |
| Uganda | 15th of the following month |
| Tanzania | 7th of the following month |
| Nigeria | 21st of the following month |
| South Africa | 7th of the following month (dividends tax) |

### 4.7 WHT Exemptions

WHT exemptions apply in the following situations and must be supported by documentation:

- **Exempt body** — government ministries, public universities, and listed companies in some jurisdictions
- **Threshold** — some jurisdictions exempt payments below a minimum amount (e.g., Uganda exempts professional fees below UGX 1,000,000)
- **Exempt income type** — certain categories of payment are explicitly excluded from WHT in the tax legislation

To configure an exemption, attach an exemption certificate to the vendor record. The engine will check for valid certificates at payment time and suppress WHT deduction where a valid certificate exists.

---

## 5. Payroll Tax (PAYE, NSSF, NHIF/SHIF, SDL, NITA)

### 5.1 Overview

Payroll taxes are calculated per employee per pay run. The system processes all applicable statutory deductions simultaneously and generates the payroll tax schedule, GL journals, and remittance reports for each authority.

### 5.2 PAYE — Pay As You Earn

#### Kenya PAYE

Kenya PAYE uses a progressive tax table applied to the employee's monthly chargeable income (gross pay minus allowable deductions).

**Personal Relief:** KES 2,400 per month  
**Insurance Relief:** 15% of premiums paid (max KES 5,000/month)

**Tax Bands (2024):**

| Monthly Chargeable Income (KES) | Rate |
|---|---|
| 0 — 24,000 | 10% |
| 24,001 — 32,333 | 25% |
| 32,334 — 500,000 | 30% |
| 500,001 — 800,000 | 32.5% |
| Above 800,000 | 35% |

**PAYE Calculation:**
```
1. Gross Pay
2. Less: Allowable deductions (NSSF employee, mortgage interest, pension contributions)
3. = Chargeable Income
4. Apply progressive tax bands → Gross Tax
5. Less: Personal Relief (KES 2,400)
6. Less: Insurance Relief (if applicable)
7. = Net PAYE Payable
```

#### Uganda PAYE

**Monthly Tax Bands (UGX):**

| Monthly Chargeable Income | Rate |
|---|---|
| 0 — 335,000 | 0% |
| 335,001 — 410,000 | 10% |
| 410,001 — 10,000,000 | 20% |
| Above 10,000,000 | 30% (plus UGX 1,713,000 fixed) |

#### Tanzania PAYE

**Monthly Tax Bands (TZS):**

| Monthly Chargeable Income | Rate |
|---|---|
| 0 — 270,000 | 0% |
| 270,001 — 520,000 | 8% |
| 520,001 — 760,000 | 20% |
| 760,001 — 1,000,000 | 25% |
| Above 1,000,000 | 30% |

#### Nigeria PAYE

Nigeria PAYE is governed by the Personal Income Tax Act (PITA). The Consolidated Relief Allowance (CRA) is the higher of NGN 200,000 or 1% of gross income, plus 20% of gross income.

**Tax Bands (Annual):**

| Annual Taxable Income (NGN) | Rate |
|---|---|
| First 300,000 | 7% |
| Next 300,000 | 11% |
| Next 500,000 | 15% |
| Next 500,000 | 19% |
| Next 1,600,000 | 21% |
| Above 3,200,000 | 24% |

#### South Africa PAYE

South Africa uses SARS-published tax tables updated annually. The system stores the full SARS table and applies the applicable bracket. Employees are also entitled to primary, secondary (65+), and tertiary (75+) rebates.

**2024/25 Annual Tax Brackets:**

| Annual Taxable Income (ZAR) | Rate |
|---|---|
| 0 — 237,100 | 18% |
| 237,101 — 370,500 | 26% |
| 370,501 — 512,800 | 31% |
| 512,801 — 673,000 | 36% |
| 673,001 — 857,900 | 39% |
| 857,901 — 1,817,000 | 41% |
| Above 1,817,000 | 45% |

### 5.3 NSSF — National Social Security Fund

#### Kenya NSSF (Post-2023 Act)

Under the NSSF Act 2013 (fully operational from February 2023):

| Component | Employee | Employer |
|---|---|---|
| Tier I (up to Lower Earnings Limit, KES 7,000) | 6% | 6% |
| Tier II (between LEL and Upper Earnings Limit, KES 36,000) | 6% | 6% |
| Maximum monthly contribution | KES 2,160 | KES 2,160 |

> Note: The NSSF rates have been subject to legal challenge. The system stores the legally current rate and allows administrators to update it promptly when court orders or legislative changes occur.

#### Uganda NSSF

| Component | Employee | Employer |
|---|---|---|
| Employee contribution | 5% of gross | — |
| Employer contribution | — | 10% of gross |

#### Tanzania NSSF / PSSSF

| Scheme | Employee | Employer |
|---|---|---|
| NSSF | 10% | 10% |
| PSSSF (Public sector) | 5% | 15% |

### 5.4 NHIF / SHIF — Health Insurance (Kenya)

#### SHIF (Social Health Insurance Fund) — Effective October 2024

SHIF replaces NHIF. The contribution rate is **2.75% of gross salary** with no cap, for both employees and employers.

**Previous NHIF scale (retained for reference and historical payroll):**

| Monthly Gross (KES) | Monthly Deduction (KES) |
|---|---|
| Up to 5,999 | 150 |
| 6,000 — 7,999 | 300 |
| 8,000 — 11,999 | 400 |
| 12,000 — 14,999 | 500 |
| 15,000 — 19,999 | 600 |
| 20,000 — 24,999 | 750 |
| 25,000 — 29,999 | 850 |
| 30,000 — 49,999 | 900 |
| 50,000 — 99,999 | 950 |
| 100,000 and above | 1,050 |

### 5.5 SDL — Skills Development Levy

**Kenya:** 0.5% of gross monthly payroll (employer only). Remitted to the National Industrial Training Authority (NITA).

**Tanzania:** 4.5% of gross payroll (employer only). Split between VETA (3.5%) and SDL (1%).

### 5.6 NITA — National Industrial Training Authority

**Kenya only.** NITA levy is part of SDL (see above). Employers can reclaim NITA contributions as training grants upon submitting approved training records.

### 5.7 Housing Levy (Kenya — Affordable Housing)

Effective March 2024: **1.5% of gross salary** deducted from employee, matched by 1.5% employer contribution. Remitted to the Kenya Revenue Authority.

### 5.8 Payroll Tax Remittance Deadlines

| Deduction | Kenya | Uganda | Tanzania |
|---|---|---|---|
| PAYE | 9th of following month | 15th | 7th |
| NSSF | 9th of following month | 15th | 7th |
| NHIF/SHIF | 9th of following month | N/A | N/A |
| SDL / NITA | 9th of following month | N/A | 7th |
| Housing Levy | 9th of following month | N/A | N/A |

### 5.9 Payroll Tax Summary Report

At each pay run, the system generates a **Payroll Tax Schedule** that shows per-employee:

- Gross pay
- PAYE computed
- NSSF employee and employer
- NHIF / SHIF employee and employer
- Housing Levy employee and employer
- SDL (employer)
- Net pay

And a **Remittance Summary** grouped by authority showing total amounts due per authority, the due date, and the payment reference format.

---

## 6. Excise Duty

### 6.1 Overview

Excise duty is a consumption tax levied on the manufacture, production, or importation of specific goods and, increasingly, services. It is typically charged upstream (at manufacturer or importer level) but cascades into the final price.

### 6.2 Applicable Products and Services

Common excise-dutiable items in the covered jurisdictions:

| Category | Examples |
|---|---|
| Alcoholic beverages | Beer, wine, spirits |
| Tobacco products | Cigarettes, cigars, smokeless tobacco |
| Petroleum products | Petrol, diesel, kerosene, lubricants |
| Motor vehicles | Passenger cars, SUVs |
| Mobile airtime / data | Telecommunications services |
| Financial services | Fees on banking transactions |
| Cosmetics and beauty products | Selected items (Kenya) |
| Betting and gaming | Gross gaming revenue |
| Sugar-sweetened beverages | Soft drinks, juices |

### 6.3 Excise Duty Rates by Jurisdiction

#### Kenya

| Product | Rate |
|---|---|
| Beer | KES 121.85 per litre |
| Wines | KES 243.98 per litre |
| Spirits | KES 356.28 per litre |
| Cigarettes | KES 5,905.91 per 1,000 sticks |
| Fuel (petrol) | KES 25.35 per litre |
| Fuel (diesel) | KES 18.42 per litre |
| Mobile airtime | 15% of excisable value |
| Mobile data | 15% of excisable value |
| Mobile money transfer | 15% of excisable value |
| Betting | 15% of stake |
| Sugar-sweetened beverages | KES 10 per litre |

#### Uganda

| Product | Rate |
|---|---|
| Beer | 60% of ex-factory / import price |
| Spirits | 80% of ex-factory / import price |
| Cigarettes | 200% of ex-factory / import price |
| Fuel | UGX 1,450 per litre (petrol) |
| Mobile airtime | 12% |
| Mobile money | 0.5% per transaction |
| Sugar-sweetened beverages | UGX 200 per litre |

#### Nigeria

| Product | Rate |
|---|---|
| Beer and stout | 20% |
| Wines | 20% |
| Spirits | 20% |
| Cigarettes | 20% |
| Petroleum (fuel) | NGN 10 per litre |
| Telecommunications (airtime) | 5% |

### 6.4 Excise Duty Calculation Methods

The system supports three excise calculation methods:

**Specific (per unit):** Fixed amount per unit of quantity (litre, kilogram, stick, etc.)
```
Excise Duty  =  Quantity x Specific Rate
Example: 500 litres of beer x KES 121.85  =  KES 60,925
```

**Ad valorem (percentage of value):** Percentage applied to the excisable value
```
Excise Duty  =  Excisable Value x Ad Valorem Rate
Example: KES 10,000 airtime x 15%  =  KES 1,500
```

**Hybrid:** Both specific and ad valorem elements (whichever is higher, or combined)
```
Excise Duty  =  MAX(Specific Amount, Ad Valorem Amount)
            OR
Excise Duty  =  Specific Amount + Ad Valorem Amount
```

### 6.5 Excise on Manufactured Goods

For businesses that manufacture excise-dutiable goods, the system tracks:

- Raw material inputs (for potential excise exemption on inputs)
- Finished goods produced (trigger point for excise liability)
- Excise stamps / duty stamps issued and applied (Kenya requirement)
- Exports (zero excise for exported goods)

### 6.6 Excise Return Filing

Excise duty is generally filed monthly. The return captures:

- Opening stock of excisable goods
- Production / purchases during the period
- Closing stock
- Sales / removals from warehouse
- Excise duty payable
- Payments made
- Net balance

---

## 7. Import Duty / Customs

### 7.1 Overview

Import duties are levied on goods brought into a country and are calculated using the customs value of the goods (generally the CIF — Cost, Insurance, and Freight — value). Multiple levies typically apply simultaneously on the same import declaration.

### 7.2 Typical Import Levy Stack (Kenya Example)

For a dutiable import into Kenya, the following apply cumulatively:

| Levy | Base | Rate |
|---|---|---|
| Import Duty | CIF Value | 0% / 10% / 25% / 35% (by HS code) |
| Import Declaration Fee (IDF) | CIF Value | 3.5% (min KES 5,000) |
| Railway Development Levy (RDL) | CIF Value | 2% |
| Excise Duty (if applicable) | CIF + Import Duty | Varies |
| VAT | CIF + Duty + Excise + IDF + RDL | 16% |
| Anti-dumping duty (if applicable) | CIF Value | Case-specific |

**Total import cost calculation:**
```
CIF Value              =  Invoice + Insurance + Freight
Import Duty            =  CIF x Duty Rate
IDF                    =  CIF x 3.5%
RDL                    =  CIF x 2%
Excise (if applicable) =  (CIF + Import Duty) x Excise Rate
VAT Base               =  CIF + Import Duty + IDF + RDL + Excise
VAT                    =  VAT Base x 16%
Total Landed Cost      =  CIF + Import Duty + IDF + RDL + Excise + VAT
```

### 7.3 HS Code Classification

Every imported item must be classified under the Harmonised System (HS) code, which determines the import duty rate. The system maintains an HS code reference table:

| Field | Description |
|---|---|
| HS Code | 6–10 digit harmonised code |
| Description | Official goods description |
| Kenya Duty Rate | EAC Common External Tariff rate for Kenya |
| Uganda Duty Rate | Rate for Uganda |
| Tanzania Duty Rate | Rate for Tanzania |
| Nigeria Duty Rate | Nigerian Customs Tariff rate |
| South Africa Duty Rate | SARS tariff rate |
| Excise Applicable | Yes/No |
| Excise Rate | If applicable |
| Prohibited/Restricted | Flag and note |

The HS code table can be updated by administrators when tariff schedules change.

### 7.4 EAC Common External Tariff (CET)

Kenya, Uganda, and Tanzania are members of the East African Community (EAC) and apply the EAC Common External Tariff for imports from outside the EAC region. Imports between EAC member states are generally zero-rated (subject to rules of origin).

**EAC CET Rate Bands:**

| Band | Rate | Typical Goods |
|---|---|---|
| Category A | 0% | Capital goods, raw materials, medical equipment |
| Category B | 10% | Intermediate goods |
| Category C | 25% | Finished consumer goods |
| Sensitive items | 35%+ | Agricultural products, textiles |

### 7.5 Import Declaration in the System

When processing an import purchase order or supplier invoice involving cross-border goods, the user initiates the **Import Declaration Wizard**:

1. **Enter / confirm the CIF value** — the system pre-fills from the purchase order
2. **Select or confirm the HS code** for each item line
3. **The system calculates all applicable levies** (duty, IDF, RDL, excise, VAT) automatically
4. **Enter the customs entry number** (C17 or local equivalent) once issued
5. **Post the import cost and tax entries** — import duty is either expensed or capitalised depending on the asset type; import VAT posts to the VAT receivable account for reclaim
6. **Attach customs entry documents** (C17, invoice, packing list, bill of lading)

### 7.6 Customs Duty Exemptions

Common exemptions and special regimes:

| Regime | Description |
|---|---|
| Manufacturing Under Bond (MUB) | Raw materials for re-export production — duty suspended |
| Export Processing Zone (EPZ) | Duty-free on imports for approved EPZ manufacturers |
| Diplomatic exemption | Imports for approved diplomatic missions |
| Aid / donor goods | Pre-approved donor-funded goods |
| Temporary importation | Short-term imports for exhibitions, events, etc. |
| Goods under EAC preferential tariff | Reduced rates for EAC originating goods |

### 7.7 Transfer of Duty to Cost or VAT Reclaim

The system automatically handles the accounting treatment:

- **Import Duty:** Added to the cost of goods (capitalised into inventory or asset cost)
- **Import VAT:** Posted to VAT Receivable (reclaimable against output VAT in the next return)
- **IDF and RDL:** Expensed to an import levies cost account (not reclaimable as VAT)

---

## 8. Corporate Tax

### 8.1 Overview

Corporate Income Tax (CIT) is an annual (or instalment-based) tax on the taxable profits of a legal entity. The Tax Management module supports CIT through:

- Permanent and temporary difference tracking (for deferred tax)
- Tax-adjusted profit computation
- Instalment tax scheduling and payment tracking
- Year-end CIT return generation
- Transfer pricing documentation links

### 8.2 CIT Rates by Jurisdiction

| Jurisdiction | Standard Rate | Special Rates |
|---|---|---|
| Kenya | 30% (resident), 37.5% (non-resident branch) | 15% for newly listed companies (3 years); 10% for EPZ companies |
| Uganda | 30% | 25% for new companies with investment >USD 50M |
| Tanzania | 30% | 25% for companies listed on DSE |
| Nigeria | 30% (large), 20% (medium), 0% (small) | Small = turnover < NGN 25M |
| South Africa | 27% | Various incentive rates |

### 8.3 Taxable Profit Computation

```
Accounting Profit (per financial statements)
+ Disallowed expenses (entertainment > statutory limit, penalties, non-business expenses)
+ Depreciation (accounting)
- Capital allowances (tax depreciation)
- Tax-exempt income (e.g., certain dividends)
+/- Other adjustments per jurisdiction rules
= Taxable Income
x CIT Rate
= Gross CIT
- WHT suffered (offsettable)
- Instalment tax paid
= CIT balance payable / (refundable)
```

### 8.4 Capital Allowances (Tax Depreciation)

Tax depreciation rates differ from accounting depreciation. The system maintains a separate capital allowance register:

#### Kenya Investment Deduction & Wear and Tear

| Asset Class | Annual Rate |
|---|---|
| Industrial buildings | 10% straight-line (or 100% in year 1 for new industrial buildings outside Nairobi/Mombasa) |
| Plant and machinery | 12.5% straight-line |
| Computers and software | 30% straight-line |
| Motor vehicles | 25% reducing balance |
| Agricultural works | 50% reducing balance |

#### Nigeria Capital Allowance

| Asset Class | Initial Allowance | Annual Allowance |
|---|---|---|
| Building (industrial) | 15% | 10% |
| Plant and machinery | 50% | 25% |
| Motor vehicles | 50% | 25% |
| Agriculture | 95% | — |

### 8.5 Instalment Tax (Advance Corporate Tax)

Most jurisdictions require companies to pay CIT in instalments during the year based on estimated annual tax. The system generates an **Instalment Tax Schedule** at the beginning of each tax year.

**Kenya Instalment Tax Schedule:**

| Instalment | Due Date | Amount |
|---|---|---|
| 1st | 20th day of 4th month of accounting year | 25% of estimated annual CIT |
| 2nd | 20th day of 6th month | 25% |
| 3rd | 20th day of 9th month | 25% |
| 4th | 20th day of 12th month | 25% |

The system alerts finance when each instalment is due, tracks payments, and adjusts estimates if interim management accounts show a significant deviation from the original estimate.

### 8.6 Deferred Tax

The system computes deferred tax arising from temporary differences:

- **Deferred tax liability:** Where accounting income > taxable income (e.g., accounting depreciation < capital allowances)
- **Deferred tax asset:** Where accounting income < taxable income (e.g., provisions not yet deductible)

Deferred tax is recognised at the enacted CIT rate expected to apply when the timing difference reverses.

### 8.7 Transfer Pricing

For groups with intercompany transactions, the system links to the transfer pricing documentation module:

- **Intercompany transaction register** — logs all related-party transactions
- **Arm's length benchmarking** — stores the benchmarking study reference and approved range
- **Country-by-Country Reporting (CbCR)** — consolidates data required for CbCR submission (applicable where group revenue exceeds the local threshold)
- **Local file / Master file** — document upload and deadline tracking

---

## 9. e-Invoicing — KRA eTIMS Integration

### 9.1 Overview

The Kenya Revenue Authority Electronic Tax Invoice Management System (eTIMS) requires all registered VAT taxpayers to transmit invoice data to KRA in real time or near-real time at the point of issuance. The system integrates natively with eTIMS — invoice submission is part of the normal invoice-posting workflow, not a separate batch process.

### 9.2 Scope of eTIMS Obligation

eTIMS applies to:

- All sales invoices (including cash sales)
- Credit notes and debit notes
- Tax invoices for taxable supplies (VAT and non-VAT registered traders above the threshold)

eTIMS does **not** currently apply to purchase invoices (those are the seller's obligation).

### 9.3 eTIMS Integration Architecture

```
User Posts Invoice in ERP
         |
         v
Invoice Validation Engine
  (checks tax codes, amounts, KRA PIN format)
         |
         v
eTIMS Payload Builder
  (maps ERP fields to eTIMS JSON schema)
         |
         v
eTIMS API Call (HTTPS POST)
  -> https://etims-api.kra.go.ke/etims-api/
         |
    +----+----+
    | Success | -> Receive KRA Invoice Number (CU Invoice Number)
    |         |   Print/email invoice with KRA QR Code
    +---------+
    | Failure | -> Queue for retry (max 3 attempts)
    |         |   Alert finance team
    +---------+
         |
         v
eTIMS Submission Log
  (status, KRA number, timestamp, payload, response)
```

### 9.4 eTIMS Invoice Fields (Required)

| ERP Field | eTIMS Field | Notes |
|---|---|---|
| Invoice Number | `invcNo` | ERP sequential number |
| Invoice Date | `salesDt` | YYYYMMDD format |
| Customer KRA PIN | `rcptTyCd` + buyer PIN | Mandatory for B2B above KES 1,000 |
| Customer Name | `custNm` | |
| Item Description | `itemNm` | Per line |
| Item HS Code | `hsCd` | Required for goods |
| Quantity | `qty` | |
| Unit Price | `unitPrice` | |
| Taxable Amount | `taxblAmt` | |
| Tax Rate Code | `taxTyCd` | A (16%), B (0%), C (Exempt) |
| Tax Amount | `taxAmt` | |
| Total Amount | `totAmt` | |
| Invoice Type | `orgInvcNo` | For credit notes, original invoice CU number |

### 9.5 Handling eTIMS Failures

If eTIMS is unavailable (KRA system downtime), the system:

1. Stores the invoice in the **eTIMS pending queue** with status `queued`.
2. The invoice can be printed and shared with the customer with a note that the KRA number is pending.
3. On system recovery, the queue auto-retries in chronological order.
4. Once submitted, the invoice is updated with the KRA CU number and QR code.
5. A corrected invoice copy with the KRA number is available for reprint.

### 9.6 eTIMS QR Code

Every KRA-submitted invoice must display a QR code containing:

- KRA CU Invoice Number
- Seller KRA PIN
- Invoice Date
- Total Amount
- Tax Amount

The system generates the QR code automatically upon successful submission and embeds it in the invoice print template.

### 9.7 eTIMS for Credit Notes

When raising a credit note:

1. The original KRA CU Invoice Number (from the original invoice) must be referenced.
2. The eTIMS payload includes the `orgInvcNo` field with this reference.
3. KRA records the credit note against the original invoice in their system.

### 9.8 Analogous Integrations in Other Jurisdictions

| Country | System | Authority | Status |
|---|---|---|---|
| Kenya | eTIMS | KRA | Live integration |
| Uganda | EFRIS (Electronic Fiscal Receipting) | URA | API integration available |
| Tanzania | EFD (Electronic Fiscal Device) | TRA | Device-based; API bridge in roadmap |
| Nigeria | e-Invoice (FIRS) | FIRS | Pilot phase; integration in roadmap |
| South Africa | No mandatory e-invoice | SARS | Not required currently |

---

## 10. Multi-Jurisdiction Rate Tables

### 10.1 Overview

The system maintains a central **Tax Rate Repository** — a versioned, effective-dated store of all tax rates across all jurisdictions. This is the single source of truth used by the tax calculation engine.

### 10.2 Rate Table Structure

Each rate record contains:

| Field | Description |
|---|---|
| Tax Code | Unique code (e.g., `KE-VAT-STD`, `NG-WHT-DIV`) |
| Tax Type | VAT / WHT / PAYE / Excise / Customs / CIT |
| Jurisdiction | Country + State/County (if applicable) |
| Rate | Percentage or fixed amount |
| Effective From | Date from which this rate applies |
| Effective To | Date after which this rate no longer applies (null = current) |
| Authority | Governing authority |
| Legislative Reference | Relevant act or statutory instrument |
| Notes | Any conditions or caveats |

### 10.3 Current Rate Summary — VAT

| Country | Standard | Zero | Registration Threshold |
|---|---|---|---|
| Kenya | 16% | 0% | KES 5M/year |
| Uganda | 18% | 0% | UGX 150M/year |
| Tanzania | 18% | 0% | TZS 200M/year |
| Nigeria | 7.5% | 0% | NGN 25M/year |
| South Africa | 15% | 0% | ZAR 1M/year |

### 10.4 Current Rate Summary — Corporate Tax

| Country | Standard CIT | Branch Rate | Small Business |
|---|---|---|---|
| Kenya | 30% | 37.5% | 30% |
| Uganda | 30% | 30% | 30% |
| Tanzania | 30% | 30% | 30% |
| Nigeria | 30% | 30% | 0% (<NGN 25M) |
| South Africa | 27% | 27% | Progressive |

### 10.5 Currency Reference

| Country | Currency | ISO Code |
|---|---|---|
| Kenya | Kenyan Shilling | KES |
| Uganda | Ugandan Shilling | UGX |
| Tanzania | Tanzanian Shilling | TZS |
| Nigeria | Naira | NGN |
| South Africa | Rand | ZAR |

All tax calculations are performed in the transaction currency and converted to base currency using the exchange rate at the transaction date.

---

## 11. Tax Return Generation

### 11.1 Overview

The **Return Generation Engine** automatically aggregates posted tax transactions for a period and compiles them into the format required for each authority's return. Prepared returns are reviewed, approved, and (where electronic filing is enabled) submitted directly from the system.

### 11.2 Supported Return Types

| Return | Jurisdiction | Frequency | Filing Method |
|---|---|---|---|
| VAT Return | KE, UG, TZ, NG, ZA | Monthly | Electronic (eTIMS / authority portal) |
| WHT Return | KE, UG, TZ, NG, ZA | Monthly | Electronic |
| PAYE Return | KE, UG, TZ, NG, ZA | Monthly | Electronic |
| NSSF Return | KE, UG, TZ | Monthly | Electronic |
| NHIF / SHIF Return | KE | Monthly | Electronic |
| Housing Levy Return | KE | Monthly | Electronic |
| Excise Return | KE, UG, NG | Monthly | Electronic |
| Provisional CIT | KE, UG, TZ, NG, ZA | Quarterly/4x annual | Electronic |
| Annual CIT Return | KE, UG, TZ, NG, ZA | Annual (6 months after year end) | Electronic |

### 11.3 Return Generation Workflow

```
1. Period Close
   └── Finance team closes the tax period in the system

2. Auto-generate Return (or manually trigger)
   └── System queries all tax transactions for the period
   └── Aggregates by return type and authority

3. Review Worksheet
   └── Finance reviews the auto-populated return lines
   └── Can view the supporting transactions behind each line
   └── Can make manual adjustments with override reasons

4. Preparer Sign-off
   └── Preparer marks return as "Prepared"

5. Reviewer Approval
   └── Tax manager reviews and approves

6. File Return
   └── System submits via API (electronic) or exports PDF/XML for manual submission
   └── Confirmation number recorded

7. Payment
   └── Payment scheduled and recorded
   └── GL journal posted (Debit: Tax Payable / Credit: Bank)

8. Period Locked
   └── Period marked Closed; no further transactions can be posted
```

### 11.4 VAT Return Auto-Population Logic

| Return Box | System Source |
|---|---|
| Total Sales | Sum of all sales invoice net amounts in period |
| Standard-rated sales | Sales with VAT code = Standard |
| Zero-rated sales | Sales with VAT code = Zero-rated |
| Exempt sales | Sales with VAT code = Exempt |
| Output VAT | Sum of VAT amounts on standard-rated sales |
| Input VAT on purchases | Sum of VAT on purchase invoices (reclaimable) |
| Input VAT on imports | Sum of import VAT posted from customs entries |
| Input VAT adjustments | Manual adjustments (credit notes, bad debt relief) |
| Net VAT Payable | Output VAT minus Total Input VAT |

### 11.5 Return Amendments

Where a filed return requires correction:

1. Open the original return and select **Amend**.
2. The system creates an **Amended Return** linked to the original.
3. Amendments show only the corrected lines with the delta amounts.
4. The amendment is submitted to the authority through the same filing channel.
5. Any additional tax due generates a payment record; overpayments generate a refund claim.

### 11.6 Penalty and Interest Calculation

If a return is filed or paid late, the system calculates:

```
Late Filing Penalty  =  Fixed penalty per authority rules
Late Payment Penalty =  Outstanding Tax x Penalty Rate per month (or part thereof)
Interest             =  Outstanding Tax x Interest Rate per annum x Days overdue / 365
Total Amount Due     =  Tax + Late Filing Penalty + Late Payment Penalty + Interest
```

Penalty and interest rates are stored per authority and updated by administrators when the authority publishes changes.

---

## 12. WHT Certificates

### 12.1 Overview

A WHT certificate (also called a WHT credit note or tax deduction certificate) is issued by the payer to the payee confirming how much WHT was deducted from a payment. The payee uses this certificate to claim a tax credit against their own income tax liability.

### 12.2 Certificate Generation

The system auto-generates a WHT certificate for every payment on which WHT is deducted. Certificates are generated at payment posting time.

**WHT Certificate Contents:**

| Field | Description |
|---|---|
| Certificate Number | Unique sequential number |
| Payer Name | Company making the payment |
| Payer Tax PIN | Payer's KRA/URA/TRA/FIRS/SARS registration number |
| Payee Name | Vendor / recipient name |
| Payee Tax PIN | Recipient's tax registration number |
| Payment Date | Date of payment |
| Payment Description | Nature of payment (e.g., consultancy fees) |
| Gross Amount | Invoice amount before WHT |
| WHT Rate | Rate applied |
| WHT Amount | Amount deducted |
| Net Amount Paid | Amount actually remitted to payee |
| Tax Period | Month and year of WHT remittance |
| Signature / Stamp | Authorised signatory |

### 12.3 Certificate Distribution

1. **Email delivery:** The system automatically emails the certificate to the vendor's registered email upon payment posting (if configured).
2. **Vendor portal:** Vendors can log in to the supplier portal and download all their WHT certificates for any period.
3. **Bulk export:** Finance can export all certificates for a period as a ZIP of PDFs for physical distribution or manual email.

### 12.4 WHT Suffered (Certificates Received)

When a customer deducts WHT from a payment to the company:

1. Finance records the customer's WHT certificate number in the payment receipt.
2. The system posts: Debit Bank, Debit WHT Receivable, Credit Accounts Receivable (full invoice amount).
3. The WHT Receivable register tracks all certificates received, the certificate number, the customer, the period, and the amount.
4. At year end, the WHT Receivable balance is presented to the CIT computation as a tax credit.

### 12.5 WHT Certificate Reconciliation

The **WHT Reconciliation Report** matches:

- WHT deducted per the payroll / payment system
- WHT remitted per the WHT return filed
- WHT certificates issued to vendors
- WHT certificates received from customers

Discrepancies are flagged for investigation before the annual WHT return is finalised.

---

## 13. Compliance Calendar & Alerts

### 13.1 Automated Compliance Calendar

The system generates a dynamic tax compliance calendar for each entity, based on the jurisdictions and tax types they are registered for. The calendar is visible in the Tax Management dashboard.

### 13.2 Sample Monthly Obligations — Kenya

| Obligation | Deadline | Days Before Alert | Filing Method |
|---|---|---|---|
| PAYE, NSSF, NHIF/SHIF, SDL, Housing Levy | 9th of following month | 5 days prior | iTax / eCitizen |
| VAT Return and payment | 20th of following month | 7 days prior | eTIMS / iTax |
| WHT Return and payment | 20th of following month | 7 days prior | iTax |
| Excise Return | 20th of following month | 7 days prior | iTax |

### 13.3 Alert Configuration

For each compliance obligation, administrators can configure:

- **Alert recipients** (by role or by named user)
- **Lead time** (how many days before the deadline to send the first alert)
- **Escalation** (a second alert if the return has not been marked filed X days before deadline)
- **Alert channel** (in-app notification, email, or both)

### 13.4 Compliance Status Dashboard

The dashboard shows, for the current and next month:

| Obligation | Period | Due Date | Status | Prepared By | Filed Date |
|---|---|---|---|---|---|
| Kenya VAT | March 2025 | 20 Apr 2025 | Prepared | J. Mwangi | — |
| Kenya PAYE | March 2025 | 9 Apr 2025 | Filed | A. Kamau | 7 Apr 2025 |
| Uganda VAT | March 2025 | 15 Apr 2025 | Open | — | — |

Status values: `Open` → `In Progress` → `Prepared` → `Filed` → `Paid` → `Closed`

---

## 14. Data Model Reference

### 14.1 Core Tables

| Table | Description |
|---|---|
| `tax_authorities` | Government tax bodies (KRA, URA, TRA, FIRS, SARS) |
| `tax_codes` | Individual tax codes with rates and rules |
| `tax_rate_history` | Effective-dated rate changes for each tax code |
| `tax_brackets` | Progressive rate brackets (for PAYE and CIT) |
| `tax_exemptions` | Exemption rules linked to tax codes |
| `tax_exemption_certificates` | Customer/vendor exemption certificates on file |
| `tax_transactions` | Tax calculated per source transaction |
| `tax_transaction_lines` | Line-level tax breakdown |
| `tax_return_periods` | Filing period definitions per authority |
| `tax_returns` | Filed returns with summary amounts |
| `tax_return_lines` | Individual return boxes/lines |
| `tax_payments` | Tax payments to authorities |
| `wht_certificates` | WHT certificates issued to vendors |
| `wht_certificates_received` | WHT certificates received from customers |
| `hs_codes` | Harmonised System code reference table with duty rates |
| `import_declarations` | Customs entries linked to purchase orders |
| `tax_treaties` | Double tax treaties with reduced WHT rates |
| `withholding_tax_rates` | WHT rates by country pair and income type |
| `etims_submission_log` | eTIMS API call log (Kenya) |
| `transfer_pricing_docs` | Transfer pricing documentation tracker |
| `compliance_calendar` | Generated compliance obligations per entity |

### 14.2 Key Relationships

```
tax_authorities ──< tax_codes ──< tax_rate_history
                                └─< tax_brackets
                                └─< tax_exemptions

tax_codes >── tax_transaction_lines ──< tax_transactions
                                         └── [source_transaction_id] ->
                                             sales_invoices / purchase_invoices /
                                             payroll_runs / expense_claims

tax_return_periods ──< tax_returns ──< tax_return_lines
                                   └─< tax_payments

tax_transactions ──> tax_return_lines (via period aggregation)

customers / vendors ──< tax_exemption_certificates
                     └─< wht_certificates
                     └─< wht_certificates_received
```

### 14.3 GL Account Mapping Requirements

Every tax code must be mapped to at least one GL account. The expected mapping:

| Tax Code Type | Required GL Mapping |
|---|---|
| Output VAT | Tax Payable (Liability) |
| Input VAT (reclaimable) | Tax Receivable (Asset) |
| Input VAT (non-reclaimable) | Tax Expense (Expense) |
| WHT payable | WHT Payable (Liability) |
| WHT receivable | WHT Receivable (Asset) |
| PAYE payable | PAYE Payable (Liability) |
| NSSF / NHIF employer | Payroll Tax Expense (Expense) |
| NSSF / NHIF payable | Social Contribution Payable (Liability) |
| Excise duty | Excise Payable (Liability) or Cost of Sales (Expense) |
| Import duty | Capitalised into asset/inventory cost |
| Import VAT | Tax Receivable (Asset) |
| CIT | Income Tax Payable (Liability) |
| Deferred tax | Deferred Tax Asset / Liability |

---

*End of Tax Management Module Documentation*

*This document covers configuration, calculation logic, multi-jurisdiction rates, e-invoicing, return generation, WHT certificates, and the data model. Implementation details for each jurisdiction should be validated against the current legislation of the respective tax authority prior to going live.*
