[<-- Back to Index](README.md)

## Sales Process Overview

### The Complete Sales Cycle

```markdown
STAGE 1: LEAD GENERATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Source: Marketing, website, referrals, events
Action: Capture lead in system
Data: Name, company, contact, interest, source
Status: NEW → CONTACTED → QUALIFIED → CONVERTED

Example:
Lead: John Kamau from ABC Manufacturing
Source: Website inquiry form
Interest: Industrial equipment
Score: High (budget confirmed, timeline immediate)
Assigned to: Sarah (Territory: Nairobi)

STAGE 2: OPPORTUNITY CREATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Trigger: Qualified lead with purchase intent
Action: Create opportunity with deal details
Data: Value, probability, expected close date
Status: PROSPECTING → QUALIFICATION → PROPOSAL

Example:
Opportunity: ABC Manufacturing - Equipment Purchase
Value: 5,000,000 KES
Probability: 60%
Expected Close: 2025-02-15
Stage: Proposal

STAGE 3: QUOTATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Trigger: Customer requests formal pricing
Action: Generate professional quote
Content: Items, quantities, prices, terms
Valid: 30 days (configurable)
Status: DRAFT → SENT → ACCEPTED/REJECTED/EXPIRED

Example:
Quotation: QTN-2025-001
Customer: ABC Manufacturing
Items: 3 different equipment models
Total: 5,000,000 KES (before VAT)
Payment Terms: 50% upfront, 50% on delivery
Validity: Valid until 2025-02-12

STAGE 4: SALES ORDER
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Trigger: Customer accepts quotation (PO received)
Action: Convert to confirmed order
Effect: 
  - Inventory reserved
  - Production/procurement triggered (if needed)
  - Delivery scheduled
Status: DRAFT → CONFIRMED → TO DELIVER → COMPLETED

Example:
Sales Order: SO-2025-001 (from QTN-2025-001)
Customer PO: PO/ABC/2025/045
Inventory Status: 2 items in stock, 1 on backorder
Expected Delivery: 2025-02-20
Deposit Received: 2,500,000 KES

STAGE 5: DELIVERY/FULFILLMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Action: Pick, pack, ship items
Document: Delivery Note with packing list
Signature: Customer representative signs
Integration: Inventory updated, COGS calculated
Status: TO DELIVER → PARTIALLY DELIVERED → DELIVERED

Example:
Delivery Note: DN-2025-001
Items: All 3 equipment units
Delivered to: ABC Manufacturing, Industrial Area
Received by: John Kamau (signed)
Delivery Date: 2025-02-20 14:30
Inventory Impact: Stock reduced, COGS: 3,200,000 KES

STAGE 6: INVOICING
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Trigger: Goods delivered (or service completed)
Action: Generate sales invoice
Integration: 
  Dr. Accounts Receivable    5,800,000
      Cr. Sales Revenue              5,000,000
      Cr. VAT Payable                  800,000
Status: DRAFT → SUBMITTED → PAID/OVERDUE

Example:
Sales Invoice: INV-2025-001
Date: 2025-02-20
Amount: 5,800,000 KES (including VAT)
Due Date: 2025-03-02 (Net 10 days)
Payment Status: Partially Paid (deposit applied)
Balance Due: 3,300,000 KES

STAGE 7: PAYMENT COLLECTION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Action: Receive and allocate payment
Methods: Bank transfer, cash, check, mobile money
Entry:
  Dr. Bank/Cash             3,300,000
      Cr. Accounts Receivable       3,300,000
Status: UNPAID → PARTIALLY PAID → PAID

Example:
Payment Entry: PAY-2025-001
Date: 2025-03-01
Amount: 3,300,000 KES (final payment)
Method: Bank transfer
Reference: TRX/2025/12345
Allocated to: INV-2025-001
Customer Balance: 0 (fully paid)

STAGE 8: POST-SALE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Activities: 
  - Customer satisfaction survey
  - Equipment installation support
  - Training provided
  - Warranty registration
  - Upsell opportunities identified
Analytics:
  - Customer lifetime value calculated
  - Margin analysis: 36% gross margin
  - Commission calculated: 250,000 KES
Status: CLOSED-WON → ACTIVE CUSTOMER
```

### Alternative Sales Flows

**1. Retail/Cash Sale (Simplified):**
```
Walk-in Customer → Sales Invoice (immediate) → Payment (immediate)
Timeline: Same day
Documents: Sales invoice only (acts as order, delivery, invoice)
```

**2. Service Sale:**
```
Lead → Opportunity → Quote → Service Order → Service Delivery → Invoice → Payment
Timeline: Varies
Special: May have milestone-based billing
```

**3. Subscription/Recurring:**
```
Initial Sale → Recurring Invoice (automatic monthly/annual) → Auto-payment
Timeline: Ongoing
Special: Automated renewal and billing
```

**4. Project-Based:**
```
RFP → Proposal → Contract → Multiple Deliveries → Progress Billing → Final Payment
Timeline: Months to years
Special: Complex milestone tracking
```

**5. Drop-Ship:**
```
Sales Order → Purchase Order to Supplier → Supplier Ships Direct → Invoice Customer
Timeline: As per supplier
Special: No inventory movement in your warehouse
```

---