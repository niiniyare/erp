# internal/core/access — Reserved for v2.0 ABAC

**Status: NOT ACTIVE in v1.0**

All `.go` files in this package and its sub-packages carry `//go:build ignore` and are excluded from all builds and tests.

## What this package contains

A relational ABAC (Attribute-Based Access Control) layer including:
- Conditional access policies with time, location, and device conditions
- Access request and approval workflows
- Per-resource permission grants with optional expiry
- Permission cache

## Why it is gated

The IAM v1.0 architecture is **Casbin RBAC only**. All authorization decisions route through `authzService.Enforce()`. This ABAC layer is dead schema — the `Service` interface is defined but no implementation exists that is wired into the application.

Activating this code without full implementation, testing, and security review would introduce an uncontrolled authorization surface.

## v2.0 activation checklist (do not activate until ALL are met)

- [ ] Full service implementation for every interface in this package
- [ ] DB migrations 000404–000413 reviewed and applied
- [ ] Integration tests covering all ABAC scenarios
- [ ] Security review of conditional access evaluation logic
- [ ] IAM architecture doc updated to reflect dual RBAC+ABAC model
- [ ] `//go:build ignore` tags removed from all files in this package
- [ ] `internal/core/service.go` re-wired with access service initialization

## Related

- DB migrations reserved for this package: `000404`–`000413` (do not drop)
- Queries: `db/queries/policies.sql` (may contain access-module-only queries)
- IAM v1.0 enforcement: `internal/core/iam/service/authz.go`
