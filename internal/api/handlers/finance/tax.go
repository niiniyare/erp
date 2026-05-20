package finance

// tax.go — HTTP handlers for tax authority and tax code management.
//
// Routes (registered in routes.go):
//
//	POST /api/v1/finance/tax/authorities          → CreateTaxAuthority
//	GET  /api/v1/finance/tax/authorities          → ListTaxAuthorities
//	GET  /api/v1/finance/tax/authorities/:id      → GetTaxAuthority
//	PUT  /api/v1/finance/tax/authorities/:id      → UpdateTaxAuthority
//
//	POST /api/v1/finance/tax/codes                → CreateTaxCode
//	GET  /api/v1/finance/tax/codes                → ListTaxCodes
//	GET  /api/v1/finance/tax/codes/:id            → GetTaxCode
//	PUT  /api/v1/finance/tax/codes/:id            → UpdateTaxCode
//
// A TaxAuthority (e.g. KRA) owns one or more TaxCodes (e.g. VAT 16%).
// Tax brackets are embedded inside TaxCode for progressive rate structures.

import (
	"time"

	"github.com/gofiber/fiber/v2"

	financeDomain "awo.so/internal/core/finance/domain"
)

// ============================================================================
// Tax authority request type
// ============================================================================

// taxAuthorityRequest is the shared body for POST and PUT tax authority endpoints.
type taxAuthorityRequest struct {
	AuthorityCode     string `json:"authority_code"      validate:"required,max=20"`
	AuthorityName     string `json:"authority_name"      validate:"required,max=200"`
	AuthorityType     string `json:"authority_type"      validate:"required"`
	CountryCode       string `json:"country_code"        validate:"required,len=2"`
	StateProvinceCode string `json:"state_province_code"`
	JurisdictionLevel string `json:"jurisdiction_level"`
	FilingFrequency   string `json:"filing_frequency"`
	FilingDueDay      int    `json:"filing_due_day"`
	PaymentDueDay     int    `json:"payment_due_day"`
	SupportsEFiling   bool   `json:"supports_e_filing"`
	EFilingEndpoint   string `json:"e_filing_endpoint"`
	IsActive          bool   `json:"is_active"`
}

// withDefaults fills optional fields that have business-defined fallbacks.
func (r *taxAuthorityRequest) withDefaults() {
	if r.JurisdictionLevel == "" {
		r.JurisdictionLevel = "national"
	}
	if r.FilingFrequency == "" {
		r.FilingFrequency = "monthly"
	}
	if r.FilingDueDay == 0 {
		r.FilingDueDay = 20
	}
	if r.PaymentDueDay == 0 {
		r.PaymentDueDay = 20
	}
}

// ============================================================================
// Tax code request types
// ============================================================================

// taxBracketRequest represents one bracket in a progressive tax structure.
// Amounts are strings to avoid float64 precision loss.
type taxBracketRequest struct {
	BracketNumber       int     `json:"bracket_number" validate:"required,min=1"`
	MinimumAmount       string  `json:"minimum_amount" validate:"required,decimal"`
	MaximumAmount       *string `json:"maximum_amount" validate:"omitempty,decimal"`
	TaxRate             string  `json:"tax_rate"       validate:"required,decimal"`
	MarginalCalculation bool    `json:"marginal_calculation"`
}

// createTaxCodeRequest is the body for POST /tax/codes.
type createTaxCodeRequest struct {
	Code              string `json:"code"               validate:"required,max=20"`
	Name              string `json:"name"               validate:"required,max=200"`
	Description       string `json:"description"`
	TaxType           string `json:"tax_type"           validate:"required"`
	TaxCategory       string `json:"tax_category"`
	TaxAuthorityID    string `json:"tax_authority_id"   validate:"required,uuid"`
	CalculationMethod string `json:"calculation_method" validate:"required"`
	// TaxRate is the flat rate as a decimal string, e.g. "16.00" for 16%.
	// Ignored when brackets are provided.
	TaxRate          string              `json:"tax_rate"       validate:"required,decimal"`
	CompoundTax      bool                `json:"compound_tax"`
	CascadeOrder     int                 `json:"cascade_order"`
	EffectiveDate    string              `json:"effective_date" validate:"required"`
	ExpiryDate       *string             `json:"expiry_date"`
	ReportingCode    string              `json:"reporting_code"`
	ReturnLineNumber string              `json:"return_line_number"`
	IsDefault        bool                `json:"is_default"`
	Brackets         []taxBracketRequest `json:"brackets"       validate:"dive"`
}

// updateTaxCodeRequest is the body for PUT /tax/codes/:id.
// Does not include code, tax_type, or tax_authority_id — those are immutable.
type updateTaxCodeRequest struct {
	Name              string              `json:"name"               validate:"required,max=200"`
	Description       string              `json:"description"`
	TaxCategory       string              `json:"tax_category"`
	CalculationMethod string              `json:"calculation_method" validate:"required"`
	TaxRate           string              `json:"tax_rate"           validate:"required,decimal"`
	CompoundTax       bool                `json:"compound_tax"`
	CascadeOrder      int                 `json:"cascade_order"`
	EffectiveDate     string              `json:"effective_date"     validate:"required"`
	ExpiryDate        *string             `json:"expiry_date"`
	ReportingCode     string              `json:"reporting_code"`
	ReturnLineNumber  string              `json:"return_line_number"`
	IsActive          bool                `json:"is_active"`
	IsDefault         bool                `json:"is_default"`
	Brackets          []taxBracketRequest `json:"brackets"           validate:"dive"`
}

// ============================================================================
// Tax authority handlers
// ============================================================================

// CreateTaxAuthority registers a new tax authority (e.g. KRA, VAT department).
//
// POST /api/v1/finance/tax/authorities
// Permission: finance.tax.authorities.create
func (h *FinanceHandler) CreateTaxAuthority(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CreateTaxAuthority")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	var req taxAuthorityRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}
	req.withDefaults()

	authority := &financeDomain.TaxAuthority{
		AuthorityCode:     req.AuthorityCode,
		AuthorityName:     req.AuthorityName,
		AuthorityType:     financeDomain.TaxAuthorityType(req.AuthorityType),
		CountryCode:       req.CountryCode,
		StateProvinceCode: req.StateProvinceCode,
		JurisdictionLevel: req.JurisdictionLevel,
		FilingFrequency:   financeDomain.FilingFrequency(req.FilingFrequency),
		FilingDueDay:      req.FilingDueDay,
		PaymentDueDay:     req.PaymentDueDay,
		SupportsEFiling:   req.SupportsEFiling,
		EFilingEndpoint:   req.EFilingEndpoint,
	}

	// 3. Delegate
	created, err := h.services.Tax.CreateAuthority(ctx, authority)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok201(c, created)
}

// GetTaxAuthority retrieves a single tax authority by ID.
//
// GET /api/v1/finance/tax/authorities/:id
// Permission: finance.tax.authorities.read
func (h *FinanceHandler) GetTaxAuthority(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetTaxAuthority")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	authID, err := parseUUID(c.Params("id"), "tax authority ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	authority, err := h.services.Tax.GetAuthority(ctx, authID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, authority)
}

// ListTaxAuthorities returns all tax authorities for the current tenant.
//
// GET /api/v1/finance/tax/authorities?active_only=true
// Permission: finance.tax.authorities.read
func (h *FinanceHandler) ListTaxAuthorities(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListTaxAuthorities")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse
	activeOnly := c.QueryBool("active_only", false)

	// 3. Delegate
	list, err := h.services.Tax.ListAuthorities(ctx, activeOnly)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, list)
}

// UpdateTaxAuthority replaces a tax authority's mutable fields.
// The authority code is immutable after creation.
//
// PUT /api/v1/finance/tax/authorities/:id
// Permission: finance.tax.authorities.update
func (h *FinanceHandler) UpdateTaxAuthority(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.UpdateTaxAuthority")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	authID, err := parseUUID(c.Params("id"), "tax authority ID")
	if err != nil {
		return h.fail(c, err)
	}

	var req taxAuthorityRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}
	req.withDefaults()

	// extractUserID is safe: auth middleware stores user_id as string
	byUserID := extractUserID(c)

	authority := &financeDomain.TaxAuthority{
		ID:                authID,
		AuthorityName:     req.AuthorityName,
		AuthorityType:     financeDomain.TaxAuthorityType(req.AuthorityType),
		CountryCode:       req.CountryCode,
		StateProvinceCode: req.StateProvinceCode,
		JurisdictionLevel: req.JurisdictionLevel,
		FilingFrequency:   financeDomain.FilingFrequency(req.FilingFrequency),
		FilingDueDay:      req.FilingDueDay,
		PaymentDueDay:     req.PaymentDueDay,
		SupportsEFiling:   req.SupportsEFiling,
		EFilingEndpoint:   req.EFilingEndpoint,
		IsActive:          req.IsActive,
		UpdatedBy:         &byUserID,
	}

	// 3. Delegate
	updated, err := h.services.Tax.UpdateAuthority(ctx, authority)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, updated)
}

// ============================================================================
// Tax code handlers
// ============================================================================

// CreateTaxCode creates a new tax code under a tax authority.
// Optionally include brackets for progressive/marginal rate structures.
//
// POST /api/v1/finance/tax/codes
// Permission: finance.tax.codes.create
func (h *FinanceHandler) CreateTaxCode(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CreateTaxCode")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	var req createTaxCodeRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	authorityID, err := parseUUID(req.TaxAuthorityID, "tax_authority_id")
	if err != nil {
		return h.fail(c, err)
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return h.fail(c, newDateParseError("effective_date"))
	}

	taxRate, err := decimalFromString(req.TaxRate)
	if err != nil {
		return h.fail(c, err)
	}

	if req.TaxCategory == "" {
		req.TaxCategory = "standard"
	}

	tc := &financeDomain.TaxCode{
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		TaxType:           financeDomain.TaxType(req.TaxType),
		TaxCategory:       req.TaxCategory,
		TaxAuthorityID:    authorityID,
		CalculationMethod: financeDomain.TaxCalculationMethod(req.CalculationMethod),
		TaxRate:           taxRate,
		CompoundTax:       req.CompoundTax,
		CascadeOrder:      req.CascadeOrder,
		EffectiveDate:     effectiveDate,
		ReportingCode:     req.ReportingCode,
		ReturnLineNumber:  req.ReturnLineNumber,
		IsDefault:         req.IsDefault,
	}
	if req.ExpiryDate != nil {
		exp, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			return h.fail(c, newDateParseError("expiry_date"))
		}
		tc.ExpiryDate = &exp
	}

	brackets, err := parseTaxBrackets(req.Brackets)
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	created, err := h.services.Tax.CreateTaxCode(ctx, tc, brackets)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok201(c, created)
}

// GetTaxCode retrieves a single tax code with its brackets.
//
// GET /api/v1/finance/tax/codes/:id
// Permission: finance.tax.codes.read
func (h *FinanceHandler) GetTaxCode(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetTaxCode")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	codeID, err := parseUUID(c.Params("id"), "tax code ID")
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	tc, err := h.services.Tax.GetTaxCode(ctx, codeID)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, tc)
}

// ListTaxCodes returns tax codes for the current tenant, optionally filtered
// by tax type.
//
// GET /api/v1/finance/tax/codes?tax_type=VAT&active_only=true
// Permission: finance.tax.codes.read
func (h *FinanceHandler) ListTaxCodes(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListTaxCodes")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse — both filters are optional
	activeOnly := c.QueryBool("active_only", false)
	var taxType *financeDomain.TaxType
	if tt := c.Query("tax_type"); tt != "" {
		t := financeDomain.TaxType(tt)
		taxType = &t
	}

	// 3. Delegate
	list, err := h.services.Tax.ListTaxCodes(ctx, taxType, activeOnly)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, list)
}

// UpdateTaxCode replaces a tax code's mutable fields and replaces its brackets.
// The code, tax_type, and tax_authority_id are immutable after creation.
//
// PUT /api/v1/finance/tax/codes/:id
// Permission: finance.tax.codes.update
func (h *FinanceHandler) UpdateTaxCode(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.UpdateTaxCode")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	codeID, err := parseUUID(c.Params("id"), "tax code ID")
	if err != nil {
		return h.fail(c, err)
	}

	var req updateTaxCodeRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return h.fail(c, newDateParseError("effective_date"))
	}

	taxRate, err := decimalFromString(req.TaxRate)
	if err != nil {
		return h.fail(c, err)
	}

	byUserID := extractUserID(c)

	tc := &financeDomain.TaxCode{
		ID:                codeID,
		Name:              req.Name,
		Description:       req.Description,
		TaxCategory:       req.TaxCategory,
		CalculationMethod: financeDomain.TaxCalculationMethod(req.CalculationMethod),
		TaxRate:           taxRate,
		CompoundTax:       req.CompoundTax,
		CascadeOrder:      req.CascadeOrder,
		EffectiveDate:     effectiveDate,
		ReportingCode:     req.ReportingCode,
		ReturnLineNumber:  req.ReturnLineNumber,
		IsActive:          req.IsActive,
		IsDefault:         req.IsDefault,
		UpdatedBy:         &byUserID,
	}
	if req.ExpiryDate != nil {
		exp, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			return h.fail(c, newDateParseError("expiry_date"))
		}
		tc.ExpiryDate = &exp
	}

	brackets, err := parseTaxBrackets(req.Brackets)
	if err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	updated, err := h.services.Tax.UpdateTaxCode(ctx, tc, brackets)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, updated)
}

// ============================================================================
// Private helpers — tax only
// ============================================================================

// parseTaxBrackets converts the request slice into domain TaxBracket slice.
// Amounts are parsed from string to preserve decimal precision.
func parseTaxBrackets(reqs []taxBracketRequest) ([]*financeDomain.TaxBracket, error) {
	if len(reqs) == 0 {
		return nil, nil
	}
	brackets := make([]*financeDomain.TaxBracket, 0, len(reqs))
	for _, b := range reqs {
		minAmt, err := decimalFromString(b.MinimumAmount)
		if err != nil {
			return nil, err
		}
		rate, err := decimalFromString(b.TaxRate)
		if err != nil {
			return nil, err
		}
		br := &financeDomain.TaxBracket{
			BracketNumber:       b.BracketNumber,
			MinimumAmount:       minAmt,
			TaxRate:             rate,
			MarginalCalculation: b.MarginalCalculation,
		}
		if b.MaximumAmount != nil {
			d, err := decimalFromString(*b.MaximumAmount)
			if err != nil {
				return nil, err
			}
			br.MaximumAmount = &d
		}
		brackets = append(brackets, br)
	}
	return brackets, nil
}

// newDateParseError returns a consistent validation error for unparseable dates.
func newDateParseError(field string) error {
	return newValidationError(
		"INVALID_DATE_FORMAT",
		field+" must be in YYYY-MM-DD format",
		field,
		"Example: 2025-01-31",
	)
}
