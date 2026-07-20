> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Health and Readiness
portal: 4 — Backend Engineering
section: 00-module-development-guide/08-instrumentation
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-instrumentation-overview.md
    title: Instrumentation Overview
---

# Health and Readiness

Health and readiness probes let orchestration platforms (Kubernetes, Railway) determine whether the service is ready to receive traffic.

## Endpoint Overview

| Endpoint | Purpose | Response |
|----------|---------|---------|
| `GET /health/live` | Liveness: is the process running? | `200 OK` always (if process is up) |
| `GET /health/ready` | Readiness: is the service ready for traffic? | `200 OK` if all deps OK, `503` if not |
| `GET /metrics` | Prometheus metrics scrape | Prometheus text format |

## Readiness Check

The readiness handler checks that all critical dependencies are reachable:

```go
// internal/api/handlers/health/handler.go
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	checks := map[string]error{
		"database":   h.checkDatabase(c.Context()),
		"cache":      h.checkCache(c.Context()),
		"temporal":   h.checkTemporal(c.Context()),
	}

	for name, err := range checks {
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "not ready",
				"check":  name,
				"error":  err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{"status": "ready"})
}
```

## Module-Level Health Contribution

Business modules do not expose their own health endpoints. They contribute to the application-level readiness check by ensuring their dependencies (DB queries, cache access) work correctly.

If a module has a background task that must be running (e.g., a Temporal worker), check its health in the readiness endpoint:

```go
// In the readiness check
"contracts_worker": h.temporalClient.CheckWorkerHealth("contracts-task-queue"),
```

## Metrics Endpoint

The `/metrics` endpoint is served by the Prometheus middleware registered in `routes.go`. No module-specific setup needed — all `s.metrics.*` calls are automatically collected and exposed.
