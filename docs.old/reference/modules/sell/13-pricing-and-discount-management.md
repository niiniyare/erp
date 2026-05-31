[<-- Back to Index](README.md)

## 13. Pricing & Discount Management

### Price List Management

```markdown
PRICE LIST STRUCTURE

Price List Hierarchy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Standard Retail Price (Base)
├─ Wholesale Price (15% discount)
├─ Corporate Price (10% discount)
├─ Distributor Price (25% discount)
├─ Export Price (USD, Euro based)
└─ Promotional Price (temporary)

Price List Master:
┌────────────────────────────────────────────┐
│ Price List Name: Standard Retail Price     │
│ Currency: KES                              │
│ Valid From: 2025-01-01                     │
│ Valid To: 2025-12-31                       │
│ Enabled: Yes                               │
│ Default: Yes (for new customers)           │
│                                            │
│ Applicable To:                             │
│ ☑ All Customers                            │
│ □ Specific Customer Group                  │
│ □ Specific Territory                       │
│ □ Specific Customers                       │
└────────────────────────────────────────────┘

Item Price Definition:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Product: Machine Model A

Price List: Standard Retail Price
  Rate: 2,200,000 KES
  Currency: KES
  UOM: Unit

Price List: Wholesale Price
  Rate: 1,870,000 KES (15% off retail)
  Minimum Quantity: 1
  
Price List: Corporate Price
  Rate: 1,980,000 KES (10% off retail)
  Valid From: 2025-01-01
  Valid To: 2025-06-30

Price List: Volume Discount
  Quantity 1-5: 2,200,000 KES
  Quantity 6-10: 2,090,000 KES (5% off)
  Quantity 11+: 1,980,000 KES (10% off)

Customer-Specific Pricing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer: ABC Manufacturing (VIP Customer)
Special Price: 1,900,000 KES
Reason: Long-term contract pricing
Valid: 2025-01-01 to 2025-12-31
Approved By: Sales Director

Priority Hierarchy:
  1. Customer-specific price (highest priority)
  2. Customer group price list
  3. Territory price list
  4. Standard price list (lowest priority)

Example Resolution:
  Customer: ABC Manufacturing
  Product: Machine Model A
  
  Available Prices:
    - Standard Retail: 2,200,000 KES
    - Corporate Price List: 1,980,000 KES (customer group)
    - Customer-Specific: 1,900,000 KES
  
  System selects: 1,900,000 KES (highest priority)
```

### Dynamic Pricing Rules

```markdown
PRICING RULE ENGINE

Rule Types:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Volume Discount
2. Product Bundle Discount
3. Promotional Discount
4. Early Payment Discount
5. Customer Loyalty Discount
6. Seasonal Discount
7. Clearance/Close-out Pricing

Volume Discount Rule:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: Bulk Purchase Discount
Item: Machine Model A
Rule:
  ├─ Quantity 1-5 units: No discount
  ├─ Quantity 6-10 units: 5% discount
  ├─ Quantity 11-20 units: 10% discount
  └─ Quantity 21+ units: 15% discount

Priority: 10 (higher number = higher priority)
Valid From: 2025-01-01
Valid To: 2025-12-31
Applies To: All customers
Can Combine With Other Rules: Yes

Example:
  Base Price: 2,000,000 KES
  Quantity: 12 units
  
  Calculation:
    Rate: 2,000,000 KES
    Volume Discount (10%): -200,000 KES
    Net Rate: 1,800,000 KES
    Total: 1,800,000 × 12 = 21,600,000 KES

Product Bundle Rule:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: Machine + Service Bundle
Condition: If buying Machine + Installation + Training
Discount: 10% on total

Products in Bundle:
  - Machine Model A: 2,000,000 KES
  - Installation: 150,000 KES
  - Training: 100,000 KES
  
Without Bundle:
  Total: 2,250,000 KES + VAT = 2,610,000 KES

With Bundle (10% off):
  Subtotal: 2,250,000 KES
  Bundle Discount (10%): -225,000 KES
  Net: 2,025,000 KES + VAT = 2,349,000 KES
  Savings: 261,000 KES

Promotional Discount Rule:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: January 2025 Promo
Type: Promotional
Discount: 15% on all products
Valid: 2025-01-15 to 2025-01-31
Applies To: Corporate Customers only
Max Discount: 500,000 KES per order

Promo Code: JAN2025
Customer must enter code at checkout

Mixed Discount Scenario:
  Base Price: 2,000,000 KES
  Customer Group Discount (10%): -200,000 = 1,800,000
  Promotional Discount (15%): -270,000
  
System calculates:
  Option A: Sequential (10% then 15%)
    After 10%: 1,800,000
    After 15%: 1,530,000 KES
  
  Option B: Best discount only (15%)
    2,000,000 - 15% = 1,700,000 KES
  
  Option C: Additive (25% total)
    2,000,000 - 25% = 1,500,000 KES

Configuration setting determines which method to use.

Customer Loyalty Discount:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: Platinum Customer Discount
Condition: 
  Customer Tier = Platinum
  (Based on annual purchase > 10M KES)

Discount: 5% on all orders
Can Combine: Yes (with volume discounts)
Cannot Combine: Promotional discounts

Example:
  Customer: ABC Manufacturing (Platinum)
  Item: Machine Model A × 8 units
  Base Price: 2,000,000 KES
  
  Calculation:
    Base: 2,000,000 × 8 = 16,000,000
    Volume Discount (5% for 6-10 units): -800,000
    Subtotal: 15,200,000
    Loyalty Discount (5%): -760,000
    Net: 14,440,000 KES
    Total Savings: 1,560,000 KES (9.75%)

Seasonal Pricing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rule Name: Q4 Year-End Sale
Products: Selected items only
Discount: 20%
Valid: 2025-10-01 to 2025-12-31
Reason: Clear old stock before new models

Items on Sale:
  - Machine Model A (2024): 20% off
  - Machine Model B (2024): 25% off
  
Regular price maintained for 2025 models
```

### Discount Approval Matrix

```markdown
DISCOUNT AUTHORIZATION LEVELS

Approval Hierarchy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Discount Level → Required Approval
┌──────────────┬────────────────────────────┐
│ 0-10%        │ Sales Person (no approval) │
│ 10-15%       │ Sales Manager              │
│ 15-20%       │ Sales Director             │
│ 20-25%       │ CFO                        │
│ 25%+         │ CEO                        │
└──────────────┴────────────────────────────┘

Order Value Consideration:
  Orders < 1M KES: Standard approval matrix
  Orders 1M-5M KES: +1 level approval
  Orders > 5M KES: +2 levels approval

Example Scenarios:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scenario 1:
  Order Value: 800,000 KES
  Discount: 12%
  Required Approval: Sales Manager ✓

Scenario 2:
  Order Value: 3,500,000 KES
  Discount: 12%
  Required Approval: Sales Director
  (Order > 1M, so +1 level)

Scenario 3:
  Order Value: 8,000,000 KES
  Discount: 18%
  Required Approval: CFO
  (Order > 5M, normally Sales Director, but +2 levels)

Approval Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales person creates quotation with 12% discount:
  
1. System detects discount > 10%
2. Status: PENDING APPROVAL
3. Notification sent to Sales Manager
4. Sales Manager reviews:
   - Customer history
   - Margin impact
   - Competitive situation
   - Order value
5. Decision:
   ☑ Approve (quotation can proceed)
   □ Reject (sales person notified)
   □ Approve with conditions (e.g., min quantity)

Approval Comments:
  "Approved. Customer is high-volume account
   with good payment history. Margin still
   acceptable at 12% discount."

Audit Trail:
  Requested by: Sarah Johnson (Sales)
  Requested on: 2025-01-20 10:30
  Discount: 12% (439,440 KES)
  Approved by: James Ndungu (Sales Manager)
  Approved on: 2025-01-20 14:15
  Comments: [As above]

Special Approval Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
New Customer Rule:
  Discount > 5% for first order
  Requires: Sales Director approval
  Reason: Establish proper pricing expectations

Loss Leader Rule:
  Discount resulting in negative margin
  Requires: CFO + Sales Director approval
  Justification: Strategic reasons only

Competitor Match:
  Discount > standard but matching competitor
  Attach: Competitor quote
  Approval: Sales Director
  Valid: One-time match only

Blanket Approval:
  Annual contracts with pre-approved discount
  Set: 15% for ABC Manufacturing (all orders)
  No per-order approval needed
  Review: Quarterly
```

### Margin Management

```markdown
GROSS MARGIN TRACKING

Margin Calculation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Selling Price: 2,000,000 KES
Cost (COGS): 1,400,000 KES

Gross Profit: 600,000 KES
Gross Margin %: 30%

Formula:
  Margin % = (Selling Price - Cost) / Selling Price × 100

With Discount:
  Original Price: 2,000,000 KES
  Discount (10%): -200,000 KES
  Net Price: 1,800,000 KES
  Cost: 1,400,000 KES
  
  Gross Profit: 400,000 KES
  Gross Margin %: 22.2%

Margin Alerts:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Company Policy:
  Minimum Margin: 20%
  Target Margin: 35%
  
Alert Levels:
   Margin > 35%: Excellent
   Margin 25-35%: Good
   Margin 20-25%: Acceptable (warning)
   Margin < 20%: Requires approval

Real-time Calculation:
  Sales person enters discount
  System immediately shows:
    - New margin %
    - Margin impact in KES
    - Alert if below threshold
    - Approval required if < 20%

Quotation Screen:
┌────────────────────────────────────────────┐
│ Item: Machine Model A                      │
│ Quantity: 1                                │
│                                            │
│ Base Price: 2,000,000 KES                  │
│ Cost: 1,400,000 KES                        │
│                                            │
│ Customer Discount: 10%                     │
│ Additional Discount: 5%  [WARNING! ]     │
│                                            │
│ Net Price: 1,710,000 KES                   │
│                                            │
│ Gross Profit: 310,000 KES                  │
│ Gross Margin: 18.1% [BELOW MINIMUM!]      │
│                                            │
│ ⚠ This margin requires Sales Director      │
│   approval before submission.              │
│                                            │
│ Justification: _____________________      │
└────────────────────────────────────────────┘

Margin Reporting:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Performance by Margin:
┌──────────────┬───────────┬──────────┬────────┐
│ Sales Person │ Sales     │ Margin   │ Avg %  │
├──────────────┼───────────┼──────────┼────────┤
│ Sarah        │ 45.2M     │ 15.8M    │  35%   │
│ Mike         │ 38.5M     │ 11.6M    │  30%   │
│ Jane         │ 52.1M     │ 13.0M    │  25%   │
│ Tom          │ 29.8M     │  5.4M    │  18% ⚠ │
└──────────────┴───────────┴──────────┴────────┘

⚠ Tom consistently below target margin
   Action: Review discount practices

Product Profitability:
┌──────────────┬─────────┬────────┬─────────┐
│ Product      │ Volume  │ Margin │ Avg %   │
├──────────────┼─────────┼────────┼─────────┤
│ Machine A    │ 120     │ 45M    │  35%    │
│ Machine B    │  85     │ 28M    │  32%    │
│ Machine C    │  50     │  8M    │  18% ⚠  │
│ Installation │ 180     │ 12M    │  40%    │
└──────────────┴─────────┴────────┴─────────┘

⚠ Machine C low margin - review pricing or cost
```

---

