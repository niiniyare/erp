package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RateType classifies how an exchange rate was obtained.
type RateType string

const (
	RateTypeSpot       RateType = "SPOT"       // Real-time market rate
	RateTypeClosing    RateType = "CLOSING"    // End-of-day bank rate
	RateTypeAverage    RateType = "AVERAGE"    // Period-average rate (for P&L revaluation)
	RateTypeHistorical RateType = "HISTORICAL" // Rate frozen at transaction date
	RateTypeFixed      RateType = "FIXED"      // Contractually fixed rate (e.g. hedging)
	RateTypeOfficial   RateType = "OFFICIAL"   // Central-bank official/published rate
)

// IsValid returns true if the RateType is a recognised value.
func (rt RateType) IsValid() bool {
	switch rt {
	case RateTypeSpot, RateTypeClosing, RateTypeAverage,
		RateTypeHistorical, RateTypeFixed, RateTypeOfficial:
		return true
	default:
		return false
	}
}

// String returns the string representation of RateType.
func (rt RateType) String() string { return string(rt) }

// Currency represents an ISO 4217 currency supported by the tenant.
type Currency struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	Code          string `json:"code"`           // ISO 4217 three-letter code, e.g. "USD"
	Name          string `json:"name"`           // Full name, e.g. "US Dollar"
	Symbol        string `json:"symbol"`         // Display symbol, e.g. "$"
	DecimalPlaces int    `json:"decimal_places"` // Number of minor units (0–4)
	IsActive      bool   `json:"is_active"`
	IsBase        bool   `json:"is_base"` // True for the tenant's functional/base currency

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// Validate returns ValidationErrors for the Currency.
func (c *Currency) Validate() []ValidationError {
	var errs []ValidationError

	if c.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if len(c.Code) != 3 {
		errs = append(errs, ValidationError{Field: "code", Message: "currency code must be exactly 3 characters (ISO 4217)", Code: "INVALID_FORMAT", Severity: ValidationSeverityError})
	}
	if strings.TrimSpace(c.Name) == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "currency name is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if c.DecimalPlaces < 0 || c.DecimalPlaces > 4 {
		errs = append(errs, ValidationError{Field: "decimal_places", Message: "decimal_places must be between 0 and 4", Code: "VALUE_OUT_OF_RANGE", Severity: ValidationSeverityError})
	}

	return errs
}

// currencyMinorUnits maps ISO 4217 codes to their number of decimal places.
// Currencies absent from this map default to 2 (the most common minor-unit exponent).
//
// Sources: ISO 4217 amendment 173 (2024-10).
var currencyMinorUnits = map[string]int{
	// Zero decimal places
	"BIF": 0, "CLP": 0, "DJF": 0, "GNF": 0, "IDR": 0,
	"ISK": 0, "JPY": 0, "KMF": 0, "KRW": 0, "PYG": 0,
	"RWF": 0, "UGX": 0, "VND": 0, "VUV": 0, "XAF": 0,
	"XOF": 0, "XPF": 0,
	// Three decimal places (Middle-Eastern high-precision currencies)
	"BHD": 3, "IQD": 3, "JOD": 3, "KWD": 3, "LYD": 3,
	"OMR": 3, "TND": 3,
	// Four decimal places
	"CLF": 4, "UYW": 4,
}

// CurrencyMinorUnits returns the number of decimal places (minor-unit exponent)
// for the given ISO 4217 currency code. Defaults to 2 for unknown codes.
func CurrencyMinorUnits(code string) int {
	if dp, ok := currencyMinorUnits[strings.ToUpper(code)]; ok {
		return dp
	}
	return 2 // most common: USD, EUR, GBP, KES, …
}

// CurrencyConversionTolerance returns the maximum acceptable rounding difference
// when converting amounts between currencies. Equals half a minor unit of the
// given currency — e.g. 0.005 for USD (2 dp), 0.0005 for KWD (3 dp), 0.5 for JPY (0 dp).
func CurrencyConversionTolerance(code string) decimal.Decimal {
	dp := CurrencyMinorUnits(code)
	// tolerance = 0.5 × 10^(-dp)
	divisor := decimal.New(1, int32(dp)) // 10^dp
	return decimal.NewFromFloat(0.5).Div(divisor)
}

// ExchangeRate records the rate between two currencies at a point in time.
type ExchangeRate struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	FromCurrency  string          `json:"from_currency"` // ISO 4217 source currency code
	ToCurrency    string          `json:"to_currency"`   // ISO 4217 target currency code
	Rate          decimal.Decimal `json:"rate"`          // How many units of ToCurrency per 1 FromCurrency
	RateType      RateType        `json:"rate_type"`
	EffectiveDate time.Time       `json:"effective_date"` // Date from which this rate is valid
	ExpiryDate    *time.Time      `json:"expiry_date,omitempty"`

	Source    string    `json:"source"` // e.g. "ECB", "manual", "open-exchange-rates"
	CreatedAt time.Time `json:"created_at"`
	CreatedBy uuid.UUID `json:"created_by"`
}

// Convert converts amount from FromCurrency to ToCurrency using this rate.
func (er *ExchangeRate) Convert(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(er.Rate)
}

// IsActiveOn returns true when this rate is in effect on the given date.
func (er *ExchangeRate) IsActiveOn(date time.Time) bool {
	d := date.Truncate(24 * time.Hour)
	start := er.EffectiveDate.Truncate(24 * time.Hour)
	if d.Before(start) {
		return false
	}
	if er.ExpiryDate != nil {
		expiry := er.ExpiryDate.Truncate(24 * time.Hour)
		if d.After(expiry) {
			return false
		}
	}
	return true
}

// Validate returns ValidationErrors for the ExchangeRate.
func (er *ExchangeRate) Validate() []ValidationError {
	var errs []ValidationError

	if er.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if len(er.FromCurrency) != 3 {
		errs = append(errs, ValidationError{Field: "from_currency", Message: "from_currency must be a valid ISO 4217 code", Code: "INVALID_FORMAT", Severity: ValidationSeverityError})
	}
	if len(er.ToCurrency) != 3 {
		errs = append(errs, ValidationError{Field: "to_currency", Message: "to_currency must be a valid ISO 4217 code", Code: "INVALID_FORMAT", Severity: ValidationSeverityError})
	}
	if er.FromCurrency == er.ToCurrency {
		errs = append(errs, ValidationError{Field: "to_currency", Message: "from_currency and to_currency must differ", Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}
	if er.Rate.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, ValidationError{Field: "rate", Message: "exchange rate must be greater than zero", Code: "VALUE_OUT_OF_RANGE", Severity: ValidationSeverityError})
	}
	if !er.RateType.IsValid() {
		errs = append(errs, ValidationError{Field: "rate_type", Message: fmt.Sprintf("invalid rate type: %s", er.RateType), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}
	if er.EffectiveDate.IsZero() {
		errs = append(errs, ValidationError{Field: "effective_date", Message: "effective_date is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if er.ExpiryDate != nil && !er.ExpiryDate.After(er.EffectiveDate) {
		errs = append(errs, ValidationError{Field: "expiry_date", Message: "expiry_date must be after effective_date", Code: "INVALID_DATE_RANGE", Severity: ValidationSeverityError})
	}

	return errs
}
