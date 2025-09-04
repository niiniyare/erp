package service

import (
	"fmt"

	"github.com/niiniyare/erp/internal/core/finance/domain"
)

// SettingsHelper provides helper functions for Finance module integration with Settings
// This is a simple helper that other Finance services can use to access configuration constants
type SettingsHelper struct{}

// NewSettingsHelper creates a new settings helper
func NewSettingsHelper() *SettingsHelper {
	return &SettingsHelper{}
}

// GetFinanceModuleName returns the finance module name constant
func (h *SettingsHelper) GetFinanceModuleName() string {
	return domain.FinanceModuleName
}

// Configuration Key Constants - these can be used by other services
// to integrate with the Settings system when it's ready

// Document Numbering Configuration Keys
func (h *SettingsHelper) GetInvoicePrefixKey() string {
	return domain.ConfigKeyInvoicePrefix
}

func (h *SettingsHelper) GetReceiptPrefixKey() string {
	return domain.ConfigKeyReceiptPrefix
}

func (h *SettingsHelper) GetPaymentPrefixKey() string {
	return domain.ConfigKeyPaymentPrefix
}

// Approval Workflow Configuration Keys
func (h *SettingsHelper) GetApprovalLimitKey() string {
	return domain.ConfigKeyApprovalLimit
}

func (h *SettingsHelper) GetAutoApprovalEnabledKey() string {
	return domain.ConfigKeyAutoApprovalEnabled
}

// Currency Configuration Keys
func (h *SettingsHelper) GetBaseCurrencyKey() string {
	return domain.ConfigKeyBaseCurrency
}

func (h *SettingsHelper) GetMultiCurrencyEnabledKey() string {
	return domain.ConfigKeyMultiCurrencyEnabled
}

// Tax Configuration Keys
func (h *SettingsHelper) GetDefaultTaxRateKey() string {
	return domain.ConfigKeyDefaultTaxRate
}

func (h *SettingsHelper) GetTaxCalculationMethodKey() string {
	return domain.ConfigKeyTaxCalculationMethod
}

// Payment Configuration Keys
func (h *SettingsHelper) GetDefaultPaymentTermsKey() string {
	return domain.ConfigKeyDefaultPaymentTerms
}

// Default Values - these can be used as fallbacks when Settings are not available
func (h *SettingsHelper) GetDefaultInvoicePrefix() string {
	return domain.DefaultInvoicePrefix
}

func (h *SettingsHelper) GetDefaultReceiptPrefix() string {
	return domain.DefaultReceiptPrefix
}

func (h *SettingsHelper) GetDefaultPaymentPrefix() string {
	return domain.DefaultPaymentPrefix
}

func (h *SettingsHelper) GetDefaultPaymentTerms() string {
	return domain.DefaultPaymentTerms
}

func (h *SettingsHelper) GetDefaultTaxRate() string {
	return domain.DefaultTaxRate
}

func (h *SettingsHelper) GetDefaultApprovalLimit() string {
	return domain.DefaultApprovalLimit
}

func (h *SettingsHelper) GetDefaultBaseCurrency() string {
	return domain.DefaultBaseCurrency
}

// Configuration Templates - can be used for entity initialization
func (h *SettingsHelper) GetManufacturingTemplate() map[string]any {
	template, _ := domain.GetConfigurationTemplate("manufacturing")
	return template
}

func (h *SettingsHelper) GetServiceTemplate() map[string]any {
	template, _ := domain.GetConfigurationTemplate("service")
	return template
}

func (h *SettingsHelper) GetRetailTemplate() map[string]any {
	template, _ := domain.GetConfigurationTemplate("retail")
	return template
}

func (h *SettingsHelper) GetNonProfitTemplate() map[string]any {
	template, _ := domain.GetConfigurationTemplate("non_profit")
	return template
}

// Validation helpers
func (h *SettingsHelper) IsValidPaymentTerms(terms string) bool {
	return domain.IsValidPaymentTerms(terms)
}

func (h *SettingsHelper) IsValidTaxCalculationMethod(method string) bool {
	validMethods := []string{
		domain.TaxCalculationExclusive,
		domain.TaxCalculationInclusive,
		domain.TaxCalculationCompound,
	}

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

func (h *SettingsHelper) IsValidRoundingMethod(method string) bool {
	return domain.IsValidRoundingMethod(method)
}

// Business Logic Helpers
// These demonstrate how Finance services can use configuration values

// ShouldRequireApprovalBasedOnAmount is an example of configuration-driven business logic
// This would be used by Transaction services to determine approval requirements
func (h *SettingsHelper) ShouldRequireApprovalBasedOnAmount(amount, approvalLimit string) bool {
	// In a real implementation, this would parse the decimal values and compare
	// For now, this is just an example showing the pattern
	return amount > approvalLimit // Simplified comparison
}

// FormatDocumentNumber shows how document numbering could use configuration
func (h *SettingsHelper) FormatDocumentNumber(prefix string, sequence int, padding int) string {
	if padding < 3 {
		padding = domain.DefaultNumberPadding
	}

	format := fmt.Sprintf("%%s%%0%dd", padding)
	return fmt.Sprintf(format, prefix, sequence)
}

// Example usage patterns for other services:
//
// // Get configuration keys for Settings integration
// helper := NewSettingsHelper()
// invoicePrefixKey := helper.GetInvoicePrefixKey()
// approvalLimitKey := helper.GetApprovalLimitKey()
//
// // Use default values as fallbacks
// defaultPrefix := helper.GetDefaultInvoicePrefix()
// defaultTerms := helper.GetDefaultPaymentTerms()
//
// // Apply business type configuration
// manufacturingConfig := helper.GetManufacturingTemplate()
//
// // Validate configuration values
// if !helper.IsValidPaymentTerms("NET30") {
//     return errors.New("invalid payment terms")
// }
//
// // Use configuration in business logic
// requiresApproval := helper.ShouldRequireApprovalBasedOnAmount(
//     transaction.Amount.String(),
//     configuredApprovalLimit,
// )
