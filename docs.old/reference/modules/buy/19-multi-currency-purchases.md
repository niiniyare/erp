[<-- Back to Index](README.md)

## Multi-Currency Purchases

### Foreign Currency Setup

```markdown
CURRENCY CONFIGURATION

Active Currencies:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Base Currency: KES (Kenyan Shilling)

Foreign Currencies Enabled:
┌──────────┬─────────────────────┬──────────────┬──────────┐
│ Currency │ Description         │ Symbol       │ Status   │
├──────────┼─────────────────────┼──────────────┼──────────┤
│ USD      │ US Dollar           │ $            │ Active   │
│ EUR      │ Euro                │ €            │ Active   │
│ GBP      │ British Pound       │ £            │ Active   │
│ CNY      │ Chinese Yuan        │ ¥            │ Active   │
│ JPY      │ Japanese Yen        │ ¥            │ Active   │
└──────────┴─────────────────────┴──────────────┴──────────┘

Exchange Rate Management:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rate Source: Central Bank of Kenya (CBK) Official Rates
Update Frequency: Daily (automated)
Manual Override: Allowed (with approval)

Exchange Rate Table (2025-04-15):
┌──────────┬────────────┬────────────┬─────────────┐
│ Currency │ Buying Rate│ Selling Rate│ Mid Rate   │
├──────────┼────────────┼────────────┼─────────────┤
│ USD/KES  │  128.50    │  129.50    │  129.00    │
│ EUR/KES  │  140.20    │  141.50    │  140.85    │
│ GBP/KES  │  163.80    │  165.20    │  164.50    │
│ CNY/KES  │   17.85    │   18.05    │   17.95    │
│ JPY/KES  │    0.86    │    0.87    │    0.865   │
└──────────┴────────────┴────────────┴─────────────┘

Rate Type Selection:
  For Purchases: Buying Rate (bank buys foreign currency)
  For Sales: Selling Rate
  Default: Mid Rate (if not specified)

Example:
  Purchase in USD: Use USD Buying Rate (128.50)
  Reason: Bank buys USD from you to pay supplier
```

### Foreign Currency Purchase Orders

```markdown
FOREIGN CURRENCY PO WORKFLOW

Creating Foreign Currency PO:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scenario: Import CNC Machine from Germany
Date: 2025-04-15
Supplier: Precision Tech GmbH (Germany)

Purchase Order: PO-2025-00325
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Supplier: Precision Tech GmbH
         Munich, Germany
         VAT No: DE123456789

Currency: EUR
Exchange Rate: 140.85 KES/EUR (mid-rate, 2025-04-15)
Rate Type: Fixed at PO creation

Items:
┌────────────────────────┬─────┬────────────┬────────────┐
│ Description            │ Qty │ Rate (EUR) │ Amount     │
├────────────────────────┼─────┼────────────┼────────────┤
│ CNC Machine Model X500 │  1  │  80,000.00 │  80,000.00 │
│                        │     │            │            │
│ Subtotal (EUR)         │     │            │  80,000.00 │
│ Freight (EUR)          │     │            │   3,500.00 │
│ Insurance (EUR)        │     │            │     500.00 │
│                        │     │            │            │
│ TOTAL (EUR)            │     │            │  84,000.00 │
│                        │     │            │            │
│ Exchange Rate          │     │ 140.85     │            │
│ TOTAL (KES)            │     │            │11,831,400  │
└────────────────────────┴─────┴────────────┴────────────┘

Payment Terms: 30% Advance, 70% on Delivery
  Advance Payment: 25,200.00 EUR (3,549,420 KES)
  Balance Payment: 58,800.00 EUR (8,281,980 KES)

Incoterms: CIF Mombasa
  Cost, Insurance, Freight included to Mombasa Port

Exchange Rate Booking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Option 1: Spot Rate at Payment
  Risk: Rate may change unfavorably
  Example: EUR strengthens to 145 KES/EUR
  Impact: 84,000 × 145 = 12,180,000 KES (+348,600 KES)
  
Option 2: Forward Contract (Hedge)
  Book rate now for future payment
  Bank charges: ~0.5% of transaction
  Rate locked: 141.00 KES/EUR
  Total: 84,000 × 141 = 11,844,000 KES
  Hedging cost: ~60,000 KES
  Protected against adverse moves

Decision: Forward contract for balance payment
  Advance (immediate): Spot rate
  Balance (60 days): Forward rate 141.00

System Entries:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PO Created:
  Dr. PO Commitment (Budget) - EUR    84,000.00
  Dr. PO Commitment (Budget) - KES              11,831,400
      Cr. Budget Reserve                                    11,831,400

Description: PO-2025-00325, CNC Machine import
Currency: EUR | Rate: 140.85 | Date: 2025-04-15
```

### Advance Payment Handling

```markdown
ADVANCE PAYMENT PROCESS

Advance Payment Request:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PO: PO-2025-00325
Amount Due: 30% × 84,000 EUR = 25,200 EUR
Date: 2025-04-20

Bank Quotation:
  Amount: 25,200 EUR
  Exchange Rate: 140.50 KES/EUR (spot rate today)
  KES Equivalent: 3,540,600 KES
  Bank Charges: 15,000 KES (wire transfer)
  Total Payment: 3,555,600 KES

Payment Approval:
  Requested By: Procurement Manager
  Approved By: CFO
  Reason: 30% advance per contract terms
  Payment Method: International Wire Transfer

Payment Execution:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: 2025-04-20

Wire Transfer Details:
  Beneficiary: Precision Tech GmbH
  Bank: Deutsche Bank, Munich
  IBAN: DE89370400440532013000
  SWIFT: DEUTDEFF
  Amount: 25,200.00 EUR
  Reference: PO-2025-00325 Advance Payment

Accounting Entry:
  Dr. Advance to Suppliers - EUR      25,200.00
  Dr. Advance to Suppliers - KES                  3,540,600
  Dr. Bank Charges                                   15,000
      Cr. Bank Account - KES                                 3,555,600

Description: Advance payment PO-2025-00325
Rate Used: 140.50 KES/EUR
Rate Variance vs PO: 140.50 vs 140.85 = -0.35 KES/EUR (favorable)
Variance Amount: 25,200 × 0.35 = 8,820 KES gain

Exchange Gain Booking:
  Dr. Advance to Suppliers                         8,820
      Cr. Foreign Exchange Gain                            8,820

Description: Favorable forex variance on advance payment
```

### Goods Receipt & Invoicing

```markdown
IMPORT CLEARANCE & RECEIPT

Shipment Arrival:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: 2025-06-10
Port: Mombasa
Documents:
  ✓ Bill of Lading
  ✓ Commercial Invoice
  ✓ Packing List
  ✓ Certificate of Origin
  ✓ Insurance Certificate

Customs Clearance:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customs Valuation:
  CIF Value: 84,000 EUR
  Exchange Rate (Customs): 141.20 KES/EUR (CBK rate on arrival)
  CIF Value (KES): 11,860,800 KES

Import Duties & Taxes:
┌──────────────────────┬──────────┬──────────┬────────────┐
│ Item                 │ Rate     │ Base     │ Amount (KES)│
├──────────────────────┼──────────┼──────────┼────────────┤
│ CIF Value            │    -     │    -     │ 11,860,800 │
│                      │          │          │            │
│ Import Duty          │  10%     │ CIF      │  1,186,080 │
│ Excise Duty          │   0%     │    -     │          0 │
│                      │          │          │            │
│ Subtotal (Dutiable)  │    -     │    -     │ 13,046,880 │
│                      │          │          │            │
│ VAT                  │  16%     │ Dutiable │  2,087,501 │
│ IDF (Import Fee)     │  3.5%    │ CIF      │    415,128 │
│ Railway Levy         │  2%      │ CIF      │    237,216 │
│                      │          │          │            │
│ TOTAL CLEARANCE COST │          │          │  3,925,925 │
└──────────────────────┴──────────┴──────────┴────────────┘

Other Costs:
  Clearing Agent Fee: 150,000 KES
  Port Charges: 85,000 KES
  Transport to Factory: 120,000 KES
  Total Logistics: 355,000 KES

Total Landed Cost:
  CIF Value: 11,860,800 KES
  Import Duties/Taxes: 3,925,925 KES
  Logistics: 355,000 KES
  ─────────────────────────────
  TOTAL: 16,141,725 KES

Goods Receipt Note:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
GRN No: GRN-2025-0345
Date: 2025-06-15
Warehouse: WH-NBI-01

Item: CNC Machine Model X500
Condition: Good (inspected by engineer)
Serial No: PTG-X500-2025-1234

Accounting Entry (GRN):
  Dr. Capital Equipment (Asset)               16,141,725
  Dr. VAT Input (Recoverable)                  2,087,501
      Cr. GRN Clearing Account                         18,229,226

Description: GRN-2025-0345, CNC Machine import
Includes: Machine, duties, logistics

Supplier Invoice Processing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice No: PTG-INV-2025-456
Date: 2025-06-10
Amount: 84,000 EUR

Balance Due: 58,800 EUR (70% - already paid 30% advance)

Forward Contract Execution:
  Date: 2025-06-20
  Amount: 58,800 EUR
  Contracted Rate: 141.00 KES/EUR (booked at PO time)
  KES Amount: 8,290,800 KES
  Bank Charges: 20,000 KES
  Total Payment: 8,310,800 KES

If paid at spot rate (say 143.00):
  Amount: 58,800 × 143 = 8,408,400 KES
  Savings from hedging: 117,600 KES ✓

Invoice Accounting:
  Dr. GRN Clearing Account                    18,229,226
      Cr. Advance to Suppliers (applied)                3,540,600
      Cr. Accounts Payable - EUR                       14,688,626

Balance payment:
  Dr. Accounts Payable - EUR                   8,290,800
  Dr. Bank Charges                                20,000
      Cr. Bank Account                                  8,310,800

Final Exchange Variance Summary:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PO Rate: 140.85
Advance Rate: 140.50 (gain 8,820 KES) ✓
Balance Rate: 141.00 (small loss vs PO)
Weighted Average Rate: ~140.85

Net Forex Impact: Minimal (hedging effective) ✓
```

### Exchange Rate Differences

```markdown
FOREX GAIN/LOSS ACCOUNTING

Unrealized Gains/Losses:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Open Foreign Currency Payables (Month-End Revaluation)

Date: 2025-04-30
Supplier: Shanghai Tools Ltd
Original Invoice:
  Date: 2025-04-01
  Amount: 50,000 USD
  Rate: 128.00 KES/USD
  KES Value: 6,400,000 KES

Month-End Rate: 129.50 KES/USD
Revalued Amount: 50,000 × 129.50 = 6,475,000 KES

Unrealized Loss: 75,000 KES

Accounting Entry (Month-End):
  Dr. Foreign Exchange Loss (P&L)              75,000
      Cr. Accounts Payable - Revaluation               75,000

Description: Month-end forex revaluation, USD payables

When Payment Made (2025-05-10):
  Actual Rate: 129.20 KES/USD
  Payment: 50,000 × 129.20 = 6,460,000 KES
  
  Reverse unrealized loss:
    Dr. Accounts Payable - Revaluation         75,000
        Cr. Foreign Exchange Loss (reversal)           75,000
  
  Book realized loss:
    Dr. Accounts Payable                    6,400,000
    Dr. Foreign Exchange Loss (realized)       60,000
        Cr. Bank Account                               6,460,000

Net Effect: 60,000 KES loss (actual rate vs original)

Realized Gains/Losses Summary:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Period: Q1 2025

┌──────────┬────────────┬───────────┬───────────┬────────────┐
│ PO No    │ Currency   │ PO Rate   │ Pay Rate  │ Gain/(Loss)│
├──────────┼────────────┼───────────┼───────────┼────────────┤
│ PO-00285 │ USD 20,000 │  128.00   │  127.50   │  +10,000 ✓ │
│ PO-00298 │ EUR 15,000 │  140.50   │  141.20   │  -10,500   │
│ PO-00312 │ USD 35,000 │  129.00   │  128.80   │   +7,000 ✓ │
│ PO-00325 │ EUR 84,000 │  140.85   │  140.68   │  +14,280 ✓ │
│ PO-00340 │ GBP  8,000 │  164.00   │  165.50   │  -12,000   │
│          │            │           │           │            │
│ TOTAL    │            │           │           │   +8,780 ✓ │
└──────────┴────────────┴───────────┴───────────┴────────────┘

Net Forex Gain: 8,780 KES (minimal impact, good hedging)

P&L Impact:
  Foreign Exchange Gains: 31,280 KES
  Foreign Exchange Losses: (22,500 KES)
  Net Forex Income: 8,780 KES
```

### Multi-Currency Reporting

```markdown
FOREIGN CURRENCY REPORTS

Open Payables by Currency:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
As at: 2025-04-30

┌──────────┬───────────────┬──────────┬────────────┬────────────┐
│ Currency │ Payable (FC)  │ Rate     │ KES Value  │ Due Date   │
├──────────┼───────────────┼──────────┼────────────┼────────────┤
│ USD      │    85,000     │ 129.00   │ 10,965,000 │ Various    │
│ EUR      │    45,000     │ 140.85   │  6,338,250 │ Various    │
│ GBP      │    12,000     │ 164.50   │  1,974,000 │ Various    │
│ CNY      │   120,000     │  17.95   │  2,154,000 │ Various    │
│          │               │          │            │            │
│ TOTAL FC PAYABLES (KES)  │          │ 21,431,250 │            │
│ KES PAYABLES             │          │ 58,750,000 │            │
│                          │          │            │            │
│ TOTAL ACCOUNTS PAYABLE   │          │ 80,181,250 │            │
└──────────┴───────────────┴──────────┴────────────┴────────────┘

Forex Exposure Analysis:
  Total FC Payables: 21.4M KES (27% of AP)
  Unhedged Exposure: 12.5M KES (58% of FC payables)
  Hedged (Forward Contracts): 8.9M KES (42%)
  
Recommendation: Increase hedging ratio to 70%

Foreign Currency Purchases Report:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Year: 2025 (Jan-Apr)

┌──────────┬───────────────┬──────────┬────────────┬─────────┐
│ Currency │ No. of POs    │ FC Value │ KES Value  │ % Total │
├──────────┼───────────────┼──────────┼────────────┼─────────┤
│ KES      │     385       │    -     │ 125,500,000│  82.5%  │
│ USD      │      18       │  180,000 │  23,220,000│  15.3%  │
│ EUR      │       6       │  120,000 │  16,902,000│  11.1%  │
│ GBP      │       2       │   15,000 │   2,467,500│   1.6%  │
│ CNY      │       4       │  250,000 │   4,487,500│   3.0%  │
│          │               │          │            │         │
│ TOTAL    │     415       │          │ 152,200,000│ 100.0%  │
└──────────┴───────────────┴──────────┴────────────┴─────────┘

Insights:
  • 82.5% local currency purchases (KES)
  • 17.5% foreign currency (mainly USD, EUR)
  • FC purchases: Capital equipment, specialized materials

Forex Gain/Loss Trend:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Quarter      Gain/(Loss)     Impact on Profit
─────────────────────────────────────────────────────────
Q1 2024      +45,000 KES     +0.03%
Q2 2024      -82,000 KES     -0.05%
Q3 2024      +12,000 KES     +0.01%
Q4 2024      -35,000 KES     -0.02%
Q1 2025      +8,780 KES      +0.01%

Trend: Well managed, minimal P&L impact ✓
Strategy: Hedging working effectively
```

---

**Next:** [Module Integration Points](./20-module-integration-points.md)
