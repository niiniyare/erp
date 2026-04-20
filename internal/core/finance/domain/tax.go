package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TaxType classifies the nature of a tax obligation.
type TaxType string

const (
	TaxTypeVAT     TaxType = "VAT"     // Value Added Tax
	TaxTypeGST     TaxType = "GST"     // Goods & Services Tax
	TaxTypeWHT     TaxType = "WHT"     // Withholding Tax
	TaxTypePAYE    TaxType = "PAYE"    // Pay As You Earn (payroll)
	TaxTypeNSSF    TaxType = "NSSF"    // National Social Security Fund
	TaxTypeNHIF    TaxType = "NHIF"    // National Hospital Insurance Fund (Kenya)
	TaxTypeSHIF    TaxType = "SHIF"    // Social Health Insurance Fund (Kenya, replaces NHIF 2024)
	TaxTypeSDL     TaxType = "SDL"     // Skills Development Levy
	TaxTypeNITA    TaxType = "NITA"    // National Industrial Training Authority levy
	TaxTypeExcise  TaxType = "EXCISE"  // Excise Duty
	TaxTypeCustoms TaxType = "CUSTOMS" // Import / Customs Duty
	TaxTypeCIT     TaxType = "CIT"     // Corporate Income Tax

	// Tax calculation methods
	TaxCalculationInclusive = "INCLUSIVE"
	TaxCalculationExclusive = "EXCLUSIVE"
	TaxCalculationCompound  = "COMPOUND"
)

// IsValid returns true if the TaxType is recognised.
func (t TaxType) IsValid() bool {
	switch t {
	case TaxTypeVAT, TaxTypeGST, TaxTypeWHT, TaxTypePAYE,
		TaxTypeNSSF, TaxTypeNHIF, TaxTypeSHIF, TaxTypeSDL,
		TaxTypeNITA, TaxTypeExcise, TaxTypeCustoms, TaxTypeCIT:
		return true
	default:
		return false
	}
}

// TaxCalculationMethod defines how the tax amount is derived.
type TaxCalculationMethod string

const (
	TaxCalcPercentage  TaxCalculationMethod = "PERCENTAGE"   // rate × taxable amount
	TaxCalcFixed       TaxCalculationMethod = "FIXED_AMOUNT" // fixed monetary amount per unit
	TaxCalcProgressive TaxCalculationMethod = "PROGRESSIVE"  // bracket-based (PAYE, CIT)
	TaxCalcLookup      TaxCalculationMethod = "LOOKUP_TABLE" // pre-built rate table
)

// IsValid returns true if the method is recognised.
func (m TaxCalculationMethod) IsValid() bool {
	switch m {
	case TaxCalcPercentage, TaxCalcFixed, TaxCalcProgressive, TaxCalcLookup:
		return true
	default:
		return false
	}
}

// TaxAuthorityType classifies the level of a tax authority.
type TaxAuthorityType string

const (
	TaxAuthorityFederal   TaxAuthorityType = "federal"
	TaxAuthorityState     TaxAuthorityType = "state"
	TaxAuthorityLocal     TaxAuthorityType = "local"
	TaxAuthorityMunicipal TaxAuthorityType = "municipal"
	TaxAuthorityVAT       TaxAuthorityType = "vat"
	TaxAuthorityCustoms   TaxAuthorityType = "customs"
)

// FilingFrequency describes how often returns must be filed.
type FilingFrequency string

const (
	FilingFrequencyWeekly    FilingFrequency = "weekly"
	FilingFrequencyMonthly   FilingFrequency = "monthly"
	FilingFrequencyQuarterly FilingFrequency = "quarterly"
	FilingFrequencyAnnual    FilingFrequency = "annual"
)

// TaxAuthority represents a government tax body (KRA, URA, TRA, FIRS, SARS, etc.).
type TaxAuthority struct {
	ID            uuid.UUID        `json:"id"`
	TenantID      uuid.UUID        `json:"tenant_id"`
	AuthorityCode string           `json:"authority_code"` // e.g. "KRA"
	AuthorityName string           `json:"authority_name"` // e.g. "Kenya Revenue Authority"
	AuthorityType TaxAuthorityType `json:"authority_type"`

	// Jurisdiction
	CountryCode       string `json:"country_code"`                  // ISO 3166-1 alpha-2
	StateProvinceCode string `json:"state_province_code,omitempty"` // for sub-national authorities
	JurisdictionLevel string `json:"jurisdiction_level"`            // national, state, local

	// Filing rules
	FilingFrequency FilingFrequency `json:"filing_frequency"`
	FilingDueDay    int             `json:"filing_due_day"`  // day of month the return is due
	PaymentDueDay   int             `json:"payment_due_day"` // day of month payment is due

	// Electronic filing
	SupportsEFiling bool   `json:"supports_e_filing"`
	EFilingEndpoint string `json:"e_filing_endpoint,omitempty"` // API URL if applicable

	IsActive bool `json:"is_active"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// Validate returns validation errors for the TaxAuthority.
func (a *TaxAuthority) Validate() []ValidationError {
	var errs []ValidationError
	if a.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if strings.TrimSpace(a.AuthorityCode) == "" {
		errs = append(errs, ValidationError{Field: "authority_code", Message: "authority_code is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if strings.TrimSpace(a.AuthorityName) == "" {
		errs = append(errs, ValidationError{Field: "authority_name", Message: "authority_name is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if len(a.CountryCode) != 2 {
		errs = append(errs, ValidationError{Field: "country_code", Message: "country_code must be a 2-letter ISO code", Code: "INVALID_FORMAT", Severity: ValidationSeverityError})
	}
	return errs
}

// TaxCode is a single tax rule — a specific rate applied by a specific authority
// for a specific tax type. Effective-dated rate changes are stored as new TaxCode
// records with the same Code but different EffectiveDate / ExpiryDate.
type TaxCode struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Code        string    `json:"code"` // unique within tenant, e.g. "KE-VAT-STD"
	Name        string    `json:"name"` // e.g. "Kenya VAT Standard Rate"
	Description string    `json:"description,omitempty"`

	TaxType           TaxType              `json:"tax_type"`
	TaxCategory       string               `json:"tax_category"` // standard, reduced, zero, exempt, reverse_charge
	TaxAuthorityID    uuid.UUID            `json:"tax_authority_id"`
	CalculationMethod TaxCalculationMethod `json:"calculation_method"`

	// TaxRate is stored as a decimal fraction: 0.16 = 16%.
	// For FIXED_AMOUNT method, this stores the fixed monetary amount instead.
	TaxRate decimal.Decimal `json:"tax_rate"`

	// Compound & ordering
	CompoundTax  bool `json:"compound_tax"`  // if true, rate applied on amount already including other taxes
	CascadeOrder int  `json:"cascade_order"` // calculation order when multiple taxes stack

	// Effective dates — the engine always selects the rate valid at the transaction date
	EffectiveDate time.Time  `json:"effective_date"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty"`

	// Amount thresholds (optional)
	MinimumAmount *decimal.Decimal `json:"minimum_amount,omitempty"` // no tax below this amount
	MaximumAmount *decimal.Decimal `json:"maximum_amount,omitempty"` // cap on tax amount

	// GL account mappings (nullable — set after chart of accounts is configured)
	TaxPayableAccountID    *uuid.UUID `json:"tax_payable_account_id,omitempty"`
	TaxExpenseAccountID    *uuid.UUID `json:"tax_expense_account_id,omitempty"`
	TaxReceivableAccountID *uuid.UUID `json:"tax_receivable_account_id,omitempty"`

	// Reporting
	ReportingCode    string `json:"reporting_code,omitempty"`     // official box reference on return
	ReturnLineNumber string `json:"return_line_number,omitempty"` // line number on tax form

	IsActive  bool `json:"is_active"`
	IsDefault bool `json:"is_default"` // used when no code is explicitly selected

	// Brackets (loaded on demand for PROGRESSIVE method)
	Brackets []*TaxBracket `json:"brackets,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// Validate returns validation errors for the TaxCode.
func (tc *TaxCode) Validate() []ValidationError {
	var errs []ValidationError
	if tc.TenantID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tenant_id", Message: "tenant_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if strings.TrimSpace(tc.Code) == "" {
		errs = append(errs, ValidationError{Field: "code", Message: "tax code is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if strings.TrimSpace(tc.Name) == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "tax name is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if !tc.TaxType.IsValid() {
		errs = append(errs, ValidationError{Field: "tax_type", Message: fmt.Sprintf("invalid tax type: %s", tc.TaxType), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}
	if !tc.CalculationMethod.IsValid() {
		errs = append(errs, ValidationError{Field: "calculation_method", Message: fmt.Sprintf("invalid calculation method: %s", tc.CalculationMethod), Code: "INVALID_VALUE", Severity: ValidationSeverityError})
	}
	if tc.TaxAuthorityID == uuid.Nil {
		errs = append(errs, ValidationError{Field: "tax_authority_id", Message: "tax_authority_id is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	if tc.TaxRate.IsNegative() {
		errs = append(errs, ValidationError{Field: "tax_rate", Message: "tax_rate cannot be negative", Code: "VALUE_OUT_OF_RANGE", Severity: ValidationSeverityError})
	}
	if tc.EffectiveDate.IsZero() {
		errs = append(errs, ValidationError{Field: "effective_date", Message: "effective_date is required", Code: "REQUIRED_FIELD", Severity: ValidationSeverityError})
	}
	return errs
}

// IsActiveOn returns true when this tax code is effective on the given date.
func (tc *TaxCode) IsActiveOn(date time.Time) bool {
	if date.Before(tc.EffectiveDate) {
		return false
	}
	if tc.ExpiryDate != nil && date.After(*tc.ExpiryDate) {
		return false
	}
	return tc.IsActive
}

// TaxBracket defines one band of a progressive tax (e.g. PAYE, CIT).
// Brackets belong to a TaxCode with CalculationMethod = PROGRESSIVE.
type TaxBracket struct {
	ID            uuid.UUID        `json:"id"`
	TaxCodeID     uuid.UUID        `json:"tax_code_id"`
	BracketNumber int              `json:"bracket_number"`           // 1-based, ordered ascending
	MinimumAmount decimal.Decimal  `json:"minimum_amount"`           // lower bound (inclusive)
	MaximumAmount *decimal.Decimal `json:"maximum_amount,omitempty"` // nil = no upper limit
	TaxRate       decimal.Decimal  `json:"tax_rate"`                 // rate for this bracket
	// MarginalCalculation: true = only the income within this band is taxed at this rate (marginal).
	// false = the entire income is taxed at the bracket rate (flat-rate-on-whole).
	MarginalCalculation bool      `json:"marginal_calculation"`
	CreatedAt           time.Time `json:"created_at"`
}

// TaxCalculationResult holds the output of a single tax code calculation.
type TaxCalculationResult struct {
	TaxCodeID     uuid.UUID       `json:"tax_code_id"`
	Code          string          `json:"code"`
	Name          string          `json:"name"`
	TaxType       TaxType         `json:"tax_type"`
	TaxableAmount decimal.Decimal `json:"taxable_amount"`
	TaxRate       decimal.Decimal `json:"tax_rate"`
	TaxAmount     decimal.Decimal `json:"tax_amount"`
	NetAmount     decimal.Decimal `json:"net_amount"`   // taxable_amount (for exclusive) or taxable_amount - tax (for inclusive)
	GrossAmount   decimal.Decimal `json:"gross_amount"` // taxable_amount + tax (for exclusive)
	IsInclusive   bool            `json:"is_inclusive"` // true = tax already included in amount
	IsCompound    bool            `json:"is_compound"`
	// GL accounts to use when posting
	TaxPayableAccountID    *uuid.UUID `json:"tax_payable_account_id,omitempty"`
	TaxExpenseAccountID    *uuid.UUID `json:"tax_expense_account_id,omitempty"`
	TaxReceivableAccountID *uuid.UUID `json:"tax_receivable_account_id,omitempty"`
}
