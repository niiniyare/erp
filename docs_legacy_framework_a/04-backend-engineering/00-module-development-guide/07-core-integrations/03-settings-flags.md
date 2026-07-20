> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Settings and Feature Flags
portal: 4 — Backend Engineering
section: 00-module-development-guide/07-core-integrations
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-iam-integration.md
    title: IAM Integration
  - path: ./04-metadata.md
    title: Metadata Integration
---

# Settings and Feature Flags

Both settings and feature flags are snapshotted into the session at login. They are read from the `ResolvedSession` without any database calls during the request.

## Feature Flags

Feature flags gate new or experimental behaviour. The flag value is a boolean.

```go
// Check a feature flag from the session
if sess.FeatureEnabled("contracts.multi_currency") {
	// New multi-currency logic
} else {
	// Legacy single-currency logic
}
```

### Flag Naming Convention

Format: `<module>.<flag_name>`

| Flag | Meaning |
|------|---------|
| `contracts.multi_currency` | Allow contracts in currencies other than tenant default |
| `contracts.auto_activation` | Auto-activate contracts on start date |
| `contracts.approval_workflow` | Use Temporal workflow for approvals |
| `contracts.vendor_validation` | Validate vendor exists via vendor service on create |

### Passing Flags to the Service

The handler reads the flag from the session and passes the resolved boolean to the service. The service receives a plain bool — it does not know about sessions or flags:

```go
// handler
multiCurrency := sess.FeatureEnabled("contracts.multi_currency")

contract, err := h.service.Create(ctx, service.CreateContractRequest{
	// ...
	AllowMultiCurrency: multiCurrency,
})
```

```go
// service
func (s *contractService) Create(ctx context.Context, req CreateContractRequest) (*domain.Contract, error) {
	if !req.AllowMultiCurrency && req.Currency != "USD" {  // or tenant default
		return nil, errors.New("multi-currency contracts not enabled for this tenant")
	}
	// ...
}
```

This pattern keeps the service testable — tests pass the flag value directly without needing a session.

## Settings

Settings are tenant-configurable values that affect business logic. They are accessed via typed accessors on the session:

```go
// String setting with default
currency := sess.SettingString("contracts.default_currency", "USD")

// Decimal setting with default
threshold := sess.SettingDecimal("contracts.approval_threshold", 50000.0)

// Bool setting with default
autoActivate := sess.SettingBool("contracts.auto_activate", false)

// Int setting with default
maxLinesPerContract := sess.SettingInt("contracts.max_lines", 100)
```

### Setting Naming Convention

Format: `<module>.<setting_name>`

| Setting | Type | Default | Meaning |
|---------|------|---------|---------|
| `contracts.default_currency` | string | `"USD"` | Default currency for new contracts |
| `contracts.approval_threshold` | decimal | `50000.0` | Contract value above which approval is required |
| `contracts.auto_activate` | bool | `false` | Automatically activate on start_date |
| `contracts.max_lines` | int | `100` | Maximum lines per contract |

### Passing Settings to the Service

Same pattern as feature flags — extract in the handler, pass resolved values to the service:

```go
// handler
approvalThreshold := sess.SettingDecimal("contracts.approval_threshold", 50000.0)

_, err := h.service.Approve(ctx, service.ApproveContractRequest{
	// ...
	ApprovalThreshold: approvalThreshold,
})
```

```go
// service
func (s *contractService) Approve(ctx context.Context, req ApproveContractRequest) (*domain.Contract, error) {
	// ...
	if contract.TotalValue.Amount().GreaterThan(decimal.NewFromFloat(req.ApprovalThreshold)) {
		// Check if requester has elevated approval permission
		allowed, _ := s.authzSvc.Enforce(ctx, iam.Request{
			Subject: iam.TenantSubject(req.Principal.UserID()),
			Domain:  iam.TenantDomain(req.TenantID),
			Object:  "contracts/contract/*",
			Action:  "approve_high_value",
		})
		if !allowed {
			return nil, domain.ErrContractValueExceedsLimit
		}
	}
	// ...
}
```

## Session Snapshot Freshness

Settings and flags are snapshotted at login. If a tenant admin changes a setting, existing sessions reflect the old value until they log in again. This is intentional — consistency within a session matters more than immediate propagation of admin changes.

For settings that must take effect immediately (e.g., disabling a tenant), the tenant module should invalidate all sessions via `SessionInvalidator.InvalidateAllForTenant(tenantID)`. This forces re-login, which picks up the new settings.
