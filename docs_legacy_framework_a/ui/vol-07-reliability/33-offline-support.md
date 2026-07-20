> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
chapter: 33
title: "Offline Support"
volume: "vol-07-reliability"
section: "Reliability"
description: "Planned offline and service worker support for schema caching and graceful degradation when the server is unreachable."
status: planned
---

# Chapter 33 — Offline Support

> **Status: PLANNED — nothing in this chapter is implemented.**
>
> The UI platform currently requires a live server connection for every navigation. This chapter describes the intended offline capabilities.

## Table of Contents

- [33.1 Current Behavior Without Connectivity](#331-current-behavior-without-connectivity)
- [33.2 PLANNED: Service Worker Schema Cache](#332-planned-service-worker-schema-cache)
- [33.3 PLANNED: Stale-While-Revalidate for Schemas](#333-planned-stale-while-revalidate-for-schemas)
- [33.4 PLANNED: Offline Data Drafts](#334-planned-offline-data-drafts)
- [33.5 PLANNED: Connectivity Detection and UI Indicators](#335-planned-connectivity-detection-and-ui-indicators)
- [33.6 Design Constraints](#336-design-constraints)

---

## 33.1 Current Behavior Without Connectivity

When the schema server is unreachable:

- The AMIS app shell attempts to fetch the schema for the current route.
- The request times out or fails with a network error.
- The page content area shows a blank state or a generic browser-level error.
- No user-friendly error message is displayed.
- No cached version of the page is available.

The application is effectively unusable without connectivity. This is acceptable for an internal enterprise application in a stable network environment but becomes a significant problem for mobile users, field workers, or sites with unreliable connectivity.

---

## 33.2 PLANNED: Service Worker Schema Cache

A service worker would intercept all `GET /schema/*` requests and serve cached schema responses when the network is unavailable.

Intended behavior:
- On first successful schema load, the service worker stores the response body (the AMIS schema JSON) in the Cache API.
- Cache key: route URL + the JWT fingerprint (so different users don't share cached schemas in the browser).
- On subsequent loads, the service worker checks for a cached version before making a network request.
- If the network is unavailable, the service worker serves the cached schema.

The cached schema is read-only — it reflects the page as it was when last fetched. The user can navigate and view data, but the data itself (from AMIS API calls to the backend) is not cached by the service worker.

---

## 33.3 PLANNED: Stale-While-Revalidate for Schemas

For frequently accessed pages, the stale-while-revalidate pattern would provide instant navigation while silently updating the cached schema in the background.

Intended behavior:
1. User navigates to `finance/invoices`.
2. Service worker immediately serves the cached schema (instant render).
3. In the background, service worker makes the network request.
4. When the response arrives, the service worker updates the cache.
5. If the new schema differs significantly from the cached schema, the app shows a "Page updated — refresh to see changes" banner.

This requires a mechanism to detect meaningful schema changes (a content hash or an ETag header — see §35.2 for planned ETag support).

---

## 33.4 PLANNED: Offline Data Drafts

Schema caching addresses the UI layer, but the data layer (AMIS API calls) is independent. For write operations (creating or editing records), offline support requires:

- Storing form data locally (IndexedDB) when the network is unavailable.
- Syncing drafts to the server when connectivity is restored.
- Conflict resolution when the server record was modified while the draft was pending.

This is significantly more complex than schema caching because it requires:
- A draft data model on the server.
- Conflict detection and merge strategies.
- UI indication of draft vs. synced state.

Offline data drafts are further out on the roadmap than schema caching.

---

## 33.5 PLANNED: Connectivity Detection and UI Indicators

A connectivity-aware layer in the app shell would:
- Listen for browser `online`/`offline` events.
- Display a persistent banner when the connection is lost.
- Disable write-action buttons (create, update, delete) while offline (read-only mode).
- Re-enable writes when connectivity is restored.
- Automatically retry failed schema fetches when the connection comes back.

Without this layer, users have no indication that they are viewing a stale cached schema — they may attempt to submit forms that silently fail.

---

## 33.6 Design Constraints

Any offline implementation must respect these constraints:

1. **Security**: Cached schemas must not be served across user sessions. The browser cache must be cleared when the user logs out. The service worker must include the user identity in the cache key.

2. **Permission correctness**: A schema cached while the user had elevated permissions must not be served after those permissions are revoked. The service worker should include the permission fingerprint in the cache key and validate the fingerprint on revalidation.

3. **Feature flag correctness**: Same as permissions — flag fingerprint must be part of the service worker cache key.

4. **Maximum staleness**: Cached schemas should have a maximum age (e.g., 24 hours) after which the service worker refuses to serve them even offline and shows a "schema too stale" error.

5. **Tenant isolation**: Browser-cached schemas from Tenant A must not be served to a user who subsequently logs into Tenant B on the same device.
