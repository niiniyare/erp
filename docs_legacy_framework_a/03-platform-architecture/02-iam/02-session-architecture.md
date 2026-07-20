> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Session Architecture
portal: 3 — Platform Architecture
section: 02-iam
audience: [architect, backend-engineer, tech-lead]
related:
  - "[IAM Overview](01-iam-overview.md)"
  - "[Authorization Model](03-authorization-model.md)"
---

# Session Architecture

## Session Lifecycle

```
POST /auth/login
    │
    ▼
SessionService.Login
    ├── Verify credentials (password hash)
    ├── Check tenant status (must be ACTIVE)
    ├── Load user roles + permissions
    ├── Load tenant feature flags + settings
    ├── Compute EntityScope
    ├── Build ResolvedSession
    ├── Serialize → store in Redis (TTL = 8h)
    └── Return opaque token (UUID → Redis key)

Every authenticated request
    │
    ▼
Authenticate middleware
    ├── Extract token from Authorization: Bearer <token>
    ├── SessionService.Resolve(token)
    │   ├── GET from Redis
    │   ├── If missing → ErrSessionExpired
    │   └── Deserialize → ResolvedSession
    └── Store in c.Locals(LocalsKeySession)

POST /auth/logout
    │
    ▼
SessionService.Revoke(token)
    └── DEL from Redis
```

## ResolvedSession Design

The session is **pre-computed at login** and cached in Redis. This means:

- No DB query on every request for roles/permissions
- Feature flags and settings are snapshotted at login
- Session invalidation (revoke) takes effect immediately

Trade-off: if roles change after login, the change takes effect only when the session expires or is revoked. For security-sensitive role removals, operators must revoke active sessions.

```go
type ResolvedSession struct {
    UserID      uuid.UUID
    TenantID    uuid.UUID
    EntityScope EntityScope
    Roles       []string            // e.g. ["contracts.reviewer", "finance.viewer"]
    Features    map[string]bool     // feature flag snapshot
    Settings    map[string]string   // tenant settings snapshot
    ExpiresAt   time.Time
}

// FeatureEnabled returns true if the flag is enabled for this session's tenant.
func (s *ResolvedSession) FeatureEnabled(key string) bool {
    return s.Features[key]
}

// SettingString returns a tenant setting or the provided default.
func (s *ResolvedSession) SettingString(key, def string) string {
    if v, ok := s.Settings[key]; ok {
        return v
    }
    return def
}

// ToPrincipal converts to IAM Principal for authz service calls.
func (s *ResolvedSession) ToPrincipal() Principal {
    return Principal{
        Subject: TenantSubject(s.UserID),
        Domain:  TenantDomain(s.TenantID),
    }
}
```

## Redis Session Storage

```
Key:   session:{token-uuid}
Value: JSON-encoded ResolvedSession
TTL:   8 hours (refreshed on each successful Resolve)
```

Sessions are stored as JSON blobs. The token is an opaque UUID — never a JWT (no client-side decoding, no signature to verify per request).

## Session Refresh

Each call to `Resolve` extends the TTL by another 8 hours (sliding expiry). A session expires only if unused for 8 consecutive hours.

```go
func (r *sessionRepo) Resolve(ctx context.Context, token string) (*ResolvedSession, error) {
    data, err := r.redis.GetEx(ctx, "session:"+token, 8*time.Hour).Bytes()
    if err == redis.Nil {
        return nil, iam.ErrSessionExpired
    }
    // ...
}
```

## What Session Does NOT Have

The session deliberately excludes:

- `Can(permission string) bool` method — use `authzSvc.Enforce` for permission checks
- `HasRole(role string) bool` method — roles are for authz service, not direct branching
- Raw permission list — permissions are evaluated by Casbin, not by iterating a list

This prevents business logic from branching on roles directly (fragile, bypasses policy engine).
