> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
chapter: 35
title: "Synchronization"
volume: "vol-07-reliability"
section: "Reliability"
description: "Current stateless schema fetch model, and planned event-driven cache invalidation and real-time schema push."
status: partially-implemented
---

# Chapter 35 — Synchronization

## Table of Contents

- [35.1 Current Model: Stateless Schema Fetch](#351-current-model-stateless-schema-fetch)
- [35.2 PLANNED: Event-Driven Cache Invalidation](#352-planned-event-driven-cache-invalidation)
- [35.3 PLANNED: Real-Time Schema Push](#353-planned-real-time-schema-push)
- [35.4 PLANNED: Schema ETags and Conditional Requests](#354-planned-schema-etags-and-conditional-requests)
- [35.5 PLANNED: Schema Delta / Patch Streaming](#355-planned-schema-delta--patch-streaming)

---

## 35.1 Current Model: Stateless Schema Fetch

The current schema delivery model is entirely stateless and pull-based. There is no persistent connection between the browser and the schema server.

### How Navigation Works Today

1. User clicks a nav link or the app shell routes to a new page.
2. The browser makes `GET /schema/<route>` with the current JWT.
3. The server runs the pipeline (cache hit or miss), returns the schema JSON.
4. AMIS renders the schema.
5. AMIS makes its own data API calls to populate tables, forms, and charts.

Steps 1–5 repeat on every navigation. The schema is re-fetched every time, with no client-side schema cache. If the user navigates back to a page they visited 30 seconds ago, a fresh `GET /schema/<route>` is issued.

### Implications

- **No stale schema risk in the browser**: The user always gets the most recently compiled schema.
- **Latency on every navigation**: Every page transition incurs at least a cache lookup round-trip (2–5ms on a hit, 20–100ms on a miss).
- **No background update mechanism**: If a schema changes while the user is looking at it, the user sees the change only on their next navigation.
- **No real-time capability**: There is no mechanism to push schema updates to connected users.

For the current use case (internal enterprise application, stable schemas, users navigating between pages), the stateless model is sufficient. The planned features in this chapter become important when schema changes are frequent or when real-time UI updates are required.

---

## 35.2 PLANNED: Event-Driven Cache Invalidation

> **Status: PLANNED — not implemented.**

Currently, cache invalidation is triggered manually via `cache.Service.DeletePattern`. This requires an operational action (a deployment, an admin API call, or a script) to clear stale entries.

The planned improvement connects cache invalidation to the domain event bus. When a domain event that affects the schema is published, a subscriber automatically invalidates the relevant cache entries.

### Intended Event-to-Invalidation Mappings

| Domain Event                    | Invalidation Scope    | Pattern                      |
|---------------------------------|-----------------------|------------------------------|
| `RoleAssignmentChanged`         | Tenant-level          | `*:<tenant_id>:*`            |
| `FeatureFlagUpdated`            | Tenant-level          | `*:<tenant_id>:*`            |
| `TenantSettingsChanged`         | Tenant-level          | `*:<tenant_id>:*`            |
| `ModuleSchemaUpdated`           | Module-level          | `<module>/*:*`               |
| `PermissionPolicyChanged`       | Global (bump version) | `*`                          |
| `TenantCurrencyChanged`         | Tenant-level          | `*:<tenant_id>:*`            |

The event subscriber would be a lightweight service registered on the internal event bus. It would consume events, determine the invalidation scope, and call `cache.DeletePattern` accordingly.

### Benefits Over Manual Invalidation

- **Self-healing**: Stale schemas are automatically cleared when the relevant data changes, without operator intervention.
- **Precise scoping**: An event knows exactly which tenant or module changed, enabling targeted invalidation rather than global flushes.
- **Audit trail**: Event-driven invalidation events are logged through the normal event bus audit trail.

### Open Questions

- How does the event subscriber handle event delivery failures? If the subscriber crashes during a cache invalidation, entries may remain stale until TTL expiry.
- Should invalidation be synchronous (blocking the event handler until the delete completes) or asynchronous (fire-and-forget)?
- What is the maximum acceptable propagation delay between a domain change and cache invalidation?

---

## 35.3 PLANNED: Real-Time Schema Push

> **Status: PLANNED — not implemented.**

Real-time schema push would allow the server to notify connected browser clients when a schema they are currently viewing has been updated, without requiring the user to navigate away and back.

### Intended Mechanism

A Server-Sent Events (SSE) or WebSocket connection would be maintained between the app shell and a lightweight schema notification service. When a schema is invalidated (via the event-driven system from §35.2), the notification service sends a message to all connected clients who are currently viewing the affected route.

```
Server                    Client
  │                         │
  │  ←── SSE connection ───  │  (maintained in background)
  │                         │
  │  [RoleAssignment event] │
  │  ──── schema_updated ─▶  │  {"route": "finance/invoices", "tenant": "tenant-x"}
  │                         │
  │                         │  App shell receives message
  │                         │  → Shows "This page has been updated" banner
  │                         │  → On user click: re-fetches GET /schema/finance/invoices
```

### Push vs. Forced Refresh

The planned model is advisory push, not forced refresh. The server notifies the client that a schema has changed; the client decides when to fetch the new version. Forced refresh (the server sending the new schema directly) is more complex and risks overwhelming clients during high-churn periods.

### Scope

Real-time push is most valuable for:
- Admin-facing screens where permission changes take effect immediately.
- Screens where the schema changes frequently (e.g., feature flag rollouts).
- Long-lived sessions where users leave a tab open for hours.

For typical navigation patterns (user moves between pages every few seconds), real-time push provides little value over the stateless re-fetch model.

---

## 35.4 PLANNED: Schema ETags and Conditional Requests

> **Status: PLANNED — not implemented.**

HTTP ETags would allow the browser to make conditional schema requests: "give me the schema for `finance/invoices` only if it has changed since I last fetched it."

```
// First request
GET /schema/finance/invoices
→ 200 OK
   ETag: "abc123"
   Body: { ... schema ... }

// Subsequent request
GET /schema/finance/invoices
   If-None-Match: "abc123"
→ 304 Not Modified
   (no body)
```

On a 304 response, the browser uses its cached schema without re-parsing the full JSON body. For large schemas (tens of kilobytes of JSON), this significantly reduces network transfer and parse time.

The ETag value would be the cache key (or a hash of it), which already encodes the schema version.

ETags interact with the service worker plan (§33.3): the stale-while-revalidate pattern uses ETags to detect whether a background revalidation fetched a new schema or the same one.

---

## 35.5 PLANNED: Schema Delta / Patch Streaming

> **Status: PLANNED — not implemented.**

Instead of returning the full schema JSON on every request, a delta/patch API would return only the parts of the schema that changed since the client's last version.

```
GET /schema/finance/invoices?since=schema-version-12
→ 200 OK
   Content-Type: application/json-patch+json
   Body: [
     {"op": "replace", "path": "/body/2/columns/4/label", "value": "Tax Amount"},
     {"op": "add", "path": "/body/0/toolbar/-", "value": { ... new button ... }}
   ]
```

The client applies the patch to its locally held schema, producing the new version without re-parsing the full schema.

This optimization is most valuable for:
- Large, complex schemas (100+ KB JSON) that change incrementally.
- Mobile clients on slow connections.
- Screens with real-time collaborative editing requirements.

The patch streaming feature requires:
1. Schema versioning on the server (each compiled schema gets a monotonic version number).
2. A diff algorithm that can compare two AMIS schema trees and produce a minimal JSON Patch document.
3. Client-side patch application logic.
4. A version reconciliation protocol for cases where the client's base version is too old (fall back to full schema).

Given the current stateless model, this is a significant architectural addition and is planned for a future phase after real-time push (§35.3) is established.
