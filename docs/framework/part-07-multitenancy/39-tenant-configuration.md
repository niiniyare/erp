---
title: "Chapter 39: Tenant Configuration"
part: "Part VII — Multi-Tenancy Operations"
chapter: 39
section: "39-tenant-configuration"
related:
  - "[Chapter 38: Tenant Lifecycle](38-tenant-lifecycle.md)"
  - "[Chapter 20: Feature Flags](../part-03-api/20-feature-flags.md)"
---

# Chapter 39: Tenant Configuration

Tenants share the same application code but have different requirements — different modules enabled, different business rules, different Kenya-specific settings. This chapter covers the TenantConfig entity, module enablement, and Kenya-specific configuration.

---

## 39.1. TenantConfig Entity

### 39.1.1. Why a Separate Config Entity?

Tenant configuration is distinct from the Tenant entity itself:
- The `tenants` table lives in the platform schema (`public.tenants`) and is accessed by the tenant middleware before any tenant schema is set
- `tenant_config` lives inside the tenant's own schema and contains richer configuration read by application code after the schema is set
- Separating them avoids bloating the platform-level tenant record with dozens of operational fields

### 39.1.2. TenantConfig Structure

```go
type TenantConfig struct {
    TenantID        uuid.UUID

    // Business identity
    LegalName       string
    TradingName     *string
    KraPIN          string
    VATNumber       *string    // if VAT-registered
    NSSFNumber      *string
    NHIFNumber      *string
    PhysicalAddress string
    PostalAddress   *string
    County          string     // Kenya county

    // Locale
    Timezone        string     // "Africa/Nairobi" (EAT, UTC+3)
    DateFormat      string     // "DD/MM/YYYY" (Kenya standard)
    NumberFormat    string     // "1,234.56" (comma thousands, dot decimal)
    Currency        string     // "KES"

    // Fiscal settings
    FiscalYearEnd   string     // "06-30" (June 30 for Kenya)
    FiscalYearLabel string     // "2025-26" format

    // VAT configuration
    VATRegistered   bool
    VATRate         decimal.Decimal  // 16% standard Kenya VAT
    VATReturnPeriod string           // "monthly" | "quarterly"

    // Modules enabled
    EnabledModules  []string   // ["finance", "inventory", "hr", "crm", "forecourt"]

    // Branding
    LogoURL         *string
    PrimaryColour   *string    // hex, e.g. "#1a6b3a"
    AccentColour    *string

    // Invoice settings
    InvoicePrefix   string     // "INV"
    QuotePrefix     string     // "QT"
    POPrefix        string     // "PO"
    InvoiceFooter   *string    // legal text, bank details
    InvoiceTemplate string     // "standard" | "compact" | "detailed"

    // Notifications
    AdminEmail      string
    SMSEnabled      bool
    SMSProvider     string     // "africastalking" | "twilio"
    SMSCredentials  EncryptedCredentials
}
```

### 39.1.3. Reading TenantConfig in Application Code

The config is loaded once per request and injected into the handler context. It is cached in Redis (TTL 5 minutes) keyed by tenant ID:

```go
func (s *TenantConfigService) Get(ctx context.Context) (*TenantConfig, error) {
    tenantID := tenant.IDFromContext(ctx)
    cacheKey := fmt.Sprintf("tenant_config:%s", tenantID)

    var cfg TenantConfig
    if err := s.redis.GetJSON(ctx, cacheKey, &cfg); err == nil {
        return &cfg, nil
    }

    cfg, err := s.repo.Get(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    s.redis.SetJSON(ctx, cacheKey, cfg, 5*time.Minute)
    return &cfg, nil
}
```

Cache invalidation on update: the `after_save` hook on `TenantConfig` deletes the Redis key. The next request re-populates it.

---

## 39.2. Module Enablement

### 39.2.1. Module Registry

Modules are defined in a registry at startup:

```go
type ModuleDef struct {
    Name        string
    DisplayName string
    Plan        string   // minimum plan: "starter" | "growth" | "enterprise"
    DependsOn   []string // modules that must also be enabled
    Routes      func(r fiber.Router, deps *wire.Dependencies)
    SeedData    func(ctx context.Context, tenantID uuid.UUID) error
    Workers     []temporal.Worker
}

var ModuleRegistry = map[string]*ModuleDef{
    "finance":   {Name: "finance", Plan: "starter", ...},
    "inventory": {Name: "inventory", Plan: "starter", DependsOn: []string{"finance"}, ...},
    "hr":        {Name: "hr", Plan: "starter", DependsOn: []string{"finance"}, ...},
    "crm":       {Name: "crm", Plan: "growth", ...},
    "forecourt": {Name: "forecourt", Plan: "enterprise", DependsOn: []string{"finance", "inventory"}, ...},
    "assets":    {Name: "assets", Plan: "growth", DependsOn: []string{"finance"}, ...},
    "projects":  {Name: "projects", Plan: "growth", ...},
}
```

### 39.2.2. Route Registration by Module

At server startup, routes are registered only for modules with the `finance` module always enabled (it's the foundation). Module routes are scoped under their module prefix:

```go
func registerModuleRoutes(app *fiber.App, deps *wire.Dependencies, enabledModules []string) {
    for _, moduleName := range enabledModules {
        mod := ModuleRegistry[moduleName]
        if mod == nil {
            continue
        }
        group := app.Group("/api/v1/" + moduleName)
        mod.Routes(group, deps)
    }
}
```

Attempting to access a disabled module's endpoint returns 404 — as if the route doesn't exist. This prevents plan-upgrade prompts being bypassed by guessing API URLs.

### 39.2.3. Runtime Module Check in Handlers

For actions that span module boundaries (e.g., posting a payroll journal entry from HR into Finance), the handler checks that both modules are enabled:

```go
func requireModule(ctx context.Context, module string) error {
    cfg, _ := tenantConfigService.Get(ctx)
    for _, m := range cfg.EnabledModules {
        if m == module {
            return nil
        }
    }
    return errs.NewBusinessError("MODULE_DISABLED",
        "module %q is not enabled for this account — contact your administrator to enable it",
        module)
}
```

### 39.2.4. Enabling a Module

Enabling a new module requires running its seed data:

```go
func EnableModuleWorkflow(ctx workflow.Context, params EnableModuleParams) error {
    ao := workflow.ActivityOptions{StartToCloseTimeout: 5 * time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)

    // Check dependencies are enabled
    mod := ModuleRegistry[params.Module]
    for _, dep := range mod.DependsOn {
        workflow.ExecuteActivity(ctx, activities.AssertModuleEnabled,
            AssertInput{TenantID: params.TenantID, Module: dep})
    }

    // Run module seed data
    workflow.ExecuteActivity(ctx, activities.SeedModuleData,
        SeedInput{TenantID: params.TenantID, Module: params.Module})

    // Update enabled_modules list
    workflow.ExecuteActivity(ctx, activities.AddEnabledModule,
        ModuleUpdateInput{TenantID: params.TenantID, Module: params.Module})

    // Invalidate tenant config cache
    workflow.ExecuteActivity(ctx, activities.InvalidateTenantConfigCache, params.TenantID)

    return nil
}
```

---

## 39.3. Kenya-Specific Configuration

### 39.3.1. Why Kenya-Specific Config?

Awo is purpose-built for Kenya. Several regulatory and operational constants differ from international defaults and change annually with government announcements:

- PAYE bands change each Finance Act (typically June budget)
- NHIF contribution tables change periodically
- NSSF moved to a new contribution structure in 2024
- VAT rate is 16% (not 15% or 20% as in other countries)
- Affordable Housing Levy introduced in 2024 at 1.5%

### 39.3.2. Statutory Rate Configuration

Rather than hard-coding rates, Awo stores them as tenant-level configuration with effective dates:

```go
type PAYEBandConfig struct {
    TenantID      uuid.UUID
    EffectiveFrom time.Time
    Bands         []PAYEBand
    PersonalRelief decimal.Decimal  // KES 2,400/month (2025)
    InsuranceReliefRate decimal.Decimal // 15% of NHIF contribution
}

type NHIFConfig struct {
    TenantID      uuid.UUID
    EffectiveFrom time.Time
    Brackets      []NHIFBracket   // gross salary range → contribution amount
}

type NSSFConfig struct {
    TenantID      uuid.UUID
    EffectiveFrom time.Time
    TierILimit    decimal.Decimal  // KES 7,000 (2025) — Tier I earnings ceiling
    TierIILimit   decimal.Decimal  // KES 36,000 (2025) — Tier II earnings ceiling
    EmployeeRate  decimal.Decimal  // 6%
    EmployerRate  decimal.Decimal  // 6%
}
```

When computing payroll, the system fetches the applicable config for the payroll month (the record with `effective_from <= payroll_month` and no later record). This ensures historical payroll re-computation uses the rates that were applicable at the time.

### 39.3.3. VAT Configuration

```go
type VATConfig struct {
    TenantID         uuid.UUID
    Registered       bool
    VATNumber        *string
    Standard Rate    decimal.Decimal  // 16%
    ZeroRated        []string         // item groups that are zero-rated
    Exempt           []string         // item groups that are VAT-exempt
    // Kenya VAT return settings
    ReturnPeriod     string           // "monthly" | "quarterly"
    iTaxPassword     EncryptedString  // for automated iTax filing (optional)
}
```

VAT-registered tenants have VAT automatically calculated on sales invoices. Line items check the item group's VAT classification:
- Standard: 16% VAT applied
- Zero-rated: 0% VAT, item appears on VAT return as zero-rated supply
- Exempt: no VAT at all, item does not appear on VAT return

### 39.3.4. M-PESA Integration Configuration

M-PESA is the primary payment method for many Kenyan SMEs. Configuration for Safaricom Daraja API:

```go
type MPESAConfig struct {
    TenantID         uuid.UUID
    Enabled          bool
    Environment      string  // "sandbox" | "production"
    ConsumerKey      EncryptedString
    ConsumerSecret   EncryptedString
    BusinessShortCode string // Paybill or Till number
    PassKey          EncryptedString
    CallbackURL      string  // e.g. "https://acme.awo.so/api/v1/mpesa/callback"
    // STK Push settings
    AccountReference string  // shown on customer's phone
    TransactionDesc  string
}
```

M-PESA STK Push triggers a payment request to the customer's phone. The callback URL receives the payment confirmation. The callback handler posts a payment receipt to Finance (Debit M-PESA Bank Account, Credit Accounts Receivable).

### 39.3.5. Africa's Talking SMS Configuration

SMS notifications (leave approvals, payslip ready, invoice due) go via Africa's Talking, the dominant SMS gateway in Kenya:

```go
type SMSConfig struct {
    TenantID    uuid.UUID
    Enabled     bool
    APIKey      EncryptedString
    Username    string
    SenderID    string  // e.g. "ACME" (up to 11 chars, registered with CA Kenya)
}
```

The `SenderID` must be registered with the Communications Authority of Kenya. Unregistered sender IDs are blocked by Kenyan networks.
