> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

[<-- Back to Index](README.md)

## IAM Contract Adoption Audit

> **Audit date:** 2026-05-18
> **Auditor:** Automated static analysis + manual review
> **Scope:** All non-IAM Go packages under `internal/`
> **Contract baseline:** `internal/core/iam/contract/` (introduced 2026-05-18)

---

### Audit criteria

| Code pattern | Verdict | Reason |
|---|---|---|
| `import "awo.so/internal/core/iam"` in non-IAM package | **CRITICAL** | Couples module to IAM internals |
| Field/param of type `iam.Service` or `iam.SessionService` | **HIGH** | Service spreads IAM dependency through call graph |
| Active `.Enforce(` call outside `middleware.Authorize` | **CRITICAL** | Bypasses single-path enforcement |
| TODO comment planning direct `.Enforce(` | **CRITICAL** | Documents planned bypass — must be deleted |
| `CapabilityContext.Permissions` / `.Features` used for access decisions | **CRITICAL** | Shadow permission system bypasses Casbin |
| `featureflag.Service` injected and called in domain service | **LOW** | Feature flags must come from `SessionContext.FeatureEnabled` |
| `ctx.Value("tenant_id")` with raw string key | **MEDIUM** | Brittle; use `shared.GetTenantID(ctx)` or `contract.FromContext` |
| `import "awo.so/internal/core/iam/contract"` | **PASS** | Correct consumer import path |
| No IAM import, reads only `context.Context` | **PASS** | Fully clean |

---

### Module-by-module findings

---

#### `internal/core/finance/service.go` — FAIL (CRITICAL + HIGH + MEDIUM + LOW)

| Line | Pattern | Severity |
|---|---|---|
| 11 | `import "awo.so/internal/core/iam"` | CRITICAL |
| 36 | `iamService iam.Service` struct field | HIGH |
| 52 | `iamService iam.Service` constructor param | HIGH |
| 38 | `featureFlagService featureflag.Service` field | LOW |
| 81 | `ctx.Value("tenant_id").(uuid.UUID)` — raw string key | MEDIUM |
| 90 | `ctx.Value("entity_id").(uuid.UUID)` — raw string key | MEDIUM |

Migration required before `finance` module can be wired.

---

#### `internal/core/finance/service/account.go` — FAIL (CRITICAL + HIGH + LOW)

| Line(s) | Pattern | Severity |
|---|---|---|
| 10 | `import "awo.so/internal/core/iam"` | CRITICAL |
| 8 | `import "awo.so/internal/core/featureflag"` | LOW |
| 92 | `iamService iam.Service` struct field | HIGH |
| 101 | `iamService iam.Service` constructor param | HIGH |
| 93 | `featureFlagService featureflag.Service` field | LOW |
| 102 | `featureFlagService featureflag.Service` constructor param | LOW |
| 159–164 | TODO comment planning direct `s.iamService.Enforce(...)` call | **CRITICAL** |
| 177–188 | `s.featureFlagService.IsEnabled(ctx, ...)` — active direct call | LOW |

The TODO comment (lines 159–164) documents an intended bypass of `middleware.Authorize`.
This is **architectural debt that must be deleted**, not implemented.
Feature flag evaluation (lines 177–188) must move to the handler layer via `SessionContext.FeatureEnabled`.

---

#### `internal/shared/context.go` — FAIL (CRITICAL)

| Lines | Pattern | Severity |
|---|---|---|
| 104–113 | `CapabilityContext` struct with `Permissions map[string]bool`, `Features map[string]bool`, `Modules map[string]bool` | CRITICAL |
| 116–124 | `WithCapabilityContext` / `GetCapabilityContext` helpers | CRITICAL |

`CapabilityContext` is a shadow authorization system. If any code reads
`CapabilityContext.Permissions` to decide whether an action is allowed, it bypasses
Casbin entirely. The struct must be audited for live usage and removed or de-armed.

> **Note:** No call site was found reading `CapabilityContext.Permissions` for access
> control during this audit. The struct exists but may be unused dead code.
> Confirm with `grep -r "CapabilityContext" internal/` before deleting.

---

#### `internal/core/finance/service/transaction.go` — PASS

No IAM import, no `iam.Service` field, no `.Enforce(` call.

---

#### `internal/core/finance/service/transaction_entry.go` — PASS

No IAM import or forbidden patterns.

---

#### `internal/core/finance/service/settings_helper.go` — PASS

No IAM import or forbidden patterns.

---

#### `internal/core/finance/service/anomaly_detector.go` — PASS

No IAM import or forbidden patterns.

---

#### `internal/core/audit/` — PASS

No IAM import. Does not evaluate permissions.

---

#### `internal/core/entity/` — PASS

No IAM import. Does not evaluate permissions.

---

#### `internal/core/featureflag/service.go` — PASS

No IAM import. Uses `tenant.Service` only. Does not evaluate permissions.
> Note: `featureflag.Service` itself is clean; the violation is in modules that inject
> it into domain services to gate business logic instead of using `SessionContext`.

---

#### `internal/core/settings/` — PASS

No IAM import. Pure config/read operations.

---

#### `internal/core/tenant/` — PASS

No IAM import. Does not evaluate permissions.

---

#### `internal/core/notification/` — PASS

No IAM import. Does not evaluate permissions.

---

#### `internal/api/middleware/` — PASS (post-fix state)

`middleware.Authorize` is the **sole** location where `Enforce()` is called.
All other middleware files are clean. `splitPermission` computed once at registration.
`cfg.AuthzService == nil` returns 500, not 403.

---

#### `internal/core/buy/`, `internal/core/sell/`, `internal/core/inventory/` — NOT YET IMPLEMENTED

These modules do not exist. When created they must follow the contract from day one:
- Import only `awo.so/internal/core/iam/contract`
- Never import `awo.so/internal/core/iam` directly
- Never inject `iam.Service` or `featureflag.Service` into domain services
- Read identity from `contract.FromContext(ctx)` only

---

#### `internal/api/handlers/`, `internal/api/routes/` — NOT YET IMPLEMENTED

No handler or route files found at audit time. When created they must follow
the patterns in `docs/reference/modules/iam/14-how-other-packages-use-authz.md`.

---

### Violation summary

| File | Status | Highest severity |
|---|---|---|
| `internal/core/finance/service.go` | **FAIL** | CRITICAL |
| `internal/core/finance/service/account.go` | **FAIL** | CRITICAL |
| `internal/shared/context.go` | **FAIL** | CRITICAL |
| `internal/core/finance/service/transaction.go` | PASS | — |
| `internal/core/finance/service/transaction_entry.go` | PASS | — |
| `internal/core/finance/service/settings_helper.go` | PASS | — |
| `internal/core/finance/service/anomaly_detector.go` | PASS | — |
| `internal/core/audit/` | PASS | — |
| `internal/core/entity/` | PASS | — |
| `internal/core/featureflag/service.go` | PASS | — |
| `internal/core/settings/` | PASS | — |
| `internal/core/tenant/` | PASS | — |
| `internal/core/notification/` | PASS | — |
| `internal/api/middleware/` | PASS | — |

---

### Contract Enforcement Migration Plan

#### Phase 1 — Delete planned bypass (immediate, blocking)

**Target:** `internal/core/finance/service/account.go` lines 159–164

Delete the TODO comment that plans a direct `Enforce()` call. There is no version of
this TODO that is correct. Authorization is `middleware.Authorize`'s job; the service
must never call `Enforce()` regardless of whether a principal is in context.

```go
// DELETE these lines entirely:
// TODO(authz): enforce finance.accounts.create via iam.Service.Enforce() once
// the session principal is wired into ctx. Example:
//   if userID, ok := shared.GetUserID(ctx); ok {
//       ok, _ := s.iamService.Enforce(ctx, iam.Request{...})
//       if !ok { return nil, errors.NewBusinessError("UNAUTHORIZED", ...) }
//   }
```

---

#### Phase 2 — Remove `iam.Service` from finance module

**Files:** `internal/core/finance/service.go`, `internal/core/finance/service/account.go`

1. Delete `iamService iam.Service` field from `financeService` and `accountService` structs.
2. Remove `iamService iam.Service` params from `NewService` and `NewAccountService`.
3. Remove `import "awo.so/internal/core/iam"` from both files.
4. Update call sites that pass `iamService` to these constructors.

Authorization is declared at routes. The service does not need a handle to `iam.Service`.

---

#### Phase 3 — Replace feature flag injection with `SessionContext`

**File:** `internal/core/finance/service/account.go` lines 177–188

Current pattern (violates boundary):
```go
evalCtx := &featureflag.EvaluationContext{TenantID: tenantID, UserID: &userID, ...}
if evalResult, err := s.featureFlagService.IsEnabled(ctx, "enhanced_account_validation", evalCtx); ...
```

Correct pattern:
```go
sc, ok := contract.FromContext(ctx)
if !ok {
    return nil, errors.New("unauthenticated")
}
useEnhancedValidation := sc.FeatureEnabled("enhanced_account_validation")
```

After replacing all `featureFlagService` calls, remove the field and import from both
`accountService` and `financeService`.

---

#### Phase 4 — Fix raw context key reads in `finance/service.go`

**File:** `internal/core/finance/service.go` lines 81, 90

Current (raw string key — breaks if key name changes):
```go
tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
entityID, ok := ctx.Value("entity_id").(uuid.UUID)
```

Correct (use typed keys from `internal/shared`):
```go
tenantID, ok := shared.GetTenantID(ctx)
entityID, ok := shared.GetEntityID(ctx)
```

Long-term: replace with `contract.FromContext(ctx).TenantID()` and
`contract.FromContext(ctx).EntityScope()` once handlers wire `InjectSessionContext`.

---

#### Phase 5 — Audit and remove `CapabilityContext`

**File:** `internal/shared/context.go` lines 104–124

1. `grep -r "CapabilityContext\|GetCapabilityContext\|WithCapabilityContext" internal/` to find all usage.
2. If any call site reads `.Permissions`, `.Features`, or `.Modules` for access decisions — remove and replace with `middleware.Authorize` at the route.
3. Delete `CapabilityContext`, `WithCapabilityContext`, `GetCapabilityContext` from `shared/context.go`.

---

#### Phase 6 — New module guard (prevention)

When `buy/`, `sell/`, `inventory/`, `api/handlers/`, `api/routes/` are created:
- Start from the route registration pattern in doc 14.
- Add their package paths to `TestAUTHZ3_*` scanner tests in `internal/core/iam/bypass_audit_test.go`.

---

### Contract Boundary Regression Tests

The following tests exist in `internal/core/iam/bypass_audit_test.go` and will **fail**
if a future commit introduces a forbidden pattern:

| Test | What it detects |
|---|---|
| `TestAUTHZ3_NoUserTypeAuthzBypassInHandlers` | `sess.UserType == "PLATFORM"` style bypasses in handlers |
| `TestAUTHZ3_NoDirectRoleAssignmentsQueryInHandlers` | Direct `role_assignments` DB queries in handlers/middleware |
| `TestAUTHZ3_NoHasRoleInHandlers` | `HasRole(` calls in handler code |
| `TestAUTHZ3_UserServiceHasNoEnforceMethod` | Reflection: `iam.UserService` interface has no Enforce/AddPolicy |

**Gaps to add** once finance violations are fixed:
- Scanner test: `internal/core/finance/` must not contain `import "awo.so/internal/core/iam"`.
- Scanner test: no file outside `internal/core/iam/` may contain `.Enforce(` (non-comment line).
- Scanner test: `internal/shared/context.go` must not contain `CapabilityContext`.

---

### SUMMARY

```
┌─────────────────────────────────────────────────────────────────┐
│  IAM CONTRACT ADOPTION AUDIT — 2026-05-18                       │
│                                                                 │
│  OVERALL STATUS:  ❌ NOT READY FOR RELEASE                      │
│  RISK LEVEL:      HIGH                                          │
│                                                                 │
│  Critical violations: 3 files                                   │
│    • finance/service.go       — iam.Service field + raw ctx     │
│    • finance/service/account.go — iam.Service + planned Enforce │
│    • shared/context.go        — CapabilityContext shadow auth   │
│                                                                 │
│  Clean modules: 10 packages (audit, entity, featureflag,        │
│    settings, tenant, notification, finance/transaction,         │
│    finance/transaction_entry, finance/settings_helper,          │
│    finance/anomaly_detector)                                    │
│                                                                 │
│  Blocking items before any finance endpoint goes live:          │
│    1. Delete planned Enforce() TODO (account.go:159-164)        │
│    2. Remove iam.Service from finance constructors              │
│    3. Replace featureflag.Service with SessionContext           │
│    4. Audit & remove CapabilityContext.Permissions              │
│    5. Fix raw ctx.Value("tenant_id") string key reads           │
│                                                                 │
│  Non-blocking (new modules not yet written):                    │
│    • buy/, sell/, inventory/, api/handlers/, api/routes/        │
│      — must adopt contract from day one when created            │
└─────────────────────────────────────────────────────────────────┘
```
