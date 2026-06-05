---
chapter: 26
title: "Security Model"
volume: "vol-06-platform"
section: "Platform"
description: "Security boundaries, identity flow, permission pre-resolution, ValidateStage assertions, and UISessionContext immutability."
status: implemented
---

# Chapter 26 — Security Model

## Table of Contents

- [26.1 Security Boundary](#261-security-boundary)
- [26.2 Identity Flow](#262-identity-flow)
- [26.3 Permission Pre-Resolution](#263-permission-pre-resolution)
- [26.4 Permission Pruning in Block Functions](#264-permission-pruning-in-block-functions)
- [26.5 ValidateStage Security Assertions](#265-validatestage-security-assertions)
- [26.6 UISessionContext Immutability](#266-uisessioncontext-immutability)
- [26.7 IAM Import Boundary](#267-iam-import-boundary)

---

## 26.1 Security Boundary

The schema handler never reads Fiber locals directly. This is the first and most important security invariant of the UI platform.

Fiber stores request-scoped values (JWT claims, tenant IDs, raw headers) in a mutable, weakly-typed locals map. Any code with access to the Fiber context could read or mutate these values. To prevent accidental bypasses, the schema pipeline does not interact with Fiber locals at all after the middleware chain has run.

Instead, the boundary is enforced at the `contract` package level:

```
Fiber Locals (raw)
      ↓
InjectSessionContext middleware
      ↓
contract.SessionContext (typed, validated)
      ↓
AuthzStage → UISessionContext (typed, immutable, permission-resolved)
      ↓
PageFn / Block functions (never touch Fiber)
```

Every component in the pipeline that needs identity information calls `contract.FromContext(ctx)`. If the context carries no session (e.g., a bug in the middleware chain), `FromContext` returns an error and the pipeline aborts — it does not fall back to reading locals or assuming an identity.

---

## 26.2 Identity Flow

```
Client
  │  Authorization: Bearer <jwt>
  ▼
Authenticate middleware
  │  Validates JWT signature and expiry
  │  Rejects → 401 AMIS envelope
  │  Accepts → attaches claims to context
  ▼
InjectSessionContext middleware
  │  Reads userID, tenantID from claims
  │  Hydrates contract.SessionContext (roles, settings, flags)
  │  Stores in context via contract key
  ▼
Schema Handler (Handle)
  │  Calls contract.FromContext(ctx) → SessionContext
  │  Starts pipeline
  ▼
AuthzStage (priority 20)
  │  Calls UIAuthzService.BulkEnforce(userID, tenantID, allKnownPermissions)
  │  Constructs UISessionContext via NewUISessionContext (sole constructor)
  ▼
PageFn / Block functions
     Use sess.Can(), sess.Flag() — no Casbin, no IAM imports
```

### What Each Stage Knows

| Stage                 | Knows user identity | Knows permissions | Knows feature flags |
|-----------------------|---------------------|-------------------|---------------------|
| Authenticate          | Partially (claims)  | No                | No                  |
| InjectSessionContext  | Yes                 | No                | No                  |
| AuthzStage            | Yes                 | Resolves them now | Resolves them now   |
| CacheStage            | Yes (for key)       | Via fingerprint   | Via fingerprint     |
| PageFn / blocks       | Yes (via sess)      | Yes (via sess)    | Yes (via sess)      |

---

## 26.3 Permission Pre-Resolution

AuthzStage resolves all permissions for the current user in a single `BulkEnforce` call before any page function runs.

```go
// AuthzStage calls this once — never called again during the request
perms, err := s.authzSvc.BulkEnforce(ctx, userID, tenantID, allUIPermissions)
// perms is map[string]bool{"finance.invoices.create": true, "hr.employees.read": false, ...}

// UISessionContext is constructed with the resolved map
sess := NewUISessionContext(sessionCtx, perms, flags)
```

`allUIPermissions` is the exhaustive list of every permission string known to the UI platform. It is defined statically and updated when new modules are added. There is no dynamic discovery at request time.

The result is that all subsequent stages — ValidateStage, the PageFn, every block function — can call `sess.Can()` without any I/O. It is a pure map lookup.

This has two consequences:

1. **No per-field Casbin calls.** A page with 50 permission-gated fields does not make 50 IAM calls. The cost is always one call, regardless of schema complexity.
2. **Permission state is frozen.** Permissions cannot change mid-compilation. A block function cannot cause a permission re-evaluation by calling external services.

---

## 26.4 Permission Pruning in Block Functions

Blocks own their permission checks. The screen layer (the `PageFn` that assembles blocks) does not gate blocks — each block function decides internally whether to include or exclude components.

### Correct Pattern

```go
// Inside a block function
func InvoiceActionsBlock(sess *UISessionContext) []Node {
    nodes := []Node{}

    if sess.Can("create", "finance.invoices") {
        nodes = append(nodes, CreateButton())
    }
    if sess.Can("delete", "finance.invoices") {
        nodes = append(nodes, DeleteButton())
    }
    return nodes
}

// The screen assembles blocks without knowing what's inside them
func InvoicePage(sess *UISessionContext) Schema {
    return Page(
        InvoiceTableBlock(sess),
        InvoiceActionsBlock(sess),  // screen never checks permissions for this
    )
}
```

### Wrong Pattern

```go
// WRONG — screen layer doing permission checks
func InvoicePage(sess *UISessionContext) Schema {
    var actions []Node
    if sess.Can("create", "finance.invoices") {  // should be inside the block
        actions = append(actions, InvoiceActionsBlock(sess))
    }
    return Page(InvoiceTableBlock(sess), actions...)
}
```

The wrong pattern breaks the separation of concerns and makes screens aware of the permissions their child blocks require. It also makes it easy to accidentally skip a permission check when a block is reused on another screen.

---

## 26.5 ValidateStage Security Assertions

ValidateStage runs on every request — it is not skipped even on cache hits. Its security assertions protect against developer mistakes that would expose permission information to the client.

### Assertion: No Permission Strings in Expressions

AMIS `visibleOn` and `disabledOn` expressions are evaluated in the browser. If a permission string appears in one of these expressions, it could leak information about the permission model to the client.

ValidateStage scans the compiled schema for any expression containing strings that match the permission key pattern (`<module>.<resource>.<action>`). If found, the request fails with a 500 and the violation is logged.

```json
// INVALID — fails ValidateStage
{
  "type": "button",
  "label": "Delete",
  "visibleOn": "${permissions['finance.invoices.delete']}"
}

// VALID — permission check happened server-side; button is absent if not permitted
{
  "type": "button",
  "label": "Delete"
}
```

The correct approach: block functions use `sess.Can()` at compile time to decide whether to include the button at all. The permission string never appears in the rendered schema.

---

## 26.6 UISessionContext Immutability

`UISessionContext` is constructed exactly once per request by `NewUISessionContext` inside `AuthzStage`. It has no public setters.

```go
// Sole constructor — called only by AuthzStage
func NewUISessionContext(
    base contract.SessionContext,
    perms map[string]bool,
    flags map[string]bool,
) *UISessionContext

// Exported read methods only
func (s *UISessionContext) Can(action, resource string) bool
func (s *UISessionContext) CanAny(action string, resources ...string) bool
func (s *UISessionContext) CanAll(action string, resources ...string) bool
func (s *UISessionContext) Flag(name string) bool
func (s *UISessionContext) UserID() string
func (s *UISessionContext) TenantID() string
func (s *UISessionContext) Currency() string
func (s *UISessionContext) Locale() string
func (s *UISessionContext) Timezone() string
```

The `perms` and `flags` maps passed to the constructor are deep-copied. After construction, no code outside `AuthzStage` can modify the permission state. This makes `UISessionContext` safe to pass across goroutines without locks.

---

## 26.7 IAM Import Boundary

No file inside the `dsl/` or `handler/` packages may import IAM internal types. The schema DSL and the pipeline stages interact with identity exclusively through:

- `contract.SessionContext` — the hydrated session from `InjectSessionContext`
- `UISessionContext` — the permission-resolved view, constructed by `AuthzStage`

If a DSL block needed to call an IAM service directly, it would create a hidden I/O dependency inside schema compilation, making schemas untestable without a live IAM backend and breaking the batch-resolve guarantee from §26.3.

The `UIAuthzService` interface is the only authorized integration point:

```go
type UIAuthzService interface {
    BulkEnforce(ctx context.Context, userID, tenantID string, permissions []string) (map[string]bool, error)
}
```

Only `AuthzStage` holds a reference to `UIAuthzService`. Block functions and page functions never hold or call it.
