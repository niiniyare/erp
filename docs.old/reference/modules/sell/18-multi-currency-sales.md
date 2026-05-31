[<-- Back to Index](README.md)

## 18. Multi-Currency Sales

### Foreign Currency Operations

```markdown
MULTI-CURRENCY SALES MANAGEMENT

Currency Setup:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Base Currency: KES (Kenya Shillings)
  - All financial statements in KES
  - Default for local customers

Foreign Currencies:
  - USD (US Dollar)
  - EUR (Euro)
  - GBP (British Pound)
  - CNY (Chinese Yuan)
  - UGX (Uganda Shilling)
  - TZS (Tanzania Shilling)

Exchange Rate Sources:
  1. Central Bank of Kenya (official)
  2. Commercial banks
  3. Forex providers (API)
  4. Manual entry

Example Rate Table:
┌──────────┬─────────────┬──────────┬──────────┐
│ Currency │ Buy Rate    │ Sell Rate│ Date     │
├──────────┼─────────────┼──────────┼──────────┤
│ USD      │ 129.50 KES  │ 130.50   │ Jan 20   │
│ EUR      │ 142.30 KES  │ 143.50   │ Jan 20   │
│ GBP      │ 165.80 KES  │ 167.20   │ Jan 20   │
└──────────┴─────────────┴──────────┴──────────┘

Multi-Currency Price Lists:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Product: Machine Model A

KES Price List (Local Sales):
  Rate: 2,000,000 KES

USD Price List (Export/Expat):
  Rate: 15,000 USD
  
  Conversion Check:
    15,000 USD × 130 = 1,950,000 KES
    (Slightly lower for competitive export pricing)

EUR Price List (European Customers):
  Rate: 14,000 EUR
  
  Conversion Check:
    14,000 EUR × 143 = 2,002,000 KES

Automatic Currency Selection:
  Customer Country → Default Currency
  Kenya → KES
  USA → USD
  Uganda → UGX or USD
  Europe → EUR or USD

Foreign Currency Sales Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Export Sale Example:

Quotation:
  Customer: Global Import Co. (USA)
  Currency: USD
  Amount: 50,000 USD
  Exchange Rate (Quote Date): 130.00 KES/USD
  Equivalent: 6,500,000 KES

Sales Order:
  Date: Jan 20, 2025
  Amount: 50,000 USD
  Rate: 130.00 KES/USD (locked)
  
  System Records:
    - Transaction Currency: 50,000 USD
    - Base Currency: 6,500,000 KES
    - Exchange Rate: 130.00
    - Rate Date: Jan 20, 2025

Invoice:
  Date: Feb 10, 2025
  Amount: 50,000 USD
  Rate: 130.00 (from sales order)
  Equivalent: 6,500,000 KES
  
  Accounting Entry:
    Dr. AR - Global Import (USD)     6,500,000 KES
        Cr. Sales Revenue                    6,500,000 KES
  
  System tracks both:
    USD: 50,000 (original currency)
    KES: 6,500,000 (functional currency)

Payment Receipt:
  Date: Mar 5, 2025
  Amount: 50,000 USD received
  Exchange Rate (Payment Date): 132.00 KES/USD
  KES Received: 6,600,000 KES
  
  Accounting Entry:
    Dr. Bank (USD)                   6,600,000 KES
        Cr. AR - Global Import               6,500,000 KES
        Cr. Foreign Exchange Gain               100,000 KES

Foreign Exchange Gain/Loss:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Realized Gain (as above):
  Invoice Rate: 130.00
  Payment Rate: 132.00
  Difference: +2.00 KES per USD
  Gain: 50,000 × 2.00 = 100,000 KES
  
  Income Statement Impact:
    Cr. FX Gain (Other Income): 100,000 KES

Realized Loss Example:
  Invoice Rate: 130.00
  Payment Rate: 128.00
  Difference: -2.00 KES per USD
  Loss: 50,000 × 2.00 = 100,000 KES
  
  Income Statement Impact:
    Dr. FX Loss (Other Expense): 100,000 KES

Unrealized Gain/Loss:
  Month-end revaluation of outstanding AR/AP
  
  Outstanding Invoice: 50,000 USD
  Invoice Rate: 130.00 → 6,500,000 KES
  Month-end Rate: 133.00 → 6,650,000 KES
  Unrealized Gain: 150,000 KES
  
  Entry:
    Dr. AR - Global Import (revaluation)  150,000
        Cr. Unrealized FX Gain                   150,000
  
  Note: Reverses next month or on payment

Currency Risk Management:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy 1: Forward Contracts
  Lock in exchange rate for future payment
  
  Export Invoice: 50,000 USD due in 60 days
  Current Rate: 130.00
  Forward Rate: 130.50 (slight premium)
  
  Action: Buy forward contract
  Benefit: Certainty on KES amount

Strategy 2: Natural Hedging
  Match USD receivables with USD payables
  
  Example:
    USD Receivables: 100,000 USD
    USD Payables (imports): 80,000 USD
    Net Exposure: 20,000 USD only

Strategy 3: Pricing Buffer
  Build in FX buffer in export prices
  
  Cost in KES: 6,000,000
  Target Margin: 20%
  Target KES: 7,200,000
  
  Expected Rate: 130 → 55,385 USD
  Buffer (5%): +2,769 USD
  Quote: 58,000 USD
  
  If rate drops to 125:
    58,000 × 125 = 7,250,000 KES (still profitable)

Strategy 4: Payment Terms
  Reduce exposure period
  
  Standard: Net 60 days
  Export: Payment in advance or LC
  
  Or: Price incentive for early payment
    60 days: 50,000 USD
    Advance: 48,500 USD (3% discount)
```

### Export Sales Specifics

```markdown
EXPORT SALES DOCUMENTATION

Required Documents:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Commercial Invoice
   Standard sales invoice
   Additional fields:
     - Incoterms (FOB/CIF/EXW)
     - Country of origin
     - HS Code (customs classification)
     - Export declaration number

2. Proforma Invoice
   Preliminary invoice for:
     - Customer to arrange payment/LC
     - Customs valuation
     - Import permit application

3. Packing List
   Detailed contents of each package:
     - Box numbers
     - Item descriptions
     - Quantities
     - Weights (gross and net)
     - Dimensions

4. Certificate of Origin
   Certifies goods manufactured in Kenya
   Required for preferential tariffs
   Issued by Kenya Chamber of Commerce

5. Bill of Lading (B/L)
   Shipping document from freight forwarder
   Evidence of shipment
   Required for customs clearance

6. Export Declaration
   Filed with Kenya Revenue Authority
   Required for VAT zero-rating
   EDF (Electronic Declaration Form)

7. Quality Certificate (if required)
   Inspection certificate
   Conformity to standards
   May require pre-shipment inspection

8. Insurance Certificate (if CIF)
   Marine cargo insurance
   Covers goods in transit
   As per Incoterms requirement

Incoterms:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EXW (Ex Works):
  Customer responsibility: Pick up from our premises
  Our cost: Production only
  Quote: 50,000 USD EXW Nairobi

FOB (Free On Board):
  Our responsibility: Deliver to ship at port
  Includes: Inland transport + export clearance
  Quote: 52,000 USD FOB Mombasa

CIF (Cost, Insurance, Freight):
  Our responsibility: Deliver to destination port
  Includes: FOB + ocean freight + insurance
  Quote: 55,000 USD CIF New York

VAT Treatment:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Export Sales = Zero-Rated (0% VAT)
  Conditions:
    ✓ Goods exported outside Kenya/EAC
    ✓ Evidence of export (B/L, customs docs)
    ✓ Export declaration filed
    ✓ Payment received in forex (usually)

Invoice Format:
  Subtotal: 50,000 USD
  VAT @ 0%: 0 USD (Zero-rated export)
  Total: 50,000 USD

Regional Sales (EAC):
  Sales to Uganda, Tanzania, Rwanda, etc.
  Still zero-rated
  Require: C2 Form (EAC goods movement cert)

Letter of Credit (LC):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Process:
  1. Proforma invoice sent to customer
  2. Customer's bank opens LC
  3. LC sent to our bank (advising bank)
  4. We ship goods
  5. We present documents to bank
  6. Bank verifies documents vs LC terms
  7. Bank pays us (3-5 days)
  8. Customer's bank debits customer

Document Requirements (typical):
  □ Commercial Invoice (3 copies)
  □ Packing List (3 copies)
  □ Bill of Lading (full set, original)
  □ Certificate of Origin (original)
  □ Insurance Certificate (if CIF)
  □ Inspection Certificate (if required)
  □ Any other docs per LC terms

Critical: Documents must match LC terms exactly
  - Customer name spelling
  - Description of goods
  - Quantities
  - Values
  - Shipment dates
  
Any discrepancy = Bank may reject payment
```

---