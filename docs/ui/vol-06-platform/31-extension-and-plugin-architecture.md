---
chapter: 31
title: "Extension and Plugin Architecture"
volume: "vol-06-platform"
section: "Platform"
description: "Planned extension points: PreConstructionHook, PostConstructionHook, plugin registry, and third-party block packages."
status: planned
---

# Chapter 31 — Extension and Plugin Architecture

> **Status: PLANNED — nothing in this chapter is implemented.**
>
> This chapter describes the intended design for allowing external packages to extend the UI platform without modifying core code.

## Table of Contents

- [31.1 Intent](#311-intent)
- [31.2 PreConstructionHook](#312-preconstructionhook)
- [31.3 PostConstructionHook](#313-postconstructionhook)
- [31.4 Plugin Registry](#314-plugin-registry)
- [31.5 Third-Party Block Packages](#315-third-party-block-packages)
- [31.6 Design Constraints](#316-design-constraints)

---

## 31.1 Intent

The current schema pipeline is closed: all stages, blocks, and page functions are registered by first-party code. There is no mechanism for an external package to insert a stage into the pipeline, add a block to an existing page, or register an entirely new module.

The extension architecture is intended to enable:

1. **Third-party modules** — an organization can ship a `forecourt` module as a Go package that registers its own pages, blocks, and permissions without touching the ERP core.
2. **Platform plugins** — cross-cutting concerns (custom logging, A/B testing, analytics instrumentation) can be injected as pipeline stages without forking the platform.
3. **Block libraries** — shared block collections can be published as Go modules and imported by multiple module packages.

---

## 31.2 PreConstructionHook

A `PreConstructionHook` would run before a `PageFn` is invoked. It would have access to the `UISessionContext` and the route, allowing it to:

- Inject additional data into the compilation context.
- Gate entire page compilations for early exit.
- Add module-specific metadata to the session.

Intended interface:

```go
// PLANNED — not implemented
type PreConstructionHook interface {
    Name() string
    BeforeConstruct(ctx context.Context, route string, sess *UISessionContext) error
}
```

Hooks would be registered globally and executed in registration order before every page function call.

---

## 31.3 PostConstructionHook

A `PostConstructionHook` would run after a `PageFn` returns its schema and before `ValidateStage`. It would receive the compiled schema and be allowed to modify it.

Intended interface:

```go
// PLANNED — not implemented
type PostConstructionHook interface {
    Name() string
    AfterConstruct(ctx context.Context, route string, schema Schema) (Schema, error)
}
```

Post-construction hooks are the intended mechanism for:
- Injecting analytics tracking attributes into all action nodes.
- Appending global footer blocks to all pages.
- Applying tenant overlay patches (§30.2).

The ordering of post-construction hooks relative to `ValidateStage` is a critical design decision: hooks that inject valid schema nodes must run before validation, but hooks that must not bypass security assertions must run after.

---

## 31.4 Plugin Registry

Plugins would self-register at startup using a `Register` function:

```go
// PLANNED — not implemented
func RegisterPlugin(plugin UIPlugin) error

type UIPlugin interface {
    Name() string
    Version() string
    PreHooks()  []PreConstructionHook
    PostHooks() []PostConstructionHook
    Stages()    []Stage
    Pages()     map[string]PageFn
    Blocks()    map[string]BlockFn
    Permissions() []string
    Flags()       []string
}
```

A plugin that provides new pages would have its `PageFn` values merged into the schema registry. A plugin that provides new permissions would have its permission strings appended to `allUIPermissions`.

Plugin registration must complete before the first request is served. Late registration (after server startup) is not supported.

---

## 31.5 Third-Party Block Packages

Block functions are currently plain Go functions. They can already be imported from external packages without any formal plugin machinery. The missing piece is the ability to inject blocks into existing pages without modifying the page function.

The intended pattern:

```go
// PLANNED — not implemented
// Forecourt module registers a block to be injected into the dashboard
ui.InjectBlockAfter("dashboard", "summary-section", ForecourtFuelSummaryBlock)
```

This would allow the forecourt module to add a fuel summary panel to the shared dashboard page without the dashboard page function knowing about forecourt.

The injection system needs conflict resolution (two plugins injecting after the same anchor), ordering guarantees, and integration with the validation and caching pipeline.

---

## 31.6 Design Constraints

Any extension mechanism must satisfy:

1. **Security boundary preserved**: Plugin-provided hooks and stages cannot bypass `ValidateStage` assertions. Plugins cannot read or modify `UISessionContext.perms` directly.

2. **Cache correctness**: If a plugin modifies a compiled schema, the cache key must reflect the plugin's version. A plugin version bump must invalidate cached schemas that the plugin touched.

3. **No IAM imports in plugin code**: Plugins interact with the session only through `UISessionContext` public methods, not through IAM internals.

4. **Compile-time registration only**: Plugins register at server startup. Runtime registration (hot-loading) is out of scope.

5. **Testability**: Plugins must be testable without a live server. This means the pipeline must support a test mode where hooks and stages can be exercised against an in-memory request.
