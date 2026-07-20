> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Architecture Principles
portal: 3 — Platform Architecture
section: 00-overview
audience: [architect, backend-engineer, tech-lead]
related:
  - "[System Overview](01-system-overview.md)"
  - "[Multi-Tenancy Model](../01-multi-tenancy/01-tenancy-model.md)"
---

# Architecture Principles

These principles govern every design decision in AwoERP. When in doubt, refer back to them.

## 1. Tenant Isolation is Non-Negotiable

Every piece of tenant data is protected at the database layer via PostgreSQL Row-Level Security. Application-level tenant filtering is a second line of defence — never the only line.

**Consequence**: `current_setting('app.tenant_id')` must never use `missing_ok = true`. A missing GUC is a programming error and must fail loudly.

## 2. Fail Closed on Authorization

When an authorization check fails or errors, access is denied. The system never defaults to permit.

```go
allowed, err := authzSvc.Enforce(ctx, req)
if err != nil || !allowed {
    return nil, iam.ErrForbidden  // deny on error AND on deny
}
```

## 3. Instrumentation Never Fails the Request

Audit writes, event publishes, notifications, and metric increments run asynchronously and never propagate errors to the HTTP response. Business operations must not be blocked by observability infrastructure.

## 4. Explicit Over Magic

- Wire over `init()` registration
- `errors.Is` over type switches
- SQLC over reflection-based ORMs
- Explicit `context.Context` propagation — never stored in structs

## 5. Money is Never a Float

All monetary values use `github.com/shopspring/decimal`. Float64 is prohibited for amounts, rates, quantities, or any value requiring exact arithmetic.

## 6. Optimistic Locking Everywhere

Every mutable entity has a `version integer NOT NULL DEFAULT 1`. Updates include `WHERE version = @version` and return the new version. This prevents lost updates in concurrent sessions without pessimistic locks.

## 7. Soft Delete, Not Hard Delete

Production data is never hard-deleted. `deleted_at timestamptz` (nullable) marks deletion. Hard deletes are reserved for maintenance scripts with explicit approval.

## 8. Async Side Effects

After a successful write, side effects (notifications, events, audit) run in goroutines with `context.Background()` + 5-second timeout + `recover()`. The HTTP response returns before side effects complete.

## 9. Module Boundary Respect

Modules communicate through:
- **Interfaces** (injected via Wire) — for synchronous calls within the process
- **Events** (published on the bus) — for asynchronous cross-module reactions
- **Never** by importing another module's repository or internal types directly

## 10. Deterministic Builds

- Wire generates `wire_gen.go` at build time — never edited manually
- SQLC generates typed DB access at build time — never written by hand
- Migrations are append-only — never modified after being applied to any environment
