[<-- Back to Index](README.md)

## Common Business Scenarios

### Scenario 1: New Company Signup

```markdown
SCENARIO: Fresh startup signs up for AWO ERP

Company:    Nyeri Farms Ltd
Contact:    Sarah Wanjiku (sarah@nyerifarms.co.ke)
Plan:       Basic
Industry:   Agriculture
Location:   Nyeri, Kenya

STEP-BY-STEP:

1. Sarah fills signup form on awo-erp.com
   → Name: Nyeri Farms Ltd
   → Email: sarah@nyerifarms.co.ke
   → Password: ********
   → Plan: Basic

2. System provisions tenant
   → ID: new-uuid-generated
   → Slug: nyeri-farms-ltd
   → Subdomain: nyerifarms.awo-erp.com
   → Status: PENDING

3. Default config created (auto-trigger)
   → max_users: 10
   → max_entities: 5
   → Modules: ["financial", "selling"]

4. Admin user created
   → User: Sarah Wanjiku
   → Role: Tenant Administrator

5. Status updated: PENDING → ACTIVE
   → Welcome email sent
   → Sarah logs in at nyerifarms.awo-erp.com
   → Sees Financial and Selling modules

6. Sarah starts setting up:
   → Creates company entity "Nyeri Farms Ltd"
   → Sets up chart of accounts
   → Adds first customer
   → Creates first invoice: INV-001 for KES 25,000
```

### Scenario 2: Growing Business Upgrades Plan

```markdown
SCENARIO: Business outgrows Basic plan limits

Company:    Nyeri Farms Ltd (6 months later)
Current:    Basic plan (10 users, 5 entities)
Problem:    Need to add 12th user, need Buying module

TRIGGER:
  Sarah tries to create user #11
  → Error: "Tenant has reached maximum user limit (10)"
  → Dashboard shows: Users 10/10 (100%)

RESOLUTION:
  Sarah contacts support: "Need more users and Buying module"
  Admin processes upgrade: Basic → Professional

CHANGES:
  max_users:        10 → 50
  max_entities:     5 → 20
  max_transactions: 1,000 → 10,000
  storage:          1 GB → 10 GB
  allowed_modules:  + buying, inventory

RESULT:
  Sarah creates users 11 and 12 successfully
  Buying module appears in navigation
  Inventory module now available
  Monthly billing updated to Professional rate
```

### Scenario 3: Multi-Company Tenant

```markdown
SCENARIO: Enterprise tenant with multiple companies

Company:    East Africa Trading Group
Plan:       Enterprise
Entities:   3 companies under one tenant

SETUP:
  Tenant: East Africa Trading Group (eat-group)
  ├── Entity 1: EAT Kenya Ltd (Nairobi)
  │   ├── Department: Sales
  │   ├── Department: Finance
  │   └── Users: 45
  │
  ├── Entity 2: EAT Uganda Ltd (Kampala)
  │   ├── Department: Operations
  │   ├── Department: Finance
  │   └── Users: 30
  │
  └── Entity 3: EAT Tanzania Ltd (Dar es Salaam)
      ├── Department: Logistics
      ├── Department: Finance
      └── Users: 25

ISOLATION:
  All 3 entities share: One tenant context
  RLS ensures:          All entities visible to all users in tenant
  Entity-level access:  Controlled by RBAC (not tenant isolation)
  Consolidation:        Cross-entity reports possible within tenant

CROSS-TENANT:
  EAT Group CANNOT see data from Coastal Coffee or Highland Textiles
  Even though all are on the same database cluster
```

### Scenario 4: Handling Payment Failure

```markdown
SCENARIO: Automated suspension for non-payment

Timeline:
  Day 0:   Invoice issued for monthly subscription
  Day 15:  Payment reminder sent (automated)
  Day 30:  Second reminder with warning
  Day 45:  Final notice: "Suspension in 5 days"

  Day 50:  AUTO-SUSPEND triggered
           Status: ACTIVE → SUSPENDED
           Reason: "Payment overdue 50 days - auto-suspend"
           Impact: 35 users switched to read-only

           Users see:
           ┌────────────────────────────────────────┐
           │ ⚠ Account Suspended                    │
           │                                         │
           │ Your account has been suspended due to  │
           │ an overdue payment. You can view your   │
           │ data but cannot make changes.            │
           │                                         │
           │ Contact: billing@awo-erp.com            │
           └────────────────────────────────────────┘

  Day 55:  Payment received
           Status: SUSPENDED → ACTIVE
           All 35 users resume normal access
```

### Scenario 5: Platform-Wide Bulk Operation

```markdown
SCENARIO: Enabling new feature across all Professional tenants

Task:     Enable "inventory" module for all Professional tenants
          that don't already have it

BULK OPERATION:
  Type:       UPDATE_FEATURES
  Actor:      Platform Admin
  Targets:    47 Professional tenants without inventory module
  Parameters: { "enable_modules": ["inventory"] }

RESULTS:
  ┌─────────────────────────────────┬──────────┐
  │ Tenant                          │ Result   │
  ├─────────────────────────────────┼──────────┤
  │ Coastal Coffee Co.              │ COMPLETED│
  │ Highland Textiles               │ SKIPPED  │ (already has it)
  │ Nyeri Farms Ltd                 │ COMPLETED│
  │ ... (44 more)                   │ ...      │
  └─────────────────────────────────┴──────────┘

  Summary:
  ├── Total:      47
  ├── Completed:  42
  ├── Skipped:    5 (already had module)
  ├── Failed:     0
  └── Duration:   12 seconds
```

### Scenario 6: Tenant Data Export Before Archive

```markdown
SCENARIO: Company closing, needs data export

Company:    Savannah Electronics Ltd
Request:    Export all data before archiving

PROCESS:
  1. Admin generates full data export
     ├── Financial: All GL entries, journals (3 years)
     ├── Sales: All invoices, customers, orders
     ├── HR: All employee records, payroll history
     └── Format: CSV + PDF reports

  2. Export delivered to tenant admin
     ├── Encrypted ZIP file
     ├── Download link (valid 7 days)
     └── Includes data dictionary

  3. Tenant archived
     ├── Status: ACTIVE → ARCHIVED
     ├── Retention: 7 years
     └── All access blocked

  4. Confirmation sent
     ├── Archive reference number
     ├── Retention expiry date
     └── Contact for data requests during retention
```

---

Next: [API Reference](./19-api-reference.md)
