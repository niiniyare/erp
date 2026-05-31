[<-- Back to Index](README.md)

## Pricing & Discount Management

### Supplier Pricing Structure

```markdown
PRICE LIST MANAGEMENT

Supplier Price List:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Supplier: Ace Steel Suppliers Ltd
Effective Date: 2025-01-01
Valid Until: 2025-12-31
Currency: KES

┌──────────────┬─────────────────────┬──────────┬────────────┐
│ Item Code    │ Description         │ UOM      │ Price (KES)│
├──────────────┼─────────────────────┼──────────┼────────────┤
│ RM-STL-001   │ Cold Rolled Steel   │ Ton      │   85,000   │
│ RM-STL-002   │ Hot Rolled Steel    │ Ton      │   75,000   │
│ RM-STL-003   │ Stainless Steel 304 │ Ton      │  250,000   │
│ RM-STL-004   │ Galvanized Sheet    │ Ton      │   95,000   │
└──────────────┴─────────────────────┴──────────┴────────────┘

Volume Discounts:
  10-20 Tons: 2% discount
  21-50 Tons: 5% discount
  51+ Tons: 8% discount

Payment Terms Discount:
  Payment within 10 days: 2% discount
  Standard: Net 30 days

Price Indexation:
  Base: London Metal Exchange (LME) Steel Index
  Review: Quarterly
  Adjustment: ±10% triggers renegotiation

Contract Pricing vs. Spot Pricing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Item: RM-STL-001 (Cold Rolled Steel)

Contract Price (Annual Agreement):
  Price: 85,000 KES/Ton (locked)
  Volume Commitment: Min 100 Tons/year
  Validity: 12 months
  Advantage: Price stability

Spot Price (Per Order):
  Price: Market rate (varies)
  Current: 87,000 KES/Ton
  No Commitment: Order as needed
  Risk: Price fluctuations

Example Scenario:
  Requirement: 120 Tons/year
  
  Option A - Contract:
    Price: 120 × 85,000 = 10,200,000 KES
    Volume Discount (5%): -510,000 KES
    Net Cost: 9,690,000 KES
    
  Option B - Spot (average 87,000):
    Price: 120 × 87,000 = 10,440,000 KES
    No volume discount
    Net Cost: 10,440,000 KES
    
  Savings with Contract: 750,000 KES (7.2%)
```

### Discount Types & Management

```markdown
DISCOUNT STRUCTURES

Volume/Quantity Discounts:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Supplier: Kenya Fasteners Ltd

┌──────────────────┬───────────┬────────────┬────────────┐
│ Order Value (KES)│ Discount  │ Example    │ Net Price  │
├──────────────────┼───────────┼────────────┼────────────┤
│ 0 - 50,000       │ 0%        │ 40,000     │  40,000    │
│ 50,001 - 100,000 │ 5%        │ 80,000     │  76,000    │
│ 100,001 - 200,000│ 8%        │ 150,000    │ 138,000    │
│ 200,001+         │ 10%       │ 300,000    │ 270,000    │
└──────────────────┴───────────┴────────────┴────────────┘

Application:
  Automatic at PO creation based on order value
  Discount shown as line item on PO

Early Payment Discounts:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Terms: 2/10 Net 30

Calculation:
  Invoice Amount: 1,000,000 KES
  Invoice Date: Apr 1
  Due Date: May 1 (30 days)
  
  Discount Period: Apr 1 - Apr 11 (10 days)
  Discount Rate: 2%
  
  If Paid by Apr 11:
    Discount: 20,000 KES
    Payment: 980,000 KES
    Annualized ROI: 36.5%
    
  If Paid After Apr 11:
    No Discount
    Payment: 1,000,000 KES

System Alert:
  "Early payment discount available: 20,000 KES
   Pay by Apr 11 to save 2%"

Seasonal/Promotional Discounts:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Promotion: Q2 2025 Office Equipment Sale
Supplier: Office Supplies Co
Period: Apr 1 - Jun 30, 2025

Discount: 15% on all office furniture
Conditions:
  Minimum order: 100,000 KES
  Categories: Desks, chairs, cabinets
  Cannot combine with other discounts

Example:
  Order: 10 office chairs @ 15,000 = 150,000 KES
  Promotional Discount (15%): -22,500 KES
  Net Price: 127,500 KES

Loyalty/Long-term Discounts:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Supplier: ChemSupply Ltd
Agreement: Strategic supplier status

Benefits:
  Base Discount: 3% on all purchases
  Applies: On top of other discounts
  Duration: Annual (renewable)
  Minimum Spend: 2M KES/year

Example with stacking:
  Item Price: 100,000 KES
  Volume Discount (5%): -5,000 KES = 95,000 KES
  Loyalty Discount (3%): -2,850 KES = 92,150 KES
  Total Savings: 7,850 KES (7.85%)

Blanket Order Discounts:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Arrangement: Annual blanket PO with call-offs
Supplier: MRO Supplies Ltd
Commitment: 1,000,000 KES minimum

Discount: 12% on all items
Call-off: As needed, 3-day delivery
Pricing: Locked for 12 months
Review: Annual

Advantage:
  Locked pricing + volume discount
  Flexibility in timing
  Reduced procurement admin
```

### Price Variance Management

```markdown
PRICE CHANGE HANDLING

Supplier Price Increase Notification:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
From: Ace Steel Suppliers Ltd
Date: 2025-06-01
Subject: Price Adjustment Notification

"Dear AWO Manufacturing,

Due to global steel price increases (LME index up 15%),
we must adjust our pricing effective August 1, 2025:

Cold Rolled Steel: 85,000 → 92,000 KES/Ton (+8.2%)

Your current contract (valid until Dec 31) allows for
quarterly price reviews with index-based adjustments.

We value our partnership and remain committed to
competitive pricing.

Regards,
Ace Steel"

Procurement Action:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1: Verify Market Conditions
  - Check LME steel index: ✓ Confirmed +15%
  - Review contract terms: ✓ Indexation clause valid
  - Check other suppliers: Similar increases reported

Step 2: Negotiate
  Meeting: 2025-06-10
  Discussion:
    - Market conditions acknowledged
    - Request gradual implementation
    - Request enhanced volume discount
    
  Outcome:
    - Price: 90,000 KES/Ton (vs 92,000 proposed)
    - Implementation: Staggered (Aug 1)
    - Volume Discount: Improved to 6% (from 5%)
    
Step 3: Update System
  - New price list: Effective Aug 1, 2025
  - PO template: Updated pricing
  - Budget impact: Calculated and reported

Step 4: Stock Strategy
  - Place large order in July at old price
  - Stock 2-3 months requirement
  - Optimize cash vs. price savings

Price Variance Analysis:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Report: Q2 2025

┌─────────────────────┬───────────┬───────────┬───────────┬────────┐
│ Item                │ Std Price │ Avg Paid  │ Variance  │ Impact │
├─────────────────────┼───────────┼───────────┼───────────┼────────┤
│ RM-STL-001          │  85,000   │  84,000   │ -1,000 ✓ │ Saving │
│ MRO-LUB-015         │   1,500   │   1,550   │ +50 ⚠    │ Loss   │
│ Office Supplies     │  10,000   │   9,200   │ -800 ✓   │ Saving │
└─────────────────────┴───────────┴───────────┴───────────┴────────┘

Positive Variances (Savings):
  Volume discounts captured: 150,000 KES
  Early payment discounts: 80,000 KES
  Competitive bidding: 120,000 KES
  Total Savings: 350,000 KES

Negative Variances (Unfavorable):
  Market price increases: 45,000 KES
  Emergency purchases: 25,000 KES
  Total Losses: 70,000 KES

Net Procurement Savings: 280,000 KES (2.8% of spend)
```

### Procurement Cost Optimization

```markdown
COST REDUCTION STRATEGIES

Supplier Consolidation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Before:
  MRO Supplies from 15 suppliers
  Average order: 25,000 KES
  No volume leverage
  
After Consolidation:
  MRO Supplies from 3 strategic suppliers
  Average order: 125,000 KES
  Volume discount: 10%
  
  Annual Spend: 3,000,000 KES
  Savings: 300,000 KES (10%)
  
Additional Benefits:
  - Simplified AP processing
  - Better service levels
  - Reduced admin cost

Competitive Bidding:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Practice: Quarterly RFQ for major items

Example:
  Item: Industrial Lubricants
  Annual Spend: 2,000,000 KES
  
  Previous Supplier: Single source at 1,500 KES/L
  RFQ Process: 4 suppliers invited
  
  Results:
    - Supplier A: 1,450 KES/L (-3.3%)
    - Supplier B: 1,400 KES/L (-6.7%) ← Selected
    - Supplier C: 1,500 KES/L (same)
    - Supplier D: 1,550 KES/L (+3.3%)
    
  Annual Savings: 133,000 KES (6.7%)

Payment Term Optimization:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategy: Capture early payment discounts selectively

Analysis:
  Available Discounts: 500,000 KES/month (2/10 Net 30)
  Early Payment Required: 20 days early average
  
  Cost of Funds: 12% p.a. (bank rate)
  Cost to pay early: 500K × 12% × (20/365) = 3,288 KES
  
  Discount Earned: 10,000 KES
  Net Benefit: 6,712 KES
  
  ROI: 204% (highly profitable)

Decision: Always take early payment discounts
Annual Savings: ~120,000 KES

Total Cost of Ownership (TCO):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Item: Industrial Machine (Capital Equipment)

Supplier A - Lowest Initial Price:
  Purchase Price: 8,000,000 KES
  Installation: 300,000 KES
  Training: 100,000 KES
  Annual Maintenance: 400,000 KES
  Energy Cost: 200,000 KES/year
  Expected Life: 10 years
  
  TCO (10 years):
    Initial: 8,400,000 KES
    Recurring: 6,000,000 KES (10 × 600K)
    Total: 14,400,000 KES

Supplier B - Higher Initial, Lower Operating:
  Purchase Price: 9,000,000 KES
  Installation: 200,000 KES (included)
  Training: Free
  Annual Maintenance: 200,000 KES (warranty)
  Energy Cost: 150,000 KES/year (efficient)
  Expected Life: 10 years
  
  TCO (10 years):
    Initial: 9,000,000 KES
    Recurring: 3,500,000 KES (10 × 350K)
    Total: 12,500,000 KES

Decision: Supplier B
Savings: 1,900,000 KES over lifetime (13%)
```

---

**Next:** [Procurement Analytics & Reporting](./16-procurement-analytics-and-reporting.md)
