package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/entity/domain"
	"awo.so/internal/core/entity/repository"
)

// Service is the public API for entity management.
type Service interface {
	// Create inserts a new entity under the current tenant (RLS must be set).
	Create(ctx context.Context, req domain.CreateEntityRequest) (*domain.EntityNode, error)
	// GetByID fetches a single entity within the current tenant.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.EntityNode, error)
	// List returns all non-deleted entities for the current tenant.
	List(ctx context.Context) ([]*domain.EntityNode, error)
	// CreateRoot creates a company-root entity for a newly provisioned tenant.
	// It opens a dedicated tenant-scoped transaction so current_tenant_id() is
	// correctly set even when called outside of an active session.
	CreateRoot(ctx context.Context, tenantID uuid.UUID, name string) (*domain.EntityNode, error)
}

// Implementation

type entityService struct {
	repo  repository.Repository
	store db.Store
}

// NewService wires the service with a Postgres-backed repository.
func NewService(repo repository.Repository, store db.Store) Service {
	return &entityService{repo: repo, store: store}
}

func (s *entityService) Create(ctx context.Context, req domain.CreateEntityRequest) (*domain.EntityNode, error) {
	return s.repo.Create(ctx, req)
}

func (s *entityService) GetByID(ctx context.Context, id uuid.UUID) (*domain.EntityNode, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *entityService) List(ctx context.Context) ([]*domain.EntityNode, error) {
	return s.repo.List(ctx)
}

// CreateRoot opens a tenant-scoped DB transaction (setting current_tenant_id())
// and creates a company-level root entity. A fresh repository is created inside
// the transaction scope so that all SQLC calls use the same RLS-aware connection.
func (s *entityService) CreateRoot(ctx context.Context, tenantID uuid.UUID, name string) (*domain.EntityNode, error) {
	var result *domain.EntityNode
	err := s.store.WithTenant(ctx, tenantID, func(ctx context.Context, scopedStore db.Store) error {
		scopedRepo := repository.NewPostgres(scopedStore)
		var err error
		result, err = scopedRepo.Create(ctx, domain.CreateEntityRequest{
			Name:     name,
			Type:     domain.EntityTypeCompany,
			IsActive: true,
		})
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("entity: create root for tenant %s: %w", tenantID, err)
	}
	return result, nil
}
