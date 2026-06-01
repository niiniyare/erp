---
title: API Versioning
portal: 4 — Backend Engineering
section: 00-module-development-guide/14-api-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-api-design-overview.md
    title: API Design Overview
---

# API Versioning

## Version Strategy

URLs carry the version prefix (`/api/v1/`, `/api/v2/`). Header-based versioning is not used.

```
/api/v1/contracts  — stable, supported
/api/v2/contracts  — new version with breaking changes (when needed)
```

## What Constitutes a Breaking Change

| Change | Breaking? | Action |
|--------|-----------|--------|
| Add optional field to response | No | Deploy freely |
| Add optional field to request | No | Deploy freely |
| Remove field from response | **Yes** | New version |
| Rename field in request or response | **Yes** | New version |
| Change field type (e.g., int → string) | **Yes** | New version |
| Change status code (e.g., 200 → 201) | **Yes** | New version |
| Remove endpoint | **Yes** | Deprecate, then new version |
| Add required field to request | **Yes** | New version |

## Deprecation Process

1. Add `Deprecation` and `Sunset` response headers to the deprecated version
2. Log a warning per-request on deprecated endpoints
3. Communicate to clients via changelog
4. Remove after Sunset date (minimum 6 months from announcement)

```go
// Deprecated v1 endpoint
func (h *v1Handler) ListContracts(c *fiber.Ctx) error {
    c.Set("Deprecation", "true")
    c.Set("Sunset", "2026-01-01T00:00:00Z")
    c.Set("Link", `</api/v2/contracts>; rel="successor-version"`)

    h.log.Warn().
        Str("path", c.Path()).
        Str("client_ip", c.IP()).
        Msg("deprecated API v1 endpoint called")

    // delegate to v1 logic
    return h.listContractsV1(c)
}
```

## Parallel Registration

Both versions registered in Fiber simultaneously:

```go
v1 := app.Group("/api/v1")
v1.Get("/contracts", v1Handlers.ListContracts)

v2 := app.Group("/api/v2")
v2.Get("/contracts", v2Handlers.ListContracts)
```

Each version has its own handler set and DTO types. They may share service layer code if the changes are response-format only.
