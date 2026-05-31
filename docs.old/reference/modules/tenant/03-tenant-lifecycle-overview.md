[<-- Back to Index](README.md)

## Tenant Lifecycle Overview

### Status Flow

Every tenant moves through defined statuses during its lifecycle:

```markdown
TENANT STATUS FLOW:

  [New Signup]
       │
       ▼
  ┌─────────┐
  │ PENDING  │ ← Initial state after provisioning
  └────┬─────┘
       │  (admin activates / auto-activation)
       ▼
  ┌─────────┐     suspend      ┌───────────┐
  │ ACTIVE  │ ───────────────→ │ SUSPENDED │
  └────┬────┘ ←─────────────── └─────┬─────┘
       │       reactivate            │
       │                             │ (cannot reactivate)
       │                             ▼
       │  archive             ┌───────────┐
       └────────────────────→ │ ARCHIVED  │
                              └───────────┘
                                    │
                                    │ (data retention period)
                                    ▼
                              [Data Purged]
```

### Status Definitions

| Status | Description | Data Access | Billing |
|--------|------------|-------------|---------|
| PENDING | Just provisioned, awaiting activation | Read-only setup | Not started |
| ACTIVE | Fully operational | Full read/write | Active |
| SUSPENDED | Temporarily disabled | Read-only | Paused or overdue |
| ARCHIVED | Permanently deactivated | No access | Stopped |

### Real-World Lifecycle Example

```markdown
EXAMPLE: Savannah Electronics Ltd

Day 1 (Jan 15, 2024):
  Status: PENDING
  Action: Platform admin provisions tenant
  Result: Tenant record created, default config applied
          Slug: savannah-electronics-ltd
          Subdomain: savannah-electronics.awo-erp.com

Day 1 (Jan 15, 2024):
  Status: ACTIVE
  Action: Admin user created, modules enabled
  Result: Tenant fully operational
          150 user slots available (Enterprise plan)
          All modules enabled

Month 6 (Jul 15, 2024):
  Status: SUSPENDED
  Action: Payment overdue for 30 days
  Reason: "Payment overdue - invoice #INV-2024-0612"
  Result: Users can view data but cannot create/edit
          API write operations blocked
          Billing team notified

Month 6 (Jul 22, 2024):
  Status: ACTIVE
  Action: Payment received, tenant reactivated
  Result: Full access restored
          All users can resume work
          Activity logged in audit trail

Year 3 (Jan 15, 2027):
  Status: ARCHIVED
  Action: Company closed, requested data archive
  Reason: "Business closure - retain data per policy"
  Retention: 7 years per regulatory requirement
  Result: Data preserved but inaccessible
          Storage moved to cold tier
```

### Lifecycle Events and Triggers

```markdown
AUTOMATED TRIGGERS:

On Provisioning (→ PENDING):
├── Generate unique slug from company name
├── Create default tenant_configurations record
├── Set timezone and currency defaults
└── Log provisioning event

On Activation (→ ACTIVE):
├── Create admin user account
├── Enable default modules
├── Send welcome notification
├── Start billing cycle
└── Initialize usage tracking

On Suspension (→ SUSPENDED):
├── Block write API operations
├── Notify all tenant admins
├── Log suspension reason
├── Pause billing (if applicable)
└── Update last_activity_at

On Reactivation (→ ACTIVE):
├── Restore full access
├── Resume billing
├── Notify tenant admins
├── Log reactivation event
└── Reset rate limit counters

On Archive (→ ARCHIVED):
├── Block all API access
├── Stop billing permanently
├── Set data retention period
├── Move to cold storage (future)
├── Final audit log entry
└── Notify platform administrators
```

### Tenant Record Structure

Each tenant stores the following core information:

```markdown
TENANT RECORD FIELDS:

Identity:
├── id           : UUID (primary key)
├── slug         : Auto-generated URL-safe identifier
├── name         : Company display name
├── email        : Primary contact email
└── subdomain    : Custom subdomain (optional)

Classification:
├── industry     : Business sector
├── company_size : Small, Medium, Large, Enterprise
├── plan_type    : basic, professional, enterprise
└── tax_id       : Tax identification number

Localization:
├── timezone     : e.g., Africa/Nairobi
├── currency_code: e.g., KES, USD, EUR
└── settings     : JSONB for tenant-specific config

Status:
├── status       : PENDING, ACTIVE, SUSPENDED, ARCHIVED
├── deleted_at   : Soft-delete timestamp (NULL = active)
└── last_activity_at : Last API activity timestamp

Audit:
├── created_at   : Tenant creation timestamp
└── updated_at   : Last modification timestamp
```

---

Next: [Initial Setup & Configuration](./04-initial-setup-and-configuration.md)
