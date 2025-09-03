package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// ConfigKey enforces business rules for configuration keys
type ConfigKey string

// keyFormatRegex validates configuration key format
var keyFormatRegex = regexp.MustCompile(`^[a-z0-9_]+$`)

// NewConfigKey creates a new configuration key with validation
func NewConfigKey(key string) (ConfigKey, error) {
	if key == "" {
		return "", ErrKeyRequired
	}

	if !isValidKeyFormat(key) {
		return "", ErrInvalidKeyFormat
	}

	if len(key) > 100 {
		return "", ErrKeyTooLong
	}

	return ConfigKey(key), nil
}

// NewFullConfigKey creates a configuration key from module and key components
func NewFullConfigKey(module ModuleName, key string) (ConfigKey, error) {
	if module == "" {
		return "", ErrModuleRequired
	}

	if !module.Valid() {
		return "", ErrInvalidModule
	}

	configKey, err := NewConfigKey(key)
	if err != nil {
		return "", err
	}

	fullKey := fmt.Sprintf("%s.%s", module, configKey)
	return ConfigKey(fullKey), nil
}

// ParseConfigKey parses a full configuration key into module and key components
func ParseConfigKey(fullKey string) (ModuleName, ConfigKey, error) {
	parts := strings.SplitN(fullKey, ".", 2)
	if len(parts) != 2 {
		return "", "", ErrInvalidFullKeyFormat
	}

	module := ModuleName(parts[0])
	if !module.Valid() {
		return "", "", ErrInvalidModule
	}

	key, err := NewConfigKey(parts[1])
	if err != nil {
		return "", "", err
	}

	return module, key, nil
}

// Module extracts the module name from a full configuration key
func (ck ConfigKey) Module() ModuleName {
	parts := strings.SplitN(string(ck), ".", 2)
	return ModuleName(parts[0])
}

// Key extracts the key component from a full configuration key
func (ck ConfigKey) Key() string {
	parts := strings.SplitN(string(ck), ".", 2)
	if len(parts) < 2 {
		return string(ck)
	}
	return parts[1]
}

// IsValid validates the configuration key format
func (ck ConfigKey) IsValid() bool {
	keyStr := string(ck)

	// Check if it's a full key (module.key)
	if strings.Contains(keyStr, ".") {
		module, key, err := ParseConfigKey(keyStr)
		return err == nil && module.Valid() && isValidKeyFormat(string(key))
	}

	// Check if it's a simple key
	return isValidKeyFormat(keyStr) && len(keyStr) <= 100
}

// IsFullKey returns true if this is a full configuration key (module.key)
func (ck ConfigKey) IsFullKey() bool {
	return strings.Contains(string(ck), ".")
}

// ToFullKey converts a simple key to a full key with the given module
func (ck ConfigKey) ToFullKey(module ModuleName) (ConfigKey, error) {
	if ck.IsFullKey() {
		return ck, nil // Already a full key
	}

	if !module.Valid() {
		return "", ErrInvalidModule
	}

	return NewFullConfigKey(module, string(ck))
}

// Matches returns true if the key matches the given pattern
// Supports wildcards: * (any characters), ? (single character)
func (ck ConfigKey) Matches(pattern string) bool {
	// Convert wildcard pattern to regex
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "?", ".")
	regexPattern = "^" + regexPattern + "$"

	matched, err := regexp.MatchString(regexPattern, string(ck))
	return err == nil && matched
}

// String returns the string representation of the configuration key
func (ck ConfigKey) String() string {
	return string(ck)
}

// Validate performs validation on the configuration key
func (ck ConfigKey) Validate() error {
	if !ck.IsValid() {
		return ErrInvalidKeyFormat
	}
	return nil
}

// Equals compares two configuration keys for equality
func (ck ConfigKey) Equals(other ConfigKey) bool {
	return string(ck) == string(other)
}

// isValidKeyFormat validates the key format using business rules
func isValidKeyFormat(key string) bool {
	if key == "" {
		return false
	}

	if len(key) > 100 {
		return false
	}

	// Keys must be lowercase with underscores and numbers
	return keyFormatRegex.MatchString(key)
}

// Common configuration key constants
const (
	// Finance module keys
	FinanceInvoicePrefix     ConfigKey = "invoice_prefix"
	FinanceAutoApprovalLimit ConfigKey = "auto_approval_limit"
	FinanceDefaultCurrency   ConfigKey = "default_currency"
	FinanceAccountingMethod  ConfigKey = "accounting_method"
	FinanceFiscalYearStart   ConfigKey = "fiscal_year_start_month"
	FinancePaymentTerms      ConfigKey = "payment_terms"
	FinanceTaxRate           ConfigKey = "default_tax_rate"

	// HR module keys
	HROvertimeThreshold   ConfigKey = "overtime_threshold"
	HRPayFrequency        ConfigKey = "pay_frequency"
	HRVacationDaysPerYear ConfigKey = "vacation_days_per_year"
	HRProbationPeriodDays ConfigKey = "probation_period_days"

	// Inventory module keys
	InventoryValuationMethod      ConfigKey = "valuation_method"
	InventoryReorderEnabled       ConfigKey = "reorder_enabled"
	InventoryNegativeStockAllowed ConfigKey = "negative_stock_allowed"
	InventoryReorderThreshold     ConfigKey = "reorder_threshold"

	// Sales module keys
	SalesOrderPrefix       ConfigKey = "order_prefix"
	SalesQuoteValidityDays ConfigKey = "quote_validity_days"
	SalesDiscountLimit     ConfigKey = "discount_limit"
	SalesAutoConfirmOrders ConfigKey = "auto_confirm_orders"

	// System-wide keys
	SystemTimezone          ConfigKey = "timezone"
	SystemDateFormat        ConfigKey = "date_format"
	SystemLanguage          ConfigKey = "language"
	SystemMaxFileUploadSize ConfigKey = "max_file_upload_size"
)
