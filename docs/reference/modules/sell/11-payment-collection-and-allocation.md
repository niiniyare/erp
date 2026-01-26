[<-- Back to Index](README.md)

## Payment Collection & Allocation

### Payment Entry Process

```markdown
PAYMENT RECEIPT WORKFLOW

Payment Sources:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Bank transfer
- Cash payment
- Check payment
- Mobile money (M-Pesa, Airtel Money)
- Credit card
- Online payment gateway

Payment Entry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment Entry No: PAY-2025-015 (auto)
Payment Date: 2025-03-15
Customer: ABC Manufacturing Ltd

Payment Details:
  Amount Received: 1,557,416 KES
  Payment Method: Bank Transfer
  Bank: Equity Bank
  Reference: TRX/2025/54321
  Received In: Company Main Account
  
Party Details:
  Paid By: ABC Manufacturing Ltd
  Account: Accounts Receivable - ABC
  
Allocation Method:
  ○ Auto-allocate (oldest first - FIFO)
  ○ Manual allocation
  ● Specific invoice selection

Outstanding Invoices:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────────┬─────────┬───────────┬───────────┐
│ Invoice      │ Date    │ Amount    │ Allocate  │
├──────────────┼─────────┼───────────┼───────────┤
│ INV-2025-001 │ Feb 10  │ 1,557,416 │ 1,557,416 │
│ INV-2024-125 │ Dec 20  │   250,000 │         0 │
│ INV-2025-010 │ Jan 30  │   180,000 │         0 │
└──────────────┴─────────┴───────────┴───────────┘

Selected for Payment:
  INV-2025-001: 1,557,416 KES (FULL PAYMENT)
  
Payment Allocation:
  Total Received: 1,557,416 KES
  Total Allocated: 1,557,416 KES
  Unallocated: 0 KES ✓

Financial Entry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Journal Entry (Automatic):

Dr. Bank - Main Account            1,557,416
    Cr. Accounts Receivable - ABC        1,557,416

Description: Payment for INV-2025-001
Reference: TRX/2025/54321
Payment Entry: PAY-2025-015

Invoice Status Update:
  INV-2025-001:
    Status: PARTIALLY PAID → PAID ✓
    Outstanding: 1,557,416 → 0
    Payment Date: 2025-03-15
    Days to Payment: 33 days (from invoice date)
    
Customer Account Summary:
  Previous Balance: 2,807,416 KES
  Payment Received: (1,557,416) KES
  ──────────────────────────────
  Current Balance: 1,250,000 KES

Automatic Actions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Email Receipt to Customer:
   Subject: Payment Receipt - PAY-2025-015
   Attachment: Official receipt PDF
   
2. Update Customer Credit:
   Available Credit Increased by 1,557,416 KES
   
3. Release Commission:
   If payment-based commission:
     Invoice: INV-2025-001
     Commission: 179,830 KES
     Payable to: Sarah Johnson
     Status: EARNED (payment received)
     
4. Cancel Payment Reminders:
   Stop reminder emails for INV-2025-001
   
5. Update Reporting:
   DSO calculation
   Collection metrics
   Cash flow forecast
```

### Payment Allocation Scenarios

```markdown
ALLOCATION SCENARIOS

Scenario 1: Exact Match
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment: 1,557,416 KES
Invoice: INV-2025-001 = 1,557,416 KES

Result: Perfect match, fully allocate ✓

Scenario 2: Partial Payment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment: 500,000 KES
Invoice: INV-2025-001 = 1,557,416 KES

Allocation:
  Paid: 500,000 KES
  Balance: 1,057,416 KES
  Status: PARTIALLY PAID
  
Follow-up:
  - Send acknowledgment
  - Request balance payment
  - Track remaining amount

Scenario 3: Overpayment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment: 1,600,000 KES
Invoice: INV-2025-001 = 1,557,416 KES

Allocation:
  To Invoice: 1,557,416 KES
  Excess: 42,584 KES
  
Excess Handling Options:
  ○ Credit to customer account (advance payment)
  ○ Refund to customer
  ○ Allocate to other outstanding invoices
  
Selected: Credit to account
Entry:
  Dr. Bank                      1,600,000
      Cr. Accounts Receivable         1,557,416
      Cr. Customer Advance               42,584

Scenario 4: Multiple Invoice Payment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment: 2,000,000 KES

Outstanding Invoices:
  INV-2025-001: 1,557,416 KES
  INV-2024-125:   250,000 KES
  INV-2025-010:   180,000 KES
  Total: 1,987,416 KES

Allocation (Auto - Oldest First):
  INV-2024-125: 250,000 KES (PAID) ✓
  INV-2025-001: 1,557,416 KES (PAID) ✓
  INV-2025-010: 180,000 KES (PAID) ✓
  Excess: 12,584 KES (advance credit)

All three invoices marked PAID ✓

Scenario 5: Payment Without Reference
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Bank Statement: 500,000 KES from "ABC Mfg"
No invoice reference

Process:
  1. Create unallocated payment entry
  2. Contact customer for invoice reference
  3. Manual allocation once confirmed
  
Temporary Entry:
  Dr. Bank                       500,000
      Cr. Unallocated Payments        500,000
  
After Confirmation:
  Dr. Unallocated Payments       500,000
      Cr. Accounts Receivable - ABC   500,000

Scenario 6: Early Payment Discount
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice Terms: 2/10 Net 30
  (2% discount if paid within 10 days)

Invoice Amount: 1,000,000 KES
Invoice Date: Mar 1
Due Date: Mar 31
Discount Valid Until: Mar 11

Payment Date: Mar 8 ✓ (within discount period)
Payment Amount: 980,000 KES (2% discount taken)

Entry:
  Dr. Bank                       980,000
  Dr. Sales Discount              20,000
      Cr. Accounts Receivable         1,000,000

Invoice Status: PAID ✓
Discount Given: 20,000 KES
Effective Discount: 2%
```

### Payment Reconciliation

```markdown
BANK RECONCILIATION INTEGRATION

Daily Bank Statement Import:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: Mar 15, 2025

Bank Statement Lines:
┌──────────┬─────────────┬──────────┬───────────┐
│ Date     │ Description │ Debit    │ Credit    │
├──────────┼─────────────┼──────────┼───────────┤
│ Mar 15   │ TRX/54321   │          │ 1,557,416 │
│          │ ABC MFG     │          │           │
├──────────┼─────────────┼──────────┼───────────┤
│ Mar 15   │ TRX/54322   │          │   850,000 │
│          │ XYZ LTD     │          │           │
├──────────┼─────────────┼──────────┼───────────┤
│ Mar 15   │ Bank Fees   │    2,500 │           │
└──────────┴─────────────┴──────────┴───────────┘

Automatic Matching:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Statement Line 1: TRX/54321, 1,557,416
  ↓ Match Rules:
    - Amount matches
    - Customer name contains "ABC"
    - Reference in description
  ↓ Matched to:
    Payment Entry: PAY-2025-015 ✓
    Status: RECONCILED

Statement Line 2: TRX/54322, 850,000
  ↓ Search for matching payment
    - Amount: 850,000
    - Customer: XYZ
  ↓ Matched to:
    Payment Entry: PAY-2025-016 ✓
    Status: RECONCILED

Statement Line 3: Bank Fees, 2,500
  ↓ No matching payment entry
  ↓ Create GL Entry:
    Dr. Bank Charges Expense    2,500
        Cr. Bank Account              2,500
  Status: RECONCILED (expense entry)

Unreconciled Items:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Payment Entries Not in Bank:
  PAY-2025-017: 450,000 KES (Check not cleared)
  Reason: Outstanding check
  Action: Wait for clearance

Bank Entries Not Matched:
  None ✓ All reconciled

Reconciliation Summary:
  Opening Balance: 5,250,000 KES
  Total Receipts: 2,407,416 KES
  Total Payments: 2,500 KES
  Closing Balance: 7,654,916 KES ✓
  
  GL Balance: 7,654,916 KES ✓
  Difference: 0 ✓ RECONCILED
```

### Payment Receipt Document

```markdown
OFFICIAL RECEIPT FORMAT

┌────────────────────────────────────────────┐
│         [COMPANY LOGO]                     │
│                                            │
│        OFFICIAL RECEIPT                    │
│                                            │
│ Receipt No: RCP-2025-015                   │
│ Date: March 15, 2025                       │
│                                            │
│ Company PIN: P987654321Z                   │
├────────────────────────────────────────────┤
│ RECEIVED FROM:                             │
│                                            │
│ ABC Manufacturing Ltd                      │
│ PIN: P000123456A                           │
│ Head Office, Industrial Area               │
│ Nairobi, Kenya                             │
├────────────────────────────────────────────┤
│ THE SUM OF:                                │
│                                            │
│ KES 1,557,416.00                           │
│                                            │
│ (One Million Five Hundred Fifty Seven      │
│  Thousand Four Hundred Sixteen Shillings)  │
├────────────────────────────────────────────┤
│ BEING PAYMENT FOR:                         │
│                                            │
│ Invoice No: INV-2025-001                   │
│ Invoice Date: February 10, 2025            │
│ Invoice Amount: 4,172,056.00               │
│ Previous Payments: 2,614,640.00            │
│ This Payment: 1,557,416.00                 │
│ Balance: 0.00                              │
│                                            │
│ Status: PAID IN FULL ✓                     │
├────────────────────────────────────────────┤
│ PAYMENT DETAILS:                           │
│                                            │
│ Payment Method: Bank Transfer              │
│ Bank: Equity Bank                          │
│ Reference: TRX/2025/54321                  │
│ Date: March 15, 2025                       │
├────────────────────────────────────────────┤
│ ACCOUNT SUMMARY:                           │
│                                            │
│ Previous Balance: 2,807,416.00             │
│ Payment Received: (1,557,416.00)           │
│ Current Balance: 1,250,000.00              │
├────────────────────────────────────────────┤
│ This is a computer-generated receipt       │
│                                            │
│ For: ABC Manufacturing Ltd                 │
│                                            │
│ Received by: _______________               │
│ Signature: _______________                │
│ Date: _______________                      │
│                                            │
│ [Company Stamp]                            │
│                                            │
│ Thank you for your payment!                │
└────────────────────────────────────────────┘
```

---

