---
title: Audit Logging Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Observability](../../../03-platform-architecture/05-observability/01-observability-overview.md)"
  - "[Service Layer](../06-service-layer/01-service-overview.md)"
  - "[Notifications](../10-notifications/01-notifications-overview.md)"
---

# Audit Logging Overview

## Purpose

Audit logging records who did what to which resource and when. Unlike application logs (operational), audit logs are:
- **Business records** — legally and compliance-relevant
- **Append-only** — never modified or deleted
- **Structured** — queryable by tenant, user, resource, action, time

## When to Write Audit Logs

Write an audit log entry for every **state-changing operation** on business data:

| Operation | Audit? |
|-----------|--------|
| Create contract | Yes |
| Update contract | Yes |
| Submit contract | Yes |
| Approve contract | Yes |
| Read contract (GET) | No |
| List contracts | No |
| User login | Yes (IAM module) |
| Role assignment | Yes (IAM module) |
| Feature flag change | Yes (tenant admin) |

## Audit Log Schema

```sql
CREATE TABLE audit_log (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      uuid NOT NULL REFERENCES tenants(id),
    user_id        uuid,             -- NULL for system actions
    action         text NOT NULL,    -- e.g. "contracts.contract.create"
    resource_type  text NOT NULL,    -- e.g. "contract"
    resource_id    uuid,             -- ID of affected record
    before_state   jsonb,            -- NULL for creates
    after_state    jsonb,            -- NULL for deletes
    ip_address     text,
    user_agent     text,
    request_id     text,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_log_tenant_id_created_at ON audit_log (tenant_id, created_at DESC);
CREATE INDEX audit_log_resource ON audit_log (tenant_id, resource_type, resource_id);
CREATE INDEX audit_log_user_id ON audit_log (tenant_id, user_id, created_at DESC);
```

`before_state` and `after_state` are JSONB snapshots of the record before and after the change.

## AuditService Interface

```go
// internal/shared/audit/service.go
type Service interface {
    Log(ctx context.Context, entry Entry) error
}

type Entry struct {
    TenantID     uuid.UUID
    UserID       uuid.UUID       // zero value for system actions
    Action       string          // "module.resource.verb"
    ResourceType string
    ResourceID   uuid.UUID
    Before       interface{}     // pointer to domain struct, or nil
    After        interface{}     // pointer to domain struct, or nil
    IPAddress    string          // from request context
    UserAgent    string
    RequestID    string
}
```

## Async Pattern

Audit writes are async — they never block the HTTP response:

```go
func (s *ContractService) Create(ctx context.Context, sess ResolvedSession, req CreateParams) (*Contract, error) {
    contract, err := s.repo.Create(ctx, sess.TenantID, req)
    if err != nil {
        return nil, err
    }

    // Audit async — never block
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        defer func() {
            if r := recover(); r != nil {
                s.logger.Error("audit log panic", "recover", r)
            }
        }()
        s.audit.Log(ctx, audit.Entry{
            TenantID:     sess.TenantID,
            UserID:       sess.UserID,
            Action:       "contracts.contract.create",
            ResourceType: "contract",
            ResourceID:   contract.ID,
            Before:       nil,
            After:        contract,
            RequestID:    requestIDFromContext(ctx),
        })
    }()

    return contract, nil
}
```

Rationale: Audit log failure must not fail the business operation. The goroutine uses `context.Background()` so it completes even after the HTTP request context is cancelled.

## Action Name Conventions

Action names follow `{module}.{resource}.{verb}`:

```
contracts.contract.create
contracts.contract.update
contracts.contract.delete
contracts.contract.submit
contracts.contract.approve
contracts.contract.activate
contracts.contract.terminate

finance.account.create
finance.transaction.post
finance.transaction.void

iam.user.create
iam.user.deactivate
iam.role.assign
iam.session.revoke

tenant.tenant.activate
tenant.tenant.suspend
tenant.feature_flag.update
```

## Before/After State

For updates, capture before and after:

```go
func (s *ContractService) Update(ctx context.Context, sess ResolvedSession, id uuid.UUID, req UpdateParams) (*Contract, error) {
    // Get current state (before)
    before, err := s.repo.GetByID(ctx, id, sess.TenantID)
    if err != nil {
        return nil, err
    }

    after, err := s.repo.Update(ctx, id, sess.TenantID, req)
    if err != nil {
        return nil, err
    }

    go s.recordAuditAsync(sess, "contracts.contract.update", after.ID, before, after)
    return after, nil
}

func (s *ContractService) recordAuditAsync(sess ResolvedSession, action string, id uuid.UUID, before, after interface{}) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    defer func() { recover() }()
    s.audit.Log(ctx, audit.Entry{
        TenantID:     sess.TenantID,
        UserID:       sess.UserID,
        Action:       action,
        ResourceType: "contract",
        ResourceID:   id,
        Before:       before,
        After:        after,
    })
}
```

## Querying Audit Logs

```sql
-- All changes to a specific contract
SELECT * FROM audit_log
WHERE tenant_id = $1
  AND resource_type = 'contract'
  AND resource_id = $2
ORDER BY created_at;

-- All actions by a user today
SELECT * FROM audit_log
WHERE tenant_id = $1
  AND user_id = $2
  AND created_at >= current_date
ORDER BY created_at DESC;

-- All approval actions in a time range
SELECT * FROM audit_log
WHERE tenant_id = $1
  AND action = 'contracts.contract.approve'
  AND created_at BETWEEN $2 AND $3
ORDER BY created_at DESC;
```
