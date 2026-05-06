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

// BudgetService manages budget headers and line items.
type BudgetService interface {
	// CreateBudget creates a new budget with optional line items.
	CreateBudget(ctx context.Context, b *domain.Budget, lines []*domain.BudgetLineItem) (*domain.Budget, error)
	// GetBudget returns a budget by ID with its line items populated.
	GetBudget(ctx context.Context, id uuid.UUID) (*domain.Budget, error)
	// ListBudgets lists budgets, optionally filtered by fiscal year.
	ListBudgets(ctx context.Context, fiscalYearID *uuid.UUID) ([]*domain.Budget, error)
	// SubmitBudget transitions a DRAFT budget to SUBMITTED.
	SubmitBudget(ctx context.Context, id uuid.UUID, byUserID uuid.UUID) (*domain.Budget, error)
	// ApproveBudget transitions a SUBMITTED budget to APPROVED.
	ApproveBudget(ctx context.Context, id uuid.UUID, byUserID uuid.UUID) (*domain.Budget, error)
	// RejectBudget transitions a SUBMITTED budget back to REJECTED.
	RejectBudget(ctx context.Context, id uuid.UUID, byUserID uuid.UUID, note string) (*domain.Budget, error)
	// CloseBudget transitions an APPROVED/ACTIVE budget to CLOSED.
	CloseBudget(ctx context.Context, id uuid.UUID) (*domain.Budget, error)
	// GetLineItems returns line items for a budget.
	GetLineItems(ctx context.Context, budgetID uuid.UUID) ([]*domain.BudgetLineItem, error)
}

type budgetService struct {
	repo    domain.BudgetRepository
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewBudgetService creates a new BudgetService.
func NewBudgetService(
	repo domain.BudgetRepository,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) BudgetService {
	return &budgetService{
		repo:    repo,
		tracing: tracing,
		metrics: metrics,
	}
}

func (s *budgetService) CreateBudget(ctx context.Context, b *domain.Budget, lines []*domain.BudgetLineItem) (*domain.Budget, error) {
	ctx, span := s.tracing.StartSpan(ctx, "budget_service.create")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	b.TenantID = tenantID

	if b.Status == "" {
		b.Status = domain.BudgetStatusDraft
	}
	if b.Version == 0 {
		b.Version = 1
	}

	if errs := b.Validate(); len(errs) > 0 {
		return nil, errors.NewBusinessError("BUDGET_VALIDATION_FAILED", errs[0].Message).
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(errors.CategoryValidation)
	}

	if err := s.repo.CreateBudget(ctx, b); err != nil {
		span.RecordError(err)
		return nil, mapBudgetError(err, "create budget")
	}

	// Attach budget ID and tenant ID to each line then persist
	if len(lines) > 0 {
		for _, li := range lines {
			li.BudgetID = b.ID
			li.TenantID = tenantID
		}
		if err := s.repo.CreateLineItems(ctx, lines); err != nil {
			span.RecordError(err)
			return nil, mapBudgetError(err, "create budget line items")
		}
		b.Lines = lines // stitch onto returned budget
	}

	s.metrics.IncrementCounter("budget_created_total", metrics.Fields{})
	return b, nil
}

func (s *budgetService) GetBudget(ctx context.Context, id uuid.UUID) (*domain.Budget, error) {
	ctx, span := s.tracing.StartSpan(ctx, "budget_service.get")
	defer span.End()

	b, err := s.repo.GetBudgetByID(ctx, id)
	if err != nil {
		return nil, mapBudgetError(err, "get budget")
	}

	// Load line items
	lines, err := s.repo.GetLineItems(ctx, id)
	if err != nil {
		return nil, mapBudgetError(err, "get budget line items")
	}
	b.Lines = lines
	return b, nil
}

func (s *budgetService) ListBudgets(ctx context.Context, fiscalYearID *uuid.UUID) ([]*domain.Budget, error) {
	ctx, span := s.tracing.StartSpan(ctx, "budget_service.list")
	defer span.End()

	tenantID, err := tenantIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	budgets, err := s.repo.ListBudgets(ctx, tenantID, fiscalYearID)
	if err != nil {
		return nil, mapBudgetError(err, "list budgets")
	}
	return budgets, nil
}

func (s *budgetService) SubmitBudget(ctx context.Context, id uuid.UUID, byUserID uuid.UUID) (*domain.Budget, error) {
	ctx, span := s.tracing.StartSpan(ctx, "budget_service.submit")
	defer span.End()

	b, err := s.repo.GetBudgetByID(ctx, id)
	if err != nil {
		return nil, mapBudgetError(err, "get budget")
	}
	if b.Status != domain.BudgetStatusDraft && b.Status != domain.BudgetStatusRejected {
		return nil, errors.NewBusinessError("BUDGET_INVALID_TRANSITION",
			fmt.Sprintf("cannot submit budget in status %s", b.Status)).
			WithHTTPStatus(http.StatusConflict)
	}

	now := time.Now()
	b.Status = domain.BudgetStatusSubmitted
	b.SubmittedAt = &now
	b.SubmittedBy = &byUserID
	b.UpdatedBy = &byUserID

	if err := s.repo.UpdateBudget(ctx, b); err != nil {
		span.RecordError(err)
		return nil, mapBudgetError(err, "submit budget")
	}
	return b, nil
}

func (s *budgetService) ApproveBudget(ctx context.Context, id uuid.UUID, byUserID uuid.UUID) (*domain.Budget, error) {
	ctx, span := s.tracing.StartSpan(ctx, "budget_service.approve")
	defer span.End()

	b, err := s.repo.GetBudgetByID(ctx, id)
	if err != nil {
		return nil, mapBudgetError(err, "get budget")
	}
	if b.Status != domain.BudgetStatusSubmitted {
		return nil, errors.NewBusinessError("BUDGET_INVALID_TRANSITION",
			fmt.Sprintf("cannot approve budget in status %s; must be SUBMITTED", b.Status)).
			WithHTTPStatus(http.StatusConflict)
	}

	now := time.Now()
	b.Status = domain.BudgetStatusApproved
	b.ApprovedAt = &now
	b.ApprovedBy = &byUserID
	b.UpdatedBy = &byUserID

	if err := s.repo.UpdateBudget(ctx, b); err != nil {
		span.RecordError(err)
		return nil, mapBudgetError(err, "approve budget")
	}
	s.metrics.IncrementCounter("budget_approved_total", metrics.Fields{})
	return b, nil
}

func (s *budgetService) RejectBudget(ctx context.Context, id uuid.UUID, byUserID uuid.UUID, note string) (*domain.Budget, error) {
	ctx, span := s.tracing.StartSpan(ctx, "budget_service.reject")
	defer span.End()

	b, err := s.repo.GetBudgetByID(ctx, id)
	if err != nil {
		return nil, mapBudgetError(err, "get budget")
	}
	if b.Status != domain.BudgetStatusSubmitted {
		return nil, errors.NewBusinessError("BUDGET_INVALID_TRANSITION",
			fmt.Sprintf("cannot reject budget in status %s; must be SUBMITTED", b.Status)).
			WithHTTPStatus(http.StatusConflict)
	}

	now := time.Now()
	b.Status = domain.BudgetStatusRejected
	b.RejectedAt = &now
	b.RejectedBy = &byUserID
	if note != "" {
		b.RejectNote = &note
	}
	b.UpdatedBy = &byUserID

	if err := s.repo.UpdateBudget(ctx, b); err != nil {
		span.RecordError(err)
		return nil, mapBudgetError(err, "reject budget")
	}
	return b, nil
}

func (s *budgetService) CloseBudget(ctx context.Context, id uuid.UUID) (*domain.Budget, error) {
	ctx, span := s.tracing.StartSpan(ctx, "budget_service.close")
	defer span.End()

	b, err := s.repo.GetBudgetByID(ctx, id)
	if err != nil {
		return nil, mapBudgetError(err, "get budget")
	}
	if !b.CanClose() {
		return nil, errors.NewBusinessError("BUDGET_INVALID_TRANSITION",
			fmt.Sprintf("cannot close budget in status %s", b.Status)).
			WithHTTPStatus(http.StatusConflict)
	}

	b.Status = domain.BudgetStatusClosed

	if err := s.repo.UpdateBudget(ctx, b); err != nil {
		span.RecordError(err)
		return nil, mapBudgetError(err, "close budget")
	}
	return b, nil
}

func (s *budgetService) GetLineItems(ctx context.Context, budgetID uuid.UUID) ([]*domain.BudgetLineItem, error) {
	ctx, span := s.tracing.StartSpan(ctx, "budget_service.get_line_items")
	defer span.End()

	lines, err := s.repo.GetLineItems(ctx, budgetID)
	if err != nil {
		return nil, mapBudgetError(err, "get line items")
	}
	return lines, nil
}

func mapBudgetError(err error, op string) error {
	switch err {
	case domain.ErrBudgetNotFound:
		return errors.NewBusinessError("BUDGET_NOT_FOUND", "budget not found").
			WithHTTPStatus(http.StatusNotFound)
	case domain.ErrBudgetNotEditable:
		return errors.NewBusinessError("BUDGET_NOT_EDITABLE", "budget is not editable in its current status").
			WithHTTPStatus(http.StatusConflict)
	case domain.ErrBudgetAlreadyApproved:
		return errors.NewBusinessError("BUDGET_ALREADY_APPROVED", "budget is already approved").
			WithHTTPStatus(http.StatusConflict)
	default:
		return errors.NewBusinessError("BUDGET_ERROR", fmt.Sprintf("%s: %v", op, err)).
			WithHTTPStatus(http.StatusInternalServerError)
	}
}
