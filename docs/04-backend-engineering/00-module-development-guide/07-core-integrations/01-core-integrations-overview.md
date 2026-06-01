---
title: Core Module Integrations
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Service Layer Overview](../06-service-layer/01-service-overview.md)"
  - "[IAM Architecture](../../../03-platform-architecture/02-iam/01-iam-overview.md)"
  - "[Tenancy Model](../../../03-platform-architecture/01-multi-tenancy/01-tenancy-model.md)"
---

# Core Module Integrations

Every business module integrates with the same set of shared platform services. This document explains each integration point and how to use it correctly.

## Overview

| Platform Service | Interface | What it provides |
|-----------------|-----------|-----------------|
| IAM | `iam.AuthzService` | Permission enforcement |
| Settings | `settings.Service` | Tenant-configurable values |
| Feature Flags | Via `ResolvedSession` | Per-tenant feature toggles |
| Metadata / Attributes | `metadata.Service` | Dynamic attributes on entities |
| Tenant Config | `tenantconfig.Service` | Structural config (currencies, etc.) |

## IAM Integration

Already handled at two levels:

1. **Route-level** — `middleware.Authorize("module.resource.action")` in `routes.go`
2. **Service-level** — `authzSvc.Enforce(ctx, iam.Request{...})` in service methods

See [Service Layer Authorization](../06-service-layer/04-authorization.md) for full patterns.

## Settings

Tenant-configurable settings are key-value pairs accessible in the service:

```go
// In service constructor
type contractService struct {
    settings settings.Service
    // ...
}

// In service method
func (s *contractService) Create(ctx context.Context, sess ResolvedSession, req CreateParams) (*domain.Contract, error) {
    // Read tenant-configured default currency
    defaultCurrency := s.settings.Get(ctx, sess.TenantID, "finance.default_currency", "USD")

    if req.Currency == "" {
        req.Currency = defaultCurrency
    }
    // ...
}
```

Settings are cached in Redis per-tenant. Update via the tenant admin UI or Settings API.

## Feature Flags

Feature flags are loaded into `ResolvedSession` at login time:

```go
// Available in service via session
func (s *contractService) Import(ctx context.Context, sess ResolvedSession, data []byte) error {
    if !sess.FeatureFlags["contracts.bulk_import"] {
        return domain.NewBusinessError(403, "FEATURE_NOT_ENABLED",
            "bulk contract import is not enabled for this tenant")
    }
    // proceed with import
}
```

Feature flags are boolean. Defined per-tenant via the platform admin UI. New flags default to `false`.

### Flag Naming Convention

```
{module}.{feature_name}
```

Examples:
```
contracts.bulk_import
contracts.auto_approve_threshold
finance.multi_currency
hr.payroll_module
```

### Adding a New Flag

1. Define the flag key as a constant in the domain package:
```go
// domain/feature_flags.go
const FlagBulkImport = "contracts.bulk_import"
```

2. Use via session in the service:
```go
if !sess.FeatureFlags[domain.FlagBulkImport] {
    return nil, domain.NewBusinessError(403, "FEATURE_NOT_ENABLED", "feature not enabled")
}
```

3. Enable per-tenant via platform admin.

## Metadata / Attribute Definitions

Modules can attach dynamic attributes to their entities. The attribute system is defined in the `metadata` module:

```go
// Reading attribute values for a contract
attrs, err := s.metaSvc.GetAttributeValues(ctx, metadata.GetRequest{
    TenantID:     sess.TenantID,
    ResourceType: "contract",
    ResourceID:   contractID,
})

// Setting an attribute
err = s.metaSvc.SetAttributeValue(ctx, metadata.SetRequest{
    TenantID:      sess.TenantID,
    ResourceType:  "contract",
    ResourceID:    contractID,
    AttributeKey:  "contract_category",
    AttributeValue: "procurement",
})
```

Attribute definitions (key, type, label, required) are managed by tenant admins. Modules do not hardcode attribute definitions — they are configured at runtime.

## Tenant Configuration

Structural configuration differs from settings. Tenant config includes:

- **Enabled modules**: which business modules are active for this tenant
- **Organizational hierarchy**: entity type definitions
- **Fiscal year**: start month for finance reports
- **Multi-currency**: enabled currencies

```go
// Check if a module is enabled
cfg, err := s.tenantCfg.Get(ctx, tenantID)
if err != nil {
    return err
}
if !cfg.IsModuleEnabled("contracts") {
    return domain.NewBusinessError(403, "MODULE_DISABLED", "contracts module is not enabled")
}
```

## Notification Service

```go
// Notify async — never block the request
s.notifSvc.NotifyAsync(ctx, notifications.Notification{
    TenantID:     sess.TenantID,
    Audience:     notifications.AudienceRole,
    Role:         "contracts.reviewer",
    Category:     domain.NotifContractSubmitted,
    Title:        "Contract Submitted for Review",
    Body:         fmt.Sprintf("Contract %s awaits review.", contract.ContractNumber),
    ResourceID:   contract.ID,
    ResourceType: "contract",
    Priority:     notifications.PriorityNormal,
})
```

See [Notifications Overview](../10-notifications/01-notifications-overview.md) for full patterns.

## Audit Service

```go
// Always async — never block the request
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    defer func() { recover() }()
    s.auditSvc.Log(ctx, audit.Entry{
        TenantID:     sess.TenantID,
        UserID:       sess.UserID,
        Action:       "contracts.contract.create",
        ResourceType: "contract",
        ResourceID:   contract.ID,
        After:        contract,
    })
}()
```

See [Audit Logging Overview](../15-audit-logging/01-audit-overview.md) for patterns.

## Service Constructor with All Integrations

A fully-integrated service constructor:

```go
type contractService struct {
    repo        repository.ContractRepository
    authzSvc    iam.AuthzService
    auditSvc    audit.Service
    notifSvc    notifications.Service
    settings    settings.Service
    metaSvc     metadata.Service
    tenantCfg   tenantconfig.Service
    eventBus    events.Bus
    logger      *zerolog.Logger
    tracer      trace.Tracer
}

func NewContractService(
    repo      repository.ContractRepository,
    authz     iam.AuthzService,
    audit     audit.Service,
    notif     notifications.Service,
    settings  settings.Service,
    meta      metadata.Service,
    tcfg      tenantconfig.Service,
    events    events.Bus,
    logger    *zerolog.Logger,
    tp        trace.TracerProvider,
) *contractService {
    return &contractService{
        repo: repo, authzSvc: authz, auditSvc: audit,
        notifSvc: notif, settings: settings, metaSvc: meta,
        tenantCfg: tcfg, eventBus: events,
        logger: logger, tracer: tp.Tracer("contracts"),
    }
}
```

All dependencies injected via Wire — the service never constructs them.
