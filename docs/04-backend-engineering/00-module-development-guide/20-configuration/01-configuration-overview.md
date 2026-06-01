---
title: Configuration Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Environment Variables](../../../06-devops/02-environment-variables.md)"
  - "[Startup Sequence](../21-server-startup/03-startup-sequence.md)"
  - "[Tenant Module](../../../03-platform-architecture/01-multi-tenancy/01-tenancy-model.md)"
---

# Configuration Overview

## Two Configuration Systems

| System | What it stores | Where |
|--------|---------------|-------|
| Environment variables | Infrastructure config (ports, URLs, secrets) | OS / k8s Secrets |
| Tenant configuration | Per-tenant feature flags and settings | PostgreSQL |

Never use environment variables for tenant-specific settings. Never use the DB for infrastructure config.

## Server Config (Environment)

```go
// internal/server/config.go
type Config struct {
    // Server
    HTTPPort   int    `env:"HTTP_PORT"   envDefault:"8080"`
    HealthPort int    `env:"HEALTH_PORT" envDefault:"8081"`
    MetricsPort int   `env:"METRICS_PORT" envDefault:"9090"`
    Env        string `env:"ENV"         envDefault:"production"`

    // Database
    DatabaseURL      string `env:"DATABASE_URL"      envRequired:"true"`
    DatabaseMaxConns int    `env:"DATABASE_MAX_CONNS" envDefault:"20"`
    DatabaseMinConns int    `env:"DATABASE_MIN_CONNS" envDefault:"5"`

    // Redis
    RedisURL string `env:"REDIS_URL" envRequired:"true"`

    // Session
    SessionSecret string `env:"SESSION_SECRET" envRequired:"true"`
    SessionTTL    int    `env:"SESSION_TTL"    envDefault:"86400"`

    // Temporal
    TemporalHost      string `env:"TEMPORAL_HOST"      envRequired:"true"`
    TemporalNamespace string `env:"TEMPORAL_NAMESPACE" envDefault:"default"`

    // Observability
    OTELEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
    LogLevel     string `env:"LOG_LEVEL" envDefault:"info"`
    LogFormat    string `env:"LOG_FORMAT" envDefault:"json"`
}

func loadConfig() (Config, error) {
    cfg := Config{}
    if err := env.Parse(&cfg); err != nil {
        return Config{}, fmt.Errorf("config: %w", err)
    }
    return cfg, nil
}
```

Use `github.com/caarlos0/env/v11` for parsing. Fails fast at startup if required vars are missing.

## Tenant Configuration (DB)

Per-tenant settings are stored in `tenant_configurations` table and accessed through `ResolvedSession`:

```go
// Access in service — session pre-loaded at login
currency := sess.SettingString("finance.default_currency", "USD")
approvalRequired := sess.SettingBool("contracts.approval_required", true)
```

The session contains a `Settings map[string]string` pre-computed at login from the tenant's configuration rows.

### Adding a New Config Key

1. Insert the config definition:

```sql
INSERT INTO config_definitions (key, description, default_value, data_type, module)
VALUES (
    'contracts.max_contract_value',
    'Maximum allowed contract value in tenant default currency',
    '1000000',
    'decimal',
    'contracts'
);
```

2. Use in service code:

```go
maxValue, err := decimal.NewFromString(sess.SettingString("contracts.max_contract_value", "1000000"))
if req.TotalValue.GreaterThan(maxValue) {
    return nil, &domain.BusinessError{
        Code:    "CONTRACT_VALUE_EXCEEDS_LIMIT",
        Message: fmt.Sprintf("contract value exceeds tenant limit of %s", maxValue),
        Status:  422,
    }
}
```

3. Document in the Tenant API reference.

## Feature Flags

Feature flags gate entire module features per tenant:

```go
// Check in service
if !sess.FeatureEnabled("contracts.bulk_import") {
    return nil, &domain.BusinessError{
        Code:    "FEATURE_NOT_ENABLED",
        Message: "bulk import is not enabled for this tenant",
        Status:  403,
    }
}
```

Feature flags are boolean. For fine-grained config, use tenant configuration keys.

### Adding a New Feature Flag

1. Insert the flag definition:

```sql
INSERT INTO feature_flags (name, description, default_enabled)
VALUES ('contracts.bulk_import', 'Enable CSV bulk import for contracts', false);
```

2. Enable for specific tenants via the Tenant API.

3. Gate the feature in service code with `sess.FeatureEnabled()`.

## Module Config Struct

For modules with multiple config values, group them into a module-specific config:

```go
// internal/core/contracts/config.go
type ContractConfig struct {
    MaxContractValue    decimal.Decimal
    ApprovalRequired    bool
    BulkImportEnabled   bool
    DefaultCurrency     string
}

func ContractConfigFromSession(sess iam.ResolvedSession) ContractConfig {
    maxVal, _ := decimal.NewFromString(sess.SettingString("contracts.max_contract_value", "1000000"))
    return ContractConfig{
        MaxContractValue:  maxVal,
        ApprovalRequired:  sess.SettingBool("contracts.approval_required", true),
        BulkImportEnabled: sess.FeatureEnabled("contracts.bulk_import"),
        DefaultCurrency:   sess.SettingString("finance.default_currency", "USD"),
    }
}
```

Build at the start of each service method — config may differ per tenant and is always sourced from the session.
