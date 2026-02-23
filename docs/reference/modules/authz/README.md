# Authorization Module (authz) — Complete Technical & Business Guide

> **Comprehensive guide covering Casbin-based authorization, role and policy management, domain isolation, performance, and integration with every module in the AWO ERP ecosystem.**

## Table of Contents

1. [Executive Summary](./01-executive-summary.md)
2. [Why Authorization in ERP](./02-why-authorization-in-erp.md)
3. [Architecture Overview](./03-architecture-overview.md)
4. [Domain Model — 4 Actor Types](./04-domain-model.md)
5. [Casbin Policy Engine](./05-casbin-policy-engine.md)
6. [Database Architecture](./06-database-architecture.md)
7. [Role Management](./07-role-management.md)
8. [Policy Management](./08-policy-management.md)
9. [Temporal Roles & Expiry](./09-temporal-roles-and-expiry.md)
10. [Domain Isolation](./10-domain-isolation.md)
11. [Middleware & HTTP Integration](./11-middleware-and-http.md)
12. [System Module Integration](./12-system-module-integration.md)
13. [Business Module Integration](./13-business-module-integration.md)
14. [How Other Packages Use authz](./14-how-other-packages-use-authz.md)
15. [Workflow Integration](./15-workflow-integration.md)
16. [Performance & Caching](./16-performance-and-caching.md)
17. [Security Considerations](./17-security-considerations.md)
18. [Common Business Scenarios](./18-common-business-scenarios.md)
19. [Troubleshooting Guide](./19-troubleshooting-guide.md)
20. [Business Rules & Validation](./20-business-rules-and-validation.md)
21. [API Reference](./21-api-reference.md)
22. [Summary](./22-summary.md)
23. [Test Cases](./testing.md)

---

## Related Modules

| Module | Relationship |
|--------|-------------|
| [Tenant](../tenant/README.md) | Every policy is scoped to a tenant domain; tenant lifecycle affects policy validity |
| [Settings](../settings/README.md) | Authorization checks respect feature-flag and module-enable settings |
| [Selling](../sell/README.md) | Invoice read/write/approve operations gated by authz policies |
| [Financial](../financial/README.md) | Journal entries, payment approval, GL access all guarded by authz |

## Quick Links

- [Casbin Model](./05-casbin-policy-engine.md)
- [4 Actor Types](./04-domain-model.md)
- [Database Schema](./06-database-architecture.md)
- [Performance Guide](./16-performance-and-caching.md)
- [Middleware Usage](./11-middleware-and-http.md)
- [Test Cases](./testing.md)
