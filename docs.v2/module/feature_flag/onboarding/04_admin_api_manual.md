# 4. Administrator's API Manual

This document provides a manual for the Feature Flag Admin API, designed for System Administrators, SREs, and the Operations team. These endpoints provide the tools to manage the system safely and at scale.

**Base Path**: `/api/v1/admin/feature-flags`  
**Authentication**: All endpoints require a valid JWT with appropriate scopes.

---

## Bulk Operations

Bulk endpoints are essential for managing multiple flags at once, such as during a large feature launch or an incident response. They are designed to be atomic and provide detailed feedback.

### Bulk Enable / Disable Flags

-   **Endpoints**: `POST /bulk/enable`, `POST /bulk/disable`
-   **Purpose**: To enable or disable a list of feature flags in a single atomic operation.
-   **When to Use**:
    -   Coordinating the launch of multiple related features.
    -   Disabling a set of features during a production incident to mitigate impact.
    -   Automating feature state changes as part of a CI/CD pipeline.

**Request Body**:
```json
{
  "flag_names": ["feature-one", "feature-two"],
  "reason": "Disabling for incident response INC-5821"
}
```

**Key Response Fields**:
-   `successful` / `failed`: A quick summary of the outcome.
-   `results`: A detailed array showing the outcome for each individual flag. This is critical for debugging partial failures.
-   `summary.error_categories`: Groups failures by type (e.g., `not_found`) so you can quickly diagnose the root cause.

---

## System Health & Monitoring

### Check System Health

-   **Endpoint**: `GET /health`
-   **Purpose**: To get a real-time,  snapshot of the feature flag system's health.
-   **When to Use**:
    -   As a primary endpoint for automated monitoring and alerting systems (e.g., Prometheus, Datadog).
    -   As the first step in troubleshooting any issue related to feature flags.

**Key Response Fields**:
-   `status`: The overall health status (`healthy`, `degraded`, `unhealthy`).
-   `component_health`: A breakdown of the health of each critical dependency (database, cache). This allows you to pinpoint the source of a problem.
-   `overall_score`: A numerical score (0-100) representing system health, useful for tracking trends over time.
-   `recommended_actions`: Plain-English suggestions for what to do if the system is not healthy.

---

## Emergency Controls

These endpoints are the system's "kill switches" and should be used with caution. They are designed for rapid incident response.

### Emergency Disable All Flags

-   **Endpoint**: `POST /emergency/disable-all`
-   **Purpose**: To immediately disable all active feature flags for the current tenant. This is the primary tool to stop system-wide "bleeding" caused by one or more new features.
-   **When to Use**:
    -   During a major production incident where new features are the suspected cause.
    -   When a critical security vulnerability is discovered that needs to be mitigated immediately.

**Request Body**:
```json
{
  "reason": "Critical performance degradation on main database cluster."
}
```

**Key Response Fields**:
-   `affected_flags`: Tells you how many flags were turned off.
-   `rollback_token`: **Crucially, this token can be used to safely re-enable the exact set of flags that were disabled by this operation.** This prevents accidental enabling of other features and allows for a controlled recovery.
-   `estimated_impact`: A human-readable assessment of how this action will affect users.

---

## Cache Management

The system relies heavily on caching for performance. These endpoints provide control over the cache's state.

### Get Cache Statistics

-   **Endpoint**: `GET /cache/stats`
-   **Purpose**: To inspect the performance and state of the Redis cache.
-   **When to Use**:
    -   When diagnosing performance issues (a low hit rate can indicate a problem).
    -   For capacity planning (monitoring memory usage and eviction rates).

### Clear Cache

-   **Endpoint**: `POST /cache/clear`
-   **Purpose**: To selectively or completely invalidate the cache for one or more tenants.
-   **When to Use**:
    -   After a manual database change to ensure stale data is not being served.
    -   As a troubleshooting step if you suspect a "stuck" flag value.
