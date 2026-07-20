> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "42 – CI/CD Considerations"
volume: "vol-09-engineering"
chapter: 42
section: "Engineering"
status: "partial — startup guard implemented; CI script planned"
---

# Chapter 42 – CI/CD Considerations

## Table of Contents
- [42.1 ValidateRegistry at Startup](#421-validateregistry-at-startup)
- [42.2 CI Architecture Guard Script](#422-ci-architecture-guard-script)
- [42.3 The Eight Guards](#423-the-eight-guards)
- [42.4 Screen Size Lint](#424-screen-size-lint)

---

## 42.1 ValidateRegistry at Startup

`ValidateRegistry()` is called inside `NewUIPipeline` during application
startup.  If any registered page has an invalid `PageRegistration`, the process
exits before the HTTP server starts — fail-fast, no partial startup.

```go
// internal/web/pipeline/pipeline.go
func NewUIPipeline(...) (*Pipeline, error) {
    if err := registry.ValidateRegistry(); err != nil {
        return nil, fmt.Errorf("UI registry validation failed at startup: %w", err)
    }
    // ... build stages
}
```

This means:
- A deploy with a malformed `RegisterPage` call will be caught during `main`
  startup, before the process becomes healthy.
- Kubernetes readiness probes will never return healthy for a binary with a
  broken registry.
- No special CI step is required for this check — it runs on every deployment.

For the testing counterpart, see §41.7.

---

## 42.2 CI Architecture Guard Script

> **PLANNED (Task 13)** — not yet implemented.

A shell script `scripts/check-arch.sh` is planned to run as a CI step after
`go build` succeeds.  It enforces structural rules that the Go compiler cannot
check:

```yaml
# .github/workflows/ui.yml (planned)
- name: UI architecture guards
  run: bash scripts/check-arch.sh
```

The script will exit non-zero if any guard fails, blocking the PR merge.

---

## 42.3 The Eight Guards

The planned `scripts/check-arch.sh` will enforce the following eight rules.
These rules exist today as conventions enforced by code review; the script
mechanises them.

| # | Guard | Rationale |
|---|---|---|
| 1 | **No `map[string]any` in `internal/web/`** | Zero tolerance. Untyped maps bypass `CompileTree` validation and produce runtime errors. Use `ast.*` struct literals. |
| 2 | **No direct IAM imports from `dsl/`** | DSL code must only access IAM data via `UISessionContext` fields. Direct IAM package imports create coupling that breaks cache isolation. |
| 3 | **No permission string checks in `visibleOn` predicates** | AMIS `visibleOn` expressions run in the browser and cannot evaluate server-side permission state. Permission gating belongs in block functions. |
| 4 | **No schema generated outside the pipeline** | Schema `any` values must not be constructed in handlers, middleware, or init functions — only inside `PageFn` / `ASTPageFn` called by `CompileStage`. |
| 5 | **Every stage has non-empty `DependsOn`** | Stages without declared dependencies execute in undefined order. Every stage except the first must declare at least one predecessor. |
| 6 | **No `sync.Map` or third-party cache in `internal/web/`** | The UI cache is managed by `CacheStage` exclusively. Ad-hoc caches inside DSL code or stages create invalidation blind spots. |
| 7 | **No block-level concerns defined in multiple `screens/` files** | If the same block logic appears in more than one screen file, it must be extracted to `blocks/`. Copy-paste in screens is a governance failure. |
| 8 | **Every `screens/` file is under 60 lines** | A screen file over 60 lines means business logic escaped the block layer. Extract a block. |

Until the script is implemented, these rules are enforced at code review time
using the governance model described in §45.

---

## 42.4 Screen Size Lint

Guard #8 (60-line screen limit) deserves elaboration because it is the most
commonly violated rule in early development.

The 60-line limit is not arbitrary — it is calibrated so that a screen file
contains only:
- One `RegisterPage` call (in `init()`)
- One screen function signature
- 3–8 block composition calls
- Minimal structural glue (e.g. wrapping nodes in a `PageNode`)

If a screen file grows beyond 60 lines, it is evidence that:
1. A block function was inlined instead of extracted, **or**
2. A conditional that belongs in a block was written in the screen.

Until the CI script exists, reviewers should flag files in `screens/` that
exceed 60 lines as a blocking review comment.

```sh
# Manual check (run from repo root):
awk 'END { if (NR > 60) print FILENAME, NR, "lines — exceeds 60-line limit" }' \
    internal/web/dsl/screens/*.go
```
