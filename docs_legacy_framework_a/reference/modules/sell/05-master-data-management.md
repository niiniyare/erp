> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## Master Data Management

### Customer Master

**Customer Record Structure:**

```markdown
CUSTOMER MASTER DATA

Basic Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customer ID: AUTO-GENERATED (CUST-00001)
Customer Name: ABC Manufacturing Ltd
Customer Type: Company / Individual
Customer Group: Corporate / Retail / Wholesale / Distributor
Territory: Nairobi / Mombasa / Kisumu / etc.
Industry: Manufacturing / Retail / Services / etc.
Customer Since: 2023-01-15
Status: Active / Inactive / Suspended

Contact Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Primary Contact Person: John Kamau
Title: Procurement Manager
Email: john.kamau@abcmanufacturing.com
Phone: +254-700-123-456
Mobile: +254-722-123-456
Website: www.abcmanufacturing.com

Address Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Billing Address:
  Address Line 1: Industrial Area
  Address Line 2: Nairobi
  City: Nairobi
  County: Nairobi
  Postal Code: 00100
  Country: Kenya

Shipping Address: □ Same as Billing
  [If different, separate fields]

Tax Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Tax ID (PIN): P000123456A
Tax Category: Taxable / Exempt / Zero-Rated
VAT Registration: Yes/No
Tax Exemption Certificate: [Upload if applicable]

Financial Settings:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Currency: KES (can support multi-currency)
Default Price List: Corporate Pricing
Payment Terms: Net 30 Days
Credit Limit: 5,000,000 KES
Credit Days: 30 days
Payment Method: Bank Transfer / Cash / Check

Credit Control:
  Check Credit Limit: Yes
  Credit Limit Override Allowed: With Approval
  Current Outstanding: 1,250,000 KES
  Available Credit: 3,750,000 KES

Banking Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Bank Name: Equity Bank
Branch: Industrial Area
Account Number: 0123456789
SWIFT Code: EQBLKENA

Sales Settings:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sales Person: Sarah Johnson
Sales Team: Enterprise Sales
Territory: Nairobi Corporate
Customer Category: Key Account / Regular / New

Loyalty & Preferences:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Loyalty Program: Gold Tier
Discount Percentage: 15% (approved)
Preferred Delivery Day: Thursday
Preferred Delivery Time: Morning (8-12)
Special Instructions: Requires delivery note in duplicate

Custom Fields (Configurable):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Industry Segment: Industrial Equipment
Business Size: Medium (50-200 employees)
Annual Revenue: 50-100M KES
Purchase Frequency: Monthly
Key Decision Maker: John Kamau

Attachments:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
□ Business Registration Certificate
□ Tax PIN Certificate
□ Bank Details Letter
□ Credit Application Form
□ Signed Terms & Conditions

Audit Trail:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Created By: admin@company.com
Created On: 2023-01-15 10:30:00
Last Modified By: sarah.johnson@company.com
Last Modified On: 2025-01-10 14:25:00
```

**Customer Categorization:**

```markdown
CUSTOMER GROUP STRUCTURE

By Business Type:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
├─ Retail
│  ├─ Walk-in Customers
│  ├─ Online Shoppers
│  └─ Small Businesses
├─ Wholesale
│  ├─ Distributors
│  ├─ Resellers
│  └─ Agents
├─ Corporate
│  ├─ Large Enterprises
│  ├─ SMEs
│  └─ Government
└─ Export
   ├─ East Africa
   ├─ Rest of Africa
   └─ International

By Territory:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
├─ Kenya
│  ├─ Nairobi
│  ├─ Mombasa
│  ├─ Kisumu
│  ├─ Nakuru
│  └─ Other Counties
├─ Uganda
├─ Tanzania
└─ Rwanda

By Industry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
├─ Manufacturing
├─ Retail & Distribution
├─ Construction
├─ Healthcare
├─ Education
├─ Hospitality
├─ Agriculture
└─ Services

By Value Tier:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
├─ Platinum (> 10M annually)
├─ Gold (5-10M annually)
├─ Silver (1-5M annually)
└─ Bronze (< 1M annually)

PRICING & DISCOUNT IMPACT:

Customer Group: Corporate
├─ Base Discount: 10%
├─ Volume Discount: Additional 5% (>100 units)
├─ Payment Discount: 2% if paid within 10 days
└─ Special Promotions: Eligible

Customer Group: Retail
├─ Base Discount: 0%
├─ Loyalty Discount: 3% (for repeat customers)
├─ Payment Discount: None
└─ Special Promotions: Limited
```

### Contact Persons (Multi-Contact Support)

```markdown
MULTIPLE CONTACTS PER CUSTOMER

Customer: ABC Manufacturing Ltd

Contact 1 (Primary):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Name: John Kamau
Designation: Procurement Manager
Department: Purchasing
Email: john.kamau@abc.com
Phone: +254-700-123-456
Mobile: +254-722-123-456
Is Primary Contact: Yes
Receives: Quotations, Orders, Invoices
Decision Maker: Yes

Contact 2 (Finance):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Name: Mary Wanjiku
Designation: Finance Manager
Department: Finance
Email: mary.wanjiku@abc.com
Phone: +254-700-234-567
Is Primary Contact: No
Receives: Invoices, Payment Receipts, Statements
Decision Maker: For payments

Contact 3 (Technical):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Name: David Omondi
Designation: Technical Manager
Department: Engineering
Email: david.omondi@abc.com
Phone: +254-700-345-678
Is Primary Contact: No
Receives: Technical Specs, Delivery Notes
Decision Maker: For specifications

USAGE IN DOCUMENTS:

Quotation:
  Attention: John Kamau (Procurement)
  CC: David Omondi (for technical review)

Invoice:
  Attention: Mary Wanjiku (Finance)
  CC: John Kamau (FYI)

Delivery Note:
  Contact: David Omondi (to receive goods)
```

### Customer Addresses (Multi-Address Support)

```markdown
MULTIPLE ADDRESSES PER CUSTOMER

Customer: ABC Manufacturing Ltd

Address 1 (Head Office - Billing):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Address Type: Billing
Address Name: Head Office
Address Line 1: Mombasa Road, Industrial Area
Address Line 2: Building 45, Floor 3
City: Nairobi
County: Nairobi County
Postal Code: 00100
Country: Kenya
Is Primary: Yes
Is Billing: Yes
Is Shipping: No

Address 2 (Factory - Shipping):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Address Type: Shipping
Address Name: Factory Warehouse
Address Line 1: Thika Road, Exit 14
Address Line 2: KM 25
City: Ruiru
County: Kiambu County
Postal Code: 00232
Country: Kenya
Is Primary: No
Is Billing: No
Is Shipping: Yes
Contact Person: David Omondi
Phone: +254-700-345-678
Delivery Instructions: Gate closes at 5 PM

Address 3 (Branch Office):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Address Type: Shipping
Address Name: Mombasa Branch
Address Line 1: Moi Avenue
City: Mombasa
County: Mombasa County
Postal Code: 80100
Country: Kenya
Is Primary: No
Is Billing: No
Is Shipping: Yes

USAGE IN DOCUMENTS:

Sales Order:
  Billing Address: Head Office
  Shipping Address: Factory Warehouse
  (User can select from saved addresses)

Delivery Note:
  Deliver To: Factory Warehouse
  Contact: David Omondi (+254-700-345-678)
  Instructions: "Gate closes at 5 PM"
```

### Customer Groups

```markdown
CUSTOMER GROUP CONFIGURATION

Group: Corporate Accounts
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Description: Large corporate customers
Default Price List: Corporate Pricing
Default Discount: 10%
Payment Terms: Net 30 Days
Credit Limit Range: 1M - 10M KES
Requires Approval For:
  - Discount above 15%
  - Credit limit increase
Special Features:
  - Dedicated account manager
  - Quarterly business reviews
  - Priority support

Group: Wholesale Distributors
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Description: Wholesale buyers and distributors
Default Price List: Wholesale Pricing
Default Discount: 20%
Payment Terms: Net 15 Days
Credit Limit Range: 500K - 5M KES
Special Features:
  - Volume discounts available
  - Flexible delivery schedules
  - Marketing support

Group: Retail Customers
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Description: Walk-in and regular retail customers
Default Price List: Standard Retail Price
Default Discount: 0%
Payment Terms: Cash / Immediate
Credit Limit: 0 (Cash only)
Special Features:
  - Loyalty program eligible
  - Seasonal promotions
  - Member discounts

Group: Export Customers
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Description: International customers
Default Price List: Export Pricing (USD)
Default Discount: 0%
Payment Terms: LC / Advance Payment
Credit Limit: Case by case
Special Features:
  - Multi-currency support
  - Export documentation
  - International shipping
  - Zero-rated VAT
```

---