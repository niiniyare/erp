[<-- Back to Index](README.md)

## Bulk Operations

### Overview

Platform administrators can perform operations on multiple tenants simultaneously. Bulk operations are tracked with full audit trails, progress monitoring, and error handling.

### Supported Operation Types

```markdown
BULK OPERATION TYPES:

1. SUSPEND
   Purpose: Mass suspend tenants (e.g., non-payment batch)
   Parameters: { "reason": "Payment overdue 30+ days" }

2. REACTIVATE
   Purpose: Mass reactivate suspended tenants
   Parameters: { "reason": "Payment received" }

3. ARCHIVE
   Purpose: Mass archive inactive tenants
   Parameters: { "reason": "Inactive 12+ months",
                  "retention_days": 2555 }

4. UPDATE_LIMITS
   Purpose: Mass update resource limits
   Parameters: { "max_users": 100,
                  "storage_quota_mb": 20480 }

5. UPDATE_FEATURES
   Purpose: Mass enable/disable features
   Parameters: { "enable_modules": ["inventory"],
                  "disable_modules": ["manufacturing"] }
```

### Bulk Operation Flow

```markdown
BULK OPERATION LIFECYCLE:

  Admin submits request
       │
       ▼
  ┌──────────────────┐
  │ Create Operation  │  Status: IN_PROGRESS
  │ Record (master)   │  Total: N tenants
  └────────┬─────────┘
           │
           ▼
  ┌──────────────────┐
  │ Create Result     │  Status: PENDING (each)
  │ Records (per      │  One row per tenant
  │ tenant)           │
  └────────┬─────────┘
           │
     ┌─────┼─────┐
     ▼     ▼     ▼
  [Process each tenant]
     │     │     │
     ▼     ▼     ▼
  COMPLETED FAILED SKIPPED
     │     │     │
     └─────┼─────┘
           │
           ▼  (trigger updates counts automatically)
  ┌──────────────────┐
  │ Final Status:     │
  │ COMPLETED         │ ← All succeeded
  │ FAILED            │ ← All failed
  │ PARTIAL_SUCCESS   │ ← Mixed results
  │ CANCELLED         │ ← Admin cancelled
  └──────────────────┘
```

### Real-World Example

```markdown
EXAMPLE: Bulk Suspend Non-Paying Tenants

REQUEST:
  Actor:     admin@awo-erp.com (Platform Admin)
  Type:      SUSPEND
  Tenants:   3 tenants with overdue payments
  Reason:    "Payment overdue 30+ days - batch Feb 2024"

OPERATION CREATED:
  ID:              op-uuid-001
  Status:          IN_PROGRESS
  Total Tenants:   3

PROCESSING:

  Tenant 1: Highland Textiles (ID: ccc-333)
    Previous Status: ACTIVE
    Action:          Set status → SUSPENDED
    Result:          COMPLETED
    Message:         "Suspended successfully"

  Tenant 2: Lake Auto Parts (ID: ddd-444)
    Previous Status: SUSPENDED
    Action:          Already suspended
    Result:          SKIPPED
    Message:         "Tenant already suspended"

  Tenant 3: Metro Supplies (ID: eee-555)
    Previous Status: ACTIVE
    Action:          Set status → SUSPENDED
    Result:          COMPLETED
    Message:         "Suspended successfully"

FINAL STATUS:
  Operation:       COMPLETED
  Successful:      2
  Failed:          0
  Skipped:         1
  Duration:        3 seconds
```

### Monitoring Bulk Operations

The `get_bulk_operation_summary()` function provides real-time monitoring:

```markdown
SUMMARY QUERY RESULT:

  operation_id:     op-uuid-001
  operation_type:   SUSPEND
  status:           COMPLETED
  total_tenants:    3
  successful_count: 2
  failed_count:     0
  in_progress_count: 0
  duration_seconds: 3
```

### Automatic Count Updates

Counts are maintained automatically via a database trigger:

```markdown
TRIGGER: tenant_bulk_operation_results_update_counts

Fires: AFTER INSERT, UPDATE, or DELETE on results table

Logic:
1. Count COMPLETED results → successful_count
2. Count FAILED results → failed_count
3. Count PENDING + PROCESSING → in_progress_count
4. Determine overall status:
   ├── in_progress > 0     → IN_PROGRESS
   ├── failed = 0          → COMPLETED
   ├── successful = 0      → FAILED
   └── mixed               → PARTIAL_SUCCESS
5. Set completed_at when in_progress = 0
```

### Go Service Interface

```markdown
PROVISIONING SERVICE METHODS:

BulkTenantOperation(ctx, request) → (*BulkOperationResult, error)
├── Creates master operation record
├── Creates individual result records (PENDING)
├── Processes each tenant in sequence
├── Updates result status per tenant
├── Returns final summary
└── Triggers count update automatically

Request Structure:
  BulkTenantOperationRequest {
    OperationType:  "SUSPEND" | "REACTIVATE" | "ARCHIVE" | ...
    TenantIDs:      []UUID
    Parameters:     map[string]interface{}
    ActorID:        UUID
    ActorName:      string
  }
```

---

Next: [Tenant Suspension & Reactivation](./13-tenant-suspension-and-reactivation.md)
