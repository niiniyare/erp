package finance

// exchange_rates.go — HTTP handlers for exchange rates and currency management.
//
// Routes (registered in routes.go):
//
//	POST /api/v1/finance/exchange-rates              → UpsertExchangeRate
//	GET  /api/v1/finance/exchange-rates              → GetExchangeRate    (point-in-time lookup)
//	GET  /api/v1/finance/exchange-rates/history      → ListExchangeRates  (date range)
//	POST /api/v1/finance/currencies                  → CreateCurrency
//	GET  /api/v1/finance/currencies                  → ListCurrencies
//	PUT  /api/v1/finance/currencies/:id              → UpdateCurrency

import (
	"time"

	"github.com/gofiber/fiber/v2"

	financeDomain "awo.so/internal/core/finance/domain"
	sharedErrors "awo.so/internal/shared/errors"
)

// ============================================================================
// Exchange rate handlers
// ============================================================================

// UpsertExchangeRate creates or updates an exchange rate for a currency pair
// on a specific date. If a rate for the same pair, type, and date already
// exists it is overwritten (upsert semantics, enforced by service).
//
// POST /api/v1/finance/exchange-rates
// Permission: finance.exchange_rates.write
func (h *FinanceHandler) UpsertExchangeRate(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.UpsertExchangeRate")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate — financeDomain.ExchangeRate carries its own
	//    validation tags so we bind directly to the domain type.
	var rate financeDomain.ExchangeRate
	if err := h.bind(c, &rate); err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	created, err := h.services.ExchangeRate.UpsertRate(ctx, &rate)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok201(c, created)
}

// GetExchangeRate returns the effective rate for a currency pair at a given date.
// When no date is supplied the current day is used.
//
// GET /api/v1/finance/exchange-rates?from=USD&to=KES&rate_type=SPOT&date=2025-01-31
// Permission: finance.exchange_rates.read
func (h *FinanceHandler) GetExchangeRate(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.GetExchangeRate")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate — required query params
	fromCurrency := c.Query("from")
	toCurrency := c.Query("to")
	if fromCurrency == "" || toCurrency == "" {
		return h.fail(c, sharedErrors.NewBusinessError("MISSING_CURRENCY_PARAMS", "'from' and 'to' query parameters are required").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithSuggestion("Example: GET /exchange-rates?from=USD&to=KES"))
	}

	// rate_type defaults to SPOT; unknown values fall back silently
	rateType := financeDomain.RateType(c.Query("rate_type", string(financeDomain.RateTypeSpot)))
	if !rateType.IsValid() {
		rateType = financeDomain.RateTypeSpot
	}

	// date defaults to today
	asOf := time.Now()
	if s := c.Query("date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			asOf = t
		}
	}

	// 3. Delegate
	rate, err := h.services.ExchangeRate.GetRate(ctx, fromCurrency, toCurrency, rateType, asOf)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, rate)
}

// ListExchangeRates returns historical rates for a currency pair within a date range.
//
// GET /api/v1/finance/exchange-rates/history?from=USD&to=KES&from_date=2025-01-01&to_date=2025-01-31&limit=100
// Permission: finance.exchange_rates.read
func (h *FinanceHandler) ListExchangeRates(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListExchangeRates")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	fromCurrency := c.Query("from")
	toCurrency := c.Query("to")
	if fromCurrency == "" || toCurrency == "" {
		return h.fail(c, sharedErrors.NewBusinessError("MISSING_CURRENCY_PARAMS", "'from' and 'to' query parameters are required").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation))
	}

	limit := c.QueryInt("limit", 100)

	var fromDate, toDate *time.Time
	if s := c.Query("from_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			fromDate = &t
		}
	}
	if s := c.Query("to_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			toDate = &t
		}
	}

	// 3. Delegate
	rates, err := h.services.ExchangeRate.ListRates(ctx, fromCurrency, toCurrency, fromDate, toDate, limit)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, rates)
}

// ============================================================================
// Currency handlers
// ============================================================================

// CreateCurrency registers a new currency for the tenant.
// The currency code must be a valid ISO 4217 code (enforced by service).
//
// POST /api/v1/finance/currencies
// Permission: finance.currencies.create
func (h *FinanceHandler) CreateCurrency(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.CreateCurrency")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	var currency financeDomain.Currency
	if err := h.bind(c, &currency); err != nil {
		return h.fail(c, err)
	}

	// 3. Delegate
	created, err := h.services.Currency.CreateCurrency(ctx, &currency)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok201(c, created)
}

// ListCurrencies returns all currencies configured for the tenant.
//
// GET /api/v1/finance/currencies?active_only=true
// Permission: finance.currencies.read
func (h *FinanceHandler) ListCurrencies(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.ListCurrencies")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse — single optional bool query param
	activeOnly := c.QueryBool("active_only", false)

	// 3. Delegate
	currencies, err := h.services.Currency.ListCurrencies(ctx, activeOnly)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, currencies)
}

// UpdateCurrency updates a currency's display name, symbol, or active flag.
// The currency code itself is immutable after creation.
//
// PUT /api/v1/finance/currencies/:id
// Permission: finance.currencies.update
func (h *FinanceHandler) UpdateCurrency(c *fiber.Ctx) error {
	// 1. Observability
	ctx, span := h.tracer.StartSpan(c.UserContext(), "finance.UpdateCurrency")
	defer span.End()
	c.SetUserContext(ctx)

	// 2. Parse & validate
	currencyID, err := parseUUID(c.Params("id"), "currency ID")
	if err != nil {
		return h.fail(c, err)
	}

	var currency financeDomain.Currency
	if err := h.bind(c, &currency); err != nil {
		return h.fail(c, err)
	}
	// Path param ID is authoritative; ignore any ID in the body
	currency.ID = currencyID

	// 3. Delegate
	updated, err := h.services.Currency.UpdateCurrency(ctx, &currency)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	// 4. Response
	return h.ok200(c, updated)
}
