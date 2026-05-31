[<-- Back to Index](README.md)

## 21. Approval Workflows

### Workflow Engine Integration

```markdown
SALES APPROVAL WORKFLOWS

Workflow Types:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Discount Approval
2. Credit Limit Override
3. Pricing Exception
4. Sales Order Confirmation
5. Return Authorization
6. Credit Note Approval
7. Write-Off Approval

Workflow Definition Example:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Workflow: Discount Approval

Trigger Conditions:
  IF discount > 10%
  THEN initiate approval workflow

Approval Levels:
  Level 1 (10-15%): Sales Manager
  Level 2 (15-20%): Sales Director
  Level 3 (20%+): CFO

Routing Rules:
  Document Type: Quotation
  Field: Additional Discount %
  Condition: > 10
  
  Route to:
    Role: Sales Manager
    User: Current territory manager
    Timeout: 24 hours
    Escalate to: Sales Director

Approval Workflow States:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DRAFT:
  - Sales person creating quotation
  - Can edit freely

PENDING APPROVAL:
  - Discount > threshold
  - Submitted for approval
  - Sales person cannot edit
  - Cannot send to customer

APPROVED:
  - Manager approved
  - Can proceed to send quote
  - Sales person can submit

REJECTED:
  - Manager rejected
  - Sales person notified
  - Must revise or cancel

CANCELLED:
  - Sales person cancelled request
  - Workflow terminated

Approval Actions:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Approver Options:
  ☑ Approve
    Proceed with request
    
  ☑ Approve with Comments
    Conditional approval
    Example: "Approved for this customer only,
             not to set precedent"
    
  ☑ Reject
    Request denied
    Reason required
    
  ☑ Request More Info
    Need clarification
    Returns to requester
    
  ☑ Delegate
    Forward to another approver
    (When on leave, etc.)

Approval Form:
┌────────────────────────────────────────────┐
│ APPROVAL REQUEST                           │
│                                            │
│ Document: QTN-2025-150                     │
│ Requester: Sarah Johnson                   │
│ Date: Jan 25, 2025                         │
│                                            │
│ Request Type: Discount Approval            │
│                                            │
│ Details:                                   │
│ Customer: ABC Manufacturing Ltd            │
│ Order Value: 5,000,000 KES                 │
│ Standard Discount: 10%                     │
│ Requested Discount: 15%                    │
│ Additional Discount: 5%                    │
│ Amount Impact: 250,000 KES                 │
│                                            │
│ Justification:                             │
│ "Customer is comparing with Competitor X   │
│ who quoted 10% lower. This discount        │
│ maintains our margin at 25% and wins a     │
│ high-value account with good repeat        │
│ potential."                                │
│                                            │
│ Supporting Documents:                      │
│ □ Competitor quote (attached)              │
│ □ Customer email                           │
│                                            │
│ Approver: James Ndungu (Sales Manager)     │
│ Action: ☐ Approve  ☐ Reject  ☐ More Info  │
│ Comments: _______________________________  │
│ __________________________________________ │
└────────────────────────────────────────────┘

Multi-Level Approval:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scenario: 22% Discount Request
  Value: 10,000,000 KES order

Approval Chain:
  Level 1: Sales Manager (approve <20%)
    Insufficient authority
    Auto-routes to next level
  
  Level 2: Sales Director (approve <25%)
    Has authority
    Reviews and approves
  
  Level 3: CFO
    Not reached (Director approved)

Parallel Approval:
  Some workflows need multiple approvers
  
  Example: Large Custom Order
    Technical Approval: Engineering Manager
    Commercial Approval: Sales Director
    Financial Approval: Finance Manager
    
  All three must approve before proceeding

Serial Approval:
  Approvals happen in sequence
  
  Example: Credit Limit Increase
    Step 1: Sales Manager recommends
    Step 2: Credit Manager reviews risk
    Step 3: CFO approves amount
```

### Specific Workflow Examples

```markdown
DETAILED WORKFLOW SCENARIOS

1. Quotation Discount Workflow:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Event: Sales person adds 12% discount

Auto-Actions:
  1. Status → PENDING APPROVAL
  2. Notification → Sales Manager
  3. Email sent with details
  4. Dashboard alert

Manager Reviews:
  - Customer value & history
  - Competitor situation
  - Margin impact
  - Strategic importance

Decision: APPROVED
  Comments: "Approved. Customer has been
            paying on time for 2 years.
            Good strategic account."

Result:
  - Status → APPROVED
  - Sales person notified
  - Can submit to customer
  - Audit trail logged

2. Credit Limit Override:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Event: Sales order exceeds credit limit

System Check:
  Customer: ABC Manufacturing
  Credit Limit: 5,000,000 KES
  Outstanding: 3,500,000 KES
  New Order: 2,500,000 KES
  Total Exposure: 6,000,000 KES
  
  ⚠ EXCEEDS LIMIT BY: 1,000,000 KES

Workflow Triggered:
  1. Order on hold
  2. Routed to Credit Manager
  3. Sales Manager notified

Credit Manager Review:
  Checks:
    - Payment history: Perfect ✓
    - Current aging: All current ✓
    - Financial health: Stable ✓
    - Order profitability: Good ✓
    - Relationship: 3-year customer ✓

Decision: APPROVED (Temporary Increase)
  New Limit: 6,000,000 KES
  Duration: This order only
  Condition: Next order reverts to 5M

Result:
  - Order proceeds
  - Invoice created
  - Limit tracked
  - Review in 90 days

3. Return Authorization:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Event: Customer requests return

Workflow:
  Step 1: Support Team Assessment
    - Troubleshoot issue
    - Confirm defect/reason
    - Recommend: Approve/Reject

  Step 2: Sales Manager Review
    - Order value: 2,000,000 KES
    - Return reason: Defective
    - Support confirms: Yes
    - Decision: APPROVED

  Step 3: Logistics Coordination
    - Schedule pickup
    - Create RMA
    - Issue credit note authority

  Step 4: Finance Approval (if >1M)
    - Credit note amount: 2,000,000
    - Finance Manager approves
    - GL entries authorized

Automated Notifications:
  → Customer: RMA number & instructions
  → Warehouse: Expect return
  → Finance: Credit note pending

4. Sales Order Confirmation:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
For High-Value Orders (>5M KES):

Event: Sales person creates SO-2025-250
  Value: 8,000,000 KES
  Customer: New customer (first order)

Auto-Workflow Trigger:
  Condition: Value > 5M AND First Order
  
Approval Required:
  Level 1: Sales Manager
    Checks: Customer verified, pricing OK
    Action: APPROVED
  
  Level 2: Credit Manager
    Checks: Credit application complete
    New customer assessment
    Action: APPROVED with conditions
    Condition: 50% deposit required
  
  Level 3: Finance Manager
    Final check on financials
    Action: APPROVED

Result:
  Order confirmed with conditions:
    - 50% deposit before delivery
    - Balance on delivery
  
  Sales person notified
  Customer contacted for deposit

Workflow Tracking:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Audit Trail Example:
┌──────────┬──────────┬────────────────────┐
│ Date     │ User     │ Action             │
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ Sarah J. │ Created QTN-2025-1 │
│ 09:00    │          │ Discount: 15%      │
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ System   │ Routed to Manager  │
│ 09:01    │          │ (Approval required)│
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ James N. │ Reviewed request   │
│ 14:30    │          │ Added comments     │
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ James N. │ APPROVED           │
│ 14:32    │          │ Conditional: This  │
│          │          │ customer only      │
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ System   │ Notified Sarah J.  │
│ 14:33    │          │ Status: Approved   │
├──────────┼──────────┼────────────────────┤
│ Jan 25   │ Sarah J. │ Submitted to       │
│ 15:00    │          │ customer           │
└──────────┴──────────┴────────────────────┘

Complete trail for audit/compliance
```

---