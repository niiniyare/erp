package entity

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// Service defines the interface for entity business logic
type Service interface {
	// Entity CRUD operations
	CreateEntity(ctx context.Context, req CreateEntityRequest) (*Entity, error)
	GetEntity(ctx context.Context, identifier string) (*Entity, error)
	GetEntityByID(ctx context.Context, id uuid.UUID) (*Entity, error)
	UpdateEntity(ctx context.Context, id uuid.UUID, req UpdateEntityRequest) (*Entity, error)
	DeleteEntity(ctx context.Context, id uuid.UUID, permanent bool) error
	RestoreEntity(ctx context.Context, id uuid.UUID) error
	ListEntities(ctx context.Context, req ListEntitiesRequest) ([]*Entity, error)

	// Hierarchy operations
	GetEntityChildren(ctx context.Context, entityID uuid.UUID) ([]*Entity, error)
	GetEntityAncestors(ctx context.Context, entityID uuid.UUID) ([]*Entity, error)
	GetEntityWithHierarchy(ctx context.Context, entityID uuid.UUID) (*EntityWithHierarchy, error)
	GetEntityTree(ctx context.Context) ([]*EntityWithHierarchy, error)

	// Entity state operations
	GetNextSequence(ctx context.Context, req EntitySequenceRequest) (int64, error)
	ResetSequence(ctx context.Context, req EntitySequenceRequest) error
}

// service implements the Service interface
type service struct {
	repo    Repository
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewService creates a new entity service
func NewService(repo Repository, tracing tracing.Service, metrics metrics.MetricsProvider) Service {
	return &service{
		repo:    repo,
		tracing: tracing,
		metrics: metrics,
	}
}

// CreateEntity creates a new entity with business validation
func (s *service) CreateEntity(ctx context.Context, req CreateEntityRequest) (*Entity, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.create_entity",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.name", req.Name),
			attribute.String("entity.code", req.Code),
			attribute.String("entity.type", string(req.Type)),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting entity creation",
		logger.Fields{
			"entity_name": req.Name,
			"entity_code": req.Code,
			"entity_type": string(req.Type),
		})

	// Validate entity type
	if !req.Type.IsValid() {
		s.metrics.IncrementCounter("entity_creation_errors", metrics.Fields{
			"error_type": "invalid_type",
		})

		logger.WarnContext(ctx, "Invalid entity type provided",
			logger.Fields{"entity_type": string(req.Type)})

		return nil, errors.ErrInvalidEntityType
	}

	// Validate unique name
	ctx, nameSpan := s.tracing.StartSpan(ctx, "service.validate_entity_name")
	if err := s.repo.ValidateEntityName(ctx, req.Name, nil); err != nil {
		nameSpan.RecordError(err)
		nameSpan.End()

		s.metrics.IncrementCounter("entity_creation_errors", metrics.Fields{
			"error_type": "name_exists",
		})

		logger.WarnContext(ctx, "Entity name already exists",
			logger.Fields{"entity_name": req.Name})

		return nil, err
	}
	nameSpan.End()

	// Validate unique code
	ctx, codeSpan := s.tracing.StartSpan(ctx, "service.validate_entity_code")
	if err := s.repo.ValidateEntityCode(ctx, req.Code, nil); err != nil {
		codeSpan.RecordError(err)
		codeSpan.End()

		s.metrics.IncrementCounter("entity_creation_errors", metrics.Fields{
			"error_type": "code_exists",
		})

		logger.WarnContext(ctx, "Entity code already exists",
			logger.Fields{"entity_code": req.Code})

		return nil, err
	}
	codeSpan.End()

	// Validate parent if provided
	if req.ParentID != nil {
		ctx, parentSpan := s.tracing.StartSpan(ctx, "service.validate_entity_parent")
		if err := s.repo.ValidateEntityParent(ctx, *req.ParentID, uuid.New()); err != nil {
			parentSpan.RecordError(err)
			parentSpan.End()

			s.metrics.IncrementCounter("entity_creation_errors", metrics.Fields{
				"error_type": "invalid_parent",
			})

			logger.WarnContext(ctx, "Invalid parent entity",
				logger.Fields{"parent_id": req.ParentID.String()})

			return nil, err
		}
		parentSpan.End()
	}

	// Create entity
	timer := s.metrics.Timer("entity_creation_duration", metrics.Fields{
		"entity_type": string(req.Type),
	})

	entity, err := s.repo.Create(ctx, &req)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("entity_creation_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to create entity",
			logger.Fields{"error": err.Error()})

		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	// Success metrics
	s.metrics.IncrementCounter("entities_created_total", metrics.Fields{
		"entity_type": string(req.Type),
		"status":      "success",
	})

	s.metrics.ObserveHistogram("entity_creation_duration",
		duration.Seconds(), metrics.Fields{
			"entity_type": string(req.Type),
		})

	logger.InfoContext(ctx, "Entity created successfully",
		logger.Fields{
			"entity_id":   entity.ID.String(),
			"entity_name": entity.Name,
			"entity_type": string(entity.Type),
			"duration_ms": duration.Milliseconds(),
		})

	return entity, nil
}

// GetEntity retrieves an entity by identifier (ID, code, or name)
func (s *service) GetEntity(ctx context.Context, identifier string) (*Entity, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_entity",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.identifier", identifier),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting entity by identifier",
		logger.Fields{"identifier": identifier})

	// Try to parse as UUID first
	if id, err := uuid.Parse(identifier); err == nil {
		return s.GetEntityByID(ctx, id)
	}

	// Try to get by code
	entity, err := s.repo.GetByCode(ctx, identifier)
	if err != nil && err != errors.ErrEntityNotFound {
		logger.ErrorContext(ctx, "Failed to get entity by code",
			logger.Fields{
				"identifier": identifier,
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to get entity by code: %w", err)
	}

	if entity != nil {
		logger.DebugContext(ctx, "Entity found by code",
			logger.Fields{
				"entity_id":   entity.ID.String(),
				"entity_code": entity.Code,
			})
		return entity, nil
	}

	// Try to get by name
	entity, err = s.repo.GetByName(ctx, identifier)
	if err != nil {
		if err == errors.ErrEntityNotFound {
			logger.WarnContext(ctx, "Entity not found",
				logger.Fields{"identifier": identifier})
			return nil, err
		}

		logger.ErrorContext(ctx, "Failed to get entity by name",
			logger.Fields{
				"identifier": identifier,
				"error":      err.Error(),
			})
		return nil, fmt.Errorf("failed to get entity by name: %w", err)
	}

	logger.DebugContext(ctx, "Entity found by name",
		logger.Fields{
			"entity_id":   entity.ID.String(),
			"entity_name": entity.Name,
		})

	return entity, nil
}

// GetEntityByID retrieves an entity by ID
func (s *service) GetEntityByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_entity_by_id",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting entity by ID",
		logger.Fields{"entity_id": id.String()})

	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == errors.ErrEntityNotFound {
			logger.WarnContext(ctx, "Entity not found",
				logger.Fields{"entity_id": id.String()})
			return nil, err
		}

		logger.ErrorContext(ctx, "Failed to get entity by ID",
			logger.Fields{
				"entity_id": id.String(),
				"error":     err.Error(),
			})
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	logger.DebugContext(ctx, "Entity retrieved successfully",
		logger.Fields{
			"entity_id":   entity.ID.String(),
			"entity_name": entity.Name,
		})

	return entity, nil
}

// UpdateEntity updates an existing entity with business validation
func (s *service) UpdateEntity(ctx context.Context, id uuid.UUID, req UpdateEntityRequest) (*Entity, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.update_entity",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting entity update",
		logger.Fields{"entity_id": id.String()})

	// Validate entity exists
	existingEntity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Entity not found for update",
			logger.Fields{
				"entity_id": id.String(),
				"error":     err.Error(),
			})
		return nil, err
	}

	// Validate entity type if provided
	if req.Type != nil && !req.Type.IsValid() {
		s.metrics.IncrementCounter("entity_update_errors", metrics.Fields{
			"error_type": "invalid_type",
		})

		logger.WarnContext(ctx, "Invalid entity type provided",
			logger.Fields{"entity_type": string(*req.Type)})

		return nil, errors.ErrInvalidEntityType
	}

	// Validate unique name if provided
	if req.Name != nil && *req.Name != existingEntity.Name {
		if err := s.repo.ValidateEntityName(ctx, *req.Name, &id); err != nil {
			s.metrics.IncrementCounter("entity_update_errors", metrics.Fields{
				"error_type": "name_exists",
			})

			logger.WarnContext(ctx, "Entity name already exists",
				logger.Fields{"entity_name": *req.Name})

			return nil, err
		}
	}

	// Validate unique code if provided
	if req.Code != nil && *req.Code != existingEntity.Code {
		if err := s.repo.ValidateEntityCode(ctx, *req.Code, &id); err != nil {
			s.metrics.IncrementCounter("entity_update_errors", metrics.Fields{
				"error_type": "code_exists",
			})

			logger.WarnContext(ctx, "Entity code already exists",
				logger.Fields{"entity_code": *req.Code})

			return nil, err
		}
	}

	// Validate parent changes
	if req.ParentID != nil {
		// Check if parent is changing
		if existingEntity.ParentID == nil || *existingEntity.ParentID != *req.ParentID {
			// Validate new parent
			if err := s.repo.ValidateEntityParent(ctx, *req.ParentID, id); err != nil {
				s.metrics.IncrementCounter("entity_update_errors", metrics.Fields{
					"error_type": "invalid_parent",
				})

				logger.WarnContext(ctx, "Invalid parent entity",
					logger.Fields{"parent_id": req.ParentID.String()})

				return nil, err
			}

			// Check for circular reference
			isAncestor, err := s.repo.IsEntityAncestor(ctx, id, *req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("failed to check circular reference: %w", err)
			}
			if isAncestor {
				s.metrics.IncrementCounter("entity_update_errors", metrics.Fields{
					"error_type": "circular_reference",
				})

				logger.WarnContext(ctx, "Circular reference detected",
					logger.Fields{
						"entity_id": id.String(),
						"parent_id": req.ParentID.String(),
					})

				return nil, errors.ErrCircularReference
			}
		}
	}

	// Update entity
	timer := s.metrics.Timer("entity_update_duration", metrics.Fields{
		"entity_type": string(existingEntity.Type),
	})

	updatedEntity, err := s.repo.Update(ctx, id, &req)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("entity_update_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to update entity",
			logger.Fields{
				"entity_id": id.String(),
				"error":     err.Error(),
			})

		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	// Success metrics
	s.metrics.IncrementCounter("entities_updated_total", metrics.Fields{
		"entity_type": string(existingEntity.Type),
		"status":      "success",
	})

	s.metrics.ObserveHistogram("entity_update_duration",
		duration.Seconds(), metrics.Fields{
			"entity_type": string(existingEntity.Type),
		})

	logger.InfoContext(ctx, "Entity updated successfully",
		logger.Fields{
			"entity_id":   updatedEntity.ID.String(),
			"entity_name": updatedEntity.Name,
			"duration_ms": duration.Milliseconds(),
		})

	return updatedEntity, nil
}

// DeleteEntity deletes an entity (soft or hard delete)
func (s *service) DeleteEntity(ctx context.Context, id uuid.UUID, permanent bool) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.delete_entity",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.id", id.String()),
			attribute.Bool("permanent", permanent),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting entity deletion",
		logger.Fields{
			"entity_id": id.String(),
			"permanent": permanent,
		})

	// Validate entity exists
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Entity not found for deletion",
			logger.Fields{
				"entity_id": id.String(),
				"error":     err.Error(),
			})
		return err
	}

	// Check for children if permanent delete
	if permanent {
		children, err := s.repo.GetChildren(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to check for children: %w", err)
		}

		if len(children) > 0 {
			s.metrics.IncrementCounter("entity_deletion_errors", metrics.Fields{
				"error_type": "has_children",
			})

			logger.WarnContext(ctx, "Cannot delete entity with children",
				logger.Fields{
					"entity_id":      id.String(),
					"children_count": len(children),
				})

			return errors.ErrEntityHasChildren
		}
	}

	// Delete entity
	timer := s.metrics.Timer("entity_deletion_duration", metrics.Fields{
		"entity_type": string(entity.Type),
		"permanent":   permanent,
	})

	err = s.repo.Delete(ctx, id, permanent)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("entity_deletion_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to delete entity",
			logger.Fields{
				"entity_id": id.String(),
				"error":     err.Error(),
			})

		return fmt.Errorf("failed to delete entity: %w", err)
	}

	// Success metrics
	s.metrics.IncrementCounter("entities_deleted_total", metrics.Fields{
		"entity_type": string(entity.Type),
		"permanent":   permanent,
		"status":      "success",
	})

	s.metrics.ObserveHistogram("entity_deletion_duration",
		duration.Seconds(), metrics.Fields{
			"entity_type": string(entity.Type),
			"permanent":   permanent,
		})

	logger.InfoContext(ctx, "Entity deleted successfully",
		logger.Fields{
			"entity_id":   id.String(),
			"entity_name": entity.Name,
			"permanent":   permanent,
			"duration_ms": duration.Milliseconds(),
		})

	return nil
}

// RestoreEntity restores a soft-deleted entity
func (s *service) RestoreEntity(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.restore_entity",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.id", id.String()),
		))
	defer span.End()

	logger.InfoContext(ctx, "Starting entity restoration",
		logger.Fields{"entity_id": id.String()})

	timer := s.metrics.Timer("entity_restoration_duration", metrics.Fields{})

	err := s.repo.Restore(ctx, id)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("entity_restoration_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to restore entity",
			logger.Fields{
				"entity_id": id.String(),
				"error":     err.Error(),
			})

		return fmt.Errorf("failed to restore entity: %w", err)
	}

	// Success metrics
	s.metrics.IncrementCounter("entities_restored_total", metrics.Fields{
		"status": "success",
	})

	s.metrics.ObserveHistogram("entity_restoration_duration",
		duration.Seconds(), metrics.Fields{})

	logger.InfoContext(ctx, "Entity restored successfully",
		logger.Fields{
			"entity_id":   id.String(),
			"duration_ms": duration.Milliseconds(),
		})

	return nil
}

// ListEntities retrieves entities with filtering and pagination
func (s *service) ListEntities(ctx context.Context, req ListEntitiesRequest) ([]*Entity, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.list_entities",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.Int("limit", req.Limit),
			attribute.Int("offset", req.Offset),
		))
	defer span.End()

	logger.DebugContext(ctx, "Listing entities",
		logger.Fields{
			"limit":  req.Limit,
			"offset": req.Offset,
		})

	// Set default limit if not provided
	if req.Limit <= 0 {
		req.Limit = 50
	}

	// Validate limit
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	timer := s.metrics.Timer("entity_list_duration", metrics.Fields{})

	entities, err := s.repo.List(ctx, &req)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("entity_list_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to list entities",
			logger.Fields{"error": err.Error()})

		return nil, fmt.Errorf("failed to list entities: %w", err)
	}

	// Success metrics
	s.metrics.IncrementCounter("entity_list_requests_total", metrics.Fields{
		"status": "success",
	})

	s.metrics.ObserveHistogram("entity_list_duration",
		duration.Seconds(), metrics.Fields{})

	logger.DebugContext(ctx, "Entities listed successfully",
		logger.Fields{
			"count":       len(entities),
			"duration_ms": duration.Milliseconds(),
		})

	return entities, nil
}

// GetEntityChildren retrieves direct children of an entity
func (s *service) GetEntityChildren(ctx context.Context, entityID uuid.UUID) ([]*Entity, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_entity_children",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.id", entityID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting entity children",
		logger.Fields{"entity_id": entityID.String()})

	// Validate entity exists
	_, err := s.repo.GetByID(ctx, entityID)
	if err != nil {
		return nil, err
	}

	children, err := s.repo.GetChildren(ctx, entityID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get entity children",
			logger.Fields{
				"entity_id": entityID.String(),
				"error":     err.Error(),
			})
		return nil, fmt.Errorf("failed to get entity children: %w", err)
	}

	logger.DebugContext(ctx, "Entity children retrieved successfully",
		logger.Fields{
			"entity_id":      entityID.String(),
			"children_count": len(children),
		})

	return children, nil
}

// GetEntityAncestors retrieves all ancestors of an entity
func (s *service) GetEntityAncestors(ctx context.Context, entityID uuid.UUID) ([]*Entity, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_entity_ancestors",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.id", entityID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting entity ancestors",
		logger.Fields{"entity_id": entityID.String()})

	// Validate entity exists
	_, err := s.repo.GetByID(ctx, entityID)
	if err != nil {
		return nil, err
	}

	ancestors, err := s.repo.GetAncestors(ctx, entityID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get entity ancestors",
			logger.Fields{
				"entity_id": entityID.String(),
				"error":     err.Error(),
			})
		return nil, fmt.Errorf("failed to get entity ancestors: %w", err)
	}

	logger.DebugContext(ctx, "Entity ancestors retrieved successfully",
		logger.Fields{
			"entity_id":       entityID.String(),
			"ancestors_count": len(ancestors),
		})

	return ancestors, nil
}

// GetEntityWithHierarchy retrieves entity with hierarchy information
func (s *service) GetEntityWithHierarchy(ctx context.Context, entityID uuid.UUID) (*EntityWithHierarchy, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_entity_with_hierarchy",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.id", entityID.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting entity with hierarchy",
		logger.Fields{"entity_id": entityID.String()})

	entityWithHierarchy, err := s.repo.GetEntityWithHierarchy(ctx, entityID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get entity with hierarchy",
			logger.Fields{
				"entity_id": entityID.String(),
				"error":     err.Error(),
			})
		return nil, fmt.Errorf("failed to get entity with hierarchy: %w", err)
	}

	logger.DebugContext(ctx, "Entity with hierarchy retrieved successfully",
		logger.Fields{
			"entity_id":    entityID.String(),
			"entity_level": entityWithHierarchy.Level,
		})

	return entityWithHierarchy, nil
}

// GetEntityTree retrieves the complete entity tree
func (s *service) GetEntityTree(ctx context.Context) ([]*EntityWithHierarchy, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_entity_tree",
		tracing.WithSpanKind(tracing.SpanKindInternal))
	defer span.End()

	logger.DebugContext(ctx, "Getting entity tree")

	timer := s.metrics.Timer("entity_tree_duration", metrics.Fields{})

	entityTree, err := s.repo.GetEntityTree(ctx)
	duration := timer.Stop()

	if err != nil {
		s.metrics.IncrementCounter("entity_tree_errors", metrics.Fields{
			"error_type": "repository_error",
		})

		logger.ErrorContext(ctx, "Failed to get entity tree",
			logger.Fields{"error": err.Error()})

		return nil, fmt.Errorf("failed to get entity tree: %w", err)
	}

	// Success metrics
	s.metrics.IncrementCounter("entity_tree_requests_total", metrics.Fields{
		"status": "success",
	})

	s.metrics.ObserveHistogram("entity_tree_duration",
		duration.Seconds(), metrics.Fields{})

	logger.DebugContext(ctx, "Entity tree retrieved successfully",
		logger.Fields{
			"entities_count": len(entityTree),
			"duration_ms":    duration.Milliseconds(),
		})

	return entityTree, nil
}

// GetNextSequence gets the next sequence number for an entity
func (s *service) GetNextSequence(ctx context.Context, req EntitySequenceRequest) (int64, error) {
	ctx, span := s.tracing.StartSpan(ctx, "service.get_next_sequence",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.id", req.EntityID.String()),
			attribute.String("sequence.key", req.Key),
			attribute.Int("sequence.fiscal_year", req.FiscalYear),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting next sequence number",
		logger.Fields{
			"entity_id":   req.EntityID.String(),
			"key":         req.Key,
			"fiscal_year": req.FiscalYear,
		})

	// Validate entity exists
	_, err := s.repo.GetByID(ctx, req.EntityID)
	if err != nil {
		return 0, err
	}

	// Get next sequence
	sequence, err := s.repo.GetNextSequence(ctx, req.EntityID, req.Key, req.FiscalYear)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get next sequence",
			logger.Fields{
				"entity_id":   req.EntityID.String(),
				"key":         req.Key,
				"fiscal_year": req.FiscalYear,
				"error":       err.Error(),
			})
		return 0, fmt.Errorf("failed to get next sequence: %w", err)
	}

	logger.DebugContext(ctx, "Next sequence retrieved successfully",
		logger.Fields{
			"entity_id":   req.EntityID.String(),
			"key":         req.Key,
			"fiscal_year": req.FiscalYear,
			"sequence":    sequence,
		})

	return sequence, nil
}

// ResetSequence resets the sequence number for an entity
func (s *service) ResetSequence(ctx context.Context, req EntitySequenceRequest) error {
	ctx, span := s.tracing.StartSpan(ctx, "service.reset_sequence",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(
			attribute.String("entity.id", req.EntityID.String()),
			attribute.String("sequence.key", req.Key),
			attribute.Int("sequence.fiscal_year", req.FiscalYear),
		))
	defer span.End()

	logger.InfoContext(ctx, "Resetting sequence number",
		logger.Fields{
			"entity_id":   req.EntityID.String(),
			"key":         req.Key,
			"fiscal_year": req.FiscalYear,
		})

	// Validate entity exists
	_, err := s.repo.GetByID(ctx, req.EntityID)
	if err != nil {
		return err
	}

	// Reset sequence
	err = s.repo.ResetSequence(ctx, req.EntityID, req.Key, req.FiscalYear)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to reset sequence",
			logger.Fields{
				"entity_id":   req.EntityID.String(),
				"key":         req.Key,
				"fiscal_year": req.FiscalYear,
				"error":       err.Error(),
			})
		return fmt.Errorf("failed to reset sequence: %w", err)
	}

	logger.InfoContext(ctx, "Sequence reset successfully",
		logger.Fields{
			"entity_id":   req.EntityID.String(),
			"key":         req.Key,
			"fiscal_year": req.FiscalYear,
		})

	return nil
}
