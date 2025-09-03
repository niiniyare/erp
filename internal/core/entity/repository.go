package entity

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"
import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/convert"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Repository defines the interface for entity data operations
type Repository interface {
	// Entity CRUD operations
	Create(ctx context.Context, entity *CreateEntityRequest) (*Entity, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Entity, error)
	GetByCode(ctx context.Context, code string) (*Entity, error)
	GetByName(ctx context.Context, name string) (*Entity, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateEntityRequest) (*Entity, error)
	Delete(ctx context.Context, id uuid.UUID, permanent bool) error
	Restore(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *ListEntitiesRequest) ([]*Entity, error)

	// Validation operations
	ValidateEntityName(ctx context.Context, name string, excludeID *uuid.UUID) error
	ValidateEntityCode(ctx context.Context, code string, excludeID *uuid.UUID) error
	ValidateEntityParent(ctx context.Context, parentID, childID uuid.UUID) error
	IsEntityAncestor(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error)

	// Hierarchy operations
	GetChildren(ctx context.Context, ancestorID uuid.UUID) ([]*Entity, error)
	GetAncestors(ctx context.Context, descendantID uuid.UUID) ([]*Entity, error)
	GetEntityWithHierarchy(ctx context.Context, entityID uuid.UUID) (*EntityWithHierarchy, error)
	GetEntityTree(ctx context.Context) ([]*EntityWithHierarchy, error)
	MoveEntityToNewParent(ctx context.Context, entityID, newParentID uuid.UUID) error

	// Entity state operations
	GetNextSequence(ctx context.Context, entityID uuid.UUID, key string, fiscalYear int) (int64, error)
	ResetSequence(ctx context.Context, entityID uuid.UUID, key string, fiscalYear int) error
}

// repository implements the Repository interface
type repository struct {
	store   db.Store
	tracing tracing.TracingService
	metrics metrics.MetricsProvider
}

// createNoopMetrics creates a noop metrics provider
func createNoopMetrics() metrics.MetricsProvider {
	service, _ := metrics.NewMetricsService(metrics.MetricsConfig{
		Enabled: false,
	})
	return service
}

// NewRepository creates a new entity repository
func NewRepository(store db.Store, tracing tracing.TracingService, metric metrics.MetricsProvider) Repository {
	return &repository{
		store:   store,
		tracing: tracing,
		metrics: metric, // Use noop metrics if not provided
	}
}

// Create creates a new entity
func (r *repository) Create(ctx context.Context, req *CreateEntityRequest) (*Entity, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.create_entity",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "create"),
			attribute.String("db.table", "entities"),
			attribute.String("entity.name", req.Name),
			attribute.String("entity.code", req.Code),
			attribute.String("entity.type", string(req.Type)),
		))
	defer span.End()

	logger.DebugContext(ctx, "Creating entity in database",
		logger.Fields{
			"entity_name": req.Name,
			"entity_code": req.Code,
			"entity_type": string(req.Type),
		})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "create_entity",
		"table":     "entities",
	})

	params, err := req.ToSQLCCreateParams()
	if err != nil {
		timer.Stop()
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to convert request to SQLC params")
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Generate UUID for new entity
	entityID := uuid.New()
	params.Uuid = entityID

	sqlcEntity, err := r.store.CreateEntity(ctx, params)
	timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation":  "create_entity",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":     err.Error(),
				"operation": "create_entity",
			})

		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	// Create hierarchy paths if parent exists
	if req.ParentID != nil {
		if err := r.createHierarchyPaths(ctx, *req.ParentID, entityID); err != nil {
			// Log error but don't fail the creation
			logger.ErrorContext(ctx, "Failed to create hierarchy paths",
				logger.Fields{
					"entity_id": entityID.String(),
					"parent_id": req.ParentID.String(),
					"error":     err.Error(),
				})
		}
	}

	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "create_entity",
		"status":    "success",
	})

	logger.DebugContext(ctx, "Entity created successfully in database",
		logger.Fields{
			"entity_id": entityID.String(),
		})

	return FromSQLCEntity(sqlcEntity)
}

// GetByID retrieves an entity by ID
func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_entity_by_id",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "entities"),
			attribute.String("entity.id", id.String()),
		))
	defer span.End()

	logger.DebugContext(ctx, "Getting entity by ID",
		logger.Fields{"entity_id": id.String()})

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_entity_by_id",
		"table":     "entities",
	})

	sqlcEntity, err := r.store.GetEntity(ctx, id)
	timer.Stop()

	if err != nil {
		if err.Error() == "no rows in result set" {
			span.SetStatus(codes.Ok, "Entity not found")
			return nil, errors.ErrEntityNotFound
		}

		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation":  "get_entity_by_id",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		logger.ErrorContext(ctx, "Database operation failed",
			logger.Fields{
				"error":     err.Error(),
				"operation": "get_entity_by_id",
			})

		return nil, fmt.Errorf("failed to get entity by ID: %w", err)
	}

	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_entity_by_id",
		"status":    "success",
	})

	logger.DebugContext(ctx, "Entity retrieved successfully",
		logger.Fields{
			"entity_id": id.String(),
		})

	return FromSQLCEntity(sqlcEntity)
}

// GetByCode retrieves an entity by code
func (r *repository) GetByCode(ctx context.Context, code string) (*Entity, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_entity_by_code",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "entities"),
			attribute.String("entity.code", code),
		))
	defer span.End()

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_entity_by_code",
		"table":     "entities",
	})

	sqlcEntity, err := r.store.GetEntityByCode(ctx, &code)
	timer.Stop()

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrEntityNotFound
		}

		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation":  "get_entity_by_code",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		return nil, fmt.Errorf("failed to get entity by code: %w", err)
	}

	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_entity_by_code",
		"status":    "success",
	})

	return FromSQLCEntity(sqlcEntity)
}

// GetByName retrieves an entity by name
func (r *repository) GetByName(ctx context.Context, name string) (*Entity, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_entity_by_name",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "entities"),
			attribute.String("entity.name", name),
		))
	defer span.End()

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "get_entity_by_name",
		"table":     "entities",
	})

	sqlcEntity, err := r.store.GetEntityByName(ctx, name)
	timer.Stop()

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrEntityNotFound
		}

		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation":  "get_entity_by_name",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		return nil, fmt.Errorf("failed to get entity by name: %w", err)
	}

	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "get_entity_by_name",
		"status":    "success",
	})

	return FromSQLCEntity(sqlcEntity)
}

// Update updates an existing entity
func (r *repository) Update(ctx context.Context, id uuid.UUID, req *UpdateEntityRequest) (*Entity, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.update_entity",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "update"),
			attribute.String("db.table", "entities"),
			attribute.String("entity.id", id.String()),
		))
	defer span.End()

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "update_entity",
		"table":     "entities",
	})

	// Handle parent changes if provided
	if req.ParentID != nil {
		if err := r.handleParentChange(ctx, id, *req.ParentID); err != nil {
			timer.Stop()
			span.RecordError(err)
			return nil, err
		}
	}

	// Build update parameters
	params := r.buildUpdateParams(id, req)

	sqlcEntity, err := r.store.UpdateEntity(ctx, params)
	timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation":  "update_entity",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "update_entity",
		"status":    "success",
	})

	return FromSQLCEntity(sqlcEntity)
}

// Delete deletes an entity (soft or hard delete)
func (r *repository) Delete(ctx context.Context, id uuid.UUID, permanent bool) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.delete_entity",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "delete"),
			attribute.String("db.table", "entities"),
			attribute.String("entity.id", id.String()),
			attribute.Bool("permanent", permanent),
		))
	defer span.End()

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "delete_entity",
		"table":     "entities",
	})

	var err error
	if permanent {
		err = r.hardDeleteEntity(ctx, id)
	} else {
		err = r.softDeleteEntity(ctx, id)
	}

	timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation":  "delete_entity",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		return err
	}

	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "delete_entity",
		"status":    "success",
	})

	return nil
}

// Restore restores a soft-deleted entity
func (r *repository) Restore(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.restore_entity",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "update"),
			attribute.String("db.table", "entities"),
			attribute.String("entity.id", id.String()),
		))
	defer span.End()

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "restore_entity",
		"table":     "entities",
	})

	err := r.store.RestoreEntity(ctx, id)
	timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation":  "restore_entity",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		return fmt.Errorf("failed to restore entity: %w", err)
	}

	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "restore_entity",
		"status":    "success",
	})

	return nil
}

// List retrieves entities with filtering and pagination
func (r *repository) List(ctx context.Context, req *ListEntitiesRequest) ([]*Entity, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.list_entities",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "entities"),
			attribute.Int("limit", req.Limit),
			attribute.Int("offset", req.Offset),
		))
	defer span.End()

	timer := r.metrics.Timer("database_query_duration", metrics.Fields{
		"operation": "list_entities",
		"table":     "entities",
	})

	sqlcEntities, err := r.store.ListEntities(ctx)
	timer.Stop()

	if err != nil {
		r.metrics.IncrementCounter("database_errors_total", metrics.Fields{
			"operation":  "list_entities",
			"error_type": "sql_error",
		})

		span.RecordError(err)
		span.SetStatus(codes.Error, "Database operation failed")

		return nil, fmt.Errorf("failed to list entities: %w", err)
	}

	entities := make([]*Entity, len(sqlcEntities))
	for i, sqlcEntity := range sqlcEntities {
		entity, err := FromSQLCEntity(sqlcEntity)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entity: %w", err)
		}
		entities[i] = entity
	}

	r.metrics.IncrementCounter("database_operations_total", metrics.Fields{
		"operation": "list_entities",
		"status":    "success",
	})

	return entities, nil
}

// ValidateEntityName validates if entity name is unique
func (r *repository) ValidateEntityName(ctx context.Context, name string, excludeID *uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.validate_entity_name")
	defer span.End()

	var excludeUUID uuid.UUID
	if excludeID != nil {
		excludeUUID = *excludeID
	} else {
		excludeUUID = uuid.Nil
	}

	params := db.ValidateEntityNameParams{
		Name: name,
		Uuid: excludeUUID,
	}

	exists, err := r.store.ValidateEntityName(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to validate entity name: %w", err)
	}

	if exists {
		return errors.ErrEntityNameExists
	}

	return nil
}

// ValidateEntityCode validates if entity code is unique
func (r *repository) ValidateEntityCode(ctx context.Context, code string, excludeID *uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.validate_entity_code")
	defer span.End()

	var excludeUUID uuid.UUID
	if excludeID != nil {
		excludeUUID = *excludeID
	} else {
		excludeUUID = uuid.Nil
	}

	params := db.ValidateEntityCodeParams{
		Code: &code,
		Uuid: excludeUUID,
	}

	exists, err := r.store.ValidateEntityCode(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to validate entity code: %w", err)
	}

	if exists {
		return errors.ErrEntityCodeExists
	}

	return nil
}

// ValidateEntityParent validates if parent is valid
func (r *repository) ValidateEntityParent(ctx context.Context, parentID, childID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.validate_entity_parent")
	defer span.End()

	// Check if parent exists
	_, err := r.store.GetEntity(ctx, parentID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return errors.ErrInvalidParentEntity
		}
		return fmt.Errorf("failed to validate parent: %w", err)
	}

	return nil
}

// IsEntityAncestor checks if one entity is ancestor of another
func (r *repository) IsEntityAncestor(ctx context.Context, ancestorID, descendantID uuid.UUID) (bool, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.is_entity_ancestor")
	defer span.End()

	exists, err := r.store.IsEntityAncestor(ctx, db.IsEntityAncestorParams{
		AncestorID:   ancestorID,
		DescendantID: descendantID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check ancestor relationship: %w", err)
	}

	return exists, nil
}

// GetChildren retrieves direct children of an entity
func (r *repository) GetChildren(ctx context.Context, ancestorID uuid.UUID) ([]*Entity, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_entity_children")
	defer span.End()

	sqlcEntities, err := r.store.GetEntityChildren(ctx, ancestorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity children: %w", err)
	}

	entities := make([]*Entity, len(sqlcEntities))
	for i, sqlcEntity := range sqlcEntities {
		entity, err := FromSQLCEntity(sqlcEntity)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entity: %w", err)
		}
		entities[i] = entity
	}

	return entities, nil
}

// GetAncestors retrieves all ancestors of an entity
func (r *repository) GetAncestors(ctx context.Context, descendantID uuid.UUID) ([]*Entity, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_entity_ancestors")
	defer span.End()

	sqlcRows, err := r.store.GetEntityAncestors(ctx, descendantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity ancestors: %w", err)
	}

	entities := make([]*Entity, len(sqlcRows))
	for i, row := range sqlcRows {
		// Convert the GetEntityAncestorsRow to db.Entity first
		dbEntity := &db.Entity{
			Uuid:          row.Uuid,
			TenantID:      row.TenantID,
			ParentID:      row.ParentID,
			Name:          row.Name,
			Code:          row.Code,
			Type:          row.Type,
			IsActive:      row.IsActive,
			Hidden:        row.Hidden,
			AccrualMethod: row.AccrualMethod,
			FyStartMonth:  row.FyStartMonth,
			Address:       row.Address,
			Picture:       row.Picture,
			Settings:      row.Settings,
			Metadata:      row.Metadata,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
			DeletedAt:     row.DeletedAt,
		}

		entity, err := FromSQLCEntity(dbEntity)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entity: %w", err)
		}
		entities[i] = entity
	}

	return entities, nil
}

// GetEntityWithHierarchy retrieves entity with hierarchy information
func (r *repository) GetEntityWithHierarchy(ctx context.Context, entityID uuid.UUID) (*EntityWithHierarchy, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_entity_with_hierarchy")
	defer span.End()

	// Get the entity
	entity, err := r.GetByID(ctx, entityID)
	if err != nil {
		return nil, err
	}

	// Get hierarchy information
	// Using a dummy tenant ID since the function uses current_tenant_id() in the query
	hierarchyInfo, err := r.store.GetEntityWithHierarchyInfo(ctx, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hierarchy info: %w", err)
	}

	// Convert any to int for Level
	level, ok := hierarchyInfo.Level.(int64)
	if !ok {
		level = 0
	}

	return &EntityWithHierarchy{
		Entity:      *entity,
		Level:       int(level),
		Path:        "root", // TODO: implement proper path from query
		HasChildren: hierarchyInfo.ChildCount > 0,
	}, nil
}

// GetEntityTree retrieves the complete entity tree
func (r *repository) GetEntityTree(ctx context.Context) ([]*EntityWithHierarchy, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_entity_tree")
	defer span.End()

	sqlcRows, err := r.store.GetEntityTreeStructure(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity tree: %w", err)
	}

	entities := make([]*EntityWithHierarchy, len(sqlcRows))
	for i, row := range sqlcRows {
		// Convert the GetEntityTreeStructureRow to db.Entity first
		dbEntity := &db.Entity{
			Uuid:          row.Uuid,
			TenantID:      row.TenantID,
			ParentID:      row.ParentID,
			Name:          row.Name,
			Code:          row.Code,
			Type:          row.Type,
			IsActive:      row.IsActive,
			Hidden:        row.Hidden,
			AccrualMethod: row.AccrualMethod,
			FyStartMonth:  row.FyStartMonth,
			Address:       row.Address,
			Picture:       row.Picture,
			Settings:      row.Settings,
			Metadata:      row.Metadata,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
			DeletedAt:     row.DeletedAt,
		}

		entity, err := FromSQLCEntity(dbEntity)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entity: %w", err)
		}

		entities[i] = &EntityWithHierarchy{
			Entity:      *entity,
			Level:       int(row.Level),
			Path:        row.PathText,
			HasChildren: false, // TODO: calculate from children count
		}
	}

	return entities, nil
}

// MoveEntityToNewParent moves an entity to a new parent
func (r *repository) MoveEntityToNewParent(ctx context.Context, entityID, newParentID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.move_entity_to_new_parent")
	defer span.End()

	return r.store.WithTx(ctx, func(ctx context.Context, tx db.Store) error {
		// Delete old hierarchy paths
		if err := tx.DeleteHierarchyPaths(ctx, entityID); err != nil {
			return fmt.Errorf("failed to delete old hierarchy paths: %w", err)
		}

		// Update parent - using the MoveEntityToNewParent method which handles both hierarchy and parent update
		if err := tx.MoveEntityToNewParent(ctx, db.MoveEntityToNewParentParams{
			Uuid:     entityID,
			ParentID: &newParentID,
		}); err != nil {
			return fmt.Errorf("failed to move entity to new parent: %w", err)
		}

		return nil
	})
}

// GetNextSequence gets the next sequence number for an entity
func (r *repository) GetNextSequence(ctx context.Context, entityID uuid.UUID, key string, fiscalYear int) (int64, error) {
	ctx, span := r.tracing.StartSpan(ctx, "repository.get_next_sequence")
	defer span.End()

	safeFiscalYear, err := convert.IntToInt16(fiscalYear)
	if err != nil {
		return 0, fmt.Errorf("invalid fiscal year: %w", err)
	}
	fiscalYearPtr := &safeFiscalYear

	sequence, err := r.store.GetNextSequenceNumber(ctx, db.GetNextSequenceNumberParams{
		EntityID:   entityID,
		Key:        key,
		FiscalYear: fiscalYearPtr,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence: %w", err)
	}

	return sequence, nil
}

// ResetSequence resets the sequence number for an entity
func (r *repository) ResetSequence(ctx context.Context, entityID uuid.UUID, key string, fiscalYear int) error {
	ctx, span := r.tracing.StartSpan(ctx, "repository.reset_sequence")
	defer span.End()

	safeFiscalYear, err := convert.IntToInt16(fiscalYear)
	if err != nil {
		return fmt.Errorf("invalid fiscal year: %w", err)
	}
	err = r.store.ResetAllEntitySequences(ctx, db.ResetAllEntitySequencesParams{
		EntityID:   entityID,
		FiscalYear: &safeFiscalYear,
	})
	if err != nil {
		return fmt.Errorf("failed to reset sequence: %w", err)
	}

	return nil
}

// Helper methods

func (r *repository) createHierarchyPaths(ctx context.Context, parentID, childID uuid.UUID) error {
	// Create direct path
	if err := r.store.CreateHierarchyPath(ctx, db.CreateHierarchyPathParams{
		AncestorID:   parentID,
		DescendantID: childID,
		Depth:        1,
	}); err != nil {
		return fmt.Errorf("failed to create direct hierarchy path: %w", err)
	}

	// Update all ancestor paths
	if err := r.store.UpdateHierarchyPaths(ctx, childID); err != nil {
		return fmt.Errorf("failed to update hierarchy paths: %w", err)
	}

	return nil
}

func (r *repository) handleParentChange(ctx context.Context, entityID, newParentID uuid.UUID) error {
	// Get current parent
	currentParent, err := r.store.GetEntityParent(ctx, entityID)
	if err != nil && err.Error() != "no rows in result set" {
		return fmt.Errorf("failed to get current parent: %w", err)
	}

	// Check if parent is actually changing
	if currentParent != nil && currentParent.ParentID != nil && *currentParent.ParentID == newParentID {
		return nil // No change needed
	}

	// Validate new parent
	if err := r.ValidateEntityParent(ctx, newParentID, entityID); err != nil {
		return err
	}

	// Check for circular reference
	isAncestor, err := r.IsEntityAncestor(ctx, entityID, newParentID)
	if err != nil {
		return fmt.Errorf("failed to check circular reference: %w", err)
	}
	if isAncestor {
		return errors.ErrCircularReference
	}

	// Move entity to new parent
	return r.MoveEntityToNewParent(ctx, entityID, newParentID)
}

func (r *repository) buildUpdateParams(id uuid.UUID, req *UpdateEntityRequest) db.UpdateEntityParams {
	// Get current entity to fill in missing fields
	currentEntity, err := r.store.GetEntity(context.Background(), id)
	if err != nil {
		// Return default params if we can't get current entity
		return db.UpdateEntityParams{
			Uuid:     id,
			Name:     "",
			Code:     nil,
			Type:     "other",
			IsActive: true,
			Hidden:   false,
		}
	}

	params := db.UpdateEntityParams{
		Uuid:          id,
		Name:          currentEntity.Name,
		Code:          currentEntity.Code,
		Type:          currentEntity.Type,
		IsActive:      currentEntity.IsActive,
		Hidden:        currentEntity.Hidden,
		AccrualMethod: currentEntity.AccrualMethod,
		FyStartMonth:  currentEntity.FyStartMonth,
		Address:       currentEntity.Address,
		Picture:       currentEntity.Picture,
		Settings:      currentEntity.Settings,
	}

	// Update only provided fields
	if req.Name != nil {
		params.Name = *req.Name
	}

	if req.Code != nil {
		params.Code = req.Code
	}

	if req.Type != nil {
		params.Type = string(*req.Type)
	}

	if req.IsActive != nil {
		params.IsActive = *req.IsActive
	}

	if req.IsHidden != nil {
		params.Hidden = *req.IsHidden
	}

	if req.Metadata != nil {
		metadata, _ := json.Marshal(req.Metadata)
		params.Settings = metadata
	}

	return params
}

func (r *repository) softDeleteEntity(ctx context.Context, id uuid.UUID) error {
	return r.store.SoftDeleteEntity(ctx, id)
}

func (r *repository) hardDeleteEntity(ctx context.Context, id uuid.UUID) error {
	return r.store.WithTx(ctx, func(ctx context.Context, tx db.Store) error {
		// Delete hierarchy paths
		if err := tx.DeleteHierarchyPaths(ctx, id); err != nil {
			return fmt.Errorf("failed to delete hierarchy paths: %w", err)
		}

		// Hard delete entity
		if err := tx.HardDeleteEntity(ctx, id); err != nil {
			return fmt.Errorf("failed to hard delete entity: %w", err)
		}

		return nil
	})
}
