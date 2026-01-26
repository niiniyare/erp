[<-- Back to Index](README.md)

## Procurement Process Overview

### End-to-End Purchase-to-Pay Cycle

```markdown
COMPLETE PROCUREMENT WORKFLOW

┌─────────────────────────────────────────────────────────┐
│                  PROCUREMENT LIFECYCLE                  │
└─────────────────────────────────────────────────────────┘

1. NEED IDENTIFICATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Triggers:
  • Inventory reorder point reached
  • Material requirement from production
  • Department request
  • Project requirement
  • Asset replacement need

Output: Purchase Requisition (PR)
Responsible: Department Head / Inventory System

2. REQUISITION APPROVAL
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Validations:
  • Budget availability check
  • Policy compliance
  • Approver hierarchy
  • Required vs. nice-to-have

Statuses: Draft → Pending → Approved → Rejected
Responsible: Manager → Director → Procurement

3. SUPPLIER SELECTION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Options:

A. Direct Purchase (Existing Supplier):
   - Use pre-approved supplier
   - Existing contract pricing
   - Fast-track PO creation

B. Request for Quotation (RFQ):
   - Create RFQ document
   - Send to multiple suppliers
   - Receive and compare quotations
   - Select best offer

C. Competitive Bidding:
   - Formal tender process
   - Supplier presentations
   - Evaluation criteria
   - Contract negotiation

Output: Selected Supplier + Price
Responsible: Procurement Team

4. PURCHASE ORDER CREATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PO Contents:
  • Line items and quantities
  • Agreed pricing
  • Payment terms
  • Delivery schedule
  • Delivery location
  • Quality specifications

Statuses: Draft → Submitted → Sent → Acknowledged
System Actions:
  ✓ Budget commitment recorded
  ✓ Expected receipt date set
  ✓ Automatic email/EDI to supplier
  ✓ Reference number generated

Responsible: Procurement Officer

5. ORDER ACKNOWLEDGMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Supplier Actions:
  • Confirm receipt of PO
  • Validate pricing and terms
  • Confirm delivery date
  • Provide production timeline

Procurement Actions:
  • Track PO acknowledgment
  • Update expected delivery
  • Flag discrepancies
  • Revise PO if needed

6. GOODS RECEIPT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Receiving Process:
  1. Verify PO reference
  2. Physical inspection
  3. Quantity verification
  4. Quality check
  5. Acceptance or rejection
  6. System entry (GRN - Goods Receipt Note)

Outcomes:
  • Full Acceptance: Update stock, close PO
  • Partial Receipt: Update partial quantities
  • Rejection: Create return note
  • Discrepancy: Note variance, inform procurement

Stock Impact:
  Dr. Inventory (Asset)
      Cr. GRN Clearing Account

Responsible: Warehouse Team

7. INVOICE RECEIPT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Supplier Invoice Arrives:
  • Paper invoice (scanned)
  • Email PDF
  • EDI transmission
  • Supplier portal upload

System Entry:
  • Invoice number
  • Invoice date
  • Invoice amount
  • Tax details
  • Payment terms

8. THREE-WAY MATCHING
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Automatic Comparison:

┌─────────────────┬──────────────┬──────────────┬─────────┐
│ Field           │ PO           │ GRN          │ Invoice │
├─────────────────┼──────────────┼──────────────┼─────────┤
│ Quantity        │ 100 units    │ 100 units    │ 100     │
│ Price/Unit      │ 10,000 KES   │ -            │ 10,000  │
│ Total           │ 1,000,000    │ -            │1,000,000│
│ Tax             │ 160,000      │ -            │ 160,000 │
│ Grand Total     │ 1,160,000    │ -            │1,160,000│
└─────────────────┴──────────────┴──────────────┴─────────┘

Match Result: ✓ MATCHED - Approve for payment

Variance Scenarios:
  • Price mismatch → Procurement review
  • Quantity mismatch → Investigate with warehouse
  • Tax mismatch → Verify with finance
  • Within tolerance (±2%) → Auto-approve
  • Outside tolerance → Manual approval

9. INVOICE APPROVAL
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
If Matched:
  • Auto-route to AP team
  • Schedule payment per terms
  • Post to GL

If Variance:
  • Route to exception queue
  • Procurement investigation
  • Resolution or adjustment
  • Manual approval

Financial Posting:
  Dr. Inventory / Expense Account    1,000,000
  Dr. VAT Input                         160,000
      Cr. Accounts Payable - Supplier        1,160,000

10. PAYMENT PROCESSING
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment Scheduling:
  • Due date calculation
  • Early payment discount check
  • Cash flow optimization
  • Payment batch grouping

Payment Methods:
  • Bank transfer (EFT)
  • Check
  • Mobile money
  • Letter of credit

Payment Entry:
  Dr. Accounts Payable - Supplier    1,160,000
      Cr. Bank Account                       1,160,000

Supplier Notification:
  ✓ Payment advice sent
  ✓ Remittance details
  ✓ Invoice references

11. RECONCILIATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Monthly Activities:
  • Supplier statement matching
  • Aging analysis review
  • Dispute resolution
  • Credit balance clearance

Variance Investigation:
  • Unmatched invoices
  • Unpaid invoices past due
  • Duplicate payments
  • Credit notes pending
```

### Process Flow Diagram

```markdown
VISUAL WORKFLOW

Need → Requisition → Approval → Sourcing → PO → Receipt → Invoice → Payment
  ↓         ↓           ↓          ↓       ↓      ↓         ↓         ↓
Identify  Budget    Manager   RFQ/     Commit Send to  Match  Schedule
Requirement Check   Review    Vendor   Budget Supplier PO/GRN Payment
                              Select                    /INV
```

### Procurement Cycle Times

```markdown
TYPICAL TIMELINES (Best Practice)

┌──────────────────────────┬─────────────┬─────────────┐
│ Process Step             │ Duration    │ Cumulative  │
├──────────────────────────┼─────────────┼─────────────┤
│ Requisition Creation     │ 1 day       │ Day 1       │
│ Approval Workflow        │ 2-3 days    │ Day 3-4     │
│ Supplier Selection (RFQ) │ 5-7 days    │ Day 8-11    │
│ PO Creation & Send       │ 1 day       │ Day 9-12    │
│ Supplier Lead Time       │ 14-30 days  │ Day 23-42   │
│ Goods Receipt            │ 1 day       │ Day 24-43   │
│ Invoice Receipt          │ 0-7 days    │ Day 24-50   │
│ Invoice Processing       │ 2-3 days    │ Day 26-53   │
│ Payment (Net 30)         │ 30 days     │ Day 56-83   │
└──────────────────────────┴─────────────┴─────────────┘

Total Cycle: 56-83 days (Requisition to Payment)
```

### Process Variations

**1. Direct Purchase (Stock Items)**
```
Need → PO → Receipt → Invoice → Payment
Timeline: 20-35 days
Use: Standard inventory items, approved suppliers
```

**2. Service Procurement**
```
Need → Requisition → RFQ → Contract → Service → Invoice → Payment
Timeline: Variable (per contract)
Use: Professional services, maintenance contracts
```

**3. Capital Equipment**
```
Need → Requisition → Budget Approval → RFQ → Evaluation →
Contract Negotiation → PO → Inspection → Payment Schedule
Timeline: 60-180 days
Use: Machinery, vehicles, IT infrastructure
```

**4. Emergency Purchase**
```
Need → Fast-track Approval → Verbal PO → Receipt → Retroactive PO → Invoice → Payment
Timeline: 1-7 days
Use: Critical breakdowns, production stoppage
Note: Requires post-purchase documentation
```

---

**Next:** [Initial Setup & Configuration](./04-initial-setup-and-configuration.md)
