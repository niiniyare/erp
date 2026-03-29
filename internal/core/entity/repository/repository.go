package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/entity/domain"
)

// Repository is the data-access interface for entities.
// All methods expect the RLS current_tenant_id() to be set on the underlying
// DB connection — either by the session middleware (for request-scoped calls)
// or by wrapping the call in store.WithTenant (for provisioning).
type Repository interface {
	// Create inserts a new entity and seeds the hierarchy-path closure table.
	Create(ctx context.Context, req domain.CreateEntityRequest) (*domain.EntityNode, error)
	// GetByID fetches a single entity by UUID within the current tenant.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.EntityNode, error)
	// List returns all non-deleted entities for the current tenant.
	List(ctx context.Context) ([]*domain.EntityNode, error)
}

// Postgres implementation

type postgresRepo struct {
	store db.Store
}

// NewPostgres returns a Repository backed by PostgreSQL.
func NewPostgres(store db.Store) Repository {
	return &postgresRepo{store: store}
}

func (r *postgresRepo) Create(ctx context.Context, req domain.CreateEntityRequest) (*domain.EntityNode, error) {
	id := uuid.New()

	// Compute materialized path and hierarchy level from parent.
	var entityPath string
	var entityLevel int32 = 1

	if req.ParentID != nil {
		parent, err := r.store.GetEntity(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("entity: get parent %s: %w", req.ParentID, err)
		}
		// Parent path is guaranteed non-empty for existing entities.
		parentPath := "/" + parent.Uuid.String() + "/"
		if parent.EntityPath != nil && *parent.EntityPath != "" {
			parentPath = *parent.EntityPath
		}
		entityPath = parentPath + id.String() + "/"
		entityLevel = parent.EntityLevel + 1
	} else {
		entityPath = "/" + id.String() + "/"
	}

	row, err := r.store.CreateEntity(ctx, db.CreateEntityParams{
		Uuid:          id,
		ParentID:      req.ParentID,
		Name:          req.Name,
		Code:          req.Code,
		Type:          string(req.Type),
		IsActive:      req.IsActive,
		Hidden:        false,
		AccrualMethod: false,
		FyStartMonth:  1,
		EntityPath:    &entityPath,
		EntityLevel:   entityLevel,
		Address:       []byte("{}"),
		Metadata:      []byte("{}"),
		Settings:      []byte("{}"),
	})
	if err != nil {
		return nil, fmt.Errorf("entity: create: %w", err)
	}

	// Closure table — every node is a depth-0 "ancestor" of itself.
	if hpErr := r.store.CreateHierarchyPath(ctx, db.CreateHierarchyPathParams{
		AncestorID:   id,
		DescendantID: id,
		Depth:        0,
	}); hpErr != nil {
		return nil, fmt.Errorf("entity: create self hierarchy path: %w", hpErr)
	}

	// Direct parent→child edge.
	if req.ParentID != nil {
		if hpErr := r.store.CreateHierarchyPath(ctx, db.CreateHierarchyPathParams{
			AncestorID:   *req.ParentID,
			DescendantID: id,
			Depth:        1,
		}); hpErr != nil {
			return nil, fmt.Errorf("entity: create parent hierarchy path: %w", hpErr)
		}
	}

	return toNode(row), nil
}

func (r *postgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EntityNode, error) {
	row, err := r.store.GetEntity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("entity: get %s: %w", id, err)
	}
	return toNode(row), nil
}

func (r *postgresRepo) List(ctx context.Context) ([]*domain.EntityNode, error) {
	rows, err := r.store.ListEntities(ctx)
	if err != nil {
		return nil, fmt.Errorf("entity: list: %w", err)
	}
	nodes := make([]*domain.EntityNode, len(rows))
	for i, row := range rows {
		nodes[i] = toNode(row)
	}
	return nodes, nil
}

// toNode maps a DB Entity row to the domain EntityNode.
func toNode(e *db.Entity) *domain.EntityNode {
	n := &domain.EntityNode{
		ID:          e.Uuid,
		TenantID:    e.TenantID,
		ParentID:    e.ParentID,
		Name:        e.Name,
		Code:        e.Code,
		Type:        domain.EntityType(e.Type),
		IsActive:    e.IsActive,
		EntityLevel: e.EntityLevel,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
	if e.EntityPath != nil {
		n.EntityPath = *e.EntityPath
	}
	return n
}
