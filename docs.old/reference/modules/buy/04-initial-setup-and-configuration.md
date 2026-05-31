[<-- Back to Index](README.md)

## Initial Setup & Configuration

### Tenant-Level Procurement Configuration

```markdown
PROCUREMENT SETTINGS BY TENANT

Basic Information:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Tenant: AWO Manufacturing Ltd
Company: AWO Mfg - Nairobi Plant
Base Currency: KES
Fiscal Year: Jan 1 - Dec 31
Default Warehouse: Main Warehouse - Nairobi

Procurement Preferences:
  ☑ Enable Purchase Requisitions
  ☑ Require approval for all PRs
  ☑ Enable RFQ process
  ☑ Automatic 3-way matching
  ☑ Budget validation on requisitions
  □ Allow negative stock (disabled for procurement)
  ☑ Track supplier performance
  ☑ Generate GRN automatically

Purchase Order Settings:
  Default Terms: Net 30 Days
  Auto-numbering: Yes
  PO Validity: 30 days
  Amendment requires approval: Yes
  Minimum suppliers for RFQ: 3
```

### Document Numbering Series

```markdown
NUMBERING CONFIGURATION

Purchase Requisition:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Prefix: PR-
Format: PR-YYYY-####
Example: PR-2025-0001
Auto-increment: Yes
Reset: Yearly

Request for Quotation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Prefix: RFQ-
Format: RFQ-YYYY-####
Example: RFQ-2025-0001

Purchase Order:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Prefix: PO-
Format: PO-YYYY-#####
Example: PO-2025-00001
Branch-specific:
  PO-NBI-2025-00001 (Nairobi)
  PO-MBA-2025-00001 (Mombasa)

Goods Receipt Note:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Prefix: GRN-
Format: GRN-YYYY-#####
Example: GRN-2025-00001

Purchase Invoice:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Prefix: PINV-
Format: PINV-YYYY-#####
Example: PINV-2025-00001

Payment Entry:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Prefix: PAY-
Format: PAY-YYYY-#####
Example: PAY-2025-00001

Purchase Return:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Prefix: PR-RET-
Format: PR-RET-YYYY-####
Example: PR-RET-2025-0001

Debit Note:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Prefix: DN-
Format: DN-YYYY-####
Example: DN-2025-0001
```

### Approval Matrix Configuration

```markdown
PROCUREMENT APPROVAL HIERARCHY

Purchase Requisition Approval:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌─────────────────┬──────────────────┬─────────────────┐
│ Amount (KES)    │ Approver Level 1 │ Approver Level 2│
├─────────────────┼──────────────────┼─────────────────┤
│ 0 - 50,000      │ Department Head  │ -               │
│ 50,001 - 200,000│ Department Head  │ Procurement Mgr │
│ 200,001 - 1M    │ Department Head  │ Director        │
│ > 1,000,000     │ Director         │ CFO             │
│ > 5,000,000     │ CFO              │ CEO             │
└─────────────────┴──────────────────┴─────────────────┘

Purchase Order Approval:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌─────────────────┬──────────────────┬─────────────────┐
│ Amount (KES)    │ Approver Level 1 │ Approver Level 2│
├─────────────────┼──────────────────┼─────────────────┤
│ 0 - 100,000     │ Procurement Off  │ -               │
│ 100,001 - 500K  │ Procurement Mgr  │ -               │
│ 500,001 - 2M    │ Procurement Mgr  │ Finance Mgr     │
│ 2,000,001 - 5M  │ Procurement Dir  │ CFO             │
│ > 5,000,000     │ CFO              │ CEO             │
└─────────────────┴──────────────────┴─────────────────┘

Invoice Approval (If variance):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Variance Type         → Approver
─────────────────────────────────────────────
Price ± 2%            → Auto-approve
Price 2-5%            → Procurement Manager
Price > 5%            → Director
Quantity mismatch     → Warehouse + Procurement
Tax variance          → Finance Manager

Payment Approval:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌─────────────────┬──────────────────┬─────────────────┐
│ Amount (KES)    │ Approver Level 1 │ Approver Level 2│
├─────────────────┼──────────────────┼─────────────────┤
│ 0 - 500,000     │ AP Officer       │ -               │
│ 500,001 - 2M    │ Finance Manager  │ -               │
│ > 2,000,000     │ Finance Manager  │ CFO             │
└─────────────────┴──────────────────┴─────────────────┘

Special Approvals:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Emergency Purchase:
  Approver: Department Director + CFO
  Post-documentation required within 24 hours

Capital Equipment (> 1M):
  Approver: CFO + CEO + Board (if > 10M)
  Requires: Business case, ROI analysis

New Supplier:
  Approver: Procurement Director
  Requires: Supplier qualification form
```

### Supplier Categories

```markdown
SUPPLIER CLASSIFICATION

By Business Type:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Raw Material Suppliers
   - Steel suppliers
   - Chemical suppliers
   - Packaging material providers

2. Finished Goods Suppliers
   - Resale inventory
   - Trading goods

3. Service Providers
   - Maintenance services
   - Professional services
   - Logistics providers

4. Utilities
   - Electricity
   - Water
   - Internet/Telecommunications

5. Asset Suppliers
   - Machinery vendors
   - IT equipment suppliers
   - Vehicle dealers

By Strategic Importance:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategic:
  - Long-term partnership
  - High spend volume
  - Critical materials
  - Limited alternatives
  → Quarterly reviews, annual contracts

Preferred:
  - Good performance
  - Competitive pricing
  - Reliable delivery
  → Annual reviews

Approved:
  - Qualified suppliers
  - Standard items
  - Multiple alternatives
  → Bi-annual reviews

Probation:
  - New suppliers
  - Recovery from issues
  - 6-month evaluation period
  → Monthly reviews

By Region:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Local (Nairobi/Kenya):
  Payment: KES
  Lead time: 1-7 days
  Tax: 16% VAT

Regional (East Africa):
  Payment: KES/USD
  Lead time: 7-21 days
  Tax: EAC protocols

International:
  Payment: USD/EUR
  Lead time: 30-90 days
  Tax: Import duty + VAT
```

### Item Categories for Procurement

```markdown
PROCUREMENT ITEM CLASSIFICATION

1. Raw Materials
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Examples:
  - Steel coils
  - Plastic pellets
  - Chemicals

Procurement Method: Long-term contracts, volume pricing
Reorder: Automatic based on production plan

2. Components & Parts
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Examples:
  - Motors
  - Bearings
  - Electronic components

Procurement Method: Blanket POs with call-offs
Reorder: Safety stock levels

3. MRO (Maintenance, Repair, Operations)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Examples:
  - Spare parts
  - Tools
  - Cleaning supplies
  - PPE (Personal Protective Equipment)

Procurement Method: Standing orders, local suppliers
Reorder: Min-max levels

4. Office Supplies
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Examples:
  - Stationery
  - Printer supplies
  - Furniture

Procurement Method: Requisition-based
Reorder: As needed

5. Capital Equipment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Examples:
  - Machinery
  - Vehicles
  - IT infrastructure

Procurement Method: RFQ/Tender process
Reorder: Budget-approved projects

6. Services
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Examples:
  - Consultancy
  - Maintenance contracts
  - Security services

Procurement Method: Service contracts
Reorder: Annual renewals
```

### Payment Terms Configuration

```markdown
STANDARD PAYMENT TERMS

Terms Configuration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Cash on Delivery (COD)
   - Payment upon receipt
   - Use: New suppliers, small orders

2. Net 7 Days
   - Payment due 7 days from invoice date
   - Use: Perishables, urgent supplies

3. Net 15 Days
   - Payment due 15 days from invoice date
   - Use: Local suppliers, MRO items

4. Net 30 Days (Most Common)
   - Payment due 30 days from invoice date
   - Use: Standard procurement

5. Net 45 Days
   - Payment due 45 days from invoice date
   - Use: Large suppliers, negotiated terms

6. Net 60 Days
   - Payment due 60 days from invoice date
   - Use: Strategic suppliers, volume purchases

7. Net 90 Days
   - Payment due 90 days from invoice date
   - Use: Capital equipment, large projects

8. Installment Terms
   - Split payments over milestones
   - Use: Projects, phased deliveries
   Example: 30% advance, 40% on delivery, 30% after 30 days

9. Early Payment Discounts
   - 2/10 Net 30: 2% discount if paid in 10 days
   - 1/7 Net 30: 1% discount if paid in 7 days

Supplier-Specific Terms:
  Supplier A: Net 30, 2/10
  Supplier B: Net 45
  Supplier C: 50% advance, 50% on delivery
```

### Tax Templates

```markdown
TAX CONFIGURATION

Kenya VAT Setup:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Standard VAT:
  Rate: 16%
  Account: VAT Input (1420)
  Applies to: Most goods and services

Zero-Rated:
  Rate: 0%
  Account: VAT Input (1420)
  Applies to: Exports, agricultural inputs, medical supplies

Exempt:
  Rate: 0%
  No VAT recovery
  Applies to: Financial services, educational services

Withholding Tax on Purchases:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WHT on Professional Fees:
  Rate: 5%
  Account: WHT Payable (2150)

WHT on Rent:
  Rate: 10%

WHT on Contractual Fees:
  Rate: 3%

Import Duty:
  Varies by HS Code
  Account: Import Duty Expense
```

### Warehouse Configuration

```markdown
WAREHOUSE SETUP FOR PROCUREMENT

Main Warehouse - Nairobi:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Code: WH-NBI-01
Type: Central Distribution
Address: Industrial Area, Nairobi
Contact: warehouse@awo.co.ke

Receiving Areas:
  - Raw Materials Bay
  - Finished Goods Bay
  - Returns Area
  - Quarantine/QC Area

Storage Zones:
  - Zone A: Fast-moving items
  - Zone B: Medium-moving items
  - Zone C: Slow-moving items
  - Zone D: Hazardous materials

Branch Warehouse - Mombasa:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Code: WH-MBA-01
Type: Regional Distribution
Replenished from: WH-NBI-01
```

### Procurement Policies

```markdown
POLICY CONFIGURATION

1. Minimum Purchase Value:
   - No PO for purchases < 5,000 KES (petty cash)
   - Requisition mandatory for all > 5,000 KES

2. Multi-Supplier Policy:
   - Orders > 500,000 KES require 3 quotations
   - Exception: Sole-source with justification

3. Budget Control:
   - All requisitions checked against budget
   - Over-budget requires CFO approval
   - Commitment recorded on PR approval

4. Quality Standards:
   - ISO certification preferred
   - Quality inspection mandatory for > 100K
   - Supplier audit for strategic suppliers

5. Local Content Policy:
   - Preference for local suppliers (10% price advantage)
   - Support for local economy
   - Faster delivery and support

6. Ethical Procurement:
   - No child labor
   - Environmental compliance
   - Anti-corruption policy
   - Conflict of interest declaration
```

---

**Next:** [Master Data Management](./05-master-data-management.md)
