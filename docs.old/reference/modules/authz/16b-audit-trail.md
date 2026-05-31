[<-- Back to Index](README.md)

## Audit Trail

### Event-Driven Approach

| Approach | Verdict |
|---|---|
| DB triggers | Captures even direct DB changes; no app context (IP, user agent); hard to maintain |
| Service-level calls | Full control; easy to omit by mistake |
| **Event-driven (chosen)** | Zero coupling; centralised; async; durable outbox prevents loss |

IAM emits domain events after every successful state change. The Audit module subscribes. IAM never imports Audit.

Audit events for **failures** (failed login, permission denied on sensitive operations) are emitted on the failure path — they are the security record.

**Durable delivery**: events are written to a `domain_events` table in the same DB transaction as the operation. A background worker processes them. No event is lost even during downtime.

---

### IAM Domain Events

```go
// internal/platform/domain/events.go

UserCreated         { UserID, Email, UserType, TenantID, CreatedBy, OccurredAt }
UserSuspended       { UserID, SuspendedBy, Reason, OccurredAt }
UserDeactivated     { UserID, DeactivatedBy, OccurredAt }
SessionStarted      { UserID, SessionID, IP, UserAgent, OccurredAt }
SessionEnded        { UserID, SessionID, Reason, OccurredAt }
LoginFailed         { Email, IP, Reason, AttemptCount, OccurredAt }
MFAEnabled          { UserID, OccurredAt }
MFADisabled         { UserID, DisabledBy, OccurredAt }
PasswordChanged     { UserID, ChangedBy, OccurredAt }
RoleAssigned        { UserID, RoleID, RoleSlug, Domain, GrantedBy, ExpiresAt, OccurredAt }
RoleRevoked         { UserID, RoleID, RoleSlug, Domain, RevokedBy, Reason, OccurredAt }
PolicyAdded         { Subject, Domain, Object, Action, Effect, AddedBy, OccurredAt }
PolicyRemoved       { Subject, Domain, Object, Action, RemovedBy, OccurredAt }
FlagChanged         { TenantID, FlagKey, OldValue, NewValue, ChangedBy, OccurredAt }
SettingChanged      { TenantID, SettingKey, OldValue, NewValue, ChangedBy, OccurredAt }
```

---

### HTTP-Level Audit (AuditWrap)

The `AuditWrap` middleware records an HTTP-level trace for all mutating requests. This coarser layer runs alongside domain event audit:

```go
// Records: method, path, status, duration, IP, tenant_id, user_id
app.Use(middleware.AuditWrap(deps.Platform.Audit))
```

This captures requests that fail before reaching the service layer (e.g., 403 at the middleware level), which domain events do not capture.

---

Next: [Performance & Caching](./16-performance-and-caching.md)
