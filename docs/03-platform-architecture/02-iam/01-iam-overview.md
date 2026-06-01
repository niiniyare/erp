---
title: IAM Overview
portal: 3 — Platform Architecture
section: 02-iam
audience: [architect, backend-engineer, tech-lead]
related:
  - "[Session Architecture](02-session-architecture.md)"
  - "[Authorization Model](03-authorization-model.md)"
  - "[Tenancy Model](../01-multi-tenancy/01-tenancy-model.md)"
---

# IAM Overview

IAM (Identity and Access Management) handles authentication, session lifecycle, and authorization. It is a platform module — all other modules depend on it via interfaces, never by direct import.

## Responsibilities

| Concern | Component |
|---------|----------|
| User authentication | `SessionService.Login` |
| Token validation | `SessionService.Resolve` |
| Session storage | Redis (TTL-based) |
| Permission check | `AuthzService.Enforce` (Casbin) |
| Role management | `RoleService` |
| Permission assignment | `PolicyService` |

## IAM Interfaces (used by all modules)

```go
// internal/core/iam/interfaces.go

// SessionService manages session lifecycle.
type SessionService interface {
    Login(ctx context.Context, req LoginRequest) (*ResolvedSession, string, error)
    Resolve(ctx context.Context, token string) (*ResolvedSession, error)
    Revoke(ctx context.Context, token string) error
}

// AuthzService checks permissions.
type AuthzService interface {
    Enforce(ctx context.Context, req Request) (bool, error)
    AddPolicy(ctx context.Context, p Policy) error
    RemovePolicy(ctx context.Context, p Policy) error
}
```

## Request / Principal

```go
// Request is the authorization check input.
type Request struct {
    Subject string // "role:{name}" or "user:{id}"
    Domain  string // "tenant:{tenantID}"
    Object  string // "contracts.contract.create" or "tenants/{id}/contracts/{id}"
    Action  string // "allow" (route-level) or the specific action (service-level)
}

// Principal is extracted from ResolvedSession.ToPrincipal().
type Principal struct {
    Subject string
    Domain  string
}
```

## Package Layout

```
internal/core/iam/
├── domain/
│   ├── session.go          # ResolvedSession, EntityScope
│   ├── principal.go        # Principal, TenantSubject, TenantDomain
│   ├── errors.go           # ErrForbidden, ErrSessionExpired, ErrSessionInvalid
│   └── locals.go           # LocalsKeySession constant
├── service/
│   ├── session_service.go  # Login, Resolve, Revoke
│   └── authz_service.go    # Casbin wrapper
├── repository/
│   └── session_repo.go     # Redis-backed session storage
└── iam.go                  # ProviderSet
```

## Error Sentinel Values

```go
var (
    ErrForbidden      = errors.New("forbidden")
    ErrSessionExpired = errors.New("session expired")
    ErrSessionInvalid = errors.New("session invalid")
    ErrUnauthorized   = errors.New("unauthorized")
)
```

All modules use `iam.ErrForbidden` — never define their own forbidden error.
