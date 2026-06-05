# Chapter 02 — Design Philosophy

> **Volume:** I — Vision & Philosophy
> **Audience:** All Engineers, Product Owners, Solution Architects
> **Prerequisites:** Chapter 01 — Introduction

---

## Table of Contents

- [2.1 The Server-Driven UI Manifesto](#21-the-server-driven-ui-manifesto)
- [2.2 Inspirations and Prior Art](#22-inspirations-and-prior-art)
- [2.3 Core Design Tenets](#23-core-design-tenets)
- [2.4 Trade-offs Accepted](#24-trade-offs-accepted)
- [2.5 Non-Goals](#25-non-goals)
- [2.6 Philosophy vs. Implementation — Where Rules Can Flex](#26-philosophy-vs-implementation--where-rules-can-flex)

---

## 2.1 The Server-Driven UI Manifesto

Every software architecture embodies a theory about where intelligence should live. In a conventional frontend architecture, intelligence is split: the backend enforces business rules; the frontend independently manages what to show and when. Both sides hold partial copies of the same truth.

AwoERP is built on a different theory: **the backend is the complete and sole authority on what the user sees; the frontend is an executor of backend intent.**

This is not a claim that frontend engineering is unimportant. Building a rendering engine that faithfully materializes a rich JSON specification — with responsive layouts, accessible interactions, and correct AMIS component semantics — is substantial work. The claim is that this work should be directed at **execution excellence**, not at re-encoding authorization or workflow decisions that the backend already owns.

Five sentences:

1. The backend knows the user, the tenant, the permissions, the workflow state, and the business rules. The backend decides what the user sees.
2. The AMIS SDK knows how to render JSON schemas into interactive web UI. The SDK decides how to present what it is told.
3. The boundary between "what" and "how" is explicit: a typed Go AST compiled to an AMIS JSON schema.
4. Anything that crosses that boundary in the wrong direction — authorization logic in the browser, pixel decisions in Go — is a defect.
5. The schema is the contract.

---

## 2.2 Inspirations and Prior Art

### 2.2.1 AMIS (Baidu)

AMIS is an open-source JSON-driven UI framework. It demonstrated that a sufficiently rich JSON vocabulary can express enterprise-grade UI complexity — forms, tables, charts, dashboards, dialog flows — without custom JavaScript per screen.

**What AwoERP inherits:** The JSON schema vocabulary and the AMIS SDK as the rendering runtime. AwoERP uses AMIS types directly: `"type": "page"`, `"type": "crud"`, `"type": "form"`, `"type": "chart"`, etc.

**What AwoERP does differently:** AMIS is typically used with static JSON files authored by frontend developers. AwoERP generates schemas dynamically at the backend, incorporating Casbin authorization, tenant feature flags, and user locale at compile time. The JSON is a computed artifact, not a static template.

### 2.2.2 Retool and Appsmith

Both platforms demonstrated that business users and developers could compose UI from a declarative component model. Their event-action systems — where user interactions trigger named actions that can be chained, conditioned, and error-branched — inform AwoERP's action design.

**What AwoERP does differently:** In Retool and Appsmith, UI definitions are constructed interactively and stored in a database. AwoERP page definitions are Go code, in the same repository as the business logic that validates form submissions. The page author and the API handler author are the same engineer, sharing the same domain model.

### 2.2.3 Meta, Airbnb SDUI Patterns

Server-Driven UI for high-scale consumer apps (Meta's news feed, Airbnb's search results) demonstrated that a backend-owned schema can drive mobile UI at production scale. These systems proved that client caching, fingerprint-based invalidation, and schema versioning are necessary engineering concerns.

**What AwoERP inherits:** The fingerprint-based cache key design — cache entries are keyed by permission fingerprint and feature flag fingerprint, so users with identical contexts share cache entries regardless of user ID.

---

## 2.3 Core Design Tenets

### 2.3.1 Business Logic Drives UI

The Go service that handles invoice creation also defines the invoice creation form. These are not separate artifacts. The form fields, required markers, and action buttons are an expression of the same domain rules that validate the submission. They live in the same file, are reviewed together, and are versioned together.

```go
// internal/web/dsl/screens/invoice.go
// The screen is co-located with the finance module code.
func InvoiceScreen(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        Title: "New Invoice",
        Body: []ast.Node{
            // ...fields that mirror the domain validation rules...
        },
    }
}
```

### 2.3.2 AST Before JSON

Every page definition starts as a typed Go struct graph (`ast.Node`), not as `map[string]any`. The AST is validated by `ast.CompileTree()` before any JSON is emitted. Structural errors — a `CRUDNode` with no columns, a `FormNode` with no fields — are caught at compile time or at the first request, never silently delivered to the browser.

```go
// ast.CompileTree validates every node before serializing any of them.
// If any node fails Validate(), the whole tree is rejected.
schema, err := ast.CompileTree(root)
```

The legacy `PageFn` path (returning raw `map[string]any`) remains supported during migration. New pages must use `ASTPageFn`.

### 2.3.3 Permissions Resolved Once, Before Compilation

The pipeline resolves all Casbin permissions via `UIAuthzService.BulkEnforce()` in a single batch call (AuthzStage, priority 20). The result is a `map[string]bool` stored in `UISessionContext`. Every subsequent operation in the request — page function, block, column filter — reads from that pre-resolved map. No Casbin call happens inside a page function.

```go
// UISessionContext.Can() reads the pre-resolved map — not Casbin.
func (u UISessionContext) Can(action, resource string) bool {
    return u.permissions[resource+"."+action]
}
```

### 2.3.4 Cache Keyed on Authorization State

The cache key includes a cryptographic fingerprint of the user's resolved permissions and active feature flags. Two users with the same permission set share the same cache entry. A permission change produces a new fingerprint and automatically bypasses the stale cache.

This design — cache after authz, not before — is why CacheStage runs at priority 30 (after AuthzStage at 20). The fingerprint must exist before the cache key can be computed.

### 2.3.5 The Frontend Is Not Trusted for Authorization

The AMIS renderer never makes an authorization decision. It renders whatever schema it receives. If a button is absent from the schema, the button does not exist for that user — there is no client-side check to bypass, no hidden DOM element to reveal. Least privilege is enforced structurally.

### 2.3.6 Pure Page Functions

`PageFn` and `ASTPageFn` are pure functions. Given the same `UISessionContext`, they must produce the same schema. They must not call the database, make HTTP requests, read files, or spawn goroutines. This purity is what makes caching safe: the output is a deterministic function of the inputs, which are captured in the cache key.

```go
// WRONG — impure page function
func BadPage(sess ui.UISessionContext) ast.Node {
    items, _ := db.Query("SELECT ...") // NO: no I/O in page functions
    return ast.PageNode{...}
}

// CORRECT — data is loaded by InitAPI at browser runtime
func GoodPage(sess ui.UISessionContext) ast.Node {
    return ast.PageNode{
        InitAPI: &ast.APISpec{Method: "get", URL: "/api/v1/finance/invoices"},
        Body:    []ast.Node{...},
    }
}
```

---

## 2.4 Trade-offs Accepted

### 2.4.1 AMIS as the Only Client Today

The platform is currently AMIS-first. The JSON schema vocabulary is AMIS's vocabulary. A Flutter or React Native client would need to understand AMIS schemas or require a translation layer. This is an accepted constraint — the AMIS SDK delivers substantial capability for internal web tooling, and mobile clients are on the roadmap.

> **Status: Planned — Not Yet Implemented**
> Mobile rendering clients (Flutter, React Native) are planned. Until then, AMIS is the only rendering target.

### 2.4.2 All UI Decisions Require a Backend Deploy

If a product manager wants to rename a form label or reorder two columns, that change requires a Go code edit, a CI pipeline, and a backend deployment. There is no no-code customization UI today.

> **Status: Planned — Not Yet Implemented**
> Tenant form customization overlays and a no-code admin UI are on the roadmap.

### 2.4.3 No Incremental / Streaming Schema Updates

The pipeline delivers a complete schema on every request. There is no mechanism to stream partial updates or push schema changes to active browser sessions. Real-time schema updates require a full page reload.

---

## 2.5 Non-Goals

The UI platform does not:

- Serve as a general-purpose low-code platform for external developers to build arbitrary applications
- Provide a visual schema designer or drag-and-drop editor (planned, not implemented)
- Manage business data — schemas describe UI, not domain records
- Replace or duplicate any IAM, notification, or event bus system — it integrates with them
- Make authorization decisions at the browser — it makes them at compilation time

---

## 2.6 Philosophy vs. Implementation — Where Rules Can Flex

### The Legacy PageFn Path

New pages must use `ASTPageFn`. The typed AST provides compile-time correctness guarantees the raw `map[string]any` path cannot. However, existing pages registered with the legacy `PageFn` continue to work. `CompileStage` dispatches to `ASTPageFn` first and falls back to `PageFn` when only the legacy registration exists.

During the migration window, mixing both is acceptable. The migration is tracked per registration in the registry.

### AMIS Expression Syntax in Raw Schemas

Legacy `PageFn` pages may use AMIS expression strings directly (`"${amount}"`, `"${!can_approve}"`). These are strings in the raw schema map; they are not validated by the AST. Pages using `ASTPageFn` and the typed AST get structural validation; raw expression strings in leaf fields remain the author's responsibility.

---

*End of Chapter 02*

**Previous:** [Chapter 01 — Introduction](./01-introduction.md)
**Next:** [Chapter 03 — Architectural Principles](./03-architectural-principles.md)
