# IAM Module — Identity, Authentication & Authorization

> **Comprehensive guide covering the full IAM platform: authentication, RBAC via Casbin, session management, entity hierarchy, MFA, SSO, and API keys.**
>
> **Package:** `internal/core/iam` (single unified package as of v1.0)
>
> **Architecture baseline**: RBAC-only, Casbin-driven, Session-as-context.
> All authorization decisions go through `authzService.Enforce()`.
> Session carries identity and configuration context only — no permission map.
>
> **Last reconciled**: 2026-05-11 — all docs verified against implementation.

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

### v1.0 Architecture Reference (NEW — reconciled 2026-05-11)

37. [RBAC Enforcement — Single-Path Model](./rbac-enforcement.md)
38. [Deferred Features — Not Active in v1.0](./deferred-features.md)
39. [Delivery Task Plan](./tasks.md)

### Operational and Administration Guides (NEW — 2026-05-11)

40. [Tenant Administration — Bootstrap, Roles, User Lifecycle](./23-tenant-administration.md)
41. [Platform Administration — Platform Users, Super Admin, Authority Model](./24-platform-administration.md)
42. [User Entity Scope — ALL / SUBTREE / ENTITY_ONLY](./25-user-entity-scope.md)
43. [API Keys and Service Accounts — Lifecycle, Security, Practices](./26-api-keys-and-service-accounts.md)
44. [Resource/Action Ownership — Module Registry, Feature Flags vs Settings vs Prefs](./27-resource-action-ownership.md)

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
