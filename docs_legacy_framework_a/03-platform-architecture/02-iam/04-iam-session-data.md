> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Session Data Reference
portal: 3 — Platform Architecture
section: 03-platform-architecture
audience: [backend-engineer, architect]
related:
  - "[IAM Overview](01-iam-overview.md)"
  - "[Session Architecture](02-session-architecture.md)"
  - "[Auth Middleware](../../04-backend-engineering/00-module-development-guide/16-middleware-chain/02-auth-middleware.md)"
---

# Session Data Reference

## ResolvedSession Fields

```go
// internal/core/iam/session.go
type ResolvedSession struct {
    // Identity
    UserID   uuid.UUID
    TenantID uuid.UUID
    Email    string
    Name     string

    // Authorization
    Roles       []string            // role names, e.g. ["contracts.editor", "finance.viewer"]
    EntityScope EntityScope         // data visibility scope

    // Feature flags (pre-loaded from tenant_feature_overrides)
    FeatureFlags map[string]bool     // "contracts.bulk_import" → true/false

    // Tenant settings (pre-loaded from tenant_configurations)
    Settings map[string]string       // "finance.default_currency" → "USD"

    // Session metadata
    Token     uuid.UUID
    ExpiresAt time.Time
    CreatedAt time.Time
}
```

## Methods

```go
// Convert to Principal for Casbin authorization
func (s ResolvedSession) ToPrincipal() Principal {
    return Principal{
        UserID:   s.UserID,
        TenantID: s.TenantID,
        Roles:    s.Roles,
    }
}

// Check feature flag (returns false if flag not found)
func (s ResolvedSession) FeatureEnabled(name string) bool {
    return s.FeatureFlags[name]
}

// Get tenant setting with fallback
func (s ResolvedSession) SettingString(key, defaultVal string) string {
    if v, ok := s.Settings[key]; ok {
        return v
    }
    return defaultVal
}

// SettingBool — parse setting as bool
func (s ResolvedSession) SettingBool(key string, defaultVal bool) bool {
    v, ok := s.Settings[key]
    if !ok {
        return defaultVal
    }
    b, err := strconv.ParseBool(v)
    if err != nil {
        return defaultVal
    }
    return b
}

// EntityID returns entity ID if scope is entity or subtree, nil if all
func (s ResolvedSession) EntityID() *uuid.UUID {
    if s.EntityScope.Type == ScopeAll {
        return nil
    }
    return &s.EntityScope.EntityID
}
```

## What ResolvedSession Does NOT Have

| Method | Why not on session |
|--------|-------------------|
| `Can(permission)` | Permission check requires Casbin — call `authzSvc.Can(ctx, sess.ToPrincipal(), perm)` |
| `HasRole(role)` | Coupling to role names — check permissions, not roles |
| `IsAdmin()` | No concept of "admin" — check specific permissions |
| `GetPermissions()` | Not stored in session — computed at check time by Casbin |

## EntityScope Types

```go
type EntityScopeType string

const (
    ScopeAll     EntityScopeType = "all"      // full tenant visibility
    ScopeSubtree EntityScopeType = "subtree"  // entity + descendants
    ScopeEntity  EntityScopeType = "entity"   // single entity only
)

type EntityScope struct {
    Type     EntityScopeType
    EntityID uuid.UUID   // zero value when Type == ScopeAll
}
```

## How to Access Session in a Handler

```go
// middleware.SessionFrom panics if session is missing
// This is intentional — missing session means the Authenticate middleware was bypassed
func (h *ContractHandler) Create(c *fiber.Ctx) error {
    sess := middleware.SessionFrom(c)  // type: iam.ResolvedSession
    // ...
}
```

Session is stored in Fiber's locals under key `"resolved_session"`. `SessionFrom` reads it:

```go
func SessionFrom(c *fiber.Ctx) iam.ResolvedSession {
    sess, ok := c.Locals("resolved_session").(iam.ResolvedSession)
    if !ok {
        panic("resolved_session not set — Authenticate middleware missing on route")
    }
    return sess
}
```

The panic is intentional: missing session on a protected route indicates a wiring bug, not a runtime error. It will be caught by Fiber's recover middleware and returned as 500, logged with stack trace.

## Session Storage Format

Redis key: `session:{token-uuid}`
Value: JSON-encoded `ResolvedSession`
TTL: 24 hours, sliding (reset on each authenticated request)

```json
{
  "user_id": "bbbbbbbb-0000-0000-0000-000000000001",
  "tenant_id": "aaaaaaaa-0000-0000-0000-000000000001",
  "email": "jane@acme.com",
  "name": "Jane Smith",
  "roles": ["contracts.editor", "finance.viewer"],
  "entity_scope": {
    "type": "all",
    "entity_id": "00000000-0000-0000-0000-000000000000"
  },
  "feature_flags": {
    "contracts.bulk_import": true,
    "finance.advanced_reporting": false
  },
  "settings": {
    "finance.default_currency": "USD",
    "contracts.approval_required": "true"
  },
  "token": "550e8400-e29b-41d4-a716-446655440000",
  "expires_at": "2025-01-16T10:30:00Z",
  "created_at": "2025-01-15T10:30:00Z"
}
```

## Invalidation Triggers

| Trigger | Invalidation method |
|---------|-------------------|
| User logout | Delete specific token |
| Role assignment change | `InvalidateUserSessions(userID)` |
| EntityScope change | `InvalidateUserSessions(userID)` |
| Feature flag change | `InvalidateTenantSessions(tenantID)` — all users must re-login |
| Tenant suspended | `InvalidateTenantSessions(tenantID)` |
| User deactivated | `InvalidateUserSessions(userID)` |
| Password change | `InvalidateUserSessions(userID)` |
