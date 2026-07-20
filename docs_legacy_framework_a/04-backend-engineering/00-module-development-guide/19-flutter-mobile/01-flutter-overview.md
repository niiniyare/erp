> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Flutter Mobile Schemas (Planned)
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, mobile-engineer]
related:
  - "[UI Schema Design](../17-ui-schema/01-ui-schema-overview.md)"
  - "[amis Web Schemas](../18-testing/01-testing-overview.md)"
---

# Flutter Mobile Schemas

> **Status: PLANNED** — Flutter mobile support is not yet implemented. This section will cover schema-driven mobile UI generation when the Flutter client is built.

## Planned Approach

The mobile client will use the same server-driven schema pattern as the web client:

- Backend serves Flutter widget tree descriptions as JSON
- Mobile client renders them using a Flutter schema interpreter
- Same API endpoints, same auth, same RBAC

## Schema Endpoint (Planned)

```
GET /api/v1/{module}/schemas/mobile/{page}
```

The `mobile` segment distinguishes mobile schemas from web (AMIS) schemas. Mobile schemas describe Flutter widget trees rather than AMIS component trees.

## When This Will Be Built

Flutter support will be scoped when:
1. Web UI reaches feature completeness
2. Mobile use cases are validated with early adopters
3. Flutter schema interpreter library is selected or built

## Temporary Guidance

Until mobile support is implemented, mobile access goes through the REST API directly (same endpoints as web, same auth). Native mobile UIs can be built against the REST API without waiting for server-driven schema support.
