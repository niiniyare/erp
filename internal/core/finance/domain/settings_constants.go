package domain

// Settings integration constants for Finance module
// This file defines all configuration keys and modules used by the Finance service
// to avoid hard-coded strings and ensure consistency across the system.

import "awo/internal/core/settings/domain"

// Finance Module Constants
const (
	// Settings module name for finance configurations
	FinanceModuleName = "Finance"
)

// Finance Configuration Keys
// These constants define all configuration keys used by the Finance module
const (
	// Document Sequence Configuration Keys
	ConfigKeyInvoicePrefix       = "invoice_prefix"
	ConfigKeyInvoiceNumberFormat = "invoice_number_format"
	ConfigKeyInvoicePadding      = "invoice_padding"
	ConfigKeyInvoiceResetFreq    = "invoice_reset_frequency"

	ConfigKeyReceiptPrefix       = "receipt_prefix"
	ConfigKeyReceiptNumberFormat = "receipt_number_format"
	ConfigKeyReceiptPadding      = "receipt_padding"

	ConfigKeyPaymentPrefix       = "payment_prefix"
	ConfigKeyPaymentNumberFormat = "payment_number_format"
	ConfigKeyPaymentPadding      = "payment_padding"

	ConfigKeyJournalPrefix       = "journal_prefix"
	ConfigKeyJournalNumberFormat = "journal_number_format"
	ConfigKeyJournalPadding      = "journal_padding"

	// Approval Workflow Configuration Keys
	ConfigKeyApprovalLimit           = "approval_limit"
	ConfigKeyAutoApprovalEnabled     = "auto_approval_enabled"
	ConfigKeyApprovalTimeoutDays     = "approval_timeout_days"
	ConfigKeyRequireManagerApproval  = "require_manager_approval"
	ConfigKeyRequireDirectorApproval = "require_director_approval"
	ConfigKeyApprovalLevels          = "approval_levels"
	ConfigKeyEscalationEnabled       = "escalation_enabled"
	ConfigKeyEscalationHours         = "escalation_hours"

	// Payment Terms Configuration Keys
	ConfigKeyDefaultPaymentTerms      = "default_payment_terms"
	ConfigKeyPaymentTermsOptions      = "payment_terms_options"
	ConfigKeyEarlyPaymentDiscountRate = "early_payment_discount_rate"
	ConfigKeyEarlyPaymentDiscountDays = "early_payment_discount_days"
	ConfigKeyLateFeeEnabled           = "late_fee_enabled"
	ConfigKeyLateFeeRate              = "late_fee_rate"
	ConfigKeyLateFeeGracePeriod       = "late_fee_grace_period"

	// Currency and Exchange Rate Configuration Keys
	ConfigKeyBaseCurrency           = "base_currency"
	ConfigKeyExchangeRateProvider   = "exchange_rate_provider"
	ConfigKeyExchangeRateUpdateFreq = "exchange_rate_update_frequency"
	ConfigKeyRoundingMethod         = "rounding_method"
	ConfigKeyRoundingPrecision      = "rounding_precision"
	ConfigKeyMultiCurrencyEnabled   = "multi_currency_enabled"
	ConfigKeyExchangeRateVariance   = "exchange_rate_variance_threshold"

	// Tax Configuration Keys
	ConfigKeyDefaultTaxRate        = "default_tax_rate"
	ConfigKeyTaxCalculationMethod  = "tax_calculation_method"
	ConfigKeyTaxInclusivePricing   = "tax_inclusive_pricing"
	ConfigKeyWithholdingTaxEnabled = "withholding_tax_enabled"
	ConfigKeyWithholdingTaxRate    = "withholding_tax_rate"
	ConfigKeyTaxReportingFrequency = "tax_reporting_frequency"
	ConfigKeyVATRegistrationNumber = "vat_registration_number"

	// Account Configuration Keys
	ConfigKeyDefaultExpenseAccount  = "default_expense_account"
	ConfigKeyDefaultRevenueAccount  = "default_revenue_account"
	ConfigKeyDefaultTaxAccount      = "default_tax_account"
	ConfigKeyDefaultBankAccount     = "default_bank_account"
	ConfigKeyDefaultCustomerAccount = "default_customer_account"
	ConfigKeyDefaultVendorAccount   = "default_vendor_account"
	ConfigKeyAccountCodeLength      = "account_code_length"
	ConfigKeyAccountNumberingScheme = "account_numbering_scheme"

	// Financial Period Configuration Keys
	ConfigKeyFiscalYearStart      = "fiscal_year_start"
	ConfigKeyFiscalYearEnd        = "fiscal_year_end"
	ConfigKeyPeriodClosingEnabled = "period_closing_enabled"
	ConfigKeyAutoClosePeriods     = "auto_close_periods"
	ConfigKeyClosingReminderDays  = "closing_reminder_days"

	// Reconciliation Configuration Keys
	ConfigKeyAutoReconciliationEnabled = "auto_reconciliation_enabled"
	ConfigKeyReconciliationTolerance   = "reconciliation_tolerance"
	ConfigKeyBankReconciliationEnabled = "bank_reconciliation_enabled"
	ConfigKeyReconciliationFrequency   = "reconciliation_frequency"

	// Reporting Configuration Keys
	ConfigKeyDefaultReportCurrency = "default_report_currency"
	ConfigKeyConsolidatedReporting = "consolidated_reporting"
	ConfigKeyReportingDecimals     = "reporting_decimals"
	ConfigKeyReportDateFormat      = "report_date_format"
	ConfigKeyReportLogo            = "report_logo"
	ConfigKeyReportFooter          = "report_footer"

	// Integration Configuration Keys
	ConfigKeyBankFeedEnabled        = "bank_feed_enabled"
	ConfigKeyBankFeedProvider       = "bank_feed_provider"
	ConfigKeyPaymentGatewayEnabled  = "payment_gateway_enabled"
	ConfigKeyPaymentGatewayProvider = "payment_gateway_provider"
	ConfigKeyInventoryIntegration   = "inventory_integration_enabled"
	ConfigKeyHRIntegration          = "hr_integration_enabled"

	// Budget and Forecasting Configuration Keys
	ConfigKeyBudgetingEnabled        = "budgeting_enabled"
	ConfigKeyBudgetVarianceThreshold = "budget_variance_threshold"
	ConfigKeyForecastingEnabled      = "forecasting_enabled"
	ConfigKeyBudgetApprovalRequired  = "budget_approval_required"

	// Audit and Compliance Configuration Keys
	ConfigKeyAuditTrailEnabled    = "audit_trail_enabled"
	ConfigKeyAuditRetentionPeriod = "audit_retention_period"
	ConfigKeyComplianceMode       = "compliance_mode"
	ConfigKeyDataRetentionYears   = "data_retention_years"
	ConfigKeyBackupFrequency      = "backup_frequency"
)

// Finance Configuration Value Constants
// These constants define commonly used configuration values
const (
	// Payment Terms Values
	PaymentTermsNet15 = "NET15"
	PaymentTermsNet30 = "NET30"
	PaymentTermsNet45 = "NET45"
	PaymentTermsNet60 = "NET60"
	PaymentTermsCOD   = "COD"
	PaymentTermsDue   = "DUE_ON_RECEIPT"

	// Rounding Methods
	RoundingMethodStandard = "STANDARD"
	RoundingMethodUp       = "UP"
	RoundingMethodDown     = "DOWN"
	RoundingMethodBankers  = "BANKERS"

	// Tax Calculation Methods - using existing constants from constant.go

	// Exchange Rate Providers
	ExchangeRateProviderManual = "MANUAL"
	ExchangeRateProviderECB    = "ECB"
	ExchangeRateProviderXE     = "XE"
	ExchangeRateProviderOpenER = "OPEN_EXCHANGE_RATES"

	// Numbering Reset Frequencies
	ResetFrequencyNever     = "NEVER"
	ResetFrequencyYearly    = "YEARLY"
	ResetFrequencyMonthly   = "MONTHLY"
	ResetFrequencyQuarterly = "QUARTERLY"

	// Bank Feed Providers
	BankFeedProviderYodlee = "YODLEE"
	BankFeedProviderPlaid  = "PLAID"
	BankFeedProviderManual = "MANUAL"

	// Compliance Modes
	ComplianceModeStandard = "STANDARD"
	ComplianceModeSOX      = "SOX"
	ComplianceModeIFRS     = "IFRS"
	ComplianceModeGAAP     = "GAAP"
)

// Helper functions to create Settings domain objects from Finance constants

// NewFinanceModuleName returns the finance module name as a Settings ModuleName
func NewFinanceModuleName() domain.ModuleName {
	return domain.ModuleName(FinanceModuleName)
}

// NewFinanceConfigKey creates a Settings ConfigKey from a finance configuration key constant
func NewFinanceConfigKey(key string) domain.ConfigKey {
	configKey, _ := domain.NewConfigKey(key)
	return configKey
}

// Finance Configuration Key Helpers
// These functions return Settings domain objects for commonly used configurations

func ConfigKeyApprovalLimitKey() domain.ConfigKey {
	return NewFinanceConfigKey(ConfigKeyApprovalLimit)
}

func ConfigKeyInvoicePrefixKey() domain.ConfigKey {
	return NewFinanceConfigKey(ConfigKeyInvoicePrefix)
}

func ConfigKeyDefaultPaymentTermsKey() domain.ConfigKey {
	return NewFinanceConfigKey(ConfigKeyDefaultPaymentTerms)
}

func ConfigKeyBaseCurrencyKey() domain.ConfigKey {
	return NewFinanceConfigKey(ConfigKeyBaseCurrency)
}

func ConfigKeyAutoApprovalEnabledKey() domain.ConfigKey {
	return NewFinanceConfigKey(ConfigKeyAutoApprovalEnabled)
}

func ConfigKeyFiscalYearStartKey() domain.ConfigKey {
	return NewFinanceConfigKey(ConfigKeyFiscalYearStart)
}

func ConfigKeyDefaultTaxRateKey() domain.ConfigKey {
	return NewFinanceConfigKey(ConfigKeyDefaultTaxRate)
}

func ConfigKeyReconciliationToleranceKey() domain.ConfigKey {
	return NewFinanceConfigKey(ConfigKeyReconciliationTolerance)
}

func ConfigKeyMultiCurrencyEnabledKey() domain.ConfigKey {
	return NewFinanceConfigKey(ConfigKeyMultiCurrencyEnabled)
}

func ConfigKeyBudgetVarianceThresholdKey() domain.ConfigKey {
	return NewFinanceConfigKey(ConfigKeyBudgetVarianceThreshold)
}

// Default Configuration Values
// These constants define default values for finance configurations

const (
	// Default Values
	DefaultInvoicePrefix = "INV-"
	DefaultReceiptPrefix = "RCP-"
	DefaultPaymentPrefix = "PAY-"
	DefaultJournalPrefix = "JNL-"
	DefaultNumberPadding = 6
	DefaultApprovalLimit = "1000.00"
	DefaultPaymentTerms  = PaymentTermsNet30
	DefaultTaxRate       = "16.00" // Kenya VAT rate
	// DefaultBaseCurrency defined in constant.go
	DefaultRoundingPrecision = 2
	// DefaultReconciliationTolerance defined in constant.go
	// DefaultBudgetVarianceThreshold defined in constant.go
	DefaultApprovalTimeoutDays = 7
	DefaultEscalationHours     = 24
	// DefaultFiscalYearStart defined in constant.go
	DefaultAuditRetentionPeriod = 2555 // 7 years in days
	DefaultDataRetentionYears   = 7
)

// Configuration Templates
// These maps define complete configuration sets for different business types

var (
	// Manufacturing Company Configuration Template
	ManufacturingFinanceConfig = map[string]any{
		ConfigKeyInvoicePrefix:        "MFG-INV-",
		ConfigKeyApprovalLimit:        "5000.00",
		ConfigKeyDefaultPaymentTerms:  PaymentTermsNet30,
		ConfigKeyMultiCurrencyEnabled: true,
		ConfigKeyBudgetingEnabled:     true,
		ConfigKeyInventoryIntegration: true,
	}

	// Service Company Configuration Template
	ServiceFinanceConfig = map[string]any{
		ConfigKeyInvoicePrefix:        "SRV-",
		ConfigKeyApprovalLimit:        "2000.00",
		ConfigKeyDefaultPaymentTerms:  PaymentTermsNet15,
		ConfigKeyMultiCurrencyEnabled: false,
		ConfigKeyBudgetingEnabled:     true,
		ConfigKeyInventoryIntegration: false,
	}

	// Retail Company Configuration Template
	RetailFinanceConfig = map[string]any{
		ConfigKeyInvoicePrefix:        "RTL-",
		ConfigKeyApprovalLimit:        "1000.00",
		ConfigKeyDefaultPaymentTerms:  PaymentTermsDue,
		ConfigKeyMultiCurrencyEnabled: false,
		ConfigKeyBudgetingEnabled:     false,
		ConfigKeyInventoryIntegration: true,
	}

	// Non-Profit Organization Configuration Template
	NonProfitFinanceConfig = map[string]any{
		ConfigKeyInvoicePrefix:        "NPO-",
		ConfigKeyApprovalLimit:        "500.00",
		ConfigKeyDefaultPaymentTerms:  PaymentTermsNet45,
		ConfigKeyMultiCurrencyEnabled: false,
		ConfigKeyBudgetingEnabled:     true,
		ConfigKeyComplianceMode:       ComplianceModeStandard,
	}
)

// Validation Constants
// These constants are used for configuration validation

const (
	MinApprovalLimit           = "0.00"
	MaxApprovalLimit           = "1000000.00"
	MinRoundingPrecision       = 0
	MaxRoundingPrecision       = 6
	MinNumberPadding           = 3
	MaxNumberPadding           = 12
	MinApprovalTimeoutDays     = 1
	MaxApprovalTimeoutDays     = 90
	MinDataRetentionYears      = 1
	MaxDataRetentionYears      = 50
	MinBudgetVarianceThreshold = "0.01"
	MaxBudgetVarianceThreshold = "100.00"
)

// IsValidPaymentTerms checks if a payment terms value is valid
func IsValidPaymentTerms(terms string) bool {
	validTerms := []string{
		PaymentTermsNet15,
		PaymentTermsNet30,
		PaymentTermsNet45,
		PaymentTermsNet60,
		PaymentTermsCOD,
		PaymentTermsDue,
	}

	for _, validTerm := range validTerms {
		if terms == validTerm {
			return true
		}
	}
	return false
}

// IsValidRoundingMethod checks if a rounding method is valid
func IsValidRoundingMethod(method string) bool {
	validMethods := []string{
		RoundingMethodStandard,
		RoundingMethodUp,
		RoundingMethodDown,
		RoundingMethodBankers,
	}

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

// IsValidTaxCalculationMethod checks if a tax calculation method is valid
func IsValidTaxCalculationMethod(method string) bool {
	validMethods := []string{
		TaxCalculationExclusive,
		TaxCalculationInclusive,
		TaxCalculationCompound,
	}

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

// GetConfigurationTemplate returns a configuration template by name
func GetConfigurationTemplate(templateName string) (map[string]any, bool) {
	switch templateName {
	case "manufacturing":
		return ManufacturingFinanceConfig, true
	case "service":
		return ServiceFinanceConfig, true
	case "retail":
		return RetailFinanceConfig, true
	case "non_profit":
		return NonProfitFinanceConfig, true
	default:
		return nil, false
	}
}
