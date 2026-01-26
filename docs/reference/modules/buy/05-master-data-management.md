[<-- Back to Index](README.md)

## Master Data Management

### Supplier Master Data

```markdown
SUPPLIER INFORMATION STRUCTURE

Basic Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Supplier ID: SUP-00123 (auto-generated)
Supplier Name: Ace Steel Suppliers Ltd
Trading Name: Ace Steel
Type: Company
Category: Raw Material Supplier
Group: Strategic Supplier

Tax & Registration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
KRA PIN: P051234567X
VAT Registered: Yes
VAT Number: 0512345678
Business Registration: C.123456
Incorporation Date: 2015-03-15

Contact Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Primary Contact:
  Name: John Mwangi
  Title: Sales Manager
  Email: john.mwangi@acesteel.co.ke
  Phone: +254 712 345 678
  Mobile: +254 722 345 678

Accounts Contact:
  Name: Mary Njeri
  Title: Accounts Manager
  Email: accounts@acesteel.co.ke
  Phone: +254 712 345 679

Physical Address:
  Street: Lunga Lunga Road, Plot 45
  Area: Industrial Area
  City: Nairobi
  Country: Kenya
  Postal: P.O. Box 12345-00100

Billing Address: Same as Physical
Delivery Terms: Ex-works

Financial Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Currency: KES
Payment Terms: Net 30 Days
Credit Limit: 5,000,000 KES
Credit Days: 30
Early Payment Discount: 2/10 Net 30

Bank Details:
  Bank Name: Equity Bank Kenya
  Branch: Industrial Area Branch
  Account Number: 0123456789
  Account Name: Ace Steel Suppliers Ltd
  SWIFT Code: EQBLKENA
  Bank Code: 68
  Branch Code: 001

Tax Withholding:
  Subject to WHT: No
  WHT Rate: 0%
  Tax Category: Standard VAT

Supplier Attributes:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status: Active
Rating: A (Excellent)
Preferred: Yes
Certified: ISO 9001:2015
Lead Time: 7 days
Minimum Order: 50,000 KES
Payment Mode: Bank Transfer

Business Information:
  Years in Business: 10
  Annual Turnover: 500M KES
  Number of Employees: 50
  Ownership: Local

Certifications:
  ☑ ISO 9001:2015 (Quality Management)
  ☑ ISO 14001 (Environmental)
  ☑ KEBS Certified
  □ Fair Trade

Documents Attached:
  ✓ Certificate of Incorporation
  ✓ KRA PIN Certificate
  ✓ Tax Compliance Certificate
  ✓ Bank Statement (last 6 months)
  ✓ ISO Certificates
  ✓ Product Catalogs
  ✓ Company Profile
```

### Item Master Data (Procurement View)

```markdown
ITEM MASTER CONFIGURATION

Item Code: RM-STL-001
Item Name: Cold Rolled Steel Sheet
Item Group: Raw Materials > Steel Products
UOM: Kilogram (Kg)

Procurement Details:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Default Supplier: Ace Steel Suppliers Ltd (SUP-00123)
Backup Supplier: Steel Corp Ltd (SUP-00456)

Purchasing UOM: Ton (1000 Kg)
Conversion: 1 Ton = 1000 Kg

Minimum Order Quantity: 1 Ton
Economic Order Quantity: 5 Tons
Lead Time: 7 days

Reorder Level: 2 Tons
Reorder Quantity: 5 Tons
Maximum Stock Level: 10 Tons
Safety Stock: 1 Ton

Valuation:
  Method: Moving Average
  Last Purchase Price: 85,000 KES/Ton
  Standard Cost: 82,000 KES/Ton
  Current Stock Value: 410,000 KES

Tax & Accounting:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Tax Category: Standard VAT (16%)
Expense Account: 5100 - Raw Material Purchases
Inventory Account: 1300 - Raw Materials Stock

Quality Specifications:
  Grade: CR1 Commercial Quality
  Thickness: 1.0mm
  Width: 1000mm
  Finish: Matte
  Standards: ASTM A1008

Supplier Pricing:
┌──────────────────────┬──────────┬────────────┬───────────┐
│ Supplier             │ Price/Ton│ Min Order  │ Lead Time │
├──────────────────────┼──────────┼────────────┼───────────┤
│ Ace Steel (Primary)  │ 85,000   │ 1 Ton      │ 7 days    │
│ Steel Corp (Backup)  │ 87,000   │ 2 Tons     │ 10 days   │
│ Global Steel (Import)│ 82,000   │ 10 Tons    │ 45 days   │
└──────────────────────┴──────────┴────────────┴───────────┘
```

### Units of Measure (UOM)

```markdown
UOM CONFIGURATION

Base Units:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Length:
  - Meter (m)
  - Centimeter (cm): 1m = 100cm
  - Millimeter (mm): 1m = 1000mm
  - Kilometer (km): 1km = 1000m

Weight:
  - Gram (g)
  - Kilogram (Kg): 1Kg = 1000g
  - Ton: 1Ton = 1000Kg
  - Milligram (mg): 1g = 1000mg

Volume:
  - Liter (L)
  - Milliliter (ml): 1L = 1000ml
  - Cubic Meter (m³): 1m³ = 1000L

Quantity:
  - Piece (Pcs)
  - Dozen: 1 Dozen = 12 Pcs
  - Box: Variable conversion per item
  - Carton: Variable conversion per item
  - Pallet: Variable conversion per item

Area:
  - Square Meter (m²)
  - Square Foot (ft²)

UOM Conversions Example:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Item: Bolts M8x50
Base UOM: Piece

Conversions:
  Box = 100 Pieces
  Carton = 10 Boxes = 1000 Pieces

Purchase: By Carton
Stock: By Piece
Issue: By Piece

Purchase Order:
  Quantity: 5 Cartons
  = 5 × 1000 = 5000 Pieces (stock updated)
```

### Tax Rate Master

```markdown
TAX CONFIGURATION

VAT Rates:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Standard VAT:
  Code: VAT-16
  Rate: 16%
  GL Account: 1420 - VAT Input
  Description: Standard VAT on taxable supplies

Zero-Rated VAT:
  Code: VAT-0
  Rate: 0%
  GL Account: 1420 - VAT Input
  Description: Zero-rated supplies (recoverable)

Exempt:
  Code: VAT-EXEMPT
  Rate: 0%
  GL Account: None
  Description: Exempt supplies (non-recoverable)

Withholding Tax:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Professional Fees:
  Code: WHT-PROF
  Rate: 5%
  GL Account: 2150 - WHT Payable
  Applied on: Gross amount
  Deducted from: Supplier payment

Rent:
  Code: WHT-RENT
  Rate: 10%
  GL Account: 2150 - WHT Payable

Contractual Fees:
  Code: WHT-CONTRACT
  Rate: 3%
  GL Account: 2150 - WHT Payable

Import Duty:
  Code: IMPORT-DUTY
  Rate: Variable (by HS Code)
  GL Account: 5800 - Import Duty Expense
  Description: Customs duty on imports

Excise Duty:
  Code: EXCISE
  Rate: Variable (by product)
  GL Account: 5810 - Excise Duty Expense
```

### Warehouse Master

```markdown
WAREHOUSE CONFIGURATION

Warehouse: Main Warehouse - Nairobi
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Code: WH-NBI-01
Type: Central Distribution Center
Status: Active

Location Details:
  Address: Industrial Area, Nairobi
  GPS: -1.3187, 36.8510
  Area: 5,000 m²

Contact:
  Manager: James Omondi
  Email: warehouse.nbi@awo.co.ke
  Phone: +254 720 123 456

Operating Hours:
  Monday - Friday: 7:00 AM - 6:00 PM
  Saturday: 8:00 AM - 2:00 PM
  Sunday: Closed

Storage Configuration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Zones:
  Zone A - Raw Materials:
    Capacity: 500 Tons
    Current: 320 Tons (64%)

  Zone B - Finished Goods:
    Capacity: 1000 Units
    Current: 650 Units (65%)

  Zone C - MRO Items:
    Capacity: 2000 SKUs
    Current: 1450 SKUs (73%)

  Zone D - Hazardous:
    Capacity: 50 Tons
    Current: 15 Tons (30%)
    Special: Fire suppression, ventilation

Receiving Areas:
  Bay 1: Raw Materials (Bulk)
  Bay 2: Packaged Goods
  Bay 3: Equipment & Machinery
  Bay 4: Returns & QC

Accounting Integration:
  Stock Account: 1300 - Inventory
  GRN Clearing: 2180 - GRN Clearing
  Stock Variance: 5900 - Stock Variance

Settings:
  ☑ Allow negative stock: No
  ☑ Batch tracking: Yes
  ☑ Serial number tracking: Yes (for equipment)
  ☑ Quality inspection: Mandatory
  ☑ Barcode scanning: Yes
```

### Cost Center Master

```markdown
COST CENTER CONFIGURATION

Cost Center: Production Department
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Code: CC-PROD-01
Name: Production Department - Nairobi
Type: Production
Status: Active
Parent: Manufacturing Division

Responsible Person:
  Name: Peter Kimani
  Title: Production Manager
  Email: peter.kimani@awo.co.ke

Budget Allocation (Annual):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Fiscal Year: 2025

┌──────────────────────┬─────────────┬─────────────┬───────────┐
│ Category             │ Budget      │ Committed   │ Available │
├──────────────────────┼─────────────┼─────────────┼───────────┤
│ Raw Materials        │ 50,000,000  │ 12,500,000  │37,500,000 │
│ MRO Supplies         │  5,000,000  │  1,200,000  │ 3,800,000 │
│ Equipment Maintenance│  2,000,000  │    450,000  │ 1,550,000 │
│ Utilities            │  3,000,000  │    750,000  │ 2,250,000 │
│ Capital Equipment    │ 10,000,000  │  5,000,000  │ 5,000,000 │
│                      │             │             │           │
│ TOTAL                │ 70,000,000  │ 19,900,000  │50,100,000 │
└──────────────────────┴─────────────┴─────────────┴───────────┘

Budget Consumption: 28.4%
Status: On Track

Cost Allocation:
  All purchases for this cost center
  Tagged in requisitions
  Tracked in GL reporting
```

### Budget Master

```markdown
BUDGET CONFIGURATION

Budget: 2025 Procurement Budget
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Period: Jan 1, 2025 - Dec 31, 2025
Status: Active
Currency: KES

Budget Breakdown:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
By Department:
┌──────────────────┬──────────────┬──────────────┬──────────┐
│ Department       │ Annual Budget│ Q1 Spent     │ Balance  │
├──────────────────┼──────────────┼──────────────┼──────────┤
│ Production       │ 70,000,000   │ 19,900,000   │50,100,000│
│ Maintenance      │ 15,000,000   │  3,200,000   │11,800,000│
│ Administration   │  5,000,000   │  1,100,000   │ 3,900,000│
│ IT               │  8,000,000   │  2,500,000   │ 5,500,000│
│ Marketing        │  3,000,000   │    600,000   │ 2,400,000│
│                  │              │              │          │
│ TOTAL            │101,000,000   │ 27,300,000   │73,700,000│
└──────────────────┴──────────────┴──────────────┴──────────┘

Budget Controls:
  ☑ Prevent over-budget purchases (hard stop)
  ☑ Alert at 80% consumption
  ☑ Require approval for budget transfer
  ☑ Monthly variance reporting

Monthly Allocation:
  Equal distribution with quarterly reviews
  Variance tracking against actuals
```

### Procurement Category Master

```markdown
PROCUREMENT CATEGORIES

Category Hierarchy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Direct Materials
   ├─ Raw Materials
   │  ├─ Metals
   │  ├─ Plastics
   │  └─ Chemicals
   ├─ Components
   │  ├─ Electrical
   │  ├─ Mechanical
   │  └─ Electronic
   └─ Packaging Materials

2. Indirect Materials
   ├─ MRO (Maintenance, Repair, Operations)
   │  ├─ Spare Parts
   │  ├─ Tools
   │  ├─ Lubricants
   │  └─ Safety Equipment
   ├─ Office Supplies
   └─ Cleaning Supplies

3. Services
   ├─ Professional Services
   │  ├─ Consultancy
   │  ├─ Legal
   │  └─ Audit
   ├─ Maintenance Services
   └─ Logistics Services

4. Capital Expenditure
   ├─ Machinery
   ├─ Vehicles
   ├─ IT Equipment
   └─ Furniture & Fixtures

5. Utilities
   ├─ Electricity
   ├─ Water
   └─ Telecommunications

Category Attributes:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Category: Raw Materials > Metals

GL Accounts:
  Expense: 5100 - Raw Material Purchases
  Inventory: 1300 - Raw Materials Stock
  COGS: 5000 - Cost of Goods Sold

Approval Required: Yes
Minimum Quotations: 3 (if > 500K)
Preferred Suppliers: Ace Steel, Steel Corp
Lead Time (Average): 7-14 days
Quality Inspection: Mandatory
```

---

**Next:** [Supplier Management](./06-supplier-management.md)
