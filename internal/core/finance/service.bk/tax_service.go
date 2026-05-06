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

// TaxService manages tax authorities, codes, and brackets.
type TaxService interface {
	// Authority operations
	CreateAuthority(ctx context.Context, a *domain.TaxAuthority) (*domain.TaxAuthority, error)
	GetAuthority(ctx context.Context, id uuid.UUID) (*domain.TaxAuthority, error)
	ListAuthorities(ctx context.Context, activeOnly bool) ([]*domain.TaxAuthority, error)
	UpdateAuthority(ctx context.Context, a *domain.TaxAuthority) (*domain.TaxAuthority, error)

	// Tax code operations
	CreateTaxCode(ctx context.Context, tc *domain.TaxCode, brackets []*domain.TaxBracket) (*domain.TaxCode, error)
	GetTaxCode(ctx context.Context, id uuid.UUID) (*domain.TaxCode, error)
	ListTaxCodes(ctx context.Context, taxType *domain.TaxType, activeOnly bool) ([]*domain.TaxCode, error)
	UpdateTaxCode(ctx context.Context, tc *domain.TaxCode, brackets []*domain.TaxBracket) (*domain.TaxCode, error)
}

type taxService struct {
	repo    domain.TaxRepository
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewTaxService creates a new TaxService.
func NewTaxService(
	repo domain.TaxRepository,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) TaxService {
	return &taxService{repo: repo, tracing: tracing, metrics: metrics}
}

// ── Authority ─────────────────────────────────────────────────────────────────

func (s *taxService) CreateAuthority(ctx context.Context, a *domain.TaxAuthority) (*domain.TaxAuthority, error) {
	ctx, span := s.tracing.StartSpan(ctx, "tax_service.create_authority")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	a.TenantID = tenantID
	a.IsActive = true

	if errs := a.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("TAX_AUTHORITY_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	if err := s.repo.CreateAuthority(ctx, a); err != nil {
		span.RecordError(err)
		return nil, mapTaxError(err, "create tax authority")
	}
	s.metrics.IncrementCounter("tax_authority_created_total", metrics.Fields{})
	return a, nil
}

func (s *taxService) GetAuthority(ctx context.Context, id uuid.UUID) (*domain.TaxAuthority, error) {
	ctx, span := s.tracing.StartSpan(ctx, "tax_service.get_authority")
	defer span.End()

	a, err := s.repo.GetAuthorityByID(ctx, id)
	if err != nil {
		return nil, mapTaxError(err, "get tax authority")
	}
	return a, nil
}

func (s *taxService) ListAuthorities(ctx context.Context, activeOnly bool) ([]*domain.TaxAuthority, error) {
	ctx, span := s.tracing.StartSpan(ctx, "tax_service.list_authorities")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := s.repo.ListAuthorities(ctx, tenantID, activeOnly)
	if err != nil {
		return nil, mapTaxError(err, "list tax authorities")
	}
	return list, nil
}

func (s *taxService) UpdateAuthority(ctx context.Context, a *domain.TaxAuthority) (*domain.TaxAuthority, error) {
	ctx, span := s.tracing.StartSpan(ctx, "tax_service.update_authority")
	defer span.End()

	// Fetch existing record to get immutable fields (code) and verify existence
	existing, err := s.repo.GetAuthorityByID(ctx, a.ID)
	if err != nil {
		return nil, mapTaxError(err, "get tax authority")
	}

	// Preserve immutable fields
	a.TenantID = existing.TenantID
	a.AuthorityCode = existing.AuthorityCode
	a.CreatedAt = existing.CreatedAt
	a.CreatedBy = existing.CreatedBy

	if errs := a.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("TAX_AUTHORITY_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	now := time.Now()
	a.UpdatedAt = now

	if err := s.repo.UpdateAuthority(ctx, a); err != nil {
		span.RecordError(err)
		return nil, mapTaxError(err, "update tax authority")
	}
	return a, nil
}

// ── Tax Code ──────────────────────────────────────────────────────────────────

func (s *taxService) CreateTaxCode(ctx context.Context, tc *domain.TaxCode, brackets []*domain.TaxBracket) (*domain.TaxCode, error) {
	ctx, span := s.tracing.StartSpan(ctx, "tax_service.create_tax_code")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	tc.TenantID = tenantID
	tc.IsActive = true

	if errs := tc.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("TAX_CODE_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	if err := s.repo.CreateTaxCode(ctx, tc); err != nil {
		span.RecordError(err)
		return nil, mapTaxError(err, "create tax code")
	}

	if len(brackets) > 0 {
		if err := s.repo.CreateBrackets(ctx, tc.ID, brackets); err != nil {
			span.RecordError(err)
			return nil, mapTaxError(err, "create tax brackets")
		}
		tc.Brackets = brackets
	}

	s.metrics.IncrementCounter("tax_code_created_total", metrics.Fields{})
	return tc, nil
}

func (s *taxService) GetTaxCode(ctx context.Context, id uuid.UUID) (*domain.TaxCode, error) {
	ctx, span := s.tracing.StartSpan(ctx, "tax_service.get_tax_code")
	defer span.End()

	tc, err := s.repo.GetTaxCodeByID(ctx, id)
	if err != nil {
		return nil, mapTaxError(err, "get tax code")
	}

	// Load brackets for progressive codes
	if tc.CalculationMethod == domain.TaxCalcProgressive {
		brackets, err := s.repo.GetBrackets(ctx, id)
		if err != nil {
			return nil, mapTaxError(err, "get tax brackets")
		}
		tc.Brackets = brackets
	}

	return tc, nil
}

func (s *taxService) ListTaxCodes(ctx context.Context, taxType *domain.TaxType, activeOnly bool) ([]*domain.TaxCode, error) {
	ctx, span := s.tracing.StartSpan(ctx, "tax_service.list_tax_codes")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := s.repo.ListTaxCodes(ctx, tenantID, taxType, activeOnly)
	if err != nil {
		return nil, mapTaxError(err, "list tax codes")
	}
	return list, nil
}

func (s *taxService) UpdateTaxCode(ctx context.Context, tc *domain.TaxCode, brackets []*domain.TaxBracket) (*domain.TaxCode, error) {
	ctx, span := s.tracing.StartSpan(ctx, "tax_service.update_tax_code")
	defer span.End()

	// Fetch existing to get immutable fields and verify existence
	existing, err := s.repo.GetTaxCodeByID(ctx, tc.ID)
	if err != nil {
		return nil, mapTaxError(err, "get tax code")
	}

	// Preserve immutable fields
	tc.TenantID = existing.TenantID
	tc.Code = existing.Code
	tc.TaxType = existing.TaxType
	tc.TaxAuthorityID = existing.TaxAuthorityID
	tc.CreatedAt = existing.CreatedAt
	tc.CreatedBy = existing.CreatedBy

	if errs := tc.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("TAX_CODE_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	if err := s.repo.UpdateTaxCode(ctx, tc); err != nil {
		span.RecordError(err)
		return nil, mapTaxError(err, "update tax code")
	}

	// Replace brackets if provided
	if brackets != nil {
		if err := s.repo.DeleteBrackets(ctx, tc.ID); err != nil {
			span.RecordError(err)
			return nil, mapTaxError(err, "delete old brackets")
		}
		if len(brackets) > 0 {
			if err := s.repo.CreateBrackets(ctx, tc.ID, brackets); err != nil {
				span.RecordError(err)
				return nil, mapTaxError(err, "create new brackets")
			}
		}
		tc.Brackets = brackets
	}

	return tc, nil
}

func mapTaxError(err error, op string) error {
	switch err {
	case domain.ErrTaxAuthorityNotFound:
		return errors.NewBusinessError("TAX_AUTHORITY_NOT_FOUND", "tax authority not found").
			WithHTTPStatus(http.StatusNotFound)
	case domain.ErrTaxAuthorityExists:
		return errors.NewBusinessError("TAX_AUTHORITY_EXISTS", "tax authority code already exists").
			WithHTTPStatus(http.StatusConflict)
	case domain.ErrTaxCodeNotFound:
		return errors.NewBusinessError("TAX_CODE_NOT_FOUND", "tax code not found").
			WithHTTPStatus(http.StatusNotFound)
	case domain.ErrTaxCodeExists:
		return errors.NewBusinessError("TAX_CODE_EXISTS", "tax code already exists").
			WithHTTPStatus(http.StatusConflict)
	default:
		return errors.NewBusinessError("TAX_ERROR", fmt.Sprintf("%s: %v", op, err)).
			WithHTTPStatus(http.StatusInternalServerError)
	}
}
