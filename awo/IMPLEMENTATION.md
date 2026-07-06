# Awo Framework — Implementation Status

Module: `awo.so` (part of root module, packages under `awo.so/awo/...`)

## Phase Status

| Phase | Status | Packages |
|---|---|---|
| 1 — Repository structure | ✅ Done | Directory tree established |
| 2 — def kernel | ✅ Done | `awo.so/awo/def` |
| 3 — filter DSL | ✅ Done | `awo.so/awo/filter` |
| 3 — registry | ✅ Done | `awo.so/awo/registry` |
| 4 — compiler | ✅ Done | `awo.so/awo/compiler` |
| 5 — runtime (partial) | 🔄 In Progress | `awo.so/awo/runtime`, `awo.so/awo/runtime/tenant` |
| 6 — driver interfaces | ✅ Done | `awo.so/awo/driver` |
| 6 — pgx driver | ⏳ Next | `awo.so/awo/driver/pgx` |
| 7 — platform/tenant | ⏳ Next | `awo.so/awo/platform/tenant` |
| 7 — platform/iam | ⏳ Next | `awo.so/awo/platform/iam` |
| 7 — platform/audit | ⏳ Next | `awo.so/awo/platform/audit` |
| 7 — platform/flags | ⏳ Next | `awo.so/awo/platform/flags` |
| 7 — platform/settings | ⏳ Next | `awo.so/awo/platform/settings` |
| 7 — platform/metadata | ⏳ Next | `awo.so/awo/platform/metadata` |
| 7 — platform/registry | ⏳ Next | `awo.so/awo/platform/registry` |
| 8 — cmd/migrate | ⏳ Next | `awo.so/awo/cmd/migrate` |
| 8 — cmd/server | ⏳ Next | `awo.so/awo/cmd/server` |
| 9 — examples | 🔄 In Progress | `awo.so/awo/examples/finance` |
| 9 — tests | 🔄 In Progress | Per-package *_test.go files |

## Package Dependency Graph

```
def          ← (stdlib, uuid, decimal only)
  ↑
filter       ← def.Filter interface
  ↑
registry     ← def (reads All(), seals)
  ↑
compiler     ← registry, def
  ↑
driver       ← def, filter, compiler
  ↑
runtime      ← compiler, def
  ↑           runtime/tenant ← (stdlib, uuid only)
internal/dberr ← runtime (for error types), pgconn

platform/*   ← driver, runtime, def
```

## Key Architectural Decisions

1. **def has zero framework deps** — only stdlib + uuid + decimal.
2. **filter.Filter implements def.Filter interface** — no circular import.
3. **Registry sealed once** — `def.Seal()` called from `registry.Build()`.
4. **Pipeline does not open transactions** — delegates to driver.
5. **TenantContext in context.Context only** — panics on absent context (intentional).
6. **No lazy loading** — edges always explicitly requested via QueryOption.
7. **No ORM types** — all persistence through EntityRepository[T] interface.
