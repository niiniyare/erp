---
title: "43 – Versioning Strategy"
volume: "vol-09-engineering"
chapter: 43
section: "Engineering"
status: "implemented (cache versioning); formal schema versioning planned"
---

# Chapter 43 – Versioning Strategy

## Table of Contents
- [43.1 CacheVersions Struct](#431-cacheversions-struct)
- [43.2 The Four Version Counters](#432-the-four-version-counters)
- [43.3 When to Bump Each Version](#433-when-to-bump-each-version)
- [43.4 Version Propagation Through the Cache Key](#434-version-propagation-through-the-cache-key)
- [43.5 Planned: Formal Schema Versioning](#435-planned-formal-schema-versioning)

---

## 43.1 CacheVersions Struct

`CacheVersions` is a value struct that encodes the current "generation" of all
compiled schemas.  It is embedded into every cache key so that bumping any
counter invalidates all entries derived under the old value.

```go
// internal/web/pipeline/versions.go
type CacheVersions struct {
    CompilerVersion   int // bump when CompileStage logic changes
    ASTVersion        int // bump when ast.* node structs change shape
    PolicyGeneration  int // bump when IAM permission definitions change
    SchemaGeneration  int // bump when DSL screen/block code changes
}
```

`CacheVersions` is passed into `NewUIPipeline` and threaded through to
`CacheStage` where it is serialised into the cache key:

```go
key := fmt.Sprintf("%s|%s|%d.%d.%d.%d|%s",
    route,
    permFingerprint,
    v.CompilerVersion,
    v.ASTVersion,
    v.PolicyGeneration,
    v.SchemaGeneration,
    flagFingerprint,
)
```

A cache entry is a hit only when **all four version counters** match the current
values.  Any mismatch triggers a recompile and counter increment of
`ui_cache_generation_mismatch_total`.

---

## 43.2 The Four Version Counters

### CompilerVersion

Tracks changes to the pipeline itself — specifically `CompileStage`,
`ValidateStage`, and any stage that transforms the schema tree.  Bump this when
the pipeline produces structurally different output for the same input.

**Examples that require a bump:**
- `ValidateStage` gains a new validation rule that rejects previously accepted
  schemas
- `CompileStage` changes how it serialises `ast.TabsNode`
- A new stage is inserted into the pipeline that modifies schema shape

### ASTVersion

Tracks changes to the `ast.*` node struct definitions.  Bump this when any
field is added, removed, renamed, or changes its JSON tag.

**Examples that require a bump:**
- `ast.CRUDNode` gains a new `FilterMode` field
- `ast.TabsNode.MountOnEnter` is renamed
- A new node type is added to the `ast` package

### PolicyGeneration

Tracks changes to IAM permission definitions that affect permission
fingerprinting.  Bump this when the set of possible permissions changes (new
module, new action, renamed permission).

**Examples that require a bump:**
- A new module registers its permissions with the IAM registry
- A permission key is renamed (e.g. `finance.invoices.view` →
  `finance.invoices.read`)
- A permission is removed from the policy set

### SchemaGeneration

Tracks changes to DSL screen and block functions.  This is the counter bumped
most frequently during active development.  Bump this whenever any file under
`internal/web/dsl/` changes.

**Examples that require a bump:**
- A new block is added to a screen
- A block function is edited to change its output
- A new screen is registered
- A `PageRegistration` title is updated

---

## 43.3 When to Bump Each Version

| Situation | Counter to bump |
|---|---|
| Deploy with no DSL changes | None — cache remains valid |
| New screen or block added | `SchemaGeneration` |
| Block or screen function edited | `SchemaGeneration` |
| `ast.*` struct field added/changed | `ASTVersion` + `SchemaGeneration` |
| Pipeline stage logic changed | `CompilerVersion` |
| New permission defined in IAM | `PolicyGeneration` |
| Permission key renamed | `PolicyGeneration` + `SchemaGeneration` |
| Emergency: invalidate everything | Bump all four |

In practice, `SchemaGeneration` is bumped on almost every deploy that touches
`internal/web/`.  The other counters change rarely.

**Convention**: store `CacheVersions` in a dedicated constant file that is
updated as part of the same commit that changes the relevant code:

```go
// internal/web/pipeline/version_consts.go
var ActiveVersions = CacheVersions{
    CompilerVersion:  1,
    ASTVersion:       3,
    PolicyGeneration: 2,
    SchemaGeneration: 47, // bump this on every DSL change
}
```

---

## 43.4 Version Propagation Through the Cache Key

The full cache key structure (for reference):

```
{route}|{permFingerprint}|{CV}.{AV}.{PG}.{SG}|{flagFingerprint}
```

Example:

```
/finance/invoices|sha256:a3f9...|1.3.2.47|sha256:b7c1...
```

A request that hits this key exactly is served from cache with no pipeline
execution.  A request where any component differs falls through to a full
pipeline run.

---

## 43.5 Planned: Formal Schema Versioning

> **PLANNED** — not yet implemented.

Current versioning is _internal_ — it governs cache invalidation within a single
running server.  It does not provide guarantees to external consumers (e.g.
mobile clients) about schema stability.

The following capabilities are on the roadmap:

| Capability | Description |
|---|---|
| Schema version header | Include `X-UI-Schema-Version` in schema HTTP responses so clients can detect breaking changes. |
| Semantic schema versioning | Separate major/minor/patch versioning: major bumps indicate breaking layout changes, minor bumps indicate additive changes. |
| Schema diff tooling | A CLI tool that compares two compiled schemas and outputs a human-readable diff of structural changes. |
| Client schema negotiation | Mobile clients declare the schema version they support; the server serves the appropriate compiled variant. |
| Deprecation annotations | `ast.*` nodes can be annotated as deprecated; `ValidateStage` emits warnings when deprecated nodes are used. |

Until these are implemented, all schema consumers (web shell, future mobile
clients) must tolerate schema changes on every deploy.
