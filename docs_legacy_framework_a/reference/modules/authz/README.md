> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

# IAM Module — Identity, Authentication & Authorization

> **Comprehensive guide covering the full IAM platform: Module/Resource/Action registry, feature flags, tenant settings, authentication, Casbin-based authorization, session management, entity hierarchy, and UI generation.**
>
> **Packages:** `internal/platform` (IAM platform) · `internal/core/authz` (Casbin engine)

## Table of Contents

### IAM Platform (internal/platform)

1. [IAM Overview — Philosophy & Scope](./00-iam-overview.md)
2. [Code Architecture & Conventions](./00b-code-architecture.md)
3. [Module/Resource/Action Registry](./03b-mra-registry.md)
4. [Feature Flags](./04b-feature-flags.md)
5. [Tenant Settings](./05b-tenant-settings.md)
6. [Authentication (AuthN)](./06b-authentication.md)
7. [Session — Pre-Computed Everything](./10b-session-precomputation.md)
8. [Entity Hierarchy & Resource Scope](./09b-entity-hierarchy.md)
9. [UI Navigation & Form Generation](./11b-ui-navigation.md)
10. [HTTP Middleware Chain](./12b-http-middleware.md)
11. [IAM Service Interfaces](./14b-iam-services.md)
12. [Cross-Module Integration](./15b-cross-module-integration.md)
13. [Audit Trail](./16b-audit-trail.md)

### Authorization Engine (internal/core/authz)

14. [Executive Summary](./01-executive-summary.md)
15. [Why Authorization in ERP](./02-why-authorization-in-erp.md)
16. [Architecture Overview](./03-architecture-overview.md)
17. [Domain Model — 4 Actor Types](./04-domain-model.md)
18. [Casbin Policy Engine](./05-casbin-policy-engine.md)
19. [Database Architecture](./06-database-architecture.md)
20. [Role Management](./07-role-management.md)
21. [Policy Management](./08-policy-management.md)
22. [Temporal Roles & Expiry](./09-temporal-roles-and-expiry.md)
23. [Domain Isolation](./10-domain-isolation.md)
24. [Middleware & HTTP Integration](./11-middleware-and-http.md)
25. [System Module Integration](./12-system-module-integration.md)
26. [Business Module Integration](./13-business-module-integration.md)
27. [How Other Packages Use authz](./14-how-other-packages-use-authz.md)
28. [Workflow Integration](./15-workflow-integration.md)

### Cross-Cutting

29. [Performance & Caching](./16-performance-and-caching.md)
30. [Security Considerations](./17-security-considerations.md)
31. [Common Business Scenarios](./18-common-business-scenarios.md)
32. [Troubleshooting Guide](./19-troubleshooting-guide.md)
33. [Business Rules & Validation](./20-business-rules-and-validation.md)
34. [API Reference](./21-api-reference.md)
35. [Summary](./22-summary.md)
36. [Test Cases](./testing.md)

---

## Related Modules

| Module | Relationship |
|--------|-------------|
| [Tenant](../tenant/README.md) | Every policy is scoped to a tenant domain; tenant lifecycle triggers IAM events |
| [Settings](../settings/README.md) | Authorization checks respect feature-flag and module-enable settings |
| [Selling](../sell/README.md) | Invoice read/write/approve operations gated by authz policies |
| [Financial](../financial/README.md) | Journal entries, payment approval, GL access all guarded by authz |

## Quick Links

- [Platform Overview](./00-iam-overview.md)
- [MRA Registry](./03b-mra-registry.md)
- [Feature Flags](./04b-feature-flags.md)
- [Authentication Flow](./06b-authentication.md)
- [Session Pre-computation](./10b-session-precomputation.md)
- [Entity Hierarchy](./09b-entity-hierarchy.md)
- [Casbin Model](./05-casbin-policy-engine.md)
- [4 Actor Types](./04-domain-model.md)
- [Database Schema](./06-database-architecture.md)
- [Performance Guide](./16-performance-and-caching.md)
- [Middleware Usage](./11-middleware-and-http.md)
- [Test Cases](./testing.md)
