---
title: Audit Implementation
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Audit Overview](01-audit-overview.md)"
  - "[Service Layer](../06-service-layer/01-service-overview.md)"
  - "[Observability Guide](../19-observability/01-observability-guide.md)"
---

# Audit Implementation

## AuditService Implementation

```go
// internal/shared/audit/service.go
type pgAuditService struct {
    store  *db.Store
    logger *slog.Logger
}

func NewAuditService(store *db.Store, logger *slog.Logger) Service {
    return &pgAuditService{
        store:  store,
        logger: logger.With("module", "audit"),
    }
}

func (s *pgAuditService) Log(ctx context.Context, entry Entry) error {
    beforeJSON, _ := toJSON(entry.Before)
    afterJSON, _ := toJSON(entry.After)

    return s.store.WithTenant(ctx, entry.TenantID, func(tx *pgx.Tx) error {
        q := sqlc.New(tx)
        return q.InsertAuditLog(ctx, sqlc.InsertAuditLogParams{
            TenantID:     entry.TenantID,
            UserID:       pgtype.UUID{Bytes: entry.UserID, Valid: entry.UserID != uuid.Nil},
            Action:       entry.Action,
            ResourceType: entry.ResourceType,
            ResourceID:   pgtype.UUID{Bytes: entry.ResourceID, Valid: entry.ResourceID != uuid.Nil},
            BeforeState:  beforeJSON,
            AfterState:   afterJSON,
            IpAddress:    pgtype.Text{String: entry.IPAddress, Valid: entry.IPAddress != ""},
            UserAgent:    pgtype.Text{String: entry.UserAgent, Valid: entry.UserAgent != ""},
            RequestID:    pgtype.Text{String: entry.RequestID, Valid: entry.RequestID != ""},
        })
    })
}

func toJSON(v interface{}) ([]byte, error) {
    if v == nil {
        return nil, nil
    }
    return json.Marshal(v)
}
```

## SQLC Query

```sql
-- name: InsertAuditLog :exec
INSERT INTO audit_log (
    tenant_id, user_id, action, resource_type, resource_id,
    before_state, after_state, ip_address, user_agent, request_id
) VALUES (
    @tenant_id, @user_id, @action, @resource_type, @resource_id,
    @before_state, @after_state, @ip_address, @user_agent, @request_id
);
```

## Extracting IP and User Agent from Context

The handler sets IP and user agent into the context before calling the service:

```go
// internal/server/middleware/context.go
type requestMetaKey struct{}

type RequestMeta struct {
    IPAddress string
    UserAgent string
    RequestID string
}

func WithRequestMeta(c *fiber.Ctx) context.Context {
    meta := RequestMeta{
        IPAddress: c.IP(),
        UserAgent: c.Get("User-Agent"),
        RequestID: c.GetRespHeader("X-Request-ID"),
    }
    return context.WithValue(c.UserContext(), requestMetaKey{}, meta)
}

func RequestMetaFromContext(ctx context.Context) RequestMeta {
    meta, _ := ctx.Value(requestMetaKey{}).(RequestMeta)
    return meta
}
```

In the handler:

```go
func (h *ContractHandler) Create(c *fiber.Ctx) error {
    sess := middleware.SessionFrom(c)
    ctx := middleware.WithRequestMeta(c)   // inject IP, UA, request ID

    contract, err := h.svc.Create(ctx, sess, req)
    // ...
}
```

In the service's audit call:

```go
meta := middleware.RequestMetaFromContext(ctx)
go s.recordAuditAsync(audit.Entry{
    TenantID:     sess.TenantID,
    UserID:       sess.UserID,
    Action:       "contracts.contract.create",
    ResourceType: "contract",
    ResourceID:   contract.ID,
    After:        contract,
    IPAddress:    meta.IPAddress,
    UserAgent:    meta.UserAgent,
    RequestID:    meta.RequestID,
})
```

## Audit Log API Endpoints

Expose audit log via the API for compliance reporting:

```
GET /api/v1/audit/log
  ?user_id={uuid}
  ?resource_type=contract
  ?resource_id={uuid}
  ?action=contracts.contract.approve
  ?date_from=2025-01-01
  ?date_to=2025-01-31
  &page=1&page_size=50
```

**Permission required**: `audit.log.read` — typically granted to `finance.auditor` and `iam.admin` roles only.

## Data Retention

Audit logs are retained indefinitely by default. For compliance:
- Never delete audit log entries
- For GDPR "right to erasure": do NOT delete audit entries — pseudonymize by removing PII from `before_state`/`after_state` but keep the event record

```sql
-- GDPR pseudonymization (not deletion)
UPDATE audit_log
SET before_state = jsonb_set(before_state, '{email}', '"[redacted]"'),
    after_state  = jsonb_set(after_state,  '{email}', '"[redacted]"')
WHERE user_id = $1;
```

## No-Mock Requirement

Audit logs must write to real PostgreSQL in integration tests. Never mock the audit service in tests that verify audit behavior:

```go
// ❌ Wrong — doesn't verify audit actually wrote to DB
func TestCreate_AuditLogged(t *testing.T) {
    mockAudit := &mockAuditService{}
    svc := NewContractService(..., mockAudit)
    svc.Create(...)
    assert.Len(t, mockAudit.logged, 1)  // only verifies call, not DB write
}

// ✅ Correct — verify the DB row
func TestCreate_AuditLogged_Integration(t *testing.T) {
    // Uses real DB
    svc := setupRealService(t)
    contract, _ := svc.Create(...)

    // Wait for async goroutine
    time.Sleep(100 * time.Millisecond)

    var count int
    db.QueryRow("SELECT COUNT(*) FROM audit_log WHERE resource_id = $1", contract.ID).Scan(&count)
    assert.Equal(t, 1, count)
}
```
