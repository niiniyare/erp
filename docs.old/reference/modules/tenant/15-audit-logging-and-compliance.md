[<-- Back to Index](README.md)

## Audit Logging & Compliance

### Overview

Every significant tenant operation is logged for compliance, debugging, and security monitoring. The provisioning service uses an `AuditLogger` interface that decouples logging from the core business logic.

### Audited Operations

```markdown
TENANT AUDIT EVENTS:

Provisioning:
├── TENANT_PROVISIONED       : New tenant created
├── TENANT_ACTIVATED         : Status changed to ACTIVE
├── TENANT_CONFIG_CREATED    : Default configuration applied
├── ADMIN_USER_CREATED       : First admin user set up
└── MODULES_ENABLED          : Modules activated for tenant

Status Changes:
├── TENANT_SUSPENDED         : Tenant suspended (with reason)
├── TENANT_REACTIVATED       : Tenant reactivated (with reason)
├── TENANT_ARCHIVED          : Tenant archived (with retention)
└── TENANT_DELETED           : Tenant soft-deleted

Configuration:
├── CONFIG_UPDATED           : Limits or settings changed
├── PLAN_UPGRADED            : Subscription plan upgraded
├── PLAN_DOWNGRADED          : Subscription plan downgraded
└── MODULES_CHANGED          : Module access modified

Bulk Operations:
├── BULK_OPERATION_STARTED   : Mass operation initiated
├── BULK_OPERATION_COMPLETED : Mass operation finished
└── BULK_OPERATION_FAILED    : Mass operation had failures

Security:
├── TENANT_CONTEXT_SET       : Tenant context established
├── TENANT_ACCESS_DENIED     : Access validation failed
├── LIMIT_EXCEEDED           : Resource limit hit
└── ISOLATION_VIOLATION      : Cross-tenant access attempt blocked
```

### Audit Log Structure

```markdown
AUDIT LOG ENTRY:

{
  "timestamp":    "2024-07-15T10:00:00Z",
  "event_type":   "TENANT_SUSPENDED",
  "tenant_id":    "ccc-333",
  "tenant_name":  "Highland Textiles",
  "actor_id":     "admin-uuid",
  "actor_name":   "admin@awo-erp.com",
  "actor_role":   "platform_admin",
  "details": {
    "reason":         "Payment overdue 30+ days",
    "previous_status": "ACTIVE",
    "new_status":      "SUSPENDED",
    "invoice_ref":     "INV-2024-0601"
  },
  "ip_address":   "203.0.113.50",
  "user_agent":   "AWO-Admin-Dashboard/1.0"
}
```

### Go Interface

```markdown
AUDIT LOGGER INTERFACE:

AuditLogger interface {
    LogProvisioningEvent(ctx, tenantID, event, details)
    LogStatusChange(ctx, tenantID, oldStatus, newStatus, reason, actor)
    LogConfigChange(ctx, tenantID, changes, actor)
    LogBulkOperation(ctx, operationID, operationType, results)
    LogSecurityEvent(ctx, tenantID, eventType, details)
}

NOTIFICATION SENDER INTERFACE:

NotificationSender interface {
    SendWelcomeEmail(ctx, tenantID, adminEmail, tenantName)
    SendSuspensionNotice(ctx, tenantID, reason)
    SendReactivationNotice(ctx, tenantID)
    SendArchiveNotice(ctx, tenantID, retentionDays)
}
```

### Compliance Requirements

```markdown
COMPLIANCE MATRIX:

Requirement                  Implementation
──────────────────────────────────────────────────────
Audit Trail                  Every status change logged
Data Isolation               RLS + enforce_tenant_isolation()
Access Control               RBAC with three database roles
Data Retention               Configurable per tenant/plan
Soft Delete                  deleted_at preserves records
Change Tracking              updated_at on every table
Actor Attribution            actor_id on all operations
Reason Documentation         reason field on status changes
Timestamp Accuracy           TIMESTAMPTZ (timezone-aware)
Immutable Audit Logs         Append-only audit entries
```

### Audit Query Examples

```markdown
COMMON AUDIT QUERIES:

1. All status changes for a tenant:
   Filter: tenant_id = X, event_type IN (SUSPENDED, REACTIVATED, ARCHIVED)
   Sort:   timestamp DESC

2. All operations by a platform admin:
   Filter: actor_id = Y
   Sort:   timestamp DESC

3. Failed access attempts (security):
   Filter: event_type = TENANT_ACCESS_DENIED
   Sort:   timestamp DESC
   Alert:  If count > threshold in time window

4. Bulk operation history:
   Filter: event_type LIKE 'BULK_OPERATION_%'
   Join:   tenant_bulk_operations for details

5. Provisioning report (monthly):
   Filter: event_type = TENANT_PROVISIONED
   Range:  Current month
   Aggregate: Count, group by plan_type
```

---

Next: [Caching & Performance](./16-caching-and-performance.md)
