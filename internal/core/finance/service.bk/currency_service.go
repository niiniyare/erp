package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// ExchangeRateService manages currency exchange rates.
type ExchangeRateService interface {
	// UpsertRate inserts or updates an exchange rate.
	UpsertRate(ctx context.Context, rate *domain.ExchangeRate) (*domain.ExchangeRate, error)
	// GetRate returns the most recent rate for the pair on or before asOfDate.
	GetRate(ctx context.Context, fromCurrency, toCurrency string, rateType domain.RateType, asOfDate time.Time) (*domain.ExchangeRate, error)
	// ListRates lists rates for a currency pair within an optional date range.
	ListRates(ctx context.Context, fromCurrency, toCurrency string, from, to *time.Time, limit int) ([]*domain.ExchangeRate, error)
}

type exchangeRateService struct {
	repo    domain.ExchangeRateRepository
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewExchangeRateService creates a new ExchangeRateService.
func NewExchangeRateService(
	repo domain.ExchangeRateRepository,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) ExchangeRateService {
	return &exchangeRateService{
		repo:    repo,
		tracing: tracing,
		metrics: metrics,
	}
}

func (s *exchangeRateService) UpsertRate(ctx context.Context, rate *domain.ExchangeRate) (*domain.ExchangeRate, error) {
	ctx, span := s.tracing.StartSpan(ctx, "exchange_rate_service.upsert_rate")
	defer span.End()

	if errs := rate.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("EXCHANGE_RATE_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	if err := s.repo.Upsert(ctx, rate); err != nil {
		span.RecordError(err)
		return nil, mapExchangeRateError(err, "upsert rate")
	}

	s.metrics.IncrementCounter("exchange_rate_upserted_total", metrics.Fields{
		"from": rate.FromCurrency,
		"to":   rate.ToCurrency,
	})
	return rate, nil
}

func (s *exchangeRateService) GetRate(ctx context.Context, fromCurrency, toCurrency string, rateType domain.RateType, asOfDate time.Time) (*domain.ExchangeRate, error) {
	ctx, span := s.tracing.StartSpan(ctx, "exchange_rate_service.get_rate")
	defer span.End()

	rate, err := s.repo.GetRate(ctx, fromCurrency, toCurrency, rateType, asOfDate)
	if err != nil {
		return nil, mapExchangeRateError(err, "get rate")
	}
	return rate, nil
}

func (s *exchangeRateService) ListRates(ctx context.Context, fromCurrency, toCurrency string, from, to *time.Time, limit int) ([]*domain.ExchangeRate, error) {
	ctx, span := s.tracing.StartSpan(ctx, "exchange_rate_service.list_rates")
	defer span.End()

	rates, err := s.repo.ListRates(ctx, fromCurrency, toCurrency, from, to, limit)
	if err != nil {
		return nil, mapExchangeRateError(err, "list rates")
	}
	return rates, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func mapExchangeRateError(err error, op string) error {
	if err == domain.ErrExchangeRateNotFound {
		return errors.NewBusinessError("EXCHANGE_RATE_NOT_FOUND", "exchange rate not found").
			WithHTTPStatus(http.StatusNotFound)
	}
	return errors.NewBusinessError("EXCHANGE_RATE_ERROR", fmt.Sprintf("%s: %v", op, err)).
		WithHTTPStatus(http.StatusInternalServerError)
}

// ── Currency (finance_currencies table) ─────────────────────────────────────

// CurrencyRepository is the minimal interface for currency CRUD, defined
// here to avoid a separate domain file for now.
type CurrencyRepository interface {
	Create(ctx context.Context, c *domain.Currency) error
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.Currency, error)
	List(ctx context.Context, tenantID uuid.UUID, activeOnly bool) ([]*domain.Currency, error)
	Update(ctx context.Context, c *domain.Currency) error
}

// CurrencyService manages the currencies a tenant supports.
type CurrencyService interface {
	CreateCurrency(ctx context.Context, c *domain.Currency) (*domain.Currency, error)
	ListCurrencies(ctx context.Context, activeOnly bool) ([]*domain.Currency, error)
	UpdateCurrency(ctx context.Context, c *domain.Currency) (*domain.Currency, error)
}

type currencyService struct {
	repo    CurrencyRepository
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewCurrencyService creates a new CurrencyService.
func NewCurrencyService(
	repo CurrencyRepository,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) CurrencyService {
	return &currencyService{
		repo:    repo,
		tracing: tracing,
		metrics: metrics,
	}
}

func (s *currencyService) CreateCurrency(ctx context.Context, c *domain.Currency) (*domain.Currency, error) {
	ctx, span := s.tracing.StartSpan(ctx, "currency_service.create")
	defer span.End()

	if errs := c.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("CURRENCY_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	if err := s.repo.Create(ctx, c); err != nil {
		span.RecordError(err)
		return nil, mapCurrencyError(err, "create currency")
	}
	s.metrics.IncrementCounter("currency_created_total", metrics.Fields{})
	return c, nil
}

func (s *currencyService) ListCurrencies(ctx context.Context, activeOnly bool) ([]*domain.Currency, error) {
	ctx, span := s.tracing.StartSpan(ctx, "currency_service.list")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	currencies, err := s.repo.List(ctx, tenantID, activeOnly)
	if err != nil {
		return nil, mapCurrencyError(err, "list currencies")
	}
	return currencies, nil
}

func (s *currencyService) UpdateCurrency(ctx context.Context, c *domain.Currency) (*domain.Currency, error) {
	ctx, span := s.tracing.StartSpan(ctx, "currency_service.update")
	defer span.End()

	if errs := c.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("CURRENCY_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	if err := s.repo.Update(ctx, c); err != nil {
		span.RecordError(err)
		return nil, mapCurrencyError(err, "update currency")
	}
	return c, nil
}

func mapCurrencyError(err error, op string) error {
	return errors.NewBusinessError("CURRENCY_ERROR", fmt.Sprintf("%s: %v", op, err)).
		WithHTTPStatus(http.StatusInternalServerError)
}
