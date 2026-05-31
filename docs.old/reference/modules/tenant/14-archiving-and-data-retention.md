[<-- Back to Index](README.md)

## Archiving & Data Retention

### Overview

Archiving permanently deactivates a tenant while preserving data according to retention policies. Unlike suspension, archival is intended to be permanent.

### Archive Flow

```markdown
ARCHIVE TENANT FLOW:

  Request: ArchiveTenantRequest
  ├── tenant_id:      UUID
  ├── reason:         "Business closure - retain per regulation"
  ├── archived_by:    admin@awo-erp.com
  └── retention_days: 2555 (7 years)

  Step 1: Validate
  ├── Tenant exists?                  YES
  ├── Current status: ACTIVE/SUSPENDED? YES
  ├── Not already archived?           YES
  └── Proceed

  Step 2: Update Status
  ├── status: ACTIVE/SUSPENDED → ARCHIVED
  ├── updated_at: NOW()
  └── metadata: { "archive_reason": "Business closure",
                   "archived_at": "2027-01-15T10:00:00Z",
                   "retention_until": "2034-01-15T10:00:00Z",
                   "archived_by": "admin@awo-erp.com" }

  Step 3: Block All Access
  ├── All API requests → 403 Forbidden
  ├── Login attempts → "Account archived" message
  └── Background jobs → Stopped

  Step 4: Audit Log
  ├── Action:    TENANT_ARCHIVED
  ├── Actor:     admin@awo-erp.com
  └── Details:   "Business closure, 7-year retention"

  Step 5: Notification
  ├── To: All tenant admin users
  └── Subject: "Account Archived"
```

### Data Retention Policies

```markdown
RETENTION REQUIREMENTS BY REGULATION:

Regulation              Retention Period    Data Types
──────────────────────────────────────────────────────
Tax Records (KRA)       7 years            Invoices, receipts, tax filings
Financial Records       7 years            GL entries, financial statements
Employment Records      5 years            Payroll, contracts, benefits
Customer Data (GDPR)    As required        Personal data, communications
Audit Trail             10 years           All system audit logs

DEFAULT RETENTION BY PLAN:

Plan          Retention After Archive
──────────────────────────────────────
Basic         1 year (365 days)
Professional  3 years (1095 days)
Enterprise    7 years (2555 days)
Custom        Configurable
```

### Archived Tenant Behavior

```markdown
ARCHIVED TENANT STATE:

Access:
├── API:        ALL requests blocked (403)
├── Login:      Blocked with archive message
├── Data:       Preserved but inaccessible
└── Subdomain:  Reserved (not reusable)

Billing:
├── Charges:    Stopped permanently
├── Invoices:   Final invoice generated
└── Refunds:    Pro-rated if prepaid

Data:
├── Database:   Records preserved in place
├── Files:      Preserved (future: cold storage)
├── Backups:    Continue per retention policy
└── Encryption: Keys preserved for retention period

After Retention Expires:
├── Data:       Permanently deleted
├── Backups:    Purged
├── Slug:       Released for reuse
├── Subdomain:  Released for reuse
└── Audit Log:  Final entry: "Data purged per retention policy"
```

### Archive vs Soft Delete

```markdown
COMPARISON:

                    Archive              Soft Delete
────────────────────────────────────────────────────
Purpose             Business closure     Error correction
Status field        ARCHIVED             (any status)
deleted_at          NULL                 Timestamp set
Data access         Blocked              Filtered from queries
Reversible          No (by policy)       Yes (restore)
Billing             Stopped              Depends on status
Retention           Policy-based         Until restored/purged
Slug reuse          After retention      After purge
```

### Example: Full Archive Lifecycle

```markdown
EXAMPLE: Savannah Electronics Closure

Year 1 (Jan 2024): Tenant ACTIVE
  ├── 150 users, Enterprise plan
  └── Full ERP operations

Year 3 (Jan 2027): Business closure announced
  ├── Admin requests archive
  ├── Reason: "Company acquired, merging into parent"
  ├── Retention: 7 years (regulatory requirement)
  └── Status: ACTIVE → ARCHIVED

Year 3-10: Data retained
  ├── Tax records preserved for KRA
  ├── Financial statements accessible to auditors (via admin)
  ├── Employee records preserved
  └── Storage: Moved to cold tier (future)

Year 10 (Jan 2034): Retention expires
  ├── Automated purge job runs
  ├── All tenant data permanently deleted
  ├── Slug "savannah-electronics-ltd" released
  ├── Subdomain "savannahelectronics" released
  └── Final audit: "Tenant data purged - retention expired"
```

---

Next: [Audit Logging & Compliance](./15-audit-logging-and-compliance.md)
