[<-- Back to Index](README.md)

## Payment Processing

### Payment Scheduling

```markdown
PAYMENT PLANNING

Payment Run Schedule:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Frequency: Weekly (Every Friday)
Next Payment Run: 2025-04-26

Due Invoices for Payment:
┌──────────────┬────────────────┬────────────┬──────────┬──────────┐
│ Invoice      │ Supplier       │ Amount     │ Due Date │ Priority │
├──────────────┼────────────────┼────────────┼──────────┼──────────┤
│ PINV-2025-234│ Ace Steel      │  960,118   │ May 01   │ Normal   │
│ PINV-2025-245│ ChemSupply     │  245,000   │ Apr 25   │ Urgent   │
│ PINV-2025-256│ Office Supply  │   85,000   │ Apr 28   │ Normal   │
│ PINV-2025-267│ Utility Co     │  290,000   │ Apr 30   │ Normal   │
│ PINV-2025-278│ Maintenance Ltd│  450,000   │ Apr 22   │ Overdue  │
│              │                │            │          │          │
│ TOTAL        │                │2,030,118   │          │          │
└──────────────┴────────────────┴────────────┴──────────┴──────────┘

Payment Selection Criteria:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Automatic Selection:
  ☑ Overdue invoices (past due date)
  ☑ Due within next 7 days
  ☑ Early payment discount available
  ☑ Strategic supplier priority

Manual Selection:
  □ Hold payments (cash flow management)
  □ Partial payments (if needed)
  □ Advance payments (special cases)

Cash Flow Check:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Available Cash: 5,500,000 KES
Planned Payments: 2,030,118 KES
Safety Reserve: 1,000,000 KES
─────────────────────────────
Cash After Payment: 3,469,882 KES ✓ Adequate

Status: ✓ PROCEED WITH PAYMENT RUN

Early Payment Discount Optimization:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice: PINV-2025-289
Supplier: Kenya Fasteners Ltd
Amount: 500,000 KES
Terms: 2/10 Net 30

Calculation:
  Invoice Date: Apr 15
  Due Date: May 15 (30 days)
  Discount Valid Until: Apr 25 (10 days)
  
  If Paid by Apr 25:
    Discount (2%): 10,000 KES
    Payment: 490,000 KES
    Savings: 10,000 KES
    
  If Paid After Apr 25:
    Payment: 500,000 KES
    No discount

Decision: Include in Apr 26 payment run to capture discount
ROI: 2% in 5 days = 146% annualized return
```

### Payment Batch Creation

```markdown
PAYMENT BATCH PROCESSING

Payment Batch No: PAY-BATCH-2025-017
Date: 2025-04-26
Created By: Finance Officer
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Batch Summary:
  Number of Payments: 5
  Total Amount: 2,030,118 KES
  Payment Method: Bank Transfer
  Bank Account: Equity Bank - Main Operating

Payment Details:
┌──────┬──────────────┬────────────────┬────────────┬──────────┐
│ No   │ Invoice      │ Supplier       │ Amount     │ Ref      │
├──────┼──────────────┼────────────────┼────────────┼──────────┤
│  1   │ PINV-2025-278│ Maintenance Ltd│  450,000   │ Overdue  │
│  2   │ PINV-2025-245│ ChemSupply     │  245,000   │ Due Soon │
│  3   │ PINV-2025-256│ Office Supply  │   85,000   │ Normal   │
│  4   │ PINV-2025-267│ Utility Co     │  290,000   │ Normal   │
│  5   │ PINV-2025-289│ Kenya Fasteners│  490,000   │ Discount │
│      │              │                │            │          │
│ TOTAL│              │                │1,560,000   │          │
└──────┴──────────────┴────────────────┴────────────┴──────────┘

Note: PINV-2025-234 (Ace Steel, 960K) excluded
Reason: Payment scheduled for May 1 (optimal timing)

Payment Entries Created:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. PAY-2025-00401 (Maintenance Ltd)
   Amount: 450,000 KES
   Invoice: PINV-2025-278
   
   Dr. Accounts Payable - Maintenance   450,000
       Cr. Bank - Equity                       450,000

2. PAY-2025-00402 (ChemSupply)
   Amount: 245,000 KES
   Invoice: PINV-2025-245
   
   Dr. Accounts Payable - ChemSupply    245,000
       Cr. Bank - Equity                       245,000

3. PAY-2025-00403 (Office Supply)
   Amount: 85,000 KES
   Invoice: PINV-2025-256
   
   Dr. Accounts Payable - Office Supply  85,000
       Cr. Bank - Equity                        85,000

4. PAY-2025-00404 (Utility Co)
   Amount: 290,000 KES
   Invoice: PINV-2025-267
   
   Dr. Accounts Payable - Utility        290,000
       Cr. Bank - Equity                       290,000

5. PAY-2025-00405 (Kenya Fasteners - with discount)
   Amount: 490,000 KES (after 2% discount)
   Invoice: PINV-2025-289
   
   Dr. Accounts Payable - Kenya Fasteners 500,000
       Cr. Purchase Discount Income              10,000
       Cr. Bank - Equity                        490,000

Total Debit: 1,570,000 KES
Total Credit: 1,560,000 KES (Bank) + 10,000 KES (Discount)
```

### Payment Approval

```markdown
PAYMENT BATCH APPROVAL

Batch: PAY-BATCH-2025-017
Amount: 1,560,000 KES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Approval Matrix (Payments):
  < 500K: Finance Officer
  500K - 2M: Finance Manager
  > 2M: Finance Manager + CFO

This Batch: 1,560,000 KES → Finance Manager Approval Required

Level 1: Finance Manager
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approver: Finance Manager
Date: 2025-04-26 10:00 AM

Review Checklist:
  ☑ All invoices properly matched
  ☑ Invoice approvals verified
  ☑ Supplier bank details confirmed
  ☑ Cash flow impact acceptable
  ☑ No duplicate payments
  ☑ Payment calculations correct

Decision: APPROVED
Comments: "All checks passed. Proceed with payment."

Authorization Code: FM-AUTH-20250426-001

Final Status: APPROVED FOR EXECUTION
Next: Execute bank transfers
```

### Payment Execution

```markdown
BANK TRANSFER PROCESSING

Payment Method: Electronic Funds Transfer (EFT)
Bank: Equity Bank - Main Operating Account
Batch: PAY-BATCH-2025-017
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Bank File Generation:
Date: 2025-04-26 11:00 AM
Format: SWIFT MT103 (International Standard)

Transfers Generated:
┌──────┬──────────────────┬────────────┬──────────────┬──────────┐
│ Ref  │ Beneficiary      │ Amount     │ Bank/Account │ Status   │
├──────┼──────────────────┼────────────┼──────────────┼──────────┤
│ 001  │ Maintenance Ltd  │  450,000   │ KCB/1234567  │ Queued   │
│ 002  │ ChemSupply       │  245,000   │ Equity/9876  │ Queued   │
│ 003  │ Office Supply    │   85,000   │ Co-op/5555   │ Queued   │
│ 004  │ Utility Company  │  290,000   │ KCB/7890     │ Queued   │
│ 005  │ Kenya Fasteners  │  490,000   │ Equity/4321  │ Queued   │
└──────┴──────────────────┴────────────┴──────────────┴──────────┘

Bank Portal Submission:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Time: 2025-04-26 11:30 AM
Portal: Equity Bank Business Online
User: Finance Officer

Upload Status:
  File Uploaded: ✓ Success
  Validation: ✓ Passed
  Total Amount: 1,560,000 KES
  Number of Transactions: 5

Bank Authorization:
  Approver 1: Finance Manager (Token)
  Approver 2: CFO (Token)
  Status: ✓ SUBMITTED TO BANK

Bank Processing:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Processing Time: 2025-04-26 2:00 PM

Transaction Status:
┌──────┬──────────────────┬──────────┬────────────────┐
│ Ref  │ Beneficiary      │ Amount   │ Status         │
├──────┼──────────────────┼──────────┼────────────────┤
│ 001  │ Maintenance Ltd  │  450,000 │ ✓ Completed    │
│ 002  │ ChemSupply       │  245,000 │ ✓ Completed    │
│ 003  │ Office Supply    │   85,000 │ ✓ Completed    │
│ 004  │ Utility Company  │  290,000 │ ✓ Completed    │
│ 005  │ Kenya Fasteners  │  490,000 │ ✓ Completed    │
└──────┴──────────────────┴──────────┴────────────────┘

All Payments: ✓ SUCCESSFUL
Batch Status: COMPLETED

System Updates:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice Status Updates:
  PINV-2025-278: PAID ✓
  PINV-2025-245: PAID ✓
  PINV-2025-256: PAID ✓
  PINV-2025-267: PAID ✓
  PINV-2025-289: PAID ✓

Supplier Balance Updates:
  Each supplier AP balance reduced accordingly

Bank Account Balance:
  Opening: 5,500,000 KES
  Payments: (1,560,000) KES
  Closing: 3,940,000 KES
```

### Payment Notification & Documentation

```markdown
PAYMENT ADVICE TO SUPPLIERS

Remittance Advice - Automatic Email:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
To: accounts@maintenanceltd.co.ke
From: ap@awo.co.ke
Subject: Payment Advice - Payment Reference PAY-2025-00401

Dear Maintenance Ltd,

We have processed payment as follows:

Payment Reference: PAY-2025-00401
Payment Date: April 26, 2025
Amount: 450,000 KES
Method: Bank Transfer to KCB Account 1234567

Invoice Details:
┌──────────────┬────────────┬────────────┬────────────┐
│ Invoice No   │ Date       │ Amount     │ Paid       │
├──────────────┼────────────┼────────────┼────────────┤
│ ML-INV-789   │ Mar 15     │  450,000   │  450,000   │
│              │            │            │            │
│ TOTAL        │            │  450,000   │  450,000   │
└──────────────┴────────────┴────────────┴────────────┘

Our Reference: PO-2025-00178

The payment should reflect in your account within 24 hours.

For any queries, contact:
Accounts Payable - +254 720 567 890
Email: ap@awo.co.ke

Regards,
AWO Manufacturing Ltd

Attachment: Remittance_PAY-2025-00401.pdf

Payment With Withholding Tax:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Supplier: Consultancy Services Ltd
Invoice: PINV-2025-00285
Amount: 580,000 KES (including VAT)

Calculation:
  Professional Fees: 500,000 KES
  VAT (16%): 80,000 KES
  Gross Invoice: 580,000 KES
  
  WHT (5% on fees): 25,000 KES
  Net Payment: 555,000 KES

Payment Advice:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Invoice Amount: 580,000 KES
Less WHT (5%): (25,000) KES
──────────────────────────
Net Paid: 555,000 KES

WHT Certificate: Will be provided after KRA filing

Breakdown:
  Professional Fees: 500,000 KES
  VAT: 80,000 KES
  Subtotal: 580,000 KES
  WHT Deducted: (25,000) KES
  Net Payment: 555,000 KES

WHT Details:
  Rate: 5%
  Base Amount: 500,000 KES
  WHT Amount: 25,000 KES
  To be remitted to KRA by AWO

WHT Certificate will be issued after KRA filing.
```

### Payment Reconciliation

```markdown
BANK RECONCILIATION

Daily Bank Statement Import:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: 2025-04-26
Bank: Equity Bank - Main Operating

Statement Lines (Debits):
┌────────────┬─────────────────┬────────────┬──────────────┐
│ Date       │ Description     │ Amount     │ Reference    │
├────────────┼─────────────────┼────────────┼──────────────┤
│ Apr 26     │ TRF/Maintenance │  450,000   │ FT123456789  │
│ Apr 26     │ TRF/ChemSupply  │  245,000   │ FT123456790  │
│ Apr 26     │ TRF/Office Sup  │   85,000   │ FT123456791  │
│ Apr 26     │ TRF/Utility     │  290,000   │ FT123456792  │
│ Apr 26     │ TRF/Kenya Fast  │  490,000   │ FT123456793  │
│ Apr 26     │ Bank Charges    │    1,500   │ -            │
└────────────┴─────────────────┴────────────┴──────────────┘

Automatic Matching:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Statement Line 1 (450,000):
  Matched to: PAY-2025-00401 ✓
  Status: RECONCILED

Statement Line 2 (245,000):
  Matched to: PAY-2025-00402 ✓
  Status: RECONCILED

Statement Line 3 (85,000):
  Matched to: PAY-2025-00403 ✓
  Status: RECONCILED

Statement Line 4 (290,000):
  Matched to: PAY-2025-00404 ✓
  Status: RECONCILED

Statement Line 5 (490,000):
  Matched to: PAY-2025-00405 ✓
  Status: RECONCILED

Statement Line 6 (1,500 - Bank Charges):
  No payment entry
  Action: Create expense entry
  
  Dr. Bank Charges (Expense)    1,500
      Cr. Bank Account                 1,500
  
  Status: RECONCILED

Reconciliation Summary:
  System Payments: 1,560,000 KES (5 transactions)
  Bank Debits: 1,561,500 KES (6 transactions)
  Difference: 1,500 KES (bank charges) ✓ Explained
  
  Status: ✓ FULLY RECONCILED
```

---

**Next:** [Purchase Returns & Debit Notes](./14-purchase-returns-and-debit-notes.md)
