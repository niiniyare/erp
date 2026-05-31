# Awo ERP — API, Multi-Tenancy & UI Architecture Guide

**Document Series:** Awo ERP Technical Reference
**Version:** 1.0.0
**Stack:** Go · Fiber · PostgreSQL · Redis · Temporal · amis-ui
**Audience:** Technical Engineers, Platform Operators, and Non-Technical Stakeholders

---

## Table of Contents

- [PART I — FOUNDATIONS](#part-i--foundations)
  - [1. Introduction & Purpose](#1-introduction--purpose)
  - [2. Glossary of Terms](#2-glossary-of-terms)
  - [3. System Overview](#3-system-overview)
- [PART II — MULTI-TENANCY ARCHITECTURE](#part-ii--multi-tenancy-architecture)
  - [4. What is Multi-Tenancy and Why Awo ERP Chose It](#4-what-is-multi-tenancy-and-why-awo-erp-chose-it)
  - [5. Tenant Identity — The X-Tenant-ID Header](#5-tenant-identity--the-x-tenant-id-header)
- [PART III — SUBDOMAIN ROUTING](#part-iii--subdomain-routing)
  - [6. Subdomain Architecture](#6-subdomain-architecture)
  - [7. The Routing Decision Tree](#7-the-routing-decision-tree)
- [PART IV — TENANT ONBOARDING](#part-iv--tenant-onboarding)
  - [8. Tenant Lifecycle](#8-tenant-lifecycle)
  - [9. The Onboarding Flow — End to End](#9-the-onboarding-flow--end-to-end)
- [PART V — PLATFORM USERS](#part-v--platform-users)
  - [10. Platform Users vs. Tenant Users](#10-platform-users-vs-tenant-users)
  - [11. Platform User Roles](#11-platform-user-roles)
  - [12. Platform User API Surface](#12-platform-user-api-surface)
- [PART VI — THE amis-UI LAYER](#part-vi--the-amis-ui-layer)
  - [13. What is amis and Why Awo ERP Chose It](#13-what-is-amis-and-why-awo-erp-chose-it)
  - [14. The amis API Response Contract](#14-the-amis-api-response-contract)
  - [15. Tenant-Aware UI Rendering](#15-tenant-aware-ui-rendering)
  - [16. The Schema Request Lifecycle](#16-the-schema-request-lifecycle)
  - [17. Security Implications of Backend-Driven UI](#17-security-implications-of-backend-driven-ui)
- [PART VII — SECURITY](#part-vii--security)
  - [18. Security Architecture Overview](#18-security-architecture-overview)
  - [19. Authentication — Session-Based Identity](#19-authentication--session-based-identity)
  - [20. Authorisation & Tenant Isolation](#20-authorisation--tenant-isolation)
  - [21. Header Security](#21-header-security)
  - [22. Transport Security](#22-transport-security)
  - [23. Rate Limiting and Abuse Prevention](#23-rate-limiting-and-abuse-prevention)
  - [24. Secret Management and Key Rotation](#24-secret-management-and-key-rotation)
- [PART VIII — CACHING STRATEGY](#part-viii--caching-strategy)
  - [25. Redis in the Awo ERP Stack](#25-redis-in-the-awo-erp-stack)
- [PART IX — TROUBLESHOOTING](#part-ix--troubleshooting)
  - [26. Common Tenant-Routing Failures](#26-common-tenant-routing-failures)
  - [27. amis-UI Failures](#27-amis-ui-failures)
  - [28. Onboarding Failures](#28-onboarding-failures)
  - [29. Platform User Access Issues](#29-platform-user-access-issues)
  - [30. Performance Troubleshooting](#30-performance-troubleshooting)
  - [31. Observability Checklist](#31-observability-checklist)
- [APPENDICES](#appendices)

---

# PART I — FOUNDATIONS

---

## 1. Introduction & Purpose

### 1.1 What This Document Is

This guide is the canonical reference for how Awo ERP's API layer works — how it accepts requests, identifies which business it is talking to, constructs a response that includes both data and user interface instructions, and does all of this securely and at scale.

It is written for two audiences simultaneously. Non-technical readers — business owners, operations managers, and implementation partners — will find plain-language explanations of *why* things work the way they do. Technical readers — Go engineers, DevOps engineers, and database administrators — will find architecture rationale, code patterns, and deep-dive comparisons with alternative approaches.

When a section is primarily technical, it is marked **[Technical]**. Sections without that label are written for general readability first.

### 1.2 What Problems This Document Solves

Awo ERP serves multiple businesses from a single running system. A petroleum retailer in Nairobi and a restaurant group in Mombasa might both use Awo ERP, and the system must never allow data or interface configurations from one to leak into the other. At the same time, the people who *build and maintain* Awo ERP — platform engineers, support agents, billing staff — need their own way to access the system that is separate from any individual tenant.

This document explains exactly how all of that is accomplished, what the security implications are, and what to do when something goes wrong.

### 1.3 Scope

This document covers:

- The API layer built with Go and the Fiber web framework
- Multi-tenancy implemented via PostgreSQL Row-Level Security
- Session-based authentication stored in Redis
- Subdomain-based routing and tenant resolution
- The `X-Tenant-ID` request header and how it interacts with subdomains
- Tenant onboarding orchestrated by Temporal
- The amis-ui backend-driven rendering system
- Platform (internal) users and their separation from tenant users
- Security architecture across all layers
- Caching strategy with Redis
- Troubleshooting guides and observability checklists

---

## 2. Glossary of Terms

Understanding Awo ERP's architecture requires clarity on a set of terms that have precise meanings in this system. This glossary is the single source of truth.

| Term | Definition |
|---|---|
| **Tenant** | A single business entity using Awo ERP. For example, "Shell Maanzoni Service Station" operated by Anika Global Limited is one tenant. |
| **Platform** | Awo ERP itself — the company and system that hosts and supports all tenants. Also used as an adjective: "platform users," "platform API." |
| **Multi-Tenancy** | The architectural property by which a single deployment of Awo ERP simultaneously serves multiple tenants while keeping their data and configurations completely separate. |
| **Subdomain** | The `shell-maanzoni` part of `shell-maanzoni.awo.app`. Each tenant gets a unique subdomain when they are onboarded. |
| **X-Tenant-ID** | An HTTP request header that carries the unique identifier of the tenant making the request. Used together with the subdomain to doubly verify tenant identity. |
| **Session** | A server-side record in Redis that proves a user is authenticated. The browser holds a session ID (cookie); the server holds the session data. |
| **Session ID** | A long, cryptographically random string stored in the user's browser as a cookie. It is the key that unlocks the session data in Redis. |
| **RLS** | Row-Level Security. A PostgreSQL feature that restricts which database rows a query can see based on a policy, automatically enforcing tenant isolation at the database layer. |
| **amis** | An open-source framework by Baidu that renders complex user interfaces from JSON configuration. Awo ERP uses it to build the entire UI from the backend. |
| **amis Schema** | A JSON document that tells amis what to render — forms, tables, charts, buttons, and all their properties. Awo ERP generates these per-tenant on the server. |
| **Temporal** | A workflow orchestration platform. Used in Awo ERP for long-running, multi-step processes like tenant onboarding where each step must be reliable and retryable. |
| **Fiber** | A high-performance Go web framework built on top of `fasthttp`. Awo ERP's API layer is built with Fiber. |
| **Middleware** | Code that runs on every request before it reaches the actual handler. Awo ERP uses middleware for session validation, tenant resolution, and rate limiting. |
| **COA** | Chart of Accounts. The financial backbone of a tenant's accounting structure, seeded during onboarding. |
| **Platform User** | An Awo ERP employee or operator who manages the platform itself — not a tenant employee. |
| **Tenant User** | A user who belongs to and operates within a specific tenant's account. |
| **Impersonation** | The act of a platform user temporarily viewing or acting as a tenant user for support purposes, with full audit logging. |
| **Outbox Pattern** | A database pattern used by Awo ERP's notification system to guarantee that events are eventually processed, even if a downstream service is temporarily down. |
| **Industry Vertical** | A business category that determines which default modules, COA structure, and amis schemas are seeded for a tenant. Examples: Petroleum Retail, Restaurant, Retail, Aviation. |
| **WAC** | Weighted Average Cost — a fuel inventory costing method relevant to petroleum tenants. |
| **Wildcard Certificate** | A TLS certificate that covers all subdomains of a domain, e.g., `*.awo.app`. |

---

## 3. System Overview

### 3.1 The Big Picture (Non-Technical)

Imagine Awo ERP as a large office building. Each tenant is a company that rents a floor of the building. They have their own entrance (their subdomain), their own locked door (the session system), and their own filing cabinets (their database rows). The building management — Awo ERP's platform team — has master keys that let them access any floor for legitimate maintenance, but every such visit is logged in a visitors' book.

When someone walks in through the Shell Maanzoni entrance (`shell-maanzoni.awo.app`), the receptionist (the API middleware) checks two things: the entrance they used, and their staff ID badge (the `X-Tenant-ID` header). If both match, they are let in. If someone tries to walk in through the Shell Maanzoni entrance with a badge from a different company, they are turned away immediately.

Once inside, the person sits at a computer that shows them exactly the forms, reports, and actions they are allowed to see — not because the computer has been physically different for each person, but because the *software* on it reads a set of instructions from the server that says "show this person these screens with these fields." That is what amis does.

### 3.2 Component Map

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLIENT BROWSER                               │
│  shell-maanzoni.awo.app  ──────────────────────────────────────────  │
│  Cookie: session_id=abc123                                           │
│  Header: X-Tenant-ID: <tenant-uuid>                                  │
└─────────────────────────┬───────────────────────────────────────────┘
                          │ HTTPS
                          ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    REVERSE PROXY / GATEWAY                           │
│   (Nginx / Caddy / Cloud Load Balancer)                              │
│   - Terminates TLS                                                   │
│   - Strips/validates internal headers                                │
│   - Forwards subdomain as X-Forwarded-Host                          │
└─────────────────────────┬───────────────────────────────────────────┘
                          │ HTTP (internal)
                          ▼
┌─────────────────────────────────────────────────────────────────────┐
│              AWO ERP API SERVER (Go + Fiber)                         │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  MIDDLEWARE CHAIN                                             │   │
│  │  1. SubdomainExtractor  → reads shell-maanzoni from host     │   │
│  │  2. TenantResolver      → looks up tenant in Redis/PG        │   │
│  │  3. HeaderValidator     → confirms X-Tenant-ID matches       │   │
│  │  4. SessionMiddleware   → validates session_id in Redis      │   │
│  │  5. RateLimiter         → per-tenant token bucket in Redis   │   │
│  │  6. RouteHandler        → business logic                     │   │
│  └──────────────────────────────────────────────────────────────┘   │
└───────────┬─────────────────────┬───────────────────────────────────┘
            │                     │
            ▼                     ▼
┌────────────────────┐  ┌──────────────────────────────────────────┐
│  REDIS             │  │  POSTGRESQL                               │
│  - Sessions        │  │  - All tenant data (RLS enforced)        │
│  - Tenant cache    │  │  - amis schema definitions               │
│  - Rate limiters   │  │  - Session metadata (optional mirror)    │
│  - amis schemas    │  │  - Audit logs                            │
└────────────────────┘  └──────────────────────┬───────────────────┘
                                               │
                         ┌─────────────────────┘
                         ▼
              ┌──────────────────────┐
              │  TEMPORAL            │
              │  - Onboarding flows  │
              │  - Notification jobs │
              │  - Retry/durability  │
              └──────────────────────┘
```

### 3.3 The Awo ERP Philosophy — Backend-Driven UI in an ERP Context

Most web applications are built as two completely separate projects: a backend API that serves JSON data, and a frontend application (React, Vue, Angular) that knows how to display that data. Every time you want to add a new field to a form or change what a user can see, a frontend developer has to write code, test it, and deploy it.

Awo ERP inverts this model. The backend does not just send data — it sends *instructions on how to display the data*. The amis-ui framework on the browser then interprets those instructions and renders the correct interface automatically. A new document type, a new field, a new permission level — all of these can be implemented purely by changing what the Go backend returns, with no frontend deployment required.

This is a deeply practical choice for an ERP system serving diverse industries. A petroleum retailer needs shift reconciliation forms and dip reading tables. A restaurant needs table management and recipe costing. Rather than maintaining separate frontends, Awo ERP's Go backend generates the right amis JSON for each tenant's industry vertical.

### 3.4 Where This Document Fits in the Awo ERP Specification Series

This document is one in a series of technical specifications for Awo ERP. It serves as the horizontal architectural reference, while other documents in the series go deep on individual modules (Finance Module, Notifications Module, Forecourt Module). When those module documents refer to "the API contract" or "tenant resolution," they are deferring to this document.

---

# PART II — MULTI-TENANCY ARCHITECTURE

---

## 4. What is Multi-Tenancy and Why Awo ERP Chose It

### 4.1 The Three Classic Isolation Models

Before explaining what Awo ERP chose, it is important to understand the landscape of options. In any system that serves multiple customers, there are three fundamental ways to isolate their data:

**Model 1 — Isolated Database per Tenant**

Each tenant gets their own completely separate database. Shell Maanzoni has a database called `awo_shell_maanzoni`. A new restaurant tenant would get `awo_ocean_grill`.

*Advantages:* Maximum isolation. A bug that corrupts one tenant's data cannot touch another's. Database backups are per-tenant. Compliance (e.g., GDPR right to erasure) is trivial — delete the database.

*Disadvantages:* Operationally nightmarish at scale. Running 100 tenants means managing 100 database servers or schemas, applying schema migrations 100 times, monitoring 100 connection pools. For a system like Awo ERP targeting 50–500 East African enterprises, the overhead per tenant would make pricing uncompetitive.

**Model 2 — Shared Database, Separate Schema**

All tenants share one PostgreSQL instance, but each gets their own schema (a PostgreSQL namespace). Shell Maanzoni's tables live in the `shell_maanzoni` schema; the restaurant's tables live in `ocean_grill`.

*Advantages:* A middle ground. Migration is still applied per schema but to the same database cluster. Isolation is good but not perfect.

*Disadvantages:* PostgreSQL schemas are not a strong isolation boundary the way separate databases are. Connection pool management is complex. Schema proliferation (hundreds of schemas) causes PostgreSQL catalog bloat that degrades query planning.

**Model 3 — Shared Database, Shared Schema (Row-Level Security)**

All tenants share one database and one schema. Every table has a `tenant_id` column. PostgreSQL Row-Level Security policies automatically filter every query to only return rows belonging to the current tenant. This is the model Awo ERP uses.

*Advantages:* Operationally simple. One migration run updates all tenants simultaneously. One connection pool serves all tenants. Adding a new tenant is a matter of inserting a record — no schema or database provisioning needed.

*Disadvantages:* Security requires careful implementation. A bug in the RLS policy or a misconfigured database role could theoretically allow cross-tenant data access. This risk is real and addressed in depth in Part VII.

### 4.2 Why Awo ERP Uses Shared Database + Row-Level Security

The decision came down to operational reality. Awo ERP is building for a market where pricing pressure is real, where the platform team is lean, and where the ability to ship schema changes quickly (adding a new field to the `fuel_deliveries` table for all petroleum tenants at once) is a core competitive advantage.

Isolated databases would mean that every schema migration is a multi-step operation across potentially hundreds of databases. That introduces deployment risk proportional to the number of tenants. Shared schema with RLS means migrations are applied once, in a single transaction, with all the standard PostgreSQL migration tooling working normally.

PostgreSQL RLS, when implemented correctly, is a mature and battle-tested isolation mechanism. It was introduced in PostgreSQL 9.5 (2016) and is used in production by systems handling far more sensitive data than ERP records.

### 4.3 Trade-offs Accepted and Trade-offs Avoided

**Trade-offs accepted:**

- RLS must be applied to every table, without exception. There is no safety net if a table is created without a policy. Awo ERP mitigates this with a CI/CD check that fails any migration that creates a table without a corresponding RLS policy.
- All tenants share the same resource pool (CPU, memory, I/O). A single tenant running very heavy reports could degrade performance for others. Awo ERP mitigates this with per-tenant rate limiting at the API layer and query timeout policies at the PostgreSQL layer.
- Cross-tenant analytics (e.g., platform-level reporting across all petroleum tenants' fuel volumes) require a special superuser role that bypasses RLS. Access to this role is tightly controlled.

**Trade-offs avoided:**

- Awo ERP deliberately does NOT use application-level `WHERE tenant_id = $1` filtering as the primary security control. Application code is fallible. RLS is enforced by the database engine regardless of what the application does. The application-level filtering is a secondary, belt-and-suspenders check.

### 4.4 How Well-Known Systems Handle This

Understanding how established platforms approach multi-tenancy gives important context for Awo ERP's choices.

**Salesforce** uses a shared database, shared schema model — exactly like Awo ERP — and is probably the most cited example of enterprise multi-tenancy at scale. They call their approach "metadata-driven" multi-tenancy. Their tenant isolation is enforced through a combination of application logic and database constraints. Salesforce serves hundreds of thousands of tenants from this model.

**Shopify** uses a sharded approach: tenants are distributed across many database shards, but within each shard, it is a shared schema model. This is a pragmatic evolution — start with shared schema, shard when you need horizontal scale. Awo ERP is not at Shopify's scale and does not need sharding, but the architecture does not preclude adding it later.

**Stripe** is interesting because they are a platform that serves other businesses (like Awo ERP), not end users directly. They use isolated data stores per customer type (cards, charges, webhooks all live in different storage systems) and rely heavily on strong application-layer access control rather than database-layer isolation. This works for Stripe because they have enormous engineering resources to keep application logic correct. For Awo ERP with a smaller team, database-enforced RLS is a safer bet.

**ERPNext (Frappe)** uses isolated databases per tenant by default. Each Frappe "site" is a fully separate MySQL/MariaDB database. This gives strong isolation but means each new tenant requires database provisioning, and running hundreds of sites requires significant infrastructure management. ERPNext targets deployments where each customer typically runs their own server, rather than a SaaS model. Awo ERP is firmly a SaaS system, which is why isolated databases are not the right fit.

---

## 5. Tenant Identity — The `X-Tenant-ID` Header

### 5.1 Why a Header and Not Something Else

When a request arrives at the Awo ERP API, the system needs to answer one question before anything else: *which tenant is this request for?* There are several ways to carry this information, and each has meaningful trade-offs.

**Option A — URL Path Prefix (e.g., `/api/v1/tenants/{id}/invoices`)**

Many APIs embed the tenant identifier in the URL path. This is simple and explicit. The problem is that it leaks the tenant identifier into server access logs, browser history, and bookmarks in a way that could expose tenant IDs to unintended parties. It also means every API client must be aware of and correctly include the tenant ID in every URL, which is error-prone.

**Option B — JWT Claim**

A JSON Web Token can carry a `tenant_id` claim inside its cryptographically signed payload. This is a common and generally good approach. However, Awo ERP uses session-based authentication rather than JWTs, and for good reasons (see Section 19). Sessions store state server-side; JWTs embed it client-side. For tenant identity — which can change (tenants can be suspended, merged, renamed) — server-side state is more controllable. A JWT claim for a suspended tenant would remain "valid" until the token expires. A session can be invalidated immediately.

**Option C — Cookie**

A cookie could carry the tenant ID. Cookies are automatically sent by browsers on every request, which is convenient. However, cookies are origin-scoped (they follow domain rules), and getting cookie scoping right across multiple subdomains (`shell-maanzoni.awo.app`, `ocean-grill.awo.app`) is fiddly and error-prone, especially when considering cross-subdomain sharing rules.

**Option D — Subdomain Only**

The subdomain itself (`shell-maanzoni`) contains the tenant slug. Could we just use that? Yes, and Awo ERP does extract the tenant from the subdomain — but the subdomain alone is not sufficient as an *authoritative* identifier because:

1. DNS is not instantaneous. During the brief period between a subdomain being set up in DNS and a tenant going live, routing could be ambiguous.
2. Subdomains are strings (slugs) not stable identifiers. A tenant might be renamed or have their slug changed for legitimate business reasons. The underlying `tenant_id` (a UUID) is the stable identifier.
3. When API clients (mobile apps, integrations, internal tools) call the API programmatically, they may not be calling through subdomains at all. They need a direct, explicit way to declare the tenant they are calling for.

**The Decision — Dual-Signal Model: Subdomain + `X-Tenant-ID` Header**

Awo ERP uses both together. The subdomain provides the *human-readable routing signal* (and allows the browser to do the right thing automatically). The `X-Tenant-ID` header carries the *stable UUID* that is the authoritative database key. The middleware validates that both signals agree before proceeding. This gives us defense-in-depth at the routing layer — an attacker who spoofs one signal still fails at the other.

### 5.2 Header Contract and Format

The `X-Tenant-ID` header must be present on all authenticated API requests except login and public endpoints.

```
X-Tenant-ID: <uuid-v4>

Example:
X-Tenant-ID: 7f3a1c2d-4e5f-6789-abcd-ef0123456789
```

**Rules:**

- Must be a valid UUID v4. Any other format is rejected with `400 Bad Request`.
- Must correspond to an existing, active tenant. Unknown IDs return `404 Not Found`. Suspended tenants return `403 Forbidden`.
- The platform API endpoints (`platform.awo.app`) do not require `X-Tenant-ID` for platform-scoped operations, but do require it when a platform user is operating on behalf of a specific tenant.
- The header name is case-insensitive in HTTP/1.1 and HTTP/2. Awo ERP normalises it to lowercase internally.

### 5.3 Where the Header Lives in the Request Lifecycle

The `X-Tenant-ID` header is set by the API client (browser JavaScript, mobile app, or integration) and travels with every request. Inside Fiber's middleware chain, the `TenantResolver` middleware (second in the chain, after subdomain extraction) reads this header and uses it to load the full tenant context from Redis (or PostgreSQL on cache miss). That tenant context is then attached to Fiber's request context (`c.Locals("tenant", tenantCtx)`) and is available to every downstream handler without further database lookups.

This means the database never receives a query without the tenant context already being established. By the time any handler calls `db.Query(...)`, the PostgreSQL session variable `app.current_tenant_id` has already been set via `SET LOCAL`, which activates the RLS policies.

### 5.4 Subdomain + Header Working Together — The Dual-Signal Model

Here is the exact validation logic that runs on every request:

```
1. Extract subdomain slug from the Host header
   "shell-maanzoni.awo.app" → slug = "shell-maanzoni"

2. Look up tenant by slug in Redis cache
   cache key: "tenant:slug:shell-maanzoni"
   → {tenant_id: "7f3a...", status: "active"}

3. Read X-Tenant-ID header from request
   → "7f3a1c2d-4e5f-6789-abcd-ef0123456789"

4. COMPARE: Does the ID from the header match the ID from the slug lookup?
   IF NO → Reject with 403 Forbidden ("Tenant identity mismatch")
   IF YES → Proceed

5. Check tenant status
   IF suspended → 403 Forbidden ("Tenant account suspended")
   IF terminated → 404 Not Found
   IF active → proceed to session validation
```

The comparison in step 4 is the critical security gate. A request that arrives on the `shell-maanzoni` subdomain but carries the `X-Tenant-ID` of a *different* tenant is rejected. This prevents a class of attack where a malicious actor attempts to use a tenant's subdomain to probe another tenant's data by swapping the header.

### 5.5 Edge Cases — What Happens When They Disagree

**Scenario: Header present, subdomain is `api.awo.app` (the platform API)**

The `api.awo.app` subdomain is a reserved platform subdomain. Requests here are handled by the platform API middleware chain, not the tenant middleware chain. `X-Tenant-ID` is optional and only required when a platform user is explicitly scoping an operation to a specific tenant.

**Scenario: Header missing**

If `X-Tenant-ID` is absent on a tenant-scoped route, the middleware returns `400 Bad Request` with body `{"error": "X-Tenant-ID header is required"}`. It does not try to infer the tenant from the subdomain alone, because subdomain-only identification is not considered authoritative.

**Scenario: Subdomain does not exist in the system**

If a request arrives for `unknown-company.awo.app` and no tenant has that slug, the DNS wildcard will still route it to the API server. The subdomain extraction middleware will find no tenant for that slug and return `404 Not Found` before any further processing. The response intentionally does not reveal whether the tenant exists or not — it simply says "not found."

**Scenario: Valid header, no subdomain (direct IP or platform URL)**

For programmatic API access (mobile apps, integrations calling `api.awo.app` directly), the subdomain part of the dual-signal check is relaxed. The request must arrive on a recognised platform endpoint, and `X-Tenant-ID` alone is used for tenant scoping. The session must still be validated and must carry tenant membership in its server-side data.

---

# PART III — SUBDOMAIN ROUTING

---

## 6. Subdomain Architecture

### 6.1 The Anatomy of an Awo ERP URL

Every tenant in Awo ERP gets a URL of the form:

```
https://<tenant-slug>.awo.app
```

For Shell Maanzoni Service Station, operated by Anika Global Limited:

```
https://shell-maanzoni.awo.app
```

Breaking this down:

- `https` — All traffic is encrypted. No unencrypted HTTP is permitted (see Section 22).
- `shell-maanzoni` — The tenant slug. This is chosen during onboarding and must be URL-safe, lowercase, and unique across the platform. It is derived from the tenant's name but is not guaranteed to be identical to it.
- `awo.app` — The platform base domain.

The platform's own interfaces use reserved subdomains:

```
https://platform.awo.app   — Platform admin and support UI
https://api.awo.app        — Direct API access for integrations
https://app.awo.app        — Marketing/public site (not part of the API)
```

### 6.2 DNS Architecture and Wildcard Records

To route all `*.awo.app` requests to the Awo ERP API servers without manually creating a DNS record for every new tenant, a DNS wildcard record is used.

```
DNS Configuration:
*.awo.app.    300   IN   A     <load-balancer-ip>
awo.app.      300   IN   A     <load-balancer-ip>

Or for IPv6:
*.awo.app.    300   IN   AAAA  <load-balancer-ipv6>
```

The TTL (Time-To-Live) of 300 seconds (5 minutes) is a deliberate choice. A low TTL means that if the IP address of the load balancer ever changes (disaster recovery, IP rotation), DNS propagation is fast. A high TTL would mean clients cache the old IP for hours.

**What the wildcard record does NOT do:** DNS is just routing. The wildcard sends all `*.awo.app` requests to the same IP address — the load balancer. The load balancer then forwards to the API server. It is the API server's job, via the middleware chain, to figure out *which tenant* the request is for. DNS has no concept of tenants.

**On-Premise or Custom Domain Tenants:** In the future, Awo ERP may support tenants bringing their own domain (e.g., `erp.anikagroup.co.ke`). In that scenario, the tenant's domain points at the Awo ERP load balancer via a CNAME, and the system maps the custom domain to the correct `tenant_id` in the database. The `X-Tenant-ID` header becomes even more important in this scenario because subdomain extraction cannot be done via the standard slug approach. This is a planned feature, not currently implemented.

### 6.3 TLS/SSL — Wildcard Certificates and Per-Tenant Certs

A wildcard TLS certificate `*.awo.app` covers all tenant subdomains with a single certificate. This is the simplest approach and is appropriate for Awo ERP's current scale. The certificate is managed via Let's Encrypt or a commercial CA and is renewed automatically.

**Limitation of wildcard certificates:** A wildcard certificate covers one level of subdomain. `*.awo.app` covers `shell-maanzoni.awo.app` but NOT `sub.shell-maanzoni.awo.app`. Awo ERP does not use nested subdomains, so this is not a current concern.

**Custom domain TLS (future):** When custom domains are supported, each custom domain will require its own TLS certificate. This is typically handled via ACME protocol automation (Let's Encrypt) with DNS-01 or HTTP-01 challenges, provisioned as part of the custom domain onboarding flow.

### 6.4 How Fiber Resolves the Subdomain to a Tenant

**[Technical]**

When a request arrives at the Fiber application, the `Host` header contains the full hostname the client requested (e.g., `shell-maanzoni.awo.app`). The `SubdomainExtractor` middleware parses this header to extract the slug.

```go
// middleware/subdomain.go

func SubdomainExtractor(baseDomain string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        host := c.Hostname() // e.g., "shell-maanzoni.awo.app"

        // Strip port if present
        if idx := strings.LastIndex(host, ":"); idx != -1 {
            host = host[:idx]
        }

        // Check if host ends with base domain
        suffix := "." + baseDomain // ".awo.app"
        if !strings.HasSuffix(host, suffix) {
            // Could be a custom domain or platform domain
            // Hand off to custom domain resolver
            c.Locals("subdomain_slug", "")
            return c.Next()
        }

        slug := strings.TrimSuffix(host, suffix)

        // Reject reserved slugs at the routing layer
        if isReservedSlug(slug) {
            // Route to platform handler, not tenant handler
            c.Locals("is_platform_request", true)
            return c.Next()
        }

        c.Locals("subdomain_slug", slug)
        return c.Next()
    }
}

func isReservedSlug(slug string) bool {
    reserved := map[string]bool{
        "api": true, "platform": true, "app": true,
        "admin": true, "mail": true, "status": true,
        "docs": true, "cdn": true,
    }
    return reserved[slug]
}
```

The `TenantResolver` middleware then takes the slug and resolves it to a full tenant context:

```go
// middleware/tenant_resolver.go

func TenantResolver(cache *redis.Client, db *pgxpool.Pool) fiber.Handler {
    return func(c *fiber.Ctx) error {
        slug, ok := c.Locals("subdomain_slug").(string)
        if !ok || slug == "" {
            // No slug = platform request or custom domain
            return c.Next()
        }

        // Try Redis first
        cacheKey := fmt.Sprintf("tenant:slug:%s", slug)
        cached, err := cache.Get(c.Context(), cacheKey).Result()
        if err == nil {
            var tenant TenantContext
            if err := json.Unmarshal([]byte(cached), &tenant); err == nil {
                return validateAndAttach(c, tenant)
            }
        }

        // Cache miss: load from PostgreSQL
        tenant, err := loadTenantBySlug(c.Context(), db, slug)
        if err != nil {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "not_found",
            })
        }

        // Cache the result (15 minutes TTL)
        data, _ := json.Marshal(tenant)
        cache.Set(c.Context(), cacheKey, data, 15*time.Minute)

        return validateAndAttach(c, tenant)
    }
}

func validateAndAttach(c *fiber.Ctx, tenant TenantContext) error {
    // Dual-signal check: header must match slug resolution
    headerID := c.Get("X-Tenant-ID")
    if headerID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "X-Tenant-ID header is required",
        })
    }

    if headerID != tenant.ID.String() {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "tenant_identity_mismatch",
        })
    }

    switch tenant.Status {
    case TenantStatusSuspended:
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "tenant_suspended",
            "message": "This account has been suspended. Please contact support.",
        })
    case TenantStatusTerminated:
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "not_found",
        })
    }

    c.Locals("tenant", tenant)
    return c.Next()
}
```

### 6.5 Reserved Subdomains

The following subdomains are permanently reserved and cannot be assigned to any tenant. Attempting to create a tenant with one of these slugs is rejected during onboarding validation.

| Subdomain | Purpose |
|---|---|
| `api` | Direct API access for programmatic clients |
| `platform` | Platform admin and support interface |
| `app` | Public marketing website |
| `admin` | Reserved for future administrative tools |
| `status` | System status page |
| `mail` | Email infrastructure |
| `cdn` | Static asset delivery |
| `docs` | Documentation site |
| `www` | Alias for root domain |
| `auth` | Authentication endpoints |
| `health` | Health check endpoint (load balancer probes) |

### 6.6 Subdomain Validation Rules

Tenant slugs (and therefore subdomains) must satisfy all of the following:

- Contain only lowercase letters (`a-z`), digits (`0-9`), and hyphens (`-`)
- Begin with a letter or digit (not a hyphen)
- End with a letter or digit (not a hyphen)
- Be between 3 and 63 characters in length (DNS label limit)
- Not be a reserved word (see above list)
- Not contain consecutive hyphens (`--`) — this could be confused with IDN labels
- Be unique across all active and recently-terminated tenants (terminated tenants hold their slug for 90 days to prevent confusion)

---

## 7. The Routing Decision Tree

### 7.1 Platform Routes vs. Tenant Routes

Every request that arrives at the Awo ERP API is one of two fundamental types:

**Platform requests** — Made to `platform.awo.app` or `api.awo.app`. These serve platform users managing the Awo ERP service itself. They require platform user sessions, not tenant user sessions. They may or may not be scoped to a specific tenant.

**Tenant requests** — Made to `<slug>.awo.app`. These serve the employees and managers of a specific tenant business. They require a tenant user session, and the tenant must be active.

These two request types go through different middleware chains and hit different route groups within Fiber. They are not interchangeable — a tenant session cannot access platform routes, and a platform session cannot access tenant data routes without explicit impersonation (see Section 12.2).

### 7.2 Fiber Middleware Chain — Illustrated Step by Step

```
Incoming Request
      │
      ▼
┌─────────────────────────────────────────────┐
│  1. GLOBAL MIDDLEWARE                        │
│     - Request ID generation                 │
│     - Structured logging (attach req ID)    │
│     - Panic recovery                        │
│     - CORS (origin validated against        │
│       *.awo.app pattern)                    │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  2. SUBDOMAIN EXTRACTOR                     │
│     - Parses Host header                    │
│     - Sets: subdomain_slug, is_platform     │
│                                             │
│     → Reserved subdomain? → Platform chain │
│     → Unknown format?     → 400            │
│     → Normal slug?        → Continue       │
└──────────────────┬──────────────────────────┘
                   │
           ┌───────┴────────┐
           ▼                ▼
   PLATFORM CHAIN    TENANT CHAIN
   (Section 12)      (continues below)
                           │
                           ▼
        ┌─────────────────────────────────────┐
        │  3. TENANT RESOLVER                 │
        │     - Reads subdomain_slug          │
        │     - Checks Redis cache            │
        │     - Falls back to PostgreSQL      │
        │     - Validates X-Tenant-ID match   │
        │     - Checks tenant status          │
        │                                     │
        │     → Not found?  → 404            │
        │     → Suspended?  → 403            │
        │     → Mismatch?   → 403            │
        │     → Active?     → Continue       │
        └─────────────────┬───────────────────┘
                          │
                          ▼
        ┌─────────────────────────────────────┐
        │  4. SESSION MIDDLEWARE              │
        │     - Reads session_id cookie       │
        │     - Looks up Redis session        │
        │     - Validates session.tenant_id   │
        │       matches resolved tenant       │
        │     - Attaches user to context      │
        │                                     │
        │     → No cookie?     → 401         │
        │     → Unknown ID?    → 401         │
        │     → Wrong tenant?  → 403         │
        │     → Expired?       → 401         │
        │     → Valid?         → Continue    │
        └─────────────────┬───────────────────┘
                          │
                          ▼
        ┌─────────────────────────────────────┐
        │  5. RATE LIMITER                    │
        │     - Per-tenant token bucket       │
        │     - Per-user rate bucket          │
        │     - Redis-backed counters         │
        │                                     │
        │     → Over limit? → 429            │
        │     → Under limit? → Continue      │
        └─────────────────┬───────────────────┘
                          │
                          ▼
        ┌─────────────────────────────────────┐
        │  6. PERMISSION CHECK                │
        │     - User role vs. route policy   │
        │     - Read from session context    │
        │                                     │
        │     → Unauthorised? → 403          │
        │     → Authorised?   → Continue     │
        └─────────────────┬───────────────────┘
                          │
                          ▼
        ┌─────────────────────────────────────┐
        │  7. ROUTE HANDLER                   │
        │     - Business logic                │
        │     - Sets PostgreSQL session var   │
        │       via SET LOCAL                 │
        │     - Returns data + amis schema    │
        └─────────────────────────────────────┘
```

### 7.3 Short-Circuit Conditions

A "short-circuit" is when the middleware chain stops processing and returns an error response immediately, without reaching the route handler. Each short-circuit condition has a specific HTTP status code and a machine-readable error code in the response body.

| Condition | HTTP Status | Error Code | Logged? |
|---|---|---|---|
| Subdomain not found | 404 | `tenant_not_found` | Yes (warn) |
| X-Tenant-ID missing | 400 | `header_required` | Yes (warn) |
| Header/slug mismatch | 403 | `tenant_identity_mismatch` | Yes (security alert) |
| Tenant suspended | 403 | `tenant_suspended` | Yes (info) |
| Tenant terminated | 404 | `not_found` | Yes (info) |
| Session cookie missing | 401 | `authentication_required` | No (noisy) |
| Session not found in Redis | 401 | `session_invalid` | Yes (info) |
| Session tenant mismatch | 403 | `session_tenant_mismatch` | Yes (security alert) |
| Session expired | 401 | `session_expired` | No (expected) |
| Rate limit exceeded | 429 | `rate_limit_exceeded` | Yes (info) |
| Insufficient permissions | 403 | `forbidden` | Yes (info) |

Security-alert level events (`tenant_identity_mismatch`, `session_tenant_mismatch`) are not just logged — they are emitted as high-priority events to the observability pipeline and should trigger alerts. These are not normal errors. They indicate either a misconfigured client or an active probing attack.

---

# PART IV — TENANT ONBOARDING

---

## 8. Tenant Lifecycle

### 8.1 States

A tenant in Awo ERP is never simply "exists" or "doesn't exist." It moves through a well-defined set of states, and the behaviour of the system changes depending on which state it is in.

```
                     ┌─────────┐
                     │ PENDING │  ← Onboarding workflow running
                     └────┬────┘
                          │ Onboarding complete
                          ▼
                     ┌─────────┐
                ┌───▶│ ACTIVE  │◀───┐  ← Normal operations
                │    └────┬────┘    │
                │         │         │
           Reinstated  Suspended  Reinstated
                │         │         │
                │         ▼         │
                └─── ┌──────────┐───┘
                     │SUSPENDED │  ← API returns 403
                     └────┬─────┘
                          │ Never reinstated /
                          │ explicit termination
                          ▼
                    ┌───────────┐
                    │TERMINATED │  ← Data retained per policy,
                    └───────────┘    API returns 404
```

| State | API Behaviour | Description |
|---|---|---|
| `PENDING` | 503 or redirect to setup | Onboarding workflow is in progress. The tenant's subdomain may already be live in DNS. |
| `ACTIVE` | Normal | All API endpoints available to authenticated users. |
| `SUSPENDED` | 403 Forbidden | Non-payment, policy violation, or manual action. Read-only data export may still be permitted. |
| `TERMINATED` | 404 Not Found | Contractually terminated. Data is retained per data retention policy but is inaccessible via API. |

### 8.2 State Transitions and Who Can Trigger Them

| Transition | From | To | Who Can Trigger |
|---|---|---|---|
| Onboarding completes | PENDING | ACTIVE | Temporal workflow (automated) |
| Onboarding fails (unrecoverable) | PENDING | TERMINATED | Temporal workflow (automated) |
| Non-payment / policy breach | ACTIVE | SUSPENDED | Billing system (automated) / Platform Super Admin |
| Customer pays outstanding balance | SUSPENDED | ACTIVE | Billing system (automated) / Platform Super Admin |
| Customer requests cancellation | ACTIVE | TERMINATED | Platform Super Admin (manual, with confirmation) |
| Forced termination (abuse) | SUSPENDED | TERMINATED | Platform Super Admin |

No tenant can change their own state. These are platform-level operations.

---

## 9. The Onboarding Flow — End to End

### 9.1 Self-Service Signup vs. Platform-Admin Provisioning

Awo ERP supports two paths to creating a new tenant:

**Self-service signup:** A prospective customer fills in a signup form at `app.awo.app/signup`. They provide their company name, industry vertical, contact information, and desired subdomain slug. This triggers the onboarding workflow automatically.

**Platform-admin provisioning:** A platform admin creates the tenant manually via the platform interface. This is used for enterprise customers with custom contracts, for tenants being migrated from another system, or for demo/sandbox environments.

Both paths result in the exact same Temporal workflow being triggered. The difference is only in who initiates it and what initial data is provided.

### 9.2 Temporal Workflow — Why Onboarding is Not a Simple HTTP Handler

A naive implementation of tenant onboarding might look like this: an HTTP handler receives the signup form, creates a database record, seeds some data, sends an email, and returns 200 OK. This is fine for simple cases. For Awo ERP, it is inadequate for several reasons.

Onboarding involves many steps across multiple systems: database writes, Redis warming, DNS or configuration updates, email sending. Any of these can fail. If step 4 (seeding the Chart of Accounts) fails after step 3 (creating the tenant record) has already succeeded, the system is in an inconsistent state. A simple HTTP handler has no built-in way to resume from step 4. It either has to retry the whole thing (risking duplicate records) or leave a broken tenant in the system.

Temporal is a workflow orchestration platform that solves exactly this problem. It maintains a durable execution history for every workflow. If an activity (a single step) fails, Temporal automatically retries it according to a configurable retry policy. If the entire Temporal worker crashes mid-workflow, when it comes back up, the workflow resumes from exactly where it left off. Every step is recorded. Every retry is logged. This is not a convenience — for a system handling businesses' financial data, it is a correctness requirement.

Comparing alternatives:
- **Simple HTTP handler with a queue:** Could work, but requires building retry logic, idempotency, dead-letter queues, and monitoring manually. Temporal provides all of this.
- **Saga pattern with a message broker (Kafka, RabbitMQ):** More complex setup, same benefits. Temporal implements the saga pattern but with a much gentler learning curve and better tooling.
- **Database transactions only:** Can handle atomicity within the database, but onboarding involves external systems (email, DNS) that are outside a database transaction. Temporal handles these gracefully.

### 9.3 Step-by-Step Onboarding Workflow Activities

The onboarding workflow consists of the following activities, executed in order:

```
OnboardingWorkflow
  ├── Activity 1: ValidateAndReserveSubdomain
  ├── Activity 2: ProvisionTenantRecord
  ├── Activity 3: SeedChartOfAccounts
  ├── Activity 4: SeedAmisSchemas
  ├── Activity 5: CreatePlatformAdminUser
  ├── Activity 6: WarmRedisCache
  ├── Activity 7: SendWelcomeNotification
  └── Activity 8: ActivateTenant
```

#### 9.3.1 Validate & Reserve Subdomain

**What it does:** Checks that the requested slug is available, valid (passes the rules in Section 6.6), and not reserved. If available, it marks the slug as "reserved" in the database to prevent a race condition where two signups for the same slug happen simultaneously.

**Why it is separate from provisioning:** If validation fails, no other steps should run. Reserving the slug as a distinct step (before creating the full tenant record) is the cleanest way to handle concurrent signups.

**Retry policy:** Retries up to 3 times with 5-second backoff. If slug is taken on retry (unlikely but possible in a race), returns a specific error code so the workflow can prompt the user to choose a different slug.

#### 9.3.2 Provision Tenant Record in PostgreSQL

**What it does:** Creates the `tenant` record with status `PENDING`, stores the industry vertical, contact details, and the reserved slug. Also creates the first platform admin user for this tenant (the signup submitter).

**Idempotency key:** The tenant's UUID is generated before this activity runs and is passed in as a workflow input. If this activity is retried, the INSERT uses `ON CONFLICT DO NOTHING` with the UUID as the primary key, ensuring no duplicate records.

#### 9.3.3 Seed Chart of Accounts

**What it does:** Inserts the standard Chart of Accounts appropriate for the tenant's industry vertical and jurisdiction. A petroleum retailer in Kenya gets the Kenyan petroleum COA. A restaurant gets the hospitality COA. These are maintained as seeded JSON files per vertical/jurisdiction combination.

**Why Temporal handles this:** The COA seed can be hundreds of records. If the database is momentarily overloaded and this activity fails halfway, Temporal retries the entire activity. The INSERT operations are wrapped in a transaction and use `ON CONFLICT DO NOTHING` to ensure retries are safe.

#### 9.3.4 Seed Default amis-UI Schema Definitions per Industry Vertical

**What it does:** Inserts the default page schemas, form schemas, and list view schemas for the tenant's industry vertical into the `tenant_ui_schemas` table. These are the starting point for what the tenant's users will see when they log in.

A petroleum tenant gets schemas for: shift management, pump readings, dip readings, cash events, fuel deliveries, reconciliation, and the standard finance module. A restaurant tenant gets schemas for: table management, orders, recipes, inventory, and finance. The schemas are stored as JSONB in PostgreSQL and can be customised per-tenant after onboarding.

**Why this is a separate activity from COA seeding:** Schema seeding and COA seeding are independent concerns. If schema seeding fails (perhaps a schema file has a syntax error), the COA is already committed and the tenant can proceed with data entry even while the UI is being fixed. Separating activities gives more granular retry control.

#### 9.3.5 Warm Redis Cache

**What it does:** Pre-populates Redis with the newly created tenant's data that will be needed on first login: the tenant context object (for the TenantResolver middleware), the user's session-related data, and the most commonly accessed amis schemas.

**Why this matters:** Without warming, the first few requests after onboarding will all be cache misses, hitting PostgreSQL simultaneously. For a tenant with a large team logging in simultaneously on day one, this could cause a thundering-herd problem. Warming the cache during onboarding prevents this.

#### 9.3.6 Send Welcome Notification

**What it does:** Sends a welcome email to the primary contact, containing their subdomain URL, a temporary login link, and getting-started documentation links.

**Why Temporal handles this:** Email delivery is an external service. If the email provider is temporarily unavailable, Temporal will retry this activity with exponential backoff for up to 24 hours before marking it as failed. A failed email does not block the tenant from going live — it is the last step before activation, but activation does not depend on it succeeding.

#### 9.3.7 Activate DNS / Mark Tenant Live

**What it does:** Updates the tenant's status from `PENDING` to `ACTIVE` in PostgreSQL and invalidates the (nonexistent) Redis cache entry for this slug so that the next request fetches the fresh `ACTIVE` state.

Because Awo ERP uses a DNS wildcard (`*.awo.app`), there is no actual DNS record to create per tenant. The subdomain is already routable the moment it is reserved. "Activating DNS" in this context means marking the tenant as active in the system so that the middleware chain will accept requests for it.

For custom domains (future), this activity would also trigger the ACME certificate provisioning.

### 9.4 Idempotency and Retry Safety in Temporal Activities

Every activity in the onboarding workflow is designed to be idempotent — running it twice produces the same result as running it once. This is not optional; it is a requirement because Temporal's at-least-once delivery guarantee means an activity *can* run more than once under failure conditions.

The patterns used to achieve idempotency:

- **Database inserts:** `INSERT ... ON CONFLICT DO NOTHING` with deterministic primary keys (the tenant UUID is decided before the workflow starts)
- **Redis writes:** `SET key value EX ttl` is inherently idempotent; overwriting with the same value is harmless
- **Email sending:** The outbox pattern is used. The email is written to the database first; a separate process sends it and marks it sent. On retry, the "marked sent" check prevents re-sending.

### 9.5 What Happens When Onboarding Fails Mid-Way

Temporal maintains complete workflow history. If an activity fails beyond its retry limit (a "non-retryable error"), the workflow transitions to a `FAILED` state. At this point:

1. A platform alert is generated with the workflow ID and the failing activity name.
2. The tenant remains in `PENDING` state. Its subdomain is reserved.
3. A platform admin can inspect the workflow history in the Temporal UI to understand exactly what failed and why.
4. The admin can either fix the underlying problem and replay the failed activity, or terminate the workflow and release the reserved slug.
5. The customer receives an automated notification that their account setup encountered an issue and support has been notified.

The Temporal workflow ID is the tenant's UUID prefixed with `onboarding-`. This makes it trivially easy to find a specific tenant's workflow in the Temporal UI.

---

# PART V — PLATFORM USERS

---

## 10. Platform Users vs. Tenant Users

### 10.1 Conceptual Separation (Non-Technical)

Think of the difference between a *tenant user* and a *platform user* the way you would think about the difference between a tenant living in an apartment and the building management company's staff.

The tenant's employees — cashiers, managers, accountants — are tenant users. They log into `shell-maanzoni.awo.app` and work within Shell Maanzoni's account. They see only Shell Maanzoni's data. They cannot see or affect any other tenant.

Awo ERP's own staff — developers, support agents, billing team — are platform users. They log into `platform.awo.app`. They can see cross-tenant dashboards (how many active tenants, system-wide error rates, billing status). They can access any specific tenant's account for legitimate support purposes, but doing so creates an audit log entry.

The two groups authenticate through separate login endpoints, hold sessions with different shapes, and have access to completely different route groups.

### 10.2 Data Model Differences

**Tenant Users** are stored in the `users` table, which has `tenant_id` as a non-nullable column, and is protected by RLS. A tenant user record is always and only visible to the tenant they belong to.

**Platform Users** are stored in a separate `platform_users` table that is *not* protected by RLS (or more precisely, has a policy that allows access only to the designated platform admin role). This table has no `tenant_id` column. It has a `platform_role` column instead.

This separation is deliberate. If both types of users were in the same table, a bug in RLS policy configuration could theoretically allow a tenant user to see platform user records, or vice versa. By putting them in completely separate tables with separate access patterns, we eliminate that risk.

```sql
-- Tenant user (RLS protected, scoped to one tenant)
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    email       TEXT NOT NULL,
    role        TEXT NOT NULL,  -- e.g., 'cashier', 'manager', 'accountant'
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Platform user (NOT RLS protected, no tenant_id)
CREATE TABLE platform_users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL UNIQUE,
    platform_role   TEXT NOT NULL,  -- 'super_admin', 'support_agent', etc.
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

### 10.3 How Platform Users Authenticate

Platform users authenticate via a login form at `platform.awo.app/login`. The authentication endpoint is on the platform API, not the tenant API. It validates credentials against the `platform_users` table and creates a platform session in Redis under the key namespace `platform:session:<session_id>`.

The platform session payload has a different shape from a tenant session:

```json
// Tenant session (in Redis at "session:<session_id>")
{
  "user_id": "<uuid>",
  "tenant_id": "<uuid>",
  "tenant_slug": "shell-maanzoni",
  "role": "manager",
  "created_at": "2025-01-15T08:00:00Z",
  "expires_at": "2025-01-15T16:00:00Z"
}

// Platform session (in Redis at "platform:session:<session_id>")
{
  "platform_user_id": "<uuid>",
  "platform_role": "support_agent",
  "created_at": "2025-01-15T08:00:00Z",
  "expires_at": "2025-01-15T17:00:00Z",
  "impersonating_tenant_id": null
}
```

The `platform:session:` prefix ensures platform sessions are never confused with tenant sessions, even if (by extraordinary coincidence) two UUIDs were identical.

---

## 11. Platform User Roles

### 11.1 Super Admin

Has full access to all platform operations. Can create, suspend, reinstate, and terminate tenants. Can access any tenant's data for any reason. Can create and manage other platform users. Can trigger manual Temporal workflow actions. Should be limited to 2–3 people.

### 11.2 Support Agent

Can view any tenant's account in read-only mode for troubleshooting purposes. Can trigger password resets for tenant users. Can view tenant session history and error logs. Cannot modify tenant data. Cannot suspend or terminate tenants.

### 11.3 Billing Admin

Can view billing and subscription information for all tenants. Can manually apply credits, update subscription tiers, and trigger suspension/reinstatement for payment reasons. Cannot access tenant operational data (invoices, financial records, etc.).

### 11.4 Read-Only Auditor

Has read access to platform audit logs only. Cannot access tenant data. Used for compliance and external audit purposes.

---

## 12. Platform User API Surface

### 12.1 Platform-Scoped Endpoints

All platform endpoints are served from `platform.awo.app/api/v1/`. They require a platform session cookie. They do not require `X-Tenant-ID` unless explicitly scoping to a tenant.

Key endpoint groups:

```
GET  /api/v1/tenants                    — List all tenants (paginated)
GET  /api/v1/tenants/:id               — Get tenant details
POST /api/v1/tenants                    — Create tenant (admin provisioning)
PUT  /api/v1/tenants/:id/suspend       — Suspend tenant
PUT  /api/v1/tenants/:id/reinstate     — Reinstate tenant
GET  /api/v1/tenants/:id/users         — List tenant's users
GET  /api/v1/platform-users            — List platform users
POST /api/v1/platform-users            — Create platform user
GET  /api/v1/audit-log                 — Platform-wide audit log
GET  /api/v1/system/health             — System health dashboard
POST /api/v1/tenants/:id/impersonate   — Begin impersonation session
DELETE /api/v1/impersonate             — End impersonation session
```

### 12.2 Tenant Impersonation — Acting On Behalf of a Tenant

Impersonation allows a support agent to see exactly what a tenant user sees, without needing the tenant user's credentials. It is a legitimate and necessary support tool, but it must be carefully controlled.

**How it works:**

1. The platform user (with Support Agent or Super Admin role) calls `POST /api/v1/tenants/:id/impersonate` with a reason for the impersonation.
2. The platform API creates a special impersonation record in the `impersonation_sessions` table, logs the action to the audit log, and returns a short-lived impersonation token (valid for 1 hour, non-renewable).
3. The platform user opens `<tenant-slug>.awo.app` with this token. A special middleware recognises the token, validates it against the impersonation session record, and grants access as a read-only tenant user.
4. Every action taken during impersonation — every page view, every query — is logged with `impersonated_by: <platform_user_id>` in the audit log.
5. When the impersonation expires or the support agent explicitly ends it (`DELETE /api/v1/impersonate`), the impersonation session is closed and no further access is possible.

**What impersonation cannot do:**

- Impersonated sessions are always read-only. Write operations (creating records, modifying data) are blocked at the middleware level regardless of what the original tenant user's role would permit.
- Impersonation sessions cannot be extended or renewed. If more time is needed, a new impersonation session must be created, creating a fresh audit trail entry.
- Impersonation of `TERMINATED` tenants is not possible. Terminated tenants' data is accessible only through a direct database query by a Super Admin, never through the API.

### 12.3 Audit Trail Requirements for Platform Actions

Every action taken by a platform user — whether through the platform API or via impersonation — must be recorded in the `platform_audit_log` table. This is non-negotiable for compliance and for trust.

```sql
CREATE TABLE platform_audit_log (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_user_id    UUID NOT NULL REFERENCES platform_users(id),
    action              TEXT NOT NULL,         -- e.g., 'tenant.suspend'
    target_tenant_id    UUID REFERENCES tenants(id),
    target_user_id      UUID,
    reason              TEXT,                  -- Required for impersonation, suspension
    request_ip          INET NOT NULL,
    user_agent          TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);
```

This table is append-only. No `UPDATE` or `DELETE` is permitted on it. Periodic archival to cold storage (S3 / object store) ensures it remains manageable in size while being permanently queryable.

---

# PART VI — THE amis-UI LAYER

---

## 13. What is amis and Why Awo ERP Chose It

### 13.1 The Backend-Driven UI Concept Explained Simply

In a traditional web application, the server's job is to answer questions: "Give me the list of invoices." "What is the balance of this account?" The *presentation* of those answers — how they look on screen, what columns the table has, which buttons appear — is decided entirely by frontend code sitting in the browser.

Backend-driven UI flips this. The server does not just send data; it sends instructions for how to display the data. The browser has a general-purpose *renderer* that can interpret these instructions and build the interface on the fly.

Imagine it like a restaurant. In the traditional model, the kitchen just sends out plates of food, and the waiters decide how to serve it based on training they received months ago. In backend-driven UI, the kitchen sends out the food *and* a card that says "serve this in a bowl, with these garnishes, at this table, to the person on the left." The waiter just follows the card.

In Awo ERP, the "card" is a JSON document describing a complete user interface — forms with fields, tables with columns, buttons with actions, validation rules, and even which elements should be hidden based on the user's role. This JSON is the amis schema.

### 13.2 amis vs. Traditional Frontend Frameworks

**React / Vue / Angular** are exceptional tools for building rich, interactive applications. They are the right choice when the interface is fundamentally dynamic, highly custom, or requires very sophisticated client-side state management. They are not the right choice when:

- The number of "document types" and "views" is large and varies per customer (ERP systems typically have hundreds of form types)
- New form types need to be added without frontend deployments
- The team building the system is primarily backend-focused
- UI configuration needs to be tenant-customisable without writing code

Building Awo ERP's full ERP UI in React or Vue would require a large, dedicated frontend team, ongoing UI deployments for every new document type, and a complex API between the frontend and backend. For every new module (Fuel Deliveries, Shift Reconciliation, Fleet Cards), a React component would need to be built, reviewed, and deployed. With amis, the same module only requires a JSON schema file.

**The trade-off:** amis schemas are less flexible than hand-crafted React components for truly custom or novel interactions. There are things you can build in React that amis cannot express. For an ERP system — which is fundamentally a collection of forms, lists, dashboards, and reports — this constraint is acceptable. The speed of development and the elimination of a frontend team more than compensate.

### 13.3 amis vs. Other Server-Driven UI Options

| Option | Strengths | Weaknesses | Why Not for Awo ERP |
|---|---|---|---|
| **Retool** | Excellent for internal tools, rich component library | Expensive, proprietary, SaaS — tenant data leaves your infrastructure | Cannot host in your own infrastructure securely for ERP data |
| **Budibase** | Open source, self-hostable | Limited component ecosystem, early-stage | Not mature enough for complex ERP workflows |
| **Appsmith** | Open source, good for internal tools | Component library not as rich as amis, less suited to complex forms | |
| **JSON Forms** | Standard-based (JSON Schema), framework-agnostic | Form-only, no page-level composition | Cannot compose full page layouts |
| **amis** | Comprehensive component library (200+ components), full page composition, data binding, actions, validation, Chinese enterprise pedigree | Documentation primarily in Chinese, smaller English-language community | — This is the chosen option |

amis was chosen primarily because of the breadth of its component library and its ability to compose complete page layouts — not just individual forms. An amis schema can describe an entire page: a header, a sidebar navigation, a main content area with tabs, each tab containing tables or forms with their own data fetching, validation, and action buttons. This is what an ERP module requires.

### 13.4 Where amis Fits in the ERP Landscape

Most traditional ERP systems (SAP, Oracle, Microsoft Dynamics) ship with their UI baked in. Customisation requires expensive development. Some newer systems (ERPNext / Frappe) use a hybrid: the framework generates forms from metadata, but customisation often requires writing Python and JavaScript.

Awo ERP's approach places it closer to the "low-code platform" end of the spectrum in terms of UI, while maintaining a fully custom, domain-modeled backend in Go. The UI is generic and configurable; the business logic is purpose-built.

---

## 14. The amis API Response Contract

### 14.1 Anatomy of an amis JSON Response — Data + UI Schema in One Payload

An amis page in Awo ERP is typically loaded in two requests:

**Request 1 — Schema fetch:** The browser asks for the UI definition. What should this page look like?

**Request 2 — Data fetch:** amis, having rendered the UI skeleton, asks for the actual data to populate it.

However, for simple pages (a form pre-populated with existing data, a small dashboard), both the schema and the initial data can be returned in a single response. Awo ERP uses a consistent envelope for both patterns.

### 14.2 The Envelope Structure Awo ERP Adopts

All amis-related API responses use the following envelope:

```json
// Schema-only response (for complex pages where data loads separately)
{
  "status": 0,
  "msg": "",
  "data": {
    "type": "page",
    "title": "Fuel Deliveries",
    "body": {
      "type": "crud",
      "api": "/api/v1/fuel-deliveries",
      "columns": [
        { "name": "delivery_date", "label": "Date", "type": "date" },
        { "name": "product",       "label": "Product", "type": "text" },
        { "name": "quantity_l",    "label": "Quantity (L)", "type": "number" },
        { "name": "supplier",      "label": "Supplier", "type": "text" }
      ],
      "headerToolbar": ["bulk-actions", "export-csv"],
      "filter": {
        "type": "form",
        "body": [
          { "type": "date-range", "name": "date_range", "label": "Date Range" },
          { "type": "select", "name": "product", "label": "Product",
            "source": "/api/v1/ref/fuel-products" }
        ]
      }
    }
  }
}
```

```json
// Combined schema + data response (for forms with existing data)
{
  "status": 0,
  "msg": "",
  "data": {
    "type": "form",
    "title": "Edit Fuel Delivery",
    "body": [
      { "type": "date",   "name": "delivery_date", "label": "Delivery Date", "required": true },
      { "type": "select", "name": "product", "label": "Product",
        "options": [
          { "label": "PMS (Petrol)", "value": "pms" },
          { "label": "AGO (Diesel)", "value": "ago" },
          { "label": "DPK (Kerosene)", "value": "dpk" }
        ]
      },
      { "type": "number", "name": "quantity_l", "label": "Quantity (Litres)" }
    ],
    "initApi": {
      "method": "get",
      "url": "/api/v1/fuel-deliveries/${id}"
    },
    "api": {
      "method": "put",
      "url": "/api/v1/fuel-deliveries/${id}"
    }
  }
}
```

**The `status` field** follows amis convention: `0` means success. Any non-zero value is treated as an error by the amis renderer. The `msg` field carries human-readable error messages.

**The `data` field** contains the amis schema (for schema endpoints) or the data records (for data endpoints). The structure of `data` is determined by what the amis component expects — a `crud` component expects `{items: [...], total: N}`, a form's `initApi` expects `{fieldName: value, ...}`.

### 14.3 Separating Concerns — When Data and Schema Travel Together vs. Separately

**Schema and data together:** Best for forms with a small, fixed set of fields. The server renders the complete schema including any tenant-specific field customisations, and the data is embedded in the same response. One round-trip, page is ready.

**Schema separate from data:** Best for lists (CRUD pages) where the schema defines the columns and filters, but the data is fetched dynamically as the user pages, sorts, and filters. The schema is cached aggressively (it changes rarely); the data is always fresh.

**The decision in Go:** The Fiber handler for a schema endpoint knows whether it is a schema-only or combined response based on the route group it belongs to:

- `GET /api/v1/schema/:page_name` — Schema-only endpoints, aggressively cached
- `GET /api/v1/:resource` — Data endpoints, returning amis-compatible data envelopes
- `GET /api/v1/:resource/:id/form` — Combined schema + data for edit forms

### 14.4 Versioning the UI Schema

amis schemas stored in the database carry a `schema_version` field. When Awo ERP's platform team makes a breaking change to a schema (removing a field, changing a field type), the version is incremented. The schema endpoint returns the version in the response, and the frontend can use this to know when to invalidate its local cache.

For non-breaking changes (adding a new optional field, adding a new action button), the version does not change. This follows a conservative versioning philosophy: only breaking changes force a cache invalidation.

---

## 15. Tenant-Aware UI Rendering

### 15.1 How the Backend Composes an amis Schema Per Tenant

When a tenant user requests a page schema, the Go handler does not simply return a static file. It goes through a composition process:

```
1. Load base schema for page (from file system or database)
   e.g., "fuel_deliveries_list_v1.json"

2. Load tenant's overrides for this page (from database)
   e.g., tenant has added a custom "Supplier Code" column

3. Load user's role-based visibility rules (from session)
   e.g., cashier cannot see "unit_cost" column

4. Merge: base + overrides + visibility rules

5. Return composed schema
```

This composition is done in Go, using a simple recursive merge strategy: tenant overrides take precedence over the base schema, and visibility rules are applied last as a post-processing step.

### 15.2 Industry Vertical Schemas

During onboarding, each tenant is assigned an industry vertical (`petroleum_retail`, `restaurant`, `retail`, `aviation`). The vertical determines which base schemas are seeded into the `tenant_ui_schemas` table.

A petroleum tenant's seeded schema set includes pages that a restaurant tenant will never see, and vice versa. If a tenant later expands into multiple verticals (a petroleum retailer that also runs a restaurant), additional schema sets can be added to their account by a platform admin.

### 15.3 Per-Tenant Schema Overrides and Customisation

The base schemas are the starting point. After onboarding, a tenant's administrator can customise their UI in limited ways (depending on their subscription tier):

- Add custom fields to forms
- Change column order in list views
- Add custom calculated columns
- Rename labels (e.g., "Supplier" → "Vendor")
- Add custom dashboard widgets

These customisations are stored as JSONB "override patches" in the `tenant_ui_schemas` table, not as complete schema replacements. The Go handler merges the override patch onto the base schema at request time. This means that when Awo ERP upgrades the base schema (adding new features), the upgrade flows through to all tenants automatically, and their overrides are preserved.

### 15.4 Role-Aware UI — Hiding Fields and Actions Based on Permissions (`visibleOn`)

amis supports conditional visibility through `visibleOn` expressions. A column or field can be conditionally shown:

```json
{
  "name": "unit_cost",
  "label": "Unit Cost",
  "type": "number",
  "visibleOn": "${role === 'manager' || role === 'accountant'}"
}
```

The `role` variable is injected into the amis data context by the Go backend when the schema is served. The amis renderer on the browser evaluates this expression client-side and shows or hides the column accordingly.

**Critical caveat:** See Section 17.1 for why `visibleOn` is NOT a security mechanism.

### 15.5 Storing amis Schemas — PostgreSQL vs. Redis vs. File System

**File System:** Simple and transparent. Schemas are version-controlled alongside the Go code. However, tenant-specific overrides cannot be stored in the file system without a per-tenant directory structure, which becomes unwieldy. Base schemas live here.

**PostgreSQL (JSONB):** The right choice for tenant-specific schemas and override patches. JSONB allows efficient storage and querying of JSON data. Schemas can be modified at runtime without deployments. Tenant overrides are stored here.

**Redis:** Schemas are cached in Redis after being composed (base + overrides + visibility). The cache key is `tenant:{tenant_id}:schema:{page_name}:{user_role}`. TTL is 15 minutes for most schemas, 1 minute for schemas that include real-time data in their `initApi` configuration.

This three-layer storage model is the right balance: base schemas are files (version-controlled, stable), customisations are in PostgreSQL (durable, queryable, auditable), and the composed result is in Redis (fast, automatically rebuilt on miss).

---

## 16. The Schema Request Lifecycle

### 16.1 How a Page Load Becomes a Schema Fetch Then a Data Fetch

```
User navigates to shell-maanzoni.awo.app/fuel-deliveries
      │
      ▼
Browser loads the amis shell (static HTML + amis JS, CDN-served)
      │
      ▼
amis shell calls: GET /api/v1/schema/fuel_deliveries_list
  Headers: X-Tenant-ID, Cookie: session_id
      │
      ▼
Go handler:
  1. Resolves tenant from middleware (already done)
  2. Checks Redis: tenant:{id}:schema:fuel_deliveries_list:manager
  3. Cache HIT → return cached schema (< 1ms)
  4. Cache MISS →
       a. Load base schema from file system
       b. Load override patch from PostgreSQL
       c. Apply role visibility rules from session
       d. Write composed schema to Redis (TTL 15min)
       e. Return composed schema
      │
      ▼
amis renders the page skeleton (table columns, filter bar, buttons)
      │
      ▼
amis calls the data API defined in the schema's "api" field:
  GET /api/v1/fuel-deliveries?date_range=...&page=1&per_page=25
  Headers: X-Tenant-ID, Cookie: session_id
      │
      ▼
Go handler:
  1. Tenant & session resolved (middleware)
  2. Sets: SET LOCAL app.current_tenant_id = '<uuid>'
  3. Queries PostgreSQL (RLS filters automatically)
  4. Returns: { "status": 0, "data": { "items": [...], "total": 47 } }
      │
      ▼
amis populates the table with data
      │
      ▼
Page is fully rendered
```

### 16.2 Schema Caching in Redis — Strategy and Invalidation

Schemas are cached because composing them (loading files, querying PostgreSQL for overrides, applying role visibility) takes time — potentially 50–100ms on every request. For a page that many users load frequently, this adds up. Caching brings it to sub-millisecond.

**Cache key structure:**

```
tenant:{tenant_id}:schema:{page_name}:{role}
e.g.: tenant:7f3a...:schema:fuel_deliveries_list:manager
```

The `role` is part of the cache key because different roles see different schemas (visibility rules). A manager schema and a cashier schema for the same page are different objects.

**Invalidation triggers:**

- A platform admin updates the base schema (global update): all `tenant:*:schema:fuel_deliveries_list:*` keys are deleted (pattern delete via Redis SCAN + DEL)
- A tenant admin saves a customisation override: all `tenant:{id}:schema:*` keys for that tenant are deleted
- TTL expiry: all schema keys have a maximum TTL (15 minutes), providing automatic eventual consistency even if explicit invalidation fails

### 16.3 Partial Schema Updates vs. Full Schema Reload

When a tenant admin changes only one field's label, it would be wasteful to invalidate and rebuild the entire page schema. Awo ERP uses a patch-based override system (Section 15.3) to minimise this, but cache invalidation is still done at the full-page-schema level for simplicity. The Redis rebuild cost is small (< 50ms), and the simplicity of "invalidate everything for this tenant" avoids subtle bugs where partial updates leave inconsistent state.

### 16.4 Handling Schema Errors Gracefully

If the schema API returns an error (500, network timeout), amis should not crash the entire application. The amis shell is configured with a global error boundary that shows a friendly message: "This page could not be loaded. Please try again or contact support." The request ID (generated in the global middleware) is displayed in the error UI, making it easy for support agents to correlate the error with the server log.

---

## 17. Security Implications of Backend-Driven UI

### 17.1 `visibleOn` is Not an Access Control Mechanism

This is the most important security concept in the entire amis-ui integration and must be understood by every engineer on the team.

`visibleOn` expressions in amis schemas are evaluated in the user's browser. They can hide fields, columns, and buttons from the UI. However:

**A determined user can bypass `visibleOn` restrictions trivially.** They can open the browser developer tools, inspect the API response, and see all the data that was sent — including data for fields that are "hidden" by `visibleOn`. They can also call the API directly (with `curl` or Postman) and receive all data the server is willing to return.

`visibleOn` is a UX mechanism, not a security mechanism. It prevents ordinary users from being confused by fields they are not supposed to interact with. It does not prevent them from accessing the data if they are motivated.

**The correct security model:**

```
VISIBILITY (amis visibleOn) → Controls what users SEE in the UI
AUTHORISATION (server-side) → Controls what data the API RETURNS
```

For every field that should not be visible to a certain role, the Go handler must *also* exclude that field from the API response payload. If a cashier should not see `unit_cost`, the API endpoint must strip `unit_cost` from the response when the session's role is `cashier`, regardless of what the schema says. The schema's `visibleOn` is the belt; the server-side omission is the suspenders.

### 17.2 Sensitive Fields — Ensuring RLS Filters Data Before It Reaches the Schema

RLS operates at the row level — it filters which *rows* a tenant can see. It does not filter columns. Column-level access control is the responsibility of the Go handler.

For sensitive fields (e.g., `cost_price`, `margin`, `employee_salary`), the handler checks the user's role and either includes or excludes the field from the response struct before serialising to JSON. Go's `json:"-"` tag alone is not sufficient here because the same struct may be used for both privileged and unprivileged responses. Explicit role checks and response DTO construction are the correct pattern.

### 17.3 Schema Injection — Can a Tenant Manipulate Another Tenant's UI?

If Tenant A were somehow able to modify Tenant B's stored schema override, they could theoretically serve malicious UI to Tenant B's users. This is prevented by:

1. The schema override API endpoint requires a tenant admin session scoped to the tenant being modified. RLS ensures the `tenant_ui_schemas` table only shows and allows writes to the authenticated tenant's records.
2. Schema JSON is validated server-side before storage. Only known amis component types and properties are accepted. An override patch that attempts to inject arbitrary JavaScript into an amis component is rejected.
3. amis itself has a content security policy that prevents execution of arbitrary JavaScript injected through schemas.

### 17.4 API Actions Embedded in amis Schemas — Preventing CSRF and Privilege Escalation

amis schemas can contain `api` definitions — URLs that amis will call when a user clicks a button or submits a form. These are constructed by the Go backend and should only point to legitimate Awo ERP API endpoints.

**CSRF protection:** All state-changing API endpoints (POST, PUT, DELETE) require the session cookie to be present AND the `X-Tenant-ID` header to match. Cross-site request forgery attacks would need to forge both. Additionally, the `SameSite=Strict` cookie attribute prevents the session cookie from being sent in cross-origin requests.

**Preventing schema-embedded escalation:** The backend validates that any `api.url` field in a stored schema override points to a URL within the Awo ERP API. External URLs (pointing to third-party services) are rejected during schema validation, preventing a tenant admin from embedding an amis schema that exfiltrates data to an external server.

---

# PART VII — SECURITY

---

## 18. Security Architecture Overview

### 18.1 Threat Model — Who Are the Adversaries?

Understanding security requires understanding who might attack the system and what they want.

**Adversary 1 — A malicious tenant user** trying to access another tenant's data. They have a legitimate account on `tenant-a.awo.app` and attempt to access data belonging to `tenant-b.awo.app`. Motivation: competitive intelligence, sabotage.

**Adversary 2 — A former employee** of a tenant whose account should have been deactivated but was not immediately. Motivation: revenge, data theft.

**Adversary 3 — An external attacker** with no legitimate account, attempting to find vulnerabilities in the API (SQL injection, auth bypass, IDOR). Motivation: data theft, ransomware.

**Adversary 4 — A malicious or compromised platform user** attempting to access tenant data beyond their authorised scope, or to manipulate platform operations. Motivation: insider threat.

**Adversary 5 — A compromised third-party integration** (a mobile app or external system with API access) whose credentials have been stolen. Motivation: exfiltrate data.

### 18.2 Defence-in-Depth Layers

Awo ERP's security is not a single gate — it is a series of overlapping layers. An attacker who bypasses one layer encounters the next.

```
Layer 1: NETWORK        — TLS encryption, HSTS, no HTTP
Layer 2: GATEWAY        — Rate limiting, DDoS protection, header stripping
Layer 3: ROUTING        — Subdomain validation, X-Tenant-ID cross-check
Layer 4: AUTHENTICATION — Session validation, expiry, rotation
Layer 5: AUTHORISATION  — Role-based access control, permission checks
Layer 6: DATABASE       — PostgreSQL RLS, tenant_id on all tables
Layer 7: APPLICATION    — Input validation, output sanitisation, DTO filtering
Layer 8: AUDIT          — Complete audit log of sensitive operations
```

---

## 19. Authentication — Session-Based Identity

### 19.1 Why Sessions, Not JWTs

This is a deliberate architectural choice that deserves explanation, because JWTs have become very popular and the question "why not JWTs?" often comes up.

**JWTs are stateless.** A JWT is a self-contained token — the server does not need to store anything. To verify a JWT, the server just checks its cryptographic signature. This is elegant and scales well.

**The statelessness of JWTs is also their weakness for this use case.** Consider what happens when you need to invalidate a token:

- A user is fired from Shell Maanzoni. Their JWT should immediately stop working.
- A tenant is suspended. All their users' JWTs should immediately stop working.
- A security breach is detected. All sessions should be invalidated.

With JWTs, you cannot do any of these instantly. The token remains valid until it expires. You can maintain a "token blocklist," but then you are essentially adding statefulness back to the system — and now you have the worst of both worlds.

With sessions, all of these scenarios are handled by deleting the session record from Redis. The session is gone. The next request fails authentication. Instantaneous.

Awo ERP's threat model specifically includes the scenarios above (tenant suspension, employee termination). Sessions are the right tool.

**Additionally,** for a multi-tenant ERP handling financial data, the session data itself (tenant_id, role, permissions) should be server-authoritative. With JWTs, the claims are baked into the token. If a user's role changes (promoted from cashier to manager), the old JWT continues to grant the old role until expiry. With sessions, the session data in Redis is updated immediately and the next request uses the new role.

### 19.2 Session Design

Sessions are stored in Redis as hashes under the key `session:{session_id}` for tenant sessions and `platform:session:{session_id}` for platform sessions.

**Session ID generation:** The session ID is a cryptographically random 32-byte value encoded as a hex string (64 characters). Using Go's `crypto/rand`:

```go
func generateSessionID() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}
```

64 hexadecimal characters give 256 bits of entropy. This is brute-force-proof. The expected time to find a valid session ID by guessing is longer than the age of the universe.

**Session cookie attributes:**

```
Set-Cookie: session_id=<64-hex-chars>;
            Path=/;
            Domain=.awo.app;
            Secure;
            HttpOnly;
            SameSite=Strict;
            Max-Age=28800
```

- `Secure` — Cookie is only sent over HTTPS. Never over HTTP.
- `HttpOnly` — JavaScript cannot read this cookie. XSS attacks cannot steal the session ID.
- `SameSite=Strict` — Cookie is not sent in cross-origin requests. CSRF attacks cannot use the cookie.
- `Domain=.awo.app` — Cookie is shared across all `*.awo.app` subdomains, allowing a login on `shell-maanzoni.awo.app` to be valid for the API calls from that page.
- `Max-Age=28800` — Cookie expires in 8 hours in the browser.

**Session expiry in Redis:** The Redis key has a TTL of 8 hours. The server-side expiry is the authoritative one. Even if the browser cookie persists past its `Max-Age` (e.g., due to a browser bug), the server will reject the stale session ID.

### 19.3 Session Rotation

**What is session rotation?** When a user logs in, they get a session ID. After a fixed interval (or on privilege changes), that session ID is retired and replaced with a new one. This limits the window of opportunity for a stolen session ID to be misused.

Awo ERP rotates session IDs:
- After login (the "unauthenticated" pre-login session is replaced with an authenticated session)
- After role elevation (e.g., when a manager approves their own action with a secondary authentication step)
- Every 2 hours for platform user sessions (higher security requirement)

Session rotation in Go/Redis:

```go
func rotateSession(ctx context.Context, redis *redis.Client,
    oldSessionID string, sessionData SessionData) (string, error) {

    newSessionID, err := generateSessionID()
    if err != nil {
        return "", err
    }

    pipe := redis.Pipeline()
    // Write new session
    pipe.HSet(ctx, "session:"+newSessionID, sessionData.ToMap())
    pipe.Expire(ctx, "session:"+newSessionID, 8*time.Hour)
    // Delete old session
    pipe.Del(ctx, "session:"+oldSessionID)
    _, err = pipe.Exec(ctx)
    return newSessionID, err
}
```

Using a Redis pipeline ensures the old session is deleted atomically with the new one being created, preventing a brief window where both are valid.

### 19.4 Session Validation in the Middleware

The session middleware runs after tenant resolution. It reads the session cookie, looks up the session in Redis, and validates:

1. Session exists (not expired or invalidated)
2. Session's `tenant_id` matches the resolved tenant from the middleware
3. Session is not flagged as compromised

Point 2 is important: a session created for Shell Maanzoni cannot be used to access Ocean Grill's API, even if the session ID is somehow obtained. This is the session layer of the dual-signal model.

### 19.5 Platform vs. Tenant Session Namespacing

As noted, platform sessions live at `platform:session:<id>` and tenant sessions live at `session:<id>`. The platform session middleware only looks in the `platform:session:` namespace; the tenant session middleware only looks in the `session:` namespace. There is no cross-namespace fallthrough. A platform session ID presented to a tenant API route is treated as invalid.

---

## 20. Authorisation & Tenant Isolation

### 20.1 PostgreSQL Row-Level Security — How It Enforces Isolation

Row-Level Security is a PostgreSQL feature where the database engine itself enforces who can see which rows, based on a *policy* attached to the table. No matter what SQL query the application runs, the database automatically adds the policy's conditions.

Think of it as the database adding an invisible `WHERE tenant_id = 'current-tenant'` to every query, automatically, without the application needing to remember to do it.

**The mechanism:**

Every table that contains tenant data has a policy like this:

```sql
-- Enable RLS on the table
ALTER TABLE fuel_deliveries ENABLE ROW LEVEL SECURITY;
ALTER TABLE fuel_deliveries FORCE ROW LEVEL SECURITY;

-- Policy: users can only see rows belonging to their tenant
CREATE POLICY tenant_isolation ON fuel_deliveries
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

`FORCE ROW LEVEL SECURITY` ensures that even the table owner (the application database user) is subject to the policy. Without `FORCE`, the table owner can bypass RLS.

### 20.2 The `SET LOCAL` Pattern in Go

Before any query runs in a database transaction, the Go code sets the PostgreSQL session variable that the RLS policy checks:

```go
func WithTenantContext(ctx context.Context, pool *pgxpool.Pool,
    tenantID uuid.UUID, fn func(pgx.Tx) error) error {

    return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
        // SET LOCAL only affects the current transaction
        _, err := tx.Exec(ctx,
            "SET LOCAL app.current_tenant_id = $1",
            tenantID.String(),
        )
        if err != nil {
            return err
        }
        return fn(tx)
    })
}
```

`SET LOCAL` (as opposed to `SET`) means the variable is only set for the duration of the current transaction. When the transaction ends (commits or rolls back), the variable returns to its default (empty). This is critical: if a connection is returned to the pool with a stale `app.current_tenant_id`, the next request using that connection could query the wrong tenant's data.

`SET LOCAL` within a transaction eliminates this risk because the variable is always cleared at transaction end.

### 20.3 What Happens If RLS Is Bypassed

If the `SET LOCAL` call is omitted — perhaps in a code path where a developer forgot to use the `WithTenantContext` helper — then `current_setting('app.current_tenant_id')` returns an empty string. The RLS policy then evaluates `tenant_id = ''::uuid`, which returns false for all rows. The query returns no results.

This is the "fail-safe" posture: a bug in tenant context setting results in empty results, not wrong-tenant data. The developer will notice that data is missing (the feature does not work) rather than seeing a data breach.

The one failure mode that is NOT safe-by-default is if a developer uses the superuser/admin role for application queries, because that role bypasses RLS. Application code must never use the superuser role. Only Temporal workflows performing cross-tenant platform operations (e.g., generating a platform-wide billing report) may use the elevated role, and only with explicit intent and audit logging.

### 20.4 Redis Namespace Isolation

Redis does not have native tenant isolation. The entire Redis instance is a shared key-value store. Isolation is achieved entirely through key naming conventions.

Every key related to a specific tenant is prefixed with `tenant:{tenant_id}:`. Platform-level keys use `platform:`. Cross-cutting keys (e.g., global rate limit counters) use `global:`.

```
tenant:7f3a...:session:abc123           → tenant user session
tenant:7f3a...:schema:fuel_deliveries   → cached amis schema
tenant:7f3a...:slug                     → slug → UUID mapping
tenant:7f3a...:rate_limit:user:xyz      → per-user rate limit counter

platform:session:def456                 → platform user session
platform:rate_limit:ip:1.2.3.4         → IP-based rate limit

global:onboarding:in_progress          → system-wide flag
```

All Redis key construction goes through a `CacheKey` builder function in Go that enforces the prefix:

```go
type CacheKey struct {
    tenantID uuid.UUID
}

func (k CacheKey) Schema(pageName, role string) string {
    return fmt.Sprintf("tenant:%s:schema:%s:%s", k.tenantID, pageName, role)
}

func (k CacheKey) Session(sessionID string) string {
    return fmt.Sprintf("tenant:%s:session:%s", k.tenantID, sessionID)
}
```

No Redis key is ever constructed by string interpolation in handler code directly. All key construction goes through this builder. This makes it easy to audit all cache key patterns and prevents typo-based namespace collisions.

---

## 21. Header Security

### 21.1 Trusting `X-Tenant-ID` — Attack Vectors

The `X-Tenant-ID` header is set by the client. This means a malicious client can set it to any value they want. The middleware must therefore not *trust* the header blindly — it must *validate* it against an authoritative source (the slug-to-UUID mapping from Redis/PostgreSQL).

The dual-signal model (Section 5.4) is the primary defence: the header value must match what the subdomain resolves to. A client cannot set `X-Tenant-ID` to a different tenant's UUID and use it on another tenant's subdomain. The cross-check catches it.

Additionally, the session validation (Section 19.4) provides a second line of defence: the session's stored `tenant_id` must also match. An attacker who somehow gets a valid session for Tenant A cannot use it against Tenant B even if they manipulate the `X-Tenant-ID` header, because the session's `tenant_id` will not match.

### 21.2 Internal-Only Headers and Gateway Enforcement

The reverse proxy (Nginx/Caddy/load balancer) is configured to *strip* certain headers from incoming client requests before forwarding to the Go application. Specifically:

- `X-Forwarded-Tenant-ID` (if Awo ERP uses this internally for inter-service communication, clients must not be able to forge it)
- `X-Platform-User-ID` (used internally for impersonation, must not be settable by clients)
- `X-Internal-*` (any internal-use header prefix)

If a client sends these headers, the gateway strips them silently. The application then only receives them when they are set by the gateway itself (for legitimate internal routing), not from client forgery.

This is a critical configuration that must be verified in the gateway configuration and tested regularly. A penetration test that checks for header injection should be part of the regular security review.

### 21.3 Header Injection and Spoofing Prevention

**CRLF injection:** An attacker might attempt to inject additional HTTP headers by including `\r\n` (carriage return + line feed) characters in a header value. Fiber (built on `fasthttp`) automatically rejects requests with CRLF characters in header values, preventing this class of attack.

**Overly long headers:** `X-Tenant-ID` must be exactly 36 characters (UUID format). The middleware rejects any value that is not a valid UUID format before attempting any database or cache lookup. This prevents large-value DoS attacks on the tenant resolution logic.

---

## 22. Transport Security

### 22.1 TLS Requirements

All traffic to and from Awo ERP is encrypted. The specific requirements:

- Minimum TLS version: **TLS 1.2**. TLS 1.0 and 1.1 are disabled.
- Preferred TLS version: **TLS 1.3** (faster handshakes, stronger security).
- Cipher suites: Only AEAD cipher suites are allowed. RC4, DES, and export ciphers are disabled.
- Certificates: Minimum RSA 2048-bit or ECDSA 256-bit. SHA-1 signature algorithm is not permitted.

These requirements are enforced at the reverse proxy level, not in the Go application. The Go application operates on the internal network where TLS termination has already occurred.

### 22.2 HSTS and Subdomain Pinning

HTTP Strict Transport Security (HSTS) tells browsers to *always* use HTTPS for the domain, even if the user types `http://`. The HSTS header is:

```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```

- `max-age=31536000` — Browsers remember to use HTTPS for 1 year.
- `includeSubDomains` — This applies to all `*.awo.app` subdomains. Critical for tenant subdomains.
- `preload` — The domain is submitted to browser HSTS preload lists, meaning browsers refuse HTTP connections even on first visit (before they have seen the HSTS header).

The `preload` directive is the highest level of HSTS protection. Once a domain is on the preload list, removing it requires a lengthy process. Awo ERP commits to always serving HTTPS before submitting to preload lists.

---

## 23. Rate Limiting and Abuse Prevention

### 23.1 Per-Tenant Rate Limits in Redis

Rate limiting prevents a single tenant or user from consuming disproportionate API resources, whether maliciously or accidentally (a runaway script, a misconfigured integration).

Awo ERP implements a **token bucket** algorithm in Redis. Each tenant has a bucket of tokens. Every API request consumes one token. Tokens refill at a fixed rate. When the bucket is empty, requests are rejected with `429 Too Many Requests`.

The parameters are configurable per-tenant subscription tier:

| Tier | Requests/minute | Burst |
|---|---|---|
| Standard | 300 | 50 |
| Professional | 1,000 | 100 |
| Enterprise | 5,000 | 500 |

Redis stores the bucket state at:
```
tenant:{tenant_id}:rate_limit:bucket
```

The implementation uses Redis's `INCRBY` + `EXPIRE` for the token count and a Lua script to make the check-and-decrement atomic.

### 23.2 Platform-Level Rate Limits

Beyond per-tenant limits, the platform enforces limits at the IP address level (to prevent credential stuffing at the login endpoint) and at the platform API level (to protect against abusive platform user scripts).

Login endpoint: Maximum 10 attempts per IP per 15-minute window. After 5 failed attempts for a specific account, the account is temporarily locked and the user is notified by email.

---

## 24. Secret Management and Key Rotation

### 24.1 What Qualifies as a Secret

In Awo ERP, the following are secrets and must never be stored in code, Git, or configuration files:

- PostgreSQL connection string (includes password)
- Redis connection string (if password-protected)
- Session signing key (used to sign session IDs before storage, if applicable)
- Temporal namespace credentials
- Email provider API keys (SendGrid, Mailgun, etc.)
- Any third-party API keys

### 24.2 Secret Storage

Secrets are stored in an environment-appropriate secrets manager:

- **Development:** `.env` file (gitignored) or OS keychain
- **Production:** Cloud provider secrets manager (AWS Secrets Manager, GCP Secret Manager, HashiCorp Vault)

The Go application loads secrets at startup via environment variables. No secrets are read at runtime from the secrets manager (which adds latency and creates a single point of failure). The secrets manager is used to populate environment variables at deployment time.

### 24.3 Key Rotation Policy

| Secret | Rotation Frequency | Rotation Method |
|---|---|---|
| Database password | Every 90 days | Rolling: create new creds, update app, revoke old |
| Redis password | Every 90 days | Same as above |
| Session signing key | Every 30 days | Dual-key: keep old key valid during transition period |
| Platform user API keys | Every 180 days | Immediately on team member departure |
| TLS certificates | Automatic (Let's Encrypt: every 60 days) | Automated via ACME |

Session signing key rotation requires special handling. If the signing key is changed and old sessions were signed with the old key, all existing sessions immediately become invalid, forcing all users to log in again. To avoid this user experience disruption, Awo ERP maintains two valid signing keys during rotation: the "current" key and the "previous" key. Sessions are verified against both. New sessions are signed with the current key. After 8 hours (the maximum session lifetime), all old sessions have either expired or been rotated, and the previous key can be retired.

---

# PART VIII — CACHING STRATEGY

---

## 25. Redis in the Awo ERP Stack

### 25.1 What Gets Cached and Why

Redis serves multiple distinct purposes in Awo ERP. Understanding each one separately prevents confusion about TTLs and invalidation strategies.

| Data Type | Key Pattern | TTL | Invalidation |
|---|---|---|---|
| Tenant session | `session:{id}` | 8 hours | Explicit on logout / rotation |
| Platform session | `platform:session:{id}` | 8 hours | Explicit on logout |
| Tenant context (slug → full object) | `tenant:slug:{slug}` | 15 minutes | On tenant state change |
| Tenant context (id → full object) | `tenant:id:{id}` | 15 minutes | On tenant state change |
| amis schema (composed) | `tenant:{id}:schema:{page}:{role}` | 15 minutes | On schema update |
| Rate limit bucket | `tenant:{id}:rate_limit:bucket` | Sliding (refill algorithm) | Never (self-managing) |
| Login attempt counter | `login:attempts:{ip}` | 15 minutes | On window expiry |

### 25.2 Cache Key Namespacing per Tenant

As described in Section 20.4, all tenant-scoped keys are prefixed with `tenant:{tenant_id}:`. This namespace approach provides two benefits:

1. **Clarity:** Any engineer reading a Redis key can immediately tell which tenant and what type of data it belongs to.
2. **Bulk invalidation:** When a tenant's state changes (suspended, schema updated), all their cache entries can be invalidated by scanning the `tenant:{id}:*` pattern and deleting matches.

Bulk invalidation via `SCAN` + `DEL` is safer than `KEYS *` which blocks the Redis event loop. The SCAN-based approach is iterative and non-blocking.

### 25.3 Cache Invalidation on Tenant State Changes

When a tenant is suspended, reinstated, or has their schema updated, a cache invalidation event is triggered. The event flows:

1. PostgreSQL trigger fires on `tenants.status` column update
2. The Go application (which performed the update) calls `InvalidateTenantCache(tenantID)` explicitly
3. This function scans `tenant:{id}:*` in Redis and deletes all matching keys
4. The next request for any resource in that tenant will be a cache miss and will load fresh data from PostgreSQL

The deliberate choice here is to have the *application* (not PostgreSQL) trigger the cache invalidation, because the application has the Redis connection. PostgreSQL triggers cannot call Redis. A listener pattern (LISTEN/NOTIFY in PostgreSQL, handled by a Go goroutine) could also work, but direct application-level invalidation is simpler to reason about.

### 25.4 amis Schema Cache — TTL Strategy and Manual Busting

Schema cache entries have a 15-minute TTL. This means that if a base schema is updated (new Awo ERP version deployed), all tenants will see the new schema within 15 minutes even without explicit cache invalidation. This is the "eventual consistency" guarantee for schema updates.

Manual cache busting is also available: a platform admin can trigger a `POST /api/v1/platform/cache/bust?type=schemas` endpoint that iterates all tenant IDs and invalidates all schema cache entries. This is used immediately after a major schema update to ensure all tenants see the new UI without waiting for TTL expiry.

### 25.5 Redis as a Rate-Limit Store

The rate-limiting role of Redis is secondary but important. The token bucket state must be shared across all API server instances (if Awo ERP runs multiple Go instances behind a load balancer). A local in-memory rate limiter would not work in this scenario because each instance would have an independent view of the bucket. Redis provides the shared state.

The rate-limit keys use a short TTL (equal to the rate window, typically 1 minute) and are self-expiring. No explicit invalidation is needed.

### 25.6 Redis Failure Handling

If Redis becomes unavailable, Awo ERP's behaviour depends on the specific Redis use:

- **Session lookup failure:** All API requests fail with 503 Service Unavailable. This is unavoidable — sessions cannot be validated without Redis.
- **Tenant cache failure:** Fall back to PostgreSQL for tenant resolution. Slower, but functional.
- **Schema cache failure:** Fall back to on-demand composition from PostgreSQL and file system. Slower, but functional.
- **Rate limit failure:** Fail open (allow requests through). Failing closed (blocking all requests) would make a Redis outage a denial of service. Failing open creates a brief window of unlimited rate, which is acceptable for a short Redis outage.

The application must distinguish between Redis "key not found" (normal cache miss) and Redis "connection error" (infrastructure failure). In Go, this is done by checking the error type returned by the Redis client.

---

# PART IX — TROUBLESHOOTING

---

## 26. Common Tenant-Routing Failures

### 26.1 Subdomain Resolves but Tenant Not Found

**Symptom:** A request to `valid-company.awo.app` returns 404 with error code `tenant_not_found`.

**Possible causes and diagnostics:**

1. **Tenant was recently terminated.** Check `SELECT status, terminated_at FROM tenants WHERE slug = 'valid-company'` in the database. If status is `TERMINATED`, the tenant has been permanently removed from API access.

2. **Tenant is in PENDING state.** Onboarding workflow has not yet completed. Check the Temporal UI for a workflow with ID `onboarding-{tenant-id}`. If the workflow is `RUNNING`, wait for it to complete. If `FAILED`, see Section 28.

3. **Cache inconsistency.** The tenant exists and is ACTIVE, but the Redis cache has a stale or missing entry. Delete the cache key: `redis-cli DEL "tenant:slug:valid-company"`. The next request will repopulate from PostgreSQL.

4. **Slug mismatch.** The subdomain slug does not match what is stored in the database. This can happen if a slug was manually changed in the database without updating the cache or without following the proper slug-change procedure. Check: `SELECT slug FROM tenants WHERE id = 'known-tenant-id'`.

### 26.2 `X-Tenant-ID` Missing or Mismatched

**Symptom:** API returns 400 `header_required` or 403 `tenant_identity_mismatch`.

**For `header_required`:** The API client is not sending the `X-Tenant-ID` header. Check the client's HTTP configuration. In a browser application, the header should be set in the Axios/fetch base configuration on login, using the `tenant_id` returned by the login response.

**For `tenant_identity_mismatch`:** The header is being sent but its value does not match the UUID associated with the subdomain. This is almost always one of:

- The client has cached a stale `tenant_id` from a previous session. Clearing the client's local storage and logging in again resolves this.
- A platform integration is configured with the wrong tenant UUID. Verify against `SELECT id FROM tenants WHERE slug = 'correct-slug'`.
- A security probe (log as a security event and investigate the source IP).

### 26.3 Tenant Suspended Mid-Session

**Symptom:** A user was working normally and suddenly gets 403 `tenant_suspended` on all requests.

**What happened:** The tenant's status was changed to `SUSPENDED` while the user had an active session. The session is still technically valid, but the tenant resolver now returns `SUSPENDED`.

**Resolution for the user:** Nothing they can do — the account is suspended. They should contact their administrator or Awo ERP support.

**Resolution for platform support:** Verify the reason for suspension. If it was an automated billing suspension and the payment has now been processed, reinstate via `PUT /api/v1/tenants/:id/reinstate`. The existing sessions will work again immediately on the next request (the tenant context cache expires in 15 minutes, or bust it manually).

---

## 27. amis-UI Failures

### 27.1 Schema Returns 200 but Page Renders Blank

**Symptom:** The API returns a `status: 0` response with a schema object, but the amis renderer shows nothing.

**Diagnostics:**

1. **Schema type mismatch.** amis requires the top-level `type` field to match a known component type. If the type is misspelled or wrong for the context, amis silently renders nothing. Check the `type` field in the returned schema JSON.

2. **Empty `body` array.** If the schema has `"body": []`, amis renders an empty page — correctly, by definition. Check that the schema composition is producing the expected `body` content.

3. **amis version incompatibility.** If the amis frontend library version does not support a component type used in the schema, it may silently skip it. Check the amis version and the component support matrix.

4. **CORS blocking.** If the schema API call is being blocked by CORS (visible in browser developer tools Network tab as a CORS error), the page will be blank. Verify the CORS configuration includes the tenant's origin.

### 27.2 Data Loads but Wrong Fields Appear (Schema-Data Mismatch)

**Symptom:** The table or form loads, but field values appear in the wrong columns, or expected fields are empty.

**Most common cause:** The schema's `name` field for a column does not match the JSON key returned by the data API. amis matches data to schema fields by name. If the schema says `"name": "delivery_date"` but the API returns `"deliveryDate"`, the column will be empty.

**Fix:** Ensure Go struct JSON tags match the field names used in amis schemas. Use snake_case throughout for consistency. Update the struct tag or the schema, and invalidate the schema cache.

### 27.3 `visibleOn` Expressions Behaving Unexpectedly

**Symptom:** Fields are visible to users who should not see them, or are hidden from users who should see them.

**Diagnostics:**

1. **Check what role is injected into the amis context.** The Go handler injects `role` into the schema's `data` section. Verify this by inspecting the raw API response in the browser developer tools.

2. **amis expression syntax.** The `visibleOn` expression uses amis's own expression language (not JavaScript). Verify the syntax in the amis documentation. For example, `${role == 'manager'}` is correct; `${role === 'manager'}` (triple equals) may not work as expected.

3. **Remember:** `visibleOn` is a UX mechanism, not a security mechanism. If security is the concern, fix it server-side (Section 17.1).

### 27.4 Schema Cache Serving Stale UI to a Tenant

**Symptom:** A schema was updated (new field added, label changed) but users still see the old version.

**Immediate fix:**

```bash
# Connect to Redis and delete the stale schema cache entry
redis-cli SCAN 0 MATCH "tenant:<tenant-id>:schema:*" COUNT 100
# For each key returned:
redis-cli DEL "tenant:<tenant-id>:schema:<page-name>:<role>"
```

Or use the platform admin API:
```
POST /api/v1/platform/cache/bust?tenant_id=<uuid>&type=schemas
```

**Root cause investigation:** Check whether the schema update triggered the cache invalidation logic. If it did not, the invalidation code path may have a bug or the PostgreSQL trigger is not firing correctly.

---

## 28. Onboarding Failures

### 28.1 Reading Temporal Workflow History

Every step of the onboarding workflow is recorded in Temporal's history. To diagnose a failed onboarding:

1. Open the Temporal Web UI (typically at `http://temporal-ui:8080`)
2. Navigate to the namespace: `awo-production` (or `awo-staging`)
3. Search for workflow ID: `onboarding-{tenant-uuid}`
4. Click into the workflow to see the event history

The event history shows every activity that ran, when it ran, how long it took, and — for failed activities — the exact error message and stack trace. This is the primary diagnostic tool.

### 28.2 Replaying a Failed Activity Safely

If an activity failed due to a transient error (database overload, external API timeout) and the underlying problem is now resolved, Temporal can resume the workflow:

```bash
# Using Temporal CLI
temporal workflow signal \
  --workflow-id "onboarding-{tenant-uuid}" \
  --name "retry-failed-activity" \
  --namespace awo-production
```

For non-transient failures (e.g., the seeded COA data has a bug), fix the bug, deploy the fix, and then use Temporal's "Reset" feature to replay the workflow from the failed activity forward. Temporal will re-run only the failed activity and those that come after it, using the already-completed results for the earlier activities.

### 28.3 Tenant Stuck in Pending State — Recovery Checklist

If a tenant has been in `PENDING` state for more than 30 minutes, something has gone wrong. Recovery checklist:

- [ ] Find the workflow in Temporal UI: `onboarding-{tenant-uuid}`
- [ ] Check workflow status: Running / Failed / Terminated
- [ ] If `FAILED`: read the failing activity and error. Fix the underlying issue, then replay.
- [ ] If `RUNNING` but blocked: check activity heartbeats. A heartbeat timeout means the activity has been running too long and will be retried by Temporal automatically.
- [ ] If `TERMINATED` unexpectedly: check the Temporal worker logs for the time the workflow was terminated.
- [ ] If workflow does not exist in Temporal: the workflow was never started. Check the API server logs for errors at the time of signup submission.
- [ ] After resolution, verify tenant status in DB: `SELECT status FROM tenants WHERE id = '{uuid}'`
- [ ] If status is ACTIVE and subdomain resolves correctly, the tenant is live.

---

## 29. Platform User Access Issues

### 29.1 Impersonation Token Errors

**Symptom:** A support agent attempts to impersonate a tenant and receives an error.

**Common causes:**

1. **Impersonation session already expired.** Impersonation sessions are valid for 1 hour only. Create a new impersonation session.
2. **Tenant is SUSPENDED or TERMINATED.** Impersonation is only available for ACTIVE tenants. Support access to suspended/terminated tenant data requires a direct database query by a Super Admin.
3. **Support agent's platform session has expired.** The agent needs to log in again at `platform.awo.app/login`.
4. **Insufficient role.** Read-Only Auditors cannot initiate impersonation. The agent may need their role elevated by a Super Admin.

### 29.2 RLS Bypasses and How to Diagnose Them

A suspected RLS bypass is a critical security event. The symptom is data appearing in a response that should not be there — data belonging to a different tenant.

**Immediate action:**

1. Preserve evidence: capture the full request and response that triggered the suspicion.
2. Check whether the Go handler used `WithTenantContext` or called the database directly without setting `app.current_tenant_id`.
3. Check whether the query ran against the correct table (which has RLS enabled) or against a view or materialized view that may not carry the RLS policy through.
4. Verify RLS is enabled on the table: `SELECT relrowsecurity FROM pg_class WHERE relname = 'table_name'`
5. Test the policy directly in PostgreSQL: connect as the application user, set `SET LOCAL app.current_tenant_id = 'tenant-a-uuid'`, and query data that should only show Tenant B's records. If Tenant B's data appears, the policy is wrong.

PostgreSQL views do not automatically inherit RLS from the underlying table unless the view has `security_invoker = true` set. Functions called from queries must also be `SECURITY INVOKER` (the default) not `SECURITY DEFINER` (which runs as the function creator, bypassing RLS).

---

## 30. Performance Troubleshooting

### 30.1 Slow Queries — Missing `tenant_id` Index

Every table with a `tenant_id` column must have an index on that column (or a composite index where `tenant_id` is the first column). Without this index, PostgreSQL performs a full table scan for every query, filtered by the RLS policy — scanning all rows across all tenants to find the current tenant's rows.

Verify with:
```sql
EXPLAIN ANALYZE
SELECT * FROM fuel_deliveries
WHERE delivery_date >= NOW() - INTERVAL '30 days';
```

If the query plan shows a `Seq Scan` on a large table, a `tenant_id` index is likely missing. The correct index for most tenant-scoped queries is:

```sql
CREATE INDEX idx_fuel_deliveries_tenant_date
    ON fuel_deliveries (tenant_id, delivery_date DESC);
```

### 30.2 Redis Eviction Causing Thundering-Herd

If Redis is under memory pressure and begins evicting keys, multiple servers may simultaneously experience cache misses for the same key and all query PostgreSQL at once. This "thundering herd" can overload PostgreSQL.

Prevention:
- Monitor Redis memory usage and `evicted_keys` counter. Alert when `evicted_keys` is non-zero.
- Set Redis `maxmemory-policy` to `allkeys-lru` (least recently used eviction). This evicts the least-used keys first, which is less likely to be a frequently-accessed tenant's schema.
- Add jitter to TTLs: instead of `15 * time.Minute` exactly, use `15*time.Minute + time.Duration(rand.Intn(120))*time.Second`. This spreads out expiry times, reducing simultaneous expiry of related keys.

### 30.3 Large amis Schema Payloads Causing Slow Page Loads

A complex amis schema can be several hundred kilobytes if it describes a full page with many components. For users on slower connections (mobile networks in some parts of Kenya), this adds noticeable load time.

Optimisations:
1. **Split schemas:** Instead of one schema for the entire page, split into top-level schema (navigation, header) and per-tab schemas (loaded lazily when a tab is activated).
2. **Schema compression:** Enable gzip compression in Fiber for API responses over 1KB. amis schemas are text (JSON) and compress extremely well (typically 80-90% reduction).
3. **Longer schema TTLs:** Schemas that change infrequently (monthly updates) can have their Redis TTL extended to 1 hour, reducing schema fetch frequency for users who load the same page repeatedly.

---

## 31. Observability Checklist

Good observability means knowing what is happening in the system *before* users report problems. The following metrics, logs, and traces should be in place.

### 31.1 Metrics to Monitor

| Metric | Alert Threshold | What It Indicates |
|---|---|---|
| `api.request_duration_p99` | > 2 seconds | Slow API, database performance issue |
| `api.error_rate` | > 1% of requests | General API health |
| `auth.session_invalid_rate` | > 5% of requests | Possible credential stuffing or session bug |
| `tenant.mismatch_count` | Any | Security probe or client misconfiguration |
| `redis.evicted_keys` | Any | Redis memory pressure |
| `redis.connection_errors` | Any | Redis connectivity issue |
| `temporal.workflow_failures` | Any | Onboarding or other workflow failures |
| `rate_limit.exceeded_count` | > 100/minute | Possible abuse or runaway client |
| `db.slow_queries` | > 500ms | Missing index or query optimisation needed |
| `amis.schema_cache_miss_rate` | > 20% | Cache invalidation too aggressive or Redis issues |

### 31.2 Log Events to Capture

Every log event should include: `request_id`, `tenant_id` (if resolved), `user_id` (if authenticated), `timestamp`, `severity`.

Key events to log:

- Every `tenant_identity_mismatch` (security/warn)
- Every `session_tenant_mismatch` (security/warn)
- Every login attempt (success and failure) with IP and user agent
- Every platform user action (info)
- Every impersonation session start and end (audit/info)
- Every tenant state change (audit/info)
- Every schema cache invalidation (debug/info)
- Every Temporal activity failure (error)
- Every rate limit exceeded event (warn)

### 31.3 Distributed Tracing

In Go, use OpenTelemetry to propagate trace context across the request lifecycle. A single user request should produce a trace that spans: Fiber handler → Redis lookup → PostgreSQL query → Redis write → response. This makes it possible to see exactly where time is spent in a slow request.

Temporal workflows should also emit traces, linking the workflow execution to the API request that triggered it.

### 31.4 Healthcheck Endpoints

```
GET /health         — Returns 200 if the Go application is running
GET /health/ready   — Returns 200 only if Redis and PostgreSQL are reachable
GET /health/live    — Returns 200 if the application is not deadlocked
```

The load balancer uses `/health/live` for instance health checks. Kubernetes (if used) uses `/health/ready` for readiness probes. `/health` is the most basic check, suitable for external monitoring.

---

# APPENDICES

---

## Appendix A — Fiber Middleware Stack — Annotated Code Reference

```go
// main.go — Fiber application setup

app := fiber.New(fiber.Config{
    // Disable Fiber's default error handler; use our custom one
    ErrorHandler: customErrorHandler,
    // Increase body size limit for large amis schema uploads
    BodyLimit: 4 * 1024 * 1024, // 4MB
    // Trust X-Forwarded-For from the load balancer only
    ProxyHeader: fiber.HeaderXForwardedFor,
    EnableTrustedProxyCheck: true,
    TrustedProxies: []string{"10.0.0.0/8"}, // Internal LB subnet
})

// LAYER 1: Global middleware (all requests)
app.Use(requestid.New())         // Assigns unique ID to every request
app.Use(logger.New())            // Structured request logging
app.Use(recover.New())           // Panic recovery
app.Use(corsMiddleware())        // CORS with *.awo.app whitelist

// LAYER 2: Subdomain extraction
app.Use(SubdomainExtractor("awo.app"))

// LAYER 3: Route groups
platform := app.Group("/api/v1", PlatformSessionMiddleware(redis))
tenant := app.Group("/api/v1",
    TenantResolver(redis, db),         // Resolves X-Tenant-ID
    TenantSessionMiddleware(redis, db), // Validates session
    RateLimiter(redis),                 // Per-tenant rate limiting
)

// Register platform routes
platform.Get("/tenants", handlers.ListTenants)
platform.Post("/tenants/:id/impersonate", handlers.ImpersonateTenant)

// Register tenant routes
tenant.Get("/schema/:page", handlers.GetAmisSchema)
tenant.Get("/fuel-deliveries", handlers.ListFuelDeliveries)
```

---

## Appendix B — PostgreSQL RLS Policy Templates

```sql
-- Template for any tenant-scoped table
-- Replace 'table_name' with the actual table name

ALTER TABLE table_name ENABLE ROW LEVEL SECURITY;
ALTER TABLE table_name FORCE ROW LEVEL SECURITY;

-- Tenant isolation policy (applies to SELECT, INSERT, UPDATE, DELETE)
CREATE POLICY tenant_isolation ON table_name
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Platform superuser bypass (for cross-tenant operations)
-- Only the 'awo_platform' role bypasses RLS
CREATE POLICY platform_admin_bypass ON table_name
    TO awo_platform
    USING (true);

-- Application role (no RLS bypass)
GRANT SELECT, INSERT, UPDATE, DELETE ON table_name TO awo_app;

-- Platform role (RLS bypass for analytics and support)
GRANT SELECT ON table_name TO awo_platform;
```

---

## Appendix C — Temporal Workflow Definitions — Reference

```go
// workflows/onboarding.go

type OnboardingWorkflowInput struct {
    TenantID        uuid.UUID
    TenantName      string
    Slug            string
    IndustryVertical string
    ContactEmail    string
    ProvisionedBy   string // "self_service" or platform_user_id
}

func OnboardingWorkflow(ctx workflow.Context, input OnboardingWorkflowInput) error {
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
            InitialInterval: 5 * time.Second,
            BackoffCoefficient: 2.0,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Activity 1
    if err := workflow.ExecuteActivity(ctx,
        activities.ValidateAndReserveSubdomain,
        input.Slug, input.TenantID).Get(ctx, nil); err != nil {
        return fmt.Errorf("reserve subdomain: %w", err)
    }

    // Activity 2
    if err := workflow.ExecuteActivity(ctx,
        activities.ProvisionTenantRecord, input).Get(ctx, nil); err != nil {
        return fmt.Errorf("provision tenant: %w", err)
    }

    // Activities 3-7 follow the same pattern...

    // Final activation
    return workflow.ExecuteActivity(ctx,
        activities.ActivateTenant, input.TenantID).Get(ctx, nil)
}
```

---

## Appendix D — amis Schema Envelope Reference Templates

### List View (CRUD Page)
```json
{
  "status": 0,
  "msg": "",
  "data": {
    "type": "page",
    "title": "Page Title",
    "body": {
      "type": "crud",
      "syncLocation": false,
      "api": "/api/v1/resource",
      "defaultParams": { "per_page": 25 },
      "columns": [],
      "headerToolbar": ["bulk-actions", "export-csv", "columns-toggler"],
      "footerToolbar": ["statistics", "pagination"],
      "filter": { "type": "form", "body": [] }
    }
  }
}
```

### Create / Edit Form
```json
{
  "status": 0,
  "msg": "",
  "data": {
    "type": "form",
    "title": "Form Title",
    "body": [],
    "api": { "method": "post", "url": "/api/v1/resource" },
    "redirect": "/list-page-url"
  }
}
```

### Data Response (for CRUD api field)
```json
{
  "status": 0,
  "msg": "",
  "data": {
    "items": [],
    "total": 0,
    "page": 1,
    "perPage": 25
  }
}
```

---

## Appendix E — Redis Key Namespace Reference

| Key Pattern | Type | TTL | Description |
|---|---|---|---|
| `session:{session_id}` | Hash | 8h | Tenant user session data |
| `platform:session:{session_id}` | Hash | 8h | Platform user session data |
| `tenant:slug:{slug}` | String (JSON) | 15m | Slug → tenant context |
| `tenant:id:{tenant_id}` | String (JSON) | 15m | ID → tenant context |
| `tenant:{id}:schema:{page}:{role}` | String (JSON) | 15m | Composed amis schema |
| `tenant:{id}:rate_limit:bucket` | String | Sliding | Token bucket state |
| `login:attempts:{ip}` | Integer | 15m | Failed login counter |
| `platform:rate_limit:ip:{ip}` | Integer | 1m | Platform IP rate limit |
| `global:onboarding:slug:{slug}` | String | 24h | Slug reservation during onboarding |

---

## Appendix F — Environment Variables and Configuration Reference

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | PostgreSQL connection string (`postgres://user:pass@host:5432/dbname?sslmode=require`) |
| `REDIS_URL` | Yes | Redis connection URL (`redis://:password@host:6379/0`) |
| `TEMPORAL_HOST` | Yes | Temporal server address (`temporal:7233`) |
| `TEMPORAL_NAMESPACE` | Yes | Temporal namespace (`awo-production`) |
| `SESSION_SECRET` | Yes | 32-byte hex key for session ID signing |
| `BASE_DOMAIN` | Yes | Platform base domain (`awo.app`) |
| `APP_ENV` | Yes | `development` / `staging` / `production` |
| `LOG_LEVEL` | No | `debug` / `info` / `warn` / `error` (default: `info`) |
| `RATE_LIMIT_DEFAULT_RPM` | No | Default requests per minute for Standard tier (default: 300) |
| `SESSION_TTL_HOURS` | No | Session lifetime in hours (default: 8) |
| `SCHEMA_CACHE_TTL_MINUTES` | No | amis schema cache TTL (default: 15) |
| `TENANT_CACHE_TTL_MINUTES` | No | Tenant context cache TTL (default: 15) |

---

## Appendix G — Subdomain Naming Rules and Reserved Words

**Allowed character set:** `[a-z0-9-]`

**Rules:**
- Length: 3–63 characters
- Must start with `[a-z0-9]`
- Must end with `[a-z0-9]`
- No consecutive hyphens (`--`)

**Reserved words (cannot be used as tenant slugs):**

`api`, `app`, `admin`, `platform`, `status`, `mail`, `www`, `cdn`, `docs`, `auth`, `health`, `static`, `assets`, `media`, `dev`, `staging`, `demo`, `test`, `sandbox`, `beta`, `v1`, `v2`, `internal`, `manage`, `dashboard`, `billing`, `support`, `help`, `signup`, `login`, `logout`, `account`, `settings`, `system`, `root`, `awo`, `null`, `undefined`

This list should be maintained in code as a constant and checked during both slug validation and the onboarding workflow's `ValidateAndReserveSubdomain` activity.

---

*End of Document — Awo ERP API, Multi-Tenancy & UI Architecture Guide v1.0.0*

*This document is part of the Awo ERP Technical Specification Series. For module-specific specifications (Finance, Notifications, Forecourt), refer to the corresponding documents in the series.*
