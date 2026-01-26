[<-- Back to Index](README.md)

## Sales Invoicing

### Invoice Creation Process

```markdown
SALES INVOICE GENERATION

Creation Methods:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Method 1: From Delivery Note (Most Common)
  Trigger: Goods delivered, customer signed
  Process:
    1. Open delivery note DN-2025-001
    2. Click "Create Sales Invoice"
    3. All data auto-populated
    4. Review and submit

Method 2: From Sales Order (Direct)
  Use Case: Service delivery, advance invoice
  Process:
    1. Select sales order
    2. Create invoice without delivery note
    3. Revenue recognized on invoice submission

Method 3: Manual Invoice
  Use Case: Ad-hoc sales, adjustments
  Process:
    1. Create new invoice
    2. Select customer
    3. Add items manually
    4. Submit

Method 4: Recurring Invoice
  Use Case: Subscriptions, maintenance contracts
  Process:
    1. Set up recurring invoice template
    2. System auto-generates on schedule
    3. Auto-email to customer

Invoice Header:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice No: INV-2025-001 (auto-generated)
Invoice Type: Tax Invoice / Proforma / Commercial
Invoice Date: 2025-02-10
Due Date: 2025-03-12 (Net 30 days)

Reference Documents:
  Sales Order: SO-2025-001
  Delivery Note: DN-2025-001
  Customer PO: PO/ABC/2025/045
  Quotation: QTN-2025-001

Customer Details:
  Customer: ABC Manufacturing Ltd
  PIN/Tax ID: P000123456A
  Billing Address: Head Office, Nairobi
  Attention: Mary Wanjiku (Finance Manager)
  Email: mary.wanjiku@abc.com

Company Details (Auto):
  Company PIN: P987654321Z
  Address: [From company master]
  Bank Details: [For payment]

Invoice Items:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Based on Delivery DN-2025-001:
┌────────┬──────────────┬─────┬────────────┬────────────┐
│ Item   │ Description  │ Qty │ Rate       │ Amount     │
├────────┼──────────────┼─────┼────────────┼────────────┤
│ MDL-A  │ Machine A    │  1  │ 1,900,000  │ 1,900,000  │
│        │ SN: 12345    │     │            │            │
│        │ Delivered:   │     │            │            │
│        │ Feb 10, 2025 │     │            │            │
├────────┼──────────────┼─────┼────────────┼────────────┤
│ MDL-B  │ Machine B    │  1  │ 1,620,000  │ 1,620,000  │
│        │ SN: 12346    │     │            │            │
├────────┼──────────────┼─────┼────────────┼────────────┤
│ SVC-01 │ Installation │  1  │   150,000  │   150,000  │
│        │ Completed:   │     │            │            │
│        │ Feb 12, 2025 │     │            │            │
└────────┴──────────────┴─────┴────────────┴────────────┘

Note: MDL-C not invoiced yet (not delivered)

Invoice Calculations:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Item Total:                3,670,000 KES
Additional Discount (2%):    (73,400) KES
───────────────────────────────────────
Net Amount:                3,596,600 KES
VAT (16%):                   575,456 KES
───────────────────────────────────────
Grand Total:               4,172,056 KES

Less: Deposit Applied:    (2,614,640) KES
───────────────────────────────────────
Balance Due:               1,557,416 KES

Payment Schedule:
  Deposit (Received Jan 25): 2,614,640 KES ✓ PAID
  Balance (Due Mar 12):      1,557,416 KES ⏳ PENDING

Financial Integration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
On Invoice Submission:

Journal Entry (Automatic):
Dr. Accounts Receivable - ABC     4,172,056
    Cr. Sales Revenue - Equipment       3,596,600
    Cr. VAT Payable                       575,456

Description: Sales Invoice INV-2025-001
Customer: ABC Manufacturing Ltd
Sales Order: SO-2025-001
Cost Center: Sales Department
Sales Person: Sarah Johnson

If Deposit Already Paid:
Dr. Customer Deposit Account      2,614,640
    Cr. Accounts Receivable - ABC       2,614,640

Description: Apply deposit from SO-2025-001

Net Receivable:
Opening AR Balance:         1,250,000
New Invoice:               4,172,056
Less: Deposit Applied:    (2,614,640)
───────────────────────────────────
Current AR Balance:        2,807,416 KES

Cost of Goods Sold (from Delivery):
Dr. Cost of Goods Sold            3,200,000
    Cr. Inventory - Equipment           3,200,000

Description: COGS for DN-2025-001
Items: MDL-A (1,600,000), MDL-B (1,600,000)

Margin Analysis:
Revenue:                   3,596,600 KES
Less: COGS:               (3,200,000) KES
───────────────────────────────────
Gross Profit:               396,600 KES
Gross Margin:                  11.0%

Commission Calculation:
Net Revenue:              3,596,600 KES
Commission Rate:                  5%
Commission Amount:          179,830 KES
Payable to: Sarah Johnson
```

### Invoice Document Format

```markdown
TAX INVOICE LAYOUT

┌──────────────────────────────────────────────┐
│           [COMPANY LOGO]                     │
│                                              │
│          TAX INVOICE                         │
│                                              │
│ Company Name                Invoice No:      │
│ PIN: P987654321Z            INV-2025-001     │
│ Address Line 1              Date: Feb 10, 25 │
│ Address Line 2              Due: Mar 12, 25  │
│ Phone: +254-20-xxx-xxxx                      │
│ Email: sales@company.com                     │
├──────────────────────────────────────────────┤
│ BILL TO:                                     │
│                                              │
│ ABC Manufacturing Ltd                        │
│ PIN: P000123456A                             │
│ Head Office, Industrial Area                 │
│ Nairobi, Kenya                               │
│                                              │
│ Attention: Mary Wanjiku, Finance Manager     │
│ Email: mary.wanjiku@abc.com                  │
│ Phone: +254-700-234-567                      │
├──────────────────────────────────────────────┤
│ REFERENCE:                                   │
│ Sales Order: SO-2025-001                     │
│ Your PO: PO/ABC/2025/045                     │
│ Delivery Note: DN-2025-001                   │
│ Quotation: QTN-2025-001                      │
├──────────────────────────────────────────────┤
│ Item  Description       Qty  Rate     Amount │
│ ────  ──────────────    ─── ──────   ─────── │
│ 1.    Machine Model A    1  1,900,000        │
│       Serial: SN12345          1,900,000     │
│       Delivered: Feb 10, 2025                │
│                                              │
│ 2.    Machine Model B    1  1,620,000        │
│       Serial: SN12346          1,620,000     │
│       Delivered: Feb 10, 2025                │
│                                              │
│ 3.    Installation       1    150,000        │
│       Service                    150,000     │
│       Completed: Feb 12, 2025                │
│                                              │
│                       Subtotal: 3,670,000    │
│                       Discount:   (73,400)   │
│                       Net:      3,596,600    │
│                       VAT(16%):   575,456    │
│                       ─────────────────────  │
│                       TOTAL:    4,172,056    │
│                                              │
│ LESS: DEPOSIT PAID                           │
│ Payment Ref: PAY-2025-001  (2,614,640)      │
│ Date: January 25, 2025                       │
│                       ─────────────────────  │
│                    BALANCE DUE: 1,557,416    │
│                                              │
│ Amount in Words:                             │
│ One Million Five Hundred Fifty Seven         │
│ Thousand Four Hundred Sixteen Shillings Only │
├──────────────────────────────────────────────┤
│ PAYMENT INSTRUCTIONS:                        │
│                                              │
│ Bank: Equity Bank Kenya                      │
│ Account Name: Company Name Ltd               │
│ Account Number: 0123456789                   │
│ Branch: Industrial Area                      │
│ Swift Code: EQBLKENA                         │
│                                              │
│ M-Pesa Till: 123456 (for amounts <100K)     │
│                                              │
│ Payment Reference: INV-2025-001              │
├──────────────────────────────────────────────┤
│ PAYMENT TERMS:                               │
│ Net 30 days from invoice date                │
│ Due Date: March 12, 2025                     │
│ Late payment interest: 2% per month          │
├──────────────────────────────────────────────┤
│ NOTES:                                       │
│ • Warranty: 12 months from delivery date     │
│ • For queries: accounts@company.com          │
│ • This is a computer-generated invoice       │
│                                              │
│ [QR Code for Payment]                        │
│                                              │
│ Thank you for your business!                 │
└──────────────────────────────────────────────┘
```

### Invoice Status Management

```markdown
INVOICE LIFECYCLE

Status Flow:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DRAFT → SUBMITTED → PAID / PARTIALLY PAID / OVERDUE

Status Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DRAFT:
  - Invoice being prepared
  - Can be edited
  - Not posted to GL
  - Not sent to customer
  - Not affecting AR

SUBMITTED:
  - Invoice finalized and sent
  - Posted to GL
  - AR balance updated
  - Email sent to customer
  - Read-only (cannot edit)
  - Payment tracking active

PAID:
  - Full payment received
  - AR cleared
  - Payment allocated
  - Receipt issued
  - Commission released

PARTIALLY PAID:
  - Part payment received
  - Balance outstanding tracked
  - Aging starts on balance
  - Follow-up for balance

OVERDUE:
  - Past due date
  - Payment not received
  - Collection actions triggered
  - Aging bucket assigned
  - Interest may apply

CANCELLED:
  - Invoice cancelled before payment
  - AR reversed
  - Must have authorization
  - Reason documented

Invoice Actions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

From SUBMITTED:
  □ Record Payment
  □ Send Reminder
  □ Generate Statement
  □ Create Credit Note
  □ Cancel (with approval)

From PAID:
  □ Generate Receipt
  □ View Payment History
  □ Issue Credit Note (for returns)

From OVERDUE:
  □ Send Reminder (automatic)
  □ Escalate to Collections
  □ Apply Late Fee
  □ Record Payment

Aging Analysis:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Invoice: INV-2025-001
Invoice Date: Feb 10, 2025
Due Date: Mar 12, 2025
Balance Due: 1,557,416 KES

As of Date: Mar 15, 2025
Status: OVERDUE
Days Overdue: 3 days
Aging Bucket: Current (0-30 days)

Reminder Actions:
  Mar 10 (2 days before): Friendly reminder ✓
  Mar 13 (1 day after): First reminder ✓
  Mar 20 (8 days after): Second reminder ⏳
  Mar 27 (15 days after): Final notice ⏳
  Apr 10 (30 days after): Escalate to collections ⏳

Payment Tracking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────┬────────────┬────────────┬────────────┐
│ Date       │ Reference  │ Amount     │ Balance    │
├────────────┼────────────┼────────────┼────────────┤
│ Jan 25     │ Deposit    │ 2,614,640  │ 1,557,416  │
│ Mar 15     │ Partial    │   500,000  │ 1,057,416  │
│ Mar 20     │ Balance    │ 1,057,416  │         0  │
└────────────┴────────────┴────────────┴────────────┘
```

### Proforma Invoice

```markdown
PROFORMA INVOICE

Purpose:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Advance invoice for customs/import
- Bank LC requirements
- Budget approval documentation
- Not a demand for payment
- Not posted to accounting

Key Differences from Tax Invoice:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Proforma Invoice:
  ✗ Not posted to GL
  ✗ Not creating AR
  ✗ No tax obligation
  ✗ Not for payment demand
  ✓ For information/planning only
  ✓ Can be revised freely
  ✓ No accounting impact

Tax Invoice:
  ✓ Posted to GL
  ✓ Creates AR
  ✓ Tax obligation created
  ✓ Legal demand for payment
  ✗ Cannot be revised (need credit note)
  ✓ Full accounting impact

Proforma Usage:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Export Sales:
   Customer needs proforma for import clearance
   
2. Large Projects:
   Customer needs quotation in invoice format
   for budget approval
   
3. Government/Tender:
   Required for procurement process

Conversion Process:
  Proforma Created → Customer Approves → 
  Sales Order → Delivery → Tax Invoice

Document Marking:
  Header: "PROFORMA INVOICE"
  Footer: "THIS IS NOT A TAX INVOICE"
  Watermark: "PROFORMA - FOR PLANNING ONLY"
```

---