[<-- Back to Index](README.md)

## Sales Order Processing

### Sales Order Creation

```markdown
SALES ORDER WORKFLOW

Creation Methods:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Method 1: From Quotation (Most Common)
  1. Open accepted quotation
  2. Click "Create Sales Order"
  3. All data auto-populated
  4. Review and confirm

Method 2: Direct Entry
  - For existing customers
  - Repeat orders
  - Phone/email orders
  - Walk-in sales

Method 3: From Portal
  - Customer self-service portal
  - Online order placement
  - Auto-validated against credit limits

Sales Order Header:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Order No: SO-2025-001 (auto-generated)
Date: 2025-01-25
Customer: ABC Manufacturing Ltd
Contact: John Kamau
Customer PO No: PO/ABC/2025/045 (required)

Reference:
  Quotation: QTN-2025-001
  Opportunity: OPP-2025-001

Addresses:
  Billing: Head Office, Nairobi
  Shipping: Factory Warehouse, Ruiru

Order Details:
  Order Type: Sales / Service / Project
  Price List: Corporate Pricing
  Currency: KES
  Payment Terms: Net 30 Days
  Delivery Date: 2025-02-10
  
Sales Team:
  Sales Person: Sarah Johnson
  Commission: 5% on net amount
  Territory: Nairobi Corporate

Order Items:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────┬─────────────┬─────┬───────┬───────────┐
│ Item   │ Description │ Qty │ Rate  │ Amount    │
├────────┼─────────────┼─────┼───────┼───────────┤
│ MDL-A  │ Machine A   │  1  │1,900K │ 1,900,000 │
│        │ Stock: 2    │     │       │           │
│        │ Reserved: 1 │     │       │           │
│        │ Available:1 │     │       │           │
│        │ Warehouse:  │     │       │           │
│        │ Main Store  │     │       │           │
├────────┼─────────────┼─────┼───────┼───────────┤
│ MDL-B  │ Machine B   │  1  │1,620K │ 1,620,000 │
│        │ Stock: 1 ✓  │     │       │           │
├────────┼─────────────┼─────┼───────┼───────────┤
│ MDL-C  │ Machine C   │  1  │1,080K │ 1,080,000 │
│        │ Stock: 0 ⚠  │     │       │           │
│        │ On Order:2  │     │       │           │
│        │ Expected:   │     │       │           │
│        │ Feb 5, 2025 │     │       │           │
└────────┴─────────────┴─────┴───────┴───────────┘

⚠ Item MDL-C on backorder
Expected delivery: Feb 5
Confirm with customer: Partial delivery or wait?

Stock Reservation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
On Order Confirmation:
  ✓ Reserve MDL-A (Qty: 1) from Main Store
  ✓ Reserve MDL-B (Qty: 1) from Main Store
  ⏳ MDL-C pending stock arrival

Reserved Stock:
  - Cannot be sold to other customers
  - Automatically allocated for this order
  - Released if order cancelled
  - Ages if not delivered (alert after 30 days)

Order Totals:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Item Total:           4,600,000 KES
  Additional Discount:    (92,000) KES  (2%)
  ───────────────────────────────────
  Net Amount:           4,508,000 KES
  VAT (16%):              721,280 KES
  ───────────────────────────────────
  Grand Total:          5,229,280 KES
  
  Deposit Required (50%): 2,614,640 KES

Payment Schedule:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────┬─────────────┬────────────┬──────────┐
│ Due Date │ Description │ Amount     │ Status   │
├──────────┼─────────────┼────────────┼──────────┤
│ Jan 25   │ Deposit(50%)│ 2,614,640  │ Pending  │
│ Feb 10   │ On Delivery │ 2,614,640  │ Pending  │
└──────────┴─────────────┴────────────┴──────────┘

Credit Check:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer: ABC Manufacturing Ltd
Credit Limit: 5,000,000 KES
Current Outstanding: 1,250,000 KES
This Order: 5,229,280 KES
─────────────────────────────────
Total Exposure: 6,479,280 KES
Credit Available: (1,479,280) KES ⚠

⚠ CREDIT LIMIT EXCEEDED

Options:
  □ Request credit limit increase
  □ Require deposit payment first
  □ Get management approval to proceed

Action: Deposit payment required before confirmation
```

### Sales Order Confirmation Process

```markdown
ORDER CONFIRMATION WORKFLOW

Step 1: Validation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
System Checks:
  ✓ Customer details complete
  ✓ Customer PO number provided
  ✓ Items available or on backorder
  ✓ Pricing authorized
  ✓ Credit limit (if applicable)
  ✓ Delivery date feasible
  ✓ Payment terms agreed

If Credit Check Fails:
  → Workflow: Send for approval
  → Approver: Credit Manager
  → Options: Approve / Reject / Require Deposit

Step 2: Approval (if required)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approval Required For:
  - Orders exceeding credit limit
  - Special discount > approval threshold
  - First order from new customer
  - High-value orders (> 5M KES)
  - Backorder situations

Approval Workflow:
  Draft Order → Pending Approval → Approved → Confirmed

Notification:
  To: Credit Manager / Sales Manager
  Subject: Sales Order Approval Required
  Content:
    - Customer details
    - Order value
    - Credit exposure
    - Reason for approval
  
Approver Actions:
  □ Approve (proceed to confirmation)
  □ Approve with conditions (e.g., deposit required)
  □ Reject (order cancelled, customer notified)

Step 3: Confirmation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
On Confirmation:
  Status: DRAFT → CONFIRMED
  
Automatic Actions:
  1. Stock Reserved:
     - Items allocated from warehouse
     - Available stock updated
     - Reservation logged
     
  2. Production/Procurement Triggered:
     If Make-to-Order:
       → Work Order created
       → Material requirements calculated
       → Production scheduled
     
     If Buy-to-Order:
       → Purchase Requisition created
       → Supplier notification
       → Expected delivery tracked
     
  3. Customer Notification:
     Email sent:
       Subject: Order Confirmation SO-2025-001
       Attachment: Order confirmation PDF
       Content: 
         - Order summary
         - Expected delivery date
         - Payment instructions
         - Tracking link
  
  4. Team Notifications:
     → Warehouse: Prepare items for delivery
     → Finance: Expect deposit payment
     → Customer Service: Order in system
  
  5. Integration Events:
     → Inventory: Stock reserved
     → Finance: Accounts receivable prepared
     → Delivery: Shipment scheduled

Step 4: Order Processing
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Order Status Tracking:
  CONFIRMED → TO DELIVER → DELIVERED → TO BILL → COMPLETED

Sub-statuses:
  - Awaiting Payment (deposit)
  - Awaiting Stock (backorder)
  - In Production
  - Ready to Deliver
  - Partially Delivered
  - Awaiting Invoice Approval
```

### Sales Order Modifications

```markdown
MODIFYING CONFIRMED ORDERS

Amendment Scenarios:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Scenario 1: Before Delivery
  Customer requests to add items
  
  Process:
    1. Create Sales Order Amendment
    2. Add new items
    3. Recalculate totals
    4. Check stock availability
    5. Get customer confirmation
    6. Update original order
  
  Impact:
    - Stock re-reserved
    - New delivery date if needed
    - Revised invoice amount
    - Customer notified

Scenario 2: Change Delivery Date
  Customer needs earlier/later delivery
  
  Process:
    1. Check warehouse availability
    2. Verify production schedule
    3. Update delivery date
    4. Notify warehouse team
    5. Email customer confirmation
  
  Impact:
    - Delivery schedule updated
    - Warehouse notified
    - No financial impact

Scenario 3: Quantity Reduction
  Customer wants fewer items
  
  Process:
    1. Update quantities
    2. Release excess reserved stock
    3. Recalculate amounts
    4. Update payment schedule
    5. Issue revised order confirmation
  
  Impact:
    - Stock released back to available
    - Lower invoice amount
    - Potential refund if deposit paid

Scenario 4: Item Substitution
  Item unavailable, offer alternative
  
  Process:
    1. Propose alternative item
    2. Get customer approval
    3. Update order with new item
    4. Adjust pricing if different
    5. Reserve new item stock
  
  Impact:
    - Original item stock released
    - New item stock reserved
    - Price adjustment (+ or -)

Restrictions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Cannot Modify If:
  ✗ Items already delivered
  ✗ Invoice already generated
  ✗ Payment already received

Must Cancel & Re-create If:
  ✗ Major changes (>50% of order)
  ✗ Complete product change
  ✗ Different customer

Modification Audit Trail:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────┬─────────┬────────────────────────┐
│ Date     │ User    │ Change                 │
├──────────┼─────────┼────────────────────────┤
│ Jan 25   │ Sarah   │ Order created          │
│ Jan 25   │ System  │ Status: Confirmed      │
│ Jan 27   │ Sarah   │ Delivery date changed  │
│          │         │ From: Feb 10 → Feb 15  │
│ Jan 28   │ Mike    │ Qty changed: Item B    │
│          │         │ From: 1 → 2 units      │
│ Jan 28   │ System  │ Amount recalculated    │
│          │         │ New total: 7,049,280   │
└──────────┴─────────┴────────────────────────┘
```

### Sales Order Cancellation

```markdown
ORDER CANCELLATION PROCESS

Cancellation Reasons:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Customer request
- Payment not received
- Items not available
- Duplicate order
- Customer credit issues
- Force majeure

Before Delivery:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Can Cancel If:
  ✓ No delivery made
  ✓ No invoice generated
  ✓ Or minor deposit paid (refundable)

Cancellation Process:
  1. Verify cancellation authority
  2. Check payment status
  3. Release reserved stock
  4. Cancel production/purchase orders
  5. Process refund if deposit paid
  6. Update customer record
  7. Notify all stakeholders

Automatic Actions:
  → Stock: Reserved items released
  → Production: Work orders cancelled
  → Purchasing: Purchase orders cancelled
  → Finance: Refund processed (if applicable)
  → Customer: Cancellation email sent

Partial Cancellation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Cancel Specific Items Only:
  - Remove items from order
  - Release their stock
  - Recalculate order total
  - Update delivery schedule

Example:
  Original Order: Items A, B, C
  Cancel Item C
  Result: Order continues with A, B only

After Delivery:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Cannot Cancel:
  ✗ Items already delivered
  ✗ Must process as RETURN instead

Return Process:
  1. Create Sales Return
  2. Receive items back
  3. Inspect condition
  4. Issue credit note
  5. Process refund

Cancellation Charges:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Policy (configurable):
  - If cancelled within 24 hours: No charge
  - If cancelled before production: 10% charge
  - If production started: 25% charge
  - If custom/special order: 50% charge

Deposit Handling:
  Original Deposit: 2,614,640 KES
  Cancellation Fee: 261,464 KES (10%)
  Refund Amount: 2,353,176 KES

Financial Entry:
  Dr. Sales Deposit Account    2,614,640
      Cr. Cash/Bank                     2,353,176
      Cr. Cancellation Income             261,464
```

---