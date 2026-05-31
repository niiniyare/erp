[<-- Back to Index](README.md)

## Tenant Suspension & Reactivation

### Suspension

Suspension temporarily disables a tenant's write access while preserving all data.

### Suspension Reasons

```markdown
COMMON SUSPENSION TRIGGERS:

1. Payment Overdue
   Trigger:   Billing system detects 30+ days overdue
   Action:    Auto-suspend or manual suspend
   Impact:    Write operations blocked, read-only access

2. Policy Violation
   Trigger:   Terms of service breach detected
   Action:    Manual suspend by platform admin
   Impact:    Full access blocked pending review

3. Security Incident
   Trigger:   Suspicious activity detected
   Action:    Immediate auto-suspend
   Impact:    Full access blocked, investigation initiated

4. Maintenance
   Trigger:   Tenant-specific data migration needed
   Action:    Temporary suspend during maintenance
   Impact:    Brief downtime for data integrity
```

### Suspension Flow

```markdown
SUSPEND TENANT FLOW:

  Request: SuspendTenantRequest
  ├── tenant_id:    UUID
  ├── reason:       "Payment overdue - INV-2024-0612"
  ├── suspended_by: admin@awo-erp.com
  └── notify:       true

  Step 1: Validate
  ├── Tenant exists?          YES
  ├── Tenant not deleted?     YES
  ├── Current status: ACTIVE? YES
  └── Proceed

  Step 2: Update Status
  ├── status: ACTIVE → SUSPENDED
  ├── updated_at: NOW()
  └── metadata: { "suspension_reason": "Payment overdue",
                   "suspended_at": "2024-07-15T10:00:00Z",
                   "suspended_by": "admin@awo-erp.com" }

  Step 3: Audit Log
  ├── Action:    TENANT_SUSPENDED
  ├── Actor:     admin@awo-erp.com
  ├── Details:   "Payment overdue - INV-2024-0612"
  └── Timestamp: 2024-07-15 10:00:00 EAT

  Step 4: Notification (if notify = true)
  ├── To: All tenant admin users
  ├── Subject: "Account Suspended - Action Required"
  └── Body: Reason + steps to resolve

  Result: Tenant suspended, users see read-only access
```

### Impact of Suspension

```markdown
SUSPENDED TENANT BEHAVIOR:

API Operations:
├── GET  requests: ALLOWED (read-only)
├── POST requests: BLOCKED (403 Forbidden)
├── PUT  requests: BLOCKED (403 Forbidden)
├── DELETE requests: BLOCKED (403 Forbidden)
└── Login:          ALLOWED (to view data / resolve issue)

User Experience:
├── Dashboard:      Visible with suspension banner
├── Reports:        Accessible (read-only)
├── Data Entry:     Disabled
├── File Upload:    Disabled
└── Settings:       View only

Background Jobs:
├── Billing:        Paused (no new charges)
├── Notifications:  Reduced to essential only
├── Analytics:      Still collected
└── Backups:        Still running
```

### Reactivation

```markdown
REACTIVATE TENANT FLOW:

  Request: ReactivateTenantRequest
  ├── tenant_id:      UUID
  ├── reason:         "Payment received - TXN-2024-0715"
  └── reactivated_by: admin@awo-erp.com

  Step 1: Validate
  ├── Tenant exists?              YES
  ├── Current status: SUSPENDED?  YES
  └── Proceed

  Step 2: Update Status
  ├── status: SUSPENDED → ACTIVE
  ├── updated_at: NOW()
  └── metadata: { "reactivation_reason": "Payment received",
                   "reactivated_at": "2024-07-22T09:00:00Z" }

  Step 3: Audit Log
  ├── Action:    TENANT_REACTIVATED
  ├── Actor:     admin@awo-erp.com
  └── Details:   "Payment received - TXN-2024-0715"

  Step 4: Notification
  ├── To: All tenant admin users
  ├── Subject: "Account Reactivated"
  └── Body: "Full access restored"

  Result: Full access restored, all operations enabled
```

### Example Timeline

```markdown
EXAMPLE: Highland Textiles Suspension & Recovery

Jun 1:  Invoice #INV-2024-0601 issued (KES 15,000)
Jun 15: Payment reminder sent
Jun 30: Payment still overdue (30 days)

Jul 1:  SUSPENDED
        Reason: "Payment overdue 30+ days"
        Users notified: admin@highlandtextiles.co.ke
        Impact: 8 users switched to read-only

Jul 5:  Highland Textiles contacts support
        "We had a banking issue, payment is being processed"

Jul 8:  Payment received (KES 15,000 + KES 500 late fee)

Jul 8:  REACTIVATED
        Reason: "Payment received - TXN ref #HT-2024-0708"
        Users notified: "Full access restored"
        Downtime: 7 days read-only

Jul 8:  All 8 users resume normal operations
```

---

Next: [Archiving & Data Retention](./14-archiving-and-data-retention.md)
