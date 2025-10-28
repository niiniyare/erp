package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// ExchangeRateEngine handles multi-currency exchange rate management and conversion
type ExchangeRateEngine interface {
	// GetExchangeRate retrieves the exchange rate between two currencies
	GetExchangeRate(ctx context.Context, req GetExchangeRateRequest) (*ExchangeRateResult, error)

	// ConvertAmount converts an amount from one currency to another
	ConvertAmount(ctx context.Context, req ConvertAmountRequest) (*ConvertAmountResult, error)

	// ValidateMultiCurrencyTransaction validates multi-currency transaction requirements
	ValidateMultiCurrencyTransaction(ctx context.Context, transaction *domain.Transaction) (*MultiCurrencyValidationResult, error)

	// UpdateExchangeRate updates or creates an exchange rate
	UpdateExchangeRate(ctx context.Context, req UpdateExchangeRateRequest) (*UpdateExchangeRateResult, error)

	// GetHistoricalRate retrieves historical exchange rate for a specific date
	GetHistoricalRate(ctx context.Context, req HistoricalRateRequest) (*ExchangeRateResult, error)

	// GetCurrencyPairRates gets all available rates for a currency pair
	GetCurrencyPairRates(ctx context.Context, req CurrencyPairRatesRequest) (*CurrencyPairRatesResult, error)

	// RefreshRates refreshes exchange rates from external sources
	RefreshRates(ctx context.Context, req RefreshRatesRequest) (*RefreshRatesResult, error)
}

// GetExchangeRateRequest represents a request to get exchange rate
type GetExchangeRateRequest struct {
	FromCurrency string     `json:"from_currency"`
	ToCurrency   string     `json:"to_currency"`
	RateDate     *time.Time `json:"rate_date,omitempty"` // If nil, uses current date
	RateType     RateType   `json:"rate_type"`
}

// ConvertAmountRequest represents a request to convert amount between currencies
type ConvertAmountRequest struct {
	Amount        decimal.Decimal `json:"amount"`
	FromCurrency  string          `json:"from_currency"`
	ToCurrency    string          `json:"to_currency"`
	RateDate      *time.Time      `json:"rate_date,omitempty"`
	RateType      RateType        `json:"rate_type"`
	RoundingMode  RoundingMode    `json:"rounding_mode"`
	DecimalPlaces int32           `json:"decimal_places"`
}

// UpdateExchangeRateRequest represents a request to update exchange rate
type UpdateExchangeRateRequest struct {
	FromCurrency string          `json:"from_currency"`
	ToCurrency   string          `json:"to_currency"`
	Rate         decimal.Decimal `json:"rate"`
	RateDate     time.Time       `json:"rate_date"`
	RateType     RateType        `json:"rate_type"`
	Source       string          `json:"source"`
	UpdatedBy    uuid.UUID       `json:"updated_by"`
}

// HistoricalRateRequest represents a request for historical exchange rate
type HistoricalRateRequest struct {
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	RateDate     time.Time `json:"rate_date"`
	RateType     RateType  `json:"rate_type"`
}

// CurrencyPairRatesRequest represents a request for currency pair rates
type CurrencyPairRatesRequest struct {
	FromCurrency string     `json:"from_currency"`
	ToCurrency   string     `json:"to_currency"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	RateType     RateType   `json:"rate_type"`
	Limit        int32      `json:"limit"`
}

// RefreshRatesRequest represents a request to refresh rates
type RefreshRatesRequest struct {
	CurrencyPairs []CurrencyPair `json:"currency_pairs,omitempty"`
	RateType      RateType       `json:"rate_type"`
	Source        string         `json:"source"`
	RefreshedBy   uuid.UUID      `json:"refreshed_by"`
}

// RateType represents the type of exchange rate
type RateType string

const (
	RateTypeSpot    RateType = "SPOT"    // Current market rate
	RateTypeAverage RateType = "AVERAGE" // Daily average rate
	RateTypeClosing RateType = "CLOSING" // End of day rate
	RateTypeOpening RateType = "OPENING" // Start of day rate
	RateTypeBuying  RateType = "BUYING"  // Bank buying rate
	RateTypeSelling RateType = "SELLING" // Bank selling rate
)

// RoundingMode represents how to round converted amounts
type RoundingMode string

const (
	RoundingModeHalfUp   RoundingMode = "HALF_UP"
	RoundingModeHalfDown RoundingMode = "HALF_DOWN"
	RoundingModeCeiling  RoundingMode = "CEILING"
	RoundingModeFloor    RoundingMode = "FLOOR"
)

// CurrencyPair represents a currency pair
type CurrencyPair struct {
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
}

// ExchangeRateResult represents the result of getting exchange rate
type ExchangeRateResult struct {
	FromCurrency string          `json:"from_currency"`
	ToCurrency   string          `json:"to_currency"`
	Rate         decimal.Decimal `json:"rate"`
	InverseRate  decimal.Decimal `json:"inverse_rate"`
	RateDate     time.Time       `json:"rate_date"`
	RateType     RateType        `json:"rate_type"`
	Source       string          `json:"source"`
	IsEstimated  bool            `json:"is_estimated"` // If rate was estimated/calculated
	LastUpdated  time.Time       `json:"last_updated"`
}

// ConvertAmountResult represents the result of amount conversion
type ConvertAmountResult struct {
	OriginalAmount    decimal.Decimal     `json:"original_amount"`
	ConvertedAmount   decimal.Decimal     `json:"converted_amount"`
	FromCurrency      string              `json:"from_currency"`
	ToCurrency        string              `json:"to_currency"`
	ExchangeRate      *ExchangeRateResult `json:"exchange_rate"`
	ConversionDetails *ConversionDetails  `json:"conversion_details"`
}

// ConversionDetails provides details about the conversion calculation
type ConversionDetails struct {
	RateUsed        decimal.Decimal `json:"rate_used"`
	RoundingApplied bool            `json:"rounding_applied"`
	RoundingMode    RoundingMode    `json:"rounding_mode"`
	DecimalPlaces   int32           `json:"decimal_places"`
	RawAmount       decimal.Decimal `json:"raw_amount"` // Before rounding
	ConversionTime  time.Time       `json:"conversion_time"`
}

// UpdateExchangeRateResult represents the result of updating exchange rate
type UpdateExchangeRateResult struct {
	Success       bool                `json:"success"`
	ExchangeRate  *ExchangeRateResult `json:"exchange_rate,omitempty"`
	PreviousRate  *ExchangeRateResult `json:"previous_rate,omitempty"`
	PercentChange decimal.Decimal     `json:"percent_change"`
	UpdatedAt     time.Time           `json:"updated_at"`
	Errors        []string            `json:"errors,omitempty"`
}

// CurrencyPairRatesResult represents historical rates for a currency pair
type CurrencyPairRatesResult struct {
	FromCurrency string               `json:"from_currency"`
	ToCurrency   string               `json:"to_currency"`
	Rates        []ExchangeRateResult `json:"rates"`
	DateRange    DateRange            `json:"date_range"`
	RateType     RateType             `json:"rate_type"`
	TotalRecords int32                `json:"total_records"`
}

// RefreshRatesResult represents the result of refreshing rates
type RefreshRatesResult struct {
	TotalPairsProcessed int32                `json:"total_pairs_processed"`
	SuccessfulUpdates   int32                `json:"successful_updates"`
	FailedUpdates       int32                `json:"failed_updates"`
	UpdatedRates        []ExchangeRateResult `json:"updated_rates"`
	FailedPairs         []CurrencyPairError  `json:"failed_pairs,omitempty"`
	RefreshTime         time.Time            `json:"refresh_time"`
	ProcessingTime      time.Duration        `json:"processing_time"`
}

// MultiCurrencyValidationResult represents validation results for multi-currency transactions
type MultiCurrencyValidationResult struct {
	IsValid              bool                    `json:"is_valid"`
	BaseCurrency         string                  `json:"base_currency"`
	ForeignCurrencies    []string                `json:"foreign_currencies"`
	MissingRates         []CurrencyPair          `json:"missing_rates,omitempty"`
	OutdatedRates        []OutdatedRate          `json:"outdated_rates,omitempty"`
	ConversionErrors     []ConversionError       `json:"conversion_errors,omitempty"`
	TotalConvertedAmount decimal.Decimal         `json:"total_converted_amount"`
	ConversionDetails    []EntryConversionDetail `json:"conversion_details"`
}

// DateRange represents a date range
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// CurrencyPairError represents an error for a specific currency pair
type CurrencyPairError struct {
	CurrencyPair CurrencyPair `json:"currency_pair"`
	Error        string       `json:"error"`
	ErrorCode    string       `json:"error_code"`
}

// OutdatedRate represents an outdated exchange rate
type OutdatedRate struct {
	CurrencyPair CurrencyPair    `json:"currency_pair"`
	RateDate     time.Time       `json:"rate_date"`
	DaysOld      int32           `json:"days_old"`
	Rate         decimal.Decimal `json:"rate"`
}

// ConversionError represents a conversion error
type ConversionError struct {
	EntryIndex   int32        `json:"entry_index"`
	CurrencyPair CurrencyPair `json:"currency_pair"`
	Error        string       `json:"error"`
	ErrorCode    string       `json:"error_code"`
}

// EntryConversionDetail represents conversion details for a transaction entry
type EntryConversionDetail struct {
	EntryID             uuid.UUID       `json:"entry_id"`
	OriginalAmount      decimal.Decimal `json:"original_amount"`
	OriginalCurrency    string          `json:"original_currency"`
	ConvertedAmount     decimal.Decimal `json:"converted_amount"`
	BaseCurrency        string          `json:"base_currency"`
	ExchangeRate        decimal.Decimal `json:"exchange_rate"`
	ConversionTimestamp time.Time       `json:"conversion_timestamp"`
}

// exchangeRateEngine implements ExchangeRateEngine
type exchangeRateEngine struct {
	tracing              tracing.Service
	rateCache            *RateCache
	externalProvider     ExternalRateProvider
	defaultDecimalPlaces int32
	defaultRoundingMode  RoundingMode
	maxRateAge           time.Duration
}

// ExchangeRateEngineDeps represents dependencies for the exchange rate engine
type ExchangeRateEngineDeps struct {
	Tracing              tracing.Service
	ExternalProvider     ExternalRateProvider
	DefaultDecimalPlaces *int32
	DefaultRoundingMode  *RoundingMode
	MaxRateAge           *time.Duration
	CacheTTL             *time.Duration
}

// ExternalRateProvider interface for external rate providers
type ExternalRateProvider interface {
	GetRate(ctx context.Context, fromCurrency, toCurrency string, rateType RateType) (*ExchangeRateResult, error)
	GetHistoricalRate(ctx context.Context, fromCurrency, toCurrency string, date time.Time, rateType RateType) (*ExchangeRateResult, error)
}

// RateCache manages in-memory caching of exchange rates
type RateCache struct {
	rates map[string]*cachedRate
	mutex sync.RWMutex
	ttl   time.Duration
}

type cachedRate struct {
	rate      *ExchangeRateResult
	timestamp time.Time
}

// NewExchangeRateEngine creates a new exchange rate engine
func NewExchangeRateEngine(deps ExchangeRateEngineDeps) ExchangeRateEngine {
	decimalPlaces := int32(2)
	if deps.DefaultDecimalPlaces != nil {
		decimalPlaces = *deps.DefaultDecimalPlaces
	}

	roundingMode := RoundingModeHalfUp
	if deps.DefaultRoundingMode != nil {
		roundingMode = *deps.DefaultRoundingMode
	}

	maxAge := 24 * time.Hour
	if deps.MaxRateAge != nil {
		maxAge = *deps.MaxRateAge
	}

	cacheTTL := 1 * time.Hour
	if deps.CacheTTL != nil {
		cacheTTL = *deps.CacheTTL
	}

	return &exchangeRateEngine{
		tracing:              deps.Tracing,
		rateCache:            NewRateCache(cacheTTL),
		externalProvider:     deps.ExternalProvider,
		defaultDecimalPlaces: decimalPlaces,
		defaultRoundingMode:  roundingMode,
		maxRateAge:           maxAge,
	}
}

// NewRateCache creates a new rate cache
func NewRateCache(ttl time.Duration) *RateCache {
	return &RateCache{
		rates: make(map[string]*cachedRate),
		ttl:   ttl,
	}
}

// GetExchangeRate retrieves the exchange rate between two currencies
func (e *exchangeRateEngine) GetExchangeRate(ctx context.Context, req GetExchangeRateRequest) (*ExchangeRateResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "ExchangeRateEngine.GetExchangeRate")
	defer span.End()

	span.SetAttributes(
		attribute.String("from_currency", req.FromCurrency),
		attribute.String("to_currency", req.ToCurrency),
		attribute.String("rate_type", string(req.RateType)),
	)

	// Handle same currency case
	if req.FromCurrency == req.ToCurrency {
		rateDate := time.Now()
		if req.RateDate != nil {
			rateDate = *req.RateDate
		}

		return &ExchangeRateResult{
			FromCurrency: req.FromCurrency,
			ToCurrency:   req.ToCurrency,
			Rate:         decimal.NewFromInt(1),
			InverseRate:  decimal.NewFromInt(1),
			RateDate:     rateDate,
			RateType:     req.RateType,
			Source:       "SYSTEM",
			IsEstimated:  false,
			LastUpdated:  time.Now(),
		}, nil
	}

	// Try cache first
	cacheKey := e.buildCacheKey(req.FromCurrency, req.ToCurrency, req.RateType, req.RateDate)
	if cachedRate := e.rateCache.Get(cacheKey); cachedRate != nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		return cachedRate, nil
	}

	// Try external provider
	if e.externalProvider != nil {
		if req.RateDate != nil {
			rate, err := e.externalProvider.GetHistoricalRate(ctx, req.FromCurrency, req.ToCurrency, *req.RateDate, req.RateType)
			if err == nil {
				e.rateCache.Set(cacheKey, rate)
				return rate, nil
			}
		} else {
			rate, err := e.externalProvider.GetRate(ctx, req.FromCurrency, req.ToCurrency, req.RateType)
			if err == nil {
				e.rateCache.Set(cacheKey, rate)
				return rate, nil
			}
		}
	}

	// TODO: Try database/repository lookup
	// For now, return a default rate with warning
	rateDate := time.Now()
	if req.RateDate != nil {
		rateDate = *req.RateDate
	}

	return &ExchangeRateResult{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Rate:         decimal.NewFromInt(1), // Default 1:1 rate
		InverseRate:  decimal.NewFromInt(1),
		RateDate:     rateDate,
		RateType:     req.RateType,
		Source:       "DEFAULT",
		IsEstimated:  true,
		LastUpdated:  time.Now(),
	}, nil
}

// ConvertAmount converts an amount from one currency to another
func (e *exchangeRateEngine) ConvertAmount(ctx context.Context, req ConvertAmountRequest) (*ConvertAmountResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "ExchangeRateEngine.ConvertAmount")
	defer span.End()

	// Get exchange rate
	rateReq := GetExchangeRateRequest{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		RateDate:     req.RateDate,
		RateType:     req.RateType,
	}

	exchangeRate, err := e.GetExchangeRate(ctx, rateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	// Calculate converted amount
	rawAmount := req.Amount.Mul(exchangeRate.Rate)

	// Apply rounding
	roundedAmount := e.applyRounding(rawAmount, req.RoundingMode, req.DecimalPlaces)

	conversionDetails := &ConversionDetails{
		RateUsed:        exchangeRate.Rate,
		RoundingApplied: !rawAmount.Equal(roundedAmount),
		RoundingMode:    req.RoundingMode,
		DecimalPlaces:   req.DecimalPlaces,
		RawAmount:       rawAmount,
		ConversionTime:  time.Now(),
	}

	result := &ConvertAmountResult{
		OriginalAmount:    req.Amount,
		ConvertedAmount:   roundedAmount,
		FromCurrency:      req.FromCurrency,
		ToCurrency:        req.ToCurrency,
		ExchangeRate:      exchangeRate,
		ConversionDetails: conversionDetails,
	}

	span.SetAttributes(
		attribute.String("original_amount", req.Amount.String()),
		attribute.String("converted_amount", roundedAmount.String()),
		attribute.String("rate", exchangeRate.Rate.String()),
	)

	return result, nil
}

// ValidateMultiCurrencyTransaction validates multi-currency transaction requirements
func (e *exchangeRateEngine) ValidateMultiCurrencyTransaction(ctx context.Context, transaction *domain.Transaction) (*MultiCurrencyValidationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "ExchangeRateEngine.ValidateMultiCurrencyTransaction")
	defer span.End()

	result := &MultiCurrencyValidationResult{
		IsValid:              true,
		BaseCurrency:         transaction.CurrencyCode,
		ForeignCurrencies:    []string{},
		MissingRates:         []CurrencyPair{},
		OutdatedRates:        []OutdatedRate{},
		ConversionErrors:     []ConversionError{},
		TotalConvertedAmount: decimal.Zero,
		ConversionDetails:    []EntryConversionDetail{},
	}

	// Collect foreign currencies
	currenciesUsed := make(map[string]bool)
	currenciesUsed[transaction.CurrencyCode] = true

	for _, entry := range transaction.Entries {
		if entry.OriginalCurrency != nil && *entry.OriginalCurrency != transaction.CurrencyCode {
			foreignCurrency := *entry.OriginalCurrency
			if !currenciesUsed[foreignCurrency] {
				result.ForeignCurrencies = append(result.ForeignCurrencies, foreignCurrency)
				currenciesUsed[foreignCurrency] = true
			}
		}
	}

	// Validate rates for each foreign currency
	for _, foreignCurrency := range result.ForeignCurrencies {
		rateReq := GetExchangeRateRequest{
			FromCurrency: foreignCurrency,
			ToCurrency:   transaction.CurrencyCode,
			RateDate:     &transaction.TransactionDate,
			RateType:     RateTypeSpot,
		}

		exchangeRate, err := e.GetExchangeRate(ctx, rateReq)
		if err != nil {
			result.IsValid = false
			result.MissingRates = append(result.MissingRates, CurrencyPair{
				FromCurrency: foreignCurrency,
				ToCurrency:   transaction.CurrencyCode,
			})
			continue
		}

		// Check if rate is outdated
		if time.Since(exchangeRate.LastUpdated) > e.maxRateAge {
			result.OutdatedRates = append(result.OutdatedRates, OutdatedRate{
				CurrencyPair: CurrencyPair{
					FromCurrency: foreignCurrency,
					ToCurrency:   transaction.CurrencyCode,
				},
				RateDate: exchangeRate.RateDate,
				DaysOld:  int32(time.Since(exchangeRate.LastUpdated).Hours() / 24),
				Rate:     exchangeRate.Rate,
			})
		}
	}

	// Convert foreign currency entries
	for i, entry := range transaction.Entries {
		if entry.OriginalCurrency != nil && *entry.OriginalCurrency != transaction.CurrencyCode {
			convertReq := ConvertAmountRequest{
				Amount:        entry.DebitAmount.Add(entry.CreditAmount),
				FromCurrency:  *entry.OriginalCurrency,
				ToCurrency:    transaction.CurrencyCode,
				RateDate:      &transaction.TransactionDate,
				RateType:      RateTypeSpot,
				RoundingMode:  e.defaultRoundingMode,
				DecimalPlaces: e.defaultDecimalPlaces,
			}

			convertResult, err := e.ConvertAmount(ctx, convertReq)
			if err != nil {
				result.IsValid = false
				result.ConversionErrors = append(result.ConversionErrors, ConversionError{
					EntryIndex: int32(i),
					CurrencyPair: CurrencyPair{
						FromCurrency: *entry.OriginalCurrency,
						ToCurrency:   transaction.CurrencyCode,
					},
					Error:     err.Error(),
					ErrorCode: "CONVERSION_FAILED",
				})
				continue
			}

			result.TotalConvertedAmount = result.TotalConvertedAmount.Add(convertResult.ConvertedAmount)
			result.ConversionDetails = append(result.ConversionDetails, EntryConversionDetail{
				EntryID:             entry.ID,
				OriginalAmount:      convertReq.Amount,
				OriginalCurrency:    *entry.OriginalCurrency,
				ConvertedAmount:     convertResult.ConvertedAmount,
				BaseCurrency:        transaction.CurrencyCode,
				ExchangeRate:        convertResult.ExchangeRate.Rate,
				ConversionTimestamp: time.Now(),
			})
		} else {
			// Add base currency entries
			result.TotalConvertedAmount = result.TotalConvertedAmount.Add(entry.DebitAmount.Add(entry.CreditAmount))
		}
	}

	return result, nil
}

// UpdateExchangeRate updates or creates an exchange rate
func (e *exchangeRateEngine) UpdateExchangeRate(ctx context.Context, req UpdateExchangeRateRequest) (*UpdateExchangeRateResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "ExchangeRateEngine.UpdateExchangeRate")
	defer span.End()

	// Get previous rate for comparison
	prevRateReq := GetExchangeRateRequest{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		RateDate:     &req.RateDate,
		RateType:     req.RateType,
	}

	previousRate, _ := e.GetExchangeRate(ctx, prevRateReq)

	// Create new rate
	newRate := &ExchangeRateResult{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Rate:         req.Rate,
		InverseRate:  decimal.NewFromInt(1).Div(req.Rate),
		RateDate:     req.RateDate,
		RateType:     req.RateType,
		Source:       req.Source,
		IsEstimated:  false,
		LastUpdated:  time.Now(),
	}

	// Calculate percent change
	percentChange := decimal.Zero
	if previousRate != nil && !previousRate.Rate.IsZero() {
		percentChange = req.Rate.Sub(previousRate.Rate).Div(previousRate.Rate).Mul(decimal.NewFromInt(100))
	}

	// Update cache
	cacheKey := e.buildCacheKey(req.FromCurrency, req.ToCurrency, req.RateType, &req.RateDate)
	e.rateCache.Set(cacheKey, newRate)

	// TODO: Save to database/repository

	result := &UpdateExchangeRateResult{
		Success:       true,
		ExchangeRate:  newRate,
		PreviousRate:  previousRate,
		PercentChange: percentChange,
		UpdatedAt:     time.Now(),
		Errors:        []string{},
	}

	span.SetAttributes(
		attribute.String("from_currency", req.FromCurrency),
		attribute.String("to_currency", req.ToCurrency),
		attribute.String("new_rate", req.Rate.String()),
		attribute.String("percent_change", percentChange.String()),
	)

	return result, nil
}

// GetHistoricalRate retrieves historical exchange rate for a specific date
func (e *exchangeRateEngine) GetHistoricalRate(ctx context.Context, req HistoricalRateRequest) (*ExchangeRateResult, error) {
	getRateReq := GetExchangeRateRequest{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		RateDate:     &req.RateDate,
		RateType:     req.RateType,
	}

	return e.GetExchangeRate(ctx, getRateReq)
}

// GetCurrencyPairRates gets all available rates for a currency pair
func (e *exchangeRateEngine) GetCurrencyPairRates(ctx context.Context, req CurrencyPairRatesRequest) (*CurrencyPairRatesResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "ExchangeRateEngine.GetCurrencyPairRates")
	defer span.End()

	// For now, return empty result - in production this would query historical data
	result := &CurrencyPairRatesResult{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Rates:        []ExchangeRateResult{},
		RateType:     req.RateType,
		TotalRecords: 0,
	}

	if req.StartDate != nil && req.EndDate != nil {
		result.DateRange = DateRange{
			StartDate: *req.StartDate,
			EndDate:   *req.EndDate,
		}
	}

	return result, nil
}

// RefreshRates refreshes exchange rates from external sources
func (e *exchangeRateEngine) RefreshRates(ctx context.Context, req RefreshRatesRequest) (*RefreshRatesResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "ExchangeRateEngine.RefreshRates")
	defer span.End()

	startTime := time.Now()

	result := &RefreshRatesResult{
		TotalPairsProcessed: int32(len(req.CurrencyPairs)),
		SuccessfulUpdates:   0,
		FailedUpdates:       0,
		UpdatedRates:        []ExchangeRateResult{},
		FailedPairs:         []CurrencyPairError{},
		RefreshTime:         startTime,
	}

	if e.externalProvider == nil {
		result.ProcessingTime = time.Since(startTime)
		return result, nil
	}

	// Process each currency pair
	for _, pair := range req.CurrencyPairs {
		rate, err := e.externalProvider.GetRate(ctx, pair.FromCurrency, pair.ToCurrency, req.RateType)
		if err != nil {
			result.FailedUpdates++
			result.FailedPairs = append(result.FailedPairs, CurrencyPairError{
				CurrencyPair: pair,
				Error:        err.Error(),
				ErrorCode:    "EXTERNAL_RATE_FETCH_FAILED",
			})
			continue
		}

		// Update the rate
		updateReq := UpdateExchangeRateRequest{
			FromCurrency: pair.FromCurrency,
			ToCurrency:   pair.ToCurrency,
			Rate:         rate.Rate,
			RateDate:     rate.RateDate,
			RateType:     req.RateType,
			Source:       req.Source,
			UpdatedBy:    req.RefreshedBy,
		}

		_, err = e.UpdateExchangeRate(ctx, updateReq)
		if err != nil {
			result.FailedUpdates++
			result.FailedPairs = append(result.FailedPairs, CurrencyPairError{
				CurrencyPair: pair,
				Error:        err.Error(),
				ErrorCode:    "RATE_UPDATE_FAILED",
			})
		} else {
			result.SuccessfulUpdates++
			result.UpdatedRates = append(result.UpdatedRates, *rate)
		}
	}

	result.ProcessingTime = time.Since(startTime)

	span.SetAttributes(
		attribute.Int64("total_pairs", int64(result.TotalPairsProcessed)),
		attribute.Int64("successful_updates", int64(result.SuccessfulUpdates)),
		attribute.Int64("failed_updates", int64(result.FailedUpdates)),
		attribute.String("processing_time", result.ProcessingTime.String()),
	)

	return result, nil
}

// Helper methods

func (e *exchangeRateEngine) buildCacheKey(fromCurrency, toCurrency string, rateType RateType, rateDate *time.Time) string {
	dateStr := "current"
	if rateDate != nil {
		dateStr = rateDate.Format("2006-01-02")
	}
	return fmt.Sprintf("%s-%s-%s-%s", fromCurrency, toCurrency, string(rateType), dateStr)
}

func (e *exchangeRateEngine) applyRounding(amount decimal.Decimal, roundingMode RoundingMode, decimalPlaces int32) decimal.Decimal {
	switch roundingMode {
	case RoundingModeHalfUp:
		return amount.Round(decimalPlaces)
	case RoundingModeHalfDown:
		return amount.RoundDown(decimalPlaces)
	case RoundingModeCeiling:
		return amount.RoundCeil(decimalPlaces)
	case RoundingModeFloor:
		return amount.RoundFloor(decimalPlaces)
	default:
		return amount.Round(decimalPlaces)
	}
}

// RateCache methods

func (c *RateCache) Get(key string) *ExchangeRateResult {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	cached, exists := c.rates[key]
	if !exists {
		return nil
	}

	// Check if cached rate has expired
	if time.Since(cached.timestamp) > c.ttl {
		delete(c.rates, key)
		return nil
	}

	return cached.rate
}

func (c *RateCache) Set(key string, rate *ExchangeRateResult) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.rates[key] = &cachedRate{
		rate:      rate,
		timestamp: time.Now(),
	}
}

func (c *RateCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.rates = make(map[string]*cachedRate)
}
