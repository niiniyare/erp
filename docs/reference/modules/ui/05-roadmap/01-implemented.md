# Implemented Features

> Last verified: 2026-05-18 | This is the authoritative feature status list.
> Update this file when features ship or when implementation changes.

---

## UI Pipeline

| Feature | Status | Code Location | Notes |
|---------|--------|---------------|-------|
| 9-stage schema pipeline | ✅ Implemented | `internal/web/ui/pipeline.go`, `internal/web/stages/` | All 9 stages at priorities 10–90 |
| Session validation (P=10) | ✅ Implemented | `internal/web/stages/session.go` | |
| Authz + permission resolution (P=20) | ✅ Implemented | `internal/web/stages/authz.go` | Casbin BulkEnforce, 35 permissions |
| Cache lookup (P=30) | ✅ Implemented | `internal/web/stages/cache.go` | L1 LRU + L2 Redis |
| Registry dispatch (P=40) | ✅ Implemented | `internal/web/stages/registry.go` | ASTPageFn preferred, PageFn fallback |
| Schema compilation (P=50) | ✅ Implemented | `internal/web/stages/compile.go` | Both `PageFn` and `ASTPageFn` paths |
| Normalize (P=60) | ✅ Implemented | `internal/web/stages/normalize.go` | Lowercase types, trim API strings, syncLocation |
| Validate (P=70) | ✅ Implemented | `internal/web/stages/validate.go` | Expression security rules, structural rules |
| Cache store (P=80) | ✅ Implemented | `internal/web/stages/cache.go` | TTL=5min, non-fatal on failure |
| Response assembly (P=90) | ✅ Implemented | `internal/web/stages/response.go` | `UISchemaOutput` struct |

---

## Page Registry

| Feature | Status | Code Location | Notes |
|---------|--------|---------------|-------|
| `PageRegistration` struct | ✅ Implemented | `internal/web/registry/registry.go` | Route, Module, Title, Fn/ASTFn |
| `RegisterPage()` | ✅ Implemented | `internal/web/registry/registry.go` | Thread-safe init() registration |
| `ValidateRegistry()` | ✅ Implemented | `internal/web/registry/registry.go` | Startup panic on invalid registrations |
| Duplicate route detection | ✅ Implemented | `internal/web/registry/registry.go` | Panics at startup |

---

## Authorization

| Feature | Status | Code Location | Notes |
|---------|--------|---------------|-------|
| Casbin permission resolution | ✅ Implemented | `internal/web/authz/service.go` | CasbinUIAuthzService |
| `BulkEnforce` (single Casbin call) | ✅ Implemented | `internal/web/authz/service.go` | ~35 permissions per request |
| `AllUIPermissions` registry | ✅ Implemented | `internal/web/authz/service.go` | Must add permission here to use it |
| Permission fingerprinting | ✅ Implemented | `internal/web/authz/service.go` | SHA256 of true-valued permission keys |
| Feature flag fingerprinting | ✅ Implemented | `internal/web/authz/service.go` | SHA256 of enabled flag names |
| `UISessionContext` | ✅ Implemented | `internal/web/ui/types.go` | `Can()`, `CanAny()`, `CanAll()`, `Flag()`, `Pref()` |
| `IsPlatform()` / `IsPortal()` | ✅ Implemented | `internal/core/iam/contract/session.go` | Binary mode detection |

---

## Caching

| Feature | Status | Code Location | Notes |
|---------|--------|---------------|-------|
| Redis schema cache | ✅ Implemented | `internal/web/stages/cache.go` | 5-min TTL |
| In-process LRU (L1) | ✅ Implemented | `internal/platform/cache/` | Checked before Redis |
| Cache key with fingerprints | ✅ Implemented | `internal/web/stages/cache.go` | `ui:schema:{tenant}:{route}:{permFP}:{flagFP}` |
| Generation-based cache key | ✅ Implemented | `internal/web/stages/cache.go` | Requires `CacheVersions` injection |
| Non-fatal cache failure | ✅ Implemented | `internal/web/stages/cache.go` | Errors treated as miss |
| Automatic invalidation on role change | ✅ Implemented | Via fingerprint — different role = different permFP = automatic miss |

---

## Schema Builder

| Feature | Status | Code Location | Notes |
|---------|--------|---------------|-------|
| `amis.Ctx` / `PageFn` (legacy) | ✅ Implemented | `internal/web/amis/builder.go` | Still supported during migration |
| `UISessionContext` / `ASTPageFn` (modern) | ✅ Implemented | `internal/web/ui/types.go` | Preferred for new pages |
| `amis.Page`, `amis.Grid`, `amis.Panel` | ✅ Implemented | `internal/web/amis/page.go` | |
| `amis.CRUD`, column builders | ✅ Implemented | `internal/web/amis/crud.go` | |
| `amis.Form`, `amis.Wizard`, field helpers | ✅ Implemented | `internal/web/amis/form.go` | |
| `amis.Chart` | ✅ Implemented | `internal/web/amis/page.go` | Transparent bg enforced |
| DSL blocks (~26) | ✅ Implemented | `internal/web/dsl/blocks/` | `UIBlock` reusable fragments |
| DSL screens | ✅ Implemented | `internal/web/dsl/screens/` | `CRUDScreen`, `FormScreen`, etc. |
| Typed AST compilation | ✅ Implemented | `internal/web/ast/` | `ast.CompileTree()` |

---

## Shell / Frontend

| Feature | Status | Code Location | Notes |
|---------|--------|---------------|-------|
| Custom HTML shell | ✅ Implemented | `web/pages/index.html` | No build step |
| Dark mode (3 layers) | ✅ Implemented | `web/pages/index.html` | Shell + AMIS tokens + portals |
| System/light/dark theme toggle | ✅ Implemented | `web/pages/index.html` | `localStorage('awo-theme')` |
| Responsive layout (mobile/tablet) | ✅ Implemented | `web/pages/index.html` | `@media` queries |
| Safe area support (notch devices) | ✅ Implemented | `web/pages/index.html` | `env(safe-area-inset-bottom)` |
| Sidebar collapse (desktop) | ✅ Implemented | `web/pages/index.html` | 260px → 60px |
| Mobile sidebar overlay | ✅ Implemented | `web/pages/index.html` | Slide-in with backdrop |
| Hash routing | ✅ Implemented | `web/pages/index.html` | `navigate()` on hashchange |
| Go-driven schema loading | ✅ Implemented | `web/pages/index.html` | `fetch('/schema/' + route)` |
| AMIS fetcher (data API) | ✅ Implemented | `web/pages/index.html` | page/perPage → offset/limit translation |

---

## Observability

| Feature | Status | Code Location | Notes |
|---------|--------|---------------|-------|
| OTel trace span per schema request | ✅ Implemented | `internal/web/handler/schema.go` | `ui.schema_handler` span |
| Trace attributes (route, tenant, cache_hit) | ✅ Implemented | `internal/web/handler/schema.go` | |
| `ui_schema_requests_total` counter | ✅ Implemented | `internal/web/handler/schema.go` | Fields: route, tenant_id, cache_hit |

---

## Known Production Issues

| Issue | Severity | Location | Description |
|-------|----------|----------|-------------|
| Font Awesome on external CDN | 🔴 Critical | `index.html:11` | Breaks offline; no SRI — needs bundling into `web/sdk/` |
| CSRF protection absent | 🟠 High | Fiber middleware | Cookie auth + no CSRF token |

## Fixed (2026-05-18)

| Issue | Was | Fix |
|-------|-----|-----|
| `API_BASE` hardcoded to localhost | 🔴 Critical | Removed; fetcher uses relative URLs |
| 401 → login redirect missing | 🔴 Critical | Added `window.location.href = '/login'` on 401 |
| XSS in breadcrumb (`innerHTML`) | 🔴 Critical | Replaced with `createElement`+`textContent` |
| `alert()` native dialog in amisEnv | 🟠 High | Replaced with `amisEnv.notify('error', msg)` |
| Sidebar collapses on main click | 🟡 Medium | Removed `#main` click→collapseSidebar listener |
| `options: json.data` in list bridge | 🟡 Medium | Removed — conflated list and select-options responses |
| `max-height: 500px` menu ceiling | 🟡 Medium | Bumped to `2000px` — no group clips |
| Sidebar search cosmetic only | 🟡 Medium | Real `<input>` + live JS filter wired in |

---

## Not Implemented

See [Planned Features](02-planned.md) for roadmap items.

| Feature | Why Not Built |
|---------|--------------|
| Notification system | Awaiting product spec finalization |
| Saved views / filter persistence | Requires user preference storage layer |
| Keyboard shortcuts | Deferred — not blockers for MVP |
| Go-driven navigation (AMIS `app`) | Migration trigger met — deferred to sprint planning |
| Per-user module ordering | Depends on navigation migration |
| 5-persona detection system | Current 2-mode (platform/portal) sufficient for now |
| Tenant-configurable branding | Deferred post-launch |
