# Finance Module Settings Integration

**Version**: 1.0  
**Date**: September 2025  
**Status**: Implementation Complete  

---

## Overview

The Finance module has been successfully integrated with the Settings system, providing centralized configuration management for financial operations. This integration follows the principle of using constants instead of hard-coded strings throughout the system.

## Architecture Components

### 1. Settings Constants (`internal/core/finance/domain/settings_constants.go`)

A  constants file containing all Finance module configuration keys and values:

```go
// Finance Module Constants
const (
    FinanceModuleName = "finance"
)

// Configuration Keys - Document Sequences
const (
    ConfigKeyInvoicePrefix        = "invoice_prefix"
    ConfigKeyInvoiceNumberFormat  = "invoice_number_format" 
    ConfigKeyApprovalLimit        = "approval_limit"
    ConfigKeyDefaultPaymentTerms  = "default_payment_terms"
    ConfigKeyBaseCurrency         = "base_currency"
    ConfigKeyDefaultTaxRate       = "default_tax_rate"
    // ... 50+ more configuration constants
)

// Configuration Values
const (
    PaymentTermsNet30 = "NET30"
    DefaultInvoicePrefix = "INV-"
    DefaultBaseCurrency = "USD"
    // ... more value constants
)

// Business Templates
var ManufacturingFinanceConfig = map[string]any{
    ConfigKeyInvoicePrefix:        "MFG-INV-",
    ConfigKeyApprovalLimit:        "5000.00",
    ConfigKeyMultiCurrencyEnabled: true,
    // ... complete template
}
```

**Key Features:**
- **50+ Configuration Keys**:  coverage of Finance settings
- **Business Type Templates**: Pre-configured templates for Manufacturing, Service, Retail, Non-Profit
- **Validation Constants**: Min/max limits for configuration values
- **Helper Functions**: Type-safe configuration key creation and validation

### 2. Finance Configuration Service (`internal/core/finance/service/config_service.go`)

Main service providing Finance-specific configuration operations:

```go
type FinanceConfigService interface {
    // Configuration Getters
    GetInvoicePrefix(ctx context.Context, entityID *uuid.UUID) (string, error)
    GetDefaultPaymentTerms(ctx context.Context, entityID *uuid.UUID) (string, error)
    GetBaseCurrency(ctx context.Context, entityID *uuid.UUID) (string, error)
    GetApprovalLimit(ctx context.Context, entityID *uuid.UUID) (decimal.Decimal, error)
    
    // Configuration Setters
    SetInvoicePrefix(ctx context.Context, entityID *uuid.UUID, prefix string) error
    SetDefaultPaymentTerms(ctx context.Context, entityID *uuid.UUID, terms string) error
    
    // Template Operations
    ApplyManufacturingTemplate(ctx context.Context, entityID *uuid.UUID) error
    ApplyBusinessTypeConfiguration(ctx context.Context, entityID *uuid.UUID, businessType string) error
    
    // Validation
    ValidateFinanceConfiguration(ctx context.Context, key string, value interface{}) error
}
```

**Key Features:**
- **Type-Safe Configuration Access**: Strongly typed getters and setters
- **Business Logic Validation**: Configuration-specific validation rules
- **Template Application**: One-click business type configuration
- **Error Handling**: Graceful degradation with default values
- **Audit Integration**: All configuration changes are logged

### 3. Finance Settings Integration (`internal/core/finance/service/finance_settings_integration.go`)

Bridge service connecting Finance business logic with configuration:

```go
type FinanceSettingsIntegration struct {
    configService FinanceConfigService
    logger        logger.Logger
}

// Business Logic Helpers
func (f *FinanceSettingsIntegration) ShouldRequireApproval(ctx context.Context, entityID *uuid.UUID, amount decimal.Decimal) (bool, error)
func (f *FinanceSettingsIntegration) FormatDocumentNumber(ctx context.Context, entityID *uuid.UUID, docType string, sequence int) (string, error)
func (f *FinanceSettingsIntegration) IsWithinReconciliationTolerance(ctx context.Context, entityID *uuid.UUID, difference decimal.Decimal) (bool, error)

// Configuration Aggregates
func (f *FinanceSettingsIntegration) GetApprovalWorkflowConfig(ctx context.Context, entityID *uuid.UUID) (*ApprovalWorkflowConfig, error)
func (f *FinanceSettingsIntegration) GetCurrencyConfig(ctx context.Context, entityID *uuid.UUID) (*CurrencyConfig, error)
```

**Key Features:**
- **Business Logic Integration**: Configuration-driven business decisions
- **Configuration Aggregates**: Grouped related settings for easier handling
- **Helper Methods**: Common configuration-based operations
- **Entity Initialization**: Setup default configurations for new entities

## Integration Patterns

### 1. Transaction Service Integration

```go
// Example: Configuration-driven transaction approval
func (s *TransactionServiceWithSettings) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*domain.Transaction, error) {
    // Get approval configuration using constants
    approvalConfig, err := s.settingsIntegration.GetApprovalWorkflowConfig(ctx, req.EntityID)
    if err != nil {
        return nil, fmt.Errorf("failed to get approval workflow config: %w", err)
    }

    // Apply configuration-based business logic
    requiresApproval, err := s.settingsIntegration.ShouldRequireApproval(ctx, req.EntityID, req.Amount)
    if err != nil {
        return nil, fmt.Errorf("failed to determine approval requirement: %w", err)
    }

    // Generate document number based on configuration
    documentNumber, err := s.settingsIntegration.FormatDocumentNumber(ctx, req.EntityID, "journal", req.Sequence)
    if err != nil {
        return nil, fmt.Errorf("failed to format document number: %w", err)
    }

    // Create transaction with settings-driven values
    transaction := &domain.Transaction{
        TransactionNumber: documentNumber,
        Status:           s.determineInitialStatus(requiresApproval, approvalConfig),
        // ... other fields
    }
    
    return s.transactionService.CreateTransaction(ctx, transaction)
}
```

### 2. Business Type Configuration

```go
// Apply Manufacturing company configuration
err := financeConfigService.ApplyManufacturingTemplate(ctx, &entityID)

// Apply custom business type
err := financeConfigService.ApplyBusinessTypeConfiguration(ctx, &entityID, "manufacturing")
```

### 3. Configuration-Driven Validation

```go
// Validate payment terms using constants
if !domain.IsValidPaymentTerms(terms) {
    return fmt.Errorf("invalid payment terms: %s", terms)
}

// Validate approval limit within bounds
minLimit, _ := decimal.NewFromString(domain.MinApprovalLimit)
maxLimit, _ := decimal.NewFromString(domain.MaxApprovalLimit)
if amount.LessThan(minLimit) || amount.GreaterThan(maxLimit) {
    return fmt.Errorf("approval limit must be between %s and %s", 
        domain.MinApprovalLimit, domain.MaxApprovalLimit)
}
```

## Configuration Inheritance

The Finance module benefits from the Settings system's three-level inheritance:

1. **System Level**: Default Finance configurations (defined in constants)
2. **Tenant Level**: Company-wide Finance settings  
3. **Entity Level**: Department or subsidiary-specific overrides

```go
// Example: Invoice prefix resolution
// 1. Check entity-specific setting: "DEPT-INV-"
// 2. Fall back to tenant setting: "ACME-INV-"  
// 3. Fall back to system default: "INV-"
prefix, err := financeConfigService.GetInvoicePrefix(ctx, &entityID)
```

## Benefits Achieved

### 1. **No Hard-Coded Strings**
- All configuration keys and values are defined as constants
- Type-safe configuration access throughout the system
- Compile-time validation of configuration keys

### 2. **Centralized Configuration Management**
- Single source of truth for all Finance settings
- Consistent configuration access patterns
- Hierarchical configuration inheritance

### 3. **Business Logic Integration**
- Configuration drives business decisions (approval workflows, document numbering)
- Easy customization per tenant or entity
- Template-based rapid configuration deployment

### 4. **Maintainability** 
- Configuration changes don't require code deployment
- Easy to add new configuration options
- Clear separation of configuration and business logic

### 5. **Auditability**
- All configuration changes are tracked and logged
- Complete audit trail for compliance requirements
- User attribution for configuration modifications

## Usage Examples

### Basic Configuration Access

```go
// Get configuration values
currency, err := configService.GetBaseCurrency(ctx, &entityID)
taxRate, err := configService.GetDefaultTaxRate(ctx, &entityID)
approvalLimit, err := configService.GetApprovalLimit(ctx, &entityID)

// Set configuration values
err = configService.SetInvoicePrefix(ctx, &entityID, "SALES-")
err = configService.SetDefaultPaymentTerms(ctx, &entityID, domain.PaymentTermsNet30)
```

### Template Application

```go
// Apply complete configuration template
err := configService.ApplyManufacturingTemplate(ctx, &entityID)

// Apply custom business type
err := configService.ApplyBusinessTypeConfiguration(ctx, &entityID, "service")
```

### Business Logic Integration

```go
// Configuration-driven approval decision
requiresApproval, err := settingsIntegration.ShouldRequireApproval(ctx, &entityID, amount)

// Configuration-driven document numbering
docNumber, err := settingsIntegration.FormatDocumentNumber(ctx, &entityID, "invoice", 1234)
// Result: "MFG-INV-001234" (based on configuration)

// Configuration-driven tolerance check
withinTolerance, err := settingsIntegration.IsWithinReconciliationTolerance(ctx, &entityID, difference)
```

## Integration Status

✅ **Completed Components:**
- Finance settings constants with 50+ configuration keys
- Finance configuration service with full CRUD operations  
- Settings integration bridge service
- Business type configuration templates
- Configuration validation with business rules
- Example integration with Transaction service
-  documentation and usage examples

 **Ready for Use:**
The Finance module is now fully integrated with the Settings system and ready for production use. All Finance services can leverage centralized configuration management using the constant-based approach.

---

**Next Steps**: Other ERP modules (HR, Inventory, Sales) can follow the same integration pattern established by the Finance module.