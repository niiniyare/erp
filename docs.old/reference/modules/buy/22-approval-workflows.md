[<-- Back to Index](README.md)

## Approval Workflows

### Approval Matrix

```markdown
PROCUREMENT APPROVAL AUTHORITY MATRIX

Purchase Requisition Approvals:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

┌──────────────────┬──────────────────────────────────────────┐
│ Amount (KES)     │ Approval Chain                           │
├──────────────────┼──────────────────────────────────────────┤
│ 0 - 100,000      │ Department Head                          │
│ 100,001 - 500,000│ Department Head → Procurement Manager    │
│ 500,001 - 2M     │ Dept Head → Proc Manager → Director     │
│ 2,000,001 - 5M   │ Dept Head → Proc Mgr → Dir → CFO       │
│ 5,000,001 - 10M  │ Dept Head → Proc Mgr → Dir → CFO → CEO │
│ > 10,000,000     │ Full chain → Board Approval              │
└──────────────────┴──────────────────────────────────────────┘

Purchase Order Approvals:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

┌──────────────────┬──────────────────────────────────────────┐
│ Amount (KES)     │ Approval Chain                           │
├──────────────────┼──────────────────────────────────────────┤
│ 0 - 500,000      │ Buyer (within approved PR scope)         │
│ 500,001 - 2M     │ Operational Procurement Manager          │
│ 2,000,001 - 5M   │ Strategic Procurement Manager            │
│ 5,000,001 - 10M  │ Procurement Director + CFO               │
│ > 10,000,000     │ Procurement Director + CFO + CEO         │
└──────────────────┴──────────────────────────────────────────┘

Payment Approvals:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

┌──────────────────┬──────────────────────────────────────────┐
│ Amount (KES)     │ Approval Chain                           │
├──────────────────┼──────────────────────────────────────────┤
│ 0 - 500,000      │ AP Manager                               │
│ 500,001 - 2M     │ AP Manager + Finance Manager             │
│ 2,000,001 - 5M   │ AP Manager + Finance Manager + CFO       │
│ > 5,000,000      │ AP Manager + FM + CFO + CEO              │
└──────────────────┴──────────────────────────────────────────┘

Special Approvals:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Emergency Purchase: Fast-track (single approver at threshold level)
Budget Override: CFO mandatory
New Supplier Onboarding: Procurement Director + Finance Manager
Single Source Justification: Procurement Director + CFO
Contract Award (>2M): Tender Committee
```

### Workflow Engine

```markdown
WORKFLOW CONFIGURATION

PR Approval Workflow Example:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Workflow: WF-PR-APPROVE
Trigger: PR Submitted
Document: Purchase Requisition

┌─────────────────────────────────────────────────────────┐
│                    PR SUBMITTED                         │
│                        │                                │
│               ┌────────▼────────┐                       │
│               │ Budget Check    │                       │
│               │ (Automated)     │                       │
│               └───┬─────────┬───┘                       │
│                   │         │                           │
│              ✓ Pass    ✗ Fail                           │
│                   │         │                           │
│                   │    ┌────▼─────────┐                 │
│                   │    │ Notify User: │                 │
│                   │    │ Over Budget  │                 │
│                   │    │ (Request     │                 │
│                   │    │ supplement)  │                 │
│                   │    └──────────────┘                 │
│                   │                                     │
│          ┌────────▼────────┐                            │
│          │ Amount Check    │                            │
│          │ (Route to       │                            │
│          │ correct level)  │                            │
│          └───┬──────┬──────┘                            │
│              │      │                                   │
│         <500K    >500K                                  │
│              │      │                                   │
│     ┌────────▼──┐ ┌─▼──────────┐                       │
│     │ Dept Head │ │ Dept Head  │                       │
│     │ Approve   │ │ Approve    │                       │
│     └─────┬─────┘ └─────┬──────┘                       │
│           │              │                              │
│           │     ┌────────▼─────────┐                    │
│           │     │ Proc Manager     │                    │
│           │     │ Approve          │                    │
│           │     └────────┬─────────┘                    │
│           │              │                              │
│     ┌─────▼──────────────▼─────┐                       │
│     │     PR APPROVED          │                       │
│     │     (Ready for PO)       │                       │
│     └──────────────────────────┘                       │
└─────────────────────────────────────────────────────────┘

Workflow Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Auto-Approval Rules:
   - PRs <10,000 KES for recurring items → Auto-approved
   - Blanket PO call-offs → Auto-approved (within limit)
   - Repeat orders matching last 3 months → Skip RFQ step

2. Escalation Rules:
   - No action within 24 hours → Reminder email
   - No action within 48 hours → Escalate to next level
   - No action within 72 hours → Escalate to Director
   
3. Delegation Rules:
   - Approver on leave → Delegate to named backup
   - Delegation period: Defined start/end dates
   - Delegate authority: Same as original approver
   - Audit trail: Shows delegate action

4. Rejection & Return Rules:
   - Any approver can reject → PR returned to creator
   - Reason mandatory for rejection
   - Creator can modify and resubmit
   - Resubmitted PR goes through full approval again

Notification Configuration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Event Notifications:
┌────────────────────────┬────────────┬────────────────────┐
│ Event                  │ Channel    │ Recipients         │
├────────────────────────┼────────────┼────────────────────┤
│ PR Awaiting Approval   │ Email+Push │ Next approver      │
│ PR Approved            │ Email      │ Requester          │
│ PR Rejected            │ Email      │ Requester          │
│ PO Created             │ Email      │ Supplier + Buyer   │
│ PO Approved            │ Email      │ Buyer + Requester  │
│ GRN Created            │ Email      │ Buyer + AP Team    │
│ Invoice Due for Payment│ Email      │ AP Manager         │
│ Payment Executed       │ Email      │ Supplier + Finance │
│ Budget Alert (>80%)    │ Email+Push │ Dept Head + CFO    │
│ Overdue Delivery       │ Email      │ Buyer + Expeditor  │
│ Approval Reminder      │ Push       │ Pending approver   │
│ Approval Escalation    │ Email+Push │ Next level approver│
└────────────────────────┴────────────┴────────────────────┘
```

### Approval Scenarios

```markdown
SCENARIO 1: Standard PR to PO Approval

PR-2025-00456
Amount: 800,000 KES
Category: Raw Materials
Department: Production
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Timeline:
  Day 1, 09:00 - PR Created by Production Planner
    Status: Submitted
    Budget Check: ✓ Within budget
    
  Day 1, 10:30 - Dept Head Reviews
    Action: Approved
    Comments: "Required for next week production run"
    Status: Dept Approved
    
  Day 1, 14:00 - Proc Manager Reviews
    Action: Approved (amount >500K requires Proc Mgr)
    Comments: "Supplier confirmed, good pricing"
    Status: Fully Approved
    
  Day 2, 09:00 - Buyer Creates PO
    PO-2025-00400
    Amount: 800,000 KES
    Supplier: Ace Steel Suppliers
    Auto-approved (within Proc Mgr limit)
    
  Day 2, 09:15 - PO Emailed to Supplier
    Status: Active

Total Approval Time: 1 business day ✓


SCENARIO 2: High-Value PO with Multi-Level Approval

PO-2025-00325
Amount: 11,300,000 KES (CNC Machine)
Category: Capital Equipment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Approval Chain (>10M requires Board):
  
  Day 1: Procurement Manager → ✓ Approved
    "RFQ evaluation complete, recommended supplier"
    
  Day 2: Procurement Director → ✓ Approved
    "Technical evaluation supports selection"
    
  Day 3: CFO → ✓ Approved
    "CAPEX budget available, ROI verified"
    
  Day 3: CEO → ✓ Approved
    "Aligned with expansion strategy"
    
  Day 5: Board (via circular resolution) → ✓ Approved
    "Board resolution BR-2025-023 passed"
    
  Day 6: PO Released
    Status: Active, sent to supplier

Total Approval Time: 6 business days
Note: Acceptable for capital purchase


SCENARIO 3: Rejected PR

PR-2025-00462
Amount: 3,500,000 KES
Category: IT Equipment (Server upgrade)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Day 1: IT Manager approves → ✓
  Day 2: Procurement Manager approves → ✓
  Day 3: CFO Reviews → ✗ REJECTED
  
  Rejection Reason:
    "Cloud migration planned for Q3. Server purchase may
    not be needed. Please discuss with CTO and resubmit
    if cloud option is not suitable."
    
  Notification to IT Manager:
    "Your PR-2025-00462 has been rejected by CFO.
    Reason: [as above]
    Please review and resubmit if appropriate."
    
  IT Manager Actions:
    Option A: Cancel PR (cloud option viable)
    Option B: Modify and resubmit with justification
    
  Outcome: IT Manager consulted with CTO
    Decision: Proceed with cloud → PR cancelled


SCENARIO 4: Emergency Approval

PR-2025-00458 (EMERGENCY)
Amount: 85,000 KES
Category: Spare Parts (Machine breakdown)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  10:15 AM: PR Created with EMERGENCY flag
  10:16 AM: Auto-budget check → ✓
  10:17 AM: Auto-approved (Emergency <100K policy)
  10:20 AM: Buyer creates PO (phone order to supplier)
  10:25 AM: PO approved (emergency protocol)
  
  Post-Approval Review:
    Within 48 hours: Procurement Manager reviews
    Validates: Emergency was genuine
    Signs off: ✓ Justified
    
  Total Approval Time: 2 minutes ✓
  (Emergency protocol working as designed)
```

### Delegation & Absence Management

```markdown
DELEGATION MANAGEMENT

Delegation Setup:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Scenario: Procurement Director on leave

Delegation Record:
  Delegator: Paul Kariuki (Procurement Director)
  Delegate: Jane Muthoni (Strategic Proc Manager)
  Period: Apr 21 - Apr 30, 2025
  Authority: All procurement approvals up to 10M KES
  Exclusions: Board-level approvals (deferred)
  
Setup in System:
  1. Director creates delegation request
  2. System validates: Delegate has appropriate role
  3. CEO approves delegation (for Director-level)
  4. System activates on start date
  5. All pending approvals route to delegate
  6. System deactivates on end date
  7. Audit trail preserved

During Delegation:
  Approval shows:
    "Approved by: Jane Muthoni
     Acting for: Paul Kariuki (Delegation D-2025-008)
     Authority: Procurement Director delegation"

Standing Delegations:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Pre-configured backup approvers:

┌──────────────────────┬─────────────────────┬───────────┐
│ Primary Approver     │ Backup              │ Authority │
├──────────────────────┼─────────────────────┼───────────┤
│ Proc Director        │ Strategic Proc Mgr  │ Up to 10M │
│ Strategic Proc Mgr   │ Operational Proc Mgr│ Up to 5M  │
│ Operational Proc Mgr │ Senior Buyer        │ Up to 2M  │
│ CFO                  │ Finance Manager     │ Up to 5M  │
│ CEO                  │ COO                 │ Up to 10M │
└──────────────────────┴─────────────────────┴───────────┘

Rules:
  • Backup activates only when primary is unavailable
  • Backup authority may be lower than primary
  • Items exceeding backup authority: Wait for primary return
  • Maximum delegation period: 30 days
  • Extension requires re-approval
```

### Audit Trail

```markdown
APPROVAL AUDIT LOG

Document: PO-2025-00400
Full Audit Trail:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

┌──────────────────┬──────────────┬────────────────────────┐
│ Timestamp        │ User         │ Action                 │
├──────────────────┼──────────────┼────────────────────────┤
│ 2025-04-15 08:30│ Mary K.      │ PR-00456 Created       │
│                  │ (Planner)    │ Amount: 800K KES       │
│                  │              │                        │
│ 2025-04-15 08:30│ SYSTEM       │ Budget check: PASS     │
│                  │              │ Available: 26.3M KES   │
│                  │              │                        │
│ 2025-04-15 10:30│ John M.      │ PR-00456 Approved      │
│                  │ (Dept Head)  │ Level 1                │
│                  │              │                        │
│ 2025-04-15 14:00│ David O.     │ PR-00456 Approved      │
│                  │ (Proc Mgr)   │ Level 2 (Final)       │
│                  │              │                        │
│ 2025-04-16 09:00│ Buyer A      │ PO-00400 Created       │
│                  │ (Buyer)      │ From PR-00456          │
│                  │              │                        │
│ 2025-04-16 09:05│ SYSTEM       │ PO Auto-approved       │
│                  │              │ (Within Proc Mgr limit)│
│                  │              │                        │
│ 2025-04-16 09:15│ SYSTEM       │ PO Emailed to supplier │
│                  │              │ Email: supplier@ace.co │
│                  │              │                        │
│ 2025-04-28 14:20│ Warehouse    │ GRN-0150 Created       │
│                  │ (James W.)   │ Qty: 10.25 Tons       │
│                  │              │                        │
│ 2025-04-29 09:00│ QC Team      │ Inspection: PASSED     │
│                  │ (Susan N.)   │ Certificate verified   │
│                  │              │                        │
│ 2025-05-02 10:00│ AP Clerk     │ Invoice PINV-234       │
│                  │ (Alice M.)   │ Matched to PO+GRN     │
│                  │              │                        │
│ 2025-05-02 10:05│ SYSTEM       │ 3-Way Match: PASS     │
│                  │              │ Variance: 0.0%        │
│                  │              │                        │
│ 2025-05-25 09:00│ AP Manager   │ Payment approved       │
│                  │ (Grace A.)   │ Batch PMT-B-045       │
│                  │              │                        │
│ 2025-05-25 14:00│ Finance      │ Payment executed       │
│                  │ (SYSTEM)     │ EFT: 930,800 KES      │
│                  │              │                        │
│ 2025-05-25 14:05│ SYSTEM       │ Document CLOSED        │
│                  │              │ Full cycle complete    │
└──────────────────┴──────────────┴────────────────────────┘

Compliance Check:
  ✓ Segregation of duties maintained
  ✓ All approvals within authority
  ✓ Budget validated before approval
  ✓ 3-way match verified
  ✓ Payment within approved terms
  ✓ Complete audit trail preserved
```

---

**Next:** [Troubleshooting Guide](./23-troubleshooting-guide.md)
