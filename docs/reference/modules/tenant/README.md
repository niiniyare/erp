# Tenant Management Module - Complete Business Domain Guide

> **Comprehensive guide covering multi-tenant architecture, tenant provisioning, configuration management, security isolation, and integration with the AWO ERP ecosystem.**

## Table of Contents

1. [Executive Summary](./01-executive-summary.md)
2. [Why Multi-Tenancy in ERP](./02-why-multi-tenancy-in-erp.md)
3. [Tenant Lifecycle Overview](./03-tenant-lifecycle-overview.md)
4. [Initial Setup & Configuration](./04-initial-setup-and-configuration.md)
5. [Core Database Architecture](./05-core-database-architecture.md)
6. [Tenant Provisioning](./06-tenant-provisioning.md)
7. [Tenant Configuration Management](./07-tenant-configuration-management.md)
8. [Row-Level Security & Isolation](./08-row-level-security-and-isolation.md)
9. [RLS: Flow, Testing & Troubleshooting](./08b-rls-testing-and-troubleshooting.md)
10. [Tenant Context Management](./09-tenant-context-management.md)
11. [Usage Tracking & Analytics](./10-usage-tracking-and-analytics.md)
12. [Subscription Plans & Limits](./11-subscription-plans-and-limits.md)
13. [Bulk Operations](./12-bulk-operations.md)
14. [Tenant Suspension & Reactivation](./13-tenant-suspension-and-reactivation.md)
15. [Archiving & Data Retention](./14-archiving-and-data-retention.md)
16. [Audit Logging & Compliance](./15-audit-logging-and-compliance.md)
17. [Caching & Performance](./16-caching-and-performance.md)
18. [Module Integration Points](./17-module-integration-points.md)
19. [Common Business Scenarios](./18-common-business-scenarios.md)
20. [API Reference](./19-api-reference.md)
21. [Troubleshooting Guide](./20-troubleshooting-guide.md)
22. [Business Rules & Validation](./21-business-rules-and-validation.md)
23. [Summary](./22-summary.md)
24. [Test Cases](./testing.md)

---

## Related Modules

| Module | Relationship |
|--------|-------------|
| [Financial](../financial/README.md) | Tenant-scoped ledgers, fiscal year settings, currency configuration |
| [Selling](../sell/README.md) | Tenant-isolated sales data, customer management, pricing |
| [Buying](../buy/README.md) | Tenant-isolated procurement, vendor management |

## Quick Links

- [Database Schema](./db.md)
- [Provisioning Workflow](./06-tenant-provisioning.md)
- [Security Model](./08-row-level-security-and-isolation.md)
- [Bulk Operations](./12-bulk-operations.md)
- [Test Cases](./testing.md)
