package service

import (
	"context"
	"fmt"

	"github.com/niiniyare/erp/internal/core/tenant/domain"
	"github.com/niiniyare/erp/internal/core/tenant/repository"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// ProvisioningService orchestrates tenant provisioning.
type ProvisioningService struct {
	repo   repository.Repository
	tracer tracing.Service
}

// NewProvisioningService creates a new ProvisioningService.
func NewProvisioningService(repo repository.Repository, tracer tracing.Service) *ProvisioningService {
	return &ProvisioningService{repo: repo, tracer: tracer}
}

// Provision creates a tenant via the database provisioning function.
func (s *ProvisioningService) Provision(ctx context.Context, input domain.ProvisioningInput) (*domain.ProvisioningResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "tenant.provisioning.Provision")
	defer span.End()
	span.SetAttributes(attribute.String("tenant.name", input.Name))

	result, err := s.repo.Provision(ctx, input)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "provisioning failed")
		return nil, fmt.Errorf("provisioning failed: %w", err)
	}

	span.SetAttributes(attribute.String("tenant.id", result.TenantID.String()))
	return result, nil
}
