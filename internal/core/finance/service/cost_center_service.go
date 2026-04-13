package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared/errors"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// CostCenterService manages cost centres and their allocation rules.
type CostCenterService interface {
	Create(ctx context.Context, cc *domain.CostCenter) (*domain.CostCenter, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CostCenter, error)
	GetByCode(ctx context.Context, code string) (*domain.CostCenter, error)
	Update(ctx context.Context, cc *domain.CostCenter) (*domain.CostCenter, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, activeOnly bool) ([]*domain.CostCenter, error)
}

type costCenterService struct {
	repo    domain.CostCenterRepository
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewCostCenterService creates a new CostCenterService.
func NewCostCenterService(
	repo domain.CostCenterRepository,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) CostCenterService {
	return &costCenterService{
		repo:    repo,
		tracing: tracing,
		metrics: metrics,
	}
}

func (s *costCenterService) Create(ctx context.Context, cc *domain.CostCenter) (*domain.CostCenter, error) {
	ctx, span := s.tracing.StartSpan(ctx, "cost_center_service.create")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	cc.TenantID = tenantID

	if errs := cc.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("COST_CENTER_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	// Validate code uniqueness
	if err := s.repo.ValidateCode(ctx, tenantID, cc.Code, nil); err != nil {
		return nil, mapCostCenterError(err, "validate code")
	}

	if err := s.repo.Create(ctx, cc); err != nil {
		span.RecordError(err)
		return nil, mapCostCenterError(err, "create cost centre")
	}

	s.metrics.IncrementCounter("cost_center_created_total", metrics.Fields{})
	return cc, nil
}

func (s *costCenterService) GetByID(ctx context.Context, id uuid.UUID) (*domain.CostCenter, error) {
	ctx, span := s.tracing.StartSpan(ctx, "cost_center_service.get_by_id")
	defer span.End()

	cc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, mapCostCenterError(err, "get cost centre")
	}
	return cc, nil
}

func (s *costCenterService) GetByCode(ctx context.Context, code string) (*domain.CostCenter, error) {
	ctx, span := s.tracing.StartSpan(ctx, "cost_center_service.get_by_code")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	cc, err := s.repo.GetByCode(ctx, tenantID, code)
	if err != nil {
		return nil, mapCostCenterError(err, "get cost centre by code")
	}
	return cc, nil
}

func (s *costCenterService) Update(ctx context.Context, cc *domain.CostCenter) (*domain.CostCenter, error) {
	ctx, span := s.tracing.StartSpan(ctx, "cost_center_service.update")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if errs := cc.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("COST_CENTER_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	// Validate code uniqueness (exclude current record)
	if err := s.repo.ValidateCode(ctx, tenantID, cc.Code, &cc.ID); err != nil {
		return nil, mapCostCenterError(err, "validate code")
	}

	if err := s.repo.Update(ctx, cc); err != nil {
		span.RecordError(err)
		return nil, mapCostCenterError(err, "update cost centre")
	}

	s.metrics.IncrementCounter("cost_center_updated_total", metrics.Fields{})
	return cc, nil
}

func (s *costCenterService) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "cost_center_service.delete")
	defer span.End()

	if err := s.repo.Delete(ctx, id); err != nil {
		span.RecordError(err)
		return mapCostCenterError(err, "delete cost centre")
	}

	s.metrics.IncrementCounter("cost_center_deleted_total", metrics.Fields{})
	return nil
}

func (s *costCenterService) List(ctx context.Context, activeOnly bool) ([]*domain.CostCenter, error) {
	ctx, span := s.tracing.StartSpan(ctx, "cost_center_service.list")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	ccs, err := s.repo.List(ctx, tenantID, activeOnly)
	if err != nil {
		return nil, mapCostCenterError(err, "list cost centres")
	}
	return ccs, nil
}

func mapCostCenterError(err error, op string) error {
	switch err {
	case domain.ErrCostCenterNotFound:
		return errors.NewBusinessError("COST_CENTER_NOT_FOUND", "cost centre not found").
			WithHTTPStatus(http.StatusNotFound)
	case domain.ErrCostCenterCodeExists:
		return errors.NewBusinessError("COST_CENTER_CODE_EXISTS", "cost centre code already exists").
			WithHTTPStatus(http.StatusConflict)
	default:
		return errors.NewBusinessError("COST_CENTER_ERROR", fmt.Sprintf("%s: %v", op, err)).
			WithHTTPStatus(http.StatusInternalServerError)
	}
}
