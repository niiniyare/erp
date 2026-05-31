[<-- Back to Index](README.md)

## 16. Commission Management

### Commission Structure

```markdown
SALES COMMISSION FRAMEWORK

Commission Models:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model 1: Flat Percentage
  Simple: X% of revenue
  
  Example:
    Rate: 5% of sales
    Sales: 10M KES
    Commission: 500K KES

Model 2: Tiered (Progressive)
  Rate increases with achievement
  
  Example:
    0-80% of quota: 3%
    80-100% of quota: 5%
    100-120% of quota: 7%
    120%+ of quota: 10%
  
  Quota: 10M KES
  Actual: 12M KES (120%)
  
  Calculation:
    First 8M (80%): 8M × 3% = 240K
    Next 2M (100%): 2M × 5% = 100K
    Last 2M (120%): 2M × 7% = 140K
    Total: 480K KES

Model 3: Profit-Based
  Commission on margin, not revenue
  Encourages profitable sales
  
  Example:
    Rate: 15% of gross profit
    Sale: 10M KES
    Cost: 7M KES
    Profit: 3M KES
    Commission: 3M × 15% = 450K KES

Model 4: Hybrid (Revenue + Profit)
  Balanced approach
  
  Example:
    Base: 2% of revenue = 200K
    Bonus: 10% of margin = 300K
    Total: 500K KES

Model 5: Team + Individual
  Split between personal and team performance
  
  Example:
    Individual Component (60%):
      Personal sales: 12M
      Rate: 5%
      Amount: 600K × 60% = 360K
    
    Team Component (40%):
      Team sales: 50M
      Rate: 3%
      Share: (12M/50M) of 1.5M
      Amount: 360K × 40% = 144K
    
    Total: 504K KES

Commission Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Eligibility:
  ✓ Invoice submitted (not draft)
  ✓ Payment received (configurable: invoice or payment)
  ✓ No return/cancellation
  ✓ Within employment period

Calculation Basis:
  Option A: On Invoice Submission
    Commission earned when invoice submitted
    Risk: Customer may not pay
  
  Option B: On Payment Receipt (Recommended)
    Commission earned when payment received
    Fair: Rep gets paid when company gets paid

Split Scenarios:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scenario 1: Mid-Period Rep Change
  Original Rep left midway through deal
  New rep closed the deal
  
  Solution:
    Original Rep: 50% (for initial work)
    New Rep: 50% (for closing)

Scenario 2: Team Selling
  Multiple people involved
  
  Solution:
    Account Executive: 60%
    Sales Engineer: 25%
    Sales Manager: 15%

Scenario 3: Referral
  Existing rep refers customer in another territory
  
  Solution:
    Territory Rep (closes): 80%
    Referring Rep: 20%

Commission Caps & Floors:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Caps (Maximum):
  Why: Prevent windfall from single large deal
  Example: Maximum 2M KES per quarter
  
  If exceeded:
    Option A: Hard cap at 2M
    Option B: Reduced rate above cap
    Option C: Defer excess to next period

Floors (Minimum):
  Guaranteed minimum (Draw)
  Even if no sales, rep gets base
  
  Example:
    Draw: 50K KES/month
    Commission: 45K earned
    Payment: 50K (company covers shortfall)
  
  Recovery:
    Future commissions offset past draws
    Or: Non-recoverable (pure guarantee)

Special Situations:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Product Launch Bonus:
  Extra incentive for new products
  New Product: 8% (vs standard 5%)
  Duration: First 6 months

Strategic Account Bonus:
  Extra for winning key accounts
  Target Account List: 20 companies
  Bonus: Additional 2% if won

Retention Bonus:
  Commission on renewals
  New Sale: 5%
  Renewal: 2%

Deal Size Accelerators:
  Larger deals = higher rate
  <1M: 4%
  1-5M: 5%
  5-10M: 6%
  10M+: 7%
```

### Commission Calculation Process

```markdown
COMMISSION PROCESSING

Monthly Commission Cycle:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Timeline:
  Day 1-31: Sales activity (Jan 2025)
  Day 32-35: Finance validates invoices & payments
  Day 36-38: Commission calculation runs
  Day 39-40: Manager reviews & approves
  Day 41-42: Payroll processing
  Day 43: Commission paid (Feb 12)

Calculation Steps:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Step 1: Gather Eligible Transactions
  Query:
    - Sales Invoices submitted in Jan 2025
    - With payments received (if policy = on payment)
    - Not cancelled or returned
    - Assigned to sales person

Step 2: Calculate Commission per Transaction
  For each invoice:
    Invoice: INV-2025-001
    Amount: 4,172,056 KES (net of discounts)
    Sales Person: Sarah Johnson
    Commission Rate: 5%
    Commission: 208,603 KES

Step 3: Apply Rules
  - Check if team-selling (split commission)
  - Apply tiered rates if applicable
  - Check for special bonuses
  - Apply any caps or clawbacks

Step 4: Aggregate
  Sarah Johnson - January 2025:
  ┌────────────┬───────────┬──────┬───────────┐
  │ Invoice    │ Amount    │ Rate │ Commission│
  ├────────────┼───────────┼──────┼───────────┤
  │ INV-2025-1 │ 4,172,056 │  5%  │  208,603  │
  │ INV-2025-5 │ 2,180,000 │  5%  │  109,000  │
  │ INV-2025-8 │ 3,650,000 │  5%  │  182,500  │
  │ ...        │ ...       │ ...  │  ...      │
  │            │           │      │           │
  │ TOTAL      │12,500,000 │      │  625,000  │
  └────────────┴───────────┴──────┴───────────┘

Step 5: Adjustments
  Gross Commission: 625,000 KES
  Less: Previous advance (draw): 0
  Less: Clawbacks (returns): -25,000
  Add: Bonuses (new product): +50,000
  ───────────────────────────────
  Net Commission: 650,000 KES

Step 6: Approval
  Manager reviews for accuracy
  Checks for disputes
  Approves payment

Step 7: Payment
  Added to payroll
  Paid with monthly salary
  Statement sent to sales person

Commission Statement:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ COMMISSION STATEMENT                       │
│ Period: January 2025                       │
│ Sales Person: Sarah Johnson                │
│                                            │
│ SALES SUMMARY                              │
│ Total Sales: 12,500,000 KES                │
│ Quota: 10,000,000 KES                      │
│ Achievement: 125% ⭐                        │
│                                            │
│ COMMISSION CALCULATION                     │
│ Base Commission (5%):      625,000         │
│ Tier Bonus (>120%):         62,500         │
│ New Product Bonus:          50,000         │
│ ──────────────────────────────────         │
│ Gross Commission:          737,500         │
│                                            │
│ ADJUSTMENTS                                │
│ Returns (INV-2024-089):    -25,000         │
│ Q4 Recovery:                    0          │
│ ──────────────────────────────────         │
│ Net Commission:            712,500         │
│                                            │
│ PAYMENT                                    │
│ Pay Date: February 12, 2025                │
│ Method: Bank Transfer                      │
│                                            │
│ YEAR-TO-DATE                               │
│ YTD Sales: 12,500,000                      │
│ YTD Commission: 712,500                    │
│ Avg Rate: 5.7%                             │
└────────────────────────────────────────────┘

Dispute Resolution:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Person Reviews Statement:
  - Checks all transactions included
  - Verifies rates applied
  - Confirms adjustments

If Dispute:
  1. Sales person submits dispute form
  2. Manager & Finance review
  3. Investigation (2-3 days)
  4. Resolution:
     - Adjust current payment OR
     - Correct next month
  5. Updated statement issued

Common Disputes:
  - Missing transaction
  - Wrong rate applied
  - Team-selling split incorrect
  - Unrecognized adjustment
```

### Commission Clawbacks

```markdown
COMMISSION RECOVERY SCENARIOS

When Clawback Applies:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Customer Non-Payment
   Invoice: INV-2025-001
   Amount: 5M KES
   Commission Paid: 250K KES
   Status: Customer defaulted (unpaid after 90 days)
   
   Action:
     If policy = "on payment": No clawback (not paid yet)
     If policy = "on invoice": Clawback 250K

2. Sales Return
   Invoice: INV-2025-005
   Amount: 3M KES
   Commission Paid: 150K KES
   Return: Full return approved
   
   Action: Clawback 150K from next commission

3. Partial Return
   Original: 5M KES, Commission: 250K
   Return: 2M KES worth of goods
   Clawback: 2M × 5% = 100K

4. Credit Note Issued
   Invoice: 4M KES, Commission: 200K
   Credit Note: 500K (price adjustment)
   Clawback: 500K × 5% = 25K

5. Deal Cancelled Pre-Delivery
   Order taken, commission paid
   Customer cancels before delivery
   Full clawback of commission

Clawback Processing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Method 1: Next Commission Offset
  Most common approach
  Deduct from next month's commission
  
  Example:
    Feb Commission Earned: 500K
    Jan Clawback: -100K
    Net Payment: 400K

Method 2: Salary Deduction
  If no future commissions expected
  (e.g., rep resigned)
  Deduct from final salary

Method 3: Direct Repayment
  Rep no longer employed
  Invoice rep for amount
  Legal action if not paid

Time Limits:
  Clawback Period: 6 months
  After 6 months: Company absorbs loss
  Exception: Fraud (no time limit)

Rep Protection:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Maximum Clawback: 50% of monthly commission
  Spreads impact over multiple months
  
  Example:
    Clawback Due: 300K
    Monthly Commission: 400K
    Max Deduction: 200K (50%)
    
    Recovery Schedule:
      Month 1: -200K
      Month 2: -100K (balance)

No Clawback Situations:
  - Customer bankruptcy (not rep's fault)
  - Force majeure events
  - Company-caused delays/issues
  - Returns due to product defects
```

---