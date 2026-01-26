[<-- Back to Index](README.md)

## Common Business Scenarios

### Emergency Purchases

```markdown
URGENT PROCUREMENT WORKFLOW

Scenario: Machine Breakdown - Urgent Spare Part Needed
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Date: 2025-04-18, 10:00 AM
Event: Production machine M-05 breaks down
Impact: Production line stopped
Required: Motor bearing (critical spare part)
Stock Status: Not in stock
Normal Lead Time: 7 days
Needed By: Today (production loss: 2M KES/day)

Emergency Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Step 1: Emergency PR (10:15 AM)
  PR No: PR-2025-00458 (URGENT flag)
  Item: Motor Bearing Model XYZ-2000
  Quantity: 1 Unit
  Required By: IMMEDIATE
  Requested By: Maintenance Manager
  Priority: EMERGENCY
  
  Approval: Fast-track (auto-approved for emergencies <100K)

Step 2: Sourcing (10:20 AM)
  Buyer Action: Phone calls to known suppliers
  
  Supplier A (Nairobi): Not in stock
  Supplier B (Mombasa): In stock, can courier today
    Price: 85,000 KES (vs normal 75,000)
    Delivery: Today by 3:00 PM
    Premium: 10,000 KES (13% over normal)
    
  Decision: Proceed with Supplier B
  Justification: Production loss > Premium cost

Step 3: Emergency PO (10:35 AM)
  PO No: PO-2025-00395 (EMERGENCY)
  Amount: 85,000 KES + VAT
  Delivery: Same day courier
  Approval: Operational Manager (expedited)
  
  Email to Supplier: Urgent - Please dispatch immediately

Step 4: Tracking (Ongoing)
  11:00 AM: Supplier confirms dispatch
  12:30 PM: Courier in transit (live tracking)
  02:45 PM: Part arrives at factory gate
  03:00 PM: GRN created, part issued to maintenance
  06:00 PM: Machine repaired, production resumed

Cost Analysis:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Premium Paid: 10,000 KES
Production Downtime Avoided: 8 hours = 670,000 KES loss
Net Benefit: 660,000 KES

Decision: Justified ✓

Post-Event Actions:
  1. Add item to critical spares inventory
  2. Set reorder level (Min: 1, Max: 2)
  3. Document for future reference
  4. Negotiate better emergency rates with supplier
```

### Capital Equipment Procurement

```markdown
CAPEX PURCHASE PROCESS

Scenario: New CNC Machine Acquisition
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Phase 1: Business Case & Approval (Week 1-2)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Capital Expenditure Request: CAPEX-2025-005
Requested By: Production Manager
Equipment: CNC Machining Center
Estimated Cost: 11M KES
Budget Year: 2025

Business Justification:
  Current Capacity: 1,000 units/month
  Demand Forecast: 1,800 units/month
  Capacity Gap: 800 units/month
  
  Financial Analysis:
    Equipment Cost: 11,000,000 KES
    Installation: 500,000 KES
    Training: 200,000 KES
    Total Investment: 11,700,000 KES
    
    Revenue Impact:
      Additional capacity: 800 units/month
      Selling price: 15,000 KES/unit
      Monthly revenue increase: 12,000,000 KES
      Annual revenue: 144,000,000 KES
      
    Payback Period: 11.7M / 144M = 0.97 months (~1 year)
    ROI (Year 1): 1,132%
    
Approval Chain:
  Production Manager → Engineering Manager → CFO → CEO → Board
  
  Result: ✓ APPROVED
  Budget Allocation: Released from CAPEX reserve

Phase 2: Technical Specification (Week 3)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Technical Team: Engineering + Production
Output: Detailed specifications (15 pages)
  - Machine specifications
  - Performance requirements
  - Quality standards
  - Service and warranty terms
  - Training requirements

Phase 3: Supplier Selection & RFQ (Week 4-6)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
RFQ Process: (See file 08-09 for details)
  - Shortlist 4 qualified suppliers
  - Issue RFQ with specifications
  - Site visits to supplier facilities
  - Technical presentations
  - Comparative evaluation
  
Winner: Precision Tech GmbH (Germany)
Price: 80,000 EUR (11.3M KES)
Lead Time: 12 weeks

Phase 4: Contract Negotiation (Week 7)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Contract Terms:
  Payment: 30% advance, 70% on delivery
  Incoterms: CIF Mombasa
  Warranty: 24 months
  Service: Annual maintenance contract available
  Training: 1 week on-site (included)
  Spares: Critical spares list provided
  
Phase 5: PO & Advance Payment (Week 8)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PO Created: PO-2025-00325
Advance Payment: 25,200 EUR (3.54M KES)
  Board approval required (>5M KES)
  Forex hedge arranged

Phase 6: Manufacture & Shipment (Week 9-20)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Weekly Progress Updates:
  Week 10: Manufacturing commenced
  Week 14: Factory acceptance test (FAT) video call
  Week 18: Shipping arranged
  Week 20: Arrival in Mombasa

Phase 7: Import Clearance (Week 21-22)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Customs Process: (See file 19 for details)
  Import duties, taxes: 3.9M KES
  Clearing and transport: 355K KES
  Total landed cost: 16.1M KES

Phase 8: Installation & Commissioning (Week 23-24)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Supplier Engineer Visit:
  Installation: 3 days
  Testing: 2 days
  Training: 5 days (3 operators, 2 maintenance staff)
  
Site Acceptance Test (SAT):
  ✓ Performance tests passed
  ✓ Quality outputs verified
  ✓ Operator training completed

Phase 9: Final Payment & Closure (Week 25)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Balance Payment: 58,800 EUR (8.29M KES)
GRN: Capitalized to Fixed Assets
Status: OPERATIONAL

Total Project Duration: 25 weeks (6 months)
Final Cost: 16.1M KES (within budget)
```

### Blanket Purchase Orders

```markdown
BLANKET PO MANAGEMENT

Scenario: Annual Office Supplies Contract
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Blanket PO Setup:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PO No: BPO-2025-001 (Blanket PO)
Supplier: Office Supplies Co Ltd
Type: Blanket Order
Duration: 12 months (Jan 1 - Dec 31, 2025)
Value: 2,000,000 KES (estimated annual spend)
Terms: Call-off as needed, 2-day delivery

Price List (Sample):
┌──────────────────────────┬──────┬────────────┬──────────┐
│ Item                     │ UOM  │ Rate (KES) │ Discount │
├──────────────────────────┼──────┼────────────┼──────────┤
│ A4 Paper (Ream)          │ Ream │    650     │   10%    │
│ Printer Toner HP-001     │ Unit │  4,500     │   10%    │
│ Ballpoint Pens (Box/50)  │ Box  │    350     │   10%    │
│ Folders (A4)             │ Unit │     45     │   10%    │
│ Staplers                 │ Unit │    280     │   10%    │
└──────────────────────────┴──────┴────────────┴──────────┘

Terms & Conditions:
  • Prices locked for 12 months
  • 10% discount on all items
  • Minimum order: None
  • Delivery: 2 working days
  • Payment: Net 30 days
  • Quality guarantee: 100% satisfaction

Call-Off Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Call-Off 1 (February 15):
  Release No: BPO-2025-001-R001
  Items:
    A4 Paper: 50 Reams @ 585 KES = 29,250 KES
    Toner: 5 Units @ 4,050 KES = 20,250 KES
    Total: 49,500 KES
    
  Process:
    1. User submits requisition
    2. System checks blanket PO
    3. Auto-create release (no re-approval needed)
    4. Email to supplier (automatic)
    5. Supplier delivers in 2 days
    6. GRN created
    7. Invoice received and processed

Call-Off 2 (March 10):
  Release: BPO-2025-001-R002
  Amount: 65,000 KES

[Multiple call-offs throughout the year]

Blanket PO Tracking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
As at June 30, 2025:

Total Value: 2,000,000 KES
Utilized: 950,000 KES (48%)
Remaining: 1,050,000 KES (52%)
Number of Releases: 28
Average Release Value: 34,000 KES

Status: On track ✓

Benefits:
  • No repeat RFQ/PO process
  • Faster procurement (2 days vs 7 days)
  • Price protection
  • Volume discount secured
  • Reduced admin work

Year-End Review:
  • Actual spend: 1,850,000 KES (93% of estimate)
  • Renew blanket PO for 2026
  • Negotiate additional 2% discount (loyalty)
```

### Service Procurement

```markdown
SERVICE CONTRACT MANAGEMENT

Scenario: Annual Maintenance Contract
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Service: Elevator Maintenance
Supplier: LiftTech Services Ltd
Contract Type: Annual Maintenance Contract (AMC)
Duration: 12 months (Jan 1 - Dec 31, 2025)
Value: 480,000 KES

Scope of Work:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Preventive Maintenance:
   - Monthly inspection: 12 visits/year
   - Lubrication and adjustment
   - Safety checks
   - Minor repairs (covered)
   
2. Breakdown Service:
   - 24/7 emergency response
   - Response time: 4 hours
   - No additional charge for labor
   
3. Spare Parts:
   - Minor parts: Included
   - Major parts: Chargeable (pre-approved rates)
   
4. Reporting:
   - Monthly service report
   - Quarterly summary
   - Annual compliance certificate

Contract Payment Structure:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total: 480,000 KES/year
Payment: Quarterly in advance
  Q1: 120,000 KES (due Jan 10)
  Q2: 120,000 KES (due Apr 10)
  Q3: 120,000 KES (due Jul 10)
  Q4: 120,000 KES (due Oct 10)

PO Structure:
  PO No: PO-2025-00015 (Service PO)
  Type: Service Contract
  Payment Schedule: Defined (4 milestones)
  
Accounting Treatment:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Q1 Invoice Processing:
  Invoice: LTS-INV-2025-001
  Amount: 120,000 KES + VAT (16%) = 139,200 KES
  Service Period: Jan 1 - Mar 31
  
  Entry:
    Dr. Prepaid Maintenance (Asset)     120,000
    Dr. VAT Input                        19,200
        Cr. Accounts Payable                    139,200
  
Monthly Accrual (Jan, Feb, Mar):
  Dr. Maintenance Expense              40,000
      Cr. Prepaid Maintenance                  40,000
  
  (120,000 / 3 months = 40,000/month)

Service Quality Tracking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Monthly Service Log:
┌──────┬──────────┬────────────┬──────────┬───────────┐
│ Month│ Scheduled│ Completed  │ Response │ Rating    │
├──────┼──────────┼────────────┼──────────┼───────────┤
│ Jan  │ Jan 15   │ Jan 15 ✓   │ On time  │ 5/5       │
│ Feb  │ Feb 15   │ Feb 16 ⚠   │ 1 day late│ 4/5      │
│ Mar  │ Mar 15   │ Mar 15 ✓   │ On time  │ 5/5       │
│ Apr  │ Apr 15   │ Apr 15 ✓   │ On time  │ 5/5       │
└──────┴──────────┴────────────┴──────────┴───────────┘

Breakdown Services (Q1):
  Date: Feb 22 (emergency)
  Issue: Elevator stuck at 3rd floor
  Response: 2.5 hours (within SLA) ✓
  Resolution: 4 hours
  Cost: Covered under AMC ✓

Performance Review (Quarterly):
  Scheduled visits: 100% completed
  Response time: Average 2.8 hours (target <4) ✓
  Customer satisfaction: 4.8/5.0 ✓
  Contract compliance: Excellent

Renewal Decision:
  Based on performance → Renew for 2026
  Negotiate: 2% price increase (accepted)
  New contract: 490,000 KES/year
```

### Consignment Stock

```markdown
CONSIGNMENT INVENTORY ARRANGEMENT

Scenario: Chemical Raw Material Consignment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Agreement:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Supplier: ChemSupply Ltd
Material: Chemical Additive X
Arrangement: Supplier-owned consignment stock at AWO premises
Duration: 12 months (renewable)

Terms:
  • Supplier maintains 3 months' stock at AWO
  • AWO has access to use as needed
  • Ownership transfer: Only when consumed
  • Stock monitoring: Weekly by supplier
  • Replenishment: Supplier's responsibility
  • Payment: Net 30 from consumption date
  • Min consumption commitment: 10,000 liters/year

Benefits to AWO:
  • Zero inventory investment
  • Always available (no stock-outs)
  • Payment only when consumed
  • Improved cash flow

Benefits to Supplier:
  • Guaranteed business
  • Better demand visibility
  • Competitive advantage

Stock Management:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Physical Stock Location:
  Warehouse: WH-NBI-01, Zone D (Hazardous)
  Section: Consignment Area (separated)
  Ownership: ChemSupply Ltd (clearly marked)

System Tracking:
  Inventory Type: Consignment Stock
  Quantity on Hand: 3,500 liters
  Ownership: Supplier
  Valuation: Not on AWO books

Consumption Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Production Requirement:
  Date: 2025-04-20
  Production Order: PRO-2025-0067
  Material needed: 500 liters

Material Requisition:
  MRS-2025-0456
  Item: Chemical Additive X (consignment)
  Quantity: 500 liters
  
System Action:
  1. Issue from consignment stock (physical)
  2. Record consumption
  3. Trigger ownership transfer
  4. Create GRN (ownership change)
  5. Generate payable

Accounting Entry (Consumption):
  Dr. Raw Material Inventory           250,000
  Dr. VAT Input                          40,000
      Cr. Accounts Payable - ChemSupply        290,000

  Description: Consumption of consignment stock, 500L @ 500 KES/L

Physical Stock After Consumption:
  Previous: 3,500 liters
  Consumed: 500 liters
  Balance: 3,000 liters (still supplier-owned)

Supplier Replenishment:
  Notification sent to supplier (auto)
  Supplier delivers 500 liters (no AWO action needed)
  Stock back to 3,500 liters

Monthly Reconciliation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Consignment Stock Report (April 2025):
┌──────────┬──────────┬──────────┬──────────┬──────────┐
│ Date     │ Opening  │ Consumed │ Replenish│ Closing  │
├──────────┼──────────┼──────────┼──────────┼──────────┤
│ Apr 1    │ 3,500 L  │    -     │    -     │ 3,500 L  │
│ Apr 8    │ 3,500 L  │  300 L   │  300 L   │ 3,500 L  │
│ Apr 15   │ 3,500 L  │  450 L   │  450 L   │ 3,500 L  │
│ Apr 20   │ 3,500 L  │  500 L   │  500 L   │ 3,500 L  │
│ Apr 28   │ 3,500 L  │  280 L   │  280 L   │ 3,500 L  │
│          │          │          │          │          │
│ Total    │    -     │ 1,530 L  │ 1,530 L  │ 3,500 L  │
└──────────┴──────────┴──────────┴──────────┴──────────┘

Invoice from Supplier:
  Invoice: CS-INV-2025-034
  Period: April 2025
  Quantity: 1,530 liters
  Rate: 500 KES/L
  Amount: 765,000 KES + VAT

Payment: Due May 30 (30 days from month-end)

Financial Impact:
  Cash flow benefit: ~90 days vs. buying upfront
  Working capital saved: ~1.75M KES (3 months stock value)
```

---

**Next:** [Approval Workflows](./22-approval-workflows.md)
