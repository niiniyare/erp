> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## Procurement Teams & Organization

### Team Structure

```markdown
PROCUREMENT DEPARTMENT ORGANIZATION

AWO Manufacturing Ltd - Procurement Function
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

                    Procurement Director
                    (Paul Kariuki)
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
  Strategic         Operational      Accounts Payable
  Procurement       Procurement           Manager
  Manager           Manager           (Grace Akinyi)
  (Jane Muthoni)    (David Otieno)         │
        │                 │                 │
        │                 │            ┌────┴────┐
        │                 │            │         │
   ┌────┴────┐       ┌────┴────┐   AP Clerk  AP Clerk
   │         │       │         │   (2 staff) (2 staff)
Contract  Supplier  Buyers   Expeditor
Specialist Relations (4 staff) (1 staff)
(1 staff) (1 staff)

Total Team Size: 15 members

Roles & Responsibilities:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Procurement Director
   - Overall procurement strategy
   - Supplier relationship management (strategic)
   - Policy development
   - Budget oversight
   - Senior management reporting
   - Approval authority: Unlimited
   
2. Strategic Procurement Manager
   - Category management
   - Sourcing strategy
   - Contract negotiations (major)
   - Supplier development
   - Cost optimization initiatives
   - Approval authority: Up to 5M KES
   
3. Operational Procurement Manager
   - Daily procurement operations
   - Team supervision (buyers, expeditor)
   - Process improvement
   - Supplier performance monitoring
   - Approval authority: Up to 2M KES
   
4. Buyers (4 staff)
   Each buyer assigned to specific categories:
   
   Buyer 1 - Raw Materials:
     • Steel, metals
     • Chemical raw materials
     • Primary inputs
     
   Buyer 2 - Production Supplies:
     • Packaging materials
     • Labels, cartons
     • Production consumables
     
   Buyer 3 - MRO & Capital Equipment:
     • Maintenance supplies
     • Spare parts
     • Machinery and equipment
     
   Buyer 4 - Services & Indirect:
     • Office supplies
     • IT equipment
     • Professional services
     • Utilities management
   
   Buyer Responsibilities:
     • Process purchase requisitions
     • Issue RFQs
     • Evaluate quotations
     • Create purchase orders
     • Follow up on deliveries
     • Approval authority: Up to 500K KES
     
5. Expeditor
   - Track open POs
   - Follow up with suppliers on deliveries
   - Coordinate with warehouse for receipts
   - Handle urgent/emergency orders
   - Update delivery schedules
   
6. Contract Specialist
   - Draft and review contracts
   - Maintain contract repository
   - Track contract renewals
   - Ensure compliance with terms
   - Support negotiations
   
7. Supplier Relations Officer
   - Onboard new suppliers
   - Maintain supplier database
   - Handle supplier queries
   - Organize supplier meetings
   - Performance scorecarding
   
8. Accounts Payable Manager
   - AP team supervision
   - Invoice processing oversight
   - Payment scheduling
   - Vendor account reconciliation
   - Month-end close activities
   - Approval authority: Payment approvals up to 2M KES
   
9. AP Clerks (4 staff)
   - Invoice entry and matching
   - Query resolution
   - Payment preparation
   - Vendor communications
   - Filing and documentation
```

### Roles Matrix

```markdown
PROCUREMENT RACI MATRIX

Activity: Purchase Requisition to Payment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

R = Responsible, A = Accountable, C = Consulted, I = Informed

┌─────────────────────┬──────┬──────┬──────┬──────┬──────┬──────┐
│ Activity            │Buyer │Oper  │Strat │Dir   │AP Mgr│Dept  │
│                     │      │Mgr   │Mgr   │      │      │Head  │
├─────────────────────┼──────┼──────┼──────┼──────┼──────┼──────┤
│ Create PR           │  I   │  I   │  -   │  -   │  -   │  R/A │
│ Approve PR (<500K)  │  -   │  A   │  -   │  -   │  -   │  C   │
│ Approve PR (>500K)  │  -   │  C   │  A   │  I   │  -   │  C   │
│ Issue RFQ           │  R   │  A   │  C   │  -   │  -   │  I   │
│ Evaluate Quotes     │  R   │  A   │  C   │  -   │  -   │  C   │
│ Negotiate           │  R   │  A   │  R   │  I   │  -   │  -   │
│ Create PO           │  R   │  A   │  I   │  -   │  -   │  I   │
│ Approve PO (<1M)    │  I   │  A   │  -   │  -   │  -   │  -   │
│ Approve PO (>1M)    │  I   │  C   │  A   │  I   │  -   │  -   │
│ Track Delivery      │  C   │  I   │  -   │  -   │  -   │  -   │
│ Receive Goods       │  I   │  I   │  -   │  -   │  -   │  -   │
│ Process Invoice     │  -   │  -   │  -   │  -   │  R/A │  -   │
│ Approve Payment     │  -   │  -   │  C   │  A   │  R   │  -   │
│ Make Payment        │  -   │  -   │  -   │  I   │  A   │  -   │
└─────────────────────┴──────┴──────┴──────┴──────┴──────┴──────┘

Note: Warehouse team (R/A for goods receipt) not shown in this matrix

Segregation of Duties:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Critical Control: No single person can:
  ✓ Create PO AND approve it
  ✓ Receive goods AND approve invoice
  ✓ Approve invoice AND make payment
  ✓ Create supplier AND approve payment to that supplier

Example Scenario:
  Buyer A creates PO for 800K
  → Operational Manager approves PO
  → Warehouse receives goods (independent team)
  → AP Clerk processes invoice
  → AP Manager approves payment
  → Finance executes payment (separate department)
  
  Result: Proper segregation maintained ✓
```

### Category Management

```markdown
CATEGORY ASSIGNMENTS

Strategic Categories:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Managed by: Strategic Procurement Manager

1. Raw Materials - Steel & Metals
   Annual Spend: 35M KES
   Suppliers: 5 strategic
   Strategy: Long-term contracts, price indexation
   Review Cycle: Quarterly
   
   Key Activities:
     • Market intelligence gathering
     • Price trend analysis
     • Supplier diversification
     • Contract negotiations
     • Risk management

2. Capital Equipment
   Annual Spend: 20M KES
   Suppliers: Various (project-based)
   Strategy: Competitive bidding, TCO analysis
   Review Cycle: Per project
   
   Key Activities:
     • Technical specification development
     • Supplier qualification
     • Bid evaluation
     • Contract terms negotiation
     • Commissioning support

Operational Categories:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Managed by: Operational Team (Buyers)

3. Packaging Materials
   Annual Spend: 25M KES
   Buyer: Buyer 2
   Suppliers: 8
   Strategy: Blanket POs, call-off system
   
4. MRO & Consumables
   Annual Spend: 18M KES
   Buyer: Buyer 3
   Suppliers: 15 (consolidation ongoing)
   Strategy: Supplier consolidation, volume leverage
   
5. Production Chemicals
   Annual Spend: 12M KES
   Buyer: Buyer 1
   Suppliers: 4
   Strategy: Quality focus, JIT delivery
   
6. Office & IT
   Annual Spend: 8M KES
   Buyer: Buyer 4
   Suppliers: 12
   Strategy: Standardization, e-commerce
   
7. Services
   Annual Spend: 15M KES
   Buyer: Buyer 4
   Suppliers: Various
   Strategy: Service level agreements

Category Review Process:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Quarterly Category Reviews:

Agenda:
  1. Spend analysis (trends, variances)
  2. Supplier performance review
  3. Market conditions update
  4. Savings opportunities
  5. Risk assessment
  6. Strategy adjustments
  
Attendees:
  - Category manager/buyer
  - Strategic/Operational Manager
  - Key stakeholders (Engineering, Production, etc.)
  
Output:
  - Action plan for next quarter
  - Sourcing initiatives
  - Supplier relationship actions
```

### Training & Development

```markdown
PROCUREMENT TEAM DEVELOPMENT

Onboarding Program (New Hires):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Week 1: ERP System & Processes
  • System navigation training
  • Document workflow overview
  • Approval matrix understanding
  • Policy and procedure review
  
Week 2: Category Familiarization
  • Product knowledge sessions
  • Supplier introductions
  • Contract review
  • Shadowing senior buyer
  
Week 3-4: Hands-on Practice
  • Process PRs under supervision
  • Participate in RFQs
  • Supplier interactions (guided)
  • First solo PO (small value)
  
Month 2: Progressive Responsibility
  • Independent PO creation (approved categories)
  • Lead RFQ process (with review)
  • Supplier negotiations (observing)
  
Month 3: Full Responsibility
  • Full buyer authority (within limits)
  • Own category assignments
  • Performance evaluation

Continuous Learning:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Internal Training (Quarterly):
  • ERP system updates
  • New policy rollouts
  • Process improvements
  • Best practice sharing
  • Case study discussions

External Training (Annual):
  • Professional certification programs:
    - CIPS (Chartered Institute of Procurement & Supply)
    - CPSM (Certified Professional in Supply Management)
  • Industry conferences
  • Supplier days/exhibitions
  • Contract negotiation workshops

Skills Matrix:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Target Competency Levels: 1=Basic, 2=Intermediate, 3=Advanced

┌──────────────────────┬────────┬────────┬────────┬────────┐
│ Skill                │ Buyer  │ Oper   │ Strat  │ Dir    │
│                      │        │ Mgr    │ Mgr    │        │
├──────────────────────┼────────┼────────┼────────┼────────┤
│ ERP System           │   3    │   3    │   2    │   1    │
│ Negotiation          │   2    │   3    │   3    │   3    │
│ Contract Management  │   1    │   2    │   3    │   3    │
│ Category Knowledge   │   3    │   2    │   2    │   2    │
│ Market Analysis      │   1    │   2    │   3    │   3    │
│ Financial Analysis   │   2    │   3    │   3    │   3    │
│ Supplier Relationship│   2    │   3    │   3    │   3    │
│ Strategic Sourcing   │   1    │   2    │   3    │   3    │
│ Risk Management      │   1    │   2    │   3    │   3    │
└──────────────────────┴────────┴────────┴────────┴────────┘

Development Plans:
  • Annual skill assessment
  • Individual development plans
  • Training budget allocation
  • Mentorship program
```

### Performance Management

```markdown
TEAM KPIs & EVALUATION

Individual Buyer KPIs:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Cost Savings
   Target: 5% of category spend
   Measurement: Documented savings vs. previous year
   Weight: 30%
   
2. PO Processing Time
   Target: 3 days (PR to PO)
   Measurement: System calculated average
   Weight: 15%
   
3. Supplier Performance
   Target: Average score >90/100
   Measurement: Supplier scorecard for assigned suppliers
   Weight: 20%
   
4. Budget Compliance
   Target: 100% (no unauthorized purchases)
   Measurement: Compliance audit
   Weight: 15%
   
5. Error Rate
   Target: <2% (PO errors requiring correction)
   Measurement: PO amendments / Total POs
   Weight: 10%
   
6. Stakeholder Satisfaction
   Target: >4.0/5.0
   Measurement: Internal customer survey (quarterly)
   Weight: 10%

Team Performance Metrics:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Department-level KPIs:

1. Total Cost Savings
   Target: 10M KES annual
   Q1 Actual: 2.9M KES (on track) ✓
   
2. Procurement Cycle Time
   Target: 15 days (PR to GRN)
   Q1 Actual: 14.2 days ✓
   
3. Supplier On-Time Delivery
   Target: 95%
   Q1 Actual: 94.2% (close)
   
4. First-Time Match Rate
   Target: 85%
   Q1 Actual: 88% ✓
   
5. Maverick Spend
   Target: <5%
   Q1 Actual: 3.2% ✓

Performance Review Cycle:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Quarterly Reviews:
  • KPI performance discussion
  • Challenges and support needs
  • Goals for next quarter
  
Annual Reviews:
  • Comprehensive performance evaluation
  • Career development discussion
  • Salary review
  • Training needs assessment

Recognition & Rewards:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Savings Incentive Program:
  • 5% of documented savings as team bonus
  • Individual contributor recognition
  • Annual awards:
    - Buyer of the Year
    - Best Cost Reduction Initiative
    - Supplier Relationship Excellence
    
Example:
  Q1 Savings: 2.9M KES
  Team Bonus Pool: 145K KES
  Distribution: Based on individual contribution
```

### Cross-Functional Collaboration

```markdown
STAKEHOLDER ENGAGEMENT

Regular Touchpoints:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Procurement-Production Weekly Meeting
   Attendees:
     • Operational Procurement Manager
     • Production Manager
     • Materials Planner
   
   Agenda:
     • Production schedule review
     • Material availability
     • Upcoming requirements
     • Expedite requests
     • Quality issues
   
   Duration: 1 hour
   Day: Monday 9:00 AM

2. Procurement-Finance Monthly Meeting
   Attendees:
     • Procurement Director
     • CFO / Finance Manager
     • AP Manager
   
   Agenda:
     • Budget performance review
     • Payment schedule review
     • Cash flow planning
     • Savings report
     • Policy matters
   
   Duration: 2 hours
   Day: Last Friday of month

3. Procurement-Engineering Quarterly Review
   Attendees:
     • Strategic Procurement Manager
     • Engineering Manager
     • Relevant buyers
   
   Agenda:
     • New product requirements
     • Specification changes
     • Supplier technical capabilities
     • Value engineering opportunities
   
   Duration: 2 hours
   Day: Third Thursday of quarter-end month

4. All-Hands Procurement Meeting
   Attendees: Full procurement team
   Frequency: Monthly
   
   Agenda:
     • Performance review
     • Process improvements
     • Team updates
     • Training sessions
     • Recognition

Service Level Agreements:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Procurement SLAs to Internal Customers:

1. Purchase Requisition Response
   • Acknowledgment: Within 4 hours
   • Initial review: Within 1 business day
   • RFQ issuance (if needed): Within 3 days
   • PO creation: Within 5 days
   
2. Query Response
   • Email queries: Within 8 hours
   • Urgent requests: Within 2 hours
   • Status updates: Real-time via system
   
3. Delivery Expediting
   • Expedite request acknowledgment: Same day
   • Supplier follow-up: Within 24 hours
   • Status update to requester: Within 48 hours

Internal Customer Satisfaction:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Quarterly Survey Results (Q1 2025):

┌────────────────────────────┬─────────┬────────┐
│ Metric                     │ Score   │ Target │
├────────────────────────────┼─────────┼────────┤
│ Responsiveness             │ 4.3/5.0 │ >4.0 ✓ │
│ Quality of Service         │ 4.1/5.0 │ >4.0 ✓ │
│ Knowledge & Expertise      │ 4.4/5.0 │ >4.0 ✓ │
│ Problem Resolution         │ 3.9/5.0 │ >4.0 ⚠ │
│ Communication              │ 4.2/5.0 │ >4.0 ✓ │
│                            │         │        │
│ OVERALL                    │ 4.2/5.0 │ >4.0 ✓ │
└────────────────────────────┴─────────┴────────┘

Action: Improve problem resolution process
```

---

**Next:** [Supplier Performance Management](./18-supplier-performance-management.md)
