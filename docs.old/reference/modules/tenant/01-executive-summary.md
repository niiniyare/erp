[<-- Back to Index](README.md)

## Executive Summary

### Module Purpose

The Tenant Management Module provides the foundational multi-tenant architecture for the AWO ERP platform, enabling:

- **Complete Data Isolation**: Every tenant's data is securely separated at the database level using PostgreSQL Row-Level Security
- **Automated Provisioning**: One-step tenant creation with configuration, admin user setup, and module enablement
- **Subscription Plan Management**: Tiered plans (Basic, Professional, Enterprise) with configurable resource limits
- **Lifecycle Management**: Full tenant lifecycle from onboarding through suspension, reactivation, and archival
- **Bulk Administrative Operations**: Mass suspend, reactivate, archive, and configuration updates across tenants
- **Usage Tracking & Analytics**: Real-time monitoring of tenant resource consumption and activity

### System Architecture Context

The Tenant Management Module is the core foundation that all other modules depend on:

```markdown
CORE SYSTEM CAPABILITIES PROVIDED:

Tenant Isolation:
✓ PostgreSQL Row-Level Security on every table
✓ Session-based tenant context (app.current_tenant_id)
✓ Cross-tenant query prevention at database level
✓ Tenant-specific configuration inheritance

Security Model:
✓ Three database roles: admin_role, application_role, readonly_role
✓ RLS policies enforced on all tenant-scoped tables
✓ Tenant context validation before any data access
✓ Soft-delete protection (deleted tenants blocked)

Provisioning Engine:
✓ Automated tenant creation with slug generation
✓ Default configuration auto-creation via triggers
✓ Admin user setup during provisioning
✓ Module enablement and welcome notifications

Resource Management:
✓ Per-tenant user limits (max_users)
✓ Entity and transaction quotas
✓ Storage quota tracking
✓ API rate limiting per tenant
```

### Integration with Other Modules

The Tenant Module integrates with every module in the AWO ERP:

```markdown
MODULE INTEGRATION MAP:

Financial Module:
- Tenant-scoped chart of accounts
- Fiscal year configuration per tenant
- Currency settings (currency_code on tenant record)
- Accounting method (FIFO, LIFO, weighted average)

Selling Module:
- Tenant-isolated customer records
- Tenant-specific pricing configurations
- Sales data scoped by tenant context

Buying Module:
- Tenant-isolated vendor records
- Procurement settings per tenant
- Purchase limits tied to tenant plan

Entity Module:
- Companies and branches scoped to tenant
- Entity limits enforced (max_entities)
- Cross-entity tenant isolation validation
```

### Key Stakeholders

| Role | Interaction |
|------|------------|
| Platform Administrator | Provisions tenants, manages bulk operations, monitors usage |
| Tenant Administrator | Configures tenant settings, manages users within limits |
| Application Services | Sets tenant context per request, validates access |
| Monitoring Systems | Tracks usage stats, API calls, error rates |

### Success Metrics

| Metric | Target |
|--------|--------|
| Tenant Provisioning Time | < 5 seconds end-to-end |
| Context Switch Overhead | < 1ms per request |
| Data Isolation Violations | Zero tolerance |
| Bulk Operation Throughput | 1000+ tenants per operation |
| Cache Hit Rate | > 95% for tenant lookups |

---

Next: [Why Multi-Tenancy in ERP](./02-why-multi-tenancy-in-erp.md)
