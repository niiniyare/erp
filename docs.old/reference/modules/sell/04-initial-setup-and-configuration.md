[<-- Back to Index](README.md)

## Initial Setup & Configuration

### Prerequisites Checklist

```markdown
BEFORE CONFIGURING SALES MODULE:

□ Finance Module Setup:
  ✓ Chart of Accounts configured
  ✓ Revenue accounts created
  ✓ Receivables account set up
  ✓ Tax accounts defined
  ✓ Fiscal year active

□ Tenant Configuration:
  ✓ Primary tenant created
  ✓ Company information complete
  ✓ Base currency set
  ✓ Timezone configured

□ Entity Structure:
  ✓ Company hierarchy defined
  ✓ Departments/divisions created
  ✓ Cost centers set up
  ✓ Locations/warehouses identified

□ User & Security:
  ✓ RBAC roles defined
  ✓ Sales team users created
  ✓ Approval hierarchies mapped
  ✓ Territory access rules

□ Inventory Module (if available):
  ✓ Product catalog ready
  ✓ Stock locations defined
  ✓ Pricing structure prepared

□ Business Rules Documentation:
  ✓ Credit policies documented
  ✓ Discount approval matrix
  ✓ Commission structure defined
  ✓ Payment terms standard list
```

### Setup Sequence

**Recommended Implementation Order:**

```markdown
PHASE 1: MASTER DATA (Week 1)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Sales Settings Configuration
2. Document Numbering Series
3. Terms & Conditions Templates
4. Email Templates
5. Print Formats

PHASE 2: CUSTOMER DATA (Week 1-2)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Customer Groups
2. Customer Categories
3. Price Lists
4. Payment Terms
5. Customer Master Data Import

PHASE 3: SALES TEAM (Week 2)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Sales Territories
2. Sales Team Setup
3. Sales Person Records
4. Target/Quota Assignment
5. Commission Structure

PHASE 4: PRICING (Week 2-3)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Pricing Rules
2. Discount Rules
3. Promotional Schemes
4. Volume Discounts
5. Customer-Specific Pricing

PHASE 5: WORKFLOWS (Week 3)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Quotation Approval Workflow
2. Discount Authorization
3. Credit Limit Override
4. Sales Order Confirmation
5. Return Authorization

PHASE 6: INTEGRATION (Week 3-4)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Finance Module GL Accounts
2. Inventory Integration Points
3. Email Server Configuration
4. Payment Gateway (if applicable)
5. External System APIs

PHASE 7: TESTING (Week 4)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. End-to-end process testing
2. Workflow validation
3. Integration testing
4. User acceptance testing
5. Performance testing

PHASE 8: GO-LIVE (Week 5)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Final data migration
2. User training
3. Parallel run (optional)
4. Cutover
5. Post-go-live support
```

### Step 1: Sales Settings Configuration

```markdown
SALES MODULE SETTINGS

General Settings:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Company: Select primary company
□ Default Customer Group: Retail/Wholesale/Corporate
□ Default Price List: Standard Selling Price
□ Default Territory: Head Office/Regional
□ Default Warehouse: Main Warehouse
□ Default Sales Person: (Optional)

Document Behavior:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Auto-create Delivery Note from Sales Order: Yes/No
□ Auto-create Invoice from Delivery: Yes/No
□ Require Customer PO Number: Yes/No
□ Allow Multiple Sales Orders against Quotation: Yes/No
□ Allow Sales Order Creation without Quotation: Yes/No

Pricing & Discount:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Apply Pricing Rule Automatically: Yes
□ Allow User to Edit Rate: Yes (with permissions)
□ Allow Discount on Item Level: Yes
□ Allow Discount on Invoice Level: Yes
□ Maximum Discount % (without approval): 10%
□ Validate Price List on Transactions: Yes

Inventory Integration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Check Stock Availability on Sales Order: Yes
□ Reserve Stock on Sales Order: Yes
□ Allow Backorders: Yes/No
□ Show Stock Balance in Quotation: Yes
□ Auto Reserve Stock on Quotation: No

Credit Control:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Enable Credit Limit: Yes
□ Check Credit Limit at: Sales Order / Delivery / Invoice
□ Credit Limit Action: Warn / Stop / Ignore
□ Allow Sales Order above Credit Limit with Approval: Yes

Financial Integration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Default Revenue Account: 4000 - Sales Revenue
□ Default Receivable Account: 1200 - Accounts Receivable
□ Default Tax Account: 2300 - VAT Payable
□ Default Cost Center: Sales Department
□ Post Accounting Entry on: Invoice Submit

Email & Notifications:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Send Email on Quotation Submit: Yes
□ Send Email on Sales Order Confirmation: Yes
□ Send Email on Invoice Submit: Yes
□ Send Email on Payment Receipt: Yes
□ Default Email Template: Standard Sales Template

Commission:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Enable Sales Commission: Yes
□ Calculate Commission on: Sales Order / Invoice / Payment
□ Commission Type: Fixed % / Tiered / Product-based
□ Default Commission Rate: 5%
```

### Step 2: Document Numbering Series

```markdown
SALES DOCUMENT NUMBERING

Quotation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: QTN-{YYYY}-{#####}
Example: QTN-2025-00001
Prefix Options:
  - By Fiscal Year: QTN-FY25-
  - By Company: QTN-KE- / QTN-UG-
  - By Territory: QTN-NAI- / QTN-MBA-

Sales Order:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: SO-{YYYY}-{#####}
Example: SO-2025-00001
Auto-increment: Yes
Starting Number: 1

Delivery Note:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: DN-{YYYY}-{#####}
Example: DN-2025-00001
Alternative: DEL-{YYYY}-{#####}

Sales Invoice:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: INV-{YYYY}-{#####}
Example: INV-2025-00001
Tax Invoice: TAX-INV-{YYYY}-{#####}
Proforma: PRO-{YYYY}-{#####}

Sales Return:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: RET-{YYYY}-{#####}
Example: RET-2025-00001
Credit Note: CN-{YYYY}-{#####}

Payment Entry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Series: PAY-{YYYY}-{#####}
Example: PAY-2025-00001
Receipt: RCP-{YYYY}-{#####}

CONFIGURATION:

For Each Document Type:
┌─────────────────────────────────────────┐
│ Document: Sales Order                   │
│ Series Format: SO-{YYYY}-{#####}       │
│ Current Number: 1                       │
│ Padding: 5 digits                       │
│ Reset Period: Yearly / Never            │
│ Active: Yes                             │
│ Set as Default: Yes                     │
└─────────────────────────────────────────┘

Multiple Series Example:
- Regular Sales: SO-2025-xxxxx
- Export Sales: EXP-2025-xxxxx
- Internal Sales: INT-2025-xxxxx
User selects series when creating document
```

### Step 3: Terms & Conditions Templates

```markdown
STANDARD TERMS & CONDITIONS

1. Payment Terms Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PAYMENT TERMS:
Payment is due within [X] days from invoice date.
Late payments will incur interest at [Y]% per month.
Accepted payment methods: Bank transfer, Cash, Check.

Bank Details:
Bank: [Bank Name]
Account: [Account Number]
Branch: [Branch Name]
SWIFT: [SWIFT Code]

2. Delivery Terms Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DELIVERY TERMS:
Delivery within [X] business days from order confirmation.
Delivery to [Address] during business hours (8 AM - 5 PM).
Customer responsible for offloading (unless agreed otherwise).
Risk transfers to customer upon delivery and signature.

3. Warranty Terms Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WARRANTY:
Products covered by [X]-month manufacturer warranty.
Warranty covers manufacturing defects only.
Damage from misuse, accidents, or unauthorized repairs void warranty.
Customer must report defects within [X] days of discovery.

4. Return Policy Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
RETURNS:
Returns accepted within [X] days of delivery.
Items must be unused, in original packaging.
Return shipping costs borne by [Customer/Seller].
Restocking fee of [X]% may apply.
Custom orders non-returnable.

5. Quotation Validity Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
VALIDITY:
This quotation is valid for [X] days from date of issue.
Prices subject to change after validity period.
Availability of items subject to prior sale.
Quotation does not constitute a binding offer until order confirmation.

CONFIGURATION:

Terms & Conditions Master:
┌─────────────────────────────────────────┐
│ Title: Standard Payment Terms           │
│ Category: Payment                       │
│ Content: [Full text as above]           │
│ Apply to: Quotation, Sales Order,       │
│           Sales Invoice                 │
│ Default: Yes/No                         │
│ Require Acceptance: Yes/No              │
└─────────────────────────────────────────┘

Multiple templates can be maintained and selected per document.
```

### Step 4: Email Templates

```markdown
EMAIL TEMPLATES CONFIGURATION

1. Quotation Email Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Quotation {quotation_number} from {company_name}

Dear {customer_name},

Thank you for your interest in our products/services.

Please find attached quotation {quotation_number} as requested.

Quotation Summary:
- Total Amount: {currency} {total_amount}
- Valid Until: {valid_till_date}
- Payment Terms: {payment_terms}

Should you have any questions or require clarification, please don't 
hesitate to contact me.

We look forward to serving you.

Best regards,
{sales_person_name}
{sales_person_email}
{company_name}

Attachment: Quotation_{quotation_number}.pdf

2. Sales Order Confirmation Email
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Order Confirmation - {sales_order_number}

Dear {customer_name},

Thank you for your order!

We are pleased to confirm your order {sales_order_number}.

Order Details:
- Order Number: {sales_order_number}
- Order Date: {order_date}
- Total Amount: {currency} {total_amount}
- Expected Delivery: {delivery_date}

Your order is being processed and you will receive a delivery 
notification once shipped.

Track your order: [tracking_link]

Thank you for your business!

Best regards,
{company_name}

3. Delivery Notification Email
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Your order {sales_order_number} has been delivered

Dear {customer_name},

Your order {sales_order_number} was successfully delivered on 
{delivery_date}.

Delivery Note: {delivery_note_number}
Delivered To: {delivery_address}
Received By: {receiver_name}

Please confirm receipt and inspect items for any damage.
Report any issues within 24 hours.

Thank you for choosing {company_name}!

Best regards,
{company_name}

4. Invoice Email Template
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Invoice {invoice_number} - {company_name}

Dear {customer_name},

Please find attached invoice {invoice_number} for recent delivery.

Invoice Summary:
- Invoice Number: {invoice_number}
- Invoice Date: {invoice_date}
- Due Date: {due_date}
- Amount Due: {currency} {outstanding_amount}

Payment Instructions:
{payment_instructions}

Pay online: [payment_link]

For any queries regarding this invoice, please contact our 
accounts team at {accounts_email}.

Thank you for your business!

Best regards,
{company_name}

Attachment: Invoice_{invoice_number}.pdf

5. Payment Reminder Email
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Payment Reminder - Invoice {invoice_number}

Dear {customer_name},

This is a friendly reminder that invoice {invoice_number} is due 
for payment on {due_date}.

Invoice Details:
- Invoice Number: {invoice_number}
- Invoice Date: {invoice_date}
- Due Date: {due_date}
- Amount Due: {currency} {outstanding_amount}
- Days Overdue: {days_overdue}

If you have already made payment, please disregard this reminder.
Otherwise, please arrange payment at your earliest convenience.

For payment arrangements or queries, contact: {accounts_email}

Thank you for your prompt attention.

Best regards,
{company_name}

6. Payment Received Email
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Payment Received - Thank You!

Dear {customer_name},

We confirm receipt of your payment for invoice {invoice_number}.

Payment Details:
- Amount Received: {currency} {amount_paid}
- Payment Date: {payment_date}
- Payment Method: {payment_method}
- Reference: {payment_reference}

Invoice {invoice_number} is now marked as PAID.

Outstanding Balance: {currency} {outstanding_balance}

Thank you for your payment!

Best regards,
{company_name}

CONFIGURATION:

Email Template Master:
┌─────────────────────────────────────────┐
│ Template Name: Quotation Submission     │
│ Document Type: Quotation                │
│ Subject: [Template with variables]      │
│ Body: [HTML/Plain text template]        │
│ Attachments: PDF, Additional Docs       │
│ Send To: Customer Email                 │
│ CC: Sales Person, Sales Manager         │
│ Trigger: On Submit                      │
│ Active: Yes                             │
└─────────────────────────────────────────┘

Available Variables:
{customer_name}, {customer_email}, {company_name},
{document_number}, {total_amount}, {date}, 
{sales_person_name}, {due_date}, etc.
```

### Step 5: Print Format Configuration

```markdown
PRINT FORMATS & BRANDING

Quotation Print Format:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ [Company Logo]          QUOTATION          │
│                                            │
│ Company Name                    QTN No:   │
│ Address Line 1                  Date:     │
│ Address Line 2                  Valid:    │
│ Phone: xxx  Email: xxx                    │
├────────────────────────────────────────────┤
│ BILL TO:                                  │
│ Customer Name                              │
│ Customer Address                           │
│ Contact Person: xxx                        │
│ Phone: xxx  Email: xxx                    │
├────────────────────────────────────────────┤
│ Item  Description    Qty  Rate    Amount  │
│ ────  ───────────    ───  ────    ──────  │
│ 1.    Product A      10   1,000   10,000  │
│ 2.    Product B      5    2,000   10,000  │
│                                            │
│                      Subtotal:    20,000  │
│                      Discount:    (1,000) │
│                      VAT (16%):    3,040  │
│                      TOTAL:       22,040  │
├────────────────────────────────────────────┤
│ TERMS & CONDITIONS:                        │
│ [Standard terms]                           │
├────────────────────────────────────────────┤
│ Prepared by: [Sales Person]                │
│ Signature: _______________                │
└────────────────────────────────────────────┘

Sales Invoice Format:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ [Company Logo]      TAX INVOICE            │
│                                            │
│ Company Name             Invoice No:      │
│ PIN: xxx                 Date:            │
│ Address                  Due Date:        │
│                          Sales Order:     │
├────────────────────────────────────────────┤
│ CUSTOMER DETAILS:                          │
│ Name: xxx              PIN: xxx            │
│ Address: xxx                               │
│                                            │
├────────────────────────────────────────────┤
│ Item  Description    Qty  Rate    Amount  │
│ ────  ───────────    ───  ────    ──────  │
│ [Line items]                               │
│                                            │
│                      Subtotal:    xxxxx   │
│                      VAT (16%):   xxxxx   │
│                      TOTAL:       xxxxx   │
├────────────────────────────────────────────┤
│ PAYMENT DETAILS:                           │
│ Bank: xxx  Account: xxx                    │
│ Branch: xxx  Swift: xxx                    │
├────────────────────────────────────────────┤
│ Amount in Words: [Amount in words]         │
│                                            │
│ This is a computer-generated invoice       │
└────────────────────────────────────────────┘

Delivery Note Format:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ [Company Logo]     DELIVERY NOTE           │
│                                            │
│ DN No: xxx           Date: xxx             │
│ Sales Order: xxx                           │
│                                            │
│ DELIVER TO:                                │
│ Customer: xxx                              │
│ Address: xxx                               │
│ Contact: xxx                               │
│                                            │
│ Item  Description         Qty  Delivered   │
│ ────  ──────────────      ───  ─────────   │
│ [Line items with checkboxes]               │
│                                            │
│ Special Instructions:                      │
│ [Any delivery notes]                       │
│                                            │
│ Delivered By: _______________             │
│ Signature: _______________                │
│ Date/Time: _______________                │
│                                            │
│ Received By: _______________              │
│ Signature: _______________                │
│ Date/Time: _______________                │
│ Condition: □ Good  □ Damaged              │
└────────────────────────────────────────────┘

CONFIGURATION:

Print Format Settings:
┌─────────────────────────────────────────┐
│ Document Type: Sales Invoice            │
│ Format Name: Standard Tax Invoice       │
│ Page Size: A4                           │
│ Orientation: Portrait                   │
│ Show Company Letterhead: Yes            │
│ Show Watermark: Draft/Paid              │
│ Show Terms & Conditions: Yes            │
│ Show Payment Instructions: Yes          │
│ Language: English / Swahili             │
│ Barcode/QR Code: Yes (for payment)     │
│ Default: Yes                            │
└─────────────────────────────────────────┘

Letterhead Configuration:
- Upload company logo
- Define header content
- Define footer content
- Set margins and spacing
- Configure colors and fonts
```

---